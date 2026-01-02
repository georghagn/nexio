|<sub>🇩🇪 [German translation →](README.de.md)</sub>|
|----:|

||[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](./LICENSE) [![Go Report Card](https://goreportcard.com/badge/github.com/georghagn/nexio)](https://goreportcard.com/report/github.com/georghagn/nexio)|
|----|----|
|![GSF-Suite-Logo](logo-suite.png)| ***nexIO Suite***<br>A collection of minimalist, independent Go modules for robust, distributed **small services**. Member of the **Go Small Frameworks (GSF)** family.|

<sup>***GSF*** stands for ***Go Small Frameworks*** — minimalist tools for robust applications.</sup>

### Overview

**nexIO** provides a set of minimal infrastructure building blocks for microservices and service-oriented systems. The focus is on **clarity, low dependencies, and composability**, following a pragmatic *"90% solution"* philosophy.

The modules are intentionally focused:
- Solve one specific problem well
- Independent usage (no forced monolithic structure)
- Prefer explicit composition over "magic" abstractions
- Minimal external dependencies

---

### Quick Start: Create a node in 3 lines
logger := nexlog.Wrap(nexlog.NewDefaultConsole())
provider := transport.NewWSProvider(logger)
node := rpc.NewNode(nil, provider, "ws://localhost:8080/ws", logger)
go node.Listen(ctx)

---

### Design Principles

- **Simplicity first** – small APIs, clear responsibilities.
- **90% solutions** – practical, stable solutions over theoretical perfection.
- **Loose coupling** – modules communicate via clean interfaces.
- **Language-agnostic architecture** – principles suitable for polyglot systems.

---

### The Smalltalk Philosophy

**nexIO** is the result of porting the flexibility and intuitive design of classic Smalltalk environments into modern systems programming with Go.

My long-standing experience with Smalltalk systems (see also my `TSF` projects) deeply influences the architecture of nexIO:

* **Messaging over Procedure Calls:** Inspired by the Smalltalk paradigm "Everything is a Message," nexIO focuses on the free flow of messages between objects rather than rigid client-server hierarchies.
* **Object Symmetry:** In Smalltalk, objects are equal actors. This philosophy is reflected in our **Symmetrical Nodes**, which act as both sender and receiver simultaneously.
* **Decoupling & Composition:** Smalltalk systems excel at combining simple, specialized components into complex systems. nexIO follows this lead with strictly decoupled modules communicating via clean interfaces.

**Why Go?** nexIO builds the bridge: The proven interaction patterns of the Smalltalk world meet the type safety, concurrency (Goroutines), and performance of Go.

---

### Modules

- [**node**](./pkg/node): Resilient P2P RPC communication.
- [**nexlog**](./pkg/nexlog): Structured logging with adapter support.
- [**rotate**](./pkg/nexlog/rotate): Safe file rotation with `.LOCK` synchronization.
- [**schedule**](./pkg/schedule): Reliable task scheduling.


#### nexIOnode (`pkg/node`)
The core of bidirectional communication. It replaces the classic client-server paradigm with a **symmetrical peer architecture**.

* **Symmetry:** Once connected, every node can register methods and call its partner simultaneously.
* **Role Agnostic:** While connections start as Client/Server, once established, all nodes act as equal peers. This is demonstrated in the cmd/node/gsfNodesExamples where a "Payment Service" and multiple "Order Services" interact bidirectionally.
* **Resilience Engine:** Integrated state machine with exponential backoff for transparent reconnection.
* **Type Safety:** Uses Go generics (`Bind[T]`) for secure JSON-RPC parameter handling.

#### nexlog & rotate (`pkg/nexlog` & `pkg/nexlog/rotate`)
A structured logging system optimized for long-term operation.

* **Interface Abstraction:** Decoupled via `LogSink`, allowing usage in any module without hard dependencies.
* **Atomic Rotation:** Robust file rotation with a `.LOCK` mechanism. 
* **Safe Operations:** Each log event follows an **Open -> Write -> Close** cycle, guaranteeing integrity even during system crashes.
* **Contextual Tracing:** Supports field enrichment via `With(key, value)` for distributed tracing.

#### nexIOschedule (`pkg/schedule`)
A precise, panic-safe scheduler for recurring tasks.

* **Interface-Driven:** Execute any Go function through a simple task interface.
* **Concurrency-Safe:** Designed to handle hundreds of parallel jobs.
* **Fault Tolerance:** Failed jobs are logged with full context via the integrated `LogSink`.

---

### Composition Model

nexIO modules are designed for explicit composition:
- `nexlog` writes to an `io.Writer`.
- `rotate.Writer` implements `io.Writer`.
- `schedule` can trigger maintenance tasks like log rotation.

Integration happens in the application layer—**no hard dependencies** between core modules.

---

### Examples

The `cmd/` directory contains self-documenting code:
- `cmd/node/gsfNodesExamples/` – **The Peer-to-Peer Demo**: Interaction of a Payment Server and multiple Order Clients.
- `cmd/rotate/main.go` – Standalone file rotation.
- `cmd/schedule/main.go` – Scheduler usage.

---

### Organizational & Standards

* **Copyright:** © 2026 Georg Hagn.
* **Namespace:** `github.com/georghagn/nexio/pkg/...`
* **License:** Apache License, Version 2.0.

nexIO is an independent open-source project and is not affiliated with any corporation of a similar name.

---

## Contributing & Security

Contributions are welcome! Please use GitHub Issues for bug reports or feature ideas.
**Security-related topics** should not be discussed publicly; please refer to `SECURITY.md`.

