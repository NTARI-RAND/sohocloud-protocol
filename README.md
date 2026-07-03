# sohocloud-protocol

The shared Go module defining the **substrate coordination protocol** for the
SoHoLINK / Cloudy network: node recognition, capability listing, and job
employment, spoken by every coordinator and every frontend.

## Scope

This module governs **coordination only**. It is **not** the JFA member
economy — no escrow, no member-issued credit, no reputation covenant, no
dialog-sealed record, no dispute adjudication. Those live in each frontend and
are deliberately absent here.

It is a **dependency leaf**: it imports no other NTARI module and nothing but
the Go standard library. Both SoHoLINK (the reference coordinator) and Cloudy
(a frontend) import *this*; nothing is imported back through it. That property
— enforced by the import graph, not by prose — is what keeps any one
coordinator from becoming a hub the whole network must route through
(open problem #7).

## Layout

```
version.go            protocol version (v0, unstable)
canon/                deterministic length-prefixed signing encoder
identity/             NodeID + the canonical SPIFFE binding predicate
listing/              CapabilityListing (node-signed)
liveness/             Heartbeat (node-signed)
employment/           Assignment (coordinator-signed); Decline, JobReport (node-signed)
fees/                 FeeDeclaration (coordinator-signed)
coordinator/          the Coordinator interface — the pluggable role, no algorithm
anchor/               STUB, not built — witnessed employment-claim layer (#6)
transport/httpjson/   reference transport; imported by no core package
```

## Status (honest)

- **Built:** message types, canonical signing bytes, ed25519 sign/verify, the
  `Coordinator` interface, and a reference HTTP+JSON transport under
  `transport/httpjson`.
- **Reference transport, not the wire:** `transport/httpjson` is optional and is
  imported by no core package. A conformant implementation MAY speak these
  messages over any transport by implementing `coordinator.Coordinator` against
  the canonical bytes documented in `SPEC.md`.
- **Stub, not built:** `anchor/` — the witnessed employment-claim layer
  (open problem #6). Labeled as a stand-in; see its package doc.

## License

AGPL-3.0-or-later. `SPEC.md` is part of the commons: it exists to be read,
reimplemented, and contested. A commons no one can read is a freedom no one can
use.

*Network Theory Applied Research Institute, Inc. — 501(c)(3) — EIN 92-3047136 — info@ntari.org*
