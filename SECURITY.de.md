
<sub>🇬🇧 [English translation →](SECURITY.en.md)</sub>


# Sicherheitsrichtlinie (Security Policy)

## Unterstützte Versionen

GSF (Go Simple Services) befindet sich derzeit in einer **frühen öffentlichen Entwicklungsphase**.

* Versionen **vor v1.0.0** werden nach dem *Best-Effort-Prinzip* bereitgestellt.
* Sicherheitsfixes können erfolgen, ohne vollständige Abwärtskompatibilität zu garantieren.

Sobald ein Modul den Status **v1.0.0** erreicht, werden **kritische Sicherheitsprobleme** mit entsprechenden Releases adressiert.

---

## Melden von Sicherheitslücken

Wenn du eine potenzielle Sicherheitslücke entdeckst, **bitte kein öffentliches Issue eröffnen**.

Stattdessen bitten wir um eine verantwortungsvolle Meldung über einen der folgenden Wege:

* Kontaktaufnahme mit dem Maintainer über GitHub (private Nachricht)
* Alternativ: Schicke eine Email an die Adresse im README
* Optional: ein **Draft Pull Request**, der das Problem und einen möglichen Fix beschreibt

Bitte gib dabei möglichst an:

* Eine klare Beschreibung der Schwachstelle
* Schritte zur Reproduktion (falls zutreffend)
* Betroffene(s) Modul(e)
* Mögliche Gegenmaßnahmen oder einen Fix-Vorschlag

Für allgemeine Bugs, Feature-Vorschläge oder Fragen nutze bitte die regulären GitHub Issues
gemäß den Hinweisen in `CONTRIBUTING.md`.

---

## Geltungsbereich (Scope)

GSF konzentriert sich bewusst auf:

* **In-Process Libraries** (z. B. Logging, Scheduling, File Rotation)
* **Minimale Abhängigkeiten** (bevorzugt Go-Standardbibliothek)

Nicht im Fokus dieses Projekts sind:

* Netzwerksicherheit (TLS, Authentifizierung, Autorisierung)
* Betriebssystem-Härtung
* Anwendungsspezifische Security-Policies

---

## Umgang mit gemeldeten Sicherheitsproblemen

Gemeldete Sicherheitslücken werden:

1. so zeitnah wie möglich geprüft
2. bei Bedarf in einem privaten Branch behoben
3. öffentlich veröffentlicht, sobald ein Fix verfügbar ist

Aktuell gibt es **keinen formalen CVE-Prozess**.

---

## Abschließende Bemerkung

GSF verfolgt das Ziel, **einfach, vorhersehbar und transparent** zu sein.

Wenn dir etwas unsicher, unklar oder überraschend erscheint:
👉 **Bitte melde es.**

Sicherheit entsteht durch gemeinsame Aufmerksamkeit.

---

