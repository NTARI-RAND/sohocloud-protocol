# lipu pi nasin sama — Janus-Facing Architecture (toki pona)

> **toki lili pi lipu ni:** lipu ni li lipu pi kulupu, li tan nasin NTARI "P2-002". lipu suli li lipu Inli "CONFORMANCE.md" (tenpo pi 2026-10-05). ilo sona li pali e lipu ni; kulupu li wile lukin e ona kepeken nasin "P2-002" kipisi "3.1".
>
> **Note:** This is a community rendering under NTARI policy P2-002. The complete document is the English original CONFORMANCE.md (snapshot 2026-10-05). Machine-assisted draft pending community review per P2-002 section 3.1.
>
> sina lukin e pakala lon toki ni la o pona e ona: o pali e "fork" lon
> https://github.com/NTARI-RAND/sohocloud-protocol, o pana e "pull request". pana sina
> li pona tawa mi mute.
>
> **lipu Inli li lawa.** lipu ni li tawa lukin, li tawa ala sona pi nasin lawa: lipu ni li ante tan `CONFORMANCE.md` la lipu Inli li lawa. nimi lawa (MUST, MUST NOT, SHOULD, SHOULD NOT, MAY), nimi pi poki sona, nimi sitelen en nimi ilo li awen lon toki Inli, tan ni: ona li lipu lawa pi nasin sama.

---

lipu ni la kulupu ilo (module) ni li toki e ona sama kepeken nimi pi nasin JFA. ona li toki e ni **lon open**, li toki e ijo pi ilo esun anu pi kepeken lon ma lon monsi, tan nasin pi nanpa toki pi nasin JFA. ona li kepeken nasin pi ken lukin pi nasin JFA: toki ale pi nasin sama li linja tawa ilo ona en tawa lukin ni: lukin ni li wawa e toki ni. linja li lon ala la sitelen li toki e ni: toki ni li ijo pi tenpo lili (stand-in). toki pi linja ala li toki esun taso.

nasin ni li **Janus-Facing Architecture (JFA)** — lipu wan pi nasin ilo pi kulupu NTARI. jan ale li ken kepeken lipu ni, li ken pana e ona; ona li lon poki pi jan ale pi pali ni (nasin "AGPL-3.0"). ona li toki e nimi pi pali en nimi pi ilo, li toki ala e nimi pi ilo esun lon tenpo ale; kulupu ilo ni li toki e pali ona, li pana e linja ona.

## toki pi pali

kulupu ilo ni li **nasin insa pi pali wan** pi **kipisi ilo** JFA: ona li toki pi linja ilo, li lipu lili pi ijo wile ala (dependency leaf). kepeken toki ni la ilo pi jo jan li kama sona tawa ilo ante, li toki e ken pali ona, li kama jo e pali, li toki e pali pini. kulupu ilo ni li pali tawa toki awen ni pi kipisi ilo: **pali wan li lon ilo ni: jan kulupu li ken jo e ona — lupa lili wan pi tomo ilo (hosting chokepoint) pi ken ala weka li lon ala.**

| nimi pi kulupu ilo ni | pali lon nasin JFA |
|---|---|
| Node | ilo pi jo jan — jan kulupu li pana e ilo ni |
| Coordinator | ilo insa: ona li lawa e pali lon poka pi ilo Node; pali ni la ilo ante li ken kama lon ma ona |
| Frontend | lupa lukin: ilo tawa jan kulupu |
| FeeDeclaration | toki mani pi ilo insa: ilo insa li sitelen e ona; jan ale li ken lukin e ona, li ken utala e ona |
| JobReport | toki pi pali (awen lon lipu kepeken jan lukin li awen tawa tenpo kama — o lukin e ijo pi tenpo lili) |

## nasin awen en linja ona

| nasin awen (nasin JFA) | ilo lon ni | lukin |
|---|---|---|
| nasin insa pi pali wan li **lipu lili pi ijo wile ala** (dependency leaf) | `go.mod` li toki e ijo wile ala (0); ona li kepeken lipu ilo pi toki Go taso (standard library) | CI: jan li pana e ijo wile sin la `.github/workflows/dependency-leaf.yml` li pakala e pali ilo (build) |
| **jan li lon ala linja lon tenpo ale** | kulupu ilo ni li sona e nimi pi pali ilo taso (workload identity) — `identity/` NodeID kepeken linja SPIFFE ona pi nasin wan (canonical); toki ala li jo e nimi jan, e nasin toki tawa jan, e poki lili pi selo jan | SPEC (lawa); selo pi kule ijo (type surface); nanpa lawa (golden vectors) |
| **lupa lili wan pi tomo ilo pi ken ala weka li lon ala** | ilo insa li pali ni: ilo ante li ken kama lon ma ona. lupa lukin MAY pali e ilo insa ona, anu ona MAY alasa e ilo Node pona lon nasin pi ilo insa ala (SPEC §1) | toki lawa pi SPEC; ilo tu li kepeken nimi nanpa open (published tag) |
| **jan ale li ken lukin e mani pana (fee)** | mani pana li lon linja lon nasin wan taso: `FeeDeclaration` (ilo insa li sitelen e nimi ona lon ona) — jan li sitelen e ona, jan ale li ken lukin e ona, li ken utala e ona | `fees/`; nanpa lawa (golden vectors) |
| **ken lukin li ijo pi pali ilo, li toki lili ala** | `SPEC.md` li lawa, li ijo pi pali ilo (build deliverable): jan pali ante li ken pali e ilo tan ona, li wile ala lukin e lipu ilo ni | `testdata/vectors.json` (lon lawa) + `testdata/reproduce_spec.py` (ilo ante pi pali sin) + lukin lawa `vectors/` — lipu ilo en nanpa awen li kama ante la pali ilo (build) li pakala |
| **sitelen pi nasin wan, pi ante ala** | `canon/`: ilo sitelen tawa nimi sitelen (signing encoder); nanpa pi suli li lon open pi ijo ale; nimi ma (domain) li tu e kipisi | tenpo ale pi lukin la ilo li pali sin e nanpa lawa (golden vectors), li lukin e sama ona lon nanpa lili ale (byte) |

## ijo li lon ala tan wile

kulupu ilo ni li esun **ala** pi jan kulupu. ijo ni li lon ala la ni li nasin awen, li weka pakala ala: mani awen (escrow) li lon ala, mani tan jan kulupu (credit) li lon ala, toki awen pi sona pona jan li lon ala, lipu awen tan toki pi jan tu li lon ala, lawa pi utala li lon ala. ijo ni li pali pi kipisi sewi (kipisi sona, kipisi pi toki awen, kipisi esun), li lon lupa lukin, li MUST NOT tawa insa pi lipu lili ni. jan wan li jo e nimi wan taso (single-participant-identity) — ni kin li wile pi lupa lukin. kipisi ni li sona ala e jan a.

## ijo pi tenpo lili en ijo pi pini ala

mi toki e nimi ona tan nasin pi toki lon pi nasin JFA: sitelen pi pali lon li len e ma pi pakala ona la ona li pana e sona pi len taso.

- **`anchor/` li ijo open (stub); lipu pi poki ilo ona li toki e ni.** awen pi toki pali lon lipu (pali li kama, ijo pini li tawa, nanpa pi kepeken) — jan lukin li lukin e ona, toki li kama suli taso — li awen tawa tenpo kama; jan li pali ala e ona. ona li lon ala la ilo insa li toki e pali pini, li wawa e toki ona kepeken ona sama taso — ni li nanpa 6 pi pini ala pi nasin JFA, li linja tawa nanpa 2 pi pini ala.
- **v0 li UNSTABLE (awen ala).** selo pi toki linja li ken ante, li pana ala e toki awen pi ken kepeken sama, tawa v1; ilo kepeken li awen e nimi nanpa open (published tags).
- **ilo nanpa pi jo jan li pana e ken weka pi ilo, li pana ala e ken weka pi mani** (nanpa 7 pi pini ala). nasin insa ni li lawa e ni: tomo ilo li ken ala kama lon ma wan mute. taso ona li ken ala pana e mani tawa jan wan tawa weka ona. mi toki e nimi ona; mi pona ala e ona.

## toki pi ijo wile

ona li wile e ni: lipu ilo pi toki Go (standard library). ona li wile ala e ijo ante — kulupu ilo ante ala pi kulupu NTARI, kulupu ilo ala pi kulupu ante, ijo ala lon sewi ona anu lon poka ona lon nasin ilo (stack). lukin CI la jan li pakala e nasin ni la pali ilo (build) li pakala; jan lukin li wile ala lukin e ni.

## toki pi ilo esun taso (lon pini, tan nasin pi nanpa toki)

ilo SoHoLINK li ilo insa pi lipu sama (reference); ilo Cloudy li lupa lukin pi lipu sama. ilo tu li kepeken kulupu ilo ni lon nimi nanpa open (published version tag). nasin insa li pana ala e ken suli tawa ilo wan: ilo ale pi nasin sama li ken kama lon ma pi ilo wan.
