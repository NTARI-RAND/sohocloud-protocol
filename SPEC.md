# sohocloud-protocol — Conformance Specification

Version: **v0 (UNSTABLE)** — message shapes and the canonical byte format may
change without compatibility guarantees until v1.

This document is normative and is a build deliverable, not an afterthought. The
AGPL right to fork the network is empty unless an implementation can be built
from a readable spec without reading this module's source. Anything below marked
MUST / MUST NOT is a conformance requirement; a change that violates one is not a
smaller version of this protocol, it is a different protocol.

---

## 1. Scope

This protocol governs the **substrate coordination layer**: how a node is
recognized, how it advertises what work it will do, and how work is offered to
it and reported back. Three parties speak it:

- **Node** — a participant-owned device offering compute and/or physical print.
  A node owns its identity and signs everything it emits. Sovereign: it chooses
  which coordinators to serve and may leave any of them.
- **Coordinator** — matches listings to work and records outcomes. A pluggable
  role. SoHoLINK is the reference coordinator; a frontend MAY run its own or
  match directly against nodes.
- **Frontend** — where members transact. It consumes coordination through the
  same interface.

**Out of scope, deliberately.** This protocol is NOT the JFA member economy. It
defines no escrow, no member-issued credit, no reputation covenant, no
dialog-sealed record, and no dispute adjudication. Those belong to each frontend.
The fiat settlement a coordinator runs with its nodes (e.g. SoHoLINK's Stripe
payouts) is the *substrate* economy and is likewise outside this spec.

---

## 2. Identity and binding

A node is named by a `NodeID` — a stable string the node chooses for itself.

A node's workload identity is a SPIFFE SVID issued under exactly:

```
/node/<NodeID>
```

A coordinator authorizes a request as coming from a node by comparing the
caller's authenticated SPIFFE path to `/node/<NodeID>` for the `NodeID` named in
the request.

- A coordinator MUST reject with **401** when no SPIFFE identity is present.
- A coordinator MUST reject with **403** when the identity is present but its
  path is not exactly `/node/<NodeID>`.
- A coordinator MUST NOT accept "any valid SVID" as authorization for a request
  naming a different node. That is an impersonation vector.
- Binding MUST be deterministic string construction, never a per-node lookup in
  mutable storage. A check gated on a field that is never populated is inert and
  MUST NOT be relied on as a security boundary.

The protocol does not distribute public keys. A coordinator resolves a node's
verification key out of band (registration, SVID material, or a registry).

---

## 3. Canonical byte encoding (for signatures)

Every signed message is signed over its **canonical bytes**, computed with the
`Signature` field EXCLUDED. Signatures are **ed25519**.

Signing is NOT performed over JSON or any transport encoding. It is performed
over the following deterministic, length-prefixed format so that signatures are
stable across platforms, Go versions, and languages.

Primitives, in the exact order fields are listed for each message in §4:

- **domain tag** — every message's canonical bytes begin with a distinct
  domain-tag string (see §4), encoded as a *string* per below. Distinct tags make
  a signature over one message type non-transferable to another.
- **string** — unsigned LEB128 varint byte-length, then the raw UTF-8 bytes.
- **bytes** — unsigned LEB128 varint byte-length, then the raw bytes. (The
  `Signature` field is never encoded; it is excluded.)
- **bool** — a single byte, `0x00` or `0x01`.
- **int64 / uint64** — fixed 8 bytes, big-endian. `int64` is encoded as its
  two's-complement `uint64`.
- **time** — the instant as `int64` UTC Unix **nanoseconds** (per the int64 rule
  above). Location and monotonic-clock components are dropped. This survives a
  JSON transport round-trip and is time-zone independent.
- **repeated field** — an unsigned LEB128 varint *count*, then each element's
  fields encoded in order, inline.

"Unsigned LEB128 varint" is the encoding produced by Go's
`encoding/binary.PutUvarint`.

A verifier recomputes the canonical bytes from the received fields and checks the
ed25519 signature against the resolved public key. A verifier MUST reject a
signature whose length is not `ed25519.SignatureSize` (64) before verifying.

---

## 4. Messages

For each message: its domain tag, its fields in canonical order, and who signs
it. All timestamps follow the `time` rule in §3.

### 4.1 CapabilityListing — signed by the NODE
Domain tag: `sohocloud/listing/v0`

Order: `NodeID` (string), `Class` (string), `Printers` (repeated: `Kind` string,
`Model` string), `Capacity.VCPUs` (int64), `Capacity.MemMB` (int64),
`Capacity.DiskMB` (int64), `Capacity.PrintQPS` (int64), `OptIn.Compute` (bool),
`OptIn.Print` (bool), `IssuedAt` (time), `Seq` (uint64).

`Class` ∈ {`micro`, `standard`, `server`}. `Kind` ∈ {`traditional`, `threed`}.

`Seq` is strictly monotonic per node. A coordinator MUST reject a listing whose
`Seq` does not strictly exceed the last one seen for that node (replay/rollback
protection).

`OptIn` is **advisory** (see §5.1).

### 4.2 Heartbeat — signed by the NODE
Domain tag: `sohocloud/heartbeat/v0`

Order: `NodeID` (string), `SentAt` (time), `Seq` (uint64).

Same strict-monotonic `Seq` discipline as §4.1.

### 4.3 Assignment — signed by the COORDINATOR
Domain tag: `sohocloud/assignment/v0`

Order: `JobID` (string), `NodeID` (string), `Spec.Workload` (string),
`Spec.Image` (string), `Spec.Args` (repeated string), `Spec.PrinterKind`
(string), `Fee.ContributorShareBps` (int64), `Fee.PlatformFeeBps` (int64),
`OfferedAt` (time).

The fee terms are attached inline so a node sees the split before it commits.
In the pull model there is no separate accept message.

### 4.4 Decline — signed by the NODE
Domain tag: `sohocloud/decline/v0`

Order: `JobID` (string), `NodeID` (string), `Reason` (string), `DeclinedAt`
(time).

`Reason` ∈ {`local_policy`, `capacity`, `unavailable`}. `local_policy` is the
opt-out path (see §5.1).

### 4.5 JobReport — signed by the NODE
Domain tag: `sohocloud/jobreport/v0`

Order: `JobID` (string), `NodeID` (string), `ExitCode` (int64), `FailureCause`
(string), `TmpfsExhausted` (bool), `StartedAt` (time), `FinishedAt` (time).

This signed report is the fact from which metering and payout derive. Its fields
mirror the executor completion parse already live in SoHoLINK.

### 4.6 FeeDeclaration — signed by the COORDINATOR
Domain tag: `sohocloud/fee/v0`

Order: `CoordinatorID` (string), `Terms.ContributorShareBps` (int64),
`Terms.PlatformFeeBps` (int64), `EffectiveAt` (time), `Seq` (uint64).

`ContributorShareBps + PlatformFeeBps` SHOULD equal `10000`.

---

## 5. Invariants

### 5.1 Opt-out is enforced locally, never on the wire
`CapabilityListing.OptIn` and the routing hints in `JobSpec` are **advisory**.
The coordinator is NOT a security boundary for opt-out. A node MUST enforce
opt-out against its own locally-trusted allowlist and MUST refuse work its local
policy forbids — via `Decline{Reason: local_policy}` — regardless of what it
previously advertised or what the coordinator offers.

### 5.2 Matching policy is not part of the protocol
The `Coordinator` interface exposes no ranking or scoring. How a coordinator
chooses which node gets which job is private to the implementation. The protocol
MUST NOT grow a field that dictates matching, because that would weld consumers
to one coordinator's policy and defeat the leaveability the interface exists to
protect.

### 5.3 Fees are legible and non-retroactive
A coordinator's fee terms are published as a signed, timestamped
`FeeDeclaration`. Terms MUST NOT change retroactively for already-offered work;
a change is a new signed declaration with a later `EffectiveAt`. This keeps fees
contestable rather than opaque (open problem #7).

### 5.4 The coordinator is a role, not a hub
Nothing a node or frontend needs may route *exclusively* through one specific
coordinator. A node MAY serve multiple coordinators; a frontend MAY run its own
coordinator or match directly. Any deployment in which a single coordinator is
in fact unavoidable is a recentralization and MUST be labeled as such, with a
path to alternatives.

### 5.5 Monotonic sequence numbers
Node `Seq` values (listings, heartbeats) are strictly monotonic per node. A
coordinator MUST reject non-increasing `Seq`.

---

## 6. Coordinator operations (pull model)

The model is **pull**: the coordinator never calls the node. Operations:

| Operation       | Direction        | Body / result                        |
|-----------------|------------------|--------------------------------------|
| `SubmitListing` | node → coord     | `CapabilityListing`                  |
| `Heartbeat`     | node → coord     | `Heartbeat`                          |
| `PollJobs`      | node → coord     | `NodeID` → `[]Assignment`            |
| `Decline`       | node → coord     | `Decline`                            |
| `ReportJob`     | node → coord     | `JobReport`                          |
| `Fees`          | anyone → coord   | → `FeeDeclaration`                   |

A coordinator MUST apply §2 binding to `PollJobs` (and to any operation naming a
node) before acting. Because the coordinator never initiates contact with a
node, there is no node-side service to specify.

---

## 7. Transport

HTTP+JSON is the **reference** transport (`transport/httpjson`), NOT the wire
itself. It is imported by no core package. A conformant implementation MAY use
any transport by implementing the coordinator operations against the canonical
bytes in §3. JSON is a transport encoding only; signatures are never computed
over JSON, so a different transport interoperates without renegotiating
signatures.

The reference HTTP+JSON endpoints:

```
POST /v0/listing     body: CapabilityListing   -> 204
POST /v0/heartbeat   body: Heartbeat           -> 204
GET  /v0/jobs?node_id=<id>                      -> 200 []Assignment
POST /v0/decline     body: Decline             -> 204
POST /v0/report      body: JobReport           -> 204
GET  /v0/fees                                   -> 200 FeeDeclaration
```

SPIFFE binding (§2) is applied by middleware in front of the handler; the
reference handler wires routes and JSON only and is not the authenticator.

---

## 8. Open problems this protocol touches

Honesty about unsolved gaps is the method, not an embarrassment.

- **#7 Sovereign compute buys mechanical, not economic, exit.** This is the
  layer #7 is about. §5.4 (coordinator as role) and §5.3 (legible fees) push
  against coordinator recentralization; cold-start remains a residual, and
  reputation portability that would ease it is a frontend/governance concern, not
  this protocol's.
- **#6 Computation honesty is out of scope.** Whether a coordinator's match or
  price was fair is not proven here, only made legible (§5.2, §5.3). `anchor/`
  reserves the place for anchoring employment *claims* (the `JobReport` facts) in
  a witnessed log so they are at least auditable. It is a labeled stub, not built.
- **#4 Membership / onboarding.** Because a node names and signs for itself
  (§2), a low-resource device (e.g. a phone) can list itself as a node. Whether
  node onboarding is open or gated is a governance/values decision left open.

## 9. Versioning

`v0` is unstable. The domain tags embed `v0`; a breaking change to a message's
canonical fields bumps its tag. Consumers MUST treat an unknown domain tag as an
unverifiable message.
