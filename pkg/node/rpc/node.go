// Copyright 2025 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/georghagn/gsf-suite/pkg/node/transport"
)

type HandlerFunc func(ctx context.Context, params json.RawMessage) (any, error)

type Node struct {
	connMu sync.RWMutex //Schützt die Verbindung während des Austauschs
	conn   transport.Connection

	handlers map[string]HandlerFunc
	mu       sync.RWMutex

	pending   map[string]pendingRequest
	pendingMu sync.Mutex
	nextID    uint64

	// Für den Reconnect-Mechanismus
	dialAddr string
	provider *transport.WSProvider

	Log transport.LogSink
}

// Wir brauchen einen Typ für ausstehende Antworten
type pendingRequest struct {
	done chan Response
}

func NewNode(
	conn transport.Connection,
	provider *transport.WSProvider,
	dialAddr string,
	logger transport.LogSink) *Node {
	n := &Node{
		conn:     conn,
		handlers: make(map[string]HandlerFunc),
		pending:  make(map[string]pendingRequest),
		provider: provider,
		dialAddr: dialAddr,
		Log:      &transport.SilentLogger{},
	}
	if logger != nil {
		n.Log = logger
	}
	return n
}

// Call sendet eine Anfrage und blockiert, bis die Antwort kommt
func (node *Node) Call(ctx context.Context, method string, params any) (json.RawMessage, error) {
	// 1. Verbindung sicher abgreifen
	node.connMu.RLock()
	currentConn := node.conn
	node.connMu.RUnlock()

	// Falls gerade ein Reconnect läuft oder die Verbindung weg ist: Kein Panic!
	if currentConn == nil {
		return nil, NewRPCError(ErrCodeInternalError, "Verbindung wird gerade wiederhergestellt")
	}

	// 2. ID generieren und in pending-Map registrieren
	node.pendingMu.Lock()
	node.nextID++
	idStr := fmt.Sprintf("%d", node.nextID)
	idJSON, _ := json.Marshal(idStr)
	ch := make(chan Response, 1)
	node.pending[idStr] = pendingRequest{done: ch}
	node.pendingMu.Unlock()

	// Aufräumen nach dem Call
	defer func() {
		node.pendingMu.Lock()
		delete(node.pending, idStr)
		node.pendingMu.Unlock()
	}()

	// 3. Request vorbereiten
	pBytes, _ := json.Marshal(params)
	req := Request{
		JSONRPC: JRPCVERSION,
		Method:  method,
		Params:  pBytes,
		ID:      idJSON,
	}

	data, _ := json.Marshal(req)

	// 4. Senden über die KOPIE der Verbindung
	if err := currentConn.Send(ctx, data); err != nil {
		return nil, err
	}

	// 5. Auf Antwort warten
	select {
	case resp := <-ch:
		if resp.Error != nil {
			return nil, fmt.Errorf("RPC error %d: %s", resp.Error.Code, resp.Error.Message)
		}
		return resp.Result, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (node *Node) Register(method string, h HandlerFunc) {
	node.mu.Lock()
	defer node.mu.Unlock()
	node.handlers[method] = h
}

func (node *Node) Listen(ctx context.Context) error {
	for {
		// 1. Verbindung sicher abgreifen
		node.connMu.RLock()
		currentConn := node.conn
		node.connMu.RUnlock()

		if currentConn == nil {
			// Wenn wir keine Adresse haben (Server-Seite), können wir nicht re-connecten
			if node.dialAddr == "" {
				return fmt.Errorf("verbindung verloren und keine Reconnect-Adresse vorhanden")
			}

			// Versuche Reconnect
			if err := node.attemptReconnect(ctx); err != nil {
				return err
			}
			continue
		}

		// 2. Normales Zuhören
		// Normales Lesen
		data, err := currentConn.Receive(ctx)
		if err != nil {
			node.Log.With("error", err).Error("Netzwerk-Fehler: Bereite Reconnect vor...") // <--

			// 1. Verbindung kappen
			node.connMu.Lock()
			node.conn = nil
			node.connMu.Unlock()

			// 2. Alle wartenden Calls abbrechen (damit sie nicht hängen)
			node.cleanupPendingRequests("Verbindung verloren")

			continue

		}

		go node.handleIncoming(ctx, data)
	}
}

// Notify sendet eine Benachrichtigung, auf die keine Antwort erwartet wird (keine ID).
func (node *Node) Notify(ctx context.Context, method string, params any) error {
	// 1. Verbindung sicher abgreifen (Read-Lock)
	node.connMu.RLock()
	currentConn := node.conn
	node.connMu.RUnlock()

	// Falls gerade ein Reconnect läuft: Fehler zurückgeben statt Panik
	if currentConn == nil {
		return NewRPCError(ErrCodeInternalError, "Benachrichtigung fehlgeschlagen: Verbindung wird wiederhergestellt")
	}

	// 2. Parameter verarbeiten
	pBytes, err := json.Marshal(params)
	if err != nil {
		return NewRPCError(ErrCodeParseError, []byte(err.Error()))
	}

	// 3. Request ohne ID erstellen (JSON-RPC Notification)
	req := Request{
		JSONRPC: "2.0",
		Method:  method,
		Params:  pBytes,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return NewRPCError(ErrCodeJSONError, []byte(err.Error()))
	}

	// 4. Über die gesicherte Verbindung senden
	if err := currentConn.Send(ctx, data); err != nil {
		return err // Hier geben wir den Netzwerkfehler direkt zurück
	}

	return nil
}

func (node *Node) attemptReconnect(ctx context.Context) error {
	backoff := 1 * time.Second
	maxBackoff := 30 * time.Second

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			node.Log.With("dialAddr", node.dialAddr).Info("Versuche Reconnect")
			newConn, err := node.provider.Dial(ctx, node.dialAddr)
			if err == nil {
				node.Log.Info("Reconnect erfolgreich!")
				node.connMu.Lock()
				node.conn = newConn
				node.connMu.Unlock()
				return nil
			}

			node.Log.With("err", err).With("backoff", backoff).Error("Fehlgeschlagen, Nächster Versuch")

			time.Sleep(backoff)
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}
}

func (node *Node) handleIncoming(ctx context.Context, data []byte) {
	// 1. Vorab-Check: Ist es ein Request oder eine Response?
	// Wir schauen einfach, ob "method" im JSON vorkommt (schnellster Weg)
	if strings.Contains(string(data), `"method"`) {
		var req Request
		if err := json.Unmarshal(data, &req); err == nil {
			node.processRequest(ctx, req)
		}
	} else {
		var resp Response
		if err := json.Unmarshal(data, &resp); err == nil {
			node.processResponse(resp)
		}
	}
}

func (node *Node) processRequest(ctx context.Context, req Request) {
	node.mu.RLock()
	handler, ok := node.handlers[req.Method]
	node.mu.RUnlock()

	var resp Response
	resp.JSONRPC = JRPCVERSION
	resp.ID = req.ID

	if !ok {
		resp.Error = NewRPCError(ErrCodeMethodNotFound, req.Method)
	} else {
		result, err := handler(ctx, req.Params)
		if err != nil {
			// Hier nutzen wir das Data Feld für die Fehlermeldung aus Go
			resp.Error = NewRPCError(ErrCodeInternalError, err.Error())
		} else {
			resBytes, _ := json.Marshal(result)
			resp.Result = resBytes
		}
	}

	// WICHTIG: Nur wenn eine ID vorhanden ist, schicken wir eine Antwort zurück.
	if req.ID != nil && string(req.ID) != "null" {
		respBytes, _ := json.Marshal(resp)
		_ = node.conn.Send(ctx, respBytes)
	} else {
		node.Log.With("req.Method", req.Method).Info("Notification erhalten")
	}
}

func (node *Node) processResponse(resp Response) {
	// Entfernt Anführungszeichen falls vorhanden
	idStr := strings.Trim(string(resp.ID), `"`)

	//	idStr := string(resp.ID)
	//    idStr = strings.Trim(idStr, `"`) // Entfernt Anführungszeichen falls vorhanden

	node.pendingMu.Lock()
	req, ok := node.pending[idStr]
	node.pendingMu.Unlock()

	if ok {
		req.done <- resp // Zugriff über das Struct-Feld .done
	}
}

func (node *Node) cleanupPendingRequests(reason string) {
	node.pendingMu.Lock()
	defer node.pendingMu.Unlock()
	for id, req := range node.pending {
		req.done <- Response{
			ID:    json.RawMessage(id),
			Error: NewRPCError(ErrCodeInternalError, reason),
		}
		delete(node.pending, id)
	}
}
