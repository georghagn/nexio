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
	// 1. Context for clean shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// initialize provider with logger
	plogger := gsflog.NewDefaultConsole()
	plogger.SetLevel(gsflog.LevelInfo)
	pBaseL := adapter.Wrap(plogger.With("order", "Bee_#38"))

	provider := transport.NewWSProvider(pBaseL)

	// 2. initialize infrastructure
	addr := "ws://localhost:8080/ws"

	// 3. Create the node (as a client, therefore `conn=nil` at the beginning)
	// We pass the provider and the target address for the reconnect
	node := rpc.NewNode(nil, provider, addr, pBaseL)

	// 4. Register handler (What should happen if the payment service calls us?)
	node.Register("order.update", func(ctx context.Context, params json.RawMessage) (any, error) {
		pBaseL.With("params", string(params)).Info("[Order] Received status update from partner")
		return "OK", nil
	})

	// 5. The active logic (e.g. requesting a payment every 10 seconds)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(10 * time.Second):
				pBaseL.Info("[Order] Attempting payment for order #Bee_#38")
				// Use Call for Request-Response
				res, err := node.Call(ctx, "payment.process", "Bee_#38")
				if err != nil {
					pBaseL.With("Error", err).Error("[Order] Error")
					continue
				}
				pBaseL.With("result", string(res)).Info("[Order] Confirmation received")
			}
		}
	}()

	// 6. Start the node (blocks and handles reconnects)
	pBaseL.With("addr", addr).Info("[Order] Service started.")
	if err := node.Listen(ctx); err != nil {
		pBaseL.With("Error", err).Error("[Order] Closed")
	}
}
