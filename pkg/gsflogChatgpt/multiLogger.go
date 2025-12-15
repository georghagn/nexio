// Copyright 2025 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

package gsflog

// --- MultiLogger ---
type MultiLogger struct {
	loggers []Logger
}

func NewMultiLogger(loggers ...Logger) *MultiLogger {
	return &MultiLogger{loggers: loggers}
}

func (m *MultiLogger) Level() Level {
	return LevelDebug
}

func (m *MultiLogger) SetLevel(level Level) {
	for _, l := range m.loggers {
		l.SetLevel(level)
	}
}

func (m *MultiLogger) Log(e Entry) {
	for _, l := range m.loggers {
		l.Log(e)
	}
}
