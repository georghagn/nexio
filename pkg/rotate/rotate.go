// Copyright 2025 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

// The rotator assumes exclusive ownership of the target file within the current process.
// Further Implemetations may change this. (DESIGN.md)
package rotate

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/georghagn/nexio/pkg/rotate/lock"
)

type RotateReason string

const (
	RotateManual   RotateReason = "manual"
	RotateSchedule RotateReason = "schedule"
	RotateSize     RotateReason = "size"
	RotateSignal   RotateReason = "signal"
)

type Writer struct {
	filename      string
	file          *os.File
	currentSize   int64
	openTimestamp time.Time // When was the current file opened?

	// The strategies
	Rotation  RotationPolicy
	Archive   ArchiveStrategy
	Retention RetentionPolicy

	fileLock *lock.FileLock
	mu       sync.Mutex
}

// New creates a writer with default strategies (if nil is passed).
func New(filename string, r RotationPolicy, a ArchiveStrategy, ret RetentionPolicy) *Writer {

	// Set defaults for "Tiny" usage (Convention over Configuration)
	if r == nil {
		r = &SizePolicy{MaxBytes: 10 * 1024 * 1024}
	} // 10MB

	if a == nil {
		a = &NoCompression{}
	} // rename only

	if ret == nil {
		ret = &KeepAll{}
	} // delete nothing

	return &Writer{
		filename:  filename,
		Rotation:  r,
		Archive:   a,
		Retention: ret,
	}
}

func (w *Writer) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 1. Open file, if necessary
	if w.file == nil {
		if err := w.openLocked(); err != nil {
			return 0, err
		}
	}

	// 2. Check if rotation is necessary (delegation to policy)
	// We hypothetically add len(p) to see if we would exceed the limit.
	timeOpen := time.Since(w.openTimestamp)
	if w.Rotation.ShouldRotate(w.currentSize+int64(len(p)), timeOpen) {
		if err := w.rotateLocked(); err != nil {
			return 0, err
		}
	}

	// 3. Write
	n, err = w.file.Write(p)
	w.currentSize += int64(n)
	return n, err
}

func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.closeLocked()
}

// RotateNow forces a log rotation regardless of the configured policy.
// It is safe to call concurrently.
func (w *Writer) RotateNow() error {
	var err error

	w.mu.Lock()
	defer w.mu.Unlock()

	fl := &lock.FileLock{
		Path:    fmt.Sprintf("%s.LOCK", w.filename),
		Timeout: 1 * time.Second,
		Expiry:  2 * time.Minute, // Falls ein Prozess crasht, nach 2 Min freigeben
	}

	if err = fl.Lock(); err == nil {
		defer fl.Unlock()
		return w.rotateLocked() // Safe to rotate now
	}
	return err
}

// --- Internal Logic ---

func (w *Writer) openLocked() error {

	dir := filepath.Dir(w.filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	f, err := os.OpenFile(w.filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	w.file = f
	w.openTimestamp = time.Now()

	info, err := f.Stat()
	if err == nil {
		w.currentSize = info.Size()
	}
	return nil
}

func (w *Writer) closeLocked() error {
	if w.file == nil {
		return nil
	}
	err := w.file.Close()
	w.file = nil
	w.currentSize = 0
	return err
}

// rotateLocked performs an in-process log rotation.
// IMPORTANT: Files must be closed before renaming.
// Renaming an open file keeps the file descriptor valid,
// causing further writes to go into the archived file.
func (w *Writer) rotateLocked() error {

	if w.file == nil {
		// Nothing to rotate yet
		return nil
	}

	// 1. Close actual file
	if err := w.closeLocked(); err != nil {
		return err
	}

	// 2. Archiving (e.g., renaming or zipping)
	if _, err := w.Archive.Archive(w.filename); err != nil {
		return err // If archiving fails, we'll try again on the next write attempt.
	}

	// 3. Cleanup (Delete old backups)
	// We are deliberately ignoring errors here; deleting logs is a "best effort" approach.
	_ = w.Retention.Prune(w.filename)

	// 4. Open new file
	return w.openLocked()
}
