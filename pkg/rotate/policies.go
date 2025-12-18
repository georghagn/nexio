// Copyright 2025 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

package rotate

import (
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// RotationPolicy decides WHEN rotation occurs.
type RotationPolicy interface {

	// ShouldRotate checks whether rotation is necessary based on the current size and opening time.
	ShouldRotate(currentSize int64, openTime time.Duration) bool
}

// ArchiveStrategy decides HOW the old file is handled (e.g., zipped).
// It takes the old path, processes it, and returns the new path.
type ArchiveStrategy interface {
	Archive(filePath string) (string, error)
}

// RetentionPolicy decides WHICH old files are deleted (cleanup).
type RetentionPolicy interface {
	Prune(baseFilename string) error
}

// Archive Strategies //
//--------------------//

// NoCompression: Simple rename (log.1, log.2 or timestamp)
type NoCompression struct{}

func (s *NoCompression) Archive(path string) (string, error) {

	// Format: name-20231027-150405.log
	timestamp := time.Now().Format("20060102-150405.000")
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)
	newName := fmt.Sprintf("%s-%s%s", base, timestamp, ext)

	err := os.Rename(path, newName)
	return newName, err
}

// GzipCompression: Zip Logs
type GzipCompression struct{}

func (s *GzipCompression) Archive(path string) (string, error) {

	timestamp := time.Now().Format("20060102-150405.000")
	gzName := fmt.Sprintf("%s-%s.gz", path, timestamp)

	// 1. Open input file
	inFile, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer inFile.Close()

	// 2. create output file (.gz)
	outFile, err := os.Create(gzName)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	// 3. Gzip Writer
	gw := gzip.NewWriter(outFile)
	defer gw.Close()

	// 4. Copy
	if _, err := io.Copy(gw, inFile); err != nil {
		return "", err
	}

	// Important: Explicitly close the browser so that the Gzip footer is written.
	gw.Close()
	inFile.Close()

	// 5. Important: Explicitly delete the original file (since it's now zipped)
	// so that a Gzip footer is written.
	os.Remove(path)

	return gzName, nil
}

// Retentaion Policies  //
// ---------------------//

// SizePolicy: Rotates at size X
type SizePolicy struct {
	MaxBytes int64
}

// DailyPolicy: Rotates every 24 hours (greatly simplified for Tiny purposes)
type DailyPolicy struct{}

func (p *SizePolicy) ShouldRotate(currentSize int64, openDuration time.Duration) bool {
	return currentSize >= p.MaxBytes
}

func (p *DailyPolicy) ShouldRotate(size int64, openDuration time.Duration) bool {
	return openDuration >= 24*time.Hour
}

// Rotation Policies //
//-------------------//

// KeepAll: keep everything (Default)
type KeepAll struct{}

func (k *KeepAll) Prune(base string) error { return nil }

// MaxFiles: Keep only newest N files
type MaxFiles struct {
	MaxBackups int
}

func (m *MaxFiles) Prune(baseFilename string) error {

	dir := filepath.Dir(baseFilename)
	baseName := filepath.Base(baseFilename)
	prefix := strings.TrimSuffix(baseName, filepath.Ext(baseName)) // z.B. "app" of "app.log"

	// 1. Read all files in the folder
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	// 2. Filter: Only our log files (those that start with the prefix and are NOT the current one)
	var backups []fs.DirEntry
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if e.Name() == baseName {
			continue
		} // Skip the current active log
		if strings.HasPrefix(e.Name(), prefix) {
			backups = append(backups, e)
		}
	}

	// 3. If less than the limit -> Done
	if len(backups) <= m.MaxBackups {
		return nil
	}

	// 4. Sort by Mod-Time (Oldest First)
	//    Since names contain timestamps, sorting by name is often sufficient, but Mod-Time is safer.
	sort.Slice(backups, func(i, j int) bool {
		infoI, _ := backups[i].Info()
		infoJ, _ := backups[j].Info()
		return infoI.ModTime().Before(infoJ.ModTime())
	})

	// 5. Delete (The oldest ones, until we're back within the limit)
	toDelete := len(backups) - m.MaxBackups
	for i := 0; i < toDelete; i++ {
		fullPath := filepath.Join(dir, backups[i].Name())
		os.Remove(fullPath)
	}

	return nil
}
