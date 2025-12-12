# gsf/rotate

Ein minimalistischer, robuster und thread-sicherer **File Rotator** für Go. 
Er implementiert `io.WriteCloser` und kann nahtlos als Backend für Logger (z.B. stdlib `log`, `zap`, `zerolog` oder `gsflog`) verwendet werden.

Teil der **GSF (Go Small Frameworks)** Suite, aber vollständig **standalone** nutzbar.

## Features

* **Zero Dependencies:** Nutzt nur die Go Standardbibliothek.
* **Thread-Safe:** Sicherer Zugriff aus mehreren Goroutinen (durch `sync.Mutex`).
* **Modular:** Nutzt das *Strategy Pattern* für maximale Flexibilität.
* **Io.Writer Kompatibel:** Einfach überall einstecken, wo ein Writer erwartet wird.

## Installation

```bash
go get github.com/georghagn/gsf-go/pkg/rotate
````

## Quick Start

Der einfachste Weg: Ein Rotator, der bei 10 MB eine neue Datei anfängt und alte Dateien behält.

```go
package main

import (
    "log"
    "github.com/georghagn/gsf-go/pkg/rotate"
)

func main() {
    // Filename, Defaults (10MB Limit, keine Kompression, alles behalten)
    r := rotate.New("app.log", nil, nil, nil) 
    defer r.Close()

    log.SetOutput(r)
    log.Println("Dieser Log-Eintrag wird automatisch rotiert!")
}
```

## Konfiguration (Advanced)

Der Rotator wird durch drei Strategien gesteuert. Du kannst jede einzeln anpassen:

1.  **RotationPolicy:** *Wann* soll rotiert werden?
2.  **ArchiveStrategy:** *Wie* soll die alte Datei verarbeitet werden?
3.  **RetentionPolicy:** *Welche* alten Dateien sollen gelöscht werden?

### Beispiel: Gzip Kompression & Aufräumen

Hier erstellen wir einen Rotator, der:

  * Bei **5 MB** rotiert.
  * Die alten Dateien mit **Gzip** komprimiert (`.gz`).
  * Nur die **5 neuesten** Backups behält.

<!-- end list -->

```go
writer := rotate.New("server.log",
    &rotate.SizePolicy{MaxBytes: 5 * 1024 * 1024}, // 5 MB Limit
    &rotate.GzipCompression{},                     // Komprimieren
    &rotate.MaxFiles{MaxBackups: 5},               // Nur 5 behalten
)
defer writer.Close()

writer.Write([]byte("Hello World\n"))
```

## Verfügbare Strategien

### Rotation

  * `SizePolicy{MaxBytes: int64}`: Rotiert, wenn die Dateigröße das Limit überschreitet.

### Archive

  * `NoCompression{}`: Benennt die Datei einfach um (Timestamp im Namen).
  * `GzipCompression{}`: Komprimiert die Datei im `.gz` Format und löscht das Original.
      * *Hinweis:* Nutzt Millisekunden-Timestamps, um Kollisionen bei hoher Last zu vermeiden.

### Retention (Aufräumen)

  * `KeepAll{}`: Behält alle Dateien (Default).
  * `MaxFiles{MaxBackups: int}`: Löscht die ältesten Backups, wenn das Limit überschritten wird.

## Concurrency

Das Paket ist **Thread-Safe**. Du kannst denselben `*rotate.Writer` Instanz an mehrere Goroutinen übergeben oder in einem `logger` nutzen, der von mehreren Routinen aufgerufen wird. Interne Locks verhindern Race Conditions beim Schreiben oder Rotieren.
