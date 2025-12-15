// Copyright 2025 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

package gsflog

import (
	"os"
)

// --- ConsoleLogger ---
type ConsoleLogger struct {
	BaseLogger
}

func NewConsoleLogger(level Level, formatter Formatter) *ConsoleLogger {

	return &ConsoleLogger{
		BaseLogger: BaseLogger{
			level:     level,
			formatter: formatter,
			out:       os.Stdout,
		},
	}
}

func NewDefaultConsoleLogger() *ConsoleLogger {
	txtFormatter := &TextFormatter{
		UseColors: true,
	}
	return NewConsoleLogger(LevelInfo, txtFormatter)
}

func (c *ConsoleLogger) Level() Level {
	return LevelDebug
}

func (c *ConsoleLogger) SetLevel(level Level) {
	c.BaseLogger.level = level
}

func (c *ConsoleLogger) Log(e Entry) {
	if !c.BaseLogger.enabled(e) {
		return
	}

	bytes, err := c.formatter.Format(&e)
	if err != nil {
		// Fallback, falls Formatierung fehlschlägt (sollte nicht passieren)
		bytes = []byte("LOG FORMAT ERROR: " + err.Error() + "\n")
	}

	// Schreiben (Thread-Safe)
	c.BaseLogger.mu.Lock()
	defer c.BaseLogger.mu.Unlock()
	c.BaseLogger.out.Write(bytes)
}
