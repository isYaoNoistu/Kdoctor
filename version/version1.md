# Version 1

## 阶段目标

本阶段的目标是把项目从一份设计文档推进为一个可以真实执行、可以做初步测试、结果基本可信的 Kafka 运维诊断工具，并完成基础工程化整理，便于后续继续迭代和对外使用。

## 本阶段完成的核心工作

### 1. 完成 V1 主干能力落地

已经按设计文档和架构文档完成 `V1` 主要检查能力实现，覆盖：

- 网络层：`NET-001~004`
- Kafka 元数据层：`KFK-001~004`
- KRaft 层：`KRF-001~003`
- Topic / Replica 层：`TOP-003~005`
- Client Probe 层：`CLI-001~005`
- Config Lint 层：`CFG-001~008`
- Host 层：`HOST-004`、`HOST-006`
- Docker 层：`DKR-001~004`
- Log 层：`LOG-001~004`

其中最关键的是：

- 支持真实 `bootstrap -> metadata -> produce -> consume -> commit` 端到端探针
- 支持 `bootstrap-only` 最小输入模式，不依赖 `compose`
- 支持 `compose` 增强模式，在上下文充分时补充配置、KRaft、宿主机、容器、日志检查

### 2. 修复多轮真实测试中暴露的误报问题

本阶段重点收敛了几类误报：

- 修复 `KRF-003` 在“外部视角 + 私网 controller 地址”下误报 `CRIT`
- 调整 `NET-002`、`KRF-002`、`KRF-003` 的视角判断逻辑
- 修复 `HOST` 层把当前代码仓库磁盘误判成 Kafka 数据盘的问题
- 保持 `compose` 为增强输入，而不是运行前提

### 3. 完成工程结构标准化

为了便于交付与后续维护，项目完成了工程重构：

- 项目目录从 `kafka-check` 重构为 `kdoctor`
- Go module 名称同步为 `kdoctor`
- 统一内部 import 路径
- 建立清晰分层：`cmd`、`internal`、`pkg`、`dist`、`scripts`

### 4. 统一二进制产物输出

新增标准构建方式：

- 构建脚本：[scripts/build.ps1](../scripts/build.ps1)
- 二进制输出目录：[dist](../dist)

默认输出：

```text
dist/kdoctor-windows-amd64.exe
```

### 5. 补齐仓库入口文档

补充了基础使用说明：

- [README.md](../README.md)
- [doc.md](../doc.md)
- [architecture.md](../architecture.md)

这些文档定义了工具作用、使用方式、设计边界和工程标准。

### 6. 完成代码入库与远端仓库落地

本阶段完成了 Git 仓库初始化、主分支整理和远端推送，为后续审核与协作打下基础。

## 验证情况

### 自动化验证

已执行并通过：

```powershell
go test ./...
```

### 构建验证

已执行并成功：

```powershell
.\scripts\build.ps1
```

### 真实环境验证

已基于真实 Kafka 地址进行多轮实际探测，验证结论如下：

- 工具主链路可以真实运行
- `bootstrap-only` 模式可直接执行
- `compose` 增强模式可输出更多部署侧结论
- 当前剩余失败主要来自环境本身，而不是工具无法运行

## 当前环境侧发现的真实问题

在此前测试环境中，工具持续发现以下真实问题：

- metadata 返回的部分 broker 端点不可达
- client probe 可能因 leader 或 coordinator 落到不可达 broker 上而失败
- 某些内部主题状态并不理想

这些结果说明：

- 工具已经具备发现真实环境问题的能力
- 当前阻塞使用测试的重点更多在 Kafka 环境本身，而不是 `kdoctor` 主干能力缺失

## 当前阶段结论

本阶段已经把 `kdoctor` 从“设计阶段”推进到“可以进入初步使用测试”的状态，达到的结果是：

- 有完整的 `V1` 功能主干
- 有标准工程结构
- 有 README、设计文档、架构文档
- 有 `dist` 二进制产物
- 有基础测试与真实探测验证

## 下一阶段建议

后续如进入 `V1.x / V2`，建议优先推进：

1. 诊断归因进一步增强
2. 更多场景夹具与 golden 输出
3. 更强的 Markdown 报告与发布物沉淀
4. 容量与趋势判断
5. 根据真实使用反馈继续收敛误报与提示文案
