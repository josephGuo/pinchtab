---
name: pinchtab-opt
description: "运行 PinchTab 优化循环。生成盲代理，仅使用 PinchTab 技能执行 39 个组的 87 个浏览器自动化步骤，然后报告通过/失败结果以及与基线相比的操作计数。当被要求'运行优化'、'运行优化循环'、'基准测试代理'或'测试 pinchtab 代理'时使用。"
---

# PinchTab 优化循环

对 87 个浏览器自动化步骤（39 个组）运行盲代理，以衡量 AI 代理在无需手动提供的选择器的情况下驱动 PinchTab 的能力。

## 路径解析

以下所有路径都相对于**项目根目录**（git 根目录）。首先解析它：

```bash
PROJECT_ROOT=$(git rev-parse --show-toplevel)
TOOLS_DIR="$PROJECT_ROOT/tests/tools"
```

子代理必须以 `$TOOLS_DIR` 作为工作目录运行，因为 `./scripts/pt` 和 `./scripts/runner` 位于那里。

## 前提条件

Docker 服务必须正在运行。在生成代理之前验证：

```bash
$TOOLS_DIR/scripts/pt health
```

如果不健康，启动服务：

```bash
docker compose -f "$TOOLS_DIR/docker-compose.yml" up -d --build
```

等待几秒钟并重新检查健康状态。

## 执行

### 0. 创建每个代理的报告文件

在生成代理之前，创建隔离的报告文件，以便并发写入不会损坏共享文件：

```bash
RESULTS_DIR="$TOOLS_DIR/../benchmark/results"
TIMESTAMP=$(date -u +%Y%m%d_%H%M%S)
mkdir -p "$RESULTS_DIR"

for agent in A B C; do
  cat > "$RESULTS_DIR/agent${agent}_${TIMESTAMP}.json" <<SEED
{
  "benchmark": {"type": "pinchtab", "timestamp": "${TIMESTAMP}", "agent": "${agent}"},
  "totals": {"steps_answered": 0},
  "steps": []
}
SEED
done
```

保存这三个文件路径——您将把一个传递给每个子代理。

### 1. 生成 3 个并行子代理

使用带有 `run_in_background: true` 的 **Agent** 工具。将 39 个组分成三个批次：

- **批次 A**：组 0-12（39 步）
- **批次 B**：组 13-25（26 步）
- **批次 C**：组 26-38（22 步）

每个子代理获得**相同的提示模板**——只有组范围和 `{REPORT_FILE}` 会改变。用实际值替换 `{START}`、`{END}`、`{START_PAD}`、`{END_PAD}`、`{PROJECT_ROOT}` 和 `{REPORT_FILE}`：

```
您正在运行 PinchTab 优化任务。您的任务是执行组 {START} 到 {END}。

关键：您的工作目录必须是所有命令的 {PROJECT_ROOT}/tests/tools。在每个 shell 命令前加上 `cd {PROJECT_ROOT}/tests/tools && `。

您的报告文件是：{REPORT_FILE}
在每次 `./scripts/runner step-end` 调用时使用 `--report-file {REPORT_FILE}`。

首先阅读这些文件以了解您的工具和任务：
1. 阅读 `{PROJECT_ROOT}/tests/optimization/subagent-context.md` — 环境、封装器和记录格式。
2. 阅读 `{PROJECT_ROOT}/skills/pinchtab/SKILL.md` — 完整的 PinchTab 命令参考。
3. 从 `{PROJECT_ROOT}/tests/optimization/group-{START_PAD}.md` 到 `{PROJECT_ROOT}/tests/optimization/group-{END_PAD}.md` 阅读每个组文件。

不要读取 `{PROJECT_ROOT}/tests/tools/scripts/baseline.sh` 或 `{PROJECT_ROOT}/tests/benchmark/` 下的任何文件。

阅读上述文件后，按顺序执行每个组中的每个步骤：
- 运行命令前始终 cd 到 {PROJECT_ROOT}/tests/tools。
- 使用 `./scripts/pt` 作为所有 PinchTab 命令的封装器。
- 每步之后，使用 `./scripts/runner step-end --report-file {REPORT_FILE} <group> <step> answer "<observation>" pass "notes"` 记录结果（如果不起作用则使用 fail）。
- 根据技能文档判断正确的 PinchTab 命令。组文件描述做什么，而不是怎么做。

完成组 {START}-{END} 中的每个步骤。不要跳过任何步骤。
```

### 2. 监控进度

代理运行时，定期计算每个代理输出文件中的 step-end 记录数：

```bash
grep -c "step-end" <output_file>
```

预期总数：批次 A ~39，批次 B ~26，批次 C ~22 = 总共 87。

### 3. 收集和汇总

所有 3 个代理完成后，使用 `./scripts/runner` 子命令合并报告（`merge-reports`）、从子代理记录中注入令牌使用量（`inject-usage`）并打印最终比较表（`opt summarize`）。将汇总输出原样呈现给用户。

## 参考数字

- **基线**：87/87 步，246 次浏览器操作，2.8 操作/步
- **预期代理范围**：350-500 次浏览器操作，4-6 操作/步（代理必须在行动前探索页面）
- **组数**：39 个组，共 87 步

## 文件位置（相对于项目根目录）

| 路径 | 用途 |
|------|------|
| `tests/optimization/subagent-context.md` | 子代理说明（环境、封装器、记录） |
| `tests/optimization/index.md` | 组列表 |
| `tests/optimization/group-00.md` .. `group-38.md` | 任务描述 |
| `skills/pinchtab/SKILL.md` | PinchTab 命令参考（由子代理读取） |
| `tests/tools/scripts/pt` | PinchTab 封装器（工作目录必须是 `tests/tools`） |
| `tests/tools/scripts/runner` | 步骤记录器（工作目录必须是 `tests/tools`） |
| `tests/tools/scripts/baseline.sh` | 基线（子代理不得读取此文件） |