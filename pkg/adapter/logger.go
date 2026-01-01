// Copyright 2025 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"github.com/georghagn/gsf-suite/pkg/gsflog"
	"github.com/georghagn/gsf-suite/pkg/node/transport"
)

// Adapter wickelt einen gsflog.Logger so ein, dass er transport.LogSink erfüllt.
type Adapter struct {
	inner *gsflog.Logger
}

// Wrap macht aus einem gsflog.Logger einen transport.LogSink.
func Wrap(g gsflog.LogSink) transport.LogSink {
	// Hier machen wir die Assertion intern einmal zentral
	return &Adapter{inner: g.(*gsflog.Logger)}
}

func (a *Adapter) Debug(msg string) { a.inner.Debug(msg) }
func (a *Adapter) Info(msg string)  { a.inner.Info(msg) }
func (a *Adapter) Warn(msg string)  { a.inner.Warn(msg) }
func (a *Adapter) Error(msg string) { a.inner.Error(msg) }

// With sorgt dafür, dass der Kontext erhalten bleibt und das Ergebnis
// wieder das korrekte gsf-suite Interface erfüllt.
func (a *Adapter) With(key string, value any) transport.LogSink {
	return &Adapter{
		inner: a.inner.With(key, value).(*gsflog.Logger),
	}
}

// SetLevel erlaubt das Ändern des Log-Levels über das Interface.
func (a *Adapter) SetLevel(l int) {
	a.inner.SetLevel(gsflog.Level(l))
}
