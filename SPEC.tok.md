# sohocloud-protocol — lipu lawa pi nasin sama (toki pona)

> **toki lili pi lipu ni:** lipu ni li lipu pi kulupu, li tan nasin NTARI "P2-002". lipu suli li lipu Inli "SPEC.md" (tenpo pi 2026-10-05). ilo sona li pali e lipu ni; kulupu li wile lukin e ona kepeken nasin "P2-002" kipisi "3.1".
>
> **Note:** This is a community rendering under NTARI policy P2-002. The complete document is the English original SPEC.md (snapshot 2026-10-05). Machine-assisted draft pending community review per P2-002 section 3.1.
>
> sina lukin e pakala lon toki ni la o pona e ona: o pali e "fork" lon
> https://github.com/NTARI-RAND/sohocloud-protocol, o pana e "pull request". pana sina
> li pona tawa mi mute.
>
> **lipu Inli li lawa.** lipu ni li tawa lukin, li tawa ala sona pi nasin lawa: lipu ni li ante tan `SPEC.md` la lipu Inli li lawa. nimi lawa (MUST, MUST NOT, SHOULD, SHOULD NOT, MAY), nimi pi poki sona, nimi sitelen en nimi ilo li awen lon toki Inli, tan ni: ona li lipu lawa pi nasin sama.

---

nanpa nasin: **v0 (UNSTABLE)** — selo pi toki en nasin pi nanpa lili pi nasin wan li ken
ante, li pana ala e toki awen pi ken kepeken sama, tawa tenpo pi v1.

lipu ni li lawa (normative), li ijo pi pali ilo (build deliverable); ona li ijo pi pilin
kama pini ala. jan li ken ala pali e ilo sin tan lipu lawa pi ken lukin (li wile ala lukin
e lipu ilo pi kulupu ilo ni) la ken AGPL pi pali sin e kulupu linja (fork) li jo e ala.
ijo ale lon anpa ni pi sitelen MUST / MUST NOT li wile pi nasin sama; ante li pakala e
wile wan la ona li kipisi lili pi nasin insa ni ala, ona li nasin insa ante.

---

## 1. ma pi lipu ni

nasin insa ni li lawa e **kipisi pi pali wan lon kipisi ilo** (substrate coordination
layer): nasin seme la jan li sona e ilo Node, nasin seme la ilo Node li toki e pali ni:
ona li wile pali e ona, nasin seme la jan li pana e pali tawa ona, li kama sona e pali
pini tan ona. jan tu wan li toki kepeken nasin ni:

- **Node** — ilo pi jo jan: ona li pana e ken pi pali nanpa (compute) anu sitelen lon ijo
  (physical print), anu ona tu. ilo Node li jo e nimi ona, li sitelen e nimi ona lon ijo
  ale ni: ona li pana e ona. ona li lawa e ona sama: ona li wile e ilo insa seme la ona li
  pali tawa ona, li ken weka tan ilo insa wan wan ale.
- **Coordinator** — ilo insa: ona li wan e lipu pi ken pali tawa pali, li awen e sona pi
  pali pini lon lipu. ona li pali; ilo ante li ken kama lon ma ona (pluggable). ilo
  SoHoLINK li ilo insa pi lipu sama (reference); lupa lukin MAY pali e ilo insa ona, anu
  ona MAY wan e pali tawa ilo Node lon nasin poka.
- **Frontend** — lupa lukin: ma ni la jan kulupu li esun. ona li kepeken pali wan kepeken
  lupa sama.

**ijo ni li lon ala tan wile.** nasin insa ni li esun pi jan kulupu pi nasin JFA ala. ona
li toki ala e mani awen (escrow), e mani tan jan kulupu, e toki awen pi sona pona jan, e
lipu awen tan toki pi jan tu (dialog-sealed record), e lawa pi utala. ijo ni li ijo pi
lupa lukin wan wan. mani pi ma lawa (fiat) ni: ilo insa li pana e ona tawa ilo Node ona
(sama mani pana Stripe pi ilo SoHoLINK) li esun pi *kipisi ilo*; ona kin li lon ala lipu
ni.

---

## 2. nimi en nasin wan ona

nimi pi ilo Node li `NodeID` — linja sitelen awen; ilo Node li wile e ona tawa ona sama.

nimi pali (workload identity) pi ilo Node li SPIFFE SVID; ona li kama lon nasin ni taso:

```
/node/<NodeID>
```

ilo insa li pana e ken tawa wile, li toki e ni: wile ni li tan ilo Node. ona li kama sona e ni kepeken nasin ni:
ona li lukin sama e nasin SPIFFE pi jan toki (lon pi nasin ni li kama sona) e
`/node/<NodeID>` tawa `NodeID` lon wile.

- nimi SPIFFE li lon ala la ilo insa MUST weka e wile kepeken **401**.
- nimi li lon, taso nasin ona li sama ala `/node/<NodeID>` la ilo insa MUST weka e wile
  kepeken **403**.
- ilo insa MUST NOT kama jo e "SVID pona ale" sama ken tawa wile pi nimi pi ilo Node ante.
  ni li nasin pi jan ike pi kepeken nimi ante (impersonation).
- nasin wan MUST pali e linja sitelen kepeken nasin pi ante ala; ona li lukin ala e lipu
  pi ken ante tawa ilo Node wan wan lon tenpo ale. lukin li awen lon poki pi sona ala lon tenpo ale la
  lukin ni li pali ala; jan MUST NOT pilin e ni: ona li selo awen.

nasin insa li pana ala e ilo open pi jan ale (public keys). ilo insa li kama jo e ilo open
pi lukin pi ilo Node lon nasin ante (lipu pi kama, sona SVID, anu lipu nimi).

---

## 3. nasin wan pi nanpa lili (tawa sitelen nimi)

toki ale pi sitelen nimi li jo e sitelen nimi lon **nanpa lili pi nasin wan** ona; nanpa
lili ni li WEKA e poki `Signature`. sitelen nimi li **ed25519**.

sitelen nimi li lon ala JSON, li lon ala nasin tawa ante. ona li lon nasin ni: ona li
ante ala, li jo e suli lon open; tan ni la sitelen nimi li awen sama lon ilo ale, lon
nanpa nasin ale pi toki Go, lon toki ilo ale.

ijo open, lon nasin sama pi poki pi toki ale lon §4:

- **unsigned LEB128 varint** — ijo open pi suli/nanpa; `string`, `bytes`, en `repeated`
  li kepeken ona. ona MUST kama lon nasin ni, tan toki ilo ala: o kama jo e nanpa (ona li
  jo ala e nanpa anpa); o pana e nanpa bit 7 lon nanpa lili wan, kulupu pi suli lili li nanpa
  wan (kulupu little-endian); o pana e nanpa bit sewi (`0x80`) lon nanpa lili ale ANTE nanpa
  lili pini, tawa toki ni: ijo li awen; nanpa bit sewi pi nanpa lili pini li ala. nanpa `0` li
  kama nanpa lili wan `0x00`. ni li sama nasin pi `encoding/binary.PutUvarint` pi toki
  Go; nimi Go li lukin sama taso, li toki lawa ala. (pali: `300` = `0b100101100` → nanpa bit 7
  anpa `0101100` kepeken awen → `0xAC`, nanpa bit sewi `0000010` lon pini → `0x02`, la ni li
  `AC 02`.)
- **domain tag** — nanpa lili pi nasin wan pi toki ale li open kepeken linja sitelen pi
  nimi ma (domain tag) ante (o lukin e §4). nimi ma MUST kama kepeken nasin pi **string**
  lon anpa — ni li ni: nanpa unsigned LEB128 varint pi suli UTF-8 ona, la nanpa lili UTF-8 ona —
  sama poki linja sitelen ante ale; ona li jo ala e nanpa ala lon pini (null terminator), li jo ala
  e nasin ante. nimi ma li ante la sitelen nimi lon kule toki wan li ken ala tawa kule
  toki ante.
- **string** — nanpa unsigned LEB128 varint pi suli (lon nanpa lili), la nanpa lili UTF-8. linja
  sitelen ala MUST kama nanpa lili wan `0x00` (suli ala, ijo ala).
- **bytes** — nanpa unsigned LEB128 varint pi suli (lon nanpa lili), la nanpa lili. (poki
  `Signature` li kama ala lon ona lon tenpo ale; ona li weka.)
- **bool** — nanpa lili wan: `0x00` (lon ala) anu `0x01` (lon) taso; nanpa lili ante ala li
  pona.
- **int64 / uint64** — nanpa lili 8 lon tenpo ale, big-endian (nanpa lili pi suli nanpa
  wan li lon open). `int64` MUST kama sama `uint64` pi nasin two's-complement: nanpa li
  lili tan ala la o pana e 2⁶⁴ tawa ona, la o pana e nanpa lili 8 big-endian. tan ni la
  `-1` → `ff ff ff ff ff ff ff ff`, `-2` → `ff ff ff ff ff ff ff fe`, en `int64` pi lili
  nanpa wan (`-9223372036854775808`) → `80 00 00 00 00 00 00 00`. nasin sign-magnitude
  ala, nasin zigzag ala, suli ante ala.
- **time** — tenpo kepeken `int64` UTC Unix **nanoseconds** (kepeken nasin int64 lon sewi,
  big-endian, nanpa lili 8). sona ma (location) en sona pi ilo tenpo pi kama suli taso
  (monotonic clock) li weka. suli pi nanoseconds li awen sama; nanpa MUST NOT kama lili
  tawa micro- anu milli-seconds. ni li awen lon tawa en kama pi JSON, li ante ala tan ma
  tenpo (time zone). tenpo li lon ala insa pi ken pi `int64` UTC Unix nanoseconds (sama
  sike suno 1678–2262) la ona li lon ala ma pi v0, li MUST kama pakala pi sitelen nanpa;
  ona li kama ala sike (wrapped), li kama ala lon selo (clamped). ilo sitelen pi lipu
  sama (reference encoder) li wawa e ni: `canon.Buffer.Time` li pana ala e ijo, li awen e
  `canon.ErrTimeOutOfRange` (sina ken kama jo e ona kepeken `Buffer.Err()`); ona li pana
  ala e nanpa sike. tan ni la tenpo tu pi ante pi sike suno ~584 li ken ala kama nanpa
  lili sama. ilo pana MUST weka e ona kepeken `canon.ValidTime` lon sinpin
  pi sitelen nimi; ilo lukin/ilo sona MUST weka e toki la nanpa tenpo ona li pakala lon
  lukin ni.
- **repeated field** — nanpa unsigned LEB128 varint pi *mute* (nanpa pi ijo ale) li kama NANPA WAN, la
  poki pi ijo ale li kama lon nasin, lon insa ale, kepeken nasin pi ijo open lon sewi.
  ijo pi poki mute li linja sitelen la ona li jo e suli lon open sama linja sitelen ante
  ale; poki pi kulupu insa (nested struct) li kama lon nasin pi ijo ni lon §4. ijo ala li
  lon insa pi ijo tu, ijo ala li lon monsi pi ijo pini. nanpa `0` li nanpa lili wan
  `0x00`, la ijo ala.

ilo lukin li pali sin e nanpa lili pi nasin wan tan poki pi toki kama, li lukin e sitelen
nimi ed25519 kepeken ilo open pi jan ale (ona li kama jo e ona). suli pi sitelen nimi li
`ed25519.SignatureSize` (64) ala la ilo lukin MUST weka e ona; ni li kama lon open pi
lukin.

nanpa lili pi nasin wan pi toki li **sama** e ni taso: nimi ma, la poki ale lon toki
`Order:` pi toki ni lon §4, lon nasin ni; nasin pi ijo open lon sewi li sitelen e poki
ale. selo ala li lon: nanpa lili pi nanpa nasin ala li lon open, nanpa suli ale ala li
lon selo pi toki, nanpa lili ala li lon monsi pi poki pini. kulupu `Order:` lon §4 li jo
e ale, li lawa — **poki pi sitelen nimi** pi toki taso li weka, en poki li lon ala ni la
ona li lon ala nanpa lili pi nasin wan. tawa toki lawa luka luka tu (12) lon §4 la poki
weka wan li `Signature`. toki pi ilo lawa (operator) pi kipisi Layer-C lon §11 li jo e
sitelen nimi **tu** (`Sig0` en `Sig1`) lon nanpa lili sama pi nasin wan; ona tu li weka
sama `Signature` lon ni, en sama `Signature` la ona tu li lon ala toki `Order:`. tan ni
la ilo lukin li pana ala e poki pi sitelen nimi lon nanpa lili ni: ona li pali sin e ona;
toki li jo e sitelen nimi wan anu tu la ni li sama.

---

## 4. toki

tawa toki ale: nimi ma ona, poki ona lon nasin wan, en ilo seme li sitelen e nimi lon
ona. nanpa tenpo ale li kepeken nasin `time` lon §3.

toki `Order:` lon anpa li lawa WAN TASO pi nasin poki, en nasin pi poki insa (sama la
`Capacity.VCPUs` li lon sinpin pi `Capacity.MemMB`, en ilo sitelen ale li `Kind` la
`Model`). jan pali MUST sitelen e poki lon nasin `Order:`, li MUST NOT kama sona e nasin
tan ijo ante. sama la lipu lukin `testdata/vectors.json` (§10) li sitelen e `fields` pi
toki ale lon **nasin pi sitelen open (alphabetical) tawa ken lukin, li nasin ALA pi
sitelen nanpa wan** — sina kepeken nasin pi nimi JSON la nanpa lili li ike, sitelen nimi
li ike.

### 4.1 CapabilityListing — ilo Node li sitelen e nimi
nimi ma: `sohocloud/listing/v0`

Order: `NodeID` (string), `Class` (string), `Printers` (repeated: `Kind` string,
`Model` string), `GPUs` (repeated: `API` string, `Model` string, `VRAMMB`
int64), `Capacity.VCPUs` (int64), `Capacity.MemMB` (int64),
`Capacity.DiskMB` (int64), `Capacity.StorageCommitMB` (int64),
`Capacity.PrintQPS` (int64), `OptIn.Compute` (bool), `OptIn.Print` (bool),
`OptIn.Storage` (bool), `IssuedAt` (time), `Seq` (uint64).

`Class` ∈ {`micro`, `standard`, `server`}. `Kind` ∈ {`traditional`, `threed`}.
`API` ∈ {`vulkan`, `nnapi`, `cuda`, `metal`}.

`Capacity.DiskMB` li ma poki pi tenpo lili tawa pali lon tenpo pali ona;
`Capacity.StorageCommitMB` li ma poki pi tenpo suli; ilo Node li toki e ni: ona li awen
e ona tawa kulupu linja (awen e kipisi sona). ona tu li ante tan wile; tan ni la ilo Node
li ken pana e wan, li wile ala pana e ante. nasin seme la sona awen li kama len, li kama
kipisi, li kama lukin? ni li wile pi lupa lukin/ilo pali (agent) lon ala nasin insa ni —
ni li pali wan taso; linja li tawa ala e sona awen lon tenpo ale.

ilo Node li toki e ilo GPU la ni li wile ona tawa pali GPU: lipu pi ken pali li jo ala e
`GPUs` la ona li kama jo ala e pali GPU; ilo Node li weka e ilo GPU kepeken ni: ona li
weka e ona lon lipu pi ken pali kama.

`Seq` li kama suli taso lon ilo Node wan. `Seq` li suli ala tan `Seq` pini pi ilo Node ni
(ilo insa li lukin e ona) la ilo insa MUST weka e lipu pi ken pali (awen tan pana sin /
tan tawa monsi).

`OptIn` li **toki wile taso** (advisory) (o lukin e §5.1).

### 4.2 Heartbeat — ilo Node li sitelen e nimi
nimi ma: `sohocloud/heartbeat/v0`

Order: `NodeID` (string), `SentAt` (time), `Seq` (uint64).

nasin `Seq` sama pi kama suli taso, sama §4.1.

### 4.3 Assignment — ilo insa li sitelen e nimi
nimi ma: `sohocloud/assignment/v0`

Order: `JobID` (string), `NodeID` (string), `Spec.Workload` (string),
`Spec.Image` (string), `Spec.Args` (repeated string), `Spec.PrinterKind`
(string), `Spec.GPUAPI` (string), `Spec.GPUMinVRAMMB` (int64),
`Fee.ContributorShareBps` (int64), `Fee.PlatformFeeBps` (int64),
`OfferedAt` (time).

`Spec.GPUAPI` ∈ {``, `vulkan`, `nnapi`, `cuda`, `metal`} — toki wile pi nasin tawa
(routing hint), sama `PrinterKind`; ona li jo e ala la ilo GPU li wile ala, la
`GPUMinVRAMMB` MUST kama `0`.

nasin mani pana li lon insa toki; tan ni la ilo Node li lukin e kipisi mani, la ona li
pana e toki awen. lon nasin pull la toki ante pi wile pona li lon ala.

### 4.4 Decline — ilo Node li sitelen e nimi
nimi ma: `sohocloud/decline/v0`

Order: `JobID` (string), `NodeID` (string), `Reason` (string), `DeclinedAt`
(time).

`Reason` ∈ {`local_policy`, `capacity`, `unavailable`}. `local_policy` li nasin pi weka
wile (opt-out) (o lukin e §5.1).

### 4.5 JobReport — ilo Node li sitelen e nimi
nimi ma: `sohocloud/jobreport/v0`

Order: `JobID` (string), `NodeID` (string), `ExitCode` (int64), `FailureCause`
(string), `TmpfsExhausted` (bool), `StartedAt` (time), `FinishedAt` (time).

toki pali ni (sitelen nimi li lon ona) li ijo lon; nanpa pi kepeken (metering) en mani
pana (payout) li kama tan ona. poki ona li sama nasin sona pi pali pini pi ilo pali
(executor completion parse) ni: ona li lon ilo SoHoLINK lon tenpo ni.

### 4.6 FeeDeclaration — ilo insa li sitelen e nimi
nimi ma: `sohocloud/fee/v0`

Order: `CoordinatorID` (string), `Terms.ContributorShareBps` (int64),
`Terms.PlatformFeeBps` (int64), `EffectiveAt` (time), `Seq` (uint64).

`ContributorShareBps + PlatformFeeBps` SHOULD sama `10000`.

### 4.7 StorageLease — ilo insa li sitelen e nimi
nimi ma: `sohocloud/lease/v0`

Order: `LeaseID` (string), `NodeID` (string), `ShardRef` (bytes),
`SizeClass` (int64), `Fee.ContributorShareBps` (int64),
`Fee.PlatformFeeBps` (int64), `IssuedAt` (time), `ExpiresAt` (time),
`Seq` (uint64).

poki awen li esun pi tenpo awen (LEASE), li pali ala: ona li wile pi tenpo awen: o awen e
kipisi sona wan (ona li len) lon tenpo ni: pini ona li lon. `ShardRef` li nimi pi ijo
insa (content address) pi nanpa lili 32; ona li len (opaque). `SizeClass` li suli pi ijo lon kulupu nanpa
(quantized) — nasin insa MUST NOT tawa e suli lon pi ijo, e sona awen; len, ijo pi suli
namako (padding), en kipisi li wile pi lupa lukin lon sewi pi insa lili ni (waist). kama
sin li `StorageLease` sin tawa `LeaseID` sama, kepeken `Seq` suli (ona li suli tan `Seq` pini; nasin sama pi tawa
monsi lon §4.1). nasin mani pana li lon insa toki, sama §4.3.

### 4.8 LeaseDecline — ilo Node li sitelen e nimi
nimi ma: `sohocloud/lease-decline/v0`

Order: `LeaseID` (string), `NodeID` (string), `Reason` (string),
`DeclinedAt` (time).

`Reason` ∈ {`local_policy`, `capacity`, `unavailable`}; `local_policy` li nasin pi weka
wile (§5.1) tawa poki awen, sama §4.4 tawa pali.

### 4.9 LeaseRelease — ilo Node li sitelen e nimi
nimi ma: `sohocloud/lease-release/v0`

Order: `LeaseID` (string), `NodeID` (string), `ReleasedAt` (time).

ilo Node li ken pini e awen lon tenpo ale (lawa pi ona sama li jo kin e ken weka). toki
pini (sitelen nimi li lon ona) li pini e nanpa pi kepeken; pana sin e kipisi sona lon ma
ante li wile pi lupa lukin.

### 4.10 ProofChallenge / ProofResponse — ilo Node li sitelen e nimi lon toki pana
nimi ma (toki pana): `sohocloud/proof/v0`

`ProofChallenge{LeaseID, Offset, Length, Nonce (16 bytes), IssuedAt}` li jo ALA e
sitelen nimi: ilo Node li kama jo e toki lukin kepeken alasa lon linja pi lon sona
(authenticated channel) (nasin pull), en toki lukin li pana ala e wile tawa jan ala. ijo
pi sitelen nimi li toki pana.

Order (toki pana): `LeaseID` (string), `NodeID` (string), `Offset` (int64),
`Length` (int64), `Nonce` (bytes), `Digest` (bytes), `RespondedAt` (time).

`Digest` MUST kama kepeken nasin ni taso:
`SHA-256(Nonce || uint64be(Offset) || uint64be(Length) || sealed[Offset : Offset+Length])`
lon nanpa lili pi kipisi sona len. toki pana li toki sin e kipisi lukin en nanpa pi
kepeken wan (nonce); tan ni la ona li ken lon ona taso sama ijo lon pi nanpa kepeken, en
jan li ken ala toki e ni: "mi pana ala e ona" — ona li sama §4.5 tawa poki awen. ilo
lukin li kama jo pini e `(LeaseID, Nonce)` pi toki pana la ona MUST weka e ona: nanpa pi
kepeken wan li ken kepeken wan taso; ni li anpa e pana sin pi toki pana pini (jan li
awen e ona).

### 4.11 KeyRotation — ilo open pi TENPO NI pi ilo Node li sitelen e nimi
nimi ma: `sohocloud/node-rotate/v0`

Order: `NodeID` (string), `NewPublicKey` (bytes), `Algo` (string),
`RotatedAt` (time), `Seq` (uint64).

jo pi ilo open weka li pana e ken tawa toki e ilo open kama. ilo lukin MUST lukin e
sitelen nimi kepeken ilo open ni: ona li jo e ona tawa ilo Node lon tenpo ni (ilo open
kama li sitelen e nimi lon ante la ni li pana e ona sama, li MUST pakala), MUST wawa e
`Seq` pi kama suli taso lon ilo Node wan, en MUST weka e ante la `NewPublicKey` ona li
ilo open pona ala tawa `Algo` (tawa `ed25519`: nanpa lili 32 ala) anu `Algo` ona li sona
ala — ante la ilo lukin li open lukin e toki kama kepeken `NewPublicKey` ike, li pakala.
lukin ni ale li pona la, ni taso la, ilo lukin li open lukin e toki kama pi ilo Node
kepeken `NewPublicKey`. `Algo` ∈ {`ed25519`} lon v0.

### 4.12 KeyRevocation — ilo open li sitelen e nimi: ona li kama PINI
nimi ma: `sohocloud/node-revoke/v0`

Order: `NodeID` (string), `RevokedPublicKey` (bytes), `RevokedAt` (time),
`Seq` (uint64).

ona li pini e ilo open, li jo ala e ilo open kama. ilo lukin MUST lukin e sitelen nimi
kepeken ilo open ni: ona li **pilin pona tawa ona lon tenpo ni** tawa `NodeID`, li MUST
wile e ni: `RevokedPublicKey` li sama ilo open ni. ilo lukin li kute e pini lon ilo open
sama taso ni: pini li toki e nimi ona, li pini e ona. ni li nasin wan pi anpa e kepeken
ike: jan ante li ken sitelen e nimi ona sama lon pini, li pana e `NodeID` pi jan ni: ona
li wile pakala e ona; jan ante li pana kin e ilo open ona sama lon `RevokedPublicKey`.
taso ilo insa li pilin pona ala tawa ilo open ni tawa jan pakala; tan ni la lukin sama li
pakala, li weka e pini — jan li jo ala e ilo open pi tenpo ni pi ilo Node la ona li ken
ala weka e ilo Node tan linja. (lipu pini li toki e ni: "o kute lon tenpo ale"; nimi ni
li weka — wan tawa ilo open pi pilin pona li wile.) ilo open pi tenpo ni li lon la, sona
pona pi pini li ni: "o pilin pona ala tawa ilo open ni". kama sin lon monsi pi pini li
lon nasin ante, sama nasin pi kama nanpa wan; ona li toki linja ala ni: jan pi kama jo
ike e ilo open li ken pali e ona.

---

## 5. nasin awen

### 5.1 weka wile li kama wawa lon ma poka, li kama ala lon linja
`CapabilityListing.OptIn` en toki wile pi nasin tawa lon `JobSpec` li **toki wile taso**.
ilo insa li selo awen ala tawa weka wile. ilo Node MUST wawa e weka wile kepeken lipu ken
ona (allowlist) pi pilin pona lon ma ona, li MUST weka e pali la nasin ona lon ma ona li
ken ala e ona — kepeken `Decline{Reason: local_policy}` — ona li toki e seme lon tenpo
pini, ilo insa li pana e seme, la ni li suli ala.

### 5.2 nasin pi wan pali li lon ala nasin insa
lupa `Coordinator` li pana ala e nasin nanpa pi pona (ranking) anu nanpa (scoring). nasin
seme la ilo insa li pana e pali seme tawa ilo Node seme? ni li len lon ilo ona. nasin
insa MUST NOT kama jo e poki sin pi lawa e wan pali, tan ni: ni li wan e ilo kepeken
tawa nasin pi ilo insa wan, li pakala e ken weka; lupa li lon tawa awen e ken weka ni.

### 5.3 mani pana li ken lukin, li ante ala tawa tenpo pini
nasin mani pana pi ilo insa li kama lon open sama `FeeDeclaration` (sitelen nimi li lon,
nanpa tenpo li lon). nasin MUST NOT ante tawa tenpo pini lon pali pi pana pini; ante li
toki sin (sitelen nimi li lon ona) kepeken `EffectiveAt` pi tenpo kama. tan ni la jan li
ken utala e mani pana; ona li len ala (ijo pi pini ala #7).

### 5.4 ilo insa li pali, li tomo insa wan ala
ijo ala pi wile pi ilo Node anu lupa lukin li ken tawa *taso* lon ilo insa wan. ilo Node
MAY pali tawa ilo insa mute; lupa lukin MAY pali e ilo insa ona anu wan e pali lon nasin
poka. pali ilo la jan li ken ala weka tan ilo insa wan la, ni li kama sin pi lawa wan
(recentralization), li MUST jo e sitelen pi nimi ni, li jo e nasin tawa ijo ante.

### 5.5 nanpa nasin pi kama suli taso
nanpa `Seq` pi ilo Node (lipu pi ken pali, toki pi lon) li kama suli taso lon ilo Node
wan. ilo insa MUST weka e `Seq` pi kama suli ala.

---

## 6. pali pi ilo insa (nasin pull)

nasin li **pull** (ilo Node li kama jo): ilo insa li toki open ala tawa ilo Node lon tenpo
ale. pali:

| pali            | nasin tawa              | insa / pini                          |
|-----------------|-------------------------|--------------------------------------|
| `SubmitListing` | ilo Node → ilo insa     | `CapabilityListing`                  |
| `Heartbeat`     | ilo Node → ilo insa     | `Heartbeat`                          |
| `PollJobs`      | ilo Node → ilo insa     | `NodeID` → `[]Assignment`            |
| `Decline`       | ilo Node → ilo insa     | `Decline`                            |
| `ReportJob`     | ilo Node → ilo insa     | `JobReport`                          |
| `Fees`          | jan ale → ilo insa      | → `FeeDeclaration`                   |

ilo insa MUST kepeken nasin wan §2 tawa `PollJobs` (en tawa pali ale pi nimi pi ilo Node)
lon open pi pali ona. ilo insa li open ala e toki tawa ilo Node lon tenpo ale; tan ni la
pali pi poka Node li lon ala, li wile ala e toki.

---

## 7. nasin tawa

HTTP+JSON li nasin tawa **pi lipu sama** (reference) (`transport/httpjson`), li linja sama
ALA. kulupu ilo insa (core package) ala li kepeken ona. ilo pi nasin sama MAY kepeken
nasin tawa ale; ona li pali e pali pi ilo insa kepeken nanpa lili pi nasin wan lon §3.
JSON li nasin tawa taso; sitelen nimi li kama ala lon JSON lon tenpo ale; tan ni la nasin
tawa ante li ken pali wan, li wile ala toki sin e sitelen nimi.

lupa HTTP+JSON pi lipu sama:

```
POST /v0/listing     body: CapabilityListing   -> 204
POST /v0/heartbeat   body: Heartbeat           -> 204
GET  /v0/jobs?node_id=<id>                      -> 200 []Assignment
POST /v0/decline     body: Decline             -> 204
POST /v0/report      body: JobReport           -> 204
GET  /v0/fees                                   -> 200 FeeDeclaration
```

nasin wan SPIFFE (§2) li kama tan ilo lon insa (middleware) lon sinpin pi ilo pali
(handler); ilo pali pi lipu sama li linja e nasin en JSON taso, li ilo pi lukin nimi ala.

---

## 8. ijo pi pini ala lon poka pi nasin insa ni

toki lon pi ijo pi pini ala li nasin, li ijo pi pilin ike ala.

- **#7 ilo nanpa pi jo sina li pana e ken weka pi ilo, li pana ala e ken weka pi mani.**
  #7 li toki e kipisi ni. §5.4 (ilo insa li pali) en §5.3 (mani pana pi ken lukin) li
  utala e kama sin pi lawa wan pi ilo insa; open lete (cold-start) li awen sama ijo pi
  pini ala, en tawa pi sona pona jan (reputation portability) li ken pona e ona, taso ona
  li wile pi lupa lukin/kipisi lawa, li wile ala pi nasin insa ni.
- **#6 lon pi pali nanpa li lon ala ma pi lipu ni.** wan pali anu mani pi ilo insa li pona
  ala pona? lipu ni li toki ala e lon ona; ona li pana e ken lukin taso (§5.2, §5.3).
  `anchor/` li awen e ma tawa awen pi *toki* pali (sona `JobReport`) lon lipu pi jan
  lukin; tan ni la jan li ken lukin e ona. ona li ijo open (stub) pi nimi sitelen; jan li
  pali ala e ona.
- **#4 kama kulupu / open.** ilo Node li pana e nimi ona sama, li sitelen e nimi ona sama
  (§2); tan ni la ilo pi ken lili (sama ilo toki lili) li ken sitelen e ona sama sama ilo
  Node. kama pi ilo Node li open anu li jo e lupa? ni li wile pi kipisi lawa/pilin; ona li
  awen open.

## 9. nanpa nasin (versioning)

`v0` li awen ala. nimi ma li jo e `v0`; ante pakala tawa poki pi nasin wan pi toki li
suli e nimi ma ona. ilo kepeken MUST lukin e toki pi nimi ma sona ala sama toki pi ken
lukin ala.

---

## 10. nanpa lukin (test vectors)

`testdata/vectors.json` li **lipu lukin lawa pi nasin sama** tawa lipu ni. ilo ante (tan
jan ante) li lon nasin sama tawa sitelen nanpa en sitelen nimi la ni taso li lon: ona li
pali sin e nanpa ale lon lipu ni tan ijo open sama, lon nanpa lili ale. ilo sitelen Go
pi lipu sama li pali e lipu lukin; ilo ante pi lipu ilo taso (stdlib-only) li lukin sin e
ona. ilo ante pi nasin sama MUST sama e ona lon ale. toki lon §3/§4 en lipu lukin li toki
ante la ni li pakala lipu, li wile kama pona; ona li pana ala e ken ante — taso lipu
lukin li toki pali pi §3/§4.

nanpa hex ale lon lipu lukin li sitelen lili (lowercase). ona li toki e ona sama: `note`
lon sewi en kulupu `encoding` li toki sin e nasin open §3 tawa jan ni: ona li open e JSON
nanpa wan.

### 10.1 selo

- **`primitives`** — sitelen nanpa pi ijo open ale §3, lon poka ante; tan ni la jan pali
  li ken lukin e ilo sitelen ona tan anpa tawa sewi, lon sinpin pi wan e toki. ijo
  ale li jo e `kind` (`uvarint`, `int64`, `uint64`, `bool`, `string`, `time`), e `input`
  (jan li ken sona e ona), e `bytes_hex` (wile). ona li jo e selo ni: ona li lawa e
  nasin: nanpa varint lon selo ale pi nanpa lili awen
  (`0, 1, 127, 128, 255, 256, 300, 16384, 1073741824`); `int64` lon `-1`, `-2`,
  `MaxInt64`, en `MinInt64`; `uint64` lon `MaxUint64`; nanpa bool tu; linja sitelen ala en
  linja sitelen UTF-8 pi nanpa lili mute; en `time` lon open pi tenpo Unix (epoch) en lon
  tenpo nanoseconds awen wan.

- **`messages`** — ijo wan tawa kule ale pi toki pi sitelen nimi (toki ale pi sitelen nimi lon §4). ijo
  ale li awen e: `name`, `domain_tag`, `signer` (`node` anu `coordinator`), `seed_hex`
  ed25519 pi nanpa lili 32 en `public_key_hex` tan ona, `fields` pi toki,
  `canonical_bytes_hex` (ijo pana pi `CanonicalBytes()`, `Signature` li weka), en
  `signature_hex` (ed25519 lon nanpa lili ni taso pi nasin wan).

- **`operator_messages`** — ijo wan tawa kule ale pi toki pi ilo lawa Layer-C (ale tu wan
  lon §11). toki pi ilo lawa li 2-of-7; tan ni la ijo ale li ante tan ijo `messages`
  lawa: `signer` li `operator` lon tenpo ale; kulupu `keys` li jo e ilo open pi jan ale
  luka tu (7) (`index`, `seed_hex`, `public_key_hex`, `algo` — nanpa open (seed) li lon
  TAN NI TASO: lipu lukin li ken lon ona taso, li ken pali sin; kulupu linja li jo ala e
  ilo open len lon tenpo ale); `idx0`/`idx1` li nanpa tu pi sitelen nimi; en sitelen nimi
  tu `sig0_hex` en `sig1_hex` li lon, li `signature_hex` wan ala; ona TU li lon
  `canonical_bytes_hex` SAMA. ilo pi nasin sama li pali sin e `canonical_bytes_hex` tan
  nasin poki §11, li lukin e sitelen nimi tu lon nanpa ona.

  > **pakala pi nasin poki.** `fields` pi toki ale li sitelen e nimi lon
  > **nasin pi sitelen open (alphabetical) tawa ken lukin, li nasin ALA pi nasin wan.**
  > nasin wan li kama TAN toki `Order:` TASO pi toki ni lon §4. jan pali li sitelen e
  > poki lon nasin pi nimi JSON la nanpa lili li ike, sitelen nimi li ken ala kama pona
  > lon lukin.

### 10.2 nasin kepeken pi lipu lukin

1. o pali sin e `primitives[].bytes_hex` ale tan `input` ona kepeken §3.
2. tawa toki ale la o kama jo e ilo open tu tan `seed_hex` (ed25519 tan nanpa open pi
   nanpa lili 32 li ante ala), o lukin e ni: ona li pana e `public_key_hex`; o sitelen e
   poki lon `Order:` §4, o lukin e ni: sina kama jo e `canonical_bytes_hex`; o lukin e
   `signature_hex` kepeken `public_key_hex` lon nanpa lili sina (sina pali e ona sina sama).
   lukin pi sitelen nimi lon nanpa lili SINA li lukin sama wawa nanpa wan: ona li ken pona
   taso la nanpa lili sina li sama nanpa lili pi ilo pali lawa.

ilo lukin pi lipu sama li lon `vectors/vectors_test.go`. ona li pali sin e nanpa lukin
ale tan lipu ilo pi tenpo ni, li wawa e ni: ona li sama JSON awen lon nanpa lili ale; tan
ni la lipu ilo en lipu lukin li kama ante la pali ilo li pakala; kin la ona li lukin sin
e lipu awen, li lukin e sitelen nimi ale. o pali sin kepeken wile kepeken
`go test ./vectors -update` (anu nimi ante
`SOHOCLOUD_UPDATE_VECTORS` (env var)). nanpa lukin ale li kepeken nanpa open pi nanpa
lili sama (byte-fill seeds) en nanpa tenpo awen — `crypto/rand` li lon ala, `time.Now()`
li lon ala; tan ni la ijo pana li ante ala lon tenpo ale.

### 10.3 pali lukin: `Heartbeat`, lon nanpa lili ale

ni li nanpa lukin `Heartbeat` (`§4.2`); poki ale li kama sona, tan ni la jan pi pali sin
li ken lukin e sitelen nanpa ale tan open tawa pini. ijo open:
`NodeID = "node-alpha"`, `SentAt = ` Unix nanoseconds `1700000050500000000`,
`Seq = 43`.

nanpa lili pi nasin wan (`canonical_bytes_hex`):

```
16 736f686f636c6f75642f6865617274626561742f7630 0a 6e6f64652d616c706861 17979d09f832d900 000000000000002b
```

sona tan poka open tawa poka pini:

| nanpa lili | ijo open | sona |
|-------|-----------|---------|
| `16` | nanpa uvarint pi suli | suli pi nimi ma = `0x16` = **22** nanpa lili. |
| `73 6f 68 6f 63 6c 6f 75 64 2f 68 65 61 72 74 62 65 61 74 2f 76 30` | nanpa lili UTF-8 | nimi ma `sohocloud/heartbeat/v0` (nanpa lili 22). linja tu ni (suli en nanpa lili) li sitelen `string` pi nimi ma (§3). |
| `0a` | nanpa uvarint pi suli | suli `NodeID` = `0x0a` = **10** nanpa lili. |
| `6e 6f 64 65 2d 61 6c 70 68 61` | nanpa lili UTF-8 | `NodeID` = `node-alpha` (nanpa lili 10). |
| `17 97 9d 09 f8 32 d9 00` | int64, big-endian | `SentAt` sama UTC Unix nanoseconds. `0x17979d09f832d900` = `1700000050500000000`. ni li nasin `time` (§3): tenpo li kama `int64`, li kama nanpa lili 8 big-endian. |
| `00 00 00 00 00 00 00 2b` | uint64, big-endian | `Seq` = `0x2b` = **43**. nanpa lili 8 awen, nanpa varint ala. |

kulupu luka wan ni taso — nimi ma (suli + nanpa lili), `NodeID` (suli + nanpa lili),
`SentAt`, `Seq` — li wan la ni li linja ale pi nanpa lili pi nasin wan. selo ala li lon
sinpin pi nimi ma, nanpa lili ala li lon monsi `Seq`. sitelen nimi ed25519
(`signature_hex`) li lon linja ni taso; poki `Signature` li lon ala ona lon tenpo ale.

o lukin: nasin poki lon ni (`NodeID`, `SentAt`, `Seq`) li kepeken `Order:` §4.2; ona li
sama kin — tan ken taso, tan ni: toki ni li lili — nasin ni: jan lukin li sitelen e ona
lon nasin ni. tawa toki suli sama `CapabilityListing` la `Order:` §4 en nimi `fields` pi
nasin sitelen open lon `vectors.json` li ante; nasin §4 taso li nasin wan.

---

## 11. nimi pi ilo lawa (Layer C)

§1–§10 li lawa e toki pi **ilo Node** en **ilo insa**. kipisi ni li pana e nimi nanpa tu
wan; ona li ante ale tan ona tu: **ilo lawa** (operator) — lupa lukin (sama ilo Cloudy)
li pini e nimi pi jan kulupu en nimi pi ilo Node lon insa ona, li pana e nimi ken wan
(ona li ante lon tenpo) tawa ilo insa. ona li ante tan nimi pali pi ilo Node (§2) tan
wile: toki lawa pi pali wan (§4) li ante ALA, li jo ALA e sitelen nimi kepeken nimi ken
pi ilo lawa.

ilo lawa li jo e kulupu **luka tu** (7) pi ilo open tu Ed25519 (nanpa `0..6`), li sitelen e nimi lon
toki ale kepeken nanpa **tu ante** (nasin **2-of-7**). kulupu linja li awen e ilo open
luka tu **pi jan ale** taso tawa ilo lawa wan; **kulupu linja li jo ala e ilo open len
pi ilo lawa lon tenpo ale**, en nasin lukin ale lon anpa li kama jo e ilo open pi jan ale
taso.

- **2-of-7 li pona e awen tan kama ante (anti-substitution) e ante pi ilo open (rotation),
  li awen pi nanpa jan (threshold security) ALA.** ilo open len luka tu ale li lon ilo
  pali wan la, tomo ilo wan li pakala la jan ike li kama jo e ale luka tu. jan li tu ala
  e jo pi ilo open lon ma mute pi pilin pona la, sina o pilin ala e ni: 2-of-7 li awen
  tan pakala pi ilo open wan. mi toki e ni tan ni: jan lukin ala o pilin pona mute tawa
  ona.
- **wan pi nasin nanpa (Algorithm binding).** toki ale pi ilo lawa li wan e linja sitelen
  `Algo` pi nasin nanpa lon nanpa lili pi nasin wan ona. `v0` li ken e `ed25519` taso.
  ilo lukin MUST weka e toki la `Algo` ona li sona ala, en MUST weka e toki la `Algo` ona
  li sama ala `algo` awen pi ilo open lon nanpa ale pi sitelen nimi. kulupu ilo open pi
  tenpo ni pi ilo lawa MUST NOT jo e nasin nanpa mute. lukin pi suli pi sitelen nimi li
  lon nasin nanpa wan wan (`ed25519` = nanpa lili 64), li nanpa awen ala; tan ni la ante
  kama pi kulupu ale lon tenpo wan tawa nasin nanpa sin pi wawa tawa ilo nanpa "quantum"
  (post-quantum) (nimi ma SIN, sama `sohocloud/operator/v1-mldsa`) li kepeken nasin sama.
  sitelen nimi li ken ala tawa nasin nanpa ante.
- **ma pi ante kama (Migration seam).** ilo pi lipu sama li tawa e sitelen nimi en lukin
  pi ilo lawa lon lupa `Signer`/`Verifier`; tenpo ni la ilo Ed25519 pi lipu ilo (stdlib)
  li pali e ona, tan ni taso: lon tenpo kama la jan li ken ante e nasin nanpa kepeken
  nimi ma `v1` sin, li wile ala ante e toki lawa. ni li toki pi pali ilo taso; toki awen
  pi linja li nasin poki en nimi ma lon anpa.

**nasin sitelen pi poki pi ilo lawa (nasin ante ala).** poki ale lon toki `Order:` §11 li
kama kepeken nasin open §3 pi lukin sama; §11 li pana ala e kule sin anu kule pi suli
awen. sama la:

- `OperatorID` en `Algo` li **linja sitelen** §3 pi nasin sama — nanpa unsigned LEB128
  varint pi suli UTF-8 (lon nanpa lili), la nanpa lili UTF-8. `Algo` li nimi taso ALA, li
  nimi kule pi suli awen (fixed-width enum tag) ALA, kin la `v0` li ken e ijo wan taso
  `ed25519`; ona MUST jo e nanpa suli lon open, sama linja sitelen ante ale (sama la
  `ed25519` li kama `07 65 64 32 35 35 31 39`). `Nonce`, `NewPublicKey`, en
  `Challenge` li **bytes** §3 pi nasin sama (nanpa suli LEB128 lon open, la nanpa lili).
- `Idx0`, `Idx1`, en `KeyIndex` li kama **`uint64` big-endian pi nanpa lili 8
  awen** kepeken nasin `uint64` §3, sama `Seq` — nanpa varint ALA, nanpa lili wan ALA — kin
  la nanpa ona li lon insa lili `0..6`. jan pali li sitelen e nanpa kepeken nanpa varint anu
  nanpa lili wan la nanpa lili li ike, sitelen nimi li ike. (sama la: nanpa `2` li kama
  `00 00 00 00 00 00 00 02`.)
- sitelen nimi tu `Sig0` en `Sig1` li **weka** tan nanpa lili pi nasin wan pi toki ale
  §11, sama `Signature` tawa toki §4 (§3). ona tu taso li poki weka, en sama `Signature`
  la ona li lon ala toki `Order:` lon anpa. poki nanpa tu wan pi sitelen nimi li lon ala.
  `Idx0`/`Idx1` (ona li toki e nimi pi ilo open pi sitelen nimi) LI lon insa pi nanpa
  lili pi sitelen nimi; *nanpa lili* pi sitelen nimi li lon ala.

### 11.0 wile pi lukin (toki ale pi ilo lawa)

toki pi ilo lawa li kama kute ala la ilo lukin MUST pali e ni:

1. o weka la nanpa pi kepeken wan (nonce) li lon ala anu li lili tan **nanpa lili 16** (lon
   toki pi nanpa pi kepeken wan). nanpa pi kepeken wan ala li MUST NEVER kama sama "o weka
   e lukin pi pana sin (replay cache)."
2. o weka la `Algo` li nasin nanpa pona ala tawa nanpa nasin ni.
3. o weka la nanpa tu pi sitelen nimi `Idx0` en `Idx1` li sama.
4. o weka la nanpa wan li lon ala insa `0..6`.
5. o weka la ilo open awen pi tenpo ni li lon ala lon nanpa wan, anu `algo` pi ilo open
   awen wan li sama ala `Algo` pi toki.
5a. o weka la ilo open pi jan ale lon `Idx0` en `Idx1` li sama lon nanpa lili ale, kin la
   nanpa li ante. ilo open luka tu pi jan ale pi ilo lawa wan MUST ante tan ante ale (tu
   tu); kama awen anu ante MUST NOT pana e ilo open pi jan ale ni: ona li lon pini lon
   nanpa ante (o lukin e §11.2). ni li kama wawa tan ni: nasin 2-of-7 pi awen tan kama
   ante li ken ala kama lili tawa 1-of-1 lon len: nanpa tu li jo e ilo open sama la ilo
   open len wan li ken pali e sitelen nimi tu. lukin ni li lon *ilo open*, li lon nanpa
   taso ala (nasin 3), li wile kin lon poka ona.
6. o weka la suli pi sitelen nimi wan li suli wile ala tawa `Algo`.
7. o pali sin e nanpa lili pi nasin wan, o weka la sitelen nimi TU li lon ala lukin pona
   lon nanpa ona lon nanpa lili ni taso.

awen tan pana sin (anti-replay) (kulupu pi `Seq` en nanpa pi kepeken wan; ona li awen, li
pini lon pakala (fail-closed), li tawa (sliding-window), li lon `(operator, coordinator)`
wan) li wile pi poka pi ilo insa, li lon SEWI pi ijo lon sewi; ona li lon ala nanpa lili
pi nasin wan, li lon ala ma pi lipu ni pi sitelen nanpa.

### 11.1 OperatorTransmission — ilo lawa li sitelen e nimi (2 tan 7)
nimi ma: `sohocloud/operator/v0`

Order: `OperatorID` (string), `Ts` (int64 UTC Unix nanoseconds — `int64` taso; ona li
kama ALA tan `time.Time`; tan ni la nanpa pi lukin suli li sama nanpa pi sitelen nimi),
`Nonce` (bytes, `len >= 16`), `Seq` (uint64), `Algo` (string), `Idx0` (uint64),
`Idx1` (uint64).

sitelen nimi tu `Sig0` (tan ilo open lon `Idx0`) en `Sig1` (tan ilo open lon `Idx1`) li
lon nanpa lili ni pi nasin wan, li WEKA tan ona, sama `Signature` lon ma ante. `Ts`,
`Nonce`, `Seq`, nanpa tu, en `Algo` li lon insa pi nanpa lili pi sitelen nimi.

### 11.2 OperatorRotation — ilo lawa li sitelen e nimi (2 tan 7)
nimi ma: `sohocloud/operator-rotate/v0`

Order: `OperatorID` (string), `KeyIndex` (uint64 — nanpa ni la ilo open ona li kama
ante), `NewPublicKey` (bytes — ilo open sin pi jan ale; ilo lawa li pali e ona), `Algo`
(string), `Ts` (int64 UTC Unix nanoseconds), `Nonce` (bytes, `len >= 16`), `Seq`
(uint64), `Idx0` (uint64), `Idx1` (uint64).

ni li pana e ken tawa ante e ilo open pi jan ale sin. ilo lawa (kulupu linja ALA lon
tenpo ale) li pali e ilo open sin (ilo open len en ilo open pi jan ale). ilo open pi jan ale sin li lon INSA pi nanpa lili
pi sitelen nimi; tan ni la jan ike lon insa (man-in-the-middle) pi kama awen pi nasin
ante, en linja lawa pakala, li ken ala pana e ijo pi ilo open: ilo insa li awen ala e ilo
open sin lon `KeyIndex` la, ona MUST lukin e toki ni (kepeken ilo open tu pi TENPO NI,
`Idx0`/`Idx1`). ilo lukin MUST kin weka e `NewPublicKey` la suli ona li ike tawa `Algo`.
ilo lukin MUST kin weka e `NewPublicKey` la ona li sama lon nanpa lili ale e ilo open
awen lon nanpa ANTE (ona li pakala e wile pi ilo open ante tu tu lon §11.0 nasin 5a, li
open e anpa 2-of-7 → 1-of-1); awen sin pi ilo open sama lon `KeyIndex` ONA (pona sin pi
ante ala) li ken. tawa awen e nasin awen pi nasin nanpa mute ala, ilo lukin SHOULD kin
weka e ante la `Algo` ona li ante tan `Algo` pi ilo open awen ante.

### 11.3 ConformanceResponse — ilo lawa li sitelen e nimi (2 tan 7)
nimi ma: `sohocloud/operator-conformance/v0`

Order: `OperatorID` (string), `Challenge` (bytes — toki lukin; ilo lukin li pana e
ona), `Algo` (string), `Idx0` (uint64), `Idx1` (uint64).

ni li toki pana pi ilo lawa (sitelen nimi li lon) tawa toki lukin pi nasin sama. nimi ma
ante ona li tu e ona tan §11.1: nimi ma li kipisi pi nanpa lili pi nasin wan; tan ni la
sitelen nimi tawa toki lukin pi nasin sama li ken ALA lukin pona sama
`OperatorTransmission` lon tenpo ale, en nasin ante sama. ona li jo ala e nanpa pi kepeken
wan/`Seq` ona — sin li kama tan `Challenge` pi ilo lukin — tan ni la nasin pi nanpa pi
kepeken wan §11.0 li lon ala ona; lukin ante ale §11.0 li lon ona.
