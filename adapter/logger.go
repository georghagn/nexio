// Copyright 2026 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"github.com/georghagn/nexio/nexlog"
	"github.com/georghagn/nexio/node/transport"
)

// The adapter wraps a nexlog.Logger in such a way that it fulfills the requirements of transport.LogSink..
type Adapter struct {
	inner *nexlog.Logger
}

// Wrap turns a nexlog.Logger into a transport.LogSink.
func Wrap(g nexlog.LogSink) transport.LogSink {
	// Here we will handle the assertion internally, centrally.
	return &Adapter{inner: g.(*nexlog.Logger)}
}

func (a *Adapter) Debug(msg string) { a.inner.Debug(msg) }
func (a *Adapter) Info(msg string)  { a.inner.Info(msg) }
func (a *Adapter) Warn(msg string)  { a.inner.Warn(msg) }
func (a *Adapter) Error(msg string) { a.inner.Error(msg) }

// With ensures that the context is preserved and the result
// again meets the correct nexio interface.
func (a *Adapter) With(key string, value any) transport.LogSink {
	return &Adapter{
		inner: a.inner.With(key, value).(*nexlog.Logger),
	}
}

// SetLevel allows you to change the log level via the interface.
func (a *Adapter) SetLevel(l int) {
	a.inner.SetLevel(nexlog.Level(l))
}
