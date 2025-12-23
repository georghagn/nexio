package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/georghagn/gsf-suite/internal/services/auth"
	"github.com/georghagn/gsf-suite/pkg/gsfconfig"
	"github.com/georghagn/gsf-suite/pkg/gsflog"
	"github.com/georghagn/gsf-suite/pkg/nexserver"
)

func main() {
	// A. Framework & Services
	// Logger Teil 1
	logFile := "debug.log"
	logger := gsflog.NewDefault(&logFile)
	logger.SetLevel(gsflog.LevelDebug)

	// Config laden
	cfg, err := gsfconfig.Load()
	if err != nil {
		logger.With("Error", err).Error("Konfigurationsfehler")
		os.Exit(1)
	}

	// Logger Teil 2
	if cfg.Log.Level == "debug" {
		logger.SetLevel(gsflog.LevelDebug)
	} else {
		logger.SetLevel(gsflog.LevelInfo)
	}

	// 2. Server mit Logger starten
	logger.With("Port", cfg.Server.Port).Info("Starte GSF Server mit Port " + cfg.Server.Port)
	hub := nexIOserver.NewServer(logger)
	auth.Configure(cfg.Auth.Secret)
	auth.RegisterRoutes(hub)

	// Hub im Hintergrund starten
	go hub.Run()

	wsHandler := nexIOserver.NewServerWS(hub)

	// 2. HTTP Server Setup (Explizit statt http.HandleFunc + ListenAndServe)
	// Wir brauchen Zugriff auf das *http.Server Objekt für Shutdown()
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", wsHandler.ServeWS)

	srv := &http.Server{
		Addr:    cfg.Server.Port,
		Handler: mux,
	}

	// 3. Server in einer Goroutine starten
	// Damit der Main-Thread weiterlaufen kann zum Warten auf Signale.
	go func() {
		logger.With("Port", cfg.Server.Port).Info("GFS-Suite Server läuft auf Port")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.With("Error", err).Error("HTTP Server Fehler")
			os.Exit(1)
		}
	}()

	// 4. Auf OS-Signale warten (Der "Trap")
	// Wir erstellen einen Channel, der auf Signale horcht
	stop := make(chan os.Signal, 1)

	// Wir registrieren uns für SIGINT (Ctrl+C) und SIGTERM (Docker stop)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Hier blockiert das Programm, bis einer drückt
	<-stop
	logger.Info("\nShutdown Signal empfangen...")

	// 5. Graceful Shutdown einleiten

	// A. Kontext mit Timeout erstellen.
	// Wir geben dem Server 5 Sekunden Zeit, laufende Requests fertig zu machen.
	// Danach wird hart abgebrochen.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// B. HTTP Server stoppen (keine neuen Verbindungen mehr)
	if err := srv.Shutdown(ctx); err != nil {
		logger.With("Error", err).Info("HTTP Shutdown Fehler")
	}

	// C. Websocket Hub stoppen (bestehende Verbindungen schließen)
	hub.Shutdown()

	logger.Info("Server erfolgreich beendet.")
}
