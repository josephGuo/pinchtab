# MCP 服务器

PinchTab 包含一个原生的 [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) 服务器，让 AI 代理通过 stdio 上的 MCP 控制浏览器。

> [!WARNING]
> MCP 服务器是 PinchTab 特权控制平面的一部分。它仅面向受信任的操作员和受信任的代理系统。不要
> 把它暴露给不受信任的用户、不受信任的客户端系统或公共互联网。如果你不确定如何保护一个非本地
> 部署，请在暴露服务之前查看 [Security](guides/security.md)，并使用 `SECURITY.md` 中的私人
> 安全联系路径。

> [!CAUTION]
> 默认情况下，PinchTab 的 IDPI 姿态旨在让 MCP 浏览保持仅限本地，直到你刻意把它放宽。把 MCP 使用
> 扩展到非本地或非受信任域，是一个降低安全性的选择。
>
> 当 MCP 工具从更广泛的域读取页面内容时，把 `pinchtab_snapshot` 和 `pinchtab_get_text` 的输出
> 当作不受信任的数据，而非指令。恶意页面可能包含提示注入内容、被投毒的文本，或其他绝不应被当作
> 操作员指导的材料。放宽域限制之前，请查看 [Security](guides/security.md#idpi)。

## 快速开始

1. 以服务器或桥接模式启动 PinchTab：
   ```bash
   pinchtab server
   # or
   pinchtab bridge
   ```
2. 在另一个终端中、或从你的 MCP 客户端配置中启动 MCP 服务器：
   ```bash
   pinchtab mcp
   ```

MCP 服务器通过 stdio 使用 JSON-RPC 通信，这是标准的 MCP 传输。

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

### Grok Build

在 PinchTab 被列入官方 xAI marketplace 之后，用以下命令安装插件：

```bash
grok plugin install pinchtab --trust
```

在此之前，或要从 PinchTab 仓库 marketplace 安装：

```bash
grok plugin marketplace add pinchtab/pinchtab
grok plugin install pinchtab --trust
```

你也可以直接从 GitHub 安装插件目录：

```bash
grok plugin install pinchtab/pinchtab#plugins/grok --trust
```

从本地检出的根目录，使用 `grok plugin install ./plugins/grok --trust`。MCP 启动前必须带
`--trust`。该插件不会安装 `pinchtab` 二进制；请单独安装 PinchTab 并运行
`pinchtab server` 或本地守护进程。

要自己接线 MCP 而非用插件，请往 `~/.grok/config.toml` 添加：

```toml
[mcp_servers.pinchtab]
command = "pinchtab"
args = ["mcp"]
```

首次使用、验证、域授权与排错，参见
[Grok 插件安装与使用指南](../plugins/grok/README.md)。

## 环境

| 变量 | 描述 |
| --- | --- |
| `PINCHTAB_TOKEN` | 受保护服务器的认证 token |

对远程服务器，使用根级 `--server` flag 并带上该主机的凭据——一个非回环主机需要
`PINCHTAB_TOKEN`（或 `PINCHTAB_SESSION`），因为命令行界面拒绝把本地配置的 `server.token`
发到本机之外：

```bash
PINCHTAB_TOKEN=<that-host-token> pinchtab --server http://remote:9867 mcp
```

## 可用工具

PinchTab 当前暴露 47 个工具：

- Navigation：9
- Interaction：8
- Keyboard：1
- Content：4
- Recording：1
- Site：1
- Tab management：6
- Human handoff：3
- Wait utilities：1
- Network：6
- Diagnostics：6
- Dialog：1

### Navigation

- `pinchtab_navigate`
- `pinchtab_back`——后退一个历史条目；返回标签页 ID 和落定的 URL
- `pinchtab_forward`——前进一个历史条目；返回标签页 ID 和落定的 URL
- `pinchtab_reload`——重新加载当前页面；返回标签页 ID 和落定的 URL
- `pinchtab_snapshot`
- `pinchtab_frame`
- `pinchtab_screenshot`
- `pinchtab_capture`——来自同一 DOM epoch 的配对截图 + 快照
- `pinchtab_get_text`——提取页面文本；`mode` 选 `readability`（默认）、`raw` 或
  `markdown`（保留链接与表格，最适合文章形态的页面）。`mode` 取代旧的布尔 `raw`。示例：
  `pinchtab_get_text {"mode":"markdown"}`

### Interaction

- `pinchtab_click`
- `pinchtab_type`
- `pinchtab_hover`
- `pinchtab_focus`
- `pinchtab_select`
- `pinchtab_scroll`
- `pinchtab_scroll_into_view`
- `pinchtab_fill`

### Keyboard

- `pinchtab_key`——`action` 为 `press`、`down`、`up`、`type` 或 `insert`

### Content

- `pinchtab_eval`
- `pinchtab_pdf`
- `pinchtab_find`
- `pinchtab_extract`——对照一个 JSON `schema` 的类型化取值（`tabId`、`scope`、`maxItems`
  可选）；每个字段都带一个供动作工具使用的 ref 和一个置信度。对结构化取值，优先它而非「先快照
  再解析」，并用 `x-pinchtab-hint` 固定一个 `low` 字段。参见 [Extract](reference/extract.md)

### Recording

- `pinchtab_record`——`action` 为 `start`、`stop` 或 `status`

### Site

- `pinchtab_scrape`——把一个站点爬取为 markdown（HTTP 优先，浏览器增强 thin/JS 页面）。用
  `preview=true` 拿廉价大纲，再用 `only` 展开选定 URL。它的 `timeoutSeconds` 是一个爬取预算、
  也是唯一的秒级参数；别处每一个等待都是 `timeoutMs`。

### Tab Management

- `pinchtab_list_tabs`
- `pinchtab_close_tab`
- `pinchtab_health`
- `pinchtab_cookies`（需要 `security.allowCookies`）
- `pinchtab_cookies_set`（需要 `security.allowCookies`）
- `pinchtab_connect_profile`

### Human Handoff

- `pinchtab_handoff`——暂停一个标签页（需要 `tabId`；`reason`、`timeoutMs` 可选），好让人工完成
  CAPTCHA、登录或同意步骤；在它恢复之前，该标签页上的每个动作工具都以 `tab_paused_handoff`
  被拒绝
- `pinchtab_resume`——恢复该标签页（需要 `tabId`；`status` 可选）。只在用户确认完成后才调用
- `pinchtab_handoff_status`——报告 `paused_handoff`（带 `reason`、`pausedAt`、`expiresAt`）或
  `active`，于是代理可以观察一次 `timeoutMs` 自动恢复，而无需自己恢复

### Wait Utilities

- `pinchtab_wait`——`for` 为 `ms`、`selector`、`text`、`url`、`load` 或 `function`，
  `value` 携带条件；`timeoutMs` 给浏览器支撑的等待设上限（`timeout` 是其弃用别名）

### Network

- `pinchtab_network`
- `pinchtab_network_detail`
- `pinchtab_network_clear`
- `pinchtab_network_route`
- `pinchtab_network_unroute`
- `pinchtab_network_rules`

### Diagnostics

- `pinchtab_console`——读取（或 `clear`）该标签页的浏览器控制台日志
- `pinchtab_errors`——读取（或 `clear`）该标签页未捕获的 JavaScript 异常；当快照看起来健康但动作
  什么都不做时，检查这里
- `pinchtab_a11y_audit`——无障碍审计；`engine=axe` 在隔离世界中运行内置 axe-core，返回带 WCAG
  标签和帮助 URL 的、ref 映射的违规
- `pinchtab_memory`——该标签页的 JavaScript 堆使用与 DOM 计数器；`gc=true` 先做垃圾回收，于是
  两次读取可比较保留内存
- `pinchtab_memory_snapshot`——把一份 V8 堆快照写到服务端文件，只返回其路径和一份摘要（顶层构造
  器、重复字符串）；需要 `security.allowMemory`。参见 [Memory](reference/memory.md)
- `pinchtab_memory_compare`——比较两个快照 id：逐构造器计数与 self-size 增量，大者在前，外加新的
  重复字符串；`retained=true` 加上头快照支配树的保留大小；需要 `security.allowMemory`

### Dialog

- `pinchtab_dialog`

## 选择器模型

对基于选择器的交互工具，优先用 `selector`。`ref`、`element` 和 `target` 在元素动作工具上仍作为
弃用别名被接受。

## 参数名

每个工具都会拒绝其 schema 未声明的参数。调用返回一个错误，命名未知键、接近时给出最近的已声明名
（`pinchtab_snapshot` 上的 `filter: "interactive"` 指向 `interactive: true`），以及该工具声明的
参数；什么都不运行。弃用写法仍保留在声明中，并在描述中命名规范写法：`selector` 的
`ref`/`element`/`target`、`pinchtab_wait` 上 `timeoutMs` 的 `timeout`、`pinchtab_click` 上
`dialogAction`/`dialogText` 的 `onDialog`/`promptText`、`pinchtab_type` 上 `text` 的
`value`、`pinchtab_select` 上 `value` 的 `option`、`pinchtab_fill` 上 `value` 的 `text`。

常见选择器形式：

- `e5`
- `#login`
- `xpath://button`
- `text:Submit`
- `find:login button`

## 实用流程

正常的 MCP 浏览器循环是：

1. 用一个 `url` 调用 `pinchtab_navigate`
2. 调用 `pinchtab_snapshot` 检查页面结构并收集 refs
3. 用结构化参数调用 `pinchtab_click`、`pinchtab_type` 或其他动作工具
4. 必要时调用 `pinchtab_wait` 或 `pinchtab_network`
5. 调用 `pinchtab_back` 离开死路，或 `pinchtab_reload` 重试该页面

`pinchtab_back`、`pinchtab_forward` 和 `pinchtab_reload` 接受一个可选的 `tabId` 和
`snap`，于是 `snap: true` 在一次往返中返回导航后的页面。

`pinchtab_snapshot` 支持 MCP 安全的输出控制：

- `compact=true` 或 `format="compact"` 用于最省 token 的文本快照
- `format="text"` 用于完整文本快照
- `noAnimations=true` 在捕获前减少动画噪声

完整参数详情，参见 [MCP Tool Reference](./reference/mcp-tools.md)。
