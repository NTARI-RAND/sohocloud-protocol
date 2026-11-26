> Tradução comunitária (rascunho) — Política NTARI P2-002, Transmissão Multilíngue Global (Global Multilingual Broadcast). Fonte: SPEC.md (original em inglês, snapshot de 2026-10-05). Rascunho comunitário assistido por máquina, pendente de revisão por mantenedor regional conforme P2-002 §3.1. As especificações técnicas centrais permanecem em inglês conforme §2.2.
>
> Notou algum erro nesta tradução? Correções de tradução são contribuições
> valiosas e muito bem-vindas: faça um fork do repositório e abra um pull
> request em https://github.com/NTARI-RAND/sohocloud-protocol.

> **O original em inglês é normativo.** Esta tradução é oferecida para leitura, não para interpretação: onde ela divergir de `SPEC.md`, prevalece o inglês. As palavras-chave normativas (MUST, MUST NOT, SHOULD, SHOULD NOT, MAY), os nomes de campos, os rótulos e os identificadores são mantidos em inglês porque constituem o contrato de conformidade.

# sohocloud-protocol — Especificação de Conformidade

Versão: **v0 (UNSTABLE)** — os formatos de mensagem e o formato canônico de bytes
podem mudar sem garantias de compatibilidade até a v1.

Este documento é normativo e é um entregável do build, não uma reflexão tardia. O
direito, garantido pela AGPL, de fazer um fork da rede é vazio a menos que uma
implementação possa ser construída a partir de uma especificação legível, sem ler o
código-fonte deste módulo. Tudo o que estiver marcado abaixo como MUST / MUST NOT é
um requisito de conformidade; uma mudança que viole um deles não é uma versão menor
deste protocolo, é um protocolo diferente.

---

## 1. Escopo

Este protocolo governa a **camada de coordenação do substrato**: como um nó é
reconhecido, como ele anuncia o trabalho que se dispõe a fazer e como o trabalho
lhe é oferecido e relatado de volta. Três partes o falam:

- **Node (nó)** — um dispositivo de propriedade do participante que oferece computação
  e/ou impressão física. Um nó é dono da sua identidade e assina tudo o que emite.
  Soberano: ele escolhe a quais coordenadores servir e pode deixar qualquer um deles.
- **Coordinator (coordenador)** — pareia listagens com trabalho e registra os
  resultados. Um papel substituível (pluggable). O SoHoLINK é o coordenador de
  referência; um frontend MAY executar o seu próprio ou fazer o pareamento
  diretamente com os nós.
- **Frontend** — onde os membros realizam transações. Ele consome a coordenação por
  meio da mesma interface.

**Fora do escopo, deliberadamente.** Este protocolo NÃO é a economia de membros da
JFA. Ele não define escrow, nem crédito emitido por membros, nem pacto de reputação,
nem registro selado por diálogo, nem julgamento de disputas. Essas coisas pertencem
a cada frontend. A liquidação em moeda fiduciária que um coordenador realiza com os
seus nós (por exemplo, os pagamentos via Stripe do SoHoLINK) é a economia do
*substrato* e também fica fora desta especificação.

---

## 2. Identidade e vínculo

Um nó é nomeado por um `NodeID` — uma string estável que o próprio nó escolhe.

A identidade de carga de trabalho de um nó é um SVID SPIFFE emitido exatamente sob:

```
/node/<NodeID>
```

Um coordenador autoriza uma requisição como vinda de um nó comparando o caminho
SPIFFE autenticado de quem chama com `/node/<NodeID>` para o `NodeID` nomeado na
requisição.

- Um coordenador MUST rejeitar com **401** quando nenhuma identidade SPIFFE estiver
  presente.
- Um coordenador MUST rejeitar com **403** quando a identidade estiver presente, mas
  o seu caminho não for exatamente `/node/<NodeID>`.
- Um coordenador MUST NOT aceitar "qualquer SVID válido" como autorização para uma
  requisição que nomeie um nó diferente. Isso é um vetor de personificação.
- O vínculo MUST ser uma construção determinística de string, nunca uma consulta por
  nó em armazenamento mutável. Uma verificação condicionada a um campo que nunca é
  preenchido é inerte e MUST NOT ser tratada como fronteira de segurança.

O protocolo não distribui chaves públicas. Um coordenador obtém a chave de
verificação de um nó por fora do canal (out of band) — no registro, no material do
SVID ou em um registro de chaves.

---

## 3. Codificação canônica de bytes (para assinaturas)

Toda mensagem assinada é assinada sobre os seus **bytes canônicos**, calculados com
o campo `Signature` EXCLUÍDO. As assinaturas são **ed25519**.

A assinatura NÃO é feita sobre JSON nem sobre qualquer codificação de transporte. Ela
é feita sobre o seguinte formato determinístico, com prefixo de comprimento, para que
as assinaturas sejam estáveis entre plataformas, versões de Go e linguagens.

Primitivas, na ordem exata em que os campos são listados para cada mensagem no §4:

- **varint LEB128 sem sinal** — a primitiva de comprimento/contagem usada por
  `string`, `bytes` e `repeated`. Ela MUST ser codificada da seguinte forma,
  independentemente da linguagem: tome o valor sem sinal; emita 7 bits por byte,
  começando pelo grupo menos significativo (grupos little-endian); ligue o bit alto
  (`0x80`) de todo byte EXCETO o último, para marcar continuação; o último byte tem o
  bit alto desligado. O valor `0` é codificado como o byte único `0x00`. Esta é
  exatamente a codificação produzida por `encoding/binary.PutUvarint` do Go; o
  símbolo do Go é uma verificação cruzada, não a definição. (Exemplo resolvido:
  `300` = `0b100101100` → 7 bits baixos `0101100` com continuação → `0xAC`, bits
  altos `0000010` por último → `0x02`, resultando em `AC 02`.)
- **tag de domínio** — os bytes canônicos de toda mensagem começam com uma string de
  tag de domínio distinta (ver §4). A tag de domínio MUST ser codificada pela regra de
  **string** abaixo — isto é, um varint LEB128 sem sinal com o seu comprimento em
  bytes UTF-8, seguido dos seus bytes UTF-8 — exatamente como qualquer outro campo de
  string, sem terminador nulo e sem tratamento especial. Tags distintas tornam uma
  assinatura sobre um tipo de mensagem intransferível para outro.
- **string** — comprimento em bytes como varint LEB128 sem sinal, seguido dos bytes
  UTF-8 brutos. Uma string vazia MUST ser codificada como o byte único `0x00` (um
  comprimento zero e nenhuma carga útil).
- **bytes** — comprimento em bytes como varint LEB128 sem sinal, seguido dos bytes
  brutos. (O campo `Signature` nunca é codificado; ele é excluído.)
- **bool** — um único byte, exatamente `0x00` (false) ou `0x01` (true); nenhum outro
  valor de byte é válido.
- **int64 / uint64** — 8 bytes fixos, big-endian (byte mais significativo primeiro).
  `int64` MUST ser codificado como a sua reinterpretação em complemento de dois como
  `uint64`: some 2⁶⁴ a um valor negativo e então emita os 8 bytes big-endian. Assim,
  `-1` → `ff ff ff ff ff ff ff ff`, `-2` → `ff ff ff ff ff ff ff fe`, e o mínimo de
  `int64` (`-9223372036854775808`) → `80 00 00 00 00 00 00 00`. Sem sinal e
  magnitude, sem zigzag, sem comprimento variável.
- **time** — o instante como **nanossegundos** Unix UTC em `int64` (pela regra de
  int64 acima, big-endian, 8 bytes). Os componentes de localização e de relógio
  monotônico são descartados. A precisão de nanossegundos é preservada exatamente; o
  valor MUST NOT ser truncado para micro ou milissegundos. Isso sobrevive a uma ida e
  volta pelo transporte JSON e independe de fuso horário. Um instante fora do
  intervalo representável de nanossegundos Unix UTC em `int64` (aproximadamente os
  anos 1678–2262) está fora do escopo da v0 e MUST ser tratado como erro de
  codificação, em vez de sofrer estouro (wrap) ou ser limitado (clamp). O codificador
  de referência impõe isso: `canon.Buffer.Time` não acrescenta nada e retém (latch)
  `canon.ErrTimeOutOfRange` (recuperável via `Buffer.Err()`) em vez de emitir um valor
  com estouro, de modo que dois instantes separados por ~584 anos nunca podem se
  confundir em bytes idênticos. Um produtor MUST rejeitar via `canon.ValidTime` antes
  de assinar; um verificador/decodificador MUST rejeitar uma mensagem cujo carimbo de
  tempo não passe nessa verificação.
- **campo repeated** — uma *contagem* (o número de elementos) como varint LEB128 sem
  sinal, codificada PRIMEIRO, e então os campos de cada elemento codificados em ordem,
  totalmente em linha, usando as regras normais de primitivas acima. Um elemento
  repetido que seja ele próprio uma string tem, portanto, prefixo de comprimento como
  qualquer outra string; campos de structs aninhadas são emitidos na ordem dada para
  aquele elemento no §4. Não há separador entre elementos nem terminador após o
  último. Uma contagem `0` é o byte único `0x00` seguido de nenhum elemento.

Um verificador recalcula os bytes canônicos a partir dos campos recebidos e confere a
assinatura ed25519 com a chave pública obtida. Um verificador MUST rejeitar, antes de
verificar, uma assinatura cujo comprimento não seja `ed25519.SignatureSize` (64).

Os bytes canônicos de uma mensagem são **exatamente** a concatenação da tag de
domínio seguida de todos os campos listados na cláusula `Order:` daquela mensagem no
§4, nessa ordem, cada um codificado pela sua regra de primitiva acima. Não há
enquadramento externo: nenhum byte de versão no início, nenhum prefixo de comprimento
total envolvendo a mensagem e nenhum byte final após o último campo. A lista `Order:`
no §4 é exaustiva e tem autoridade — o(s) **campo(s) de assinatura** da mensagem são
os ÚNICOS campos excluídos, e não existe nos bytes canônicos nenhum campo que não
esteja nomeado ali. Para as doze mensagens centrais do §4, esse único campo excluído é
`Signature`. As mensagens de operador da Camada C (Layer-C) do §11 carregam **duas**
assinaturas (`Sig0` e `Sig1`) sobre os mesmos bytes canônicos; ambas são excluídas
exatamente como `Signature` é aqui e, assim como `Signature`, nenhuma delas aparece em
uma cláusula `Order:`. Portanto, um verificador nunca codifica nenhum campo de
assinatura nos bytes que recalcula, quer a mensagem tenha uma assinatura, quer duas.

---

## 4. Mensagens

Para cada mensagem: a sua tag de domínio, os seus campos em ordem canônica e quem a
assina. Todos os carimbos de tempo seguem a regra `time` do §3.

A cláusula `Order:` abaixo é a ÚNICA autoridade sobre a ordem canônica dos campos,
incluindo a ordem dos campos de structs aninhadas (por exemplo, `Capacity.VCPUs` antes
de `Capacity.MemMB`, e cada impressora como `Kind` e depois `Model`). Um implementador
MUST codificar os campos na sequência de `Order:` e MUST NOT inferir a ordem de
qualquer outra fonte. Em particular, o exemplo `testdata/vectors.json` (§10) lista os
`fields` de cada mensagem em **ordem alfabética de chaves para facilitar a leitura, o
que NÃO é a ordem canônica de codificação** — seguir a ordem das chaves do JSON
produziria bytes errados e assinaturas inválidas.

### 4.1 CapabilityListing — assinada pelo NÓ
Tag de domínio: `sohocloud/listing/v0`

Order: `NodeID` (string), `Class` (string), `Printers` (repeated: `Kind` string,
`Model` string), `GPUs` (repeated: `API` string, `Model` string, `VRAMMB`
int64), `Capacity.VCPUs` (int64), `Capacity.MemMB` (int64),
`Capacity.DiskMB` (int64), `Capacity.StorageCommitMB` (int64),
`Capacity.PrintQPS` (int64), `OptIn.Compute` (bool), `OptIn.Print` (bool),
`OptIn.Storage` (bool), `IssuedAt` (time), `Seq` (uint64).

`Class` ∈ {`micro`, `standard`, `server`}. `Kind` ∈ {`traditional`, `threed`}.
`API` ∈ {`vulkan`, `nnapi`, `cuda`, `metal`}.

`Capacity.DiskMB` é o espaço de rascunho (scratch) disponível para um job em execução;
`Capacity.StorageCommitMB` é o armazenamento de longa duração que o nó se compromete a
manter para a rede (hospedagem de shards). Os dois são deliberadamente distintos, para
que um nó possa oferecer um sem o outro. Como os dados armazenados são cifrados,
fragmentados em shards e auditados é uma questão do frontend/agente, fora deste
protocolo — apenas coordenação; a comunicação (wire) nunca carrega conteúdo
armazenado.

Anunciar uma GPU é a adesão (opt-in) ao trabalho de GPU: uma listagem sem `GPUs` não
recebe nenhum, e um nó retira uma GPU omitindo-a da sua próxima listagem.

`Seq` é estritamente monotônico por nó. Um coordenador MUST rejeitar uma listagem cujo
`Seq` não exceda estritamente o último visto para aquele nó (proteção contra
repetição/reversão — replay/rollback).

`OptIn` é **consultivo** (ver §5.1).

### 4.2 Heartbeat — assinada pelo NÓ
Tag de domínio: `sohocloud/heartbeat/v0`

Order: `NodeID` (string), `SentAt` (time), `Seq` (uint64).

Mesma disciplina de `Seq` estritamente monotônico do §4.1.

### 4.3 Assignment — assinada pelo COORDENADOR
Tag de domínio: `sohocloud/assignment/v0`

Order: `JobID` (string), `NodeID` (string), `Spec.Workload` (string),
`Spec.Image` (string), `Spec.Args` (repeated string), `Spec.PrinterKind`
(string), `Spec.GPUAPI` (string), `Spec.GPUMinVRAMMB` (int64),
`Fee.ContributorShareBps` (int64), `Fee.PlatformFeeBps` (int64),
`OfferedAt` (time).

`Spec.GPUAPI` ∈ {``, `vulkan`, `nnapi`, `cuda`, `metal`} — uma dica de roteamento
consultiva, como `PrinterKind`; vazio significa que nenhuma GPU é exigida, e
`GPUMinVRAMMB` MUST então ser `0`.

Os termos de taxa são anexados em linha para que o nó veja a divisão antes de se
comprometer. No modelo pull não há mensagem separada de aceite.

### 4.4 Decline — assinada pelo NÓ
Tag de domínio: `sohocloud/decline/v0`

Order: `JobID` (string), `NodeID` (string), `Reason` (string), `DeclinedAt`
(time).

`Reason` ∈ {`local_policy`, `capacity`, `unavailable`}. `local_policy` é o caminho de
recusa (opt-out) (ver §5.1).

### 4.5 JobReport — assinada pelo NÓ
Tag de domínio: `sohocloud/jobreport/v0`

Order: `JobID` (string), `NodeID` (string), `ExitCode` (int64), `FailureCause`
(string), `TmpfsExhausted` (bool), `StartedAt` (time), `FinishedAt` (time).

Este relatório assinado é o fato do qual derivam a medição e o pagamento. Os seus
campos espelham a análise de conclusão do executor já em produção no SoHoLINK.

### 4.6 FeeDeclaration — assinada pelo COORDENADOR
Tag de domínio: `sohocloud/fee/v0`

Order: `CoordinatorID` (string), `Terms.ContributorShareBps` (int64),
`Terms.PlatformFeeBps` (int64), `EffectiveAt` (time), `Seq` (uint64).

`ContributorShareBps + PlatformFeeBps` SHOULD ser igual a `10000`.

### 4.7 StorageLease — assinada pelo COORDENADOR
Tag de domínio: `sohocloud/lease/v0`

Order: `LeaseID` (string), `NodeID` (string), `ShardRef` (bytes),
`SizeClass` (int64), `Fee.ContributorShareBps` (int64),
`Fee.PlatformFeeBps` (int64), `IssuedAt` (time), `ExpiresAt` (time),
`Seq` (uint64).

Armazenamento é uma LOCAÇÃO (lease), não um job: uma obrigação contínua de manter um
shard selado por um prazo limitado. `ShardRef` é um endereço de conteúdo opaco de 32
bytes e `SizeClass` é um tamanho de carga útil quantizado — o protocolo MUST NOT
carregar tamanhos reais de objetos nem conteúdo armazenado; cifragem, preenchimento
(padding) e fragmentação em shards são questões do frontend, acima desta cintura. A
renovação é uma nova `StorageLease` para o mesmo `LeaseID` com `Seq` estritamente
maior (mesma regra de reversão do §4.1). Os termos de taxa vão em linha, como no §4.3.

### 4.8 LeaseDecline — assinada pelo NÓ
Tag de domínio: `sohocloud/lease-decline/v0`

Order: `LeaseID` (string), `NodeID` (string), `Reason` (string),
`DeclinedAt` (time).

`Reason` ∈ {`local_policy`, `capacity`, `unavailable`}; `local_policy` é o caminho de
recusa (opt-out) (§5.1) para armazenamento, exatamente como o §4.4 é para jobs.

### 4.9 LeaseRelease — assinada pelo NÓ
Tag de domínio: `sohocloud/lease-release/v0`

Order: `LeaseID` (string), `NodeID` (string), `ReleasedAt` (time).

Um nó pode sempre deixar de manter o shard (a soberania inclui sair). A liberação
assinada encerra a sua medição; recolocar o shard é uma questão do frontend.

### 4.10 ProofChallenge / ProofResponse — resposta assinada pelo NÓ
Tag de domínio (resposta): `sohocloud/proof/v0`

Um `ProofChallenge{LeaseID, Offset, Length, Nonce (16 bytes), IssuedAt}` NÃO é
assinado: os nós buscam desafios por sondagem (polling) sobre o canal autenticado
(modelo pull), e um desafio não compromete ninguém com nada. O artefato assinado é a
resposta.

Order da resposta: `LeaseID` (string), `NodeID` (string), `Offset` (int64),
`Length` (int64), `Nonce` (bytes), `Digest` (bytes), `RespondedAt` (time).

`Digest` MUST ser calculado exatamente como
`SHA-256(Nonce || uint64be(Offset) || uint64be(Length) || sealed[Offset : Offset+Length])`
sobre os bytes do shard selado. A resposta reafirma o intervalo e o nonce desafiados,
para que se sustente sozinha como um fato de medição irrefutável (não repudiável) — a
contraparte de armazenamento do §4.5. Um verificador MUST rejeitar uma resposta cujo
par `(LeaseID, Nonce)` ele já tenha aceitado: nonces são de uso único, e é isso que impede a
repetição de uma resposta gravada.

### 4.11 KeyRotation — assinada pela chave ATUAL do nó
Tag de domínio: `sohocloud/node-rotate/v0`

Order: `NodeID` (string), `NewPublicKey` (bytes), `Algo` (string),
`RotatedAt` (time), `Seq` (uint64).

A posse da chave que está saindo autoriza nomear a sua sucessora. Um verificador MUST
validar a assinatura com a chave que mantém atualmente para o nó (uma rotação assinada
pela sucessora é autoinstalação e MUST falhar), MUST impor `Seq` estritamente
monotônico por nó e MUST rejeitar uma rotação cuja `NewPublicKey` não seja uma chave
bem formada para `Algo` (no caso de `ed25519`, se não tiver exatamente 32 bytes) ou cujo `Algo`
seja desconhecido — caso contrário, um verificador que passasse a conferir as
mensagens seguintes com uma `NewPublicKey` malformada entraria em falha. Somente depois
que todas essas verificações passarem o verificador começa a verificar as mensagens
seguintes do nó com `NewPublicKey`. `Algo` ∈ {`ed25519`} na v0.

### 4.12 KeyRevocation — assinada pela chave que está sendo REVOGADA
Tag de domínio: `sohocloud/node-revoke/v0`

Order: `NodeID` (string), `RevokedPublicKey` (bytes), `RevokedAt` (time),
`Seq` (uint64).

Anula uma chave sem sucessora. Um verificador MUST verificar a assinatura com a chave
em que **confia atualmente** para `NodeID` e MUST exigir que `RevokedPublicKey` seja
igual a essa chave confiável. Uma revogação só é acatada em relação à chave exata que
ela nomeia e anula. Este é o vínculo antiabuso: um estranho pode autoassinar uma
revogação que carregue o `NodeID` de uma vítima e a própria chave do estranho em
`RevokedPublicKey`, mas, como essa chave não é aquela em que o coordenador confia para
a vítima, a verificação de igualdade falha e a revogação é rejeitada — um nó não pode
ser derrubado por quem ainda não detenha a sua chave atual. (Rascunhos anteriores
diziam "acatar incondicionalmente"; essa redação está superada — o vínculo com a chave
confiável é obrigatório.) Dada a chave atual, a leitura segura de uma revogação é
"pare de confiar nesta chave". O re-cadastramento após a revogação é feito por fora do
canal (out of band), pelo mesmo caminho do primeiro cadastramento, nunca por uma
mensagem de comunicação (wire) que um ladrão de chaves pudesse forjar.

---

## 5. Invariantes

### 5.1 A recusa (opt-out) é imposta localmente, nunca na comunicação (wire)
`CapabilityListing.OptIn` e as dicas de roteamento em `JobSpec` são **consultivos**.
O coordenador NÃO é uma fronteira de segurança para a recusa. Um nó MUST impor a
recusa com base na sua própria lista de permissões (allowlist) de confiança local e
MUST recusar o trabalho que a sua política local proíbe — via
`Decline{Reason: local_policy}` — independentemente do que tenha anunciado antes ou do
que o coordenador ofereça.

### 5.2 A política de pareamento não faz parte do protocolo
A interface `Coordinator` não expõe nenhuma classificação nem pontuação. Como um
coordenador escolhe qual nó recebe qual job é assunto privado da implementação. O
protocolo MUST NOT ganhar um campo que dite o pareamento, porque isso soldaria os
consumidores à política de um único coordenador e anularia a possibilidade de saída
(leaveability) que a interface existe para proteger.

### 5.3 As taxas são legíveis e não retroativas
Os termos de taxa de um coordenador são publicados como uma `FeeDeclaration` assinada
e com carimbo de tempo. Os termos MUST NOT mudar retroativamente para trabalho já
oferecido; uma mudança é uma nova declaração assinada com um `EffectiveAt` posterior.
Isso mantém as taxas contestáveis em vez de opacas (problema em aberto #7).

### 5.4 O coordenador é um papel, não um hub
Nada de que um nó ou frontend precise pode passar *exclusivamente* por um coordenador
específico. Um nó MAY servir a vários coordenadores; um frontend MAY executar o seu
próprio coordenador ou fazer o pareamento diretamente. Qualquer implantação em que um
único coordenador seja, de fato, inevitável é uma recentralização e MUST ser rotulada
como tal, com um caminho para alternativas.

### 5.5 Números de sequência monotônicos
Os valores de `Seq` do nó (listagens, heartbeats) são estritamente monotônicos por nó.
Um coordenador MUST rejeitar `Seq` não crescente.

---

## 6. Operações do coordenador (modelo pull)

O modelo é **pull**: o coordenador nunca chama o nó. Operações:

| Operação        | Direção          | Corpo / resultado                    |
|-----------------|------------------|--------------------------------------|
| `SubmitListing` | nó → coord       | `CapabilityListing`                  |
| `Heartbeat`     | nó → coord       | `Heartbeat`                          |
| `PollJobs`      | nó → coord       | `NodeID` → `[]Assignment`            |
| `Decline`       | nó → coord       | `Decline`                            |
| `ReportJob`     | nó → coord       | `JobReport`                          |
| `Fees`          | qualquer um → coord | → `FeeDeclaration`                |

Um coordenador MUST aplicar o vínculo do §2 a `PollJobs` (e a qualquer operação que
nomeie um nó) antes de agir. Como o coordenador nunca inicia contato com um nó, não há
serviço do lado do nó a especificar.

---

## 7. Transporte

HTTP+JSON é o transporte de **referência** (`transport/httpjson`), NÃO a comunicação
(wire) em si. Ele não é importado por nenhum pacote do núcleo. Uma implementação
conforme MAY usar qualquer transporte, implementando as operações do coordenador em
relação aos bytes canônicos do §3. JSON é apenas uma codificação de transporte; as
assinaturas nunca são calculadas sobre JSON, de modo que um transporte diferente
interopera sem renegociar assinaturas.

Os endpoints HTTP+JSON de referência:

```
POST /v0/listing     body: CapabilityListing   -> 204
POST /v0/heartbeat   body: Heartbeat           -> 204
GET  /v0/jobs?node_id=<id>                      -> 200 []Assignment
POST /v0/decline     body: Decline             -> 204
POST /v0/report      body: JobReport           -> 204
GET  /v0/fees                                   -> 200 FeeDeclaration
```

O vínculo SPIFFE (§2) é aplicado por um middleware à frente do handler; o handler de
referência apenas conecta rotas e JSON e não é o autenticador.

---

## 8. Problemas em aberto que este protocolo toca

A honestidade sobre lacunas não resolvidas é o método, não um constrangimento.

- **#7 Computação soberana compra saída mecânica, não econômica.** Esta é a camada de
  que trata o #7. O §5.4 (coordenador como papel) e o §5.3 (taxas legíveis) fazem
  pressão contra a recentralização em coordenadores; a partida a frio (cold-start)
  continua sendo um resíduo, e a portabilidade de reputação que a aliviaria é uma
  questão do frontend/da governança, não deste protocolo.
- **#6 A honestidade da computação está fora do escopo.** Se o pareamento ou o preço
  de um coordenador foram justos não é provado aqui, apenas tornado legível (§5.2,
  §5.3). `anchor/` reserva o lugar para ancorar *alegações* de emprego (os fatos do
  `JobReport`) em um log testemunhado, para que sejam ao menos auditáveis. É um stub
  rotulado, não construído.
- **#4 Associação / integração.** Como um nó se nomeia e assina por si mesmo (§2), um
  dispositivo de poucos recursos (por exemplo, um celular) pode listar a si mesmo como
  nó. Se a integração de nós é aberta ou controlada é uma decisão de governança/de
  valores deixada em aberto.

## 9. Versionamento

`v0` é instável. As tags de domínio incorporam `v0`; uma mudança incompatível nos
campos canônicos de uma mensagem incrementa a sua tag. Os consumidores MUST tratar uma
tag de domínio desconhecida como uma mensagem não verificável.

---

## 10. Vetores de teste

`testdata/vectors.json` é a **fixture normativa de conformidade** desta especificação.
Uma implementação independente é conforme quanto à codificação canônica e à
assinatura se e somente se reproduzir, byte a byte, todos os valores desse arquivo a
partir das mesmas entradas. A fixture é gerada pelos codificadores Go de referência e
conferida por uma reprodução independente que usa apenas a biblioteca padrão; uma
implementação externa conforme MUST corresponder a ela exatamente. Se a prosa do
§3/§4 e a fixture alguma vez discordarem, isso é um defeito da especificação a ser
corrigido, não uma licença para divergir — mas a fixture é a declaração executável do
§3/§4.

Todo hexadecimal na fixture está em minúsculas. Ela é autodescritiva: um `note` de
nível superior e um bloco `encoding` reafirmam as regras de primitivas do §3 para quem
abrir o JSON primeiro.

### 10.1 Estrutura

- **`primitives`** — codificações isoladas de cada primitiva do §3, para que um
  implementador possa validar o codificador de baixo para cima antes de montar
  mensagens. Cada entrada tem um `kind` (`uvarint`, `int64`, `uint64`, `bool`,
  `string`, `time`), um `input` legível por humanos e os `bytes_hex` esperados. Elas
  cobrem os casos-limite que fixam o formato: valores de varint em cada limiar de byte
  de continuação (`0, 1, 127, 128, 255, 256, 300, 16384, 1073741824`); `int64`
  incluindo `-1`, `-2`, `MaxInt64` e `MinInt64`; `uint64` incluindo `MaxUint64`;
  ambos os bools; a string vazia e uma string UTF-8 multibyte; e `time` na época Unix
  e em um instante fixo em nanossegundos.

- **`messages`** — uma entrada por tipo de mensagem assinada (todas as mensagens
  assinadas do §4). Cada entrada registra: `name`, `domain_tag`, `signer` (`node` ou
  `coordinator`), o `seed_hex` ed25519 de 32 bytes e a `public_key_hex` derivada, os
  `fields` da mensagem, o `canonical_bytes_hex` (a saída de `CanonicalBytes()`, com
  `Signature` excluído) e o `signature_hex` (ed25519 sobre exatamente esses bytes
  canônicos).

- **`operator_messages`** — uma entrada por tipo de mensagem de operador da Camada C
  (Layer-C) (todos os três do §11). Como as mensagens de operador são 2 de 7, cada
  entrada difere de uma entrada central de `messages`: `signer` é sempre `operator`;
  um array `keys` lista todas as sete chaves públicas registradas (`index`,
  `seed_hex`, `public_key_hex`, `algo` — as seeds estão presentes SOMENTE para que a
  fixture seja autocontida e reproduzível, nunca porque a rede detenha chaves
  privadas); `idx0`/`idx1` são os dois índices de assinatura; e, em vez de um único
  `signature_hex`, há `sig0_hex` e `sig1_hex`, AMBOS calculados sobre o MESMO
  `canonical_bytes_hex`. Uma implementação conforme reproduz `canonical_bytes_hex` a
  partir da ordem de campos do §11 e verifica ambas as assinaturas nos seus índices.

  > **Armadilha da ordem dos campos.** O objeto `fields` de cada mensagem lista as
  > chaves em **ordem alfabética para facilitar a leitura, NÃO em ordem canônica.** A
  > ordem canônica vem SOMENTE da cláusula `Order:` daquela mensagem no §4. Um
  > implementador que codifique os campos na ordem das chaves do JSON produzirá bytes
  > errados e assinaturas que não verificam.

### 10.2 Como usar a fixture

1. Reproduza cada `primitives[].bytes_hex` a partir do seu `input`, usando o §3.
2. Para cada mensagem, derive o par de chaves a partir de `seed_hex` (ed25519 a partir
   de uma seed de 32 bytes é determinístico), confirme que ele produz
   `public_key_hex`, codifique os campos na `Order:` do §4 e confirme que obtém
   `canonical_bytes_hex`; então verifique `signature_hex` com `public_key_hex` sobre
   os bytes que você derivou de forma independente. A verificação da assinatura sobre
   os SEUS PRÓPRIOS bytes é a verificação cruzada mais forte: ela só pode passar se os
   seus bytes forem idênticos aos do gerador de referência.

O teste de referência fica em `vectors/vectors_test.go`. Ele regenera todos os vetores
a partir do código atual e afirma igualdade byte a byte com o JSON registrado
(committed), de modo que uma divergência entre código e fixture faz o build falhar;
ele também reanalisa o arquivo registrado e verifica todas as assinaturas. Regenere
intencionalmente com `go test ./vectors -update` (ou com a variável de ambiente
`SOHOCLOUD_UPDATE_VECTORS`). Todos os vetores usam seeds fixas preenchidas byte a byte
e carimbos de tempo fixos no código — não há `crypto/rand` nem `time.Now()`, de modo
que a saída é totalmente determinística.

### 10.3 Exemplo resolvido: um `Heartbeat`, byte a byte

Este é o vetor `Heartbeat` (`§4.2`) decodificado campo a campo, para que quem for
reimplementar possa acompanhar toda a codificação do início ao fim. Entradas:
`NodeID = "node-alpha"`, `SentAt = ` nanossegundos Unix `1700000050500000000`,
`Seq = 43`.

Bytes canônicos (`canonical_bytes_hex`):

```
16 736f686f636c6f75642f6865617274626561742f7630 0a 6e6f64652d616c706861 17979d09f832d900 000000000000002b
```

Decodificado da esquerda para a direita:

| Bytes | Primitiva | Significado |
|-------|-----------|---------|
| `16` | comprimento uvarint | Comprimento da tag de domínio = `0x16` = **22** bytes. |
| `73 6f 68 6f 63 6c 6f 75 64 2f 68 65 61 72 74 62 65 61 74 2f 76 30` | bytes UTF-8 | A tag de domínio `sohocloud/heartbeat/v0` (22 bytes). Junto com o prefixo de comprimento, estas duas linhas são a codificação `string` da tag de domínio (§3). |
| `0a` | comprimento uvarint | Comprimento de `NodeID` = `0x0a` = **10** bytes. |
| `6e 6f 64 65 2d 61 6c 70 68 61` | bytes UTF-8 | `NodeID` = `node-alpha` (10 bytes). |
| `17 97 9d 09 f8 32 d9 00` | int64, big-endian | `SentAt` como nanossegundos Unix UTC. `0x17979d09f832d900` = `1700000050500000000`. Esta é a regra `time` (§3): o instante reduzido a um `int64` e emitido como 8 bytes big-endian. |
| `00 00 00 00 00 00 00 2b` | uint64, big-endian | `Seq` = `0x2b` = **43**. 8 bytes fixos, sem varint. |

A concatenação exatamente destes seis grupos — tag de domínio (comprimento + bytes),
`NodeID` (comprimento + bytes), `SentAt`, `Seq` — é a string de bytes canônica
completa. Não há enquadramento antes da tag de domínio nem bytes depois de `Seq`. A
assinatura ed25519 (`signature_hex`) é calculada precisamente sobre esta string; o
próprio campo `Signature` nunca faz parte dela.

Observe como a ordem dos campos aqui (`NodeID`, `SentAt`, `Seq`) segue a `Order:` do
§4.2, que é também — por coincidência, nesta mensagem pequena — a forma como um leitor
os listaria; para mensagens maiores, como `CapabilityListing`, a `Order:` do §4 e as
chaves alfabéticas de `fields` em `vectors.json` diferem, e somente a ordem do §4 é
canônica.

---

## 11. Identidade de operador (Camada C)

Os §1–§10 governam as mensagens de **nó** e de **coordenador**. Esta seção acrescenta
uma terceira identidade, ortogonal: o **operador** — um frontend (por exemplo, o
Cloudy) que encerra em si mesmo a identidade de membros e de nós e apresenta uma única
credencial rotativa a um coordenador. Ela é deliberadamente separada da identidade de
carga de trabalho do nó (§2): as mensagens centrais de coordenação (§4) NÃO MUDAM e
NÃO são assinadas com uma credencial de operador.

Um operador detém **sete** pares de chaves Ed25519 (índices `0..6`) e assina cada
mensagem com **dois índices distintos** (uma disciplina **2 de 7**). A rede registra
apenas as sete chaves **públicas** de cada operador; **nenhuma chave privada de
operador é jamais mantida pela rede**, e todo caminho de verificação abaixo recebe
apenas chaves públicas.

- **2 de 7 é higiene antissubstituição / de rotação, NÃO segurança por limiar
  (threshold).** Se todas as sete chaves privadas residirem em um único processo, o
  comprometimento de um único host entrega todas as sete. Não se conta com o 2 de 7
  para sobreviver ao comprometimento de uma única chave, a menos que a custódia seja
  dividida entre domínios de confiança. Isso é declarado para que nenhum leitor confie
  nele em excesso.
- **Vínculo de algoritmo.** Toda mensagem de operador vincula uma string de algoritmo
  `Algo` aos seus bytes canônicos. A `v0` permite exatamente `ed25519`. Um verificador
  MUST rejeitar uma mensagem cujo `Algo` seja desconhecido, e MUST rejeitar uma
  mensagem cujo `Algo` não seja igual ao `algo` registrado da chave em cada índice de
  assinatura. O conjunto de chaves ativo de um operador MUST NOT misturar algoritmos.
  O controle do comprimento da assinatura é por algoritmo (`ed25519` = 64 bytes), não
  uma constante fixa no código, para que a rotação atômica planejada do conjunto
  inteiro para um algoritmo pós-quântico (uma NOVA tag de domínio, por exemplo,
  `sohocloud/operator/v1-mldsa`) reutilize a mesma regra. Uma assinatura é
  intransferível entre algoritmos.
- **Ponto de migração.** A implementação de referência encaminha a assinatura e a
  verificação de operador por meio de uma interface `Signer`/`Verifier`, implementada
  hoje com o Ed25519 da biblioteca padrão, justamente para que o algoritmo possa ser
  trocado mais tarde com o acréscimo de uma tag de domínio `v1`, sem tocar nas
  mensagens centrais. Esta é uma nota de implementação; o contrato de comunicação
  (wire) são as ordens de campos e as tags abaixo.

**Codificação dos campos de operador (sem tratamento especial).** Todo campo nas
cláusulas `Order:` do §11 é codificado pela sua regra comum de primitiva do §3; o §11
não introduz tipos novos nem de largura fixa. Em particular:

- `OperatorID` e `Algo` são **strings** comuns do §3 — um varint LEB128 sem sinal com
  o comprimento em bytes UTF-8, seguido dos bytes UTF-8. `Algo` NÃO é um token simples
  nem uma tag de enum de largura fixa, apesar de a `v0` permitir exatamente o único
  valor `ed25519`; ele MUST carregar o seu prefixo de comprimento como qualquer outra
  string (por exemplo, `ed25519` é codificado como `07 65 64 32 35 35 31 39`).
  `Nonce`, `NewPublicKey` e `Challenge` são **bytes** comuns do §3 (prefixo de
  comprimento LEB128 seguido dos bytes brutos).
- `Idx0`, `Idx1` e `KeyIndex` são cada um codificados como um **`uint64` big-endian
  fixo de 8 bytes**, pela regra de `uint64` do §3, exatamente como `Seq` — NÃO como
  varint e NÃO como um único byte — mesmo que os seus valores estejam no pequeno
  intervalo `0..6`. Um implementador que codifique um índice como varint ou como um
  byte produz bytes errados e uma assinatura inválida. (Exemplo: o índice `2` é
  codificado como `00 00 00 00 00 00 00 02`.)
- As duas assinaturas `Sig0` e `Sig1` são **excluídas** dos bytes canônicos de toda
  mensagem do §11, exatamente como `Signature` é para as mensagens do §4 (§3). Elas
  são os únicos campos excluídos e, assim como `Signature`, não aparecem em nenhuma
  cláusula `Order:` abaixo. Não há um terceiro campo de assinatura. `Idx0`/`Idx1`
  (que nomeiam as chaves de assinatura) ESTÃO dentro dos bytes assinados; os *bytes*
  das assinaturas não estão.

### 11.0 Requisitos de verificação (todas as mensagens de operador)

Um verificador MUST, antes de acatar qualquer mensagem de operador:

1. Rejeitar se o nonce da mensagem estiver ausente ou for menor que **16 bytes**
   (quando a mensagem tiver nonce). Um nonce vazio MUST NEVER ser tratado como
   "pular o cache de repetição (replay)".
2. Rejeitar se `Algo` não for um algoritmo suportado nesta versão.
3. Rejeitar se os dois índices de assinatura `Idx0` e `Idx1` forem iguais.
4. Rejeitar se algum dos índices estiver fora de `0..6`.
5. Rejeitar se não existir chave registrada ativa em algum dos índices, ou se o
   `algo` de alguma das chaves registradas não for igual ao `Algo` da mensagem.
5a. Rejeitar se as chaves públicas registradas em `Idx0` e `Idx1` forem idênticas
   byte a byte, mesmo que os índices sejam diferentes. As sete chaves públicas
   registradas de um operador MUST ser distintas duas a duas; um registro ou rotação
   MUST NOT instalar uma chave pública que já exista em outro índice (ver §11.2). Isso
   é imposto para que a propriedade antissubstituição do 2 de 7 não possa se degradar
   silenciosamente para 1 de 1: se dois índices contivessem a mesma chave, uma única
   chave privada poderia produzir ambas as assinaturas. Esta verificação é sobre as
   *chaves*, não apenas sobre os índices (regra 3), e é exigida além dela.
6. Rejeitar se o comprimento de alguma das assinaturas não for o esperado para
   `Algo`.
7. Recalcular os bytes canônicos e rejeitar, a menos que AMBAS as assinaturas
   verifiquem nos seus índices sobre exatamente esses bytes.

A proteção contra repetição (anti-replay) — um conjunto durável, que falha de forma
fechada (fail-closed), de janela deslizante de `Seq` + nonces, com escopo por
`(operator, coordinator)` — é uma obrigação do lado do coordenador, sobreposta ao que
foi descrito acima; ela não faz parte dos bytes canônicos e está fora do escopo desta
especificação de codificação.

### 11.1 OperatorTransmission — assinada pelo OPERADOR (2 de 7)
Tag de domínio: `sohocloud/operator/v0`

Order: `OperatorID` (string), `Ts` (int64, nanossegundos Unix UTC — o `int64` bruto,
NÃO derivado novamente por meio de um `time.Time`, para que o valor verificado quanto
ao intervalo seja exatamente o valor assinado), `Nonce` (bytes, `len >= 16`), `Seq`
(uint64), `Algo` (string), `Idx0` (uint64), `Idx1` (uint64).

As duas assinaturas `Sig0` (pela chave em `Idx0`) e `Sig1` (pela chave em `Idx1`) são
calculadas sobre estes bytes canônicos e são EXCLUÍDAS deles, exatamente como
`Signature` é nos demais casos. `Ts`, `Nonce`, `Seq`, ambos os índices e `Algo` estão
todos dentro dos bytes assinados.

### 11.2 OperatorRotation — assinada pelo OPERADOR (2 de 7)
Tag de domínio: `sohocloud/operator-rotate/v0`

Order: `OperatorID` (string), `KeyIndex` (uint64 — o índice cuja chave está sendo
substituída), `NewPublicKey` (bytes — a chave pública substituta gerada pelo
operador), `Algo` (string), `Ts` (int64, nanossegundos Unix UTC), `Nonce` (bytes,
`len >= 16`), `Seq` (uint64), `Idx0` (uint64), `Idx1` (uint64).

Esta mensagem autoriza a troca por uma nova chave pública. O operador (nunca a rede)
gera o novo par de chaves. A nova chave pública fica DENTRO dos bytes assinados para
que nem um intermediário (man-in-the-middle) no registro por fora do canal, nem um
canal administrativo comprometido possam injetar material de chave: um coordenador
MUST verificar esta mensagem (com duas chaves ATUAIS, `Idx0`/`Idx1`) antes de
registrar a nova chave em `KeyIndex`. Um verificador MUST, adicionalmente, rejeitar
uma `NewPublicKey` cujo comprimento esteja errado para `Algo`. Um verificador MUST
também rejeitar uma `NewPublicKey` que seja idêntica byte a byte a uma chave já
registrada em um índice DIFERENTE (ela violaria o requisito de chaves distintas duas a
duas da regra 5a do §11.0 e prepararia a degradação de 2 de 7 → 1 de 1); registrar de
novo a mesma chave no seu PRÓPRIO `KeyIndex` (uma atualização sem efeito) é permitido.
Para preservar a invariante de não misturar algoritmos, um verificador SHOULD também
rejeitar uma rotação cujo `Algo` seja diferente do das demais chaves registradas.

### 11.3 ConformanceResponse — assinada pelo OPERADOR (2 de 7)
Tag de domínio: `sohocloud/operator-conformance/v0`

Order: `OperatorID` (string), `Challenge` (bytes — o desafio fornecido pelo
verificador), `Algo` (string), `Idx0` (uint64), `Idx1` (uint64).

Esta é a resposta assinada de um operador a um desafio de conformidade. A sua tag de
domínio distinta faz a separação de domínio em relação ao §11.1: como a tag faz parte
dos bytes canônicos, uma assinatura produzida para um desafio de conformidade NUNCA
pode ser verificada como uma `OperatorTransmission`, e vice-versa. Ela não carrega
nonce/`Seq` próprios — o frescor vem do `Challenge` do verificador — de modo que a
regra de nonce do §11.0 não se aplica a ela; todas as demais verificações do §11.0 se
aplicam.
