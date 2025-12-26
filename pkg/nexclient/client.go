package nexIOclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/georghagn/gsf-suite/pkg/gsflog"
	"github.com/georghagn/gsf-suite/pkg/nexproto"
	"github.com/gorilla/websocket"
)

// Client repräsentiert eine Verbindung zum GSF Server
type Client struct {
	conn *websocket.Conn
	//url  string      // Damit wir wissen, wohin wir uns wiederverbinden
	auth interface{} // Damit wir uns automatisch neu einloggen können

	seq       uint64                                  // Atomic Counter für eindeutige Request IDs
	pending   map[string]chan *nexIOproto.RPCResponse // Hier warten wir auf Antworten: Map[ID] -> Antwort-Kanal
	pendingMu sync.Mutex

	OnNotification func(method string, params json.RawMessage) // Optional: Callback für Broadcasts/Notifications
	OnStatusChange func(connected bool)

	Log     gsflog.LogSink
	Options *Options

	done      chan struct{}
	closeOnce sync.Once  // Zum Beenden
	connected bool       // Interner Status
	mu        sync.Mutex // Schützt den Zugriff auf c.conn während Reconnects
}

func New(logger gsflog.LogSink, o *Options) *Client {
	//func New(logger gsflog.LogSink, opts ...func(*Options)) *Client {

	// Default-Optionen
	opts := defaultOptions()
	if o != nil {

		// Update with user-Options
		if o.PongWait > 0 {
			opts.PongWait = o.PongWait
		}
		if o.PingPeriod > 0 {
			opts.PingPeriod = o.PingPeriod
		}
		if o.MaxBackoff > 0 {
			opts.MaxBackoff = o.MaxBackoff
		}
		if o.WriteTimeout > 0 {
			opts.WriteTimeout = o.WriteTimeout
		}
		if o.Logger.LogFile != "" {
			opts.Logger.LogFile = o.Logger.LogFile
		}
		if o.Logger.LogLevel != "" {
			opts.Logger.LogLevel = o.Logger.LogLevel
		}
		if o.Logger.LogFormat != "" {
			opts.Logger.LogFormat = o.Logger.LogFormat
		}
		if o.Auth.User != "" {
			opts.Auth.User = o.Auth.User
		}
		if o.Auth.Secret != "" {
			opts.Auth.Secret = o.Auth.Secret
		}
	}
	return &Client{
		Options: opts, // Wir speichern die Optionen im Client-Struct
		pending: make(map[string]chan *nexIOproto.RPCResponse),
		done:    make(chan struct{}),
		Log:     logger.With("component", "nexIOclient"),
	}

}

// Run startet den Client und hält die Verbindung aktiv.
func (c *Client) Run(ctx context.Context, authParams interface{}) {
	//c.url = url
	c.auth = authParams

	backoff := time.Second
	maxBackoff := c.Options.MaxBackoff

	for {
		c.Log.With("url", c.Options.Url).Info("Versuche Verbindung")

		err := c.connectAndAuth()
		if err != nil {
			c.Log.With("error", err).With("backoff", backoff.String()).Error("Verbindung fehlgeschlagen")

			select {
			case <-time.After(backoff):
				// Exponential Backoff: Wir warten jedes Mal etwas länger (bis max 32s)
				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
				continue
			case <-ctx.Done():
				return
			}
		}

		// Wenn wir hier sind, steht die Verbindung!
		backoff = time.Second // Backoff zurücksetzen
		c.Log.Info("Verbindung steht und ist authentifiziert.")

		// Warten bis die Verbindung stirbt oder der Kontext beendet wird
		select {
		case <-c.done:
			c.Log.Info("Verbindung zum Server verloren. Reconnect eingeleitet...")
			// WICHTIG: c.done für den nächsten Versuch zurücksetzen
			c.mu.Lock()
			c.done = make(chan struct{})
			c.closeOnce = sync.Once{} // CloseOnce auch resetten für neue Verbindung
			c.mu.Unlock()
		case <-ctx.Done():
			c.Close()
			return
		}
	}
}

func (c *Client) setupHandlers() {
	c.conn.SetReadDeadline(time.Now().Add(c.Options.PongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(c.Options.PongWait))
		return nil
	})
}

// Interne Hilfsmethode für Dial + Auth
func (c *Client) connectAndAuth() error {
	conn, _, err := websocket.DefaultDialer.Dial(c.Options.Url, nil)
	if err != nil {
		return err
	}
	c.mu.Lock()
	c.conn = conn
	c.connected = true
	c.mu.Unlock()

	// Ping/Read Deadline Handler wieder setzen (wie gehabt)
	c.setupHandlers()

	// readLoop starten
	go c.readLoop()

	// Automatischer Login, falls Auth-Parameter vorhanden sind
	if c.auth != nil {
		ctx, cancel := context.WithTimeout(context.Background(), c.Options.CtxTimeout) //  5*time.Second)
		defer cancel()

		var loginRes interface{} // Hier könnte man ein spezielles Struct nehmen
		err := c.Call(ctx, "auth.login", c.auth, &loginRes)
		if err != nil {
			c.Close()
			return fmt.Errorf("auth failed: %w", err)
		}
	}

	if c.OnStatusChange != nil {
		c.OnStatusChange(true)
	}

	return nil
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
		c.mu.Lock()
		wasConnected := c.connected
		c.connected = false
		if c.conn != nil {
			c.conn.Close()
		}
		c.mu.Unlock()
		close(c.done)

		if wasConnected && c.OnStatusChange != nil {
			c.OnStatusChange(false)
		}
	})
}

func (c *Client) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connected
}

// readLoop läuft im Hintergrund und verteilt eingehende Nachrichten
func (c *Client) readLoop() {
	defer c.Close()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			c.Log.With("Error", err).Error("CLIENT READ ERROR")
			return
		}

		// DEBUG: Wir loggen ALLES, was vom Server reinkommt
		c.Log.With("Message", string(message)).Debug("DEBUG: Roh-Daten vom Server erhalten")

		// Erster Versuch: Ist es eine Response?
		var resp nexIOproto.RPCResponse

		err = json.Unmarshal(message, &resp)
		if err == nil && resp.ID != nil {
			// Es ist eine Antwort auf einen Call
			var idStr string
			json.Unmarshal(*resp.ID, &idStr)
			c.pendingMu.Lock()
			ch, found := c.pending[idStr]
			c.pendingMu.Unlock()
			if found {
				ch <- &resp
			}
		} else {
			// Wenn es keine ID hat ODER der erste Parse fehlgeschlagen ist
			// Versuchen wir es als Notification
			var notif nexIOproto.RPCNotification
			if errNotif := json.Unmarshal(message, &notif); errNotif == nil {
				if c.OnNotification != nil {
					go c.OnNotification(notif.Method, notif.Params)
				}
			} else {
				// Nur wenn BEIDES fehlschlägt, loggen wir den Fehler
				c.Log.With("errNotif", errNotif).Debug("DEBUG: Weder Response noch Notification")
			}
		}
	}
}
