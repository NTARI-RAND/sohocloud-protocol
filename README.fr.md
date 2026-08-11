> Traduction communautaire (version préliminaire) — politique NTARI P2-002, Diffusion multilingue mondiale. Source : README.md (original anglais, instantané du 2026-07-29). Brouillon communautaire assisté par machine, en attente de relecture par le mainteneur régional conformément à la politique P2-002 §3.1. Les spécifications techniques fondamentales restent en anglais conformément au §2.2.
>
> Vous avez remarqué une erreur de traduction ? N'hésitez pas à la corriger
> vous-même : forkez le dépôt https://github.com/NTARI-RAND/sohocloud-protocol
> et ouvrez une pull request. Les corrections de traduction sont des
> contributions précieuses, tout autant que le code.

# sohocloud-protocol

Le module Go partagé qui définit le **protocole de coordination du substrat** du
réseau SoHoLINK / Cloudy : reconnaissance des nœuds, déclaration des capacités
et embauche pour les tâches — un protocole parlé par chaque coordinateur et
chaque frontend.

## Périmètre

Ce module régit **uniquement la coordination**. Il n'est **pas** l'économie des
membres JFA : pas de séquestre, pas de crédit émis par les membres, pas
d'engagement de réputation, pas d'enregistrement scellé par le dialogue, pas
d'arbitrage des litiges. Ces éléments résident dans chaque frontend et sont
délibérément absents ici.

C'est une **feuille de dépendances** : il n'importe aucun autre module NTARI, et
rien d'autre que la bibliothèque standard de Go. SoHoLINK (le coordinateur de
référence) comme Cloudy (un frontend) importent *ce* module ; rien n'est
réimporté à travers lui. Cette propriété — garantie par le graphe
d'importation, et non par la prose — est ce qui empêche un coordinateur
quelconque de devenir un hub par lequel tout le réseau devrait passer
(problème ouvert n° 7).

## Rôles

Les frontends parlent la surface côté nœud — `SubmitListing`, `Heartbeat`,
`PollJobs`, `Decline`, `ReportJob`, `Fees` — pour le compte des machines que
leurs membres mettent à disposition ; les coordinateurs implémentent l'interface
`Coordinator` et coordonnent (et fédèrent) les frontends. Les personnes
n'apparaissent jamais sur le réseau : l'identité des membres relève du
frontend, et la seule identité que connaît ce module est l'identité de charge de
travail (workload) — le NodeID et sa liaison SPIFFE canonique dans `identity/`.

## Arborescence

```
version.go            protocol version (v0, unstable)
canon/                deterministic length-prefixed signing encoder
identity/             NodeID + the canonical SPIFFE binding predicate
listing/              CapabilityListing (node-signed)
liveness/             Heartbeat (node-signed)
employment/           Assignment (coordinator-signed); Decline, JobReport (node-signed)
fees/                 FeeDeclaration (coordinator-signed)
coordinator/          the Coordinator interface — the pluggable role, no algorithm
anchor/               STUB, not built — witnessed employment-claim layer (#6)
transport/httpjson/   reference transport; imported by no core package
vectors/              conformance test: regenerates and checks testdata/vectors.json
testdata/vectors.json cross-language conformance vectors (normative fixture; see SPEC.md §10)
```

## État d'avancement (sans fard)

- **Implémenté :** les types de messages, les octets canoniques de signature, la
  signature et la vérification ed25519, l'interface `Coordinator`, et un
  transport HTTP+JSON de référence sous `transport/httpjson`.
- **Vecteurs de conformité :** `testdata/vectors.json` est la fixture de
  conformité normative multilangages — 25 cas primitifs plus les six types de
  messages signés, avec leurs octets canoniques et leurs signatures ed25519. Une
  implémentation externe est conforme si et seulement si elle reproduit la
  fixture octet par octet à partir des mêmes entrées (`SPEC.md` §10). Le test
  `vectors/` régénère la fixture depuis les encodeurs actuels et fait échouer le
  build à la moindre divergence.
- **Transport de référence, et non le protocole de transmission :**
  `transport/httpjson` est facultatif et n'est importé par aucun paquet du cœur.
  Une implémentation conforme PEUT véhiculer ces messages sur n'importe quel
  transport, en implémentant `coordinator.Coordinator` par rapport aux octets
  canoniques documentés dans `SPEC.md`.
- **Ébauche (stub), non implémenté :** `anchor/` — la couche de revendication
  d'embauche avec témoins (problème ouvert n° 6). Signalé comme provisoire ;
  voir la documentation de son paquet.

## Licence

AGPL-3.0-or-later. `SPEC.md` fait partie des communs : il existe pour être lu,
réimplémenté et contesté. Des communs que personne ne peut lire sont une liberté
que personne ne peut exercer.

*Network Theory Applied Research Institute, Inc. — 501(c)(3) — EIN 92-3047136 — info@ntari.org*
