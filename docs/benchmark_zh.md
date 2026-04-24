# 基准测试

本页总结了 PinchTab 与 `agent-browser` 在真实代理循环令牌成本方面的比较。完整的方法、每次运行的表格和原始记录存放在 [基准测试深度分析](./deep-dive/benchmark.md) 中。

## 主要结果

PinchTab 在我们测量的每个范围内都比 agent-browser 更便宜，使用的 API 往返更少。以下百分比表示 "PinchTab 在这个指标上比 agent-browser 便宜 N%"：

| 范围                      | 每通道 n | 成本更便宜 | 更少请求 | 更少令牌 |
|----------------------------|-----------:|-------------:|---------------:|-------------:|
| 基础 Haiku 4.5（10 步） | 5          | **9.5%**     | 23.0%          | 17.9%        |
| 扩展 Haiku 4.5（24 步） | 3       | **19.6%**    | 31.1%          | 26.2%        |
| 扩展 Sonnet 4.6（24 步） | 2      | **20.3%**    | 29.4%          | 25.3%        |

每次运行的绝对成本：

| 范围                      | PinchTab 平均 | agent-browser 平均 |
|----------------------------|-------------:|------------------:|
| 基础 Haiku 4.5            | $0.1024      | $0.1132           |
| 扩展 Haiku 4.5         | $0.3516      | $0.4372           |
| 扩展 Sonnet 4.6        | $0.8932      | $1.1204           |

## 测量内容

该数字是**整个代理循环的端到端令牌成本** —— 系统提示、技能、工具调用、工具输出、模型推理和重试 —— 在一次完整的基准测试运行中求和。使用情况直接从 Anthropic 的每个响应的 `usage` 对象读取；模型不进行自我报告。

两个通道在相同的 Docker Compose 环境中运行相同的任务集，针对相同的基准测试固定服务器，由相同的 Go 运行器驱动。通道之间唯一的变化是代理与之通信的浏览器表面以及教导它命令形状的匹配技能。

## PinchTab 在成本方面获胜的原因

两个结构差异驱动了差距：

1. **更少的 API 往返**。agent-browser 遵循点击然后快照的模式：每个突变步骤成本两个 API 调用。PinchTab 通过 `--snap`/`--snap-diff` 将动作和结果快照批处理为一次往返，因此相同的步骤成本一个 API 调用。
2. **更少的重复缓存读取**。agent-browser 上的这些额外往返不仅每次成本一轮 —— 它们还在每轮重新读取缓存的系统提示和技能。在 24 步运行中，额外的缓存读取令牌主导了令牌差距（尽管不是成本差距，因为缓存读取仅为未缓存输入定价的 10%）。

## 差距如何扩展

- **范围**：差距随着步骤计数而扩大（10 步时为 9.5%，24 步时为 19.6%）。涉及操作后快照的每一个额外步骤都会在 agent-browser 上增加另一次往返。
- **模型**：在扩展范围内，Haiku 4.5 和 Sonnet 4.6 的差距基本相同（19.6% 对 20.3%）。更强的推理不会崩溃点击→快照模式 —— 额外的往返是工具表面的属性，而不是模型纠正的规划失败。

## 注意事项

- 10 步任务套件是与 PinchTab 的开发一起设计的，包含在 agent-browser 上笨拙多调用的任务。共同设计或更大的任务集将减少任务套件偏见。
- 两个通道都运行其完整技能的修剪子集（标题 + 代理实际使用的一个参考文件），以使比较集中在工具表面而不是文档权重上。在两侧使用完整技能的生产重新运行将给出不同的数字。
- 每次运行级别的方差为平均值的 ~25–30%；n=5 基础 / n=3 扩展 Haiku / n=2 扩展 Sonnet 给出可用的中心趋势，但置信区间较宽，尤其是 Sonnet 对。
- agent-browser 在每次扩展运行中都有一个异常值（lae3）；排除它会有意义地缩小差距。

## 重现

```bash
cd tests/benchmark
docker compose up -d --build
./scripts/run-optimization.sh

# 基线（确定性，~30s）
./scripts/baseline.sh

# PinchTab 和 agent-browser 通道（需要 Anthropic API 密钥）
ANTHROPIC_API_KEY=... ./scripts/run-api-benchmark.ts --lane pinchtab --groups 0,1
ANTHROPIC_API_KEY=... ./scripts/run-api-benchmark.ts --lane agent-browser --groups 0,1

# 检查使用情况
jq '.run_usage' results/pinchtab_benchmark_*.json
jq '.run_usage' results/agent_browser_benchmark_*.json
```

有关每次运行的表格、原始记录、令牌分解、方差讨论和完整的测量注意事项列表，请参阅 [基准测试深度分析](./deep-dive/benchmark.md)。