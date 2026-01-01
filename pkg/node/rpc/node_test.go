package rpc

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/georghagn/gsf-suite/pkg/node/transport"
)

func TestPeerRPC(t *testing.T) {
	// Erstelle einen Kontext, den wir explizit abbrechen können
	ctx, cancel := context.WithCancel(context.Background())

	// t.Cleanup sorgt dafür, dass die Goroutine am Ende des Tests gestoppt wird
	t.Cleanup(func() {
		cancel()
		time.Sleep(5 * time.Millisecond) // Kleiner Puffer zum "Ausschwingen"
	})

	// 1. Zwei Channels für die bidirektionale Kommunikation
	ch1 := make(chan []byte, 10)
	ch2 := make(chan []byte, 10)

	// 2. Zwei verbundene Transport-Enden erstellen
	clientConn := &transport.MemConnection{In: ch1, Out: ch2}
	serverConn := &transport.MemConnection{In: ch2, Out: ch1}

	serverNode := NewNode(serverConn, nil, "", nil)

	// 3. Server-Loop im Hintergrund starten
	go serverNode.Listen(ctx)

	// 4. Test-Case: Ping
	serverNode.Register("ping", func(ctx context.Context, p json.RawMessage) (any, error) {
		return "pong", nil
	})

	t.Run("Ping-Pong", func(t *testing.T) {
		req := Request{
			JSONRPC: "2.0",
			Method:  "ping",
			ID:      json.RawMessage(`1`),
		}
		reqBytes, _ := json.Marshal(req)

		// Request senden
		clientConn.Send(ctx, reqBytes)

		// Antwort empfangen
		respBytes, err := clientConn.Receive(ctx)
		if err != nil {
			t.Fatalf("Failed to receive: %v", err)
		}

		var resp Response
		json.Unmarshal(respBytes, &resp)

		if string(resp.Result) != `"pong"` {
			t.Errorf("Expected pong, got %s", string(resp.Result))
		}
	})
	serverNode.Register("echo", func(ctx context.Context, p json.RawMessage) (any, error) {
		return p, nil
	})

	t.Run("Echo-Complex", func(t *testing.T) {
		complexData := `{"foo":"bar","value":42}`
		req := Request{
			JSONRPC: "2.0",
			Method:  "echo",
			Params:  json.RawMessage(complexData),
			ID:      json.RawMessage(`2`),
		}
		reqBytes, _ := json.Marshal(req)

		clientConn.Send(ctx, reqBytes)
		respBytes, _ := clientConn.Receive(ctx)

		var resp Response
		json.Unmarshal(respBytes, &resp)

		if string(resp.Result) != complexData {
			t.Errorf("Expected %s, got %s", complexData, string(resp.Result))
		}
	})

}

func TestBidirectionalCall(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Hier reicht defer, da der Test keine Sub-Tests hat

	ch1 := make(chan []byte, 10)
	ch2 := make(chan []byte, 10)
	clientConn := &transport.MemConnection{In: ch1, Out: ch2}
	serverConn := &transport.MemConnection{In: ch2, Out: ch1}

	serverNode := NewNode(serverConn, nil, "", nil)
	clientNode := NewNode(clientConn, nil, "", nil)

	// WICHTIG: Der Server muss registrieren, damit er auf den Call des Clients antworten kann
	serverNode.Register("ping", func(ctx context.Context, p json.RawMessage) (any, error) {
		return "pong", nil
	})
	//ctx := context.Background()
	go serverNode.Listen(ctx)
	go clientNode.Listen(ctx)

	// Der Client ruft den Server auf
	res, err := clientNode.Call(ctx, "ping", nil)

	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}

	if string(res) != `"pong"` {
		t.Errorf("Expected pong, got %s", string(res))
	}
}
