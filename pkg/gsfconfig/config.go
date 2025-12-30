package gsfconfig

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config bildet die Struktur unserer config.yaml ab

/*
// ProtocolConfig enthält alle technischen WebSocket-Parameter
type ProtocolConfig struct {
	PongWait     time.Duration `mapstructure:"pong_wait"`
	PingPeriod   time.Duration `mapstructure:"ping_period"`
	MaxBackoff   time.Duration `mapstructure:"max_backoff"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}
*/

type ServerConfig struct {
	Port            string        `mapstructure:"port"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	ShutdownDelay   time.Duration `mapstructure:"shutdown_delay"`
	ReadbufferSize  int           `mapstructure:"readbuffer_size"`
	WritebufferSize int           `mapstructure:"writebuffer_size"`
	Log             LogConfig     `mapstructure:"log"`
	Auth            AuthConfig    `mapstructure:"auth"`
}

type ClientConfig struct {
	Url            string        `mapstructure:"url"`
	PongWait       time.Duration `mapstructure:"pong_wait"`
	WriteDeadline  time.Duration `mapstructure:"write_deadline"`
	MaxMessageSize time.Duration `mapstructure:"max_messagesize"`
	MaxBackoff     time.Duration `mapstructure:"max_backoff"`
	CtxTimeout     time.Duration `mapstructure:"ctx_timeout"`
	Log            LogConfig     `mapstructure:"log"`
	Auth           AuthConfig    `mapstructure:"auth"`
}

// AuthConfig hält die Sicherheits-Parameter
type AuthConfig struct {
	User   string `mapstructure:"user"`
	Secret string `mapstructure:"secret"`
}

// LogConfig hält die Logger-Parameter
type LogConfig struct {
	Level   string `mapstructure:"level"` // z.B. "debug", "info"
	LogFile string `mapstructure:"log_file"`
}

// Config ist das Haupt-Struct, das alles bündelt
type Config struct {
	Server ServerConfig `mapstructure:"server"`
	Client ClientConfig `mapstructure:"client"`
	//Protocol ProtocolConfig `mapstructure:"protocol"`
	//Auth     AuthConfig     `mapstructure:"auth"`
	//Log      LogConfig      `mapstructure:"log"`
}

// Load liest die Config aus Dateien und Environment-Variablen
func Load() (*Config, error) {
	v := viper.New()

	// 1. No Defaults-Setting: we initialize default in Module-Options

	// 2A. Config-Datei suchen
	v.SetConfigName("config")   // Name der Datei (ohne Endung)
	v.SetConfigType("yaml")     // Endung (yaml, json, toml...)
	v.AddConfigPath(".")        // Suche im aktuellen Ordner
	v.AddConfigPath("./config") // Suche im Unterordner config/

	// 2B. Für Entwicklung (wenn man "go run" aus cmd/server startet)
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
