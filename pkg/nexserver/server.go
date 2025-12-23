package nexIOserver

import (
	"encoding/json"
	"sync"

	"github.com/georghagn/gsf-suite/pkg/gsflog"
	"github.com/georghagn/gsf-suite/pkg/nexproto"
	"github.com/gorilla/websocket"
)

type Server struct {
	// Alle aktiven Verbindungen
	sessions map[*Session]bool

	// Registrierte RPC-Handler (z.B. auth.login)
	handlers map[string]RPCHandlerFunc

	// Kanäle für die Steuerung
	broadcast  chan []byte   // Nachrichten an alle
	register   chan *Session // Neue Verbindung
	unregister chan *Session // Verbindung beendet
	shutdown   chan struct{} // Server stoppen (ehemals quit)

	Log gsflog.LogSink

	// Mutex für Threadsicherheit bei Map-Zugriffen
	mu sync.RWMutex
}

func NewServer(logger gsflog.LogSink) *Server {
	return &Server{
		broadcast:  make(chan []byte),
		register:   make(chan *Session),
		unregister: make(chan *Session),
		sessions:   make(map[*Session]bool),
		handlers:   make(map[string]RPCHandlerFunc),
		shutdown:   make(chan struct{}),

		Log: logger.With("component", "nexServer"),
	}
}

// Register fügt einen Handler hinzu (Thread-Safe)
func (s *Server) Register(method string, handler RPCHandlerFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[method] = handler
}

// BroadcastToAuthenticated sendet eine Notification an alle eingeloggten User
// Dies ist eine Helper-Methode, die du von außen aufrufen kannst.
// Sie baut das JSON und wirft es in den broadcast-Channel.
func (s *Server) BroadcastToAuthenticated(method string, params interface{}) {
	notification := nexIOproto.RPCNotification{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
	// Wir ignorieren Fehler beim Marshaling für dieses Beispiel
	// da wir die Struktur selbst kontrollieren.
	if bytes, err := json.Marshal(notification); err == nil {
		s.broadcast <- bytes
	}
}

// Shutdown initiiert das Herunterfahren
func (s *Server) Shutdown() {
	close(s.shutdown)
}

// Run ist die Hauptschleife des Servers
func (s *Server) Run() {
	s.Log.Info("NexIO Hub gestartet")

	for {
		select {
		// 1. Neuer Client verbindet sich
		case session := <-s.register:
			s.mu.Lock()
			s.sessions[session] = true
			s.mu.Unlock()

		// 2. Client trennt Verbindung
		case session := <-s.unregister:
			s.mu.Lock()
			if _, ok := s.sessions[session]; ok {
				delete(s.sessions, session)
				close(session.send) // Schließt den Send-Kanal der Session
			}
			s.mu.Unlock()

		// 3. Nachricht an ALLE (Der Broadcast Fall)
		case message := <-s.broadcast:
			s.mu.RLock()
			for session := range s.sessions {
				// Hier filtern wir: Nur authentifizierte User bekommen Broadcasts
				if session.IsAuth {
					select {
					case session.send <- message:
					default:
						// Wenn der Client blockiert oder der Buffer voll ist, kicken wir ihn
						close(session.send)
						delete(s.sessions, session)
					}
				}
			}
			s.mu.RUnlock()

		// 4. Server herunterfahren (Ersetzt 'quit')
		case <-s.shutdown:
			s.Log.Info("NexIO Hub fährt herunter...")
			s.mu.Lock()
			for session := range s.sessions {
				// Höfliches "Tschüss" an den Browser senden
				session.conn.WriteMessage(
					websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseGoingAway, "Server shutting down"),
				)
				session.conn.Close()
				close(session.send)
			}
			// Map leeren
			s.sessions = make(map[*Session]bool)
			s.mu.Unlock()
			s.Log.Info("NexIO Hub gestoppt.")
			return
		}
	}
}
