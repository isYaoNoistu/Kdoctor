# Kdoctor 架构与工程标准

## 1. 文档定位

本文档定义 `Kdoctor` 在 **V2 封版状态** 下的工程边界与实现标准。

它不再讨论“要不要继续扩功能”，而是回答：

- 代码如何组织
- 模块如何协作
- 哪些依赖方向是允许的
- 什么样的改动才符合封版后的工程标准

## 2. 当前架构目标

封版后的 `Kdoctor` 需要保持这些特征：

- 单二进制交付
- `bootstrap-only` 可运行
- `profile / compose / docker / logs / host` 为增强输入
- 默认中文输出
- 默认终端低噪声
- 单项数据源失败不拖死整体
- 以“可信度优先”替代“功能越多越好”

## 3. 标准目录结构

```text
Kdoctor/
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
    buildinfo/
    model/
  scripts/
  version/
  README.md
  USER_GUIDE.md
  doc.md
  architecture.md
```

约束：

- `cmd/kdoctor` 只放 CLI 入口
- `internal` 放业务实现
- `pkg/model` 放稳定报告模型
- `pkg/buildinfo` 放版本与提交信息
- `version` 只放阶段文档与 release 基线

## 4. 模块职责

### 4.1 `cmd/kdoctor`

职责：

- 参数解析
- `--version`
- 进程退出码

禁止：

- 直接写检查逻辑
- 直接访问 Kafka / Docker / Shell

### 4.2 `internal/app`

职责：

- 配置装配
- profile 选择
- 输出格式分发
- 驱动主流程

### 4.3 `internal/config`

职责：

- 配置结构
- 默认值
- merge 语义
- 校验
- 运行时展开

### 4.4 `internal/runner`

职责：

- 组织 collector / checks / diagnose / output
- 控制 task timeout
- 处理 soft degrade
- 生成覆盖摘要

### 4.5 `internal/collector`

职责：

- 采集外部事实
- 构造 snapshot

约束：

- 不直接产出 `PASS/WARN/FAIL`
- 不直接做终端渲染
- `Collected` 与 `Available` 语义必须分开

### 4.6 `internal/checks`

职责：

- 基于 snapshot 做规则判断
- 产出稳定 `CheckResult`

约束：

- 一个检查器对应一个稳定问题域
- 不在检查器之间做链式调用
- 证据只输出真正命中的项

### 4.7 `internal/probe`

职责：

- 跑 `metadata -> topic-ready -> produce -> consume -> commit -> e2e`
- 记录阶段边界
- 对自动创建 probe topic 做最小副作用控制

### 4.8 `internal/diagnose`

职责：

- 根因归并
- 动作收敛
- incident 摘要

约束：

- 优先做同源归并
- 不把上下文提示抬成主因

### 4.9 `internal/output`

职责：

- `terminal / json / markdown`

约束：

- 渲染层不再做业务判断
- 默认终端折叠 `PASS / SKIP`
- JSON 字段结构稳定

### 4.10 `internal/localize`

职责：

- 报告中文化
- 术语统一
- 乱码与中英混排收口

### 4.11 `internal/transport`

职责：

- 封装 Kafka / TCP / Docker / Disk / Shell 调用

约束：

- transport 只负责访问与返回
- 不负责业务结论

## 5. 数据流

主流程固定为：

1. CLI 解析参数
2. App 装配运行时配置
3. Runner 组织采集
4. Collectors 产出 Snapshot
5. Checks 产出 CheckResult
6. Diagnose 汇总主因与动作
7. Output 渲染为终端 / JSON / Markdown

## 6. 依赖方向

允许依赖：

```text
cmd -> internal/app -> internal/runner
runner -> collector / checks / diagnose / output
collector -> parser / snapshot / transport
checks -> snapshot / rule / model
output -> model / localize
pkg/model -> 不反向依赖 internal
```

禁止反向依赖：

- output 反向依赖 collector
- checks 直接调用 transport 做外部访问
- cmd 绕过 app / runner 直接拼装检查链

## 7. 报告标准

### 7.1 报告层

至少包含：

- `mode`
- `profile`
- `checked_at`
- `elapsed_ms`
- `summary`
- `checks`
- `exit_code`
- `tool_version`
- `schema_version`

### 7.2 摘要层

至少包含：

- 总体状态
- broker 总数 / 存活数
- 概览
- 证据覆盖
- 主因判断
- 建议动作

### 7.3 检查项

至少包含：

- `id`
- `module`
- `status`
- `summary`
- `evidence`
- `possible_causes`
- `next_actions`

## 8. 默认输出标准

### 8.1 Terminal

- 默认只展开重点问题
- `PASS / SKIP` 默认折叠
- 证据去重并截断
- 适合值班现场快速扫读

### 8.2 JSON

- 字段稳定
- 向自动化友好
- 新增字段只能向后兼容地加

### 8.3 Markdown

- 适合留档与工单
- 章节固定
- 与 terminal 保持同语义

## 9. 封版后的硬约束

- 默认链路里不再注册额外指标扩展检查
- 覆盖摘要按“有无证据”展示，不按“是否尝试过采集”展示
- `TOP-011` 只输出真正命中的 topic
- 同一检查内证据必须可去重、可截断
- 中文输出无乱码、无显著中英混排
- 版本、二进制、README、用户手册必须对齐
