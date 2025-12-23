package main

import (
	"context"
	"log"
	"time"

	"github.com/georghagn/gsf-suite/pkg/nexclient"
)

// Response-Structs definieren (Typsicherheit!)
type LoginResult struct {
	Status string `json:"status"`
	Token  string `json:"token"`
}

func main() {
	// 1. Verbindung herstellen
	client, err := nexclient.Dial("ws://localhost:9090/ws")
	if err != nil {
		log.Fatal("Konnte Printer nicht verbinden:", err)
	}
	defer client.Close()

	log.Println("Printer Service verbunden.")

	// 2. Login durchführen (Synchron!)
	// Wir geben dem Login max 5 Sekunden
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	params := map[string]string{
		"username": "admin",
		"password": "default-secret",
	}

	var loginRes LoginResult

	log.Println("Versuche Login...")
	// HIER PASSIERT DIE MAGIE:
	if err := client.Call(ctx, "auth.login", params, &loginRes); err != nil {
		log.Fatalf("Login fehlgeschlagen: %v", err)
	}

	log.Printf("Login erfolgreich! Token: %s", loginRes.Token)

	// 3. Service Loop
	// Hier würde der Printer jetzt einfach warten oder ab und zu seinen Status senden
	for {
		time.Sleep(10 * time.Second)
		// Ping / KeepAlive Logik oder Status Update senden...
		// client.Call(context.Background(), "printer.status", "idle", nil)
		log.Println("Printer wartet auf Aufträge...")

		// Service Loop (Schlauer gemacht)
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Alle 10 Sekunden:
				log.Println("Printer wartet auf Aufträge...")
				// Optional: Ping senden um sicherzugehen
				client.Call(ctx, "printer.ping", nil, nil)

			case <-client.Done():
				// HIER merken wir den Tod des readLoops!
				log.Println("Service beendet: Verbindung zum Server verloren.")
				return // Beendet das Programm
			}
		}
	}
}
