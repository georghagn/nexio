// Copyright 2025 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

package gsflog

import (
	"os"
)

// NewConsoleLogger (Helper with colours)
func NewConsoleSink(level Level, formatter Formatter) LogSink {
	return NewSink(os.Stdout, level, formatter)
}

func NewDefaultConsoleSink() LogSink {
	txtFormatter := &TextFormatter{
		UseColors: true,
	}
	return NewConsoleSink(LevelInfo, txtFormatter)
}
