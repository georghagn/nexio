Verstanden. Eine spezifische Dokumentation für `pkg/gsflog` ist sehr sinnvoll, da das Modul durch die Einführung von `ReopenableWriter`, Formattern und Context (`With`) sehr mächtig geworden ist.

Hier ist der Entwurf für **`pkg/gsflog/README.md`**. Er konzentriert sich besonders auf die Unterscheidung der Rotations-Strategien, da das oft der Punkt ist, an dem Missverständnisse entstehen.

-----

# gsflog - Go Small Framework Logger

**gsflog** ist ein strukturierter, modularer Logger für Go, entwickelt nach der "Tiny Frameworks" Philosophie. Er trennt strikt zwischen Datenerfassung (`Logger`), Formatierung (`Formatter`) und Ausgabe (`io.Writer`).

## 🌟 Features

  * **Structured Logging:** Keine String-Verkettung mehr. Nutze Key-Value Paare (`With("user_id", 42)`).
  * **Formatters:**
      * `TextFormatter`: Bunte Ausgabe für die Konsole (Dev-Mode).
      * `JSONFormatter`: Maschinenlesbares JSON für Produktion (ELK, Splunk, CloudWatch).
  * **Rotation Strategies:** Unterstützt sowohl **interne** (automatische) als auch **externe** (Signal/Scheduler-basierte) Datei-Rotation.
  * **Thread-Safe:** Sicherer Zugriff aus beliebig vielen Goroutinen.

## 🚀 Quick Start

```go
package main

import "github.com/DEIN_USER/gsf-go/pkg/gsflog"

func main() {
    // 1. Einfacher Konsolen-Logger (mit Farben)
    log := gsflog.NewConsole(gsflog.LevelDebug)

    // 2. Kontext hinzufügen
    reqLog := log.With("request_id", "req-123")

    reqLog.Info("Server gestartet")
    reqLog.Warn("Speicher wird knapp")
}
```

## ⚙️ Strategien für Datei-Logs & Rotation

`gsflog` bietet zwei Wege, um Log-Dateien zu verwalten. Wähle den, der zu deiner Infrastruktur passt.

### A. Interne Rotation (Empfohlen / Standalone)

Der Logger kümmert sich selbstständig um die Rotation. Du musst keine externen Tools konfigurieren.
Hierbei nutzen wir das `pkg/rotate` Paket als Backend.

**Vorteil:** "Set and Forget". Funktioniert überall (Docker, Bare Metal, Windows/Linux).

```go
import (
    "github.com/DEIN_USER/gsf-go/pkg/gsflog"
    "github.com/DEIN_USER/gsf-go/pkg/rotate"
)

func main() {
    // Der Rotator verwaltet die Datei-Größe
    rotator := rotate.New("app.log", 
        &rotate.SizePolicy{MaxBytes: 10*1024*1024}, // 10 MB
        &rotate.GzipCompression{}, 
        nil,
    )
    defer rotator.Close()

    // Der Logger schreibt einfach in den Rotator
    log := gsflog.NewJSON(rotator, gsflog.LevelInfo)
    
    log.Info("Dies landet in einer rotierenden Datei")
}
```

### B. Externe Rotation (Linux Way / Scheduler)

Der Logger schreibt in eine Datei, aber ein **externer Prozess** (z.B. `logrotate`, ein Kubernetes Sidecar oder der GSF Scheduler) verschiebt die Datei. Der Logger muss danach angewiesen werden, die Datei neu zu öffnen.

**Vorteil:** Integration in System-Tools oder zeitgesteuerte Rotation (Cron).

```go
import (
    "github.com/DEIN_USER/gsf-go/pkg/gsflog"
    "github.com/DEIN_USER/gsf-go/pkg/schedule"
)

func main() {
    // 1. Nutze den ReopenableWriter
    writer, _ := gsflog.NewReopenableWriter("app.log")
    defer writer.Close()

    log := gsflog.NewJSON(writer, gsflog.LevelInfo)

    // 2. Ein externer Trigger (hier simuliert durch Scheduler) rotiert
    sched := schedule.New()
    sched.Every(24*time.Hour, func() {
        // A. Datei umbenennen (Simulation von logrotate)
        os.Rename("app.log", "app.log.bak")
        
        // B. WICHTIG: Dem Logger sagen, er soll neu öffnen
        writer.Reopen() 
    })
}
```

## 🎨 Formatierung

### JSON (Production)

Ideal für Log-Aggregatoren.

```go
log := gsflog.NewJSON(os.Stdout, gsflog.LevelInfo)
log.With("id", 1).Error("Fail")
// Output: {"level":"ERROR","msg":"Fail","id":1,"time":"2023-..."}
```

### Text / Konsole (Development)

Menschenlesbar, sortierte Felder, optionale ANSI-Farben.

```go
log := gsflog.NewConsole(gsflog.LevelDebug)
log.With("id", 1).Error("Fail")
// Output: 2023/... [ERROR] Fail id=1 (in Rot)
```

## 🛠 API Referenz

### Logger erstellen

  * `New(out io.Writer, level Level, fmt Formatter)`: Der generische Konstruktor.
  * `NewConsole(level Level)`: Shortcut für Stdout + TextFormatter + Farben.
  * `NewJSON(out io.Writer, level Level)`: Shortcut für JSONFormatter.

### Kontext (Fluent Interface)

  * `With(key string, value interface{}) *Logger`: Erstellt eine **Kopie** des Loggers mit dem neuen Feld. Der ursprüngliche Logger bleibt unverändert (Immutability).

### Writer

  * `NewReopenableWriter(path string)`: Erstellt einen Writer, der `Reopen()` unterstützt. Thread-safe.
