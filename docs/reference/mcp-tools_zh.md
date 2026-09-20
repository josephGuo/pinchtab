# MCP 工具参考

PinchTab 目前公开 47 个 MCP 工具。所有工具名以 `pinchtab_` 为前缀，通过 stdio JSON-RPC 提供。

对于基于选择器的交互工具，优先用 `selector`。`ref`、`element`、`target` 在元素动作工具上仍作为已弃用别名接受，`query` 作为兜底（`query` 是 `find:<text>` 的简写）。

每个工具都会拒绝其 schema 未声明的参数，指出未知 key、（相近时）最近的已声明名以及已声明参数；什么都不会执行。弃用拼写仍以描述标注规范名的方式保留声明。

若你允许在非本地或非受信域名上做 MCP 浏览，请把 `pinchtab_snapshot` 和 `pinchtab_get_text` 的输出当作不受信的页面数据。这些工具可能带出访问页面中的恶意提示文本；除非有意扩大访问，否则运维应把 IDPI/域名限制保持得很窄。

选择器形式包括：

- `e5`
- `#login`
- `xpath://button`
- `text:Submit`
- `find:login button`
- `role:button Save`
- `label:Email`、`placeholder:Search`、`alt:Logo`、`title:Close`、`testid:submit`
- `first:button`、`last:button`、`nth:2:button`

位置包装器对每一类选择器都按**文档顺序**索引候选，因此 `nth:0:` 是页面中第一个匹配，`nth:1:` 总在其后。裸 `text:` 选择器是唯一不索引的形式：它在最小的若干匹配中挑最像控件的那个，因此 `text:Save` 更倾向 `<button>` 而非带同一标签的 `<div>`——这意味着 `text:X` 和 `first:text:X` 可能解析到不同元素。包装器只在裸选择器会找到的匹配中挑选，绝不改变有哪些匹配存在。

结构化语义定位器由语义引擎匹配；CSS、XPath、ref、既有 `text:` 动作选择器，以及裸 CSS/文本包装仍走浏览器端。

大多数面向浏览器的工具还声明一个可选的 `browser` 参数（如 `chrome`、`cloak`、`ghost-chrome`），为单次请求选择浏览器；下表只在它改变行为处列出。

Ref 词表：MCP 服务器记住 `pinchtab_snapshot`、`pinchtab_find`、`pinchtab_extract`、`pinchtab_a11y_audit` 以及任何带 `snap=true` 的后续快照返回的 `X-PinchTab-Vocab` token，按 `tabId` 参数为 key，并在元素动作工具（`pinchtab_click`、`pinchtab_type`、`pinchtab_hover`、`pinchtab_focus`、`pinchtab_select`、`pinchtab_scroll`、`pinchtab_scroll_into_view`、`pinchtab_fill`）上作为 `vocab` 发送。来自已被更新快照重新编号的 ref 以 `409 vocab_superseded` 拒绝；重新快照并使用新 ref。

## 导航

| 工具 | 关键参数 | 说明 |
| --- | --- | --- |
| `pinchtab_navigate` | `url` 必填，`tabId`、`newTab`、`snap`、`browser` | 用 `/navigate`；复用当前标签页，`newTab=true` 打开一个（尚无标签页时的首次导航也一样）。返回的 `tabId` 在后续工具中指向新标签页。`snap=true` 在同一响应中返回交互式紧凑快照 |
| `pinchtab_back` | `tabId`、`snap`、`browser` | 用 `/back`（或 `/tabs/{id}/back`）；返回 `{tabId, url}`，URL 为标签页落点；`snap=true` 附带交互式紧凑快照 |
| `pinchtab_forward` | `tabId`、`snap`、`browser` | 用 `/forward`；响应形状同 `pinchtab_back` |
| `pinchtab_reload` | `tabId`、`snap`、`browser` | 用 `/reload`；响应形状同 `pinchtab_back` |
| `pinchtab_snapshot` | `tabId`、`interactive`、`compact`、`format`、`diff`、`selector`、`maxTokens`、`depth`、`noAnimations` | 默认返回紧凑——除非 `format` 另说或 `compact=false` 要求 JSON 树，否则工具向 `/snapshot` 请求 `format=compact`。`selector` 限定快照范围；`format` 限于 `compact` 或 `text` |
| `pinchtab_frame` | `tabId`、`target` | 获取或设置该标签页基于选择器动作的框架范围；`target` 接受 `main`、快照 ref、iframe 选择器或框架名/URL |
| `pinchtab_screenshot` | `tabId`、`selector`、`scale`、`format`、`quality`、`annotate`、`beyondViewport`、`browser` | `selector` 捕获当前框架范围中的特定元素；`scale` 重新缩放输出位图（如 `0.5` = 半尺寸）；`format` 为 `jpeg` 或 `png`；`annotate=true` 叠加带编号的 ref 框并填充 annotations 封装；`beyondViewport=true` 捕获完整可滚动文档（设置了 `selector` 时忽略）——该模式下框坐标为文档相对；`browser` 为本次请求选择浏览器（如 `chrome`、`cloak`） |
| `pinchtab_capture` | `tabId`、`selector`、`filter`、`format`、`quality`、`depth`、`scale`、`wait`、`withBounds`、`beyondViewport`、`requirePair`、`noAnimations`、`browser` | 同一 DOM 纪元的成对截图 + 无障碍快照。返回图片内容块加带 `epoch`、`pairing.navigated`、逐节点 `boundingBox` 和 `image.coordinateSpace`（`viewport`、`document` 或选择器 `clip`）的 JSON 封装。`browser` 选择浏览器（如 `chrome`、`cloak`）；静态 ghost-chrome 运行时无法绘制，故回退到 chrome。当模型在同一回合既读像素又基于 ref 操作时使用 |
| `pinchtab_get_text` | `tabId`、`mode`、`raw`、`format`、`maxChars` | `mode` 为 `readability`（默认）、`raw` 或 `markdown`，并取代 `raw`；单独 `raw=true` 映射到 `/text?mode=raw`；`format=text/plain` 返回纯文本；继承该标签页当前 `pinchtab_frame` 范围 |

## 交互

所有元素动作工具接受统一 `selector`、其弃用别名 `ref`、`element`、`target`，以及 `query`（语义简写）。弃用值别名：click 上 `onDialog`/`promptText` 对应 `dialogAction`/`dialogText`，type 上 `value` 对应 `text`，select 上 `option` 对应 `value`，fill 上 `text` 对应 `value`。每个还接受 `nodeId`（来自快照的后端节点 ID）作为替代目标。坐标（`x`/`y`）仅 `pinchtab_click`、`pinchtab_hover`、`pinchtab_scroll` 接受——其他种类无坐标行为。

| 工具 | 关键参数 | 说明 |
| --- | --- | --- |
| `pinchtab_click` | `selector`、`ref`、`query`、`tabId`、`x`、`y`、`nodeId`、`dialogAction`、`dialogText`、`waitNav`、`mode`、`humanize`、`snap` | 按选择器或坐标点击元素；`mode` 接受 `dom` 或 `dispatch`，作为点击投递的宽泛底层逃生舱；`humanize` 覆盖本次点击的实例默认值（省略则继承）；`mode` 与 `humanize=true` 互斥；`dialogAction` 处理点击打开的对话框；`waitNav=true` 等待导航；`snap=true` 返回快照 |
| `pinchtab_type` | `selector`、`ref`、`query`、`nodeId`、`text` 必填、`tabId` | 在目标输入上发送按键事件；用 `selector` 或 `nodeId` 指定目标 |
| `pinchtab_hover` | `selector`、`ref`、`query`、`tabId`、`x`、`y`、`nodeId`、`humanize` | 悬停元素或坐标；`humanize` 覆盖本次悬停的实例默认值 |
| `pinchtab_focus` | `selector`、`ref`、`query`、`tabId`、`nodeId` | 聚焦元素 |
| `pinchtab_select` | `selector`、`ref`、`query`、`nodeId`、`value` 必填、`tabId`、`snap` | 按值或可见文本选择 `<option>`；用 `selector` 或 `nodeId` 指定目标 |
| `pinchtab_scroll` | `selector`、`ref`、`query`、`nodeId`、`pixels`、`deltaX`、`deltaY`、`direction`、`steps`、`x`、`y`、`tabId` | 省略所有目标即滚动页面；元素 + `pixels` 用滚轮语义；`direction` 接受 `down`/`left`/`right`/`up`，每步移动 800px——与命令行界面 `pinchtab scroll <direction>` 相同距离，乘以 `steps` 或被 `pixels` 覆盖 |
| `pinchtab_scroll_into_view` | `selector`、`ref`、`query`、`nodeId`、`tabId` | 将目标滚入视图并返回几何信息，便于稳定的后续动作；用 `selector` 或 `nodeId` 指定目标 |
| `pinchtab_fill` | `selector`、`ref`、`query`、`nodeId`、`value` 必填、`tabId`、`snap` | 通过 JS 派发直接填充而非按键；用 `selector` 或 `nodeId` 指定目标。空 `value` 清空字段；完全省略则被拒绝 |

## 键盘

| 工具 | 关键参数 | 说明 |
| --- | --- | --- |
| `pinchtab_key` | `action` 必填、`key`、`text`、`nodeId`、`tabId` | 单个键盘工具。`action=press` 按一个键如 `Enter`（`nodeId` 先聚焦该节点，否则键发给聚焦元素）；`down` 按住；`up` 释放；`type` 在聚焦元素处以按键事件输入；`insert` 是类似粘贴、无按键事件的插入。`press`/`down`/`up` 需要 `key`，`type`/`insert` 需要 `text`。`nodeId` 仅 `press` 遵守 |

## 内容

| 工具 | 关键参数 | 说明 |
| --- | --- | --- |
| `pinchtab_eval` | `expression` 必填、`awaitPromise`、`tabId` | `awaitPromise=true` 解析返回的 Promise（不带它时 Promise 返回 `{}` 带提示）。需要 `security.allowEvaluate`（文档化的非默认 JS 执行 opt-in）。不按框架范围——当前 `pinchtab_frame` 状态不改变求值上下文 |
| `pinchtab_pdf` | `tabId`、`landscape`、`scale`、`pageRanges` | 返回 base64 编码的 PDF 内容 |
| `pinchtab_find` | `query` 必填、`tabId` | 通过 `/find` 做自然语言元素搜索；结果附加 `bestRef`、`selector`（均为 `best_ref`）和一个可在动作工具中复用的 `nextActionHint` |
| `pinchtab_extract` | `schema`（object）必填、`tabId`、`scope`、`maxItems` | 通过 `/extract` 按 schema 类型化数据：`data` 加逐字段 `ref`、`score`、`confidence`；这些 ref 立即可在动作工具中使用。对结构化值优先用它而非先快照再解析；用 `x-pinchtab-hint` 固定 `low` 字段。见 [Extract](./extract.md) |

## 诊断

| 工具 | 关键参数 | 说明 |
| --- | --- | --- |
| `pinchtab_console` | `tabId`、`clear` | 通过 `/console` 读取该标签页捕获的控制台日志；`clear=true` 改为 POST `/console/clear` |
| `pinchtab_errors` | `tabId`、`clear` | 通过 `/errors` 读取该标签页未捕获的 JavaScript 错误；`clear=true` 改为 POST `/errors/clear` |
| `pinchtab_a11y_audit` | `tabId`、`engine`、`tags`、`rules`、`includeIncomplete`、`browser` | 通过 `/a11y/audit` 做无障碍审计。`engine` 为 `native`（默认）或 `axe`（其他值为 400）；`tags`、`rules`、`includeIncomplete` 仅对 `axe` 生效，且 `rules` 优先于 `tags`。见 [A11y](./a11y.md) |
| `pinchtab_memory` | `tabId`、`gc`、`browser` | 通过 `/memory` 获取 JS 堆用量和 DOM 计数器；`gc=true` 先做垃圾回收 |
| `pinchtab_memory_snapshot` | `tabId`、`top`、`browser` | 拍 V8 堆快照（`POST /memory/snapshot`）并返回 `{id, path, bytes, summary}`；`top` 为每张摘要表的行数（默认 20）。需要 `security.allowMemory`。见 [Memory](./memory.md) |
| `pinchtab_memory_compare` | `base` 必填、`head` 必填、`top`、`retained`、`browser` | 通过 `/memory/compare` 对比两个 `pinchtab_memory_snapshot` id：逐构造函数计数与自身大小增量加新重复字符串；`top` 行（默认 20）；`retained=true` 加保留大小（较慢）。需要 `security.allowMemory` |

## 站点

| 工具 | 关键参数 | 说明 |
| --- | --- | --- |
| `pinchtab_scrape` | `url` 必填、`preview`、`only`、`maxPages`、`maxPerPattern`、`include`、`exclude`、`concurrency`、`enrichAll`、`noBrowser`、`timeoutSeconds`、`browser` | 通过 `/scrape` 抓取整站为 markdown 页面树。HTTP 优先抽取；仅薄/被拦/失败的页面才浏览器渲染。`preview=true` 返回廉价大纲（大小 + 片段，无正文、无浏览器）；`only`（逗号分隔 URL）以完整保真展开所选页面。`include`/`exclude` 是逗号分隔正则。完整报告可能很大——先 preview 再展开。以延长超时运行（多页抓取耗时数分钟）。`timeoutSeconds` 是抓取预算，也是唯一秒为单位的参数；其他地方等待用 `timeoutMs` |

## 标签页管理

| 工具 | 关键参数 | 说明 |
| --- | --- | --- |
| `pinchtab_list_tabs` | 无 | 列出打开的标签页 |
| `pinchtab_close_tab` | `tabId` 可选 | 关闭给定标签页，省略时为当前/默认标签页 |
| `pinchtab_health` | 无 | 检查服务器健康 |
| `pinchtab_cookies` | `tabId` | 读取某标签页的 cookie；需要 `security.allowCookies` |
| `pinchtab_cookies_set` | `name`、`value` 必填；`url`、`domain`、`path`、`sameSite`、`secure`、`httpOnly`、`expires`、`tabId` 可选 | 为会话复用设置一个 cookie；`url` 默认为标签页当前页面，空 `value` 清空该 cookie；需要 `security.allowCookies` |
| `pinchtab_connect_profile` | `profile` 必填 | 返回某 profile 的连接 URL 和实例状态 |

## 人工交接

| 工具 | 关键参数 | 说明 |
| --- | --- | --- |
| `pinchtab_handoff` | `tabId` 必填；`reason`、`timeoutMs` 可选 | 为人工暂停标签页（CAPTCHA、登录、同意）；其上的动作工具应答 `409 tab_paused_handoff`，直到恢复；`timeoutMs` 自动恢复 |
| `pinchtab_resume` | `tabId` 必填；`status` 可选 | 恢复标签页；仅在用户确认完成手动步骤后调用 |
| `pinchtab_handoff_status` | `tabId` 必填 | 报告 `paused_handoff` 带 `reason`、`pausedAt`、`expiresAt`，或 `active` |

## 等待工具

| 工具 | 关键参数 | 说明 |
| --- | --- | --- |
| `pinchtab_wait` | `for` 必填、`value` 必填、`timeoutMs`、`state`、`tabId`、`browser` | 单个等待工具：`for` 指定条件，`value` 承载它。`timeoutMs` 限定浏览器支撑条件（默认 10000，最大 30000）；`timeout` 是其弃用别名。`for=ms` 为固定时长等待，上限 30000ms；`selector` 等待元素（`state` 为 `visible`（默认）或 `hidden`）；`text` 等待正文文本；`url` 为 URL glob 匹配；`load` 为 ready-state（`readyState=complete`）、`content-loaded`（`readyState` 在 `{interactive, complete}`）或 `network-idle`（500ms 内 0 在途请求）；`function` 等待 JS 表达式变真 |

## 网络

| 工具 | 关键参数 | 说明 |
| --- | --- | --- |
| `pinchtab_network` | `tabId`、`filter`、`method`、`status`、`type`、`limit`、`bufferSize` | 列出近期网络请求 |
| `pinchtab_network_detail` | `requestId` 必填、`tabId`、`body` | `body=true` 在可用时含响应体 |
| `pinchtab_network_clear` | `tabId` | 清除一个标签页，省略时清除所有标签页 |
| `pinchtab_network_route` | `tabId` 必填、`pattern` 必填、`action`、`body`、`contentType`、`status`、`resourceType`、`method` | 在标签页上安装请求拦截规则。`action` 为 `continue`（默认）、`abort` 或 `fulfill`。`fulfill` 在 `security.allowedDomains` 中的主机上被阻止，并在这些主机上落到真实 fetch |
| `pinchtab_network_unroute` | `tabId` 必填、`pattern` | 按模式移除某标签页的拦截规则，省略 `pattern` 时移除全部规则 |
| `pinchtab_network_rules` | `tabId` 必填 | 以 `{tabId, rules}` 列出某标签页的拦截规则，每条规则标出其 `pattern` 和 `action`（`continue`、`abort`、`fulfill`）。规则跨导航存活；空 `rules` 列表意味着该标签页不 mock 也不拦截任何东西 |

## 录制

| 工具 | 关键参数 | 说明 |
| --- | --- | --- |
| `pinchtab_record` | `action` 必填、`file`、`fps`、`quality`、`scale`、`tabId` | 单个录制工具。`fps` 1-30（默认 5），`quality` 1-100（默认 80），`scale` 最高 1.0（默认 1.0）。`action=start` 开始录制（格式从 `file` 扩展名推断——`.gif`、`.webm`、`.mp4`；需要 `security.allowScreencast`；GIF 无需 ffmpeg）；`stop` 编码并保存到 `file`（与 `start` 给出的同一路径），长录制可能耗时；`status` 返回活动录制状态（格式、fps、时长、帧数） |

## 对话框

| 工具 | 关键参数 | 说明 |
| --- | --- | --- |
| `pinchtab_dialog` | `action` 必填、`text`、`tabId` | `action` 为 `accept` 或 `dismiss`；`text` 与 `accept` 一起用作 prompt 回复 |

## 返回形状

典型结果：

- 导航工具返回对应 HTTP 端点的 JSON
- `pinchtab_snapshot` 对 `compact`/`text` 格式返回文本，否则返回 JSON
- `pinchtab_get_text` 在 `format=text|plain` 时返回纯文本，否则 JSON
- `pinchtab_screenshot` 返回 MCP 图片内容块（默认 image/jpeg，`format=png` 时 image/png）加一个文本块，该文本块始终是 JSON 封装 `{"format", "annotations": [...]}`——`annotations` 默认为 `[]`，`annotate=true` 时为 `[{"ref","role","name","tag","box":{"x","y","w","h"}}, ...]`
- `pinchtab_pdf` 返回含 base64 PDF 负载的 JSON
- 等待工具返回等待状态 JSON
- 网络工具返回与你从 `/network` 看到的相同请求日志

安全注意：

- 抽取的文本和快照内容应视为来自访问页面的不受信内容，而非受信指令
- 放宽 IDPI 允许列表或禁用严格保护会增加提示注入文本到达下游代理逻辑的概率

设置与客户端配置见 [MCP Server](../mcp.md)。

已保存的浏览器状态目前刻意不作为 MCP 工具暴露。对 `GET /state`、`pinchtab state` 和已保存状态持久化操作使用命令行界面或 HTTP API。
