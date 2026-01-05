// Copyright 2025 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

/*
Package schedule provides a lightweight, thread-safe in-process job scheduler.

The scheduler is designed to manage recurring tasks and one-off jobs without
external dependencies. It is a core component of the GSF (Go Small Frameworks)
suite, often used to trigger maintenance tasks like log rotation or
session cleanups.

Core Concepts:

The scheduler manages two types of jobs:
  - Recurring Jobs (Every): Tasks that execute repeatedly at a fixed interval.
  - One-shot Jobs (At): Tasks that execute once at a specific point in time.

Concurrency and Safety:

Every job is executed in its own goroutine. To ensure the stability of the
main application, the scheduler provides built-in Panic Recovery. If a
job triggers a panic, the scheduler catches it, logs the error (if a
logger is provided), and continues to manage the remaining jobs.

The scheduler is fully thread-safe, allowing jobs to be added or canceled
from multiple goroutines simultaneously.

Graceful Shutdown:

When an application exits, it is crucial to let running jobs finish their
current execution. The StopAll() method provides a graceful shutdown
mechanism by:
 1. Signalling all job tickers to stop.
 2. Waiting for currently active job executions to complete.

Logging and Observability:

To remain decoupled from specific logging frameworks, the scheduler defines

	a minimal Logger interface (Info, Error). This allows users to inject

any logger, such as the GSF-Logger or standard library loggers.

The List() method provides introspection into the current state of the
scheduler, returning details about all registered jobs, their running
status, and their next scheduled execution time.

Example Usage:

	s := schedule.New()

	// Schedule a recurring task
	id := s.Every(5*time.Minute, func() {
	    // Perform maintenance
	})

	// Schedule a one-off task
	s.At(time.Now().Add(1*time.Hour), func() {
	    // Execute later
	})

	// Clean up on shutdown
	defer s.StopAll()

Limitations:

The scheduler is strictly in-process and does not persist jobs to disk.
If the process restarts, the job list must be repopulated. For critical
tasks requiring persistence or distributed coordination, a more complex
queue-based system should be considered.
*/
package schedule
