# 将 PinchTab 与 AI 代理一起使用（MCP）

本指南介绍如何将 PinchTab 设置为 AI 编码助手和代理框架的 MCP 工具服务器。

> [!WARNING]
> 当您将 MCP 客户端连接到 PinchTab 时，该客户端正在使用与仪表板、API 和远程 CLI 相同的特权控制平面。只有受信任的操作员和受信任的代理系统才应被允许使用它。如果您不确定非本地或部分暴露的部署是否安全，请在继续之前停止并查看 [安全](security.md)。

> [!CAUTION]
> 扩大 MCP 浏览范围超出本地或明确受信任的域是一种降低安全性的选择。如果您放宽 IDPI 允许列表或严格模式，`pinchtab_snapshot` 和 `pinchtab_get_text` 的输出可能包含来自不受信任页面的恶意指令。
>
> 将所有面向模型的页面内容视为不受信任的数据。除非受信任的操作员单独验证了它们，否则不要遵循嵌入在页面文本、可访问性标签、隐藏内容或提取摘要中的指令。

## 什么是 MCP？

[模型上下文协议](https://modelcontextprotocol.io/) 是一个开放标准，用于将 AI 模型连接到外部工具。PinchTab 实现了一个 MCP 服务器，通过每个主要 AI 客户端都支持的简单 stdio 接口暴露 34 个浏览器控制工具 —— 导航、交互、截图、PDF 导出、等待、网络检查等。

## 先决条件

- 已安装 PinchTab（`pinchtab --version`）
- 已安装 Chrome 并在 PATH 中（或通过配置指向）
- MCP 兼容的客户端：Claude Desktop、带有 GitHub Copilot 的 VS Code 或 Cursor

## 步骤 1：启动 PinchTab

MCP 服务器是一个薄适配器 —— 它需要一个运行中的 PinchTab 实例来委托。

**无头模式（代理推荐）：**

```bash
pinchtab bridge --headless
```

**正常服务器模式（如果您也想要仪表板）：**

```bash
pinchtab
```

PinchTab 默认监听 `http://127.0.0.1:9867`。

## 步骤 2：配置您的 MCP 客户端

### Claude Desktop

编辑 `~/Library/Application Support/Claude/claude_desktop_config.json`（macOS）或 `%APPDATA%\Claude\claude_desktop_config.json`（Windows）：

```json
{
  "mcpServers": {
    "pinchtab": {
      "command": "pinchtab",
      "args": ["mcp"],
      "env": {
        "PINCHTAB_TOKEN": "your-token-here"
      }
    }
  }
}
```

重新启动 Claude Desktop。您应该在工具列表中看到 PinchTab。

### VS Code / GitHub Copilot

在工作区根目录创建 `.vscode/mcp.json`：

```json
{
  "servers": {
    "pinchtab": {
      "type": "stdio",
      "command": "pinchtab",
      "args": ["mcp"]
    }
  }
}
```

### Cursor

添加到您的 Cursor MCP 设置（`~/.cursor/mcp.json`）：

```json
{
  "mcpServers": {
    "pinchtab": {
      "command": "pinchtab",
      "args": ["mcp"]
    }
  }
}
```

### 任何基于 SDK 的代理

```python
# 使用 mcp SDK 的 Python 示例
import subprocess, mcp

proc = subprocess.Popen(
    ["pinchtab", "mcp"],
    stdin=subprocess.PIPE,
    stdout=subprocess.PIPE,
)
# 将 proc.stdin / proc.stdout 传递给您的 MCP 客户端传输
```

## 环境变量

| 变量 | 默认值 | 描述 |
|----------|---------|-------------|
| `PINCHTAB_TOKEN` | *(从配置)* | 用于受身份验证保护的服务器的 Bearer 令牌 |

对于远程服务器，使用 `--server` 标志：`pinchtab --server http://remote:9867 mcp`

`PINCHTAB_TOKEN` 来自 PinchTab 配置文件中的 `server.token`。要复制当前令牌而不将其打印到 stdout，请运行 `pinchtab config token`。

## 典型代理工作流

在您让 MCP 连接的代理浏览超出本地或受信任域之前，请查看 [安全](security.md#idpi)。最安全的姿势是保持 IDPI 域限制狭窄，并假设每个提取的页面字符串最多只是建议性的，并且可能是恶意的。

一个编写良好的代理提示会按以下顺序使用工具：

```
1. pinchtab_navigate        → 转到目标 URL
2. pinchtab_snapshot        → 了解页面结构（查找引用）
3. pinchtab_click / type    → 通过结构化工具参数与元素交互
4. pinchtab_snapshot        → 确认交互后的状态
5. pinchtab_get_text / pdf  → 提取或导出结果
```

### 示例：填写搜索表单

```
代理：在维基百科上搜索“气候变化”

工具调用：
  pinchtab_navigate({url: "https://www.wikipedia.org"})
  pinchtab_snapshot({interactive: true})
    → ...input[ref=e3] placeholder="Search Wikipedia"...
  pinchtab_click({selector: "e3"})
  pinchtab_type({selector: "e3", text: "climate change"})
  pinchtab_press({key: "Enter"})
  pinchtab_snapshot({format: "compact"})
  pinchtab_get_text({})
```

## 启用 JavaScript 评估

默认情况下，`pinchtab_eval` 作为安全措施被禁用。要启用它：

```bash
pinchtab config set security.allowEvaluate true
```

或在 `~/.pinchtab/config.json` 中：

```json
{
  "security": {
    "allowEvaluate": true
  }
}
```

更改此设置后重新启动 PinchTab。

> **警告：** 启用评估是一个有文档记录的、非默认的、降低安全性的配置更改。它允许代理（以及它访问的任何页面）在浏览器中运行任意 JavaScript。仅在设置了令牌的受信任网络上启用它。

## 连接到远程 PinchTab

如果 PinchTab 在另一台机器上运行（例如 Docker 容器），请使用 `--server` 标志：

```json
{
  "mcpServers": {
    "pinchtab": {
      "command": "pinchtab",
      "args": ["--server", "http://192.168.1.50:9867", "mcp"],
      "env": {
        "PINCHTAB_TOKEN": "your-secure-token"
      }
    }
  }
}
```

`pinchtab mcp` 进程在本地运行（在代理机器上），并向远程 PinchTab 实例发出 HTTP 调用。Chrome 在远程机器上 —— 只有 stdio MCP 传输是本地的。

## 故障排除

**所有工具都显示“连接被拒绝”**

PinchTab 未运行，或在不同端口上。检查：

```bash
curl http://127.0.0.1:9867/health
```

**工具显示“HTTP 401”**

令牌不匹配。将 `PINCHTAB_TOKEN` 设置为与 PinchTab 配置中的 `server.token` 匹配。

**`pinchtab_eval` 显示“HTTP 403”**

JavaScript 评估被禁用。请参见上面的 [启用 JavaScript 评估](#enabling-javascript-evaluation)。

**“ref not found”错误**

元素引用在每次导航或重大 DOM 更新后都会更改。在使用之前快照中的引用之前，始终在页面更改后再次调用 `pinchtab_snapshot`。

**MCP 服务器未在客户端中显示**

- 检查 `command` 值 —— `pinchtab` 必须在 PATH 中，或使用绝对路径。
- 在终端中手动运行 `pinchtab mcp` 以检查启动错误。
- 检查 MCP 进程的 stderr 输出（客户端特定，通常在日志文件中）。

## 相关页面

- [MCP 概述](../mcp.md)
- [MCP 工具参考](../reference/mcp-tools.md)
- [MCP 架构](../architecture/mcp.md)
- [安全指南](./security.md)