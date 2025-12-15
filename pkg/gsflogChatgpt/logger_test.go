// logger_test
package gsflog

import (
	"testing"
)

func Test_MultiLogger(t *testing.T) {

	console := NewDefaultConsoleLogger()
	file := NewDefaultFileLogger()
	log := NewMultiLogger(console, file)
	Info(log, "server started", nil)
	Debug(log, "payload received", Fields{
		"bytes": 512,
		"user":  "alice",
	})
}
