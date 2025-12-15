// Copyright 2025 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

package gsflog

import (
	"os"
)

// NewConsoleLogger (Helper mit Farben)
func NewConsoleLogger(level Level, formatter Formatter) *Logger {
	return New(os.Stdout, level, formatter)
}

func NewDefaultConsoleLogger() *Logger {
	txtFormatter := &TextFormatter{
		UseColors: true,
	}
	return NewConsoleLogger(LevelInfo, txtFormatter)
}
