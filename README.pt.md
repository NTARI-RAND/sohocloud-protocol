# sohocloud-protocol

> Tradução comunitária (rascunho) — política P2-002 da NTARI, Transmissão Multilíngue Global. Fonte: README.md (original em inglês, snapshot de 2026-07-29). Rascunho comunitário assistido por máquina, pendente de revisão por mantenedor regional conforme P2-002 §3.1. As especificações técnicas centrais permanecem em inglês conforme §2.2.
>
> Notou algum erro nesta tradução? Correções de tradução são contribuições
> valiosas e muito bem-vindas: faça um fork do repositório e abra um pull
> request em https://github.com/NTARI-RAND/sohocloud-protocol.

O módulo Go compartilhado que define o **protocolo de coordenação do substrato**
da rede SoHoLINK / Cloudy: reconhecimento de nós, listagem de capacidades e
emprego de jobs, falado por todos os coordenadores e todos os frontends.

## Escopo

Este módulo governa **apenas a coordenação**. Ele **não é** a economia de
membros da JFA — sem escrow, sem crédito emitido por membros, sem pacto de
reputação, sem registro selado por diálogo, sem julgamento de disputas. Essas
coisas ficam em cada frontend e estão deliberadamente ausentes aqui.

Ele é uma **folha de dependência**: não importa nenhum outro módulo da NTARI e
nada além da biblioteca padrão do Go. Tanto o SoHoLINK (o coordenador de
referência) quanto o Cloudy (um frontend) importam *este* módulo; nada é
importado de volta através dele. Essa propriedade — imposta pelo grafo de
importação, e não por prosa — é o que impede que um único coordenador se torne
um hub pelo qual toda a rede tenha de trafegar (problema aberto #7).

## Papéis

Os frontends falam a superfície do lado do nó — `SubmitListing`, `Heartbeat`,
`PollJobs`, `Decline`, `ReportJob`, `Fees` — em nome das máquinas que seus
membros contribuem; os coordenadores implementam a interface `Coordinator` e
coordenam (e federam) os frontends. Pessoas nunca aparecem no tráfego da rede:
a identidade do membro é uma questão do frontend, e a única identidade que
este módulo conhece é a identidade de carga de trabalho (workload) — o NodeID
com sua vinculação SPIFFE canônica em `identity/`.

## Estrutura

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

## Status (honesto)

- **Implementado:** tipos de mensagem, bytes canônicos de assinatura,
  assinatura/verificação ed25519, a interface `Coordinator` e um transporte
  HTTP+JSON de referência em `transport/httpjson`.
- **Vetores de conformidade:** `testdata/vectors.json` é a fixture normativa
  de conformidade entre linguagens — 25 casos primitivos mais todos os seis
  tipos de mensagem assinada com seus bytes canônicos e assinaturas ed25519.
  Uma implementação externa é conformante se e somente se reproduzir a fixture
  byte a byte a partir das mesmas entradas (`SPEC.md` §10). O teste em
  `vectors/` regenera a fixture a partir dos codificadores atuais e faz o build
  falhar diante de qualquer divergência.
- **Transporte de referência, não o protocolo de rede:** `transport/httpjson` é
  opcional e não é importado por nenhum pacote do núcleo. Uma implementação
  conformante PODE falar essas mensagens sobre qualquer transporte,
  implementando `coordinator.Coordinator` em relação aos bytes canônicos
  documentados em `SPEC.md`.
- **Stub, não implementado:** `anchor/` — a camada testemunhada de
  reivindicação de emprego (problema aberto #6). Rotulada como substituto
  provisório; veja a documentação do pacote.

## Licença

AGPL-3.0-or-later. `SPEC.md` faz parte do commons: ele existe para ser lido,
reimplementado e contestado. Um commons que ninguém pode ler é uma liberdade
que ninguém pode usar.

*Network Theory Applied Research Institute, Inc. — 501(c)(3) — EIN 92-3047136 — info@ntari.org*
