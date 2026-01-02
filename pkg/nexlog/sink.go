// Copyright 2025 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

package nexlog

import (
	"io"
	"sync"
	"time"
)

// Level Definitions
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

// Fields is an alias for map, to make the code more readable.
type Fields map[string]any

// Entry holds all data for a log event.
type Entry struct {
	Level  Level
	Msg    string
	Time   time.Time
	Fields Fields
}

// Logger Structure
type Sink struct {
	threshold Level
	output    io.Writer
	formatter Formatter
	fields    Fields // Kontext-Data (immutable concept)
	mu        sync.Mutex
}

// New Sink = Logger
func NewSink(out io.Writer, threshold Level, formatter Formatter) LogSink {
	if formatter == nil {
		formatter = &TextFormatter{UseColors: false}
	}
	return &Sink{
		output:    out,
		threshold: threshold,
		formatter: formatter,
		fields:    make(Fields),
	}
}

func (l *Sink) SetLevel(level Level) {
	l.threshold = level
}

// With adds context and returns a NEW logger (Fluent Interface).
// The old logger remains untouched.
func (l *Sink) With(key string, value interface{}) LogSink {

	// 1. Copy existing fields
	newFields := make(Fields)
	l.mu.Lock()
	for k, v := range l.fields {
		newFields[k] = v
	}
	l.mu.Unlock()

	// 2. Add new fields
	newFields[key] = value

	// 3. Create a clone of the logger
	return &Sink{
		threshold: l.threshold,
		output:    l.output,
		formatter: l.formatter,
		fields:    newFields, // Neue Map
	}
}

// log is an intern method
func (l *Sink) log(level Level, msg string) {
	if level < l.threshold {
		return
	}

	entry := &Entry{
		Level:  level,
		Msg:    msg,
		Time:   time.Now(),
		Fields: l.fields,
	}

	// Formate
	bytes, err := l.formatter.Format(entry)
	if err != nil {
		// Fallback in case formatting fails (should not happen)
		bytes = []byte("LOG FORMAT ERROR: " + err.Error() + "\n")
	}

	// Write (Thread-Safe)
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output.Write(bytes)
}

// public api
func (l *Sink) Debug(msg string) {
	l.log(LevelDebug, msg)
}

func (l *Sink) Info(msg string) {
	l.log(LevelInfo, msg)
}

func (l *Sink) Warn(msg string) {
	l.log(LevelWarn, msg)
}

func (l *Sink) Error(msg string) {
	l.log(LevelError, msg)
}
