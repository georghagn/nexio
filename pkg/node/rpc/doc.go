// Copyright 2025 Georg Hagn
// SPDX-License-Identifier: Apache-2.0

/*
Package rpc implementiert ein bidirektionales JSON-RPC 2.0 Protokoll über abstrakte Verbindungen.

Im Gegensatz zu klassischem Client-Server-RPC erlaubt dieses Paket eine symmetrische
Kommunikation (Peer-to-Peer). Jeder Endpunkt kann sowohl Methoden registrieren (Server-Rolle)
als auch Methoden beim Partner aufrufen (Client-Rolle).

Hauptmerkmale:
  - Unterstützung für Call (Request/Response) und Notify (Fire-and-Forget).
  - Robuste Fehlerbehandlung mit standardisierten JSON-RPC Error-Codes.
  - Generische Bind-Funktion für typsicheres Unmarshaling ohne Reflection.
  - Unterstützung für automatische Reconnect-Logik bei Verbindungsabbruch.

Beispiel für die Registrierung eines Handlers:

	peer.Register("sum", func(ctx context.Context, params json.RawMessage) (any, error) {
	    vals, _ := rpc.Bind[[]int](params)
	    return vals[0] + vals[1], nil
	})
*/
package rpc
