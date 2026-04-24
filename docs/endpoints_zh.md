# 端点参考

本页总结了 PinchTab 暴露的实时 HTTP 表面。一些路由仅在桥接模式下可用，一些仅在完整服务器模式下可用，一些由安全设置控制。

## 健康和服务器元数据

```text
GET  /health
POST /ensure-chrome
POST /browser/restart
GET  /openapi.json
GET  /help          (别名 for /openapi.json)
GET  /metrics
GET  /api/metrics
POST /shutdown
GET  /api/events
```

注意：

- 在桥接模式下，`/health` 报告桥接健康和标签页计数
- 在完整服务器模式下，`/health` 报告仪表板健康、认证状态和实例计数
- `/metrics` 代理到桥接实例（每个实例的运行时指标）
- 完整服务器模式下的 `/api/metrics` 是服务器级指标快照（聚合）

## 仪表板认证和配置

```text
POST /api/auth/login
POST /api/auth/elevate
POST /api/auth/logout
GET  /api/config
PUT  /api/config
```

注意：

- `server.token` 被 `PUT /api/config` 视为只写
- 认证路由用于仪表板会话流

## 仪表板事件和代理

```text
GET  /api/events
GET  /api/agents
GET  /api/agents/{id}
GET  /api/agents/{id}/events
POST /api/agents/{id}/events
```

注意：

- `/api/events` 是仪表板 SSE 流
- `/api/agents/{id}/events` 流式传输一个代理的最近事件
- `POST /api/agents/{id}/events` 将代理活动摄取到仪表板 feed

## 导航和标签页

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
POST /tabs/{id}/close
GET  /tabs/{id}/metrics
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

重要行为：

- 当省略 `tabId` 时，`POST /navigate` 创建新标签页
- `POST /tab` 支持 `new`、`close` 和 `focus`

## 切换和手动干预

```text
POST /tabs/{id}/handoff
GET  /tabs/{id}/handoff
POST /tabs/{id}/resume
```

注意：

- 这些路由仅在标签页范围内
- `POST /tabs/{id}/handoff` 将标签页标记为 `paused_handoff` 并记录原因
- `GET /tabs/{id}/handoff` 返回当前切换状态，或当未设置切换时返回 `active`（当配置超时时间时包括 `expiresAt` 和 `timeoutMs`）
- `POST /tabs/{id}/resume` 清除切换状态，并可以为调用者携带恢复元数据
- 动作执行路由（`/action`、`/actions`、`/macro`）拒绝暂停的标签页，返回 `409 tab_paused_handoff`
- 可用的 CLI 包装器：`pinchtab tab handoff`、`pinchtab tab handoff-status` 和 `pinchtab tab resume`

## 标签页锁定

```text
POST /lock
POST /unlock
POST /tabs/{id}/lock
POST /tabs/{id}/unlock
```

## 交互和分析

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
POST /find
POST /tabs/{id}/find
POST /evaluate
POST /tabs/{id}/evaluate
```

`/evaluate` 故意与选择器框架范围分开。`GET/POST /frame` 仅影响基于选择器的 `/snapshot` 和 `/action` 调用，不影响任意 JavaScript 评估。

当前的动作类型包括：

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

选择器查找仅限于当前框架范围。默认范围是 `main`。在基于选择器的 iframe 动作之前使用 `/frame` 或 `/tabs/{id}/frame`。支持同源 iframe 范围；当前不暴露跨域 iframe 后代。

快照查询参数：

- `interactive`
- `compact`
- `diff`
- `selector`
- `maxTokens`
- `depth`
- `format`
- `noAnimations`
- `output`

`/snapshot` 上的 `selector` 遵循相同的规则：它仅搜索当前框架范围。它不会自动刺穿到 iframes 中，并且跨域 iframe 后代不会内联。

文本查询参数：

- `mode=raw`
- `format`
- `maxChars`
- `frameId`

`/text` 默认模式选择第一个 **可见** `<article>` / `[role="main"]` / `<main>`（跳过 `display:none`）并剥离导航/页脚/广告。使用 `mode=raw` 获得完整的 `innerText`，或使用 `/snapshot` 获得结构化 UI 文本，如价格和按钮标签。

`/text` 也是框架感知的。`frameId` 针对特定 iframe 进行一次性读取；否则，端点继承标签页的当前 `/frame` 范围。

查找主体字段：

- `query`
- `tabId`
- `threshold`
- `topK`
- `lexicalWeight`
- `embeddingWeight`
- `explain`

## 截图、PDF 和屏幕录制

```text
GET  /screenshot
GET  /tabs/{id}/screenshot
GET  /pdf
POST /pdf
GET  /tabs/{id}/pdf
POST /tabs/{id}/pdf
GET  /screencast
GET  /screencast/tabs
GET  /instances/{id}/screencast
GET  /instances/{id}/proxy/screencast
```

截图查询参数：

- `tabId`
- `format=jpeg|png`
- `quality`
- `raw=true`
- `output=file`
- `noAnimations=true`

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

## 下载、上传、Cookies 和剪贴板

```text
GET  /download
GET  /tabs/{id}/download
POST /upload
POST /tabs/{id}/upload
GET  /cookies
POST /cookies
GET  /tabs/{id}/cookies
POST /tabs/{id}/cookies
GET  /clipboard/read
POST /clipboard/write
POST /clipboard/copy
GET  /clipboard/paste
POST /cache/clear
GET  /cache/status
```

注意：

- 下载和上传端点由 `security.allowDownload` 和 `security.allowUpload` 控制
- 下载自动解压缩 `.gz` 文件并返回解压缩的内容
- `security.downloadAllowedDomains` 可以白名单特定域（对这些域绕过 SSRF 检查）。设置 `["*"]` 匹配每个主机并禁用下载端点上的所有私有 IP 保护。
- 剪贴板端点由 `security.allowClipboard` 控制
- 上传使用带有 `selector` 和 `files` 的 JSON 主体

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

所有存储路由由 `security.allowStateExport` 控制。

GET 查询参数：

- `type` — `local`、`session` 或空（两者）
- `key` — 可选，要检索的特定键
- `tabId` — 可选标签页标识符

POST 主体字段：

- `key` — 必需
- `value` — 必需
- `type` — `local` 或 `session`（必需）
- `tabId` — 可选

DELETE 主体字段：

- `type` — `local` 或 `session`（必需）
- `key` — 可选（如果省略，清除整个存储）
- `tabId` — 可选

## 状态管理

```text
GET    /state/list
GET    /state/show
POST   /state/save
POST   /state/load
DELETE /state
POST   /state/clean
```

状态管理将浏览器状态（cookies、localStorage、sessionStorage、元数据）保存到磁盘并从磁盘恢复。

注意：

- 所有状态和存储端点由 `security.allowStateExport` 控制：`/storage`、`/tabs/{id}/storage`、`GET /state/list`、`GET /state/show`、`POST /state/save`、`POST /state/load`、`DELETE /state` 和 `POST /state/clean`
- 状态文件存储在 `{stateDir}/sessions/` 中，权限为 `0600`
- 通过 `security.stateEncryptionKey` 配置设置可选的 AES-256-GCM 加密
- 存储仅为当前源（活动标签页）捕获

`POST /state/save` 主体字段：

- `name` — 状态文件名
- `encrypt` — 可选，加密状态文件
- `tabId` — 可选标签页标识符
- `metadata` — 可选附加元数据

`POST /state/load` 主体字段：

- `name` — 状态文件名（必需）
- `tabId` — 可选标签页标识符

`DELETE /state` 查询参数：

- `name` — 状态文件名（必需）

`POST /state/clean` 主体字段：

- `olderThanHours` — 可选（默认：24）

## 等待、网络、对话框、控制台和错误

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
POST /dialog
POST /tabs/{id}/dialog
GET  /console
POST /console/clear
GET  /errors
POST /errors/clear
```

等待主体字段：

- `selector`、`text`、`url`、`load`、`fn` 或 `ms` 之一
- 可选 `tabId`
- 可选 `timeout`
- 选择器等待的可选 `state`

网络查询参数：

- `tabId`
- `filter`
- `method`
- `status`
- `type`
- `limit`
- `bufferSize`
- 详细请求的 `body=true`

网络导出查询参数：

- `format` — `har`（默认）或 `ndjson`。可插拔：新格式在启动时注册。
- `output=file` — 保存到磁盘而不是流式传输到响应
- `path` — 当 `output=file` 时的文件名（如果省略则自动生成，`/export/stream` 必需）
- `body=true` — 包含响应主体（按需获取，每个条目 10 MB 上限）
- `redact` — `true`（默认）编辑 Cookie/Authorization/Set-Cookie。`false` 导出原始标头。
- 所有标准网络过滤器（`filter`、`method`、`status`、`type`、`limit`）

`/export` 端点以单个响应返回完整捕获。`/export/stream` 端点在条目到达时将其写入文件（向调用者发送 SSE 进度事件）。流式文件在完成时被原子重命名。

对话框主体字段：

- `action`：`accept` 或 `dismiss`
- `text`：可选提示文本
- `tabId`：`/dialog` 上的可选

控制台和错误路由使用查询参数：

- `tabId`
- `limit`

## 挑战解决器

```text
GET  /solvers
POST /solve
POST /solve/{name}
POST /tabs/{id}/solve
POST /tabs/{id}/solve/{name}
```

解决器框架自动检测并解决浏览器挑战（Cloudflare Turnstile 等）。有关详细信息，请参阅 [Solve 参考](./reference/solve.md)。

解决主体字段：

- `solver` 可选解决器名称（省略时自动检测）
- `tabId` 可选
- `maxAttempts` 可选（默认：3）
- `timeout` 可选，以毫秒为单位（默认：30000）

## 配置文件和实例

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
PATCH /profiles/meta
GET  /instances
GET  /instances/{id}
GET  /instances/tabs
GET  /instances/metrics
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
```

注意：

- `/instances/start` 和 `/instances/launch` 使用 `mode`，而不是 `headless`
- `/instances/launch` 是 `/instances/start` 上的兼容性别名
- 实例响应包括 `mode` 和 `headless`
- 实例启动表面接受 `securityPolicy.allowedDomains`，用于附加的实例范围 IDPI/域允许列表覆盖
- 使用 `POST /profiles` 显式创建配置文件；`name` 不再支持于 `/instances/launch`
- `/profiles/{id}/start` 使用 `headless`
- 附加路由由 `security.attach` 控制

## 活动和调度器

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
- `engine`
- `pathPrefix`

活动归因和源行为：

- 标记有 `X-Agent-Id` 的请求被记录为 `agentId`，可以使用 `GET /api/activity?agentId=<id>` 过滤
- 未过滤的 `GET /api/activity` 返回主要活动 feed
- 命名的非客户端源（如 `dashboard` 或 `orchestrator`）仅在 `observability.activity.events` 下启用时存储在特定于源的每日文件中，然后可以使用 `?source=<name>` 查询

调度器路由仅在 `scheduler.enabled` 为 true 时存在。

## 代理会话

| 方法 | 路径 | 描述 |
|--------|------|-------------|
| `POST` | `/sessions` | 创建新的代理会话（主体：`{agentId, label?}`） |
| `GET` | `/sessions` | 列出所有代理会话 |
| `GET` | `/sessions/me` | 获取当前会话（需要 `Authorization: Session` 认证） |
| `GET` | `/sessions/{id}` | 通过 ID 获取会话详细信息 |
| `POST` | `/sessions/{id}/revoke` | 撤销会话 |

`POST /sessions`、`GET /sessions` 和 `GET /sessions/{id}` 需要仪表板认证（bearer 或 cookie）。`/me` 端点需要会话认证。`POST /sessions/{id}/revoke` 允许仪表板认证或拥有的会话。

创建返回 `sessionToken` — 只显示一次的明文令牌。

会话认证的调用者不能访问仪表板/管理员端点系列，如配置、仪表板代理列表、仪表板事件流、会话管理、配置文件管理、实例管理或缓存控制。它们旨在用于受控环境中的受信任自动化，而不是用于不受信任的多租户隔离。

## 功能门

一些端点有意禁用，除非匹配的配置允许它们：

这些门不是普通的功能开关。启用它们是一个有文档记录的、非默认的、降低安全性的选择，它扩大了调用者可用的控制表面。

- `/evaluate` 和 `/tabs/{id}/evaluate` -> `security.allowEvaluate`
- `/download` 和 `/tabs/{id}/download` -> `security.allowDownload`
- `/upload` 和 `/tabs/{id}/upload` -> `security.allowUpload`
- 剪贴板路由 -> `security.allowClipboard`
- 附加路由 -> `security.attach`
- 屏幕录制路由 -> `security.allowScreencast`
- 存储路由（`/storage`、`/tabs/{id}/storage`）和完整的状态管理系列（`/state/list`、`/state/show`、`/state/save`、`/state/load`、`DELETE /state`、`POST /state/clean`）-> `security.allowStateExport`

## 错误响应格式

PinchTab 目前在过渡期间使用两种 JSON 错误形状：

- 遗留 JSON 错误：`application/json`，带有 `error` 和 `code` 等字段
- 问题详细信息错误：`application/problem+json`（RFC 7807 风格）

问题详细信息目前用于选定的前置条件和能力失败，包括：

- websocket 代理预升级后端/劫持失败
- 网络流不支持的流能力
- 仪表板 SSE 不支持的流能力或截止时间控制
- 实例日志 SSE 不支持的流能力或截止时间控制
- 屏幕录制标签页未找到前置条件失败

随着时间的推移，可能会迁移其他端点。客户端应容忍两种错误内容类型，并在解析失败时按 `Content-Type` 分支。