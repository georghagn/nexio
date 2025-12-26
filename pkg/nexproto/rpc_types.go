package nexIOproto

import (
	"encoding/json"
)

// RPCRequest: What comes from the client
type RPCRequest struct {
	JSONRPC string           `json:"jsonrpc"`
	Method  string           `json:"method"`
	Params  json.RawMessage  `json:"params,omitempty"` // Raw for performance
	ID      *json.RawMessage `json:"id,omitempty"`     // Pointer, because null by notifications
}

// RPCResponse: What we send back
type RPCResponse struct {
	JSONRPC string           `json:"jsonrpc"`
	Result  interface{}      `json:"result,omitempty"`
	Error   *RPCError        `json:"error,omitempty"`
	ID      *json.RawMessage `json:"id"`
}

// RPCError: Structured errors
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// RPCNotification: Structure for Server-to-Client Messages (Broadcasts)
type RPCNotification struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// --- Standard JSON-RPC 2.0 Error Codes ---
const (
	ErrCodeParse          = -32700
	ErrCodeInvalidRequest = -32600
	ErrCodeMethodNotFound = -32601
	ErrCodeInvalidParams  = -32602
	ErrCodeInternal       = -32603

	// Custom App Error Codes (Beispiel)
	ErrCodeUnauthorized = 401
	ErrCodeForbidden    = 403
)

// --- Error Messages Map ---
// Hier definieren wir den Standard-Text zu jedem Code
var stdErrorMessages = map[int]string{
	ErrCodeParse:          "Parse error",
	ErrCodeInvalidRequest: "Invalid Request",
	ErrCodeMethodNotFound: "Method not found",
	ErrCodeInvalidParams:  "Invalid params",
	ErrCodeInternal:       "Internal error",
	ErrCodeUnauthorized:   "Unauthorized",
	ErrCodeForbidden:      "Forbidden",
}

// NewRPCError erstellt einen Fehler und füllt die Message automatisch aus.
// data ist optional (kann nil sein).
func NewRPCError(code int, data interface{}) *RPCError {
	msg, ok := stdErrorMessages[code]
	if !ok {
		msg = "Server error" // Fallback, falls Code unbekannt
	}

	return &RPCError{
		Code:    code,
		Message: msg,
		Data:    data,
	}
}
