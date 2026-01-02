
<sub>🇬🇧 [English translation →](README.en.md)</sub>

---

### Überblick

Der **GSF-Suite/Logger** ist ein strukturierter, modularer komponierbarer Logger für Go, entwickelt nach der "Tiny Frameworks" Philosophie. Er trennt strikt zwischen Datenerfassung (`Logger`), Formatierung (`Formatter`) und Ausgabe (`io.Writer`).

Er richtet sich an kleine Services und Infrastruktur-Code, bei denen Einfachheit und explizite Kontrolle wichtiger sind als Funktionsvielfalt.


### Features

  * **Structured Logging:** Keine String-Verkettung mehr. Nutze Key-Value Paare (`With("user_id", 42)`).
  * **Formatters:**
      * `TextFormatter`: Bunte Ausgabe für die Konsole (Dev-Mode).
      * `JSONFormatter`: Maschinenlesbares JSON für Produktion (ELK, Splunk, CloudWatch).
  * **Rotation Strategies:** Unterstützt sowohl **interne** (automatische) als auch **externe** (Signal/Scheduler-basierte) Datei-Rotation.
  * **Thread-Safe:** Sicherer Zugriff aus beliebig vielen Goroutinen.


### Installation

```bash
go get github.com/georghagn/nexio/pkg/gsflog
````
---

### `gsflog` bietet:

- Loglevel (`Debug`, `Info`, `Warn`, `Error`)
- strukturierte Felder (`With(key, value)`)
- austauschbare Ausgabe über `io.Writer`

Es versucht bewusst **nicht** mit vollwertigen Logging-Frameworks wie `slog`, `zap` oder `zerolog` zu konkurrieren.

---

### Designziele

- Kleine, überschaubare API
- Kein globaler Zustand
- Keine versteckten Hintergrund-Goroutinen
- Ausgabe über Standard-Interfaces (`io.Writer`)
- Einfache Integration mit externen Modulen

---

### Nicht-Ziele

`gsflog` stellt bewusst **keine** Funktionen bereit für:

- Logrotation
- Retention oder Archivierung
- Asynchrones Logging
- Verteiltes Logging
- Anbindung an spezielle Log-Backends

Diese Aufgaben werden an externe Module delegiert.

---

### Ausgabemodell

`gsflog` schreibt Logeinträge auf ein `io.Writer`.

Dadurch kann der Logger unter anderem mit folgenden Zielen verwendet werden:

- `Stdout` / `Stderr`
- Dateien
- rotierenden Writer-Implementierungen
- eigenen Writer-Typen

Die Verantwortung für Dateihandling und Synchronisation liegt beim Writer.

---

### Reopenable Writer

Für externe Rotationsstrategien stellt `gsflog` einen `ReopenableWriter` bereit.

Dieser erlaubt es, Logdateien zur Laufzeit zu schließen und erneut zu öffnen,
beispielsweise nachdem sie von außen verschoben oder rotiert wurden.

Typische use cases:

- Time-based Rotation via Scheduler
- Externe Rools (z.B.: logrotate-style workflows)

---

### Beispiele

Lauffähiges Beispiele befindet sich unter:

- `cmd/rotate-example1/main.go` – Logging mit Rotation
- `cmd/rotate-example2/main.go` – Logging mit Rotation
- `cmd/gsflog-example/main.go` – Individuelle Konfiguration

Die Beispiele sind bewusst einfach gehalten und zeigen explizite Verdrahtung.

---

### Fehlerbehandlung

`gsflog` folgt einer klaren Regel:

> Fehler werden zurückgegeben, nicht geloggt.

Die Behandlung von Fehlern erfolgt auf Anwendungsebene.

---

## License / Kontakt

LICENSE, CONTRIBUTE.md, SECURITY.md und Kontaktinformationen findest du im Root der Suite


