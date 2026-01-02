
<sub>🇩🇪 [German translation →](README.de.md)</sub>

---

## Overview

**GSF-Suite/Rotator** is a minimalist, robust, and thread-safe **file rotator** for Go, designed to be zero-dependency and fully standalone.

It implements `io.WriteCloser` and can be seamlessly used as a backend for loggers (e.g., stdlib `log`, `zap`, `zerolog`, or `gsflog`). It is part of the **GSF (Go Small Frameworks)** suite but is fully **standalone**.

## Features

* **Zero Dependencies:** Uses only the Go standard library.
* **Thread-Safe:** Safe access from multiple goroutines (via `sync.Mutex`).
* **Inter-Process Ready:** Optional atomic file-locking for external schedulers/loggers.
* **Modular:** Uses the *Strategy Pattern* for maximum flexibility.
* **io.Writer Compatible:** Drop it in anywhere a writer is expected.

## Core Concept

The central type is `rotate.Writer`, which implements the `io.Writer` interface.

For every write operation, the rotator:

1. Opens the file if necessary.
2. Checks if a rotation is required.
3. Rotates the file if the policy triggers.
4. Writes the content.

The decision to rotate is delegated to specific **Policies**.

## Installation

```bash
go get github.com/georghagn/nexio/pkg/rotate

```

## Quick Start

The simplest way: A rotator that starts a new file at 10 MB and keeps all old files.

```go
package main

import (
    "github.com/georghagn/nexio/pkg/rotate"
)

func main() {
    // Filename, Defaults (10MB Limit, no compression, keep all)
    r := rotate.New("app.log", nil, nil, nil) 
    defer r.Close()
}

```

## Configuration (Advanced)

The rotator is controlled by three strategies. You can customize each one individually:

1. **RotationPolicy:** *When* should the file be rotated?
2. **ArchiveStrategy:** *How* should the old file be processed?
3. **RetentionPolicy:** *Which* old files should be deleted?

### Example: Gzip Compression & Cleanup

Here we create a rotator that:

* Rotates at **5 MB**.
* Compresses old files using **Gzip** (`.gz`).
* Keeps only the **5 newest** backups.

```go
writer := rotate.New("server.log",
    &rotate.SizePolicy{MaxBytes: 5 * 1024 * 1024}, // 5 MB Limit
    &rotate.GzipCompression{},                     // Compress
    &rotate.MaxFiles{MaxBackups: 5},               // Keep only 5
)
defer writer.Close()

writer.Write([]byte("Hello World\n"))

```

## Inter-Process Locking (Multi-Process)

If you use an external process (like a standalone Scheduler) to trigger rotation while another process is writing, use the **Atomic File Lock**:

```go
import "github.com/georghagn/nexio/pkg/rotate/lock"

// See documentation (doc.go) for implementation details on 
// cross-platform safety (Windows vs. iOS/Linux).

```

## Available Strategies

### Rotation

* `SizePolicy{MaxBytes: int64}`: Rotates when the file size exceeds the limit.

### Archive

* `NoCompression{}`: Simply renames the file (includes a timestamp in the name).
* `GzipCompression{}`: Compresses the file into `.gz` format and deletes the original.
* *Note:* Uses millisecond timestamps to avoid collisions under high load.



### Retention (Cleanup)

* `KeepAll{}`: Keeps all files (Default).
* `MaxFiles{MaxBackups: int}`: Deletes the oldest backups when the limit is reached.

## Concurrency

The package is **thread-safe**. You can pass the same `*rotate.Writer` instance to multiple goroutines or use it within a logger called by multiple routines. Internal locks prevent race conditions during writing or rotating.

## Examples

Typical integrations:

* Logging via `gsflog`
* Scheduled rotation via `schedule`
* Custom triggers

A runnable example can be found at `cmd/rotate-example/main.go`.

---

## License / Contact

LICENSE, CONTRIBUTE.md, SECURITY.md, and contact information can be found in the root of the suite.



