package nexIOserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// Ein simpler Handler für den Test: Gibt einfach die Params zurück
func rpcEcho(ctx context.Context, s *Session, params json.RawMessage) (interface{}, *RPCError) {
	// Wir erwarten einen String als Params
	var input string
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, &RPCError{Code: -32602, Message: "Invalid Params"}
	}
	return "Echo: " + input, nil
}

func TestIntegration(t *testing.T) {
	// 1. Setup Server (Hub & Transport)
	hub := NewServer()
	hub.Register("test.echo", rpcEcho)

	// Server im Hintergrund starten
	go hub.Run()
	// Sicherstellen, dass der Hub am Ende gestoppt wird
	defer hub.Stop()

	// 2. HTTP Test Server starten
	wsHandler := NewServerWS(hub)
	testServer := httptest.NewServer(http.HandlerFunc(wsHandler.ServeWS))
	defer testServer.Close()

	// URL von http:// zu ws:// umwandeln
	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http")

	// 3. Client verbinden
	// Wichtig: Der Dialer kümmert sich um den Handshake
	clientConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Verbindung fehlgeschlagen: %v", err)
	}
	defer clientConn.Close()

	// --- TEST SZENARIO 1: Einfacher RPC Call ---
	t.Run("RPC Call Response", func(t *testing.T) {
		// A. Request senden
		req := RPCRequest{
			JSONRPC: "2.0",
			Method:  "test.echo",
			Params:  json.RawMessage(`"Hallo Welt"`),
			ID:      rawJSON("1"), // Helper-Funktion unten
		}
		if err := clientConn.WriteJSON(req); err != nil {
			t.Fatalf("Konnte Request nicht senden: %v", err)
		}

		// B. Response lesen
		// Da wir synchron testen, erwarten wir sofort die Antwort
		var resp RPCResponse
		if err := clientConn.ReadJSON(&resp); err != nil {
			t.Fatalf("Konnte Response nicht lesen: %v", err)
		}

		// C. Validieren
		if resp.Error != nil {
			t.Fatalf("RPC Error erhalten: %v", resp.Error)
		}

		// ID prüfen
		if string(*resp.ID) != "1" {
			t.Errorf("Falsche ID. Erwartet 1, bekommen %s", string(*resp.ID))
		}

		// Result prüfen (ist interface{}, daher Typ-Assertion oder Cast nötig)
		resultStr, ok := resp.Result.(string)
		if !ok || resultStr != "Echo: Hallo Welt" {
			t.Errorf("Falsches Result: %v", resp.Result)
		}
	})

	// --- TEST SZENARIO 2: Broadcast empfangen ---
	t.Run("Broadcast Receive", func(t *testing.T) {
		// Wir simulieren, dass der Server (z.B. durch einen anderen Prozess) einen Broadcast auslöst.
		// WICHTIG: Broadcasts gehen nur an "authentifizierte" User.
		// Wir müssen unsere Session im Test also "auth=true" setzen.
		// Das ist tricky im Integration Test ohne echten Login-Handler.
		// Workaround für Test: Wir bauen einen "Login"-Handler oder warten kurz und setzen es intern (dirty hack für Unit Test ok).

		// Sauberer Weg: Wir registrieren schnell einen Auth-Handler
		hub.Register("auth.fake", func(ctx context.Context, s *Session, p json.RawMessage) (interface{}, *RPCError) {
			s.IsAuth = true
			return true, nil
		})

		// Fake-Login durchführen
		authReq := RPCRequest{
			JSONRPC: "2.0",
			Method:  "auth.fake",
			ID:      rawJSON("2"),
		}
		clientConn.WriteJSON(authReq)
		var authResp RPCResponse
		clientConn.ReadJSON(&authResp) // Antwort weglesen

		// Jetzt Broadcast triggern
		expectedMsg := "System shutting down"
		hub.BroadcastToAuthenticated("system.alert", expectedMsg)

		// Client muss Notification empfangen
		// Wir setzen ein Timeout, falls nichts kommt
		clientConn.SetReadDeadline(time.Now().Add(1 * time.Second))

		var notif RPCNotification // Wir nutzen das Struct, das wir vorhin erstellt haben
		if err := clientConn.ReadJSON(&notif); err != nil {
			t.Fatalf("Kein Broadcast empfangen: %v", err)
		}

		if notif.Method != "system.alert" {
			t.Errorf("Falsche Methode: %s", notif.Method)
		}
		if notif.Params.(string) != expectedMsg {
			t.Errorf("Falsche Params: %v", notif.Params)
		}
	})
}

// Helper um *json.RawMessage schnell zu erzeugen
func rawJSON(s string) *json.RawMessage {
	msg := json.RawMessage(s)
	return &msg
}
