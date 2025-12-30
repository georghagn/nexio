package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/georghagn/gsf-suite/pkg/gsfconfig"
	"github.com/georghagn/gsf-suite/pkg/gsflog"
	"github.com/georghagn/gsf-suite/pkg/nexclient"
)

// Response-Structs definieren (Typsicherheit!)
type LoginResult struct {
	Status string `json:"status"`
	Token  string `json:"token"`
}

func main() {

	// Config aus YAML laden
	cfg, err := gsfconfig.Load()
	if err != nil {
		fmt.Println("Fehler beim Laden der Config: %v", err)
		os.Exit(1)
	}

	// Brücke bauen: gsfconfig -> nexIOclient.ProtocolSettings
	defaultLogger := nexIOclient.LogOptions{
		LogFile:  cfg.Client.Log.LogFile,
		LogLevel: cfg.Client.Log.Level,
	}

	defaultAuth := nexIOclient.AuthOptions{
		User:   cfg.Client.Auth.User,
		Secret: cfg.Client.Auth.Secret,
	}

	clientSettings := nexIOclient.ClientSettings{
		Url:          cfg.Client.Url,
		PongWait:     cfg.Client.PongWait,
		MaxBackoff:   cfg.Client.MaxBackoff,
		WriteTimeout: cfg.Client.WriteDeadline,
		CtxTimeout:   cfg.Client.CtxTimeout,
		Logger:       defaultLogger,
		Auth:         defaultAuth,
	}

	loglevel := gsflog.LevelInfo
	logger := gsflog.NewDefault(&clientSettings.Logger.LogFile)
	switch strings.ToUpper(clientSettings.Logger.LogLevel) {
	case "DEBUG":
		loglevel = gsflog.LevelDebug
	case "INFO":
		loglevel = gsflog.LevelInfo
	case "WARN":
		loglevel = gsflog.LevelWarn
	case "ERROR":
		loglevel = gsflog.LevelError
	default:
		loglevel = gsflog.LevelInfo
	}
	logger.SetLevel(loglevel)

	// Client erstellen mit dem Bulk-Update
	client := nexIOclient.New(
		logger.With("component", "printer-service"),
		nexIOclient.WithClientSettings(clientSettings),
	)

	// 2. Business-Logik registrieren (Callbacks)
	// Diese bleiben über alle Reconnects hinweg aktiv!
	client.OnNotification = func(method string, params json.RawMessage) {
		if method == "printer.print" {
			logger.With("params", string(params)).Info("!!! DRUCKAUFTRAG ERHALTEN")
		}
	}

	client.OnStatusChange = func(connected bool) {
		if connected {
			logger.Info("Online: Drucker ist bereit.")
		} else {
			logger.Info("Offline: Warte auf Wiederverbindung...")
		}
	}

	// 3. Authentifizierungs-Daten vorbereiten
	// Das wird bei jedem (Wieder-)Verbinden automatisch mitgeschickt

	authParams := map[string]string{
		"username": "admin",
		"password": "admin-01",
	}

	/*
		authParams := map[string]string{
			"username": client.Options.Auth.User,
			"password": client.Options.Auth.Secret,
		}
	*/
	// 4. Den Client starten
	// Run blockiert, solange der Kontext aktiv ist.
	logger.Info("Printer Service startet...")
	ctx := context.Background()

	// Hier gibst du URL und Auth mit. Den Rest erledigt die Library.
	client.Run(ctx, authParams)

}
