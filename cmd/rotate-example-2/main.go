// Copyright 2025 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"time"

	"github.com/georghagn/nexio/pkg/rotate"
)

func main() {
	fmt.Println("Start GSF Demo...")

	simple()
	profi()
}

func simple() {
	fmt.Println("Start GSF Simple Rotator Demo...")

	// 1. Simpel (Defaults)
	rotator := rotate.New("simple.log", nil, nil, nil)

	// Simuliere Log-Einträge
	for i := 0; i < 10; i++ {
		msg := fmt.Sprintf("Log Entry %d at %v\n", i, time.Now().Format(time.TimeOnly))

		fmt.Print("Schreibe: " + msg)
		_, err := rotator.Write([]byte(msg))
		if err != nil {
			panic(err)
		}

		time.Sleep(1 * time.Second) // Kurz warten, damit Zeitstempel unterschiedlich sind
	}

	rotator.Close()
	fmt.Println("Fertig. Prüfe deinen Ordner auf .log Dateien!")
}

func profi() {
	fmt.Println("Start GSF Profi Demo...")

	// Profi-Konfiguration
	rotator := rotate.New("production.log",
		&rotate.SizePolicy{MaxBytes: 50}, // 5 MB
		&rotate.GzipCompression{},        // Platz sparen
		&rotate.MaxFiles{MaxBackups: 5},  // Nur 5 alte behalten
	)

	// Simuliere Log-Einträge
	for i := 0; i < 10; i++ {
		msg := fmt.Sprintf("Log Entry %d at %v\n", i, time.Now().Format(time.TimeOnly))

		fmt.Print("Schreibe: " + msg)
		_, err := rotator.Write([]byte(msg))
		if err != nil {
			panic(err)
		}

		time.Sleep(1 * time.Second) // Kurz warten, damit Zeitstempel unterschiedlich sind
	}

	rotator.Close()
	fmt.Println("Fertig. Prüfe deinen Ordner auf .log Dateien!")
}
