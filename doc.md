# Kdoctor 设计文档

## 1. 文档定位

本文档说明 `Kdoctor` 在 **V2 封版状态** 下的产品边界、输入方式、输出目标与诊断原则。

与 [architecture.md](./architecture.md) 的分工：

- `doc.md` 回答“为什么做、做什么、边界在哪”
- `architecture.md` 回答“代码怎么组织、模块怎么协作、工程标准是什么”

## 2. 工具目标

`Kdoctor` 的目标不是替代 Kafka 平台，而是成为一线现场可直接执行的 Go 二进制诊断工具。

它需要做到：

- 只有一个 `bootstrap` 地址时也能给出有价值判断
- 在有 `profile / compose / docker / logs` 时逐层增强证据
- 尽量把零散症状收敛成“主因判断 + 建议动作”
- 同时支持人工阅读与自动化消费

## 3. 当前边界

### 3.1 这版保留

- 网络与 listener 路径
- Kafka metadata / broker / internal topics
- KRaft controller / quorum
- Topic / ISR / leader / 规划
- Client probe
- Compose lint
- Docker / Host / Logs
- Producer / Consumer / Transaction 上下文

### 3.2 这版不做

- 图形化平台
- 自动修复
- 默认扩展指标链路
- K8s 控制面解析
- 平台化资产管理

## 4. 核心设计原则

### 4.1 最小输入可运行

`bootstrap` 是最小输入。

没有 `compose`、没有 `profile`、没有 Docker、没有日志目录时，也必须能跑出最小可用结果。

### 4.2 分层增强

输入层级按能力增强，而不是按部署方式硬分叉：

1. `bootstrap-only`
2. `bootstrap + profile`
3. `bootstrap + compose`
4. `bootstrap + compose + docker/logs/host`

### 4.3 证据优先

每条结论尽量落到：

- 状态
- 摘要
- 证据
- 风险解释
- 下一步

### 4.4 减少误导

当当前视角不能可靠判断时，应优先输出：

- `SKIP`
- `WARN`
- `无可用证据`

而不是在证据不足时直接抬成严重故障。

### 4.5 面向值班人员

默认终端报告必须满足：

- 中文
- 短
- 先看主因和动作
- 默认不刷 `PASS / SKIP`

## 5. 输入模式

### 5.1 `bootstrap-only`

最小可用模式，至少支持：

- bootstrap TCP 检查
- metadata 拉取
- broker endpoint 可达性检查
- Topic / leader / ISR 检查
- `metadata / produce / consume / commit / e2e` 探针

### 5.2 `bootstrap + profile`

用于补充环境预期，例如：

- broker 数
- controller 端点
- min ISR
- replication factor
- producer / consumer / transaction 上下文

### 5.3 `bootstrap + compose`

用于部署结构与配置对照，例如：

- `listeners`
- `advertised.listeners`
- `controller.quorum.voters`
- `node.id`
- `process.roles`
- `inter.broker.listener.name`

### 5.4 `bootstrap + docker/logs/host`

用于增强运行态证据，例如：

- 容器存在 / 运行 / OOMKilled / mount
- 宿主机磁盘 / 端口 / FD / 内存
- 日志指纹与上下文

## 6. 运行模式

- `quick`
  快速巡检
- `probe`
  真实链路探针
- `lint`
  偏静态配置审计
- `full`
  尽量完整
- `incident`
  更强调主因摘要

## 7. 输出目标

### 7.1 Terminal

- 默认只展开重点问题
- 证据截断
- 面向值班排障

### 7.2 JSON

- 结构稳定
- 包含 `schema_version`
- 包含 `tool_version`
- 面向自动化

### 7.3 Markdown

- 章节固定
- 面向留档 / 工单

## 8. 证据覆盖语义

封版后的覆盖摘要不再使用“尝试过采集”的乐观说法，而是按证据语义输出：

- `已获取证据`
- `无可用证据`
- `未纳入本次运行`
- `探针=已执行`

## 9. 封版目标

这版封版追求的不是“继续做大”，而是：

- 更稳
- 更准
- 更短
- 更像给值班人员看的报告

只要默认输出不再制造额外噪声，覆盖摘要与检查明细不再互相打架，证据不再重复误导，文档、二进制和输出行为保持一致，就达到这版封版目标。
