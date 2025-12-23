// Copyright 2025 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

package lockWriter

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/georghagn/gsf-suite/pkg/rotate/lock"
)

type LockWriter struct {
	filename string
	fileLock *lock.FileLock
	file     *os.File
	mu       sync.Mutex
}

// NewLockWriter opens the file and prepares the writing.
func NewLockingWriter(filename string) (*LockWriter, error) {
	fl := &lock.FileLock{
		Path:    fmt.Sprintf("%s.LOCK", filename),
		Timeout: 1 * time.Second,
		Expiry:  2 * time.Minute, // if a prozess crash: release after 2 min
	}

	w := &LockWriter{
		filename: filename,
		fileLock: fl,
	}
	return w, nil
}

// Write implements an io.Writer. It is thread-safe.
func (w *LockWriter) Write(p []byte) (n int, err error) {

	// 1. Extern prozess-Lock (e.g. for Rotator)
	if err = w.fileLock.Lock(); err != nil {
		return 0, err
	}
	defer w.fileLock.Unlock()

	// 2. Intern Thread-Lock (for Goroutine)
	w.mu.Lock()
	defer w.mu.Unlock()

	// 3. Open file for writing
	f, err := os.OpenFile(w.filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return 0, err
	}
	// Ensure that the file is ALWAYS closed at the end of the method.
	defer f.Close()

	return f.Write(p)

}

// Close is only included here for formality (io.Closer),
// since no file remains permanently open in Atomic mode.
func (w *LockWriter) Close() error {
	return nil
}
