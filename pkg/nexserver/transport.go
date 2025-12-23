package nexIOserver

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// ServerWS kümmert sich nur um das Upgrade und die Verknüpfung zum Hub
type ServerWS struct {
	hub      *Server
	Upgrader websocket.Upgrader
}

func NewServerWS(hub *Server) *ServerWS {
	return &ServerWS{
		hub: hub,
		Upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
	}
}

func (ws *ServerWS) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		ws.hub.Log.With("Error", err).Error("Upgrade Error")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	sid := fmt.Sprintf("%d", time.Now().UnixNano())

	session := &Session{
		ID:      sid,
		server:  ws.hub,
		conn:    conn,
		send:    make(chan []byte, 256),
		Store:   make(map[string]interface{}),
		Context: ctx,
		cancel:  cancel,
		log:     ws.hub.Log.With("ip", r.RemoteAddr),
	}

	ws.hub.register <- session

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
