// Copyright 2025 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

package gsflog

import (
	"sync"
)

// --- MultiLogger ---
type MultiLogger struct {
	loggers map[string]*Logger
	mu      sync.Mutex
}

func NewMultiLogger() *MultiLogger {
	lgrs := make(map[string]*Logger)
	return &MultiLogger{loggers: lgrs}
}

func NewDefaultMultiLogger(logName *string) *MultiLogger {
	consoleL := NewDefaultConsoleLogger()
	fileL := NewDefaultFileLogger(logName)

	mLogger := NewMultiLogger()
	mLogger.AddNamed("Console", consoleL)
	mLogger.AddNamed("File", fileL)
	return mLogger
}

func (m *MultiLogger) AddNamed(name string, logger *Logger) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.loggers[name] = logger
}

func (m *MultiLogger) RemoveNamed(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.loggers, name)
}

func (m *MultiLogger) List() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	keys := make([]string, 0, len(m.loggers))
	for k, _ := range m.loggers {
		keys = append(keys, k)

	}
	return keys
}

func (m *MultiLogger) Level() Level {
	return LevelDebug
}

func (m *MultiLogger) SetLevel(level Level) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, l := range m.loggers {
		l.SetLevel(level)
	}
}

func (m *MultiLogger) Debug(msg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, l := range m.loggers {
		l.log(LevelDebug, msg)
	}
}

func (m *MultiLogger) Info(msg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, l := range m.loggers {
		l.log(LevelInfo, msg)
	}
}

func (m *MultiLogger) Warn(msg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, l := range m.loggers {
		l.log(LevelWarn, msg)
	}
}

func (m *MultiLogger) Error(msg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, l := range m.loggers {
		l.log(LevelError, msg)
	}
}

func (m *MultiLogger) With(key string, value interface{}) *MultiLogger {
	m.mu.Lock()
	defer m.mu.Unlock()
	lgrs := make(map[string]*Logger)
	for k, l := range m.loggers {
		lgrs[k] = l.With(key, value)
	}
	return &MultiLogger{loggers: lgrs}
}
