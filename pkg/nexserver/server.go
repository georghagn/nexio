package nexIOserver

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/georghagn/gsf-suite/pkg/gsfconfig"
	"github.com/georghagn/gsf-suite/pkg/gsflog"
	"github.com/georghagn/gsf-suite/pkg/nexproto"
	"github.com/gorilla/websocket"
)

type Server struct {
	// Alle aktiven Verbindungen
	sessions map[*Session]bool
	// Das Telefonbuch (UserID -> Liste von Sessions)
	userSessions map[string][]*Session
	// Registrierte RPC-Handler (z.B. auth.login)
	handlers map[string]RPCHandlerFunc

	// Kanäle für die Steuerung
	broadcast  chan []byte   // Nachrichten an alle
	register   chan *Session // Neue Verbindung
	unregister chan *Session // Verbindung beendet
	shutdown   chan struct{} // Server stoppen (ehemals quit)

	Log     gsflog.LogSink
	options *Options

	mu sync.RWMutex // Mutex für Threadsicherheit bei Map-Zugriffen
}

func New(logger gsflog.LogSink, cfg gsfconfig.ProtocolConfig) *Server {
	return &Server{
		broadcast:    make(chan []byte),
		register:     make(chan *Session),
		unregister:   make(chan *Session),
		sessions:     make(map[*Session]bool),
		userSessions: make(map[string][]*Session),
		handlers:     make(map[string]RPCHandlerFunc),
		shutdown:     make(chan struct{}),

		Log: logger.With("component", "nexIOserver"),
		//options: cfg,
		options: nil,
	}
}

// Register fügt einen Handler hinzu (Thread-Safe)
func (s *Server) Register(method string, handler RPCHandlerFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[method] = handler
}

// BindUser verknüpft eine Session mit einer UserID
func (s *Server) BindUser(session *Session, userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Prüfen, ob diese Session bereits für diesen User registriert ist
	for _, sess := range s.userSessions[userID] {
		if sess == session {
			return // Schon drin, nichts tun
		}
	}

	s.userSessions[userID] = append(s.userSessions[userID], session)
	session.Store["userID"] = userID

	s.Log.With("user_id", userID).Info("User bound to session")
}

// BroadcastToAuthenticated sendet eine Notification an alle eingeloggten User
// Dies ist eine Helper-Methode, die du von außen aufrufen kannst.
// Sie baut das JSON und wirft es in den broadcast-Channel.
func (s *Server) BroadcastToAuthenticated(method string, params interface{}) {
	// 1. Die params (interface{}) in JSON-Bytes umwandeln
	pBytes, err := json.Marshal(params)
	if err != nil {
		s.Log.With("error", err).Error("Broadcast: Marshal params failed")
		return
	}
	notification := nexIOproto.RPCNotification{
		JSONRPC: "2.0",
		Method:  method,
		Params:  json.RawMessage(pBytes),
	}

	// 2. In Bytes umwandeln für den Versand
	msgBytes, err := json.Marshal(notification)
	if err != nil {
		s.Log.With("error", err).Error("Broadcast: Marshal notification failed")
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for session := range s.sessions {
		if session.IsAuth {
			select {
			case session.send <- msgBytes:
			default:
				s.Log.Warn("Broadcast: session buffer full, skipping")
			}
		}
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
				// 1. Aus der Haupt-Map löschen
				delete(s.sessions, session)

				// 2. Aus dem User-Telefonbuch löschen
				if userID, ok := session.Store["userID"].(string); ok {
					sessions := s.userSessions[userID]
					for i, sess := range sessions {
						// WICHTIG: Wir vergleichen den Pointer!
						if sess == session {
							s.userSessions[userID] = append(sessions[:i], sessions[i+1:]...)
							break
						}
					}
					if len(s.userSessions[userID]) == 0 {
						delete(s.userSessions, userID)
					}
				}

				// 3. Kanal schließen
				close(session.send)
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

// SendToUser sendet eine Notification an alle Verbindungen eines bestimmten Users
func (s *Server) SendToUser(userID string, method string, params interface{}) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessions, found := s.userSessions[userID]
	if !found || len(sessions) == 0 {
		return fmt.Errorf("user '%s' not connected", userID)
	}

	// 1. Die params (interface{}) in JSON-Bytes umwandeln
	pBytes, err := json.Marshal(params)
	if err != nil {
		return err
	}
	notification := nexIOproto.RPCNotification{
		JSONRPC: "2.0",
		Method:  method,
		Params:  json.RawMessage(pBytes),
	}
	bytes, err := json.Marshal(notification)
	if err != nil {
		return err
	}

	for _, sess := range sessions {
		select {
		case sess.send <- bytes:
		default:
			// Buffer voll, Pech gehabt (oder kicken)
		}
	}
	return nil
}
