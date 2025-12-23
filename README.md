
|<sub>🇩🇪 [German translation →](README.de.md)</sub>|
|----:|
|    |

||[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](./LICENSE) [![Dependencies](https://img.shields.io/badge/dependencies-zero-brightgreen.svg)](#)|
|----|----|
|![GSF-Suite-Logo](logo-suite.png)| ***GSF-Suite***<br>GSF is a collection of small, independent Go modules for building **small services**. Member of **Go Small Frameworks Suite**|

<sup>***GSF*** stands für ***Go Small Frameworks*** — a collection of minimalist tools for robust applications.</sup>

### Overview

GSF is a collection of small, independent Go modules for building **small and simple services**.
The focus is on **clarity, low dependencies, and composability**, following a pragmatic *"90% solution"* philosophy.

This repository (`gsf-go`) contains the Go implementation of GSF.

GSF (Go Small Frameworks) provides a set of minimal infrastructure building blocks for microservices and service-oriented systems.

The modules are intentionally small and focused. Each module:
- solves one problem
- can be used independently
- avoids unnecessary abstractions
- prefers explicit composition over magic

GSF is **not** a full-stack framework and does not try to replace existing ecosystems.

---

### Design Principles

- **Simplicity first** – small APIs, clear responsibilities
- **90% solutions** – practical over perfect
- **Low dependencies** – standard library where possible
- **Loose coupling** – modules communicate via interfaces
- **Language-agnostic architecture** – suitable for polyglot systems

---

### Modules

Please also refer to the README files in the respective modules.

#### `gsflog`
A minimal logger with log levels and structured fields.

- Writes to any `io.Writer`
- No log retention, rotation, or compression logic
- No replacement for `slog`, `zerolog`, etc.

Responsibility:
> Format and emit log messages

---

#### `rotate`
A generic file rotation module.

- Works on files only (not log-specific)
- Rotation based on size and/or time
- Archive and retention strategies are pluggable
- No logging dependency

Responsibility:
> Handle files according to rotation policies

---

#### `schedule`
A small job scheduler.

- Periodic jobs (`Every`)
- One-shot jobs (`At`)
- Panic-safe execution
- Optional logger interface

Responsibility:
> Execute jobs at specific times

---

### Composition Model

GSF modules are designed to be composed explicitly:

- `gsflog` writes to an `io.Writer`
- `rotate.Writer` implements `io.Writer`
- `schedule` can trigger rotations or reopen operations

There are **no hard dependencies** between modules.
Integration happens in the application layer.

---

### Examples

The `cmd/` directory contains runnable examples:

- `cmd/main.go` – full example (logger + rotation + scheduler)
- `cmd/rotate/main.go` – standalone rotation example
- `cmd/schedule/main.go` – scheduler example

Each example is self-contained and meant as documentation by code.

---

### Non-Goals

GSF deliberately does **not** provide:

- distributed logging
- tracing
- metrics
- service discovery
- configuration frameworks

GSF is infrastructure glue, not a platform.

---

### License

Licensed under the Apache License, Version 2.0.

---

## Contributing & Security

Contributions to the GSF Suite are welcome — including bug reports,
improvements, and pull requests.

Please refer to:
- Contribution guidelines: see `CONTRIBUTING.md`
- Responsible disclosure of security issues: see `SECURITY.md`

For general bugs or feature ideas, please use GitHub Issues.
Security-related topics should **not** be discussed publicly.

