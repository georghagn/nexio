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

// RotationPolicy entscheidet, WANN rotiert wird.
type RotationPolicy interface {

	// ShouldRotate prüft anhand der aktuellen Größe und Öffnungszeit, ob rotiert werden muss.
	ShouldRotate(currentSize int64, openTime time.Duration) bool
}

// ArchiveStrategy entscheidet, WIE das alte File behandelt wird (z.B. Zippen).
// Sie nimmt den alten Pfad, verarbeitet ihn und gibt den neuen Pfad zurück.
type ArchiveStrategy interface {
	Archive(filePath string) (string, error)
}

// RetentionPolicy entscheidet, WELCHE alten Dateien gelöscht werden (Aufräumen).
type RetentionPolicy interface {
	Prune(baseFilename string) error
}

// Archive Strategies //
//--------------------//

// NoCompression: Einfaches Umbenennen (log.1, log.2 oder timestamp)
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

// GzipCompression: Zippen des Logs
type GzipCompression struct{}

func (s *GzipCompression) Archive(path string) (string, error) {

	timestamp := time.Now().Format("20060102-150405.000")
	gzName := fmt.Sprintf("%s-%s.gz", path, timestamp)

	// 1. Input Datei öffnen
	inFile, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer inFile.Close()

	// 2. Output Datei (.gz) erstellen
	outFile, err := os.Create(gzName)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	// 3. Gzip Writer
	gw := gzip.NewWriter(outFile)
	defer gw.Close()

	// 4. Kopieren
	if _, err := io.Copy(gw, inFile); err != nil {
		return "", err
	}

	// Wichtig: Explizit schließen, damit Gzip Footer geschrieben wird
	gw.Close()
	inFile.Close()

	// 5. Originaldatei löschen (da jetzt gezippt)
	os.Remove(path)

	return gzName, nil
}

// Retentaion Policies  //
// ---------------------//

// SizePolicy: Rotiert bei Größe X
type SizePolicy struct {
	MaxBytes int64
}

// DailyPolicy: Rotiert alle 24h (stark vereinfacht für Tiny-Zwecke)
type DailyPolicy struct{}

func (p *SizePolicy) ShouldRotate(currentSize int64, openDuration time.Duration) bool {
	return currentSize >= p.MaxBytes
}

func (p *DailyPolicy) ShouldRotate(size int64, openDuration time.Duration) bool {
	return openDuration >= 24*time.Hour
}

// Rotation Policies //
//-------------------//

// KeepAll: Behält alles (Default)
type KeepAll struct{}

func (k *KeepAll) Prune(base string) error { return nil }

// MaxFiles: Behält nur die neuesten N Dateien
type MaxFiles struct {
	MaxBackups int
}

func (m *MaxFiles) Prune(baseFilename string) error {

	dir := filepath.Dir(baseFilename)
	baseName := filepath.Base(baseFilename)
	prefix := strings.TrimSuffix(baseName, filepath.Ext(baseName)) // z.B. "app" von "app.log"

	// 1. Alle Dateien im Ordner lesen
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	// 2. Filtern: Nur unsere Logfiles (die mit dem Prefix anfangen und NICHT das aktuelle sind)
	var backups []fs.DirEntry
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if e.Name() == baseName {
			continue
		} // Das aktuelle aktive Log überspringen
		if strings.HasPrefix(e.Name(), prefix) {
			backups = append(backups, e)
		}
	}

	// 3. Wenn weniger als Limit -> Fertig
	if len(backups) <= m.MaxBackups {
		return nil
	}

	// 4. Sortieren nach Mod-Time (Älteste zuerst)
	// Da Namen Zeitstempel enthalten, reicht oft Namenssortierung, aber ModTime ist sicherer
	sort.Slice(backups, func(i, j int) bool {
		infoI, _ := backups[i].Info()
		infoJ, _ := backups[j].Info()
		return infoI.ModTime().Before(infoJ.ModTime())
	})

	// 5. Löschen (Die Ältesten, bis wir wieder im Limit sind)
	toDelete := len(backups) - m.MaxBackups
	for i := 0; i < toDelete; i++ {
		fullPath := filepath.Join(dir, backups[i].Name())
		os.Remove(fullPath)
	}

	return nil
}
