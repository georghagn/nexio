// Copyright 2026 Georg Hagn
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
	connMu sync.RWMutex //Protects the connection during the exchange.
	conn   transport.Connection

	handlers map[string]HandlerFunc
	mu       sync.RWMutex

	pending   map[string]pendingRequest
	pendingMu sync.Mutex
	nextID    uint64

	// For the reconnect mechanism
	dialAddr string
	provider *transport.WSProvider

	Log transport.LogSink
}

// We need someone to handle outstanding answers.
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

// Call sends a request and blocks until the response arrives.
func (node *Node) Call(ctx context.Context, method string, params any) (json.RawMessage, error) {
	// 1. Secure connection
	node.connMu.RLock()
	currentConn := node.conn
	node.connMu.RUnlock()

	// If a reconnect is in progress or the connection is lost: No panic!
	if currentConn == nil {
		return nil, NewRPCError(ErrCodeInternalError, "The connection is currently being re-established.")
	}

	// 2. ID generieren und in pending-Map registrieren
	node.pendingMu.Lock()
	node.nextID++
	idStr := fmt.Sprintf("%d", node.nextID)
	idJSON, _ := json.Marshal(idStr)
	ch := make(chan Response, 1)
	node.pending[idStr] = pendingRequest{done: ch}
	node.pendingMu.Unlock()

	// Cleaning up after the call
	defer func() {
		node.pendingMu.Lock()
		delete(node.pending, idStr)
		node.pendingMu.Unlock()
	}()

	// 3. Prepare request
	pBytes, _ := json.Marshal(params)
	req := Request{
		JSONRPC: JRPCVERSION,
		Method:  method,
		Params:  pBytes,
		ID:      idJSON,
	}

	data, _ := json.Marshal(req)

	// 4. Send via COPY of the connection
	if err := currentConn.Send(ctx, data); err != nil {
		return nil, err
	}

	// 5. Wait for answer
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
		// 1. Secure connection
		node.connMu.RLock()
		currentConn := node.conn
		node.connMu.RUnlock()

		if currentConn == nil {
			// if we don't have an address (server-side), we can't reconnect.
			if node.dialAddr == "" {
				return fmt.Errorf("Connection lost and no reconnect address available")
			}

			// try Reconnect
			if err := node.attemptReconnect(ctx); err != nil {
				return err
			}
			continue
		}

		// 2.Normal listening
		data, err := currentConn.Receive(ctx)
		if err != nil {
			node.Log.With("error", err).Error("Network error: Preparing to reconnect...")

			// 1. Cut connection
			node.connMu.Lock()
			node.conn = nil
			node.connMu.Unlock()

			// 2. Cancel all pending calls (so they don't get stuck)
			node.cleanupPendingRequests("Connection lost")

			continue

		}

		go node.handleIncoming(ctx, data)
	}
}

// Notify sends a notification to which no response is expected (no ID).
func (node *Node) Notify(ctx context.Context, method string, params any) error {
	// 1. Securely intercept connection (Read-Lock)
	node.connMu.RLock()
	currentConn := node.conn
	node.connMu.RUnlock()

	// If a reconnect is currently in progress: Report the error instead of panicking
	if currentConn == nil {
		return NewRPCError(ErrCodeInternalError, "Notification failed: Reconnecting")
	}

	// 2. Process parameters
	pBytes, err := json.Marshal(params)
	if err != nil {
		return NewRPCError(ErrCodeParseError, []byte(err.Error()))
	}

	// 3. Create request without ID (JSON-RPC Notification)
	req := Request{
		JSONRPC: JRPCVERSION,
		Method:  method,
		Params:  pBytes,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return NewRPCError(ErrCodeJSONError, []byte(err.Error()))
	}

	// 4. Send via secure connection
	if err := currentConn.Send(ctx, data); err != nil {
		return err // Here we are directly returning the network error.
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
			node.Log.With("dialAddr", node.dialAddr).Info("Try Reconnect")
			newConn, err := node.provider.Dial(ctx, node.dialAddr)
			if err == nil {
				node.Log.Info("Reconnect successful!")
				node.connMu.Lock()
				node.conn = newConn
				node.connMu.Unlock()
				return nil
			}

			node.Log.With("err", err).With("backoff", backoff).Error("Failed, next attempt")

			time.Sleep(backoff)
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}
}

func (node *Node) handleIncoming(ctx context.Context, data []byte) {
	// 1. Preliminary Check: Is it a request or a response?
	// We simply check if "method" appears in the JSON (fastest way)
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
			// Here we use the Data field for the error message from Go
			resp.Error = NewRPCError(ErrCodeInternalError, err.Error())
		} else {
			resBytes, _ := json.Marshal(result)
			resp.Result = resBytes
		}
	}

	// IMPORTANT: We will only send a reply if an ID is provided.
	if req.ID != nil && string(req.ID) != "null" {
		respBytes, _ := json.Marshal(resp)
		_ = node.conn.Send(ctx, respBytes)
	} else {
		node.Log.With("req.Method", req.Method).Info("Notification received")
	}
}

func (node *Node) processResponse(resp Response) {
	// Removes quotation marks if present.
	idStr := strings.Trim(string(resp.ID), `"`)

	node.pendingMu.Lock()
	req, ok := node.pending[idStr]
	node.pendingMu.Unlock()

	if ok {
		req.done <- resp // Access via the .done struct field
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
