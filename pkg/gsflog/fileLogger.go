// Copyright 2025 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

package gsflog

import (
	"io"

	"github.com/georghagn/gsf-go/pkg/rotate"
)

func NewFileLogger(out io.Writer, level Level, formatter Formatter) *Logger {
	return New(out, level, formatter)
}

func NewDefaultFileLogger(fileName *string) *Logger {
	fName := "app.log"
	if fileName != nil {
		fName = *fileName
	}
	jsonFormatter := &JSONFormatter{}
	rotator := rotate.New(fName, nil, nil, nil)
	return NewFileLogger(rotator, LevelInfo, jsonFormatter)
}
