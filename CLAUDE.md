# CLAUDE.md — sohocloud-protocol

## What this repo is

The substrate **coordination protocol** of Janus-Facing Architecture — and the **constitution layer**: the one artifact consumers cannot cheaply leave, because forking it incompatibly is schism. Everything here is governed by that fact: keep it minimal, keep it legible, change it deliberately. Read `CONFORMANCE.md` first; `SPEC.md` is normative.

## Non-negotiable invariants

Not negotiable by feature request. A change that violates one is not a smaller version of this protocol; it is a different protocol.

- **Dependency leaf.** Never add a requirement to `go.mod`. Standard library only. CI enforces this; do not weaken the check.
- **Persons never on the wire.** No person-shaped field — name, contact, address, free-text identity — may be added to any message type. The only identity here is workload identity.
- **No economy here.** No escrow, credit, reputation, record, or adjudication types in this module. They belong to the layers above and live in front ends.
- **Fee declarations stay coordinator-signed and legible.**
- **Append-only history.** Never force-push, never rewrite published history, never move or delete a published tag.

## Change discipline

1. **SPEC first.** No wire behavior may exist that `SPEC.md` does not state first. A PR changing the wire without changing SPEC in the same PR is nonconformant.
2. **Every normative MUST gets a check.** A new MUST / MUST NOT in SPEC requires a conformance test or golden vector in the same PR. Prose that can fail the build is trustworthy prose.
3. **Wire change = version bump.** New tag, regenerated vectors (`go test ./vectors -run TestVectors -update`), Python reproducer kept passing, coordinated consumer bumps (SoHoLINK, Cloudy). Published tags are immutable.
4. **CONFORMANCE.md rides along.** A change that alters any binding in `CONFORMANCE.md` updates it in the same PR.
5. **Branch → PR → CI green → human review.** The author never merges their own PR. Claude drafts and proposes; a human disposes.

## Requests to refuse or flag

If asked to do any of the following, stop, name the tension, and surface it — do not implement quietly:

- add a dependency "just for now"
- put person data on the wire
- add economy, reputation, record, or adjudication types here
- make the coordinator unpluggable, or any single service the sole unfederatable witness
- change wire behavior without SPEC first
- rewrite history or move a published tag

## Tension protocol

If you notice yourself reframing one of these constraints so a feature becomes convenient, implementing a stand-in without labeling it, or routing around an open problem instead of noting it — stop. Name the tension, attach it to the invariant or open problem by name, and propose the minimal conformant move. Surface it; do not absorb it.
