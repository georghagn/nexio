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

// Logger Interface Signatur
type Logger interface {
	Level() Level
	SetLevel(Level)
	Log(Entry)
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

// --- BaseLogger ---
type BaseLogger struct {
	level     Level
	formatter Formatter
	out       io.Writer
	mu        sync.Mutex
}

func (b *BaseLogger) Level() Level {
	return b.level
}

func (b *BaseLogger) SetLevel(l Level) {
	b.level = l
}

func (b *BaseLogger) enabled(e Entry) bool {
	return e.Level >= b.level
}
