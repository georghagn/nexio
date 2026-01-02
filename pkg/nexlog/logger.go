// Copyright 2025 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

package nexlog

import (
	"sync"
)

// --- MultiLogger ---
type Logger struct {
	loggers map[string]LogSink
	mu      sync.RWMutex
}

func New() *Logger {
	lgrs := make(map[string]LogSink)
	return &Logger{loggers: lgrs}
}

func NewDefault(logFileName *string) *Logger {
	consoleSink := NewDefaultConsoleSink()
	fileSink := NewDefaultFileSink(logFileName)

	mLogger := New()
	mLogger.AddNamed("Console", consoleSink)
	mLogger.AddNamed("File", fileSink)

	return mLogger
}

func NewDefaultConsole() *Logger {
	consoleSink := NewDefaultConsoleSink()

	mLogger := New()
	mLogger.AddNamed("Console", consoleSink)

	return mLogger
}

func (m *Logger) AddNamed(name string, logger LogSink) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.loggers[name] = logger
}

func (m *Logger) RemoveNamed(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.loggers, name)
}

func (m *Logger) List() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	keys := make([]string, 0, len(m.loggers))
	for k, _ := range m.loggers {
		keys = append(keys, k)

	}
	return keys
}

func (m *Logger) Level() Level {
	return LevelDebug
}

func (m *Logger) SetLevel(level Level) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, l := range m.loggers {
		l.SetLevel(level)
	}
}

func (m *Logger) Debug(msg string) {
	m.mu.RLock()
	loggers := make([]LogSink, 0, len(m.loggers))
	for _, l := range m.loggers {
		loggers = append(loggers, l)
	}
	m.mu.RUnlock()
	for _, l := range m.loggers {
		l.Debug(msg)
	}
}

func (m *Logger) Info(msg string) {
	m.mu.RLock()
	loggers := make([]LogSink, 0, len(m.loggers))
	for _, l := range m.loggers {
		loggers = append(loggers, l)
	}
	m.mu.RUnlock()
	for _, l := range m.loggers {
		l.Info(msg)
	}
}

func (m *Logger) Warn(msg string) {
	m.mu.RLock()
	loggers := make([]LogSink, 0, len(m.loggers))
	for _, l := range m.loggers {
		loggers = append(loggers, l)
	}
	m.mu.RUnlock()
	for _, l := range m.loggers {
		l.Warn(msg)
	}
}

func (m *Logger) Error(msg string) {
	m.mu.RLock()
	loggers := make([]LogSink, 0, len(m.loggers))
	for _, l := range m.loggers {
		loggers = append(loggers, l)
	}
	m.mu.RUnlock()
	for _, l := range m.loggers {
		l.Error(msg)
	}
}

func (m *Logger) With(key string, value interface{}) LogSink {
	m.mu.Lock()
	defer m.mu.Unlock()
	lgrs := make(map[string]LogSink)
	for k, l := range m.loggers {
		lgrs[k] = l.With(key, value)
	}
	return &Logger{loggers: lgrs}
}
