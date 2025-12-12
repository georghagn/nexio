// Copyright 2025 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

package nexio

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// Upgrader konfiguriert den HTTP->WS Handshake
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	// Wir cke erlauben hier erstmals alle Origins (CORS). In Produktion: NO
	CheckOrigin: func(r *http.Request) bool { return true },
}

// ServeWS behandelt WebSocket Verbindungen.
func (s *Server) ServeWS(w http.ResponseWriter, r *http.Request) {

	// 1. Upgrade: Aus HTTP wird WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if s.logger != nil {
			s.logger.Errorf("Upgrade failed: %v", err)
		}
		return
	}
	defer conn.Close()

	// 2. Loop: Solange Verbindung offen ist
	for {
		var req RPCRequest

		// Nachricht lesen (JSON)
		err := conn.ReadJSON(&req)
		if err != nil {
			if s.logger != nil {
				s.logger.Errorf("Fehler beim Lesen, (Client weg?) : %v", err)
			}
			// Fehler beim Lesen bedeutet meist: Client ist weg
			break
		}

		// 3. RPC ausführen (Nutzt unsere isolierte Logik!)
		resp := s.ProcessRequest(req)

		// 4. Antwort schreiben
		if err := conn.WriteJSON(resp); err != nil {
			if s.logger != nil {
				s.logger.Errorf("Write failed: %v", err)
			}
			break
		}
	}
}
