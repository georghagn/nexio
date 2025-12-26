package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"sync"

	// Framework importieren
	"github.com/georghagn/gsf-suite/pkg/nexproto"
	"github.com/georghagn/gsf-suite/pkg/nexserver"
)

// --- 1. State & Helper (Private) ---

var (
	tokenStore = make(map[string]string)
	tokenMu    sync.RWMutex

	// Kein Hardcoding mehr nötig, aber Default hilft beim Testen
	secret = "default-secret"
)

func generateRandomToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// --- 2. Params Structs (Private) ---

type loginParams struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type resumeParams struct {
	Token string `json:"token"`
}

// --- 3. Handler Logik (Private) ---
// Wir müssen die Funktionen nicht exportieren (Großschreiben),
// da wir sie nur unten in RegisterRoutes nutzen.

func handleLogin(ctx context.Context, s *nexIOserver.Session, params json.RawMessage) (interface{}, *nexIOproto.RPCError) {
	var p loginParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, nexIOproto.NewRPCError(nexIOproto.ErrCodeInvalidParams, err.Error())
	}

	// Fake DB Check
	if p.Username == "admin" && p.Password == secret {
		token := generateRandomToken()

		tokenMu.Lock()
		tokenStore[token] = p.Username
		tokenMu.Unlock()

		s.IsAuth = true
		s.BindUser(p.Username)
		s.Store["user"] = p.Username

		return map[string]string{"status": "Login OK", "token": token}, nil
	}

	return nil, nexIOproto.NewRPCError(nexIOproto.ErrCodeUnauthorized, "Bad credentials")
}

func handleResume(ctx context.Context, s *nexIOserver.Session, params json.RawMessage) (interface{}, *nexIOproto.RPCError) {
	var p resumeParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, nexIOproto.NewRPCError(nexIOproto.ErrCodeInvalidParams, err.Error())
	}

	tokenMu.RLock()
	user, ok := tokenStore[p.Token]
	tokenMu.RUnlock()

	if !ok {
		return nil, nexIOproto.NewRPCError(nexIOproto.ErrCodeUnauthorized, "Token invalid")
	}

	s.IsAuth = true
	s.Store["user"] = user

	return map[string]string{"status": "Resumed", "user": user}, nil
}

// --- 4. Public API (Das "Interface" nach außen) ---
func Configure(passwd string) {
	secret = passwd
}

// RegisterRoutes verbindet diesen Service mit dem Server
func RegisterRoutes(hub *nexIOserver.Server) {
	hub.Register("auth.login", handleLogin)
	hub.Register("auth.resume", handleResume)
}
