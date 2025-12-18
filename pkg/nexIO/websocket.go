// Copyright 2025 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

package nexIO

import (
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Default Upgrader: Erlaubt alle Origins (für einfachere Entwicklung)
// In Production könnte man CheckOrigin anpassen.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Erlaubt Zugriff von überall (CORS-ähnlich)
	},
}

// ServeWS upgradet eine HTTP-Verbindung zu einem WebSocket
// und verarbeitet RPC-Requests in einer Loop.
func (s *Server) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if s.logger != nil {
			s.logger.Errorf("NexIO: WebSocket Upgrade failed: %v", err)
		}
		return
	}
	defer conn.Close()

	if s.logger != nil {
		s.logger.Debugf("NexIO: New WebSocket connection from %s", r.RemoteAddr)
	}

	// Mutex für Thread-Safe Writes auf diesen Socket
	// (Gorilla Websocket unterstützt keine konkurrierenden Writes)
	var writeMu sync.Mutex

	// Helper zum Schreiben von Antworten
	writeResponse := func(resp RPCResponse) {
		writeMu.Lock()
		defer writeMu.Unlock()

		// Setze Write Deadline um hängende Verbindungen zu vermeiden
		conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		if err := conn.WriteJSON(resp); err != nil {
			if s.logger != nil {
				s.logger.Errorf("NexIO: WebSocket Write Error: %v", err)
			}
			// Bei Schreibfehlern ist die Connection meist kaputt -> Loop wird beim nächsten Read abbrechen
		}
	}

	// Loop: Nachrichten lesen
	for {
		// Optional: Read Deadline (Keepalive Logik könnte hier erweitert werden)
		// conn.SetReadDeadline(time.Now().Add(60 * time.Second))

		var req RPCRequest
		// Blockiert, bis eine Nachricht kommt
		err := conn.ReadJSON(&req)
		if err != nil {
			// Normale Disconnects (CloseGoingAway etc.) nicht als Error loggen
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				if s.logger != nil {
					s.logger.Errorf("NexIO: WebSocket Read Error: %v", err)
				}
			} else {
				if s.logger != nil {
					s.logger.Debugf("NexIO: WebSocket disconnected: %v", err)
				}
			}
			break
		}

		// Asynchrone Verarbeitung!
		// Damit lange Tasks den Socket nicht für andere Incoming-Messages blockieren.
		go func(r RPCRequest) {
			// Deine existierende Logik nutzen:
			resp := s.ProcessRequest(r)

			// Antwort zurücksenden (Thread-Safe)
			writeResponse(resp)
		}(req)
	}
}
