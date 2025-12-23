
<sub>🇬🇧 [English translation →](CONTRIBUTING.md)</sub>


# Beitragen zur GSF Suite

Erstmal: **Vielen Dank**, dass du dir die Zeit nimmst, zur **GSF‑Suite** beizutragen!
Open Source lebt von Menschen, die ihr Wissen, ihre Zeit und ihre Erfahrung teilen.

Dieses Dokument beschreibt einige Leitlinien, die helfen sollen, das Projekt **übersichtlich, stabil und langfristig wartbar** zu halten.

---

## 1. Bugs melden

Bevor du ein neues Issue erstellst, prüfe bitte:

* ob das Problem bereits gemeldet wurde
* ob du die **neueste verfügbare Version** des jeweiligen Moduls verwendest

Wenn du ein Bug‑Report erstellst, helfen uns folgende Informationen sehr:

* verwendete Version(en)
* Schritte zur Reproduktion
* erwartetes Verhalten vs. tatsächliches Verhalten
* optional: ein **minimales Code‑Beispiel**

> **Hinweis zu Sicherheitslücken:**
> Wenn du vermutest, dass es sich um ein **sicherheitsrelevantes Problem** handelt,
> **bitte kein öffentliches Issue eröffnen**, sondern die Hinweise in der `SECURITY.md` beachten.

---

## 2. Pull Requests (Code beitragen)

Wir freuen uns über Pull Requests – egal ob Bugfix, Verbesserung oder neues Feature.

Damit dein PR gut nachvollziehbar ist und zügig geprüft werden kann, beachte bitte:

1. **Fork & Branch**
   Erstelle einen Fork des Repositories und arbeite in einem eigenen Feature‑Branch:

   ```bash
   git checkout -b feature/mein-feature
   ```

2. **Coding Style**
   Halte dich bitte an den bestehenden Code‑Stil und die Design‑Prinzipien des jeweiligen Moduls.

3. **Tests**

   * Neue Funktionalität sollte durch passende Tests begleitet werden
   * Bugfixes sollten – wenn sinnvoll – einen Regressionstest enthalten

4. **Lizenz‑Header**
   Neue Dateien müssen den korrekten SPDX‑Lizenz‑Header enthalten:

   ```go
   // SPDX-License-Identifier: Apache-2.0
   ```

5. **Security‑relevante Änderungen**
   Wenn dein Pull Request eine potenzielle Sicherheitslücke betrifft,
   orientiere dich bitte an der `SECURITY.md` und reiche den PR ggf. zunächst als **Draft** ein.

---

## 3. Rechtliches & Lizenzierung

Durch das Einreichen eines Pull Requests bestätigst du, dass:

1. du der Urheber des beigetragenen Codes bist **oder** die notwendigen Rechte besitzt
2. dein Beitrag unter der **Apache License 2.0** veröffentlicht werden darf

Dieses Projekt folgt dem Prinzip **„Inbound = Outbound“**:

* Es wird **kein separater Contributor License Agreement (CLA)** benötigt
* Alle Beiträge stehen automatisch unter derselben Lizenz wie das Projekt selbst

---

## 4. Philosophie

Die GSF‑Suite folgt bewusst einer **Tiny / Simple‑Philosophie**:

* minimale Abhängigkeiten
* explizite APIs statt Magie
* kleine, klar abgegrenzte Module
* Vorhersehbarkeit vor Feature‑Reichtum

Beiträge sollten diese Grundhaltung respektieren.

---

Vielen Dank für deine Unterstützung!
**Wir** freuen uns über konstruktive Diskussionen, saubere Beiträge und gemeinsame Weiterentwicklung.

—
Georg Hagn

