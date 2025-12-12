
# GSF (Go Small Frameworks) - Suite

**GSF-Suite** ist eine Suite von kleinen, spezialisierten und modularen Go-Paketen, inspiriert von der "Tiny Smalltalk Frameworks" Philosophie.
Das Ziel: **Maximale Funktionalität bei minimalen Abhängigkeiten.** Wir nutzen fast ausschließlich die Go Standardbibliothek.

## Philosophie

  * **Zero Dependencies:** Keine aufgeblähten externen Libraries.
  * **Idiomatic Go:** Nutzung von Interfaces (`io.Writer`, `context.Context`), Goroutines und Channels.
  * **Modular:** Jedes Paket (`pkg/*`) kann unabhängig voneinander genutzt werden.
  * **Robust:** Thread-Safety und Panic-Recovery sind standardmäßig eingebaut.

## Die Module

### 1\. `pkg/rotate` - Der File Rotator

Ein robuster `io.WriteCloser`, der Dateien automatisch rotiert, wenn sie zu groß oder zu alt werden.

  * **Features:** Thread-safe, Strategy Pattern (Rotation, Archive, Retention).
  * **Strategies:** Size-based, Time-based, Gzip Compression, Max Files Retention.
  * **Besonderheit:** Funktioniert als Backend für *jeden* Logger.

```go
w := rotate.New("app.log",
    &rotate.SizePolicy{MaxBytes: 10*1024*1024}, // 10 MB
    &rotate.GzipCompression{},                  // .gz Kompression
    &rotate.MaxFiles{MaxBackups: 5},            // Max 5 Backups
)
defer w.Close()
w.Write([]byte("Log Entry..."))
```

### 2\. `pkg/gsflog` - Der Logger

Ein strukturierter Logger mit Unterstützung für JSON, Farben und Kontext-Feldern.

  * **Features:** Structured Logging (JSON/Text), Log-Levels, `With(key, val)` Kontext, Color-Support.
  * **Modi:**
      * **Inline Rotation:** Nutzt `pkg/rotate` direkt.
      * **External Rotation:** Nutzt `ReopenableWriter` für externe Tools (logrotate/Scheduler).

```go
// JSON Output + Rotation
log := gsflog.NewJSON(rotator, gsflog.LevelInfo)

// Kontext hinzufügen (Fluent Interface)
reqLog := log.With("request_id", "123").With("user_id", 42)
reqLog.Info("Processing started") 
// Output: {"level":"INFO","msg":"Processing started","request_id":"123","user_id":42,...}
```

### 3\. `pkg/schedule` - Der Scheduler

Ein Ticker-basierter Task Runner für Hintergrundaufgaben.

  * **Features:** One-Shot (`At`) & Interval (`Every`), Panic Recovery (kein Server-Crash bei Job-Fehlern), Graceful Shutdown, Introspection (`List`).

```go
sched := schedule.New()

// Job starten
id := sched.Every(5*time.Minute, func() {
    fmt.Println("DB Cleanup...")
})

// Job stoppen
sched.Cancel(id)
```

### 4\. `pkg/nexio` - JSON-RPC 2.0 Server

Ein flexibler RPC-Server, der Go-Methoden via Reflection automatisch verfügbar macht.

  * **Features:** JSON-RPC 2.0 Spec, HTTP & WebSocket Support, Reflection-based Service Registry.
  * **Highlight:** Trennung von Transport (HTTP/WS) und Logik.

```go
type MyService struct{}
func (s *MyService) Echo(args EchoArgs) (string, error) {
    return args.Text, nil
}

// ...
server := nexio.New()
server.RegisterService(&MyService{}) // Exposes "MyService.Echo"
http.Handle("/rpc", server)
```

## Integration: Die Suite

Hier sehen wir, wie alle Module zu einer robusten Anwendung verschmelzen:

```go
func main() {
    // 1. Logging mit Rotation
    rotator := rotate.New("server.log", nil, nil, nil)
    log := gsflog.NewJSON(rotator, gsflog.LevelInfo)
    
    // 2. Scheduler
    sched := schedule.New()
    sched.Every(1*time.Hour, func() { log.Info("System Check OK") })

    // 3. RPC Service
    server := nexio.New()
    // ... Register Services ...

    // 4. Start
    log.Info("GSF App startet auf :8080")
    http.ListenAndServe(":8080", nil)
}
```

## Testing

Jedes Modul verfügt über eine eigene Test-Suite inklusive Race-Detection.

```bash
# Alle Tests ausführen
go test ./pkg/... -v

# Race Conditions prüfen (Wichtig für Concurrency!)
go test ./pkg/... -race
```

## Lizenz

Apache 2.0 License - Feel free to use and modify.

