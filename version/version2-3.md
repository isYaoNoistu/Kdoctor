# version2-3

## 阶段定位

这是第二阶段的第三轮推进，目标是补上 `P0` 里最关键但此前仍缺位的一块能力：

- `JMX / metrics` 采集基础
- 指标型检查项的第一批落地

这轮的定位不是一次性把所有 JMX 相关能力做完，而是先把“能采、能判、能接入主流程”的底座搭起来，为后续 `MET-* / JVM-* / STG-*` 继续扩展打基础。

## 本轮完成内容

### 1. JMX / metrics 采集底座已接入

这轮新增了 `metrics snapshot` 与 `collector`，已经正式进入主执行流程。

当前支持两种来源：

- 显式配置 `jmx.endpoints`
- 在 `compose` 模式下，从 Kafka 服务环境变量中自动推断 JMX 端点

新增配置字段：

- `jmx.path`
- `jmx.endpoints`

默认仍然是：

- `path=/metrics`
- `enabled=false`

也就是说，这轮不会强迫所有环境都打开 JMX，但一旦配置了，工具就已经具备采集与判定能力。

### 2. 第一批指标检查已落地

本轮新增并接入：

- `MET-001` UnderReplicatedPartitions
- `MET-002` UnderMinISR / AtMinISR
- `MET-004` OfflineLogDirectoryCount
- `JVM-001` NetworkProcessorAvgIdlePercent
- `JVM-002` RequestHandlerAvgIdlePercent

这些检查会在有可用 JMX 指标时真正执行；如果当前环境没有可用 JMX 来源，会明确 `SKIP`，不会误报。

### 3. 采集覆盖与根因归并已吸收 JMX

这轮同时补了两件对现场很重要的事：

- 报告摘要里的“采集覆盖”现在会展示 `JMX` 是否已采到
- root cause 归并已经能吸收：
  - `MET-001`
  - `MET-002`
  - `MET-004`
  - `JVM-001`
  - `JVM-002`

这样输出不只是“多几条指标检查”，而是会把：

- 副本复制压力
- UnderMinISR 压力
- OfflineLogDirectory
- broker 线程池压力

提升到主因层。

## 验证情况

本轮已执行：

- `gofmt`
- `go test ./...`

结果通过。

## 当前阶段判断

第二阶段 `P0` 现在已经不只是配置/消费组/存储/安全的静态与半动态检查了，开始进入真正的运行态指标诊断阶段。

按 `version2.md` 的主线看：

- `group / lag`：已落地
- `scheduler timeout / degrade`：已落地
- `security` 基础域：已落地
- `storage / logdir` 基础域：已落地
- `JMX / metrics` 基础能力：已落地第一批

## 下一步建议

最自然的下一步是继续推进：

1. `TOP-006 / TOP-007 / STG-002` 和指标之间做更强的双证据关联
2. 继续补 `TOP-008 / MET-003`
3. 往 `KRF-004 / KRF-006 / KRF-007` 这批 KRaft 指标型判断推进

这样第二阶段 `P0` 会逐步从“能看见”走向“能更稳定地定位”。 
