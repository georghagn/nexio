// Copyright 2025 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

package nexIO

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect" // <--- Wichtig für RegisterService
	"sync"
)

// --- Daten-Strukturen ---

type RPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	ID      interface{}     `json:"id"`
}

type RPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// --- Interfaces & Typen ---

type HandlerFunc func(params json.RawMessage) (interface{}, error)

// Logger Interface für Injection
type Logger interface {
	Errorf(format string, args ...interface{})
	Debugf(format string, args ...interface{})
}

// --- Server ---

type Server struct {
	handlers map[string]HandlerFunc
	mu       sync.RWMutex
	logger   Logger // Optionaler Logger
}

func New() *Server {
	return &Server{
		handlers: make(map[string]HandlerFunc),
	}
}

// SetLogger injiziert den Logger.
func (s *Server) SetLogger(l Logger) {
	s.logger = l
}

// Register registriert eine manuelle Handler-Funktion.
func (s *Server) Register(methodName string, handler HandlerFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[methodName] = handler
}

// RegisterService nimmt ein Struct und registriert alle Methoden als RPC-Handler.
// Name-Schema: "StructName.MethodName"
func (s *Server) RegisterService(service interface{}) error {
	val := reflect.ValueOf(service)
	typ := reflect.TypeOf(service)

	// Name des Structs ermitteln (z.B. "SchedulerRPC")
	serviceName := reflect.Indirect(val).Type().Name()

	// Über alle Methoden iterieren
	for i := 0; i < typ.NumMethod(); i++ {
		method := typ.Method(i)
		methodName := method.Name

		// Private Methoden ignorieren
		if method.PkgPath != "" {
			continue
		}

		// RPC Name bauen: "SchedulerRPC.Start"
		rpcName := fmt.Sprintf("%s.%s", serviceName, methodName)

		// Wrapper erstellen
		handler := s.createReflectHandler(val, method.Func)

		s.Register(rpcName, handler)

		if s.logger != nil {
			s.logger.Debugf("NexIO: Registered Method '%s'", rpcName)
		}
	}
	return nil
}

// createReflectHandler baut den Wrapper um die Go-Methode
func (s *Server) createReflectHandler(receiver reflect.Value, methodFunc reflect.Value) HandlerFunc {
	return func(rawParams json.RawMessage) (interface{}, error) {
		methodType := methodFunc.Type()

		// Wir erwarten Receiver + 1 Argument (Params)
		if methodType.NumIn() != 2 {
			return nil, fmt.Errorf("method needs exactly 1 argument struct")
		}

		// Typ des Arguments ermitteln
		argType := methodType.In(1)
		argValue := reflect.New(argType) // Neue Instanz davon

		// JSON in das Argument-Struct parsen
		if len(rawParams) > 0 {
			if err := json.Unmarshal(rawParams, argValue.Interface()); err != nil {
				return nil, fmt.Errorf("invalid params: %v", err)
			}
		}

		// Aufrufen: method(receiver, arg)
		returnValues := methodFunc.Call([]reflect.Value{receiver, argValue.Elem()})

		// Rückgabe prüfen: (Result, error)
		if len(returnValues) != 2 {
			return nil, fmt.Errorf("method must return (result, error)")
		}

		// Error prüfen (2. Rückgabewert)
		errInter := returnValues[1].Interface()
		if errInter != nil {
			return nil, errInter.(error)
		}

		// Resultat zurückgeben (1. Rückgabewert)
		return returnValues[0].Interface(), nil
	}
}

// --- Runtime Logik ---

// ProcessRequest ist die transport-unabhängige Kernlogik.
func (s *Server) ProcessRequest(req RPCRequest) RPCResponse {
	if s.logger != nil {
		s.logger.Debugf("NexIO Request: %s", req.Method)
	}

	s.mu.RLock()
	handler, exists := s.handlers[req.Method]
	s.mu.RUnlock()

	if !exists {
		return RPCResponse{
			JSONRPC: "2.0",
			Error:   &RPCError{Code: -32601, Message: "Method not found"},
			ID:      req.ID,
		}
	}

	result, err := handler(req.Params)
	if err != nil {
		if s.logger != nil {
			s.logger.Errorf("NexIO Error in %s: %v", req.Method, err)
		}
		return RPCResponse{
			JSONRPC: "2.0",
			Error:   &RPCError{Code: -32000, Message: err.Error()},
			ID:      req.ID,
		}
	}

	return RPCResponse{
		JSONRPC: "2.0",
		Result:  result,
		ID:      req.ID,
	}
}

// ServeHTTP Implementierung
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "JSON-RPC must be POST", http.StatusMethodNotAllowed)
		return
	}

	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	var req RPCRequest
	if err := json.Unmarshal(body, &req); err != nil {
		s.writeJSON(w, RPCResponse{
			JSONRPC: "2.0",
			Error:   &RPCError{Code: -32700, Message: "Parse error"},
		})
		return
	}

	resp := s.ProcessRequest(req)
	s.writeJSON(w, resp)
}

func (s *Server) writeJSON(w http.ResponseWriter, resp RPCResponse) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Helper
func ParseParams(raw json.RawMessage, target interface{}) error {
	if len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, target)
}
