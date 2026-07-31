> Traducción comunitaria (borrador) — Política P2-002 de NTARI, Difusión Multilingüe Global. Fuente: README.md (original en inglés, instantánea del 2026-07-29). Borrador comunitario asistido por máquina, pendiente de revisión por el mantenedor regional conforme a P2-002 §3.1. Las especificaciones técnicas centrales permanecen en inglés conforme al §2.2.
>
> ¿Encontraste un error en esta traducción? Tu corrección es una contribución
> bienvenida y valorada: haz un fork del repositorio y abre un pull request en
> https://github.com/NTARI-RAND/sohocloud-protocol.

# sohocloud-protocol

El módulo de Go compartido que define el **protocolo de coordinación del sustrato** para la
red SoHoLINK / Cloudy: reconocimiento de nodos, listado de capacidades y empleo de
trabajos, hablado por cada coordinador y cada frontend.

## Alcance

Este módulo rige **únicamente la coordinación**. **No** es la economía de miembros
JFA — sin depósito en garantía (escrow), sin crédito emitido por miembros, sin pacto de
reputación, sin registro sellado por diálogo, sin adjudicación de disputas. Todo eso vive en
cada frontend y está deliberadamente ausente aquí.

Es una **hoja de dependencias**: no importa ningún otro módulo de NTARI ni nada más que
la biblioteca estándar de Go. Tanto SoHoLINK (el coordinador de referencia) como Cloudy
(un frontend) importan *este módulo*; nada se importa de vuelta a través de él. Esa propiedad
— impuesta por el grafo de importaciones, no por la prosa — es lo que impide que un
coordinador cualquiera se convierta en un concentrador (hub) por el que toda la red deba enrutarse
(problema abierto #7).

## Roles

Los frontends hablan la superficie del lado del nodo — `SubmitListing`, `Heartbeat`,
`PollJobs`, `Decline`, `ReportJob`, `Fees` — en nombre de las máquinas que aportan sus
miembros; los coordinadores implementan la interfaz `Coordinator` y
coordinan (y federan) los frontends. Las personas nunca aparecen en el cable (wire):
la identidad de los miembros es asunto del frontend, y la única identidad que este módulo
conoce es la identidad de carga de trabajo — el NodeID con su vínculo canónico SPIFFE en
`identity/`.

## Estructura

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

## Estado (honesto)

- **Construido:** los tipos de mensaje, los bytes canónicos de firma, firma/verificación
  ed25519, la interfaz `Coordinator` y un transporte de referencia HTTP+JSON bajo
  `transport/httpjson`.
- **Vectores de conformidad:** `testdata/vectors.json` es el fixture normativo de
  conformidad entre lenguajes — 25 casos primitivos más los seis tipos de mensaje
  firmados con sus bytes canónicos y firmas ed25519. Una implementación externa es
  conforme si y solo si reproduce el fixture byte por byte a partir de
  las mismas entradas (`SPEC.md` §10). La prueba de `vectors/` regenera el fixture
  a partir de los codificadores actuales y hace fallar la compilación ante cualquier desviación.
- **Transporte de referencia, no el protocolo de cable:** `transport/httpjson` es opcional y
  no es importado por ningún paquete central. Una implementación conforme PUEDE hablar estos
  mensajes sobre cualquier transporte implementando `coordinator.Coordinator` contra
  los bytes canónicos documentados en `SPEC.md`.
- **Stub, no construido:** `anchor/` — la capa de reclamos de empleo con testigos
  (problema abierto #6). Etiquetado como sustituto provisional; véase la documentación de su paquete.

## Licencia

AGPL-3.0-or-later. `SPEC.md` es parte del procomún: existe para ser leído,
reimplementado y disputado. Un procomún que nadie puede leer es una libertad que nadie puede
usar.

*Network Theory Applied Research Institute, Inc. — 501(c)(3) — EIN 92-3047136 — info@ntari.org*
