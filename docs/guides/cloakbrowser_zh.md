# CloakBrowser

使用本指南来让 PinchTab 配合一个 CloakBrowser Chromium 二进制运行。PinchTab
仍然是 API 与自动化控制平面，CloakBrowser 则提供浏览器可执行文件和原生指纹补丁。

```text
agent -> PinchTab API -> PinchTab bridge -> CloakBrowser Chromium
```

## PinchTab 支持的内容

PinchTab 可以启动用户自行安装的 CloakBrowser 二进制，支持：

- `browsers.default=cloak`
- `browser.binary=/absolute/path/to/cloakbrowser/chrome`
- `browser.cloak` 下的结构化指纹设置
- PinchTab Profile 目录、标签页生命周期、截图、快照、动作、下载、上传、剪贴板以及 evaluate 接口
- `/stealth/status` 上报当前活动浏览器、原生 Cloak 模式以及 PinchTab 覆盖层状态

当 `browser.cloak.disableDefaultStealthArgs=true` 时，PinchTab 会禁用与其重叠的
JavaScript 隐身覆盖层以及隐藏自动化特征的启动 flag。PinchTab 仍然控制浏览器生命周期、用户数据目录、无头/有头模式、远程调试、标签页上限以及动作拟人化。

PinchTab 不会在发布二进制、npm 包、Homebrew formula 或默认发布的 Docker 镜像中捆绑 CloakBrowser。

### 控制台日志捕获

CloakBrowser 的原生补丁会在首次导航之后抑制 CDP `Runtime` 域事件（启用该域是众所周知的机器人检测手段），因此 PinchTab 在 Chrome 上使用的标准 `Runtime.consoleAPICalled` 流永远无法送达页面控制台输出。PinchTab 通过浏览器能力注册表检测到这一点，转而在每个标签页上附加第二个只读 CDP 会话，监听已废弃但仍可用的 `Console` 域。控制台日志（`/console`、审计的 `consoleLogs`）以及未捕获的页面错误（`/errors`）通过该回退机制在 Cloak 上工作。该机制不会向页面注入任何内容，因此不增加任何指纹面。

## 许可证

CloakBrowser 有两个相关部分：

- 包装器包与源代码
- 编译后的 CloakBrowser Chromium 二进制

包装器源代码采用 MIT 许可证。编译后的 Chromium 二进制有单独的许可证。除非许可证明确允许该用途，否则不要在 PinchTab 发布产物或公开 Docker 镜像中转售（vendor）、复制或再分发该二进制。

PinchTab 对 CloakBrowser 的分发策略是：

- PinchTab **不**在发布产物中捆绑任何 CloakBrowser 二进制——不在 `./dev binaries` 中，不在 npm 包中，不在 Homebrew formula 中，不在发布的 `pinchtab/pinchtab` 或 `ghcr.io/pinchtab/pinchtab` Docker 镜像中。
- 用户自行在磁盘上提供 CloakBrowser 二进制，并将 `browser.binary`（或 `browser.targets.<name>.binary`）指向它。
- 本地 Docker 文件 `tests/tools/docker/cloakbrowser-smoke.Dockerfile` 仅用于本地冒烟测试。它不会被推送到任何仓库，也不是发布产物。
- 目前没有公开的捆绑 CloakBrowser 的 PinchTab 发布镜像，也没有托管服务。在与上游 CloakBrowser 维护者明确再分发、OEM 或 SaaS 条款之前，暂无此类计划。

安全的默认做法是：自行安装或下载 CloakBrowser；若它无法被自动发现，再把 PinchTab 指向该本地二进制路径。当省略 `browsers.default` 时（包括新生成的配置），PinchTab 会自动优先使用可发现的本地 CloakBrowser 安装。显式选择 Chrome 的现有配置仍继续使用 Chrome。

请查阅：

- [CloakBrowser 仓库](https://github.com/CloakHQ/CloakBrowser)
- [CloakBrowser 二进制许可证](https://github.com/CloakHQ/CloakBrowser/blob/main/BINARY-LICENSE.md)

## 安装 CloakBrowser

通过其官方包路径之一安装 CloakBrowser，然后把报告出的 Chromium 二进制路径用作 PinchTab 的 `browser.binary`。

Python：

```bash
pip install cloakbrowser
python -m cloakbrowser install
python -m cloakbrowser info
```

JavaScript：

```bash
npm install cloakbrowser playwright-core
node -e "import('cloakbrowser').then(async m => { await m.ensureBinary(); console.log(m.binaryInfo()); })"
```

使用安装程序报告的 `chrome` 可执行文件的绝对路径。

## 配置 PinchTab

创建或更新你的 PinchTab 配置：

```bash
pinchtab config init
pinchtab config set browsers.default cloak      # CLI: --browser=cloak
pinchtab config set browser.binary /absolute/path/to/cloakbrowser/chrome
pinchtab config set browser.cloak.fingerprintSeed 42069
pinchtab config set browser.cloak.platform windows
pinchtab config set browser.cloak.timezone Europe/London
pinchtab config set browser.cloak.locale en-GB
```

配置片段示例：

```json
{
  "browsers": {
    "default": "cloak"
  },
  "browser": {
    "binary": "/absolute/path/to/cloakbrowser/chrome",
    "cloak": {
      "fingerprintSeed": "42069",
      "platform": "windows",
      "timezone": "Europe/London",
      "locale": "en-GB",
      "webrtcIP": "auto",
      "disableDefaultStealthArgs": true
    }
  },
  "instanceDefaults": {
    "mode": "headless",
    "humanize": true
  }
}
```

当从同一 Profile 回访同一站点时，请使用固定的 `browser.cloak.fingerprintSeed`。除非你有意要把 PinchTab 旧版隐身行为叠加到 CloakBrowser 原生补丁之上，否则请保持 `browser.cloak.disableDefaultStealthArgs` 为 `true`。

## Cloak 选项

PinchTab 将以下 `browser.cloak` 字段映射到 CloakBrowser 启动 flag：

| 配置字段 | 启动 flag |
| --- | --- |
| `fingerprintSeed` | `--fingerprint=<seed>` |
| `platform` | `--fingerprint-platform=<windows|macos|linux>` |
| `timezone` | `--fingerprint-timezone=<iana timezone>` |
| `locale` | `--fingerprint-locale=<locale>` |
| `webrtcIP` | `--fingerprint-webrtc-ip=<ip|auto>` |
| `fontsDir` | `--fingerprint-fonts-dir=<path>` |
| `storageQuotaMB` | `--fingerprint-storage-quota=<mb>` |

仅当高级 CloakBrowser flag 没有对应的结构化 PinchTab 字段时，才使用 `browser.extraFlags`。

```json
{
  "browsers": {
    "default": "cloak"
  },
  "browser": {
    "binary": "/absolute/path/to/cloakbrowser/chrome",
    "cloak": {
      "fingerprintSeed": "42069",
      "timezone": "Europe/London",
      "locale": "en-GB"
    },
    "extraFlags": "--fingerprint-brand=Chrome"
  }
}
```

对原始 flag 要保守使用。PinchTab 掌管生命周期 flag，例如远程调试端口、用户数据目录、无头模式、窗口尺寸、扩展路径和 user agent。

## 启动 PinchTab

启动服务器：

```bash
pinchtab server
```

在另一个 shell 中，检查服务器是否健康并能驱动页面：

```bash
pinchtab health
pinchtab nav https://example.com --snap
```

如果你的安全策略禁止公网导航，请在使用公网 URL 之前，把确切的主机加入 `security.allowedDomains`。

## 用 `pinchtab doctor` 验证配置

在启动服务器之前，运行这个只读诊断命令。它会检查已配置的二进制、短暂执行它，并在不改动任何状态的情况下验证指纹 flag：

```bash
pinchtab doctor                       # human-readable report
pinchtab doctor --json                # machine-readable JSON
pinchtab doctor browser cloak-eu      # scope to one browser.targets entry
pinchtab doctor --check binary_exists # run a single check by name
```

退出码：`0` 表示全部检查通过或跳过，`1` 表示至少一项检查失败，`2` 表示用法错误（例如未知的检查名）。如何应对具体失败，请参见下方[故障排查](#故障排查)一节。

## 验证 CloakBrowser 是否生效

在受管浏览器实例启动后，检查 `/stealth/status`：

```bash
TOKEN="$(pinchtab config token --stdout)"
curl -sS -H "Authorization: Bearer ${TOKEN}" \
  http://127.0.0.1:9867/stealth/status
```

关注以下字段：

```json
{
  "provider": "cloak",
  "native": true,
  "pinchtabOverlaysDisabled": true,
  "fingerprintSeed": "42069"
}
```

如果启动失败，检查该实例：

```bash
pinchtab instance list
pinchtab instance logs <instance-id>
```

常见原因：

- 配置的二进制路径不存在
- 该二进制不可执行
- 下载的 CloakBrowser 二进制适用于另一个操作系统
- Profile 目录被另一个浏览器进程锁定
- 某个原始 extra flag 与 PinchTab 掌管的生命周期行为冲突

## Docker

发布的 `pinchtab/pinchtab` 和 `ghcr.io/pinchtab/pinchtab` 镜像包含 Alpine Chromium，不包含 CloakBrowser。

要在本地 Docker 中配合 CloakBrowser 使用，请构建专用的本地镜像：

```bash
docker build \
  -f tests/tools/docker/cloakbrowser-smoke.Dockerfile \
  -t pinchtab-cloakbrowser:local \
  .
```

创建一个配置，把 PinchTab 指向该镜像内的 CloakBrowser 二进制：

```json
{
  "server": {
    "bind": "0.0.0.0",
    "port": "9867",
    "token": "replace-me",
    "stateDir": "/data"
  },
  "browsers": {
    "default": "cloak"
  },
  "browser": {
    "binary": "/opt/cloakbrowser/chrome",
    "cloak": {
      "fingerprintSeed": "42069",
      "platform": "linux",
      "timezone": "UTC",
      "locale": "en-US",
      "disableDefaultStealthArgs": true
    }
  },
  "instanceDefaults": {
    "mode": "headless",
    "humanize": true
  },
  "profiles": {
    "baseDir": "/data/profiles",
    "defaultProfile": "default"
  }
}
```

以只读方式挂载配置来运行容器：

```bash
docker run -d \
  --name pinchtab-cloak \
  -p 127.0.0.1:9867:9867 \
  --shm-size=2g \
  -v pinchtab-cloak-data:/data \
  -v "$PWD/pinchtab-cloak.json":/config/pinchtab.json:ro \
  -e PINCHTAB_CONFIG=/config/pinchtab.json \
  pinchtab-cloakbrowser:local
```

除非 CloakBrowser 二进制许可证明确允许在你的使用场景下再分发，否则不要发布或推送捆绑 CloakBrowser 的镜像。

## 维护者验证

项目维护者可以用以下命令验证 Docker 集成：

```bash
./dev e2e --browser=cloak
```

该命令会把运行时浏览器切换为 `browsers.default=cloak`（`/opt/cloakbrowser/chrome`、`fingerprintSeed=42069`、`disableDefaultStealthArgs=true`），从 `tests/tools/docker/cloakbrowser-smoke.Dockerfile` 构建 `pinchtab-cloakbrowser:test` 镜像（设置 `SKIP_BUILD=1` 可复用已有镜像），断言 `/stealth/status` 报告 `provider=cloak, native=true, pinchtabOverlaysDisabled=true, fingerprintSeed=42069`，并针对 cloak 容器运行完整 E2E 套件。

使用 `./dev smoke --browser=cloak` 运行完整的本地 CloakBrowser 冒烟集。仅需要专门的对等校验环节时使用 `./dev smoke cloakbrowser`，包括 `--multi-target`、`--profile-persistence` 和 `--profile-lock-recovery`。

## 附加到一个正在运行的 CloakBrowser

如果 CloakBrowser 已经在启用远程调试的情况下运行（例如在 CloakBrowser Manager 下），你可以通过常规 CDP 附加路径把它接入 PinchTab。PinchTab 会在外部端点周围派生一个子进程 `pinchtab bridge --cdp-attach ...` 包装器，并在本地端口上暴露标准 API。

```bash
curl -X POST http://localhost:9867/instances/attach \
  -H "Authorization: Bearer $(pinchtab config token --stdout)" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "cloak-manager-profile",
    "cdpUrl": "ws://127.0.0.1:9222/devtools/browser/abc123",
    "browser": "cloak"
  }'
```

远程附加 CloakBrowser 时：

- `browser`（或其别名 `provider`）接受提供者名 `cloak`，而不是 target 名；当配置了 `browser.targets` 时，必须存在一个 provider 为 `cloak` 的 target；省略该值则以默认 target 的 provider 进行附加
- 既无 `browser`/`provider` 也无 targets 时，附加默认为 `chrome`；若同时传两个字段，二者必须一致
- `/stealth/status` 报告 `provider=cloak`、`launchMode=remote-cdp`、`native=true`、`pinchtabOverlaysDisabled=true`——PinchTab 不注入其 JS 指纹覆盖层，前提是假定外部浏览器自行负责原生指纹行为
- 停止被附加的 PinchTab 实例只会关闭包装器 bridge；外部 CloakBrowser 进程保持运行

可选的冒烟测试 `./dev smoke cdp-attach` 会针对本地构建的 CloakBrowser 镜像演练这条路径。

## 故障排查

大多数 CloakBrowser 启动失败都可归入以下某一类。请先运行 `pinchtab doctor`——它会把失败定位到某个具名检查，省去猜测。

### 缺少二进制

症状：`pinchtab doctor` 报告 `binary_exists: FAIL`，或服务器在启动实例时日志打印 `stat /path/to/cloakbrowser/chrome: no such file or directory`。

修复：把 `browser.binary`（或具名 target 的 `browser.targets.<name>.binary`）设为一个存在且可执行的绝对路径：

```bash
pinchtab config set browser.binary /absolute/path/to/cloakbrowser/chrome
# or, for a named target:
pinchtab config set browser.targets.cloak-eu.binary /absolute/path/to/cloakbrowser/chrome
```

自动发现只会在 `$PATH` 上寻找名为 `cloakbrowser` 的可执行文件，然后查找 `/opt/cloakbrowser/chrome` 和 `~/.cloakbrowser/chrome`（Linux 与 macOS）。把 `chrome` 放到别处的安装程序不会被发现，因此请使用 `python -m cloakbrowser info` 或 `cloakbrowser.binaryInfo()` 报告的绝对路径。

### 不受支持的平台构建

症状：`pinchtab doctor` 以非零退出码、段错误或 `exec format error` 报告 `binary_starts: FAIL`。

修复：CloakBrowser 二进制是平台相关的（CPU 架构和操作系统）。Linux x86_64 二进制无法在 macOS arm64 上运行，反之亦然。请为你实际运行的平台重新安装 CloakBrowser，并参阅上游 CloakBrowser 发布说明中支持的构建矩阵。

PinchTab 冒烟镜像（`tests/tools/docker/cloakbrowser-smoke.Dockerfile`）按主机架构构建；如果你在 arm64 上构建，再尝试在 x86_64 上运行所得镜像（或反过来），内部的 CloakBrowser 二进制会以同样方式失败。

### Profile 锁冲突

症状：启动失败并提示 `profile in use`，或 PinchTab 记录一条警告称 Profile 锁被持有。

当没有任何运行中的 PinchTab 拥有该 Profile 时，PinchTab 的陈旧 `SingletonLock` 恢复机制会自动清除锁：

- 若有运行中的 PinchTab 拥有该 Profile，PinchTab 拒绝抢占它，并记录 `browser profile lock appears active and owned by another pinchtab; leaving singleton files in place`
- 若锁的 PID 仍存活但其 PinchTab 拥有者已消失，PinchTab 记录 `browser profile lock appears active but pinchtab owner is dead; proceeding with stale cleanup`
- 若浏览器进程仍持有该 Profile 且无 PinchTab 拥有者，PinchTab 记录 `browser profile lock appears active but no pinchtab owner found; killing stale processes`，然后移除 singleton 文件

若恢复未自动发生，请在容器或主机日志中查找这些 `browser profile lock` 行：

```bash
docker logs pinchtab-cloak | grep "profile lock"
```

手动恢复（仅当你确信没有真实的 CloakBrowser 进程正对着该 Profile 运行时）：

```bash
rm -f /path/to/profile/{SingletonLock,SingletonCookie,SingletonSocket}
```

### 错误的指纹 / 代理 / 时区组合

症状：尽管 `/stealth/status` 报告 `provider=cloak, native=true`，目标站点仍把会话标记为自动化。

CloakBrowser 的指纹 flag 与代理地理位置必须**自洽**。把美国住宅代理与 `cloak.timezone=Europe/London`、`cloak.locale=en-GB` 配对，对任何会交叉校验客户端 locale 与出口 IP 的检测器来说都很可疑。请让它们对齐：

- 让 `browser.proxy.geo.timezone` 与代理的实际位置匹配
- 让 `browser.proxy.geo.locale` 与 `browser.cloak.locale` 匹配
- 让 `browser.cloak.platform` 与该 locale 下可信的设备匹配
- 若你设置了 `browser.cloak.webrtcIP`，确保它不会泄露主机真实的局域网网段

如果你无法手动设置代理的地理位置，请去掉 `browser.proxy.geo` 块，让 PinchTab 在没有地理位置覆盖的情况下使用代理——缺失一个信号总比自相矛盾的信号好。

代理地理位置对齐仅适用于 CloakBrowser。普通 Chrome target 保持其正常的 Chrome 启动行为，不会收到由代理推导的 `--lang`、`TZ` 或 WebRTC flag。

### 检测站点仍标记为自动化

症状：即使启用了 CloakBrowser 原生隐身，机器人检测演示页面（CreepJS、BrowserScan、Sannysoft 等）仍报告非零检测分。

这种情况预期会有波动。检测分是一个移动靶：检测器启发式会在 PinchTab 发布之外变化，CloakBrowser 发布新补丁，Chromium 版本漂移，且各站点信号独立于任何单一组件演化。

要排查，请对已配置的浏览器重新运行实时检测冒烟测试：

```bash
./dev smoke live-detection                       # Chrome leg (baseline)
./dev smoke live-detection --browser=chrome
./dev smoke live-detection --browser=cloak
```

该冒烟测试会把各站点截图和提取出的摘要保存到 `./tests/e2e/results/live-detection/<browser>-<timestamp>/`。把 cloak 运行结果与 chrome 基线对比：若 cloak 的非零分数与 Chrome 基线一致，则表明这是一个通用 Chromium 信号，而非 CloakBrowser 回归。在上游 CloakBrowser 或检测器更新后重新运行冒烟测试，而不是把某一次运行当作权威结论。

实时检测冒烟测试的结果仅供参考，绝不会作为 CI 的门禁。

## 相关指南

- [docker.md](docker.md)——本地 CloakBrowser 冒烟镜像与持久化配置卷
- [attach-chrome.md](attach-chrome.md)——通过 CDP 附加到外部托管的 CloakBrowser
- [headed-mode.md](headed-mode.md)——手动有头模式设置（捆绑镜像与冒烟镜像均仅限无头）
- [security.md](security.md)——底层安全模型、附加策略、IDPI 与令牌处理

## 安全

CloakBrowser 改变的是浏览器指纹行为，它不改变 PinchTab 的安全模型。

请遵守以下规则：

- 除非你有意运维一个加固过的远程部署，否则保持 PinchTab 绑定 localhost
- 当服务可从 localhost 之外访问时，务必设置 `server.token`
- 保持 `security.allowedDomains` 收窄
- 不要公网暴露 CDP 端点
- 不要对你不拥有或未获授权测试的系统进行自动化

如果目标站点阻止自动化，请把它视为一个策略与可靠性信号，而不仅仅是技术障碍。
