// Copyright 2025 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

/*
Package rotate provides a thread-safe, modular file rotator.

It is designed to be used as a backend for loggers like zap, zerolog, or the
standard library's log package. The package supports various rotation policies
(size-based) and retention strategies (max files, compression).

Architecture:

The central component is the Writer, which implements io.WriteCloser. It delegates
decisions to three strategies:
  - RotationPolicy: When to rotate.
  - ArchiveStrategy: How to rename/compress the old file.
  - RetentionPolicy: Which old files to keep or delete.

The global function Write() is designed for inbound use like a "normal"
io.WriteCloser. This is the safest and fastest scenario in conjunction with
the Logger. The global function WriteNow() is designed to be used as external
"Rotator". The subpackage lock provides atomic file-system based locking.
Unfortunately, this scenario comes at the cost of performance loss.

Concurrency & Safety:

The Writer is thread-safe for use within a single process via sync.Mutex.
For coordination between multiple OS processes (e.g., a Logger and a
separate Scheduler process), the sub-package "lock" provides atomic
file-system based locking.

Operating System Considerations:

The "Atomic (Safe)" mode (Open-Write-Close) is recommended for maximum
portability. On Windows, it prevents "file in use" errors during rotation.
On Unix-like systems (Linux, macOS, iOS), it ensures the logger always
writes to the current file name instead of a stale file descriptor
pointing to an archived file.

For high-performance scenarios where the overhead of opening/closing files
is too high, consider using the "Normal" writer with the inbound function Write().
keeping in mind the platform-specific locking behaviors.

# Examples

Typical integrations:

	Logging via `nexlog`
	Scheduled rotation via `schedule`
	Custom triggers

A runnable example can be found at `cmd/rotate-example/main.go`.
*/
package rotate
