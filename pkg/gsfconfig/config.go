package gsfconfig

import (
	"strings"

	"github.com/spf13/viper"
)

// Config bildet die Struktur unserer config.yaml ab
type Config struct {
	Server ServerConfig `mapstructure:"server"`
	Log    LogConfig    `mapstructure:"log"`
	Auth   AuthConfig   `mapstructure:"auth"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`  // z.B. "debug", "info"
	Format string `mapstructure:"format"` // z.B. "json", "text"
}

type AuthConfig struct {
	Secret string `mapstructure:"secret"` // Für Token Generierung
}

// Load liest die Config aus Dateien und Environment-Variablen
func Load() (*Config, error) {
	v := viper.New()

	// 1. Defaults setzen (Falls keine Config-Datei da ist)
	v.SetDefault("server.port", ":8080")
	v.SetDefault("server.host", "localhost")
	v.SetDefault("log.level", "info")
	v.SetDefault("auth.secret", "change-me-in-prod")

	// 2. Config-Datei suchen
	v.SetConfigName("config")   // Name der Datei (ohne Endung)
	v.SetConfigType("yaml")     // Endung (yaml, json, toml...)
	v.AddConfigPath(".")        // Suche im aktuellen Ordner
	v.AddConfigPath("./config") // Suche im Unterordner config/

	// B. Für Entwicklung (wenn man "go run" aus cmd/server startet)
	// Wir suchen 2 Ebenen höher im Project Root
	v.AddConfigPath("../..")

	// 3. Environment Variablen lesen (Automatisch)
	// Macht aus "server.port" -> "GSF_SERVER_PORT"
	v.SetEnvPrefix("GSF")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 4. Einlesen
	if err := v.ReadInConfig(); err != nil {
		// Es ist okay, wenn keine Config-Datei da ist (wir haben Defaults),
		// aber Syntax-Fehler wollen wir wissen.
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	// 5. In Struct unmarshaln
	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return nil, err
	}

	return &c, nil
}
