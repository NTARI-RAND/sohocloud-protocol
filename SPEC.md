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

- **unsigned LEB128 varint** — the length/count primitive used by `string`,
  `bytes`, and `repeated`. It MUST be encoded as follows, language-independently:
  take the unsigned value; emit 7 bits per byte, least-significant group first
  (little-endian groups); set the high bit (`0x80`) of every byte EXCEPT the last
  to mark continuation; the last byte has its high bit clear. The value `0`
  encodes as the single byte `0x00`. This is exactly the encoding produced by
  Go's `encoding/binary.PutUvarint`; the Go symbol is a cross-check, not the
  definition. (Worked: `300` = `0b100101100` → low 7 bits `0101100` with
  continuation → `0xAC`, high bits `0000010` last → `0x02`, giving `AC 02`.)
- **domain tag** — every message's canonical bytes begin with a distinct
  domain-tag string (see §4). The domain tag MUST be encoded with the **string**
  rule below — i.e. an unsigned LEB128 varint of its UTF-8 byte length followed by
  its UTF-8 bytes — exactly like any other string field, with no null terminator
  and no special casing. Distinct tags make a signature over one message type
  non-transferable to another.
- **string** — unsigned LEB128 varint byte-length, then the raw UTF-8 bytes. An
  empty string MUST encode as the single byte `0x00` (a zero length and no
  payload).
- **bytes** — unsigned LEB128 varint byte-length, then the raw bytes. (The
  `Signature` field is never encoded; it is excluded.)
- **bool** — a single byte, exactly `0x00` (false) or `0x01` (true); no other
  byte value is valid.
- **int64 / uint64** — fixed 8 bytes, big-endian (most-significant byte first).
  `int64` MUST be encoded as its two's-complement `uint64` reinterpretation: add
  2⁶⁴ to a negative value, then emit the 8 big-endian bytes. Thus `-1` →
  `ff ff ff ff ff ff ff ff`, `-2` → `ff ff ff ff ff ff ff fe`, and the minimum
  `int64` (`-9223372036854775808`) → `80 00 00 00 00 00 00 00`. No sign-magnitude,
  no zigzag, no variable length.
- **time** — the instant as `int64` UTC Unix **nanoseconds** (per the int64 rule
  above, big-endian, 8 bytes). Location and monotonic-clock components are
  dropped. Nanosecond precision is preserved exactly; the value MUST NOT be
  truncated to micro- or milliseconds. This survives a JSON transport round-trip
  and is time-zone independent. An instant outside the representable range of
  `int64` UTC Unix nanoseconds (approximately the years 1678–2262) is out of scope
  for v0 and MUST be treated as an encoding error rather than wrapped or clamped.
  The reference encoder enforces this: `canon.Buffer.Time` appends nothing and
  latches `canon.ErrTimeOutOfRange` (retrievable via `Buffer.Err()`) instead of
  emitting a wrapped value, so two instants ~584 years apart can never alias to
  identical bytes. A producer MUST reject via `canon.ValidTime` before signing;
  a verifier/decoder MUST reject a message whose timestamp fails it.
- **repeated field** — an unsigned LEB128 varint *count* (the number of elements)
  encoded FIRST, then each element's fields encoded in order, fully inline, using
  the normal primitive rules above. A repeated element that is itself a string is
  therefore length-prefixed like any other string; nested struct fields are
  emitted in the order given for that element in §4. There is no separator between
  elements and no terminator after the last one. A count of `0` is the single
  byte `0x00` followed by no elements.

A verifier recomputes the canonical bytes from the received fields and checks the
ed25519 signature against the resolved public key. A verifier MUST reject a
signature whose length is not `ed25519.SignatureSize` (64) before verifying.

The canonical bytes of a message are **exactly** the concatenation of the domain
tag followed by every field listed in that message's `Order:` clause in §4, in
that order, each encoded by its primitive rule above. There is no outer framing:
no leading version byte, no total-length prefix wrapping the message, and no
trailing bytes after the last field. The `Order:` list in §4 is exhaustive and
authoritative — the message's **signature field(s)** are the ONLY fields
excluded, and no field exists in the canonical bytes that is not named there.
For the twelve core messages of §4 that single excluded field is `Signature`. The
Layer-C operator messages of §11 carry **two** signatures (`Sig0` and `Sig1`)
over the same canonical bytes; both are excluded exactly as `Signature` is here,
and, like `Signature`, neither appears in an `Order:` clause. A verifier
therefore never encodes any signature field into the bytes it recomputes,
whether the message has one signature or two.

---

## 4. Messages

For each message: its domain tag, its fields in canonical order, and who signs
it. All timestamps follow the `time` rule in §3.

The `Order:` clause below is the SOLE authority for canonical field order,
including the order of nested struct fields (e.g. `Capacity.VCPUs` before
`Capacity.MemMB`, and each printer as `Kind` then `Model`). An implementer MUST
encode fields in the `Order:` sequence and MUST NOT infer order from any other
source. In particular, the sample `testdata/vectors.json` (§10) lists each
message's `fields` in **alphabetical key order for readability, which is NOT the
canonical encoding order** — following the JSON key order would produce wrong
bytes and invalid signatures.

### 4.1 CapabilityListing — signed by the NODE
Domain tag: `sohocloud/listing/v0`

Order: `NodeID` (string), `Class` (string), `Printers` (repeated: `Kind` string,
`Model` string), `GPUs` (repeated: `API` string, `Model` string, `VRAMMB`
int64), `Capacity.VCPUs` (int64), `Capacity.MemMB` (int64),
`Capacity.DiskMB` (int64), `Capacity.StorageCommitMB` (int64),
`Capacity.PrintQPS` (int64), `OptIn.Compute` (bool), `OptIn.Print` (bool),
`OptIn.Storage` (bool), `IssuedAt` (time), `Seq` (uint64).

`Class` ∈ {`micro`, `standard`, `server`}. `Kind` ∈ {`traditional`, `threed`}.
`API` ∈ {`vulkan`, `nnapi`, `cuda`, `metal`}.

`Capacity.DiskMB` is scratch space available to a running job;
`Capacity.StorageCommitMB` is long-lived storage the node commits to hold for
the network (shard hosting). The two are deliberately distinct so a node can
offer either without the other. How stored data is encrypted, sharded, and
audited is a frontend/agent concern outside this protocol — coordination only,
the wire never carries stored content.

Advertising a GPU is the opt-in for GPU work: a listing with no `GPUs`
receives none, and a node withdraws a GPU by omitting it from its next
listing.

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
(string), `Spec.GPUAPI` (string), `Spec.GPUMinVRAMMB` (int64),
`Fee.ContributorShareBps` (int64), `Fee.PlatformFeeBps` (int64),
`OfferedAt` (time).

`Spec.GPUAPI` ∈ {``, `vulkan`, `nnapi`, `cuda`, `metal`} — an advisory routing
hint like `PrinterKind`; empty means no GPU is required and `GPUMinVRAMMB`
MUST then be `0`.

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

### 4.7 StorageLease — signed by the COORDINATOR
Domain tag: `sohocloud/lease/v0`

Order: `LeaseID` (string), `NodeID` (string), `ShardRef` (bytes),
`SizeClass` (int64), `Fee.ContributorShareBps` (int64),
`Fee.PlatformFeeBps` (int64), `IssuedAt` (time), `ExpiresAt` (time),
`Seq` (uint64).

Storage is a LEASE, not a job: an ongoing obligation to hold one sealed
shard for a bounded term. `ShardRef` is an opaque 32-byte content address
and `SizeClass` a quantized payload size — the protocol MUST NOT carry true
object sizes or stored content; encryption, padding, and sharding are
frontend concerns above this waist. Renewal is a new `StorageLease` for the
same `LeaseID` with strictly higher `Seq` (same rollback rule as §4.1). Fee
terms ride inline, as in §4.3.

### 4.8 LeaseDecline — signed by the NODE
Domain tag: `sohocloud/lease-decline/v0`

Order: `LeaseID` (string), `NodeID` (string), `Reason` (string),
`DeclinedAt` (time).

`Reason` ∈ {`local_policy`, `capacity`, `unavailable`}; `local_policy` is the
opt-out path (§5.1) for storage exactly as §4.4 is for jobs.

### 4.9 LeaseRelease — signed by the NODE
Domain tag: `sohocloud/lease-release/v0`

Order: `LeaseID` (string), `NodeID` (string), `ReleasedAt` (time).

A node may always stop holding (sovereignty includes leaving). The signed
release ends its metering; re-placing the shard is the frontend's concern.

### 4.10 ProofChallenge / ProofResponse — response signed by the NODE
Domain tag (response): `sohocloud/proof/v0`

A `ProofChallenge{LeaseID, Offset, Length, Nonce (16 bytes), IssuedAt}` is
NOT signed: nodes fetch challenges by polling over the authenticated channel
(pull model), and a challenge commits no one to anything. The signed
artifact is the response.

Response Order: `LeaseID` (string), `NodeID` (string), `Offset` (int64),
`Length` (int64), `Nonce` (bytes), `Digest` (bytes), `RespondedAt` (time).

`Digest` MUST be computed exactly as
`SHA-256(Nonce || uint64be(Offset) || uint64be(Length) || sealed[Offset : Offset+Length])`
over the sealed shard bytes. The response restates the challenged range and
nonce so it stands alone as a non-repudiable metering fact — the storage
counterpart of §4.5. A verifier MUST reject a response whose
`(LeaseID, Nonce)` it has already accepted: nonces are single-use, which is
what defeats replaying a recorded answer.

### 4.11 KeyRotation — signed by the node's CURRENT key
Domain tag: `sohocloud/node-rotate/v0`

Order: `NodeID` (string), `NewPublicKey` (bytes), `Algo` (string),
`RotatedAt` (time), `Seq` (uint64).

Possession of the outgoing key authorizes naming its successor. A verifier
MUST validate the signature against the key it currently holds for the node
(a rotation signed by the successor is self-installation and MUST fail),
MUST enforce strictly monotonic `Seq` per node, and MUST reject a rotation
whose `NewPublicKey` is not a well-formed key for `Algo` (for `ed25519`, not
exactly 32 bytes) or whose `Algo` is unknown — otherwise a verifier that
begins checking subsequent messages against a malformed `NewPublicKey` would
fault. Only after all of these pass does the verifier begin verifying
subsequent node messages against `NewPublicKey`. `Algo` ∈ {`ed25519`} in v0.

### 4.12 KeyRevocation — signed by the key being REVOKED
Domain tag: `sohocloud/node-revoke/v0`

Order: `NodeID` (string), `RevokedPublicKey` (bytes), `RevokedAt` (time),
`Seq` (uint64).

Kills a key with no successor. A verifier MUST verify the signature against
the key it **currently trusts** for `NodeID`, and MUST require that
`RevokedPublicKey` equals that trusted key. A revocation is only ever honored
against the exact key it names and kills. This is the anti-abuse binding: a
stranger can self-sign a revocation carrying a victim's `NodeID` and the
stranger's own key in `RevokedPublicKey`, but because that key is not the one
the coordinator trusts for the victim, the equality check fails and the
revocation is rejected — a node cannot be knocked offline by anyone who does
not already hold its current key. (Earlier drafts said "honor unconditionally";
that wording is superseded — the binding to the trusted key is mandatory.)
Given the current key, the safe reading of a revocation is "stop trusting this
key". Re-enrollment after revocation is out-of-band, via the same path as
first enrollment, never a wire message a key thief could forge.

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

---

## 10. Test vectors

`testdata/vectors.json` is the **normative conformance fixture** for this spec.
An independent implementation is conformant with respect to canonical encoding
and signing if and only if it reproduces, byte-for-byte, every value in that
file from the same inputs. The fixture is generated by the reference Go encoders
and is cross-checked by an independent stdlib-only reproduction; a conforming
foreign implementation MUST match it exactly. Where prose in §3/§4 and the
fixture ever disagree, that is a spec defect to be fixed, not a license to
diverge — but the fixture is the executable statement of §3/§4.

All hex in the fixture is lowercase. It is self-describing: a top-level `note`
and an `encoding` block restate the §3 primitive rules for a reader who opens the
JSON first.

### 10.1 Structure

- **`primitives`** — isolated encodings of each §3 primitive, so an implementer
  can validate the encoder bottom-up before assembling messages. Each entry has a
  `kind` (`uvarint`, `int64`, `uint64`, `bool`, `string`, `time`), a
  human-readable `input`, and the expected `bytes_hex`. These cover the boundary
  cases that pin the format: varint values across each continuation-byte
  threshold (`0, 1, 127, 128, 255, 256, 300, 16384, 1073741824`); `int64`
  including `-1`, `-2`, `MaxInt64`, and `MinInt64`; `uint64` including
  `MaxUint64`; both bools; the empty string and a multibyte UTF-8 string; and
  `time` at the Unix epoch and at a fixed nanosecond instant.

- **`messages`** — one entry per signed message type (every signed message of §4). Each entry
  records: `name`, `domain_tag`, `signer` (`node` or `coordinator`), the 32-byte
  ed25519 `seed_hex` and the derived `public_key_hex`, the message `fields`, the
  `canonical_bytes_hex` (the output of `CanonicalBytes()`, `Signature` excluded),
  and the `signature_hex` (ed25519 over exactly those canonical bytes).

- **`operator_messages`** — one entry per Layer-C operator message type (all
  three of §11). Because operator messages are 2-of-7, each entry differs from a
  core `messages` entry: `signer` is always `operator`; a `keys` array lists all
  seven registered public keys (`index`, `seed_hex`, `public_key_hex`, `algo` —
  the seeds are present ONLY so the fixture is self-contained and reproducible,
  never because the network holds private keys); `idx0`/`idx1` are the two
  signing indices; and instead of one `signature_hex` there are `sig0_hex` and
  `sig1_hex`, BOTH taken over the SAME `canonical_bytes_hex`. A conformant
  implementation reproduces `canonical_bytes_hex` from the §11 field order and
  verifies both signatures at their indices.

  > **Field order trap.** Each message's `fields` object lists keys in
  > **alphabetical order for readability, NOT canonical order.** Canonical order
  > comes ONLY from the `Order:` clause of that message in §4. An implementer who
  > encodes fields in the JSON key order will produce wrong bytes and signatures
  > that do not verify.

### 10.2 How to use the fixture

1. Reproduce every `primitives[].bytes_hex` from its `input` using §3.
2. For each message, derive the keypair from `seed_hex` (ed25519 from a 32-byte
   seed is deterministic), confirm it yields `public_key_hex`, encode the fields
   in §4 `Order:` and confirm you get `canonical_bytes_hex`, then verify
   `signature_hex` against `public_key_hex` over your independently derived bytes.
   Signature verification over your OWN bytes is the strongest cross-check: it can
   only pass if your bytes are identical to the reference generator's.

The reference test lives in `vectors/vectors_test.go`. It regenerates every
vector from the current code and asserts byte-for-byte equality with the
committed JSON, so a drift between code and fixture fails the build; it also
re-parses the committed file and verifies every signature. Regenerate
intentionally with `go test ./vectors -update` (or the
`SOHOCLOUD_UPDATE_VECTORS` env var). All vectors use fixed byte-fill seeds and
hardcoded timestamps — there is no `crypto/rand` and no `time.Now()`, so the
output is fully deterministic.

### 10.3 Worked example: a `Heartbeat`, byte by byte

This is the `Heartbeat` vector (`§4.2`) decoded field by field, so a
reimplementer can trace the whole encoding start to finish. Inputs:
`NodeID = "node-alpha"`, `SentAt = ` Unix nanoseconds `1700000050500000000`,
`Seq = 43`.

Canonical bytes (`canonical_bytes_hex`):

```
16 736f686f636c6f75642f6865617274626561742f7630 0a 6e6f64652d616c706861 17979d09f832d900 000000000000002b
```

Decoded left to right:

| Bytes | Primitive | Meaning |
|-------|-----------|---------|
| `16` | uvarint length | Domain-tag length = `0x16` = **22** bytes. |
| `73 6f 68 6f 63 6c 6f 75 64 2f 68 65 61 72 74 62 65 61 74 2f 76 30` | UTF-8 bytes | The domain tag `sohocloud/heartbeat/v0` (22 bytes). Together with the length prefix, these two rows are the `string` encoding of the domain tag (§3). |
| `0a` | uvarint length | `NodeID` length = `0x0a` = **10** bytes. |
| `6e 6f 64 65 2d 61 6c 70 68 61` | UTF-8 bytes | `NodeID` = `node-alpha` (10 bytes). |
| `17 97 9d 09 f8 32 d9 00` | int64, big-endian | `SentAt` as UTC Unix nanoseconds. `0x17979d09f832d900` = `1700000050500000000`. This is the `time` rule (§3): the instant reduced to an `int64` and emitted as 8 big-endian bytes. |
| `00 00 00 00 00 00 00 2b` | uint64, big-endian | `Seq` = `0x2b` = **43**. Fixed 8 bytes, no varint. |

The concatenation of exactly these six groups — domain tag (length + bytes),
`NodeID` (length + bytes), `SentAt`, `Seq` — is the complete canonical byte
string. There is no framing before the domain tag and no bytes after `Seq`. The
ed25519 signature (`signature_hex`) is taken over precisely this string; the
`Signature` field itself is never part of it.

Note how the field order here (`NodeID`, `SentAt`, `Seq`) follows §4.2 `Order:`,
which is also — by coincidence for this small message — how a reader would list
them; for larger messages such as `CapabilityListing`, the §4 `Order:` and the
alphabetical `fields` keys in `vectors.json` differ, and only the §4 order is
canonical.

---

## 11. Operator identity (Layer C)

§1–§10 govern **node** and **coordinator** messages. This section adds a third,
orthogonal identity: the **operator** — a frontend (e.g. Cloudy) that terminates
member and node identity inside itself and presents a single rotating credential
to a coordinator. It is deliberately separate from node workload identity (§2):
the core coordination messages (§4) are UNCHANGED and are NOT signed with an
operator credential.

An operator holds **seven** Ed25519 keypairs (indices `0..6`) and signs each
message with **two distinct** indices (a **2-of-7** discipline). The network
registers only the seven **public** keys per operator; **no operator private key
is ever held by the network**, and every verify path below takes public keys
only.

- **2-of-7 is anti-substitution / rotation hygiene, NOT threshold security.** If
  all seven private keys live in one process, a single host compromise yields all
  seven. 2-of-7 is not relied on to survive single-key compromise unless custody
  is split across trust domains. This is stated so no reader over-trusts it.
- **Algorithm binding.** Every operator message binds an algorithm string `Algo`
  into its canonical bytes. `v0` permits exactly `ed25519`. A verifier MUST
  reject a message whose `Algo` is unknown, and MUST reject a message whose
  `Algo` does not equal the registered `algo` of the key at each signing index.
  An operator's active key set MUST NOT mix algorithms. The signature length gate
  is per-algorithm (`ed25519` = 64 bytes), not a hardcoded constant, so the
  planned whole-set atomic rotation to a post-quantum algorithm (a NEW domain
  tag, e.g. `sohocloud/operator/v1-mldsa`) reuses the same rule. A signature is
  non-transferable across algorithms.
- **Migration seam.** The reference implementation routes operator signing and
  verification through a `Signer`/`Verifier` interface, implemented with stdlib
  Ed25519 today, precisely so the algorithm can be swapped by adding a `v1`
  domain tag later without touching the core messages. This is an
  implementation note; the wire contract is the field orders and tags below.

**Encoding of operator fields (no special casing).** Every field in the §11
`Order:` clauses is encoded by its ordinary §3 primitive rule; §11 introduces no
new or fixed-width types. In particular:

- `OperatorID` and `Algo` are ordinary §3 **strings** — an unsigned LEB128
  varint of the UTF-8 byte length followed by the UTF-8 bytes. `Algo` is NOT a
  bare token or a fixed-width enum tag despite `v0` permitting exactly the one
  value `ed25519`; it MUST carry its length prefix like any other string (e.g.
  `ed25519` encodes as `07 65 64 32 35 35 31 39`). `Nonce`, `NewPublicKey`, and
  `Challenge` are ordinary §3 **bytes** (LEB128 length prefix then raw bytes).
- `Idx0`, `Idx1`, and `KeyIndex` are each encoded as a **fixed 8-byte
  big-endian `uint64`** per the §3 `uint64` rule, exactly like `Seq` — NOT as a
  varint and NOT as a single byte — even though their values lie in the small
  range `0..6`. An implementer who encodes an index as a varint or one byte
  produces wrong bytes and an invalid signature. (Example: index `2` encodes as
  `00 00 00 00 00 00 00 02`.)
- The two signatures `Sig0` and `Sig1` are **excluded** from the canonical bytes
  of every §11 message, exactly as `Signature` is for the §4 messages (§3). They
  are the only excluded fields and, like `Signature`, appear in no `Order:`
  clause below. There is no third signature field. `Idx0`/`Idx1` (which name the
  signing keys) ARE inside the signed bytes; the signature *bytes* are not.

### 11.0 Verification requirements (all operator messages)

A verifier MUST, before honoring any operator message:

1. Reject if the message's nonce is absent or shorter than **16 bytes** (where the
   message has a nonce). An empty nonce MUST NEVER be treated as "skip the replay
   cache."
2. Reject if `Algo` is not a supported algorithm for this version.
3. Reject if the two signing indices `Idx0` and `Idx1` are equal.
4. Reject if either index is outside `0..6`.
5. Reject if no active registered key exists at either index, or if either
   registered key's `algo` does not equal the message's `Algo`.
5a. Reject if the registered public keys at `Idx0` and `Idx1` are byte-identical,
   even though the indices differ. The seven registered public keys of an
   operator MUST be pairwise distinct; a registration or rotation MUST NOT
   install a public key that already exists at another index (see §11.2). This
   is enforced so the 2-of-7 anti-substitution property cannot silently degrade
   to 1-of-1: if two indices held the same key, a single private key could
   produce both signatures. This check is on the *keys*, not merely the indices
   (rule 3), and is required in addition to it.
6. Reject if either signature's length is not the expected length for `Algo`.
7. Recompute the canonical bytes and reject unless BOTH signatures verify at
   their indices over exactly those bytes.

Anti-replay (a durable, fail-closed sliding-window `Seq` + nonce set scoped per
`(operator, coordinator)`) is a coordinator-side obligation layered ON TOP of the
above; it is not part of the canonical bytes and is out of scope for this
encoding spec.

### 11.1 OperatorTransmission — signed by the OPERATOR (2 of 7)
Domain tag: `sohocloud/operator/v0`

Order: `OperatorID` (string), `Ts` (int64 UTC Unix nanoseconds — the raw `int64`,
NOT re-derived through a `time.Time`, so the range-checked value is exactly the
signed value), `Nonce` (bytes, `len >= 16`), `Seq` (uint64), `Algo` (string),
`Idx0` (uint64), `Idx1` (uint64).

The two signatures `Sig0` (by the key at `Idx0`) and `Sig1` (by the key at
`Idx1`) are taken over these canonical bytes and are EXCLUDED from them, exactly
as `Signature` is elsewhere. `Ts`, `Nonce`, `Seq`, both indices, and `Algo` are
all inside the signed bytes.

### 11.2 OperatorRotation — signed by the OPERATOR (2 of 7)
Domain tag: `sohocloud/operator-rotate/v0`

Order: `OperatorID` (string), `KeyIndex` (uint64 — the index whose key is being
replaced), `NewPublicKey` (bytes — the operator-generated replacement public
key), `Algo` (string), `Ts` (int64 UTC Unix nanoseconds), `Nonce` (bytes,
`len >= 16`), `Seq` (uint64), `Idx0` (uint64), `Idx1` (uint64).

This authorizes swapping in a new public key. The operator (never the network)
generates the new keypair. The new public key is INSIDE the signed bytes so that
neither a man-in-the-middle of the out-of-band registration nor a compromised
admin channel can inject key material: a coordinator MUST verify this message
(with two CURRENT keys, `Idx0`/`Idx1`) before recording the new key at
`KeyIndex`. A verifier MUST additionally reject a `NewPublicKey` whose length is
wrong for `Algo`. A verifier MUST also reject a `NewPublicKey` that is
byte-identical to a key already registered at a DIFFERENT index (it would
violate the pairwise-distinct-keys requirement of §11.0 rule 5a and set up the
2-of-7 → 1-of-1 degradation); re-registering the same key at its OWN `KeyIndex`
(a no-op refresh) is permitted. To preserve the no-mixed-algorithms invariant, a
verifier SHOULD also reject a rotation whose `Algo` differs from that of the
other registered keys.

### 11.3 ConformanceResponse — signed by the OPERATOR (2 of 7)
Domain tag: `sohocloud/operator-conformance/v0`

Order: `OperatorID` (string), `Challenge` (bytes — the verifier-supplied
challenge), `Algo` (string), `Idx0` (uint64), `Idx1` (uint64).

This is an operator's signed answer to a conformance challenge. Its distinct
domain tag domain-separates it from §11.1: because the tag is part of the
canonical bytes, a signature produced for a conformance challenge can NEVER
verify as an `OperatorTransmission`, and vice versa. It carries no nonce/`Seq`
of its own — freshness comes from the verifier's `Challenge` — so the §11.0
nonce rule does not apply to it; all other §11.0 checks do.
