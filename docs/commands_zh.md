# 命令参考

## 服务器与运行时

```bash
pinchtab server                         # Start the full server (dashboard + API)
pinchtab server -b                      # Start it detached; logs go to <stateDir>/server.log
pinchtab server -v                      # Full startup banner, and log at debug level
pinchtab server --log-level warn        # Record warnings and errors only
pinchtab server stop                    # Stop the running server (foreground or background)
pinchtab server restart                 # Stop + restart in background (applies config changes)
pinchtab bridge                         # Start the bridge-only runtime
pinchtab bridge --log-level debug       # Bridge threshold (same precedence as server)
pinchtab mcp                            # Start the MCP stdio server
pinchtab dashboard                      # Open the dashboard in your browser (--no-open prints the URL)
pinchtab session create --agent-id <id> # Create an agent session (--grant limits it to capability groups)
pinchtab session list                   # List agent sessions
pinchtab session info                   # Show the current agent session
pinchtab session revoke <session-id>    # Revoke an agent session
pinchtab daemon                         # Show daemon status
pinchtab daemon install                 # Install as a background service
pinchtab daemon start                   # Start the background service
pinchtab daemon stop                    # Stop the background service
pinchtab daemon restart                 # Restart the background service
pinchtab daemon uninstall               # Remove the background service
pinchtab completion <shell>             # Generate shell completions
```

日志是一个级别（level），而不是一个开/关开关。每一次运行 —— 前台或 `--background`—— 都会记录逐请求的访问日志（含其 `requestId`）、实例生命周期变迁、警告与错误。阈值取自以下第一个被设置的项：`--log-level debug|info|warn|error`，然后是配置文件中的 `server.logLevel`，然后是 `-v`，最后是默认的 `info`。`-v` 总会加上完整的启动横幅，并且只在另外两项都未设置时才把级别提升到 debug。

这些日志行落在哪里，取决于服务器是如何启动的 —— 这也是为什么不带参数运行 `pinchtab` 会打印一行 `logs`，指明当前活跃的日志去向。服务器有四种启动方式，而日志去向有三个：`<stateDir>/server.log`（对应 `pinchtab server -b`，**也对应由裸的 `pinchtab nav` 或 `pinchtab mcp` 自动启动的服务器**—— 两者都以 detached 方式派生，都追加到同一文件）；`~/.pinchtab/logs/daemon.err.log`（对应以 daemon 方式安装的服务）；以及终端（对应前台运行）。早先一次 detached 运行遗留下来的 `server.log` 会被特别标注为「并非由当前服务器写入」，以免被误认为是活跃日志。

一个失败的请求会在那里**不加脱敏**地记录其原因 —— 绝对路径原样保留 —— 并挂在访问日志所记录的同一个 `requestId` 下，于是一个 5xx 可以与其发生原因关联起来：

```bash
grep '"request failed"' <stateDir>/server.log        # causes, with requestId and status
grep <requestId> ~/.pinchtab/activity/*.jsonl        # the access-log line it belongs to
```

跨越 HTTP 边界的消息仍然是路径脱敏过的（`fork/exec [path]`），因此那份未脱敏的副本只存在于服务器自身的日志中。服务器故障（5xx）以 error 级别记录；4xx 属于调用方输入，以 debug 级别记录。

访问日志正是打开着的仪表板所付出的代价：其错误面板和控制台面板都以 3 秒间隔轮询，因此一个一直开着的仪表板每分钟大约写出 40 行。这是为「一次事后能自解释的运行」所做的刻意权衡，而 `--log-level warn` 就是当你想要记录、却不想要轮询时的逃生口。

以 daemon 安装的服务器，与由裸的 `pinchtab nav` 自动启动的服务器，都不带任何 flag 运行 `pinchtab server`，因此 `server.logLevel` 是设置它们阈值的唯一途径（用 `pinchtab config set server.logLevel warn`）。这也意味着在默认 `info` 级别下，4xx 的原因不会被写出；需要时请把级别调高。

本段中的一切同样适用于 `pinchtab bridge`：它读取同一个 `server.logLevel`、接受同一个 `--log-level`，并按相同的优先级解析 —— 这一点很重要，因为 bridge 持有 CDP 会话，因而它拥有「目标崩溃、实例生命周期、选择器解析」这些日志的归属权。唯一的区别是 `bridge` 没有 `-v`：`-v` 还会打开服务器的启动横幅，而 bridge 没有横幅可开，所以 `--log-level debug` 才是调高 bridge 的方式。由编排器派生的 bridge 会通过其子配置继承该级别。

## 导航

```bash
pinchtab nav <url>                      # Navigate current tab, or create one if needed
pinchtab nav <url> --tab <id>           # Reuse a specific tab
pinchtab nav <url> --new-tab            # Explicitly force a new tab
pinchtab nav <url> --timeout 90          # Allow up to 90s (maximum 120s)
pinchtab nav <url> --block-images       # Block images for this navigation
pinchtab nav <url> --block-ads          # Block ads for this navigation
pinchtab nav <url> --snap               # Navigate and output interactive snapshot
pinchtab nav <url> --text               # Navigate and output page text
pinchtab nav <url> --print-tab-id       # Print only the tab ID, whatever stdout is (with --snap/--text the tab ID goes to stderr)
pinchtab back                           # Go back in the active tab
pinchtab back --tab <id>                # Go back in a specific tab
pinchtab forward                        # Go forward in the active tab
pinchtab reload                         # Reload the active tab
```

`back`、`forward` 和 `reload` 总会报告页面实际落定的 URL，因此即使不加 `--snap`，重定向、登录墙或错误页面也都可见。

`nav` 则不同，因为它的标准输出是其他命令要消费的值：它先打印标签页 ID，并且只在标准输出是终端时才追加落定 URL。在 `--print-tab-id` 下、在管道或重定向中，标准输出只携带标签页 ID，因此 `TAB=$(pinchtab nav <url>)` 能捕获到一个可用的标签页 ID。脚本化导航若需要知道自己落定在哪里，应请求 `--json`（它打印携带 `url` 的响应体），或 `--snap`（它把落定 URL 放在节点上方的标题行里）。不要用 `--text`：URL 只在 IDPI 信任边界包装层内部才到达它，因此在 `security.idpi.enabled` 或 `security.idpi.wrapContent` 关闭的地方它会消失。

## 标签页

`tab` 命令只列出、聚焦和关闭标签页。它不代理其余的浏览器命令集。

```bash
pinchtab tab                            # List tabs
pinchtab tab <id>                       # Focus a tab by ID or 1-based index
pinchtab nav <url> --new-tab            # Open a new tab and navigate it
pinchtab tab close <id>                 # Close a tab
pinchtab close <id>                     # Same as tab close
```

对标签页作用域内的工作，使用带 `--tab` 的顶级命令：

```bash
pinchtab snap --tab <id>
pinchtab click --tab <id> <selector>
pinchtab pdf --tab <id> -o page.pdf
```

当调用方可被识别时，未加作用域的标签页命令会使用服务端的当前标签页状态：`PINCHTAB_SESSION` 按会话限定当前标签页，而在没有会话时，`--agent-id` 或 `PINCHTAB_AGENT_ID` 按代理 ID 限定。匿名的命令行界面调用仍使用共享的本地当前标签页状态文件。

## 交互

大多数元素命令接受统一的选择器：

- 快照引用，如 `e5`
- CSS 选择器，如 `#login`
- XPath，如 `xpath://button`
- 文本选择器，如 `text:Submit`
- 语义选择器，如 `find:login button`
- role/name 选择器，如 `role:button Save`
- label、placeholder、alt、title 或 test id 选择器，如 `label:Email`、`placeholder:Search`、`alt:Logo`、`title:Close`、`testid:submit`
- 位置包装器，如 `first:button`、`last:role:button` 或 `nth:2:button`（`nth` 从零开始）

位置包装器对每一种选择器都按**文档顺序**为候选者编号，因此 `nth:0:` 是页面中的第一个匹配，`nth:1:` 总在它之后。裸的 `text:` 选择器是唯一不编号的形式：它在那些最小的候选中挑选最像控件的那个，因此 `text:Save` 会优先选 `<button>` 而不是携带同一 label 的 `<div>`—— 这意味着 `text:X` 与 `first:text:X` 可能解析到不同元素。包装器永远只在裸选择器能找到的匹配中做选择；它从不改变存在哪些匹配。

选择器前缀不区分大小写，因此 `CSS:#login` 与 `css:#login` 含义相同。只有前缀做大小写折叠；其后的值原样透传。

`role:`、`label:`、`testid:` 这类结构化形式，由语义引擎对照增强后的快照描述符做匹配。CSS、XPath、引用、既有的 `text:` 动作选择器，以及裸的 CSS/text 包装器，仍走浏览器侧的选择器解析。

选择器查找按 frame 显式进行。未加作用域的选择器只搜索当前 frame 作用域，默认为 `main`。在基于选择器的 iframe 工作之前使用 `pinchtab frame ...`。支持同源 iframe 作用域；当前不暴露跨源 iframe 后代。

```bash
pinchtab frame                         # Show current frame scope
pinchtab frame "#payment-frame"        # Scope selectors to an iframe
pinchtab frame main                    # Return selector scope to the top document
pinchtab click [selector]               # Click an element or coordinates with --x/--y
pinchtab click <selector> --submit      # One terminal submit click; reports bounded post-submit state
pinchtab click --css <selector>         # Force CSS selector mode
pinchtab click --wait-nav <selector>    # Click and wait for navigation
pinchtab click --snap <selector>        # Click and output interactive snapshot
pinchtab dblclick [selector]            # Double-click
pinchtab type <selector> <text>         # Type via key events
pinchtab fill <selector> <text>         # Fill directly
pinchtab press <key>                    # Press a key
pinchtab hover [selector]               # Hover an element
pinchtab mouse move <x> <y>             # Move the mouse to coordinates
pinchtab mouse move [selector]          # Or move to an element center
pinchtab mouse down [selector]          # Press a mouse button
pinchtab mouse up [selector]            # Release a mouse button
pinchtab mouse wheel [dy|selector]      # Dispatch wheel deltas
pinchtab drag <from> <to>               # Drag between targets (selector/ref or x,y)
pinchtab drag <selector> --drag-x <n> --drag-y <n>  # Drag by a pixel offset
pinchtab focus [selector]               # Focus an element
pinchtab scroll <selector|pixels>       # Scroll an element or the page
pinchtab scroll down --snap             # Scroll and output snapshot
pinchtab scroll 800 --snap-diff         # Scroll and output snapshot diff
pinchtab select <selector> <value>      # Select a <select> option
pinchtab check <selector>               # Check a checkbox or radio
pinchtab uncheck <selector>             # Uncheck a checkbox or radio
pinchtab scrollintoview <selector>      # Scroll an element into view
```

低层鼠标命令在拖动手柄、类似画布的 UI，以及 DOM 原生的 click 或 hover 抽象不够用的流程中很有用：

```bash
pinchtab mouse move e5
pinchtab mouse down --button left
pinchtab mouse up --button left
pinchtab mouse wheel 240 --dx 40
pinchtab mouse move --x 400 --y 320
pinchtab drag e5 400,320
```

## 页面分析

```bash
pinchtab snap [selector]                # Accessibility snapshot, optionally scoped
pinchtab snap -i -c                     # Interactive + compact
pinchtab snap -d                        # Diff from previous snapshot
pinchtab snap --selector <css>          # Scope snapshot
pinchtab snap --max-tokens <n>          # Limit token budget
pinchtab snap --depth <n>               # Limit tree depth
pinchtab snap --text                    # Text output
pinchtab text                           # Extract readable text
pinchtab text --full                    # Full page innerText
pinchtab text --raw                     # Raw extraction
pinchtab text --markdown                 # Markdown (preserves links, tables)
pinchtab text --markdown --output page.md # Write Markdown to a file (one-line confirmation)
pinchtab text --frame <frameId>         # Read text from one iframe
pinchtab html [selector]                # Document or element HTML (--max-chars, --frame)
pinchtab styles [selector]              # Computed styles (root element when omitted; --prop for one)
pinchtab title                          # Current tab title
pinchtab url                            # Current tab URL
pinchtab value <ref>                    # Current value of a form element
pinchtab attr <ref> <name>              # Value of one HTML attribute
pinchtab box <ref>                      # Bounding box of an element
pinchtab checked <ref>                  # Whether an element is checked
pinchtab enabled <ref>                  # Whether an element is enabled
pinchtab visible <ref>                  # Whether an element is rendered
pinchtab count <selector>               # Count elements matching a CSS selector
pinchtab find <query>                   # Semantic element search
pinchtab find --threshold <0-1>         # Minimum similarity score
pinchtab find --explain                 # Include score breakdown
pinchtab find --ref-only                # Print only the best ref
pinchtab extract --schema <file>        # Schema-typed data as JSON (see reference/extract.md)
pinchtab extract --schema -             # Read the schema from stdin
pinchtab extract --schema <file> --fields   # + field<TAB>ref<TAB>confidence table
pinchtab extract --schema <file> --scope role:table --max-items 5  # Confine and cap
pinchtab eval <expression>              # Evaluate JavaScript
pinchtab a11y audit                     # Accessibility score + findings (native engine)
pinchtab a11y audit --axe               # Run axe-core in the page (industry rule ids)
pinchtab a11y audit --axe --tags wcag2a,wcag2aa   # axe: filter by WCAG tags
pinchtab a11y audit --axe --rules image-alt,label # axe: run only these rule ids
pinchtab a11y audit --axe --json        # Full JSON envelope with per-node refs
pinchtab memory                         # JS heap usage and DOM counters for the tab
pinchtab memory --gc --json             # Collect garbage first, raw JSON
pinchtab memory snapshot                # V8 heap snapshot to a server-side file (security.allowMemory)
pinchtab memory snapshot --out app.heapsnapshot  # Also copy it to a local path
pinchtab memory summary <id> --top 5    # Top constructors and duplicate strings of a snapshot
pinchtab memory compare <a> <b> --top 5 # Constructor growth from snapshot a to snapshot b
pinchtab memory compare <a> <b> --retained  # Also retained sizes from b's dominator tree
```

`pinchtab memory snapshot`、`summary` 和 `compare` 需要 `security.allowMemory`，因为堆快照持有页面上的每一个字符串。参见 [reference/memory.md](reference/memory.md)。

`pinchtab a11y audit --axe` 在页面的隔离世界（isolated world）中运行内置的 axe-core 引擎（见 [reference/a11y.md](reference/a11y.md)），因此页面脚本无法篡改结果。每一个能映射到快照引用的违规节点都会带上该引用，于是失败元素可以直接用 `pinchtab click` 和其他元素命令来处理。

`pinchtab eval` 刻意不做 frame 作用域限定。当前的 `pinchtab frame` 状态会影响基于选择器的命令（如 `snap`、`click`、`fill`、`type`），并且在未显式提供 `--frame` 时也会影响 `text`。

基于选择器的操作现在在选择器不匹配时会快速失败。如果 UI 仍在加载，请先使用 `pinchtab wait`，而不是依赖动作超时。

对于一个动作绝不能被重试的表单按钮，使用 `pinchtab click <selector> --submit`。它精确执行一次 DOM 点击，并报告一段简短的提交后观察，而不是重试投递。它不能与 `--wait-nav`、`--mode` 或 `--humanize` 组合使用；这些工作流请用普通点击。`pinchtab fill <selector> <text> --submit` 仍是「填充后按回车」的独立快捷方式。

## 独立桥接（Standalone Bridges）

独立的 `pinchtab bridge` 进程在运行时会把自己注册到本地状态目录。无需发送信号即可查看它们：

```bash
pinchtab bridges list
pinchtab bridges list --json
pinchtab bridges list --prune   # Remove only records whose original PID is dead or reused
```

`--prune` 绝不杀死任何进程。它只移除确属陈旧的注册表记录；一个 PID 仍存活或未知、但不可达的监听器仍会保持可见，供运维排查。

## 键盘、等待与诊断

```bash
pinchtab keyboard type <text>           # Type at the focused element
pinchtab keyboard inserttext <text>     # Insert text without key events
pinchtab keydown <key>                  # Hold a key down
pinchtab keyup <key>                    # Release a key
pinchtab wait <selector>                # Wait for selector to be visible
pinchtab wait <selector> --state hidden # Wait for selector to disappear
pinchtab wait <ms>                      # Fixed duration sleep (escape hatch; max 30000ms — prefer condition-based waits)
pinchtab wait --text <text>             # Wait for page text to appear
pinchtab wait --not-text <text>         # Wait for page text to disappear
pinchtab wait --url <glob>              # Wait for URL match (glob: **, *, ?)
pinchtab wait --load <state>            # state: ready-state | content-loaded | network-idle
                                        #   ready-state    → document.readyState === 'complete'
                                        #   content-loaded → readyState in {interactive, complete}
                                        #   network-idle   → 0 in-flight requests for 500ms (HTTP `idleFor` overrides; no CLI flag)
pinchtab wait --fn <expression>         # Wait for JS to become truthy
pinchtab wait ... --timeout-ms <ms>     # Override timeout in ms (default 10000, max 30000); --timeout is a deprecated alias
pinchtab network                        # List captured network requests
pinchtab network <requestId>            # Show one request in detail
pinchtab network --stream               # Stream network entries
pinchtab network --clear                # Clear captured network data
pinchtab network route <url> --abort    # Block matching requests (--body '<json>' fulfills instead)
pinchtab network unroute [url]          # Remove one interception rule, or all of them
pinchtab network rules                  # List interception rules
# HAR / NDJSON export is available over HTTP (no dedicated CLI subcommand):
#   curl http://127.0.0.1:9867/network/export                 → HAR 1.2 archive
#   curl http://127.0.0.1:9867/network/export?format=ndjson   → NDJSON (one entry per line)
#   curl http://127.0.0.1:9867/network/export?body=1          → include response bodies
#   curl http://127.0.0.1:9867/network/export/stream          → live HAR stream
# Per-tab variants live under /tabs/{id}/network/export[/stream].
pinchtab dialog accept [text]           # Accept alert/confirm/prompt
pinchtab dialog dismiss                 # Dismiss dialog
pinchtab console                        # Show console logs
pinchtab console --clear                # Clear console logs
pinchtab errors                         # Show browser error logs
pinchtab errors --clear                 # Clear browser error logs
pinchtab clipboard read                 # Read server-side clipboard text
pinchtab clipboard write <text>         # Write clipboard text
pinchtab clipboard copy <text>          # Alias for write
pinchtab clipboard paste                # Alias for read
pinchtab cache clear                    # Clear browser HTTP disk cache
pinchtab cache status                   # Check if cache can be cleared
```

手动交接（handoff）与恢复可通过命令行界面和 API 获得：

```bash
pinchtab tab handoff <tabId> --reason captcha --timeout-ms 120000
pinchtab tab handoff-status <tabId>
pinchtab tab resume <tabId> --status completed
```

`pinchtab handoff`、`pinchtab handoff-status` 和 `pinchtab resume` 是同样这三个命令的顶级写法。

API 等效项：

暂停中的交接状态会以 `409 tab_paused_handoff` 阻断动作执行路由（`/action`、`/actions`、`/macro`），直到通过超时恢复或过期。

```bash
curl -X POST "$PINCHTAB_SERVER/tabs/<tabId>/handoff"
curl "$PINCHTAB_SERVER/tabs/<tabId>/handoff"
curl -X POST "$PINCHTAB_SERVER/tabs/<tabId>/resume"
```

## 捕获与导出

```bash
pinchtab screenshot                     # Save a screenshot to a generated .jpg path
pinchtab screenshot -o <path>           # Save to a chosen path (.png infers PNG; otherwise JPEG)
pinchtab screenshot --format <jpeg|png>  # Override the inferred output format
pinchtab screenshot -q <0-100>          # JPEG quality
pinchtab screenshot -s <selector>       # Capture a specific element by selector
pinchtab screenshot --scale 0.5         # Half-size output (quarter the pixels)
pinchtab screenshot --beyond-viewport   # Capture the full scrollable document (ignored with -s)
pinchtab screenshot --annotate          # Bake numbered ref boxes into the image (for vision models)
pinchtab annotate                       # Inject a persistent, clickable overlay on the LIVE page
pinchtab annotate -s <selector>         # Scope the overlay to elements within a selector
pinchtab annotate --clear               # Remove the persistent overlay
pinchtab capture                        # Paired screenshot + accessibility snapshot from the same DOM epoch
pinchtab capture -o <path>              # Save the paired image to a chosen path
pinchtab capture --beyond-viewport      # Capture the full document; bounds in page coords
pinchtab capture --require-pair         # Fail (409) if the page navigated mid-capture
pinchtab capture --with-bounds=false    # Skip per-node DOM.getBoxModel round trips
pinchtab capture --scale 0.5            # Half-size image (snapshot/bounds unchanged)
pinchtab pdf                            # Export the active page as PDF
pinchtab pdf -o <path>                  # Save PDF to a chosen path
pinchtab pdf --landscape                # Landscape orientation
pinchtab pdf --scale <n>                # Print scale
pinchtab pdf --paper-width <in>         # Paper width in inches
pinchtab pdf --paper-height <in>        # Paper height in inches
pinchtab pdf --page-ranges <r>          # Page ranges such as 1-3
pinchtab pdf --prefer-css-page-size     # Use CSS page size
pinchtab pdf --display-header-footer    # Show header/footer
pinchtab download <url>                 # Download through the browser session
pinchtab download <url> -o <path>       # Save downloaded file to a path
pinchtab upload <file>                  # Upload to the default file input
pinchtab upload <file> -s <css>         # Upload to a specific file input
pinchtab record start <file>            # Start recording (.webm, .mp4, .gif)
pinchtab record start <file> --fps 10   # Custom frame rate (default 5)
pinchtab record start <file> --quality 90 # JPEG capture quality (default 80)
pinchtab record start <file> --scale 0.5  # Half resolution
pinchtab record stop                    # Stop recording and save
pinchtab record status                  # Check recording status
```

## 存储与状态

key 或状态名是第一个参数。为兼容既有脚本，`--key` / `--name` 被接受为同一值；同时传位置参数和 flag 会被拒绝。

```bash
pinchtab storage get                    # Both localStorage and sessionStorage for the tab's origin
pinchtab storage get --type local       # One store
pinchtab storage get <key>              # A single item (same as --key <key>)
pinchtab storage set <key> <value>      # Write an item (localStorage unless --type session)
pinchtab storage delete <key>           # Remove one key (same as --key <key>); a key is required
pinchtab storage clear                  # Wipe localStorage (--type session for sessionStorage)
pinchtab storage clear --all            # Wipe both stores
pinchtab state                          # Current browser state for the tab
pinchtab state list                     # List saved state files
pinchtab state save [name]              # Save cookies and storage (name auto-generated if omitted)
pinchtab state save <name> --encrypt    # Save encrypted (needs security.stateEncryptionKey)
pinchtab state load <name>              # Restore a saved state; <name> may be a prefix (newest match)
pinchtab state show <name>              # Print a saved state file
pinchtab state delete <name>            # Delete a saved state file
pinchtab state clean --older-than <h>   # Remove state files older than <h> hours (default 24)
pinchtab cookies get                    # Cookies for the tab's current page (--name, --url)
pinchtab cookies set <name> <value>     # Set a cookie (--domain, --path, --secure, --http-only, --same-site)
pinchtab cookies clear                  # Clear ALL browser cookies, every origin
```

`storage delete` 从不清空一个存储：裸的 `pinchtab storage delete` 会被拒绝，并提示 `storage clear`—— 那是唯一一个清空动词。每一个 storage 动词都接受 `--tab <id>`，`state`、`state save` 和 `state load` 也是。`state` 家族需要 `security.allowStateExport`。

## 模拟（Emulation）

```bash
pinchtab set viewport <w> <h>           # Viewport size (--dpr, --mobile)
pinchtab set geo <lat> <lon>            # Geolocation (--accuracy)
pinchtab set media <feature> <value>    # CSS media feature, e.g. prefers-color-scheme dark
pinchtab set offline <true|false>       # Toggle offline mode
pinchtab set headers '<json>'           # Extra HTTP headers ({} clears them)
pinchtab set credentials <user> <pass>  # HTTP auth credentials
```

## 站点审计、对比与抓取

```bash
pinchtab audit <url>                    # Browser-enriched page audit (see audit.md)
pinchtab compare <live-url> <staging-url>  # Visual + data diff of two site versions
pinchtab scrape <url>                   # HTTP crawl, browser-render only thin pages (see scrape.md)
```

flags 与报告结构见 [audit.md](audit.md) 和 [scrape.md](scrape.md)。

## 实例、配置文件与活动

```bash
pinchtab instance list                  # List running instances
pinchtab instance start                 # Start an instance
pinchtab instance start --profile <id-or-name>
pinchtab instance start --mode headed
pinchtab instance start --port <n>
pinchtab instance start --extension /path/to/ext
pinchtab instance stop <id>             # Stop an instance
pinchtab instance restart <id>          # Soft-restart an instance's browser process
pinchtab instance logs <id>             # Show instance logs
pinchtab instance navigate <id> <url>   # Open a tab in an instance already on <url> (one step)
pinchtab profiles                       # List profiles
pinchtab profiles create <name>         # Create a profile for human setup and login
pinchtab profiles prune                 # List reclaimable quarantined profiles (removes nothing)
pinchtab profiles prune --confirm       # Remove them and report the disk freed
pinchtab profiles prune --profile <dir> # Reclaim just one quarantined directory
pinchtab activity                       # List recorded activity events
pinchtab activity tab <tab-id>          # Filter activity by tab
pinchtab health                         # Check server health
```

### 回收隔离（quarantine）的配置文件

当某个配置文件的浏览器数据变得不可读时，PinchTab 会把该目录改名为 `<profile>.quarantine-<timestamp>`，并从空配置文件重新启动它。这些目录此后再也不会被读取，因此纯粹是磁盘开销。

有两种方式移除它们，二者回答的是不同问题：

- **自动清理（automatic prune）** 在*每次隔离时、按单个配置文件*约束累积。每次一个配置文件被隔离时，会移除**同一配置文件**的更旧隔离副本，保留其中的 `profiles.quarantineKeep` 份（默认 1）。它只作为一次新隔离的副作用运行，因此一个只被隔离过一次、此后再未被隔离的配置文件，会无限期保留其副本；而 `profiles.quarantineKeep: 0` —— 文档记载的「全部保留」方式 —— 会把它完全关掉。
- **`pinchtab profiles prune`** *按需、跨所有配置文件*回收。它回答的是「现在就把磁盘还给我」，包括那些自动清理永远不会再回头处理的隔离。它完全无视 `quarantineKeep`，因此即使自动策略保留一切，仍留有一条显式回收的途径。

没有任何东西被调度，也没有任何东西在启动时运行；按需路径只在你要求时才运行。

裸命令是一次空跑（dry run）—— 它打印将要移除什么、将要释放多少总量，但不删除任何东西，因此供代理安全运行：

```bash
$ pinchtab profiles prune
default.quarantine-1748100001	412.6 MB
work.quarantine-1748100002	1.1 GB

2 quarantined profile(s), 1.5 GB reclaimable. Nothing was removed; re-run with --confirm.
```

资格判定依据是隔离名模式 `<profile>.quarantine-<timestamp>`，而不是 PinchTab 实际隔离了什么的记录 —— 因此你以这种形状的名字创建的配置文件也符合资格，并且在其他所有地方也被列为已隔离。裸的空跑正是你在任何东西被移除之前看到这一点的地方。模式之外的东西无论传什么都不会被移除。

`--profile` 指定的是一个隔离目录，绝不是文件系统路径 —— 传路径会被拒绝。要删除普通配置文件，请改用配置文件删除路由；本命令无法触及它。

经 HTTP 的同一操作是 `POST /profiles/prune`，用 `{"confirm": true}` 执行移除，用可选的 `"profile"` 收窄范围。每一次移除都会记录其路径与释放的字节数。

## 配置与安全

```bash
pinchtab config                         # Interactive config overview/editor
pinchtab config init                    # Create a default config file
pinchtab config show                    # Print effective runtime config
pinchtab config token                   # Copy server.token to the clipboard without printing it
pinchtab config token --stdout          # Print server.token to stdout (headless hosts, $(...) capture)
pinchtab config path                    # Print config file path
pinchtab config validate                # Validate the current config file
pinchtab config schema                  # Print the config JSON Schema URL (--print for the schema)
pinchtab config get <path>              # Read one file-config value
pinchtab config set <path> <val>        # Set one file-config value
pinchtab config patch <json>            # Merge JSON into the config file
pinchtab security                       # Interactive security overview
pinchtab security up                    # Apply stricter defaults
pinchtab security down                  # Apply documented guards-down preset
pinchtab doctor                         # Read-only install and browser checks (--json, --check <name>)
pinchtab doctor browsers                # Configured and known browsers with availability
pinchtab doctor browser [name]          # Browser availability, or checks for one target
pinchtab version                        # Print the PinchTab version
```

## 全局 Flags

根命令支持：

```bash
PINCHTAB_TOKEN=<that-host-token> pinchtab --server http://host:9867 <command>
pinchtab --help
pinchtab --version
```

一个非回环（non-loopback）的 `--server` 主机，要求在同一条命令上通过 `PINCHTAB_TOKEN`（或 `PINCHTAB_SESSION`）提供其凭据 —— 命令行界面拒绝把本地配置的 `server.token` 发到本机之外。

当前带 `--tab` 的命令包括：

- `nav`
- `back`
- `forward`
- `reload`
- `snap`
- `screenshot`
- `capture`
- `pdf`
- `find`
- `extract`
- `text`
- `click`
- `dblclick`
- `hover`
- `mouse move`
- `mouse down`
- `mouse up`
- `mouse wheel`
- `focus`
- `type`
- `press`
- `fill`
- `scroll`
- `select`
- `eval`
- `check`
- `uncheck`
- `keyboard type`
- `keyboard inserttext`
- `keydown`
- `keyup`
- `scrollintoview`
- `network`
- `wait`
- `dialog accept`
- `dialog dismiss`
- `console`
- `errors`
- `frame`、`html`、`styles`、`title`、`url`
- `value`、`attr`、`box`、`checked`、`enabled`、`visible`、`count`
- `drag`、`download`、`upload`、`annotate`
- `a11y audit`、`memory`、`memory snapshot`、`record start`
- `network route`、`network unroute`、`network rules`
- `cookies get`、`cookies set`、`set`（每一个子命令）
- `storage`（每一个子命令）、`state`、`state save`、`state load`

## 输出格式

大多数命令默认输出人类可读文本。使用 `--json` 获得机器可解析的 JSON 输出：

```bash
pinchtab tab                            # Human-readable: *abc123  https://...  Page Title
pinchtab tab --json                     # JSON: {"tabs":[...]}
pinchtab frame                          # Human-readable: main
pinchtab frame --json                   # JSON: {"tabId":"...","scoped":false,...}
pinchtab network                        # Human-readable: GET  200  https://...
pinchtab network --json                 # JSON: {"entries":[...],"count":5}
```

**对脚本与自动化而言**：在管道传递输出或以编程方式解析时，始终使用 `--json`。人类可读格式可能在版本之间变化，不保证稳定。JSON 模式才是稳定契约。

带 `--json` 的命令包括：`tab`、`frame`、`network`、`click`、`type`、`scroll`、`nav`、`back`、`forward`、`reload`、`wait`、`find`、`extract`、`eval`，以及大多数动作命令。
