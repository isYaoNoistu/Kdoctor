# Version 1-4

## 阶段目标

本阶段的目标是把 `version-1-release.md` 中定义的 `V1` 最终收口项真正落到代码里，让 `kdoctor` 达到“可以直接投入一线使用”的状态，而不只是“能运行、能展示结果”。

重点收口的方向有四类：

1. 修复 probe 链路的级联失败噪声
2. 补齐 probe topic 生命周期
3. 把内部主题判断改成上下文感知
4. 继续提升中文输出与最终使用体验

## 本阶段完成的核心工作

### 1. probe 改成按阶段执行、按阶段结论输出

本阶段重构了 `probe` 主链路，补齐了阶段状态模型：

- `metadata`
- `topic_ready`
- `produce`
- `consume`
- `commit`
- `complete`

现在探针不再把“上游失败”扩散成多条重复失败，而是：

- 在快照里记录 `executed_stage`、`failure_stage`
- 对未执行到的下游阶段输出明确 `SKIP`
- 在报告中直接说明“为什么没有继续执行”

这让现场报告明显更干净，也更容易一眼看到真正的阻断点。

### 2. 补齐 probe topic 生命周期

本轮把 `_kdoctor_probe` 从“直接拿来发消息”的弱前提，升级成了真正可控的探针资源：

- 先检查 topic 是否存在
- 不存在时自动尝试创建
- 自动按 broker 数量推导合理副本数
- 支持 `produce_count`
- 在 `cleanup=true` 且本轮由 `kdoctor` 创建 topic 时，尝试自动清理

这意味着：

- fresh cluster 不会再因为 probe topic 缺失被误判成业务链路故障
- 配置中的 `produce_count` 和 `cleanup` 不再是空参数

### 3. `KFK-004` 改成上下文感知判断

`__consumer_offsets` 的判断逻辑已经从“绝对 FAIL”改成“结合 probe 上下文”：

- 如果 commit 探针已经执行，而 `__consumer_offsets` 仍缺失，会判为 `FAIL`
- 如果当前更像是 fresh cluster 或 commit 链路还没真正跑到，会降级为 `WARN`
- `__transaction_state` 仍维持合理 `WARN`，避免把未启用事务场景误判成严重故障

这项调整直接提升了 `bootstrap-only` 场景下的可信度。

### 4. 运行后刷新 Kafka / Topic 快照

本阶段把 probe 后的快照刷新接回了主流程：

- 如果本轮创建了 probe topic
- 或执行了 commit
- 或执行了 cleanup

则会在 probe 后重新采集 Kafka / Topic 元数据。

这样 `KFK-004`、`TOP-*` 等检查看到的是更接近“probe 后真实状态”的数据，而不是过早采集到的旧快照。

### 5. 中文输出继续收口

这一轮继续补齐了中文输出覆盖，重点包括：

- 新增 probe 阶段跳过原因的中文化
- 新增 `executed_stage` / `failure_stage` / `topic_ready_reason` 等证据翻译
- 去掉了重复证据
- 让摘要不再重复堆叠 fallback 主因

现在终端输出、JSON 输出里的关键结论已经基本是中文可读状态。

### 6. 补充回归测试

本阶段新增并通过了与本轮修复直接相关的测试：

- client 阶段跳过行为测试
- `KFK-004` 上下文判断测试

## 验证情况

### 自动化验证

已执行并通过：

```powershell
go test ./...
```

### 真实环境验证

本阶段使用以下测试节点进行了真实验证：

- `192.168.100.78:9192`
- `192.168.100.78:9194`
- `192.168.100.78:9196`

典型命令：

```powershell
go run ./cmd/kdoctor probe --bootstrap 192.168.100.78:9192
go run ./cmd/kdoctor probe --bootstrap 192.168.100.78:9192 --json
```

### 当前测试结论

本轮真实测试验证出两个重要结果：

1. `probe` 主链路已经达到“基本可用可信”

- `CLI-001~005` 可以在真实环境中完整跑通
- `_kdoctor_probe` 不存在时，工具会自动准备 topic
- 下游检查不会再因为上游失败发生级联误报

2. 当前环境里剩余的主要告警已经变成真实环境信号

- 当前测试环境下主要剩余告警是 `KFK-004` 提示 `__transaction_state` 尚未出现
- 在“未使用事务”的场景下，这属于合理 `WARN`，不是阻塞性故障

## 当前阶段结论

本阶段完成后，`kdoctor` 的状态可以描述为：

- `V1` 设计中的关键收口项已经补齐
- probe 已达到可实际使用、结果基本可信
- 中文输出已经适合直接给运维同学阅读
- 报告噪声相比前几轮明显下降

换句话说，这一轮之后，`kdoctor` 已经进入：

**可以交给真实环境做初步上线使用测试的状态。**

## 下一阶段建议

如果后续进入 `V1.x` 持续迭代，建议优先看真实使用反馈，再决定是否继续推进：

1. 更细粒度的 commit / coordinator 归因
2. 更多 golden case
3. 更完整的 Markdown 样例与发布物沉淀
4. 调度层更强的 timeout / degrade 机制
5. 容量与趋势判断
