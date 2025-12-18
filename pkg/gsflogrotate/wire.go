package gsflogrotate

import (
	"github.com/georghagn/gsf-go/pkg/gsflog"
	"github.com/georghagn/gsf-go/pkg/rotate"
)

func Wire(log gsflog.LogSink, w *rotate.Writer) {
	w.OnEvent = func(e rotate.Event) {
		switch e.Type {

		case rotate.EventRotate:
			log.With("file", e.Filename).
				Info("log file rotated")

		case rotate.EventError:
			log.With("file", e.Filename).
				With("error", e.Err).
				Error("log rotation failed")
		}
	}
}
