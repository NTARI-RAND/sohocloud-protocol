// Package anchor is a STUB. It is not built.
//
// It reserves the place for a witnessed, append-only anchoring of employment
// claims — the JobReport facts (job in, result out, meter) — so a contributor
// can verify that the work it did and the payment it is owed are recorded
// honestly and cannot be silently rewritten by a coordinator. This bears on
// open problem #6 (computation honesty) and closes the integrity gap that a
// bare coordination protocol otherwise leaves: without it, a coordinator could
// equivocate about a completed job and stiff the contributor.
//
// When it is built, these invariants are inherited from Janus-Facing
// Architecture and are NOT negotiable:
//
//   - It MUST NOT contain PII. Anchor references and hashes only; any
//     identifying narrative stays in erasable, consumer-local storage.
//   - It MUST be append-only. Corrections are new entries; no update, no delete.
//   - It MUST NOT be one global ledger over unrelated exchanges, and MUST NOT
//     introduce a consensus layer or a central authority. Per-operator logs plus
//     independent, federatable witnesses (the Certificate Transparency model):
//     signed monotonic checkpoints and cross-operator inclusion proofs.
//   - No single service, SoHoLINK included, may be the sole witness in a way
//     that cannot be federated. A single-witness deployment MUST be labeled the
//     stand-in it is, with a config path to independent witnesses.
//
// Building this revives a deferred decision: how many independent witnesses run
// at launch. That is a governance/values call and is intentionally left open.
package anchor
