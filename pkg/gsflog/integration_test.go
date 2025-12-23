package gsflog_test

import (
	"io"
	"os"
	"testing"

	"github.com/georghagn/gsf-suite/pkg/gsflog"
	"github.com/georghagn/gsf-suite/pkg/rotate"
)

func TestLoggerWithRotatorIntegration(t *testing.T) {
	t.Logf("Test: TestLoggerWithRotatorIntegration")
	// --- Arrange --------------------------------------------------

	tmpFile, err := os.CreateTemp("", "gsflog-*.log")
	if err != nil {
		t.Fatalf("cannot create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Rotator
	writer := rotate.New(
		tmpFile.Name(),
		nil, // default rotation policy
		nil, // default archive strategy
		nil, // default retention
	)

	// Logger
	logger := gsflog.NewSink(
		writer,
		gsflog.LevelDebug,
		nil, // default formatter
	)

	// Integration (EXPLICIT!)

	// --- Act ------------------------------------------------------

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	// --- Assert ---------------------------------------------------

	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("cannot read log file: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("log file is empty")
	}
}

func TestMultiLoggerWithRotatorIntegration(t *testing.T) {
	t.Logf("Test: TestLoggerWithRotatorIntegration")

	tmpFile, err := os.CreateTemp("", "gsf-multi-*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// File logger
	rot := rotate.New(tmpFile.Name(), nil, nil, nil)
	fileLogger := gsflog.NewSink(rot, gsflog.LevelDebug, nil)

	// Console logger (discard output)
	consoleLogger := gsflog.NewSink(io.Discard, gsflog.LevelDebug, nil)

	// MultiLogger
	m := gsflog.New()
	m.AddNamed("file", fileLogger)
	m.AddNamed("console", consoleLogger)

	// Act
	m.Info("hello multilogger")

	// Assert
	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	if len(data) == 0 {
		t.Fatal("file logger did not receive log")
	}
}
