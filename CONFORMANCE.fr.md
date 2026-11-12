> Traduction communautaire (brouillon) — politique NTARI P2-002, Diffusion multilingue mondiale. Source : CONFORMANCE.md (original anglais, instantané du 2026-10-05). Brouillon communautaire assisté par machine, en attente de relecture par le mainteneur régional conformément à P2-002 §3.1. Les spécifications techniques essentielles restent en anglais conformément au §2.2.
>
> Vous avez remarqué une erreur de traduction ? N'hésitez pas à la corriger
> vous-même : forkez le dépôt https://github.com/NTARI-RAND/sohocloud-protocol et
> ouvrez une pull request. Les corrections de traduction sont des
> contributions précieuses, tout autant que le code.

> L'original anglais est normatif. Cette traduction est proposée pour la lecture, non pour l'interprétation : là où elle diverge de `CONFORMANCE.md`, l'anglais prévaut. Les mots-clés normatifs (MUST, MUST NOT, SHOULD, SHOULD NOT, MAY), les noms de champs, les libellés et les identifiants sont conservés en anglais, car ils constituent le contrat de conformité.

# Conformité — Janus-Facing Architecture

Ce document est l'autodescription du module dans les termes propres à l'architecture, énoncée **avant** tout élément propre au produit ou au déploiement, conformément à la règle d'ordonnancement de l'architecture. Il suit la discipline de lisibilité de l'architecture : une affirmation de conformité est rattachée au mécanisme et à la vérification qui la font respecter, ou bien elle est étiquetée comme substitution. Une prose non rattachée relève du marketing.

L'architecture est **Janus-Facing Architecture (JFA)** — le document d'architecture unifié de NTARI, documentation libre placée sous les communs AGPL-3.0 du projet. Elle nomme des rôles et des mécanismes, jamais des produits ; ce module déclare le rôle qu'il remplit et montre ses rattachements.

## Déclaration de rôle

Ce module est le **protocole de coordination** d'un **substrat** JFA : le langage filaire, feuille de dépendances, par lequel des machines détenues par les participants sont reconnues, annoncent leurs capacités, puis prennent en charge du travail et en rendent compte. La garantie du substrat qu'il sert : **une coordination sur une infrastructure que les participants peuvent détenir — aucun point d'étranglement d'hébergement inamovible.**

| Terme de ce module | Rôle dans l'architecture |
|---|---|
| Node | l'infrastructure détenue par les participants — les machines que les membres apportent |
| Coordinator | le coordinateur : l'orchestration côté nœud, un rôle interchangeable |
| Frontend | un frontend : l'application destinée aux membres |
| FeeDeclaration | la déclaration de frais du coordinateur, rédigée par lui, lisible et contestable |
| JobReport | une déclaration de travail (l'ancrage attesté par témoins est réservé — voir les substitutions) |

## Invariants et leurs rattachements

| Invariant (architecture) | Mécanisme ici | Vérification |
|---|---|---|
| Le protocole de coordination est une **feuille de dépendances** | `go.mod` ne déclare aucune dépendance ; bibliothèque standard uniquement | CI : `.github/workflows/dependency-leaf.yml` fait échouer la compilation à tout ajout de dépendance |
| **Les personnes n'apparaissent jamais sur le fil** | La seule identité que ce module connaît est l'identité de charge de travail — le NodeID de `identity/` avec son rattachement SPIFFE canonique ; aucun message ne porte de nom, de coordonnées ni de champ ayant la forme d'une personne | SPEC (normatif) ; la surface des types ; vecteurs de référence |
| **Aucun point d'étranglement d'hébergement inamovible** | Le coordinateur est un rôle interchangeable : un frontend MAY exécuter son propre coordinateur ou s'apparier directement avec des nœuds (SPEC §1) | Texte normatif de la SPEC ; deux consommateurs s'appuient sur l'étiquette publiée |
| **Lisibilité des frais** | Les frais n'existent sur le fil que sous la forme d'une `FeeDeclaration` signée par le coordinateur — rédigée par lui, lisible, contestable | `fees/` ; vecteurs de référence |
| **La lisibilité est un livrable, pas un commentaire** | `SPEC.md` est normatif et constitue un livrable de compilation : un implémenteur extérieur construit à partir de lui sans lire ce code source | `testdata/vectors.json` (vérité de référence) + `testdata/reproduce_spec.py` (reproducteur indépendant) + test de référence `vectors/` — toute dérive entre le code et les vecteurs versionnés fait échouer la compilation |
| **Encodage canonique et déterministe** | `canon/` : encodeur de signature à préfixe de longueur et à séparation de domaine | vecteurs de référence régénérés et comparés octet par octet à chaque exécution des tests |

## Délibérément hors périmètre

Ce module n'est **pas** l'économie des membres, et cette absence est un invariant, non une lacune : aucun séquestre, aucun crédit émis par les membres, aucun pacte de réputation, aucun registre scellé par dialogue, aucun arbitrage des litiges. Ces éléments relèvent des couches supérieures (registre, pacte, économie), résident dans les frontends et MUST NOT migrer dans cette feuille. L'identité unique du participant est de même une obligation du frontend — cette couche ne connaît absolument aucune personne.

## Substitutions et résidus ouverts

Nommés conformément à la règle d'honnêteté de l'architecture : un exemple concret qui dissimulerait ses insuffisances n'enseignerait que la dissimulation.

- **`anchor/` est une ébauche, étiquetée comme telle dans la documentation de son paquet.** L'ancrage en ajout seul, attesté par témoins, des déclarations de travail (tâche en entrée, résultat en sortie, compteur) est réservé, non construit. Tant qu'il n'existe pas, le compte rendu du travail accompli par un coordinateur est auto-attesté — le problème ouvert 6 de l'architecture, lié au problème ouvert 2.
- **v0 est UNSTABLE.** Les formes sur le fil peuvent changer sans garantie de compatibilité jusqu'à v1 ; les consommateurs épinglent des étiquettes publiées.
- **Le calcul souverain offre une sortie mécanique, non économique** (problème ouvert 7). Ce protocole plafonne le degré de concentration que peut atteindre l'hébergement ; il ne peut financer la sortie de personne. Nommé, non résolu.

## Déclaration de dépendances

Dépend de : la bibliothèque standard de Go. Rien d'autre — aucun autre module NTARI, aucun module tiers, rien au-dessus ni à côté de lui dans la pile. La vérification CI fait d'une violation un échec de compilation plutôt qu'une détection en revue.

## Notes propres au produit (en dernier, conformément à la règle d'ordonnancement)

SoHoLINK est le coordinateur de référence ; Cloudy est un frontend de référence. Tous deux consomment ce module par étiquette de version publiée. Aucun n'est privilégié par le protocole : toute implémentation conforme peut remplacer l'un ou l'autre.
