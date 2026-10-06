> 社区翻译（草稿）——NTARI 政策 P2-002《全球多语言广播》。来源：SPEC.md（英文原文，2026-10-05 快照）。本文为机器辅助的社区翻译草稿，依照 P2-002 §3.1 待区域维护者审校。根据 §2.2，核心技术规范仍以英文为准。
>
> 如发现译文有误，欢迎 fork 仓库并提交 Pull Request
> 来改进翻译：https://github.com/NTARI-RAND/sohocloud-protocol。翻译修正与代码贡献同样宝贵，我们诚挚欢迎。

> 英文原文具有规范效力。本译文仅供阅读，不作解释之用：凡与 `SPEC.md` 存在分歧之处，以英文为准。规范性关键词（MUST、MUST NOT、SHOULD、SHOULD NOT、MAY）、字段名、标签与标识符均保留英文，因为它们构成符合性契约。

# sohocloud-protocol——符合性规范

版本：**v0 (UNSTABLE)**——在 v1 之前，消息形态与规范字节格式可能发生变化，
不提供兼容性保证。

本文档具有规范性，是一项构建交付物，而非事后补充。除非一个实现能够在不阅读
本模块源码的情况下，依据一份可读的规范构建出来，否则 AGPL 赋予的分叉网络之权
便是空洞的。下文凡标注 MUST / MUST NOT 之处均为符合性要求；违反其中任何一条的
变更并不是本协议的一个缩小版本，而是另一个协议。

---

## 1. 范围

本协议管辖**基质协调层**：节点如何被识别、如何通告自己愿意承担的工作，以及
工作如何提供给它、又如何回报。使用本协议的有三方：

- **Node（节点）**——参与者所拥有的设备，提供计算和/或实体打印。
  节点拥有自己的身份，并对其发出的一切内容签名。它是自主的：它选择为哪些
  协调者服务，并可以离开其中任何一个。
- **Coordinator（协调者）**——将能力列示与工作相匹配，并记录结果。这是一个
  可插拔的角色。SoHoLINK 是参考协调者；前端 MAY 运行自己的协调者，或直接与
  节点进行匹配。
- **Frontend（前端）**——成员进行交易之处。它通过同一接口使用协调服务。

**刻意排除在范围之外。** 本协议**不是** JFA 成员经济。它不定义托管
（escrow）、成员发行的信用、声誉公约、经对话密封的记录，也不定义争议裁断。
这些属于各个前端。协调者与其节点之间进行的法币结算（例如 SoHoLINK 的 Stripe
付款）属于*基质*经济，同样不在本规范范围之内。

---

## 2. 身份与绑定

节点以一个 `NodeID` 命名——这是节点为自己选定的一个稳定字符串。

节点的工作负载身份是一个 SPIFFE SVID，其签发路径严格为：

```
/node/<NodeID>
```

协调者通过将调用方经认证的 SPIFFE 路径与请求中所指名 `NodeID` 对应的
`/node/<NodeID>` 进行比较，来授权某一请求确实来自该节点。

- 当不存在 SPIFFE 身份时，协调者 MUST 以 **401** 拒绝。
- 当身份存在但其路径并非严格为 `/node/<NodeID>` 时，协调者 MUST 以 **403**
  拒绝。
- 协调者 MUST NOT 将“任何有效的 SVID”视为对指名另一节点之请求的授权。那是一
  个冒充途径。
- 绑定 MUST 是确定性的字符串构造，绝不能是在可变存储中逐节点查找。以一个从未
  被填充的字段为门槛的检查是无效的，MUST NOT 被当作安全边界依赖。

本协议不分发公钥。协调者在带外（注册、SVID 材料或某个登记处）解析节点的验证
密钥。

---

## 3. 规范字节编码（用于签名）

每一条签名消息都是对其**规范字节**进行签名，计算规范字节时**排除** `Signature`
字段。签名算法为 **ed25519**。

签名**不**是对 JSON 或任何传输编码进行的。签名是对以下这种确定性的、带长度前
缀的格式进行的，从而使签名在不同平台、不同 Go 版本与不同语言之间保持稳定。

基元，按 §4 中为每条消息列出字段的确切顺序使用：

- **unsigned LEB128 varint（无符号 LEB128 变长整数）**——`string`、`bytes` 与
  `repeated` 所使用的长度/计数基元。它 MUST 按如下方式编码，与语言无关：
  取无符号值；每个字节承载 7 位，最低有效组在前（小端分组）；除最后一个字节外，
  每个字节都设置高位（`0x80`）以表示后续还有字节；最后一个字节的高位清零。值 `0`
  编码为单个字节 `0x00`。这正是 Go 的 `encoding/binary.PutUvarint` 所产生的编码；
  该 Go 符号仅用于交叉核对，而非定义。（示例：`300` = `0b100101100` → 低 7 位
  `0101100` 加续位 → `0xAC`，高位 `0000010` 为最后一字节 → `0x02`，得到
  `AC 02`。）
- **domain tag（域标签）**——每条消息的规范字节都以一个独特的域标签字符串开头
  （见 §4）。域标签 MUST 按下文的 **string** 规则编码——即其 UTF-8 字节长度的无
  符号 LEB128 varint，后接其 UTF-8 字节——与其他任何字符串字段完全相同，没有空
  终止符，也没有特殊处理。各不相同的标签使针对一种消息类型的签名无法转移到另一
  种消息类型上。
- **string（字符串）**——无符号 LEB128 varint 字节长度，后接原始 UTF-8 字节。
  空字符串 MUST 编码为单个字节 `0x00`（长度为零且无载荷）。
- **bytes（字节串）**——无符号 LEB128 varint 字节长度，后接原始字节。（`Signature`
  字段从不被编码；它被排除在外。）
- **bool（布尔值）**——单个字节，严格为 `0x00`（false）或 `0x01`（true）；其他
  任何字节值均无效。
- **int64 / uint64**——固定 8 字节，大端序（最高有效字节在前）。
  `int64` MUST 按其二进制补码的 `uint64` 重新解释进行编码：对负值加上 2⁶⁴，然后
  输出 8 个大端字节。因此 `-1` → `ff ff ff ff ff ff ff ff`，`-2` →
  `ff ff ff ff ff ff ff fe`，`int64` 最小值（`-9223372036854775808`）→
  `80 00 00 00 00 00 00 00`。不使用符号-数值表示，不使用 zigzag，不使用变长。
- **time（时间）**——以 `int64` UTC Unix **纳秒**表示的时刻（依照上文 int64 规
  则，大端序，8 字节）。位置与单调时钟成分被丢弃。纳秒精度被完整保留；该值
  MUST NOT 被截断为微秒或毫秒。这一表示在 JSON 传输往返后依然保持不变，且与时区
  无关。超出 `int64` UTC Unix 纳秒可表示范围（约为 1678–2262 年）的时刻不在 v0
  范围之内，MUST 被视为编码错误，而不是被回绕或钳制到边界值。
  参考编码器强制执行这一点：`canon.Buffer.Time` 不追加任何内容，而是锁存
  `canon.ErrTimeOutOfRange`（可通过 `Buffer.Err()` 取得），而不是输出一个回绕后
  的值，因此两个相隔约 584 年的时刻绝不会映射为相同的字节。生产者 MUST 在签名前
  通过 `canon.ValidTime` 予以拒绝；验证者/解码者 MUST 拒绝时间戳未通过该检查的
  消息。
- **repeated field（重复字段）**——**首先**编码一个无符号 LEB128 varint *计数*
  （元素个数），然后按顺序、完全内联地编码每个元素的字段，使用上述常规基元规则。
  因此，本身为字符串的重复元素与其他字符串一样带长度前缀；嵌套结构体字段按 §4
  中为该元素给出的顺序输出。元素之间没有分隔符，最后一个元素之后也没有终止符。
  计数 `0` 即单个字节 `0x00`，其后没有任何元素。

验证者根据收到的字段重新计算规范字节，并以解析得到的公钥核验 ed25519 签名。
验证者 MUST 在验证之前拒绝长度不等于 `ed25519.SignatureSize`（64）的签名。

一条消息的规范字节**严格**等于域标签，后接该消息在 §4 中 `Order:` 子句所列出的
每一个字段，按该顺序拼接，每个字段均按上述相应基元规则编码。不存在外层封装：
没有前导版本字节，没有包裹整条消息的总长度前缀，最后一个字段之后也没有尾随字节。
§4 中的 `Order:` 列表是穷尽且权威的——消息的**签名字段**是**唯一**被排除的字段，
规范字节中不存在任何未在该处指名的字段。对于 §4 的十二条核心消息，这个唯一被
排除的字段是 `Signature`。§11 的 Layer-C 运营者消息对同一组规范字节携带**两个**
签名（`Sig0` 与 `Sig1`）；二者的排除方式与此处的 `Signature` 完全相同，并且与
`Signature` 一样，二者都不出现在任何 `Order:` 子句中。因此，无论消息带有一个还
是两个签名，验证者在重新计算字节时都绝不会把任何签名字段编码进去。

---

## 4. 消息

对于每条消息：给出其域标签、按规范顺序排列的字段，以及由谁签名。所有时间戳均
遵循 §3 中的 `time` 规则。

下文的 `Order:` 子句是规范字段顺序的**唯一**权威，包括嵌套结构体字段的顺序（例
如 `Capacity.VCPUs` 在 `Capacity.MemMB` 之前，每台打印机先 `Kind` 后 `Model`）。
实现者 MUST 按 `Order:` 序列编码字段，MUST NOT 从任何其他来源推断顺序。特别是，
示例文件 `testdata/vectors.json`（§10）中每条消息的 `fields` 是**为便于阅读而按
字母顺序排列键的，这并非规范编码顺序**——按 JSON 键顺序编码会产生错误的字节和
无效的签名。

### 4.1 CapabilityListing——由 NODE 签名
域标签：`sohocloud/listing/v0`

Order：`NodeID`（string）、`Class`（string）、`Printers`（repeated：`Kind` string、
`Model` string）、`GPUs`（repeated：`API` string、`Model` string、`VRAMMB`
int64）、`Capacity.VCPUs`（int64）、`Capacity.MemMB`（int64）、
`Capacity.DiskMB`（int64）、`Capacity.StorageCommitMB`（int64）、
`Capacity.PrintQPS`（int64）、`OptIn.Compute`（bool）、`OptIn.Print`（bool）、
`OptIn.Storage`（bool）、`IssuedAt`（time）、`Seq`（uint64）。

`Class` ∈ {`micro`, `standard`, `server`}。`Kind` ∈ {`traditional`, `threed`}。
`API` ∈ {`vulkan`, `nnapi`, `cuda`, `metal`}。

`Capacity.DiskMB` 是运行中的作业可用的临时空间；
`Capacity.StorageCommitMB` 是节点承诺为网络长期保存数据的存储空间（分片托管）。
二者被刻意区分，使节点可以只提供其中之一而不提供另一个。存储数据如何加密、分片
和审计是前端/代理层面的事务，不属于本协议——本协议只负责协调，线路上从不承载
所存储的内容。

通告一块 GPU 即表示选择承接 GPU 工作：不含 `GPUs` 的能力列示不会收到任何 GPU
工作；节点若要撤回某块 GPU，只需在下一次能力列示中省略它。

`Seq` 对每个节点严格单调递增。协调者 MUST 拒绝 `Seq` 未严格大于该节点上一次所
见值的能力列示（防重放/回滚保护）。

`OptIn` 是**建议性的**（见 §5.1）。

### 4.2 Heartbeat——由 NODE 签名
域标签：`sohocloud/heartbeat/v0`

Order：`NodeID`（string）、`SentAt`（time）、`Seq`（uint64）。

与 §4.1 相同的严格单调 `Seq` 规则。

### 4.3 Assignment——由 COORDINATOR 签名
域标签：`sohocloud/assignment/v0`

Order：`JobID`（string）、`NodeID`（string）、`Spec.Workload`（string）、
`Spec.Image`（string）、`Spec.Args`（repeated string）、`Spec.PrinterKind`
（string）、`Spec.GPUAPI`（string）、`Spec.GPUMinVRAMMB`（int64）、
`Fee.ContributorShareBps`（int64）、`Fee.PlatformFeeBps`（int64）、
`OfferedAt`（time）。

`Spec.GPUAPI` ∈ {``, `vulkan`, `nnapi`, `cuda`, `metal`}——与 `PrinterKind` 一样
是建议性的路由提示；为空表示不需要 GPU，此时 `GPUMinVRAMMB` MUST 为 `0`。

费用条款以内联方式附带，使节点在作出承诺之前就能看到分成比例。
在拉取模型中，不存在单独的接受消息。

### 4.4 Decline——由 NODE 签名
域标签：`sohocloud/decline/v0`

Order：`JobID`（string）、`NodeID`（string）、`Reason`（string）、`DeclinedAt`
（time）。

`Reason` ∈ {`local_policy`, `capacity`, `unavailable`}。`local_policy` 是选择
退出的途径（见 §5.1）。

### 4.5 JobReport——由 NODE 签名
域标签：`sohocloud/jobreport/v0`

Order：`JobID`（string）、`NodeID`（string）、`ExitCode`（int64）、`FailureCause`
（string）、`TmpfsExhausted`（bool）、`StartedAt`（time）、`FinishedAt`（time）。

这份经签名的报告是计量与付款所依据的事实。其字段与 SoHoLINK 中已上线的执行器
完成状态解析相对应。

### 4.6 FeeDeclaration——由 COORDINATOR 签名
域标签：`sohocloud/fee/v0`

Order：`CoordinatorID`（string）、`Terms.ContributorShareBps`（int64）、
`Terms.PlatformFeeBps`（int64）、`EffectiveAt`（time）、`Seq`（uint64）。

`ContributorShareBps + PlatformFeeBps` SHOULD 等于 `10000`。

### 4.7 StorageLease——由 COORDINATOR 签名
域标签：`sohocloud/lease/v0`

Order：`LeaseID`（string）、`NodeID`（string）、`ShardRef`（bytes）、
`SizeClass`（int64）、`Fee.ContributorShareBps`（int64）、
`Fee.PlatformFeeBps`（int64）、`IssuedAt`（time）、`ExpiresAt`（time）、
`Seq`（uint64）。

存储是一份**租约**，而不是一项作业：即在有限期限内持有一个已密封分片的持续性
义务。`ShardRef` 是一个不透明的 32 字节内容地址，`SizeClass` 是经量化的载荷
大小——本协议 MUST NOT 承载真实的对象大小或所存储的内容；加密、填充与分片是位
于这一“腰部”之上的前端事务。续约是针对同一 `LeaseID`、带有严格更高 `Seq` 的一份
新 `StorageLease`（回滚规则与 §4.1 相同）。费用条款以内联方式附带，与 §4.3 相同。

### 4.8 LeaseDecline——由 NODE 签名
域标签：`sohocloud/lease-decline/v0`

Order：`LeaseID`（string）、`NodeID`（string）、`Reason`（string）、
`DeclinedAt`（time）。

`Reason` ∈ {`local_policy`, `capacity`, `unavailable`}；`local_policy` 是存储的
选择退出途径（§5.1），与 §4.4 之于作业完全相同。

### 4.9 LeaseRelease——由 NODE 签名
域标签：`sohocloud/lease-release/v0`

Order：`LeaseID`（string）、`NodeID`（string）、`ReleasedAt`（time）。

节点始终可以停止持有（自主权包括离开）。经签名的释放消息结束其计量；重新放置
该分片是前端的事务。

### 4.10 ProofChallenge / ProofResponse——响应由 NODE 签名
域标签（响应）：`sohocloud/proof/v0`

`ProofChallenge{LeaseID, Offset, Length, Nonce (16 bytes), IssuedAt}` **不**签名：
节点通过在经认证的通道上轮询来获取挑战（拉取模型），而挑战不使任何人承担任何
义务。经签名的产物是响应。

响应 Order：`LeaseID`（string）、`NodeID`（string）、`Offset`（int64）、
`Length`（int64）、`Nonce`（bytes）、`Digest`（bytes）、`RespondedAt`（time）。

`Digest` MUST 严格按
`SHA-256(Nonce || uint64be(Offset) || uint64be(Length) || sealed[Offset : Offset+Length])`
在已密封的分片字节上计算。响应重述被挑战的范围与随机数，使其能够作为一项不可
抵赖的计量事实独立成立——它是 §4.5 在存储方面的对应物。验证者 MUST 拒绝其已接
受过的 `(LeaseID, Nonce)` 所对应的响应：随机数是一次性的，正是这一点使重放已记
录的答案失效。

### 4.11 KeyRotation——由节点的 CURRENT（当前）密钥签名
域标签：`sohocloud/node-rotate/v0`

Order：`NodeID`（string）、`NewPublicKey`（bytes）、`Algo`（string）、
`RotatedAt`（time）、`Seq`（uint64）。

持有即将退役的密钥，即有权指定其继任者。验证者 MUST 以其当前为该节点持有的密钥
验证签名（由继任密钥签名的轮换属于自我安装，MUST 失败），MUST 对每个节点强制执
行严格单调的 `Seq`，并且 MUST 拒绝 `NewPublicKey` 对 `Algo` 而言格式不正确（对
`ed25519` 而言即不是恰好 32 字节）或 `Algo` 未知的轮换——否则，一个开始以格式错
误的 `NewPublicKey` 检查后续消息的验证者将会出错。只有在以上各项全部通过之后，
验证者才开始以 `NewPublicKey` 验证该节点的后续消息。v0 中 `Algo` ∈ {`ed25519`}。

### 4.12 KeyRevocation——由被 REVOKED（吊销）的密钥签名
域标签：`sohocloud/node-revoke/v0`

Order：`NodeID`（string）、`RevokedPublicKey`（bytes）、`RevokedAt`（time）、
`Seq`（uint64）。

在没有继任者的情况下废止一把密钥。验证者 MUST 以其为 `NodeID` **当前信任**的密
钥验证签名，并且 MUST 要求 `RevokedPublicKey` 等于该受信任密钥。吊销只会针对其所
指名并废止的那把确切密钥生效。这就是防滥用绑定：陌生人可以自签一份吊销消息，其
中携带受害者的 `NodeID`，并在 `RevokedPublicKey` 中填入陌生人自己的密钥；但由于
该密钥并非协调者为受害者所信任的那一把，相等性检查失败，吊销被拒绝——任何尚未
持有节点当前密钥的人，都无法使该节点下线。（早期草稿写的是“无条件生效”；该措辞
已被取代——与受信任密钥的绑定是强制性的。）
在持有当前密钥的前提下，对吊销的稳妥理解是“停止信任这把密钥”。吊销之后的重新
注册在带外进行，经由与首次注册相同的途径，绝不是密钥窃取者可以伪造的线路消息。

---

## 5. 不变式

### 5.1 选择退出在本地强制执行，绝不在线路上
`CapabilityListing.OptIn` 与 `JobSpec` 中的路由提示是**建议性的**。
协调者**不是**选择退出的安全边界。节点 MUST 依据自己本地受信任的允许列表强制执
行选择退出，并且 MUST 拒绝其本地策略所禁止的工作——通过
`Decline{Reason: local_policy}`——无论它此前通告过什么，也无论协调者提供什么。

### 5.2 匹配策略不属于本协议
`Coordinator` 接口不暴露任何排序或评分。协调者如何选择由哪个节点承接哪项作业，
由实现私有决定。本协议 MUST NOT 增设任何规定匹配方式的字段，因为那会把使用方焊
死在某一个协调者的策略上，从而破坏该接口存在所要保护的可离开性。

### 5.3 费用可理解且不溯及既往
协调者的费用条款以经签名、带时间戳的 `FeeDeclaration` 形式发布。对于已经提供的
工作，条款 MUST NOT 溯及既往地变更；变更即是一份带有更晚 `EffectiveAt` 的新签名
声明。这使费用保持可质疑，而非不透明（开放问题 #7）。

### 5.4 协调者是一个角色，而非一个枢纽
节点或前端所需的任何东西都不得*只能*经由某一个特定的协调者路由。节点 MAY 为多
个协调者服务；前端 MAY 运行自己的协调者或直接匹配。任何事实上无法避开单一协调
者的部署都是一种再中心化，MUST 被如实标注为再中心化，并给出通往替代方案的路径。

### 5.5 单调序列号
节点的 `Seq` 值（能力列示、心跳）对每个节点严格单调递增。协调者 MUST 拒绝非递增
的 `Seq`。

---

## 6. 协调者操作（拉取模型）

模型为**拉取**：协调者从不调用节点。操作如下：

| 操作            | 方向             | 消息体 / 结果                        |
|-----------------|------------------|--------------------------------------|
| `SubmitListing` | node → coord     | `CapabilityListing`                  |
| `Heartbeat`     | node → coord     | `Heartbeat`                          |
| `PollJobs`      | node → coord     | `NodeID` → `[]Assignment`            |
| `Decline`       | node → coord     | `Decline`                            |
| `ReportJob`     | node → coord     | `JobReport`                          |
| `Fees`          | anyone → coord   | → `FeeDeclaration`                   |

协调者 MUST 在采取行动之前，对 `PollJobs`（以及任何指名某个节点的操作）应用 §2
的绑定。由于协调者从不主动联系节点，因此没有需要规定的节点侧服务。

---

## 7. 传输

HTTP+JSON 是**参考**传输（`transport/httpjson`），**不是**线路本身。没有任何核心
包导入它。符合规范的实现 MAY 使用任何传输方式，只要它针对 §3 中的规范字节实现协
调者操作即可。JSON 仅是一种传输编码；签名从不在 JSON 上计算，因此不同的传输方式
无需重新协商签名即可互操作。

参考 HTTP+JSON 端点：

```
POST /v0/listing     body: CapabilityListing   -> 204
POST /v0/heartbeat   body: Heartbeat           -> 204
GET  /v0/jobs?node_id=<id>                      -> 200 []Assignment
POST /v0/decline     body: Decline             -> 204
POST /v0/report      body: JobReport           -> 204
GET  /v0/fees                                   -> 200 FeeDeclaration
```

SPIFFE 绑定（§2）由处理程序前方的中间件施加；参考处理程序只负责连接路由与
JSON，并不是认证者。

---

## 8. 本协议涉及的开放问题

如实面对尚未解决的缺口是方法本身，而不是一种难堪。

- **#7 主权计算换来的是机械意义上的退出，而非经济意义上的退出。** 开放问题 #7
  所涉及的正是这一层。§5.4（协调者作为角色）与 §5.3（可理解的费用）抵制协调者的
  再中心化；冷启动仍是一项残余问题，而能够缓解它的声誉可移植性属于前端/治理层面
  的事务，而非本协议的事务。
- **#6 计算诚实性不在范围之内。** 协调者的匹配或价格是否公平，在此并不予以证明，
  只是使之可理解（§5.2、§5.3）。`anchor/` 为将雇佣*主张*（`JobReport` 中的事实）
  锚定到经见证的日志中预留了位置，使其至少可被审计。它是一个已标注的桩，尚未构建。
- **#4 成员资格 / 入驻。** 由于节点为自己命名并签名（§2），一台低资源设备（例如
  手机）也可以将自己列示为节点。节点入驻是开放的还是设有门槛，是一项留待决定的
  治理/价值取向问题。

## 9. 版本管理

`v0` 是不稳定的。域标签中嵌入了 `v0`；对某条消息规范字段的破坏性变更会提升其标签
版本。使用方 MUST 将未知的域标签视为无法验证的消息。

---

## 10. 测试向量

`testdata/vectors.json` 是本规范的**规范性符合性测试夹具**。
一个独立实现在规范编码与签名方面是符合规范的，当且仅当它能够从相同输入逐字节
复现该文件中的每一个值。该夹具由参考 Go 编码器生成，并经过一个仅使用标准库的
独立复现程序交叉核对；符合规范的外部实现 MUST 与之完全一致。若 §3/§4 中的文字与
该夹具出现任何不一致，那是需要修正的规范缺陷，而不是可以偏离的许可——但该夹具
是 §3/§4 的可执行表述。

夹具中的所有十六进制均为小写。它是自描述的：顶层的 `note` 与 `encoding` 块为首先
打开 JSON 的读者重述了 §3 的基元规则。

### 10.1 结构

- **`primitives`**——§3 每个基元的独立编码，使实现者能够在组装消息之前自底向上
  验证编码器。每个条目包含一个 `kind`（`uvarint`、`int64`、`uint64`、`bool`、
  `string`、`time`）、一个人类可读的 `input`，以及预期的 `bytes_hex`。这些条目涵盖
  了确定格式的边界情况：跨越每个续位字节阈值的 varint 值
  （`0, 1, 127, 128, 255, 256, 300, 16384, 1073741824`）；`int64`，包括 `-1`、
  `-2`、`MaxInt64` 和 `MinInt64`；`uint64`，包括 `MaxUint64`；两个布尔值；空字符串
  与一个多字节 UTF-8 字符串；以及位于 Unix 纪元和某个固定纳秒时刻的 `time`。

- **`messages`**——每种签名消息类型一个条目（§4 的每一条签名消息）。每个条目
  记录：`name`、`domain_tag`、`signer`（`node` 或 `coordinator`）、32 字节的
  ed25519 `seed_hex` 及其派生的 `public_key_hex`、消息 `fields`、
  `canonical_bytes_hex`（`CanonicalBytes()` 的输出，排除 `Signature`），
  以及 `signature_hex`（恰好对这些规范字节进行的 ed25519 签名）。

- **`operator_messages`**——每种 Layer-C 运营者消息类型一个条目（§11 的全部
  三种）。由于运营者消息采用 2-of-7，每个条目与核心 `messages` 条目有所不同：
  `signer` 始终为 `operator`；`keys` 数组列出全部七把已登记的公钥（`index`、
  `seed_hex`、`public_key_hex`、`algo`——种子之所以出现，**仅仅**是为了使该夹具
  自成一体、可以复现，绝不是因为网络持有私钥）；`idx0`/`idx1` 是两个签名索引；
  并且没有单一的 `signature_hex`，而是有 `sig0_hex` 与 `sig1_hex`，**二者**都针对
  **同一个** `canonical_bytes_hex`。符合规范的实现依据 §11 的字段顺序复现
  `canonical_bytes_hex`，并在各自的索引上验证两个签名。

  > **字段顺序陷阱。** 每条消息的 `fields` 对象**为便于阅读而按字母顺序列出键，
  > 并非规范顺序。** 规范顺序**只**来自 §4 中该消息的 `Order:` 子句。按 JSON 键
  > 顺序编码字段的实现者将产生错误的字节，以及无法通过验证的签名。

### 10.2 如何使用该夹具

1. 使用 §3，根据每个 `primitives[].bytes_hex` 的 `input` 复现该值。
2. 对每条消息，从 `seed_hex` 派生密钥对（由 32 字节种子派生 ed25519 密钥是确定性
   的），确认得到 `public_key_hex`，按 §4 `Order:` 编码各字段并确认得到
   `canonical_bytes_hex`，然后在你独立推导出的字节上以 `public_key_hex` 验证
   `signature_hex`。
   在你**自己**的字节上进行签名验证是最强的交叉核对：只有当你的字节与参考生成器
   的字节完全相同时，它才可能通过。

参考测试位于 `vectors/vectors_test.go`。它以当前代码重新生成每一个向量，并断言其
与已提交的 JSON 逐字节相等，因此代码与夹具之间的漂移会使构建失败；它还会重新解析
已提交的文件并验证每一个签名。如需有意重新生成，请使用 `go test ./vectors -update`
（或 `SOHOCLOUD_UPDATE_VECTORS` 环境变量）。所有向量都使用固定字节填充的种子和
硬编码的时间戳——没有 `crypto/rand`，也没有 `time.Now()`，因此输出完全确定。

### 10.3 详解示例：逐字节解析一条 `Heartbeat`

这是 `Heartbeat` 向量（`§4.2`）逐字段的解码，使重新实现者能够从头到尾追踪整个
编码过程。输入：
`NodeID = "node-alpha"`，`SentAt = ` Unix 纳秒 `1700000050500000000`，
`Seq = 43`。

规范字节（`canonical_bytes_hex`）：

```
16 736f686f636c6f75642f6865617274626561742f7630 0a 6e6f64652d616c706861 17979d09f832d900 000000000000002b
```

从左到右解码：

| 字节 | 基元 | 含义 |
|-------|-----------|---------|
| `16` | uvarint 长度 | 域标签长度 = `0x16` = **22** 字节。 |
| `73 6f 68 6f 63 6c 6f 75 64 2f 68 65 61 72 74 62 65 61 74 2f 76 30` | UTF-8 字节 | 域标签 `sohocloud/heartbeat/v0`（22 字节）。这两行连同长度前缀，构成域标签的 `string` 编码（§3）。 |
| `0a` | uvarint 长度 | `NodeID` 长度 = `0x0a` = **10** 字节。 |
| `6e 6f 64 65 2d 61 6c 70 68 61` | UTF-8 字节 | `NodeID` = `node-alpha`（10 字节）。 |
| `17 97 9d 09 f8 32 d9 00` | int64，大端序 | `SentAt`，以 UTC Unix 纳秒表示。`0x17979d09f832d900` = `1700000050500000000`。这就是 `time` 规则（§3）：将时刻化为一个 `int64`，并以 8 个大端字节输出。 |
| `00 00 00 00 00 00 00 2b` | uint64，大端序 | `Seq` = `0x2b` = **43**。固定 8 字节，不使用 varint。 |

恰好这六组的拼接——域标签（长度 + 字节）、`NodeID`（长度 + 字节）、`SentAt`、
`Seq`——就是完整的规范字节串。域标签之前没有封装，`Seq` 之后也没有任何字节。
ed25519 签名（`signature_hex`）恰好针对这一字节串计算；`Signature` 字段本身从不
是其中的一部分。

请注意，这里的字段顺序（`NodeID`、`SentAt`、`Seq`）遵循 §4.2 的 `Order:`，而对这
条小消息来说，这也恰巧是读者自然会列出的顺序；对于 `CapabilityListing` 等较大的
消息，§4 的 `Order:` 与 `vectors.json` 中按字母顺序排列的 `fields` 键并不相同，
只有 §4 的顺序才是规范的。

---

## 11. 运营者身份（Layer C）

§1–§10 管辖**节点**与**协调者**消息。本节增加第三种相互正交的身份：**运营者**
（operator）——一个前端（例如 Cloudy），它在自身内部终结成员身份与节点身份，并
向协调者出示单一的、可轮换的凭证。它被刻意与节点工作负载身份（§2）分开：
核心协调消息（§4）**保持不变**，并且**不**使用运营者凭证签名。

运营者持有**七**对 Ed25519 密钥对（索引 `0..6`），并以**两个不同的**索引为每条
消息签名（**2-of-7** 规则）。网络为每个运营者只登记七把**公钥**；**网络从不持有任
何运营者私钥**，下文的每一条验证路径都只接受公钥。

- **2-of-7 是防替换 / 轮换卫生措施，并非门限安全。** 如果全部七把私钥都存放在
  同一个进程中，一次主机入侵就会泄露全部七把。除非密钥保管被分散到多个信任域，
  否则不能依赖 2-of-7 来抵御单一密钥泄露。此处明确说明，以免任何读者过度信任它。
- **算法绑定。** 每条运营者消息都将一个算法字符串 `Algo` 绑定进其规范字节。
  `v0` 只允许 `ed25519`。验证者 MUST 拒绝 `Algo` 未知的消息，并且 MUST 拒绝
  `Algo` 不等于各签名索引上密钥所登记之 `algo` 的消息。
  运营者的活动密钥集合 MUST NOT 混用算法。签名长度门槛按算法而定
  （`ed25519` = 64 字节），而不是一个硬编码常量，因此计划中向后量子算法进行的整
  套原子轮换（使用一个**新的**域标签，例如 `sohocloud/operator/v1-mldsa`）可以沿用
  同一规则。签名不可跨算法转移。
- **迁移接缝。** 参考实现将运营者签名与验证经由 `Signer`/`Verifier` 接口路由，
  目前以标准库 Ed25519 实现，正是为了日后只需新增一个 `v1` 域标签即可替换算法，
  而无需触动核心消息。这是一条实现说明；线路契约是下文的字段顺序与标签。

**运营者字段的编码（无特殊处理）。** §11 各 `Order:` 子句中的每个字段都按其普通
的 §3 基元规则编码；§11 不引入任何新的或固定宽度的类型。具体而言：

- `OperatorID` 与 `Algo` 是普通的 §3 **字符串**——UTF-8 字节长度的无符号 LEB128
  varint，后接 UTF-8 字节。尽管 `v0` 只允许 `ed25519` 这一个值，`Algo` 也**不是**
  一个裸令牌或固定宽度的枚举标签；它 MUST 像其他任何字符串一样携带长度前缀（例如
  `ed25519` 编码为 `07 65 64 32 35 35 31 39`）。`Nonce`、`NewPublicKey` 与
  `Challenge` 是普通的 §3 **字节串**（LEB128 长度前缀，后接原始字节）。
- `Idx0`、`Idx1` 与 `KeyIndex` 均按 §3 的 `uint64` 规则编码为**固定 8 字节大端
  `uint64`**，与 `Seq` 完全相同——**不是** varint，也**不是**单个字节——尽管它们
  的取值位于很小的范围 `0..6` 之内。将索引编码为 varint 或单个字节的实现者会产生
  错误的字节和无效的签名。（示例：索引 `2` 编码为
  `00 00 00 00 00 00 00 02`。）
- 两个签名 `Sig0` 与 `Sig1` 被**排除**在每条 §11 消息的规范字节之外，与 §4 消息中
  的 `Signature` 完全相同（§3）。它们是唯一被排除的字段，并且与 `Signature` 一样，
  不出现在下文任何 `Order:` 子句中。不存在第三个签名字段。`Idx0`/`Idx1`（用于指名
  签名密钥）**位于**被签名的字节之内；签名*字节*则不在其中。

### 11.0 验证要求（所有运营者消息）

在采信任何运营者消息之前，验证者 MUST：

1. 若消息的随机数缺失或短于 **16 字节**（对于带有随机数的消息），予以拒绝。空随机
   数 MUST NEVER 被视为“跳过重放缓存”。
2. 若 `Algo` 不是本版本所支持的算法，予以拒绝。
3. 若两个签名索引 `Idx0` 与 `Idx1` 相等，予以拒绝。
4. 若任一索引超出 `0..6`，予以拒绝。
5. 若任一索引上不存在活动的已登记密钥，或任一已登记密钥的 `algo` 不等于消息的
   `Algo`，予以拒绝。
5a. 若 `Idx0` 与 `Idx1` 上已登记的公钥逐字节相同，即使索引不同，也予以拒绝。运营者
   的七把已登记公钥 MUST 两两不同；登记或轮换 MUST NOT 安装一把已存在于另一个索引
   上的公钥（见 §11.2）。强制执行这一点，是为了使 2-of-7 防替换属性不会悄然退化为
   1-of-1：如果两个索引持有同一把密钥，单独一把私钥就能产生两个签名。此项检查针对
   的是*密钥*，而不仅仅是索引（规则 3），并且是在规则 3 之外额外要求的。
6. 若任一签名的长度不是 `Algo` 的预期长度，予以拒绝。
7. 重新计算规范字节，除非**两个**签名都在各自索引上针对恰好这些字节验证通过，
   否则予以拒绝。

防重放（一个持久的、失败即关闭的滑动窗口 `Seq` + 随机数集合，按
`(operator, coordinator)` 划定范围）是叠加在上述要求**之上**的协调者侧义务；它不是
规范字节的一部分，不在本编码规范的范围之内。

### 11.1 OperatorTransmission——由 OPERATOR 签名（2 of 7）
域标签：`sohocloud/operator/v0`

Order：`OperatorID`（string）、`Ts`（int64 UTC Unix 纳秒——原始 `int64`，**不**经由
`time.Time` 重新推导，从而使经过范围检查的值与被签名的值完全相同）、`Nonce`（bytes，
`len >= 16`）、`Seq`（uint64）、`Algo`（string）、`Idx0`（uint64）、`Idx1`（uint64）。

两个签名 `Sig0`（由 `Idx0` 上的密钥签署）与 `Sig1`（由 `Idx1` 上的密钥签署）针对这些
规范字节计算，并被**排除**在其之外，与其他地方的 `Signature` 完全相同。`Ts`、
`Nonce`、`Seq`、两个索引以及 `Algo` 全都位于被签名的字节之内。

### 11.2 OperatorRotation——由 OPERATOR 签名（2 of 7）
域标签：`sohocloud/operator-rotate/v0`

Order：`OperatorID`（string）、`KeyIndex`（uint64——其密钥正被替换的索引）、
`NewPublicKey`（bytes——由运营者生成的替换公钥）、`Algo`（string）、`Ts`（int64 UTC
Unix 纳秒）、`Nonce`（bytes，`len >= 16`）、`Seq`（uint64）、`Idx0`（uint64）、
`Idx1`（uint64）。

此消息授权换入一把新公钥。新密钥对由运营者（绝不是网络）生成。新公钥**位于**被签
名的字节之内，因此无论是带外注册过程中的中间人，还是遭入侵的管理通道，都无法注入
密钥材料：协调者 MUST 先（以两把**当前**密钥 `Idx0`/`Idx1`）验证此消息，然后才能
在 `KeyIndex` 上记录新密钥。验证者 MUST 另外拒绝长度对 `Algo` 而言不正确的
`NewPublicKey`。验证者还 MUST 拒绝与已登记在**另一个**索引上的密钥逐字节相同的
`NewPublicKey`（这将违反 §11.0 规则 5a 的两两不同密钥要求，并为 2-of-7 → 1-of-1 的
退化埋下伏笔）；在其**自身** `KeyIndex` 上重新登记同一把密钥（无操作的刷新）是允许
的。为维护不混用算法的不变式，验证者 SHOULD 同时拒绝 `Algo` 与其他已登记密钥不同的
轮换。

### 11.3 ConformanceResponse——由 OPERATOR 签名（2 of 7）
域标签：`sohocloud/operator-conformance/v0`

Order：`OperatorID`（string）、`Challenge`（bytes——由验证者提供的挑战）、
`Algo`（string）、`Idx0`（uint64）、`Idx1`（uint64）。

这是运营者对符合性挑战作出的签名回答。其独特的域标签使它与 §11.1 实现域分隔：由于
标签是规范字节的一部分，为符合性挑战产生的签名**绝不**可能被验证为
`OperatorTransmission`，反之亦然。它不携带自己的随机数/`Seq`——新鲜性来自验证者的
`Challenge`——因此 §11.0 的随机数规则不适用于它；§11.0 的其他所有检查均适用。
