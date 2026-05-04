# Docker 部署

PinchTab 可以在 Docker 中运行，使用挂载的数据卷来存储配置、配置文件和状态。
捆绑的镜像现在在 `/data/.config/pinchtab/config.json` 下管理其默认配置。
如果您想完全控制配置文件路径，您仍然可以挂载自己的文件并将 `PINCHTAB_CONFIG` 指向它。

## 快速开始

从这个仓库构建镜像：

```bash
docker build -t pinchtab .
```

使用持久数据卷运行容器：

```bash
docker run -d \
  --name pinchtab \
  -p 127.0.0.1:9867:9867 \
  -v pinchtab-data:/data \
  --shm-size=2g \
  pinchtab
```

在首次启动时，镜像会创建 `/data/.config/pinchtab/config.json`，其中包含 `bind: 0.0.0.0`（Docker 端口发布所需），并在需要时生成令牌。

如果您从 Docker 内部检查启动安全摘要，环回绑定检查仍会报告有效的运行时绑定为非环回。这是预期的：进程在容器内部监听 `0.0.0.0`，以便 Docker 端口发布可以将流量转发给它。

这并不自动意味着服务暴露在您的机器之外。主机暴露仍然取决于您如何发布容器端口。例如：

- `-p 127.0.0.1:9867:9867` 使服务只能从主机机器访问
- `-p 9867:9867` 在主机的网络接口上暴露它

将 Docker 运行时绑定和主机发布的地址视为单独的层。如果您将 PinchTab 暴露在 localhost 之外，请保持设置身份验证令牌，并将其放在 TLS 或受信任的反向代理后面。

## 健康检查和就绪状态

PinchTab 在 Docker 中有两阶段的就绪模型：

1. **仪表板就绪**：`/health` 返回 HTTP 200 — 服务器进程已启动
2. **浏览器就绪**：`/health` 响应具有 `defaultInstance.status == "running"` — Chrome 已就绪

### 为什么有两个阶段？

使用 `always-on` 策略（默认），PinchTab 在启动时启动管理的 Chrome 实例。仪表板立即变得健康，但 Chrome 需要几秒钟才能初始化。如果您的应用在 Chrome 就绪之前访问 `/navigate` 或 `/snapshot`，它会获得 HTTP 503。

### Docker Compose 健康检查

标准健康检查在仪表板响应时将容器标记为"健康"：

```yaml
healthcheck:
  test: ["CMD-SHELL", "wget -q -O /dev/null http://localhost:9867/health"]
  interval: 3s
  timeout: 10s
  retries: 20
  start_period: 15s
```

这对于容器编排是正确的 — Docker 知道进程是活动的，服务是可达的。

### 应用级就绪状态

如果您的应用需要 Chrome 在发出请求之前就绪，请轮询 `/health` 并检查 `defaultInstance.status`：

```bash
# 等待浏览器就绪
until curl -sf http://localhost:9867/health | jq -e '.defaultInstance.status == "running"' > /dev/null 2>&1; do
  sleep 1
done
echo "Browser ready"
```

或在代码中：

```javascript
async function waitForBrowser(baseUrl, timeoutMs = 60000) {
  const start = Date.now();
  while (Date.now() - start < timeoutMs) {
    try {
      const res = await fetch(`${baseUrl}/health`);
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

有关完整详细信息，请参阅 [健康参考](../reference/health.md)。

## 提供您自己的 `config.json`

如果您想自己管理配置文件，请挂载它并将 `PINCHTAB_CONFIG` 指向它：

```text
docker-data/
└── config.json
```

示例 `docker-data/config.json`：

```json
{
  "server": {
    "bind": "0.0.0.0",
    "port": "9867",
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

使用显式配置文件运行：

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
curl http://localhost:9867/health
curl http://localhost:9867/instances
```

## 要持久化的内容

如果您希望数据在容器重启后仍然存在，请持久化：

- 管理的配置目录或您挂载的配置文件
- 配置文件目录
- 状态目录

没有挂载卷，配置文件和保存的会话状态是临时的。

## 运行时配置

支持的环境变量：

- `PINCHTAB_CONFIG` — 自定义配置文件的路径（如果不使用管理的配置）
- `PINCHTAB_TOKEN` — 身份验证令牌（首选 Docker 密钥；见下文）

其他所有内容，包括绑定地址和端口，都应在 `config.json` 中设置。

### 关于容器中的 `bind: 0.0.0.0`

入口点在首次启动时在配置中设置 `bind: 0.0.0.0`。这是必要的，因为 Docker 端口发布要求进程在容器内部监听 `0.0.0.0`。

示例：`docker run -p 127.0.0.1:9867:9867` 使 PinchTab 只能从您的主机机器访问，即使进程在内部监听 `0.0.0.0`。

### Docker 密钥（敏感配置）

对于生产部署，使用 Docker 密钥而不是环境变量来设置 `PINCHTAB_TOKEN`：

```bash
# 创建密钥
echo "your-secret-token" | docker secret create pinchtab_token -

# 在 docker-compose.yml 中使用它
services:
  pinchtab:
    image: pinchtab/pinchtab
    secrets:
      - pinchtab_token
    environment:
      PINCHTAB_TOKEN_FILE: /run/secrets/pinchtab_token
    # ... 其余配置
```

或使用 `docker run`：

```bash
docker run -d \
  --name pinchtab \
  --secret pinchtab_token \
  -e PINCHTAB_TOKEN_FILE=/run/secrets/pinchtab_token \
  pinchtab/pinchtab
```

密钥以只读方式挂载，永远不会出现在 `docker ps` 或日志中。

## Compose

仓库包含一个 `docker-compose.yml`，它遵循管理配置模式：

1. 挂载持久的 `/data` 卷
2. 让入口点创建和维护 `/data/.config/pinchtab/config.json`
3. 可选地传递 `PINCHTAB_TOKEN`

如果您更喜欢完全用户管理的配置文件，请单独挂载它并设置 `PINCHTAB_CONFIG`。

如果您将 PinchTab 暴露在 localhost 之外，请设置身份验证令牌并将其放在 TLS 或受信任的反向代理后面。

## 安全

### 容器中禁用 Chrome 沙箱

PinchTab 在容器中使用 `--no-sandbox` 运行 Chrome。这是标准做法，因为：

- **用户命名空间不可用**：容器没有 Chrome 沙箱所需的完整命名空间隔离
- **容器安全补偿**：Docker 镜像使用：
  - `cap_drop: ALL`（无权限）
  - `read_only: true`（不可变文件系统）
  - `seccomp` 默认配置文件（系统调用过滤）
  - 非根用户
- **容器层的隔离**：容器运行时（cgroups、seccomp、AppArmor/SELinux）提供安全边界

此配置被主要的无头浏览器服务（Puppeteer、Playwright、Browserless）使用。

PinchTab 在运行时管理这种兼容性。不要在 `browser.extraFlags` 中放入 `--no-sandbox`。

## 资源说明

容器中的 Chrome 通常需要：

- 更大的共享内存，例如 `--shm-size=2g`
- 足够的 RAM 用于您的标签页数量和工作负载

对于更重的抓取或测试工作负载，还考虑：

- 降低 `instanceDefaults.maxTabs`
- 在配置中设置阻止选项，如 `blockImages`
- 运行多个较小的容器，而不是一个过大的浏览器

## 容器中的多实例

您可以在一个容器内运行编排器模式并从 API 启动管理实例，但许多团队更喜欢每个容器一个浏览器服务，因为：

- 生命周期更简单
- 容器级资源限制更清晰
- 重启行为更容易推理

根据您是想要容器级隔离还是 PinchTab 管理的多实例编排来选择。