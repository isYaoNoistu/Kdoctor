# 审计收口记录

## 1. 目标

本轮依据 `version/end.md` 的审计结论，对 `Kdoctor` 做封版前最后一轮收口优化。

目标不是继续扩功能，而是把现有能力磨到：

- 更稳
- 更准
- 更短
- 更适合值班人员直接使用

## 2. 本轮完成项

### 2.1 默认链路进一步降噪

- 默认报告继续保持只展开重点问题
- 覆盖摘要统一按“是否拿到证据”表达
- 额外指标类默认链路不再进入封版主流程

### 2.2 证据与输出收口

- 修正 `TOP-011`，只输出真正命中的 topic
- 修正 `NET-003 / NET-005 / KFK-005` 一类端点证据的重复与歧义
- 修正日志、宿主机、探针、Topic 规划等输出中的中英混排
- 终端、JSON、Markdown 三种输出继续对齐

### 2.3 文档收口

- 重写 `README.md`
- 重写 `USER_GUIDE.md`
- 重写 `doc.md`
- 重写 `architecture.md`
- 统一成 V2 封版说明
- 清理主要文档中的乱码与漂移表述

### 2.4 工程化收口

- `--version`、`tool_version`、`schema_version` 保持可用
- 构建脚本继续向工作区根目录 `dist/` 输出产物
- `.gitignore` 补充忽略测试缓存与运行报告

## 3. 本轮验证

已执行：

- `go test ./...`
- `go run ./cmd/kdoctor --version`
- `go run ./cmd/kdoctor probe --bootstrap 192.168.100.78:9192`
- `go run ./cmd/kdoctor probe --bootstrap 192.168.100.78:9192 --json --output ./probe.json`
- `go run ./cmd/kdoctor probe --bootstrap 192.168.100.78:9192 --format markdown --output ./report.md`

## 4. 当前真实环境结论

在 `192.168.100.78:9192` 这组环境上，工具本身已经可用，当前剩余重点问题属于环境真实状态：

- `KRF-003 / KRF-004`
  当前执行视角下 `192.168.119.7:9193/9195/9197` 不可达，多数派证据不足
- `NET-002`
  显式 controller listener 不可达
- `TOP-005 / TOP-007`
  `__consumer_offsets` 多个分区低于 `min.insync.replicas=2`
- `TOP-011`
  `_kdoctor_probe` 分区数低于 broker 数，属于规划型提示

同时也确认：

- `NET-001 / NET-003 / NET-005` 可正常工作
- `CLI-001 ~ CLI-005` 探针链路可通过
- `KFK-001 ~ KFK-006` 主链路可运行

## 5. 结论

这轮完成后，`Kdoctor` 已经达到“可继续用于内部审计、可用性测试和封版交付”的状态。
