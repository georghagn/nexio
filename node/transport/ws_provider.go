// Copyright 2026 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

package transport

import (
	"context"

	"net/http"

	"github.com/coder/websocket"
)

// WSProvider encapsulates the logic for establishing the connection.
type WSProvider struct {
	server *http.Server
	Log    LogSink
}

func NewWSProvider(logger LogSink) *WSProvider {
	p := &WSProvider{}
	if logger == nil {
		p.Log = &SilentLogger{}
	} else {
		p.Log = logger
	}
	return p
}

// Server is waiting for a connection (server-side)
// We use a channel to reconnect after the upgrade
func (p *WSProvider) Listen(ctx context.Context, addr string, found chan<- Connection) error {
	p.Log.With("addr", addr).Info("WebSocket Server startet...")

	mux := http.NewServeMux()

	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: true,
		})
		if err != nil {
			return
		}
		// Send new connection to the main inbox
		found <- &WSConnection{Conn: c}
	})

	p.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// A goroutine that waits for the context to be terminated.
	go func() {
		<-ctx.Done()
		p.Log.Info("HTTP-Server fährt herunter...")
		p.server.Shutdown(context.Background())
	}()

	//ListenAndServe blocks here until Shutdown() is called.
	return p.server.ListenAndServe()

}

// Dial connects to a server (client side)
func (p *WSProvider) Dial(ctx context.Context, url string) (Connection, error) {
	p.Log.With("url", url).Info("Dial...")
	c, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		return nil, err
	}
	return &WSConnection{Conn: c}, nil
}
