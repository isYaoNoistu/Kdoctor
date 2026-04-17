# Kdoctor

`Kdoctor` 是一个面向 Kafka 运维排障场景的 Go 二进制诊断工具。

它的目标不是替代 Kafka 管理平台，而是在你手里只有一个 `bootstrap` 地址，或者额外还能拿到 `profile`、`compose`、日志目录、Docker 运行时信息时，尽快判断问题更像是网络、`advertised.listeners`、KRaft、Topic/ISR、客户端链路，还是宿主机、容器和日志层。

详细使用手册见 [USER_GUIDE.md](./USER_GUIDE.md)。
手册已经区分 `Windows` / `Linux` 两套命令和路径写法，可直接按系统照抄。

## 作用

- 支持最小输入模式：只给 `bootstrap` 也能执行网络、metadata、topic、probe 检查。
- 支持增强输入模式：附加 `profile`、`compose`、日志目录后，可补充配置校验、KRaft 对照、宿主机、容器、日志指纹检查。
- 默认输出中文终端报告，也支持 `json` 和 `markdown`。
- 面向“现场可用性”和“误报控制”，缺少上下文时优先 `SKIP` 或降级，而不是机械报错。
- `probe` 会优先检查 `_kdoctor_probe` 是否可用；主题不存在时会尝试自动创建，避免 fresh cluster 被误判成链路故障。
- `probe` 结果已按阶段收口：上游阶段失败时，下游未执行项会标记为 `SKIP`，不会再把一处失败扩散成多条重复 `FAIL`。

## 当前 V1 能力

- 网络：`NET-001~004`
- Kafka 元数据：`KFK-001~004`
- KRaft：`KRF-001~003`
- Topic/Replica：`TOP-003~005`
- Client Probe：`CLI-001~005`
- Config Lint：`CFG-001~008`
- Host：`HOST-004`、`HOST-006`
- Docker：`DKR-001~004`
- Logs：`LOG-001~004`

## 输入模式

### 1. `bootstrap-only`

这是最小可用模式，也是工具默认应该支持的模式。

```powershell
go run ./cmd/kdoctor probe --bootstrap 192.168.1.1:9192
```

### 2. `bootstrap + profile`

在只有地址的基础上补充环境预期值，例如 broker 数量、controller 端点、min ISR 等。

```powershell
go run ./cmd/kdoctor probe --bootstrap 192.168.1.1:9192 --profile generic-bootstrap
```

### 3. `bootstrap + compose`

在前两者基础上增加静态配置 lint 和部署侧对照检查。

```powershell
go run ./cmd/kdoctor probe --bootstrap 192.168.1.1:9192 --compose .\docker-compose.yml
```

### 4. `bootstrap + 输出文件`

支持 JSON 和 Markdown 两种文件输出。

```powershell
go run ./cmd/kdoctor probe --bootstrap 192.168.1.1:9192 --json
go run ./cmd/kdoctor probe --bootstrap 192.168.1.1:9192 --format markdown --output .\report.md
```

## 运行模式

- `quick`：快速巡检，优先给出核心健康结论。
- `probe`：执行真实 client probe，是当前最常用模式。
- `lint`：偏静态配置与部署校验。
- `full`：尽量执行完整检查。
- `incident`：面向故障现场，输出更聚焦的摘要。

## 构建

在项目根目录执行：

```powershell
go test ./...
.\scripts\build.ps1
```

默认产物在：

```text
dist/kdoctor-windows-amd64.exe
```

如需交叉编译：

```powershell
.\scripts\build.ps1 -GOOS linux -GOARCH amd64
```

## 使用示例

### 1. 直接使用源码运行

```powershell
go run ./cmd/kdoctor probe --bootstrap 192.168.1.1:9192
```

### 2. 输出 JSON

```powershell
go run ./cmd/kdoctor probe --bootstrap 192.168.1.1:9192 --json
```

### 3. 输出 Markdown 报告

```powershell
go run ./cmd/kdoctor probe --bootstrap 192.168.1.1:9192 --format markdown --output .\report.md
```

### 4. 使用已构建二进制

```powershell
.\dist\kdoctor-windows-amd64.exe probe --bootstrap 192.168.1.1:9192
```

## 输出与退出码

- 默认输出：中文终端报告
- `--json`：机器可读 JSON
- `--format markdown`：便于发群、贴工单、留档

`probe` 模式下的结果语义：

- `CLI-001` 对应 `bootstrap -> metadata`
- `CLI-002` 对应 `produce`
- `CLI-003` 对应 `consume`
- `CLI-004` 对应 `commit`
- `CLI-005` 对应整条端到端链路

如果某个上游阶段已经失败，后续未执行阶段会显示为 `SKIP`，并附带失败阶段说明。

退出码约定：

- `0`：最高状态为 `PASS`
- `1`：最高状态为 `WARN`
- `2`：最高状态为 `FAIL`
- `3`：最高状态为 `CRIT`
- `5`：最高状态为 `ERROR`
- `6`：最高状态为 `TIMEOUT`

参数解析或配置初始化失败时，CLI 直接以错误码退出。

## 目录

```text
kdoctor/
  cmd/kdoctor/           CLI 入口
  internal/              内部实现
  pkg/model/             统一报告模型
  dist/                  二进制输出目录
  scripts/build.ps1      构建脚本
  kdoctor.example.yaml   示例配置
  architecture.md        架构与工程标准
  doc.md                 设计文档
  version/               阶段记录
```

## 说明

- `compose`、Docker、日志目录都不是前置条件；没有这些输入时，相关检查会合理 `SKIP`。
- `dist/` 只放构建产物，不放源码。
- 当前默认输出已经是中文，适合直接做初步人工排障。
- 设计说明见 [doc.md](./doc.md)，工程标准见 [architecture.md](./architecture.md)。

