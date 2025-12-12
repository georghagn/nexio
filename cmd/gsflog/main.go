// Copyright 2025 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/georghagn/gsf-go/pkg/gsflog"
	"github.com/georghagn/gsf-go/pkg/rotate"
)

func main() {
	// 1. Setup File Logger (JSON, Rotiert)
	rotator := rotate.New("prod.log", nil, nil, nil)
	defer rotator.Close()

	fileLog := gsflog.NewJSON(rotator, gsflog.LevelInfo)

	// 2. Setup Console Logger (Bunt, Text)
	consoleLog := gsflog.NewConsole(gsflog.LevelDebug)

	fmt.Println("--- GSF Logger 2.0 Demo ---")

	// Szenario: Ein Request kommt rein
	requestID := "req-abc-123"

	// Wir erstellen Logger, die diesen Kontext kennen
	// "With" erstellt eine leichte Kopie.
	fLog := fileLog.With("req_id", requestID)
	cLog := consoleLog.With("req_id", requestID)

	// Loggen
	cLog.Info("Verarbeite Request...")
	fLog.Info("Processing Request") // JSON ins File

	// ... Irgendwo tief im Code ...
	userID := 42

	// Kontext erweitern
	cLogUser := cLog.With("user_id", userID)
	fLogUser := fLog.With("user_id", userID)

	cLogUser.Warn("User hat kein Guthaben mehr!")
	// Konsole: 2023/... [WARN] User hat kein Guthaben mehr! req_id=req-abc-123 user_id=42 (in Gelb)

	fLogUser.Error("Transaction failed")
	// File: {"level":"ERROR","msg":"Transaction failed","req_id":"req-abc-123","time":"...","user_id":42}

	fmt.Println("Check 'prod.log' for JSON output!")
}
