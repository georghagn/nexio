// logger_test
package gsflog

import (
	"testing"
)

func TestLogger(t *testing.T) {
	t.Logf("Test: TestLogger")

	// a Logger with Consoleoutput and Fileoutput to app.log
	// Fileoutput with
	//     - Json-Formatter,
	//     - SizePolicy = MaxBytes 1GB,
	//     - ArchiveStrategy = NoCompression
	//     - RetentiationPolicy = KeepAll
	log := NewDefault(nil)
	log.Info("server started")

	// Szenario: Ein Request kommt rein
	requestID := "req-abc-123"

	log.With("req_id", requestID).With("bytes", 512).Info("bytes received")

	wLog := log.With("req_id", requestID).With("user", "alice")
	wLog.Info("payload received")
	wLog.Info("eveything allright now")

	t.Logf("Loggers: %v", log.List())
}

func TestLoggerDispatches(t *testing.T) {
	t.Logf("Test: TestLoggerDispatches")

	a := NewTestSink()
	b := NewTestSink()

	m := New()
	m.AddNamed("a", a)
	m.AddNamed("b", b)

	m.Info("hello")

	if len(a.Calls) != 1 || a.Calls[0] != "hello" {
		t.Fatalf("sink a did not receive message")
	}
	if len(b.Calls) != 1 || b.Calls[0] != "hello" {
		t.Fatalf("sink b did not receive message")
	}
}
