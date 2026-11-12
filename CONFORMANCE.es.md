> Traducción comunitaria (borrador) — Política P2-002 de NTARI, Difusión Multilingüe Global. Fuente: CONFORMANCE.md (original en inglés, instantánea del 2026-10-05). Borrador comunitario asistido por máquina, pendiente de revisión por mantenedores regionales conforme a P2-002 §3.1. Las especificaciones técnicas centrales permanecen en inglés conforme a §2.2.
>
> ¿Encontraste un error en esta traducción? Tu corrección es una contribución
> bienvenida y valorada: haz un fork del repositorio y abre un pull request en
> https://github.com/NTARI-RAND/sohocloud-protocol.

> **El original en inglés es normativo.** Esta traducción se ofrece para lectura, no para interpretación: donde difiera de `CONFORMANCE.md`, rige el inglés. Las palabras clave normativas (MUST, MUST NOT, SHOULD, SHOULD NOT, MAY), los nombres de campos, las etiquetas y los identificadores se mantienen en inglés porque constituyen el contrato de conformidad.

# Conformidad — Janus-Facing Architecture

Este documento es la autodescripción del módulo en los propios términos de la arquitectura, expuesta **antes** de cualquier cosa específica del producto o del despliegue, según la regla de orden de la arquitectura. Sigue la disciplina de legibilidad de la arquitectura: una afirmación de conformidad está vinculada al mecanismo y a la verificación que la hacen cumplir, o se etiqueta como suplencia. La prosa no vinculada es marketing.

La arquitectura es **Janus-Facing Architecture (JFA)**: el documento de arquitectura unificada de NTARI, documentación libre bajo el bien común AGPL-3.0 del proyecto. Nombra roles y mecanismos, nunca productos; este módulo declara qué rol ocupa y muestra sus vínculos.

## Declaración de rol

Este módulo es el **protocolo de coordinación** de un **sustrato** de JFA: el lenguaje de transmisión, hoja de dependencias, mediante el cual las máquinas propiedad de los participantes son reconocidas, anuncian su capacidad, y toman e informan trabajo. La garantía del sustrato a la que sirve: **coordinación sobre infraestructura que los participantes pueden poseer — sin un punto de estrangulamiento de alojamiento imposible de eliminar.**

| Término de este módulo | Rol en la arquitectura |
|---|---|
| Node | infraestructura propiedad de los participantes — las máquinas que aportan los miembros |
| Coordinator | el coordinador: orquestación del lado de los nodos, un rol intercambiable |
| Frontend | un frontend: la aplicación de cara a los miembros |
| FeeDeclaration | la declaración de tarifas del coordinador, de su autoría, legible e impugnable |
| JobReport | una afirmación de trabajo (anclaje atestiguado reservado — véanse las suplencias) |

## Invariantes y sus vínculos

| Invariante (arquitectura) | Mecanismo aquí | Verificación |
|---|---|---|
| El protocolo de coordinación es una **hoja de dependencias** | `go.mod` declara cero requisitos; solo biblioteca estándar | CI: `.github/workflows/dependency-leaf.yml` hace fallar la compilación ante cualquier dependencia añadida |
| **Las personas nunca aparecen en la transmisión** | La única identidad que este módulo conoce es la identidad de carga de trabajo — el NodeID de `identity/` con su vinculación SPIFFE canónica; ningún mensaje lleva un nombre, un contacto ni un campo con forma de persona | SPEC (normativo); la superficie de tipos; vectores de referencia |
| **Sin punto de estrangulamiento de alojamiento imposible de eliminar** | El coordinador es un rol intercambiable: un frontend MAY ejecutar su propio coordinador o emparejar directamente con los nodos (SPEC §1) | Texto normativo de SPEC; existen dos consumidores contra la etiqueta publicada |
| **Legibilidad de las tarifas** | Las tarifas existen en la transmisión solo como una `FeeDeclaration` firmada por el coordinador — de su autoría, legible, impugnable | `fees/`; vectores de referencia |
| **La legibilidad es un producto, no un comentario** | `SPEC.md` es normativo y un entregable de la compilación: un implementador externo construye a partir de él sin leer este código fuente | `testdata/vectors.json` (verdad de referencia) + `testdata/reproduce_spec.py` (reproductor independiente) + prueba de referencia de `vectors/` — cualquier divergencia entre el código y los vectores confirmados hace fallar la compilación |
| **Codificación canónica y determinista** | Codificador de firma de `canon/` con prefijo de longitud y separación por dominio | vectores de referencia regenerados y comparados byte a byte en cada ejecución de pruebas |

## Deliberadamente fuera de alcance

Este módulo **no** es la economía de miembros, y esa ausencia es un invariante, no una carencia: sin depósito de garantía (escrow), sin crédito emitido por los miembros, sin pacto de reputación, sin registro sellado por diálogo, sin adjudicación de disputas. Esas funciones pertenecen a las capas superiores (registro, pacto, economía), viven en los frontends y MUST NOT migrar a esta hoja. La identidad única de participante es igualmente una obligación de los frontends — esta capa no conoce personas en absoluto.

## Suplencias y residuos abiertos

Nombrados según la regla de honestidad de la arquitectura: un ejemplo práctico que ocultara dónde se queda corto solo enseñaría a ocultar.

- **`anchor/` es un stub, etiquetado como tal en la documentación de su paquete.** El anclaje atestiguado y de solo adición de las afirmaciones de empleo (trabajo que entra, resultado que sale, medición) está reservado, no construido. Hasta que exista, el relato de un coordinador sobre el trabajo completado es autoatestiguado — el problema abierto 6 de la arquitectura, acoplado al problema abierto 2.
- **v0 es UNSTABLE.** Las formas de transmisión pueden cambiar sin garantías de compatibilidad hasta v1; los consumidores fijan etiquetas publicadas.
- **El cómputo soberano compra una salida mecánica, no económica** (problema abierto 7). Este protocolo limita cuán concentrado puede llegar a estar el alojamiento; no puede financiar la salida de nadie. Nombrado, no resuelto.

## Declaración de dependencias

Depende de: la biblioteca estándar de Go. Nada más — ningún otro módulo de NTARI, ningún módulo de terceros, nada por encima ni al lado de él en la pila. La verificación de CI convierte una infracción en una compilación fallida en lugar de algo que se detecta en una revisión.

## Notas específicas del producto (al final, según la regla de orden)

SoHoLINK es el coordinador de referencia; Cloudy es un frontend de referencia. Ambos consumen este módulo mediante una etiqueta de versión publicada. Ninguno de los dos es privilegiado por el protocolo: cualquier implementación conforme puede reemplazar a cualquiera de ellos.
