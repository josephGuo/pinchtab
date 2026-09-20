# 将 PinchTab 与 AI 代理一起使用（MCP）

本指南带你把 PinchTab 搭建为 AI 编码助手与代理框架的 MCP 工具服务器。

> [!WARNING]
> 当你把一个 MCP 客户端接到 PinchTab 时，该客户端行使的是与仪表板、API 和远程 CLI 相同的特权控制平面。只应允许受信任的操作员和受信任的代理系统使用它。如果你不确定一个非本地或部分暴露的部署是否安全，请停下并在继续之前阅读 [安全](security.md)。

> [!CAUTION]
> 把 MCP 浏览范围放宽到本地或明确受信任域之外，是一个降低安全性的选择。如果你放宽 IDPI 允许列表或严格模式，`pinchtab_snapshot` 和 `pinchtab_get_text` 的输出可能包含来自不受信任页面的恶意指令。
>
> 把所有面向模型的页面内容都视为不可信数据。除非受信任的操作员单独验证过，否则不要遵循嵌入在页面文本、可访问性标签、隐藏内容或提取摘要中的指令。

## 什么是 MCP？

[Model Context Protocol](https://modelcontextprotocol.io/) 是一个把 AI 模型连接到外部工具的开放标准。PinchTab 实现了一个 MCP 服务器，在一个所有主流 AI 客户端都支持的简单 stdio 接口上暴露 47 个浏览器控制工具——导航、交互、截图、PDF 导出、等待、网络检查等等。

## 前置条件

- 已安装 PinchTab（`pinchtab --version`）
- 已安装 Chrome 并在 PATH 中（或通过配置指向它）
- 一个 MCP 兼容客户端：Claude Desktop、带 GitHub Copilot 的 VS Code、Cursor，或 Grok Build

## 第 1 步：启动 PinchTab

MCP 服务器是一个薄适配器——它需要一个正在运行的 PinchTab 实例来代理调用。

**单实例 bridge（默认无头，推荐给代理）：**

```bash
pinchtab bridge
```

**普通服务器模式（如果你也想要仪表板）：**

```bash
pinchtab server
```

PinchTab 默认监听 `http://127.0.0.1:9867`。

## 第 2 步：配置你的 MCP 客户端

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

重启 Claude Desktop。你应该能在工具列表中看到 PinchTab。

### VS Code / GitHub Copilot

在你的工作区根目录创建 `.vscode/mcp.json`：

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

加入你的 Cursor MCP 设置（`~/.cursor/mcp.json`）：

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

### Grok Build

在 PinchTab 被列入 xAI 官方 marketplace 之后，用以下命令安装插件：

```bash
grok plugin install pinchtab --trust
```

在此之前，或要从 PinchTab 仓库 marketplace 安装：

```bash
grok plugin marketplace add pinchtab/pinchtab
grok plugin install pinchtab --trust
```

你也可以直接从 GitHub 安装该插件目录：

```bash
grok plugin install pinchtab/pinchtab#plugins/grok --trust
```

从本地检出的根目录，使用 `grok plugin install ./plugins/grok --trust`。该插件不安装二进制。请单独安装 PinchTab、启动 `pinchtab server`，然后信任该插件，MCP 工具就会出现在 `/mcps`。首次使用、验证、域名授权与故障排查见 [Grok 插件安装与使用指南](../../plugins/grok/README.md)。

要不带插件配置 MCP，在 `~/.grok/config.toml` 中添加：

```toml
[mcp_servers.pinchtab]
command = "pinchtab"
args = ["mcp"]
```

### 任何基于 SDK 的代理

```python
# Python example using the mcp SDK
import subprocess, mcp

proc = subprocess.Popen(
    ["pinchtab", "mcp"],
    stdin=subprocess.PIPE,
    stdout=subprocess.PIPE,
)
# pass proc.stdin / proc.stdout to your MCP client transport
```

## 环境变量

| 变量 | 默认值 | 描述 |
|----------|---------|-------------|
| `PINCHTAB_TOKEN` | *（来自配置）* | 受认证保护服务器的 Bearer 令牌 |

对于远程服务器，用 `--server` flag 并带上该主机的凭证——非环回主机要求 `PINCHTAB_TOKEN`（或 `PINCHTAB_SESSION`），因为 CLI 拒绝把本地配置的 `server.token` 发到本机之外：

```bash
PINCHTAB_TOKEN=<that-host-token> pinchtab --server http://remote:9867 mcp
```

`PINCHTAB_TOKEN` 来自你 PinchTab 配置文件里的 `server.token`——那是你**本地**服务器的凭证，不是远程服务器的。要复制当前令牌而不把它打到 stdout，运行 `pinchtab config token`。在没有剪贴板工具的主机上——CI、Docker、无头 Linux——用 `PINCHTAB_TOKEN=$(pinchtab config token --stdout)`，它只打印令牌、别无其他。

## 典型代理工作流

在让一个接了 MCP 的代理浏览本地或受信任域之外之前，请阅读 [安全](security.md#idpi)。最安全的姿态是把 IDPI 域名限制收窄，并假设每个提取出的页面字符串最多只是参考性的、且可能是恶意的。

一个写得好的代理提示会按以下顺序使用工具：

```
1. pinchtab_navigate        → go to the target URL
2. pinchtab_snapshot        → understand the page structure (find refs)
3. pinchtab_click / type    → interact with elements by structured tool arguments
4. pinchtab_snapshot        → confirm state after interaction
5. pinchtab_get_text / pdf  → extract or export results
```

### 示例：填写搜索表单

```
Agent: Search for "climate change" on Wikipedia

Tool calls:
  pinchtab_navigate({url: "https://www.wikipedia.org"})
  pinchtab_snapshot({interactive: true})
    → ...input[ref=e3] placeholder="Search Wikipedia"...
  pinchtab_click({selector: "e3"})
  pinchtab_type({selector: "e3", text: "climate change"})
  pinchtab_key({action: "press", key: "Enter"})
  pinchtab_snapshot({format: "compact"})
  pinchtab_get_text({})
```

## 启用 JavaScript 求值

作为一项安全措施，`pinchtab_eval` 默认禁用。要启用它：

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

更改此设置后重启 PinchTab。

> **警告：** 启用 evaluate 是一个有文档记录、非默认、降低安全性的配置变更。它允许代理（以及它访问的任何页面）在浏览器中运行任意 JavaScript。只在设置了令牌的受信任网络上启用它。

## 连接到远程 PinchTab

如果 PinchTab 运行在另一台机器上（例如一个 Docker 容器），用 `--server` flag：

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

`pinchtab mcp` 进程在本地（代理机器上）运行，对远程 PinchTab 实例发起 HTTP 调用。Chrome 在远程机器上——只有 stdio MCP 传输是本地的。

## 故障排查

**所有工具都报 "Connection refused"**

PinchTab 没在运行，或在不同端口上。检查：

```bash
pinchtab health
```

（裸 `curl http://127.0.0.1:9867/health` 也能证明端口有应答，但不带 `-H "Authorization: Bearer <token>"` 时它会返回 `401`。）

**工具报 "HTTP 401"**

令牌不匹配。把 `PINCHTAB_TOKEN` 设成与你 PinchTab 配置里的 `server.token` 一致。

**`pinchtab_eval` 报 "HTTP 403"**

JavaScript 求值被禁用。见上文[启用 JavaScript 求值](#启用-javascript-求值)。

**"ref not found" 错误**

元素 ref 在每次导航或显著 DOM 更新后都会变化。页面变化后、再用上一次快照里的 ref 之前，务必重新调用 `pinchtab_snapshot`。

**MCP 服务器没出现在客户端里**

- 检查 `command` 值——`pinchtab` 必须在 PATH 上，否则用绝对路径。
- 在终端里手动运行 `pinchtab mcp`，检查启动错误。
- 查看 MCP 进程的 stderr 输出（因客户端而异，通常在某个日志文件里）。

## 相关页面

- [MCP 概述](../mcp.md)
- [MCP 工具参考](../reference/mcp-tools.md)
- [MCP 架构](../architecture/mcp.md)
- [安全指南](./security.md)
