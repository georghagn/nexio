
**GSF – Go Small Framework** 
Design - Decisions

## Purpose of this Document

This document explains the **design decisions, scope, and non-goals** of the GSF modules.
It is intended for contributors, reviewers, and users who want to understand *why* things are built the way they are – not just *how*.

GSF follows a **“small, explicit, and predictable”** philosophy.

---

## Core Design Principles

### 1. Small & Composable

Each module:

* solves **one clearly defined problem**
* can be used **standalone**
* avoids hidden coupling to other modules

Composition happens **at the application level**, not inside the framework.

---

### 2. In-Process First

All GSF modules are designed as **in-process libraries**.

This means:

* thread-safe within **one Go process**
* no assumptions about external coordination
* no background daemons
* no persistent runtime state unless explicitly stated

This keeps behavior predictable and testable.

---

### 3. Explicit over Implicit

GSF avoids:

* implicit retries
* hidden background goroutines
* automatic recovery that hides errors

If something can fail, it should **return an error** and let the caller decide.

---

### 4. Zero or Minimal Dependencies

Whenever possible:

* Go standard library only
* no heavy abstractions
* no dependency chains

This improves:

* auditability
* long-term maintainability
* portability

---

## Module-Specific Design Notes

---

## Rotator (`pkg/rotate`)

### Responsibility

The rotator is responsible for:

* writing to a single file
* deciding **when** to rotate (via policies)
* performing rotation (rename / archive)
* cleaning up old files

It deliberately **does not**:

* manage multiple files
* coordinate with other processes
* log its own success messages

If no error is returned, the operation succeeded.

---

### Single File Ownership

**Design decision:**

> One `rotate.Writer` manages exactly **one file**.

Reasons:

* simpler locking model
* clearer lifecycle
* easier reasoning about failure modes

If multiple files need rotation, the scheduler should create **multiple jobs**, each with its own rotator instance.

---


### Rotator has 2 different Locking strategies:

####  1) In-Process Locking Only (default)

The default part of rotator uses **in-process synchronization** (`sync.Mutex`).
This in-process does **not** implement:

* lock files (`.lock`)
* OS-level advisory locks
* network-wide or distributed locking

**Rationale**

While cross-process locking (e.g. via lock files) is sometimes useful, it introduces:

* platform-specific behavior
* failure modes that are hard to recover from
* implicit coupling between unrelated processes

GSF deliberately avoids this complexity.

*If multiple processes write to the same log file, this is considered a deployment concern, not a library concern.*

**File Descriptor Handling**

The rotator keeps the file *open* during normal operation.
It does *not* use a workflow like:

```
open → write → close (per write)
```

Reasons:

* poor performance
* increased syscall overhead
* higher risk of race conditions under load

Instead:

* the file is opened lazily
* kept open
* closed and reopened only during rotation or explicit reopen


**Reopen Mechanism**

With the `ReopenableWriter` a `Reopen()` operation exists to support scenarios like:

* external log rotation (e.g. `logrotate`)
* signal-based reopen (`SIGHUP`)
* scheduler-triggered maintenance

The rotator itself **does not decide when to reopen**.
This keeps responsibilities clean:

* **Rotator:** file lifecycle
* **Scheduler / Signals:** timing & orchestration

####  2) External locking synchronizationprocess 

This part of rotator was designed for maximum data security and process independence. It follows the principle of "atomic writes" and a strict file locking logic.

**Features:**
- *Lock file synchronization:* Before each write operation, a `.LOCK` file is checked or created. This prevents data corruption if multiple instances or external tools access the logs simultaneously.
- *Stateless Writing (Open-Write-Close):* The log file does not remain open permanently. Each write operation follows the following cycle:
  1. Check the lock (`.LOCK`).
  2. Open file in append-mode.
  3. Write the data.
  4. Close the file and unlock it.
- *Size based rotation:* Once a defined file size is exceeded, the file is rotated. The current file is archived (timestamp suffix) and a new log file is started.
- *Resource conservation:* Closing the file immediately after writing prevents file descriptor leaks and increases compatibility with file system backups (e.g., rsync).

With the `LockWriter` rotator and logger support this scenarios

---

## Scheduler (`pkg/schedule`)

### Responsibility

The scheduler:

* executes jobs **in-process**
* supports recurring and one-shot jobs
* provides panic recovery
* allows job introspection
* supports graceful shutdown

It is **not** a replacement for:

* cron
* distributed schedulers
* persistent job queues

---

### Panic Recovery

Each job execution is wrapped with `recover()`.

A panic in a job:

* does **not** crash the application
* does **not** stop other jobs

This is a deliberate safety feature for long-running services.

---

### No Persistence

The scheduler does **not** persist jobs.

If the process restarts:

* all scheduled jobs are gone
* the application must recreate them

This keeps the scheduler:

* simple
* predictable
* free of storage concerns

---

## Logging Philosophy

GSF modules:

* **do not log by default**
* may accept a minimal logger interface
* never depend on a concrete logging implementation

### Rationale

Logging inside low-level libraries often leads to:

* duplicate log entries
* unclear ownership of log messages
* difficulty testing error paths

Instead:

> Errors are returned.
> Logging happens at the application boundary.

---

## Non-Goals (Explicit)

GSF intentionally does **not** aim to provide:

* distributed locking
* cross-process coordination
* network-level synchronization
* guaranteed exactly-once execution
* persistent scheduling
* cluster-wide log rotation

These problems are valid – but belong to **other layers** of the system.

---

## Design Heritage

Some design decisions are informed by experience with:

* Smalltalk systems
* long-running services
* microprocess and multi-runtime architectures

In other environments, different trade-offs were made (e.g. lock files, shared coordination).

In GSF, the focus is clarity, simplicity, and explicit responsibility boundaries.

---

## Final Note

GSF prefers **boring correctness** over clever abstractions.

If a feature adds significant complexity but only serves rare edge cases, it is likely out of scope.

Simplicity is not a lack of capability –
it is a deliberate constraint.

---

