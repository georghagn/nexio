// Copyright 2025 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

package nexlog

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

// The formatter determines how a log entry is converted into bytes.
type Formatter interface {
	Format(entry *Entry) ([]byte, error)
}

// --- 1. JSON Formatter (for machines/files) ---
type JSONFormatter struct{}

func (f *JSONFormatter) Format(e *Entry) ([]byte, error) {

	// We are building a flat map for the JSON
	data := make(Fields)

	// Base Data
	data["time"] = e.Time.Format(time.RFC3339)
	data["level"] = e.Level.String()
	data["msg"] = e.Msg

	// Adding user-fields
	for k, v := range e.Fields {
		data[k] = v
	}

	// convert to JSON (with Newline at end)
	bytes, err := json.Marshal(data)
	return append(bytes, '\n'), err
}

// --- 2. Text Formatter (for humans/console) ---
type TextFormatter struct {
	UseColors bool
}

func (f *TextFormatter) Format(e *Entry) ([]byte, error) {

	timestamp := e.Time.Format("2006/01/02 15:04:05")

	// Level String
	lvl := e.Level.String()

	// Colours (only if desired)
	if f.UseColors {
		lvl = colorize(e.Level, lvl)
	}

	// Foramte fields (key=value)
	var fieldStr string
	if len(e.Fields) > 0 {

		// Sort for consistent output
		keys := make([]string, 0, len(e.Fields))
		for k := range e.Fields {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		var sb strings.Builder
		for _, k := range keys {
			sb.WriteString(fmt.Sprintf(" %s=%v", k, e.Fields[k]))
		}
		fieldStr = sb.String()
	}

	// Building: YYYY/MM/DD HH:MM:SS [LEVEL] Message key=value
	line := fmt.Sprintf("%s [%s] %s%s\n", timestamp, lvl, e.Msg, fieldStr)
	return []byte(line), nil
}

// --- Helper: ANSI Colours ---
func colorize(l Level, s string) string {
	const (
		Reset  = "\033[0m"
		Red    = "\033[31m"
		Yellow = "\033[33m"
		Blue   = "\033[34m"
		Gray   = "\033[37m"
	)

	var color string
	switch l {
	case LevelDebug:
		color = Gray
	case LevelInfo:
		color = Blue
	case LevelWarn:
		color = Yellow
	case LevelError:
		color = Red
	default:
		color = Reset
	}
	return color + s + Reset
}
