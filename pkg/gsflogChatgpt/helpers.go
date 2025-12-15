// Copyright 2025 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

package gsflog

import "time"

func Debug(l Logger, msg string, f Fields) {
	log(l, LevelDebug, msg, f)
}

func Info(l Logger, msg string, f Fields) {
	log(l, LevelInfo, msg, f)
}

func Warn(l Logger, msg string, f Fields) {
	log(l, LevelWarn, msg, f)
}

func Error(l Logger, msg string, f Fields) {
	log(l, LevelError, msg, f)
}

func log(l Logger, level Level, msg string, f Fields) {
	l.Log(Entry{
		Time:   time.Now(),
		Level:  level,
		Msg:    msg,
		Fields: f,
	})
}
