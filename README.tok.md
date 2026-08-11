*toki pona:* lipu ni li lipu lili pi kulupu jan. ona li tan lawa NTARI "P2-002". lipu suli ale li lipu "README.md" pi toki Inli (tenpo suno 2026-07-29). ilo sona li pali e lipu ni; kulupu o lukin o pona e ona kepeken "P2-002" § 3.1.

*English:* This is a condensed community rendering under NTARI policy P2-002. The complete document is the English original README.md (snapshot 2026-07-29). Machine-assisted draft pending community review per P2-002 section 3.1.

sina lukin e pakala lon toki ni la o pona e ona: o pali e "fork" lon
https://github.com/NTARI-RAND/sohocloud-protocol, o pana e "pull request".
pana sina li pona tawa mi mute.

# sohocloud-protocol

## ni li seme

ona li kulupu lipu pi toki ilo "Go". ona li pana e nasin toki wan tawa kulupu ilo "SoHoLINK" en "Cloudy": ilo li sona e ilo ante, li toki e ken pali ona, li kama jo e pali.

## ona li lawa e nasin toki taso

mani awen ala. sona pi pona jan ala. utala pi jan tu ala. ijo mani ni pi kulupu "JFA" li lon ilo "frontend".

ona li kepeken taso ijo pi toki ilo "Go". ilo "SoHoLINK" en ilo "Cloudy" li kepeken ona; ona li kepeken ala ona. tan ni la ilo lawa wan li ken ala kama nena wan pi kulupu ale ("open problem" #7).

## jan en ilo

ilo "frontend" li toki tan ilo pi jan kulupu ona (`SubmitListing`, `Heartbeat`, `PollJobs`, `Decline`, `ReportJob`, `Fees`). ilo lawa li pali e nasin `Coordinator`, li lawa e ilo "frontend" mute. jan li lon ala nasin toki ni: lipu ni li sona e nimi pi ilo pali taso (`NodeID` kepeken nasin "SPIFFE", lon `identity/`).

## tenpo ni la

- pali pini: nimi toki, sitelen pi nasin wan, sitelen ken "ed25519", nasin `Coordinator`, en nasin pana "HTTP+JSON" lon `transport/httpjson` (nasin wan taso — sina ken kepeken nasin ante).
- `testdata/vectors.json` li lipu lawa pi lukin pona: ilo sin li pona la ona li pana e sitelen sama ale (`SPEC.md` §10).
- `anchor/` li nimi taso, li pali ala ("open problem" #6).

## sina ken kama pali

lipu `SPEC.md` li ijo pi jan ale: o lukin e ona, o pali sin e ona, o toki utala e ona. jan ale li ken ala lukin e ijo pi jan ale la, ken ni li pona ala tawa jan. lawa lipu: AGPL-3.0-or-later.

*Network Theory Applied Research Institute, Inc. — 501(c)(3) — EIN 92-3047136 — info@ntari.org*
