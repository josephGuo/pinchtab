# 命令行界面 概览

`pinchtab` 有两种常规使用方式：

- 交互式菜单模式
- 直接命令模式

当你需要一个引导式的本地控制界面时使用菜单模式。当你需要使用 shell 历史记录、脚本或通过 `--server` 进行远程目标操作时使用直接命令模式。

当你使用 `--server` 目标指向远程服务器时，命令行界面 正在使用与仪表板和 HTTP API 相同的特权控制平面。不要将其用作不受信任用户或不受信任系统的访问路径。有关部署指南，请参阅 [安全](../guides/security.md)。

## 交互式菜单

在交互式终端中运行 `pinchtab` 不带子命令会打开菜单。它不会立即启动服务器。

典型流程：

```text
listen    running  127.0.0.1:9867
str,plc   simple,fcfs
daemon    ok
security  [■■■■■■■■■■]  LOCKED

Main Menu
  1. Start server
  2. Daemon
  3. Start bridge
  4. Start MCP server
  5. Config
  6. Security
  7. Help
  8. Exit
```

## 直接命令

当你已经知道想要执行的操作时使用直接命令：

```bash
pinchtab server
pinchtab bridge
pinchtab mcp
pinchtab config
pinchtab --agent-id agent-main nav https://pinchtab.com
pinchtab nav https://pinchtab.com
pinchtab snap -i -c
pinchtab click e5
pinchtab find "login button"
pinchtab network --limit 20
```

全局标志如 `--server` 和 `--agent-id` 适用于直接命令模式。`--agent-id` 会记录在活动日志和仪表板代理视图中，以便区分多个由 命令行界面 驱动的代理。

## 代理归因

命令行界面 请求通过 `X-Agent-Id` 请求头携带代理标识。

- `--agent-id <value>` 为该命令显式设置头部
- `PINCHTAB_AGENT_ID` 为当前 shell 或脚本设置默认代理 ID
- 如果两者都未设置，命令行界面 使用 `命令行界面`

该代理 ID 以 `agentId` 的形式出现在 `/api/activity`、代理页面和调度器驱动的活动中。

示例：

```bash
PINCHTAB_AGENT_ID=agent-crawl-01 pinchtab nav https://pinchtab.com
curl 'http://127.0.0.1:9867/api/activity?agentId=agent-crawl-01'
```

## 输出格式

大多数命令默认输出人类可读的文本。使用 `--json` 获得结构化输出：

```bash
pinchtab tab                  # *abc123  https://...  Title
pinchtab tab --json           # {"tabs":[...]}
pinchtab frame                # main
pinchtab network              # GET  200  https://...
```

**对于脚本**：在通过管道传输或编程解析时始终使用 `--json`。人类可读的输出可能在版本之间发生变化。JSON 是稳定的契约。

## 核心命令

| 命令 | 用途 |
| --- | --- |
| `pinchtab server` | 启动完整服务器和仪表板 |
| `pinchtab bridge` | 启动单实例桥接运行时 |
| `pinchtab mcp` | 启动标准输入/输出 MCP 服务器 |
| `pinchtab daemon` | 显示守护进程状态并管理后台服务 |
| `pinchtab config` | 打开交互式配置概览/编辑器 |
| `pinchtab security` | 打开交互式安全概览 |
| `pinchtab completion <shell>` | 生成 shell 补全脚本 |

## 浏览器命令

浏览器控制界面是顶级的。`tab` 仅用于列表/聚焦/新建/关闭。

常用命令：

| 命令 | 用途 |
| --- | --- |
| `pinchtab nav <url>` | 打开新标签页并导航 |
| `pinchtab quick <url>` | 导航并快照 |
| `pinchtab snap` | 可访问性快照 |
| `pinchtab frame [target\|main]` | 显示或设置选择器框架范围 |
| `pinchtab click <selector>` | 点击元素 |
| `pinchtab mouse move <x> <y>` | 将指针移动到坐标 |
| `pinchtab mouse down [selector]` | 在当前指针或新目标处按下鼠标按钮 |
| `pinchtab mouse up [selector]` | 在当前指针或新目标处释放鼠标按钮 |
| `pinchtab mouse wheel [dy\|selector]` | 在当前指针或新目标处调度滚轮增量 |
| `pinchtab drag <from> <to>` | 从一个目标拖动到另一个目标 |
| `pinchtab type <selector> <text>` | 通过按键事件输入 |
| `pinchtab fill <selector> <text>` | 直接填充 |
| `pinchtab text` | 提取页面文本 (`--full`, `--raw`, `--frame <frameId>`) |
| `pinchtab find <query>` | 语义元素搜索 |
| `pinchtab screenshot` | 保存屏幕截图 (`-s/--selector` 捕获特定元素, `--css-1x` 以 CSS 大小导出选择器截图) |
| `pinchtab pdf` | 将页面导出为 PDF |
| `pinchtab network` | 检查捕获的网络请求 |
| `pinchtab wait ...` | 等待选择器、文本、URL、JS 或时间 |
| `pinchtab console` | 显示浏览器控制台日志 |
| `pinchtab errors` | 显示浏览器错误日志 |

许多浏览器命令接受 `--tab <id>` 来针对现有标签页而不是活动标签页。

选择器查找按框架显式进行。未限定范围的选择器会保持在主文档中，除非你先使用 `pinchtab frame` 设置框架。支持同源 iframe 范围；目前不公开跨域 iframe 后代。

`pinchtab text` 也遵循该框架模型：它使用活动框架范围，除非你使用 `--frame` 覆盖它。

`pinchtab eval` 与该模型分离，不继承当前框架范围。

基于选择器的操作在选择器不匹配时会快速失败。如果你期望动态内容很快出现，请先使用 `pinchtab wait`。

手动切换可通过 `tab` 命令使用：

```bash
pinchtab tab handoff <tabId> --reason captcha --timeout-ms 120000
pinchtab tab handoff-status <tabId>
pinchtab tab resume <tabId> --status completed
```

API 等价物：

暂停的切换状态会阻止操作执行路由 (`/action`, `/actions`, `/macro`) 并返回 `409 tab_paused_handoff`，直到通过超时恢复或过期。

```bash
curl -X POST http://localhost:9867/tabs/<tabId>/handoff \
  -H "Content-Type: application/json" \
  -d '{"reason":"captcha"}'
curl http://localhost:9867/tabs/<tabId>/handoff
curl -X POST http://localhost:9867/tabs/<tabId>/resume \
  -H "Content-Type: application/json" \
  -d '{"status":"completed"}'
```

## 标签页命令

`pinchtab tab` 故意设计得很小：

```bash
pinchtab tab
pinchtab tab <id>
pinchtab tab new [url]
pinchtab tab close <id>
pinchtab tab handoff <id>
pinchtab tab handoff-status <id>
pinchtab tab resume <id>
```

对于标签页范围的操作，使用带有 `--tab` 的普通顶级命令：

```bash
pinchtab click --tab <id> e5
pinchtab pdf --tab <id> -o page.pdf
```

## 从 命令行界面 配置

`pinchtab config` 显示：

- `multiInstance.strategy`
- `multiInstance.allocationPolicy`
- `instanceDefaults.stealthLevel`
- `instanceDefaults.tabEvictionPolicy`
- 活动配置文件路径
- 服务器运行时的仪表板 URL
- 掩码服务器令牌
- `Copy token` 操作

有关文件架构详情和 `config get/set/patch`，请参阅 [配置](./config.md)。

## 从 命令行界面 安全

`pinchtab security` 是交互式安全屏幕。

直接子命令：

```bash
pinchtab security up
pinchtab security down
```

`pinchtab security down` 为本地操作员工作流应用文档中记录的、非默认的、降低安全性的预设。它不是基线安全状态。

有关更广泛的安全指南，请参阅 [安全指南](../guides/security.md)。

## 守护进程

`pinchtab daemon` 支持：

- macOS 通过 `launchd`
- Linux 通过用户 `systemd`

Windows 二进制文件存在，但目前不支持守护进程工作流。直接使用 `pinchtab server` 或 `pinchtab bridge`。

有关操作详情，请参阅 [后台服务（守护进程）](../guides/daemon.md)。

## 完整命令树

使用内置帮助获取实时命令树：

```bash
pinchtab --help
```

有关每个命令的页面，请从 [参考索引](./index.md) 开始。