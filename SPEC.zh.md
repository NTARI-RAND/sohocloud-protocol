> 社区翻译（草案）—— NTARI 政策 P2-002《全球多语言广播》。来源：SPEC.md（英文原版，2026-09-16 快照）。本文为机器辅助的社区草案，依据 P2-002 §3.1 尚待区域维护者审校。依据 §2.2，核心技术规范仍以英文为准。
>
> **英文原件具有规范效力。** 本译文供阅读之用，不供解释之用：凡与 `SPEC.md`
> 有出入者，以英文为准。以下内容因构成一致性契约而保留英文：规范性关键词
> （MUST、MUST NOT、SHOULD、MAY）、字段名、域标签（domain tag）、类型标识符，
> 以及全部代码块。
>
> 如发现译文有误，欢迎 fork 仓库并提交 Pull Request
> 来改进翻译：https://github.com/NTARI-RAND/sohocloud-protocol。翻译修正与代码贡献同样宝贵。

# sohocloud-protocol —— 一致性规范

版本：**v0（不稳定）** —— 在 v1 之前，消息形态与规范字节格式可能变更，且不提供
兼容性保证。

本文件具有规范效力，是一项构建交付物，而非事后补充。若无法凭一份可读的规范、
不读本模块源码即构建出实现，则 AGPL 赋予的分叉网络之权利便是空的。下文凡标记
MUST / MUST NOT 者，皆为一致性要求；违反其一的变更不是本协议的缩减版本，而是
另一个协议。

---

## 1. 范围

本协议管辖**基底协调层**：节点如何被识别，如何公告它将承接何种工作，以及工作
如何被提议给它并回报。三方使用本协议：

- **Node** —— 参与者自有的设备，提供算力和/或实体打印。节点拥有自己的身份，并
  对其发出的一切进行签名。主权性：它选择服务哪些协调器，并可离开其中任何一个。
- **Coordinator** —— 将列示与工作撮合，并记录结果。这是一个可插拔的角色。
  SoHoLINK 是参考协调器；frontend MAY 运行自有协调器，或直接与节点撮合。
- **Frontend** —— 成员进行交易之处。它通过同一接口消费协调能力。

**刻意排除在范围之外。** 本协议**不是** JFA 的成员经济。它不定义托管（escrow）、
不定义成员发行的信用、不定义声誉盟约、不定义对话封存的记录，也不定义争议裁断。
那些属于各个 frontend。协调器与其节点之间运行的法币结算（例如 SoHoLINK 的
Stripe 付款）属于**基底**经济，同样不在本规范之内。

---

## 2. 身份与绑定

节点以 `NodeID` 命名 —— 一个由节点自行选定的稳定字符串。

节点的工作负载身份是一份恰好签发于以下路径之下的 SPIFFE SVID：

```
/node/<NodeID>
```

协调器通过将调用方已认证的 SPIFFE 路径与请求中所指 `NodeID` 对应的
`/node/<NodeID>` 相比较，来授权某个请求确实来自该节点。

- 当不存在 SPIFFE 身份时，协调器 MUST 以 **401** 拒绝。
- 当身份存在但其路径并非恰好为 `/node/<NodeID>` 时，协调器 MUST 以 **403** 拒绝。
- 对于指名另一个节点的请求，协调器 MUST NOT 接受「任何有效 SVID」作为授权。那是
  一条冒充路径。
- 绑定 MUST 是确定性的字符串构造，绝不可是在可变存储中按节点查表。以一个从不被
  填充的字段为条件的检查是惰性的，MUST NOT 依赖它作为安全边界。

本协议不分发公钥。协调器在带外（注册、SVID 材料，或某个登记处）解析节点的验证
密钥。

---

## 3. 规范字节编码（用于签名）

每一条签名消息都在其**规范字节**上签名，计算时**排除** `Signature` 字段。签名为
**ed25519**。

签名**不**在 JSON 或任何传输编码之上进行，而是在下述确定性的、带长度前缀的格式
之上进行，从而使签名在不同平台、不同 Go 版本与不同语言之间保持稳定。

原语，按 §4 中每条消息所列字段的确切顺序：

- **unsigned LEB128 varint** —— 由 `string`、`bytes` 与 `repeated` 使用的长度/
  计数原语。MUST 按如下方式编码，且与语言无关：取无符号值；每字节输出 7 位，
  最低有效组在前（小端分组）；除最后一个字节外，每个字节的高位（`0x80`）置位以
  标记续接；最后一个字节高位清零。值 `0` 编码为单字节 `0x00`。这正是 Go 的
  `encoding/binary.PutUvarint` 所产生的编码；Go 符号只是交叉校验，不是定义。
  （演算示例：`300` = `0b100101100` → 低 7 位 `0101100` 加续接位 → `0xAC`，高位
  `0000010` 置于最后 → `0x02`，得 `AC 02`。）
- **domain tag** —— 每条消息的规范字节都以一个各不相同的域标签字符串开头（见
  §4）。域标签 MUST 按下文 **string** 规则编码 —— 即其 UTF-8 字节长度的 unsigned
  LEB128 varint，其后接其 UTF-8 字节 —— 与任何其他字符串字段完全一样，没有空终止
  符，也没有特例处理。不同的标签使得针对某一消息类型的签名无法转用于另一类型。
- **string** —— 字节长度的 unsigned LEB128 varint，其后为原始 UTF-8 字节。空字符
  串 MUST 编码为单字节 `0x00`（长度为零且无载荷）。
- **bytes** —— 字节长度的 unsigned LEB128 varint，其后为原始字节。（`Signature`
  字段从不被编码；它被排除。）
- **bool** —— 单个字节，恰为 `0x00`（假）或 `0x01`（真）；其他任何字节值均无效。
- **int64 / uint64** —— 固定 8 字节，大端（最高有效字节在前）。`int64` MUST 编码
  为其二进制补码的 `uint64` 重解释：对负值加 2⁶⁴，然后输出 8 个大端字节。于是
  `-1` → `ff ff ff ff ff ff ff ff`，`-2` → `ff ff ff ff ff ff ff fe`，而最小的
  `int64`（`-9223372036854775808`）→ `80 00 00 00 00 00 00 00`。不用符号-数值表示，
  不用 zigzag，不用变长。
- **time** —— 将该瞬间表示为 UTC Unix **纳秒**的 `int64`（依上述 int64 规则，大端，
  8 字节）。时区与单调时钟分量被丢弃。纳秒精度被完整保留；该值 MUST NOT 被截断到
  微秒或毫秒。这可经受 JSON 传输的往返，且与时区无关。超出 `int64` UTC Unix 纳秒
  可表示范围（约公元 1678–2262 年）的瞬间不在 v0 范围之内，MUST 按编码错误处理，
  而不得回绕或截断。参考编码器强制这一点：`canon.Buffer.Time` 不追加任何内容，而是
  锁存 `canon.ErrTimeOutOfRange`（可经 `Buffer.Err()` 取回），而非输出回绕值，
  从而使相距约 584 年的两个瞬间绝不会混同为相同字节。生产方 MUST 在签名前经
  `canon.ValidTime` 拒绝；验证方/解码方 MUST 拒绝时间戳未通过该检查的消息。
- **repeated field** —— 先编码一个表示*计数*（元素数量）的 unsigned LEB128
  varint，然后按顺序、完全内联地编码每个元素的各字段，使用上述常规原语规则。
  因此，本身是字符串的重复元素同样带长度前缀；嵌套结构字段按 §4 中为该元素给出的
  顺序输出。元素之间没有分隔符，最后一个元素之后也没有终止符。计数为 `0` 即单字节
  `0x00`，其后没有任何元素。

验证方根据收到的字段重新计算规范字节，并针对解析出的公钥校验 ed25519 签名。验证
方 MUST 在验证之前拒绝长度不等于 `ed25519.SignatureSize`（64）的签名。

一条消息的规范字节**恰好**是：域标签，后接 §4 中该消息 `Order:` 子句所列的每个
字段，按该顺序，各自依其原语规则编码，二者的串接。没有任何外层封装：开头没有版本
字节，没有包裹整条消息的总长度前缀，最后一个字段之后也没有尾随字节。§4 的
`Order:` 列表是穷尽且权威的 —— 消息的**签名字段**是**唯一**被排除的字段，规范字节
中不存在任何未在其中列名的字段。对 §4 的十二条核心消息而言，那唯一被排除的字段是
`Signature`。§11 的 Layer-C 操作方消息在同一批规范字节上携带**两个**签名（`Sig0`
与 `Sig1`）；二者被排除的方式与此处的 `Signature` 完全相同，且与 `Signature` 一样，
都不出现在任何 `Order:` 子句中。因此，无论消息带一个签名还是两个，验证方在其重新
计算的字节中从不编码任何签名字段。

---

## 4. 消息

每条消息给出：其域标签、按规范顺序排列的字段，以及由谁签名。所有时间戳遵循 §3 的
`time` 规则。

下文的 `Order:` 子句是规范字段顺序的**唯一**权威，包括嵌套结构字段的顺序（例如
`Capacity.VCPUs` 先于 `Capacity.MemMB`，每台打印机为 `Kind` 后接 `Model`）。实现者
MUST 按 `Order:` 序列编码字段，且 MUST NOT 从任何其他来源推断顺序。特别地，示例
`testdata/vectors.json`（§10）为便于阅读，将每条消息的 `fields` 按**键名字母序**
列出，**那并不是规范编码顺序** —— 若照 JSON 键序编码，将产生错误的字节和无效的
签名。

### 4.1 CapabilityListing —— 由 NODE 签名
Domain tag：`sohocloud/listing/v0`

Order: `NodeID` (string), `Class` (string), `Printers` (repeated: `Kind` string,
`Model` string), `GPUs` (repeated: `API` string, `Model` string, `VRAMMB`
int64), `Capacity.VCPUs` (int64), `Capacity.MemMB` (int64),
`Capacity.DiskMB` (int64), `Capacity.StorageCommitMB` (int64),
`Capacity.PrintQPS` (int64), `OptIn.Compute` (bool), `OptIn.Print` (bool),
`OptIn.Storage` (bool), `IssuedAt` (time), `Seq` (uint64).

`Class` ∈ {`micro`, `standard`, `server`}。`Kind` ∈ {`traditional`, `threed`}。
`API` ∈ {`vulkan`, `nnapi`, `cuda`, `metal`}。

`Capacity.DiskMB` 是可供运行中作业使用的临时空间；`Capacity.StorageCommitMB` 是
节点承诺为网络保有的长期存储（分片托管）。二者刻意区分，使节点可以只提供其中之一。
存储数据如何加密、分片与审计，属于 frontend/代理的事务，不在本协议之内 —— 本协议
只管协调；线路上从不承载被存储的内容。

公告 GPU 即为对 GPU 工作的选择加入：不含 `GPUs` 的列示不会收到任何 GPU 工作，
节点在下一次列示中省略某块 GPU 即为撤回它。

`Seq` 按节点严格单调。协调器 MUST 拒绝 `Seq` 未严格大于该节点上一次所见值的列示
（重放/回滚保护）。

`OptIn` 是**建议性的**（见 §5.1）。

### 4.2 Heartbeat —— 由 NODE 签名
Domain tag：`sohocloud/heartbeat/v0`

Order: `NodeID` (string), `SentAt` (time), `Seq` (uint64).

与 §4.1 相同的严格单调 `Seq` 纪律。

### 4.3 Assignment —— 由 COORDINATOR 签名
Domain tag：`sohocloud/assignment/v0`

Order: `JobID` (string), `NodeID` (string), `Spec.Workload` (string),
`Spec.Image` (string), `Spec.Args` (repeated string), `Spec.PrinterKind`
(string), `Spec.GPUAPI` (string), `Spec.GPUMinVRAMMB` (int64),
`Fee.ContributorShareBps` (int64), `Fee.PlatformFeeBps` (int64),
`OfferedAt` (time).

`Spec.GPUAPI` ∈ {``, `vulkan`, `nnapi`, `cuda`, `metal`} —— 与 `PrinterKind` 同类
的建议性路由提示；为空表示不需要 GPU，此时 `GPUMinVRAMMB` MUST 为 `0`。

费用条款内联附上，使节点在承诺之前即可看到分成。在拉取模型中没有单独的接受消息。

### 4.4 Decline —— 由 NODE 签名
Domain tag：`sohocloud/decline/v0`

Order: `JobID` (string), `NodeID` (string), `Reason` (string), `DeclinedAt`
(time).

`Reason` ∈ {`local_policy`, `capacity`, `unavailable`}。`local_policy` 是选择退出
的路径（见 §5.1）。

### 4.5 JobReport —— 由 NODE 签名
Domain tag：`sohocloud/jobreport/v0`

Order: `JobID` (string), `NodeID` (string), `ExitCode` (int64), `FailureCause`
(string), `TmpfsExhausted` (bool), `StartedAt` (time), `FinishedAt` (time).

这份签名报告是计量与结算所依据的事实。其字段对应 SoHoLINK 中已上线的执行器完成
解析。

### 4.6 FeeDeclaration —— 由 COORDINATOR 签名
Domain tag：`sohocloud/fee/v0`

Order: `CoordinatorID` (string), `Terms.ContributorShareBps` (int64),
`Terms.PlatformFeeBps` (int64), `EffectiveAt` (time), `Seq` (uint64).

`ContributorShareBps + PlatformFeeBps` SHOULD 等于 `10000`。

### 4.7 StorageLease —— 由 COORDINATOR 签名
Domain tag：`sohocloud/lease/v0`

Order: `LeaseID` (string), `NodeID` (string), `ShardRef` (bytes),
`SizeClass` (int64), `Fee.ContributorShareBps` (int64),
`Fee.PlatformFeeBps` (int64), `IssuedAt` (time), `ExpiresAt` (time),
`Seq` (uint64).

存储是一项**租约**，而非一项作业：在有限期限内持有一个封存分片的持续义务。
`ShardRef` 是 32 字节的不透明内容地址，`SizeClass` 是量化后的载荷规模 —— 本协议
MUST NOT 承载真实对象大小或被存储的内容；加密、填充与分片是这条腰线之上 frontend
的事务。续期是为同一 `LeaseID` 签发、`Seq` 严格更大的新 `StorageLease`（与 §4.1
相同的回滚规则）。费用条款如 §4.3 那样内联随行。

### 4.8 LeaseDecline —— 由 NODE 签名
Domain tag：`sohocloud/lease-decline/v0`

Order: `LeaseID` (string), `NodeID` (string), `Reason` (string),
`DeclinedAt` (time).

`Reason` ∈ {`local_policy`, `capacity`, `unavailable`}；`local_policy` 之于存储，
正如 §4.4 之于作业，是选择退出的路径（§5.1）。

### 4.9 LeaseRelease —— 由 NODE 签名
Domain tag：`sohocloud/lease-release/v0`

Order: `LeaseID` (string), `NodeID` (string), `ReleasedAt` (time).

节点随时可以停止持有（主权包含离开）。这份签名的释放终止其计量；重新安置该分片是
frontend 的事务。

### 4.10 ProofChallenge / ProofResponse —— 响应由 NODE 签名
Domain tag（响应的）：`sohocloud/proof/v0`

`ProofChallenge{LeaseID, Offset, Length, Nonce (16 bytes), IssuedAt}` **不**签名：
节点通过在已认证通道上轮询来取得挑战（拉取模型），而挑战不使任何人承担任何义务。
被签名的制品是响应。

Response Order: `LeaseID` (string), `NodeID` (string), `Offset` (int64),
`Length` (int64), `Nonce` (bytes), `Digest` (bytes), `RespondedAt` (time).

`Digest` MUST 恰按下式在封存分片字节上计算：
`SHA-256(Nonce || uint64be(Offset) || uint64be(Length) || sealed[Offset : Offset+Length])`
响应重述被挑战的区间与 nonce，使其能够独立成立为不可抵赖的计量事实 —— 这是 §4.5
在存储侧的对应物。验证方 MUST 拒绝其 `(LeaseID, Nonce)` 已被接受过的响应：nonce
一次性使用，正是这一点挫败了对已录得答案的重放。

### 4.11 KeyRotation —— 由节点的**当前**密钥签名
Domain tag：`sohocloud/node-rotate/v0`

Order: `NodeID` (string), `NewPublicKey` (bytes), `Algo` (string),
`RotatedAt` (time), `Seq` (uint64).

持有即将退出的密钥，即授权指定其后继。验证方 MUST 针对它当前为该节点持有的密钥来
校验签名（由后继密钥签名的轮换属于自我安装，MUST 失败），MUST 按节点强制严格单调
的 `Seq`，并且 MUST 拒绝这样的轮换：其 `NewPublicKey` 对 `Algo` 而言不是良构密钥
（对 `ed25519` 而言即非恰好 32 字节），或其 `Algo` 未知 —— 否则，开始针对一个畸形
`NewPublicKey` 校验后续消息的验证方将会出错。唯有以上全部通过之后，验证方才开始
针对 `NewPublicKey` 验证该节点的后续消息。v0 中 `Algo` ∈ {`ed25519`}。

### 4.12 KeyRevocation —— 由**被吊销**的那把密钥签名
Domain tag：`sohocloud/node-revoke/v0`

Order: `NodeID` (string), `RevokedPublicKey` (bytes), `RevokedAt` (time),
`Seq` (uint64).

在没有后继的情况下终止一把密钥。验证方 MUST 针对它当前**为 `NodeID` 所信任**的
密钥校验签名，并且 MUST 要求 `RevokedPublicKey` 等于那把受信密钥。吊销永远只对它
所指名并终止的那把密钥生效。这就是反滥用绑定：陌生人可以自签一份吊销，其中携带
受害者的 `NodeID` 以及陌生人自己的密钥作为 `RevokedPublicKey`，但由于那把密钥并非
协调器为受害者所信任的密钥，相等性检查失败，吊销被拒绝 —— 不持有节点当前密钥的人
无法把该节点打下线。（早期草案曾写「无条件遵从」；该措辞已被取代 —— 与受信密钥的
绑定是强制的。）在给定当前密钥的前提下，对吊销的安全解读是「停止信任这把密钥」。
吊销之后的重新登记在带外进行，走与首次登记相同的路径，绝不是密钥窃取者能够伪造的
线上消息。

---

## 5. 不变式

### 5.1 选择退出在本地强制，绝不在线路上
`CapabilityListing.OptIn` 以及 `JobSpec` 中的路由提示都是**建议性的**。协调器
**不是**选择退出的安全边界。节点 MUST 依据自己本地受信的允许清单强制执行选择退出，
并且 MUST 拒绝其本地策略所禁止的工作 —— 经由 `Decline{Reason: local_policy}` ——
无论它先前公告过什么，也无论协调器提供什么。

### 5.2 撮合策略不属于本协议
`Coordinator` 接口不暴露排序或打分。协调器如何选择由哪个节点承接哪项作业，属于
实现私有。本协议 MUST NOT 增设支配撮合的字段，因为那会把消费方焊死在某一个协调器
的策略上，并挫败该接口本就是为保护而存在的可离开性。

### 5.3 费用可读且不溯及既往
协调器的费用条款以签名且带时间戳的 `FeeDeclaration` 发布。对已提议的工作，条款
MUST NOT 溯及既往地更改；变更即是一份带更晚 `EffectiveAt` 的新签名声明。这使费用
保持可争议，而非不透明（开放问题 #7）。

### 5.4 协调器是一个角色，不是一个枢纽
节点或 frontend 所需要的任何东西，都不得**排他地**取道某一特定协调器。节点 MAY
服务多个协调器；frontend MAY 运行自有协调器或直接撮合。任何使单一协调器事实上不可
绕开的部署都是再集中化，MUST 如实标注，并给出通往替代方案的路径。

### 5.5 单调的序列号
节点的 `Seq` 值（列示、心跳）按节点严格单调。协调器 MUST 拒绝非递增的 `Seq`。

---

## 6. 协调器操作（拉取模型）

模型是**拉取**：协调器从不呼叫节点。操作如下：

| 操作            | 方向              | 主体 / 结果                          |
|-----------------|-------------------|--------------------------------------|
| `SubmitListing` | 节点 → 协调器     | `CapabilityListing`                  |
| `Heartbeat`     | 节点 → 协调器     | `Heartbeat`                          |
| `PollJobs`      | 节点 → 协调器     | `NodeID` → `[]Assignment`            |
| `Decline`       | 节点 → 协调器     | `Decline`                            |
| `ReportJob`     | 节点 → 协调器     | `JobReport`                          |
| `Fees`          | 任何人 → 协调器   | → `FeeDeclaration`                   |

协调器 MUST 在行动之前，对 `PollJobs`（以及任何指名某节点的操作）施加 §2 的绑定。
由于协调器从不主动联系节点，因此不存在需要规定的节点侧服务。

---

## 7. 传输

HTTP+JSON 是**参考**传输（`transport/httpjson`），**不是**线路本身。没有任何核心
包导入它。符合本规范的实现 MAY 使用任何传输，只需针对 §3 的规范字节实现协调器
操作即可。JSON 只是一种传输编码；签名从不在 JSON 之上计算，因此换用另一种传输也能
互通，而无需重新协商签名。

参考 HTTP+JSON 端点：

```
POST /v0/listing     body: CapabilityListing   -> 204
POST /v0/heartbeat   body: Heartbeat           -> 204
GET  /v0/jobs?node_id=<id>                      -> 200 []Assignment
POST /v0/decline     body: Decline             -> 204
POST /v0/report      body: JobReport           -> 204
GET  /v0/fees                                   -> 200 FeeDeclaration
```

SPIFFE 绑定（§2）由处理器前置的中间件施加；参考处理器只负责接线路由与 JSON，
并不是认证者。

---

## 8. 本协议触及的开放问题

对未解缺口保持诚实是方法，不是难堪。

- **#7 主权算力买到的是机械性退出，而非经济性退出。** #7 讲的正是这一层。§5.4
  （协调器作为角色）与 §5.3（费用可读）抵抗协调器的再集中化；冷启动仍是残余问题，
  而能缓解它的声誉可携性属于 frontend/治理的事务，不属于本协议。
- **#6 计算诚实性不在范围之内。** 协调器的撮合或定价是否公平，在此并不被证明，
  只是被变得可读（§5.2、§5.3）。`anchor/` 为把雇用*主张*（`JobReport` 的事实）
  锚定到受见证日志中预留了位置，以使其至少可审计。那是一个带标注的存根，尚未构建。
- **#4 成员资格 / 接纳。** 由于节点自行命名并为自己签名（§2），低资源设备（例如
  手机）也能把自己列为节点。节点接纳应当开放还是设限，是一项被留待决定的治理/
  价值问题。

## 9. 版本化

`v0` 不稳定。域标签内嵌 `v0`；对某条消息规范字段的破坏性变更会提升其标签。消费方
MUST 将未知域标签视为不可验证的消息。

---

## 10. 测试向量

`testdata/vectors.json` 是本规范的**规范性一致性夹具**。当且仅当一个独立实现能够
从相同输入逐字节复现该文件中的每一个值时，它在规范编码与签名方面才是一致的。该夹具
由参考 Go 编码器生成，并由一份仅依赖标准库的独立复现进行交叉校验；符合规范的外部
实现 MUST 与之完全一致。凡 §3/§4 的叙述与该夹具出现分歧，即为应予修正的规范缺陷，
而不是背离的许可 —— 但夹具是 §3/§4 的可执行陈述。

夹具中的十六进制一律小写。它是自描述的：顶层的 `note` 与一个 `encoding` 块为先打开
JSON 的读者重述 §3 的原语规则。

### 10.1 结构

- **`primitives`** —— §3 各原语的孤立编码，使实现者能在拼装消息之前自底向上验证
  编码器。每个条目含一个 `kind`（`uvarint`、`int64`、`uint64`、`bool`、`string`、
  `time`）、一个可读的 `input`，以及期望的 `bytes_hex`。它们覆盖了钉死该格式的边界
  情形：跨越每个续接字节阈值的 varint 值（`0, 1, 127, 128, 255, 256, 300, 16384,
  1073741824`）；包含 `-1`、`-2`、`MaxInt64` 与 `MinInt64` 的 `int64`；包含
  `MaxUint64` 的 `uint64`；两个布尔值；空字符串与一个多字节 UTF-8 字符串；以及位于
  Unix 纪元和某个固定纳秒瞬间的 `time`。

- **`messages`** —— 每种签名消息类型各一条（§4 的全部签名消息）。每个条目记录：
  `name`、`domain_tag`、`signer`（`node` 或 `coordinator`）、32 字节的 ed25519
  `seed_hex` 及其派生的 `public_key_hex`、消息的 `fields`、`canonical_bytes_hex`
  （`CanonicalBytes()` 的输出，已排除 `Signature`），以及 `signature_hex`
  （在恰好那些规范字节上的 ed25519 签名）。

- **`operator_messages`** —— 每种 Layer-C 操作方消息类型各一条（§11 的全部三种）。
  由于操作方消息是 2-of-7，每个条目都与核心的 `messages` 条目不同：`signer` 恒为
  `operator`；一个 `keys` 数组列出全部七把已注册公钥（`index`、`seed_hex`、
  `public_key_hex`、`algo` —— 其中的种子**仅仅**是为了让夹具自洽且可复现，绝不是
  因为网络持有私钥）；`idx0`/`idx1` 是两个签名索引；并且不是一个 `signature_hex`，
  而是 `sig0_hex` 与 `sig1_hex`，**两者**都取自**同一批** `canonical_bytes_hex`。
  符合规范的实现应由 §11 的字段顺序复现 `canonical_bytes_hex`，并在各自索引上验证
  这两个签名。

  > **字段顺序的陷阱。** 每条消息的 `fields` 对象为便于阅读，按**字母序**列出键，
  > **而非规范顺序**。规范顺序**只**来自 §4 中该消息的 `Order:` 子句。若实现者按
  > JSON 键序编码字段，将产生错误的字节和无法通过验证的签名。

### 10.2 如何使用夹具

1. 依 §3，从每个 `primitives[].bytes_hex` 的 `input` 复现其字节。
2. 对每条消息，由 `seed_hex` 派生密钥对（由 32 字节种子派生 ed25519 是确定性的），
   确认它得出 `public_key_hex`；按 §4 的 `Order:` 编码字段，确认你得到
   `canonical_bytes_hex`；然后在你独立推导出的字节之上，针对 `public_key_hex`
   验证 `signature_hex`。在**你自己**的字节上做签名验证是最强的交叉校验：只有当
   你的字节与参考生成器的完全相同时，它才可能通过。

参考测试位于 `vectors/vectors_test.go`。它从当前代码重新生成每个向量，并断言与已
提交 JSON 的逐字节相等，因此代码与夹具之间的漂移会使构建失败；它还会重新解析已提交
的文件并验证全部签名。有意重新生成时使用 `go test ./vectors -update`（或环境变量
`SOHOCLOUD_UPDATE_VECTORS`）。所有向量都使用固定的字节填充种子与硬编码时间戳 ——
既没有 `crypto/rand` 也没有 `time.Now()`，因此输出完全确定。

### 10.3 演算示例：逐字节解一条 `Heartbeat`

这是 `Heartbeat` 向量（`§4.2`）按字段逐一解码的结果，以便重新实现者能够从头到尾
追踪整个编码。输入：`NodeID = "node-alpha"`，`SentAt = ` Unix 纳秒
`1700000050500000000`，`Seq = 43`。

规范字节（`canonical_bytes_hex`）：

```
16 736f686f636c6f75642f6865617274626561742f7630 0a 6e6f64652d616c706861 17979d09f832d900 000000000000002b
```

自左向右解码：

| 字节 | 原语 | 含义 |
|-------|-----------|---------|
| `16` | uvarint 长度 | 域标签长度 = `0x16` = **22** 字节。 |
| `73 6f 68 6f 63 6c 6f 75 64 2f 68 65 61 72 74 62 65 61 74 2f 76 30` | UTF-8 字节 | 域标签 `sohocloud/heartbeat/v0`（22 字节）。连同长度前缀，这两行即域标签的 `string` 编码（§3）。 |
| `0a` | uvarint 长度 | `NodeID` 长度 = `0x0a` = **10** 字节。 |
| `6e 6f 64 65 2d 61 6c 70 68 61` | UTF-8 字节 | `NodeID` = `node-alpha`（10 字节）。 |
| `17 97 9d 09 f8 32 d9 00` | int64，大端 | `SentAt`，以 UTC Unix 纳秒计。`0x17979d09f832d900` = `1700000050500000000`。这就是 `time` 规则（§3）：该瞬间被归约为 `int64` 并以 8 个大端字节输出。 |
| `00 00 00 00 00 00 00 2b` | uint64，大端 | `Seq` = `0x2b` = **43**。固定 8 字节，不用 varint。 |

恰好这六组的串接 —— 域标签（长度 + 字节）、`NodeID`（长度 + 字节）、`SentAt`、
`Seq` —— 就是完整的规范字节串。域标签之前没有封装，`Seq` 之后也没有字节。ed25519
签名（`signature_hex`）正是在这一串之上取得的；`Signature` 字段本身从不属于它。

请注意，此处的字段顺序（`NodeID`、`SentAt`、`Seq`）遵循 §4.2 的 `Order:`，而在这条
小消息中它恰好也是读者会采用的罗列顺序；对于 `CapabilityListing` 这类较大的消息，
§4 的 `Order:` 与 `vectors.json` 中按字母序排列的 `fields` 键并不相同，只有 §4 的
顺序才是规范的。

---

## 11. 操作方身份（Layer C）

§1–§10 管辖 **node** 与 **coordinator** 消息。本节增加第三种、彼此正交的身份：
**operator** —— 一个 frontend（例如 Cloudy），它在自身内部终结成员与节点身份，
并向协调器出示单一的轮换凭证。它与节点的工作负载身份（§2）刻意分离：核心协调消息
（§4）**保持不变**，且**不**以操作方凭证签名。

一个操作方持有**七**对 Ed25519 密钥（索引 `0..6`），并以**两个不同的**索引为每条
消息签名（一种 **2-of-7** 纪律）。网络每个操作方只登记那七把**公**钥；**网络从不
持有任何操作方私钥**，且下文每一条验证路径都只取公钥。

- **2-of-7 是反替换 / 轮换卫生，不是门限安全。** 若七把私钥都存活在同一个进程中，
  单一主机被攻破即交出全部七把。除非保管分散在不同信任域，否则不应指望 2-of-7 能
  在单钥泄露下幸存。此处明说，是为了不让任何读者过度信任它。
- **算法绑定。** 每条操作方消息都把算法字符串 `Algo` 绑入其规范字节。`v0` 恰好只
  允许 `ed25519`。验证方 MUST 拒绝 `Algo` 未知的消息，并且 MUST 拒绝其 `Algo` 不
  等于各签名索引处已登记密钥之 `algo` 的消息。操作方的活跃密钥集合 MUST NOT 混用
  算法。签名长度闸门按算法而定（`ed25519` = 64 字节），而非硬编码常量，因此计划中
  向后量子算法的整集原子轮换（采用一个**新的**域标签，例如
  `sohocloud/operator/v1-mldsa`）可以复用同一规则。签名在不同算法之间不可转用。
- **迁移接缝。** 参考实现把操作方的签名与验证都路由经过一个 `Signer`/`Verifier`
  接口，今天以标准库 Ed25519 实现，正是为了日后只需添加一个 `v1` 域标签即可替换
  算法，而不必触动核心消息。这是实现说明；线路契约是下文的字段顺序与标签。

**操作方字段的编码（没有特例）。** §11 各 `Order:` 子句中的每个字段，都按 §3 的
普通原语规则编码；§11 不引入任何新类型或定宽类型。特别地：

- `OperatorID` 与 `Algo` 是 §3 的普通 **string** —— UTF-8 字节长度的 unsigned
  LEB128 varint，其后为 UTF-8 字节。尽管 `v0` 恰好只允许 `ed25519` 这一个值，
  `Algo` 仍**不是**裸记号，也不是定宽枚举标签；它 MUST 像任何其他字符串一样带上
  自己的长度前缀（例如 `ed25519` 编码为 `07 65 64 32 35 35 31 39`）。`Nonce`、
  `NewPublicKey` 与 `Challenge` 都是 §3 的普通 **bytes**（LEB128 长度前缀后接
  原始字节）。
- `Idx0`、`Idx1` 与 `KeyIndex` 各自按 §3 的 `uint64` 规则编码为**固定 8 字节的
  大端 `uint64`**，与 `Seq` 完全一样 —— **不是** varint，也**不是**单个字节 ——
  尽管它们的取值落在 `0..6` 这一小范围内。把索引编码成 varint 或一个字节的实现者
  将产生错误的字节和无效的签名。（示例：索引 `2` 编码为
  `00 00 00 00 00 00 00 02`。）
- 两个签名 `Sig0` 与 `Sig1` 被**排除**在 §11 每条消息的规范字节之外，与 §4 消息中
  `Signature` 的做法完全相同（§3）。它们是仅有的被排除字段，并且与 `Signature`
  一样，不出现在下文任何 `Order:` 子句中。没有第三个签名字段。`Idx0`/`Idx1`
  （指名签名密钥）**位于**被签名的字节之内；签名*字节本身*则不在其中。

### 11.0 验证要求（所有操作方消息）

在遵从任何操作方消息之前，验证方 MUST：

1. 若消息的 nonce 缺失或短于 **16 字节**（在消息带有 nonce 之处），则拒绝。空
   nonce MUST 永远不被当作「跳过重放缓存」。
2. 若 `Algo` 不是本版本支持的算法，则拒绝。
3. 若两个签名索引 `Idx0` 与 `Idx1` 相等，则拒绝。
4. 若任一索引落在 `0..6` 之外，则拒绝。
5. 若任一索引处不存在活跃的已登记密钥，或任一已登记密钥的 `algo` 不等于消息的
   `Algo`，则拒绝。
5a. 若 `Idx0` 与 `Idx1` 处已登记的公钥逐字节相同，则即便索引不同也要拒绝。一个
   操作方的七把已登记公钥 MUST 两两互异；登记或轮换 MUST NOT 安装一把已存在于另一
   索引处的公钥（见 §11.2）。强制这一点，是为了使 2-of-7 的反替换性质不会悄然退化
   为 1-of-1：若两个索引持有同一把密钥，则单独一把私钥即可产生两个签名。这项检查
   针对的是*密钥*，而不仅是索引（规则 3），并且是在规则 3 之外额外要求的。
6. 若任一签名的长度不是 `Algo` 所期望的长度，则拒绝。
7. 重新计算规范字节，除非**两个**签名都在各自索引上、恰在那些字节之上通过验证，
   否则拒绝。

反重放（一个按 `(operator, coordinator)` 划分作用域的、持久且失败即关闭的滑动窗口
`Seq` 加 nonce 集合）是叠加在**以上之上**的协调器侧义务；它不属于规范字节，也不在
本编码规范的范围之内。

### 11.1 OperatorTransmission —— 由 OPERATOR 签名（7 取 2）
Domain tag：`sohocloud/operator/v0`

Order: `OperatorID` (string), `Ts` (int64 UTC Unix 纳秒 —— 原始的 `int64`，
**不**经由 `time.Time` 重新推导，从而使经过范围检查的值恰好就是被签名的值),
`Nonce` (bytes, `len >= 16`), `Seq` (uint64), `Algo` (string), `Idx0` (uint64),
`Idx1` (uint64).

两个签名 `Sig0`（由 `Idx0` 处的密钥产生）与 `Sig1`（由 `Idx1` 处的密钥产生）都取自
这些规范字节，并被**排除**在其外，与别处的 `Signature` 完全一样。`Ts`、`Nonce`、
`Seq`、两个索引以及 `Algo`，全都在被签名的字节之内。

### 11.2 OperatorRotation —— 由 OPERATOR 签名（7 取 2）
Domain tag：`sohocloud/operator-rotate/v0`

Order: `OperatorID` (string), `KeyIndex` (uint64 —— 其密钥正被替换的那个索引),
`NewPublicKey` (bytes —— 由操作方生成的替换公钥), `Algo` (string), `Ts` (int64
UTC Unix 纳秒), `Nonce` (bytes, `len >= 16`), `Seq` (uint64), `Idx0` (uint64),
`Idx1` (uint64).

这授权换入一把新的公钥。新密钥对由操作方（绝非网络）生成。新公钥位于被签名字节
**之内**，从而使带外登记环节的中间人与被攻破的管理通道都无法注入密钥材料：协调器
MUST 在把新密钥记入 `KeyIndex` 之前验证本消息（使用两把**当前**密钥，
`Idx0`/`Idx1`）。验证方 MUST 另行拒绝长度对 `Algo` 而言不正确的 `NewPublicKey`。
验证方 MUST 还要拒绝与**另一**索引处已登记密钥逐字节相同的 `NewPublicKey`（那会
违反 §11.0 规则 5a 的两两互异要求，并埋下 2-of-7 → 1-of-1 退化的伏笔）；把同一把
密钥重新登记在它**自己**的 `KeyIndex` 上（一次无操作的刷新）则是允许的。为保持
不混用算法这一不变式，验证方 SHOULD 同时拒绝其 `Algo` 与其余已登记密钥不同的轮换。

### 11.3 ConformanceResponse —— 由 OPERATOR 签名（7 取 2）
Domain tag：`sohocloud/operator-conformance/v0`

Order: `OperatorID` (string), `Challenge` (bytes —— 由验证方提供的挑战), `Algo`
(string), `Idx0` (uint64), `Idx1` (uint64).

这是操作方对一致性挑战作出的签名回答。它各不相同的域标签使它与 §11.1 实现域分隔：
因为标签是规范字节的一部分，为一致性挑战所产生的签名**绝不**可能作为
`OperatorTransmission` 通过验证，反之亦然。它不携带自己的 nonce 或 `Seq` ——
新鲜性来自验证方的 `Challenge` —— 因此 §11.0 的 nonce 规则不适用于它；§11.0 的
其余检查则全部适用。
