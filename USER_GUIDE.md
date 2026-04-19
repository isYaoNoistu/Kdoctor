# Kdoctor 用户使用手册

## 1. 手册目的

这份手册回答的是“拿到工具以后，怎么真的用起来”。

它重点解决四类问题：

- 只有一个 Kafka 地址时怎么查
- 有 `kdoctor.yaml`、`docker-compose.yml`、日志目录时怎么增强检查
- 报告里的 `通过 / 告警 / 失败 / 错误 / 跳过` 分别代表什么
- Windows 和 Linux 下命令、路径怎么写才不踩坑

如果你是第一次使用，建议按下面顺序阅读：

1. 先看“快速上手”
2. 再看“最常用命令”
3. 然后看“报告怎么读”
4. 最后按需要查“配置文件”和“常见问题”

## 2. Kdoctor 是什么

`Kdoctor` 是一个 Kafka 现场排障工具，不是管理平台，也不是监控平台。

它更像一个“值班现场的第一把刀”：

- 先判断问题大概落在网络、返回地址、KRaft、Topic/ISR、客户端链路，还是宿主机 / 容器 / 日志层
- 再把零散症状收敛成 1 到 3 条主因判断
- 最后给出一线能直接执行的下一步动作

## 3. 这版工具的边界

这次封版后的定位很明确：

- 聚焦 Kafka 内部使用场景
- 聚焦首轮巡检和排障加速
- 聚焦已有能力做稳、做准、做短

这版默认保留：

- 网络
- Kafka metadata / broker / internal topics
- KRaft
- Topic / ISR / leader / 规划
- Client probe
- Compose lint
- Docker
- Host
- Logs
- 事务上下文

这版默认移除：

- JMX / Metrics / JVM / Quota 检查
- 依赖 JMX 的默认输出和 `SKIP` 噪声

## 4. 快速上手

### 4.1 最小可用方式

如果你手里只有一个 Kafka 地址，直接这样跑。

Windows：

```powershell
.\kdoctor.exe probe --bootstrap 192.168.1.1:9192
```

Linux：

```bash
./kdoctor probe --bootstrap 192.168.1.1:9192
```

这条命令已经能做：

- bootstrap 连通性
- metadata 拉取
- broker 返回地址探测
- Topic / leader / ISR 检查
- metadata / produce / consume / commit probe

### 4.2 使用配置文件

如果你有一份常用环境配置，最推荐直接走 `--config`。

Windows：

```powershell
.\kdoctor.exe probe --config .\kdoctor.yaml
```

Linux：

```bash
./kdoctor probe --config ./kdoctor.yaml
```

### 4.3 再加上 compose 做增强检查

如果你还有 `docker-compose.yml`，可以继续增强成配置对照模式。

Windows：

```powershell
.\kdoctor.exe probe --config .\kdoctor.yaml --compose .\docker-compose.yml
```

Linux：

```bash
./kdoctor probe --config ./kdoctor.yaml --compose ./docker-compose.yml
```

这会额外补上：

- `listeners`
- `advertised.listeners`
- `controller.quorum.voters`
- `node.id`
- `process.roles`
- `inter.broker.listener.name`
- 容器名 / 挂载 / 持久化相关检查

## 5. Windows 和 Linux 的关键区别

### 5.1 路径写法

Windows 常见写法：

```powershell
.\kdoctor.exe
.\kdoctor.yaml
.\docker-compose.yml
.\report.md
```

Linux 常见写法：

```bash
./kdoctor
./kdoctor.yaml
./docker-compose.yml
./report.md
```

### 5.2 一个最常见坑

Linux 下不要把路径写成：

```bash
.\kdoctor.yaml
```

请写成：

```bash
./kdoctor.yaml
```

这是最容易导致“明明传了 `--config`，结果还是走默认 profile”的原因之一。

## 6. 运行模式

### 6.1 `quick`

快速巡检模式。

适合：

- 日常看一眼集群是否大致正常
- 不想跑真实 probe，只想先看基础健康情况

说明：

- 默认不会执行完整 client probe
- 终端输出更短

### 6.2 `probe`

最常用模式。

适合：

- 排障现场
- 希望确认真实链路是不是通的
- 希望看到 `metadata -> produce -> consume -> commit` 整条链路

### 6.3 `lint`

偏静态配置和部署对照。

适合：

- 有 `compose`
- 更关注配置是否写错

### 6.4 `full`

尽量执行完整检查。

适合：

- 需要比较完整的留档
- 想把 probe、配置、日志、Docker、Host 一次性都跑出来

### 6.5 `incident`

更偏故障现场的摘要模式。

适合：

- 当前已经有故障
- 你更想看最值得先处理的主因，而不是看所有细节

## 7. 最常用命令

### 7.1 最小探测

Windows：

```powershell
.\kdoctor.exe probe --bootstrap 192.168.1.1:9192
```

Linux：

```bash
./kdoctor probe --bootstrap 192.168.1.1:9192
```

### 7.2 多个 bootstrap

Windows：

```powershell
.\kdoctor.exe probe --bootstrap 192.168.1.1:9192,192.168.1.1:9194,192.168.1.1:9196
```

Linux：

```bash
./kdoctor probe --bootstrap 192.168.1.1:9192,192.168.1.1:9194,192.168.1.1:9196
```

### 7.3 使用 profile

Windows：

```powershell
.\kdoctor.exe probe --bootstrap 192.168.1.1:9192 --profile generic-bootstrap
```

Linux：

```bash
./kdoctor probe --bootstrap 192.168.1.1:9192 --profile generic-bootstrap
```

### 7.4 使用 config

Windows：

```powershell
.\kdoctor.exe probe --config .\kdoctor.yaml
```

Linux：

```bash
./kdoctor probe --config ./kdoctor.yaml
```

### 7.5 输出 JSON

Windows：

```powershell
.\kdoctor.exe probe --config .\kdoctor.yaml --json
```

Linux：

```bash
./kdoctor probe --config ./kdoctor.yaml --json
```

### 7.6 输出 Markdown 报告

Windows：

```powershell
.\kdoctor.exe probe --config .\kdoctor.yaml --format markdown --output .\report.md
```

Linux：

```bash
./kdoctor probe --config ./kdoctor.yaml --format markdown --output ./report.md
```

### 7.7 展开全部明细

默认终端报告会折叠 `PASS / SKIP`。

如果你要看完整明细，加 `--verbose`：

Windows：

```powershell
.\kdoctor.exe probe --config .\kdoctor.yaml --verbose
```

Linux：

```bash
./kdoctor probe --config ./kdoctor.yaml --verbose
```

## 8. 配置文件怎么写

最常用的是根目录这份：

- `kdoctor.yaml`

完整示例在：

- `kdoctor.example.yaml`

### 8.1 最小配置示例

```yaml
version: 2

default_profile: generic-bootstrap

profiles:
  generic-bootstrap:
    bootstrap_external:
      - "192.168.1.1:9192"
```

### 8.2 生产环境常用配置示例

```yaml
version: 2

default_profile: prod

profiles:
  prod:
    bootstrap_internal:
      - "192.168.119.7:9192"
      - "192.168.119.7:9194"
      - "192.168.119.7:9196"
    controller_endpoints:
      - "192.168.119.7:9193"
      - "192.168.119.7:9195"
      - "192.168.119.7:9197"
    broker_count: 3
    expected_min_isr: 2
    expected_replication_factor: 3
    execution_view: "internal"

docker:
  enabled: true
  container_names: ["kafka1", "kafka2", "kafka3"]

logs:
  enabled: true
  lookback_minutes: 30
  tail_lines: 500

probe:
  enabled: true
  topic: "_kdoctor_probe"

output:
  max_evidence_items: 8
  show_pass_checks: false
  show_skip_checks: false
  verbose: false
```

## 9. 关键参数说明

### 9.1 `profiles.*`

最重要的字段：

- `bootstrap_external`
- `bootstrap_internal`
- `controller_endpoints`
- `broker_count`
- `expected_min_isr`
- `expected_replication_factor`
- `execution_view`

### 9.2 `docker.*`

最常用：

- `enabled`
- `compose_file`
- `container_names`
- `inspect_mounts`

### 9.3 `logs.*`

这版封版后推荐值：

- `lookback_minutes: 30`
- `tail_lines: 500`
- `max_files: 12`
- `max_bytes_per_source: 1048576`
- `custom_patterns_dir: ""`

### 9.4 `probe.*`

最关键：

- `topic`
- `timeout`
- `message_bytes`
- `produce_count`

封版默认值：

- `probe.topic = _kdoctor_probe`
- `probe.message_bytes = 1024`

### 9.5 `output.*`

这版封版后新增并建议保留默认值：

- `output.max_evidence_items = 8`
- `output.show_pass_checks = false`
- `output.show_skip_checks = false`
- `output.verbose = false`

含义：

- 默认终端不被 PASS / SKIP 淹没
- 单个检查最多展示前 8 条证据
- 如果真要全量明细，再显式打开 `--verbose`

## 10. 报告怎么读

默认终端输出骨架是：

1. 页头摘要
2. 证据覆盖
3. 主因判断
4. 建议动作
5. 重点问题
6. 附加错误

### 10.1 总体状态

- `通过`：没有发现问题
- `告警`：存在风险，但不一定已经形成硬故障
- `失败`：已经有明确异常
- `严重`：高危异常，应优先处理
- `错误`：当前检查执行失败或关键前提缺失
- `跳过`：本次运行没有纳入这个检查，或者当前证据不足

### 10.2 证据覆盖

这版开始，覆盖语义只看三种状态：

- `已启用，已获取证据`
- `已启用，未获取证据`
- `未纳入本次运行`

重点理解：

- “已启用”不等于“已经拿到了有用证据”
- 如果某模块没纳入本次运行，就不应该在默认报告里制造大段 `SKIP`

### 10.3 主因判断

主因判断不是简单重复检查项，而是把同源问题做归并。

例如：

- `NET-003 + NET-005 + KFK-005`
  可能会收敛成“metadata 返回地址与当前客户端视角不匹配”
- `TOP-006 + TOP-007 + TOP-009`
  可能会收敛成“复制 / leader 风险”

### 10.4 重点问题

默认只展开：

- `CRIT`
- `FAIL`
- `WARN`
- `ERROR`

`PASS / SKIP` 只在：

- `--verbose`
- Markdown
- JSON

里看完整明细。

## 11. Probe 结果怎么理解

`probe` 模式下最关键的是这 5 项：

- `CLI-001`：metadata
- `CLI-002`：produce
- `CLI-003`：consume
- `CLI-004`：commit
- `CLI-005`：end-to-end

### 11.1 一个常见误解

如果 `CLI-002` 失败，后面的 `CLI-003~005` 很可能不会继续执行。

这不是工具坏了，而是为了避免把一个上游失败扩散成一堆重复 `FAIL`。

现在它会：

- 在失败阶段给出真正的 `FAIL`
- 对未执行阶段给出合理 `SKIP`

## 12. 日志结果怎么理解

这版日志模块最重要的变化是：不再把“采集成功”直接等价成“日志健康”。

### 12.1 `LOG-001`

表示“日志来源与样本质量”。

它会重点看：

- `source`
- `line_count`
- `byte_count`
- `latest_timestamp`
- `freshness`
- `sample_sufficient`

### 12.2 `LOG-002`

只表示：

- 有没有命中已知错误指纹

它不再暗示：

- “没命中就一定健康”

## 13. 常见排障入口

### 13.1 bootstrap 可达，但 probe produce 失败

优先看：

- `NET-003`
- `NET-005`
- `KFK-005`
- `CLI-002`

这类问题最常见是：

- `advertised.listeners` 返回地址不对
- 端口没有真正暴露
- 客户端所在网络和 broker 返回地址不匹配

### 13.2 metadata 正常，但消费组异常

优先看：

- `CSM-001`
- `CSM-002`
- `CSM-006`
- `KFK-004`

常见方向：

- `__consumer_offsets`
- coordinator
- rebalance 风暴

### 13.3 看起来像 Kafka 问题，其实是宿主机 / 容器问题

优先看：

- `HOST-*`
- `DKR-*`
- `STG-*`
- `LOG-*`

## 14. FAQ

### 14.1 为什么我明明传了 `--config`，结果还是走默认 profile？

最常见原因：

- Linux 下把路径写成了 `.\kdoctor.yaml`
- 配置文件路径根本没读到
- 配置文件里 `default_profile` 不对

建议直接用：

Linux：

```bash
./kdoctor probe --config ./kdoctor.yaml
```

Windows：

```powershell
.\kdoctor.exe probe --config .\kdoctor.yaml
```

### 14.2 为什么有些检查没有出现？

因为这版封版前做了“按输入收口注册”。

例如：

- 没有 `compose`，就不会默认注册整批 compose lint
- 没有日志来源，就不会注册整批日志明细
- 没有消费组目标，就不会注册消费组检查

这样做是为了减少默认报告噪声。

### 14.3 为什么默认看不到 PASS / SKIP？

因为默认终端是给值班人员看的，不是给开发者看所有细节的。

如果你要完整明细：

```bash
./kdoctor probe --config ./kdoctor.yaml --verbose
```

或者：

```powershell
.\kdoctor.exe probe --config .\kdoctor.yaml --verbose
```

### 14.4 这版为什么没有 JMX 相关输出了？

因为封版前已经明确移除了这条路径。

原因不是“功能做不出来”，而是：

- 现实环境无法稳定启用
- 默认报告里只会制造背景噪声
- 对内部使用价值反而是负面

## 15. 构建与打包

在仓库根目录执行：

```powershell
go test ./...
.\scripts\build.ps1
.\scripts\build.ps1 -GOOS linux -GOARCH amd64
```

构建产物会输出到工作区根目录 `dist/`。

## 16. 最后建议

第一次用时，不要一上来把所有输入都堆满。

建议顺序是：

1. 先跑 `probe --bootstrap`
2. 再补 `--config`
3. 最后再补 `--compose`、日志目录、Docker 视角

这样最容易看清：

- 基础链路到底通不通
- 额外输入到底是在增强证据，还是在引入背景噪声
