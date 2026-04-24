# 配置

`pinchtab config` 是创建、检查、验证和编辑 PinchTab 配置文件的 CLI 入口点。

有关安全状态、令牌使用、敏感端点策略和 IDPI 指南，请参阅 [安全](../guides/security.md)。

## 命令

### `pinchtab config`

打开交互式配置概览/编辑器。

它当前直接暴露这些高信号设置：

- `multiInstance.strategy`
- `multiInstance.allocationPolicy`
- `instanceDefaults.stealthLevel`
- `instanceDefaults.tabEvictionPolicy`

它还显示：

- 活动配置文件路径
- 服务器运行时的仪表板 URL
- 掩码服务器令牌
- `Copy token` 操作

```bash
pinchtab config
```

### `pinchtab config init`

在当前配置路径创建默认配置文件。

```bash
pinchtab config init
```

`config init` 尊重 `PINCHTAB_CONFIG`。如果设置了该环境变量，文件将在那里创建。

### `pinchtab config show`

显示有效的运行时配置。

```bash
pinchtab config show
```

秘密值（如 `server.token`）在此输出中保持掩码状态。

### `pinchtab config token`

将配置的 `server.token` 复制到系统剪贴板，而不将其打印到 stdout。

```bash
pinchtab config token
```

如果剪贴板访问不可用，命令会安全地报告这一点，并且仍然不会打印令牌。

### `pinchtab config path`

打印 PinchTab 将读取的配置文件路径。

```bash
pinchtab config path
```

### `pinchtab config validate`

验证当前配置文件。

```bash
pinchtab config validate
```

### `pinchtab config get`

从文件配置中读取单个点路径值。

```bash
pinchtab config get server.port
pinchtab config get instanceDefaults.mode
pinchtab config get security.attach.allowHosts
```

### `pinchtab config set`

在文件配置中设置单个点路径值。

```bash
pinchtab config set server.port 8080
pinchtab config set instanceDefaults.mode headed
pinchtab config set multiInstance.strategy explicit
```

### `pinchtab config patch`

将 JSON 对象合并到配置文件中。

```bash
pinchtab config patch '{"server":{"port":"8080"}}'
pinchtab config patch '{"instanceDefaults":{"mode":"headed","maxTabs":50}}'
pinchtab config patch '{"observability":{"activity":{"retentionDays":14}}}'
```

## 加载顺序

PinchTab 按以下顺序应用配置：

1. 内置默认值
2. 由 `PINCHTAB_CONFIG` 或默认路径选择的配置文件
3. `PINCHTAB_TOKEN`（如果设置），在运行时覆盖 `server.token`

支持的环境变量：

- `PINCHTAB_CONFIG`：选择配置文件路径
- `PINCHTAB_TOKEN`：在运行时覆盖 API 令牌

对于远程 CLI 目标，使用根 `--server` 标志而不是配置。

## 配置文件位置

按 OS 的默认位置：

- macOS：`~/.pinchtab/config.json`
- Linux：`~/.pinchtab/config.json`
- Windows：`%APPDATA%\pinchtab\config.json`

在 macOS 和 Linux 上，PinchTab 默认使用 `~/.pinchtab`，因此 CLI、npm 管理的二进制文件和配置文件都使用相同的基本目录。

使用以下命令覆盖配置路径：

```bash
export PINCHTAB_CONFIG=/path/to/config.json
```

## 配置形状

当前嵌套文件配置形状：

```json
{
  "configVersion": "0.8.0",
  "server": {
    "port": "9867",
    "bind": "127.0.0.1",
    "token": "your-secret-token",
    "stateDir": "/path/to/state",
    "engine": "chrome",
    "networkBufferSize": 100,
    "trustProxyHeaders": false,
    "cookieSecure": null
  },
  "browser": {
    "version": "144.0.7559.133",
    "binary": "/path/to/chrome",
    "extraFlags": "--disable-gpu",
    "extensionPaths": ["/path/to/pinchtab/extensions"]
  },
  "instanceDefaults": {
    "mode": "headless",
    "noRestore": false,
    "timezone": "Europe/Rome",
    "blockImages": false,
    "blockMedia": false,
    "blockAds": false,
    "maxTabs": 20,
    "maxParallelTabs": 0,
    "userAgent": "",
    "noAnimations": false,
    "stealthLevel": "light",
    "tabEvictionPolicy": "close_lru",
    "dialogAutoAccept": false
  },
  "security": {
    "allowEvaluate": false,
    "allowMacro": false,
    "allowScreencast": false,
    "allowDownload": false,
    "allowedDomains": ["127.0.0.1", "localhost", "::1"],
    "downloadAllowedDomains": [],
    "downloadMaxBytes": 20971520,
    "allowUpload": false,
    "allowClipboard": false,
    "uploadMaxRequestBytes": 10485760,
    "uploadMaxFiles": 8,
    "uploadMaxFileBytes": 5242880,
    "uploadMaxTotalBytes": 10485760,
    "maxRedirects": -1,
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
      "customPatterns": [],
      "scanTimeoutSec": 5,
      "shieldThreshold": 30
    }
  },
  "profiles": {
    "baseDir": "/path/to/profiles",
    "defaultProfile": "default"
  },
  "multiInstance": {
    "strategy": "always-on",
    "allocationPolicy": "fcfs",
    "instancePortStart": 9868,
    "instancePortEnd": 9968,
    "restart": {
      "maxRestarts": 20,
      "initBackoffSec": 2,
      "maxBackoffSec": 60,
      "stableAfterSec": 300
    }
  },
  "timeouts": {
    "actionSec": 30,
    "navigateSec": 60,
    "shutdownSec": 10,
    "waitNavMs": 1000
  },
  "autoSolver": {
    "enabled": false,
    "maxAttempts": 8,
    "solvers": ["cloudflare", "semantic", "capsolver", "twocaptcha"],
    "llmProvider": "",
    "llmFallback": false,
    "external": {
      "capsolverKey": "",
      "twoCaptchaKey": ""
    }
  },
  "scheduler": {
    "enabled": false,
    "strategy": "fair-fifo",
    "maxQueueSize": 1000,
    "maxPerAgent": 100,
    "maxInflight": 20,
    "maxPerAgentInflight": 10,
    "resultTTLSec": 300,
    "workerCount": 4
  },
  "observability": {
    "activity": {
      "enabled": true,
      "sessionIdleSec": 1800,
      "retentionDays": 1,
      "events": {
        "dashboard": false,
        "server": false,
        "bridge": false,
        "orchestrator": false,
        "scheduler": false,
        "mcp": false,
        "other": false
      }
    }
  }
}
```

`autoSolver.external` 仅在配置文件中。Capsolver 和 2Captcha 凭证存储在那里。

仪表板设置页面公开非秘密的 AutoSolver 设置，并显示活动配置文件路径。提供商密钥仍然直接在配置文件中管理。

`browser.extraFlags` 经过验证和清理。它仅用于用户安全的 Chrome 标志，这些标志不会削弱浏览器安全性，也不会覆盖 PinchTab 拥有的启动行为。

被拒绝的示例包括：

- `--no-sandbox`
- `--disable-web-security`
- `--ignore-certificate-errors`
- `--user-agent=...`
- `--enable-automation=...`
- `--disable-blink-features=...`

改用专用配置字段：

- `instanceDefaults.userAgent` 用于 UA 覆盖
- `instanceDefaults.mode` 用于有头/无头
- `instanceDefaults.timezone` 用于时区
- `browser.extensionPaths` 用于扩展加载
- `browser.remoteDebuggingPort` 用于远程调试端口

对于 Linux 容器兼容性，使用运行时管理的路径而不是 `browser.extraFlags`。PinchTab 在需要时自动启用 `--no-sandbox`。

默认情况下，PinchTab 在 `<server.stateDir>/extensions` 中查找未打包的 Chrome 扩展。在正常的本地安装中，这意味着特定于 OS 的 PinchTab 配置目录加上 `extensions/`，例如：

- macOS: `~/.pinchtab/extensions`
- Linux: `~/.pinchtab/extensions`
- Windows: `%APPDATA%\pinchtab\extensions`

您可以使用 `browser.extensionPaths` 更改或清除该默认值。

## 部分

| 部分 | 目的 |
| --- | --- |
| `server` | HTTP 服务器设置、引擎选择、代理信任和网络缓冲区默认值 |
| `browser` | Chrome 可执行文件、版本固定、额外标志和扩展路径 |
| `instanceDefaults` | 管理实例的默认行为 |
| `security` | 敏感功能门、传输限制、附加策略和 IDPI |
| `profiles` | 配置文件存储默认值 |
| `multiInstance` | 编排器策略、分配、端口范围和重启策略 |
| `timeouts` | 操作、导航、关闭和导航等待延迟 |
| `scheduler` | 可选任务队列 |
| `observability` | 活动日志记录、源选择和保留 |

## `config get` 和 `config set` 支持

`pinchtab config get` 和 `pinchtab config set` 仅支持这些顶级部分：

- `server`
- `browser`
- `instanceDefaults`
- `security`
- `profiles`
- `multiInstance`
- `timeouts`
- `observability`

它们不暴露这些部分中的每个字段，也不支持 `scheduler.*`。

对于以下字段，使用 `pinchtab config patch` 或直接编辑 `config.json`：

- `server.engine`
- `server.networkBufferSize`
- `browser.extensionPaths`
- `instanceDefaults.dialogAutoAccept`
- `security.allowClipboard`
- `security.idpi.scanTimeoutSec`
- `security.idpi.shieldThreshold`
- `scheduler.*`
- `observability.activity.events.*`

## 常见示例

### 有头模式

```json
{
  "instanceDefaults": {
    "mode": "headed"
  }
}
```

### 带令牌的网络绑定

```bash
pinchtab config set server.bind 0.0.0.0
pinchtab config set server.token secret
pinchtab server
```

将 `server.bind` 更改为非环回是记录的、非默认的、降低安全性的部署更改。仅当远程可达性是有意的时使用它，保持令牌设置，并明确审查外部网络边界。

如果仪表板通过非环回绑定上的纯 HTTP 提供服务，PinchTab 会显示产品内警告，因为会话 cookie 不再通过传输加密。尽可能使用 HTTPS 或 localhost。

### 仪表板 Cookie 传输

`server.cookieSecure` 控制仪表板会话 cookie 是否必须使用 `Secure` 标志：

- `null` / 未设置 / `auto`：默认行为。会话 cookie 在 HTTPS 上是 `Secure`，在纯 HTTP 上是非 `Secure`。
- `true`：始终要求 `Secure`。仪表板登录仅在 HTTPS 上工作。
- `false`：始终省略 `Secure`，即使在 HTTPS 上也是如此。仅用于操作员管理的边缘情况。

示例：

```bash
pinchtab config set server.cookieSecure true
pinchtab config set server.cookieSecure false
pinchtab config set server.cookieSecure auto
```

当 `server.cookieSecure = true` 时，纯 HTTP 仪表板登录会明确失败，显示 HTTPS 要求错误，而不是看起来成功并循环。

如果 TLS 在 PinchTab 前面终止，当代理受信任并且正确重写 `Forwarded` / `X-Forwarded-*` 头时，也设置 `server.trustProxyHeaders=true`。

### 自定义实例端口范围

```json
{
  "multiInstance": {
    "instancePortStart": 8100,
    "instancePortEnd": 8200
  }
}
```

### 附加策略

```json
{
  "security": {
    "attach": {
      "enabled": true,
      "allowHosts": ["127.0.0.1", "localhost", "chrome.internal"],
      "allowSchemes": ["ws", "wss", "http", "https"]
    }
  }
}
```

`security.attach.allowHosts` 是一个允许列表。如果您将其设置为 `["*"]`，PinchTab 接受任何具有允许方案的可访问附加主机。这是记录的、非默认的、降低安全性的覆盖：它完全删除了主机允许列表，仅应在隔离的、操作员控制的网络上使用。

### 活动保留

```json
{
  "observability": {
    "activity": {
      "retentionDays": 14,
      "sessionIdleSec": 1800
    }
  }
}
```

`server.trustProxyHeaders` 应保持 `false`，除非 PinchTab 在信任的反向代理后面，该代理覆盖 `Forwarded` 和 `X-Forwarded-*` 头。不要在直接暴露部署或通过未更改的客户端提供的转发头的代理后面启用它。

## 旧版扁平格式

较旧的扁平配置仍然接受以向后兼容：

```json
{
  "port": "9867",
  "headless": true,
  "maxTabs": 20,
  "allowEvaluate": false,
  "timeoutSec": 30,
  "navigateSec": 60
}
```

使用 `pinchtab config init` 创建当前的嵌套格式。

## 验证

`pinchtab config validate` 检查，除其他外：

- 有效的 `instanceDefaults.mode`
- 有效的 `instanceDefaults.stealthLevel`
- 有效的 `instanceDefaults.tabEvictionPolicy`
- `instanceDefaults.maxTabs >= 1`
- `instanceDefaults.maxParallelTabs >= 0`
- 有效的 `multiInstance.strategy`
- 有效的 `multiInstance.allocationPolicy`
- 有效的 `multiInstance.restart.*` 值
- 有效的 `security.attach.allowSchemes`
- `multiInstance.instancePortStart <= multiInstance.instancePortEnd`
- `multiInstance.restart.initBackoffSec <= multiInstance.restart.maxBackoffSec`
- 非负超时值
- 非负 `server.networkBufferSize`
- 非负 `security.idpi.scanTimeoutSec`
- 正 `observability.activity.sessionIdleSec` 和 `retentionDays`

有效的枚举值：

| 字段 | 值 |
| --- | --- |
| `instanceDefaults.mode` | `headless`, `headed` |
| `instanceDefaults.stealthLevel` | `light`, `medium`, `full` |
| `instanceDefaults.tabEvictionPolicy` | `reject`, `close_oldest`, `close_lru` |
| `multiInstance.strategy` | `simple`, `explicit`, `simple-autorestart`, `always-on`, `no-instance` |
| `multiInstance.allocationPolicy` | `fcfs`, `round_robin`, `random` |
| `security.attach.allowSchemes` | `ws`, `wss`, `http`, `https` |

## 注意事项

- `config show` 报告有效的运行时值，而不仅仅是原始文件内容。
- `config get`、`set` 和 `patch` 操作文件配置模型，而不是瞬态运行时覆盖。
- 仪表板配置 API 将 `server.token` 视为只写；使用 CLI 或文件编辑来管理它。