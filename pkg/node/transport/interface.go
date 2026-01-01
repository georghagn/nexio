// Copyright 2025 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

package transport

import "context"

type Connection interface {
	Send(ctx context.Context, data []byte) error
	Receive(ctx context.Context) ([]byte, error)
	Close(reason string) error
}

type WSService interface {
	Listen(addr string, found chan<- Connection) error
	Dial(ctx context.Context, url string) (Connection, error)
}

type LogSink interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)

	With(key string, value any) LogSink
}

// default in Konstruktor: Silentlogger No-Op-Logger
// => no panic if logger is not explicitly set
type SilentLogger struct{}

func (s *SilentLogger) Debug(msg string) {}
func (s *SilentLogger) Info(msg string)  {}
func (s *SilentLogger) Warn(msg string)  {}
func (s *SilentLogger) Error(msg string) {}
func (s *SilentLogger) With(key string, value any) LogSink {
	return s
}
