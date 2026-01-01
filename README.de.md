
|<sub>🇬🇧 [English translation →](README.md)</sub>|
|----:|
|    |

||[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](./LICENSE) [![Dependencies](https://img.shields.io/badge/dependencies-zero-brightgreen.svg)](#)|
|----|----|
|![GSF-Suite-Logo](logo-suite.png)| ***GSF-Suite***<br>Die GSF - Suite ist eine Sammlung kleiner, unabhängiger Go-Module zum Bau **einfacher Services**. Teil der **Go Small Frameworks Suite**|

<sup>***GSF*** steht für ***Go Small Frameworks*** — eine Sammlung von minimalistischen Tools für robuste Anwendungen.</sup>

### Überblick

GSF (Go Small Frameworks) ist eine Sammlung kleiner, unabhängiger Go-Module zum Bau **einfacher Services**.

Der Fokus liegt auf **Klarheit, wenigen Abhängigkeiten und expliziter Zusammensetzung**, nach dem Prinzip einer pragmatischen *90%-Lösung*.

Dieses Repository (`gsf-go`) enthält die Go-Implementierung von GSF.

---

### Designprinzipien

- **Einfachheit zuerst** – kleine APIs, klare Verantwortung
- **90%-Lösungen** – praktikabel statt perfekt
- **Wenig Abhängigkeiten** – Standardbibliothek bevorzugt
- **Lose Kopplung** – Kommunikation über Interfaces
- **Sprachunabhängige Architektur** – geeignet für Polyglot-Systeme



### Module

Bitte beachten Sie auch die README in den jeweiligen Modulen.

---

#### nexIOnode (pkg/node)

Das Herzstück der bidirektionalen Kommunikation. Es bricht das klassische Client-Server-Paradigma auf und ersetzt es durch eine symmetrische Peer-Architektur.

* **Symmetrie:** Jeder Node kann Methoden registrieren und gleichzeitig als Client beim Partner Anfragen stellen.
* **Resilienz-Engine:** Ein integrierter Zustandsautomat überwacht die Verbindung und nutzt einen exponentiellen Backoff für die Wiederverbindung, ohne die laufende Applikationslogik zu blockieren.
* **Typ-Sicherheit:** Durch Go Generics (`Bind[T]`) werden JSON-RPC Parameter sicher in native Go-Strukturen überführt.

---

#### nexIOlog & nexIOlog/rotate (pkg/gsflog)

Ein hochperformantes, strukturiertes Logging-System, das für den Langzeitbetrieb in Microservices optimiert wurde.

* **Interface-Abstraktion:** Über das `LogSink`-Interface entkoppelt, kann der Logger in jedem Modul (RPC, Transport, Scheduler) eingesetzt werden, ohne harte Abhängigkeiten zu erzeugen.
* **Atomic Rotation:** Implementiert eine robuste Dateirotation mit `.LOCK`-Mechanismus. Jedes Log-Event wird atomar geschrieben (Open -> Write -> Close), was maximale Integrität auch bei Systemabstürzen garantiert.
* **Contextual Logging:** Unterstützt das Anreichern von Log-Einträgen mit Kontext-Daten (`With`), um Tracing über verteilte Nodes hinweg zu ermöglichen.

##### `gsflog`
Ein minimaler Logger mit Loglevels und strukturierten Feldern.

- Schreibt auf beliebige `io.Writer`
- Keine Archivierung, Rotation oder Retention
- Kein Ersatz für etablierte Logging-Frameworks

Verantwortung:
> Logmeldungen formatieren und ausgeben


##### `rotate`
Ein generisches Modul zur Dateirotation.

- Arbeitet ausschließlich auf Dateien
- Rotation nach Größe und/oder Zeit
- Archivierungs- und Retention-Strategien austauschbar
- Keine Abhängigkeit zu Logging

Verantwortung:
> Dateien nach Policies behandeln

---

#### nexIOschedule (pkg/schedule)

Ein präziser Zeitplaner für wiederkehrende Aufgaben innerhalb der nexIO-Ökosystems.

* **Interface-Driven:** Aufgaben werden über ein einfaches Interface definiert, was die Ausführung beliebiger Go-Funktionen ermöglicht.
* **Concurrency-Safe:** Der Scheduler ist darauf ausgelegt, hunderte parallele Jobs zu verwalten, ohne die Echtzeitfähigkeit der RPC-Kommunikation zu beeinträchtigen.
* **Fehlertoleranz:** Schlägt ein Job fehl, wird dies über das integrierte `nexlog`-System mit vollem Kontext protokolliert.


- Periodische Jobs (`Every`)
- Einmalige Jobs (`At`)
- Panic-sichere Ausführung
- Optionales Logger-Interface

Verantwortung:
> Jobs zeitgesteuert ausführen

---

### Zusammenspiel

Die Module werden explizit zusammengesetzt:

- `gsflog` schreibt auf ein `io.Writer`
- `rotate.Writer` implementiert `io.Writer`
- `schedule` kann Rotation oder Reopen auslösen

Es gibt **keine festen Abhängigkeiten** zwischen den Modulen.
Die Integration erfolgt auf Anwendungsebene.

---

### Beispiele

Im Verzeichnis `cmd/` befinden sich lauffähige Beispiele:

- `cmd/main.go` – vollständiges Beispiel
- `cmd/rotate/main.go` – Rotation isoliert
- `cmd/schedule/main.go` – Scheduler isoliert
- `cmd/node/gsfNodeExample/.../main.go` – Zusammenspiel von 3 Nodes

Die Beispiele dienen bewusst als ausführbare Dokumentation.

---

### Nicht-Ziele

GSF stellt bewusst **keine** Plattform bereit für:

- verteiltes Logging
- Tracing
- Metriken
- Service Discovery
- Konfigurations-Frameworks

GSF ist Infrastruktur-Baustein, kein Framework.

---
### Organisatorisches & Standards

* **Copyright:** © 2026 Georg Hagn.
* **Namespace:** Alle Module folgen der Namenskonvention `github.com/georghagn/gsf-suite/pkg/...`.
* **Clean Code:** Strikte Trennung von Transport-Logik (WebSockets) und Applikations-Logik (RPC).

---

### Lizenz

Lizenziert unter der Apache License, Version 2.0.

---

## Contributing & Security

Beiträge zur GSF Suite sind willkommen – sei es in Form von Bug-Reports,
Verbesserungsvorschlägen oder Pull Requests.

Bitte beachte dazu:
- Hinweise zum Beitragen: siehe `CONTRIBUTING.md`
- Verantwortungsvolle Meldung von Sicherheitslücken: siehe `SECURITY.md`

Für normale Bugs oder Feature-Ideen nutze bitte GitHub Issues.
Sicherheitsrelevante Themen sollten **nicht öffentlich** diskutiert werden.


