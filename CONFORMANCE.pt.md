> Tradução comunitária (rascunho) — Política NTARI P2-002, Transmissão Multilíngue Global (Global Multilingual Broadcast). Fonte: CONFORMANCE.md (original em inglês, snapshot de 2026-10-05). Rascunho comunitário assistido por máquina, pendente de revisão por mantenedor regional conforme P2-002 §3.1. As especificações técnicas centrais permanecem em inglês conforme §2.2.
>
> Notou algum erro nesta tradução? Correções de tradução são contribuições
> valiosas e muito bem-vindas: faça um fork do repositório e abra um pull
> request em https://github.com/NTARI-RAND/sohocloud-protocol.

> **O original em inglês é normativo.** Esta tradução é oferecida para leitura, não para interpretação: onde ela divergir de `CONFORMANCE.md`, prevalece o inglês. As palavras-chave normativas (MUST, MUST NOT, SHOULD, SHOULD NOT, MAY), os nomes de campos, os rótulos e os identificadores são mantidos em inglês porque constituem o contrato de conformidade.

# Conformidade — Janus-Facing Architecture

Este documento é a autodescrição do módulo nos próprios termos da arquitetura, apresentada **antes** de qualquer coisa específica de produto ou de implantação, conforme a regra de ordenação da arquitetura. Ele segue a disciplina de legibilidade da arquitetura: uma alegação de conformidade está vinculada ao mecanismo e à verificação que a fazem cumprir, ou então recebe um rótulo provisório (stand-in). Prosa não vinculada é marketing.

A arquitetura é a **Janus-Facing Architecture (JFA)** (Arquitetura de Dupla Face) — o documento de arquitetura unificada da NTARI, documentação livre sob o commons AGPL-3.0 do projeto. Ela nomeia papéis e mecanismos, nunca produtos; este módulo declara qual papel preenche e mostra seus vínculos.

## Declaração de papel

Este módulo é o **protocolo de coordenação** de um **substrato** da JFA: a linguagem de comunicação (wire), folha de dependência, pela qual máquinas de propriedade dos participantes são reconhecidas, anunciam capacidade e assumem e relatam trabalho. A garantia de substrato a que ele serve: **coordenação sobre infraestrutura que os participantes podem possuir — nenhum ponto de estrangulamento de hospedagem impossível de remover.**

| Termo deste módulo | Papel na arquitetura |
|---|---|
| Node | infraestrutura de propriedade dos participantes — as máquinas com que os membros contribuem |
| Coordinator | o coordenador: orquestração do lado dos nós, um papel substituível (pluggable) |
| Frontend | um frontend: o aplicativo voltado aos membros |
| FeeDeclaration | a declaração de taxas do coordenador, de sua autoria, legível e contestável |
| JobReport | uma alegação de trabalho (ancoragem testemunhada reservada — ver rótulos provisórios) |

## Invariantes e seus vínculos

| Invariante (arquitetura) | Mecanismo aqui | Verificação |
|---|---|---|
| O protocolo de coordenação é uma **folha de dependência** | `go.mod` declara zero requisitos; somente a biblioteca padrão | CI: `.github/workflows/dependency-leaf.yml` faz o build falhar diante de qualquer dependência adicionada |
| **Pessoas nunca aparecem na comunicação (wire)** | A única identidade que este módulo conhece é a identidade de carga de trabalho — o NodeID de `identity/` com seu vínculo SPIFFE canônico; nenhuma mensagem carrega nome, contato ou campo com formato de pessoa | SPEC (normativa); a superfície de tipos; vetores de referência (golden vectors) |
| **Nenhum ponto de estrangulamento de hospedagem impossível de remover** | O coordenador é um papel substituível: um frontend MAY executar seu próprio coordenador ou fazer o pareamento diretamente com os nós (SPEC §1) | Texto normativo da SPEC; existem dois consumidores usando a tag publicada |
| **Legibilidade das taxas** | Taxas existem na comunicação apenas como uma `FeeDeclaration` assinada pelo coordenador — de sua autoria, legível, contestável | `fees/`; vetores de referência |
| **Legibilidade é um produto, não um comentário** | `SPEC.md` é normativo e um entregável do build: um implementador externo constrói a partir dele sem ler este código-fonte | `testdata/vectors.json` (verdade de referência) + `testdata/reproduce_spec.py` (reprodutor independente) + teste de referência em `vectors/` — divergência entre o código e os vetores registrados faz o build falhar |
| **Codificação canônica e determinística** | Codificador de assinatura de `canon/`, com prefixo de comprimento e separação por domínio | vetores de referência regenerados e comparados byte a byte a cada execução de testes |

## Deliberadamente fora do escopo

Este módulo **não** é a economia de membros, e essa ausência é uma invariante, não uma lacuna: sem custódia (escrow), sem crédito emitido por membros, sem pacto de reputação, sem registro selado por diálogo, sem julgamento de disputas. Essas coisas pertencem às camadas acima (registro, pacto, economia), ficam nos frontends e MUST NOT migrar para esta folha. A identidade única por participante é igualmente uma obrigação do frontend — esta camada não conhece pessoa alguma.

## Rótulos provisórios e resíduos em aberto

Nomeados conforme a regra de honestidade da arquitetura: um exemplo prático que ocultasse onde fica aquém só ensinaria a ocultação.

- **`anchor/` é um stub, rotulado como tal na documentação do seu pacote.** A ancoragem testemunhada e somente de adição (append-only) das alegações de emprego (job que entra, resultado que sai, medição) está reservada, não construída. Até que exista, o relato de um coordenador sobre o trabalho concluído é autodeclarado — o problema em aberto 6 da arquitetura, acoplado ao problema em aberto 2.
- **v0 é UNSTABLE.** Os formatos de comunicação podem mudar sem garantias de compatibilidade até a v1; os consumidores fixam tags publicadas.
- **Computação soberana compra saída mecânica, não econômica** (problema em aberto 7). Este protocolo limita o quanto a hospedagem pode se concentrar; não pode financiar a saída de ninguém. Nomeado, não resolvido.

## Declaração de dependências

Depende de: a biblioteca padrão do Go. Nada mais — nenhum outro módulo da NTARI, nenhum módulo de terceiros, nada acima ou ao lado dele na pilha. A verificação de CI faz de uma violação um build com falha, em vez de algo detectado em revisão.

## Notas específicas de produto (por último, conforme a regra de ordenação)

O SoHoLINK é o coordenador de referência; o Cloudy é um frontend de referência. Ambos consomem este módulo por meio de uma tag de versão publicada. Nenhum dos dois é privilegiado pelo protocolo: qualquer implementação conforme pode substituir qualquer um deles.
