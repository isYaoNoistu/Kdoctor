# Kdoctor

`Kdoctor` 是一个面向 Kafka 运维排障场景的 Go 二进制诊断工具。

它的目标不是替代 Kafka 管理平台，而是在你手里只有一个 `bootstrap` 地址，或者额外还能拿到 `profile`、`compose`、日志目录、Docker 运行时信息时，尽快判断问题更像是网络、`advertised.listeners`、KRaft、Topic/ISR、客户端链路，还是宿主机、容器和日志层。

详细使用手册见 [USER_GUIDE.md](./USER_GUIDE.md)。  
手册已区分 `Windows` / `Linux` 两套命令和路径写法，可直接照抄使用。

## 工具定位

- 只给一个 `bootstrap` 地址，也能执行最小可用巡检。
- 补充 `profile`、`compose`、日志目录、Docker 信息后，会增强规则和证据，但不会改写主流程。
- 默认输出中文终端报告，也支持 `json` 和 `markdown`。
- 默认终端只展开 `CRIT / FAIL / WARN / ERROR`，`PASS / SKIP` 默认折叠，方便值班人员快速扫读。
- 报告会优先给出 `主因判断 + 建议动作`，而不是把所有检查项平铺给人看。

## 当前封版能力

- 网络：`NET-001~009`
- Kafka 元数据与拓扑：`KFK-001~009`
- KRaft：`KRF-001~005`
- Topic / ISR / 规划：`TOP-003~009`、`TOP-011`
- Client Probe：`CLI-001~005`
- Config Lint：`CFG-001~014`
- Producer / Consumer / Transaction 上下文：`PRD-001~004`、`PRD-006`、`CSM-001~006`、`TXN-001~005`
- Security / Storage / Host / Docker / Logs：`SEC-001~005`、`STG-001`、`STG-003`、`STG-005~006`、`HOST-004`、`HOST-007~011`、`DKR-001~007`、`LOG-001~008`

以下内容已从封版默认能力中移除：

- JMX / Metrics / JVM / Quota 相关检查
- 依赖 JMX 的 Host / KRaft 路径
- 默认报告中的 JMX 噪声和相关 `SKIP`

## 最常用命令

Windows：

```powershell
.\kdoctor.exe probe --bootstrap 192.168.1.1:9192
.\kdoctor.exe probe --config .\kdoctor.yaml
.\kdoctor.exe probe --config .\kdoctor.yaml --compose .\docker-compose.yml
.\kdoctor.exe probe --bootstrap 192.168.1.1:9192 --format markdown --output .\report.md
```

Linux：

```bash
./kdoctor probe --bootstrap 192.168.1.1:9192
./kdoctor probe --config ./kdoctor.yaml
./kdoctor probe --config ./kdoctor.yaml --compose ./docker-compose.yml
./kdoctor probe --bootstrap 192.168.1.1:9192 --format markdown --output ./report.md
```

## 输出说明

- 默认终端报告：压缩版，只展开重点问题。
- `--verbose`：展开 `PASS / SKIP` 明细。
- `--json`：适合脚本或自动化处理。
- `--format markdown`：适合留档、发群、贴工单。

默认输出行为：

- `output.max_evidence_items=8`
- `output.show_pass_checks=false`
- `output.show_skip_checks=false`
- `output.verbose=false`

## 构建

在仓库根目录执行：

```powershell
go test ./...
.\scripts\build.ps1
.\scripts\build.ps1 -GOOS linux -GOARCH amd64
```

构建产物默认输出到工作区根目录的 `dist/`，不放在代码仓库目录里。

## 目录

```text
Kdoctor/
  cmd/kdoctor/           CLI 入口
  internal/              内部实现
  pkg/model/             统一报告模型
  scripts/build.ps1      构建脚本
  kdoctor.example.yaml   示例配置
  kdoctor.yaml           常用环境配置
  USER_GUIDE.md          详细用户手册
  architecture.md        架构与工程标准
  doc.md                 设计文档
  version/               版本阶段记录
```

## 说明

- `compose`、Docker、日志目录都不是前置条件；没有这些输入时，相关检查不会再在默认报告里制造噪声。
- 终端、Markdown、JSON 三种输出已经统一了中文术语、证据顺序和问题排序。
- 设计说明见 [doc.md](./doc.md)，工程标准见 [architecture.md](./architecture.md)。
