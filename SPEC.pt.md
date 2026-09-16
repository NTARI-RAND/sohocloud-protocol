> Tradução comunitária (rascunho) — Política P2-002 da NTARI, Difusão Multilíngue Global. Fonte: SPEC.md (original em inglês, instantâneo de 2026-09-16). Rascunho comunitário assistido por máquina, pendente de revisão pelo mantenedor regional conforme P2-002 §3.1. As especificações técnicas centrais permanecem em inglês conforme o §2.2.
>
> **O original em inglês é normativo.** Esta tradução destina-se à leitura, não à
> interpretação: onde divergir de `SPEC.md`, prevalece o inglês. Permanecem em
> inglês, por constituírem o contrato de conformidade, as palavras-chave
> normativas (MUST, MUST NOT, SHOULD, MAY), os nomes de campos, as etiquetas de
> domínio, os identificadores de tipos e todos os blocos de código.
>
> Notou algum erro nesta tradução? Correções de tradução são contribuições
> valiosas e muito bem-vindas: faça um fork do repositório e abra um pull
> request em https://github.com/NTARI-RAND/sohocloud-protocol.

# sohocloud-protocol — Especificação de conformidade

Versão: **v0 (INSTÁVEL)** — os formatos das mensagens e o formato canônico de
bytes podem mudar sem garantias de compatibilidade até a v1.

Este documento é normativo e é um entregável de construção, não um acessório. O
direito da AGPL de bifurcar a rede é vazio se uma implementação não puder ser
construída a partir de uma especificação legível sem ler o código-fonte deste
módulo. Tudo que abaixo estiver marcado MUST / MUST NOT é um requisito de
conformidade; uma mudança que viole um deles não é uma versão menor deste
protocolo, é um protocolo diferente.

---

## 1. Escopo

Este protocolo rege a **camada de coordenação do substrato**: como um nó é
reconhecido, como ele anuncia que trabalho aceitará, e como o trabalho lhe é
oferecido e reportado de volta. Três partes o falam:

- **Node** — um dispositivo de propriedade de um participante que oferece
  computação e/ou impressão física. Um nó é dono de sua identidade e assina tudo
  que emite. Soberano: escolhe quais coordenadores servir e pode deixar qualquer
  um deles.
- **Coordinator** — casa anúncios com trabalho e registra resultados. Um papel
  plugável. SoHoLINK é o coordenador de referência; um frontend MAY executar o seu
  próprio ou casar diretamente com nós.
- **Frontend** — onde os membros transacionam. Consome coordenação pela mesma
  interface.

**Fora de escopo, deliberadamente.** Este protocolo NÃO é a economia de membros do
JFA. Ele não define escrow, nem crédito emitido por membros, nem pacto de
reputação, nem registro selado por diálogo, nem julgamento de disputas. Isso
pertence a cada frontend. A liquidação em moeda fiduciária que um coordenador
opera com seus nós (p. ex. os repasses via Stripe do SoHoLINK) é a economia do
*substrato* e também fica fora desta especificação.

---

## 2. Identidade e vinculação

Um nó é nomeado por um `NodeID` — uma cadeia estável que o próprio nó escolhe.

A identidade de carga de trabalho de um nó é um SPIFFE SVID emitido exatamente
sob:

```
/node/<NodeID>
```

Um coordenador autoriza uma requisição como vinda de um nó comparando o caminho
SPIFFE autenticado do chamador com `/node/<NodeID>` para o `NodeID` nomeado na
requisição.

- Um coordenador MUST rejeitar com **401** quando não houver identidade SPIFFE
  presente.
- Um coordenador MUST rejeitar com **403** quando a identidade estiver presente
  mas seu caminho não for exatamente `/node/<NodeID>`.
- Um coordenador MUST NOT aceitar "qualquer SVID válido" como autorização para uma
  requisição que nomeie um nó diferente. Isso é um vetor de personificação.
- A vinculação MUST ser construção determinística de cadeias, nunca uma busca por
  nó em armazenamento mutável. Uma verificação condicionada a um campo que nunca é
  preenchido é inerte e MUST NOT ser usada como fronteira de segurança.

O protocolo não distribui chaves públicas. Um coordenador resolve a chave de
verificação de um nó fora de banda (registro, material do SVID ou um cadastro).

---

## 3. Codificação canônica de bytes (para assinaturas)

Toda mensagem assinada é assinada sobre seus **bytes canônicos**, computados com o
campo `Signature` EXCLUÍDO. As assinaturas são **ed25519**.

A assinatura NÃO é feita sobre JSON nem sobre qualquer codificação de transporte.
É feita sobre o seguinte formato determinístico com prefixo de comprimento, de
modo que as assinaturas sejam estáveis entre plataformas, versões de Go e
linguagens.

Primitivas, na ordem exata em que os campos são listados para cada mensagem no §4:

- **unsigned LEB128 varint** — a primitiva de comprimento/contagem usada por
  `string`, `bytes` e `repeated`. Ela MUST ser codificada como segue,
  independentemente da linguagem: tome-se o valor sem sinal; emitam-se 7 bits por
  byte, primeiro o grupo menos significativo (grupos little-endian); ponha-se o
  bit alto (`0x80`) de cada byte EXCETO o último para marcar continuação; o último
  byte tem seu bit alto zerado. O valor `0` codifica-se como o byte único `0x00`.
  Esta é exatamente a codificação produzida por `encoding/binary.PutUvarint` do
  Go; o símbolo Go é uma verificação cruzada, não a definição. (Exemplo
  trabalhado: `300` = `0b100101100` → 7 bits baixos `0101100` com continuação →
  `0xAC`, bits altos `0000010` por último → `0x02`, resultando `AC 02`.)
- **domain tag** — os bytes canônicos de toda mensagem começam com uma cadeia de
  etiqueta de domínio distinta (ver §4). A etiqueta de domínio MUST ser codificada
  pela regra de **string** abaixo — isto é, um unsigned LEB128 varint do seu
  comprimento em bytes UTF-8 seguido dos seus bytes UTF-8 — exatamente como
  qualquer outro campo de cadeia, sem terminador nulo e sem tratamento especial.
  Etiquetas distintas tornam uma assinatura sobre um tipo de mensagem
  intransferível para outro.
- **string** — unsigned LEB128 varint com o comprimento em bytes, depois os bytes
  UTF-8 crus. Uma cadeia vazia MUST codificar-se como o byte único `0x00`
  (comprimento zero e nenhum conteúdo).
- **bytes** — unsigned LEB128 varint com o comprimento em bytes, depois os bytes
  crus. (O campo `Signature` nunca é codificado; ele é excluído.)
- **bool** — um único byte, exatamente `0x00` (falso) ou `0x01` (verdadeiro);
  nenhum outro valor de byte é válido.
- **int64 / uint64** — 8 bytes fixos, big-endian (byte mais significativo
  primeiro). `int64` MUST ser codificado como sua reinterpretação `uint64` em
  complemento de dois: some-se 2⁶⁴ a um valor negativo e emitam-se os 8 bytes
  big-endian. Assim `-1` → `ff ff ff ff ff ff ff ff`, `-2` →
  `ff ff ff ff ff ff ff fe`, e o `int64` mínimo (`-9223372036854775808`) →
  `80 00 00 00 00 00 00 00`. Sem sinal-magnitude, sem zigzag, sem comprimento
  variável.
- **time** — o instante como `int64` de **nanossegundos** Unix UTC (pela regra de
  int64 acima, big-endian, 8 bytes). Localização e componentes de relógio monótono
  são descartados. A precisão de nanossegundos é preservada exatamente; o valor
  MUST NOT ser truncado a micro- ou milissegundos. Isso sobrevive a uma ida e
  volta por transporte JSON e independe do fuso horário. Um instante fora da faixa
  representável de nanossegundos Unix UTC em `int64` (aproximadamente os anos
  1678–2262) está fora de escopo para a v0 e MUST ser tratado como erro de
  codificação em vez de envolvido ou limitado. O codificador de referência impõe
  isso: `canon.Buffer.Time` não acrescenta nada e trava
  `canon.ErrTimeOutOfRange` (recuperável via `Buffer.Err()`) em vez de emitir um
  valor envolvido, de modo que dois instantes separados por ~584 anos nunca possam
  coincidir em bytes idênticos. Um produtor MUST rejeitar via `canon.ValidTime`
  antes de assinar; um verificador/decodificador MUST rejeitar uma mensagem cuja
  marca de tempo falhe nessa checagem.
- **repeated field** — um unsigned LEB128 varint com a *contagem* (o número de
  elementos) codificado PRIMEIRO, depois os campos de cada elemento codificados em
  ordem, totalmente em linha, usando as regras primitivas normais acima. Um
  elemento repetido que seja ele próprio uma cadeia leva, portanto, prefixo de
  comprimento como qualquer outra cadeia; campos de estruturas aninhadas são
  emitidos na ordem dada para aquele elemento no §4. Não há separador entre
  elementos nem terminador após o último. Uma contagem de `0` é o byte único
  `0x00` seguido de nenhum elemento.

Um verificador recomputa os bytes canônicos a partir dos campos recebidos e checa
a assinatura ed25519 contra a chave pública resolvida. Um verificador MUST
rejeitar uma assinatura cujo comprimento não seja `ed25519.SignatureSize` (64)
antes de verificar.

Os bytes canônicos de uma mensagem são **exatamente** a concatenação da etiqueta
de domínio seguida de cada campo listado na cláusula `Order:` daquela mensagem no
§4, nessa ordem, cada um codificado por sua regra primitiva. Não há enquadramento
externo: nenhum byte de versão inicial, nenhum prefixo de comprimento total
envolvendo a mensagem e nenhum byte final após o último campo. A lista `Order:` do
§4 é exaustiva e autoritativa — os **campos de assinatura** da mensagem são os
ÚNICOS campos excluídos, e não existe nos bytes canônicos nenhum campo que não
esteja nomeado ali. Para as doze mensagens centrais do §4, esse único campo
excluído é `Signature`. As mensagens de operador da Camada C do §11 carregam
**duas** assinaturas (`Sig0` e `Sig1`) sobre os mesmos bytes canônicos; ambas são
excluídas exatamente como `Signature` aqui e, como `Signature`, nenhuma aparece em
uma cláusula `Order:`. Portanto, um verificador nunca codifica nenhum campo de
assinatura nos bytes que recomputa, tenha a mensagem uma assinatura ou duas.

---

## 4. Mensagens

Para cada mensagem: sua etiqueta de domínio, seus campos em ordem canônica e quem
a assina. Todas as marcas de tempo seguem a regra `time` do §3.

A cláusula `Order:` abaixo é a ÚNICA autoridade sobre a ordem canônica dos campos,
inclusive a ordem de campos de estruturas aninhadas (p. ex. `Capacity.VCPUs` antes
de `Capacity.MemMB`, e cada impressora como `Kind` e depois `Model`). Quem
implementar MUST codificar os campos na sequência `Order:` e MUST NOT inferir a
ordem de qualquer outra fonte. Em particular, o exemplo `testdata/vectors.json`
(§10) lista os `fields` de cada mensagem em **ordem alfabética de chaves para
legibilidade, o que NÃO é a ordem canônica de codificação** — seguir a ordem das
chaves do JSON produziria bytes errados e assinaturas inválidas.

### 4.1 CapabilityListing — assinada pelo NODE
Domain tag: `sohocloud/listing/v0`

Order: `NodeID` (string), `Class` (string), `Printers` (repeated: `Kind` string,
`Model` string), `GPUs` (repeated: `API` string, `Model` string, `VRAMMB`
int64), `Capacity.VCPUs` (int64), `Capacity.MemMB` (int64),
`Capacity.DiskMB` (int64), `Capacity.StorageCommitMB` (int64),
`Capacity.PrintQPS` (int64), `OptIn.Compute` (bool), `OptIn.Print` (bool),
`OptIn.Storage` (bool), `IssuedAt` (time), `Seq` (uint64).

`Class` ∈ {`micro`, `standard`, `server`}. `Kind` ∈ {`traditional`, `threed`}.
`API` ∈ {`vulkan`, `nnapi`, `cuda`, `metal`}.

`Capacity.DiskMB` é espaço de rascunho disponível a um trabalho em execução;
`Capacity.StorageCommitMB` é armazenamento de longa duração que o nó se compromete
a manter para a rede (hospedagem de fragmentos). Os dois são deliberadamente
distintos para que um nó possa oferecer um sem o outro. Como os dados armazenados
são cifrados, fragmentados e auditados é assunto do frontend/agente, fora deste
protocolo — apenas coordenação; o fio nunca transporta conteúdo armazenado.

Anunciar uma GPU é a adesão ao trabalho com GPU: um anúncio sem `GPUs` não recebe
nenhum, e um nó retira uma GPU omitindo-a de seu anúncio seguinte.

`Seq` é estritamente monotônico por nó. Um coordenador MUST rejeitar um anúncio
cujo `Seq` não exceda estritamente o último visto para aquele nó (proteção contra
repetição/retrocesso).

`OptIn` é **consultivo** (ver §5.1).

### 4.2 Heartbeat — assinado pelo NODE
Domain tag: `sohocloud/heartbeat/v0`

Order: `NodeID` (string), `SentAt` (time), `Seq` (uint64).

Mesma disciplina de `Seq` estritamente monotônico do §4.1.

### 4.3 Assignment — assinada pelo COORDINATOR
Domain tag: `sohocloud/assignment/v0`

Order: `JobID` (string), `NodeID` (string), `Spec.Workload` (string),
`Spec.Image` (string), `Spec.Args` (repeated string), `Spec.PrinterKind`
(string), `Spec.GPUAPI` (string), `Spec.GPUMinVRAMMB` (int64),
`Fee.ContributorShareBps` (int64), `Fee.PlatformFeeBps` (int64),
`OfferedAt` (time).

`Spec.GPUAPI` ∈ {``, `vulkan`, `nnapi`, `cuda`, `metal`} — uma dica consultiva de
roteamento como `PrinterKind`; vazio significa que nenhuma GPU é requerida e
`GPUMinVRAMMB` MUST então ser `0`.

Os termos de taxa são anexados em linha para que um nó veja a divisão antes de se
comprometer. No modelo de extração não há mensagem de aceitação separada.

### 4.4 Decline — assinada pelo NODE
Domain tag: `sohocloud/decline/v0`

Order: `JobID` (string), `NodeID` (string), `Reason` (string), `DeclinedAt`
(time).

`Reason` ∈ {`local_policy`, `capacity`, `unavailable`}. `local_policy` é a via de
recusa (ver §5.1).

### 4.5 JobReport — assinado pelo NODE
Domain tag: `sohocloud/jobreport/v0`

Order: `JobID` (string), `NodeID` (string), `ExitCode` (int64), `FailureCause`
(string), `TmpfsExhausted` (bool), `StartedAt` (time), `FinishedAt` (time).

Este relatório assinado é o fato do qual derivam a medição e o pagamento. Seus
campos espelham a análise de conclusão do executor já em uso no SoHoLINK.

### 4.6 FeeDeclaration — assinada pelo COORDINATOR
Domain tag: `sohocloud/fee/v0`

Order: `CoordinatorID` (string), `Terms.ContributorShareBps` (int64),
`Terms.PlatformFeeBps` (int64), `EffectiveAt` (time), `Seq` (uint64).

`ContributorShareBps + PlatformFeeBps` SHOULD ser igual a `10000`.

### 4.7 StorageLease — assinado pelo COORDINATOR
Domain tag: `sohocloud/lease/v0`

Order: `LeaseID` (string), `NodeID` (string), `ShardRef` (bytes),
`SizeClass` (int64), `Fee.ContributorShareBps` (int64),
`Fee.PlatformFeeBps` (int64), `IssuedAt` (time), `ExpiresAt` (time),
`Seq` (uint64).

Armazenamento é um ARRENDAMENTO, não um trabalho: uma obrigação contínua de manter
um fragmento selado por um prazo limitado. `ShardRef` é um endereço de conteúdo
opaco de 32 bytes e `SizeClass` um tamanho de carga quantizado — o protocolo MUST
NOT transportar tamanhos reais de objetos nem conteúdo armazenado; cifragem,
preenchimento e fragmentação são assuntos do frontend acima desta cintura. A
renovação é um novo `StorageLease` para o mesmo `LeaseID` com `Seq` estritamente
maior (mesma regra de retrocesso do §4.1). Os termos de taxa viajam em linha, como
no §4.3.

### 4.8 LeaseDecline — assinada pelo NODE
Domain tag: `sohocloud/lease-decline/v0`

Order: `LeaseID` (string), `NodeID` (string), `Reason` (string),
`DeclinedAt` (time).

`Reason` ∈ {`local_policy`, `capacity`, `unavailable`}; `local_policy` é a via de
recusa (§5.1) para armazenamento exatamente como o §4.4 o é para trabalhos.

### 4.9 LeaseRelease — assinada pelo NODE
Domain tag: `sohocloud/lease-release/v0`

Order: `LeaseID` (string), `NodeID` (string), `ReleasedAt` (time).

Um nó pode sempre parar de manter (soberania inclui partir). A liberação assinada
encerra sua medição; recolocar o fragmento é assunto do frontend.

### 4.10 ProofChallenge / ProofResponse — resposta assinada pelo NODE
Domain tag (da resposta): `sohocloud/proof/v0`

Um `ProofChallenge{LeaseID, Offset, Length, Nonce (16 bytes), IssuedAt}` NÃO é
assinado: os nós buscam desafios sondando pelo canal autenticado (modelo de
extração), e um desafio não compromete ninguém a nada. O artefato assinado é a
resposta.

Response Order: `LeaseID` (string), `NodeID` (string), `Offset` (int64),
`Length` (int64), `Nonce` (bytes), `Digest` (bytes), `RespondedAt` (time).

`Digest` MUST ser computado exatamente como
`SHA-256(Nonce || uint64be(Offset) || uint64be(Length) || sealed[Offset : Offset+Length])`
sobre os bytes do fragmento selado. A resposta reafirma a faixa desafiada e o
nonce para que se sustente sozinha como fato de medição não repudiável — a
contraparte de armazenamento do §4.5. Um verificador MUST rejeitar uma resposta
cujo par `(LeaseID, Nonce)` já tenha aceitado: nonces são de uso único, e é isso
que derrota a repetição de uma resposta gravada.

### 4.11 KeyRotation — assinada pela chave ATUAL do nó
Domain tag: `sohocloud/node-rotate/v0`

Order: `NodeID` (string), `NewPublicKey` (bytes), `Algo` (string),
`RotatedAt` (time), `Seq` (uint64).

A posse da chave que sai autoriza nomear sua sucessora. Um verificador MUST
validar a assinatura contra a chave que atualmente detém para o nó (uma rotação
assinada pela sucessora é autoinstalação e MUST falhar), MUST impor `Seq`
estritamente monotônico por nó, e MUST rejeitar uma rotação cuja `NewPublicKey`
não seja uma chave bem-formada para `Algo` (para `ed25519`, não exatamente 32
bytes) ou cujo `Algo` seja desconhecido — do contrário, um verificador que
começasse a checar mensagens subsequentes contra uma `NewPublicKey` malformada
falharia. Só depois de todos esses passos o verificador começa a verificar as
mensagens subsequentes do nó contra `NewPublicKey`. `Algo` ∈ {`ed25519`} na v0.

### 4.12 KeyRevocation — assinada pela chave que está sendo REVOGADA
Domain tag: `sohocloud/node-revoke/v0`

Order: `NodeID` (string), `RevokedPublicKey` (bytes), `RevokedAt` (time),
`Seq` (uint64).

Mata uma chave sem sucessora. Um verificador MUST verificar a assinatura contra a
chave em que **atualmente confia** para `NodeID`, e MUST exigir que
`RevokedPublicKey` seja igual a essa chave confiável. Uma revogação só é honrada
contra a chave exata que ela nomeia e mata. Esta é a vinculação antiabuso: um
estranho pode autoassinar uma revogação portando o `NodeID` de uma vítima e a
própria chave do estranho em `RevokedPublicKey`, mas como essa chave não é a que o
coordenador considera confiável para a vítima, a checagem de igualdade falha e a
revogação é rejeitada — um nó não pode ser derrubado por ninguém que já não
detenha sua chave atual. (Rascunhos anteriores diziam "honrar
incondicionalmente"; essa redação está superada — a vinculação à chave confiável
é obrigatória.) Dada a chave atual, a leitura segura de uma revogação é "pare de
confiar nesta chave". A reinscrição após revogação é fora de banda, pela mesma via
da inscrição inicial, nunca uma mensagem no fio que um ladrão de chaves pudesse
forjar.

---

## 5. Invariantes

### 5.1 A recusa é imposta localmente, nunca no fio
`CapabilityListing.OptIn` e as dicas de roteamento em `JobSpec` são
**consultivas**. O coordenador NÃO é uma fronteira de segurança para a recusa. Um
nó MUST impor a recusa contra sua própria lista de permissões de confiança local e
MUST recusar o trabalho que sua política local proíba — via
`Decline{Reason: local_policy}` — independentemente do que tenha anunciado antes
ou do que o coordenador ofereça.

### 5.2 A política de casamento não faz parte do protocolo
A interface `Coordinator` não expõe classificação nem pontuação. Como um
coordenador escolhe qual nó recebe qual trabalho é privado da implementação. O
protocolo MUST NOT ganhar um campo que dite o casamento, porque isso soldaria os
consumidores à política de um único coordenador e derrotaria a possibilidade de
partir que a interface existe para proteger.

### 5.3 As taxas são legíveis e não retroativas
Os termos de taxa de um coordenador são publicados como uma `FeeDeclaration`
assinada e com marca de tempo. Os termos MUST NOT mudar retroativamente para
trabalho já oferecido; uma mudança é uma nova declaração assinada com
`EffectiveAt` posterior. Isso mantém as taxas contestáveis em vez de opacas
(problema aberto #7).

### 5.4 O coordenador é um papel, não um concentrador
Nada de que um nó ou um frontend precise pode passar *exclusivamente* por um
coordenador específico. Um nó MAY servir múltiplos coordenadores; um frontend MAY
executar seu próprio coordenador ou casar diretamente. Qualquer implantação em que
um único coordenador seja de fato inevitável é uma recentralização e MUST ser
rotulada como tal, com um caminho para alternativas.

### 5.5 Números de sequência monotônicos
Os valores de `Seq` do nó (anúncios, batimentos) são estritamente monotônicos por
nó. Um coordenador MUST rejeitar `Seq` não crescente.

---

## 6. Operações do coordenador (modelo de extração)

O modelo é de **extração**: o coordenador nunca chama o nó. Operações:

| Operação        | Direção          | Corpo / resultado                    |
|-----------------|------------------|--------------------------------------|
| `SubmitListing` | nó → coord       | `CapabilityListing`                  |
| `Heartbeat`     | nó → coord       | `Heartbeat`                          |
| `PollJobs`      | nó → coord       | `NodeID` → `[]Assignment`            |
| `Decline`       | nó → coord       | `Decline`                            |
| `ReportJob`     | nó → coord       | `JobReport`                          |
| `Fees`          | qualquer um → coord | → `FeeDeclaration`                |

Um coordenador MUST aplicar a vinculação do §2 a `PollJobs` (e a qualquer operação
que nomeie um nó) antes de agir. Como o coordenador nunca inicia contato com um
nó, não há serviço do lado do nó a especificar.

---

## 7. Transporte

HTTP+JSON é o transporte de **referência** (`transport/httpjson`), NÃO o fio em
si. Nenhum pacote central o importa. Uma implementação conforme MAY usar qualquer
transporte implementando as operações do coordenador contra os bytes canônicos do
§3. JSON é apenas uma codificação de transporte; assinaturas nunca são computadas
sobre JSON, de modo que um transporte diferente interopera sem renegociar
assinaturas.

Os endpoints HTTP+JSON de referência:

```
POST /v0/listing     body: CapabilityListing   -> 204
POST /v0/heartbeat   body: Heartbeat           -> 204
GET  /v0/jobs?node_id=<id>                      -> 200 []Assignment
POST /v0/decline     body: Decline             -> 204
POST /v0/report      body: JobReport           -> 204
GET  /v0/fees                                   -> 200 FeeDeclaration
```

A vinculação SPIFFE (§2) é aplicada por middleware à frente do manipulador; o
manipulador de referência apenas conecta rotas e JSON e não é o autenticador.

---

## 8. Problemas abertos que este protocolo toca

Honestidade sobre lacunas não resolvidas é o método, não um constrangimento.

- **#7 A computação soberana compra saída mecânica, não econômica.** Esta é a
  camada de que trata o #7. O §5.4 (coordenador como papel) e o §5.3 (taxas
  legíveis) empurram contra a recentralização do coordenador; a partida a frio
  permanece um resíduo, e a portabilidade de reputação que a aliviaria é assunto
  de frontend/governança, não deste protocolo.
- **#6 A honestidade da computação está fora de escopo.** Se o casamento ou o
  preço de um coordenador foram justos não se prova aqui, apenas se torna legível
  (§5.2, §5.3). `anchor/` reserva o lugar para ancorar *alegações* de emprego (os
  fatos do `JobReport`) em um registro testemunhado, para que sejam ao menos
  auditáveis. É um esboço rotulado, não construído.
- **#4 Adesão / integração.** Como um nó se nomeia e assina por si (§2), um
  dispositivo de poucos recursos (p. ex. um telefone) pode se listar como nó. Se a
  integração de nós é aberta ou restrita é uma decisão de governança/valores
  deixada em aberto.

## 9. Versionamento

`v0` é instável. As etiquetas de domínio embutem `v0`; uma mudança incompatível
nos campos canônicos de uma mensagem incrementa sua etiqueta. Os consumidores MUST
tratar uma etiqueta de domínio desconhecida como mensagem não verificável.

---

## 10. Vetores de teste

`testdata/vectors.json` é o **fixture normativo de conformidade** desta
especificação. Uma implementação independente é conforme quanto à codificação
canônica e à assinatura se e somente se reproduzir, byte a byte, cada valor daquele
arquivo a partir das mesmas entradas. O fixture é gerado pelos codificadores Go de
referência e é conferido de forma cruzada por uma reprodução independente que usa
apenas a biblioteca padrão; uma implementação externa conforme MUST corresponder a
ele exatamente. Onde a prosa dos §3/§4 e o fixture discordarem, isso é um defeito
da especificação a ser corrigido, não uma licença para divergir — mas o fixture é
o enunciado executável dos §3/§4.

Todo o hexadecimal do fixture é minúsculo. Ele é autodescritivo: um `note` de
nível superior e um bloco `encoding` reapresentam as regras primitivas do §3 para
quem abrir o JSON primeiro.

### 10.1 Estrutura

- **`primitives`** — codificações isoladas de cada primitiva do §3, para que quem
  implementa possa validar o codificador de baixo para cima antes de montar
  mensagens. Cada entrada tem um `kind` (`uvarint`, `int64`, `uint64`, `bool`,
  `string`, `time`), um `input` legível e os `bytes_hex` esperados. Elas cobrem os
  casos-limite que fixam o formato: valores varint em cada limiar de byte de
  continuação (`0, 1, 127, 128, 255, 256, 300, 16384, 1073741824`); `int64`
  incluindo `-1`, `-2`, `MaxInt64` e `MinInt64`; `uint64` incluindo `MaxUint64`;
  ambos os booleanos; a cadeia vazia e uma cadeia UTF-8 multibyte; e `time` na
  época Unix e em um instante fixo de nanossegundos.

- **`messages`** — uma entrada por tipo de mensagem assinada (todas as mensagens
  assinadas do §4). Cada entrada registra: `name`, `domain_tag`, `signer` (`node`
  ou `coordinator`), o `seed_hex` ed25519 de 32 bytes e o `public_key_hex`
  derivado, os `fields` da mensagem, os `canonical_bytes_hex` (a saída de
  `CanonicalBytes()`, com `Signature` excluído) e o `signature_hex` (ed25519 sobre
  exatamente esses bytes canônicos).

- **`operator_messages`** — uma entrada por tipo de mensagem de operador da Camada
  C (as três do §11). Como as mensagens de operador são 2-de-7, cada entrada
  difere de uma entrada central de `messages`: `signer` é sempre `operator`; um
  arranjo `keys` lista todas as sete chaves públicas registradas (`index`,
  `seed_hex`, `public_key_hex`, `algo` — as sementes estão presentes APENAS para
  que o fixture seja autocontido e reproduzível, nunca porque a rede detenha
  chaves privadas); `idx0`/`idx1` são os dois índices assinantes; e, em vez de um
  `signature_hex`, há `sig0_hex` e `sig1_hex`, AMBOS tomados sobre os MESMOS
  `canonical_bytes_hex`. Uma implementação conforme reproduz
  `canonical_bytes_hex` a partir da ordem de campos do §11 e verifica ambas as
  assinaturas em seus índices.

  > **Armadilha da ordem dos campos.** O objeto `fields` de cada mensagem lista as
  > chaves em **ordem alfabética para legibilidade, NÃO em ordem canônica.** A
  > ordem canônica vem SOMENTE da cláusula `Order:` daquela mensagem no §4. Quem
  > codificar os campos na ordem das chaves do JSON produzirá bytes errados e
  > assinaturas que não verificam.

### 10.2 Como usar o fixture

1. Reproduza cada `primitives[].bytes_hex` a partir do seu `input` usando o §3.
2. Para cada mensagem, derive o par de chaves a partir de `seed_hex` (ed25519 a
   partir de uma semente de 32 bytes é determinístico), confirme que ele produz
   `public_key_hex`, codifique os campos na `Order:` do §4 e confirme que você
   obtém `canonical_bytes_hex`; então verifique `signature_hex` contra
   `public_key_hex` sobre os seus bytes derivados independentemente. A verificação
   da assinatura sobre os SEUS PRÓPRIOS bytes é a conferência cruzada mais forte:
   ela só pode passar se os seus bytes forem idênticos aos do gerador de
   referência.

O teste de referência vive em `vectors/vectors_test.go`. Ele regenera cada vetor a
partir do código atual e afirma igualdade byte a byte com o JSON versionado, de
modo que um desvio entre código e fixture quebra a construção; ele também
reanalisa o arquivo versionado e verifica todas as assinaturas. Regenere
intencionalmente com `go test ./vectors -update` (ou a variável de ambiente
`SOHOCLOUD_UPDATE_VECTORS`). Todos os vetores usam sementes de preenchimento de
bytes fixas e marcas de tempo codificadas — não há `crypto/rand` nem `time.Now()`,
então a saída é plenamente determinística.

### 10.3 Exemplo trabalhado: um `Heartbeat`, byte a byte

Este é o vetor `Heartbeat` (`§4.2`) decodificado campo a campo, para que quem
reimplementa possa traçar toda a codificação do início ao fim. Entradas:
`NodeID = "node-alpha"`, `SentAt = ` nanossegundos Unix `1700000050500000000`,
`Seq = 43`.

Bytes canônicos (`canonical_bytes_hex`):

```
16 736f686f636c6f75642f6865617274626561742f7630 0a 6e6f64652d616c706861 17979d09f832d900 000000000000002b
```

Decodificados da esquerda para a direita:

| Bytes | Primitiva | Significado |
|-------|-----------|-------------|
| `16` | comprimento uvarint | Comprimento da etiqueta de domínio = `0x16` = **22** bytes. |
| `73 6f 68 6f 63 6c 6f 75 64 2f 68 65 61 72 74 62 65 61 74 2f 76 30` | bytes UTF-8 | A etiqueta de domínio `sohocloud/heartbeat/v0` (22 bytes). Junto com o prefixo de comprimento, estas duas linhas são a codificação `string` da etiqueta de domínio (§3). |
| `0a` | comprimento uvarint | Comprimento de `NodeID` = `0x0a` = **10** bytes. |
| `6e 6f 64 65 2d 61 6c 70 68 61` | bytes UTF-8 | `NodeID` = `node-alpha` (10 bytes). |
| `17 97 9d 09 f8 32 d9 00` | int64, big-endian | `SentAt` como nanossegundos Unix UTC. `0x17979d09f832d900` = `1700000050500000000`. Esta é a regra `time` (§3): o instante reduzido a um `int64` e emitido como 8 bytes big-endian. |
| `00 00 00 00 00 00 00 2b` | uint64, big-endian | `Seq` = `0x2b` = **43**. 8 bytes fixos, sem varint. |

A concatenação exatamente destes seis grupos — etiqueta de domínio (comprimento +
bytes), `NodeID` (comprimento + bytes), `SentAt`, `Seq` — é a cadeia canônica de
bytes completa. Não há enquadramento antes da etiqueta de domínio nem bytes depois
de `Seq`. A assinatura ed25519 (`signature_hex`) é tomada precisamente sobre essa
cadeia; o próprio campo `Signature` nunca faz parte dela.

Note que a ordem de campos aqui (`NodeID`, `SentAt`, `Seq`) segue a `Order:` do
§4.2, que também é — por coincidência nesta mensagem pequena — como um leitor os
listaria; para mensagens maiores como `CapabilityListing`, a `Order:` do §4 e as
chaves alfabéticas de `fields` em `vectors.json` diferem, e só a ordem do §4 é
canônica.

---

## 11. Identidade de operador (Camada C)

Os §1–§10 regem as mensagens de **node** e **coordinator**. Esta seção acrescenta
uma terceira identidade, ortogonal: o **operator** — um frontend (p. ex. Cloudy)
que termina dentro de si a identidade de membros e nós e apresenta uma única
credencial rotativa a um coordenador. Ela é deliberadamente separada da identidade
de carga de trabalho do nó (§2): as mensagens centrais de coordenação (§4) estão
INALTERADAS e NÃO são assinadas com credencial de operador.

Um operador detém **sete** pares de chaves Ed25519 (índices `0..6`) e assina cada
mensagem com **dois índices distintos** (uma disciplina **2-de-7**). A rede
registra apenas as sete chaves **públicas** por operador; **nenhuma chave privada
de operador é jamais detida pela rede**, e todo caminho de verificação abaixo toma
somente chaves públicas.

- **2-de-7 é antissubstituição / higiene de rotação, NÃO segurança de limiar.** Se
  as sete chaves privadas vivem em um mesmo processo, um único comprometimento do
  host entrega todas as sete. Não se confia em 2-de-7 para sobreviver ao
  comprometimento de uma única chave, a menos que a custódia esteja repartida
  entre domínios de confiança. Isso é declarado para que nenhum leitor confie
  demais.
- **Vinculação de algoritmo.** Toda mensagem de operador vincula uma cadeia de
  algoritmo `Algo` em seus bytes canônicos. A `v0` permite exatamente `ed25519`.
  Um verificador MUST rejeitar uma mensagem cujo `Algo` seja desconhecido, e MUST
  rejeitar uma mensagem cujo `Algo` não seja igual ao `algo` registrado da chave em
  cada índice assinante. O conjunto de chaves ativas de um operador MUST NOT
  misturar algoritmos. A barreira de comprimento de assinatura é por algoritmo
  (`ed25519` = 64 bytes), não uma constante fixa, de modo que a planejada rotação
  atômica de todo o conjunto para um algoritmo pós-quântico (uma NOVA etiqueta de
  domínio, p. ex. `sohocloud/operator/v1-mldsa`) reutiliza a mesma regra. Uma
  assinatura não é transferível entre algoritmos.
- **Costura de migração.** A implementação de referência encaminha a assinatura e
  a verificação de operador por uma interface `Signer`/`Verifier`, implementada
  hoje com Ed25519 da biblioteca padrão, precisamente para que o algoritmo possa
  ser trocado adicionando depois uma etiqueta de domínio `v1` sem tocar nas
  mensagens centrais. Isto é uma nota de implementação; o contrato do fio são as
  ordens de campos e as etiquetas abaixo.

**Codificação dos campos de operador (sem tratamento especial).** Todo campo das
cláusulas `Order:` do §11 é codificado pela sua regra primitiva ordinária do §3; o
§11 não introduz tipos novos nem de largura fixa. Em particular:

- `OperatorID` e `Algo` são **strings** ordinárias do §3 — um unsigned LEB128
  varint do comprimento em bytes UTF-8 seguido dos bytes UTF-8. `Algo` NÃO é um
  token nu nem uma etiqueta de enumeração de largura fixa, apesar de a `v0`
  permitir exatamente o único valor `ed25519`; ele MUST levar seu prefixo de
  comprimento como qualquer outra cadeia (p. ex. `ed25519` codifica-se como
  `07 65 64 32 35 35 31 39`). `Nonce`, `NewPublicKey` e `Challenge` são **bytes**
  ordinários do §3 (prefixo de comprimento LEB128 e então os bytes crus).
- `Idx0`, `Idx1` e `KeyIndex` são cada um codificados como um **`uint64`
  big-endian de 8 bytes fixos** pela regra `uint64` do §3, exatamente como `Seq` —
  NÃO como varint e NÃO como um único byte — ainda que seus valores fiquem na
  pequena faixa `0..6`. Quem codificar um índice como varint ou como um byte
  produz bytes errados e assinatura inválida. (Exemplo: o índice `2` codifica-se
  como `00 00 00 00 00 00 00 02`.)
- As duas assinaturas `Sig0` e `Sig1` são **excluídas** dos bytes canônicos de toda
  mensagem do §11, exatamente como `Signature` o é para as mensagens do §4 (§3).
  São os únicos campos excluídos e, como `Signature`, não aparecem em nenhuma
  cláusula `Order:` abaixo. Não há um terceiro campo de assinatura. `Idx0`/`Idx1`
  (que nomeiam as chaves assinantes) ESTÃO dentro dos bytes assinados; os *bytes*
  da assinatura, não.

### 11.0 Requisitos de verificação (todas as mensagens de operador)

Um verificador MUST, antes de honrar qualquer mensagem de operador:

1. Rejeitar se o nonce da mensagem estiver ausente ou for menor que **16 bytes**
   (onde a mensagem tiver nonce). Um nonce vazio MUST NUNCA ser tratado como
   "pular o cache de repetição".
2. Rejeitar se `Algo` não for um algoritmo suportado para esta versão.
3. Rejeitar se os dois índices assinantes `Idx0` e `Idx1` forem iguais.
4. Rejeitar se qualquer índice estiver fora de `0..6`.
5. Rejeitar se não existir chave registrada ativa em algum dos índices, ou se o
   `algo` de alguma chave registrada não for igual ao `Algo` da mensagem.
5a. Rejeitar se as chaves públicas registradas em `Idx0` e `Idx1` forem idênticas
   byte a byte, ainda que os índices difiram. As sete chaves públicas registradas
   de um operador MUST ser distintas duas a duas; um registro ou rotação MUST NOT
   instalar uma chave pública que já exista em outro índice (ver §11.2). Isso é
   imposto para que a propriedade antissubstituição 2-de-7 não possa degradar-se
   silenciosamente para 1-de-1: se dois índices tivessem a mesma chave, uma única
   chave privada poderia produzir ambas as assinaturas. Esta checagem é sobre as
   *chaves*, não meramente sobre os índices (regra 3), e é exigida além dela.
6. Rejeitar se o comprimento de alguma assinatura não for o esperado para `Algo`.
7. Recomputar os bytes canônicos e rejeitar a menos que AMBAS as assinaturas
   verifiquem em seus índices sobre exatamente esses bytes.

A antirrepetição (um conjunto durável e de fechamento seguro de `Seq` em janela
deslizante mais nonces, com escopo por `(operator, coordinator)`) é uma obrigação
do lado do coordenador sobreposta AO ACIMA; não faz parte dos bytes canônicos e
está fora do escopo desta especificação de codificação.

### 11.1 OperatorTransmission — assinada pelo OPERATOR (2 de 7)
Domain tag: `sohocloud/operator/v0`

Order: `OperatorID` (string), `Ts` (int64 nanossegundos Unix UTC — o `int64` cru,
NÃO rederivado através de um `time.Time`, de modo que o valor com faixa verificada
seja exatamente o valor assinado), `Nonce` (bytes, `len >= 16`), `Seq` (uint64),
`Algo` (string), `Idx0` (uint64), `Idx1` (uint64).

As duas assinaturas `Sig0` (pela chave em `Idx0`) e `Sig1` (pela chave em `Idx1`)
são tomadas sobre estes bytes canônicos e são EXCLUÍDAS deles, exatamente como
`Signature` em outros pontos. `Ts`, `Nonce`, `Seq`, ambos os índices e `Algo`
estão todos dentro dos bytes assinados.

### 11.2 OperatorRotation — assinada pelo OPERATOR (2 de 7)
Domain tag: `sohocloud/operator-rotate/v0`

Order: `OperatorID` (string), `KeyIndex` (uint64 — o índice cuja chave está sendo
substituída), `NewPublicKey` (bytes — a chave pública de substituição gerada pelo
operador), `Algo` (string), `Ts` (int64 nanossegundos Unix UTC), `Nonce` (bytes,
`len >= 16`), `Seq` (uint64), `Idx0` (uint64), `Idx1` (uint64).

Isto autoriza a inserção de uma nova chave pública. O operador (nunca a rede) gera
o novo par de chaves. A nova chave pública está DENTRO dos bytes assinados para
que nem um intermediário no registro fora de banda nem um canal administrativo
comprometido possam injetar material de chave: um coordenador MUST verificar esta
mensagem (com duas chaves ATUAIS, `Idx0`/`Idx1`) antes de gravar a nova chave em
`KeyIndex`. Um verificador MUST adicionalmente rejeitar uma `NewPublicKey` cujo
comprimento esteja errado para `Algo`. Um verificador MUST também rejeitar uma
`NewPublicKey` idêntica byte a byte a uma chave já registrada em um índice
DIFERENTE (violaria o requisito de chaves distintas duas a duas da regra 5a do
§11.0 e prepararia a degradação 2-de-7 → 1-de-1); reregistrar a mesma chave em seu
PRÓPRIO `KeyIndex` (uma atualização sem efeito) é permitido. Para preservar a
invariante de não misturar algoritmos, um verificador SHOULD também rejeitar uma
rotação cujo `Algo` difira do das demais chaves registradas.

### 11.3 ConformanceResponse — assinada pelo OPERATOR (2 de 7)
Domain tag: `sohocloud/operator-conformance/v0`

Order: `OperatorID` (string), `Challenge` (bytes — o desafio fornecido pelo
verificador), `Algo` (string), `Idx0` (uint64), `Idx1` (uint64).

Esta é a resposta assinada de um operador a um desafio de conformidade. Sua
etiqueta de domínio distinta a separa por domínio do §11.1: como a etiqueta faz
parte dos bytes canônicos, uma assinatura produzida para um desafio de
conformidade NUNCA pode verificar como uma `OperatorTransmission`, e vice-versa.
Ela não carrega nonce nem `Seq` próprios — o frescor vem do `Challenge` do
verificador — de modo que a regra de nonce do §11.0 não se aplica a ela; todas as
demais checagens do §11.0, sim.
