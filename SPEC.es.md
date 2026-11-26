> Traducción comunitaria (borrador) — Política P2-002 de NTARI, Difusión Multilingüe Global. Fuente: SPEC.md (original en inglés, instantánea del 2026-10-05). Borrador comunitario asistido por máquina, pendiente de revisión por mantenedores regionales conforme a P2-002 §3.1. Las especificaciones técnicas centrales permanecen en inglés conforme a §2.2.
>
> ¿Encontraste un error en esta traducción? Tu corrección es una contribución
> bienvenida y valorada: haz un fork del repositorio y abre un pull request en
> https://github.com/NTARI-RAND/sohocloud-protocol.

> **Precedencia.** El original en inglés es normativo. Esta traducción se ofrece
> para su lectura, no para su interpretación: donde difiera de `SPEC.md`, rige el
> inglés. Las palabras clave normativas (MUST, MUST NOT, SHOULD, SHOULD NOT, MAY),
> los nombres de campo, las etiquetas y los identificadores se mantienen en inglés
> porque constituyen el contrato de conformidad.

# sohocloud-protocol — Especificación de conformidad

Versión: **v0 (INESTABLE)** — las estructuras de los mensajes y el formato canónico
de bytes pueden cambiar sin garantías de compatibilidad hasta la v1.

Este documento es normativo y es un entregable de la construcción, no una idea de
último momento. El derecho que otorga la AGPL a bifurcar (fork) la red está vacío
a menos que pueda construirse una implementación a partir de una especificación
legible, sin leer el código fuente de este módulo. Todo lo que aparece a
continuación marcado como MUST / MUST NOT es un requisito de conformidad; un
cambio que viole uno de ellos no es una versión reducida de este protocolo, es un
protocolo diferente.

---

## 1. Alcance

Este protocolo rige la **capa de coordinación del sustrato**: cómo se reconoce un
nodo, cómo anuncia qué trabajo realizará y cómo se le ofrece trabajo y cómo este
se reporta de vuelta. Lo hablan tres partes:

- **Nodo** — un dispositivo propiedad de un participante que ofrece cómputo y/o
  impresión física. Un nodo es dueño de su identidad y firma todo lo que emite.
  Soberano: elige a qué coordinadores servir y puede abandonar cualquiera de ellos.
- **Coordinador** — empareja anuncios con trabajo y registra los resultados. Es un
  rol intercambiable. SoHoLINK es el coordinador de referencia; un frontend MAY
  ejecutar el suyo propio o emparejar directamente contra los nodos.
- **Frontend** — donde los miembros realizan sus transacciones. Consume la
  coordinación a través de la misma interfaz.

**Fuera de alcance, deliberadamente.** Este protocolo NO es la economía de miembros
de JFA. No define custodia (escrow), ni crédito emitido por los miembros, ni pacto
de reputación, ni registro sellado por diálogo, ni adjudicación de disputas. Todo
eso corresponde a cada frontend. La liquidación en moneda fiduciaria que un
coordinador realiza con sus nodos (p. ej., los pagos de SoHoLINK a través de
Stripe) es la economía del *sustrato* y queda igualmente fuera de esta
especificación.

---

## 2. Identidad y vinculación

Un nodo se nombra mediante un `NodeID` — una cadena estable que el nodo elige para
sí mismo.

La identidad de carga de trabajo de un nodo es un SVID de SPIFFE emitido
exactamente bajo:

```
/node/<NodeID>
```

Un coordinador autoriza una solicitud como proveniente de un nodo comparando la
ruta SPIFFE autenticada de quien llama con `/node/<NodeID>` para el `NodeID`
nombrado en la solicitud.

- Un coordinador MUST rechazar con **401** cuando no haya ninguna identidad SPIFFE
  presente.
- Un coordinador MUST rechazar con **403** cuando la identidad esté presente pero
  su ruta no sea exactamente `/node/<NodeID>`.
- Un coordinador MUST NOT aceptar "cualquier SVID válido" como autorización para
  una solicitud que nombre a un nodo diferente. Eso es un vector de suplantación.
- La vinculación MUST ser una construcción determinista de cadenas, nunca una
  búsqueda por nodo en un almacenamiento mutable. Una comprobación condicionada a
  un campo que nunca se rellena es inerte y MUST NOT usarse como frontera de
  seguridad.

El protocolo no distribuye claves públicas. Un coordinador resuelve la clave de
verificación de un nodo por un canal externo (registro, material del SVID o un
directorio de registro).

---

## 3. Codificación canónica de bytes (para firmas)

Cada mensaje firmado se firma sobre sus **bytes canónicos**, calculados con el
campo `Signature` EXCLUIDO. Las firmas son **ed25519**.

La firma NO se realiza sobre JSON ni sobre ninguna codificación de transporte. Se
realiza sobre el siguiente formato determinista con prefijo de longitud, de modo
que las firmas sean estables entre plataformas, versiones de Go y lenguajes.

Primitivas, en el orden exacto en que se enumeran los campos de cada mensaje en §4:

- **varint LEB128 sin signo** — la primitiva de longitud/conteo que usan `string`,
  `bytes` y `repeated`. MUST codificarse de la siguiente manera, con independencia
  del lenguaje: tomar el valor sin signo; emitir 7 bits por byte, empezando por el
  grupo menos significativo (grupos en little-endian); activar el bit alto (`0x80`)
  de cada byte EXCEPTO el último para señalar continuación; el último byte tiene su
  bit alto desactivado. El valor `0` se codifica como el byte único `0x00`. Esta es
  exactamente la codificación que produce `encoding/binary.PutUvarint` de Go; el
  símbolo de Go es una verificación cruzada, no la definición. (Ejemplo resuelto:
  `300` = `0b100101100` → 7 bits bajos `0101100` con continuación → `0xAC`, bits
  altos `0000010` al final → `0x02`, lo que da `AC 02`.)
- **etiqueta de dominio** — los bytes canónicos de cada mensaje comienzan con una
  cadena de etiqueta de dominio distinta (véase §4). La etiqueta de dominio MUST
  codificarse con la regla de **string** que aparece más abajo — es decir, un
  varint LEB128 sin signo con su longitud en bytes UTF-8 seguido de sus bytes
  UTF-8 — exactamente igual que cualquier otro campo de cadena, sin terminador nulo
  y sin casos especiales. Las etiquetas distintas hacen que una firma sobre un tipo
  de mensaje no sea transferible a otro.
- **string** — longitud en bytes como varint LEB128 sin signo, seguida de los bytes
  UTF-8 sin procesar. Una cadena vacía MUST codificarse como el byte único `0x00`
  (una longitud cero y sin carga útil).
- **bytes** — longitud en bytes como varint LEB128 sin signo, seguida de los bytes
  sin procesar. (El campo `Signature` nunca se codifica; queda excluido.)
- **bool** — un único byte, exactamente `0x00` (falso) o `0x01` (verdadero);
  ningún otro valor de byte es válido.
- **int64 / uint64** — 8 bytes fijos, big-endian (el byte más significativo
  primero). `int64` MUST codificarse como su reinterpretación `uint64` en
  complemento a dos: sumar 2⁶⁴ a un valor negativo y luego emitir los 8 bytes en
  big-endian. Así, `-1` → `ff ff ff ff ff ff ff ff`, `-2` →
  `ff ff ff ff ff ff ff fe`, y el `int64` mínimo (`-9223372036854775808`) →
  `80 00 00 00 00 00 00 00`. Sin signo-magnitud, sin zigzag, sin longitud variable.
- **time** — el instante como `int64` de **nanosegundos** Unix UTC (según la regla
  de int64 anterior, big-endian, 8 bytes). Los componentes de ubicación y de reloj
  monotónico se descartan. La precisión de nanosegundos se conserva exactamente; el
  valor MUST NOT truncarse a micro- ni milisegundos. Esto sobrevive a un viaje de
  ida y vuelta por un transporte JSON y es independiente de la zona horaria. Un
  instante fuera del rango representable de `int64` en nanosegundos Unix UTC
  (aproximadamente los años 1678–2262) queda fuera del alcance de v0 y MUST
  tratarse como un error de codificación en lugar de envolverse o recortarse. El
  codificador de referencia lo hace cumplir: `canon.Buffer.Time` no añade nada y
  retiene `canon.ErrTimeOutOfRange` (recuperable mediante `Buffer.Err()`) en lugar
  de emitir un valor envuelto, de modo que dos instantes separados por ~584 años
  nunca puedan coincidir en bytes idénticos. Un productor MUST rechazar mediante
  `canon.ValidTime` antes de firmar; un verificador/decodificador MUST rechazar un
  mensaje cuya marca de tiempo no la supere.
- **campo repetido** — un *conteo* como varint LEB128 sin signo (el número de
  elementos) codificado PRIMERO, y luego los campos de cada elemento codificados en
  orden, completamente en línea, usando las reglas de primitivas normales
  anteriores. Un elemento repetido que sea en sí una cadena lleva, por tanto,
  prefijo de longitud como cualquier otra cadena; los campos de estructuras
  anidadas se emiten en el orden indicado para ese elemento en §4. No hay separador
  entre elementos ni terminador después del último. Un conteo de `0` es el byte
  único `0x00` sin elementos a continuación.

Un verificador recalcula los bytes canónicos a partir de los campos recibidos y
comprueba la firma ed25519 contra la clave pública resuelta. Un verificador MUST
rechazar una firma cuya longitud no sea `ed25519.SignatureSize` (64) antes de
verificar.

Los bytes canónicos de un mensaje son **exactamente** la concatenación de la
etiqueta de dominio seguida de cada campo enumerado en la cláusula `Order:` de ese
mensaje en §4, en ese orden, cada uno codificado según su regla de primitiva
anterior. No hay ningún encuadre exterior: ni byte de versión inicial, ni prefijo
de longitud total que envuelva el mensaje, ni bytes finales después del último
campo. La lista `Order:` de §4 es exhaustiva y autoritativa — el o los **campos de
firma** del mensaje son los ÚNICOS campos excluidos, y no existe en los bytes
canónicos ningún campo que no esté nombrado allí. Para los doce mensajes centrales
de §4, ese único campo excluido es `Signature`. Los mensajes de operador de la
Capa C de §11 llevan **dos** firmas (`Sig0` y `Sig1`) sobre los mismos bytes
canónicos; ambas se excluyen exactamente como se excluye `Signature` aquí y, al
igual que `Signature`, ninguna aparece en una cláusula `Order:`. Por lo tanto, un
verificador nunca codifica ningún campo de firma en los bytes que recalcula, tenga
el mensaje una firma o dos.

---

## 4. Mensajes

Para cada mensaje: su etiqueta de dominio, sus campos en orden canónico y quién lo
firma. Todas las marcas de tiempo siguen la regla de `time` de §3.

La cláusula `Order:` que aparece a continuación es la ÚNICA autoridad sobre el
orden canónico de los campos, incluido el orden de los campos de estructuras
anidadas (p. ej., `Capacity.VCPUs` antes de `Capacity.MemMB`, y cada impresora como
`Kind` y luego `Model`). Un implementador MUST codificar los campos en la secuencia
de `Order:` y MUST NOT inferir el orden de ninguna otra fuente. En particular, el
ejemplo `testdata/vectors.json` (§10) enumera los `fields` de cada mensaje en
**orden alfabético de claves para facilitar la lectura, que NO es el orden de
codificación canónico** — seguir el orden de claves del JSON produciría bytes
erróneos y firmas inválidas.

### 4.1 CapabilityListing — firmado por el NODO
Domain tag: `sohocloud/listing/v0`

Order: `NodeID` (string), `Class` (string), `Printers` (repeated: `Kind` string,
`Model` string), `GPUs` (repeated: `API` string, `Model` string, `VRAMMB`
int64), `Capacity.VCPUs` (int64), `Capacity.MemMB` (int64),
`Capacity.DiskMB` (int64), `Capacity.StorageCommitMB` (int64),
`Capacity.PrintQPS` (int64), `OptIn.Compute` (bool), `OptIn.Print` (bool),
`OptIn.Storage` (bool), `IssuedAt` (time), `Seq` (uint64).

`Class` ∈ {`micro`, `standard`, `server`}. `Kind` ∈ {`traditional`, `threed`}.
`API` ∈ {`vulkan`, `nnapi`, `cuda`, `metal`}.

`Capacity.DiskMB` es el espacio temporal (scratch) disponible para un trabajo en
ejecución; `Capacity.StorageCommitMB` es el almacenamiento de larga duración que el
nodo se compromete a mantener para la red (alojamiento de fragmentos). Ambos son
deliberadamente distintos para que un nodo pueda ofrecer uno sin el otro. Cómo se
cifran, fragmentan y auditan los datos almacenados es un asunto del
frontend/agente, fuera de este protocolo — solo coordinación; el cable nunca
transporta contenido almacenado.

Anunciar una GPU es la adhesión voluntaria (opt-in) al trabajo con GPU: un anuncio
sin `GPUs` no recibe ninguno, y un nodo retira una GPU omitiéndola de su siguiente
anuncio.

`Seq` es estrictamente monotónico por nodo. Un coordinador MUST rechazar un anuncio
cuyo `Seq` no supere estrictamente el último visto para ese nodo (protección contra
repetición/reversión).

`OptIn` es **orientativo** (véase §5.1).

### 4.2 Heartbeat — firmado por el NODO
Domain tag: `sohocloud/heartbeat/v0`

Order: `NodeID` (string), `SentAt` (time), `Seq` (uint64).

La misma disciplina de `Seq` estrictamente monotónico que en §4.1.

### 4.3 Assignment — firmado por el COORDINADOR
Domain tag: `sohocloud/assignment/v0`

Order: `JobID` (string), `NodeID` (string), `Spec.Workload` (string),
`Spec.Image` (string), `Spec.Args` (repeated string), `Spec.PrinterKind`
(string), `Spec.GPUAPI` (string), `Spec.GPUMinVRAMMB` (int64),
`Fee.ContributorShareBps` (int64), `Fee.PlatformFeeBps` (int64),
`OfferedAt` (time).

`Spec.GPUAPI` ∈ {``, `vulkan`, `nnapi`, `cuda`, `metal`} — una indicación de
enrutamiento orientativa, como `PrinterKind`; vacío significa que no se requiere
GPU y, en ese caso, `GPUMinVRAMMB` MUST ser `0`.

Las condiciones de la tarifa se adjuntan en línea para que el nodo vea el reparto
antes de comprometerse. En el modelo de sondeo (pull) no hay un mensaje de
aceptación separado.

### 4.4 Decline — firmado por el NODO
Domain tag: `sohocloud/decline/v0`

Order: `JobID` (string), `NodeID` (string), `Reason` (string), `DeclinedAt`
(time).

`Reason` ∈ {`local_policy`, `capacity`, `unavailable`}. `local_policy` es la vía de
exclusión voluntaria (opt-out) (véase §5.1).

### 4.5 JobReport — firmado por el NODO
Domain tag: `sohocloud/jobreport/v0`

Order: `JobID` (string), `NodeID` (string), `ExitCode` (int64), `FailureCause`
(string), `TmpfsExhausted` (bool), `StartedAt` (time), `FinishedAt` (time).

Este reporte firmado es el hecho del que derivan la medición y el pago. Sus campos
reflejan el análisis de finalización del ejecutor que ya está en funcionamiento en
SoHoLINK.

### 4.6 FeeDeclaration — firmado por el COORDINADOR
Domain tag: `sohocloud/fee/v0`

Order: `CoordinatorID` (string), `Terms.ContributorShareBps` (int64),
`Terms.PlatformFeeBps` (int64), `EffectiveAt` (time), `Seq` (uint64).

`ContributorShareBps + PlatformFeeBps` SHOULD ser igual a `10000`.

### 4.7 StorageLease — firmado por el COORDINADOR
Domain tag: `sohocloud/lease/v0`

Order: `LeaseID` (string), `NodeID` (string), `ShardRef` (bytes),
`SizeClass` (int64), `Fee.ContributorShareBps` (int64),
`Fee.PlatformFeeBps` (int64), `IssuedAt` (time), `ExpiresAt` (time),
`Seq` (uint64).

El almacenamiento es un ARRENDAMIENTO, no un trabajo: una obligación continua de
conservar un fragmento sellado durante un plazo acotado. `ShardRef` es una
dirección de contenido opaca de 32 bytes y `SizeClass` un tamaño de carga útil
cuantizado — el protocolo MUST NOT transportar tamaños reales de objetos ni
contenido almacenado; el cifrado, el relleno y la fragmentación son asuntos del
frontend por encima de esta cintura. La renovación es un nuevo `StorageLease` para
el mismo `LeaseID` con un `Seq` estrictamente mayor (la misma regla de reversión
que en §4.1). Las condiciones de la tarifa viajan en línea, como en §4.3.

### 4.8 LeaseDecline — firmado por el NODO
Domain tag: `sohocloud/lease-decline/v0`

Order: `LeaseID` (string), `NodeID` (string), `Reason` (string),
`DeclinedAt` (time).

`Reason` ∈ {`local_policy`, `capacity`, `unavailable`}; `local_policy` es la vía de
exclusión voluntaria (§5.1) para el almacenamiento, exactamente como lo es §4.4
para los trabajos.

### 4.9 LeaseRelease — firmado por el NODO
Domain tag: `sohocloud/lease-release/v0`

Order: `LeaseID` (string), `NodeID` (string), `ReleasedAt` (time).

Un nodo siempre puede dejar de conservar (la soberanía incluye marcharse). La
liberación firmada pone fin a su medición; volver a colocar el fragmento es asunto
del frontend.

### 4.10 ProofChallenge / ProofResponse — respuesta firmada por el NODO
Domain tag (response): `sohocloud/proof/v0`

Un `ProofChallenge{LeaseID, Offset, Length, Nonce (16 bytes), IssuedAt}` NO se
firma: los nodos obtienen los desafíos mediante sondeo por el canal autenticado
(modelo pull), y un desafío no compromete a nadie a nada. El artefacto firmado es
la respuesta.

Response Order: `LeaseID` (string), `NodeID` (string), `Offset` (int64),
`Length` (int64), `Nonce` (bytes), `Digest` (bytes), `RespondedAt` (time).

`Digest` MUST calcularse exactamente como
`SHA-256(Nonce || uint64be(Offset) || uint64be(Length) || sealed[Offset : Offset+Length])`
sobre los bytes del fragmento sellado. La respuesta reitera el rango y el nonce
desafiados para valerse por sí misma como un hecho de medición no repudiable — la
contraparte de almacenamiento de §4.5. Un verificador MUST rechazar una respuesta
cuyo `(LeaseID, Nonce)` ya haya aceptado: los nonces son de un solo uso, que es lo
que impide reproducir una respuesta grabada.

### 4.11 KeyRotation — firmado por la clave ACTUAL del nodo
Domain tag: `sohocloud/node-rotate/v0`

Order: `NodeID` (string), `NewPublicKey` (bytes), `Algo` (string),
`RotatedAt` (time), `Seq` (uint64).

La posesión de la clave saliente autoriza a nombrar a su sucesora. Un verificador
MUST validar la firma contra la clave que tiene actualmente para el nodo (una
rotación firmada por la sucesora es una autoinstalación y MUST fallar), MUST hacer
cumplir un `Seq` estrictamente monotónico por nodo, y MUST rechazar una rotación
cuya `NewPublicKey` no sea una clave bien formada para `Algo` (para `ed25519`, que
no tenga exactamente 32 bytes) o cuyo `Algo` sea desconocido — de lo contrario, un
verificador que empezara a comprobar los mensajes siguientes contra una
`NewPublicKey` mal formada fallaría. Solo después de que todas estas comprobaciones
se superen comienza el verificador a verificar los mensajes siguientes del nodo
contra `NewPublicKey`. `Algo` ∈ {`ed25519`} en v0.

### 4.12 KeyRevocation — firmado por la clave que se REVOCA
Domain tag: `sohocloud/node-revoke/v0`

Order: `NodeID` (string), `RevokedPublicKey` (bytes), `RevokedAt` (time),
`Seq` (uint64).

Anula una clave sin sucesora. Un verificador MUST verificar la firma contra la
clave en la que **confía actualmente** para `NodeID`, y MUST exigir que
`RevokedPublicKey` sea igual a esa clave de confianza. Una revocación solo se
acepta contra la clave exacta que nombra y anula. Esta es la vinculación
antiabuso: un extraño puede autofirmar una revocación que lleve el `NodeID` de una
víctima y la propia clave del extraño en `RevokedPublicKey`, pero como esa clave no
es la que el coordinador tiene como de confianza para la víctima, la comprobación
de igualdad falla y la revocación se rechaza — nadie que no posea ya la clave
actual de un nodo puede dejarlo fuera de línea. (Borradores anteriores decían
"aceptar incondicionalmente"; esa redacción queda reemplazada — la vinculación a la
clave de confianza es obligatoria.) Dada la clave actual, la lectura segura de una
revocación es "dejar de confiar en esta clave". La reinscripción después de una
revocación se realiza por un canal externo, por la misma vía que la primera
inscripción, nunca mediante un mensaje en el cable que un ladrón de claves pudiera
falsificar.

---

## 5. Invariantes

### 5.1 La exclusión voluntaria se hace cumplir localmente, nunca en el cable
`CapabilityListing.OptIn` y las indicaciones de enrutamiento de `JobSpec` son
**orientativas**. El coordinador NO es una frontera de seguridad para la exclusión
voluntaria. Un nodo MUST hacer cumplir la exclusión voluntaria contra su propia
lista de permitidos de confianza local y MUST rechazar el trabajo que su política
local prohíba — mediante `Decline{Reason: local_policy}` — independientemente de lo
que haya anunciado antes o de lo que ofrezca el coordinador.

### 5.2 La política de emparejamiento no forma parte del protocolo
La interfaz `Coordinator` no expone ninguna clasificación ni puntuación. Cómo elige
un coordinador qué nodo recibe qué trabajo es privado de la implementación. El
protocolo MUST NOT incorporar un campo que dicte el emparejamiento, porque eso
soldaría a los consumidores a la política de un solo coordinador y anularía la
posibilidad de abandono que la interfaz existe para proteger.

### 5.3 Las tarifas son legibles y no retroactivas
Las condiciones de tarifa de un coordinador se publican como una `FeeDeclaration`
firmada y con marca de tiempo. Las condiciones MUST NOT cambiar retroactivamente
para el trabajo ya ofrecido; un cambio es una nueva declaración firmada con un
`EffectiveAt` posterior. Esto mantiene las tarifas impugnables en lugar de opacas
(problema abierto n.º 7).

### 5.4 El coordinador es un rol, no un centro
Nada de lo que necesite un nodo o un frontend puede pasar *exclusivamente* por un
coordinador específico. Un nodo MAY servir a varios coordinadores; un frontend MAY
ejecutar su propio coordinador o emparejar directamente. Cualquier despliegue en el
que un único coordinador sea de hecho inevitable es una recentralización y MUST
etiquetarse como tal, con una vía hacia alternativas.

### 5.5 Números de secuencia monotónicos
Los valores `Seq` de los nodos (anuncios, heartbeats) son estrictamente monotónicos
por nodo. Un coordinador MUST rechazar un `Seq` que no sea creciente.

---

## 6. Operaciones del coordinador (modelo pull)

El modelo es de **sondeo (pull)**: el coordinador nunca llama al nodo. Operaciones:

| Operación       | Dirección           | Cuerpo / resultado                   |
|-----------------|---------------------|--------------------------------------|
| `SubmitListing` | nodo → coord.       | `CapabilityListing`                  |
| `Heartbeat`     | nodo → coord.       | `Heartbeat`                          |
| `PollJobs`      | nodo → coord.       | `NodeID` → `[]Assignment`            |
| `Decline`       | nodo → coord.       | `Decline`                            |
| `ReportJob`     | nodo → coord.       | `JobReport`                          |
| `Fees`          | cualquiera → coord. | → `FeeDeclaration`                   |

Un coordinador MUST aplicar la vinculación de §2 a `PollJobs` (y a cualquier
operación que nombre a un nodo) antes de actuar. Como el coordinador nunca inicia
el contacto con un nodo, no hay ningún servicio del lado del nodo que especificar.

---

## 7. Transporte

HTTP+JSON es el transporte de **referencia** (`transport/httpjson`), NO el cable en
sí. Ningún paquete central lo importa. Una implementación conforme MAY usar
cualquier transporte implementando las operaciones del coordinador contra los bytes
canónicos de §3. JSON es solo una codificación de transporte; las firmas nunca se
calculan sobre JSON, por lo que un transporte diferente interopera sin renegociar
las firmas.

Los endpoints HTTP+JSON de referencia:

```
POST /v0/listing     body: CapabilityListing   -> 204
POST /v0/heartbeat   body: Heartbeat           -> 204
GET  /v0/jobs?node_id=<id>                      -> 200 []Assignment
POST /v0/decline     body: Decline             -> 204
POST /v0/report      body: JobReport           -> 204
GET  /v0/fees                                   -> 200 FeeDeclaration
```

La vinculación SPIFFE (§2) la aplica un middleware situado delante del manejador;
el manejador de referencia solo conecta las rutas y el JSON y no es el
autenticador.

---

## 8. Problemas abiertos que toca este protocolo

La honestidad sobre las brechas sin resolver es el método, no una vergüenza.

- **n.º 7 El cómputo soberano compra una salida mecánica, no económica.** Esta es la
  capa de la que trata el n.º 7. §5.4 (el coordinador como rol) y §5.3 (tarifas
  legibles) empujan en contra de la recentralización de los coordinadores; el
  arranque en frío sigue siendo un residuo, y la portabilidad de la reputación que
  lo aliviaría es un asunto del frontend/de la gobernanza, no de este protocolo.
- **n.º 6 La honestidad del cómputo queda fuera de alcance.** Si el emparejamiento o
  el precio de un coordinador fueron justos no se demuestra aquí; solo se hace
  legible (§5.2, §5.3). `anchor/` reserva el lugar para anclar las *afirmaciones*
  de empleo (los hechos de `JobReport`) en un registro atestiguado, para que al
  menos sean auditables. Es un esbozo etiquetado, no construido.
- **n.º 4 Membresía / incorporación.** Como un nodo se nombra y firma por sí mismo
  (§2), un dispositivo de pocos recursos (p. ej., un teléfono) puede anunciarse
  como nodo. Si la incorporación de nodos es abierta o restringida es una decisión
  de gobernanza/valores que queda abierta.

## 9. Versionado

`v0` es inestable. Las etiquetas de dominio incorporan `v0`; un cambio
incompatible en los campos canónicos de un mensaje incrementa su etiqueta. Los
consumidores MUST tratar una etiqueta de dominio desconocida como un mensaje no
verificable.

---

## 10. Vectores de prueba

`testdata/vectors.json` es el **dispositivo de conformidad normativo** (fixture)
de esta especificación. Una implementación independiente es conforme en cuanto a
codificación canónica y firma si y solo si reproduce, byte por byte, cada valor de
ese archivo a partir de las mismas entradas. El fixture lo generan los
codificadores de referencia en Go y se verifica de forma cruzada mediante una
reproducción independiente que solo usa la biblioteca estándar; una implementación
externa conforme MUST coincidir con él exactamente. Si alguna vez la prosa de
§3/§4 y el fixture discrepan, se trata de un defecto de la especificación que debe
corregirse, no de una licencia para divergir — pero el fixture es la declaración
ejecutable de §3/§4.

Todo el hexadecimal del fixture está en minúsculas. Es autodescriptivo: una `note`
de nivel superior y un bloque `encoding` reiteran las reglas de primitivas de §3
para quien abra primero el JSON.

### 10.1 Estructura

- **`primitives`** — codificaciones aisladas de cada primitiva de §3, para que un
  implementador pueda validar el codificador de abajo hacia arriba antes de
  ensamblar mensajes. Cada entrada tiene un `kind` (`uvarint`, `int64`, `uint64`,
  `bool`, `string`, `time`), un `input` legible por humanos y los `bytes_hex`
  esperados. Estas cubren los casos límite que fijan el formato: valores varint en
  cada umbral de byte de continuación (`0, 1, 127, 128, 255, 256, 300, 16384, 1073741824`);
  `int64` incluidos `-1`, `-2`, `MaxInt64` y `MinInt64`; `uint64` incluido
  `MaxUint64`; ambos bools; la cadena vacía y una cadena UTF-8 multibyte; y `time`
  en la época Unix y en un instante fijo en nanosegundos.

- **`messages`** — una entrada por tipo de mensaje firmado (cada mensaje firmado de §4). Cada entrada
  registra: `name`, `domain_tag`, `signer` (`node` o `coordinator`), la `seed_hex`
  ed25519 de 32 bytes y la `public_key_hex` derivada, los `fields` del mensaje, los
  `canonical_bytes_hex` (la salida de `CanonicalBytes()`, con `Signature`
  excluido) y la `signature_hex` (ed25519 sobre exactamente esos bytes canónicos).

- **`operator_messages`** — una entrada por tipo de mensaje de operador de la Capa C
  (los tres de §11). Como los mensajes de operador son 2 de 7, cada entrada difiere
  de una entrada central de `messages`: `signer` es siempre `operator`; un arreglo
  `keys` enumera las siete claves públicas registradas (`index`, `seed_hex`,
  `public_key_hex`, `algo` — las semillas están presentes SOLO para que el fixture
  sea autocontenido y reproducible, nunca porque la red tenga claves privadas);
  `idx0`/`idx1` son los dos índices firmantes; y en lugar de una `signature_hex`
  hay `sig0_hex` y `sig1_hex`, AMBAS tomadas sobre los MISMOS
  `canonical_bytes_hex`. Una implementación conforme reproduce
  `canonical_bytes_hex` a partir del orden de campos de §11 y verifica ambas firmas
  en sus índices.

  > **Trampa del orden de campos.** El objeto `fields` de cada mensaje enumera las
  > claves en **orden alfabético para facilitar la lectura, NO en orden canónico.**
  > El orden canónico proviene ÚNICAMENTE de la cláusula `Order:` de ese mensaje en
  > §4. Un implementador que codifique los campos en el orden de claves del JSON
  > producirá bytes erróneos y firmas que no verifican.

### 10.2 Cómo usar el fixture

1. Reproducir cada `primitives[].bytes_hex` a partir de su `input` usando §3.
2. Para cada mensaje, derivar el par de claves a partir de `seed_hex` (ed25519 a
   partir de una semilla de 32 bytes es determinista), confirmar que produce
   `public_key_hex`, codificar los campos en el `Order:` de §4 y confirmar que se
   obtiene `canonical_bytes_hex`, y luego verificar `signature_hex` contra
   `public_key_hex` sobre tus bytes derivados de forma independiente.
   La verificación de la firma sobre tus PROPIOS bytes es la verificación cruzada
   más sólida: solo puede superarse si tus bytes son idénticos a los del generador
   de referencia.

La prueba de referencia está en `vectors/vectors_test.go`. Regenera todos los
vectores a partir del código actual y afirma la igualdad byte por byte con el JSON
confirmado en el repositorio, de modo que cualquier desviación entre el código y el
fixture hace fallar la construcción; también vuelve a analizar el archivo
confirmado y verifica cada firma. Para regenerar de forma intencionada, usa
`go test ./vectors -update` (o la variable de entorno
`SOHOCLOUD_UPDATE_VECTORS`). Todos los vectores usan semillas de relleno de bytes
fijo y marcas de tiempo codificadas de forma fija — no hay `crypto/rand` ni
`time.Now()`, por lo que la salida es totalmente determinista.

### 10.3 Ejemplo resuelto: un `Heartbeat`, byte por byte

Este es el vector `Heartbeat` (`§4.2`) decodificado campo por campo, para que quien
reimplemente pueda seguir la codificación completa de principio a fin. Entradas:
`NodeID = "node-alpha"`, `SentAt = ` nanosegundos Unix `1700000050500000000`,
`Seq = 43`.

Bytes canónicos (`canonical_bytes_hex`):

```
16 736f686f636c6f75642f6865617274626561742f7630 0a 6e6f64652d616c706861 17979d09f832d900 000000000000002b
```

Decodificados de izquierda a derecha:

| Bytes | Primitiva | Significado |
|-------|-----------|-------------|
| `16` | longitud uvarint | Longitud de la etiqueta de dominio = `0x16` = **22** bytes. |
| `73 6f 68 6f 63 6c 6f 75 64 2f 68 65 61 72 74 62 65 61 74 2f 76 30` | bytes UTF-8 | La etiqueta de dominio `sohocloud/heartbeat/v0` (22 bytes). Junto con el prefijo de longitud, estas dos filas son la codificación `string` de la etiqueta de dominio (§3). |
| `0a` | longitud uvarint | Longitud de `NodeID` = `0x0a` = **10** bytes. |
| `6e 6f 64 65 2d 61 6c 70 68 61` | bytes UTF-8 | `NodeID` = `node-alpha` (10 bytes). |
| `17 97 9d 09 f8 32 d9 00` | int64, big-endian | `SentAt` en nanosegundos Unix UTC. `0x17979d09f832d900` = `1700000050500000000`. Esta es la regla de `time` (§3): el instante reducido a un `int64` y emitido como 8 bytes en big-endian. |
| `00 00 00 00 00 00 00 2b` | uint64, big-endian | `Seq` = `0x2b` = **43**. 8 bytes fijos, sin varint. |

La concatenación de exactamente estos seis grupos — etiqueta de dominio (longitud +
bytes), `NodeID` (longitud + bytes), `SentAt`, `Seq` — es la cadena completa de
bytes canónicos. No hay encuadre antes de la etiqueta de dominio ni bytes después
de `Seq`. La firma ed25519 (`signature_hex`) se toma sobre precisamente esta
cadena; el campo `Signature` en sí nunca forma parte de ella.

Observa cómo el orden de los campos aquí (`NodeID`, `SentAt`, `Seq`) sigue el
`Order:` de §4.2, que es también — por coincidencia en este mensaje pequeño — como
los enumeraría un lector; en mensajes más grandes como `CapabilityListing`, el
`Order:` de §4 y las claves alfabéticas de `fields` en `vectors.json` difieren, y
solo el orden de §4 es canónico.

---

## 11. Identidad del operador (Capa C)

§1–§10 rigen los mensajes de **nodo** y de **coordinador**. Esta sección añade una
tercera identidad, ortogonal: el **operador** — un frontend (p. ej., Cloudy) que
termina la identidad de los miembros y de los nodos dentro de sí mismo y presenta
una única credencial rotativa a un coordinador. Está deliberadamente separada de la
identidad de carga de trabajo del nodo (§2): los mensajes centrales de coordinación
(§4) NO CAMBIAN y NO se firman con una credencial de operador.

Un operador posee **siete** pares de claves Ed25519 (índices `0..6`) y firma cada
mensaje con **dos índices distintos** (una disciplina de **2 de 7**). La red
registra solo las siete claves **públicas** por operador; **la red nunca posee
ninguna clave privada de operador**, y toda ruta de verificación descrita a
continuación recibe solo claves públicas.

- **2 de 7 es higiene contra la sustitución / de rotación, NO seguridad de
  umbral.** Si las siete claves privadas residen en un mismo proceso, el compromiso
  de un solo host entrega las siete. No se confía en 2 de 7 para sobrevivir al
  compromiso de una sola clave a menos que la custodia esté repartida entre
  dominios de confianza. Se declara para que ningún lector confíe en ello más de la
  cuenta.
- **Vinculación del algoritmo.** Cada mensaje de operador vincula una cadena de
  algoritmo `Algo` dentro de sus bytes canónicos. `v0` permite exactamente
  `ed25519`. Un verificador MUST rechazar un mensaje cuyo `Algo` sea desconocido, y
  MUST rechazar un mensaje cuyo `Algo` no sea igual al `algo` registrado de la clave
  en cada índice firmante. El conjunto de claves activas de un operador MUST NOT
  mezclar algoritmos. El control de longitud de firma es por algoritmo (`ed25519` =
  64 bytes), no una constante fija en el código, de modo que la rotación atómica
  planificada de todo el conjunto a un algoritmo poscuántico (una NUEVA etiqueta de
  dominio, p. ej., `sohocloud/operator/v1-mldsa`) reutiliza la misma regla. Una
  firma no es transferible entre algoritmos.
- **Junta de migración.** La implementación de referencia encamina la firma y la
  verificación del operador a través de una interfaz `Signer`/`Verifier`,
  implementada hoy con Ed25519 de la biblioteca estándar, precisamente para que el
  algoritmo pueda sustituirse más adelante añadiendo una etiqueta de dominio `v1`
  sin tocar los mensajes centrales. Esta es una nota de implementación; el contrato
  del cable son los órdenes de campos y las etiquetas que siguen.

**Codificación de los campos de operador (sin casos especiales).** Cada campo de
las cláusulas `Order:` de §11 se codifica según su regla de primitiva ordinaria de
§3; §11 no introduce tipos nuevos ni de ancho fijo. En particular:

- `OperatorID` y `Algo` son **strings** ordinarias de §3 — un varint LEB128 sin
  signo con la longitud en bytes UTF-8 seguido de los bytes UTF-8. `Algo` NO es un
  token desnudo ni una etiqueta de enumeración de ancho fijo, a pesar de que `v0`
  permite exactamente el único valor `ed25519`; MUST llevar su prefijo de longitud
  como cualquier otra cadena (p. ej., `ed25519` se codifica como
  `07 65 64 32 35 35 31 39`). `Nonce`, `NewPublicKey` y `Challenge` son **bytes**
  ordinarios de §3 (prefijo de longitud LEB128 y luego los bytes sin procesar).
- `Idx0`, `Idx1` y `KeyIndex` se codifican cada uno como un **`uint64` big-endian
  fijo de 8 bytes** según la regla de `uint64` de §3, exactamente igual que `Seq` —
  NO como varint y NO como un solo byte — aunque sus valores estén en el rango
  pequeño `0..6`. Un implementador que codifique un índice como varint o como un
  byte produce bytes erróneos y una firma inválida. (Ejemplo: el índice `2` se
  codifica como `00 00 00 00 00 00 00 02`.)
- Las dos firmas `Sig0` y `Sig1` se **excluyen** de los bytes canónicos de todo
  mensaje de §11, exactamente como `Signature` para los mensajes de §4 (§3). Son
  los únicos campos excluidos y, al igual que `Signature`, no aparecen en ninguna
  cláusula `Order:` de las siguientes. No hay un tercer campo de firma.
  `Idx0`/`Idx1` (que nombran las claves firmantes) SÍ están dentro de los bytes
  firmados; los *bytes* de la firma no.

### 11.0 Requisitos de verificación (todos los mensajes de operador)

Un verificador MUST, antes de aceptar cualquier mensaje de operador:

1. Rechazar si el nonce del mensaje está ausente o tiene menos de **16 bytes**
   (cuando el mensaje tiene nonce). Un nonce vacío MUST NEVER tratarse como "omitir
   la caché de repetición".
2. Rechazar si `Algo` no es un algoritmo admitido para esta versión.
3. Rechazar si los dos índices firmantes `Idx0` e `Idx1` son iguales.
4. Rechazar si alguno de los índices está fuera de `0..6`.
5. Rechazar si no existe una clave registrada activa en alguno de los índices, o si
   el `algo` de alguna de las claves registradas no es igual al `Algo` del mensaje.
5a. Rechazar si las claves públicas registradas en `Idx0` e `Idx1` son idénticas
   byte por byte, aunque los índices difieran. Las siete claves públicas
   registradas de un operador MUST ser distintas dos a dos; un registro o una
   rotación MUST NOT instalar una clave pública que ya exista en otro índice (véase
   §11.2). Esto se hace cumplir para que la propiedad antisustitución de 2 de 7 no
   pueda degradarse silenciosamente a 1 de 1: si dos índices contuvieran la misma
   clave, una sola clave privada podría producir ambas firmas. Esta comprobación
   recae sobre las *claves*, no solo sobre los índices (regla 3), y se exige además
   de ella.
6. Rechazar si la longitud de alguna de las firmas no es la esperada para `Algo`.
7. Recalcular los bytes canónicos y rechazar a menos que AMBAS firmas verifiquen en
   sus índices sobre exactamente esos bytes.

La protección contra repetición (un conjunto duradero, de cierre ante fallos
(fail-closed), de ventana deslizante de `Seq` + nonce, con alcance por
`(operator, coordinator)`) es una obligación del lado del coordinador que se
superpone a lo ANTERIOR; no forma parte de los bytes canónicos y queda fuera del
alcance de esta especificación de codificación.

### 11.1 OperatorTransmission — firmado por el OPERADOR (2 de 7)
Domain tag: `sohocloud/operator/v0`

Order: `OperatorID` (string), `Ts` (int64 UTC Unix nanoseconds — el `int64` sin
procesar, NO rederivado a través de un `time.Time`, de modo que el valor con rango
comprobado sea exactamente el valor firmado), `Nonce` (bytes, `len >= 16`), `Seq`
(uint64), `Algo` (string), `Idx0` (uint64), `Idx1` (uint64).

Las dos firmas `Sig0` (por la clave en `Idx0`) y `Sig1` (por la clave en `Idx1`) se
toman sobre estos bytes canónicos y quedan EXCLUIDAS de ellos, exactamente como
`Signature` en los demás casos. `Ts`, `Nonce`, `Seq`, ambos índices y `Algo` están
todos dentro de los bytes firmados.

### 11.2 OperatorRotation — firmado por el OPERADOR (2 de 7)
Domain tag: `sohocloud/operator-rotate/v0`

Order: `OperatorID` (string), `KeyIndex` (uint64 — el índice cuya clave se
reemplaza), `NewPublicKey` (bytes — la clave pública de reemplazo generada por el
operador), `Algo` (string), `Ts` (int64 UTC Unix nanoseconds), `Nonce` (bytes,
`len >= 16`), `Seq` (uint64), `Idx0` (uint64), `Idx1` (uint64).

Esto autoriza a incorporar una nueva clave pública. El operador (nunca la red)
genera el nuevo par de claves. La nueva clave pública está DENTRO de los bytes
firmados para que ni un intermediario (man-in-the-middle) del registro por canal
externo ni un canal de administración comprometido puedan inyectar material de
claves: un coordinador MUST verificar este mensaje (con dos claves ACTUALES,
`Idx0`/`Idx1`) antes de registrar la nueva clave en `KeyIndex`. Un verificador MUST
además rechazar una `NewPublicKey` cuya longitud sea incorrecta para `Algo`. Un
verificador MUST también rechazar una `NewPublicKey` que sea idéntica byte por byte
a una clave ya registrada en un índice DIFERENTE (violaría el requisito de claves
distintas dos a dos de la regla 5a de §11.0 y prepararía la degradación de 2 de 7 →
1 de 1); volver a registrar la misma clave en su PROPIO `KeyIndex` (una
actualización sin efecto) está permitido. Para preservar la invariante de no
mezclar algoritmos, un verificador SHOULD también rechazar una rotación cuyo `Algo`
difiera del de las demás claves registradas.

### 11.3 ConformanceResponse — firmado por el OPERADOR (2 de 7)
Domain tag: `sohocloud/operator-conformance/v0`

Order: `OperatorID` (string), `Challenge` (bytes — el desafío proporcionado por el
verificador), `Algo` (string), `Idx0` (uint64), `Idx1` (uint64).

Esta es la respuesta firmada de un operador a un desafío de conformidad. Su
etiqueta de dominio distinta la separa por dominio de §11.1: como la etiqueta forma
parte de los bytes canónicos, una firma producida para un desafío de conformidad
NUNCA puede verificarse como una `OperatorTransmission`, y viceversa. No lleva
nonce/`Seq` propios — la frescura proviene del `Challenge` del verificador — por lo
que la regla de nonce de §11.0 no se le aplica; todas las demás comprobaciones de
§11.0 sí.
