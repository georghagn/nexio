// Copyright 2025 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

// Design Note:
//
// The Rotator keeps the target file open for the lifetime of the writer.
// Opening and closing the file on every write is intentionally avoided.
// A stable file descriptor is required for correct rotation semantics,
// atomic writes and predictable error handling.
//
// The Reopener is an explicit recovery mechanism, not part of the write path.
// It is intended to be triggered by external events (e.g. SIGHUP, logrotate,
// administrative action) or by higher-level supervision logic.
//
// Write errors are returned to the caller and are not silently recovered
// by implicit reopen or retry logic.
package nexlog

import (
	"os"
	"sync"
)

// ReopenableWriter is an io.WriteCloser that can be reopened at runtime.
// This is useful for logrotate strategies (e.g., via scheduler or SIGHUP),
// where the file is moved externally and the process needs to renew the file handle.
type ReopenableWriter struct {
	filename string
	file     *os.File
	mu       sync.Mutex
}

// NewReopenableWriter opens the file and prepares the writing.
func NewReopenableWriter(filename string) (*ReopenableWriter, error) {
	w := &ReopenableWriter{filename: filename}
	if err := w.Reopen(); err != nil {
		return nil, err
	}
	return w, nil
}

// Write implements an io.Writer. It is thread-safe.
func (w *ReopenableWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		// Fallback: Try reopening if it was closed.
		if err := w.reopenInternal(); err != nil {
			return 0, err
		}
	}

	return w.file.Write(p)

}

// Close implements an io.Closer.
func (w *ReopenableWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.closeInternal()
}

// Reopen closes the current file and reopens it under the same path.
func (w *ReopenableWriter) Reopen() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 1. Close first (Flush)
	_ = w.closeInternal()

	// 2. New opening
	return w.reopenInternal()
}

// --- Interne Helper (Without a lock, as it's locked by the caller.) ---
func (w *ReopenableWriter) reopenInternal() error {
	f, err := os.OpenFile(w.filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	w.file = f
	return nil
}

func (w *ReopenableWriter) closeInternal() error {
	if w.file != nil {
		err := w.file.Close()
		w.file = nil
		return err
	}
	return nil
}
