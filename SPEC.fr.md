> Traduction communautaire (brouillon) — politique NTARI P2-002, Diffusion multilingue mondiale. Source : SPEC.md (original anglais, instantané du 2026-10-05). Brouillon communautaire assisté par machine, en attente de relecture par le mainteneur régional conformément à P2-002 §3.1. Les spécifications techniques essentielles restent en anglais conformément au §2.2.
>
> Vous avez remarqué une erreur de traduction ? N'hésitez pas à la corriger
> vous-même : forkez le dépôt https://github.com/NTARI-RAND/sohocloud-protocol et
> ouvrez une pull request. Les corrections de traduction sont des
> contributions précieuses, tout autant que le code.

> L'original anglais est normatif. Cette traduction est proposée pour la lecture, non pour l'interprétation : là où elle diverge de `SPEC.md`, l'anglais prévaut. Les mots-clés normatifs (MUST, MUST NOT, SHOULD, SHOULD NOT, MAY), les noms de champs, les libellés et les identifiants sont conservés en anglais, car ils constituent le contrat de conformité.

# sohocloud-protocol — Spécification de conformité

Version : **v0 (INSTABLE)** — les formes des messages et le format d'octets canonique
peuvent changer sans garantie de compatibilité jusqu'à la v1.

Ce document est normatif et constitue un livrable de construction, non une réflexion
après coup. Le droit de forker le réseau que confère l'AGPL est vide si une mise en
œuvre ne peut pas être construite à partir d'une spécification lisible sans lire le code
source de ce module. Tout ce qui est marqué ci-dessous MUST / MUST NOT est une exigence
de conformité ; une modification qui en viole une n'est pas une version réduite de ce
protocole, c'est un protocole différent.

---

## 1. Portée

Ce protocole régit la **couche de coordination du substrat** : comment un nœud est
reconnu, comment il annonce le travail qu'il accepte d'effectuer, et comment le travail
lui est proposé puis rapporté. Trois parties le parlent :

- **Nœud** — un appareil détenu par un participant, qui offre du calcul et/ou de
  l'impression physique. Un nœud possède son identité et signe tout ce qu'il émet.
  Souverain : il choisit les coordinateurs qu'il sert et peut quitter n'importe lequel
  d'entre eux.
- **Coordinateur** — apparie les annonces au travail et enregistre les résultats. Un
  rôle interchangeable. SoHoLINK est le coordinateur de référence ; un frontend MAY
  faire tourner le sien ou apparier directement avec les nœuds.
- **Frontend** — l'endroit où les membres effectuent leurs transactions. Il consomme la
  coordination par la même interface.

**Hors de portée, délibérément.** Ce protocole n'est PAS l'économie des membres de JFA.
Il ne définit ni séquestre, ni crédit émis par les membres, ni pacte de réputation, ni
enregistrement scellé par dialogue, ni arbitrage des litiges. Ces éléments relèvent de
chaque frontend. Le règlement en monnaie fiduciaire qu'un coordinateur opère avec ses
nœuds (par ex. les versements Stripe de SoHoLINK) constitue l'économie du *substrat* et
se situe également hors de cette spécification.

---

## 2. Identité et liaison

Un nœud est nommé par un `NodeID` — une chaîne stable que le nœud choisit pour lui-même.

L'identité de charge de travail d'un nœud est un SVID SPIFFE émis exactement sous :

```
/node/<NodeID>
```

Un coordinateur autorise une requête comme provenant d'un nœud en comparant le chemin
SPIFFE authentifié de l'appelant à `/node/<NodeID>` pour le `NodeID` nommé dans la
requête.

- Un coordinateur MUST rejeter avec **401** lorsqu'aucune identité SPIFFE n'est
  présente.
- Un coordinateur MUST rejeter avec **403** lorsque l'identité est présente mais que son
  chemin n'est pas exactement `/node/<NodeID>`.
- Un coordinateur MUST NOT accepter « n'importe quel SVID valide » comme autorisation
  pour une requête nommant un autre nœud. C'est un vecteur d'usurpation d'identité.
- La liaison MUST être une construction de chaîne déterministe, jamais une recherche par
  nœud dans un stockage modifiable. Une vérification conditionnée à un champ qui n'est
  jamais renseigné est inerte et MUST NOT être utilisée comme frontière de sécurité.

Le protocole ne distribue pas de clés publiques. Un coordinateur obtient la clé de
vérification d'un nœud hors bande (enregistrement, matériel du SVID ou registre).

---

## 3. Encodage canonique en octets (pour les signatures)

Chaque message signé est signé sur ses **octets canoniques**, calculés en EXCLUANT le
champ `Signature`. Les signatures sont **ed25519**.

La signature n'est PAS effectuée sur du JSON ni sur aucun encodage de transport. Elle
est effectuée sur le format déterministe, préfixé par la longueur, décrit ci-dessous,
afin que les signatures soient stables d'une plateforme, d'une version de Go et d'un
langage à l'autre.

Primitives, dans l'ordre exact où les champs sont listés pour chaque message au §4 :

- **varint LEB128 non signé** — la primitive de longueur/nombre utilisée par `string`,
  `bytes` et `repeated`. Elle MUST être encodée comme suit, indépendamment du langage :
  prendre la valeur non signée ; émettre 7 bits par octet, le groupe de poids faible en
  premier (groupes petit-boutistes) ; mettre à 1 le bit de poids fort (`0x80`) de chaque
  octet SAUF le dernier pour marquer la continuation ; le dernier octet a son bit de
  poids fort à 0. La valeur `0` s'encode comme l'octet unique `0x00`. C'est exactement
  l'encodage produit par `encoding/binary.PutUvarint` de Go ; le symbole Go est une
  vérification croisée, non la définition. (Exemple détaillé : `300` = `0b100101100` →
  7 bits de poids faible `0101100` avec continuation → `0xAC`, bits de poids fort
  `0000010` en dernier → `0x02`, ce qui donne `AC 02`.)
- **étiquette de domaine** — les octets canoniques de chaque message commencent par une
  chaîne d'étiquette de domaine distincte (voir §4). L'étiquette de domaine MUST être
  encodée selon la règle **string** ci-dessous — c'est-à-dire un varint LEB128 non signé
  de sa longueur en octets UTF-8 suivi de ses octets UTF-8 — exactement comme tout autre
  champ de chaîne, sans terminateur nul et sans traitement particulier. Des étiquettes
  distinctes rendent une signature portant sur un type de message intransférable à un
  autre.
- **string** — longueur en octets sous forme de varint LEB128 non signé, puis les octets
  UTF-8 bruts. Une chaîne vide MUST s'encoder comme l'octet unique `0x00` (une longueur
  nulle et aucune charge utile).
- **bytes** — longueur en octets sous forme de varint LEB128 non signé, puis les octets
  bruts. (Le champ `Signature` n'est jamais encodé ; il est exclu.)
- **bool** — un seul octet, exactement `0x00` (faux) ou `0x01` (vrai) ; aucune autre
  valeur d'octet n'est valide.
- **int64 / uint64** — 8 octets fixes, gros-boutiste (octet de poids fort en premier).
  `int64` MUST être encodé selon sa réinterprétation en complément à deux sous forme de
  `uint64` : ajouter 2⁶⁴ à une valeur négative, puis émettre les 8 octets gros-boutistes.
  Ainsi `-1` → `ff ff ff ff ff ff ff ff`, `-2` → `ff ff ff ff ff ff ff fe`, et le
  minimum `int64` (`-9223372036854775808`) → `80 00 00 00 00 00 00 00`. Ni signe et
  magnitude, ni zigzag, ni longueur variable.
- **time** — l'instant sous forme d'`int64` en **nanosecondes** Unix UTC (selon la règle
  int64 ci-dessus, gros-boutiste, 8 octets). Les composantes de fuseau et d'horloge
  monotone sont abandonnées. La précision à la nanoseconde est préservée exactement ; la
  valeur MUST NOT être tronquée aux micro- ou millisecondes. Cela survit à un aller-retour
  par transport JSON et est indépendant du fuseau horaire. Un instant hors de la plage
  représentable des nanosecondes Unix UTC en `int64` (approximativement les années
  1678–2262) est hors de portée pour la v0 et MUST être traité comme une erreur
  d'encodage plutôt que replié ou borné. L'encodeur de référence l'impose :
  `canon.Buffer.Time` n'ajoute rien et mémorise `canon.ErrTimeOutOfRange` (récupérable via
  `Buffer.Err()`) au lieu d'émettre une valeur repliée, de sorte que deux instants
  séparés d'environ 584 ans ne peuvent jamais aboutir aux mêmes octets. Un producteur
  MUST rejeter via `canon.ValidTime` avant de signer ; un vérificateur/décodeur MUST
  rejeter un message dont l'horodatage échoue à ce contrôle.
- **champ répété** — un *nombre* (le nombre d'éléments) sous forme de varint LEB128 non
  signé, encodé EN PREMIER, puis les champs de chaque élément encodés dans l'ordre,
  entièrement en ligne, selon les règles de primitives normales ci-dessus. Un élément
  répété qui est lui-même une chaîne est donc préfixé par sa longueur comme toute autre
  chaîne ; les champs de structures imbriquées sont émis dans l'ordre donné pour cet
  élément au §4. Il n'y a pas de séparateur entre les éléments ni de terminateur après le
  dernier. Un nombre égal à `0` est l'octet unique `0x00` suivi d'aucun élément.

Un vérificateur recalcule les octets canoniques à partir des champs reçus et vérifie la
signature ed25519 au regard de la clé publique obtenue. Un vérificateur MUST rejeter une
signature dont la longueur n'est pas `ed25519.SignatureSize` (64) avant de vérifier.

Les octets canoniques d'un message sont **exactement** la concaténation de l'étiquette
de domaine suivie de chaque champ listé dans la clause `Order:` de ce message au §4, dans
cet ordre, chacun encodé selon sa règle de primitive ci-dessus. Il n'y a pas
d'encapsulation externe : pas d'octet de version en tête, pas de préfixe de longueur
totale enveloppant le message, et pas d'octets de queue après le dernier champ. La liste
`Order:` du §4 est exhaustive et fait autorité — le ou les **champs de signature** du
message sont les SEULS champs exclus, et aucun champ ne figure dans les octets canoniques
sans y être nommé. Pour les douze messages de base du §4, ce champ exclu unique est
`Signature`. Les messages opérateur de la couche C du §11 portent **deux** signatures
(`Sig0` et `Sig1`) sur les mêmes octets canoniques ; toutes deux sont exclues exactement
comme l'est `Signature` ici et, comme `Signature`, aucune n'apparaît dans une clause
`Order:`. Un vérificateur n'encode donc jamais aucun champ de signature dans les octets
qu'il recalcule, que le message porte une signature ou deux.

---

## 4. Messages

Pour chaque message : son étiquette de domaine, ses champs dans l'ordre canonique, et qui
le signe. Tous les horodatages suivent la règle `time` du §3.

La clause `Order:` ci-dessous est la SEULE autorité pour l'ordre canonique des champs, y
compris l'ordre des champs de structures imbriquées (par ex. `Capacity.VCPUs` avant
`Capacity.MemMB`, et chaque imprimante comme `Kind` puis `Model`). Un implémenteur MUST
encoder les champs selon la séquence `Order:` et MUST NOT déduire l'ordre d'aucune autre
source. En particulier, l'exemple `testdata/vectors.json` (§10) liste les `fields` de
chaque message dans l'**ordre alphabétique des clés pour la lisibilité, qui n'est PAS
l'ordre d'encodage canonique** — suivre l'ordre des clés JSON produirait des octets
erronés et des signatures invalides.

### 4.1 CapabilityListing — signé par le NŒUD
Étiquette de domaine : `sohocloud/listing/v0`

Order : `NodeID` (string), `Class` (string), `Printers` (repeated : `Kind` string,
`Model` string), `GPUs` (repeated : `API` string, `Model` string, `VRAMMB`
int64), `Capacity.VCPUs` (int64), `Capacity.MemMB` (int64),
`Capacity.DiskMB` (int64), `Capacity.StorageCommitMB` (int64),
`Capacity.PrintQPS` (int64), `OptIn.Compute` (bool), `OptIn.Print` (bool),
`OptIn.Storage` (bool), `IssuedAt` (time), `Seq` (uint64).

`Class` ∈ {`micro`, `standard`, `server`}. `Kind` ∈ {`traditional`, `threed`}.
`API` ∈ {`vulkan`, `nnapi`, `cuda`, `metal`}.

`Capacity.DiskMB` est l'espace de travail temporaire disponible pour une tâche en cours ;
`Capacity.StorageCommitMB` est le stockage de longue durée que le nœud s'engage à
conserver pour le réseau (hébergement de fragments). Les deux sont délibérément distincts
afin qu'un nœud puisse offrir l'un sans l'autre. La manière dont les données stockées sont
chiffrées, fragmentées et auditées est une préoccupation du frontend/de l'agent hors de ce
protocole — coordination uniquement, le fil ne transporte jamais de contenu stocké.

Annoncer un GPU constitue l'adhésion au travail GPU : une annonce sans `GPUs`
ne reçoit aucun travail GPU, et un nœud retire un GPU en l'omettant de son annonce
suivante.

`Seq` est strictement monotone par nœud. Un coordinateur MUST rejeter une annonce dont le
`Seq` ne dépasse pas strictement le dernier vu pour ce nœud (protection contre le rejeu et
le retour en arrière).

`OptIn` est **indicatif** (voir §5.1).

### 4.2 Heartbeat — signé par le NŒUD
Étiquette de domaine : `sohocloud/heartbeat/v0`

Order : `NodeID` (string), `SentAt` (time), `Seq` (uint64).

Même discipline de `Seq` strictement monotone qu'au §4.1.

### 4.3 Assignment — signé par le COORDINATEUR
Étiquette de domaine : `sohocloud/assignment/v0`

Order : `JobID` (string), `NodeID` (string), `Spec.Workload` (string),
`Spec.Image` (string), `Spec.Args` (repeated string), `Spec.PrinterKind`
(string), `Spec.GPUAPI` (string), `Spec.GPUMinVRAMMB` (int64),
`Fee.ContributorShareBps` (int64), `Fee.PlatformFeeBps` (int64),
`OfferedAt` (time).

`Spec.GPUAPI` ∈ {``, `vulkan`, `nnapi`, `cuda`, `metal`} — un indice de routage
indicatif comme `PrinterKind` ; vide signifie qu'aucun GPU n'est requis et `GPUMinVRAMMB`
MUST alors valoir `0`.

Les conditions de frais sont jointes en ligne afin qu'un nœud voie la répartition avant de
s'engager. Dans le modèle pull, il n'y a pas de message d'acceptation distinct.

### 4.4 Decline — signé par le NŒUD
Étiquette de domaine : `sohocloud/decline/v0`

Order : `JobID` (string), `NodeID` (string), `Reason` (string), `DeclinedAt`
(time).

`Reason` ∈ {`local_policy`, `capacity`, `unavailable`}. `local_policy` est la voie de
désengagement (voir §5.1).

### 4.5 JobReport — signé par le NŒUD
Étiquette de domaine : `sohocloud/jobreport/v0`

Order : `JobID` (string), `NodeID` (string), `ExitCode` (int64), `FailureCause`
(string), `TmpfsExhausted` (bool), `StartedAt` (time), `FinishedAt` (time).

Ce rapport signé est le fait dont découlent le comptage et le versement. Ses champs
reflètent l'analyse de fin d'exécution de l'exécuteur déjà en service dans SoHoLINK.

### 4.6 FeeDeclaration — signé par le COORDINATEUR
Étiquette de domaine : `sohocloud/fee/v0`

Order : `CoordinatorID` (string), `Terms.ContributorShareBps` (int64),
`Terms.PlatformFeeBps` (int64), `EffectiveAt` (time), `Seq` (uint64).

`ContributorShareBps + PlatformFeeBps` SHOULD être égal à `10000`.

### 4.7 StorageLease — signé par le COORDINATEUR
Étiquette de domaine : `sohocloud/lease/v0`

Order : `LeaseID` (string), `NodeID` (string), `ShardRef` (bytes),
`SizeClass` (int64), `Fee.ContributorShareBps` (int64),
`Fee.PlatformFeeBps` (int64), `IssuedAt` (time), `ExpiresAt` (time),
`Seq` (uint64).

Le stockage est un BAIL, non une tâche : une obligation continue de conserver un
fragment scellé pendant une durée limitée. `ShardRef` est une adresse de contenu opaque
de 32 octets et `SizeClass` une taille de charge utile quantifiée — le protocole MUST NOT
transporter les tailles réelles des objets ni le contenu stocké ; le chiffrement, le
remplissage et la fragmentation sont des préoccupations du frontend, au-dessus de ce
goulot. Le renouvellement est un nouveau `StorageLease` pour le même `LeaseID` avec un
`Seq` strictement supérieur (même règle anti-retour en arrière qu'au §4.1). Les conditions
de frais voyagent en ligne, comme au §4.3.

### 4.8 LeaseDecline — signé par le NŒUD
Étiquette de domaine : `sohocloud/lease-decline/v0`

Order : `LeaseID` (string), `NodeID` (string), `Reason` (string),
`DeclinedAt` (time).

`Reason` ∈ {`local_policy`, `capacity`, `unavailable`} ; `local_policy` est la voie de
désengagement (§5.1) pour le stockage, exactement comme le §4.4 l'est pour les tâches.

### 4.9 LeaseRelease — signé par le NŒUD
Étiquette de domaine : `sohocloud/lease-release/v0`

Order : `LeaseID` (string), `NodeID` (string), `ReleasedAt` (time).

Un nœud peut toujours cesser de conserver (la souveraineté inclut le départ). La
libération signée met fin à son comptage ; replacer le fragment relève du frontend.

### 4.10 ProofChallenge / ProofResponse — réponse signée par le NŒUD
Étiquette de domaine (réponse) : `sohocloud/proof/v0`

Un `ProofChallenge{LeaseID, Offset, Length, Nonce (16 bytes), IssuedAt}` n'est
PAS signé : les nœuds récupèrent les défis en interrogeant par le canal authentifié
(modèle pull), et un défi n'engage personne à rien. L'artefact signé est la
réponse.

Order de la réponse : `LeaseID` (string), `NodeID` (string), `Offset` (int64),
`Length` (int64), `Nonce` (bytes), `Digest` (bytes), `RespondedAt` (time).

`Digest` MUST être calculé exactement comme
`SHA-256(Nonce || uint64be(Offset) || uint64be(Length) || sealed[Offset : Offset+Length])`
sur les octets du fragment scellé. La réponse reformule la plage et le nonce du défi afin
de valoir à elle seule comme fait de comptage non répudiable — le pendant pour le
stockage du §4.5. Un vérificateur MUST rejeter une réponse dont il a déjà accepté le
couple `(LeaseID, Nonce)` : les nonces sont à usage unique, et c'est ce qui déjoue le
rejeu d'une réponse enregistrée.

### 4.11 KeyRotation — signé par la clé ACTUELLE du nœud
Étiquette de domaine : `sohocloud/node-rotate/v0`

Order : `NodeID` (string), `NewPublicKey` (bytes), `Algo` (string),
`RotatedAt` (time), `Seq` (uint64).

La possession de la clé sortante autorise à nommer son successeur. Un vérificateur
MUST valider la signature au regard de la clé qu'il détient actuellement pour le nœud
(une rotation signée par le successeur est une auto-installation et MUST échouer),
MUST imposer un `Seq` strictement monotone par nœud, et MUST rejeter une rotation
dont la `NewPublicKey` n'est pas une clé bien formée pour `Algo` (pour `ed25519`, pas
exactement 32 octets) ou dont l'`Algo` est inconnu — sans quoi un vérificateur qui
commencerait à vérifier les messages suivants au regard d'une `NewPublicKey` malformée
tomberait en défaut. Ce n'est qu'après la réussite de tous ces contrôles que le
vérificateur commence à vérifier les messages ultérieurs du nœud au regard de
`NewPublicKey`. `Algo` ∈ {`ed25519`} en v0.

### 4.12 KeyRevocation — signé par la clé RÉVOQUÉE
Étiquette de domaine : `sohocloud/node-revoke/v0`

Order : `NodeID` (string), `RevokedPublicKey` (bytes), `RevokedAt` (time),
`Seq` (uint64).

Supprime une clé sans successeur. Un vérificateur MUST vérifier la signature au regard
de la clé à laquelle il **fait actuellement confiance** pour `NodeID`, et MUST exiger que
`RevokedPublicKey` soit égale à cette clé de confiance. Une révocation n'est jamais
honorée qu'à l'égard de la clé exacte qu'elle nomme et supprime. C'est la liaison
anti-abus : un inconnu peut autosigner une révocation portant le `NodeID` d'une victime
et sa propre clé dans `RevokedPublicKey`, mais comme cette clé n'est pas celle à laquelle
le coordinateur fait confiance pour la victime, le contrôle d'égalité échoue et la
révocation est rejetée — un nœud ne peut pas être mis hors ligne par quiconque ne détient
pas déjà sa clé actuelle. (Des versions antérieures disaient « honorer sans condition » ;
cette formulation est remplacée — la liaison à la clé de confiance est obligatoire.)
Étant donné la clé actuelle, la lecture sûre d'une révocation est « cesser de faire
confiance à cette clé ». La réinscription après révocation se fait hors bande, par la
même voie que l'inscription initiale, jamais par un message sur le fil qu'un voleur de
clé pourrait forger.

---

## 5. Invariants

### 5.1 Le désengagement est imposé localement, jamais sur le fil
`CapabilityListing.OptIn` et les indices de routage de `JobSpec` sont **indicatifs**.
Le coordinateur n'est PAS une frontière de sécurité pour le désengagement. Un nœud MUST
imposer le désengagement au regard de sa propre liste d'autorisation de confiance locale
et MUST refuser le travail que sa politique locale interdit — via
`Decline{Reason: local_policy}` — quoi qu'il ait annoncé auparavant et quoi que le
coordinateur propose.

### 5.2 La politique d'appariement ne fait pas partie du protocole
L'interface `Coordinator` n'expose aucun classement ni aucune notation. La manière dont un
coordinateur choisit quel nœud reçoit quelle tâche est propre à la mise en œuvre. Le
protocole MUST NOT se doter d'un champ qui dicte l'appariement, car cela souderait les
consommateurs à la politique d'un seul coordinateur et irait à l'encontre de la
possibilité de partir que l'interface existe pour protéger.

### 5.3 Les frais sont lisibles et non rétroactifs
Les conditions de frais d'un coordinateur sont publiées sous forme de `FeeDeclaration`
signée et horodatée. Les conditions MUST NOT changer rétroactivement pour un travail déjà
proposé ; un changement est une nouvelle déclaration signée avec un `EffectiveAt`
ultérieur. Cela maintient les frais contestables plutôt qu'opaques (problème ouvert n° 7).

### 5.4 Le coordinateur est un rôle, non un pivot
Rien de ce dont un nœud ou un frontend a besoin ne peut transiter *exclusivement* par un
coordinateur spécifique. Un nœud MAY servir plusieurs coordinateurs ; un frontend MAY
faire tourner son propre coordinateur ou apparier directement. Tout déploiement dans
lequel un coordinateur unique est de fait incontournable constitue une recentralisation
et MUST être étiqueté comme tel, avec une voie vers des alternatives.

### 5.5 Numéros de séquence monotones
Les valeurs `Seq` d'un nœud (annonces, heartbeats) sont strictement monotones par nœud.
Un coordinateur MUST rejeter un `Seq` non croissant.

---

## 6. Opérations du coordinateur (modèle pull)

Le modèle est **pull** : le coordinateur n'appelle jamais le nœud. Opérations :

| Opération       | Sens             | Corps / résultat                     |
|-----------------|------------------|--------------------------------------|
| `SubmitListing` | node → coord     | `CapabilityListing`                  |
| `Heartbeat`     | node → coord     | `Heartbeat`                          |
| `PollJobs`      | node → coord     | `NodeID` → `[]Assignment`            |
| `Decline`       | node → coord     | `Decline`                            |
| `ReportJob`     | node → coord     | `JobReport`                          |
| `Fees`          | quiconque → coord | → `FeeDeclaration`                  |

Un coordinateur MUST appliquer la liaison du §2 à `PollJobs` (et à toute opération
nommant un nœud) avant d'agir. Comme le coordinateur n'initie jamais de contact avec un
nœud, il n'y a pas de service côté nœud à spécifier.

---

## 7. Transport

HTTP+JSON est le transport **de référence** (`transport/httpjson`), PAS le fil lui-même.
Il n'est importé par aucun paquet du cœur. Une mise en œuvre conforme MAY utiliser
n'importe quel transport en implémentant les opérations du coordinateur au regard des
octets canoniques du §3. JSON n'est qu'un encodage de transport ; les signatures ne sont
jamais calculées sur du JSON, de sorte qu'un transport différent interopère sans
renégocier les signatures.

Les points de terminaison HTTP+JSON de référence :

```
POST /v0/listing     body: CapabilityListing   -> 204
POST /v0/heartbeat   body: Heartbeat           -> 204
GET  /v0/jobs?node_id=<id>                      -> 200 []Assignment
POST /v0/decline     body: Decline             -> 204
POST /v0/report      body: JobReport           -> 204
GET  /v0/fees                                   -> 200 FeeDeclaration
```

La liaison SPIFFE (§2) est appliquée par un intergiciel placé devant le gestionnaire ; le
gestionnaire de référence ne fait que câbler les routes et le JSON et n'est pas
l'authentificateur.

---

## 8. Problèmes ouverts que touche ce protocole

L'honnêteté sur les lacunes non résolues est la méthode, non un embarras.

- **n° 7 Le calcul souverain offre une sortie mécanique, non économique.** C'est la
  couche dont traite le n° 7. Le §5.4 (coordinateur comme rôle) et le §5.3 (frais
  lisibles) s'opposent à la recentralisation des coordinateurs ; le démarrage à froid
  reste un résidu, et la portabilité de la réputation qui l'atténuerait est une
  préoccupation du frontend/de la gouvernance, non de ce protocole.
- **n° 6 L'honnêteté du calcul est hors de portée.** Le caractère équitable de
  l'appariement ou du prix d'un coordinateur n'est pas prouvé ici, seulement rendu lisible
  (§5.2, §5.3). `anchor/` réserve la place pour ancrer les *affirmations* d'emploi (les
  faits de `JobReport`) dans un journal attesté par témoins afin qu'elles soient au moins
  auditables. C'est une ébauche étiquetée, non construite.
- **n° 4 Adhésion / intégration.** Comme un nœud se nomme et signe pour lui-même (§2), un
  appareil à faibles ressources (par ex. un téléphone) peut s'inscrire comme nœud. La
  question de savoir si l'intégration des nœuds est ouverte ou soumise à condition est une
  décision de gouvernance/de valeurs laissée ouverte.

## 9. Gestion des versions

`v0` est instable. Les étiquettes de domaine intègrent `v0` ; une modification incompatible
des champs canoniques d'un message incrémente son étiquette. Les consommateurs MUST
traiter une étiquette de domaine inconnue comme un message invérifiable.

---

## 10. Vecteurs de test

`testdata/vectors.json` est le **jeu de données de conformité normatif** de cette
spécification. Une mise en œuvre indépendante est conforme en matière d'encodage canonique
et de signature si et seulement si elle reproduit, octet pour octet, chaque valeur de ce
fichier à partir des mêmes entrées. Le jeu de données est généré par les encodeurs Go de
référence et vérifié de manière croisée par une reproduction indépendante n'utilisant que
la bibliothèque standard ; une mise en œuvre tierce conforme MUST le reproduire
exactement. Si la prose des §3/§4 et le jeu de données venaient à diverger, ce serait un
défaut de la spécification à corriger, non une permission de diverger — mais le jeu de
données est l'énoncé exécutable des §3/§4.

Tout l'hexadécimal du jeu de données est en minuscules. Il est autodescriptif : une `note`
de premier niveau et un bloc `encoding` reformulent les règles de primitives du §3 pour un
lecteur qui ouvre d'abord le JSON.

### 10.1 Structure

- **`primitives`** — encodages isolés de chaque primitive du §3, afin qu'un implémenteur
  puisse valider l'encodeur de bas en haut avant d'assembler des messages. Chaque entrée
  comporte un `kind` (`uvarint`, `int64`, `uint64`, `bool`, `string`, `time`), une
  `input` lisible par un humain, et les `bytes_hex` attendus. Ceux-ci couvrent les cas
  limites qui fixent le format : des valeurs varint franchissant chaque seuil d'octet de
  continuation (`0, 1, 127, 128, 255, 256, 300, 16384, 1073741824`) ; `int64`, y compris
  `-1`, `-2`, `MaxInt64` et `MinInt64` ; `uint64`, y compris `MaxUint64` ; les deux
  booléens ; la chaîne vide et une chaîne UTF-8 multioctet ; et `time` à l'époque Unix et
  à un instant fixe à la nanoseconde.

- **`messages`** — une entrée par type de message signé (chaque message signé du §4). Chaque entrée
  enregistre : `name`, `domain_tag`, `signer` (`node` ou `coordinator`), la `seed_hex`
  ed25519 de 32 octets et la `public_key_hex` dérivée, les `fields` du message, les
  `canonical_bytes_hex` (la sortie de `CanonicalBytes()`, `Signature` exclue), et la
  `signature_hex` (ed25519 sur exactement ces octets canoniques).

- **`operator_messages`** — une entrée par type de message opérateur de la couche C (les
  trois du §11). Comme les messages opérateur sont en 2-of-7, chaque entrée diffère d'une
  entrée `messages` de base : `signer` vaut toujours `operator` ; un tableau `keys` liste
  les sept clés publiques enregistrées (`index`, `seed_hex`, `public_key_hex`, `algo` —
  les graines ne sont présentes QUE pour que le jeu de données soit autonome et
  reproductible, jamais parce que le réseau détiendrait des clés privées) ; `idx0`/`idx1`
  sont les deux indices de signature ; et au lieu d'une seule `signature_hex`, il y a
  `sig0_hex` et `sig1_hex`, TOUTES DEUX prises sur les MÊMES `canonical_bytes_hex`. Une
  mise en œuvre conforme reproduit `canonical_bytes_hex` à partir de l'ordre des champs du
  §11 et vérifie les deux signatures à leurs indices.

  > **Piège de l'ordre des champs.** L'objet `fields` de chaque message liste les clés
  > dans l'**ordre alphabétique pour la lisibilité, NON dans l'ordre canonique.** L'ordre
  > canonique provient UNIQUEMENT de la clause `Order:` de ce message au §4. Un
  > implémenteur qui encode les champs dans l'ordre des clés JSON produira des octets
  > erronés et des signatures qui ne se vérifient pas.

### 10.2 Comment utiliser le jeu de données

1. Reproduire chaque `primitives[].bytes_hex` à partir de son `input` à l'aide du §3.
2. Pour chaque message, dériver la paire de clés à partir de `seed_hex` (ed25519 à partir
   d'une graine de 32 octets est déterministe), confirmer qu'elle donne `public_key_hex`,
   encoder les champs selon l'`Order:` du §4 et confirmer que l'on obtient
   `canonical_bytes_hex`, puis vérifier `signature_hex` au regard de `public_key_hex` sur
   vos octets dérivés indépendamment. La vérification de la signature sur vos PROPRES
   octets est la vérification croisée la plus forte : elle ne peut réussir que si vos
   octets sont identiques à ceux du générateur de référence.

Le test de référence se trouve dans `vectors/vectors_test.go`. Il régénère chaque vecteur
à partir du code actuel et affirme l'égalité octet pour octet avec le JSON validé dans le
dépôt, de sorte qu'une dérive entre le code et le jeu de données fait échouer la
construction ; il réanalyse aussi le fichier validé et vérifie chaque signature.
Régénérez intentionnellement avec `go test ./vectors -update` (ou la variable
d'environnement `SOHOCLOUD_UPDATE_VECTORS`). Tous les vecteurs utilisent des graines à
remplissage d'octets fixe et des horodatages codés en dur — il n'y a ni `crypto/rand` ni
`time.Now()`, de sorte que la sortie est entièrement déterministe.

### 10.3 Exemple détaillé : un `Heartbeat`, octet par octet

Voici le vecteur `Heartbeat` (`§4.2`) décodé champ par champ, afin qu'un
réimplémenteur puisse suivre l'encodage complet du début à la fin. Entrées :
`NodeID = "node-alpha"`, `SentAt = ` nanosecondes Unix `1700000050500000000`,
`Seq = 43`.

Octets canoniques (`canonical_bytes_hex`) :

```
16 736f686f636c6f75642f6865617274626561742f7630 0a 6e6f64652d616c706861 17979d09f832d900 000000000000002b
```

Décodage de gauche à droite :

| Octets | Primitive | Signification |
|-------|-----------|---------|
| `16` | longueur uvarint | Longueur de l'étiquette de domaine = `0x16` = **22** octets. |
| `73 6f 68 6f 63 6c 6f 75 64 2f 68 65 61 72 74 62 65 61 74 2f 76 30` | octets UTF-8 | L'étiquette de domaine `sohocloud/heartbeat/v0` (22 octets). Avec le préfixe de longueur, ces deux lignes forment l'encodage `string` de l'étiquette de domaine (§3). |
| `0a` | longueur uvarint | Longueur de `NodeID` = `0x0a` = **10** octets. |
| `6e 6f 64 65 2d 61 6c 70 68 61` | octets UTF-8 | `NodeID` = `node-alpha` (10 octets). |
| `17 97 9d 09 f8 32 d9 00` | int64, gros-boutiste | `SentAt` en nanosecondes Unix UTC. `0x17979d09f832d900` = `1700000050500000000`. C'est la règle `time` (§3) : l'instant réduit à un `int64` et émis sous forme de 8 octets gros-boutistes. |
| `00 00 00 00 00 00 00 2b` | uint64, gros-boutiste | `Seq` = `0x2b` = **43**. 8 octets fixes, pas de varint. |

La concaténation d'exactement ces six groupes — étiquette de domaine (longueur + octets),
`NodeID` (longueur + octets), `SentAt`, `Seq` — est la chaîne d'octets canonique complète.
Il n'y a aucune encapsulation avant l'étiquette de domaine ni aucun octet après `Seq`. La
signature ed25519 (`signature_hex`) est prise sur précisément cette chaîne ; le champ
`Signature` lui-même n'en fait jamais partie.

Remarquez que l'ordre des champs ici (`NodeID`, `SentAt`, `Seq`) suit l'`Order:` du §4.2,
qui est aussi — par coïncidence pour ce petit message — l'ordre dans lequel un lecteur les
listerait ; pour des messages plus grands comme `CapabilityListing`, l'`Order:` du §4 et
les clés alphabétiques de `fields` dans `vectors.json` diffèrent, et seul l'ordre du §4
est canonique.

---

## 11. Identité d'opérateur (couche C)

Les §1–§10 régissent les messages de **nœud** et de **coordinateur**. Cette section ajoute
une troisième identité, orthogonale : l'**opérateur** — un frontend (par ex. Cloudy) qui
termine en son sein l'identité des membres et des nœuds et présente un unique justificatif
tournant à un coordinateur. Elle est délibérément séparée de l'identité de charge de
travail des nœuds (§2) : les messages de coordination de base (§4) sont INCHANGÉS et ne
sont PAS signés avec un justificatif d'opérateur.

Un opérateur détient **sept** paires de clés Ed25519 (indices `0..6`) et signe chaque
message avec **deux indices distincts** (une discipline **2-of-7**). Le réseau
n'enregistre que les sept clés **publiques** par opérateur ; **aucune clé privée
d'opérateur n'est jamais détenue par le réseau**, et chaque chemin de vérification
ci-dessous ne prend que des clés publiques.

- **Le 2-of-7 relève de l'anti-substitution / de l'hygiène de rotation, NON de la
  sécurité à seuil.** Si les sept clés privées résident dans un même processus, une seule
  compromission de l'hôte livre les sept. On ne compte pas sur le 2-of-7 pour résister à
  la compromission d'une seule clé, sauf si la garde est répartie entre domaines de
  confiance. Ceci est énoncé pour qu'aucun lecteur ne lui accorde une confiance excessive.
- **Liaison à l'algorithme.** Chaque message opérateur lie une chaîne d'algorithme `Algo`
  dans ses octets canoniques. `v0` autorise exactement `ed25519`. Un vérificateur MUST
  rejeter un message dont l'`Algo` est inconnu, et MUST rejeter un message dont l'`Algo`
  n'est pas égal à l'`algo` enregistré de la clé à chaque indice de signature.
  L'ensemble de clés actif d'un opérateur MUST NOT mélanger les algorithmes. Le contrôle
  de longueur de signature est propre à chaque algorithme (`ed25519` = 64 octets), et non
  une constante codée en dur, de sorte que la rotation atomique planifiée de l'ensemble
  complet vers un algorithme post-quantique (une NOUVELLE étiquette de domaine, par ex.
  `sohocloud/operator/v1-mldsa`) réutilise la même règle. Une signature est
  intransférable d'un algorithme à l'autre.
- **Point de jonction pour la migration.** La mise en œuvre de référence fait passer la
  signature et la vérification des opérateurs par une interface `Signer`/`Verifier`,
  implémentée aujourd'hui avec l'Ed25519 de la bibliothèque standard, précisément pour que
  l'algorithme puisse être remplacé plus tard en ajoutant une étiquette de domaine `v1`
  sans toucher aux messages de base. Il s'agit d'une remarque de mise en œuvre ; le contrat
  sur le fil, ce sont les ordres de champs et les étiquettes ci-dessous.

**Encodage des champs opérateur (sans traitement particulier).** Chaque champ des clauses
`Order:` du §11 est encodé selon sa règle de primitive ordinaire du §3 ; le §11
n'introduit aucun type nouveau ou à largeur fixe. En particulier :

- `OperatorID` et `Algo` sont des **chaînes** ordinaires du §3 — un varint LEB128 non
  signé de la longueur en octets UTF-8 suivi des octets UTF-8. `Algo` n'est PAS un jeton
  nu ni une étiquette d'énumération à largeur fixe, bien que `v0` n'autorise que l'unique
  valeur `ed25519` ; il MUST porter son préfixe de longueur comme toute autre chaîne (par
  ex. `ed25519` s'encode `07 65 64 32 35 35 31 39`). `Nonce`, `NewPublicKey` et
  `Challenge` sont des **bytes** ordinaires du §3 (préfixe de longueur LEB128 puis octets
  bruts).
- `Idx0`, `Idx1` et `KeyIndex` sont chacun encodés comme un **`uint64` gros-boutiste fixe
  de 8 octets** selon la règle `uint64` du §3, exactement comme `Seq` — NON comme un
  varint et NON comme un seul octet — même si leurs valeurs se situent dans la petite
  plage `0..6`. Un implémenteur qui encode un indice sous forme de varint ou d'un seul
  octet produit des octets erronés et une signature invalide. (Exemple : l'indice `2`
  s'encode `00 00 00 00 00 00 00 02`.)
- Les deux signatures `Sig0` et `Sig1` sont **exclues** des octets canoniques de chaque
  message du §11, exactement comme `Signature` l'est pour les messages du §4 (§3). Ce sont
  les seuls champs exclus et, comme `Signature`, ils n'apparaissent dans aucune clause
  `Order:` ci-dessous. Il n'existe pas de troisième champ de signature. `Idx0`/`Idx1` (qui
  nomment les clés de signature) SONT dans les octets signés ; les *octets* de signature
  ne le sont pas.

### 11.0 Exigences de vérification (tous les messages opérateur)

Un vérificateur MUST, avant d'honorer tout message opérateur :

1. Rejeter si le nonce du message est absent ou plus court que **16 octets** (lorsque le
   message comporte un nonce). Un nonce vide MUST NEVER être traité comme « ignorer le
   cache anti-rejeu ».
2. Rejeter si `Algo` n'est pas un algorithme pris en charge pour cette version.
3. Rejeter si les deux indices de signature `Idx0` et `Idx1` sont égaux.
4. Rejeter si l'un ou l'autre indice est hors de `0..6`.
5. Rejeter si aucune clé enregistrée active n'existe à l'un ou l'autre indice, ou si
   l'`algo` de l'une ou l'autre clé enregistrée n'est pas égal à l'`Algo` du message.
5a. Rejeter si les clés publiques enregistrées à `Idx0` et `Idx1` sont identiques octet
   pour octet, même si les indices diffèrent. Les sept clés publiques enregistrées d'un
   opérateur MUST être distinctes deux à deux ; un enregistrement ou une rotation MUST NOT
   installer une clé publique qui existe déjà à un autre indice (voir §11.2). Ceci est
   imposé pour que la propriété anti-substitution du 2-of-7 ne puisse pas se dégrader
   silencieusement en 1-of-1 : si deux indices portaient la même clé, une seule clé privée
   pourrait produire les deux signatures. Ce contrôle porte sur les *clés*, et non
   seulement sur les indices (règle 3), et il est requis en plus de celle-ci.
6. Rejeter si la longueur de l'une ou l'autre signature n'est pas la longueur attendue
   pour `Algo`.
7. Recalculer les octets canoniques et rejeter à moins que les DEUX signatures ne se
   vérifient à leurs indices sur exactement ces octets.

L'anti-rejeu (un ensemble durable, à défaillance fermée, de `Seq` à fenêtre glissante et
de nonces, délimité par `(operator, coordinator)`) est une obligation côté coordinateur
superposée AU-DESSUS de ce qui précède ; il ne fait pas partie des octets canoniques et
est hors de portée de cette spécification d'encodage.

### 11.1 OperatorTransmission — signé par l'OPÉRATEUR (2 sur 7)
Étiquette de domaine : `sohocloud/operator/v0`

Order : `OperatorID` (string), `Ts` (int64 en nanosecondes Unix UTC — l'`int64` brut,
NON redérivé via un `time.Time`, de sorte que la valeur contrôlée en plage est exactement
la valeur signée), `Nonce` (bytes, `len >= 16`), `Seq` (uint64), `Algo` (string),
`Idx0` (uint64), `Idx1` (uint64).

Les deux signatures `Sig0` (par la clé à `Idx0`) et `Sig1` (par la clé à
`Idx1`) sont prises sur ces octets canoniques et en sont EXCLUES, exactement
comme `Signature` l'est ailleurs. `Ts`, `Nonce`, `Seq`, les deux indices et `Algo` sont
tous dans les octets signés.

### 11.2 OperatorRotation — signé par l'OPÉRATEUR (2 sur 7)
Étiquette de domaine : `sohocloud/operator-rotate/v0`

Order : `OperatorID` (string), `KeyIndex` (uint64 — l'indice dont la clé est
remplacée), `NewPublicKey` (bytes — la clé publique de remplacement générée par
l'opérateur), `Algo` (string), `Ts` (int64 en nanosecondes Unix UTC), `Nonce` (bytes,
`len >= 16`), `Seq` (uint64), `Idx0` (uint64), `Idx1` (uint64).

Ce message autorise la substitution d'une nouvelle clé publique. L'opérateur (jamais le
réseau) génère la nouvelle paire de clés. La nouvelle clé publique est DANS les octets
signés afin que ni un intermédiaire malveillant sur l'enregistrement hors bande ni un
canal d'administration compromis ne puissent injecter du matériel de clé : un
coordinateur MUST vérifier ce message (avec deux clés ACTUELLES, `Idx0`/`Idx1`) avant
d'enregistrer la nouvelle clé à `KeyIndex`. Un vérificateur MUST en outre rejeter une
`NewPublicKey` dont la longueur est incorrecte pour `Algo`. Un vérificateur MUST aussi
rejeter une `NewPublicKey` identique octet pour octet à une clé déjà enregistrée à un
AUTRE indice (elle violerait l'exigence de clés distinctes deux à deux de la règle 5a du
§11.0 et préparerait la dégradation 2-of-7 → 1-of-1) ; réenregistrer la même clé à son
PROPRE `KeyIndex` (un rafraîchissement sans effet) est permis. Pour préserver l'invariant
d'absence de mélange d'algorithmes, un vérificateur SHOULD aussi rejeter une rotation dont
l'`Algo` diffère de celui des autres clés enregistrées.

### 11.3 ConformanceResponse — signé par l'OPÉRATEUR (2 sur 7)
Étiquette de domaine : `sohocloud/operator-conformance/v0`

Order : `OperatorID` (string), `Challenge` (bytes — le défi fourni par le
vérificateur), `Algo` (string), `Idx0` (uint64), `Idx1` (uint64).

Il s'agit de la réponse signée d'un opérateur à un défi de conformité. Son étiquette de
domaine distincte la sépare par domaine du §11.1 : comme l'étiquette fait partie des
octets canoniques, une signature produite pour un défi de conformité ne peut JAMAIS se
vérifier comme `OperatorTransmission`, et inversement. Elle ne porte ni nonce ni `Seq`
propres — la fraîcheur provient du `Challenge` du vérificateur — de sorte que la règle de
nonce du §11.0 ne s'y applique pas ; tous les autres contrôles du §11.0 s'appliquent.
