// Copyright 2025 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

/*
Package log provides a lightweight, robust logging framework for the GSF suite.

The logger is designed with a focus on reliability and interoperability,
especially in environments where log files are managed by external processes
like the GSF Rotator. It strictly separates data acquisition (`Logger`),
formatting (`Formatter`), and output (`io.Writer`). It is designed for small
services and infrastructure code where simplicity, explicitness, and low
dependencies matter more than features.

Key Features:

  - Pluggable Appenders: Support for multiple output targets (File, Console, etc.).

  - Thread-Safety: Safe for concurrent use across multiple goroutines.

  - Internal `io.Writer`: The package nexlog offer 3 io.Writer implemntations:
    LockWriter, ReopenableWriter and the default implementation `rotator'

    -- LockWriter: In conjunction with `Lockwriter` as the io.Writer, it offers
    `Inter-Process Compatibility` with an "Open-Write-Close" strategy to ensure
    that files are not locked indefinitely, allowing external tools to rotate
    or archive logs safely. The goal is NOT performance but safety
    (see pkg/nexlog/writer/lockwriter).

    -- ReopenableWriter: Is an io.WriteCloser that can be reopened at runtime.
    This is useful for logrotate strategies (e.g., via scheduler or SIGHUP),
    where the file is moved externally and the process needs to renew the file
    handle (see pkg/nexlog/writer/reopen).

    -- Rotator: It is the default io.Writer implementation for nexlog. Used as
    inline-process it offers the fastest and safest write-strategy
    (see pkg/nexlog/rotate).

  - Log Levels: Standard levels including Debug, Info, Warn, Error, and Fatal.

Architecture:

The Logger acts as the central entry point. It dispatches log entries to one
or more Appenders. Each Appender is responsible for formatting the entry
and writing it to its respective destination.

Usage Example:

Examples can be found in tests and cmd/..example../main.go
*/
package nexlog
