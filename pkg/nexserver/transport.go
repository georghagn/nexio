package nexIOserver

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// Diese Methode gehört jetzt direkt zum Hub!
func (h *Hub) ServeWebSocket(w http.ResponseWriter, r *http.Request) {
	// 1. Upgrader direkt in der Methode definieren (oder als Feld im Hub)
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	// 2. HTTP Upgrade
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.Log.With("Error", err).Error("Upgrade Error")
		return
	}

	// 3. Kontext und Session-ID vorbereiten
	ctx, cancel := context.WithCancel(context.Background())
	sid := fmt.Sprintf("%d", time.Now().UnixNano())

	// 4. Session initialisieren
	session := &Session{
		UserID:  "", // Wird erst nach auth.login befüllt
		Hub:     h,  // Verweis auf den Hub statt auf den Server
		conn:    conn,
		send:    make(chan []byte, 256),
		Context: ctx,
		cancel:  cancel,
		// Wir nutzen den Logger des Hubs
		log: h.Log.With("sid", sid).With("ip", r.RemoteAddr),
	}

	// 5. Registrierung über den Hub-Kanal
	// Da ServeWebSocket eine Methode von Hub ist, nutzen wir h.register
	h.register <- session

	// 6. Loops starten
	go session.writePump()
	go session.readPump()
}

// Zukunftsmusik: Ein WebTransport Handler
/*
type ServerWT struct {
    hub *Server
}

func (wt *ServerWT) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // 1. Upgrade auf WebTransport (statt WebSocket)
    conn, _ := webtransport.Upgrade(w, r)

    // 2. Erstelle Session
    // Der Hub merkt gar nicht, dass die Daten jetzt per UDP/QUIC kommen!
    session := &Session{ ... }
    wt.hub.register <- session
}
*/
