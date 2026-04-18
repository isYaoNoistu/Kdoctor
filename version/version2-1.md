# version2-1

## 阶段定位

这是第二阶段的第一轮落地，不是把 `version2.md` 里所有能力一次性做完，而是先按 `P0` 优先级补最关键的诊断盲区，为后续 JMX、安全、quota、upgrade 等域打基础。

本轮目标：

- 把 V2 配置骨架接进工程
- 补齐消费组 / lag 采集与检查
- 让 host 与 threshold 进入参数化阶段

## 本轮完成内容

### 1. V2 配置骨架已接入

这次扩展了配置模型，已支持第二阶段后续能力继续生长，主要新增：

- `profiles.execution_view`
- `profiles.security_mode`
- `profiles.group_probe_targets`
- `docker.inspect_mounts`
- `probe.acks`
- `probe.enable_idempotence`
- `probe.cleanup_mode`
- `probe.tx_probe_enabled`
- `jmx.*`
- `host.*`
- `thresholds.*`
- `diagnosis.suppress_downstream_symptoms`
- `diagnosis.rule_packs`

同时：

- `kdoctor.example.yaml`
- `kdoctor.yaml`

都已升级到 `version: 2` 结构。

### 2. 消费组 lag / 状态盲区已开始补齐

这是 V2 第一轮里最重要的实装项之一。

新增了消费组快照模型与采集能力：

- 采集 group state
- 采集 coordinator
- 采集 committed offsets
- 采集 end offsets
- 计算 total lag / max partition lag / missing offsets

对应新增检查：

- `CSM-001` 消费组 lag
- `CSM-002` 消费组 rebalance / 状态稳定性
- `CSM-006` coordinator 与位点视图可用性

这些检查已经接入主执行流程，不再是占位目录。

### 3. host / threshold 进入参数化

这轮没有新增完整的 `HOST-007~011`，但已经把第二阶段需要的阈值基础接起来了：

- `HOST-004` 磁盘阈值不再写死，开始使用 `thresholds.disk_warn_pct` / `disk_crit_pct`
- host collector 现在支持 `host.disk_paths`
- host collector 现在支持 `host.check_ports`

这意味着后续 storage / logdir / port drift 相关能力可以直接继续叠加，而不用再回头改底层配置结构。

### 4. 根因摘要已接入消费组链路

root cause 规则现在已经能吸收：

- lag 高
- rebalance 异常
- coordinator / offsets 视图异常

输出不会只停留在检查项层面，而会开始把问题归并到“真实消费链路异常”这一层。

### 5. 执行超时 / 降级可见性已结构化

这轮同时把第二阶段 `scheduler timeout/degrade` 的第一层能力补上了：

- 采集任务的超时 / 软失败不再只混在附加错误里
- 报告摘要现在会单独给出“采集覆盖”
- 报告摘要现在会单独给出“采集降级”
- terminal / markdown / json 输出都会保留这些信息

这会直接提升值班现场的可用性，因为可以更快区分：

- 是工具没拿到证据
- 还是证据已经拿到，但检查项判断出了真实异常

## 验证情况

本轮已执行：

- `gofmt`
- `go test ./...`

结果通过。

## 当前阶段判断

第二阶段现在已经正式开始进入“功能落地”阶段，不再只是文档规划阶段。

如果按 `version2.md` 的路线图看：

- `P0` 已开始推进
- 当前已完成 `P0` 中的第一块关键基础能力
- 下一批最自然的推进方向是：
  - `JMX / metrics`
  - `storage / logdir`
  - `security` 基础域

## 下一步建议

最合适的下一步是直接继续推进：

1. `JMX` 采集基础能力
2. `TOP-006 / TOP-007 / MET-*` 这批指标型检查
3. `SEC-001~003` 的基础协议 / TLS 诊断

这样第二阶段的主干就会真正成形。  
