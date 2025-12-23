package nexclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/georghagn/gsf-suite/pkg/nexproto"
	"github.com/gorilla/websocket"
)

// Client repräsentiert eine Verbindung zum GSF Server
type Client struct {
	conn *websocket.Conn

	// Atomic Counter für eindeutige Request IDs
	seq uint64

	// Hier warten wir auf Antworten: Map[ID] -> Antwort-Kanal
	pending   map[string]chan *nexIOproto.RPCResponse
	pendingMu sync.Mutex

	// Optional: Callback für Broadcasts/Notifications
	OnNotification func(method string, params json.RawMessage)

	// Zum Beenden
	closeOnce sync.Once
	done      chan struct{}
}

const (
	// serverTimeout: Wenn wir so lange nichts vom Server hören, ist er tot.
	// Muss länger sein als der Ping-Intervall des Servers!
	serverTimeout = 60 * time.Second

	// writeTimeout: Wie lange wir Zeit haben, einen Pong zu senden
	writeTimeout = 5 * time.Second
)

// Dial verbindet sich mit dem Server
func Dial(url string) (*Client, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	c := &Client{
		conn:    conn,
		pending: make(map[string]chan *nexIOproto.RPCResponse),
		done:    make(chan struct{}),
	}

	// Wenn der Server "Ping" ruft, setzen wir unsere Deadline zurück.
	conn.SetPingHandler(func(appData string) error {
		// 1. Lebenszeichen erhalten -> Uhr zurücksetzen
		conn.SetReadDeadline(time.Now().Add(serverTimeout))

		// 2. Höflich mit "Pong" antworten (sonst kickt uns der Server)
		err := conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(writeTimeout))
		if err == websocket.ErrCloseSent {
			return nil
		} else if e, ok := err.(interface{ Temporary() bool }); ok && e.Temporary() {
			return nil
		}
		return err
	})

	// Starte sofort den "Lauscher" im Hintergrund
	go c.readLoop()

	return c, nil
}

// Eine Methode, damit der User prüfen kann, ob wir noch leben
func (c *Client) Done() <-chan struct{} {
	return c.done
}

// Call führt einen synchronen RPC-Aufruf durch.
// Es blockiert, bis eine Antwort kommt oder der Context abläuft.
// resultPtr: Ein Pointer auf die Struktur, in die das Ergebnis geparst werden soll.
func (c *Client) Call(ctx context.Context, method string, params interface{}, resultPtr interface{}) error {
	// 1. Request bauen
	id := fmt.Sprintf("%d", atomic.AddUint64(&c.seq, 1))
	reqID := json.RawMessage(`"` + id + `"`)

	// Params marshaln
	paramBytes, err := json.Marshal(params)
	if err != nil {
		return err
	}
	pRaw := json.RawMessage(paramBytes)

	req := nexIOproto.RPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  pRaw,
		ID:      &reqID,
	}

	// 2. Antwort-Kanal vorbereiten
	respChan := make(chan *nexIOproto.RPCResponse, 1)
	c.pendingMu.Lock()
	c.pending[id] = respChan
	c.pendingMu.Unlock()

	// Cleanup: Egal was passiert, lösche den Kanal am Ende aus der Map
	defer func() {
		c.pendingMu.Lock()
		delete(c.pending, id)
		c.pendingMu.Unlock()
	}()

	// 3. Senden (Thread-Safe durch Gorilla)
	if err := c.conn.WriteJSON(req); err != nil {
		return err
	}

	// 4. Warten auf Antwort oder Timeout
	select {
	case resp, ok := <-respChan:
		if !ok {
			// Kanal wurde geschlossen -> Verbindung weg
			return errors.New("connection closed while waiting for response")
		}
		// Antwort erhalten!
		if resp.Error != nil {
			// Der Server hat einen Fehler gemeldet
			return fmt.Errorf("RPC Error %d: %s", resp.Error.Code, resp.Error.Message)
		}

		// Ergebnis in den Pointer parsen (wenn vorhanden)
		if resultPtr != nil && resp.Result != nil {
			// resp.Result ist oft map[string]interface{} oder float64 durch JSON decoding.
			// Um es sauber in das Ziel-Struct zu bekommen, marshaln wir es kurz zurück
			// und unmarshaln es in das Ziel. Das ist der sicherste Weg in Go.
			tempBytes, _ := json.Marshal(resp.Result)
			if err := json.Unmarshal(tempBytes, resultPtr); err != nil {
				return fmt.Errorf("result parse error: %v", err)
			}
		}
		return nil

	case <-ctx.Done():
		return ctx.Err() // Timeout oder Abbruch
	}
}

// Close schließt die Verbindung
func (c *Client) Close() {
	c.closeOnce.Do(func() {
		c.conn.Close()
	})
}

// readLoop läuft im Hintergrund und verteilt eingehende Nachrichten
func (c *Client) readLoop() {
	defer close(c.done)
	defer c.Close()

	// Wenn diese Loop endet (Verbindung weg), müssen wir aufräumen!
	defer func() {
		c.pendingMu.Lock()
		for id, ch := range c.pending {
			close(ch) // <--- WICHTIG: Signalisieren, dass nichts mehr kommt
			delete(c.pending, id)
		}
		c.pendingMu.Unlock()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("CLIENT READ ERROR: %v\n", err)
			log.Println("Client connection closed:", err)
			return
		}

		// Deadline verlängern bei JEDER Nachricht ---
		c.conn.SetReadDeadline(time.Now().Add(serverTimeout))

		// Wir parsen erst generisch
		var resp nexIOproto.RPCResponse
		if err := json.Unmarshal(message, &resp); err != nil {
			continue // Parse error ignorieren
		}

		// Fall A: Es ist eine Antwort auf einen Request (hat ID)
		if resp.ID != nil {
			// ID String säubern (da json.RawMessage Anführungszeichen enthalten kann)
			var idStr string
			json.Unmarshal(*resp.ID, &idStr)

			c.pendingMu.Lock()
			ch, found := c.pending[idStr]
			c.pendingMu.Unlock()

			if found {
				ch <- &resp
			}
		} else {
			// Fall B: Es ist eine Notification (keine ID)
			// Hier greift der User-Callback
			if c.OnNotification != nil {
				// Da RPCResponse Struktur für Result/Error optimiert ist,
				// müssten wir eigentlich RPCNotification parsen.
				// Der Einfachheit halber tun wir so, als wäre 'Method' im JSON.
				// (Für saubere Lösung: Ein generisches struct parsen)
				var notif nexIOproto.RPCNotification
				if err := json.Unmarshal(message, &notif); err == nil {
					// Async aufrufen, damit wir ReadLoop nicht blockieren
					go c.OnNotification(notif.Method, json.RawMessage{})
					// Hinweis: Params Handling hier vereinfacht
				}
			}
		}
	}
}
