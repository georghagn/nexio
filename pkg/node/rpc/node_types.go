// Copyright 2025 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

package rpc

import (
	"encoding/json"
	"fmt"
)

type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      json.RawMessage `json:"id,omitempty"`
}

type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
	ID      json.RawMessage `json:"id"`
}

type RPCError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

const JRPCVERSION string = "2.0"

const (
	ErrCodeParseError      = -32700
	ErrConnectionLostError = -32701
	ErrCodeJSONError       = -32702
	ErrCodeInvalidRequest  = -32600
	ErrCodeMethodNotFound  = -32601
	ErrCodeInvalidParams   = -32602
	ErrCodeInternalError   = -32603

	// Custom App Error Codes (Beispiel)
	ErrCodeUnauthorized = 401
	ErrCodeForbidden    = 403
)

// --- Error Messages Map ---
// Hier definieren wir den Standard-Text zu jedem Code
var stdErrorMessages = map[int]string{
	ErrCodeParseError:      "Parse error",
	ErrConnectionLostError: "Connection lost during request",
	ErrCodeJSONError:       "JSON konnte nicht erstellt werden",
	ErrCodeInvalidRequest:  "Invalid Request",
	ErrCodeMethodNotFound:  "Method not found",
	ErrCodeInvalidParams:   "Invalid params",
	ErrCodeInternalError:   "Internal error",
	ErrCodeUnauthorized:    "Unauthorized",
	ErrCodeForbidden:       "Forbidden",
}

// Hilfsfunktion zur Erstellung von Fehlern
func NewRPCError(code int, data any) *RPCError {
	msg, ok := stdErrorMessages[code]
	if !ok {
		msg = "Server error" // Fallback, falls Code unbekannt
	}
	rpcErr := &RPCError{
		Code:    code,
		Message: msg,
	}

	if data != nil {
		// Wir serialisieren die Zusatzdaten direkt hier in RawMessage
		if b, err := json.Marshal(data); err == nil {
			rpcErr.Data = b
		}
	}
	return rpcErr
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("RPC Error %d: %s %s", e.Code, e.Message, e.Data)
}
