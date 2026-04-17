# Kdoctor 设计文档

## 1. 文档定位

本文档回答四个问题：

1. 这个工具为什么做。
2. 这个工具要解决什么问题。
3. V1 必须做到什么程度才算可用。
4. 后续开发应当围绕什么边界继续推进。

与 [architecture.md](./architecture.md) 的关系如下：

- `doc.md` 负责产品目标、能力范围、输入输出和诊断边界。
- `architecture.md` 负责工程结构、模块职责、依赖规则和实施标准。

## 2. 工具目标

`Kdoctor` 的定位不是 Kafka 管理平台，而是一个可在运维现场直接执行的 Go 二进制诊断工具。

它需要做到：

- 在只有一个 `bootstrap` 地址时，也能快速给出有价值的判断。
- 在存在 `profile`、`compose`、Docker、日志目录时，能够分层增强，而不是改写主流程。
- 把“症状”尽量收敛成“主因判断 + 下一步动作”。
- 输出适合人读，也适合自动化集成。

## 3. V1 要解决的问题

V1 重点覆盖以下五类问题：

- 网络与 listener 问题
- Kafka 元数据、broker 注册、KRaft controller / quorum 问题
- Topic、leader、ISR、副本健康问题
- 真实 client 链路问题：metadata、produce、consume、commit
- 配置错误、部署错误、宿主机 / Docker / 日志侧问题

V1 不追求的平台能力：

- 不做图形化平台
- 不做自动修复
- 不深做 JMX 指标体系
- 不直接接入 K8s 控制面解析
- 不覆盖全部安全协议场景

## 4. 核心设计原则

### 4.1 通用输入优先

`bootstrap` 是最小必需输入，`profile` 和 `compose` 都只是增强输入。

换句话说：

- 没有 `compose` 也必须能查。
- 没有 `profile` 也必须能查。
- 没有 Docker 和日志目录也必须能查。

### 4.2 分层增强，而不是条件分叉

工具输入层应该按能力增强，不应该按部署方式硬分叉。

输入层级如下：

1. `bootstrap-only`
2. `bootstrap + profile`
3. `bootstrap + compose`
4. `bootstrap + compose + log-dir/docker`

### 4.3 证据优先

每个检查项都要输出：

- 状态
- 摘要
- 证据
- 可能原因
- 下一步动作

### 4.4 控制误报

如果当前视角无法可靠判断，就优先：

- `SKIP`
- `WARN`
- 降级说明

而不是在证据不足时直接给出 `CRIT`。

### 4.5 输出要能让一线人员直接使用

默认终端输出必须是中文，且足够“人话化”。

例如不应只写：

```text
under replicated
```

而应尽量表达为：

```text
某些分区 ISR 不足，acks=all 写入可能失败。
```

## 5. 支持的输入模式

### 5.1 `bootstrap-only`

最小可用输入，适合只有一个 Kafka 地址的现场。

在该模式下，V1 至少要支持：

- bootstrap TCP 检查
- metadata 拉取
- broker endpoint 可达性检查
- topic / leader / ISR 检查
- metadata / produce / consume / commit probe

### 5.2 `bootstrap + profile`

适合你知道一些环境事实，但没有完整配置文件的场景。

典型用途：

- 期望 broker 数量
- 期望 controller 端点
- 期望 min ISR
- 场景标签

### 5.3 `bootstrap + compose`

适合对部署结构做静态校验。

V1 在该模式下增强：

- `CFG-001~008`
- controller quorum 配置对照
- listener / advertised.listeners 对照
- 宿主机端口和 Docker 挂载校验

### 5.4 `bootstrap + log-dir/docker`

适合现场排障时进一步收集：

- 容器运行态
- OOMKilled / restart
- 数据目录挂载
- 关键日志指纹

## 6. 检查能力分层

### 6.1 网络层

- `NET-001` bootstrap 可达性
- `NET-002` 显式 listener 可达性
- `NET-003` metadata 返回端点可达性
- `NET-004` DNS / 主机名解析

### 6.2 Kafka 元数据层

- `KFK-001` metadata 拉取
- `KFK-002` broker 注册
- `KFK-003` endpoint 合法性
- `KFK-004` 内部主题健康

### 6.3 KRaft 层

- `KRF-001` quorum 配置一致性
- `KRF-002` active controller 合法性
- `KRF-003` quorum 多数派与视角判断

### 6.4 Topic / Replica 层

- `TOP-003` leader 健康
- `TOP-004` ISR / replica 健康
- `TOP-005` min ISR 风险

### 6.5 Client Probe 层

- `CLI-001` metadata probe
- `CLI-002` producer probe
- `CLI-003` consumer probe
- `CLI-004` commit probe
- `CLI-005` end-to-end probe

### 6.6 Config Lint 层

- `CFG-001~008`

包括：

- node.id
- cluster.id
- process.roles
- controller.quorum.voters
- listeners / advertised.listeners
- inter.broker.listener.name
- replication / ISR 参数

### 6.7 运维增强层

- Host
- Docker
- Logs

这些能力不能阻塞主流程，但应在可用时增强判断准确性。

## 7. 运行模式

### 7.1 `quick`

目标：

- 快速判断是否有明显故障
- 优先跑轻量级检查

### 7.2 `probe`

目标：

- 执行真实业务链路探针
- 作为当前最重要的可用性模式

### 7.3 `lint`

目标：

- 偏静态配置与部署校验

### 7.4 `full`

目标：

- 在上下文足够时尽量跑完整检查

### 7.5 `incident`

目标：

- 输出更聚焦的摘要
- 更强调“主因”和“建议优先动作”

## 8. 输出与退出码

### 8.1 输出格式

V1 支持三种输出：

- 终端文本
- JSON
- Markdown

### 8.2 输出要求

- 默认终端输出必须是中文
- JSON 保持稳定结构，便于自动化
- Markdown 适合发群、贴工单和留档

### 8.3 退出码

- `0`：最高状态为 `PASS`
- `1`：最高状态为 `WARN`
- `2`：最高状态为 `FAIL`
- `3`：最高状态为 `CRIT`
- `5`：最高状态为 `ERROR`
- `6`：最高状态为 `TIMEOUT`

CLI 参数解析或初始化失败时，进程直接退出。

## 9. 报告模型要求

统一报告至少包含：

- 工具版本
- 模式
- profile
- 检查时间
- 耗时
- 汇总状态
- broker 总数与存活数
- 主因判断
- 建议动作
- 逐项检查结果

逐项检查至少包含：

- `id`
- `module`
- `status`
- `summary`
- `evidence`
- `possible_causes`
- `next_actions`

## 10. Probe 设计要求

Probe 是 V1 的核心能力之一，但要控制副作用。

约束如下：

- 仅使用探针主题和探针消费组
- 默认写入最小消息
- 不复用业务消费组
- 失败时清楚标识失败阶段：metadata / produce / consume / commit

## 11. 归因层要求

V1 不能只停留在“列出问题”，至少要做到：

- 把网络、metadata、KRaft、topic、probe、日志信号做基本关联
- 尽量给出 1 到 3 个优先级最高的主因
- 输出建议动作顺序

当前阶段尤其要重点关联：

- `NET-003` 与 `advertised.listeners`
- `KFK-004` 与内部主题 / coordinator
- `KRF-*` 与 controller / quorum
- `TOP-*` 与 ISR / 写入风险
- `CLI-*` 与真实客户端链路

## 12. V1 验收标准

V1 达到“基本可用可信”至少满足：

- `go test ./...` 通过
- 默认终端输出为中文
- `bootstrap-only` 可直接运行
- `compose` 为增强输入而不是前提
- 真实 probe 可跑通 metadata / produce / consume / commit 主流程
- 对视角不足的场景不做明显误报
- JSON / Markdown 输出可正常生成

## 13. 后续路线

V1 之后优先考虑：

1. 容量与趋势判断
2. 更多场景夹具和 golden 输出
3. 更强的调度隔离、timeout 和 degrade 模型
4. 更多 profile 模板
5. 根据真实使用反馈继续收敛误报和提示文案

