
### The nexIO role paradigm: From connection to symmetry

The examples under `cmd/node/gsfNodesExamples/` (Payment Service & Order Services) demonstrate the fundamental nexIO principle: **The decoupling of connection establishment and communication role.**

#### 1. Initial connection setup (asymmetric)

At startup, we use the classic client-server model to establish the physical connection:

* **Payment Service:** Acts as the initial server (waits for incoming connections).
* **Order Services:** Act as the initial clients (initiate the connection to the payment service).

#### 2.Established communication (symmetrical)

Once the WebSocket handshake is complete, the distinction between client and server disappears.

* **Unified Nodes:** All instances are now simply `nexIOnode` objects.
* **Full-Duplex RPC:** The Payment Service can now call methods on the Order Service (e.g., `UpdateOrderStatus`) just as the Order Service can call the Payment Service (e.g., `ProcessPayment`).

* **Same Logic:** All nodes use the same `listen` loop, the same handler mechanism, and the same resilience roadmap.

#### Example scenario:

Although the `Order-Service` has physically established the connection, the `Payment-Service` can send a `Notify` or a `Call` to the Order-Service at any time, for example, if a payment status has changed in the background. This makes nexIO ideal for event-driven microservices.

---


### nexIO Startup Configuration

**Quickstart Matrix**, illustrating the difference during startup (and the similarity thereafter):

| Feature | Initial Server (e.g., Payment) | Initial Client (e.g., Order) |
| --- | --- | --- |
| *Provider Setup* | `provider.Listen(ctx, addr, foundChan)` | `provider.Dial(ctx, addr)` |
| *Node Creation* | `NewNode(connFromChan, provider, "")` | `NewNode(dialedConn, provider, addr)` |
| *Start Command* | `go node.Listen(ctx)` | `go node.Listen(ctx)` |
| *Subsequent Role* | **Peer** (Symmetric) | **Peer** (Symmetric) |
| *Capabilities* | `Call`, `Notify`, `Register` | `Call`, `Notify`, `Register` |


>**Architectural Note:**
>In this example setup (`paymentservice` and `orderservice`), you can see nexIO in action.
>1. The **Payment Service** opens a port and waits for connections.
>2. The **Order Services** actively connect to it.
>
>**The Special Feature:** Once the connection is established, there is no further difference in the programming. The Payment Service can proactively call the Order Service (e.g., for confirmations), and vice versa. Both use the identical `nexIOnode` logic and the same `nexlog` adapter.

---

### Usage

**Things to look out for in the "complete work"**:

- Log Identity: Thanks to the adapter and `With()`, you should now be able to see exactly who is speaking in the console.

- Parallel Connections: Your payment server will now start a separate `handleClient` goroutine for each incoming client. Check whether the payment server can process both orders simultaneously.

- The Reconnect Show: You can now stop and restart one of the order nodes—the payment server should handle this without issue. It's more interesting to briefly stop the payment server: Both order nodes should start "searching" simultaneously and reconnect immediately as soon as the server comes back online.

**A quick tip for an overview**:

If you open three terminals side by side, you can best see the "heartbeat" of your system:
- Left: Payment server (waiting for :8080)
- Middle: Order (connecting)
- Right: Bee (connecting)
