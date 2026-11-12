"""Independently check SPEC 7.1 response framing and byte-budget fixtures."""

import json
from pathlib import Path


def reject_constant(value):
    """JSON has no NaN or Infinity literals."""
    raise ValueError(f"non-JSON constant: {value}")


def accepts(body, limit):
    """Zero is unbounded; otherwise count UTF-8 bytes before parsing one document."""
    if limit < 0 or (limit > 0 and len(body.encode("utf-8")) > limit):
        return False
    try:
        json.loads(body, parse_constant=reject_constant)
    except ValueError:
        return False
    return True


def main():
    path = Path(__file__).resolve().parents[1] / "testdata/response_vectors.json"
    cases = json.loads(path.read_text(encoding="utf-8"))
    for case in cases:
        if accepts(case["body"], case["limit"]) != case["accept"]:
            raise SystemExit(f"response conformance failed: {case['name']}")
    print(f"All {len(cases)} response framing/budget vectors match SPEC 7.1")


if __name__ == "__main__":
    main()
