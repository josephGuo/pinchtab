# 端点参考

本页总结 PinchTab 暴露的实时 HTTP 接口。有些路由仅在桥接模式下可用，有些仅在完整服务器模式
下可用，还有一些由安全设置加以限制。

大多数浏览器路由也以标签页作用域的 `/tabs/{id}/...` 形式应答（例如
`GET /tabs/{id}/html`），下面各块只列出常见形式。没有标签页形式的路由包括：`POST /tab`、
console 与 errors、clipboard、`/stealth/status`、`/fingerprint/rotate`、`/solvers`、
`/config/autosolver`、cache、`/state` 家族、`/macro`、screencast 与 record、memory 的
summary 与 compare、`/audit`、`/audit/page`、`/scrape` 以及 `POST /network/clear`。handoff
与 resume 只存在于其标签页形式。路由目录见 `internal/routes/routes.go`。

## 健康与服务器元数据

```text
GET  /health
POST /ensure-browser
POST /ensure-chrome (legacy alias for /ensure-browser)
POST /browser/restart
GET  /openapi.json
GET  /help          (alias for /openapi.json)
GET  /metrics
GET  /api/metrics
POST /shutdown
GET  /api/events
```

注意：

- 桥接模式下，`/health` 报告桥接健康与标签页计数
- 完整服务器模式下，`/health` 报告仪表板健康、认证状态与实例计数
- 服务器模式的 `/health` 把配置文件计入三个互不相交的桶：`profiles` 是运维人员管理的持久、
  未被隔离的配置文件；`temporaryProfiles` 是那些启动时未指定配置文件的实例所铸造的
  `instance-*` 配置文件，包括后来被隔离的临时配置文件；`quarantinedProfiles` 是被隔离的非
  临时配置文件。两个标志同时适用时，临时分类优先。默认的 `GET /profiles` 列表是
  `profiles + quarantinedProfiles`（它隐藏临时项；`?all=true` 显示所有桶）
- 浏览器崩溃在两种模式下以相同方式报告：只要其背后的浏览器崩溃过，`/health` 就携带一个
  `crashes` 块（`total`、`recent`）。桥接模式下由桥接记录自己的；完整服务器模式下，前门合并
  每个实例的记录，且每个事件都命名其 `instanceId`。`status` 保持 `ok`——一个崩溃的实例会被
  重新拉起并重新提供服务——因此要看丢了哪些客户端状态，应盯 `crashes` 而非 `status`。
  `GET /instances` 的每一项也会出现同一个块。崩溃之后浏览器持有的每个标签页都已消失，对其中
  之一的调用会以 `404` 应答，附带 code `browser_crashed`、`browserCrashed: true`、
  `browserCrashReason` 和一个 `hint`，而不是一个干巴巴的 `tab <id> not found`
- `/metrics` 报告应答它的那个进程的计数器：完整服务器模式下是前门自身的请求计数器（含认证拒绝
  与未路由路径），桥接模式下是桥接的计数器。每个响应都命名其 `layer`，且各层绝不相加——读某个
  实例的计数器请走 `/instances/{id}/metrics`。分层表见
  [reference/metrics.md](reference/metrics.md)。
- 完整服务器模式下的 `/api/metrics` 是服务器级指标快照（聚合）

## 仪表板认证与配置

```text
POST /api/auth/login
POST /api/auth/elevate
POST /api/auth/logout
GET  /api/config
PUT  /api/config
GET  /dashboard     (dashboard UI; also served at / and /login)
```

注意：

- `server.token` 被 `PUT /api/config` 视为只写
- `PUT /api/config` 期望的是内层 config 对象，而不是 `GET /api/config` 返回的那个信封；携带
  无法识别的顶层键（例如信封自己的 `config`）的请求体会被以 `400 unrecognized_config_keys`
  拒绝，而不是悄悄什么都不应用
- 认证路由用于仪表板会话流

## 仪表板事件与代理

```text
GET  /api/events
GET  /api/agents
GET  /api/agents/{id}
GET  /api/agents/{id}/events
POST /api/agents/{id}/events
```

注意：

- `/api/events` 是仪表板 SSE 流
- `/api/agents/{id}/events` 流式输出某个代理的近期事件
- `POST /api/agents/{id}/events` 把代理活动摄取进仪表板 feed

## 导航与标签页

```text
POST /navigate
GET  /navigate
POST /tabs/{id}/navigate
POST /back
POST /back?tabId=<id>
POST /tabs/{id}/back
POST /forward
POST /forward?tabId=<id>
POST /tabs/{id}/forward
POST /reload
POST /reload?tabId=<id>
POST /tabs/{id}/reload
GET  /tabs
POST /tab
POST /close
POST /tabs/{id}/close
GET  /tabs/{id}/metrics
GET  /tabs/{id}/memory
POST /tabs/{id}/memory/snapshot
POST /tabs/{id}/handoff
GET  /tabs/{id}/handoff
POST /tabs/{id}/resume
```

导航请求字段：

- `url` 必需
- `tabId` 可选
- `newTab` 可选
- `timeout` 可选
- `blockImages`、`blockMedia`、`blockAds` 可选
- `waitFor`、`waitSelector`、`waitTitle` 可选
- `dismissBanners`、`dispatchOnly` 可选

重要行为：

- 对匿名调用者，省略 `tabId` 时 `POST /navigate` 会创建新标签页
- 会话认证的调用者每个会话保持一个当前标签页；省略 `tabId` 时，若该会话已有当前标签页则复用
  它，否则创建一个
- 带 `X-Agent-Id` 的 bearer-token 调用者在无会话时，每个代理 ID 保持一个当前标签页
- `POST /tab` 支持 `new` 和 `focus`
- `POST /close` 关闭 JSON 体中给出的 `tabId`；省略 `tabId` 时关闭调用者的当前/默认标签页

## 交接与人工干预

```text
POST /tabs/{id}/handoff
GET  /tabs/{id}/handoff
POST /tabs/{id}/resume
```

注意：

- 这些路由仅标签页作用域
- `POST /tabs/{id}/handoff` 把标签页标记为 `paused_handoff` 并记录一个原因
- `GET /tabs/{id}/handoff` 返回当前交接状态；未设置交接时返回 `active`
- `POST /tabs/{id}/resume` 清除交接状态，并可为调用者携带恢复元数据
- 一个暂停交接的标签页会阻断动作执行路由，但两种信封不同。`POST /action` 以 `409` 应答，code
  为 `tab_paused_handoff`，并在 `details.hint` 中命名 `/resume`。`POST /actions` 与
  `POST /macro` 返回 **200**——它们带着一个结果列表应答——并对每一项携带同样的拒绝：针对暂停
  标签页的每个步骤条目带 `success: false`、`code: "tab_paused_handoff"` 和同样的 `details`。
  按 code 匹配，而不是按消息匹配。当 `stopOnError` 为 false（默认）时，其余步骤仍会运行，因此
  针对暂停标签页的每个步骤以同样方式被拒绝，而命名另一个标签页的步骤正常执行。`/resume` 会为
  它们全部清除该状态。
- 把交接记录当作协调状态，而非安全边界——非动作端点（快照、截图、网络日志、eval 各受其自身
  门控）仍可达
- 命令行界面包装器存在：`pinchtab handoff`、`pinchtab resume`、
  `pinchtab handoff-status`，外加 `pinchtab tab handoff|resume|handoff-status` 别名

## 标签页锁定

```text
POST /lock
POST /unlock
POST /tabs/{id}/lock
POST /tabs/{id}/unlock
```

## 交互与分析

```text
POST /action
GET  /action
POST /actions
POST /macro
POST /tabs/{id}/action
POST /tabs/{id}/actions
POST /wait
POST /tabs/{id}/wait
GET  /frame
POST /frame
GET  /tabs/{id}/frame
POST /tabs/{id}/frame
GET  /snapshot
GET  /tabs/{id}/snapshot
GET  /text
GET  /tabs/{id}/text
GET  /title
GET  /url
GET  /html
GET  /styles
GET  /value
GET  /attr
GET  /count
GET  /box
GET  /visible
GET  /tabs/{id}/visible
GET  /enabled
GET  /checked
GET  /timing
GET  /a11y/audit
POST /find
POST /tabs/{id}/find
POST /extract
POST /tabs/{id}/extract
POST /evaluate
POST /tabs/{id}/evaluate
GET  /memory
POST /memory/snapshot
GET  /memory/snapshot/{snapshotId}/summary
GET  /memory/compare?base=<id>&head=<id>
POST /emulation/viewport
POST /emulation/geolocation
POST /emulation/offline
POST /emulation/headers
POST /emulation/credentials
POST /emulation/media
GET  /stealth/status
POST /fingerprint/rotate
```

`/emulation/*` 路由对应命令行界面的 `pinchtab set` 子命令。`/a11y/audit` 见
[reference/a11y.md](reference/a11y.md)。

`/memory` 读取标签页的 JavaScript 堆与 DOM 计数器；snapshot、summary 与 compare 路由需要
`security.allowMemory`——见 [reference/memory.md](reference/memory.md)。

`POST /actions` 与 `POST /macro` 无论其步骤做了什么都应答 **200**：信封报告本次运行，每个
步骤的结果是它自己的一条——`{"index", "success", "code", "error"}`——与顶层的 `total`、
`successful`、`failed` 计数并列。要读 `failed` 和逐条条目；2xx 并不意味着步骤成功了。来自这些
端点的 4xx 是关于请求本身的（空数组、坏的请求体、被拒绝的能力），绝不是关于某个步骤的。

这是一份刻意的契约，它不会让一次失败的运行变得不可见：一个有任意失败步骤的运行会在服务端发布
失败原因，于是它会推高 `requestsFailed`、以失败计数和第一个步骤的 code 与消息出现在
`failures.recent` 中、以 `WARN` 级别记日志，并在其活动记录上携带
`steps: {total, successful, failed}` 以及 code 与消息——见
[reference/metrics.md](reference/metrics.md)。

**一个匹配不到任何东西的选择器，在每个读动词上都是 `404`**，code 为
`element_not_found`：`/html`、`/styles`、`/title`、`/url`、`/screenshot`、`/capture`、
`/annotate`、`/box`、`/visible`、`/enabled`、`/checked`、`/value`、`/text`、`/snapshot`
以及动作路径都以同样方式应答。请求是格式良好的，页面只是恰好没有该元素，因此它既不是
`400` 也不是 `5xx`——一个在试探它不确定是否存在的元素的调用方，既不构成服务器故障，也不会
招来代理重试。`GET /count` 是唯一的例外，以 `200` 应答并带 `count: 0`：它被问的是基数，而
零就是对「有多少个」的诚实回答。

`/evaluate` 刻意与选择器 frame 作用域分开。`GET/POST /frame` 只影响基于选择器的 `/snapshot`
与 `/action` 调用，不影响任意 JavaScript 求值。

`GET /action` 会解码动作字段的一个子集，并以 `400`（命名该字段）拒绝任何它无法表达的参数，
而不是悄悄丢弃——因此修饰符组合键、拖拽、`waitNav` 或 `humanize` 必须以带 JSON 体的
`POST /action` 发送。动作请求根本没有声明的参数也以同样方式被拒绝，并对接近拼写给出
`did you mean` 提示，于是 `?modifers=8` 或 `?Modifiers=8` 不再派发一次普通点击并应答 200。
被接受的集合是动作请求自身的字段，加上只有 GET 形式才携带的参数——今天是 `timeout`。`timeout`
是一个逐请求的动作超时（秒），当其大于 0 且至多 60 时生效（任何其他值回退到配置的动作超时）。
POST 体没有等价字段，因此 POST 动作总是使用配置的动作超时。缓存破坏参数和游离参数必须从 URL
中去掉。

`GET /snapshot` 以同样方式校验其成本控制，但以不同方式处理未知参数问题，且这个差异是刻意的。
坏的**值**以 `400` 拒绝并命名可接受集合：`format` 为 `json`、`compact`、`text` 或
`yaml`；`filter` 为 `all` 或 `interactive`；`maxTokens` 为正整数；`depth` 为
`>= -1` 的整数。`format` 与 `filter` 比较时忽略大小写与空白，因此 `INTERACTIVE` 和
`" interactive "` 都会选中 interactive 子集——此前它们每个都会漏到整棵树，因为比较是精确
字符串匹配，而这些控制项每一个过去都会朝*更昂贵*的答案失败，却不告诉调用方。

`interactive` 是文档记载的 `filter` 的布尔别名：`interactive=true` 即 `filter=interactive`，
`interactive=false` 即 `filter=all`。它接受 Go 的 `ParseBool` 所接受的值
（`true`/`false`、`1`/`0`、`t`/`f`），任何其他值都是 `400`，而不是漏到整棵树。两者一致时
同时发送无妨；`filter=all` 旁跟着 `interactive=true` 会被拒绝，因为用一条调用方看不见的优先级
规则去裁决，意味着它发的两个参数之一没起作用。这个别名在这个端点上宣传了很久却从未被读取，于是
一个照着文档来的原始 HTTP 调用方买到了整棵树，却没被告知任何事。

未知的参数**名**会被报告而非拒绝：JSON 与 YAML 响应上的 `ignoredParams`，`compact` 与
`text` 响应上的一行 `# ignored params: ...`。`/action` 可以拒绝，因为它的参数集就是动作
请求自身的字段，而调用方发了别的东西，确实就是要求一个不会发生的行为。`/snapshot` 是一种读
取，较新的客户会带上较旧服务器还没学到的参数来调用它，因此拒绝会朝版本错位通常发生的方向打破
它，而披露仍然终结了沉默——这才是关键，因为 `quick` 命令行界面命令很长一段时间都发送
`compact=true`（一个 `/snapshot` 从未读取过的参数），结果收到的是 JSON 快照，而不是它写出来
想要请求的 compact 快照。

`GET /visible`（以及 `pinchtab visible <ref>`）回答的是 CSS 渲染状态——`display`、
`visibility`、`opacity`，以及一个有非零尺寸的已布局盒。滚动位置不是输入：一个在首屏下方很
远处、或已被滚过的元素仍报告 `visible: true`。「是否在屏上」是响应的 `onScreen` 字段，它与
capture 快照共享视口相交谓词（见 [reference/capture.md](reference/capture.md)）；当元素无法
被测量时 `onScreen` 被省略——缺席意味着未知，绝不是「否」。

当前动作类型包括：

- `click`
- `dblclick`
- `type`
- `fill`
- `press`
- `hover`
- `mouse-move`
- `mouse-down`
- `mouse-up`
- `mouse-wheel`
- `focus`
- `select`
- `scroll`
- `drag`
- `check`
- `uncheck`
- `keyboard-type`
- `keyboard-inserttext`
- `keydown`
- `keyup`
- `scrollintoview`

动作目标字段：

- `ref`
- `selector`
- `nodeId`
- `x` 和 `y`
- `button`
- `deltaX` 和 `deltaY`
- `waitNav`
- `dialogAction` 和 `dialogText`
- `humanize`
- `toSelector`、`toX`/`toY` 和 `dragX`/`dragY`（拖拽）
- `submit`、`mode` 和 `modifiers`（点击）

`fill` 与 `type` 把字符串写在 `text` 里；`fill` 也接受它作为 `value`，而 `select` 读的正是
这个字段。一个既不带 `text` 也不带 `value` 的 `fill` 会被拒绝——发送 `"text": ""` 来清空字段，
于是「清空」始终有别于「文本从未到达」的请求。

`select` 先匹配 `<option value="...">` 属性，其次匹配选项的可见文本，因此两种写法都行。它与
`fill` 一样区分「缺席」与「提供」：发送 `"value": ""` 来选中一个
`<option value="">` 占位项并重置下拉框，一个两个键都不带的 `select` 会被拒绝。每一个入口都
表达这一点——带 `"value": ""` 的 `POST /action`、`pinchtab select <ref> ""`，以及带
`value: ""` 的 `pinchtab_select` MCP 工具。

`button` 接受 `left`、`right`、`middle`——即命令行界面 `--button` 帮助里列出的同一套词汇，
容忍大小写和周围空白，因此 `RIGHT` 和 ` middle ` 就是它们所命名的按钮。任何其他值都以
`400 invalid_mouse_button` 拒绝并命名这三者，适用于任何携带该字段的动作体：`primary`、
`secondary` 和 `0` 过去会被重新解释为 `left` 并报告为成功。省略 `button` 意味着 `left`，这是
一个默认值，而不是对服务器不知道的名字的宽恕。

`humanize` 是一个逐动作的输入风格覆盖。省略时，动作使用 `instanceDefaults.humanize`，其默认
为 `false`。当一个页面需要更慢、更拟人化的指针或打字路径时，用 `kind:"click"` 或
`kind:"type"` 并配 `humanize:true`。

指针回退行为：

- `mouse-move` 先尝试一次真实的 CDP `mouseMoved` 事件。
- 如果无头 Chromium 在等待渲染器确认时让这次移动卡住，PinchTab 会在同一目标回退到 DOM 的
  `mouseover`/`mouseenter`/`mousemove` 事件，使悬停类检查仍保持响应。
- 非超时的 CDP 错误与调用方上下文取消不会被回退隐藏。
- `mouse-wheel` 在目标点派发一个 DOM `WheelEvent`，并在事件未被取消时滚动窗口。

选择器查找仅限于当前 frame 作用域。默认作用域是 `main`。在基于选择器的 iframe 动作之前使用
`/frame` 或 `/tabs/{id}/frame`。支持同源 iframe 作用域；当前不暴露跨源 iframe 后代。

快照查询参数：

- `filter`
- `interactive`（`filter` 的布尔别名）
- `diff`
- `selector`
- `maxTokens`
- `depth`
- `format`
- `noAnimations`
- `output`（`output=file` 时带 `path`）
- `tabId`

`/snapshot` 上的 `selector` 遵循同样规则：它只搜索当前 frame 作用域。它不会自动刺穿到
iframe 中，跨源 iframe 后代也不会被内联。

文本查询参数：

- `mode=raw`（`mode=full` 是别名）、`mode=markdown`（任何其他值都是 400 并命名可接受值）
- `format`
- `maxChars`
- `frameId`
- 读取单个元素用的 `selector` 或 `ref`
- `tabId`

`/text` 默认模式挑选第一个**可见**的 `<article>` / `[role="main"]` / `<main>`
（跳过 `display:none`），并剥离 nav/footer/广告。要完整 `innerText` 用 `mode=raw`，要价格、
按钮标签这类结构化 UI 文本用 `/snapshot`。

`mode=markdown` 通过 seaportal 转换器把渲染后的页面返回为 Markdown——即站点抓取器所应用的
同一转换——JSON 信封在 `text` 之外携带转换器的 `title` 和 `description`。它读取当前 frame
作用域的渲染后 HTML（即 `/html` 返回的文档），因此一个经 `/frame` 选中的 iframe 会转换那个
frame。`format=text` 以
`Content-Type: text/markdown; charset=utf-8` 返回原始 Markdown 正文，`maxChars` 保持整行、
对最后越界的一行按符文截断，绝不切分表格行（整行丢弃）或链接（截断会退回到该链接之前）。当转换
器什么都产出不了时，响应回退到原始页面文本并回显 `extraction: "markdown_fallback"`。

`mode=raw` 与 `mode=full` 是同一种提取——整个未过滤的页面——也就是命令行界面 `--raw` 与
`--full` 所发送的。默认提取保留块与表格单元格边界：相邻单元格以制表符分隔，相邻块以换行
分隔，于是相邻单元格里的一个状态码和一个时间戳仍保持为两个字段，而不是一个数字。

`/text` 也是 frame 感知的。`frameId` 针对特定 iframe 做一次性读取；否则该端点继承标签页当前的
`/frame` 作用域。

### 作用域读取上的 `frame` 披露

`/frame` 作用域是逐标签页的服务端状态，而非逐请求参数：它会在之后每条命令中存活，直到有东西
清除它，因此读取一个已作用域标签页的调用方，往往并不是当初作用域它的那个调用方。于是
`/snapshot` 与 `/text` 会披露它所服务的那个 frame：

```json
"frame": {
  "frameId": "886601397BFA0B332880152438BD0153",
  "frameUrl": "http://127.0.0.1:18798/inner.html",
  "frameName": "payment-frame",
  "frameTitle": "Inner",
  "ownerRef": "e3"
}
```

- 整文档读取时该键**缺席**，因此对未作用域的调用方什么都不变。
- 对一次性的 `?frameId=` 读取**也会**发布它——在一个没有存储作用域的标签页上：该披露命名的是
  这次读取实际所服务的 frame，而不是标签页恰好作用域到的那个。一次性读取返回一个片段，原因与
  作用域读取相同，因此它以同样方式说明这一点。
- `frameUrl` 与 `frameTitle` 在请求时从该 frame 读取，是返回内容所归属的对象；一个在设置作用
  域后又导航过的 frame，报告的是它现在所在的位置。
- 顶层 `url` 与 `title` 在每个响应中（无论是否作用域）都保持其含义：它们是该**标签页**的文档。
  它们绝不会被重新指向 frame——一个平时指一个意思、在不可见状态下又指另一个意思的字段，正是
  这份披露要消除的缺陷。
- `format=compact` 与 `format=text` 在头部以同一事实表达，形如
  `# Outer | http://127.0.0.1:18798/ | frame e3 | 3 nodes`。当 owner ref 已知时，该标记命名
  它，因为那正是 `POST /frame` 作为 `target` 接受的句柄；原始 frame id 则不是。没有已知 ref
  时，它命名一个缩短的 frame id。
- 该对象即 `GET /frame` 在 `frame` 下返回的那个，外加 `frameTitle`。

`/capture` 在作用域读取时也发布同一个 `frame` 对象，原因相同：它的快照一半被过滤到作用域
frame，而顶层 `url` 与 `title` 命名的是标签页文档。

`epoch.frameId` **不是**作用域，从来也不是。它是 frame 树的**根** id，在 capture 之前取得，
用于把图片与其拍摄时所对照的 DOM epoch 配对，且无论是否设置了作用域它都持有同一值——因此一个
作用域调用方读到它时，会被告知内容来自主文档。读 `frame.frameId` 得到作用域，读
`epoch.frameId` 得到 epoch；它们回答不同的问题，只在标签页未作用域时才一致。

`/html` 与 `/styles` 早已把它们的 frame 作为顶层 `frameId` 披露，且它们的 `url`、`title`
来自该 frame 自己的文档而非标签页的，因此那里的作用域读取从未被归到父级头上。

查找主体字段：

- `query`
- `tabId`
- `threshold`
- `topK`
- `lexicalWeight`
- `embeddingWeight`
- `explain`

## 截图、PDF 与 screencast

```text
GET  /screenshot
GET  /tabs/{id}/screenshot
GET  /annotate
GET  /tabs/{id}/annotate
GET  /capture
GET  /tabs/{id}/capture
GET  /pdf
POST /pdf
GET  /tabs/{id}/pdf
POST /tabs/{id}/pdf
GET  /screencast
GET  /screencast/tabs
GET  /instances/{id}/screencast
GET  /instances/{id}/proxy/screencast
POST /record/start
POST /record/stop
GET  /record/status
```

截图查询参数：

- `tabId`
- `format=jpeg|png`
- `quality`
- `raw=true`
- `output=file`
- `noAnimations=true`
- `selector`——捕获单个元素
- `annotate=true`——把编号的 ref 框烘焙进图片
- `beyondViewport=true`——捕获完整文档（与 `selector` 同用时被忽略）
- `scale=<float>`——重新缩放输出位图（例如 `0.5` = 一半大小，`0.25` = 四分之一）。默认 `1`。

`/annotate` 向活页面注入一个持久、可点击的标注覆盖层——每个可交互元素一个带标签的框——并把
它留在那里（`screenshot?annotate=true` 的覆盖层是瞬态的，反而会被烘焙进图片）。面向有头
浏览器：点击一个标签会把一个引用块（页面、ref、role、可访问名、CSS 选择器、XPath）复制到
剪贴板。`?clear=true` 移除它；`?selector=` 对它做作用域。参见
[Fix your website faster with an LLM](guides/annotate-for-llm-fixes.md)。

Annotate 查询参数：

- `tabId`
- `selector`——把覆盖层作用域到该选择器内的元素
- `clear=true`——移除覆盖层而非注入

`/capture` 在单次调用中从同一个 DOM epoch 返回一张截图和一份可访问性快照。它是「背靠背各发一次
`/screenshot` 和 `/snapshot`」的有视觉锚定的替代——这两次未配对调用不共享任何同步原语，因此
页面可能在它们之间发生变化，快照里的 ref 可能指向拍摄图片时并不存在的节点。

Capture 查询参数：

- `tabId`
- `selector`——把截图裁剪到该元素，并把快照子树过滤到同一元素
- `filter=interactive|all`
- `depth`——快照最大深度（默认 `-1` 表示全部）
- `format=jpeg|png`
- `quality`
- `output=file|inline|raw`——默认 `file`
- `requirePair=true`——在捕获窗口期间观察到导航时返回 `409 Conflict`
- `noAnimations=true`
- `scale=<float>`——经 CDP 的 `clip.scale` 重新缩放输出图片。默认 `1`（原生像素）。
  `scale=0.5` 把每个轴减半（四分之一像素）。响应里的 `image.devicePixelRatio` 告诉你你的原生
  DPR 是多少，于是需要时你可以算出 CSS 像素等价物。
- `wait=stable|load|none`——默认 `stable`。`stable` 在打开捕获窗口之前等待
  `Page.lifecycleEvent` 静默（250ms 无动静，上限 750ms），使截图与 AX 树描述的是一个已安定的
  页面。`none` 跳过等待。`load` 当前是 `none` 的别名；留给将来的 `document.readyState` 门控。
- `withBounds=true|false`——默认 `true`。开启时，每个带非零后端 node id 的快照节点都获得一个
  `boundingBox` 字段和一个 `visible` 标志。`boundingBox` 是元素的**边框盒（border box）**——
  绘出的边缘，即 `GET /box`、`screenshot?annotate=true` 和 `getBoundingClientRect` 所报告的
  同一矩形，于是一个盒可以与它们任意一个交叉核对。它不是内容盒：否则一个带边框或内边距的元素
  会报告一个从观看者识别控件所依据的边缘向内缩进的矩形。每个带界节点花费一次
  `DOM.getBoxModel` 往返（约 5ms）；对典型的 interactive-filter 快照，预算在 250ms 以内。传
  `withBounds=false` 跳过逐节点工作。
- `beyondViewport=true|false`——默认 `false`。开启时，图片横跨整个文档，而不只是可见视口。
  响应把 `image.coordinateSpace` 设为 `"document"`，且边界盒以页面（文档）坐标表达，使它们
  叠加在完整图片上。当同时给了 `selector` 时，selector 裁剪胜出，`beyondViewport` 被悄悄
  忽略——这正是 `/screenshot` 所强制的同一规则。超视口捕获会强制一次布局，可解析懒加载图片并
  触发 `IntersectionObserver`；AX 树获取与边界收集在截图之后运行，因此它们反映的是重排之后的
  状态。

响应携带 `image.coordinateSpace`、`image.devicePixelRatio` 和 `image.viewport`（捕获时刻的
`w`、`h`、`scrollX`、`scrollY`，CSS 像素），使客户无需猜测即可在图片像素与 `boundingBox`
值之间换算。要让这一点成立，必须钉住两个轴：坐标**原点**（由 `image.coordinateSpace` 命名，
`viewport` 或 `document`），以及盒模型**边缘**（永远是边框盒）。

表述为保证：**从 `image.coordinateSpace` 所命名的原点出发，把一个 `boundingBox` 乘以
`image.devicePixelRatio`，就落到那张图片自身的像素上——在每一种模式下，无论是否有模拟视口。**
图片测量的恰好是所报告空间乘以所报告比例，因此客户永远不需要按配置分支，也永远不必检测是否
恰好存在一个线性映射。默认（视口）捕获为了信守这一承诺会合成页面，这使它失去了 `/screenshot`
仍享有的那条更快的「读视口」路径：在一个空闲的有头浏览器上，`/capture` 因而可能阻塞到其截止
时间，而 `/screenshot` 立即返回。

响应携带一个 `epoch.domEpoch` token，缓存在标签页的 ref 缓存上。将来的客户工作可以把
`expectedEpoch` 传给动作端点，以在使用点检测过期 ref；在 P1 它只是信息性的。当主 frame 的
`loaderId` 在捕获中途变化时，`pairing.navigated` 为 `true`——这是 P1 检测的唯一漂移模式。
文档内变动（重渲染、observer 突变）是后续阶段要处理的残余风险。

响应形状：

```json
{
  "status": "ok",
  "tabId": "tab_abc",
  "url": "https://example.com",
  "title": "Example",
  "capturedAt": "2026-05-29T10:11:12.345Z",
  "epoch": { "frameId": "...", "loaderId": "...", "domEpoch": "ep_..." },
  "pairing": { "navigated": false, "captureDurationMs": 312 },
  "image": { "format": "jpeg", "path": "/.../captures/cap-...jpg", "bytes": 184223 },
  "snapshot": { "filter": "interactive", "nodeCount": 14, "nodes": [...] }
}
```

PDF 查询参数：

- `tabId`
- `raw=true`
- `output=file`
- `path`
- `landscape`
- `scale`
- `paperWidth`
- `paperHeight`
- `marginTop`
- `marginBottom`
- `marginLeft`
- `marginRight`
- `pageRanges`
- `preferCSSPageSize`
- `displayHeaderFooter`
- `headerTemplate`
- `footerTemplate`
- `generateTaggedPDF`
- `generateDocumentOutline`

Record 启动主体字段（JSON POST `/record/start`）：

- `format`：`gif`、`webm` 或 `mp4`。
- `fps`：每秒帧数，1-30（默认 5）。
- `quality`：JPEG 捕获质量 1-100（默认 80）。
- `scale`：分辨率倍率（默认 1.0）。
- `tabId`：针对特定标签页。

注意：

- 录制端点由 `security.allowScreencast` 门控。
- `.webm` 与 `.mp4` 格式要求服务器 PATH 上有 `ffmpeg`。
- `.gif` 格式使用纯 Go 编码（始终可用）。
- 每个桥接实例只允许一次录制。

## 站点审计与抓取

```text
POST /audit/page
POST /audit
POST /scrape
```

`POST /audit/page` 审计一个 `url`；`POST /audit` 接受 `urls`、一个 `sitemapUrl` 或 SeaPortal
结果；`POST /scrape` 接受爬取根 `url`。它们支撑 `pinchtab audit` 与 `pinchtab scrape`——请求体
与报告结构见 [audit.md](audit.md) 和 [scrape.md](scrape.md)。

## 下载、上传、Cookies 与剪贴板

```text
GET  /download
GET  /tabs/{id}/download
POST /upload
POST /tabs/{id}/upload
GET  /cookies
POST /cookies
DELETE /cookies
GET  /tabs/{id}/cookies
POST /tabs/{id}/cookies
DELETE /tabs/{id}/cookies
GET  /clipboard/read
POST /clipboard/write
POST /clipboard/copy
GET  /clipboard/paste
POST /cache/clear
GET  /cache/status
```

注意：

- 下载与上传端点由 `security.allowDownload` 和 `security.allowUpload` 门控
- cookie 端点（`GET/POST/DELETE /cookies`，加标签页作用域变体）由 `security.allowCookies` 门控
- 下载自动解压 `.gz` 文件并返回解压后内容
- `security.downloadAllowedDomains` 可以把特定域名加入白名单（对匹配域绕过 SSRF 检查）。设为
  `["*"]` 匹配每个主机并对该端点禁用私有 IP 保护，包括回环。命名一个回环主机（`127.0.0.1`、
  `localhost`）或使用 `"*"`，会让下载端点能触及服务器本机上的服务，包括 PinchTab 自己的本地
  端点。一个既不匹配也非公网的主机以 code `download_host_blocked` 被拒绝，且响应携带命名它的
  `config set` 行。
- 剪贴板端点由 `security.allowClipboard` 门控
- 上传使用一个带 `selector`（默认 `input[type=file]`）、`files`（base64）和/或 `paths`
  （已位于 `<stateDir>/uploads` 内的文件）以及可选 `fileNames` 的 JSON 体
- `fileNames` 与 `files` 按索引对齐，设置页面在 `file.name` 中看到的名字——务必传它，否则每次
  上传都以 `upload-<i>.bin` 抵达，而按 `accept=".csv"` 或 `file.name.endsWith(...)` 门控的表单
  会拒绝它。不给名字时，扩展名从内容嗅探，但嗅探无法识别文本格式（`.csv`、`.json`、`.txt`、
  `.md`、`.html`），因为它们没有魔数字节。提供的名字优先于嗅探类型，即使两者不一致——与浏览器
  所发送的一致。只使用文件名部分：任何目录部分都被丢弃。

## 存储

```text
GET    /storage
POST   /storage
DELETE /storage
GET    /tabs/{id}/storage
POST   /tabs/{id}/storage
DELETE /tabs/{id}/storage
```

存储仅为当前源（活动标签页）捕获。不支持多源存储。

所有存储路由由 `security.allowStateExport` 门控。

GET 查询参数：

- `type`——`local`、`session` 或空（两者）
- `key`——可选，要检索的特定键
- `tabId`——可选标签页标识符

POST 主体字段：

- `key`——必需
- `value`——必需
- `type`——`local` 或 `session`（必需）
- `tabId`——可选

DELETE 主体字段（主体本身可选）：

- `type`——`local`、`session` 或 `all`（默认 `all`，两个存储）
- `key`——可选（省略时清除整个存储）；与 `type: all` 同时给出会被拒绝
- `tabId`——可选

## 状态

```text
GET    /state
GET    /state/list
GET    /state/show
POST   /state/save
POST   /state/load
DELETE /state
POST   /state/clean
```

`GET /state` 返回当前标签页或显式 `tabId` 的当前完整浏览器状态，包括 cookies、当前源存储、
元数据和基本标签页信息。

`/state/save|load|list|show|delete|clean` 管理磁盘上持久保存的浏览器状态。

这与 `GET /tabs/{id}/state` 不同，后者返回用于就绪与阻断检查的、实时的标签页/页面运行时状态。

注意：

- 所有 state 与 storage 端点由 `security.allowStateExport` 门控：`/storage`、
  `/tabs/{id}/storage`、`GET /state`、`GET /state/list`、`GET /state/show`、
  `POST /state/save`、`POST /state/load`、`DELETE /state` 和 `POST /state/clean`
- `GET /state` 与 `GET /state/show` 中的 cookie 值另外还需要 `security.allowCookies`；仅有
  `allowStateExport` 时，这些响应返回 cookie 计数但扣留值。save/load 仍可用，因为它们在服务端
  移动 cookie 值，只返回计数。
- 状态文件以 `0600` 权限存于 `{stateDir}/sessions/`
- 通过 `security.stateEncryptionKey` 配置项提供可选的 AES-256-GCM 加密
- 存储仅为当前源（活动标签页）捕获

`GET /state` 查询参数：

- `tabId`——可选标签页标识符；省略时使用当前标签页

`POST /state/save` 主体字段：

- `name`——状态文件名
- `encrypt`——可选，加密该状态文件
- `tabId`——可选标签页标识符
- `metadata`——可选附加元数据

`POST /state/load` 主体字段：

- `name`——状态文件名（必需）
- `tabId`——可选标签页标识符

`DELETE /state` 查询参数：

- `name`——状态文件名（必需）

`POST /state/clean` 主体字段：

- `olderThanHours`——可选（默认：24）

## 标签页状态

```text
GET /tabs/{id}/state
```

为一个标签页返回轻量的实时标签页/页面运行时状态，包括加载状态、对话框存在与否、可操作性。

把它当作动作之前廉价的就绪探针。详细语义请保留在 API/skill 参考中，而非此处。

## 等待、网络、对话框、控制台与错误

```text
POST /wait
POST /tabs/{id}/wait
GET  /network
GET  /network/stream
GET  /network/export
GET  /network/export/stream
GET  /network/{requestId}
POST /network/clear
GET  /tabs/{id}/network
GET  /tabs/{id}/network/stream
GET  /tabs/{id}/network/export
GET  /tabs/{id}/network/export/stream
GET  /tabs/{id}/network/{requestId}
GET  /network/route
POST /network/route
DELETE /network/route
GET  /tabs/{id}/network/route
POST /tabs/{id}/network/route
DELETE /tabs/{id}/network/route
POST /dialog
POST /tabs/{id}/dialog
GET  /console
POST /console/clear
GET  /errors
POST /errors/clear
```

等待主体字段：

- 以下恰好其一：
  - `selector`——CSS / XPath（`xpath:` 前缀或前导 `//`）/ 文本（`text:` 前缀）
  - `text`——`document.body.innerText` 的子串
  - `notText`——等待该子串不再出现
  - `url`——对 `window.location.href` 匹配的 glob 模式（`**`、`*`、`?`）
  - `load`——以下之一：
    - `ready-state` → `document.readyState === 'complete'`
    - `content-loaded` → `document.readyState` 属于 {`interactive`, `complete`}
    - `network-idle` → 零在途 CDP 请求保持 `idleFor` 毫秒（默认 500，最大 10000）。接受遗留别名
      `networkidle`。
  - `fn`——被轮询直到为真的 JS 表达式（需要 `security.allowEvaluate`）
  - `ms`——固定睡眠毫秒数，最大 30000（逃生口；优先条件式等待）
- 可选 `tabId`
- 可选 `timeout`——毫秒，默认 10000，钳制到 100–30000
- 选择器等待的可选 `state`——`visible`（默认）或 `hidden`
- `load: network-idle` 的可选 `idleFor`——毫秒静默期，默认 500，钳制到 0–10000

网络查询参数：

- `tabId`
- `filter`
- `method`
- `status`
- `type`
- `limit`
- `bufferSize`
- `broken=true`——以失效资源列表（`broken`、`count`）应答，而非条目
- 详细请求上的 `body=true`
- 详细请求上的 `bodyMode=auto|retained-preferred|retained-only|live-only`，选择如何解析响应体
- 详细请求上的 `timeoutMs`，用于钳制保留体等待窗口（默认 2000，最大 30000）

网络详细/导出的响应体行为：

- 默认情况下，响应体按需从实时 CDP 状态获取，对较旧的请求可能已不可用
- 当 `server.retainNetworkBodies=true` 时，PinchTab 机会性地把有界响应体保留在内存网络缓冲区中，
  并优先返回保留体
- `bodyMode=retained-preferred` 在回退到实时 CDP 之前短暂等待挂起的保留体捕获
- `bodyMode=retained-only` 绝不回退到实时 CDP，而是返回显式的 pending/skipped/error 状态
- 详细响应可能暴露 `bodySource=retained|live`，以区分哪条路径产生了所返回的体
- 保留体详细响应在捕获仍在进行时可能暴露 `bodyPending=true`；当保留未完成时可能暴露
  `bodySkipped=true` 并带 `bodySkipReason`——或是一开始就被跳过（保留被禁用、该标签页保留预算
  耗尽、达到并发上限），或是因为一个超预算的 base64 体被整体丢弃而非截断
- 保留体有双重上限：逐体受 `server.retainNetworkBodyMaxBytes` 限制（`bodySkipReason` 会说
  "retention limit"），并受该标签页剩余保留缓冲限制（"retention budget"）。一个超尺寸文本体被
  截断到字节精确的前缀并标记 `bodyTruncated=true`；一个超尺寸 base64 体被整体丢弃，带
  `bodySkipped=true` 及原因，因为一个 base64 片段是不可解码的
- `base64Encoded=true` 把所返回体（保留或实时）标记为 base64——使用前请解码。对文本体该字段被
  省略、绝不设为 `false`，因此其存在性才是要据以分支的东西。两个上限都度量编码后长度，因此一个
  二进制响应的有效原始预算约为配置字节数的四分之三
- 保留响应可能包含 `bodyRetained=true`

请求体（`postData`）行为：

- `postData` 以页面发送时的形式、解码后持有请求体。Chrome 以 base64 编码并分块交付它；
  PinchTab 解码并拼接，因此调用方无需 base64 解码
- 它被限制为 64 KiB 解码后体，在字符边界处截断，被截断的体标记 `postDataTruncated=true`
  ——否则一个被裁的请求体会被读成客户端发送的体
- 当体不是文本时它被省略——例如 multipart 上传中的二进制部分——因为该字段不带编码标记。省略的
  体会说明原因：`postDataSkipped=true` 带 `postDataSkipReason`
  （"request body entry is not base64"、"request body is not valid UTF-8"），因此缺席的
  `postData` 绝不会被误当作「一个没发体的请求」
- `postDataTruncated` 与 `postDataSkipped` 是不同的答案，绝不会同时设置：truncated 表示被截断
  但可用，skipped 表示没有体可读。一个根本就没体的请求两者都不带
- HAR 导出把同一个解码值放进 `request.postData.text`，当没有可发布的体时整块省略

网络导出查询参数：

- `format`——`har`（默认）或 `ndjson`。可插拔：新格式在启动时注册。
- `output=file`——保存到磁盘而非流式输出到响应
- `path`——`output=file` 时的文件名（省略则自动生成；`/export/stream` 必需）
- `body=true`——包含响应体（默认按需获取；保留体模式可使有界条目持久化）
- `redact`——`true`（默认）对 Cookie/Authorization/Set-Cookie 做脱敏。`false` 导出原始头。
- 所有标准网络过滤器（`filter`、`method`、`status`、`type`、`limit`）

`/export` 端点以单个响应返回完整捕获。`/export/stream` 端点在条目到达时把它们写入文件（向调用
方发送 SSE 进度事件）。流式文件在完成时被原子重命名。

拦截规则（`/network/route`，由 `security.allowNetworkIntercept` 门控）：`POST` 接受
`pattern`（子串或 `*`/`?` glob）、`action`（`continue`、`abort` 或 `fulfill`），对 fulfill
还接受 `body`、`contentType`、`status`，外加可选 `resourceType` 与 `method`。`DELETE` 在
查询或体中接受 `pattern`，省略它时移除全部规则；`GET` 列出该标签页的规则。

对话框主体字段：

- `action`：`accept` 或 `dismiss`
- `text`：可选提示文本
- `tabId`：`/dialog` 上可选

控制台与错误路由使用查询参数：

- `tabId`
- `limit`

## 挑战解决器

```text
GET  /solvers
GET  /config/autosolver
POST /solve
POST /solve/{name}
POST /tabs/{id}/solve
POST /tabs/{id}/solve/{name}
```

自动解决器框架自动检测并解决浏览器挑战（Cloudflare Turnstile、CAPTCHA、插页等）。详见
[Solve reference](./reference/solve.md)。

Solve 主体字段：

- `solver` 可选解决器名（省略时自动检测）
- `tabId` 可选
- `maxAttempts` 可选（默认取 `autoSolver.maxAttempts`，默认 `8`）
- `timeout` 可选，毫秒（省略时自动估计，最小 `30000`）

一个被命名却无法运行的 `solver`，会在解决任何东西之前以 `400` 被拒绝，且两种原因携带不同
code：

| Code | 含义 | 示例消息 |
| --- | --- | --- |
| `unknown_solver` | 没有解决器应答这个名字——通常是拼写错误。列出可用者。 | `unknown solver "cloudlfare" (available: [cloudflare semantic jschallenge])` |
| `solver_key_missing` | 一个已知的、由 key 门控的解决器，其 API key 未设置。命名要设置的配置键。 | `solver "capsolver" is configured but its API key is not set; set autoSolver.external.capsolverKey to use it` |

这里的 API 刻意比配置校验更严格——后者在 `autoSolver.solvers` 中接受未设 key 的
`capsolver` 或 `twocaptcha`，而不报错——在 key 就位之前配置一个付费解决器是合法的先后顺序，
本次运行会回退到能运行的解决器。一个命名了某个解决器的请求没有这种回退：它绝不能悄悄运行另一个
解决器，因此它被拒绝，并被告知哪个 key 会启用它。

`GET /config/autosolver` 返回生效的自动解决器运行时设置与当前可用的解决器列表。

示例响应：

```json
{
	"enabled": true,
	"autoTrigger": true,
	"triggerOnNavigate": true,
	"triggerOnAction": true,
	"maxAttempts": 8,
	"solverTimeoutSec": 30,
	"retryBaseDelayMs": 500,
	"retryMaxDelayMs": 10000,
	"solvers": ["cloudflare", "semantic", "jschallenge"],
	"llmProvider": "",
	"llmFallback": false
}
```

注意：

- `capsolver` 与 `twocaptcha` 只在其 API key 已配置时才出现在 `solvers` 中。

## 配置文件与实例

```text
GET  /profiles
POST /profiles
POST /profiles/create
GET  /profiles/{id}
PATCH /profiles/{id}
DELETE /profiles/{id}
POST /profiles/{id}/start
POST /profiles/{id}/stop
GET  /profiles/{id}/instance
POST /profiles/{id}/reset
GET  /profiles/{id}/logs
GET  /profiles/{id}/analytics
POST /profiles/import
POST /profiles/prune
PATCH /profiles/meta
GET  /instances
GET  /instances/{id}
GET  /instances/tabs
GET  /instances/metrics
GET  /instances/{id}/metrics
POST /instances/start
POST /instances/launch
POST /instances/attach
POST /instances/attach-bridge
POST /instances/{id}/start
POST /instances/{id}/restart
POST /instances/{id}/stop
GET  /instances/{id}/logs
GET  /instances/{id}/logs/stream
GET  /instances/{id}/tabs
POST /instances/{id}/tabs/open
POST /instances/{id}/tab
POST /instances/{id}/close
POST /instances/{id}/cookies
POST /instances/{id}/audit
POST /instances/{id}/scrape
POST /instances/{id}/cache/clear
GET  /instances/{id}/cache/status
```

注意：

- `/instances/start` 与 `/instances/launch` 用 `profileId` 指定既有配置文件 ID 或名字，用
  `mode`，而非 `headless`。请求体以 400 拒绝无法识别的字段，并命名问题键与可接受形状。
- `/instances/launch` 是 `/instances/start` 的兄弟端点（独立处理器
  `handleLaunchByName`），为「按名字启动」工作流保留；体上的 `name` 不再被支持，配置文件必须
  已存在
- 实例响应同时包含 `mode` 和 `headless`
- 实例启动表面接受 `securityPolicy.allowedDomains`，用于附加的实例作用域 IDPI/域允许列表覆盖
- 用 `POST /profiles` 显式创建配置文件；`/instances/launch` 上不再支持 `name`
- `/profiles/{id}/start` 使用 `headless`
- attach 路由由 `security.attach` 门控
- `POST /profiles/prune` 移除被隔离的配置文件目录；见
  [commands.md](commands.md) 的 `pinchtab profiles prune` 一节
- `/instances/{id}/close|cookies|audit|scrape|cache/*` 把同一代理路由到该实例

## 活动与调度器

```text
GET  /api/activity
POST /tasks
GET  /tasks
GET  /tasks/{id}
POST /tasks/{id}/cancel
POST /tasks/batch
GET  /scheduler/stats
```

活动查询参数包括：

- `limit`
- `ageSec`
- `since`
- `until`
- `source`
- `requestId`
- `sessionId`
- `agentId`
- `instanceId`
- `profileId`
- `profileName`
- `tabId`
- `action`
- `pathPrefix`

活动归因与源行为：

- 带 `X-Agent-Id` 标记的请求被记录为 `agentId`，可用 `GET /api/activity?agentId=<id>` 过滤
- 未过滤的 `GET /api/activity` 返回主活动 feed
- 命名的非客户端源（如 `dashboard` 或 `orchestrator`）仅在 `observability.activity.events`
  下启用时，才存入按源划分的每日文件，此后可用 `?source=<name>` 查询

调度器路由仅在 `scheduler.enabled` 为 true 时存在。

## 代理会话

| 方法 | 路径 | 描述 |
|--------|------|-------------|
| `POST` | `/sessions` | 创建新的代理会话（体：`{agentId, label?, grants?, browser?}`） |
| `GET` | `/sessions` | 列出所有代理会话 |
| `GET` | `/sessions/me` | 获取当前会话（需要 `Authorization: Session` 认证） |
| `GET` | `/sessions/{id}` | 按 ID 获取会话详情 |
| `POST` | `/sessions/{id}/revoke` | 撤销会话 |

`POST /sessions`、`GET /sessions` 和 `GET /sessions/{id}` 需要仪表板认证（bearer 或
cookie）。`/me` 端点需要会话认证。`POST /sessions/{id}/revoke` 允许仪表板认证或所属会话。

创建返回 `sessionToken`——只显示一次的明文 token。

代理会话路由仅在完整服务器模式、且代理会话开启时存在——`sessions.agent.enabled` 为 true
且 `sessions.agent.mode` 不为 `off`。该家族总会应答，因此其状态可从错误码读出，而非从一个干
巴巴的 404：桥接返回 `sessions_unavailable_bridge_mode`，其补救是运行 `pinchtab server`；
一个把它们关掉的完整服务器返回 `sessions_disabled`。没有任何配置值会在桥接模式下挂载这个家族。

`sessions_disabled` 覆盖两种状态，其 `details.hint` 说明是哪一种。一个**启动时**就把代理会话
关掉的服务器从未挂载这个家族，因此开启它们需要一次配置编辑*和*一次重启，且该拒绝不带
`details.remedy`，因为那不是一条命令。一个启动时开着、后被一次配置保存关掉的服务器已经挂载了
这个家族：编辑实时生效、无需重启，并作为该拒绝的 `details.remedy` 携带。

有两个设置会关掉代理会话，因此提示会命名真正关掉的那一个——`sessions.agent.enabled` 为
false、`sessions.agent.mode` 为 `off`，或两者皆是（此时补救同时设置两者）。照所开的命令做，会在
一个被保存关掉的服务器上恢复服务；不存在「照做之后仍停在同一拒绝」的状态。

会话认证的调用方无法触及仪表板/管理员端点家族，如 config、仪表板代理列表、仪表板事件流、会话
管理、配置文件管理、实例管理或缓存控制。它们旨在用于受控环境中的受信任自动化，而非不受信任的多
租户隔离。

## 功能门

有些端点刻意被禁用，除非匹配的配置允许它们：

这些门不是普通的功能开关。启用它们是一个有文档记载的、非默认的、降低安全性的选择，它扩大了调用
方可用的控制范围。

- `/evaluate` 和 `/tabs/{id}/evaluate` -> `security.allowEvaluate`
- `/macro` -> `security.allowMacro`
- `GET /network/{requestId}`、`POST /network/clear` 和 `/network/route` 家族（加标签页作用域
  变体）-> `security.allowNetworkIntercept`
- `/memory/snapshot`、`/memory/snapshot/{snapshotId}/summary` 和 `/memory/compare` ->
  `security.allowMemory`
- `/download` 和 `/tabs/{id}/download` -> `security.allowDownload`
- `GET/POST/DELETE /cookies` 和 `GET/POST/DELETE /tabs/{id}/cookies` ->
  `security.allowCookies`
- `/upload` 和 `/tabs/{id}/upload` -> `security.allowUpload`
- 剪贴板路由 -> `security.allowClipboard`
- attach 路由 -> `security.attach`
- screencast 路由 -> `security.allowScreencast`
- 存储路由（`/storage`、`/tabs/{id}/storage`）和完整的状态管理家族
  （`GET /state`、`/state/list`、`/state/show`、`/state/save`、`/state/load`、
  `DELETE /state`、`POST /state/clean`）-> `security.allowStateExport`

## 错误响应格式

PinchTab 在过渡期间目前使用两种 JSON 错误形状：

- 遗留 JSON 错误：`application/json`，带 `error`、`code` 等字段
- Problem Details 错误：`application/problem+json`（RFC 7807 风格）

Problem Details 目前用于选定的前置条件与能力失败，包括：

- websocket 代理预升级后端/劫持失败
- 网络流不支持的流能力
- 仪表板 SSE 不支持的流能力或截止时间控制
- 实例日志 SSE 不支持的流能力或截止时间控制

随着时间推移，可能会迁移更多端点。客户端应容忍两种错误内容类型，并在解析失败时按
`Content-Type` 分支。

### 拒绝提示：`details.hint` 与 `details.remedy`

一个调用方可据以行动的拒绝会携带一个带两个字段的 `details` 对象。它们是两类不同的答案，谁也不
替代谁：

- `hint`——供人或模型阅读的文字。解释、替代方案、前置条件，以及任何不是单条命令的东西都放在
  这里。
- `remedy`——**一行 shell 能接受的东西**，于是代理无需解析英文即可逐字运行。

`remedy` 保证以下全部：

- 一行，一条或多条 `pinchtab` 调用，需要多条时用 `&&` 连接。允许 `$(...)` 命令替换，并在需要
  先读回一个值的地方使用——例如放宽域允许列表是追加到当前列表而非替换它
- 没有文字连接词（`then:`、`or`、括号尾巴），没有管道、分号、重定向、反引号、注释或花括号
  展开。`pinchtab dialog accept|dismiss` 对 shell 来说不是两个建议——它是一条进入名为
  `dismiss` 的命令的管道——因此那样的一行不是 remedy，也绝不会出现在该字段中
- 其中每个命令与 flag 都在命令行界面中存在
- 一个空位是一个 `<name>` 占位符，沿用命令行界面自身 `--help` 所用的同一尖括号约定，别无其他。
  拒绝产生时已知的值已被插值，因此一个占位符意味着该值确实是调用方要提供的
- **当没有单条命令能修复该拒绝时，该字段缺席。** 缺席即答案，而非遗漏：它如实说明没有可运行
  的东西，那种情况的提示在 `hint` 中。不要把缺失的 `remedy` 当作响应中的错误

```json
{
  "error": "this endpoint requires the evaluate capability; enable security.allowEvaluate in config to use it",
  "code": "evaluate_disabled",
  "details": {
    "setting": "security.allowEvaluate",
    "hint": "Enable security.allowEvaluate to use this feature, then restart PinchTab to apply the change.",
    "remedy": "pinchtab config set security.allowEvaluate true"
  }
}
```

`details` 在这两个字段之外可能还携带更多机器可读字段——上面的能力拒绝命名了 `setting`，一个
允许列表块命名了被阻断的 `url` 与 `domain`——因此按键读取该对象，而不要假设它只装着提示。

`pinchtab` 在请求失败时会渲染这两个字段，把 `remedy` 打印进一行 `Remedy:`。该空位中的每个
值都满足上述契约，无论它来自服务器还是命令行界面自身的客户端拒绝。
