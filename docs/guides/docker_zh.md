# Docker 部署

PinchTab 可以在 Docker 中运行，用一个挂载的数据卷存放配置、Profile 和状态。
捆绑镜像把默认配置管理在 `/data/.pinchtab/config.json` 下（镜像设置了 `HOME=/data`，而 PinchTab 在 Linux 上把配置放在 `~/.pinchtab`）。
如果你想完全控制配置文件路径，仍然可以挂载自己的文件并把 `PINCHTAB_CONFIG` 指向它。

## 快速开始

从本仓库构建镜像：

```bash
docker build -t pinchtab .
```

用持久数据卷运行容器：

```bash
docker run -d \
  --name pinchtab \
  -p 127.0.0.1:9867:9867 \
  -v pinchtab-data:/data \
  --shm-size=2g \
  pinchtab
```

首次启动时，镜像会创建带 `bind: 0.0.0.0`（Docker 端口发布所需）的 `/data/.pinchtab/config.json`，并在需要时生成一个令牌。用 `docker exec pinchtab pinchtab config token --stdout` 读取它；每次 API 调用（包括 `/health`）都需要以 `Authorization: Bearer <token>` 携带它。

如果你从 Docker 内部检查启动安全摘要，环回绑定检查仍会把生效的运行时绑定报告为非环回。这是预期行为：进程在容器内监听 `0.0.0.0`，以便 Docker 端口发布能把流量转发给它。

这并不自动意味着服务被暴露到你机器之外。主机暴露仍然取决于你如何发布容器端口。例如：

- `-p 127.0.0.1:9867:9867` 让服务只能从主机本机访问
- `-p 9867:9867` 把它暴露到主机的网络接口上

把 Docker 运行时绑定与主机发布地址视为两层。如果你把 PinchTab 暴露到 localhost 之外，请保持设置认证令牌，并把它放在 TLS 或受信任的反向代理之后。

## 健康检查与就绪状态

PinchTab 在 Docker 中有两阶段就绪模型：

1. **仪表板就绪**：`/health` 返回 HTTP 200——服务器进程已起来
2. **浏览器就绪**：`/health` 响应里 `defaultInstance.status == "running"`——Chrome 就绪

### 为什么有两个阶段？

在 `always-on` 策略（默认）下，PinchTab 启动时会拉起一个受管 Chrome 实例。仪表板立即变健康，但 Chrome 需要几秒钟初始化。如果你的应用在 Chrome 就绪之前请求 `/navigate` 或 `/snapshot`，会得到 HTTP 503。

### Docker Compose 健康检查

镜像内置的 `HEALTHCHECK` 运行 `pinchtab health`，它从容器配置中读取令牌，并在仪表板响应时把容器标记为「healthy」。在 Compose 中等价写法为：

```yaml
healthcheck:
  test: ["CMD-SHELL", "pinchtab health >/dev/null"]
  interval: 3s
  timeout: 10s
  retries: 20
  start_period: 15s
```

这对容器编排是正确的——Docker 知道进程存活且服务可达。

### 应用级就绪

如果你的应用需要 Chrome 就绪后再发请求，轮询 `/health` 并检查 `defaultInstance.status`：

```bash
# Wait for browser to be ready
until curl -sf -H "Authorization: Bearer $PINCHTAB_TOKEN" http://localhost:9867/health | jq -e '.defaultInstance.status == "running"' > /dev/null 2>&1; do
  sleep 1
done
echo "Browser ready"
```

或在代码中：

```javascript
async function waitForBrowser(baseUrl, token, timeoutMs = 60000) {
  const start = Date.now();
  while (Date.now() - start < timeoutMs) {
    try {
      const res = await fetch(`${baseUrl}/health`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      const data = await res.json();
      if (data.defaultInstance?.status === "running") return;
    } catch {}
    await new Promise(r => setTimeout(r, 1000));
  }
  throw new Error("Browser not ready within timeout");
}
```

### 完整健康响应（服务器模式）

```json
{
  "status": "ok",
  "mode": "dashboard",
  "version": "0.8.0",
  "uptime": 12345,
  "authRequired": true,
  "profiles": 1,
  "instances": 1,
  "defaultInstance": {
    "id": "inst_abc12345",
    "status": "running"
  },
  "agents": 0,
  "restartRequired": false
}
```

完整细节见 [健康参考](../reference/health.md)。

## 提供你自己的 `config.json`

如果你想自己管理配置文件，挂载它并把 `PINCHTAB_CONFIG` 指向它：

```text
docker-data/
└── config.json
```

`docker-data/config.json` 示例：

```json
{
  "server": {
    "bind": "0.0.0.0",
    "port": "9867",
    "token": "replace-with-a-generated-token",
    "stateDir": "/data/state"
  },
  "profiles": {
    "baseDir": "/data/profiles",
    "defaultProfile": "default"
  },
  "instanceDefaults": {
    "mode": "headless",
    "noRestore": true
  }
}
```

PinchTab 不会向你通过 `PINCHTAB_CONFIG` 提供的文件里写入令牌，因此要么像上面那样设置 `server.token`，要么传 `-e PINCHTAB_TOKEN=...`；两者都没有时，服务器拒绝启动。

用显式配置文件运行：

```bash
docker run -d \
  --name pinchtab \
  -p 127.0.0.1:9867:9867 \
  -e PINCHTAB_CONFIG=/config/config.json \
  -v "$PWD/docker-data:/data" \
  -v "$PWD/docker-data/config.json:/config/config.json:ro" \
  --shm-size=2g \
  pinchtab
```

检查它：

```bash
curl -H "Authorization: Bearer $PINCHTAB_TOKEN" http://localhost:9867/health
curl -H "Authorization: Bearer $PINCHTAB_TOKEN" http://localhost:9867/instances
```

## 要持久化什么

如果你希望数据在容器重启后存活，请持久化：

- 受管配置目录或你挂载的配置文件
- Profile 目录
- 状态目录

没有挂载卷，Profile 和已保存的会话状态就是临时的。

## 运行时配置

支持的环境变量：

- `PINCHTAB_CONFIG`——自定义配置文件路径（不使用受管配置时）
- `PINCHTAB_TOKEN`——认证令牌（推荐用 Docker secrets；见下文）

其他一切，包括绑定地址和端口，都应放在 `config.json` 里。

### 关于容器中的 `bind: 0.0.0.0`

entrypoint 会在首次启动时在配置里设置 `bind: 0.0.0.0`。这是必要的，因为 Docker 端口发布要求进程在容器内监听 `0.0.0.0`。

例如：`docker run -p 127.0.0.1:9867:9867` 让 PinchTab 只能从你的主机本机访问，即使进程内部监听 `0.0.0.0`。

### Docker secrets（敏感配置）

镜像只从环境变量消费 `PINCHTAB_TOKEN`——它不读取 `PINCHTAB_TOKEN_FILE` 这种间接方式。要使用 Docker secrets，请在你自己的 entrypoint 或 wrapper 中先 source 该 secret 文件，再运行镜像：

```bash
# Create a secret
echo "your-secret-token" | docker secret create pinchtab_token -

# Use it in docker-compose.yml — read the secret file and export PINCHTAB_TOKEN
services:
  pinchtab:
    image: pinchtab/pinchtab
    secrets:
      - pinchtab_token
    entrypoint: ["/bin/sh", "-c"]
    command:
      - |
        export PINCHTAB_TOKEN="$(cat /run/secrets/pinchtab_token)"
        exec /usr/local/bin/docker-entrypoint.sh pinchtab server
```

挂载在 `/run/secrets/...` 的 secrets 是只读的，绝不会出现在 `docker ps` 或日志中。

## Compose

仓库附带一个 `docker-compose.yml`，遵循受管配置模式：

1. 挂载一个持久的 `/data` 卷
2. 让 entrypoint 创建并维护 `/data/.pinchtab/config.json`
3. 可选地传入 `PINCHTAB_TOKEN`

如果你偏好完全由用户管理的配置文件，请单独挂载它并设置 `PINCHTAB_CONFIG`。

如果你把 PinchTab 暴露到 localhost 之外，请设置认证令牌并把它放在 TLS 或受信任的反向代理之后。

## 安全

### 容器中禁用 Chrome 沙箱

PinchTab 在容器中以 `--no-sandbox` 运行 Chrome。这是标准做法，因为：

- **用户命名空间不可用**：容器不具备 Chrome 沙箱所需的完整命名空间隔离
- **容器安全作为补偿**：Docker 镜像使用：
  - `cap_drop: ALL`（无 capability）
  - `read_only: true`（不可变文件系统）
  - `seccomp` 默认 profile（系统调用过滤）
  - 非 root 用户
- **容器层的隔离**：容器运行时（cgroups、seccomp、AppArmor/SELinux）提供安全边界

这一配置被主流无头浏览器服务（Puppeteer、Playwright、Browserless）采用。

PinchTab 在运行时管理这种兼容性。不要把 `--no-sandbox` 放进 `browser.extraFlags`。

## 资源说明

容器中的 Chrome 通常需要：

- 更大的共享内存，例如 `--shm-size=2g`
- 与你的标签页数量和工作负载相称的 RAM

对于更重的抓取或测试负载，还可以考虑：

- 调低 `instanceDefaults.maxTabs`
- 在配置里设置 `blockImages` 等屏蔽选项
- 运行多个较小的容器，而不是一个过大的浏览器

## 容器中的多实例

你可以在一个容器内运行编排器模式并从 API 启动受管实例，但很多团队倾向于每个容器一个浏览器服务，因为：

- 生命周期更简单
- 容器级资源限制更清晰
- 重启行为更容易推理

根据你想要容器级隔离还是 PinchTab 管理的多实例编排来选择。

## Docker 中的 CloakBrowser（仅本地镜像）

发布的 `pinchtab/pinchtab` 和 `ghcr.io/pinchtab/pinchtab` 镜像自带原版 Chromium。它们**不**包含 CloakBrowser，PinchTab 也不发布捆绑 CloakBrowser 的镜像。原因见 [cloakbrowser.md → 许可证](cloakbrowser.md#licensing)。

用于本地测试时，仓库自带一个自建 Dockerfile，把 CloakBrowser 叠加在 PinchTab 镜像之上。它被标记为「smoke」，因为它仅用于本地冒烟测试与对等校验——不分发、不用于生产。

### 构建本地 CloakBrowser 冒烟镜像

```bash
docker build \
  -f tests/tools/docker/cloakbrowser-smoke.Dockerfile \
  -t pinchtab-cloakbrowser:local \
  .
```

该镜像仅限本地：

- 它不被推送到任何仓库
- 它不由 `./dev binaries` 产出
- `./dev smoke cloakbrowser` 每次运行都从同一个 Dockerfile 自建一份，tag 为 `pinchtab-cloakbrowser:test`；设置 `SKIP_BUILD=1` 可改为复用已有镜像

### 本地运行一个由 CloakBrowser 支撑的容器

创建一个配置，把 PinchTab 指向镜像内的 CloakBrowser 二进制（配置文件里的 `"browsers": {"default": "cloak"}` 等价于 CLI 上的 `--browser cloak`）：

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

以只读方式挂载配置，并挂上持久化 Profile 卷来运行：

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

`pinchtab-cloak-data` 命名卷在容器重启之间持有 `/data/profiles`，因此 CloakBrowser 在多次运行之间保留 cookies、本地存储和历史。删除该卷即可从零开始：

```bash
docker volume rm pinchtab-cloak-data
```

### 仅无头设计

该冒烟镜像与捆绑的 `pinchtab/pinchtab` 镜像遵循同样的仅无头设计：没有 X11、没有 Wayland、没有 Xvfb。Docker 内有头的 CloakBrowser 不是受支持的配置。如果你需要一个可见浏览器窗口来调试，请在主机上直接运行 PinchTab + CloakBrowser——手动本地设置见 [headed-mode.md](headed-mode.md)。

### 相关指南

- [cloakbrowser.md](cloakbrowser.md)——完整的 CloakBrowser 配置、指纹 flag 与故障排查
- [attach-chrome.md](attach-chrome.md)——通过 CDP 附加到外部托管的 CloakBrowser
- [headed-mode.md](headed-mode.md)——在捆绑镜像之外的手动有头设置
- [security.md](security.md)——容器与非本地部署的安全模型
