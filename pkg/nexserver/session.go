package nexIOserver

import (
	"context"
	"encoding/json"
	"time"

	"github.com/georghagn/gsf-suite/pkg/gsflog"
	"github.com/georghagn/gsf-suite/pkg/nexproto"
	"github.com/gorilla/websocket"
)

type RPCHandlerFunc func(ctx context.Context, s *Session, params json.RawMessage) (interface{}, *nexIOproto.RPCError)

// --- Konstanten für stabiles Verbindungs-Management ---
const (
	// Zeit, die wir warten, um eine Nachricht zu schreiben (Write Deadline)
	writeWait = 10 * time.Second

	// Zeit, wie lange wir auf ein Pong vom Client warten (Read Deadline)
	pongWait = 60 * time.Second

	// Intervall, in dem wir Pings an den Client senden.
	// MUSS kleiner sein als pongWait (z.B. 90% davon).
	pingPeriod = (pongWait * 9) / 10

	// Maximale Nachrichtengröße (in Bytes)
	maxMessageSize = 512
)

// Session ... (Struct bleibt gleich)
type Session struct {
	ID string

	IsAuth  bool
	Store   map[string]interface{}
	Context context.Context

	cancel context.CancelFunc

	server *Server
	conn   *websocket.Conn
	send   chan []byte

	log gsflog.LogSink
}

// writePump: Sendet Nachrichten UND Pings
func (s *Session) writePump() {
	// Ticker startet den Herzschlag (Ping)
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		s.conn.Close()
	}()

	for {
		select {
		case message, ok := <-s.send:
			// Wir setzen eine Deadline, damit wir nicht ewig blockieren
			s.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				s.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := s.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Optimierung: Alle wartenden Nachrichten im Channel direkt mitsenden
			n := len(s.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-s.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		// --- NEU: Der Ticker feuert regelmäßig ---
		case <-ticker.C:
			s.conn.SetWriteDeadline(time.Now().Add(writeWait))
			// Wir senden einen PING (OpCode 9). Der Browser antwortet automatisch mit PONG.
			if err := s.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return // Wenn Ping fehlschlägt, ist die Verbindung tot -> Abbruch
			}

		case <-s.Context.Done():
			return
		}
	}
}

// readPump: Empfängt Nachrichten UND verarbeitet Pongs
func (s *Session) readPump() {
	defer func() {
		s.server.unregister <- s
		s.conn.Close()
		s.cancel()
	}()

	s.conn.SetReadLimit(maxMessageSize)

	// --- NEU: Deadlines initialisieren ---
	s.conn.SetReadDeadline(time.Now().Add(pongWait))

	// Wenn ein PONG (OpCode 10) reinkommt, verlängern wir die Deadline
	s.conn.SetPongHandler(func(string) error {
		s.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := s.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				//s.log.Warn("WS Error: %v", err)
				s.log.With("Error", err).Warn("WS Error")
			} else {
				s.log.Debug("WebSocket closed normally")
			}
			break
		}

		// LOGGING: Nachricht empfangen
		s.log.With("Message", string(message)).Debug("RX Raw Message")

		var req nexIOproto.RPCRequest
		if err := json.Unmarshal(message, &req); err != nil {
			s.log.Warn("JSON Parse Error")
			s.sendError(nil, nexIOproto.NewRPCError(nexIOproto.ErrCodeParse, nil))
			continue
		}

		s.server.mu.RLock()
		handler, exists := s.server.handlers[req.Method]
		s.server.mu.RUnlock()

		if !exists {
			s.log.With("req.Method", req.Method).Warn("Method not found")
			s.sendError(req.ID, nexIOproto.NewRPCError(nexIOproto.ErrCodeMethodNotFound, req.Method))
			continue
		}

		go func(r nexIOproto.RPCRequest, h RPCHandlerFunc) {
			// Wenn der Kanal geschlossen ist, stirbt diese Goroutine leise,
			// statt den ganzen Server mitzureißen.
			defer func() {
				if r := recover(); r != nil {
					s.log.Error("PANIC in Handler!")
					// Optional: log.Printf("Senden fehlgeschlagen (Verbindung weg): %v", r)
				}
			}()

			s.log.With("r.Method", r.Method).Debug("Executing Handler")

			start := time.Now() // Zeitmessung
			res, rpcErr := h(s.Context, s, r.Params)
			duration := time.Since(start)

			//s.log.Debug("Handler finished: " + r.Method + " took " + duration.String())
			s.log.With("r.Method", r.Method).With("took", duration.String()).Debug("Handler finished")

			res, rpcErr = h(s.Context, s, r.Params)
			if r.ID != nil {
				resp := nexIOproto.RPCResponse{JSONRPC: "2.0", ID: r.ID}
				if rpcErr != nil {
					resp.Error = rpcErr
					s.log.With("rpcErr.Message", rpcErr.Message).Warn("Handler returned Error")
				} else {
					resp.Result = res
				}
				bytes, _ := json.Marshal(resp)
				select {
				case s.send <- bytes:
				default:
					s.log.Error("Send Buffer full or channel closed, dropping response")
				}
			}
		}(req, handler)
	}
}

// Helper zum Senden von Fehlern
func (s *Session) sendError(id *json.RawMessage, errObj *nexIOproto.RPCError) {
	resp := nexIOproto.RPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   errObj,
	}
	b, _ := json.Marshal(resp)
	s.send <- b
}

// Helper: Holt den User aus dem Store, wenn eingeloggt
// Gibt nil zurück, wenn nicht eingeloggt.
func (s *Session) GetUser() interface{} {
	if !s.IsAuth {
		return nil
	}
	if u, ok := s.Store["user"]; ok {
		return u
	}
	return nil
}

// Helper: Typsicherer Zugriff (erfordert, dass du das User-Struct kennst)
// Da "User" aber Teil deiner App-Logik ist und nicht nexIO,
// belassen wir es hier beim generischen Interface.
