# Kdoctor V2 封版收口记录

## 本轮目标

按照 `version2-release.md` 完成封版前最后一轮完善，不再继续扩展功能域，只做：

- 代码层规则优化
- 输出格式优化
- 参数与默认值收口

## 本轮完成项

### 1. JMX 路径从默认能力中移除

- 从检查注册链路中移除了 `MET-*`、`JVM-*`、`QTA-*`
- 同步下线了依赖 JMX 的 `HOST-009`、`KRF-006`、`KRF-007`
- `KRF-008` 也从封版默认注册中移除
- 默认终端、Markdown、摘要覆盖里都不再出现 JMX 相关 `SKIP`

### 2. 证据覆盖语义修正

覆盖状态统一成三态：

- 已启用，已获取证据
- 已启用，未获取证据
- 未纳入本次运行

不再出现“摘要说已采集，但明细全是 SKIP”的冲突表达。

### 3. 证据去重与限流

- `NET-003`
- `NET-005`
- `KFK-005`

相关 endpoint 证据已做归一化和去重。  
终端与 Markdown 单检查证据展示上限收口为 `8` 条，超出后显示“其余 X 条已省略”。

### 4. TOP-011 误导修正

`TOP-011` 现在只输出真正命中的 topic：

- `replication factor > broker count`
- `partition count < broker count`

正常 topic 不再混进告警 evidence。

### 5. 日志语义修正

- `LOG-001` 明确为“日志来源与样本质量”
- `LOG-002` 只表示“是否命中已知错误指纹”

`LOG-001` 证据现在统一展示：

- `source`
- `line_count`
- `byte_count`
- `latest_timestamp`
- `freshness`
- `sample_sufficient`

### 6. 终端与 Markdown 输出收口

默认终端输出现在只展开：

- `CRIT`
- `FAIL`
- `WARN`
- `ERROR`

`PASS / SKIP` 默认折叠，仅在：

- `--verbose`
- Markdown 附录
- JSON

中展开。

Markdown 输出改成了更适合留档的结构：

- 摘要表格
- 证据覆盖
- 主因判断
- 建议动作
- 重点问题表
- 重点问题详情
- 可展开完整附录

### 7. 文档与示例配置收口

已更新：

- `README.md`
- `USER_GUIDE.md`
- `kdoctor.example.yaml`
- `kdoctor.yaml`
- `version2-release.md`

文档已统一到：

- 无 JMX 默认能力
- 默认折叠 PASS / SKIP
- Windows / Linux 双平台用法
- 输出参数 `output.*`

### 8. 构建输出目录收口

`scripts/build.ps1` 已改为输出到工作区根目录 `dist/`，不再把二进制放在仓库目录里。

## 测试与验证

本轮已执行：

- `gofmt`
- `go test ./...`

结果通过。

同时补了三套固定样例输出测试：

- `bootstrap-only`
- `probe-only`
- `compose + docker + logs`

并分别覆盖：

- 终端输出
- Markdown 输出

## 封版结论

`Kdoctor V2` 已完成封版前最后一轮收口，当前状态可以认定为：

- 功能边界已冻结
- 默认输出已减噪
- 中文术语已统一
- 证据语义已收口
- 二进制与用户手册可直接交付内部使用
