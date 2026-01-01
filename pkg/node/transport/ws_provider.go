// Copyright 2025 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

package transport

import (
	"context"

	"net/http"

	"github.com/coder/websocket"
)

// WSProvider kapselt die Logik für den Verbindungsaufbau
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

// Serve wartet auf eine Verbindung (Server-Seite)
// Wir nutzen einen Channel, um die Connection nach dem Upgrade zurückzugeben
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
		// Neue Verbindung in den Posteingang der main schicken
		found <- &WSConnection{Conn: c}
	})

	p.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Eine Goroutine, die auf den Abbruch des Contexts wartet
	go func() {
		<-ctx.Done()
		p.Log.Info("HTTP-Server fährt herunter...")
		p.server.Shutdown(context.Background())
	}()

	// ListenAndServe blockiert hier, bis Shutdown() aufgerufen wird
	return p.server.ListenAndServe()

}

// Dial verbindet sich mit einem Server (Client-Seite)
func (p *WSProvider) Dial(ctx context.Context, url string) (Connection, error) {
	p.Log.With("url", url).Info("Dial...")
	c, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		return nil, err
	}
	return &WSConnection{Conn: c}, nil
}
