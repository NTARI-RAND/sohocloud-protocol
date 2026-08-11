> 社区翻译（草案）—— NTARI 政策 P2-002《全球多语言广播》。来源：README.md（英文原版，2026-07-29 快照）。本文为机器辅助的社区草案，依据 P2-002 §3.1 尚待区域维护者审校。依据 §2.2，核心技术规范仍以英文为准。
>
> 如发现译文有误，欢迎 fork 仓库并提交 Pull Request
> 来改进翻译：https://github.com/NTARI-RAND/sohocloud-protocol。翻译修正与代码贡献同样宝贵，我们诚挚欢迎。

# sohocloud-protocol

本共享 Go 模块定义了 SoHoLINK / Cloudy 网络的**基底协调协议**（substrate coordination protocol）：节点识别、能力列示与作业雇用，由每个协调器和每个前端共同遵循。

## 适用范围

本模块仅管辖**协调**层面。它**不是** JFA 会员经济体系 —— 不含资金托管（escrow）、不含会员发行的信用、不含声誉契约、不含对话封存记录、也不含争议裁决。这些均存在于各前端之中，并被有意地排除在本模块之外。

它是一个**依赖叶节点**（dependency leaf）：除 Go 标准库外，它不导入任何其他 NTARI 模块，也不导入任何别的东西。SoHoLINK（参考协调器）和 Cloudy（一个前端）都导入*本模块*；而没有任何东西通过它被反向导入。这一特性——由导入图（import graph）而非文字约定来强制保证——正是防止任何单一协调器演变为整个网络都必须经由的中心枢纽的关键所在（开放问题 #7）。

## 角色

前端代表其成员贡献的机器，使用节点侧接口 —— `SubmitListing`、`Heartbeat`、`PollJobs`、`Decline`、`ReportJob`、`Fees`；协调器实现 `Coordinator` 接口，负责协调（并联合）各前端。自然人从不出现在线路协议中：成员身份是前端层面的事务，本模块所认知的唯一身份是工作负载身份（workload identity）—— 即 NodeID 及其在 `identity/` 中的规范 SPIFFE 绑定。

## 目录结构

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

## 状态（如实说明）

- **已构建：** 消息类型、规范签名字节（canonical signing bytes）、ed25519 签名/验证、`Coordinator` 接口，以及位于 `transport/httpjson` 下的参考 HTTP+JSON 传输。
- **一致性向量：** `testdata/vectors.json` 是规范性的跨语言一致性测试夹具（fixture）—— 包含 25 个基元用例，以及全部六种带签名的消息类型及其规范字节和 ed25519 签名。当且仅当一个外部实现能够从相同输入逐字节复现该夹具时，才被视为符合一致性要求（`SPEC.md` §10）。`vectors/` 测试会用当前的编码器重新生成该夹具，一旦出现任何漂移即令构建失败。
- **参考传输，而非线路协议本身：** `transport/httpjson` 是可选的，且不被任何核心包导入。符合一致性要求的实现可以（MAY）通过任意传输方式传递这些消息，只需针对 `SPEC.md` 中记载的规范字节实现 `coordinator.Coordinator` 即可。
- **桩代码（stub），尚未构建：** `anchor/` —— 带见证的雇用声明层（witnessed employment-claim layer，开放问题 #6）。已标注为占位实现；参见其包文档。

## 许可证

AGPL-3.0-or-later。`SPEC.md` 是公共资源（commons）的一部分：它的存在就是为了被阅读、被重新实现、被质疑。无人能读的公共资源，就是无人能用的自由。

*Network Theory Applied Research Institute, Inc. — 501(c)(3) — EIN 92-3047136 — info@ntari.org*
