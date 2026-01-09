
|<sub>🇬🇧 [English translation →](README.md)</sub>|
|----:|
|    |

||[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](./LICENSE) [![Dependencies](https://img.shields.io/badge/dependencies-zero-brightgreen.svg)](#)|
|----|----|
|![GSF-Suite-Logo](logo-suite.png)| ***GSF-nexIO Suite***<br>Die GSF - Suite ist eine Sammlung kleiner, unabhängiger Go-Module zum Bau **einfacher Services**. Teil der **Go Small Frameworks Suite**|

<sup>***GSF*** steht für ***Go Small Frameworks*** — eine Sammlung von minimalistischen Tools für robuste Anwendungen.</sup>

### Übersicht

**GSF-nexIO** bietet eine Auswahl an minimalen Infrastruktur-Bausteinen für Microservices und serviceorientierte Systeme. Der Fokus liegt auf **Klarheit, geringen Abhängigkeiten und Komponierbarkeit** – ganz nach der pragmatischen Philosophie der *"90%-Lösung"*.

Die Module sind bewusst fokussiert:

* Lösen jeweils ein spezifisches Problem gründlich.
* Unabhängige Nutzung (kein Zwang zu monolithischen Strukturen).
* Bevorzugen explizite Komposition gegenüber "magischen" Abstraktionen.
* Minimale externe Abhängigkeiten.

---

### Quick Start: Ein Node in 3 Zeilen

```go
logger := nexlog.Wrap(nexlog.NewDefaultConsole())
provider := transport.NewWSProvider(logger)
node := rpc.NewNode(nil, provider, "ws://localhost:8080/ws", logger)
go node.Listen(ctx)

```

---

### Design-Prinzipien

* **Einfachheit zuerst** – kleine APIs, klare Verantwortlichkeiten.
* **90%-Lösungen** – praktische, stabile Lösungen vor theoretischer Perfektion.
* **Lose Kopplung** – Module kommunizieren über saubere Interfaces.
* **Sprachunabhängige Architektur** – Konzepte, die auch für polyglotte Systeme geeignet sind.

---

### Die Smalltalk-Philosophie

**GSF-nexIO** ist das Ergebnis des Versuchs, die Flexibilität und das intuitive Design klassischer Smalltalk-Umgebungen in die moderne Systemprogrammierung mit Go zu übertragen.

Meine langjährige Erfahrung mit Smalltalk-Systemen (siehe auch meine `TSF`-Projekte) prägt die Architektur von nexIO entscheidend:

* **Nachrichtenaustausch statt Funktionsaufrufe:** Inspiriert durch das Smalltalk-Paradigma „Everything is a Message“, konzentriert sich nexIO auf den freien Fluss von Nachrichten zwischen Objekten, anstatt auf starre Client-Server-Hierarchien.
* **Objekt-Symmetrie:** In Smalltalk sind Objekte gleichberechtigte Akteure. Diese Philosophie spiegelt sich in unseren **Symmetrical Nodes** wider, die gleichzeitig Sender und Empfänger sein können.
* **Entkopplung & Komposition:** Smalltalk-Systeme glänzen durch ihre Fähigkeit, einfache, spezialisierte Komponenten zu komplexen Systemen zu kombinieren. nexIO folgt diesem Vorbild durch strikt entkoppelte Module, die über Interfaces kommunizieren.

**Warum Go?** nexIO schlägt die Brücke: Die bewährten Interaktionsmuster aus der Smalltalk-Welt treffen hier auf die Typsicherheit, Nebenläufigkeit (Goroutines) und Performance von Go.

---

### Module

* [**node**](./node): Resiliente P2P RPC-Kommunikation.
* [**nexlog**](./nexlog): Strukturiertes Logging mit Adapter-Unterstützung.
* [**rotate**](./nexlog/rotate): Sicherer Datei-Rotator mit `.LOCK`-Synchronisierung.
* [**schedule**](./schedule): Zuverlässige Aufgabenplanung (Scheduling).

#### nexIOnode (`node`)

Das Herzstück der bidirektionalen Kommunikation. Es bricht mit dem klassischen Client-Server-Paradigma und ersetzt es durch eine **symmetrische Peer-Architektur**.

* **Symmetrie:** Sobald die Verbindung steht, kann jeder Node Methoden registrieren und gleichzeitig seinen Partner als Client aufrufen.
* **Rollenunabhängig:** Während Verbindungen als Client/Server starten, agieren nach dem Aufbau alle Teilnehmer als gleichberechtigte Peers. Dies wird im Beispiel `cmd/node/gsfNodesExamples` verdeutlicht, wo ein "Payment Service" und mehrere "Order Services" bidirektional interagieren.
* **Resilienz-Engine:** Integrierter Zustandsautomat mit exponentiellem Backoff für transparente Wiederverbindungen.
* **Typsicherheit:** Nutzt Go Generics (`Bind[T]`), um JSON-RPC-Parameter sicher in native Go-Strukturen zu überführen.

#### nexlog & rotate (`nexlog` & `nexlog/rotate`)

Ein strukturiertes Logging-System, optimiert für den Langzeitbetrieb.

* **Interface-Abstraktion:** Entkoppelt über `LogSink`, was die Nutzung in jedem Modul ohne harte Abhängigkeiten ermöglicht.
* **Atomare Rotation:** Robuste Dateirotation mit einem `.LOCK`-Mechanismus.
* **Sichere Operationen:** Jedes Log-Ereignis folgt einem **Open -> Write -> Close** Zyklus, was die Integrität auch bei Systemabstürzen garantiert.
* **Kontextuelles Tracing:** Unterstützt die Anreicherung von Feldern via `With(key, value)` für verteiltes Tracing.

#### nexIOschedule (`schedule`)

Ein präziser, "panic-sicherer" Scheduler für wiederkehrende Aufgaben.

* **Interface-gesteuert:** Führen Sie jede beliebige Go-Funktion über ein einfaches Task-Interface aus.
* **Konkurrenz-sicher:** Entwickelt, um hunderte parallele Jobs zu verwalten.
* **Fehlertoleranz:** Fehlgeschlagene Jobs werden mit vollem Kontext über das integrierte `LogSink` protokolliert.

---

### Kompositions-Modell

nexIO-Module sind für die explizite Komposition konzipiert:

* `nexlog` schreibt in einen `io.Writer`.
* `rotate.Writer` implementiert `io.Writer`.
* `schedule` kann Wartungsaufgaben wie die Log-Rotation auslösen.

Die Integration findet in der Anwendungsschicht statt – **keine harten Abhängigkeiten** zwischen den Kernmodulen.

---

### Beispiele

Das Verzeichnis `cmd/` enthält selbsterklärenden Code:

* `cmd/node/gsfNodesExamples/` – **Die Peer-to-Peer Demo**: Interaktion eines Payment-Servers mit mehreren Order-Clients.
* `cmd/rotate/main.go` – Eigenständige Datei-Rotation.
* `cmd/schedule/main.go` – Nutzung des Schedulers.

---

### Organisation & Standards

* **Copyright:** © 2026 Georg Hagn.
* **Namespace:** `github.com/georghagn/nexio/pkg/...`
* **Lizenz:** Apache License, Version 2.0.

GSF-nexIO ist ein unabhängiges open-source project und ist mit keinem Unternehmen ähnlichen Namens verbunden.

---

## Mitwirken & Sicherheit

Beiträge sind willkommen! Bitte nutzen Sie GitHub Issues für Fehlerberichte oder Feature-Ideen.
**Sicherheitsrelevante Themen** sollten nicht öffentlich diskutiert werden; bitte beachten Sie hierzu die `SECURITY.md`.

---

## Kontakt

Bei Fragen oder Interesse an diesem Projekt erreichen Sie mich unter:
📧 *georghagn [at] tiny-frameworks.io*

<sup>*(Bitte keine Anfragen an die privaten GitHub-Account-Adressen)*</sup>


