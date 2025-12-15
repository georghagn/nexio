// Copyright 2025 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

package gsflog

import (
	"io"

	"github.com/georghagn/gsf-go/pkg/rotate"
)

// --- FileLogger JSON ---
type FileLogger struct {
	BaseLogger
}

func NewFileLogger(level Level, formatter Formatter, out io.Writer) *FileLogger {
	return &FileLogger{
		BaseLogger: BaseLogger{
			level:     level,
			formatter: formatter,
			out:       out,
		},
	}
}

func NewDefaultFileLogger() *FileLogger {
	formatter := &JSONFormatter{}
	rotator := rotate.New("app.log", nil, nil, nil)
	return &FileLogger{
		BaseLogger: BaseLogger{
			level:     LevelInfo,
			formatter: formatter,
			out:       rotator,
		},
	}
}

func (f *FileLogger) Level() Level {
	return LevelDebug
}

func (f *FileLogger) SetLevel(level Level) {
	f.BaseLogger.level = level
}

func (f *FileLogger) Log(e Entry) {
	if !f.BaseLogger.enabled(e) {
		return
	}

	bytes, err := f.formatter.Format(&e)
	if err != nil {
		// Fallback, falls Formatierung fehlschlägt (sollte nicht passieren)
		bytes = []byte("LOG FORMAT ERROR: " + err.Error() + "\n")
	}

	// Schreiben (Thread-Safe)
	f.BaseLogger.mu.Lock()
	defer f.BaseLogger.mu.Unlock()
	f.BaseLogger.out.Write(bytes)
}
