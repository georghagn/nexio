// Copyright 2026 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/georghagn/nexio/pkg/adapter"
	"github.com/georghagn/nexio/pkg/gsflog"
	"github.com/georghagn/nexio/pkg/node/rpc"
	"github.com/georghagn/nexio/pkg/node/transport"
)

func main() {
	mode := flag.String("mode", "server", "server oder client")
	addr := flag.String("addr", "localhost:8080", "Addresse")
	flag.Parse()

	// Der zentrale Schalter für das ganze Programm
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	found := make(chan transport.Connection)

	// Provider incl. Log initialize
	// zu DemoZwecken provider und node mit verschieden LogLevels ausrüsten
	plogger := gsflog.NewDefaultConsole()
	plogger.SetLevel(gsflog.LevelDebug)
	pBaseL := plogger.With("module", "transport.WSProvider")
	provider := transport.NewWSProvider(adapter.Wrap(pBaseL))

	outFile := "main.log"
	nlogger := gsflog.NewDefault(&outFile)
	nlogger.SetLevel(gsflog.LevelDebug)

	//nBaseL := nlogger.With("module", "rpc.Node")
	nodeLogger := adapter.Wrap(nlogger.With("module", "rpc.Node"))

	if *mode == "server" {
		go func() {
			// Wir übergeben den ctx, damit der Server weiß, wann er stoppen muss
			if err := provider.Listen(ctx, *addr, found); err != nil && err != http.ErrServerClosed {
				log.Printf("Server Fehler: %v", err)
			}
		}()
	} else {
		// Client Modus: Wir starten direkt handleNewPeer,
		// da dieser nun selbst für das erste Dial und Reconnects zuständig ist.
		go handleNewPeer(ctx, nil, provider, "ws://"+*addr+"/ws", nodeLogger)
	}

	// Der Dispatcher-Loop
	for {
		select {
		case conn := <-found:
			go handleNewPeer(ctx, conn, provider, "", nodeLogger)
		case <-ctx.Done():
			log.Println("Alle Prozesse werden beendet...")
			// Wir geben den Goroutinen einen Moment Zeit zum Aufräumen
			time.Sleep(500 * time.Millisecond)
			return
		}
	}
}

func handleNewPeer(
	ctx context.Context,
	conn transport.Connection,
	provider *transport.WSProvider,
	dialAddr string,
	logger transport.LogSink) {

	node := rpc.NewNode(conn, provider, dialAddr, logger)

	node.Register("ping", func(ctx context.Context, p json.RawMessage) (any, error) {
		return "pong", nil
	})
	//node.Register("system.status", handleStatusRequest)

	node.Register("system.echo", func(ctx context.Context, p json.RawMessage) (any, error) {
		return p, nil // Unser alter Bekannter für Tests
	})
	node.Register("payment.confirmed", func(ctx context.Context, p json.RawMessage) (any, error) {
		var paymentID string
		json.Unmarshal(p, &paymentID)
		log.Printf("✅ Order System: Markiere Zahlung %s als erledigt", paymentID)
		return nil, nil // Return wird ignoriert, da Notification
	})

	// AKTIVITÄT: Der Node kann auch selbst aktiv werden!
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
				result, err := node.Call(ctx, "ping", nil)
				if err != nil {
					// Wenn der Fehler "Verbindung wird wiederhergestellt" ist, loggen wir es diskreter
					if strings.Contains(err.Error(), "wiederhergestellt") {
						log.Println("Ping: Warte auf Reconnect...")
					} else {
						log.Printf("Ping fehlgeschlagen: %v", err)
					}
					continue
				}
				log.Printf("Antwort: %s", string(result))
			}
		}
	}()

	// Listen läuft so lange, wie die Verbindung steht ODER der ctx abgebrochen wird
	if err := node.Listen(ctx); err != nil {
		log.Printf("Peer Verbindung beendet: %v", err)
	}

}
