// Copyright 2025 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

package gsflog

import (
	"io"
	"sync"
	"time"
)

// Level Definitionen
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO "
	case LevelWarn:
		return "WARN "
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Fields ist ein Alias für Map, damit der Code lesbarer wird.
type Fields map[string]any

// Entry hält alle Daten eines Log-Ereignisses.
type Entry struct {
	Level  Level
	Msg    string
	Time   time.Time
	Fields Fields
}

// Logger Struktur
type Logger struct {
	threshold Level
	output    io.Writer
	formatter Formatter
	fields    Fields // Kontext-Daten (immutable concept)
	mu        sync.Mutex
}

// New Logger
func New(out io.Writer, threshold Level, formatter Formatter) *Logger {
	if formatter == nil {
		formatter = &TextFormatter{UseColors: false}
	}
	return &Logger{
		output:    out,
		threshold: threshold,
		formatter: formatter,
		fields:    make(Fields),
	}
}

func (l *Logger) SetLevel(level Level) {
	l.threshold = level
}

// With fügt Kontext hinzu und gibt einen NEUEN Logger zurück (Fluent Interface).
// Der alte Logger bleibt unberührt.
func (l *Logger) With(key string, value interface{}) *Logger {

	// 1. Kopiere bestehende Felder
	newFields := make(Fields)
	l.mu.Lock()
	for k, v := range l.fields {
		newFields[k] = v
	}
	l.mu.Unlock()

	// 2. Füge neues Feld hinzu
	newFields[key] = value

	// 3. Erstelle Klon des Loggers
	return &Logger{
		threshold: l.threshold,
		output:    l.output,
		formatter: l.formatter,
		fields:    newFields, // Neue Map
	}
}

// log ist die interne Methode
func (l *Logger) log(level Level, msg string) {
	if level < l.threshold {
		return
	}

	entry := &Entry{
		Level:  level,
		Msg:    msg,
		Time:   time.Now(),
		Fields: l.fields,
	}

	// Formatieren
	bytes, err := l.formatter.Format(entry)
	if err != nil {
		// Fallback, falls Formatierung fehlschlägt (sollte nicht passieren)
		bytes = []byte("LOG FORMAT ERROR: " + err.Error() + "\n")
	}

	// Schreiben (Thread-Safe)
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output.Write(bytes)
}

/*
// Infof etc. unterstützen wir auch noch, aber "With()" ist moderner
func (l *Logger) Infof(format string, args ...interface{}) {
	// Hinweis: Wir nutzen hier fmt.Sprintf intern, bevor wir es dem Logger geben
	// Alternativ könnte man Message als interface{} definieren.
	// Für Tiny halten wir es simpel:
	l.log(LevelInfo, fmt.Sprintf(format, args...))
}

// Helper für formatierte Strings (wie printf)
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(LevelWarn, fmt.Sprintf(format, args...))
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(LevelError, fmt.Sprintf(format, args...))
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(LevelDebug, fmt.Sprintf(format, args...))
}
*/
