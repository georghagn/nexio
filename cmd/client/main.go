package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

// Typen für Requests/Responses (lokal definiert für den Client)
type RPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      string      `json:"id"`
}

type RPCResponse struct {
	Result json.RawMessage `json:"result"`
	Error  interface{}     `json:"error"`
	ID     string          `json:"id"`
}

func main() {
	// 1. Verbindung aufbauen
	url := "ws://localhost:8080/ws"
	fmt.Printf("Verbinde zu %s ...\n", url)

	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("Dial Error:", err)
	}
	defer c.Close()

	// 2. Read-Loop (Wichtig für Ping/Pong und um Antworten zu sehen)
	done := make(chan struct{})

	// Channel, um das empfangene Token an den Main-Thread zu geben
	tokenChan := make(chan string)

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("Read Error:", err)
				return
			}
			fmt.Printf("SERVER ANTWORT: %s\n", message)

			// Versuchen wir, ein Token aus der Antwort zu fischen (quick & dirty Parsing)
			var resp RPCResponse
			json.Unmarshal(message, &resp)

			// Wenn wir ein Result haben, schauen wir ob ein "token" drin ist
			if len(resp.Result) > 0 {
				var resMap map[string]string
				if err := json.Unmarshal(resp.Result, &resMap); err == nil {
					if token, ok := resMap["token"]; ok {
						// Token gefunden! Senden an Main
						select {
						case tokenChan <- token:
						default:
						}
					}
				}
			}
		}
	}()

	// 3. Szenario: Login
	fmt.Println("\n--- Schritt 1: Login versenden ---")
	loginReq := RPCRequest{
		JSONRPC: "2.0",
		Method:  "auth.login",
		Params:  map[string]string{"username": "admin", "password": "secret"},
		ID:      "1",
	}
	c.WriteJSON(loginReq)

	// 4. Warten auf Token
	var savedToken string
	select {
	case t := <-tokenChan:
		savedToken = t
		fmt.Printf("\n>>> TOKEN ERHALTEN: %s <<<\n", savedToken)
	case <-time.After(2 * time.Second):
		fmt.Println("!!! Timeout: Kein Token erhalten !!!")
		return
	}

	// Kurze Pause
	time.Sleep(1 * time.Second)

	// 5. Szenario: Resume (mit dem erhaltenen Token)
	fmt.Println("\n--- Schritt 2: Session Resume testen ---")
	resumeReq := RPCRequest{
		JSONRPC: "2.0",
		Method:  "auth.resume",
		Params:  map[string]string{"token": savedToken},
		ID:      "2",
	}
	c.WriteJSON(resumeReq)

	// 6. Warten auf Abbruch (Ctrl+C)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
	fmt.Println("Client beendet.")
}
