# Conformance — Janus-Facing Architecture

This document is the module's self-description in the architecture's own terms, stated **before** anything product- or deployment-specific, per the architecture's ordering rule. It follows the architecture's legibility discipline: a conformance claim is bound to the mechanism and the check that enforces it, or it is labeled a stand-in. Unbound prose is marketing.

The architecture is **Janus-Facing Architecture (JFA)** — NTARI's unified architecture document, free documentation under the project's AGPL-3.0 commons. It names roles and mechanisms, never products; this module declares which role it fills and shows its bindings.

## Role declaration

This module is the **coordination protocol** of a JFA **substrate**: the dependency-leaf wire language by which participant-owned machines are recognized, advertise capability, and take and report work. The substrate guarantee it serves: **coordination on infrastructure participants can own — no unremovable hosting chokepoint.**

| This module's term | Architecture role |
|---|---|
| Node | participant-owned infrastructure — the machines members contribute |
| Coordinator | the coordinator: node-side orchestration, a pluggable role |
| Frontend | a front end: the member-facing application |
| FeeDeclaration | the coordinator's authored, legible, contestable fee declaration |
| JobReport | a work claim (witnessed anchoring reserved — see stand-ins) |

## Invariants and their bindings

| Invariant (architecture) | Mechanism here | Check |
|---|---|---|
| The coordination protocol is a **dependency leaf** | `go.mod` declares zero requirements; standard library only | CI: `.github/workflows/dependency-leaf.yml` fails the build on any added dependency; `TestInvariant_P_lean_code_StdlibOnlyCopyleft` (see below) |
| **Persons never appear on the wire** | The only identity this module knows is workload identity — `identity/` NodeID with its canonical SPIFFE binding; no message carries a name, contact, or person-shaped field | SPEC (normative); the type surface; golden vectors |
| **No unremovable hosting chokepoint** | The coordinator is a pluggable role: a frontend MAY run its own coordinator or match directly against nodes (SPEC §1) | SPEC normative text; two consumers exist against the published tag |
| **Fee legibility** | Fees exist on the wire only as a coordinator-signed `FeeDeclaration` — authored, legible, contestable | `fees/`; golden vectors |
| **Legibility is an output, not a comment** | `SPEC.md` is normative and a build deliverable: a foreign implementer builds from it without reading this source | `testdata/vectors.json` (ground truth) + `testdata/reproduce_spec.py` (independent reproducer) + `vectors/` golden test — drift between code and committed vectors fails the build |
| **Canonical, deterministic encoding** | `canon/` length-prefixed, domain-separated signing encoder | golden vectors regenerated-and-compared byte-for-byte on every test run |
| **Representable timestamps** (SPEC §3) | Every timestamp-bearing message verifier checks `canon.ValidTime` before checking the signature; a signature over incomplete canonical bytes cannot authorize an invalid timestamp | `TestVerifyRejectsUnrepresentableTime`, `TestVerifyRepresentableTimeBoundaries`, and `FuzzMessageVerification` |

## JFA registry bindings

Tests that bind an ID from the Janus invariant registry (`jfa-conformance-suite.py` in NTARI-RAND/Janus) are named `TestInvariant_<ID>_<Name>`, with any `-` in the ID written `_`. `go test -run TestInvariant_ -v ./...` lists them; `jfa-conformance-suite.py --bindings ../sohocloud-protocol` counts them.

| Registry ID | Test | What it checks |
|---|---|---|
| **P-lean-code** — protocol software depends on nothing but its language's standard library, is auditable whole, and is copyleft | `TestInvariant_P_lean_code_StdlibOnlyCopyleft` (`conformance_test.go`) | `go.mod` has no require/replace/exclude/retract, and there is no `go.sum` or `vendor/`; every Go file, tests included, imports only the standard library or this module, and none uses cgo; the Python reproducer imports only the Python standard library; `LICENSE` is the GNU AGPL v3 and no file declares another SPDX license; CI still runs `go test ./...`, the reproducer, and the dependency-leaf workflow |
| **SUB-no-inbound-requirement** — a node contributes without open inbound ports or a static address | `TestInvariant_SUB_no_inbound_requirement_NodeDialsOut` (`no_inbound_test.go`) | SPEC §6 still states the pull model and lists every operation as node → coord; the coordinator interfaces are pinned and no new interface appears; no wire package imports a networking package or has a field name or json tag naming an address, host, port, endpoint, URL, IP, callback or webhook (plurals included); nothing in the module opens a listening socket or serves on an inherited one, and no file dot-imports a networking package; over the reference transport a node with no listener completes listing, heartbeat, fees, poll, decline and report on a fresh outbound connection per call, across an address change. Residual, named: Cloudy's node agent does not yet run this loop; a real NAT/CGNAT path is a multi-node (Phase 2) test; address synonyms are left to review; storage and key-lifecycle operations are covered by the interface pin only |

## Deliberately out of scope

This module is **not** the member economy, and the absence is an invariant, not a gap: no escrow, no member-issued credit, no reputation covenant, no dialog-sealed record, no dispute adjudication. Those belong to the layers above (record, covenant, economy), live in front ends, and MUST NOT migrate into this leaf. Single-participant-identity is likewise a front-end obligation — this layer knows no persons at all.

## Stand-ins and open residuals

Named per the architecture's honesty rule: a worked example that concealed where it falls short would only teach concealment.

- **`anchor/` is a stub, labeled as such in its package doc.** Witnessed, append-only anchoring of employment claims (job in, result out, meter) is reserved, not built. Until it exists, a coordinator's account of completed work is self-attested — the architecture's open problem 6, coupling to open problem 2.
- **v0 is UNSTABLE.** Wire shapes may change without compatibility guarantees until v1; consumers pin published tags.
- **Sovereign compute buys mechanical, not economic, exit** (open problem 7). This protocol caps how concentrated hosting can become; it cannot fund anyone's exit. Named, not solved.

## Dependency declaration

Depends on: the Go standard library. Nothing else — no other NTARI module, no third-party module, nothing above or beside it in the stack. The CI check makes a violation a failed build rather than a review catch.

## Product-specific notes (last, per the ordering rule)

SoHoLINK is the reference coordinator; Cloudy is a reference frontend. Both consume this module by published version tag. Neither is privileged by the protocol: any conformant implementation may replace either.
