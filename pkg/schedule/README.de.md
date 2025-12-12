Natürlich\! Eine README für den Scheduler ist wichtig, da sich hier oft Fragen zur **Concurrency** (Nebenläufigkeit), **Sicherheit** (Panics) und zum **Shutdown-Verhalten** stellen.

Hier ist der Entwurf für **`pkg/schedule/README.md`**.

-----

# gsf/schedule - Go Small Framework Scheduler

**gsf/schedule** ist ein leichtgewichtiger, robuster **In-Process Job Scheduler** für Go.
Er wurde entwickelt, um wiederkehrende Aufgaben oder einmalige Tasks auszuführen, ohne externe Abhängigkeiten (wie Cron-Daemons) zu benötigen.

Im Gegensatz zu einem einfachen `time.Ticker` bietet dieses Paket **Panic Recovery**, **Job-Management (Start/Stop)** und **Graceful Shutdown**.

## 🌟 Features

  * **Simple API:** Intuitive Methoden wie `Every` und `At`.
  * **Panic Recovery:** Wenn ein Job abstürzt (panic), fängt der Scheduler den Fehler ab. Deine Hauptanwendung stürzt nicht ab.
  * **Thread-Safe:** Sicherer Zugriff auf die Job-Liste aus mehreren Goroutinen.
  * **Graceful Shutdown:** `StopAll()` wartet, bis laufende Jobs beendet sind, bevor das Programm beendet wird.
  * **Introspection:** Abfragen von Laufzeit-Statistiken (`NextRun`, `Interval`) via `List()` – ideal für Status-Dashboards oder RPC.

## 🚀 Quick Start

```go
package main

import (
    "fmt"
    "time"
    "github.com/DEIN_USER/gsf-go/pkg/schedule"
)

func main() {
    // 1. Scheduler erstellen
    sched := schedule.New()

    // 2. Job starten (Alle 500ms)
    jobID := sched.Every(500*time.Millisecond, func() {
        fmt.Println("Tick...")
    })

    // 3. Einen One-Shot Job planen (in 2 Sekunden)
    sched.At(time.Now().Add(2*time.Second), func() {
        fmt.Println("Boom! (Einmalig)")
    })

    // Lass es kurz laufen
    time.Sleep(3 * time.Second)

    // 4. Job stoppen
    sched.Cancel(jobID)
    fmt.Println("Ticker gestoppt.")
}
```

## ⚙️ Kern-Konzepte

### Recurring Jobs (`Every`)

Führt eine Funktion in einem festen Intervall aus. Der Task läuft in einer eigenen Goroutine.

```go
id := sched.Every(1*time.Minute, func() {
    // DB Backup Logik
})
```

### One-Shot Jobs (`At`)

Führt eine Funktion einmalig zu einem bestimmten Zeitpunkt aus.

```go
targetTime := time.Now().Add(10 * time.Minute)
sched.At(targetTime, func() {
    // Reminder Email senden
})
```

### Panic Recovery (Crash Protection)

Ein häufiges Problem bei selbstgebauten `go func()` Lösungen: Wenn der Code in der Goroutine "panic-ed", stürzt das **gesamte Programm** ab.
`gsf/schedule` kapselt jeden Job in einer `recover()` Funktion.

```go
sched.Every(1*time.Second, func() {
    panic("Datenbank weg!") // Führt NICHT zum Absturz der App
})
// Output auf Stdout: "SCHEDULER PANIC in Job 1: Datenbank weg!"
// Der Scheduler und andere Jobs laufen weiter.
```

## 🛠 Management & Introspection

### Jobs stoppen

Jeder Aufruf von `Every` oder `At` gibt eine `JobID` (int64) zurück. Damit kann der Job gezielt abgebrochen werden.

```go
err := sched.Cancel(jobID)
if err != nil {
    log.Println("Job wurde bereits beendet oder nicht gefunden")
}
```

### Jobs auflisten (`List`)

Das `List()` Feature ist besonders mächtig, um via RPC (z.B. mit `pkg/nexio`) oder in einem Admin-Panel zu sehen, was gerade passiert.

```go
jobs := sched.List()
for _, job := range jobs {
    fmt.Printf("ID: %d, Running: %v, Next Run: %v\n", 
        job.ID, job.IsRunning, job.NextRun)
}
```

### Graceful Shutdown

Wenn die Anwendung beendet wird (z.B. SIGTERM), sollte man nicht mitten in einem Schreibvorgang abbrechen.

```go
// ... Signal empfangen ...
sched.StopAll() // 1. Sendet Stop-Signal an alle Jobs
                // 2. Wartet (blockierend), bis alle aktuell laufenden Ausführungen fertig sind
```

## ⚠️ Grenzen (Design Philosophy)

  * **In-Process:** Die Jobs leben im RAM. Startet die App neu, sind alle dynamisch geplanten Jobs weg (es sei denn, du lädst sie beim Start neu).
  * **Nicht Persistent:** Es gibt keine eingebaute Datenbank. Für kritische Jobs, die einen Neustart überleben müssen, sollte eine externe Queue oder DB genutzt werden.
  * **Kein "Distributed Lock":** Wenn du deine App 10x skalierst (z.B. in Kubernetes), läuft der Scheduler 10x.
