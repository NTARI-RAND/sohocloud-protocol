> Traducción comunitaria (borrador) — Política P2-002 de NTARI, Difusión Multilingüe Global. Fuente: SPEC.md (original en inglés, instantánea del 2026-09-16). Borrador comunitario asistido por máquina, pendiente de revisión por el mantenedor regional conforme a P2-002 §3.1. Las especificaciones técnicas centrales permanecen en inglés conforme al §2.2.
>
> **El original en inglés es normativo.** Esta traducción se ofrece para lectura,
> no para interpretación: donde diverja de `SPEC.md`, manda el inglés. Se
> conservan en inglés, por ser el contrato de conformidad, las palabras clave
> normativas (MUST, MUST NOT, SHOULD, MAY), los nombres de campo, las etiquetas
> de dominio, los identificadores de tipo y todos los bloques de código.
>
> ¿Encontraste un error en esta traducción? Tu corrección es una contribución
> bienvenida y valorada: haz un fork del repositorio y abre un pull request en
> https://github.com/NTARI-RAND/sohocloud-protocol.

# sohocloud-protocol — Especificación de conformidad

Versión: **v0 (INESTABLE)** — las formas de los mensajes y el formato canónico de
bytes pueden cambiar sin garantías de compatibilidad hasta la v1.

Este documento es normativo y es un entregable de construcción, no un añadido
posterior. El derecho AGPL a bifurcar la red es vacío si una implementación no
puede construirse a partir de una especificación legible sin leer el código
fuente de este módulo. Todo lo que aquí se marque MUST / MUST NOT es un requisito
de conformidad; un cambio que viole uno de ellos no es una versión menor de este
protocolo, es un protocolo distinto.

---

## 1. Alcance

Este protocolo gobierna la **capa de coordinación del sustrato**: cómo se
reconoce a un nodo, cómo anuncia qué trabajo hará, y cómo se le ofrece trabajo y
se informa de vuelta. Tres partes lo hablan:

- **Node** — un dispositivo propiedad de un participante que ofrece cómputo y/o
  impresión física. Un nodo es dueño de su identidad y firma todo lo que emite.
  Soberano: elige a qué coordinadores servir y puede abandonar cualquiera de
  ellos.
- **Coordinator** — empareja listados con trabajo y registra resultados. Un rol
  intercambiable. SoHoLINK es el coordinador de referencia; un frontend MAY
  ejecutar el suyo propio o emparejar directamente contra nodos.
- **Frontend** — donde los miembros transaccionan. Consume coordinación a través
  de la misma interfaz.

**Fuera de alcance, deliberadamente.** Este protocolo NO es la economía de
miembros de JFA. No define escrow, ni crédito emitido por miembros, ni pacto de
reputación, ni registro sellado por diálogo, ni adjudicación de disputas. Eso
pertenece a cada frontend. La liquidación en moneda fiduciaria que un coordinador
ejecuta con sus nodos (p. ej. los pagos por Stripe de SoHoLINK) es la economía
del *sustrato* y queda igualmente fuera de esta especificación.

---

## 2. Identidad y vinculación

Un nodo se nombra mediante un `NodeID` — una cadena estable que el propio nodo
elige.

La identidad de carga de trabajo de un nodo es un SPIFFE SVID emitido exactamente
bajo:

```
/node/<NodeID>
```

Un coordinador autoriza una petición como proveniente de un nodo comparando la
ruta SPIFFE autenticada del llamante con `/node/<NodeID>` para el `NodeID`
nombrado en la petición.

- Un coordinador MUST rechazar con **401** cuando no hay identidad SPIFFE
  presente.
- Un coordinador MUST rechazar con **403** cuando la identidad está presente pero
  su ruta no es exactamente `/node/<NodeID>`.
- Un coordinador MUST NOT aceptar "cualquier SVID válido" como autorización para
  una petición que nombre a un nodo distinto. Eso es un vector de suplantación.
- La vinculación MUST ser construcción determinista de cadenas, nunca una
  búsqueda por nodo en almacenamiento mutable. Una comprobación condicionada a un
  campo que nunca se puebla es inerte y MUST NOT usarse como frontera de
  seguridad.

El protocolo no distribuye claves públicas. Un coordinador resuelve la clave de
verificación de un nodo fuera de banda (registro, material del SVID o un
registro de claves).

---

## 3. Codificación canónica de bytes (para firmas)

Todo mensaje firmado se firma sobre sus **bytes canónicos**, calculados con el
campo `Signature` EXCLUIDO. Las firmas son **ed25519**.

La firma NO se realiza sobre JSON ni sobre ninguna codificación de transporte. Se
realiza sobre el siguiente formato determinista con prefijo de longitud, de modo
que las firmas sean estables entre plataformas, versiones de Go y lenguajes.

Primitivas, en el orden exacto en que se listan los campos de cada mensaje en §4:

- **unsigned LEB128 varint** — la primitiva de longitud/conteo usada por
  `string`, `bytes` y `repeated`. MUST codificarse como sigue, con independencia
  del lenguaje: tómese el valor sin signo; emítanse 7 bits por byte, primero el
  grupo menos significativo (grupos little-endian); póngase el bit alto (`0x80`)
  de cada byte EXCEPTO el último para marcar continuación; el último byte tiene
  su bit alto a cero. El valor `0` se codifica como el byte único `0x00`. Esta es
  exactamente la codificación que produce `encoding/binary.PutUvarint` de Go; el
  símbolo Go es una verificación cruzada, no la definición. (Ejemplo resuelto:
  `300` = `0b100101100` → 7 bits bajos `0101100` con continuación → `0xAC`, bits
  altos `0000010` al final → `0x02`, dando `AC 02`.)
- **domain tag** — los bytes canónicos de todo mensaje comienzan con una cadena
  de etiqueta de dominio distinta (véase §4). La etiqueta de dominio MUST
  codificarse con la regla de **string** de abajo — es decir, un unsigned LEB128
  varint de su longitud en bytes UTF-8 seguido de sus bytes UTF-8 — exactamente
  como cualquier otro campo de cadena, sin terminador nulo y sin tratamiento
  especial. Etiquetas distintas hacen que una firma sobre un tipo de mensaje no
  sea transferible a otro.
- **string** — unsigned LEB128 varint con la longitud en bytes, luego los bytes
  UTF-8 crudos. Una cadena vacía MUST codificarse como el byte único `0x00` (una
  longitud cero y ningún contenido).
- **bytes** — unsigned LEB128 varint con la longitud en bytes, luego los bytes
  crudos. (El campo `Signature` nunca se codifica; queda excluido.)
- **bool** — un solo byte, exactamente `0x00` (falso) o `0x01` (verdadero);
  ningún otro valor de byte es válido.
- **int64 / uint64** — 8 bytes fijos, big-endian (primero el byte más
  significativo). `int64` MUST codificarse como su reinterpretación `uint64` en
  complemento a dos: súmese 2⁶⁴ a un valor negativo y luego emítanse los 8 bytes
  big-endian. Así `-1` → `ff ff ff ff ff ff ff ff`, `-2` →
  `ff ff ff ff ff ff ff fe`, y el `int64` mínimo (`-9223372036854775808`) →
  `80 00 00 00 00 00 00 00`. Sin signo-magnitud, sin zigzag, sin longitud
  variable.
- **time** — el instante como `int64` de **nanosegundos** Unix UTC (según la
  regla de int64 anterior, big-endian, 8 bytes). Se descartan la ubicación y los
  componentes de reloj monótono. La precisión de nanosegundos se preserva
  exactamente; el valor MUST NOT truncarse a micro- ni milisegundos. Esto
  sobrevive a un viaje de ida y vuelta por transporte JSON y es independiente de
  la zona horaria. Un instante fuera del rango representable de nanosegundos
  Unix UTC en `int64` (aproximadamente los años 1678–2262) queda fuera de alcance
  para v0 y MUST tratarse como error de codificación en lugar de envolverse o
  recortarse. El codificador de referencia lo impone: `canon.Buffer.Time` no
  añade nada y fija `canon.ErrTimeOutOfRange` (recuperable mediante
  `Buffer.Err()`) en vez de emitir un valor envuelto, de modo que dos instantes
  separados por ~584 años nunca puedan colapsar en bytes idénticos. Un productor
  MUST rechazar mediante `canon.ValidTime` antes de firmar; un
  verificador/decodificador MUST rechazar un mensaje cuya marca de tiempo no la
  supere.
- **repeated field** — un unsigned LEB128 varint con el *conteo* (el número de
  elementos) codificado PRIMERO, luego los campos de cada elemento codificados en
  orden, completamente en línea, usando las reglas primitivas normales de arriba.
  Por tanto, un elemento repetido que sea a su vez una cadena lleva prefijo de
  longitud como cualquier otra cadena; los campos de estructuras anidadas se
  emiten en el orden dado para ese elemento en §4. No hay separador entre
  elementos ni terminador tras el último. Un conteo de `0` es el byte único
  `0x00` seguido de ningún elemento.

Un verificador recalcula los bytes canónicos a partir de los campos recibidos y
comprueba la firma ed25519 contra la clave pública resuelta. Un verificador MUST
rechazar una firma cuya longitud no sea `ed25519.SignatureSize` (64) antes de
verificar.

Los bytes canónicos de un mensaje son **exactamente** la concatenación de la
etiqueta de dominio seguida de cada campo listado en la cláusula `Order:` de ese
mensaje en §4, en ese orden, cada uno codificado por su regla primitiva. No hay
encuadre externo: ningún byte de versión inicial, ningún prefijo de longitud
total que envuelva el mensaje, y ningún byte final tras el último campo. La lista
`Order:` de §4 es exhaustiva y autoritativa — los **campos de firma** del mensaje
son los ÚNICOS campos excluidos, y no existe en los bytes canónicos ningún campo
que no esté nombrado allí. Para los doce mensajes centrales de §4, ese único
campo excluido es `Signature`. Los mensajes de operador de la Capa C de §11
llevan **dos** firmas (`Sig0` y `Sig1`) sobre los mismos bytes canónicos; ambas
quedan excluidas exactamente igual que `Signature` aquí, y, como `Signature`,
ninguna aparece en una cláusula `Order:`. Por tanto, un verificador nunca
codifica ningún campo de firma en los bytes que recalcula, tenga el mensaje una
firma o dos.

---

## 4. Mensajes

Para cada mensaje: su etiqueta de dominio, sus campos en orden canónico y quién
lo firma. Todas las marcas de tiempo siguen la regla `time` de §3.

La cláusula `Order:` de abajo es la ÚNICA autoridad sobre el orden canónico de
campos, incluido el orden de los campos de estructuras anidadas (p. ej.
`Capacity.VCPUs` antes de `Capacity.MemMB`, y cada impresora como `Kind` y luego
`Model`). Quien implemente MUST codificar los campos en la secuencia `Order:` y
MUST NOT inferir el orden de ninguna otra fuente. En particular, el ejemplo
`testdata/vectors.json` (§10) lista los `fields` de cada mensaje en **orden
alfabético de claves por legibilidad, lo que NO es el orden canónico de
codificación** — seguir el orden de claves del JSON produciría bytes incorrectos
y firmas inválidas.

### 4.1 CapabilityListing — firmado por el NODE
Domain tag: `sohocloud/listing/v0`

Order: `NodeID` (string), `Class` (string), `Printers` (repeated: `Kind` string,
`Model` string), `GPUs` (repeated: `API` string, `Model` string, `VRAMMB`
int64), `Capacity.VCPUs` (int64), `Capacity.MemMB` (int64),
`Capacity.DiskMB` (int64), `Capacity.StorageCommitMB` (int64),
`Capacity.PrintQPS` (int64), `OptIn.Compute` (bool), `OptIn.Print` (bool),
`OptIn.Storage` (bool), `IssuedAt` (time), `Seq` (uint64).

`Class` ∈ {`micro`, `standard`, `server`}. `Kind` ∈ {`traditional`, `threed`}.
`API` ∈ {`vulkan`, `nnapi`, `cuda`, `metal`}.

`Capacity.DiskMB` es espacio de trabajo disponible para un trabajo en ejecución;
`Capacity.StorageCommitMB` es almacenamiento de larga duración que el nodo se
compromete a mantener para la red (hospedaje de fragmentos). Ambos son
deliberadamente distintos para que un nodo pueda ofrecer uno sin el otro. Cómo se
cifran, fragmentan y auditan los datos almacenados es asunto del frontend/agente,
fuera de este protocolo — solo coordinación; el cable nunca transporta contenido
almacenado.

Anunciar una GPU es la adhesión al trabajo con GPU: un listado sin `GPUs` no
recibe ninguno, y un nodo retira una GPU omitiéndola de su siguiente listado.

`Seq` es estrictamente monótono por nodo. Un coordinador MUST rechazar un listado
cuyo `Seq` no supere estrictamente el último visto para ese nodo (protección
contra repetición/retroceso).

`OptIn` es **consultivo** (véase §5.1).

### 4.2 Heartbeat — firmado por el NODE
Domain tag: `sohocloud/heartbeat/v0`

Order: `NodeID` (string), `SentAt` (time), `Seq` (uint64).

Misma disciplina de `Seq` estrictamente monótono que en §4.1.

### 4.3 Assignment — firmado por el COORDINATOR
Domain tag: `sohocloud/assignment/v0`

Order: `JobID` (string), `NodeID` (string), `Spec.Workload` (string),
`Spec.Image` (string), `Spec.Args` (repeated string), `Spec.PrinterKind`
(string), `Spec.GPUAPI` (string), `Spec.GPUMinVRAMMB` (int64),
`Fee.ContributorShareBps` (int64), `Fee.PlatformFeeBps` (int64),
`OfferedAt` (time).

`Spec.GPUAPI` ∈ {``, `vulkan`, `nnapi`, `cuda`, `metal`} — una pista consultiva
de enrutado como `PrinterKind`; vacío significa que no se requiere GPU y
`GPUMinVRAMMB` MUST ser entonces `0`.

Los términos de la tarifa se adjuntan en línea para que un nodo vea el reparto
antes de comprometerse. En el modelo de extracción no hay un mensaje de
aceptación aparte.

### 4.4 Decline — firmado por el NODE
Domain tag: `sohocloud/decline/v0`

Order: `JobID` (string), `NodeID` (string), `Reason` (string), `DeclinedAt`
(time).

`Reason` ∈ {`local_policy`, `capacity`, `unavailable`}. `local_policy` es la vía
de exclusión (véase §5.1).

### 4.5 JobReport — firmado por el NODE
Domain tag: `sohocloud/jobreport/v0`

Order: `JobID` (string), `NodeID` (string), `ExitCode` (int64), `FailureCause`
(string), `TmpfsExhausted` (bool), `StartedAt` (time), `FinishedAt` (time).

Este informe firmado es el hecho del que se derivan la medición y el pago. Sus
campos reflejan el análisis de finalización del ejecutor ya activo en SoHoLINK.

### 4.6 FeeDeclaration — firmado por el COORDINATOR
Domain tag: `sohocloud/fee/v0`

Order: `CoordinatorID` (string), `Terms.ContributorShareBps` (int64),
`Terms.PlatformFeeBps` (int64), `EffectiveAt` (time), `Seq` (uint64).

`ContributorShareBps + PlatformFeeBps` SHOULD ser igual a `10000`.

### 4.7 StorageLease — firmado por el COORDINATOR
Domain tag: `sohocloud/lease/v0`

Order: `LeaseID` (string), `NodeID` (string), `ShardRef` (bytes),
`SizeClass` (int64), `Fee.ContributorShareBps` (int64),
`Fee.PlatformFeeBps` (int64), `IssuedAt` (time), `ExpiresAt` (time),
`Seq` (uint64).

El almacenamiento es un ARRENDAMIENTO, no un trabajo: una obligación continuada
de mantener un fragmento sellado durante un plazo acotado. `ShardRef` es una
dirección de contenido opaca de 32 bytes y `SizeClass` un tamaño de carga
cuantizado — el protocolo MUST NOT transportar tamaños reales de objeto ni
contenido almacenado; el cifrado, el relleno y la fragmentación son asuntos del
frontend por encima de esta cintura. La renovación es un nuevo `StorageLease`
para el mismo `LeaseID` con `Seq` estrictamente mayor (misma regla de retroceso
que §4.1). Los términos de tarifa viajan en línea, como en §4.3.

### 4.8 LeaseDecline — firmado por el NODE
Domain tag: `sohocloud/lease-decline/v0`

Order: `LeaseID` (string), `NodeID` (string), `Reason` (string),
`DeclinedAt` (time).

`Reason` ∈ {`local_policy`, `capacity`, `unavailable`}; `local_policy` es la vía
de exclusión (§5.1) para el almacenamiento exactamente como §4.4 lo es para los
trabajos.

### 4.9 LeaseRelease — firmado por el NODE
Domain tag: `sohocloud/lease-release/v0`

Order: `LeaseID` (string), `NodeID` (string), `ReleasedAt` (time).

Un nodo siempre puede dejar de mantener (la soberanía incluye marcharse). La
liberación firmada termina su medición; recolocar el fragmento es asunto del
frontend.

### 4.10 ProofChallenge / ProofResponse — respuesta firmada por el NODE
Domain tag (de la respuesta): `sohocloud/proof/v0`

Un `ProofChallenge{LeaseID, Offset, Length, Nonce (16 bytes), IssuedAt}` NO se
firma: los nodos obtienen desafíos sondeando por el canal autenticado (modelo de
extracción), y un desafío no compromete a nadie a nada. El artefacto firmado es
la respuesta.

Response Order: `LeaseID` (string), `NodeID` (string), `Offset` (int64),
`Length` (int64), `Nonce` (bytes), `Digest` (bytes), `RespondedAt` (time).

`Digest` MUST calcularse exactamente como
`SHA-256(Nonce || uint64be(Offset) || uint64be(Length) || sealed[Offset : Offset+Length])`
sobre los bytes del fragmento sellado. La respuesta reafirma el rango desafiado y
el nonce para sostenerse por sí sola como hecho de medición no repudiable — la
contraparte de almacenamiento de §4.5. Un verificador MUST rechazar una respuesta
cuyo par `(LeaseID, Nonce)` ya haya aceptado: los nonces son de un solo uso, y eso
es lo que derrota la repetición de una respuesta grabada.

### 4.11 KeyRotation — firmado por la clave ACTUAL del nodo
Domain tag: `sohocloud/node-rotate/v0`

Order: `NodeID` (string), `NewPublicKey` (bytes), `Algo` (string),
`RotatedAt` (time), `Seq` (uint64).

La posesión de la clave saliente autoriza a nombrar a su sucesora. Un verificador
MUST validar la firma contra la clave que actualmente tiene para el nodo (una
rotación firmada por la sucesora es autoinstalación y MUST fallar), MUST imponer
un `Seq` estrictamente monótono por nodo, y MUST rechazar una rotación cuya
`NewPublicKey` no sea una clave bien formada para `Algo` (para `ed25519`, no
exactamente 32 bytes) o cuyo `Algo` sea desconocido — de lo contrario, un
verificador que empezara a comprobar mensajes posteriores contra una
`NewPublicKey` malformada fallaría. Solo después de que todo esto pase comienza el
verificador a verificar los mensajes posteriores del nodo contra `NewPublicKey`.
`Algo` ∈ {`ed25519`} en v0.

### 4.12 KeyRevocation — firmado por la clave que se REVOCA
Domain tag: `sohocloud/node-revoke/v0`

Order: `NodeID` (string), `RevokedPublicKey` (bytes), `RevokedAt` (time),
`Seq` (uint64).

Mata una clave sin sucesora. Un verificador MUST verificar la firma contra la
clave en la que **actualmente confía** para `NodeID`, y MUST exigir que
`RevokedPublicKey` sea igual a esa clave de confianza. Una revocación solo se
honra frente a la clave exacta que nombra y mata. Esta es la vinculación
antiabuso: un extraño puede autofirmar una revocación que lleve el `NodeID` de
una víctima y la clave propia del extraño en `RevokedPublicKey`, pero como esa
clave no es la que el coordinador considera de confianza para la víctima, la
comprobación de igualdad falla y la revocación se rechaza — un nodo no puede ser
sacado de línea por nadie que no posea ya su clave actual. (Borradores anteriores
decían "honrar incondicionalmente"; esa redacción queda superada — la vinculación
a la clave de confianza es obligatoria.) Dada la clave actual, la lectura segura
de una revocación es "deja de confiar en esta clave". La reinscripción tras una
revocación es fuera de banda, por la misma vía que la inscripción inicial, nunca
un mensaje por el cable que un ladrón de claves pudiera falsificar.

---

## 5. Invariantes

### 5.1 La exclusión se aplica localmente, nunca en el cable
`CapabilityListing.OptIn` y las pistas de enrutado en `JobSpec` son
**consultivas**. El coordinador NO es una frontera de seguridad para la
exclusión. Un nodo MUST imponer la exclusión contra su propia lista de permitidos
de confianza local y MUST rechazar el trabajo que su política local prohíba — vía
`Decline{Reason: local_policy}` — con independencia de lo que haya anunciado
antes o de lo que el coordinador ofrezca.

### 5.2 La política de emparejamiento no forma parte del protocolo
La interfaz `Coordinator` no expone clasificación ni puntuación. Cómo elige un
coordinador qué nodo recibe qué trabajo es privado de la implementación. El
protocolo MUST NOT hacer crecer un campo que dicte el emparejamiento, porque eso
soldaría a los consumidores a la política de un solo coordinador y derrotaría la
posibilidad de marcharse que la interfaz existe para proteger.

### 5.3 Las tarifas son legibles y no retroactivas
Los términos de tarifa de un coordinador se publican como una `FeeDeclaration`
firmada y con marca de tiempo. Los términos MUST NOT cambiar retroactivamente
para trabajo ya ofrecido; un cambio es una nueva declaración firmada con un
`EffectiveAt` posterior. Esto mantiene las tarifas impugnables en lugar de opacas
(problema abierto #7).

### 5.4 El coordinador es un rol, no un concentrador
Nada de lo que un nodo o un frontend necesiten puede enrutarse *exclusivamente*
por un coordinador específico. Un nodo MAY servir a varios coordinadores; un
frontend MAY ejecutar su propio coordinador o emparejar directamente. Todo
despliegue en el que un único coordinador sea de hecho inevitable es una
recentralización y MUST etiquetarse como tal, con una vía hacia alternativas.

### 5.5 Números de secuencia monótonos
Los valores `Seq` de nodo (listados, latidos) son estrictamente monótonos por
nodo. Un coordinador MUST rechazar un `Seq` no creciente.

---

## 6. Operaciones del coordinador (modelo de extracción)

El modelo es de **extracción**: el coordinador nunca llama al nodo. Operaciones:

| Operación       | Dirección        | Cuerpo / resultado                   |
|-----------------|------------------|--------------------------------------|
| `SubmitListing` | nodo → coord     | `CapabilityListing`                  |
| `Heartbeat`     | nodo → coord     | `Heartbeat`                          |
| `PollJobs`      | nodo → coord     | `NodeID` → `[]Assignment`            |
| `Decline`       | nodo → coord     | `Decline`                            |
| `ReportJob`     | nodo → coord     | `JobReport`                          |
| `Fees`          | cualquiera → coord | → `FeeDeclaration`                 |

Un coordinador MUST aplicar la vinculación de §2 a `PollJobs` (y a cualquier
operación que nombre a un nodo) antes de actuar. Como el coordinador nunca inicia
contacto con un nodo, no hay servicio del lado del nodo que especificar.

---

## 7. Transporte

HTTP+JSON es el transporte de **referencia** (`transport/httpjson`), NO el cable
mismo. Ningún paquete central lo importa. Una implementación conforme MAY usar
cualquier transporte implementando las operaciones del coordinador contra los
bytes canónicos de §3. JSON es solo una codificación de transporte; las firmas
nunca se computan sobre JSON, de modo que un transporte distinto interopera sin
renegociar firmas.

Los endpoints HTTP+JSON de referencia:

```
POST /v0/listing     body: CapabilityListing   -> 204
POST /v0/heartbeat   body: Heartbeat           -> 204
GET  /v0/jobs?node_id=<id>                      -> 200 []Assignment
POST /v0/decline     body: Decline             -> 204
POST /v0/report      body: JobReport           -> 204
GET  /v0/fees                                   -> 200 FeeDeclaration
```

La vinculación SPIFFE (§2) la aplica un middleware delante del manejador; el
manejador de referencia solo conecta rutas y JSON y no es el autenticador.

---

## 8. Problemas abiertos que este protocolo toca

La honestidad sobre las brechas sin resolver es el método, no una vergüenza.

- **#7 El cómputo soberano compra salida mecánica, no económica.** Esta es la capa
  de la que trata el #7. §5.4 (coordinador como rol) y §5.3 (tarifas legibles)
  empujan contra la recentralización del coordinador; el arranque en frío sigue
  siendo un residuo, y la portabilidad de reputación que lo aliviaría es asunto
  del frontend/gobernanza, no de este protocolo.
- **#6 La honestidad del cómputo queda fuera de alcance.** Si el emparejamiento o
  el precio de un coordinador fueron justos no se prueba aquí, solo se hace
  legible (§5.2, §5.3). `anchor/` reserva el lugar para anclar *afirmaciones* de
  empleo (los hechos de `JobReport`) en un registro atestiguado, de modo que al
  menos sean auditables. Es un esbozo etiquetado, no construido.
- **#4 Membresía / incorporación.** Como un nodo se nombra y firma por sí mismo
  (§2), un dispositivo de pocos recursos (p. ej. un teléfono) puede listarse como
  nodo. Si la incorporación de nodos es abierta o restringida es una decisión de
  gobernanza/valores que queda abierta.

## 9. Versionado

`v0` es inestable. Las etiquetas de dominio incrustan `v0`; un cambio
incompatible en los campos canónicos de un mensaje incrementa su etiqueta. Los
consumidores MUST tratar una etiqueta de dominio desconocida como un mensaje no
verificable.

---

## 10. Vectores de prueba

`testdata/vectors.json` es el **fixture normativo de conformidad** de esta
especificación. Una implementación independiente es conforme respecto de la
codificación canónica y la firma si y solo si reproduce, byte a byte, cada valor
de ese archivo a partir de las mismas entradas. El fixture lo generan los
codificadores Go de referencia y se verifica de forma cruzada con una
reproducción independiente que solo usa la biblioteca estándar; una
implementación ajena conforme MUST coincidir exactamente con él. Cuando la prosa
de §3/§4 y el fixture discrepen, eso es un defecto de la especificación que hay
que corregir, no una licencia para divergir — pero el fixture es el enunciado
ejecutable de §3/§4.

Todo el hexadecimal del fixture está en minúsculas. Es autodescriptivo: un `note`
de nivel superior y un bloque `encoding` reexponen las reglas primitivas de §3
para quien abra primero el JSON.

### 10.1 Estructura

- **`primitives`** — codificaciones aisladas de cada primitiva de §3, para que
  quien implemente pueda validar el codificador de abajo arriba antes de ensamblar
  mensajes. Cada entrada tiene un `kind` (`uvarint`, `int64`, `uint64`, `bool`,
  `string`, `time`), un `input` legible por humanos y los `bytes_hex` esperados.
  Cubren los casos límite que fijan el formato: valores varint a través de cada
  umbral de byte de continuación (`0, 1, 127, 128, 255, 256, 300, 16384,
  1073741824`); `int64` incluyendo `-1`, `-2`, `MaxInt64` y `MinInt64`; `uint64`
  incluyendo `MaxUint64`; ambos booleanos; la cadena vacía y una cadena UTF-8
  multibyte; y `time` en la época Unix y en un instante de nanosegundos fijo.

- **`messages`** — una entrada por cada tipo de mensaje firmado (todos los
  mensajes firmados de §4). Cada entrada registra: `name`, `domain_tag`, `signer`
  (`node` o `coordinator`), el `seed_hex` ed25519 de 32 bytes y el
  `public_key_hex` derivado, los `fields` del mensaje, los `canonical_bytes_hex`
  (la salida de `CanonicalBytes()`, con `Signature` excluido) y el
  `signature_hex` (ed25519 sobre exactamente esos bytes canónicos).

- **`operator_messages`** — una entrada por cada tipo de mensaje de operador de la
  Capa C (los tres de §11). Como los mensajes de operador son 2-de-7, cada entrada
  difiere de una entrada central de `messages`: `signer` es siempre `operator`; un
  arreglo `keys` lista las siete claves públicas registradas (`index`, `seed_hex`,
  `public_key_hex`, `algo` — las semillas están presentes SOLO para que el fixture
  sea autocontenido y reproducible, nunca porque la red posea claves privadas);
  `idx0`/`idx1` son los dos índices firmantes; y en lugar de un `signature_hex`
  hay `sig0_hex` y `sig1_hex`, AMBOS tomados sobre los MISMOS
  `canonical_bytes_hex`. Una implementación conforme reproduce
  `canonical_bytes_hex` a partir del orden de campos de §11 y verifica ambas
  firmas en sus índices.

  > **Trampa del orden de campos.** El objeto `fields` de cada mensaje lista las
  > claves en **orden alfabético por legibilidad, NO en orden canónico.** El orden
  > canónico procede ÚNICAMENTE de la cláusula `Order:` de ese mensaje en §4.
  > Quien codifique los campos en el orden de claves del JSON producirá bytes
  > incorrectos y firmas que no verifican.

### 10.2 Cómo usar el fixture

1. Reproduce cada `primitives[].bytes_hex` a partir de su `input` usando §3.
2. Para cada mensaje, deriva el par de claves desde `seed_hex` (ed25519 a partir
   de una semilla de 32 bytes es determinista), confirma que produce
   `public_key_hex`, codifica los campos en el `Order:` de §4 y confirma que
   obtienes `canonical_bytes_hex`; luego verifica `signature_hex` contra
   `public_key_hex` sobre tus bytes derivados de forma independiente. La
   verificación de la firma sobre tus PROPIOS bytes es la verificación cruzada más
   fuerte: solo puede pasar si tus bytes son idénticos a los del generador de
   referencia.

La prueba de referencia vive en `vectors/vectors_test.go`. Regenera cada vector a
partir del código actual y afirma la igualdad byte a byte con el JSON confirmado,
de modo que una divergencia entre código y fixture rompe la construcción; también
vuelve a analizar el archivo confirmado y verifica todas las firmas. Regenera
intencionadamente con `go test ./vectors -update` (o la variable de entorno
`SOHOCLOUD_UPDATE_VECTORS`). Todos los vectores usan semillas de relleno de bytes
fijas y marcas de tiempo codificadas — no hay `crypto/rand` ni `time.Now()`, así
que la salida es plenamente determinista.

### 10.3 Ejemplo resuelto: un `Heartbeat`, byte a byte

Este es el vector `Heartbeat` (`§4.2`) decodificado campo a campo, para que quien
reimplemente pueda seguir toda la codificación de principio a fin. Entradas:
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
| `17 97 9d 09 f8 32 d9 00` | int64, big-endian | `SentAt` como nanosegundos Unix UTC. `0x17979d09f832d900` = `1700000050500000000`. Esta es la regla `time` (§3): el instante reducido a un `int64` y emitido como 8 bytes big-endian. |
| `00 00 00 00 00 00 00 2b` | uint64, big-endian | `Seq` = `0x2b` = **43**. 8 bytes fijos, sin varint. |

La concatenación de exactamente estos seis grupos — etiqueta de dominio (longitud
+ bytes), `NodeID` (longitud + bytes), `SentAt`, `Seq` — es la cadena canónica de
bytes completa. No hay encuadre antes de la etiqueta de dominio ni bytes después
de `Seq`. La firma ed25519 (`signature_hex`) se toma precisamente sobre esa
cadena; el campo `Signature` mismo nunca forma parte de ella.

Obsérvese que el orden de campos aquí (`NodeID`, `SentAt`, `Seq`) sigue el
`Order:` de §4.2, que además es — por coincidencia en este mensaje pequeño — como
un lector los enumeraría; para mensajes mayores como `CapabilityListing`, el
`Order:` de §4 y las claves alfabéticas de `fields` en `vectors.json` difieren, y
solo el orden de §4 es canónico.

---

## 11. Identidad de operador (Capa C)

§1–§10 gobiernan los mensajes de **node** y **coordinator**. Esta sección añade
una tercera identidad, ortogonal: el **operator** — un frontend (p. ej. Cloudy)
que termina dentro de sí la identidad de miembros y nodos y presenta una única
credencial rotatoria a un coordinador. Es deliberadamente distinta de la identidad
de carga de trabajo del nodo (§2): los mensajes centrales de coordinación (§4)
están SIN CAMBIOS y NO se firman con una credencial de operador.

Un operador posee **siete** pares de claves Ed25519 (índices `0..6`) y firma cada
mensaje con **dos índices distintos** (una disciplina **2-de-7**). La red registra
solo las siete claves **públicas** por operador; **la red nunca posee ninguna
clave privada de operador**, y toda vía de verificación de abajo toma únicamente
claves públicas.

- **2-de-7 es antisustitución / higiene de rotación, NO seguridad de umbral.** Si
  las siete claves privadas viven en un mismo proceso, un único compromiso del
  host las entrega todas. No se confía en 2-de-7 para sobrevivir al compromiso de
  una sola clave a menos que la custodia esté repartida entre dominios de
  confianza. Esto se declara para que ningún lector confíe en exceso.
- **Vinculación de algoritmo.** Todo mensaje de operador vincula una cadena de
  algoritmo `Algo` en sus bytes canónicos. `v0` permite exactamente `ed25519`. Un
  verificador MUST rechazar un mensaje cuyo `Algo` sea desconocido, y MUST
  rechazar un mensaje cuyo `Algo` no sea igual al `algo` registrado de la clave en
  cada índice firmante. El conjunto de claves activas de un operador MUST NOT
  mezclar algoritmos. La puerta de longitud de firma es por algoritmo (`ed25519` =
  64 bytes), no una constante fija, de modo que la rotación atómica planificada de
  todo el conjunto a un algoritmo poscuántico (una NUEVA etiqueta de dominio, p.
  ej. `sohocloud/operator/v1-mldsa`) reutiliza la misma regla. Una firma no es
  transferible entre algoritmos.
- **Costura de migración.** La implementación de referencia encamina la firma y la
  verificación de operador a través de una interfaz `Signer`/`Verifier`,
  implementada hoy con Ed25519 de la biblioteca estándar, precisamente para que el
  algoritmo pueda intercambiarse añadiendo después una etiqueta de dominio `v1`
  sin tocar los mensajes centrales. Esto es una nota de implementación; el
  contrato del cable son los órdenes de campos y las etiquetas de abajo.

**Codificación de los campos de operador (sin tratamiento especial).** Todo campo
de las cláusulas `Order:` de §11 se codifica por su regla primitiva ordinaria de
§3; §11 no introduce tipos nuevos ni de ancho fijo. En particular:

- `OperatorID` y `Algo` son **strings** ordinarias de §3 — un unsigned LEB128
  varint de la longitud en bytes UTF-8 seguido de los bytes UTF-8. `Algo` NO es un
  token desnudo ni una etiqueta enumerada de ancho fijo, pese a que `v0` permita
  exactamente el único valor `ed25519`; MUST llevar su prefijo de longitud como
  cualquier otra cadena (p. ej. `ed25519` se codifica como
  `07 65 64 32 35 35 31 39`). `Nonce`, `NewPublicKey` y `Challenge` son **bytes**
  ordinarios de §3 (prefijo de longitud LEB128 y luego los bytes crudos).
- `Idx0`, `Idx1` y `KeyIndex` se codifican cada uno como un **`uint64` big-endian
  de 8 bytes fijos** según la regla `uint64` de §3, exactamente como `Seq` — NO
  como varint y NO como un solo byte — aunque sus valores estén en el pequeño
  rango `0..6`. Quien codifique un índice como varint o como un byte produce bytes
  incorrectos y una firma inválida. (Ejemplo: el índice `2` se codifica como
  `00 00 00 00 00 00 00 02`.)
- Las dos firmas `Sig0` y `Sig1` quedan **excluidas** de los bytes canónicos de
  todo mensaje de §11, exactamente como `Signature` lo está para los mensajes de
  §4 (§3). Son los únicos campos excluidos y, como `Signature`, no aparecen en
  ninguna cláusula `Order:` de abajo. No hay un tercer campo de firma.
  `Idx0`/`Idx1` (que nombran las claves firmantes) SÍ están dentro de los bytes
  firmados; los *bytes* de la firma no.

### 11.0 Requisitos de verificación (todos los mensajes de operador)

Un verificador MUST, antes de honrar cualquier mensaje de operador:

1. Rechazar si el nonce del mensaje está ausente o es más corto de **16 bytes**
   (donde el mensaje tenga nonce). Un nonce vacío MUST NO tratarse NUNCA como
   "omitir la caché de repetición".
2. Rechazar si `Algo` no es un algoritmo admitido para esta versión.
3. Rechazar si los dos índices firmantes `Idx0` e `Idx1` son iguales.
4. Rechazar si algún índice queda fuera de `0..6`.
5. Rechazar si no existe clave registrada activa en alguno de los índices, o si el
   `algo` de alguna clave registrada no es igual al `Algo` del mensaje.
5a. Rechazar si las claves públicas registradas en `Idx0` e `Idx1` son idénticas
   byte a byte, aunque los índices difieran. Las siete claves públicas registradas
   de un operador MUST ser distintas dos a dos; un registro o rotación MUST NOT
   instalar una clave pública que ya exista en otro índice (véase §11.2). Esto se
   impone para que la propiedad antisustitución 2-de-7 no pueda degradarse en
   silencio a 1-de-1: si dos índices tuvieran la misma clave, una sola clave
   privada podría producir ambas firmas. Esta comprobación es sobre las *claves*,
   no meramente sobre los índices (regla 3), y se exige además de ella.
6. Rechazar si la longitud de alguna firma no es la esperada para `Algo`.
7. Recalcular los bytes canónicos y rechazar salvo que AMBAS firmas verifiquen en
   sus índices sobre exactamente esos bytes.

La antirrepetición (un conjunto duradero y de cierre seguro de `Seq` con ventana
deslizante más nonces, con ámbito por `(operator, coordinator)`) es una obligación
del lado del coordinador superpuesta A LO ANTERIOR; no forma parte de los bytes
canónicos y queda fuera del alcance de esta especificación de codificación.

### 11.1 OperatorTransmission — firmado por el OPERATOR (2 de 7)
Domain tag: `sohocloud/operator/v0`

Order: `OperatorID` (string), `Ts` (int64 nanosegundos Unix UTC — el `int64`
crudo, NO rederivado a través de un `time.Time`, de modo que el valor con rango
comprobado sea exactamente el valor firmado), `Nonce` (bytes, `len >= 16`), `Seq`
(uint64), `Algo` (string), `Idx0` (uint64), `Idx1` (uint64).

Las dos firmas `Sig0` (por la clave en `Idx0`) y `Sig1` (por la clave en `Idx1`)
se toman sobre estos bytes canónicos y quedan EXCLUIDAS de ellos, exactamente como
`Signature` en otras partes. `Ts`, `Nonce`, `Seq`, ambos índices y `Algo` están
todos dentro de los bytes firmados.

### 11.2 OperatorRotation — firmado por el OPERATOR (2 de 7)
Domain tag: `sohocloud/operator-rotate/v0`

Order: `OperatorID` (string), `KeyIndex` (uint64 — el índice cuya clave se
reemplaza), `NewPublicKey` (bytes — la clave pública de reemplazo generada por el
operador), `Algo` (string), `Ts` (int64 nanosegundos Unix UTC), `Nonce` (bytes,
`len >= 16`), `Seq` (uint64), `Idx0` (uint64), `Idx1` (uint64).

Esto autoriza a intercambiar una nueva clave pública. El operador (nunca la red)
genera el nuevo par de claves. La nueva clave pública está DENTRO de los bytes
firmados para que ni un intermediario del registro fuera de banda ni un canal
administrativo comprometido puedan inyectar material de claves: un coordinador
MUST verificar este mensaje (con dos claves ACTUALES, `Idx0`/`Idx1`) antes de
registrar la nueva clave en `KeyIndex`. Un verificador MUST además rechazar una
`NewPublicKey` cuya longitud sea incorrecta para `Algo`. Un verificador MUST
también rechazar una `NewPublicKey` idéntica byte a byte a una clave ya registrada
en un índice DISTINTO (violaría el requisito de claves distintas dos a dos de la
regla 5a de §11.0 y prepararía la degradación 2-de-7 → 1-de-1); volver a registrar
la misma clave en su PROPIO `KeyIndex` (una renovación sin efecto) sí está
permitido. Para preservar la invariante de no mezclar algoritmos, un verificador
SHOULD rechazar además una rotación cuyo `Algo` difiera del de las demás claves
registradas.

### 11.3 ConformanceResponse — firmado por el OPERATOR (2 de 7)
Domain tag: `sohocloud/operator-conformance/v0`

Order: `OperatorID` (string), `Challenge` (bytes — el desafío suministrado por el
verificador), `Algo` (string), `Idx0` (uint64), `Idx1` (uint64).

Esta es la respuesta firmada de un operador a un desafío de conformidad. Su
etiqueta de dominio distinta la separa por dominio de §11.1: como la etiqueta
forma parte de los bytes canónicos, una firma producida para un desafío de
conformidad NUNCA puede verificar como `OperatorTransmission`, ni viceversa. No
lleva nonce ni `Seq` propios — la frescura procede del `Challenge` del verificador
— por lo que la regla de nonce de §11.0 no se le aplica; todas las demás
comprobaciones de §11.0 sí.
