# Protocol reliability checks

The reference transport remains optional; middleware owns SPIFFE authentication,
coordinators own replay/sequence enforcement, and callers verify signed messages.
The tests here establish protocol behavior, not production deployment readiness.

## Local checks

From the module root, using Go 1.22 or newer and Python 3:

```sh
test -z "$(gofmt -l .)"
go vet ./...
go build ./...
go test ./...
go test -race ./...
python3 testdata/reproduce_spec.py
go list -m all
```

The final command should list only this module. Tests using `httptest.NewServer`
need permission to bind loopback sockets; do not skip them to declare a pass.

Run the five bounded fuzz targets independently:

```sh
go test . -run '^$' -fuzz '^FuzzMessageVerification$' -fuzztime 10s -parallel 2
go test ./canon -run '^$' -fuzz '^FuzzVerifySig$' -fuzztime 10s -parallel 2
go test ./operator -run '^$' -fuzz '^FuzzOperatorVerification$' -fuzztime 10s -parallel 2
go test ./transport/httpjson -run '^$' -fuzz '^FuzzHandlerJSON$' -fuzztime 10s -parallel 2
go test ./transport/httpjson -run '^$' -fuzz '^FuzzClientJSON$' -fuzztime 10s -parallel 2
```

`go test ./...` runs seed cases without starting a fuzz campaign. The signature
helper target seeds from the normative golden fixture. The other targets seed
valid typed messages and malformed input. JSON fuzz targets use in-process HTTP
recorders/transports, so each iteration needs no listener or remote service.

## What fails the build

- Transport tests cover all six operations, signed payload preservation, escaped
  node IDs, methods, malformed JSON, error propagation, body cleanup, concurrent
  calls, and cancellation before and during an HTTP request.
- Timestamp tests exercise every `time.Time` field on the twelve signed message
  types, including each field of two-timestamp messages independently. They
  reject zero/out-of-range instants even when signed over incomplete encoder
  output; both valid int64-nanosecond endpoints still verify (SPEC §3).
- Fuzz targets exercise both JSON decode paths, panic-safe signature verification,
  all twelve timestamp-bearing message families, and all three operator families.
  Beyond its signed seeds, `FuzzOperatorVerification` primarily checks panic
  safety: random mutations are not expected to produce valid Ed25519 signatures.
  On successful HTTP responses, `FuzzClientJSON` checks body closure, not the
  decoded payload's correctness.
  Operator replay state and key-registration policy remain coordinator concerns.
- Existing golden-vector comparison, independent Python reproduction,
  dependency-leaf audit, and JFA invariant bindings continue to run in
  `.github/workflows/ci.yml` and `.github/workflows/dependency-leaf.yml`.

`.github/workflows/reliability.yml` runs race checks and five 10-second fuzz jobs
on PRs and pushes to `main`. Its daily schedule runs each target for 60 seconds;
manual runs use 10 seconds. Scheduled execution begins after this workflow lands
on the default branch. Existing build/vet/test and Python CI jobs remain intact.

The fuzz step has a 10-minute timeout within the 15-minute job budget, leaving
time for the failure upload. If the step fails or times out, the workflow
attempts to retain its log and any `testdata/fuzz` inputs for 14 days. Cancellation
or runner loss can still prevent an upload. Only PR runs cancel in-progress runs;
push, scheduled and manual runs let the current run finish.
Reproduce a failure with the package/seed command printed by Go, add a focused
regression assertion, and commit the confirmed failing input alongside its fix.
Fuzzing is bounded evidence, not an exhaustive proof. Hosted CI and branch
protection configuration require maintainer review; the new job names do not
become required status checks automatically.
