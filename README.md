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

## Roles

Frontends speak the node-side surface — `SubmitListing`, `Heartbeat`,
`PollJobs`, `Decline`, `ReportJob`, `Fees` — on behalf of the machines their
members contribute; coordinators implement the `Coordinator` interface and
coordinate (and federate) frontends. Persons never appear on the wire:
member identity is a frontend concern, and the only identity this module
knows is workload identity — the NodeID with its canonical SPIFFE binding in
`identity/`.

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
vectors/              conformance test: regenerates and checks testdata/vectors.json
testdata/vectors.json cross-language conformance vectors (normative fixture; see SPEC.md §10)
```

## Status (honest)

- **Built:** message types, canonical signing bytes, ed25519 sign/verify, the
  `Coordinator` interface, and a reference HTTP+JSON transport under
  `transport/httpjson`.
- **Conformance vectors:** `testdata/vectors.json` is the normative
  cross-language conformance fixture — 25 primitive cases plus all six signed
  message types with their canonical bytes and ed25519 signatures. A foreign
  implementation is conformant iff it reproduces the fixture byte-for-byte from
  the same inputs (`SPEC.md` §10). The `vectors/` test regenerates the fixture
  from the current encoders and fails the build on any drift.
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

## Successful response limits and migration

The reference client rejects trailing JSON values or junk in `PollJobs` and
`Fees` responses, including with its zero-value configuration. JSON whitespace
and unknown object fields remain accepted. Failed decoding returns no partial
result. This stricter framing is proposed for module release v0.4.0; it does not
change canonical message bytes or the `/v0` endpoints.

To opt into a local success-body byte budget, set `MaxResponseBytes` on
`httpjson.Client`. Zero keeps response size unbounded; negative values fail
before sending. Positive values include whitespace, ignored POST content, and
one detection byte beyond the configured limit. Overflow returns
`httpjson.ErrResponseTooLarge` without exposing the payload. Limits count the
bytes exposed by the HTTP response body, including automatically decompressed
content, and are not exact process-memory ceilings: buffering and decoded Go
objects have overhead. Combine a budget with an HTTP timeout or request deadline.

Before rollout, merge and publish the reviewed protocol as a new immutable
v0.4.0 tag, then finish the Cloudy and SoHoLINK dependency drafts against that
published tag. Fetch real module checksums and run each consumer's required
checks without a local workspace override. Do not move existing tags or commit
a local `replace`. Existing deployments get no automatic byte budget; their
owners choose one for their workloads. Framing failures and body-limit errors
add no retry and do not establish whether an earlier request was applied.

Transport conformance: `go test ./transport/httpjson -count=1` and
`python3 vectors/reproduce_responses.py`. Canonical conformance remains
`go test ./vectors` and `python3 testdata/reproduce_spec.py`.

The consumer drafts upgrade from v0.1.0, whose listing and assignment canonical
encodings differ from current main because of prior GPU/storage additions under
existing `v0` domains. Mixed old/new agents, frontends/relays, and coordinators
reject those signatures even when added fields are empty. Fees alone do not
exercise that boundary. `testdata/legacy_v0_1_signatures.json` was signed and
verified with the actual v0.1.0 module using a public all-zero test seed;
`TestLegacySignaturesRequireCoordinatedMigration` records the candidate's
rejection of both old signatures. This is evidence of incompatibility, not a
migration implementation. Agree and verify the coordinated rollout before
consumer deployment; do not reinterpret persisted signatures in this patch.
