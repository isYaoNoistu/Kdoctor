# Kdoctor 架构与工程标准

## 1. 文档定位

本文档定义 `Kdoctor` 的工程标准，目标是保证后续开发、重构、评审和测试都围绕同一套结构进行。

它回答的是“代码该怎么组织、模块该如何协作、什么样的改动算符合工程标准”。

## 2. 工程目标

`Kdoctor` 应当保持以下特征：

- 单二进制交付
- 默认可在 Linux 服务器和运维工作站场景使用
- 支持 `bootstrap-only` 最小输入
- 支持分层增强输入
- 输出稳定、可读、可自动化
- 单项数据源失败不会轻易拖死整体

## 3. 标准目录结构

```text
kdoctor/
  cmd/
    kdoctor/
      main.go
  internal/
    app/
    checks/
    collector/
    config/
    diagnose/
    exitcode/
    localize/
    output/
    parser/
    probe/
    profile/
    runner/
    rule/
    snapshot/
    transport/
  pkg/
    model/
  dist/
  scripts/
  version/
  README.md
  doc.md
  architecture.md
```

目录约束：

- `cmd/kdoctor` 只放 CLI 入口。
- `internal` 放业务实现。
- `pkg/model` 放稳定模型。
- `dist` 只放构建产物。
- `version` 只放阶段记录。

## 4. 模块职责

### 4.1 `cmd/kdoctor`

职责：

- 解析命令行参数
- 初始化应用
- 负责退出码

禁止：

- 直接写检查逻辑
- 直接访问 Kafka、Docker、Shell

### 4.2 `internal/app`

职责：

- 装配配置、profile、运行时上下文
- 选择输出格式
- 驱动主流程

### 4.3 `internal/config`

职责：

- 配置结构定义
- 默认值
- 配置校验
- 运行时配置合并

### 4.4 `internal/profile`

职责：

- 管理内置环境模板
- 表达环境预期，而不是运行结果

### 4.5 `internal/runner`

职责：

- 编排阶段执行
- 聚合 collector / checks / diagnose / output 所需上下文
- 控制整体 timeout

要求：

- 允许数据源降级
- 保持主流程清晰

### 4.6 `internal/snapshot`

职责：

- 定义本轮检查统一快照
- 作为 collector 与 checks 之间的稳定边界

要求：

- 结构化优先
- 命名稳定
- 避免把渲染用文本塞进快照

### 4.7 `internal/collector`

职责：

- 向外部系统采集事实
- 构造快照

禁止：

- 直接产出 `PASS/WARN/FAIL`
- 直接做终端渲染

### 4.8 `internal/checks`

职责：

- 基于快照做规则判断
- 输出标准 `CheckResult`

要求：

- 一个检查器负责一个稳定问题域
- 每个检查器有唯一编号
- 不在检查器之间互相强耦合调用

### 4.9 `internal/probe`

职责：

- 执行 metadata / produce / consume / commit 探针
- 输出 probe 快照

要求：

- 这是少数允许写入副作用的模块
- 必须控制写入规模
- 不能复用业务消费组

### 4.10 `internal/diagnose`

职责：

- 汇总多个检查结果
- 进行主因判断
- 生成 incident 摘要

要求：

- 优先做相关性归并，而不是简单列出问题

### 4.11 `internal/output`

职责：

- 把统一 `Report` 渲染为 terminal / json / markdown

要求：

- 渲染层不再做业务判断
- 默认终端输出为中文

### 4.12 `internal/localize`

职责：

- 对报告做文本本地化
- 集中管理中文化映射

要求：

- 不把本地化逻辑散落到每个 checker

### 4.13 `internal/transport`

职责：

- 封装 Kafka、TCP、Docker、Shell 等外部访问

要求：

- transport 只处理调用与返回，不处理业务结论

## 5. 数据流标准

主流程应保持如下顺序：

1. CLI 解析参数
2. App 合并运行时配置
3. Runner 组织采集
4. Collectors 产出 Snapshot
5. Checks 基于 Snapshot 产出 CheckResult
6. Diagnose 汇总为 Summary / Root Causes / Actions
7. Output 渲染为终端、JSON 或 Markdown

禁止反向依赖：

- output 不能反向依赖 collector
- checks 不能直接控制 transport
- cmd 不能跳过 app / runner 直接调用底层包

## 6. 依赖规则

允许的依赖方向：

```text
cmd -> internal/app -> internal/runner
runner -> collector / checks / diagnose / output
collector -> parser / snapshot / transport
checks -> snapshot / rule / model
output -> model / localize
pkg/model -> 不反向依赖 internal
```

## 7. 报告标准

### 7.1 Summary

必须包含：

- 总体状态
- broker 总数与存活数
- 概览
- 主因判断
- 建议动作

### 7.2 CheckResult

必须包含：

- 编号
- 模块
- 状态
- 摘要

建议包含：

- 证据
- 可能原因
- 下一步动作

## 8. 中文输出标准

终端报告和 Markdown 报告必须满足：

- 标题、标签、状态使用中文
- 错误说明尽量中文化
- 保留必要技术名词，例如 Kafka、broker、KRaft、ISR

JSON 的键名可以保持稳定英文，但值应尽量中文化。

## 9. 测试标准

至少覆盖三层：

### 9.1 单元测试

- 规则检查器
- 归因层
- 渲染层

### 9.2 契约测试

- JSON 输出结构
- Markdown 输出关键段落
- 退出码映射

### 9.3 真实环境验证

至少要验证：

- `bootstrap-only`
- `compose` 增强模式
- JSON 输出
- Markdown 输出

## 10. 代码评审标准

任何改动进入主分支前，应至少确认：

- 是否破坏了 `bootstrap-only`
- 是否引入新的误报
- 是否保持了中文输出
- 是否破坏 JSON / Markdown 输出
- 是否更新了对应文档
- 是否补了必要测试

## 11. 文档标准

文档分工如下：

- `README.md`：工具定位与使用方法
- `doc.md`：设计目标与能力边界
- `architecture.md`：工程标准
- `version/*.md`：阶段记录

要求：

- 文档编码统一为 UTF-8
- 示例优先使用脱敏地址
- 阶段文档说明目标、完成项、验证、问题和下一步

## 12. 当前工程结论

当前 `kdoctor` 已进入“V1 已交付、开始稳定化优化”的阶段。

后续所有实现，应继续遵守这几个底线：

- 不把 `compose` 变成前置依赖
- 不牺牲误报控制换取表面覆盖率
- 不把本地化逻辑散落到各层
- 不让输出层重新承担诊断逻辑

