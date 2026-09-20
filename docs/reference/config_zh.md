# 配置（Config）

`pinchtab config` 是创建、查看、校验和编辑 PinchTab 配置文件的命令行界面入口。

安全态势、token 用法、敏感端点策略和 IDPI 指南见 [Security](../guides/security.md)。

## 命令

### `pinchtab config`

打印只读的配置概览。它不是编辑器；用 `config set`、`config patch` 或直接编辑文件来改值。

它显示这些高信号的生效设置：

- `multiInstance.strategy`
- `multiInstance.allocationPolicy`
- `instanceDefaults.stealthLevel`
- `instanceDefaults.tabEvictionPolicy`
- `instanceDefaults.tabPolicy.lifecycle`（启用了空闲 lifecycle 时附带上 close delay）

它还显示：

- 活动配置文件路径
- 掩码后的服务器 token
- 服务器运行时的仪表板 URL，否则为 `not running`
- `config get`、`config set`、`config show`、`config token` 和 `pinchtab security` 的提示

```bash
pinchtab config
```

### `pinchtab config init`

在当前配置路径创建默认配置文件。

```bash
pinchtab config init
```

`config init` 尊重 `PINCHTAB_CONFIG`。若设置了该环境变量，文件创建在那里。

生成的配置文件包含用于 IDE 补全和校验的 `$schema` URL。新文件会获得一个生成的 `server.token`，打印到 stderr。若该路径已存在文件，`config init` 在覆盖前会询问。

### `pinchtab config schema`

打印当前 PinchTab 构建的 JSON Schema URL。源码构建、开发构建以及没有已发布 schema 的版本使用 `main` schema URL。对于发布构建，PinchTab 使用不高于自身版本的最近已发布 schema tag，因此生成的配置绝不会指向更新版本的规则；若无此类 tag 则使用 `main`。

```bash
pinchtab config schema
```

打印内置的 schema JSON：

```bash
pinchtab config schema --print
```

### `pinchtab config show`

显示生效的运行时配置。

```bash
pinchtab config show
```

`server.token` 等密钥值在此输出中保持掩码。Security 部分还以 `Trust Loopback Proxy` 包含 `security.trustLoopbackProxy`，使代理信任态势显式可见。

### `pinchtab config token`

将配置的 `server.token` 复制到系统剪贴板。stdout 不写任何内容，因此 shell 捕获得到空而非一条消息。

```bash
pinchtab config token             # copy to clipboard; status goes to stderr
TOKEN=$(pinchtab config token --stdout)   # print the token and nothing else
```

`--stdout` 是在没有剪贴板工具的主机上读取 token 的受支持方式。不带它时 token 绝不打印。

若剪贴板不可用，命令以非零退出码退出，并在 stderr 同时指出 `--stdout` 和配置文件路径，因此脚本看到的是失败而非 "复制了空内容却成功"。

### `pinchtab config path`

打印 PinchTab 将读取的配置文件路径。

```bash
pinchtab config path
```

### `pinchtab config validate`

校验当前配置文件。

```bash
pinchtab config validate
```

### `pinchtab config get`

读取单个点路径值，并报告**生效中的值**：文件设置了什么，否则运行时解析出的值 —— 内置默认值，或从另一个 key 派生的值。`profiles.baseDir` 是最清晰的例子：把它从文件中略去，`config get` 报告 `<server.stateDir>/profiles`，即 profiles 实际存放的目录，而非空。一个真正未设置的 key—— 没人配置过的密钥、如 `browser.binary` 这样的可选覆盖 —— 仍回答空。

```bash
pinchtab config get server.port
pinchtab config get instanceDefaults.mode
pinchtab config get security.attach.allowHosts
pinchtab config get profiles.baseDir      # derived from server.stateDir when unset
```

### `pinchtab config set`

在文件配置中设置单个点路径值。

```bash
pinchtab config set server.port 8080
pinchtab config set instanceDefaults.mode headed
pinchtab config set multiInstance.strategy explicit
```

### `pinchtab config patch`

将一个 JSON 对象合并进配置文件。

```bash
pinchtab config patch '{"server":{"port":"8080"}}'
pinchtab config patch '{"instanceDefaults":{"mode":"headed","maxTabs":50}}'
pinchtab config patch '{"observability":{"activity":{"retentionDays":14}}}'
```

## 加载顺序

PinchTab 按以下顺序应用配置：

1. 内置默认值
2. 由 `PINCHTAB_CONFIG` 或默认路径选择的配置文件
3. `PINCHTAB_TOKEN`（若设置），在运行时覆盖 `server.token`

支持的环境变量：

- `PINCHTAB_CONFIG`：选择配置文件路径
- `PINCHTAB_TOKEN`：在运行时覆盖 API token
- `PINCHTAB_RATE_LIMIT_MAX`：每客户端每 10 秒窗口的请求上限（默认 3000，按代理驱动的快照 / 操作突发量设计）。将端口暴露到 localhost 之外时调低（例如 300）。子实例从编排器环境继承它。
- `PINCHTAB_STATE_KEY`：状态文件加密 key；设置时优先于 `security.stateEncryptionKey`

远程命令行界面目标用根 `--server` flag 而非配置。

## 配置文件位置

按操作系统的默认位置：

- macOS：`~/.pinchtab/config.json`
- Linux：`~/.pinchtab/config.json`
- Windows：`%APPDATA%\pinchtab\config.json`

在 macOS 和 Linux 上，PinchTab 默认使用 `~/.pinchtab`，因此命令行界面、npm 管理的二进制和配置文件都使用同一个基础目录。

若你从较早的 macOS 设置升级，仍在 `~/Library/Application Support/pinchtab/config.json` 有配置文件，请将其视为旧位置，迁移或合并到 `~/.pinchtab/config.json`。

用以下方式覆盖配置路径：

```bash
export PINCHTAB_CONFIG=/path/to/config.json
```

## 配置形状

当前嵌套的文件配置形状：

```json
{
  "$schema": "https://raw.githubusercontent.com/pinchtab/pinchtab/main/schema/config.json",
  "configVersion": "0.8.0",
  "server": {
    "port": "9867",
    "bind": "127.0.0.1",
    "token": "your-secret-token",
    "stateDir": "/path/to/state",
    "logLevel": "info",
    "networkBufferSize": 100,
    "retainNetworkBodies": false,
    "retainNetworkBodyMaxBytes": 262144,
    "trustProxyHeaders": false,
    "cookieSecure": null
  },
  "browser": {
    "version": "144.0.7559.133",
    "binary": "/path/to/chrome",
    "remoteDebuggingPort": null,
    "extraFlags": "--disable-gpu",
    "cloak": {
      "fingerprintSeed": "42069",
      "platform": "windows",
      "locale": "en-GB",
      "timezone": "Europe/London",
      "webrtcIP": "auto",
      "fontsDir": "/path/to/fonts",
      "storageQuotaMB": 2048,
      "disableDefaultStealthArgs": true
    },
    "extensionPaths": ["/path/to/pinchtab/extensions"]
  },
  "browsers": {
    "default": "chrome"
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
    "captureAllowActivation": true,
    "humanize": false,
    "stealthLevel": "light",
    "tabEvictionPolicy": "close_lru",
    "tabPolicy": {
      "lifecycle": "keep",
      "closeDelaySec": 300,
      "restore": false
    },
    "dialogAutoAccept": false
  },
  "security": {
    "allowEvaluate": false,
    "allowMacro": false,
    "allowScreencast": false,
    "allowDownload": false,
    "allowCookies": false,
    "allowNetworkIntercept": false,
    "allowMemory": false,
    "allowFileScheme": false,
    "allowedDomains": ["127.0.0.1", "localhost", "::1"],
    "downloadAllowedDomains": [],
    "downloadMaxBytes": 20971520,
    "memorySnapshotMaxBytes": 536870912,
    "allowUpload": false,
    "allowClipboard": false,
    "allowStateExport": false,
    "stateEncryptionKey": null,
    "uploadMaxRequestBytes": 10485760,
    "uploadMaxFiles": 8,
    "uploadMaxFileBytes": 5242880,
    "uploadMaxTotalBytes": 10485760,
    "maxRedirects": -1,
    "trustedProxyCIDRs": [],
    "trustedResolveCIDRs": [],
    "trustLoopbackProxy": false,
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
      "customPatterns": [],
      "scanTimeoutSec": 5,
      "shieldThreshold": 30
    }
  },
  "profiles": {
    "baseDir": "/path/to/profiles",
    "defaultProfile": "default",
    "quarantineKeep": 1
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
    "autoTrigger": true,
    "triggerOnNavigate": true,
    "triggerOnAction": true,
    "maxAttempts": 8,
    "solverTimeoutSec": 30,
    "retryBaseDelayMs": 500,
    "retryMaxDelayMs": 10000,
    "solvers": ["cloudflare", "semantic"],
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
    "workerCount": 4,
    "maxBatchSize": 50
  },
  "observability": {
    "activity": {
      "enabled": true,
      "sessionIdleSec": 1800,
      "retentionDays": 30,
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
  },
  "sessions": {
    "dashboard": {
      "persist": true,
      "idleTimeoutSec": 604800,
      "maxLifetimeSec": 604800,
      "elevationWindowSec": 900,
      "persistElevationAcrossRestart": false,
      "requireElevation": false
    },
    "agent": {
      "enabled": true,
      "mode": "preferred",
      "idleTimeoutSec": 1800,
      "maxLifetimeSec": 86400
    }
  }
}
```

`sessions.agent.*` 见 [Agent Identity](../guides/agent-identity.md)；`sessions.agent.mode` 接受 `off` 或 `preferred`。`browser.proxy`、`browser.targets`、`browser.defaultTarget` 和 `browser.fallbackOrder` 见下文浏览器选择。

`autoSolver.external` 仅存在于配置文件。Capsolver 和 2Captcha 凭证存放在那里。

### 语义流凭证

语义优先的自动求解流会将凭证值注入识别出的登录 / 注册 / 表单字段。在配置文件的 `autoSolver.credentials` 下配置：

```json
{
  "autoSolver": {
    "credentials": {
      "login":  { "user": "you@example.com", "password": "..." },
      "signup": { "name": "Jane Doe", "email": "you@example.com", "password": "..." },
      "form":   { "field1": "...", "field2": "...", "email": "you@example.com" }
    }
  }
}
```

注意：

- 通过直接写配置文件来编辑凭证。仪表板 config API 在读取时会脱敏（GET 返回空白），并在 PUT 字段为空时保留磁盘上的值，因此密钥绝不通过 UI 往返。
- 表单求解步骤 2 在 `form.field2` 为空时回退到 `form.email`。
- 未配置值的步骤落入纯点击流（例如无密码的登录尝试变成 "点击提交" 尝试）。

仪表板 Settings 页面暴露非秘密的 AutoSolver 设置并显示活动配置文件路径。provider key 仍直接在配置文件中管理。

### 浏览器选择

命令行界面用 `--browser <name>` 选择浏览器。配置文件中等价的字段是 `browsers.default`：

```json
{
  "browsers": { "default": "cloak" }
}
```

`browsers.default` 显式选择本地浏览器后端：

- `chrome` 走常规 Chrome/Chromium 启动路径。
- `ghost-chrome` 用轻量 fetcher 服务静态友好的读取，并在页面需要渲染时升级到 Chrome。
- `cloak` 使用发现或配置的本地 CloakBrowser Chromium 二进制。

省略 `browsers.default` 时，PinchTab 优先使用发现的本地 CloakBrowser 安装，否则回退到 Chrome。新生成的配置文件做同样选择，而显式将 `browsers.default` 设为 `chrome` 的现有配置继续使用 Chrome。

`browsers.available` 可选地限制请求可选择哪些浏览器。同一 provider 的多个命名配置（不同二进制、代理或指纹），使用 `browser.targets` 配合 `browser.defaultTarget` 和 `browser.fallbackOrder`—— 见 [Terminology](../architecture/terminology.md)。旧的 `browser.provider` 字段不再支持，在校验时被拒绝。

所选浏览器为 `cloak` 时，PinchTab 在设置了 `browser.binary` 时使用它，否则搜索其常规的本地 CloakBrowser 路径。命名的 CloakBrowser 目标必须显式设置该目标的 `binary`。PinchTab 不下载、不打包、不重新分发 CloakBrowser 二进制。

在 **macOS** 上，优先使用专用自动化浏览器而非你日常的 Google Chrome。无头启动 `/Applications/Google Chrome.app` 会让 macOS 认为 Chrome 已在运行，因此从 Dock 打开你的日常 Chrome 只会激活无窗口的自动化进程，不出现窗口（issue #583）。PinchTab 的发现现在优先 Google Chrome for Testing、Chromium，然后 Chrome Canary，仅在最后手段回退到你的主 Chrome。若只装了日常 Chrome，安装 [Google Chrome for Testing](https://developer.chrome.com/blog/chrome-for-testing)，或把 `browser.binary` 指向单独的 Chrome/Chromium 构建。`pinchtab doctor browsers` 在自动化将使用你的主 Chrome 时发出警告。

`browser.cloak` 将支持的 CloakBrowser 指纹设置映射到原生启动 flags：

- `fingerprintSeed` -> `--fingerprint`
- `platform` -> `--fingerprint-platform`
- `locale` -> `--fingerprint-locale`
- `timezone` -> `--fingerprint-timezone`
- `webrtcIP` -> `--fingerprint-webrtc-ip`
- `fontsDir` -> `--fingerprint-fonts-dir`
- `storageQuotaMB` -> `--fingerprint-storage-quota`

对 CloakBrowser 目标，`disableDefaultStealthArgs` 默认为 true。设置时，PinchTab 保留其进程、profile、标签页、扩展和动作控制行为，但不添加自己的 JS 隐身覆盖层或隐藏自动化的启动 flags。仅当你有意要在 CloakBrowser 原生补丁之上再叠加 PinchTab 旧版隐身层时才设为 false。

高级 CloakBrowser flags 若不是 PinchTab 自有的生命周期 flags，仍可通过 `browser.extraFlags` 传入。

`browser.proxy.server` 是该块其余部分的前提：`browser.proxy.username`、`password`、`bypassList` 和每个 `browser.proxy.geo.*` key 在没有它时都不起作用，因为没有可路由、可认证或对齐 geo 的代理。这些值仍会被保留 ——`config set` 写入它们，`config get` 返回它们，之后设置 server 即生效 ——PinchTab 在每次配置读写时报告这个不完整的块，直到设置 server。清除 `browser.proxy.server` 会关闭代理，其他值留在磁盘留给下一个 server。

`browser.proxy.username` 和 `password` 是例外，且与 server 无关：两者互相要求对方，因此无论你先设哪一个，`config set` 都会在单独设其中一个时中止。改为在一次写入中发送这一对：

```bash
pinchtab config patch '{"browser":{"proxy":{"username":"bob","password":"s3cret"}}}'
```

`browser.proxy.geo` 是 CloakBrowser 指纹对齐提示。当为某个 CloakBrowser 目标配置了代理 server 和 geo 块时，PinchTab 将 geo 值映射到原生 CloakBrowser 指纹 flags，除非该目标已设置对应的 `browser.cloak` 字段。原生的 `chrome` 浏览器不从代理 geo 数据推导 `--lang`、`TZ` 或 WebRTC 启动设置。

### 浏览器额外 Flags

`browser.extraFlags` 经过校验和清洗。它只用于不削弱浏览器安全、也不覆盖 PinchTab 自有启动行为的用户安全 Chrome flags。

被拒绝的示例包括：

- `--no-sandbox`
- `--disable-web-security`
- `--ignore-certificate-errors`
- `--user-agent=...`
- `--enable-automation=...`
- `--disable-blink-features=...`

改用专用配置字段：

- `instanceDefaults.userAgent` 用于 UA 覆盖
- `instanceDefaults.mode` 用于有头 / 无头
- `instanceDefaults.timezone` 用于时区
- `browser.extensionPaths` 用于扩展加载
- `browser.remoteDebuggingPort` 用于远程调试端口

对于 Linux 容器兼容性，使用运行时管理的路径而非 `browser.extraFlags`。PinchTab 在需要时自动启用 `--no-sandbox`。

默认情况下，PinchTab 在 `<server.stateDir>/extensions` 中查找未打包的 Chrome 扩展。正常本地安装下，这意味着特定于 OS 的 PinchTab 配置目录加上 `extensions/`，例如：

- macOS：`~/.pinchtab/extensions`
- Linux：`~/.pinchtab/extensions`
- Windows：`%APPDATA%\\pinchtab\\extensions`

你可以用 `browser.extensionPaths` 更改或清除该默认值。

### 标签页策略

`instanceDefaults.tabPolicy` 对标签页生命周期行为分组：

```json
{
  "instanceDefaults": {
    "tabPolicy": {
      "eviction": "close_lru",
      "lifecycle": "keep",
      "closeDelaySec": 300,
      "restore": false
    }
  }
}
```

- `eviction` 控制达到 `maxTabs` 时的行为：`close_lru`、`close_oldest` 或 `reject`。
- `lifecycle` 控制空闲生命周期行为：`keep` 禁用生命周期自动关闭，为默认；`close_idle` 在某标签页处理完一次授权的 `/text`、`/snapshot` 或 `/action` 请求后自动关闭它；`freeze_idle` 则冻结在空闲延迟内未被任何请求触及的标签页（定时器和 JavaScript 停止，页面和会话保留）；每个请求都会重置该时钟并先解冻标签页。若渲染进程不接受解冻，请求以 `503` 应答，代码 `tab_unfreeze_failed` 且 `retryable: true`；标签页和会话的当前标签页指针都保留，下一次请求重试。标签页在其上有请求运行时（包括 screencast 流）、因交接暂停时、或持有网络拦截规则时，绝不被冻结。
- `closeDelaySec` 是 `close_idle` 和 `freeze_idle` 的空闲延迟。启用任一时默认为 `300` 秒。
- `restore` 控制启动时是否恢复会话标签页。默认 `false`。

`instanceDefaults.tabEvictionPolicy` 仍为兼容而接受。新配置应使用 `instanceDefaults.tabPolicy.eviction`。

### 人性化输入

`instanceDefaults.humanize` 控制点击和输入动作默认是否走较慢的人性化路径。默认 `false`，用原始 CDP 输入保持自动化快速且确定。

调用方可按动作覆盖实例默认值，用 JSON 字段 `humanize`：

- `{"kind":"click","selector":"#submit","humanize":true}` 让单个动作走贝塞尔鼠标移动和类人延迟。
- `{"kind":"type","selector":"#name","text":"Ada","humanize":true}` 让单次 type 动作走较慢的逐字符路径。

理由：人性化输入对那些对原始输入反应不佳的页面有用，但它加入了 sleep 和多步指针移动。保持可选开启可避免默认 E2E 和代理运行中意外多出数秒开销。

## 各节

| 节 | 用途 |
| --- | --- |
| `server` | HTTP 服务器设置、日志级别、代理信任、cookie 传输和网络缓冲默认值 |
| `browser` | Chrome 可执行文件、版本固定、额外 flags、扩展路径、CloakBrowser flags、代理和命名目标 |
| `browsers` | 默认浏览器选择和可选的 `available` 允许列表 |
| `instanceDefaults` | 受管实例的默认行为 |
| `security` | 敏感功能门、传输限制、attach 策略和 IDPI |
| `profiles` | profile 存储默认值 |
| `multiInstance` | 编排器策略、分配、端口范围和重启策略 |
| `timeouts` | 动作、导航、关闭和导航等待延迟 |
| `scheduler` | 可选任务队列 |
| `observability` | 活动日志、源选择和保留 |
| `sessions` | 仪表板会话 cookie 和代理会话 |
| `autoSolver` | 挑战自动求解行为、provider key 和凭证 |

## `config get` 与 `config set` 支持

`pinchtab config get` 和 `pinchtab config set` 接受以下顶级节中的 `section.field` 点路径：

- `server`
- `browser`
- `browsers`
- `instanceDefaults`
- `security`
- `profiles`
- `multiInstance`
- `timeouts`
- `scheduler`
- `observability`
- `sessions`
- `autoSolver`

这些节中的每个叶子都可达，例外如下：

- `server.engine` 和 `browser.provider` 是已移除的设置；设置它们被拒绝
- `instanceDefaults.headless` 已被 `instanceDefaults.mode` 取代
- `browsers.config.*` 是已退役的块，被 `browser.targets` 取代
- `observability.activity.stateDir` 可读但不可设置（见活动保留）

列表值如 `security.allowedDomains` 或 `browser.extensionPaths` 以逗号分隔字符串设置。`$schema` 和 `configVersion` 是文档元数据，不是设置。

## 常见示例

### 有头模式

```json
{
  "instanceDefaults": {
    "mode": "headed"
  }
}
```

### 守护进程或自动启动服务器的日志级别

`pinchtab server --log-level` 只影响你手工启动的服务器。守护进程安装的服务器（`pinchtab daemon install`）和裸 `pinchtab nav` 自动启动的服务器都是不带 flags 启动 `pinchtab server`，因此 `server.logLevel` 是设置其阈值的唯一方式 —— 且无需改动 unit 文件。

```bash
pinchtab config set server.logLevel warn    # Warnings and errors only
pinchtab config set server.logLevel debug   # Full debug detail while diagnosing
pinchtab config get server.logLevel
```

接受值为 `debug`、`info`（默认）、`warn` 和 `error`；无法解析的值在配置加载时失败，并列出接受值。`--log-level` 仍优先于配置值。`-v` 无论如何都保留其启动横幅，但它仅在 `--log-level` 和 `server.logLevel` 都未设置时才提升级别 —— 因此持久化的 `warn` 能经住 `pinchtab server -v`，而 `--log-level debug` 就是你为单次运行覆盖它的方式。`pinchtab bridge` 读取同一个 key 并接受同一个 flag。

### 内存诊断

| 设置 | 默认值 | 效果 |
| --- | --- | --- |
| `security.allowMemory` | `false` | 启用 `POST /memory/snapshot`、`GET /memory/snapshot/{snapshotId}/summary` 和 `GET /memory/compare`（关闭时代码 `memory_disabled`）。`GET /memory` 无需能力门。 |
| `security.memorySnapshotMaxBytes` | `536870912`（512 MB） | 单个堆快照的大小上限，最大 4 GB；更大的快照以 `memory_snapshot_too_large` 中止，不留文件 |

堆快照包含页面上的每一个字符串，包括 token，这就是它单独受能力门保护的原因。见 [memory.md](memory.md)。

```bash
pinchtab config set security.allowMemory true
```

### 带 token 的网络绑定

```bash
pinchtab config set server.bind 0.0.0.0
pinchtab config set server.token secret
pinchtab server
```

将 `server.bind` 从环回改掉是文档化的、非默认的、降低安全性的部署变更。仅在远程可达是有意时使用它，保持设置 token，并明确审查外部网络边界。

若仪表板通过非环回绑定上的纯 HTTP 提供服务，PinchTab 会显示产品内警告，因为会话 cookie 不再受传输加密。可能时优先 HTTPS 或 localhost。

### 仪表板 Cookie 传输

`server.cookieSecure` 控制仪表板会话 cookie 是否必须使用 `Secure` 标志：

- `null` / 未设置 / `auto`：默认行为。会话 cookie 在 HTTPS 上为 `Secure`，纯 HTTP 上为非 `Secure`。
- `true`：始终要求 `Secure`。仪表板登录仅在 HTTPS 上工作。
- `false`：始终省略 `Secure`，即使在 HTTPS 上。仅用于操作员管理的边缘情况。

示例：

```bash
pinchtab config set server.cookieSecure true
pinchtab config set server.cookieSecure false
pinchtab config set server.cookieSecure auto
```

当 `server.cookieSecure = true` 时，纯 HTTP 仪表板登录会明确失败，给出 "需要 HTTPS" 错误，而非看似成功后循环。

若 TLS 在 PinchTab 前面终止，也仅当代理受信任且正确重写 `Forwarded` / `X-Forwarded-*` 头时才设置 `server.trustProxyHeaders=true`。

### 自定义实例端口范围

```json
{
  "multiInstance": {
    "instancePortStart": 8100,
    "instancePortEnd": 8200
  }
}
```

### Attach 策略

```json
{
  "security": {
    "attach": {
      "enabled": true,
      "allowHosts": ["127.0.0.1", "localhost", "chrome.internal"],
      "allowSchemes": ["ws", "wss", "http", "https"],
      "forwardProxyAuth": false
    }
  }
}
```

`security.attach.allowHosts` 是允许列表。若设为 `["*"]`，PinchTab 接受任何带允许 scheme 的可达 attach 主机。这是文档化的、非默认的、降低安全性的覆盖：它完全移除主机允许列表，仅应在隔离的、操作员控制的网络上使用。

`security.attach.forwardProxyAuth` 控制 PinchTab 是否可在远程 CDP attach 上发送配置的代理认证凭证。默认 `false`；仅当被 attach 的浏览器进程和 CDP 传输都受信任时才启用。

### 隔离 profile 保留

当某浏览器 profile 被证明不可用，PinchTab 将其改名搁置为 `<profile>.quarantine-<unix>` 并全新开始。PinchTab 保留每个 profile 最近一次的隔离副本，并在创建新隔离时修剪更旧的 —— 最新副本最可能与当前正在调查的问题相关，产品中没有任何东西读取更旧的。每次移除都记录路径和回收的字节数。

```json
{
  "profiles": {
    "quarantineKeep": 1
  }
}
```

`profiles.quarantineKeep` 默认为 `1`。设为 `0` 保留每个隔离 profile，这正是该设置存在前的行为。修剪只在创建新隔离时运行，从不在启动时或定时运行，因此一个不再失败的 profile 的隔离副本保持不动 —— 清理这批积压是单独的显式操作。

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

活动日志始终写入 `<server.stateDir>/activity`，因此两个实例不会共用一个日志目录。`observability.activity.stateDir` 因而不可设置：`config set` 拒绝它，`config get observability.activity.stateDir` 报告生效的派生目录。用 `server.stateDir` 移动日志。

已带该 key 的文件继续加载并继续工作。该 key 被**忽略**，说明它只是*建议*而非校验错误：PinchTab 会报告它 —— 在加载时，以及在你运行 `config set`、`config patch` 或 `config validate` 时进 stderr—— 它绝不阻塞任何事。没有要修的也没有要删的；该值就是没有效果。校验错误（如超出范围的 `server.port`）仍会阻止保存。

`server.trustProxyHeaders` 应保持 `false`，除非 PinchTab 在一个会重写 `Forwarded` 和 `X-Forwarded-*` 头的受信任反向代理后面。不要在直接暴露部署、或在原样透传客户端提供的转发头的代理后面启用它。启用时，最靠近客户端的转发地址也就是客户端身份：限流桶、并发流上限和每一行审计都以它为 key，而非代理地址。

## 旧版扁平格式

旧的扁平配置仍为向后兼容而接受：

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

用 `pinchtab config init` 创建当前的嵌套格式。

## 校验

`pinchtab config validate` 检查（包括但不限于）：

- 有效的 `instanceDefaults.mode`
- 有效的 `instanceDefaults.stealthLevel`
- 有效的 `instanceDefaults.tabEvictionPolicy`
- 有效的 `instanceDefaults.tabPolicy.eviction`
- 有效的 `instanceDefaults.tabPolicy.lifecycle`
- 非负的 `instanceDefaults.tabPolicy.closeDelaySec`
- `instanceDefaults.maxTabs >= 1`
- `instanceDefaults.maxParallelTabs >= 0`
- 有效的 `multiInstance.strategy`
- 有效的 `multiInstance.allocationPolicy`
- 有效的 `multiInstance.restart.*` 值
- 有效的 `security.attach.allowSchemes`
- `multiInstance.instancePortStart <= multiInstance.instancePortEnd`
- `multiInstance.restart.initBackoffSec <= multiInstance.restart.maxBackoffSec`
- 非负的超时值
- `server.networkBufferSize` 在 1 到 10000 之间
- `server.retainNetworkBodyMaxBytes` 在 0 到 10 MiB 之间
- `security.downloadMaxBytes`、`memorySnapshotMaxBytes` 和 `upload*` 限制在 1 到其上限之间，且 `uploadMaxFileBytes <= uploadMaxTotalBytes`
- 非负的 `security.idpi.scanTimeoutSec`
- 非负的 `observability.activity.sessionIdleSec` 和正的 `retentionDays`
- 有效的 `sessions.agent.mode` 和正的 `sessions.dashboard.*Sec` 值
- `server.engine` 和 `browser.provider` 作为已移除设置被拒绝

有效枚举值：

| 字段 | 值 |
| --- | --- |
| `instanceDefaults.mode` | `headless`, `headed` |
| `instanceDefaults.stealthLevel` | `light`, `medium`, `full` |
| `instanceDefaults.tabEvictionPolicy` | `reject`, `close_oldest`, `close_lru` |
| `instanceDefaults.tabPolicy.eviction` | `reject`, `close_oldest`, `close_lru` |
| `instanceDefaults.tabPolicy.lifecycle` | `keep`, `close_idle`, `freeze_idle` |
| `multiInstance.strategy` | `simple`, `explicit`, `simple-autorestart`, `always-on`, `no-instance` |
| `multiInstance.allocationPolicy` | `fcfs`, `round_robin`, `random` |
| `security.attach.allowSchemes` | `ws`, `wss`, `http`, `https` |
| `security.attach.forwardProxyAuth` | `true`, `false` |
| `sessions.agent.mode` | `off`, `preferred` |

## 注意事项

- `config show` 报告生效的运行时值，而非仅原始文件内容。
- `config get` 报告生效中的值，包括内置默认值和从其他 key 派生的值。`config set` 和 `patch` 写文件配置模型，不携带瞬态运行时覆盖。
- 仪表板 config API 把 `server.token` 视为只写；用命令行界面或文件编辑来管理它。
