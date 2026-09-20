# 安全

PinchTab 的设计目标是：默认在本地机器上即可使用，除非你明确开启，否则不暴露高风险的浏览器控制功能。

PinchTab 默认且主要的部署模型是本地优先：一个用户、一台机器、一个操作员控制的浏览器控制平面。更复杂的拓扑——如 Docker、LAN 访问、远程桥接或分布式编排器部署——也受支持，但它们属于高级部署。PinchTab 不应被当作开箱即用的、面向公网的服务；确保这些部署的安全是操作员的责任。

如果你在另一台机器上运行 PinchTab，只有在你理解自己所操作的安全模型时才这么做。优先使用私有或其他封闭网络，避免把服务直接暴露到公网，并保持高风险能力禁用，除非该部署确实需要它们。如果必须启用，就加以限制，让只有需要它们的最小受信任系统才能触达。

> [!WARNING]
> PinchTab 的仪表板、HTTP API、远程 CLI 目标、MCP 集成和自动化路由都属于同一个特权控制平面。它们只面向受信任的操作员和受信任的系统。不要把它们暴露给不受信任的用户、不受信任的客户端系统或公网。
>
> 如果你不确定一个非本地或部分暴露的部署是否安全，先不要暴露它。先阅读本指南，并在继续之前使用 `SECURITY.md` 中的私人安全联络渠道。

默认安全姿态为：

- `server.bind = 127.0.0.1`
- `server.token` 在默认设置时生成，应保持设置
- `security.allowEvaluate = false`
- `security.allowMacro = false`
- `security.allowScreencast = false`
- `security.allowDownload = false`
- `security.allowCookies = false`
- `security.allowUpload = false`
- `security.allowStateExport = false`
- `security.allowNetworkIntercept = false`
- `security.allowMemory = false`
- `security.allowClipboard = false`
- `security.allowFileScheme = false`
- `autoSolver.enabled = false`
- `instanceDefaults.stealthLevel = "light"`（仅做最少量的指纹归一化；反机器人绕过需要显式 opt-in 到 `medium` 或 `full`）
- `security.attach.enabled = false`
- `security.attach.allowHosts = ["127.0.0.1", "localhost", "::1"]`
- `security.attach.allowSchemes = ["ws", "wss", "http", "https"]`
- `security.attach.forwardProxyAuth = false`
- `security.allowedDomains` 未设置（不做域名限制；SSRF/私网 IP 守卫仍然生效）
- `security.trustedProxyCIDRs = []`
- `security.trustedResolveCIDRs = []`
- `security.idpi.enabled = true`
- `security.idpi.strictMode = true`
- `security.idpi.scanContent = true`
- `security.idpi.wrapContent = true`

用 `pinchtab security` 查看当前姿态，用 `pinchtab security up` 恢复推荐默认值（环回绑定、所有能力关闭、仅本地 attach、IDPI 开启）。`security up` 会把整个 `security` 块重置为上面的默认值，因此也会清除 `security.allowedDomains`；如果你想要一个仅本地的允许列表，请另行设置（见 [IDPI](#idpi)）。

## 安全理念

PinchTab 遵循几条简单规则：

- 默认仅本地访问
- 默认关闭危险能力
- 把传输访问与功能暴露分开
- 当无法建立内容或域名信任时，默认拒绝

这意味着有两个独立的问题：

1. 谁能访问到服务器
2. 服务器被访问后被允许做什么

两者都重要。

## 信任边界

重要的操作规则很简单：

- 如果某个人或系统不应该被允许控制浏览器状态、Profile、配置、attach 或敏感端点族，它就不应该能访问到 PinchTab，也不应该被授予 PinchTab 的凭据

这包括：

- 浏览器仪表板
- 直接的 HTTP API 客户端
- 用 `--server` 针对远程服务器的 CLI 用法
- MCP 客户端、插件、脚本以及构建在 API 之上的其他自动化层

这些是同一控制平面的不同接口，不是各自独立的信任域。

## 高级部署

如果你有意在默认本地设置之外运行 PinchTab，操作员最低清单是：

- 把 `server.token` 设为一个强随机值
- 用受信任的网络边界、VPN、防火墙或反向代理收窄网络可达性
- 当流量离开本机时，在代理或传输层加 TLS
- 仅当确实有一个受信任的反向代理在为你剥离并重建 `Forwarded` / `X-Forwarded-*` 头时，才启用 `server.trustProxyHeaders`
- 保持敏感端点族禁用，除非明确需要；若启用，则把它们限制在必须访问的最小受信任调用者或网络路径上
- 为你所操作的远程拓扑有意地设定 `security.attach` 和 `security.idpi`

这些选择是部署责任，不是 PinchTab 能替你安全推断的默认值。

当服务器不与用户或代理在同一台机器上运行时，门槛应更高：知道哪些主机能访问它、哪些凭据保护它、哪些端点族已启用、以及哪个网络边界在围堵它。

绑定到环回减少了谁能访问 API。令牌减少了谁能成功使用它。敏感端点门减少了成功调用者能做什么。IDPI 减少了哪些网站和提取内容被信任到足以深入代理工作流。

## API 令牌

`server.token` 是主 API 令牌。

对于非浏览器客户端，请求应发送：

```http
Authorization: Bearer <token>
```

浏览器仪表板使用另一种流程：

1. 用户在登录页输入一次令牌
2. 服务器把它交换成一个同源 `HttpOnly` 会话 cookie
3. 敏感的仪表板操作可以要求重新输入令牌，以进行短期提权

默认情况下，PinchTab 自动检测仪表板会话 cookie 是否应使用 `Secure` flag。在 `auto` 模式下，HTTPS 请求得到 `Secure` cookie，纯 HTTP 请求则不。

这意味着：

- 反向代理后的 HTTPS 保持 `Secure` 启用
- 纯 `http://localhost:9867` 在仅本地使用时继续工作
- 纯 `http://192.168.x.x:9867` 或 `http://10.x.x.x:9867` 可用，但仪表板会警告会话正在不安全的 HTTP 上运行

如果你想要求仪表板登录必须走 HTTPS，强制 `server.cookieSecure` 为 `true`：

```json
{
  "server": {
    "cookieSecure": true
  }
}
```

在纯 HTTP 上，这会明确失败，返回一个「需要 HTTPS」的登录错误，而不是看似成功然后循环跳转。

如果你有意在受信任的 LAN 上需要纯 HTTP，也可以显式强制 `cookieSecure` 关闭：

```json
{
  "server": {
    "cookieSecure": false
  }
}
```

推荐用法：

- 除非有理由覆盖，否则保持 `cookieSecure` 不设置（`auto`）
- TLS 在 PinchTab 前面时用 `cookieSecure: true`
- 仅在操作员控制的纯 HTTP 部署上用 `cookieSecure: false`
- 如果 TLS 在受信任的反向代理处终止，启用 `server.trustProxyHeaders`，以便识别被转发的 HTTPS 请求

为什么这很重要：

- 没有令牌，任何能访问到服务器的进程都能调用 API
- 在 `127.0.0.1` 上，这仍然包括本地脚本、浏览器页面、同机其他用户和恶意软件
- 在 `0.0.0.0` 或 LAN 绑定上，缺令牌是大得多的风险

推荐做法：

- 把 `server.bind` 保持在 `127.0.0.1`
- 设一个强随机 `server.token`
- 仅当远程访问是有意的时才扩大绑定

`pinchtab config init` 会在默认设置时生成并存储一个令牌：

```bash
pinchtab config init
```

仪表板设置页面不暴露也不轮换 `server.token`。用 `pinchtab config token` 复制当前令牌（在没有剪贴板的地方用 `pinchtab config token --stdout` 打印它），或者在 `server.token` 为空时让 `pinchtab security up` 创建一个。

如果你手动调用 API：

```bash
curl -H "Authorization: Bearer <token>" http://127.0.0.1:9867/health
```

CLI 命令默认使用配置中的本地服务器设置，`PINCHTAB_TOKEN` 可以在单个 shell 会话内覆盖令牌。

## 代理会话

代理会话是面向受信任自动化的、缩小分发范围的凭据，不是给不受信任客户端的沙箱。

- 经过会话认证的调用者被禁止访问仪表板/管理端点族，例如配置、会话管理、Profile 管理、实例管理、仪表板代理列表和缓存控制
- 会话记录可以选择携带显式 grant，进一步收窄访问范围——在创建时用 `pinchtab session create --agent-id <id> --grant browse` 设置，或在 `POST /sessions` 上用 `grants` 字段（见 [sessions](../reference/sessions.md#session-grants)）
- 没有显式 grant 的会话默认仍可使用常规的非管理自动化 API
- grant 只会收窄、不会放宽：每一道服务器级能力门仍然在其之上生效，因此 `--grant evaluate` 不会重新启用 `security.allowEvaluate`，任何 grant 也够不到管理路由

这意味着代理会话适合受控环境——调用者已被信任来驱动浏览器自动化，但不应拿到完整的仪表板 bearer 令牌。它们不足以应对恶意多租户共享或公网暴露。那种隔离需要在各自独立的网络和凭据边界后面运行单独的 PinchTab 实例。

## 敏感端点

有些端点族比普通导航和检查暴露大得多的权力。PinchTab 默认保持它们禁用：

- `security.allowEvaluate`
- `security.allowMacro`
- `security.allowScreencast`
- `security.allowDownload`
- `security.allowCookies`
- `security.allowUpload`
- `security.allowStateExport`
- `security.allowNetworkIntercept`
- `security.allowMemory`
- `security.allowClipboard`
- `security.allowFileScheme`

为什么它们被认为危险：

- `evaluate` 可以在页面上下文里执行 JavaScript
- `macro` 可以触发更高级的自动化流程
- `screencast` 可以实时串流页面内容
- `download` 可以抓取并持久化远程内容。当设置了 `security.downloadAllowedDomains` 时，匹配的域名会绕过私网 IP SSRF 检查（预期用于内部主机，例如 Docker 服务）。一个裸 `"*"` 匹配所有主机，因此会为该端点禁用私网 IP 保护，包括环回。指定一个环回主机（`127.0.0.1`、`localhost`）或使用 `"*"` 都会让下载端点访问到服务器本机上的服务，包括 PinchTab 自己的本地端点，因此请像对待 `allowFileScheme` 一样对待这两种配置。
- `cookies` 可以读取、写入或清除当前页面的浏览器会话令牌
- `upload` 可以把本地文件推进浏览器流程
- `stateExport` 把 cookies 和浏览器存储写到磁盘并重新加载（`/state/*`、`/storage`）
- `networkIntercept` 可以重写、拦截或 mock 请求，并读取完整请求细节（`/network/route`、`/network/{requestId}`）
- `memory` 写 V8 堆快照，其中包含页面上的每一个字符串，包括令牌（`/memory/snapshot`、`/memory/compare`）
- `clipboard` 可以读写浏览器剪贴板（`/clipboard/*`）
- `allowFileScheme` 允许导航到 `file://` URL。因为 `file://` URL 没有 host，它**不**受 `allowedDomains` 或 SSRF/私网 IP 守卫约束，因此启用它就授予了对服务器进程能读的任何本地文件的读取访问（通过 snapshot/screenshot/scrape）。当严格模式的 `allowedDomains` 允许列表生效时它保持被阻止。只在受信任的单租户主机上启用。`javascript:`、`chrome://` 和 `data:` 无论如何都被拒绝。

这些与认证不是一回事。

- 认证决定谁可以调用 API
- 敏感端点门决定哪些高风险能力到底存在

例如，一个受令牌保护、但 `security.allowEvaluate = true` 的服务器，仍然是有意地把 JavaScript 执行暴露给任何持有令牌的调用者。

禁用时，这些路由被锁死并返回 `403`，说明该端点族在配置中被禁用（错误码如 `evaluate_disabled`、`upload_disabled`、`memory_disabled`）。

## Attach 策略

Attach 是一个高级功能，用于通过 CDP URL 注册一个外部受管的 Chrome 实例。它默认禁用：

```json
{
  "security": {
    "attach": {
      "enabled": false,
      "allowHosts": ["127.0.0.1", "localhost", "::1"],
      "allowSchemes": ["ws", "wss", "http", "https"],
      "forwardProxyAuth": false
    }
  }
}
```

如果你启用 attach：

- 保持 `allowHosts` 范围狭窄
- 除非外部 Chrome 目标或远程桥接是有意的，否则优先仅本地主机
- 只附加到你信任的浏览器和 CDP 端点
- `allowHosts: ["*"]` 是有文档记录、非默认、降低安全性的覆盖。它完全禁用主机允许列表，允许任何具备允许 scheme 的可访问 attach 主机。只在隔离的、操作员控制的网络上使用。
- 除非所附加的浏览器进程和 CDP 传输都受信任，否则保持 `forwardProxyAuth` 禁用；启用它会允许 PinchTab 通过 CDP WebSocket 发送已配置的代理凭据。

有两个 attach 端点，信任形态不同：

- `POST /instances/attach`——通过 `cdpUrl` 附加一个已有的 **CDP 浏览器**。PinchTab 派生一个子进程 `pinchtab bridge --cdp-attach ...`，包装该外部端点，并把 bridge 的本地 HTTP URL（而不是原始 `ws://` CDP URL）注册为可路由的实例 URL。CDP URL 作为元数据保留，并在日志中**脱敏**。
- `POST /instances/attach-bridge`——通过其 HTTP `baseUrl` 附加一个已在运行的 **PinchTab bridge**。编排器在注册前先跑一次 `/health` 检查。

Scheme 允许列表规则：

- `ws`、`wss`——用浏览器 WebSocket URL 做 CDP attach 时必需
- `http`、`https`——用 HTTP DevTools 源做 CDP attach *以及* `POST /instances/attach-bridge` 时必需

当 `allowHosts` 包含 `"*"` 时，`security.attach.allowSchemes` 和 `security.attach.enabled` 仍然生效，但在该配置下主机允许列表不再提供保护。

对于 `attach-bridge`，`baseUrl` 应是一个裸 bridge 源，例如 `http://bridge.internal:9868`。不要包含凭据、查询字符串、片段或路径。对于通过 HTTP 的 CDP attach，只接受裸源或 `/json/version` 路径。

停止一个 CDP 附加实例只会关闭子 PinchTab bridge，绝不会杀掉外部浏览器进程——PinchTab 不拥有那个进程。

## IDPI

IDPI 代表 Indirect Prompt Injection defense（间接提示注入防御）。

它的存在是为了降低不受信任的网站内容通过隐藏指令、投毒文本或不安全导航影响下游代理的概率。

PinchTab 的 IDPI 层目前做四件事：

- 把导航限制到一个已批准域名的允许列表
- 当某个 URL 无法匹配该允许列表时，阻止或警告
- 扫描提取出的内容，查找可疑的提示注入模式
- 包装文本输出，让下游系统能把它当作不可信内容处理

IDPI 默认开启，但 `allowedDomains` 未设置，因此在你设置它之前导航不做域名限制。一个仅本地的 IDPI 配置长这样：

```json
{
  "security": {
    "allowedDomains": ["127.0.0.1", "localhost", "::1"],
    "trustedProxyCIDRs": [],
    "trustedResolveCIDRs": [],
    "idpi": {
      "enabled": true,
      "strictMode": true,
      "scanContent": true,
      "wrapContent": true,
      "customPatterns": []
    }
  }
}
```

重要说明：

- 空的 `allowedDomains` **解除域名限制但不授予任何东西**。它不是一种放行：没有任何主机算被显式允许，因此 SSRF/私网 IP 守卫会像 IDPI 关闭时一样继续拒绝私网和内部地址。（过去用空列表启用 IDPI 反而会*移除*该保护，因为「扫描器没发现可疑内容」被误读为「操作员允许了这个主机」。）
- 显式列出一个私网或内部主机——`["10.0.0.5"]`，或仅本地的 `["127.0.0.1", "localhost", "::1"]`——才是允许导航到它的做法。这是放宽私网 IP 守卫的唯一途径，而且它是由一个**指明主机**的条目做正向匹配，绝不是因为列表为空，也绝不是一个裸 `"*"`。
- 如果 `allowedDomains` 包含 `"*"`，白名单实际上允许一切：对所有主机解除域名限制。它**不**授予私网 IP 豁免——`"*"` 不指明任何主机，意图与空列表相同——因此配 `["*"]` 时 SSRF 守卫仍拒绝私网和内部地址。要访问某个私网地址，就把它点名：`["*", "10.0.0.5"]` 保持无限制导航并允许该主机。
- `security.allowedDomains` 是规范配置路径。加载旧配置文件时仍接受 `security.idpi.allowedDomains`，但新保存会被规范化到 `security.allowedDomains`
- `strictMode = true` 阻止不允许的域名和可疑内容
- `strictMode = false` 允许请求但发出警告
- `scanContent` 覆盖所有返回页面可控内容的端点：`/text`、`/snapshot`、`/capture`、`/find`、`/scrape`、`/pdf`、`/html` 和 `/styles`，以及它们的 `/tabs/{id}/...` 形式。一个连 `/text` 都拒绝的页面，会被它们全部拒绝，因此没有哪个单点端点能绕开扫描器
- `wrapContent` 为下游消费者加上显式的不可信内容包装。它作用于文本形态的响应（`/text`、`/snapshot`、`/capture`），不作用于 `/html` 或 `/styles`：包装原始标记会把它弄得无法交给 HTML 解析器调用方解析。这两个端点会被扫描并阻止或警告，并在 `X-IDPI-Warning` 以及响应的 `idpiWarning` 字段中携带提示
- 把导航放宽到非本地或非受信任站点仍然是一个降低安全性的选择；IDPI 降低风险，但它不会让恶意页面变安全，也不会移除浏览器攻击面

关于导航信任覆盖：

- `security.trustedResolveCIDRs` 允许主机名在导航预检期间解析到一个非公网 IP。预期用于操作员控制的 DNS 或代理设置，例如内部代理、实验室网络或基准测试网段
- `security.trustedProxyCIDRs` 在运行时导航检查中信任来自已知内部代理的、浏览器上报的远程 IP
- 两个列表都保持收窄。像 `10.0.0.0/8` 这样的大段范围会削弱 SSRF 保护，只有在整段网络都被有意信任时才应使用
- 已知局限：从浏览器缓存或 service worker 提供的响应不上报远程 IP，因此运行时远程 IP 检查按设计会放行它们；解析时检查仍是这类导航的主要门槛

支持的域名模式为：

- 精确主机：`example.com`
- 子域通配：`*.example.com`
- 全通配：`*`

`*` 很方便，但它会打掉主要的允许列表防线，除非你有意禁用域名限制，否则应避免。

如果你只需要为某一个受管浏览器放宽信任，优先用实例级覆盖，而不是改全局服务器策略。`POST /instances/start`、`POST /instances/launch` 和 `POST /profiles/{id}/start` 接受：

```json
{
  "securityPolicy": {
    "allowedDomains": ["*"]
  }
}
```

该覆盖仅对该实例附加生效。例如，你可以保持服务器基线仅本地，同时启动一个带 `allowedDomains: ["*"]` 或一个窄的额外主机列表（如 `["wikipedia.org"]`）的临时实例，而不放宽服务器的其余部分。

## 推荐配置

一个安全的本地部署：

```json
{
  "server": {
    "bind": "127.0.0.1",
    "token": "replace-with-a-generated-token"
  },
  "security": {
    "allowEvaluate": false,
    "allowMacro": false,
    "allowScreencast": false,
    "allowDownload": false,
    "allowCookies": false,
    "allowUpload": false,
    "allowedDomains": ["127.0.0.1", "localhost", "::1"],
    "trustedProxyCIDRs": [],
    "trustedResolveCIDRs": [],
    "attach": {
      "enabled": false,
      "allowHosts": ["127.0.0.1", "localhost", "::1"],
      "allowSchemes": ["ws", "wss", "http", "https"],
      "forwardProxyAuth": false
    },
    "idpi": {
      "enabled": true,
      "strictMode": true,
      "scanContent": true,
      "wrapContent": true,
      "customPatterns": []
    }
  }
}
```

如果你有意把 PinchTab 暴露到 localhost 之外，请把令牌视为必需，并保持敏感端点族禁用，除非你有特定理由启用它们。对于任何比单机本地部署更暴露的场景，都按高级部署对待，逐项审查每个安全控制。

## 已认证的浏览器会话

当代理复用人类已认证（已登录）的浏览器会话时，请遵循以下做法：

- 使用一个**专用的低权限 Profile**——不要用用户的个人浏览 Profile
- 在复用会话中执行改账户类操作（改密码、支付、删除、权限）之前，**先与用户确认**
- 通过 `security.allowedDomains` 或实例级 `securityPolicy.allowedDomains` 把导航限制在任务所需的站点

这些是在代理/skill 层执行的操作指引，不是 API 级门槛。Profile 系统（`POST /profiles`）支持元数据字段（`name`、`description`、`useWhen`），帮助代理为任务挑选合适的 Profile。

## Daemon 生命周期

后台 daemon 是为持久本地浏览器控制提供便利的。安装后它持续运行（macOS 上 `KeepAlive`，Linux 上 `Restart=always`）。

当不再需要浏览器自动化时，停用它：

- `pinchtab daemon stop`——停止服务但不移除它
- `pinchtab daemon uninstall`——停止、禁用并移除服务文件

对于短命或一次性使用，更倾向于 `pinchtab server`（前台进程，终端关闭即退出）而非 daemon。

代理会话凭据默认在**空闲 30 分钟**后自动过期（`sessions.agent.idleTimeoutSec: 1800`），并有 **24 小时最大生命周期**（`sessions.agent.maxLifetimeSec: 86400`）。

## 代理会话令牌

对于自动化代理，请使用 **代理会话**，而不是共享服务器 bearer 令牌。每个代理拿到一个专用会话令牌（`PINCHTAB_SESSION`），它：

- 映射到特定 `agentId`，用于活动追踪
- 可单独撤销而不影响其他代理
- 有可配置的空闲超时和最大生命周期
- 绝不向代理暴露服务器 bearer 令牌

**重要：** 代理会话设计用于受信任环境。会话管理 API（`/sessions`）没有按代理的授权——任何经过 bearer 认证的调用者都能管理所有会话。不要把这些端点暴露给不受信任的网络。

配置与 API 细节见 [Reference: Agent Sessions](../reference/sessions.md)。

## 相关指南

- [cloakbrowser.md](cloakbrowser.md)——浏览器相关的指纹 flag 以及 CloakBrowser 二进制的许可证/分发策略
- [attach-chrome.md](attach-chrome.md)——通过 CDP 附加外部受管的 Chrome 或 CloakBrowser，以及上文概述的 attach 策略字段
- [docker.md](docker.md)——容器部署、仅无头设计以及本地 CloakBrowser 冒烟镜像
- [headed-mode.md](headed-mode.md)——手动有头设置（捆绑镜像和 CI 中不受支持）
