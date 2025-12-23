package main

import (
	"os"
	"time"

	"github.com/georghagn/gsf-go/pkg/gsflog"
	"github.com/georghagn/gsf-go/pkg/rotate"
	"github.com/georghagn/gsf-go/pkg/schedule"
)

// Canonical example for GSF (Go Simple Frameworks)
//
// This example demonstrates explicit composition of:
//   - gsflog   : logging
//   - rotate   : file rotation
//   - schedule : time-based triggering
//
// There is no hidden wiring, no global state, and no implicit background magic.
func main() {
	// --- Logger ------------------------------------------------------------

	// Create a basic logger writing to an io.Writer
	// (the writer will be provided by the rotator below)

	// --- Rotator -----------------------------------------------------------

	rotator := rotate.New(
		"./logs/app.log",
		&rotate.SizePolicy{MaxBytes: 1024 * 1024}, // 1 MB
		nil, // default archive strategy
		nil, // default retention policy
	)

	// Wire rotation events to the logger (no cyclic dependency)
	logger := gsflog.New(rotator)
	gsflogrotate.Wire(logger, rotator)

	// --- Scheduler ---------------------------------------------------------

	scheduler := schedule.New()
	scheduler.SetLogger(logger)

	// Periodic job: force rotation every minute (time-based rotation)
	scheduler.Every(1*time.Minute, func() {
		_ = rotator.Close()
	})

	// --- Application Logic -------------------------------------------------

	logger.Info("application started")

	for i := 0; i < 10; i++ {
		logger.With("counter", i).Info("working")
		time.Sleep(2 * time.Second)
	}

	logger.Info("application shutting down")

	// --- Graceful Shutdown -------------------------------------------------

	scheduler.StopAll()
	_ = rotator.Close()
	_ = os.Stdout.Sync()
}
