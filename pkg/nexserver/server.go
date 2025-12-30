package nexIOserver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/georghagn/gsf-suite/pkg/gsflog"
	"github.com/georghagn/gsf-suite/pkg/nexproto"
)

type Server struct {
	authenticator Authenticator // Das Interface-Feld
	Hub           *Hub
	Log           gsflog.LogSink
	options       *Options
}

func New(logger gsflog.LogSink, auth Authenticator, opts *Options) *Server {
	if auth == nil {
		auth = &DummyAuthenticator{}
	}
	server := &Server{
		Hub:           NewHub(logger, opts),
		authenticator: auth,
		Log:           logger.With("component", "nexIOserver"),
		options:       opts,
	}

	server.registerDefaultHandlers()
	return server
}

// Register fügt einen Handler hinzu (Thread-Safe)
func (s *Server) Register(method string, handler RPCHandlerFunc) {
	s.Hub.mu.Lock()
	defer s.Hub.mu.Unlock()
	s.Hub.handlers[method] = handler
}

func (s *Server) registerDefaultHandlers() {
	// Wir hüllen s.handleLogin ein, um die Typ-Unterschiede (Context & Error-Typ) auszugleichen
	s.Hub.Handle("auth.login", func(ctx context.Context, session *Session, params json.RawMessage) (interface{}, *nexIOproto.RPCError) {
		result, err := s.handleLogin(session, params)
		if err != nil {
			// Wir wandeln den normalen Go-Error in einen nexIOproto.RPCError um
			return nil, nexIOproto.NewRPCError(nexIOproto.ErrCodeInternal, err.Error())
		}
		return result, nil
	})
}

// BindUser verknüpft eine Session mit einer UserID
func (s *Server) BindUser(session *Session, userID string) {
	s.Hub.mu.Lock()
	defer s.Hub.mu.Unlock()

	// Prüfen, ob diese Session bereits für diesen User registriert ist
	for _, sess := range s.Hub.userSessions[userID] {
		if sess == session {
			return // Schon drin, nichts tun
		}
	}

	s.Hub.userSessions[userID] = append(s.Hub.userSessions[userID], session)
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

	s.Hub.mu.RLock()
	defer s.Hub.mu.RUnlock()

	for session := range s.Hub.sessions {
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
	close(s.Hub.shutdown)
}

// SendToUser sendet eine Notification an alle Verbindungen eines bestimmten Users
func (s *Server) SendToUser(userID string, method string, params interface{}) error {
	s.Hub.mu.RLock()
	defer s.Hub.mu.RUnlock()

	sessions, found := s.Hub.userSessions[userID]
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

func (s *Server) handleLogin(session *Session, params json.RawMessage) (interface{}, error) {
	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	// 1. Parameter parsen
	if err := json.Unmarshal(params, &creds); err != nil {
		return nil, fmt.Errorf("invalid login parameters")
	}

	// 2. Authentifizierung prüfen
	userID, success := s.authenticator.Authenticate(creds.Username, creds.Password)

	// FEHLERFALL
	if !success {
		s.Log.With("User", creds.Username).Warn("Login failed for user")
		return nil, fmt.Errorf("auth failed") // Hier brechen wir ab!
	}

	// ERFOLGSFALL
	// 3. Session im Hub branden (Identity-Management)
	// Nutze hier den exakten Namen der Methode, die du in hub.go erstellt hast
	s.Hub.BindSessionToUser(session, userID)

	s.Log.With("UserID", userID).Info("User logged in successfully")

	return map[string]string{
		"status":  "success",
		"user_id": userID,
	}, nil
}
