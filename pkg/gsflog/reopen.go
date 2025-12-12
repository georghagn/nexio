// Copyright 2025 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

package gsflog

import (
	"os"
	"sync"
)

// ReopenableWriter ist ein io.WriteCloser, der zur Laufzeit neu geöffnet werden kann.
// Das ist nützlich für logrotate-Strategien (z.B. via Scheduler oder SIGHUP),
// bei denen die Datei extern verschoben wird und der Prozess das Filehandle erneuern muss.
type ReopenableWriter struct {
	filename string
	file     *os.File
	mu       sync.Mutex
}

// NewReopenableWriter öffnet die Datei und bereitet das Schreiben vor.
func NewReopenableWriter(filename string) (*ReopenableWriter, error) {
	w := &ReopenableWriter{filename: filename}
	if err := w.Reopen(); err != nil {
		return nil, err
	}
	return w, nil
}

// Write implementiert io.Writer. Es ist thread-safe.
func (w *ReopenableWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		// Fallback: Versuchen neu zu öffnen, falls geschlossen war
		if err := w.reopenInternal(); err != nil {
			return 0, err
		}
	}
	return w.file.Write(p)
}

// Close implementiert io.Closer.
func (w *ReopenableWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.closeInternal()
}

// Reopen schließt die aktuelle Datei und öffnet sie unter dem gleichen Pfad neu.
func (w *ReopenableWriter) Reopen() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 1. Erst schließen (Flush)
	_ = w.closeInternal()

	// 2. Neu öffnen
	return w.reopenInternal()
}

// --- Interne Helper (ohne Lock, da vom Caller gelockt) ---

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
