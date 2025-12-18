package main

import (
	"fmt"
	"time"

	"github.com/georghagn/gsf-go/pkg/rotate"
)

// rotate-example demonstrates the rotate module in isolation.
//
// It shows:
//   - size-based rotation
//   - archive + retention defaults
//   - event handling without any logging framework
func main() {
	writer := rotate.New(
		"./logs/rotate-example.log",
		&rotate.SizePolicy{MaxBytes: 256}, // very small to trigger rotation quickly
		nil, // default archive strategy (rename)
		nil, // default retention policy (keep all)
	)

	// Attach an event handler
	writer.OnEvent = func(e rotate.Event) {
		switch e.Type {
		case rotate.EventRotate:
			fmt.Printf("ROTATE: %s\n", e.Filename)
		case rotate.EventError:
			fmt.Printf("ERROR: %s (%v)\n", e.Filename, e.Err)
		}
	}

	// Write data until rotations occur
	for i := 0; i < 20; i++ {
		line := fmt.Sprintf("line %02d: some example data...\n", i)
		_, err := writer.Write([]byte(line))
		if err != nil {
			fmt.Printf("write failed: %v\n", err)
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	_ = writer.Close()
}
