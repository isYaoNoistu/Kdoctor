# Kdoctor

`Kdoctor` 是一个面向内部 Kafka 运维与排障场景的单二进制诊断工具。

它不是管理平台，也不是监控平台。它的目标很直接：当你手里只有一个 `bootstrap` 地址，或者还能补充 `kdoctor.yaml`、`docker-compose.yml`、Docker 运行时、日志目录时，尽快把问题收敛到“更像网络、listener、KRaft、Topic/ISR、客户端链路，还是宿主机 / 容器 / 日志层”。

当前仓库已经收口到 **V2 封版可用状态**。默认输出以“可信、短、清楚、低噪声”为目标，适合值班与现场初诊。

详细使用说明见 [USER_GUIDE.md](./USER_GUIDE.md)。

## 1. 工具定位

- 面向 Kafka 首轮巡检与现场排障
- 面向内部使用，不做平台化扩展
- 单二进制交付
- `bootstrap-only` 可运行，`profile / compose / docker / logs` 是增强输入
- 默认中文输出
- 默认终端只展开 `CRIT / FAIL / WARN / ERROR`

## 2. 默认能力范围

当前默认保留：

- 网络：`NET-*`
- Kafka 元数据与 broker：`KFK-*`
- KRaft：`KRF-*`
- Topic / ISR / leader / 规划：`TOP-*`
- 客户端链路探针：`CLI-*`
- Compose 配置审计：`CFG-*`
- 消费组：`CSM-*`
- Producer / Transaction 上下文：`PRD-*`、`TXN-*`
- Security / Storage / Host / Docker / Logs：`SEC-*`、`STG-*`、`HOST-*`、`DKR-*`、`LOG-*`

当前默认不再纳入封版主链路：

- 历史指标扩展检查
- 依赖额外指标采集的默认检查注册
- 默认报告中的额外指标类 `SKIP` 噪声

## 3. 最常用命令

Windows:

```powershell
.\kdoctor.exe probe --bootstrap 192.168.1.1:9192
.\kdoctor.exe probe --config .\kdoctor.yaml
.\kdoctor.exe probe --config .\kdoctor.yaml --compose .\docker-compose.yml
.\kdoctor.exe probe --bootstrap 192.168.1.1:9192 --json
.\kdoctor.exe probe --bootstrap 192.168.1.1:9192 --format markdown --output .\report.md
.\kdoctor.exe --version
```

Linux:

```bash
./kdoctor probe --bootstrap 192.168.1.1:9192
./kdoctor probe --config ./kdoctor.yaml
./kdoctor probe --config ./kdoctor.yaml --compose ./docker-compose.yml
./kdoctor probe --bootstrap 192.168.1.1:9192 --json
./kdoctor probe --bootstrap 192.168.1.1:9192 --format markdown --output ./report.md
./kdoctor --version
```

## 4. 输入模式

### 4.1 最小输入

只给一个 `bootstrap` 地址也能运行，至少支持：

- bootstrap 连通性
- metadata 拉取
- metadata 返回地址检查
- Topic / leader / ISR 检查
- `metadata -> produce -> consume -> commit -> e2e` 探针

### 4.2 配置增强

如果提供 `kdoctor.yaml`，会增强：

- 期望 broker 数
- 期望 replication factor
- 期望 min ISR
- producer / consumer / transaction 上下文
- 输出与摘要参数

### 4.3 Compose 增强

如果提供 `docker-compose.yml`，会额外增强：

- `listeners`
- `advertised.listeners`
- `controller.quorum.voters`
- `node.id`
- `process.roles`
- `inter.broker.listener.name`
- 容器 / volume / 数据目录规划

### 4.4 Docker / Logs / Host 增强

如果当前执行机能访问 Docker、宿主机信息或日志来源，会继续增强：

- 容器状态、重启、OOM、挂载
- 宿主机磁盘、端口、fd、内存
- 日志来源、指纹、上下文与聚合

## 5. 输出格式

- `terminal`
  适合值班现场，默认只看重点问题。
- `json`
  适合自动化处理，字段稳定，包含 `schema_version` 与 `tool_version`。
- `markdown`
  适合留档、工单和审计。

## 6. 构建

仓库内源码根目录是 `Kdoctor/`，构建产物输出到工作区根目录 `dist/`，不放进代码仓库。

Windows PowerShell:

```powershell
.\scripts\build.ps1 -GOOS windows -GOARCH amd64
.\scripts\build.ps1 -GOOS linux -GOARCH amd64
```

构建后可在工作区根目录 `dist/` 下打包分发。

## 7. 当前建议用法

如果你是第一次接入环境，建议按这个顺序：

1. 先跑 `bootstrap-only`
2. 再补 `kdoctor.yaml`
3. 最后补 `docker-compose.yml`

这样最容易看清“环境本身有问题”，还是“部署配置与运行状态不一致”。

## 8. 版本信息

- CLI 支持 `--version`
- JSON 报告包含 `schema_version`
- 构建时会注入版本号与 commit

## 9. 仓库文档

- [USER_GUIDE.md](./USER_GUIDE.md)：完整用户手册
- [doc.md](./doc.md)：设计文档
- [architecture.md](./architecture.md)：架构与工程标准
- [version](./version)：阶段记录与封版文档
