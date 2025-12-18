// Copyright 2025 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

package nexIO

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn      *websocket.Conn
	idCounter int
}

func NewClient(host string, path string) (*Client, error) {
	u := url.URL{Scheme: "ws", Host: host, Path: path}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}

	return &Client{conn: conn, idCounter: 1}, nil
}

func (c *Client) Close() {
	c.conn.Close()
}

// Call ruft eine RPC Methode auf und speichert das Resultat in resultPtr
func (c *Client) Call(method string, params interface{}, resultPtr interface{}) error {
	c.idCounter++

	// 1. Parameter zu JSON serialisieren
	paramBytes, _ := json.Marshal(params)

	req := RPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  paramBytes,
		ID:      c.idCounter,
	}

	// 2. Senden
	if err := c.conn.WriteJSON(req); err != nil {
		return err
	}

	// 3. Antwort lesen
	var resp RPCResponse
	if err := c.conn.ReadJSON(&resp); err != nil {
		return err
	}

	if resp.Error != nil {
		return fmt.Errorf("RPC Error %d: %s", resp.Error.Code, resp.Error.Message)
	}

	// 4. Resultat unmarshallen in den Ziel-Pointer
	if resultPtr != nil {
		resultBytes, _ := json.Marshal(resp.Result)
		return json.Unmarshal(resultBytes, resultPtr)
	}

	return nil
}
