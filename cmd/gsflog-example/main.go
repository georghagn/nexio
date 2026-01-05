// Copyright 2025 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/georghagn/nexio/nexlog"
	"github.com/georghagn/nexio/rotate"
)

func main() {
	// Setup Logger-Sinks
	//   - Console: coloured Textformatter,
	//   - File:  Json-Formatter, rotating)

	// 1. Setup Rotator
	logFile := "nexlog-example-main.log"
	rotator := rotate.New(logFile, nil, nil, nil)
	defer rotator.Close()

	// 2. Setup Console Logger (Bunt, Text)
	colouredTextFormatter := &nexlog.TextFormatter{UseColors: true}
	consoleLoggerSink := nexlog.NewConsoleSink(nexlog.LevelDebug, colouredTextFormatter)

	// 3. Setup Console FileLogger (Json)
	jsonFormatter := &nexlog.JSONFormatter{}
	fileLoggerSink := nexlog.NewFileSink(rotator, nexlog.LevelInfo, jsonFormatter)

	// 4. Build Logger
	// we don't use the Convenience Method: mainLogger := NewDefault(logFile)
	// instead we use the hardcore way :-) Different LogLevels! for console and file
	mainLogger := nexlog.New()
	mainLogger.AddNamed("Console", consoleLoggerSink)
	mainLogger.AddNamed("File", fileLoggerSink)

	// 5. Start demo
	fmt.Println("--- GSF Logger Demo ---")

	// Szenario: Ein Request kommt rein
	requestID := "req-abc-123"

	// output to console and file (InfoLevel)
	mainLogger.With("req_id", requestID).Info("Processing Request...")

	// output only to (DebugLevel)
	mainLogger.With("req_id", requestID).Debug("Should only be in console output...")

	// ... somewhere deep in sourcecode ...
	userID := 42

	// we extend Context
	ctxLogger := mainLogger.With("user_id", userID)
	ctxLogger.With("req_id", requestID).Info("Processing Request...")
	ctxLogger.Warn("Use has no credit")
	ctxLogger.Info("Stop Processing...")

	// Konsole: 2023/... [WARN] User hat kein Guthaben mehr! req_id=req-abc-123 user_id=42 (in Gelb)

	mainLogger.Error("Transaction failed")
	// File: {"level":"ERROR","msg":"Transaction failed","req_id":"req-abc-123","time":"...","user_id":42}

	fmt.Println("Check ", logFile, " for JSON output!")
}
