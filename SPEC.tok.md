# sohocloud-protocol — nasin pi wan nasin (lipu lili)

> **toki pi lipu ni:** lipu ni li lipu lili tan lipu suli. nasin NTARI "P2-002" li pana e ken tawa ona, tawa kulupu. sina wile e sona ale la, o lukin e lipu suli pi toki Inli: `SPEC.md` (tan tenpo suno 2026-09-16). ilo sona li pali e lipu ni. kulupu li lukin ala e ona lon tenpo ni. nasin P2-002 kipisi 3.1 li wile e ni: kulupu o lukin e lipu ni.
>
> **Note:** This is a condensed community rendering produced under NTARI policy P2-002. The complete document is the English original `SPEC.md` (snapshot 2026-09-16), and **the English is normative**: where this rendering differs, the English governs. This spec is a conformance contract — the normative keywords (MUST, MUST NOT, SHOULD, MAY), all field names, domain tags, type identifiers and code blocks are kept in English and are NOT translated, because an implementation is checked against those exact strings. Machine-assisted draft pending community review per P2-002 section 3.1; core technical specifications remain in English per section 2.2.
>
> sina lukin e pakala lon toki ni la o pona e ona: o pali e "fork" lon
> https://github.com/NTARI-RAND/sohocloud-protocol, o pana e "pull request".
> pana sina li pona tawa mi mute.

nanpa: **v0** — ona li awen ala. lipu li ken ante lon tenpo kama, tawa tenpo pi
nanpa v1.

lipu ni li **lawa**. ona li ijo pi pali ilo, li ijo poka ala. nasin AGPL li pana e
ken ni tawa jan: jan li ken pali sin e ilo sama. taso ken ni li ken ala lon la jan
li ken ala pali e ilo sin tan lipu taso. ni la lipu o pona, o lukin ken. ijo pi
nimi `MUST` anu `MUST NOT` li **wile lawa**. jan li pakala e wan la ilo ona li ilo
**ante**, li ilo lili ala pi ilo ni.

## 1. lipu ni li toki e seme

lipu ni li lawa e **kipisi pi wan ilo** (substrate coordination layer): ilo li
kama sona kepeken nasin seme; ona li toki e pali ona kepeken nasin seme; pali li
tawa ona, en toki pini li tawa monsi, kepeken nasin seme. jan pali tu wan li
kepeken lipu ni:

1. **Node** — ilo pi jan lili. ona li pana e wawa sona (compute) anu ilo pi pali
   sijelo (print). ona li jo e nimi ona. ona li **pana e nimi ona** (sign) lon ijo
   ale pi tan ona. ona li **lawa e ona sama**: ona li wile e `Coordinator` seme
   la ni li pona; ona li ken weka tan ona lon tenpo ale.
2. **Coordinator** — ona li wan e pali e ilo, li awen e pini. ona li ijo pi
   weka ken (pluggable). SoHoLINK li `Coordinator` nanpa wan, taso `Frontend` li
   MAY pali e `Coordinator` ona sama, anu li toki tawa `Node` lon nasin pi ijo
   insa ala.
3. **Frontend** — ma pi esun pi jan kulupu. ona li kepeken nasin sama.

**lipu ni li toki ALA e ni:** esun pi kulupu JFA. lipu ni li jo ala e `escrow`,
li jo ala e mani pi pana jan, li jo ala e nasin pi toki awen (covenant), li jo ala
e lipu pi toki pini, li jo ala e nasin pi pini utala. ijo ni ale li lon
`Frontend`. kin la mani pi ma suli (fiat) lon poka pi `Coordinator` en `Node` li
lon ala lipu ni.

## 2. nimi en awen pi nimi

`Node` li jo e nimi `NodeID` — nimi awen pi ona sama.

nimi pi pali ona li `SPIFFE SVID`. nasin ona li ni taso:

```
/node/<NodeID>
```

`Coordinator` li lukin e nasin `SPIFFE` pi jan toki, li wan e ona tawa
`/node/<NodeID>` pi `NodeID` lon toki.

- `SPIFFE` li lon ala la `Coordinator` MUST toki e **401**.
- `SPIFFE` li lon taso nasin ona li sama ala `/node/<NodeID>` la `Coordinator`
  MUST toki e **403**.
- `Coordinator` MUST NOT toki e ni: "`SVID` pona ale li pona tawa `Node` ale."
  ni li nasin pi jan ike (impersonation).
- awen ni MUST kama tan nimi taso (deterministic string). ona MUST NOT kama tan
  poki pi ante ken. lukin pi ijo lon ala li lukin **ala**; ona MUST NOT awen e
  sina.

lipu ni li pana **ala** e nimi len (public key). `Coordinator` li kama jo e ona
lon nasin ante (kama lon kulupu, `SVID`, anu poki nimi).

## 3. sike nanpa pi nimi pana (canonical bytes)

toki ale pi nimi pana li pana e nimi lon **sike nanpa pi nasin wan** (canonical
bytes). ijo `Signature` li **lon ala** insa ona. nasin pi nimi pana li
**ed25519**.

nimi li pana lon `JSON` **ala**, li pana lon nasin tawa **ala**. ona li pana lon
nasin nanpa ni taso, tan ni: nimi pana o sama lon ilo ale, lon toki ilo ale.

ijo nanpa, lon nasin pi lipu kipisi 4:

- **unsigned LEB128 varint** — nanpa pi suli en nanpa pi mute. nanpa `0` li sike
  `0x00`.
- **domain tag** — nimi pi toki kulipisi wan. sike nanpa ale li open kepeken ona.
  ona MUST kama nanpa kepeken nasin `string` (nanpa pi suli, lon open; nimi
  UTF-8, lon monsi). nimi ante pi kipisi ante la nimi pana pi toki wan li ken ala
  tawa toki ante. ni li pona.
- **string** — nanpa pi suli (varint), lon open. nimi UTF-8, lon monsi. nimi ala
  li sike `0x00` taso.
- **bytes** — nanpa pi suli (varint), lon open. sike, lon monsi.
- **bool** — sike wan taso: `0x00` (lon ala) anu `0x01` (lon). sike ante li pona
  ala.
- **int64 / uint64** — sike 8, nanpa suli lon open (big-endian). `int64` MUST
  kama `uint64` kepeken nasin "two's complement": `-1` li kama
  `ff ff ff ff ff ff ff ff`.
- **time** — tenpo li kama `int64` pi nanpa lili mute (nanoseconds) tan tenpo
  `Unix UTC`. ma tenpo li weka. nanpa MUST NOT kama lili. tenpo li lon poka pi
  tenpo sike (1678–2262) ala la ni li **pakala**; ilo MUST toki e pakala, li ante
  ala e nanpa. (ilo `canon.ValidTime` li lukin e ni. `canon.Buffer.Time` li awen
  e `canon.ErrTimeOutOfRange`.)
- **repeated field** — nanpa pi **mute** li lon open. ijo ale li lon monsi, lon
  nasin pi lipu kipisi 4. ijo pi kipisi ijo li lon ala. mute `0` li sike `0x00`
  taso.

jan lukin li pali sin e sike nanpa tan ijo pi kama, li lukin e nimi pana kepeken
nimi len. jan lukin MUST weka e nimi pana pi suli ante tan `ed25519.SignatureSize`
(64), lon open pi lukin.

sike nanpa pi toki wan li **ni taso**: `domain tag`, en ijo ale pi lipu `Order:`
lon kipisi 4, lon nasin ni. ijo ante li lon ala: sike pi nanpa lipu li lon ala,
sike pi suli ale li lon ala, sike pi monsi li lon ala. lipu `Order:` li **ale**.
ijo `Signature` (anu `Sig0` en `Sig1` lon kipisi 11) li **lon ala** ona.

## 4. toki (messages)

lipu `Order:` lon kipisi 4 pi lipu suli li **lawa wan** pi nasin ijo. jan pali
MUST kepeken nasin ona. jan pali MUST NOT kama sona e nasin tan ijo ante.

> **ike pi nasin ijo:** lipu `testdata/vectors.json` li pana e `fields` lon nasin
> pi sitelen (alphabetical). ni li nasin **ala** pi sike nanpa. jan li kepeken
> nasin JSON la sike nanpa li pakala, en nimi pana li pona ala.

toki li lon ni (nimi ona li toki Inli):

| toki | jan pana nimi | domain tag |
|---|---|---|
| `CapabilityListing` | `NODE` | `sohocloud/listing/v0` |
| `Heartbeat` | `NODE` | `sohocloud/heartbeat/v0` |
| `Assignment` | `COORDINATOR` | `sohocloud/assignment/v0` |
| `Decline` | `NODE` | `sohocloud/decline/v0` |
| `JobReport` | `NODE` | `sohocloud/jobreport/v0` |
| `FeeDeclaration` | `COORDINATOR` | `sohocloud/fee/v0` |
| `StorageLease` | `COORDINATOR` | `sohocloud/lease/v0` |
| `LeaseDecline` | `NODE` | `sohocloud/lease-decline/v0` |
| `LeaseRelease` | `NODE` | `sohocloud/lease-release/v0` |
| `ProofResponse` | `NODE` | `sohocloud/proof/v0` |
| `KeyRotation` | `NODE` (nimi pi tenpo ni) | `sohocloud/node-rotate/v0` |
| `KeyRevocation` | `NODE` (nimi pi weka) | `sohocloud/node-revoke/v0` |

sona lili pi toki ni:

- **`CapabilityListing`** — `Node` li toki e ijo ona: `Class`, `Printers`,
  `GPUs`, `Capacity`, `OptIn`. `Capacity.DiskMB` li ma pi pali lon tenpo ni;
  `Capacity.StorageCommitMB` li ma pi awen suli. ona li ijo tu, tan ni: `Node` li
  ken pana e wan taso. `Node` li toki e `GPUs` la ona li wile e pali `GPU`; ona li
  toki ala e ona lon toki sin la ona li weka e ken ni. `Seq` li **suli ale** lon
  tenpo ale. `Coordinator` MUST weka e toki pi `Seq` lili anu sama.
- **`Assignment`** — mani (`Fee`) li lon insa. ni la `Node` li lukin e mani lon
  open, lon tenpo pi wile ala. toki pi "mi wile" li lon ala: nasin li nasin pi
  kama jo (pull).
- **`JobReport`** — ni li **ijo lon** pi mani. nanpa mani li kama tan ona.
- **`StorageLease`** — awen li **esun tenpo** (lease), li pali ala. `ShardRef` li
  sike 32 taso; `SizeClass` li nanpa pi suli lili. lipu ni MUST NOT pana e suli
  lon, li MUST NOT pana e ijo insa. sin la: `StorageLease` sin pi `LeaseID` sama,
  kepeken `Seq` suli.
- **`ProofResponse`** — `ProofChallenge` li **jo ala** e nimi pana (`Node` li kama
  jo e ona kepeken nasin pi kama jo). `Digest` MUST kama tan ni:
  `SHA-256(Nonce || uint64be(Offset) || uint64be(Length) || sealed[Offset : Offset+Length])`.
  jan lukin MUST weka e toki pi `(LeaseID, Nonce)` sama pi tenpo pini: `Nonce` li
  pali **wan** taso.
- **`KeyRotation`** — jan li jo e nimi len pi tenpo ni la ona li ken toki e nimi
  sin. jan lukin MUST lukin kepeken nimi pi **tenpo ni** (nimi sin li ken ala
  pana e ona sama), MUST wile e `Seq` suli ale, MUST weka e `NewPublicKey` pi
  suli ike (`ed25519`: sike 32 taso) anu `Algo` sona ala.
- **`KeyRevocation`** — ni li weka e nimi len, li pana ala e nimi sin. jan lukin
  MUST lukin kepeken nimi pi **pilin pona lon tenpo ni**, en MUST wile e ni:
  `RevokedPublicKey` li **sama** nimi ni. ni li awen e `Node` tan jan ike: jan
  ante li ken pali e toki weka kepeken nimi ona sama, taso nimi ona li sama ala
  nimi pi pilin pona, la toki ni li weka. (lipu pi tenpo pini li toki e ni: "o
  kute lon tenpo ale." ni li **lon ala** lon tenpo ni.)

## 5. ijo awen (invariants)

1. **weka pi pali (opt-out) li lon `Node`, li lon linja ala.** `OptIn` en nasin pi
   `JobSpec` li **toki taso** (advisory). `Coordinator` li **awen ala** e ni.
   `Node` MUST awen e ona sama kepeken lipu ona, li MUST weka e pali ike kepeken
   `Decline{Reason: local_policy}`.
2. **nasin pi wan (matching) li lon ala lipu ni.** `Coordinator` li lawa e ni lon
   insa ona. lipu ni MUST NOT kama jo e ijo pi lawa e wan, tan ni: ni li wan e jan
   tawa `Coordinator` wan, li pakala e ken pi weka.
3. **mani li lukin ken, li tawa monsi ala.** mani li lon `FeeDeclaration` pi nimi
   pana en tenpo. mani MUST NOT ante lon pali pi tenpo pini; ante li
   `FeeDeclaration` sin pi `EffectiveAt` sin.
4. **`Coordinator` li nasin, li ilo insa ala.** ijo wile ala o tawa `Coordinator`
   wan taso. `Node` MAY pali tawa `Coordinator` mute; `Frontend` MAY pali e
   `Coordinator` ona. `Coordinator` wan li wile lon la ni li **ilo wan**, li
   pakala e nasin 11 pi lipu JFA; ona MUST kama sitelen kepeken nimi ni.
5. **`Seq` li suli ale.** `Coordinator` MUST weka e `Seq` pi suli ala.

## 6. pali pi `Coordinator` (nasin pi kama jo)

`Coordinator` li toki **ala** tawa `Node` lon open. `Node` taso li open.

| pali | nasin | ijo |
|---|---|---|
| `SubmitListing` | `Node` → `Coordinator` | `CapabilityListing` |
| `Heartbeat` | `Node` → `Coordinator` | `Heartbeat` |
| `PollJobs` | `Node` → `Coordinator` | `NodeID` → `[]Assignment` |
| `Decline` | `Node` → `Coordinator` | `Decline` |
| `ReportJob` | `Node` → `Coordinator` | `JobReport` |
| `Fees` | jan ale → `Coordinator` | → `FeeDeclaration` |

`Coordinator` MUST kepeken awen pi kipisi 2 lon `PollJobs` (en lon pali ale pi
nimi `Node`) lon open pi pali. `Coordinator` li open ala e toki la ilo pi kute lon
`Node` li lon ala.

## 7. nasin tawa (transport)

`HTTP+JSON` li nasin tawa **nanpa wan** (`transport/httpjson`), li linja ala.
ilo insa ale li kama jo ala e ona. ilo pona MAY kepeken nasin tawa ante; ona o
kepeken sike nanpa pi kipisi 3 taso. `JSON` li nasin tawa taso — nimi pana li
kama **ala** tan `JSON`. ni la nasin tawa ante li pona, li wile ala e toki sin pi
nimi pana.

```
POST /v0/listing     body: CapabilityListing   -> 204
POST /v0/heartbeat   body: Heartbeat           -> 204
GET  /v0/jobs?node_id=<id>                      -> 200 []Assignment
POST /v0/decline     body: Decline             -> 204
POST /v0/report      body: JobReport           -> 204
GET  /v0/fees                                   -> 200 FeeDeclaration
```

awen `SPIFFE` (kipisi 2) li lon ilo poka (middleware) lon sinpin pi ilo kute. ilo
kute nanpa wan li wan e nasin en `JSON` taso; ona li ilo pi awen nimi ala.

## 8. wile sona pi pini ala

toki lon pi ijo pini ala li nasin pona, li ijo pi pilin ike ala.

- **#7 wawa sona pi lawa ona li pana e weka pi ilo, li pana ala e weka pi mani.**
  ijo 5.4 en 5.3 li utala e ni: `Coordinator` wan li kama ilo insa. taso open pi
  lete (cold start) li awen ijo. sona pi jan (reputation) li ijo pi `Frontend` en
  lawa, li ijo pi lipu ni ala.
- **#6 lon pi nanpa li lon ala lipu ni.** lipu ni li toki ala e ni: "wan pi
  `Coordinator` li pona anu ike." ona li pana e **lukin** taso (5.2, 5.3). ma
  `anchor/` li awen e ma tawa awen pi toki `JobReport` lon lipu pi jan lukin.
  ona li nimi taso; ona li pali ala.
- **#4 kama lon kulupu.** `Node` li pana e nimi ona sama (kipisi 2) la ilo lili
  (ilo toki) li ken kama `Node`. kama ni li open tawa jan ale anu li lawa — ni li
  wile sona pi lawa kulupu, li pini ala.

## 9. nanpa pi lipu (versioning)

`v0` li awen ala. `domain tag` li jo e `v0`. ijo pi sike nanpa li ante suli la
`domain tag` li kama nanpa sin. jan kepeken MUST toki e ni tawa `domain tag` sona
ala: "toki ni li ken ala kama lon."

## 10. ijo pi lukin (test vectors)

`testdata/vectors.json` li **ijo pi lukin lawa** pi lipu ni. ilo ante li pona lon
nasin pi sike nanpa en nimi pana **lon tenpo ni taso**: ona li pali sin e nanpa
ale pi lipu ni, sike kepeken sike, tan ijo sama. ilo Go nanpa wan li pali e lipu
ni; ilo ante pi ilo poka ala li lukin e ona. ilo pi toki ante MUST sama e ona.
lipu pi toki (kipisi 3, 4) en lipu pi nanpa li ante la ni li **pakala pi lipu**;
jan o pona e lipu. jan o ante ala e ilo.

- **`primitives`** — sike nanpa pi ijo nanpa ale. jan pali li ken lukin e ilo ona
  tan anpa tawa sewi.
- **`messages`** — toki ale pi nimi pana: `name`, `domain_tag`, `signer`,
  `seed_hex`, `public_key_hex`, `fields`, `canonical_bytes_hex`, `signature_hex`.
- **`operator_messages`** — toki tu wan pi kipisi 11. ona li jo e `keys` (nimi
  len luka tu), `idx0`, `idx1`, en nimi pana **tu**: `sig0_hex` en `sig1_hex`.
  nimi pana tu li lon `canonical_bytes_hex` **sama**. (nimi len pi len — `seed` —
  li lon lipu tan ni taso: lipu o ken pali sin. kulupu li jo **ala** e nimi len pi
  len pi jan `operator`.)

ilo pi lukin li lon `vectors/vectors_test.go`. ona li pali sin e ijo ale tan ilo
pi tenpo ni, li wile e sama pi sike ale. ilo en lipu li ante la pali li pakala.
sina wile e ante la o kepeken `go test ./vectors -update`.

**ijo pi lukin: `Heartbeat`.** ijo: `NodeID = "node-alpha"`, `SentAt =`
`1700000050500000000`, `Seq = 43`.

```
16 736f686f636c6f75642f6865617274626561742f7630 0a 6e6f64652d616c706861 17979d09f832d900 000000000000002b
```

| sike | ijo nanpa | kon |
|---|---|---|
| `16` | uvarint | suli pi `domain tag` = **22**. |
| `73 6f 68 6f ... 2f 76 30` | UTF-8 | `sohocloud/heartbeat/v0`. |
| `0a` | uvarint | suli pi `NodeID` = **10**. |
| `6e 6f 64 65 2d 61 6c 70 68 61` | UTF-8 | `NodeID` = `node-alpha`. |
| `17 97 9d 09 f8 32 d9 00` | int64 | `SentAt` = `1700000050500000000`. |
| `00 00 00 00 00 00 00 2b` | uint64 | `Seq` = **43**. |

ijo luka wan ni taso li sike nanpa ale. ijo li lon ala lon open; ijo li lon ala
lon monsi pi `Seq`. nimi pana li lon ni taso.

## 11. nimi pi jan `operator` (Layer C)

kipisi 1–10 li lawa e toki pi `Node` en `Coordinator`. kipisi ni li pana e nimi
nanpa tu wan: **`operator`** — `Frontend` (Cloudy, sama) li awen e nimi pi jan
kulupu en `Node` lon insa ona, li pana e nimi **wan** tawa `Coordinator`. nimi ni
li **ante** tan nimi pi kipisi 2: toki pi kipisi 4 li **ante ala**, li kepeken
nimi `operator` **ala**.

jan `operator` li jo e nimi len **luka tu (7)** pi nasin `Ed25519` (nanpa `0..6`).
ona li pana e nimi pana **tu** (nanpa ante tu) lon toki ale — nasin **2-of-7**.
kulupu li awen e nimi len **lukin** (public) luka tu taso. **kulupu li jo ala e
nimi len pi len (private) pi jan `operator`.**

- **2-of-7 li awen tan ante pi nimi; ona li awen ala tan jan ike wan.** nimi len
  luka tu ale li lon ilo wan la jan ike wan li ken kama jo e ale. mi toki e ni tan
  ni: jan o pilin ike ala lon tenpo kama.
- **`Algo` li lon insa pi sike nanpa.** `v0` li ken e `ed25519` taso. jan lukin
  MUST weka e toki pi `Algo` sona ala, en MUST weka e toki pi `Algo` ante tan
  `algo` pi nimi len lon nanpa ale. nimi len pi jan `operator` MUST NOT jo e
  `Algo` ante lon tenpo sama. suli pi nimi pana li kama tan `Algo` (ni li ken e
  nasin sin lon tenpo kama: `domain tag` sin, `sohocloud/operator/v1-mldsa`
  sama).

**sike nanpa pi ijo `operator` — nasin ante li lon ala.**

- `OperatorID` en `Algo` li **string** pi kipisi 3. `Algo` MUST jo e nanpa pi
  suli, sama nimi ale (`ed25519` li kama `07 65 64 32 35 35 31 39`).
- `Idx0`, `Idx1`, `KeyIndex` li **`uint64` pi sike 8** (big-endian), sama `Seq`.
  ona li **varint ala**, li **sike wan ala**, kin la nanpa ona li lili
  (`0..6`). (nanpa `2` li kama `00 00 00 00 00 00 00 02`.)
- `Sig0` en `Sig1` li **lon ala** sike nanpa, sama `Signature` lon kipisi 4. nimi
  pana nanpa tu wan li lon ala. taso `Idx0` en `Idx1` li **lon insa** pi sike
  nanpa.

**wile pi lukin (kipisi 11.0).** jan lukin MUST pali e ni lon open:

1. weka, `Nonce` li lon ala anu li lili tan sike **16**. `Nonce` ala li **ken
   ala** toki e ni: "o lukin ala e poki pi toki sin."
2. weka, `Algo` li pona ala tawa nanpa lipu ni.
3. weka, `Idx0` li sama `Idx1`.
4. weka, nanpa li lon ala poka `0..6`.
5. weka, nimi len li lon ala lon nanpa wan, anu `algo` ona li sama ala `Algo`.
   **5a.** weka kin, nimi len lon `Idx0` en `Idx1` li **sama sike** — nanpa li
   ante la ni li ike kin. nimi len luka tu ale MUST ante. kama sin MUST NOT pana
   e nimi len lon nanpa tu. tan ni: 2-of-7 li ken kama 1-of-1 lon len, en nimi
   len pi len wan li ken pali e nimi pana tu. lukin ni li lukin e **nimi len**,
   li lukin ala e nanpa taso.
6. weka, suli pi nimi pana li sama ala suli pi `Algo`.
7. pali sin e sike nanpa. nimi pana **tu** li pona ala lon nanpa ona lon sike ni
   la o weka.

awen tan toki sin (anti-replay: `Seq` en poki `Nonce` lon poka pi
`(operator, coordinator)`) li pali pi `Coordinator`, li lon **sewi** pi ijo ni.
ona li lon ala sike nanpa, li lon ala lipu ni.

toki `operator` tu wan:

| toki | domain tag | ijo suli |
|---|---|---|
| `OperatorTransmission` | `sohocloud/operator/v0` | `Ts`, `Nonce` (≥16), `Seq`, `Algo`, `Idx0`, `Idx1` |
| `OperatorRotation` | `sohocloud/operator-rotate/v0` | `KeyIndex`, `NewPublicKey`, `Algo`, `Ts`, `Nonce`, `Seq`, `Idx0`, `Idx1` |
| `ConformanceResponse` | `sohocloud/operator-conformance/v0` | `Challenge`, `Algo`, `Idx0`, `Idx1` |

- **`OperatorRotation`** — jan `operator` (kulupu **ala**) li pali e nimi len sin.
  `NewPublicKey` li **lon insa** pi sike nanpa: ni la jan ike li ken ala pana e
  nimi len ona lon nasin ante. `Coordinator` MUST lukin e toki ni (kepeken nimi
  len **tu pi tenpo ni**) lon open pi awen e nimi sin lon `KeyIndex`. jan lukin
  MUST weka e `NewPublicKey` pi suli ike, en MUST weka e `NewPublicKey` sama nimi
  len lon nanpa **ante** (o lukin e 5a). nimi len sama lon `KeyIndex` **ona
  sama** li pona.
- **`ConformanceResponse`** — ona li jo e `Nonce` ala en `Seq` ala. tenpo sin li
  kama tan `Challenge` pi jan lukin. ni la wile nanpa wan (pi `Nonce`) li lon ala
  tawa ona; wile ante ale pi kipisi 11.0 li lon.

---

*Network Theory Applied Research Institute, Inc. — 501(c)(3) — EIN 92-3047136 — info@ntari.org*
