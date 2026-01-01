package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	"github.com/georghagn/gsf-suite/pkg/adapter"
	"github.com/georghagn/gsf-suite/pkg/gsflog"
	"github.com/georghagn/gsf-suite/pkg/node/rpc"
	"github.com/georghagn/gsf-suite/pkg/node/transport"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Provider incl. Log initialize
	// zu DemoZwecken provider und node mit verschieden LogLevels ausrüsten
	plogger := gsflog.NewDefaultConsole()
	plogger.SetLevel(gsflog.LevelInfo)
	pBaseL := adapter.Wrap(plogger.With("module", "payment-service"))

	provider := transport.NewWSProvider(pBaseL)

	// Der Channel, über den wir neue Verbindungen empfangen
	found := make(chan transport.Connection)

	// 1. Den Port öffnen und auf "Anklopfen" warten (Server-Rolle)
	go func() {
		pBaseL.Info("[Payment] Öffne Port :8080 und warte auf Order-Service...")
		if err := provider.Listen(ctx, ":8080", found); err != nil {
			pBaseL.With("Error", err).Error("[Payment] Server-Fehler")
		}
	}()

	// 2. Dispatcher-Loop: Was tun, wenn eine Verbindung reinkommt?
	for {
		select {
		case conn := <-found:
			pBaseL.Info("[Payment] Order-Service hat sich verbunden!")
			// Wir erstellen den Peer mit der aktiven Verbindung.
			// dialAddr ist leer, da der Server auf diesen Client nicht re-connecten muss.
			go handleClient(ctx, conn, pBaseL)

		case <-ctx.Done():
			pBaseL.Info("[Payment] Service wird beendet...")
			return
		}
	}
}

func handleClient(ctx context.Context, conn transport.Connection, logger transport.LogSink) {
	// Provider und leere Addresse, da dieser Peer passiv erzeugt wurde
	node := rpc.NewNode(conn, nil, "", logger)

	// 3. Handler registrieren: Was können andere bei uns aufrufen?
	node.Register("payment.process", func(ctx context.Context, p json.RawMessage) (any, error) {
		var orderID string
		json.Unmarshal(p, &orderID)
		node.Log.With("OrderID", orderID).Info("[Payment] 💳 Verarbeite Zahlung")

		// Simuliere Erfolg
		return "Payment_Success_ID_9988", nil
	})

	// 4. Eigene Logik: Wir können den Order-Service auch aktiv benachrichtigen
	go func() {
		// Wir warten kurz und schicken dann eine Bestätigung (Notify)
		logger.Info("[Payment] Schicke Status-Update an Order-Service...")
		err := node.Notify(ctx, "order.update", "Zahlung verbucht")
		if err != nil {
			logger.With("Error", err).Error("[Payment] Notify Fehler")
		}
	}()

	// 5. Den Peer am Leben erhalten
	if err := node.Listen(ctx); err != nil {
		logger.With("Error", err).Error("[Payment] Verbindung zu Client verloren")
	}
}
