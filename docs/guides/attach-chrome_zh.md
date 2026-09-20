# 附加 Chrome

当出现以下情况时使用本指南：

- Chrome 已经存在于 PinchTab 之外
- 你希望 PinchTab 服务器把该浏览器注册为一个实例
- 你已经有一个浏览器级别的 DevTools WebSocket URL

如果你的目标只是：

- 为你的代理启动一个浏览器
- 运行常规的本地 PinchTab 工作流

则不要使用本指南。

那种情况下，请使用受管实例，配合 `pinchtab` 和 `POST /instances/start`。

---

## 启动 vs 附加

心智模型是：

```text
launch = PinchTab starts and owns the browser
attach = PinchTab registers an already running browser
```

使用附加时：

- Chrome 在别处启动
- PinchTab 收到一个 `cdpUrl`
- 服务器把该浏览器注册为一个附加实例

---

## 当前已实现的内容

当前代码库实现了：

- `POST /instances/attach`——启动一个子进程 `pinchtab bridge --cdp-attach ...`，包装外部浏览器。该 bridge 提供标准的 PinchTab HTTP API；编排器把 bridge 的 HTTP URL（而不是原始的 `ws://` CDP URL）注册为可路由的实例 URL。
- `POST /instances/attach-bridge`——把一个已经在运行的 PinchTab bridge 注册为实例（未改动）。
- 配置中 `security.attach` 下的附加策略
- `GET /instances` 中的附加实例元数据

附加请求体为：

```json
{
  "name": "shared-chrome",
  "cdpUrl": "ws://127.0.0.1:9222/devtools/browser/...",
  "browser": "chrome"
}
```

`browser`（或其别名 `provider`）是可选的，它接受一个提供者名（`chrome`、`cloak`、`ghost-chrome`），而不是 target 名。当配置了 `browser.targets` 时，该提供者必须至少有一个已配置的 target；省略该值则用默认 target 的提供者进行附加。如果同时传 `browser` 和 `provider`，二者必须一致。没有 browser targets 时，提供者默认为 `chrome`；对 CloakBrowser 端点使用 `cloak`（等价于 CLI 上的 `--browser cloak`）。

可接受的 `cdpUrl` 形式：

- 浏览器级 WebSocket URL：`ws://host:port/devtools/browser/<id>`
- HTTP DevTools 源：`http://host:port`（通过 `/json/version` 解析）
- HTTP `/json/version` URL

页面级 URL（`/devtools/page/...`）会被拒绝。

目前还没有 CLI 附加命令。

---

## 第 1 步：启用附加策略

除非你在配置中允许，否则附加是禁用的。

示例：

```json
{
  "security": {
    "attach": {
      "enabled": true,
      "allowHosts": ["127.0.0.1", "localhost", "::1"],
      "allowSchemes": ["ws", "wss"],
      "forwardProxyAuth": false
    }
  }
}
```

这会做的事：

- 启用附加端点
- 限制接受哪些主机
- 限制接受哪些 URL scheme

这不会做的事：

- 它不会启动 Chrome
- 它不会定义一个全局远程浏览器
- 它不会取代受管实例

---

## 第 2 步：以远程调试方式启动 Chrome

示例：

```bash
google-chrome --remote-debugging-port=9222
# Or on some systems:
# chromium --remote-debugging-port=9222
```

这会让 Chrome 暴露一个浏览器级别的 DevTools 端点。

---

## 第 3 步：获取浏览器 WebSocket URL

查询 Chrome：

```bash
curl -s http://127.0.0.1:9222/json/version | jq .
# Response
{
  "webSocketDebuggerUrl": "ws://127.0.0.1:9222/devtools/browser/abc123"
}
```

`webSocketDebuggerUrl` 的值就是你传给 PinchTab 的 `cdpUrl`。

---

## 第 4 步：把它附加到 PinchTab

```bash
curl -X POST http://localhost:9867/instances/attach \
  -H "Authorization: Bearer $(pinchtab config token --stdout)" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "shared-chrome",
    "cdpUrl": "ws://127.0.0.1:9222/devtools/browser/abc123",
    "browser": "chrome"
  }'
# Response (201, abridged)
{
  "id": "inst_0a89a5bb",
  "profileId": "prof_278be873",
  "profileName": "shared-chrome",
  "port": "9868",
  "url": "http://127.0.0.1:9868",
  "mode": "headed",
  "headless": false,
  "status": "running",
  "attached": true,
  "attachType": "cdp-bridge",
  "cdpUrl": "ws://127.0.0.1:9222/devtools/browser/abc123",
  "browser": "chrome"
}
```

注意：

- `name` 是可选的；若省略，服务器会生成一个形如 `attached-...` 的名称
- 服务器会对照 `security.attach.allowHosts` 和 `security.attach.allowSchemes` 校验该 URL（被拒绝的 URL 返回 `403`）
- `port` 和 `url` 属于 PinchTab 从实例端口范围派生的子 bridge，而不属于 Chrome

---

## 第 5 步：确认它已注册

```bash
curl -s -H "Authorization: Bearer $(pinchtab config token --stdout)" http://localhost:9867/instances | jq .
# CLI Alternative
pinchtab instance list
```

一个附加实例会出现在常规实例列表中，并带有：

- `attached: true`
- `attachType: "cdp-bridge"`
- `cdpUrl: ...`
- `status: "running"`

---

## 所有权与生命周期

附加实例由外部拥有，但 PinchTab 仍然持有一个包裹它们的 *bridge 包装器*。

这意味着：

- PinchTab 没有启动该浏览器
- PinchTab 派生了一个子进程 `pinchtab bridge --cdp-attach ...`，它包装外部 CDP 端点并提供常规的 PinchTab 路由
- 外部 Chrome/CloakBrowser 进程始终在 PinchTab 生命周期所有权之外

具体来说：

- `POST /instances/{id}/stop` 只会关闭子 PinchTab bridge——外部 Chrome 进程**保持运行**
- 像 `/tabs`、`/snapshot`、`/action`、`/screenshot` 这样的路由会到达子 bridge，后者通过 `chromedp.NewRemoteAllocator` 与外部浏览器进行 CDP 通信

---

## 何时适合使用附加

在以下情况使用附加：

- Chrome 由另一个系统管理
- Chrome 已经在某个独立服务或容器中运行
- 你希望服务器知道一个外部托管的浏览器
- 你希望把浏览器的所有权保留在 PinchTab 之外

---

## 安全

附加拓宽了信任边界，因此请把它牢牢锁住。

推荐规则：

- 除非需要，否则保持附加禁用
- 保持 `allowHosts` 收窄
- 保持 `allowSchemes` 收窄
- 除非所附加的浏览器进程和 CDP 传输都可信，否则保持 `forwardProxyAuth` 禁用
- 当服务器可从 localhost 之外访问时设置 `PINCHTAB_TOKEN`
- 只附加到你信任的 CDP 端点

如果你把 `allowHosts` 设为 `["*"]`，PinchTab 会接受任何具备允许 scheme 的可访问附加主机。这是一个有文档记录的、非默认的、降低安全性的覆盖项：它完全移除了主机允许列表，只应在隔离的、由操作员控制的网络上使用。

还要记住：

- Chrome DevTools 提供强大的浏览器控制能力
- 一个可访问的 CDP 端点应被视为敏感基础设施

如果 Chrome 是远程的，更倾向于用隧道，而不是大范围暴露调试端口。

---

## 操作模型

预期的模型是：

```text
agent -> PinchTab server -> attached external Chrome
```

这是一条专家路径，不是默认的用户路径。

默认路径仍然是：

```bash
pinchtab
```

然后通过以下方式启动受管实例：

```bash
curl -X POST http://localhost:9867/instances/start \
  -H "Authorization: Bearer $(pinchtab config token --stdout)" \
  -H "Content-Type: application/json" \
  -d '{"mode":"headless"}'
# CLI Alternative
pinchtab instance start
```

---

## 相关指南

- [cloakbrowser.md](cloakbrowser.md)——完整的 CloakBrowser 配置以及 `--browser cloak` 附加变体
- [docker.md](docker.md)——在容器中运行 PinchTab（以及本地 CloakBrowser 冒烟镜像）
- [headed-mode.md](headed-mode.md)——在捆绑镜像之外的手动有头设置
- [security.md](security.md)——附加策略细节、IDPI、令牌处理
