package main

import (
	"fmt"
	"time"

	"github.com/georghagn/gsf-go/pkg/schedule"
)

// simpleLogger is a minimal logger implementation used only for the example.
// It demonstrates how the scheduler can optionally integrate with logging
// without depending on a concrete logging framework.
type simpleLogger struct{}

func (l *simpleLogger) Info(format string, args ...interface{}) {
	fmt.Printf("[INFO] "+format+"\n", args...)
}

func (l *simpleLogger) Error(format string, args ...interface{}) {
	fmt.Printf("[ERROR] "+format+"\n", args...)
}

func main() {
	fmt.Println("Starting scheduler example...")

	s := schedule.New()

	// Optional: inject a logger
	s.SetLogger(&simpleLogger{})

	// Periodic job: runs every 2 seconds
	s.Every(2*time.Second, func() {
		fmt.Println("Periodic job executed at", time.Now().Format(time.RFC3339))
	})

	// One-shot job: runs once after 5 seconds
	s.At(time.Now().Add(5*time.Second), func() {
		fmt.Println("One-shot job executed at", time.Now().Format(time.RFC3339))
	})

	// Let the scheduler run for a while
	time.Sleep(10 * time.Second)

	fmt.Println("Stopping scheduler...")
	s.StopAll()

	fmt.Println("Scheduler stopped cleanly.")
}
