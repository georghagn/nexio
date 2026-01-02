<sub>🇩🇪 [German translation →](README.de.md)</sub>

---

### Overview

The **GSF-Suite/Logger** is a structured, modular, composable logger for Go, developed according to the "Tiny Frameworks" philosophy. It strictly separates data acquisition (`Logger`), formatting (`Formatter`), and output (`io.Writer`).

It is designed for small services and infrastructure code where simplicity, explicitness, and low dependencies matter more than features.

### Features

* **Structured Logging:** No more string concatenation. Use key-value pairs (`With("user_id", 42)`).
* **Formatters:**
* `TextFormatter`: Colorful output for the console (Dev Mode).
* `JSONFormatter`: Machine-readable JSON for production (ELK, Splunk, CloudWatch).
* **Rotation Strategies:** Supports both **internal** (automatic) and **external** (signal/scheduler-based) file rotation.
* **Thread-Safe:** Secure access from any number of goroutines.


### Installation

```bash
go get github.com/georghagn/nexio/pkg/gsflog
````
---

### `gsflog` provides:

- log levels (`Debug`, `Info`, `Warn`, `Error`)
- structured fields (`With(key, value)`)
- pluggable output via `io.Writer`

It deliberately does **not** try to compete with full-featured logging frameworks such as `slog`, `zap`, or `zerolog`.

---

### Design Goals

- Small and predictable API
- No global state
- No hidden background goroutines
- Output abstraction via standard interfaces
- Easy integration with external components (rotation, scheduling)

---

### Non-Goals

`gsflog` intentionally does **not** provide:

- log file rotation
- retention or archival
- async buffering
- distributed logging
- structured log backends

Those concerns are expected to be handled by **external modules**.

---

### Output Model

`gsflog` writes log entries to an `io.Writer`.

This allows seamless integration with:

- `os.Stdout` / `os.Stderr`
- files
- rotating writers (e.g. `rotate.Writer`)
- custom writers

The logger assumes that the writer is responsible for:

- file lifecycle
- reopening files if required
- synchronization guarantees

---

### Reopenable Writers

For external log rotation strategies (e.g. via scheduler or signals), `gsflog` provides a `ReopenableWriter`.

This allows a log file to be closed and reopened at runtime after it has been moved or replaced.

Typical use cases:

- time-based rotation via scheduler
- external tools (e.g. logrotate-style workflows)

---

### Integration Example

Working examples can be found at:

- `cmd/rotate-example1/main.go` – Logging with rotation
- `cmd/rotate-example2/main.go` – Logging with rotation
- `cmd/gsflog-example/main.go` – Individuell Configuration

The examples demonstrate explicit wiring instead of implicit configuration.

---

### Error Handling Philosophy

`gsflog` follows a simple rule:

> Errors are returned, not logged.

The logger itself avoids logging internal errors.
Handling and interpretation of errors is expected at the application boundary.

---

## License / Contact

LICENSE, CONTRIBUTE.md, SECURITY.md and contact information can be found in the root of the suite.


