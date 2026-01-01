// Copyright 2026 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

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

	// provider incl. Log initialize
	// for demonstration purposes, equip the provider and node with different log levels.
	plogger := gsflog.NewDefaultConsole()
	plogger.SetLevel(gsflog.LevelInfo)
	pBaseL := adapter.Wrap(plogger.With("module", "payment-service"))

	provider := transport.NewWSProvider(pBaseL)

	// the channel through which we receive new connections
	found := make(chan transport.Connection)

	// 1. open the port and wait for "knocking" (server role)
	go func() {
		pBaseL.Info("[Payment] Open port :8080 and wait for order-service...")
		if err := provider.Listen(ctx, ":8080", found); err != nil {
			pBaseL.With("Error", err).Error("[Payment] Server-Error")
		}
	}()

	// 2. dispatcher-Loop: What to do when a connection comes in?
	for {
		select {
		case conn := <-found:
			pBaseL.Info("[Payment] Order service has connected!")
			// we create the peer with the active connection.
			// dialAddr is empty because the server doesn't need to reconnect to this client.
			go handleClient(ctx, conn, pBaseL)

		case <-ctx.Done():
			pBaseL.Info("[Payment] Close service...")
			return
		}
	}
}

func handleClient(ctx context.Context, conn transport.Connection, logger transport.LogSink) {
	// Provider and empty address, because this peer was passively created.
	node := rpc.NewNode(conn, nil, "", logger)

	// 3. register your handler: What information can others access from us?
	node.Register("payment.process", func(ctx context.Context, p json.RawMessage) (any, error) {
		var orderID string
		json.Unmarshal(p, &orderID)
		node.Log.With("OrderID", orderID).Info("[Payment] 💳 Process payment")

		// simuliere success
		return "Payment_Success_ID_9988", nil
	})

	// 4. own logic: We can also actively notify the order service.
	go func() {
		// we'll wait a moment and then send a confirmation (notify).
		logger.Info("[Payment] Send status update to order service...")
		err := node.Notify(ctx, "order.update", "Payment recorded")
		if err != nil {
			logger.With("Error", err).Error("[Payment] Notify Error")
		}
	}()

	// 5. keeping the node alive
	if err := node.Listen(ctx); err != nil {
		logger.With("Error", err).Error("[Payment] Connection to client lost")
	}
}
