---
name: pinchtab-coldstart
description: "运行 PinchTab 冷启动测试。生成一个子代理，从源代码构建 PinchTab，本地启动服务器（不使用 Docker），并仅使用技能文档执行组 0-1（14 个步骤）。当被要求'运行冷启动'、'冷启动测试'或'测试代理入职流程'时使用。"
---

# PinchTab 冷启动测试

验证 AI 代理能否从零开始仅使用技能文档与 PinchTab 一起工作——无需手把手指导。

## 前提条件

在生成子代理之前，清理环境：

```bash
PROJECT_ROOT=$(git rev-parse --show-toplevel)
pkill -f 'pinchtab' 2>/dev/null
pkill -f 'Google Chrome.*pinchtab' 2>/dev/null
sleep 2
rm -f ~/.local/state/pinchtab/current-tab 2>/dev/null
rm -f "$PROJECT_ROOT/pinchtab" 2>/dev/null
lsof -ti:9867 2>/dev/null | xargs kill 2>/dev/null
```

清理后等待 2 秒再生成代理。

## 执行

使用下面的提示生成单个子代理。用实际值替换 `{PROJECT_ROOT}` 和 `{TIMESTAMP}`。

```
您正在运行 PinchTab 冷启动验证。您的工作目录是 {PROJECT_ROOT}。

首先读取上下文文件，然后按照其说明操作：

1. 读取 `tests/coldstart/subagent-context.md` — 您的完整说明。
2. 读取它引用的技能文件。
3. 读取它引用的组文件。
4. 执行组 0 和组 1 中的所有步骤。

报告每个步骤的通过/失败。将完整结果写入 `/tmp/pinchtab-coldstart-{TIMESTAMP}.md`。
```

## 结果解释

- **14/14 通过**：技能文档足以完成冷启动。
- **任何失败**：检查代理卡在哪个步骤——这是技能文档或命令行界面易用性的差距。

在报告中要关注的关键事项：
- 代理使用的是默认端口（9867）还是自定义端口？
- 代理读取服务器的 READY 输出还是循环轮询健康状态？
- 代理使用的是 `./pinchtab` 命令行界面还是回退到 curl/HTTP API？
- 代理修改了 `~/.pinchtab/config.json` 还是使用带 `PINCHTAB_CONFIG` 的临时配置？
- 步骤 1.2（点击跟随链接）是否在没有 eval 解决方法的情况下通过？

## 比较运行

跟踪各次运行的令牌使用量和工具调用次数以衡量改进：

| 指标 | 良好 | 需要改进 |
|------|------|----------|
| 总令牌数 | < 40k | > 50k |
| 工具调用 | < 40 | > 50 |
| 端口 | 9867（默认） | 自定义端口 |
| 服务器等待 | 读取 READY | 轮询健康状态 |
| API 使用 | 仅命令行界面 | curl/HTTP 回退 |