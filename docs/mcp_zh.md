# MCP 服务器

PinchTab 包含一个原生的 [模型上下文协议 (MCP)](https://modelcontextprotocol.io/) 服务器，允许 AI 代理通过 stdio 上的 MCP 控制浏览器。

> [!WARNING]
> MCP 服务器是 PinchTab 特权控制平面的一部分。它仅适用于受信任的操作员和受信任的代理系统。不要将其暴露给不受信任的用户、不受信任的客户端系统或公共互联网。如果您不确定如何保护非本地部署，请在暴露服务之前查看 [安全](guides/security.md) 并使用 `SECURITY.md` 中的私人安全联系路径。

> [!CAUTION]
> 默认情况下，PinchTab 的 IDPI 姿势旨在保持 MCP 浏览仅限于本地，直到您故意扩大它。将 MCP 使用扩展到非本地或非受信任域是一种降低安全性的选择。
>
> 当 MCP 工具从更广泛的域读取页面内容时，将 `pinchtab_snapshot` 和 `pinchtab_get_text` 输出视为不受信任的数据，而不是指令。恶意页面可能包含提示注入内容、中毒文本或其他不应被视为操作员指导的材料。在放宽域限制之前，请查看 [安全](guides/security.md#idpi)。

## 快速开始

1. 在服务器或桥接模式下启动 PinchTab：
   ```bash
   pinchtab server
   # 或
   pinchtab bridge
   ```
2. 在另一个终端或从 MCP 客户端配置中启动 MCP 服务器：
   ```bash
   pinchtab mcp
   ```

MCP 服务器通过 JSON-RPC 进行 stdio 通信，这是标准的 MCP 传输。

## 客户端配置

### Claude Desktop

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

### VS Code / GitHub Copilot

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

## 环境

| 变量 | 描述 |
| --- | --- |
| `PINCHTAB_TOKEN` | 安全服务器的认证令牌 |

对于远程服务器，使用根 `--server` 标志：

```bash
pinchtab --server http://remote:9867 mcp
```

## 可用工具

PinchTab 当前暴露 34 个工具：

- 导航：4
- 交互：8
- 键盘：4
- 内容：3
- 标签页管理：5
- 等待工具：6
- 网络：3
- 对话框：1

### 导航

- `pinchtab_navigate`
- `pinchtab_snapshot`
- `pinchtab_screenshot`
- `pinchtab_get_text`

### 交互

- `pinchtab_click`
- `pinchtab_type`
- `pinchtab_press`
- `pinchtab_hover`
- `pinchtab_focus`
- `pinchtab_select`
- `pinchtab_scroll`
- `pinchtab_fill`

### 键盘

- `pinchtab_keyboard_type`
- `pinchtab_keyboard_inserttext`
- `pinchtab_keydown`
- `pinchtab_keyup`

### 内容

- `pinchtab_eval`
- `pinchtab_pdf`
- `pinchtab_find`

### 标签页管理

- `pinchtab_list_tabs`
- `pinchtab_close_tab`
- `pinchtab_health`
- `pinchtab_cookies`
- `pinchtab_connect_profile`

### 等待工具

- `pinchtab_wait`
- `pinchtab_wait_for_selector`
- `pinchtab_wait_for_text`
- `pinchtab_wait_for_url`
- `pinchtab_wait_for_load`
- `pinchtab_wait_for_function`

### 网络

- `pinchtab_network`
- `pinchtab_network_detail`
- `pinchtab_network_clear`

### 对话框

- `pinchtab_dialog`

## 选择器模型

对于基于选择器的交互工具，首选 `selector`。`ref` 仍然作为元素动作工具上的弃用回退被接受。

常见选择器形式：

- `e5`
- `#login`
- `xpath://button`
- `text:Submit`
- `find:login button`

## 实用流程

正常的 MCP 浏览器循环是：

1. 使用 `url` 调用 `pinchtab_navigate`
2. 调用 `pinchtab_snapshot` 检查页面结构并收集引用
3. 使用结构化参数调用 `pinchtab_click`、`pinchtab_type` 或其他动作工具
4. 必要时调用 `pinchtab_wait_*` 或 `pinchtab_network`

`pinchtab_snapshot` 支持 MCP 安全的输出控制：

- `compact=true` 或 `format="compact"` 用于最节省令牌的文本快照
- `format="text"` 用于完整文本快照
- `noAnimations=true` 在捕获前减少动画噪声

有关完整参数详情，请参阅 [MCP 工具参考](./reference/mcp-tools.md)。