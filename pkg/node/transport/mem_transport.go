// Copyright 2025 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

package transport

import (
	"context"
	"errors"
)

// MemConnection implementiert transport.Connection für In-Memory Tests
type MemConnection struct {
	In  chan []byte
	Out chan []byte
}

func (m *MemConnection) Send(ctx context.Context, data []byte) error {
	select {
	case m.Out <- data:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (m *MemConnection) Receive(ctx context.Context) ([]byte, error) {
	select {
	case data, ok := <-m.In:
		if !ok {
			return nil, errors.New("connection closed")
		}
		return data, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (m *MemConnection) Close(reason string) error {
	// In einem echten Test würde man hier vorsichtig sein,
	// um "close of closed channel" zu vermeiden.
	return nil
}

// NewMemPair erzeugt zwei fertig verbundene Endpunkte
func NewMemPair() (*MemConnection, *MemConnection) {
	aToB := make(chan []byte, 10)
	bToA := make(chan []byte, 10)

	return &MemConnection{In: bToA, Out: aToB},
		&MemConnection{In: aToB, Out: bToA}
}
