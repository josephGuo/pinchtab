# 安全

PinchTab 设计为默认在本地机器上使用，除非您明确开启，否则不会暴露高风险的浏览器控制功能。

PinchTab 的默认和主要部署模型是本地优先：一个用户，一台机器，一个操作员控制的浏览器控制平面。支持更复杂的拓扑，如 Docker、LAN 访问、远程桥接或分布式编排器设置，但这些是高级部署。PinchTab 不应被视为现成的面向互联网的服务，确保这些部署的安全是操作员的责任。

如果您在不同的机器上运行 PinchTab，只有在您理解您正在操作的安全模型时才这样做。首选私有或其他封闭网络，避免将服务直接暴露给公共互联网，并保持高风险功能禁用，除非该部署需要它们。如果必须启用它们，限制它们，以便只有需要它们的最小受信任系统才能访问它们。

> [!WARNING]
> PinchTab 的仪表板、HTTP API、远程 CLI 目标、MCP 集成和自动化路由都属于同一个特权控制平面。它们仅适用于受信任的操作员和受信任的系统。不要将它们暴露给不受信任的用户、不受信任的客户端系统或公共互联网。
>
> 如果您不确定非本地或部分暴露的部署是否安全，请暂时不要暴露它。先查看本指南，并在继续之前使用 `SECURITY.md` 中的私人安全联系路径。

默认安全状态是：

- `server.bind = 127.0.0.1`
- `server.token` 在默认设置期间生成，应保持设置
- `security.allowEvaluate = false`
- `security.allowMacro = false`
- `security.allowScreencast = false`
- `security.allowDownload = false`
- `security.allowUpload = false`
- `security.attach.enabled = false`
- `security.attach.allowHosts = ["127.0.0.1", "localhost", "::1"]`
- `security.attach.allowSchemes = ["ws", "wss"]`
- `security.allowedDomains = ["127.0.0.1", "localhost", "::1"]`
- `security.trustedProxyCIDRs = []`
- `security.trustedResolveCIDRs = []`
- `security.idpi.enabled = true`
- `security.idpi.strictMode = true`
- `security.idpi.scanContent = true`
- `security.idpi.wrapContent = true`

使用 `pinchtab security` 查看当前状态并恢复推荐的默认值。

## 安全理念

PinchTab 遵循几个简单的规则：

- 默认仅本地访问
- 默认关闭危险功能
- 将传输访问与功能暴露分开
- 当无法建立内容或域信任时，默认拒绝

这意味着有两个独立的问题：

1. 谁可以访问服务器
2. 服务器被访问后被允许做什么

两者都很重要。

## 信任边界

重要的操作规则很简单：

- 如果一个人或系统不应该被允许控制浏览器状态、配置文件、配置、附件或敏感端点系列，它就不应该能够访问 PinchTab，也不应该获得 PinchTab 的凭据

这包括：

- 浏览器仪表板
- 直接 HTTP API 客户端
- 使用 `--server` 针对远程服务器的 CLI 使用
- MCP 客户端、插件、脚本和其他构建在 API 之上的自动化层

这些是同一控制平面的不同接口，不是单独的信任域。

## 高级部署

如果您有意运行超出默认本地设置的 PinchTab，操作员的最低清单是：

- 将 `server.token` 设置为强随机值
- 通过受信任的网络边界、VPN、防火墙或反向代理缩小网络可达性
- 当流量离开本地机器时，在代理或传输层添加 TLS
- 仅当受信任的反向代理实际为您剥离和重建 `Forwarded` / `X-Forwarded-*` 头时，才启用 `server.trustProxyHeaders`
- 保持敏感端点系列禁用，除非明确需要它们，如果启用，将它们限制在必须访问它们的最小受信任调用者或网络路径
- 为您正在操作的远程拓扑故意设置 `security.attach` 和 `security.idpi`

这些选择是部署责任，不是 PinchTab 可以代表您安全推断的默认值。

当服务器不在用户或代理所在的同一台机器上运行时，标准应该更高：知道哪些主机可以访问它，知道哪些凭据保护它，知道哪些端点系列已启用，以及知道哪些网络边界包含它。

绑定到环回减少了谁可以访问 API。令牌减少了谁可以成功使用它。敏感端点门减少了成功调用者可以做什么。IDPI 减少了哪些网站和提取的内容足够受信任，可以更深入地传递到代理工作流中。

## API 令牌

`server.token` 是主 API 令牌。

对于非浏览器客户端，请求应发送：

```http
Authorization: Bearer <token>
```

浏览器仪表板使用不同的流程：

1. 用户在登录页面上输入令牌一次
2. 服务器将其交换为同源 `HttpOnly` 会话 cookie
3. 敏感的仪表板操作可能需要短期提升的令牌重新输入

默认情况下，PinchTab 自动检测仪表板会话 cookie 是否应使用 `Secure` 标志。在 `auto` 模式下，HTTPS 请求获取 `Secure` cookie，而纯 HTTP 请求则不。

这意味着：

- 反向代理的 HTTPS 保持 `Secure` 启用
- 纯 `http://localhost:9867` 保持仅本地使用
- 纯 `http://192.168.x.x:9867` 或 `http://10.x.x.x:9867` 工作，但仪表板警告会话在不安全的 HTTP 上运行

如果您想要求 HTTPS 进行仪表板登录，强制 `server.cookieSecure` 为 `true`：

```json
{
  "server": {
    "cookieSecure": true
  }
}
```

在纯 HTTP 上，这现在会明确失败，显示 HTTPS 要求的登录错误，而不是看起来成功然后循环。

如果您有意在受信任的 LAN 上需要纯 HTTP，您也可以明确强制 `cookieSecure` 关闭：

```json
{
  "server": {
    "cookieSecure": false
  }
}
```

推荐用法：

- 除非有理由覆盖，否则保持 `cookieSecure` 未设置（`auto`）
- 当 TLS 在 PinchTab 前面时使用 `cookieSecure: true`
- 仅在操作员控制的纯 HTTP 部署上使用 `cookieSecure: false`
- 如果 TLS 在受信任的反向代理终止，启用 `server.trustProxyHeaders`，以便识别转发的 HTTPS 请求

为什么这很重要：

- 没有令牌，任何可以访问服务器的进程都可以调用 API
- 在 `127.0.0.1` 上，这仍然包括本地脚本、浏览器页面、同一机器上的其他用户和恶意软件
- 在 `0.0.0.0` 或 LAN 绑定上，缺少令牌是更大的风险

推荐做法：

- 保持 `server.bind` 在 `127.0.0.1`
- 设置强随机 `server.token`
- 仅当远程访问是有意的时才扩大绑定

`pinchtab config init` 在默认设置过程中生成并存储令牌：

```bash
pinchtab config init
```

仪表板设置页面不暴露或轮换 `server.token`。使用 `pinchtab config token` 复制当前令牌，或如果 `server.token` 为空，让 `pinchtab security` 恢复或创建一个。

如果您手动调用 API：

```bash
curl -H "Authorization: Bearer <token>" http://127.0.0.1:9867/health
```

CLI 命令默认使用配置的本地服务器设置，`PINCHTAB_TOKEN` 可以覆盖单个 shell 会话的令牌。

## 代理会话

代理会话是受信任自动化的减少分发凭据，而不是不受信任客户端的沙箱。

- 会话认证的调用者被阻止访问仪表板/管理端点系列，如配置、会话管理、配置文件管理、实例管理、仪表板代理列表和缓存控制
- 会话记录可以选择性地携带明确的授权，进一步缩小访问范围
- 默认情况下，没有明确授权的会话仍然可以使用正常的非管理自动化 API

这意味着代理会话适合受控环境，其中调用者已经被信任驱动浏览器自动化，但不应接收完整的仪表板承载令牌。它们不足以用于敌对的多租户共享或公共互联网暴露。对于那种隔离，在单独的网络和凭据边界后面运行单独的 PinchTab 实例。

## 敏感端点

一些端点系列比正常的导航和检查暴露更多的权力。PinchTab 默认保持它们禁用：

- `security.allowEvaluate`
- `security.allowMacro`
- `security.allowScreencast`
- `security.allowDownload`
- `security.allowUpload`

为什么它们被认为是危险的：

- `evaluate` 可以在页面上下文中执行 JavaScript
- `macro` 可以触发更高级的自动化流程
- `screencast` 可以流式传输实时页面内容
- `download` 可以获取并持久化远程内容。当设置 `security.downloadAllowedDomains` 时，列出的域绕过私有 IP SSRF 检查（用于内部主机，如 Docker 服务）。`["*"]` 匹配每个主机并禁用下载端点上的所有私有 IP 保护。
- `upload` 可以将本地文件推入浏览器流程

这些与身份验证不同。

- 身份验证决定谁可以调用 API
- 敏感端点门决定哪些高风险功能存在

例如，带有 `security.allowEvaluate = true` 的令牌保护服务器仍然有意向任何拥有令牌的调用者暴露 JavaScript 执行。

禁用时，这些路由被锁定并返回 `403`，解释该端点系列在配置中被禁用。

## 附加策略

附加是通过 CDP URL 注册外部管理的 Chrome 实例的高级功能。默认情况下它被禁用：

```json
{
  "security": {
    "attach": {
      "enabled": false,
      "allowHosts": ["127.0.0.1", "localhost", "::1"],
      "allowSchemes": ["ws", "wss"]
    }
  }
}
```

如果您启用附加：

- 保持 `allowHosts` 范围狭窄
- 除非外部 Chrome 目标或远程桥接是有意的，否则首选仅本地主机
- 仅附加到您信任的浏览器和 CDP 端点
- `allowHosts: ["*"]` 是记录的、非默认的、降低安全性的覆盖。它完全禁用主机允许列表，并允许任何具有允许方案的可访问附加主机。仅在隔离的、操作员控制的网络上使用。

如果您使用 `POST /instances/attach-bridge`，`security.attach.allowSchemes` 还必须包含 `http` 或 `https`。

当 `allowHosts` 包含 `"*"` 时，`security.attach.allowSchemes` 和 `security.attach.enabled` 仍然适用，但在该配置中主机允许列表不再提供保护。

对于 `attach-bridge`，`baseUrl` 应该是裸桥接源，例如 `http://bridge.internal:9868`。不要包含凭据、查询字符串、片段或路径。

## IDPI

IDPI 代表间接提示注入防御。

它的存在是为了减少不受信任的网站内容通过隐藏指令、中毒文本或不安全导航影响下游代理的机会。

PinchTab 的 IDPI 层目前做四件事：

- 将导航限制为已批准域的允许列表
- 当 URL 无法与该允许列表匹配时阻止或警告
- 扫描提取的内容以查找可疑的提示注入模式
- 包装文本输出，以便下游系统可以将其视为不受信任的内容

默认的仅本地 IDPI 配置是：

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

- 如果 `allowedDomains` 为空，主要域限制不会做有用的工作
- 如果 `allowedDomains` 包含 `"*"`，白名单实际上允许一切
- `security.allowedDomains` 是规范配置路径。加载旧配置文件时仍然接受 `security.idpi.allowedDomains`，但新保存被规范化为 `security.allowedDomains`
- `strictMode = true` 阻止不允许的域和可疑内容
- `strictMode = false` 允许请求但发出警告
- `scanContent` 保护 `/text` 和 `/snapshot` 样式的提取路径
- `wrapContent` 为下游消费者添加明确的不受信任内容框架
- 将导航扩大到非本地或非受信任站点仍然是降低安全性的选择；IDPI 降低风险，但它不会使敌对页面安全或移除浏览器攻击面

对于导航信任覆盖：

- `security.trustedResolveCIDRs` 允许主机名在导航预检期间解析为非公共 IP。这适用于操作员控制的 DNS 或代理设置，如内部代理、实验室网络或基准测试范围
- `security.trustedProxyCIDRs` 在运行时导航检查期间信任来自已知内部代理的浏览器报告的远程 IP
- 保持两个列表狭窄。广泛的范围如 `10.0.0.0/8` 会减少 SSRF 保护，仅当整个网络段被有意信任时才应使用

支持的域模式是：

- 精确主机：`example.com`
- 子域通配符：`*.example.com`
- 完全通配符：`*`

`*` 很方便，但它会破坏主要的允许列表防御，除非您故意禁用域限制，否则应避免使用。

如果您只需要为一个管理的浏览器扩大信任，优先使用实例范围的覆盖，而不是更改全局服务器策略。`POST /instances/start`、`POST /instances/launch` 和 `POST /profiles/{id}/start` 接受：

```json
{
  "securityPolicy": {
    "allowedDomains": ["*"]
  }
}
```

该覆盖仅对该实例是附加的。例如，您可以保持服务器基线仅本地，并启动一个带有 `allowedDomains: ["*"]` 或狭窄的额外主机列表（如 `["wikipedia.org"]`）的临时实例，而不扩大服务器的其余部分。

## 推荐配置

对于安全的本地设置：

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
    "allowUpload": false,
    "allowedDomains": ["127.0.0.1", "localhost", "::1"],
    "trustedProxyCIDRs": [],
    "trustedResolveCIDRs": [],
    "attach": {
      "enabled": false,
      "allowHosts": ["127.0.0.1", "localhost", "::1"],
      "allowSchemes": ["ws", "wss"]
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

如果您有意将 PinchTab 暴露在 localhost 之外，将令牌视为必需，并保持敏感端点系列禁用，除非您有特定理由启用它们。对于任何比单机本地设置更暴露的内容，假设您正在操作高级部署，并明确审查每个安全控制。

## 代理会话

对于自动化代理，使用**代理会话**而不是共享服务器承载令牌。每个代理获得一个专用的会话令牌 (`PINCHTAB_SESSION`)，它：

- 映射到特定的 `agentId` 用于活动跟踪
- 可以单独撤销而不影响其他代理
- 具有可配置的空闲超时和最大生命周期
- 从不向代理暴露服务器承载令牌

**重要：**代理会话设计用于受信任的环境。会话管理 API (`/sessions`) 没有每个代理的授权 — 任何承载认证的调用者都可以管理所有会话。不要将这些端点暴露给不受信任的网络。

有关配置和 API 详细信息，请参阅 [参考：代理会话](../reference/sessions.md)。