
|<sub>🇬🇧 [English translation →](README.en.md)</sub>|
|----:|
|    |

||[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](./LICENSE) [![Dependencies](https://img.shields.io/badge/dependencies-zero-brightgreen.svg)](#)|
|----|----|
|![GSF-Suite-Logo](logo-suite.png)| ***GSF-Suite***<br>Die GSF - Suite ist eine Sammlung kleiner, unabhängiger Go-Module zum Bau **einfacher Services**. Teil der **Go Simple Frameworks Suite**|

<sup>***GSF*** steht für ***Go Simple Frameworks*** — eine Sammlung von minimalistischen Tools für robuste Anwendungen.</sup>

### Überblick

GSF (Go Simple Frameworks) ist eine Sammlung kleiner, unabhängiger Go-Module zum Bau **einfacher Services**.

Der Fokus liegt auf **Klarheit, wenigen Abhängigkeiten und expliziter Zusammensetzung**, nach dem Prinzip einer pragmatischen *90%-Lösung*.

Dieses Repository (`gsf-go`) enthält die Go-Implementierung von GSF.

---

### Designprinzipien

- **Einfachheit zuerst** – kleine APIs, klare Verantwortung
- **90%-Lösungen** – praktikabel statt perfekt
- **Wenig Abhängigkeiten** – Standardbibliothek bevorzugt
- **Lose Kopplung** – Kommunikation über Interfaces
- **Sprachunabhängige Architektur** – geeignet für Polyglot-Systeme

---

### Module

Bitte beachten Sie auch die README in den jeweiligen Modulen.

#### `gsflog`
Ein minimaler Logger mit Loglevels und strukturierten Feldern.

- Schreibt auf beliebige `io.Writer`
- Keine Archivierung, Rotation oder Retention
- Kein Ersatz für etablierte Logging-Frameworks

Verantwortung:
> Logmeldungen formatieren und ausgeben

---

#### `rotate`
Ein generisches Modul zur Dateirotation.

- Arbeitet ausschließlich auf Dateien
- Rotation nach Größe und/oder Zeit
- Archivierungs- und Retention-Strategien austauschbar
- Keine Abhängigkeit zu Logging

Verantwortung:
> Dateien nach Policies behandeln

---

#### `schedule`
Ein einfacher Scheduler für zeitgesteuerte Jobs.

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


