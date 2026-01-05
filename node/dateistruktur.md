Here is an overview of `nexio/node` architecture as a reference for code study.

### Projektstructure

```text
nexio/
├── cmd/node/
│   └── main.go           # Dispatcher: Controls lifecycle & flags
├── pkg/transport/
│   ├── interface.go      # The abstraktion (Connection & Provider)
│   ├── mem_transport.go  # Implementation of Read/Write-Interfaces for test-usage
│   ├── ws_transport.go   # Implementation of Read/Write-Interfaces for production-usage
│   └── ws_provider.go    # Logic for Dial (Client) & Listen (Server)
└── pkg/rpc/
│   ├── doc.go          # a little documentation
│   ├── node_types.go   # global types
│   ├── node.go         # The heart: JSON-RPC dispatcher & request tracking
    └── README.md       

```

---

### Core Components & Responsibilities

| Component | Responsibility | focus |
| --- | --- | --- |
| **`Connection` (Interface)** | `Send`, `Receive`, `Close` | Byte-Transport (blind to content) |
| **`WSProvider`** | Creates Connections via WebSocket | Connections-Management |
| **`Node` (RPC)** | Linking `Connection` with `JSON-RPC` | Logic & Symmetry |
| **`main.go`** | Connects `Provider` with `Peer` | Orchestration & Signals |

---

### The data flow (visualization)

1. **Transport Layer:** The `WSProvider` accepts a connection (server) or establishes one (client).
2. **Dispatcher:** The connection is reported to `main.go` via the `found` channel.
3. **Promotion:** In `handleNewNode`, the "dumb" byte connection (`Connection`) is promoted to the "smart" `rpc.Node`.
4. **Lifecycle:** The `context.Context` ensures that everything from the server to the smallest RPC handler stops simultaneously when you disconnect.

---

### A little tip for reflection:

When reading, pay particular attention to the **`pendingRequests` map** in `node.go`. It's the reason why you can still write something as simple as `result, err := node.Call(...)` in a bidirectional, asynchronous world (WebSockets). That's the "magic" of this design.
