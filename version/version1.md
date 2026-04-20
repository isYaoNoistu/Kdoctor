# Kdoctor V1

## 版本定位

`V1` 是 `Kdoctor` 从设计文档进入可执行、可验证、可初步投入现场使用的第一个稳定阶段。

它重点解决的是“先把 Kafka 首轮排障主链路跑通”，而不是一开始就把所有边角能力堆满。  
这一版的目标很明确：

- 只给一个 `bootstrap` 地址也能运行
- 能把问题快速收敛到网络、metadata、KRaft、Topic / ISR、客户端链路几个主干层
- 结果尽量短、尽量清楚，适合值班场景

## 核心特性

### 1. 最小输入可运行

`V1` 明确支持 `bootstrap-only` 模式，不再要求必须提供 `docker-compose.yml` 才能诊断。

在最小输入模式下，至少可以完成：

- `NET-*` 网络连通与返回端点检查
- `KFK-*` metadata 与 broker 基础检查
- `TOP-*` leader / ISR / 内部主题基础检查
- `CLI-*` 从 `metadata -> produce -> consume -> commit -> e2e` 的真实链路探针

### 2. 增强输入模式成型

在只给地址之外，`V1` 也建立了增强输入体系：

- `kdoctor.yaml`：提供期望 broker 数、期望副本数、探针参数
- `docker-compose.yml`：补齐 listeners、advertised.listeners、controller.quorum.voters、node.id、process.roles 等配置检查
- Docker / Host / Logs：在环境允许时补充容器、宿主机、日志层证据

### 3. 主干检查体系落地

`V1` 最终形成的主干能力包括：

- 网络：`NET-001~004`
- Kafka 元数据：`KFK-001~004`
- KRaft：`KRF-001~003`
- Topic / ISR：`TOP-003~005`
- Client Probe：`CLI-001~005`
- Config Lint：`CFG-001~008`
- Host / Docker / Logs：首批可用检查

这些能力让工具从“设计文档”变成了真正能跑的 Kafka 首轮巡检脚本。

### 4. 工程化基础完成

`V1` 阶段完成了项目重构和交付基础：

- 仓库从 `kafka-check` 收口为 `Kdoctor`
- Go module 与内部 import 统一
- 建立 `cmd / internal / pkg / scripts / version` 结构
- 二进制输出统一到工作区根目录 `dist/`
- README、设计文档、架构文档、用户手册全部补齐

## 优化与改变方向

`V1` 的优化重点不是继续扩功能，而是把“能跑”逐步收敛成“基本可信”。

### 1. 误报控制

`V1` 期间多轮真实测试推动了误报收敛，重点包括：

- 外部视角下的 KRaft controller 误判修正
- `compose` 从“运行前提”调整为“增强输入”
- 宿主机 / 数据盘 / listener 视角误判修正

这一阶段确立了一个原则：

- 看不到证据时，宁可 `SKIP / WARN`
- 不把上下文不足硬判成 Kafka 故障

### 2. 结果表达开始面向值班

`V1` 不是简单列检查项，而是开始形成：

- 概览
- 主因判断
- 建议动作
- 重点问题

也正是在这个阶段，`Kdoctor` 的中文输出、摘要式报告和低噪声方向被确定下来。

### 3. 真实链路优先于静态推测

`V1` 的一条核心路线是：尽量让结论建立在真实 probe 上，而不是只看配置和端口。

所以 `V1` 最重要的能力并不是某一个静态检查，而是：

- 能真实访问 Kafka
- 能真实发送消息
- 能真实消费与提交 offset

## V1 结论

如果用一句话概括 `V1`，它更像：

**一个已经可以上现场的 Kafka 首轮巡检工具。**
