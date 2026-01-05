
<sub>🇩🇪 [German translation →](README.de.md)</sub>

---

## Overview

The **GSF-Suite/Scheduler** is a lightweight, robust **in-process job scheduler** for Go.

It was designed to execute recurring tasks or one-off jobs without the need for external dependencies (like cron daemons).

Unlike a simple `time.Ticker`, this package provides **panic recovery**, **job management (start/stop)**, and **graceful shutdown**.

## Features

* **Simple API:** Intuitive methods like `Every` and `At`.
* **Panic Recovery:** If a job crashes (panic), the scheduler catches the error. Your main application remains unaffected.
* **Thread-Safe:** Safe access to the job list from multiple goroutines.
* **Graceful Shutdown:** `StopAll()` waits for running jobs to finish before the program exits.
* **Introspection:** Query runtime statistics (`NextRun`, `Interval`) via `List()` – ideal for status dashboards or RPC.
* **Logging:** Optional logging support.

## Installation

```bash
go get github.com/georghagn/nexio/schedule

```

## Quick Start

```go
package main

import (
    "fmt"
    "time"
    "github.com/georghagn/nexio/schedule"
)

func main() {
    // 1. Create scheduler
    sched := schedule.New()

    // 2. Start a recurring job (Every 500ms)
    jobID := sched.Every(500*time.Millisecond, func() {
        fmt.Println("Tick...")
    })

    // 3. Schedule a one-shot job (in 2 seconds)
    sched.At(time.Now().Add(2*time.Second), func() {
        fmt.Println("Boom! (One-shot)")
    })

    // Let it run briefly
    time.Sleep(3 * time.Second)

    // 4. Stop a job
    sched.Cancel(jobID)
    fmt.Println("Ticker stopped.")
}

```

## Core Concepts

### Recurring Jobs (`Every`)

Executes a function at a fixed interval. The task runs in its own goroutine.

```go
id := sched.Every(1*time.Minute, func() {
    // DB backup logic
})

```

### One-Shot Jobs (`At`)

Executes a function once at a specific point in time.

```go
targetTime := time.Now().Add(10 * time.Minute)
sched.At(targetTime, func() {
    // Send reminder email
})

```

### Logging

By default, no logger is used. Optionally, a logger implementation can be injected.

The scheduler defines a minimal logger interface:

* `Info(format, ...args)`
* `Error(format, ...args)`

This keeps the module independent of specific logging frameworks.

### Panic Recovery (Crash Protection)

A common problem with DIY `go func()` solutions: If the code inside the goroutine triggers a panic, the **entire program** crashes.

`gsf/schedule` wraps every job in a `recover()` function.

```go
sched.Every(1*time.Second, func() {
    panic("database gone!") // Does NOT crash the app
})
// Output on Stdout: "SCHEDULER PANIC in Job 1: database gone!"
// The scheduler and other jobs continue running.

```

## Management & Introspection

### Stopping Jobs

Every call to `Every` or `At` returns a `JobID` (int64). This allows you to cancel specific jobs.

```go
err := sched.Cancel(jobID)
if err != nil {
    log.Println("Job already finished or not found")
}

```

### Listing Jobs (`List`)

The `List()` feature is powerful for monitoring what is currently happening via RPC (e.g., with `pkg/nexio`) or in an admin panel.

```go
jobs := sched.List()
for _, job := range jobs {
    fmt.Printf("ID: %d, Running: %v, Next Run: %v\n", 
        job.ID, job.IsRunning, job.NextRun)
}

```

### Graceful Shutdown

When an application exits (e.g., SIGTERM), you should avoid interrupting jobs mid-execution.

```go
// ... receive signal ...
sched.StopAll() // 1. Sends stop signal to all jobs
                // 2. Blocks until all currently executing jobs are finished

```

## Examples

Typical use cases:

* Triggering file rotations
* Periodic cleanup jobs
* Scheduled maintenance tasks

A working example is available at `cmd/schedule-example/main.go`.

## Limitations (Design Philosophy)

* **In-Process:** Jobs live in RAM. If the app restarts, all dynamically scheduled jobs are lost (unless you reload them on startup).
* **Non-Persistent:** There is no built-in database. For critical jobs that must survive a restart, use an external queue or DB.
* **No "Distributed Lock":** If you scale your app 10x (e.g., in Kubernetes), the scheduler will run 10x independently.

---

## License / Contact

LICENSE, CONTRIBUTE.md, SECURITY.md, and contact information can be found in the root of the suite.


