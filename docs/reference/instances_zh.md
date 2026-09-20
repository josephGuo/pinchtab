# 实例（Instances）

实例是 PinchTab 管理的运行中 Chrome 进程。每个受管实例都有：

- 实例 ID
- 一个 profile
- 一个端口
- 一个模式（`headless` 或 `headed`）
- 一个执行状态

一个 profile 同一时间最多只能有一个活动的受管实例。

## 列出实例

```bash
curl http://localhost:9867/instances
# Response: JSON array (see below)

# CLI Alternative (human-readable by default)
pinchtab instance list
# Output (tab-separated): inst_0a89a5bb  9999  headed  running

pinchtab instance list --json          # Full JSON response
```

`pinchtab instance list` 是从命令行界面查看当前实例群最简单的方式。旧的 `pinchtab instances` 拼写仍作为已弃用别名运行。

响应形状：

```json
[
  {
    "id": "inst_0a89a5bb",
    "profileId": "prof_278be873",
    "profileName": "instance-1741410000000000000-3fa9c2d1",
    "port": "9999",
    "mode": "headed",
    "headless": false,
    "status": "running",
    "startTime": "2026-03-08T05:00:00Z",
    "attached": false,
    "responsiveness": "responsive",
    "securityPolicy": {
      "allowedDomains": ["127.0.0.1", "localhost", "::1", "wikipedia.org"]
    }
  }
]
```

`GET /instances` 返回裸 JSON 数组，而非 `{"instances":[...]}` 这样的封装。每个实例响应同时包含 `mode`（`"headless"` 或 `"headed"`）和兼容旧版的 `headless` 布尔值。可选字段在适用时出现：`url`（bridge 支撑的实例）、`error`（`status` 为 `error` 时）、`attachType` 和 `cdpUrl`（attach 的实例）、`browser`、`fallbackFrom`/`fallbackReason`（回退到其他目标的一次启动）以及 `crashes`。

## 启动实例

### `POST /instances/start`

当你想按 profile ID 或 profile 名启动，或让 PinchTab 创建临时 profile 时，使用 `/instances/start`。

```bash
curl -X POST http://localhost:9867/instances/start \
  -H "Content-Type: application/json" \
  -d '{"profileId":"prof_278be873","mode":"headed","port":"9999","securityPolicy":{"allowedDomains":["wikipedia.org","wikimedia.org"]}}'
# CLI Alternative
pinchtab instance start --profile prof_278be873 --mode headed --port 9999 --allow-domain wikipedia.org --allow-domain wikimedia.org
```

请求体：

- `profileId`：可选；接受 profile ID 或现有 profile 名
- `mode`：可选；用 `headed` 表示可见浏览器，其他值视为无头
- `port`：可选
- `securityPolicy.allowedDomains`：可选的、附加性的实例级 IDPI/域名允许列表条目
- `browser`：可选的 provider 或 `browser.targets` 名（命令行界面 `--browser`）
- `fallbackTargets`：可选的有序目标名，首次启动失败时依次尝试（命令行界面 `--browser-fallback`，可重复）

响应为 `201`，带实例对象。

注意：

- 省略 `profileId` 时，PinchTab 创建自动生成的临时 profile
- 省略 `port` 时，PinchTab 从配置的实例端口范围分配一个
- 命令行界面 flag 是 `--profile`，尽管 API 字段是 `profileId`
- `securityPolicy.allowedDomains` 仅为该实例与服务器级 `security.allowedDomains` 基线合并
- 你可以放宽单个实例而不改服务器默认值。例如 `{"securityPolicy":{"allowedDomains":["*"]}}` 让该实例不受限制，而其他实例仍用服务器基线
- 请求提供的扩展路径被拒绝（因此 `pinchtab instance start --extension` 以 `400` 失败）；改为在服务器上配置 `browser.extensionPaths`。默认情况下，PinchTab 使用其状态/配置文件夹下的本地 `extensions/` 目录

### `POST /instances/launch`

`/instances/launch` 是 `/instances/start` 的兼容别名。

```bash
curl -X POST http://localhost:9867/instances/launch \
  -H "Content-Type: application/json" \
  -d '{"profileId":"prof_278be873","mode":"headed","securityPolicy":{"allowedDomains":["wikipedia.org"]}}'
```

请求体：

- `profileId`：可选的现有 profile ID 或现有 profile 名
- `mode`：可选；`headed` 或默认为无头
- `port`：可选
- `securityPolicy.allowedDomains`：可选的、附加性的实例级 IDPI/域名允许列表条目
- `browser`、`fallbackTargets`：可选，同 `/instances/start`

重要：

- `/instances/launch` 不读取 `headless` 字段。想要有头浏览器时用 `mode:"headed"`。
- `name` 在 `/instances/launch` 上不再支持。先通过 `POST /profiles` 创建 profile，再用返回的 `id` 作为 `profileId`。
- 请求提供的扩展路径被拒绝；改为在服务器上配置 `browser.extensionPaths`。默认情况下，PinchTab 使用其状态/配置文件夹下的本地 `extensions/` 目录。

## 获取一个实例

```bash
curl http://localhost:9867/instances/inst_ea2e747f
```

常见状态值：

- `starting`
- `running`
- `stopping`
- `stopped`
- `error`

实例响应包括：

- `mode`：`"headless"` 或 `"headed"`
- `headless`：为兼容保留的布尔值
- `responsiveness`：该实例的浏览器路由是否应答，由对该实例 `/tabs` 在短预算下最近一次完成的探测测量，front door `/health` 或监控快照会启动它并最多等待 250ms；更慢的探测由下一次读取报告。应答了为 `responsive`，`/health` 应答但 `/tabs` 在预算内未应答为 `unresponsive`，尚未探测或探测无法连接为 `unknown`。它绝不改变 `status`，无响应实例的 `status` 仍为 `running`，也不触发重启。任一实例 `unresponsive` 时 front door 的 `/health` 降级

## 获取实例日志

```bash
curl http://localhost:9867/instances/inst_ea2e747f/logs
# CLI Alternative
pinchtab instance logs inst_ea2e747f
```

响应为纯文本。`GET /instances/{id}/logs/stream` 还有一个 SSE 流。

## 停止实例

```bash
curl -X POST http://localhost:9867/instances/inst_ea2e747f/stop
# CLI Alternative
pinchtab instance stop inst_ea2e747f
```

停止实例会保留 profile，除非它是临时自动生成的 profile。响应为 `{"status":"stopped","id":"<instanceId>"}`。

## 按 ID 重启或启动实例

```bash
curl -X POST http://localhost:9867/instances/inst_ea2e747f/restart
# CLI Alternative
pinchtab instance restart inst_ea2e747f
```

`POST /instances/{id}/restart` 对运行中实例的浏览器进程做软重启（未运行时 `503`）。`POST /instances/{id}/start` 用之前的 profile、端口和模式重新启动一个已知的已停止实例（`201` 带实例），或确保仍活动实例的浏览器就绪；attach 的 CDP 实例不能以此方式启动（`409`）。

## 按 profile 启动

你也可以从面向 profile 的路由启动实例：

```bash
curl -X POST http://localhost:9867/profiles/prof_278be873/start \
  -H "Content-Type: application/json" \
  -d '{"headless":false,"port":"9999","securityPolicy":{"allowedDomains":["wikipedia.org"]}}'
```

此路由在路径中接受 profile ID 或 profile 名。与 `/instances/start` 和 `/instances/launch` 不同，其请求体用 `headless` 而非 `mode`。`headless` 默认为 `false`，因此不带它的请求启动有头浏览器。主体还接受 `browser` 和 `fallbackTargets`。

## 在实例中打开标签页

```bash
curl -X POST http://localhost:9867/instances/inst_ea2e747f/tabs/open \
  -H "Content-Type: application/json" \
  -d '{"url":"https://pinchtab.com"}'
```

没有专用的实例级 `tab open` 命令行界面命令。命令行界面快捷方式是：

```bash
pinchtab instance navigate inst_ea2e747f https://pinchtab.com
```

该命令在一次 `tabs/open` 调用中就为该实例打开已在该 URL 上的标签页。

## 列出一个实例的标签页

```bash
curl http://localhost:9867/instances/inst_ea2e747f/tabs
```

## 列出所有运行实例的标签页

```bash
curl http://localhost:9867/instances/tabs
```

这是全实例群的标签页列表端点。它不同于 `GET /tabs`，后者是简写或 bridge 范围的。

两个路由都返回 `{"id","instanceId","url","title"}` 对象的裸 JSON 数组，按实例缓存；加 `?fresh=1` 重新拉取。实例未运行时 `GET /instances/{id}/tabs` 应答 `503`。

## 列出跨实例指标

```bash
curl http://localhost:9867/instances/metrics
```

`GET /instances/{id}/metrics` 返回单个实例自己的请求计数器；服务器上裸 `GET /metrics` 应答服务器自己的计数器（见 [Metrics](./metrics.md)）。

## 附加现有浏览器（CDP）

```bash
curl -X POST http://localhost:9867/instances/attach \
  -H "Content-Type: application/json" \
  -d '{
    "name":"shared-chrome",
    "cdpUrl":"ws://127.0.0.1:9222/devtools/browser/...",
    "provider":"chrome",
    "browser":"chrome-local"
  }'
```

它做的事：

- 派生一个子 `pinchtab bridge --cdp-attach ...` 进程，通过 `chromedp.NewRemoteAllocator` 包装外部 CDP 浏览器
- 以 `attached=true`、`attachType="cdp-bridge"` 注册该实例，可路由 `url` 设为 **HTTP bridge URL**（非原始 `ws://` CDP URL）
- 在 API 响应中保留原始 `cdpUrl` 作为元数据
- 通过子 bridge 路由 `/tabs`、`/snapshot`、`/action`、`/screenshot` 等

接受的 `cdpUrl` 形态：

- 浏览器级 WebSocket URL：`ws://host:port/devtools/browser/<id>`
- HTTP DevTools 源：`http://host:port`（通过 `/json/version` 解析）
- HTTP `/json/version` URL

页面级 URL（`/devtools/page/...`）被拒绝。

`browser` 可选，接受 provider 名（`chrome`、`cloak`）或 `browser.targets` 中配置的目标名。配置了 `browser.targets` 时，省略值则 attach 到配置的默认目标并使用该目标的浏览器。若同时有 `provider`，它必须与 `browser` 值一致。无 `browser.targets` 时，`provider` 为 `chrome`（默认）或 `cloak`；cloak 浏览器假定外部浏览器原生拥有指纹行为，因此禁用 PinchTab JS 覆盖层。

注意：

- 没有命令行界面 attach 命令
- 仅当配置中 `security.attach` 下启用时才允许 attach
- `security.attach.allowHosts` 必须允许 `cdpUrl` 主机
- 若传 HTTP DevTools 源，`security.attach.allowSchemes` 必须包含 `http` 或 `https`
- 停止 attach 的实例会关闭子 PinchTab bridge，但外部浏览器进程继续运行
- `allowHosts: ["*"]` 是文档化的、非默认的、降低安全性的覆盖

## 附加现有 Bridge

```bash
curl -X POST http://localhost:9867/instances/attach-bridge \
  -H "Content-Type: application/json" \
  -d '{
    "name":"shared-bridge",
    "baseUrl":"http://10.0.12.24:9868",
    "token":"bridge-secret-token",
    "browser":"chrome-local"
  }'
```

注意：

- `baseUrl` 必须是裸 bridge 源；不要包含凭据、查询串、片段或路径
- `browser` 可选，接受 provider 或目标名；配置了 `browser.targets` 时，省略值 attach 到默认目标
- 编排器在注册前做一次健康检查
- `security.attach.allowHosts` 必须允许 bridge 主机
- `allowHosts: ["*"]` 是文档化的、非默认的、降低安全性的覆盖。它完全禁用主机允许列表，允许任何带允许 scheme 的可达 bridge 主机。仅在隔离的、操作员控制的网络上使用。
