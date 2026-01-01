package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/georghagn/gsf-suite/pkg/adapter"
	"github.com/georghagn/gsf-suite/pkg/gsflog"
	"github.com/georghagn/gsf-suite/pkg/node/rpc"
	"github.com/georghagn/gsf-suite/pkg/node/transport"
)

func main() {
	// 1. Kontext für sauberes Beenden
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Provider incl. Log initialize
	plogger := gsflog.NewDefaultConsole()
	plogger.SetLevel(gsflog.LevelInfo)
	pBaseL := adapter.Wrap(plogger.With("order", "Bee_#38"))

	provider := transport.NewWSProvider(pBaseL)

	// 2. Infrastruktur initialisieren
	addr := "ws://localhost:8080/ws"

	// 3. Den Node erstellen (als Client, daher conn=nil am Anfang)
	// Wir übergeben den Provider und die Zieladresse für den Reconnect
	node := rpc.NewNode(nil, provider, addr, pBaseL)

	// 4. Handler registrieren (Was soll passieren, wenn der Payment-Service uns anruft?)
	node.Register("order.update", func(ctx context.Context, params json.RawMessage) (any, error) {
		pBaseL.With("params", string(params)).Info("[Order] Status-Update vom Partner erhalten")
		return "OK", nil
	})

	// 5. Die aktive Logik (z.B. alle 10 Sekunden eine Zahlung anfordern)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(10 * time.Second):
				pBaseL.Info("[Order] Versuche Zahlung für Order #Bee_#38")
				// Nutze Call für Request-Response
				res, err := node.Call(ctx, "payment.process", "Bee_#38")
				if err != nil {
					pBaseL.With("Error", err).Error("[Order] Fehler")
					continue
				}
				pBaseL.With("result", string(res)).Info("[Order] Bestätigung erhalten")
			}
		}
	}()

	// 6. Den Node starten (blockiert und kümmert sich um Reconnects)
	pBaseL.With("addr", addr).Info("[Order] Service gestartet. Ziel")
	if err := node.Listen(ctx); err != nil {
		pBaseL.With("Error", err).Error("[Order] Beendet")
	}
}
