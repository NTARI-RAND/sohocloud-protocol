#!/usr/bin/env python3
"""
Independent reproduction of sohocloud-protocol v0 canonical byte encoding.

Written from SPEC.md (SS3 canonical encoding, SS4 message field orders) and the
INPUT field values in testdata/vectors.json ALONE. No Go source was consulted.

Standard library only. Hand-rolled uvarint, big-endian int64/uint64, UnixNano
time. Verifies each committed canonical_bytes_hex byte-for-byte, and (if a
pure-python ed25519 is provided below) verifies signatures over the
independently derived bytes against public_key_hex.
"""

import sys

# Print UTF-8 regardless of the host console codepage (Windows defaults to cp1252,
# which cannot encode the multibyte string test case). Reproducibility is the point
# of this script, so it must run cleanly on any forker's machine.
try:
    sys.stdout.reconfigure(encoding="utf-8")
except (AttributeError, ValueError):
    pass

import json
import os
import hashlib

# --------------------------------------------------------------------------
# Canonical primitives, hand-rolled strictly from SPEC.md SS3.
# --------------------------------------------------------------------------


def uvarint(n: int) -> bytes:
    """Unsigned LEB128 varint, as produced by Go encoding/binary.PutUvarint.

    Little-endian base-128 groups, high bit set on all but the last byte.
    """
    if n < 0:
        raise ValueError("uvarint requires non-negative")
    out = bytearray()
    while True:
        b = n & 0x7F
        n >>= 7
        if n:
            out.append(b | 0x80)
        else:
            out.append(b)
            break
    return bytes(out)


def enc_string(s: str) -> bytes:
    raw = s.encode("utf-8")
    return uvarint(len(raw)) + raw


def enc_bytes(p: bytes) -> bytes:
    return uvarint(len(p)) + p


def enc_bool(v: bool) -> bytes:
    return b"\x01" if v else b"\x00"


def enc_uint64(n: int) -> bytes:
    if n < 0 or n > 0xFFFFFFFFFFFFFFFF:
        raise ValueError("uint64 out of range")
    return n.to_bytes(8, "big", signed=False)


def enc_int64(n: int) -> bytes:
    # int64 encoded as its two's-complement uint64, fixed 8 bytes big-endian.
    if n < -(2 ** 63) or n > 2 ** 63 - 1:
        raise ValueError("int64 out of range")
    return (n & 0xFFFFFFFFFFFFFFFF).to_bytes(8, "big", signed=False)


def enc_time(unixnano: int) -> bytes:
    # time == int64 UTC Unix nanoseconds.
    return enc_int64(unixnano)


def enc_domain(tag: str) -> bytes:
    # domain tag encoded as a string.
    return enc_string(tag)


# --------------------------------------------------------------------------
# Message encoders, field order strictly from SPEC.md SS4. Signature excluded.
# --------------------------------------------------------------------------


def canon_capability_listing(f) -> bytes:
    out = bytearray()
    out += enc_domain("sohocloud/listing/v0")
    out += enc_string(f["NodeID"])
    out += enc_string(f["Class"])
    printers = f["Printers"]
    out += uvarint(len(printers))          # repeated: count then elements inline
    for p in printers:
        out += enc_string(p["Kind"])
        out += enc_string(p["Model"])
    gpus = f.get("GPUs", [])
    out += uvarint(len(gpus))              # repeated: count then elements inline
    for g in gpus:
        out += enc_string(g["API"])
        out += enc_string(g["Model"])
        out += enc_int64(g["VRAMMB"])
    cap = f["Capacity"]
    out += enc_int64(cap["VCPUs"])
    out += enc_int64(cap["MemMB"])
    out += enc_int64(cap["DiskMB"])
    out += enc_int64(cap["StorageCommitMB"])
    out += enc_int64(cap["PrintQPS"])
    out += enc_bool(f["OptIn"]["Compute"])
    out += enc_bool(f["OptIn"]["Print"])
    out += enc_bool(f["OptIn"]["Storage"])
    out += enc_time(f["_IssuedAtNanos"])
    out += enc_uint64(f["Seq"])
    return bytes(out)


def canon_heartbeat(f) -> bytes:
    out = bytearray()
    out += enc_domain("sohocloud/heartbeat/v0")
    out += enc_string(f["NodeID"])
    out += enc_time(f["_SentAtNanos"])
    out += enc_uint64(f["Seq"])
    return bytes(out)


def canon_assignment(f) -> bytes:
    out = bytearray()
    out += enc_domain("sohocloud/assignment/v0")
    out += enc_string(f["JobID"])
    out += enc_string(f["NodeID"])
    spec = f["Spec"]
    out += enc_string(spec["Workload"])
    out += enc_string(spec["Image"])
    args = spec["Args"]
    out += uvarint(len(args))              # repeated string
    for a in args:
        out += enc_string(a)
    out += enc_string(spec["PrinterKind"])
    out += enc_string(spec["GPUAPI"])
    out += enc_int64(spec["GPUMinVRAMMB"])
    fee = f["Fee"]
    out += enc_int64(fee["ContributorShareBps"])
    out += enc_int64(fee["PlatformFeeBps"])
    out += enc_time(f["_OfferedAtNanos"])
    return bytes(out)


def canon_decline(f) -> bytes:
    out = bytearray()
    out += enc_domain("sohocloud/decline/v0")
    out += enc_string(f["JobID"])
    out += enc_string(f["NodeID"])
    out += enc_string(f["Reason"])
    out += enc_time(f["_DeclinedAtNanos"])
    return bytes(out)


def canon_jobreport(f) -> bytes:
    out = bytearray()
    out += enc_domain("sohocloud/jobreport/v0")
    out += enc_string(f["JobID"])
    out += enc_string(f["NodeID"])
    out += enc_int64(f["ExitCode"])
    out += enc_string(f["FailureCause"])
    out += enc_bool(f["TmpfsExhausted"])
    out += enc_time(f["_StartedAtNanos"])
    out += enc_time(f["_FinishedAtNanos"])
    return bytes(out)


def canon_feedeclaration(f) -> bytes:
    out = bytearray()
    out += enc_domain("sohocloud/fee/v0")
    out += enc_string(f["CoordinatorID"])
    out += enc_int64(f["ContributorShareBps"])
    out += enc_int64(f["PlatformFeeBps"])
    out += enc_time(f["_EffectiveAtNanos"])
    out += enc_uint64(f["Seq"])
    return bytes(out)


def canon_storage_lease(f) -> bytes:
    out = bytearray()
    out += enc_domain("sohocloud/lease/v0")
    out += enc_string(f["LeaseID"])
    out += enc_string(f["NodeID"])
    out += enc_bytes(bytes.fromhex(f["ShardRef"]))
    out += enc_int64(f["SizeClass"])
    fee = f["Fee"]
    out += enc_int64(fee["ContributorShareBps"])
    out += enc_int64(fee["PlatformFeeBps"])
    out += enc_time(f["_IssuedAtNanos"])
    out += enc_time(f["_ExpiresAtNanos"])
    out += enc_uint64(f["Seq"])
    return bytes(out)


def canon_lease_decline(f) -> bytes:
    out = bytearray()
    out += enc_domain("sohocloud/lease-decline/v0")
    out += enc_string(f["LeaseID"])
    out += enc_string(f["NodeID"])
    out += enc_string(f["Reason"])
    out += enc_time(f["_DeclinedAtNanos"])
    return bytes(out)


def canon_lease_release(f) -> bytes:
    out = bytearray()
    out += enc_domain("sohocloud/lease-release/v0")
    out += enc_string(f["LeaseID"])
    out += enc_string(f["NodeID"])
    out += enc_time(f["_ReleasedAtNanos"])
    return bytes(out)


def canon_proof_response(f) -> bytes:
    out = bytearray()
    out += enc_domain("sohocloud/proof/v0")
    out += enc_string(f["LeaseID"])
    out += enc_string(f["NodeID"])
    out += enc_int64(f["Offset"])
    out += enc_int64(f["Length"])
    out += enc_bytes(bytes.fromhex(f["Nonce"]))
    out += enc_bytes(bytes.fromhex(f["Digest"]))
    out += enc_time(f["_RespondedAtNanos"])
    return bytes(out)


def canon_key_rotation(f) -> bytes:
    out = bytearray()
    out += enc_domain("sohocloud/node-rotate/v0")
    out += enc_string(f["NodeID"])
    out += enc_bytes(bytes.fromhex(f["NewPublicKey"]))
    out += enc_string(f["Algo"])
    out += enc_time(f["_RotatedAtNanos"])
    out += enc_uint64(f["Seq"])
    return bytes(out)


def canon_key_revocation(f) -> bytes:
    out = bytearray()
    out += enc_domain("sohocloud/node-revoke/v0")
    out += enc_string(f["NodeID"])
    out += enc_bytes(bytes.fromhex(f["RevokedPublicKey"]))
    out += enc_time(f["_RevokedAtNanos"])
    out += enc_uint64(f["Seq"])
    return bytes(out)


# --------------------------------------------------------------------------
# Layer-C OPERATOR message encoders (SPEC.md SS11). Two signatures each,
# excluded from the canonical bytes. Field order strictly from SS11.
# --------------------------------------------------------------------------


def canon_operator_transmission(f) -> bytes:
    # SS11.1 Order: OperatorID, Ts, Nonce, Seq, Algo, Idx0, Idx1.
    out = bytearray()
    out += enc_domain("sohocloud/operator/v0")
    out += enc_string(f["OperatorID"])
    out += enc_int64(f["_TsNanos"])
    out += enc_bytes(bytes.fromhex(f["Nonce"]))
    out += enc_uint64(f["Seq"])
    out += enc_string(f["Algo"])
    out += enc_uint64(f["Idx0"])
    out += enc_uint64(f["Idx1"])
    return bytes(out)


def canon_operator_rotation(f) -> bytes:
    # SS11.2 Order: OperatorID, KeyIndex, NewPublicKey, Algo, Ts, Nonce, Seq,
    # Idx0, Idx1.
    out = bytearray()
    out += enc_domain("sohocloud/operator-rotate/v0")
    out += enc_string(f["OperatorID"])
    out += enc_uint64(f["KeyIndex"])
    out += enc_bytes(bytes.fromhex(f["NewPublicKey"]))
    out += enc_string(f["Algo"])
    out += enc_int64(f["_TsNanos"])
    out += enc_bytes(bytes.fromhex(f["Nonce"]))
    out += enc_uint64(f["Seq"])
    out += enc_uint64(f["Idx0"])
    out += enc_uint64(f["Idx1"])
    return bytes(out)


def canon_conformance_response(f) -> bytes:
    # SS11.3 Order: OperatorID, Challenge, Algo, Idx0, Idx1.
    out = bytearray()
    out += enc_domain("sohocloud/operator-conformance/v0")
    out += enc_string(f["OperatorID"])
    out += enc_bytes(bytes.fromhex(f["Challenge"]))
    out += enc_string(f["Algo"])
    out += enc_uint64(f["Idx0"])
    out += enc_uint64(f["Idx1"])
    return bytes(out)


OPERATOR_ENCODERS = {
    "OperatorTransmission": canon_operator_transmission,
    "OperatorRotation": canon_operator_rotation,
    "ConformanceResponse": canon_conformance_response,
}


def prepare_operator_fields(name, fields):
    """Flatten the 'unixnano <int>' time inputs into raw nanos for operator
    messages that carry a timestamp."""
    f = dict(fields)
    if "Ts" in fields:
        f["_TsNanos"] = int(str(fields["Ts"]).split()[-1])
    return f


# --------------------------------------------------------------------------
# Minimal pure-python ed25519 verify (RFC 8032), stdlib only.
# --------------------------------------------------------------------------

_p = 2 ** 255 - 19
_L = 2 ** 252 + 27742317777372353535851937790883648493
_d = (-121665 * pow(121666, _p - 2, _p)) % _p
_I = pow(2, (_p - 1) // 4, _p)


def _inv(x):
    return pow(x, _p - 2, _p)


def _xrecover(y):
    xx = (y * y - 1) * _inv(_d * y * y + 1)
    x = pow(xx, (_p + 3) // 8, _p)
    if (x * x - xx) % _p != 0:
        x = (x * _I) % _p
    if x % 2 != 0:
        x = _p - x
    return x


_By = (4 * _inv(5)) % _p
_Bx = _xrecover(_By)
_B = (_Bx % _p, _By % _p, 1, (_Bx * _By) % _p)


def _edwards_add(P, Q):
    x1, y1, z1, t1 = P
    x2, y2, z2, t2 = Q
    a = ((y1 - x1) * (y2 - x2)) % _p
    b = ((y1 + x1) * (y2 + x2)) % _p
    c = (t1 * 2 * _d * t2) % _p
    dd = (z1 * 2 * z2) % _p
    e = b - a
    f = dd - c
    g = dd + c
    h = b + a
    x3 = e * f
    y3 = g * h
    t3 = e * h
    z3 = f * g
    return (x3 % _p, y3 % _p, z3 % _p, t3 % _p)


def _scalarmult(P, e):
    Q = (0, 1, 1, 0)
    while e > 0:
        if e & 1:
            Q = _edwards_add(Q, P)
        P = _edwards_add(P, P)
        e >>= 1
    return Q


def _encodepoint(P):
    x, y, z, t = P
    zi = _inv(z)
    x = (x * zi) % _p
    y = (y * zi) % _p
    bits = [(y >> i) & 1 for i in range(255)] + [x & 1]
    return bytes(sum(bits[i * 8 + j] << j for j in range(8)) for i in range(32))


def _decodepoint(s):
    y = int.from_bytes(s, "little") & ((1 << 255) - 1)
    x = _xrecover(y)
    if x & 1 != (int.from_bytes(s, "little") >> 255):
        x = _p - x
    P = (x, y, 1, (x * y) % _p)
    return P


def _Hint(m):
    return int.from_bytes(hashlib.sha512(m).digest(), "little")


def ed25519_verify(public_key: bytes, message: bytes, signature: bytes) -> bool:
    if len(signature) != 64 or len(public_key) != 32:
        return False
    R = _decodepoint(signature[:32])
    A = _decodepoint(public_key)
    S = int.from_bytes(signature[32:], "little")
    h = _Hint(signature[:32] + public_key + message) % _L
    left = _scalarmult(_B, S)
    right = _edwards_add(R, _scalarmult(A, h))
    return _encodepoint(left) == _encodepoint(right)


# --------------------------------------------------------------------------
# Driver
# --------------------------------------------------------------------------

ENCODERS = {
    "CapabilityListing": canon_capability_listing,
    "Heartbeat": canon_heartbeat,
    "Assignment": canon_assignment,
    "Decline": canon_decline,
    "JobReport": canon_jobreport,
    "FeeDeclaration": canon_feedeclaration,
    "StorageLease": canon_storage_lease,
    "LeaseDecline": canon_lease_decline,
    "LeaseRelease": canon_lease_release,
    "ProofResponse": canon_proof_response,
    "KeyRotation": canon_key_rotation,
    "KeyRevocation": canon_key_revocation,
}


def prepare_fields(name, fields):
    """Flatten the time-string inputs from vectors.json into raw nanos.

    The vectors.json 'time' inputs are strings like 'unixnano 1700000000000000123'.
    Extract the integer nanos.
    """
    f = dict(fields)

    def nanos(key):
        s = fields[key]
        # format: 'unixnano <int>'
        return int(s.split()[-1])

    if name == "CapabilityListing":
        f["_IssuedAtNanos"] = nanos("IssuedAt")
    elif name == "Heartbeat":
        f["_SentAtNanos"] = nanos("SentAt")
    elif name == "Assignment":
        f["_OfferedAtNanos"] = nanos("OfferedAt")
    elif name == "Decline":
        f["_DeclinedAtNanos"] = nanos("DeclinedAt")
    elif name == "JobReport":
        f["_StartedAtNanos"] = nanos("StartedAt")
        f["_FinishedAtNanos"] = nanos("FinishedAt")
    elif name == "FeeDeclaration":
        f["_EffectiveAtNanos"] = nanos("EffectiveAt")
    elif name == "StorageLease":
        f["_IssuedAtNanos"] = nanos("IssuedAt")
        f["_ExpiresAtNanos"] = nanos("ExpiresAt")
    elif name == "LeaseDecline":
        f["_DeclinedAtNanos"] = nanos("DeclinedAt")
    elif name == "LeaseRelease":
        f["_ReleasedAtNanos"] = nanos("ReleasedAt")
    elif name == "ProofResponse":
        f["_RespondedAtNanos"] = nanos("RespondedAt")
    elif name == "KeyRotation":
        f["_RotatedAtNanos"] = nanos("RotatedAt")
    elif name == "KeyRevocation":
        f["_RevokedAtNanos"] = nanos("RevokedAt")
    return f


def check_primitives(prims):
    results = []
    for p in prims:
        kind = p["kind"]
        inp = p["input"]
        want = p["bytes_hex"]
        if kind == "uvarint":
            got = uvarint(int(inp)).hex()
        elif kind == "int64":
            got = enc_int64(int(inp)).hex()
        elif kind == "uint64":
            got = enc_uint64(int(inp)).hex()
        elif kind == "bool":
            got = enc_bool(inp == "true").hex()
        elif kind == "string":
            got = enc_string(inp).hex()
        elif kind == "time":
            # input like 'unixnano 0' or 'unixnano 1700000000000000123' or with a prefix
            n = int(inp.split()[-1].strip("()")) if "unixnano" in inp else int(inp)
            # handle '1970-...Z (unixnano 0)' form
            if "unixnano" in inp:
                # take token right after 'unixnano'
                toks = inp.replace("(", " ").replace(")", " ").split()
                idx = toks.index("unixnano")
                n = int(toks[idx + 1])
            got = enc_time(n).hex()
        else:
            results.append((kind + ":" + inp, False, "unknown kind", want))
            continue
        results.append((kind + ":" + inp, got == want, got, want))
    return results


def main():
    here = os.path.dirname(os.path.abspath(__file__))
    with open(os.path.join(here, "vectors.json"), encoding="utf-8") as fh:
        v = json.load(fh)

    all_ok = True
    total = 0

    print("=== PRIMITIVES ===")
    for label, ok, got, want in check_primitives(v["primitives"]):
        total += 1
        if not ok:
            all_ok = False
        print(f"[{'OK' if ok else 'FAIL'}] {label:40s} got={got} want={want}")

    print("\n=== MESSAGES (canonical bytes) ===")
    for m in v["messages"]:
        name = m["name"]
        total += 1
        f = prepare_fields(name, m["fields"])
        got = ENCODERS[name](f).hex()
        want = m["canonical_bytes_hex"]
        ok = got == want
        if not ok:
            all_ok = False
        print(f"[{'OK' if ok else 'FAIL'}] {name}")
        if not ok:
            print(f"    got : {got}")
            print(f"    want: {want}")

    print("\n=== SIGNATURES (ed25519 over independently-derived bytes) ===")
    for m in v["messages"]:
        name = m["name"]
        f = prepare_fields(name, m["fields"])
        canon = ENCODERS[name](f)
        pub = bytes.fromhex(m["public_key_hex"])
        sig = bytes.fromhex(m["signature_hex"])
        total += 1
        ok = ed25519_verify(pub, canon, sig)
        if not ok:
            all_ok = False
        print(f"[{'OK' if ok else 'FAIL'}] {name} sig verifies over derived bytes")

    # ---- Layer-C operator messages (2-of-7, two signatures each) ----
    op_msgs = v.get("operator_messages", [])
    if op_msgs:
        print("\n=== OPERATOR MESSAGES (canonical bytes) ===")
        for m in op_msgs:
            name = m["name"]
            total += 1
            f = prepare_operator_fields(name, m["fields"])
            got = OPERATOR_ENCODERS[name](f).hex()
            want = m["canonical_bytes_hex"]
            ok = got == want
            if not ok:
                all_ok = False
            print(f"[{'OK' if ok else 'FAIL'}] {name}")
            if not ok:
                print(f"    got : {got}")
                print(f"    want: {want}")

        print("\n=== OPERATOR SIGNATURES (both ed25519 sigs over derived bytes) ===")
        for m in op_msgs:
            name = m["name"]
            f = prepare_operator_fields(name, m["fields"])
            canon_bytes = OPERATOR_ENCODERS[name](f)
            # Resolve the two signing public keys from the registry by index.
            pubs = {k["index"]: bytes.fromhex(k["public_key_hex"]) for k in m["keys"]}
            idx0, idx1 = m["idx0"], m["idx1"]
            total += 1
            distinct = idx0 != idx1
            ok0 = ed25519_verify(pubs[idx0], canon_bytes, bytes.fromhex(m["sig0_hex"]))
            ok1 = ed25519_verify(pubs[idx1], canon_bytes, bytes.fromhex(m["sig1_hex"]))
            ok = distinct and ok0 and ok1
            if not ok:
                all_ok = False
            print(f"[{'OK' if ok else 'FAIL'}] {name} both 2-of-7 sigs verify over derived bytes")

    print(f"\nTOTAL CHECKS: {total}  RESULT: {'ALL MATCH' if all_ok else 'MISMATCH'}")
    return 0 if all_ok else 1


if __name__ == "__main__":
    raise SystemExit(main())
