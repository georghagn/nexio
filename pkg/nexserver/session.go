package nexIOserver

import (
	"context"
	"encoding/json"
	"time"

	"github.com/georghagn/gsf-suite/pkg/gsflog"
	"github.com/georghagn/gsf-suite/pkg/nexproto" // Achte darauf, ob nexproto oder nexIOproto!
	"github.com/gorilla/websocket"
)

// RPCHandlerFunc Signatur (angepasst an deine Bedürfnisse)
type RPCHandlerFunc func(ctx context.Context, s *Session, params json.RawMessage) (interface{}, *nexIOproto.RPCError)

type Session struct {
	UserID string
	IsAuth bool
	Store  map[string]interface{}

	Context context.Context
	cancel  context.CancelFunc

	// REFAC: Wir zeigen jetzt direkt auf den Hub, nicht mehr auf den Server
	Hub  *Hub
	conn *websocket.Conn
	send chan []byte

	log gsflog.LogSink
}

func (s *Session) writePump() {
	// Wir nutzen die Werte aus den Hub-Optionen (kein Hardcoding mehr!)
	ticker := time.NewTicker(s.Hub.options.PingPeriod)

	defer func() {
		ticker.Stop()
		s.conn.Close()
	}()

	for {
		select {
		case message, ok := <-s.send:
			s.conn.SetWriteDeadline(time.Now().Add(s.Hub.options.WriteDeadline))

			if !ok {
				s.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := s.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Optimierung: Wartende Nachrichten mitsenden
			n := len(s.send)
			for i := 0; i < n; i++ {
				w.Write(<-s.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			s.conn.SetWriteDeadline(time.Now().Add(s.Hub.options.WriteDeadline))
			if err := s.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-s.Context.Done():
			return
		}
	}
}

func (s *Session) readPump() {
	defer func() {
		// REFAC: Zugriff auf Hub Kanäle
		s.Hub.unregister <- s
		s.conn.Close()
		s.cancel()
	}()

	// Hier nutzen wir wieder die Hub-Optionen
	s.conn.SetReadLimit(4096) // Oder s.Hub.options.MaxMessageSize
	s.conn.SetReadDeadline(time.Now().Add(s.Hub.options.PongWait))

	s.conn.SetPongHandler(func(string) error {
		s.conn.SetReadDeadline(time.Now().Add(s.Hub.options.PongWait))
		return nil
	})

	for {
		_, message, err := s.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				s.log.With("Error", err).Warn("WS Unexpected Close")
			}
			break
		}

		s.log.With("Message", string(message)).Debug("RX Raw")

		var req nexIOproto.RPCRequest
		if err := json.Unmarshal(message, &req); err != nil {
			s.sendError(nil, nexIOproto.NewRPCError(nexIOproto.ErrCodeParse, nil))
			continue
		}

		// REFAC: Zugriff auf Handler über den Hub
		s.Hub.mu.RLock()
		handler, exists := s.Hub.handlers[req.Method]
		s.Hub.mu.RUnlock()

		if !exists {
			s.sendError(req.ID, nexIOproto.NewRPCError(nexIOproto.ErrCodeMethodNotFound, req.Method))
			continue
		}

		go func(r nexIOproto.RPCRequest, h RPCHandlerFunc) {
			defer func() {
				if r := recover(); r != nil {
					s.log.Error("PANIC in Handler")
				}
			}()

			res, rpcErr := h(s.Context, s, r.Params)

			if r.ID != nil {
				resp := nexIOproto.RPCResponse{JSONRPC: "2.0", ID: r.ID}
				if rpcErr != nil {
					resp.Error = rpcErr
				} else {
					resp.Result = res
				}
				bytes, _ := json.Marshal(resp)
				s.send <- bytes
			}
		}(req, handler)
	}
}

func (s *Session) sendError(id *json.RawMessage, errObj *nexIOproto.RPCError) {
	resp := nexIOproto.RPCResponse{JSONRPC: "2.0", ID: id, Error: errObj}
	b, _ := json.Marshal(resp)
	select {
	case s.send <- b:
	default:
	}
}

// Bind entfernt die alte s.server Abhängigkeit
func (s *Session) Bind(userID string) {
	s.Hub.BindSessionToUser(s, userID)
}
