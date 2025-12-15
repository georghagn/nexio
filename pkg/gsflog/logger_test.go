// logger_test
package gsflog

import (
	"testing"
)

func Test_MultiLogger(t *testing.T) {

	// a Logger with Consoleoutput and Fileoutput to app.log
	// Fileoutput with
	//     - Json-Formatter,
	//     - SizePolicy = MaxBytes 1GB,
	//     - ArchiveStrategy = NoCompression
	//     - RetentiationPolicy = KeepAll
	log := NewDefaultMultiLogger(nil)
	log.Info("server started")

	// Szenario: Ein Request kommt rein
	requestID := "req-abc-123"

	log.With("req_id", requestID).With("bytes", 512).Info("bytes received")

	wLog := log.With("req_id", requestID).With("user", "alice")
	wLog.Info("payload received")
	wLog.Info("eveything allright now")

	t.Logf("Loggers: %v", log.List())
}
