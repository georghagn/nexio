// Copyright 2025 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

package gsflog

import (
	"io"

	"github.com/georghagn/gsf-suite/pkg/rotate"
)

func NewFileSink(out io.Writer, level Level, formatter Formatter) LogSink {
	return NewSink(out, level, formatter)
}

func NewDefaultFileSink(fileName *string) LogSink {
	fName := "app.log"
	if fileName != nil {
		fName = *fileName
	}
	jsonFormatter := &JSONFormatter{}
	rotator := rotate.New(fName, nil, nil, nil)
	return NewFileSink(rotator, LevelInfo, jsonFormatter)
}
