# 基准测试

本页总结了 PinchTab 与 `agent-browser` 在真实代理循环令牌成本上的对比。完整方法论、每次运行的表格以及原始记录见[基准测试深度分析](./deep-dive/benchmark.md)。

## 主要结果

在我们测量的每一个范围内，PinchTab 都比 agent-browser 更便宜，且 API 往返次数更少。下表中的百分比读作"PinchTab 在该指标上比 agent-browser 便宜 N%"：

| 范围                        | 每条通道 n | 成本更低     | 请求更少  | 令牌更少  |
| ------------------------- | ------ | -------- | ----- | ----- |
| 基础 Haiku 4.5（10 步）       | 5      | **9.5%** | 23.0% | 17.9% |
| 扩展 Haiku 4.5（24 步）       | 3      | **19.6%** | 31.1% | 26.2% |
| 扩展 Sonnet 4.6（24 步）      | 2      | **20.3%** | 29.4% | 25.3% |

每次运行的绝对成本：

| 范围            | PinchTab 平均 | agent-browser 平均 |
| ------------- | ----------- | ---------------- |
| 基础 Haiku 4.5  | $0.1024     | $0.1132          |
| 扩展 Haiku 4.5  | $0.3516     | $0.4372          |
| 扩展 Sonnet 4.6 | $0.8932     | $1.1204          |

## 测量的是什么

这个数字是**整个代理循环的端到端令牌成本**——系统提示、技能、工具调用、工具输出、模型推理以及重试——在一次完整基准测试运行中的总和。用量直接取自 Anthropic 每个响应的 `usage` 对象；不依赖模型自我报告。

两条通道在同一个 Docker Compose 环境内、针对同一台基准测试夹具服务器运行相同的任务集，并由同一个 Go 运行器驱动。两条通道之间唯一变化的，是代理与之交互的浏览器接口，以及教会它命令形式的匹配技能。

## 为什么 PinchTab 在成本上占优

两个结构性差异造成了这一差距：

1. **更少的 API 往返。** agent-browser 遵循"先点击再快照"的模式：每一步变更操作都要花两次 API 调用。PinchTab 通过 `--snap`/`--snap-diff` 把动作和随之产生的快照合并为一次往返，因此同一步只需要一次 API 调用。
2. **更少的重复缓存读取。** agent-browser 上那些额外的往返不只是每一步多花一轮——它们还会在每一轮重新读取缓存的系统提示和技能。在一次 24 步的运行中，额外的缓存读取令牌主导了令牌差距（尽管不是成本差距，因为缓存读取只按未缓存输入定价的 10% 计费）。

## 差距如何随规模变化

- **范围：** 差距随步数增加而扩大（10 步时 9.5%，24 步时 19.6%）。每一个涉及动作后快照的额外步骤，都会在 agent-browser 上再增加一次往返。
- **模型：** 在扩展范围下，Haiku 4.5 与 Sonnet 4.6 上的差距基本相同（19.6% 对 20.3%）。更强的推理并不会瓦解"点击→快照"这一模式——额外的往返是工具接口的属性，而不是模型能够纠正的规划失误。

## 注意事项

- 这套 10 步任务集是与 PinchTab 的开发一同设计的，其中包含在 agent-browser 上显得笨拙、需要多次调用的任务。一个共同设计或大得多的任务集能减少任务集偏差。
- 两条通道都只运行各自完整技能的裁剪子集（标题 + 代理实际会用到的那一个参考文件），以便让比较聚焦于工具接口而非文档体量。两侧都使用完整技能重新运行生产测试会得到不同的数字。
- 在单次运行层面，方差约为均值的 25–30%；n=5 基础 / n=3 扩展 Haiku / n=2 扩展 Sonnet 能给出可用的集中趋势，但置信区间较宽，Sonnet 那一对尤其如此。
- agent-browser 在每次扩展运行中都有一个离群值（lae3）；把它排除会显著缩小差距。

## 复现

```bash
# From the repo root. The benchmark runs in Docker and the entry points
# live in tests/tools/, not tests/benchmark/.

# Deterministic baseline (no API key required)
./dev opt baseline

# PinchTab and agent-browser lanes (Anthropic API key required)
ANTHROPIC_API_KEY=... ./dev bench pinchtab --groups 0,1
ANTHROPIC_API_KEY=... ./dev bench agent-browser --groups 0,1

# Inspect usage (results land under tests/benchmark/results/)
jq '.run_usage' tests/benchmark/results/pinchtab_benchmark_*.json
jq '.run_usage' tests/benchmark/results/agent_browser_benchmark_*.json
```

shell 入口点和 Docker compose 文件位于 `tests/tools/scripts/` 和 `tests/tools/docker-compose*.yml` 下；代理循环是位于 `tests/tools/runner/` 的 Go 二进制文件。

每次运行的表格、原始记录、令牌分解、方差讨论以及完整的测量注意事项列表，见[基准测试深度分析](./deep-dive/benchmark.md)。
