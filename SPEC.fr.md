> Traduction communautaire (version préliminaire) — Politique P2-002 de NTARI, Diffusion mondiale multilingue. Source : SPEC.md (original en anglais, instantané du 2026-09-16). Version préliminaire communautaire assistée par machine, en attente de révision par le mainteneur régional conformément à P2-002 §3.1. Les spécifications techniques centrales demeurent en anglais conformément au §2.2.
>
> **L'original anglais fait foi.** Cette traduction est fournie pour la lecture,
> non pour l'interprétation : en cas de divergence avec `SPEC.md`, l'anglais
> prévaut. Sont conservés en anglais, parce qu'ils constituent le contrat de
> conformité, les mots-clés normatifs (MUST, MUST NOT, SHOULD, MAY), les noms de
> champs, les étiquettes de domaine, les identifiants de types et tous les blocs
> de code.
>
> Vous avez repéré une erreur de traduction ? N'hésitez pas à la corriger
> vous-même : forkez le dépôt https://github.com/NTARI-RAND/sohocloud-protocol
> et ouvrez une pull request. Les corrections de traduction sont des
> contributions précieuses, tout autant que le code.

# sohocloud-protocol — Spécification de conformité

Version : **v0 (INSTABLE)** — la forme des messages et le format canonique des
octets peuvent changer sans garantie de compatibilité jusqu'à la v1.

Ce document est normatif et constitue un livrable de construction, non un
accessoire. Le droit AGPL de bifurquer le réseau est vide si une implémentation
ne peut pas être construite à partir d'une spécification lisible sans lire le
code source de ce module. Tout ce qui est marqué ci-dessous MUST / MUST NOT est
une exigence de conformité ; un changement qui en viole une n'est pas une version
réduite de ce protocole, c'est un autre protocole.

---

## 1. Portée

Ce protocole régit la **couche de coordination du substrat** : comment un nœud
est reconnu, comment il annonce le travail qu'il acceptera, et comment du travail
lui est proposé puis rendu compte. Trois parties le parlent :

- **Node** — un appareil détenu par un participant, offrant du calcul et/ou de
  l'impression physique. Un nœud possède son identité et signe tout ce qu'il
  émet. Souverain : il choisit quels coordinateurs servir et peut quitter
  n'importe lequel d'entre eux.
- **Coordinator** — apparie les offres au travail et enregistre les résultats. Un
  rôle enfichable. SoHoLINK est le coordinateur de référence ; un frontend MAY
  faire tourner le sien ou apparier directement avec des nœuds.
- **Frontend** — là où les membres effectuent leurs transactions. Il consomme la
  coordination par la même interface.

**Hors périmètre, délibérément.** Ce protocole n'est PAS l'économie des membres
de JFA. Il ne définit ni séquestre, ni crédit émis par les membres, ni pacte de
réputation, ni registre scellé par dialogue, ni règlement des litiges. Cela
relève de chaque frontend. Le règlement en monnaie fiduciaire qu'un coordinateur
opère avec ses nœuds (p. ex. les versements Stripe de SoHoLINK) relève de
l'économie du *substrat* et sort également de cette spécification.

---

## 2. Identité et liaison

Un nœud est nommé par un `NodeID` — une chaîne stable que le nœud choisit
lui-même.

L'identité de charge de travail d'un nœud est un SPIFFE SVID émis exactement
sous :

```
/node/<NodeID>
```

Un coordinateur autorise une requête comme provenant d'un nœud en comparant le
chemin SPIFFE authentifié de l'appelant à `/node/<NodeID>` pour le `NodeID` nommé
dans la requête.

- Un coordinateur MUST rejeter avec **401** lorsqu'aucune identité SPIFFE n'est
  présente.
- Un coordinateur MUST rejeter avec **403** lorsque l'identité est présente mais
  que son chemin n'est pas exactement `/node/<NodeID>`.
- Un coordinateur MUST NOT accepter « n'importe quel SVID valide » comme
  autorisation pour une requête nommant un autre nœud. C'est un vecteur
  d'usurpation.
- La liaison MUST être une construction de chaîne déterministe, jamais une
  recherche par nœud dans un stockage mutable. Un contrôle conditionné à un champ
  qui n'est jamais renseigné est inerte et MUST NOT servir de frontière de
  sécurité.

Le protocole ne distribue pas de clés publiques. Un coordinateur résout la clé de
vérification d'un nœud hors bande (enregistrement, matériel du SVID, ou un
registre).

---

## 3. Encodage canonique des octets (pour les signatures)

Tout message signé est signé sur ses **octets canoniques**, calculés avec le champ
`Signature` EXCLU. Les signatures sont en **ed25519**.

La signature n'est PAS effectuée sur du JSON ni sur un quelconque encodage de
transport. Elle est effectuée sur le format déterministe préfixé par longueur
ci-dessous, afin que les signatures soient stables entre plateformes, versions de
Go et langages.

Primitives, dans l'ordre exact où les champs sont listés pour chaque message au
§4 :

- **unsigned LEB128 varint** — la primitive de longueur/comptage utilisée par
  `string`, `bytes` et `repeated`. Elle MUST être encodée comme suit,
  indépendamment du langage : prendre la valeur non signée ; émettre 7 bits par
  octet, groupe de poids faible en premier (groupes petit-boutiens) ; mettre le
  bit de poids fort (`0x80`) de chaque octet SAUF le dernier pour marquer la
  continuation ; le dernier octet a son bit de poids fort à zéro. La valeur `0`
  s'encode comme l'unique octet `0x00`. C'est exactement l'encodage produit par
  `encoding/binary.PutUvarint` de Go ; le symbole Go est un recoupement, non la
  définition. (Exemple : `300` = `0b100101100` → 7 bits de poids faible `0101100`
  avec continuation → `0xAC`, bits de poids fort `0000010` en dernier → `0x02`,
  d'où `AC 02`.)
- **domain tag** — les octets canoniques de tout message commencent par une chaîne
  d'étiquette de domaine distincte (voir §4). L'étiquette de domaine MUST être
  encodée selon la règle **string** ci-dessous — c'est-à-dire un unsigned LEB128
  varint de sa longueur en octets UTF-8 suivi de ses octets UTF-8 — exactement
  comme tout autre champ chaîne, sans terminateur nul et sans traitement
  particulier. Des étiquettes distinctes rendent une signature portant sur un type
  de message non transférable à un autre.
- **string** — unsigned LEB128 varint donnant la longueur en octets, puis les
  octets UTF-8 bruts. Une chaîne vide MUST s'encoder comme l'unique octet `0x00`
  (longueur zéro et aucune charge utile).
- **bytes** — unsigned LEB128 varint donnant la longueur en octets, puis les
  octets bruts. (Le champ `Signature` n'est jamais encodé ; il est exclu.)
- **bool** — un seul octet, exactement `0x00` (faux) ou `0x01` (vrai) ; aucune
  autre valeur d'octet n'est valide.
- **int64 / uint64** — 8 octets fixes, gros-boutiens (octet de poids fort en
  premier). `int64` MUST être encodé comme sa réinterprétation `uint64` en
  complément à deux : ajouter 2⁶⁴ à une valeur négative, puis émettre les 8 octets
  gros-boutiens. Ainsi `-1` → `ff ff ff ff ff ff ff ff`, `-2` →
  `ff ff ff ff ff ff ff fe`, et l'`int64` minimal (`-9223372036854775808`) →
  `80 00 00 00 00 00 00 00`. Pas de signe-magnitude, pas de zigzag, pas de
  longueur variable.
- **time** — l'instant sous forme d'`int64` de **nanosecondes** Unix UTC (selon la
  règle int64 ci-dessus, gros-boutien, 8 octets). Le fuseau et les composantes
  d'horloge monotone sont abandonnés. La précision à la nanoseconde est préservée
  exactement ; la valeur MUST NOT être tronquée à la micro- ou à la milliseconde.
  Cela survit à un aller-retour par transport JSON et est indépendant du fuseau
  horaire. Un instant hors de la plage représentable en nanosecondes Unix UTC sur
  `int64` (approximativement les années 1678–2262) sort du périmètre de la v0 et
  MUST être traité comme une erreur d'encodage plutôt qu'enroulé ou borné.
  L'encodeur de référence l'impose : `canon.Buffer.Time` n'ajoute rien et fixe
  `canon.ErrTimeOutOfRange` (récupérable via `Buffer.Err()`) au lieu d'émettre une
  valeur enroulée, de sorte que deux instants distants d'environ 584 ans ne
  peuvent jamais produire des octets identiques. Un producteur MUST rejeter via
  `canon.ValidTime` avant de signer ; un vérificateur/décodeur MUST rejeter un
  message dont l'horodatage échoue à ce contrôle.
- **repeated field** — un unsigned LEB128 varint donnant le *compte* (le nombre
  d'éléments) encodé EN PREMIER, puis les champs de chaque élément encodés dans
  l'ordre, entièrement en ligne, selon les règles primitives normales ci-dessus.
  Un élément répété qui est lui-même une chaîne est donc préfixé par sa longueur
  comme toute autre chaîne ; les champs de structures imbriquées sont émis dans
  l'ordre donné pour cet élément au §4. Il n'y a ni séparateur entre éléments ni
  terminateur après le dernier. Un compte de `0` est l'unique octet `0x00` suivi
  d'aucun élément.

Un vérificateur recalcule les octets canoniques à partir des champs reçus et
contrôle la signature ed25519 face à la clé publique résolue. Un vérificateur MUST
rejeter une signature dont la longueur n'est pas `ed25519.SignatureSize` (64)
avant de vérifier.

Les octets canoniques d'un message sont **exactement** la concaténation de
l'étiquette de domaine suivie de chaque champ listé dans la clause `Order:` de ce
message au §4, dans cet ordre, chacun encodé par sa règle primitive. Il n'y a
aucun encadrement externe : ni octet de version en tête, ni préfixe de longueur
totale enveloppant le message, ni octets de queue après le dernier champ. La liste
`Order:` du §4 est exhaustive et fait autorité — les **champs de signature** du
message sont les SEULS champs exclus, et aucun champ n'existe dans les octets
canoniques qui n'y soit nommé. Pour les douze messages centraux du §4, ce champ
exclu unique est `Signature`. Les messages d'opérateur de la Couche C du §11
portent **deux** signatures (`Sig0` et `Sig1`) sur les mêmes octets canoniques ;
toutes deux sont exclues exactement comme `Signature` ici et, comme `Signature`,
n'apparaissent dans aucune clause `Order:`. Un vérificateur n'encode donc jamais
de champ de signature dans les octets qu'il recalcule, que le message porte une
signature ou deux.

---

## 4. Messages

Pour chaque message : son étiquette de domaine, ses champs en ordre canonique, et
qui le signe. Tous les horodatages suivent la règle `time` du §3.

La clause `Order:` ci-dessous est la SEULE autorité sur l'ordre canonique des
champs, y compris l'ordre des champs de structures imbriquées (p. ex.
`Capacity.VCPUs` avant `Capacity.MemMB`, et chaque imprimante sous la forme `Kind`
puis `Model`). L'implémenteur MUST encoder les champs dans la séquence `Order:` et
MUST NOT déduire l'ordre d'une quelconque autre source. En particulier, l'exemple
`testdata/vectors.json` (§10) liste les `fields` de chaque message par **ordre
alphabétique des clés, pour la lisibilité, ce qui N'EST PAS l'ordre canonique
d'encodage** — suivre l'ordre des clés JSON produirait des octets erronés et des
signatures invalides.

### 4.1 CapabilityListing — signé par le NODE
Domain tag : `sohocloud/listing/v0`

Order: `NodeID` (string), `Class` (string), `Printers` (repeated: `Kind` string,
`Model` string), `GPUs` (repeated: `API` string, `Model` string, `VRAMMB`
int64), `Capacity.VCPUs` (int64), `Capacity.MemMB` (int64),
`Capacity.DiskMB` (int64), `Capacity.StorageCommitMB` (int64),
`Capacity.PrintQPS` (int64), `OptIn.Compute` (bool), `OptIn.Print` (bool),
`OptIn.Storage` (bool), `IssuedAt` (time), `Seq` (uint64).

`Class` ∈ {`micro`, `standard`, `server`}. `Kind` ∈ {`traditional`, `threed`}.
`API` ∈ {`vulkan`, `nnapi`, `cuda`, `metal`}.

`Capacity.DiskMB` est l'espace de travail disponible pour une tâche en cours ;
`Capacity.StorageCommitMB` est le stockage de longue durée que le nœud s'engage à
conserver pour le réseau (hébergement de fragments). Les deux sont délibérément
distincts afin qu'un nœud puisse offrir l'un sans l'autre. La façon dont les
données stockées sont chiffrées, fragmentées et auditées relève du frontend ou de
l'agent, hors de ce protocole — coordination seulement ; le fil ne transporte
jamais de contenu stocké.

Annoncer un GPU vaut adhésion au travail GPU : une offre sans `GPUs` n'en reçoit
aucun, et un nœud retire un GPU en l'omettant de son offre suivante.

`Seq` est strictement monotone par nœud. Un coordinateur MUST rejeter une offre
dont le `Seq` ne dépasse pas strictement le dernier vu pour ce nœud (protection
contre le rejeu et le retour en arrière).

`OptIn` est **indicatif** (voir §5.1).

### 4.2 Heartbeat — signé par le NODE
Domain tag : `sohocloud/heartbeat/v0`

Order: `NodeID` (string), `SentAt` (time), `Seq` (uint64).

Même discipline de `Seq` strictement monotone qu'au §4.1.

### 4.3 Assignment — signé par le COORDINATOR
Domain tag : `sohocloud/assignment/v0`

Order: `JobID` (string), `NodeID` (string), `Spec.Workload` (string),
`Spec.Image` (string), `Spec.Args` (repeated string), `Spec.PrinterKind`
(string), `Spec.GPUAPI` (string), `Spec.GPUMinVRAMMB` (int64),
`Fee.ContributorShareBps` (int64), `Fee.PlatformFeeBps` (int64),
`OfferedAt` (time).

`Spec.GPUAPI` ∈ {``, `vulkan`, `nnapi`, `cuda`, `metal`} — un indice de routage
indicatif comme `PrinterKind` ; vide signifie qu'aucun GPU n'est requis et
`GPUMinVRAMMB` MUST alors valoir `0`.

Les conditions tarifaires sont jointes en ligne afin qu'un nœud voie la
répartition avant de s'engager. Dans le modèle « pull », il n'y a pas de message
d'acceptation distinct.

### 4.4 Decline — signé par le NODE
Domain tag : `sohocloud/decline/v0`

Order: `JobID` (string), `NodeID` (string), `Reason` (string), `DeclinedAt`
(time).

`Reason` ∈ {`local_policy`, `capacity`, `unavailable`}. `local_policy` est la voie
de retrait (voir §5.1).

### 4.5 JobReport — signé par le NODE
Domain tag : `sohocloud/jobreport/v0`

Order: `JobID` (string), `NodeID` (string), `ExitCode` (int64), `FailureCause`
(string), `TmpfsExhausted` (bool), `StartedAt` (time), `FinishedAt` (time).

Ce compte rendu signé est le fait dont dérivent la mesure et le paiement. Ses
champs reflètent l'analyse de fin d'exécution déjà en service dans SoHoLINK.

### 4.6 FeeDeclaration — signé par le COORDINATOR
Domain tag : `sohocloud/fee/v0`

Order: `CoordinatorID` (string), `Terms.ContributorShareBps` (int64),
`Terms.PlatformFeeBps` (int64), `EffectiveAt` (time), `Seq` (uint64).

`ContributorShareBps + PlatformFeeBps` SHOULD être égal à `10000`.

### 4.7 StorageLease — signé par le COORDINATOR
Domain tag : `sohocloud/lease/v0`

Order: `LeaseID` (string), `NodeID` (string), `ShardRef` (bytes),
`SizeClass` (int64), `Fee.ContributorShareBps` (int64),
`Fee.PlatformFeeBps` (int64), `IssuedAt` (time), `ExpiresAt` (time),
`Seq` (uint64).

Le stockage est un BAIL, non une tâche : une obligation continue de conserver un
fragment scellé pour une durée bornée. `ShardRef` est une adresse de contenu opaque
de 32 octets et `SizeClass` une taille de charge utile quantifiée — le protocole
MUST NOT transporter les tailles réelles des objets ni le contenu stocké ; le
chiffrement, le remplissage et la fragmentation sont des affaires de frontend
au-dessus de cette taille. Le renouvellement est un nouveau `StorageLease` pour le
même `LeaseID` avec un `Seq` strictement supérieur (même règle de retour en
arrière qu'au §4.1). Les conditions tarifaires voyagent en ligne, comme au §4.3.

### 4.8 LeaseDecline — signé par le NODE
Domain tag : `sohocloud/lease-decline/v0`

Order: `LeaseID` (string), `NodeID` (string), `Reason` (string),
`DeclinedAt` (time).

`Reason` ∈ {`local_policy`, `capacity`, `unavailable`} ; `local_policy` est la
voie de retrait (§5.1) pour le stockage exactement comme le §4.4 l'est pour les
tâches.

### 4.9 LeaseRelease — signé par le NODE
Domain tag : `sohocloud/lease-release/v0`

Order: `LeaseID` (string), `NodeID` (string), `ReleasedAt` (time).

Un nœud peut toujours cesser de conserver (la souveraineté inclut le départ). La
libération signée met fin à sa mesure ; replacer le fragment est l'affaire du
frontend.

### 4.10 ProofChallenge / ProofResponse — réponse signée par le NODE
Domain tag (de la réponse) : `sohocloud/proof/v0`

Un `ProofChallenge{LeaseID, Offset, Length, Nonce (16 bytes), IssuedAt}` n'est PAS
signé : les nœuds récupèrent les défis en interrogeant le canal authentifié
(modèle « pull »), et un défi n'engage personne à rien. L'artefact signé est la
réponse.

Response Order: `LeaseID` (string), `NodeID` (string), `Offset` (int64),
`Length` (int64), `Nonce` (bytes), `Digest` (bytes), `RespondedAt` (time).

`Digest` MUST être calculé exactement comme
`SHA-256(Nonce || uint64be(Offset) || uint64be(Length) || sealed[Offset : Offset+Length])`
sur les octets du fragment scellé. La réponse rappelle la plage interrogée et le
nonce afin de tenir seule comme fait de mesure non répudiable — le pendant, côté
stockage, du §4.5. Un vérificateur MUST rejeter une réponse dont le couple
`(LeaseID, Nonce)` a déjà été accepté : les nonces sont à usage unique, et c'est
ce qui met en échec le rejeu d'une réponse enregistrée.

### 4.11 KeyRotation — signé par la clé ACTUELLE du nœud
Domain tag : `sohocloud/node-rotate/v0`

Order: `NodeID` (string), `NewPublicKey` (bytes), `Algo` (string),
`RotatedAt` (time), `Seq` (uint64).

La possession de la clé sortante autorise à nommer sa remplaçante. Un vérificateur
MUST valider la signature face à la clé qu'il détient actuellement pour le nœud
(une rotation signée par la remplaçante est une auto-installation et MUST échouer),
MUST imposer un `Seq` strictement monotone par nœud, et MUST rejeter une rotation
dont la `NewPublicKey` n'est pas une clé bien formée pour `Algo` (pour `ed25519`,
pas exactement 32 octets) ou dont l'`Algo` est inconnu — sinon un vérificateur qui
se mettrait à contrôler les messages suivants face à une `NewPublicKey` malformée
tomberait en faute. Ce n'est qu'après la réussite de tous ces contrôles que le
vérificateur commence à vérifier les messages ultérieurs du nœud face à
`NewPublicKey`. `Algo` ∈ {`ed25519`} en v0.

### 4.12 KeyRevocation — signé par la clé RÉVOQUÉE
Domain tag : `sohocloud/node-revoke/v0`

Order: `NodeID` (string), `RevokedPublicKey` (bytes), `RevokedAt` (time),
`Seq` (uint64).

Tue une clé sans remplaçante. Un vérificateur MUST vérifier la signature face à la
clé à laquelle il **fait actuellement confiance** pour `NodeID`, et MUST exiger que
`RevokedPublicKey` soit égale à cette clé de confiance. Une révocation n'est jamais
honorée que face à la clé exacte qu'elle nomme et qu'elle tue. C'est la liaison
anti-abus : un inconnu peut auto-signer une révocation portant le `NodeID` d'une
victime et sa propre clé dans `RevokedPublicKey`, mais comme cette clé n'est pas
celle que le coordinateur tient pour digne de confiance pour la victime, le
contrôle d'égalité échoue et la révocation est rejetée — un nœud ne peut pas être
mis hors ligne par quiconque ne détient pas déjà sa clé actuelle. (Des versions
antérieures disaient « honorer inconditionnellement » ; cette formulation est
caduque — la liaison à la clé de confiance est obligatoire.) Étant donné la clé
actuelle, la lecture sûre d'une révocation est « cesser de faire confiance à cette
clé ». La réinscription après révocation se fait hors bande, par la même voie que
la première inscription, jamais par un message sur le fil qu'un voleur de clé
pourrait forger.

---

## 5. Invariants

### 5.1 Le retrait s'applique localement, jamais sur le fil
`CapabilityListing.OptIn` et les indices de routage de `JobSpec` sont
**indicatifs**. Le coordinateur n'est PAS une frontière de sécurité pour le
retrait. Un nœud MUST appliquer le retrait face à sa propre liste d'autorisations
de confiance locale et MUST refuser le travail que sa politique locale interdit —
via `Decline{Reason: local_policy}` — quoi qu'il ait annoncé auparavant et quoi que
le coordinateur propose.

### 5.2 La politique d'appariement ne fait pas partie du protocole
L'interface `Coordinator` n'expose ni classement ni score. La façon dont un
coordinateur choisit quel nœud reçoit quelle tâche est privée à l'implémentation.
Le protocole MUST NOT se voir ajouter un champ dictant l'appariement, car cela
souderait les consommateurs à la politique d'un seul coordinateur et anéantirait
la possibilité de partir que l'interface existe pour protéger.

### 5.3 Les frais sont lisibles et non rétroactifs
Les conditions tarifaires d'un coordinateur sont publiées sous forme de
`FeeDeclaration` signée et horodatée. Les conditions MUST NOT changer
rétroactivement pour du travail déjà proposé ; un changement est une nouvelle
déclaration signée avec un `EffectiveAt` postérieur. Cela maintient les frais
contestables plutôt qu'opaques (problème ouvert #7).

### 5.4 Le coordinateur est un rôle, pas un concentrateur
Rien de ce dont un nœud ou un frontend a besoin ne peut transiter *exclusivement*
par un coordinateur particulier. Un nœud MAY servir plusieurs coordinateurs ; un
frontend MAY faire tourner son propre coordinateur ou apparier directement. Tout
déploiement dans lequel un unique coordinateur est en fait incontournable est une
recentralisation et MUST être étiqueté comme telle, avec une voie vers des
alternatives.

### 5.5 Numéros de séquence monotones
Les valeurs `Seq` d'un nœud (offres, battements) sont strictement monotones par
nœud. Un coordinateur MUST rejeter un `Seq` non croissant.

---

## 6. Opérations du coordinateur (modèle « pull »)

Le modèle est celui de l'**extraction** : le coordinateur n'appelle jamais le
nœud. Opérations :

| Opération       | Sens             | Corps / résultat                     |
|-----------------|------------------|--------------------------------------|
| `SubmitListing` | nœud → coord     | `CapabilityListing`                  |
| `Heartbeat`     | nœud → coord     | `Heartbeat`                          |
| `PollJobs`      | nœud → coord     | `NodeID` → `[]Assignment`            |
| `Decline`       | nœud → coord     | `Decline`                            |
| `ReportJob`     | nœud → coord     | `JobReport`                          |
| `Fees`          | quiconque → coord | → `FeeDeclaration`                  |

Un coordinateur MUST appliquer la liaison du §2 à `PollJobs` (et à toute opération
nommant un nœud) avant d'agir. Comme le coordinateur n'initie jamais de contact
avec un nœud, il n'y a aucun service côté nœud à spécifier.

---

## 7. Transport

HTTP+JSON est le transport de **référence** (`transport/httpjson`), et NON le fil
lui-même. Aucun paquet central ne l'importe. Une implémentation conforme MAY
utiliser n'importe quel transport en implémentant les opérations du coordinateur
face aux octets canoniques du §3. JSON n'est qu'un encodage de transport ; les
signatures ne sont jamais calculées sur du JSON, de sorte qu'un autre transport
interopère sans renégocier les signatures.

Les points d'entrée HTTP+JSON de référence :

```
POST /v0/listing     body: CapabilityListing   -> 204
POST /v0/heartbeat   body: Heartbeat           -> 204
GET  /v0/jobs?node_id=<id>                      -> 200 []Assignment
POST /v0/decline     body: Decline             -> 204
POST /v0/report      body: JobReport           -> 204
GET  /v0/fees                                   -> 200 FeeDeclaration
```

La liaison SPIFFE (§2) est appliquée par un intergiciel placé devant le
gestionnaire ; le gestionnaire de référence ne fait que câbler les routes et le
JSON et n'est pas l'authentificateur.

---

## 8. Problèmes ouverts que ce protocole touche

L'honnêteté sur les lacunes non résolues est la méthode, non une gêne.

- **#7 Le calcul souverain achète une sortie mécanique, non économique.** C'est la
  couche dont traite le #7. Le §5.4 (coordinateur comme rôle) et le §5.3 (frais
  lisibles) s'opposent à la recentralisation du coordinateur ; le démarrage à
  froid demeure un résidu, et la portabilité de réputation qui l'atténuerait
  relève du frontend et de la gouvernance, non de ce protocole.
- **#6 L'honnêteté du calcul est hors périmètre.** La justesse de l'appariement ou
  du prix d'un coordinateur n'est pas prouvée ici, seulement rendue lisible (§5.2,
  §5.3). `anchor/` réserve la place pour ancrer les *revendications* d'emploi (les
  faits de `JobReport`) dans un journal attesté afin qu'elles soient au moins
  auditables. C'est une ébauche étiquetée, non construite.
- **#4 Adhésion / intégration.** Parce qu'un nœud se nomme et signe pour lui-même
  (§2), un appareil peu doté (p. ex. un téléphone) peut se déclarer nœud. Que
  l'intégration des nœuds soit ouverte ou filtrée est une décision de gouvernance
  et de valeurs laissée ouverte.

## 9. Versionnage

`v0` est instable. Les étiquettes de domaine incorporent `v0` ; un changement
cassant sur les champs canoniques d'un message incrémente son étiquette. Les
consommateurs MUST traiter une étiquette de domaine inconnue comme un message non
vérifiable.

---

## 10. Vecteurs de test

`testdata/vectors.json` est le **jeu de fixtures normatif de conformité** de cette
spécification. Une implémentation indépendante est conforme quant à l'encodage
canonique et à la signature si et seulement si elle reproduit, octet pour octet,
chaque valeur de ce fichier à partir des mêmes entrées. Le jeu de fixtures est
généré par les encodeurs Go de référence et recoupé par une reproduction
indépendante n'utilisant que la bibliothèque standard ; une implémentation
étrangère conforme MUST y correspondre exactement. Là où la prose des §3/§4 et le
jeu de fixtures divergeraient, il s'agit d'un défaut de spécification à corriger,
non d'une licence pour diverger — mais le jeu de fixtures est l'énoncé exécutable
des §3/§4.

Tout l'hexadécimal du jeu de fixtures est en minuscules. Il est auto-descriptif :
une `note` de premier niveau et un bloc `encoding` rappellent les règles
primitives du §3 pour qui ouvrirait le JSON en premier.

### 10.1 Structure

- **`primitives`** — encodages isolés de chaque primitive du §3, afin qu'un
  implémenteur puisse valider l'encodeur de bas en haut avant d'assembler des
  messages. Chaque entrée a un `kind` (`uvarint`, `int64`, `uint64`, `bool`,
  `string`, `time`), un `input` lisible et les `bytes_hex` attendus. Ils couvrent
  les cas limites qui fixent le format : des valeurs varint à chaque seuil d'octet
  de continuation (`0, 1, 127, 128, 255, 256, 300, 16384, 1073741824`) ; `int64` y
  compris `-1`, `-2`, `MaxInt64` et `MinInt64` ; `uint64` y compris `MaxUint64` ;
  les deux booléens ; la chaîne vide et une chaîne UTF-8 multi-octets ; et `time` à
  l'époque Unix et à un instant fixé à la nanoseconde.

- **`messages`** — une entrée par type de message signé (tous les messages signés
  du §4). Chaque entrée consigne : `name`, `domain_tag`, `signer` (`node` ou
  `coordinator`), le `seed_hex` ed25519 de 32 octets et le `public_key_hex`
  dérivé, les `fields` du message, les `canonical_bytes_hex` (la sortie de
  `CanonicalBytes()`, `Signature` exclu) et le `signature_hex` (ed25519 sur
  exactement ces octets canoniques).

- **`operator_messages`** — une entrée par type de message d'opérateur de la Couche
  C (les trois du §11). Les messages d'opérateur étant en 2-de-7, chaque entrée
  diffère d'une entrée `messages` centrale : `signer` vaut toujours `operator` ; un
  tableau `keys` liste les sept clés publiques enregistrées (`index`, `seed_hex`,
  `public_key_hex`, `algo` — les graines ne sont présentes QUE pour que le jeu de
  fixtures soit autonome et reproductible, jamais parce que le réseau détiendrait
  des clés privées) ; `idx0`/`idx1` sont les deux indices signataires ; et au lieu
  d'un unique `signature_hex`, il y a `sig0_hex` et `sig1_hex`, TOUS DEUX pris sur
  les MÊMES `canonical_bytes_hex`. Une implémentation conforme reproduit
  `canonical_bytes_hex` à partir de l'ordre des champs du §11 et vérifie les deux
  signatures à leurs indices.

  > **Piège de l'ordre des champs.** L'objet `fields` de chaque message liste les
  > clés par **ordre alphabétique pour la lisibilité, ET NON en ordre canonique.**
  > L'ordre canonique vient UNIQUEMENT de la clause `Order:` de ce message au §4.
  > Un implémenteur qui encoderait les champs dans l'ordre des clés JSON
  > produirait des octets erronés et des signatures qui ne vérifient pas.

### 10.2 Comment utiliser le jeu de fixtures

1. Reproduisez chaque `primitives[].bytes_hex` à partir de son `input` en suivant
   le §3.
2. Pour chaque message, dérivez la paire de clés depuis `seed_hex` (ed25519 à
   partir d'une graine de 32 octets est déterministe), confirmez qu'elle donne
   `public_key_hex`, encodez les champs selon l'`Order:` du §4 et confirmez que
   vous obtenez `canonical_bytes_hex`, puis vérifiez `signature_hex` face à
   `public_key_hex` sur vos octets dérivés indépendamment. La vérification de la
   signature sur VOS PROPRES octets est le recoupement le plus fort : elle ne peut
   réussir que si vos octets sont identiques à ceux du générateur de référence.

Le test de référence vit dans `vectors/vectors_test.go`. Il régénère chaque
vecteur à partir du code courant et affirme l'égalité octet pour octet avec le
JSON versionné, de sorte qu'une dérive entre le code et le jeu de fixtures fait
échouer la construction ; il relit également le fichier versionné et vérifie
toutes les signatures. Régénérez intentionnellement avec
`go test ./vectors -update` (ou la variable d'environnement
`SOHOCLOUD_UPDATE_VECTORS`). Tous les vecteurs utilisent des graines à remplissage
d'octets fixes et des horodatages codés en dur — il n'y a ni `crypto/rand` ni
`time.Now()`, la sortie est donc pleinement déterministe.

### 10.3 Exemple détaillé : un `Heartbeat`, octet par octet

Voici le vecteur `Heartbeat` (`§4.2`) décodé champ par champ, pour qu'un
réimplémenteur puisse suivre tout l'encodage du début à la fin. Entrées :
`NodeID = "node-alpha"`, `SentAt = ` nanosecondes Unix `1700000050500000000`,
`Seq = 43`.

Octets canoniques (`canonical_bytes_hex`) :

```
16 736f686f636c6f75642f6865617274626561742f7630 0a 6e6f64652d616c706861 17979d09f832d900 000000000000002b
```

Décodés de gauche à droite :

| Octets | Primitive | Signification |
|--------|-----------|---------------|
| `16` | longueur uvarint | Longueur de l'étiquette de domaine = `0x16` = **22** octets. |
| `73 6f 68 6f 63 6c 6f 75 64 2f 68 65 61 72 74 62 65 61 74 2f 76 30` | octets UTF-8 | L'étiquette de domaine `sohocloud/heartbeat/v0` (22 octets). Avec le préfixe de longueur, ces deux lignes constituent l'encodage `string` de l'étiquette de domaine (§3). |
| `0a` | longueur uvarint | Longueur de `NodeID` = `0x0a` = **10** octets. |
| `6e 6f 64 65 2d 61 6c 70 68 61` | octets UTF-8 | `NodeID` = `node-alpha` (10 octets). |
| `17 97 9d 09 f8 32 d9 00` | int64, gros-boutien | `SentAt` en nanosecondes Unix UTC. `0x17979d09f832d900` = `1700000050500000000`. C'est la règle `time` (§3) : l'instant réduit à un `int64` et émis sur 8 octets gros-boutiens. |
| `00 00 00 00 00 00 00 2b` | uint64, gros-boutien | `Seq` = `0x2b` = **43**. 8 octets fixes, pas de varint. |

La concaténation d'exactement ces six groupes — étiquette de domaine (longueur +
octets), `NodeID` (longueur + octets), `SentAt`, `Seq` — constitue la chaîne
canonique d'octets complète. Il n'y a pas d'encadrement avant l'étiquette de
domaine ni d'octets après `Seq`. La signature ed25519 (`signature_hex`) est prise
précisément sur cette chaîne ; le champ `Signature` lui-même n'en fait jamais
partie.

Notez que l'ordre des champs ici (`NodeID`, `SentAt`, `Seq`) suit l'`Order:` du
§4.2, qui se trouve être aussi — par coïncidence pour ce petit message — celui dans
lequel un lecteur les énumérerait ; pour des messages plus grands comme
`CapabilityListing`, l'`Order:` du §4 et les clés alphabétiques de `fields` dans
`vectors.json` diffèrent, et seul l'ordre du §4 est canonique.

---

## 11. Identité d'opérateur (Couche C)

Les §1 à §10 régissent les messages de **node** et de **coordinator**. Cette
section ajoute une troisième identité, orthogonale : l'**operator** — un frontend
(p. ex. Cloudy) qui termine en lui-même l'identité des membres et des nœuds et
présente un unique justificatif rotatif à un coordinateur. Elle est délibérément
distincte de l'identité de charge de travail du nœud (§2) : les messages centraux
de coordination (§4) sont INCHANGÉS et ne sont PAS signés avec un justificatif
d'opérateur.

Un opérateur détient **sept** paires de clés Ed25519 (indices `0..6`) et signe
chaque message avec **deux indices distincts** (une discipline **2-de-7**). Le
réseau n'enregistre que les sept clés **publiques** par opérateur ; **aucune clé
privée d'opérateur n'est jamais détenue par le réseau**, et chaque chemin de
vérification ci-dessous ne prend que des clés publiques.

- **Le 2-de-7 est de l'anti-substitution et de l'hygiène de rotation, PAS de la
  sécurité à seuil.** Si les sept clés privées vivent dans un même processus, une
  seule compromission d'hôte les livre toutes. On ne compte pas sur le 2-de-7 pour
  survivre à la compromission d'une seule clé, à moins que la garde ne soit
  répartie entre domaines de confiance. Ceci est énoncé pour qu'aucun lecteur n'y
  place une confiance excessive.
- **Liaison d'algorithme.** Tout message d'opérateur lie une chaîne d'algorithme
  `Algo` dans ses octets canoniques. `v0` permet exactement `ed25519`. Un
  vérificateur MUST rejeter un message dont l'`Algo` est inconnu, et MUST rejeter
  un message dont l'`Algo` n'est pas égal à l'`algo` enregistré de la clé à chaque
  indice signataire. Le jeu de clés actives d'un opérateur MUST NOT mélanger les
  algorithmes. Le contrôle de longueur de signature se fait par algorithme
  (`ed25519` = 64 octets), et non par une constante en dur, de sorte que la
  rotation atomique planifiée de tout le jeu vers un algorithme post-quantique
  (une NOUVELLE étiquette de domaine, p. ex. `sohocloud/operator/v1-mldsa`)
  réutilise la même règle. Une signature n'est pas transférable d'un algorithme à
  l'autre.
- **Couture de migration.** L'implémentation de référence achemine la signature et
  la vérification d'opérateur au travers d'une interface `Signer`/`Verifier`,
  implémentée aujourd'hui avec l'Ed25519 de la bibliothèque standard, précisément
  pour que l'algorithme puisse être remplacé en ajoutant plus tard une étiquette
  de domaine `v1` sans toucher aux messages centraux. Il s'agit d'une note
  d'implémentation ; le contrat du fil, ce sont les ordres de champs et les
  étiquettes ci-dessous.

**Encodage des champs d'opérateur (aucun traitement particulier).** Chaque champ
des clauses `Order:` du §11 est encodé par sa règle primitive ordinaire du §3 ; le
§11 n'introduit aucun type nouveau ni de largeur fixe. En particulier :

- `OperatorID` et `Algo` sont des **strings** ordinaires au sens du §3 — un
  unsigned LEB128 varint de la longueur en octets UTF-8 suivi des octets UTF-8.
  `Algo` n'est PAS un jeton nu ni une étiquette d'énumération de largeur fixe, bien
  que `v0` ne permette exactement que la seule valeur `ed25519` ; il MUST porter
  son préfixe de longueur comme toute autre chaîne (p. ex. `ed25519` s'encode en
  `07 65 64 32 35 35 31 39`). `Nonce`, `NewPublicKey` et `Challenge` sont des
  **bytes** ordinaires au sens du §3 (préfixe de longueur LEB128 puis octets
  bruts).
- `Idx0`, `Idx1` et `KeyIndex` sont chacun encodés comme un **`uint64`
  gros-boutien sur 8 octets fixes** selon la règle `uint64` du §3, exactement comme
  `Seq` — NON comme un varint et NON comme un octet unique — bien que leurs
  valeurs se situent dans la petite plage `0..6`. Un implémenteur qui encoderait
  un indice en varint ou sur un octet produit des octets erronés et une signature
  invalide. (Exemple : l'indice `2` s'encode en `00 00 00 00 00 00 00 02`.)
- Les deux signatures `Sig0` et `Sig1` sont **exclues** des octets canoniques de
  chaque message du §11, exactement comme `Signature` l'est pour les messages du
  §4 (§3). Ce sont les seuls champs exclus et, comme `Signature`, ils
  n'apparaissent dans aucune clause `Order:` ci-dessous. Il n'y a pas de troisième
  champ de signature. `Idx0`/`Idx1` (qui nomment les clés signataires) SONT à
  l'intérieur des octets signés ; les *octets* de signature, non.

### 11.0 Exigences de vérification (tous les messages d'opérateur)

Un vérificateur MUST, avant d'honorer tout message d'opérateur :

1. Rejeter si le nonce du message est absent ou plus court que **16 octets** (là
   où le message comporte un nonce). Un nonce vide MUST NE JAMAIS être traité
   comme « ignorer le cache de rejeu ».
2. Rejeter si `Algo` n'est pas un algorithme pris en charge pour cette version.
3. Rejeter si les deux indices signataires `Idx0` et `Idx1` sont égaux.
4. Rejeter si l'un des indices est hors de `0..6`.
5. Rejeter s'il n'existe aucune clé enregistrée active à l'un des indices, ou si
   l'`algo` de l'une des clés enregistrées n'est pas égal à l'`Algo` du message.
5a. Rejeter si les clés publiques enregistrées à `Idx0` et `Idx1` sont identiques
   octet pour octet, même si les indices diffèrent. Les sept clés publiques
   enregistrées d'un opérateur MUST être distinctes deux à deux ; un enregistrement
   ou une rotation MUST NOT installer une clé publique qui existe déjà à un autre
   indice (voir §11.2). Cela est imposé pour que la propriété d'anti-substitution
   2-de-7 ne puisse pas se dégrader silencieusement en 1-de-1 : si deux indices
   portaient la même clé, une seule clé privée pourrait produire les deux
   signatures. Ce contrôle porte sur les *clés*, et pas seulement sur les indices
   (règle 3) ; il est exigé en plus de celle-ci.
6. Rejeter si la longueur de l'une des signatures n'est pas celle attendue pour
   `Algo`.
7. Recalculer les octets canoniques et rejeter à moins que LES DEUX signatures ne
   vérifient à leurs indices sur exactement ces octets.

L'anti-rejeu (un ensemble durable et à fermeture sûre combinant fenêtre glissante
de `Seq` et nonces, à portée `(operator, coordinator)`) est une obligation côté
coordinateur superposée À CE QUI PRÉCÈDE ; elle ne fait pas partie des octets
canoniques et sort du périmètre de cette spécification d'encodage.

### 11.1 OperatorTransmission — signé par l'OPERATOR (2 sur 7)
Domain tag : `sohocloud/operator/v0`

Order: `OperatorID` (string), `Ts` (int64 nanosecondes Unix UTC — l'`int64` brut,
NON re-dérivé au travers d'un `time.Time`, afin que la valeur contrôlée en plage
soit exactement la valeur signée), `Nonce` (bytes, `len >= 16`), `Seq` (uint64),
`Algo` (string), `Idx0` (uint64), `Idx1` (uint64).

Les deux signatures `Sig0` (par la clé à `Idx0`) et `Sig1` (par la clé à `Idx1`)
sont prises sur ces octets canoniques et en sont EXCLUES, exactement comme
`Signature` ailleurs. `Ts`, `Nonce`, `Seq`, les deux indices et `Algo` sont tous à
l'intérieur des octets signés.

### 11.2 OperatorRotation — signé par l'OPERATOR (2 sur 7)
Domain tag : `sohocloud/operator-rotate/v0`

Order: `OperatorID` (string), `KeyIndex` (uint64 — l'indice dont la clé est
remplacée), `NewPublicKey` (bytes — la clé publique de remplacement générée par
l'opérateur), `Algo` (string), `Ts` (int64 nanosecondes Unix UTC), `Nonce` (bytes,
`len >= 16`), `Seq` (uint64), `Idx0` (uint64), `Idx1` (uint64).

Ceci autorise l'installation d'une nouvelle clé publique. L'opérateur (jamais le
réseau) génère la nouvelle paire de clés. La nouvelle clé publique est À
L'INTÉRIEUR des octets signés afin que ni un intercepteur de l'enregistrement hors
bande ni un canal d'administration compromis ne puissent injecter du matériel de
clé : un coordinateur MUST vérifier ce message (avec deux clés ACTUELLES,
`Idx0`/`Idx1`) avant d'enregistrer la nouvelle clé à `KeyIndex`. Un vérificateur
MUST en outre rejeter une `NewPublicKey` dont la longueur est incorrecte pour
`Algo`. Un vérificateur MUST également rejeter une `NewPublicKey` identique octet
pour octet à une clé déjà enregistrée à un indice DIFFÉRENT (cela violerait
l'exigence de clés distinctes deux à deux de la règle 5a du §11.0 et préparerait la
dégradation 2-de-7 → 1-de-1) ; réenregistrer la même clé à son PROPRE `KeyIndex`
(un rafraîchissement sans effet) est permis. Pour préserver l'invariant
d'absence de mélange d'algorithmes, un vérificateur SHOULD aussi rejeter une
rotation dont l'`Algo` diffère de celui des autres clés enregistrées.

### 11.3 ConformanceResponse — signé par l'OPERATOR (2 sur 7)
Domain tag : `sohocloud/operator-conformance/v0`

Order: `OperatorID` (string), `Challenge` (bytes — le défi fourni par le
vérificateur), `Algo` (string), `Idx0` (uint64), `Idx1` (uint64).

C'est la réponse signée d'un opérateur à un défi de conformité. Son étiquette de
domaine distincte la sépare par domaine du §11.1 : puisque l'étiquette fait partie
des octets canoniques, une signature produite pour un défi de conformité ne peut
JAMAIS se vérifier comme une `OperatorTransmission`, et réciproquement. Elle ne
porte ni nonce ni `Seq` propres — la fraîcheur vient du `Challenge` du
vérificateur — de sorte que la règle de nonce du §11.0 ne s'y applique pas ; tous
les autres contrôles du §11.0, si.
