// Copyright 2025 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

package rotate

import (
	"os"
	"path/filepath"
	"sync"
	"time"
)

// DebugLogger Interface für interne Diagnose
type DebugLogger interface {
	Debug(msg string)
}

type Writer struct {
	filename      string
	file          *os.File
	currentSize   int64
	openTimestamp time.Time // Wann wurde das aktuelle File geöffnet?

	// Die Strategien
	Rotation  RotationPolicy
	Archive   ArchiveStrategy
	Retention RetentionPolicy

	mu sync.Mutex

	Logger DebugLogger
}

// New erzeugt einen Writer mit Standard-Strategien (falls nil übergeben wird).
func New(filename string, r RotationPolicy, a ArchiveStrategy, ret RetentionPolicy) *Writer {

	// Defaults setzen für "Tiny" Usage (Convention over Configuration)
	if r == nil {
		r = &SizePolicy{MaxBytes: 10 * 1024 * 1024}
	} // 10MB

	if a == nil {
		a = &NoCompression{}
	} // Nur Umbenennen

	if ret == nil {
		ret = &KeepAll{}
	} // Nichts löschen

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

	// 1. Datei öffnen falls nötig
	if w.file == nil {
		if err := w.open(); err != nil {
			return 0, err
		}
	}

	// 2. Prüfen ob Rotation nötig (Delegation an Policy)
	// Wir addieren len(p) hypothetisch dazu, um zu sehen, ob wir das Limit sprengen würden
	timeOpen := time.Since(w.openTimestamp)
	if w.Rotation.ShouldRotate(w.currentSize+int64(len(p)), timeOpen) {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}

	// 3. Schreiben
	n, err = w.file.Write(p)
	w.currentSize += int64(n)
	return n, err
}

func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.closeFile()
}

func (w *Writer) SetLogger(l DebugLogger) {
	w.Logger = l
}

// --- Interne Logik ---
func (w *Writer) open() error {

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

func (w *Writer) closeFile() error {
	if w.file == nil {
		return nil
	}
	err := w.file.Close()
	w.file = nil
	w.currentSize = 0
	return err
}

func (w *Writer) rotate() error {
	if w.Logger != nil {
		w.Logger.Debug("Rotating file: " + w.filename)
	}

	if err := w.closeFile(); err != nil {
		return err
	}

	// 1. Aktuelles File zu
	if err := w.closeFile(); err != nil {
		return err
	}

	// 2. Archivieren (z.B. Umbenennen oder Zippen)
	if _, err := w.Archive.Archive(w.filename); err != nil {
		return err // Wenn Archivieren fehlschlägt, versuchen wir es beim nächsten Write wieder
	}

	// 3. Aufräumen (Alte Backups löschen)
	// Wir ignorieren Fehler hier bewusst, Logs löschen ist "Best Effort"
	_ = w.Retention.Prune(w.filename)

	// 4. Neues File auf
	return w.open()
}
