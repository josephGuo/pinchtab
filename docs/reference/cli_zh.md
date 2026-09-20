# 命令行界面概览

`pinchtab` 由直接命令驱动。不带任何子命令运行它会打印一份状态摘要和下一步提示。

当你用 `--server` 指向远程服务器时，命令行界面正在使用与仪表板和 HTTP API 相同的特权控制平面。不要将其用作不受信任用户或不受信任系统的访问路径。部署指南见 [安全](../guides/security.md)。

## 裸 `pinchtab`

不带子命令运行 `pinchtab` 不会启动服务器。当配置文件是新建的，或其 `configVersion` 比当前构建旧时，它会先运行安全设置（若缺失则生成 `server.token`；提示仅出现在交互式终端中）。然后它打印服务器状态、日志去向、`allowedDomains`、IDPI 状态以及建议的下一步命令，例如：

```text
PinchTab dev

  server               protected listener
  logs                 stdout/stderr of the terminal running `pinchtab server`
  allowedDomains       a.com, b.com
  idpi                 enabled

Next steps:
  pinchtab config token --stdout               # print the API token (capture with $(...))
  pinchtab health --json                       # retry health with the current token
  pinchtab config show                         # inspect configured port and token
```

关于你可能已选择的稳态的提示性建议每次运行只打印一次；设置 `PINCHTAB_HINTS=off` 可将其静音。

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

`pinchtab nav <url>` 在本地 PinchTab 服务器尚未运行时会自动启动它。显式的 `--server` 和 `PINCHTAB_SERVER` 目标按原样使用，不会自动启动。默认配置使用无头浏览器，因此 `nav` 可能在不打开可见窗口的情况下成功。安装后要在一条命令中同时导航并快照，运行：

```bash
pinchtab nav https://pinchtab.com --snap
```

全局 flags 如 `--server` 和 `--agent-id` 适用于直接命令模式。`--agent-id` 会记录在活动日志和仪表板的代理视图中，以便区分多个由命令行界面驱动的代理。

## 代理归因

命令行界面请求通过 `X-Agent-Id` 请求头携带代理身份。

- `--agent-id <value>` 为该命令显式设置该头
- `PINCHTAB_AGENT_ID` 为当前 shell 或脚本设置默认代理 ID
- 若两者都未设置，命令行界面不发送 `X-Agent-Id`；用代理会话（`PINCHTAB_SESSION`）认证的请求会归到该会话的代理 ID

该代理 ID 就是 `/api/activity`、Agents 页面和调度器驱动活动中显示为 `agentId` 的那个值。

示例：

```bash
PINCHTAB_AGENT_ID=agent-crawl-01 pinchtab nav https://pinchtab.com
curl 'http://127.0.0.1:9867/api/activity?agentId=agent-crawl-01'
```

## 输出格式

大多数命令默认输出人类可读文本。使用 `--json` 获得结构化输出：

```bash
pinchtab tab                  # *abc123  https://...  Title
pinchtab tab --json           # {"tabs":[...]}
pinchtab frame                # main
pinchtab network              # GET  200  https://...
```

**对于脚本**：在通过管道传输或编程解析时始终使用 `--json`。人类可读输出可能在版本之间变化。JSON 是稳定的契约。

## 退出码

动词拼写错误在任何地方都是错误：无法识别的命令或子命令以 `1` 退出，并列出有效的子命令，在顶层和每个分组内部均如此。`pinchtab cache clera` 与 `pinchtab cache clear` 不共享退出码，因此 `set -e` 和 `&&` 链会在拼写错误处停止，而不是当作状态已重置继续执行。

```bash
$ pinchtab cache clera ; echo exit=$?
unknown command "clera" for "pinchtab cache"
Valid subcommands: clear, status
exit=1
```

有两条命令接受参数而非子命令，不受影响：`pinchtab tab <id>` 聚焦某个标签页，`pinchtab network <requestId>` 检查一个已捕获的请求（过滤器使用 `--filter`、`--method`、`--status`、`--type`）。那里的未知值是服务器拒绝的数据，而非命令行界面能捕获的拼写错误。

## 核心命令

| 命令 | 用途 |
| --- | --- |
| `pinchtab server` | 启动完整服务器和仪表板 |
| `pinchtab server stop` | 停止运行中的服务器（前台或后台） |
| `pinchtab server restart` | 后台停止 + 重启（应用配置变更） |
| `pinchtab bridge` | 启动单实例 bridge 运行时 |
| `pinchtab mcp` | 启动 stdio MCP 服务器 |
| `pinchtab daemon` | 显示守护进程状态并管理后台服务 |
| `pinchtab config` | 打印配置概览（只读） |
| `pinchtab security` | 打印运行时安全态势 |
| `pinchtab completion <shell>` | 生成 shell 补全脚本 |
| `pinchtab version` | 打印 PinchTab 版本（同 `--version`） |

### Server Flags

```bash
pinchtab server [flags]
```

| Flag | 简写 | 用途 |
| --- | --- | --- |
| `--yolo` | `-y` | 仅本次运行应用"放下守卫"预设；配置文件不变（除状态导出外的每个能力门控打开、attach 打开、IDPI 关闭） |
| `--headed` | `-H` | 以有头（可见）模式启动默认实例 |
| `--extension <path>` | `-e` | 加载浏览器扩展（可重复） |
| `--browser <name>` | | `chrome`、`cloak` 或 `ghost-chrome`（覆盖配置） |
| `--bind <addr>` | | HTTP 绑定地址（覆盖 `server.bind`） |
| `--port <port>` | | HTTP 端口（覆盖 `server.port`） |
| `--log-level <level>` | | `debug`、`info`（默认）、`warn` 或 `error` |
| `--verbose` | `-v` | 完整启动横幅；仅当 `--log-level` 和 `server.logLevel` 都未设置时日志才为 debug |
| `--background` | `-b` | 分离式启动服务器并打印含 pid/url/token 的 JSON |

示例：

```bash
pinchtab server -y                  # guards down for local dev
pinchtab server -H                  # visible browser for debugging
pinchtab server -yH                 # both combined
pinchtab server -e ./my-extension   # load extension
```

**注意：** 仅在需要视觉反馈（调试、手工测试）时使用 `--headed`。无头模式对自动化更节省资源。

## 浏览器命令

浏览器控制界面是顶级的。`tab` 仅用于列表/聚焦/关闭。

常用命令：

| 命令 | 用途 |
| --- | --- |
| `pinchtab nav <url>` | 导航当前被跟踪的标签页，必要时新建一个 |
| `pinchtab nav <url> --timeout 90` | 覆盖导航超时时间，秒为单位（最大 120） |
| `pinchtab nav <url> --snap` | 导航并输出交互式紧凑快照 |
| `pinchtab snap [selector]` | 当前标签页的无障碍快照，可限定范围 |
| `pinchtab frame [target\|main]` | 显示或设置选择器框架范围 |
| `pinchtab click <selector>` | 点击元素 |
| `pinchtab mouse move <x> <y>` | 将指针移动到坐标 |
| `pinchtab mouse down [selector]` | 在当前指针或新目标处按下鼠标按钮 |
| `pinchtab mouse up [selector]` | 在当前指针或新目标处释放鼠标按钮 |
| `pinchtab mouse wheel [dy\|selector]` | 在当前指针或新目标处派发滚轮增量 |
| `pinchtab drag <from> <to>` | 从一个目标拖动到另一个目标 |
| `pinchtab type <selector> <text>` | 通过按键事件输入 |
| `pinchtab fill <selector> <text>` | 直接填充 |
| `pinchtab text` | 提取页面文本（`--full`、`--raw`、`--frame <frameId>`） |
| `pinchtab find <query>` | 语义元素搜索 |
| `pinchtab extract --schema <file\|->` | 从页面按 schema 类型化的 JSON（`--scope`、`--max-items`、`--fields`、`--explain`） |
| `pinchtab screenshot` | 保存截图（`-s/--selector` 捕获特定元素，`--scale <f>` 重新缩放位图，`--beyond-viewport` 捕获完整可滚动文档） |
| `pinchtab capture` | 同一 DOM 纪元的成对截图 + 无障碍快照（`--scale`、`--beyond-viewport`、`--require-pair`、`--with-bounds`） |
| `pinchtab pdf` | 将页面导出为 PDF |
| `pinchtab network` | 检查捕获的网络请求 |
| `pinchtab wait ...` | 等待选择器、文本、URL、网络空闲、JS 或时间 |
| `pinchtab console` | 显示浏览器控制台日志 |
| `pinchtab errors` | 显示浏览器错误日志 |

许多浏览器命令接受 `--tab <id>` 来针对现有标签页而非活动标签页。

选择器查找按框架显式进行。未限定范围的选择器保持在主文档中，除非你先用 `pinchtab frame` 设置框架。支持同源 iframe 范围；目前不公开跨域 iframe 后代。

`pinchtab text` 也遵循该框架模型：它使用活动框架范围，除非你用 `--frame` 覆盖。

`pinchtab eval` 与该模型分离，不继承当前框架范围。

基于选择器的操作在选择器不匹配时快速失败。如果你预期动态内容很快出现，请先使用 `pinchtab wait`。

手动交接可通过 `tab` 命令使用：

```bash
pinchtab tab handoff <tabId> --reason captcha --timeout-ms 120000
pinchtab tab handoff-status <tabId>
pinchtab tab resume <tabId> --status completed
```

API 等价物：

暂停的交接状态会以 `409 tab_paused_handoff` 阻止操作执行路由（`/action`、`/actions`、`/macro`），直到通过恢复或超时过期。

```bash
curl -X POST http://localhost:9867/tabs/<tabId>/handoff \
  -H "Content-Type: application/json" \
  -d '{"reason":"captcha"}'
curl http://localhost:9867/tabs/<tabId>/handoff
curl -X POST http://localhost:9867/tabs/<tabId>/resume \
  -H "Content-Type: application/json" \
  -d '{"status":"completed"}'
```

## Tab 命令

`pinchtab tab` 刻意保持精简：

```bash
pinchtab tab
pinchtab tab <id>
pinchtab tab close <id>
pinchtab tab handoff <id>
pinchtab tab handoff-status <id>
pinchtab tab resume <id>
```

对于标签页范围的操作，使用带 `--tab` 的普通顶级命令：

```bash
pinchtab click --tab <id> e5
pinchtab pdf --tab <id> -o page.pdf
```

## 从命令行界面配置

`pinchtab config` 打印只读概览：

- `multiInstance.strategy`
- `multiInstance.allocationPolicy`
- `instanceDefaults.stealthLevel`
- `instanceDefaults.tabEvictionPolicy`
- `instanceDefaults.tabPolicy.lifecycle`
- 活动配置文件路径
- 掩码后的服务器 token
- 服务器运行时的仪表板 URL
- `config get/set/show`、`config token` 和 `pinchtab security` 的提示

文件 schema 详情和 `config get/set/patch` 见 [Config](./config.md)。

## 从命令行界面安全

`pinchtab security` 打印运行时安全态势和推荐默认值。

直接子命令：

```bash
pinchtab security up
pinchtab security down
```

`pinchtab security down` 为本地操作员工作流应用文档中记录的、非默认的、降低安全性的预设。它不是基线安全态势。

更广泛的安全指南见 [Security Guide](../guides/security.md)。

## 守护进程

`pinchtab daemon` 支持：

- macOS 通过 `launchd`
- Linux 通过用户 `systemd`

Windows 二进制文件存在，但目前不支持守护进程工作流。直接使用 `pinchtab server` 或 `pinchtab bridge`。

操作详情见 [后台服务（Daemon）](../guides/daemon.md)。

## 完整命令树

用内置帮助查看实时命令树：

```bash
pinchtab --help
```

每个命令的页面从 [Reference Index](./index.md) 开始。
