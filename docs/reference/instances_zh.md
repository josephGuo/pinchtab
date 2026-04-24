# 实例

实例是由 PinchTab 管理的运行中 Chrome 进程。每个管理实例都有：

- 实例 ID
- 配置文件
- 端口
- 模式（`headless` 或 `headed`）
- 执行状态

一个配置文件一次最多只能有一个活动的管理实例。

## 列出实例

```bash
curl http://localhost:9867/instances
# 响应：JSON 数组（见下文）

# CLI 替代方案（默认人类可读）
pinchtab instances
# 输出：inst_0a89  9999  headed  running

pinchtab instances --json              # 完整 JSON 响应
```

`pinchtab instances` 是从 CLI 检查当前实例群的最简单方法。

响应形状：

```json
[
  {
    "id": "inst_0a89a5bb",
    "profileId": "prof_278be873",
    "profileName": "instance-1741410000000",
    "port": "9999",
    "mode": "headed",
    "headless": false,
    "status": "running",
    "securityPolicy": {
      "allowedDomains": ["127.0.0.1", "localhost", "::1", "wikipedia.org"]
    }
  }
]
```

`GET /instances` 返回一个裸 JSON 数组，而不是像 `{"instances":[...]}` 这样的信封。每个实例响应都包含 `mode`（`"headless"` 或 `"headed"`）和与旧版兼容的 `headless` 布尔值。

## 启动实例

### `POST /instances/start`

当您想通过配置文件 ID 或配置文件名称启动，或让 PinchTab 创建临时配置文件时，使用 `/instances/start`。

```bash
curl -X POST http://localhost:9867/instances/start \
  -H "Content-Type: application/json" \
  -d '{"profileId":"prof_278be873","mode":"headed","port":"9999","securityPolicy":{"allowedDomains":["wikipedia.org","wikimedia.org"]}}'
# CLI 替代方案
pinchtab instance start --profile prof_278be873 --mode headed --port 9999 --allow-domain wikipedia.org --allow-domain wikimedia.org
```

请求体：

- `profileId`：可选；接受配置文件 ID 或现有配置文件名称
- `mode`：可选；使用 `headed` 表示可见浏览器，其他任何值都被视为无头
- `port`：可选
- `securityPolicy.allowedDomains`：可选的附加实例范围 IDPI/域允许列表条目

注意：

- 如果省略 `profileId`，PinchTab 会创建一个自动生成的临时配置文件
- 如果省略 `port`，PinchTab 会从配置的实例端口范围中分配一个
- CLI 标志是 `--profile`，即使 API 字段是 `profileId`
- `securityPolicy.allowedDomains` 仅为该实例与服务器级 `security.allowedDomains` 基线合并
- 您可以在不更改服务器默认值的情况下扩展单个实例。例如，`{"securityPolicy":{"allowedDomains":["*"]}}` 使该实例不受限制，而其他实例仍使用服务器基线
- 请求提供的扩展路径被拒绝；改为在服务器上配置 `browser.extensionPaths`。默认情况下，PinchTab 使用其状态/配置文件夹下的本地 `extensions/` 目录。

### `POST /instances/launch`

`/instances/launch` 是 `/instances/start` 的兼容性别名。

```bash
curl -X POST http://localhost:9867/instances/launch \
  -H "Content-Type: application/json" \
  -d '{"profileId":"prof_278be873","mode":"headed","securityPolicy":{"allowedDomains":["wikipedia.org"]}}'
```

请求体：

- `profileId`：可选的现有配置文件 ID 或现有配置文件名称
- `mode`：可选；`headed` 或默认为无头
- `port`：可选
- `securityPolicy.allowedDomains`：可选的附加实例范围 IDPI/域允许列表条目

重要：

- `/instances/launch` 不读取 `headless` 字段。当您想要有头浏览器时，使用 `mode:"headed"`。
- `name` 在 `/instances/launch` 上不再支持。首先通过 `POST /profiles` 创建配置文件，然后使用返回的 `id` 作为 `profileId`。
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
- `headless`：为兼容性保留的布尔值

## 获取实例日志

```bash
curl http://localhost:9867/instances/inst_ea2e747f/logs
# CLI 替代方案
pinchtab instance logs inst_ea2e747f
```

响应是纯文本。在 `GET /instances/{id}/logs/stream` 也有一个 SSE 流。

## 停止实例

```bash
curl -X POST http://localhost:9867/instances/inst_ea2e747f/stop
# CLI 替代方案
pinchtab instance stop inst_ea2e747f
```

停止实例会保留配置文件，除非它是临时自动生成的配置文件。

## 按配置文件启动

您也可以从面向配置文件的路由启动实例：

```bash
curl -X POST http://localhost:9867/profiles/prof_278be873/start \
  -H "Content-Type: application/json" \
  -d '{"headless":false,"port":"9999","securityPolicy":{"allowedDomains":["wikipedia.org"]}}'
```

此路由在路径中接受配置文件 ID 或配置文件名称。与 `/instances/start` 和 `/instances/launch` 不同，其请求体使用 `headless` 而不是 `mode`。

## 在实例中打开标签页

```bash
curl -X POST http://localhost:9867/instances/inst_ea2e747f/tabs/open \
  -H "Content-Type: application/json" \
  -d '{"url":"https://pinchtab.com"}'
```

没有专用的实例范围 `tab open` CLI 命令。CLI 快捷方式是：

```bash
pinchtab instance navigate inst_ea2e747f https://pinchtab.com
```

该命令为实例打开一个空白标签页，然后导航它。

## 列出一个实例的标签页

```bash
curl http://localhost:9867/instances/inst_ea2e747f/tabs
```

## 列出所有运行实例的标签页

```bash
curl http://localhost:9867/instances/tabs
```

这是整个实例群的标签页列表面端点。它与 `GET /tabs` 不同，后者是简写或桥接范围的。

## 列出实例的指标

```bash
curl http://localhost:9867/instances/metrics
```

## 附加现有 Chrome

```bash
curl -X POST http://localhost:9867/instances/attach \
  -H "Content-Type: application/json" \
  -d '{"name":"shared-chrome","cdpUrl":"ws://127.0.0.1:9222/devtools/browser/..."}'
```

注意：

- 没有 CLI 附加命令
- 仅当在 `security.attach` 下的配置中启用时才允许附加
- `security.attach.allowHosts` 必须允许 `cdpUrl` 主机
- `allowHosts: ["*"]` 是记录的、非默认的、降低安全性的覆盖。它完全禁用主机允许列表，并允许任何具有允许方案的可访问 CDP 主机。仅在隔离的、操作员控制的网络上使用。

## 附加现有桥接

```bash
curl -X POST http://localhost:9867/instances/attach-bridge \
  -H "Content-Type: application/json" \
  -d '{
    "name":"shared-bridge",
    "baseUrl":"http://10.0.12.24:9868",
    "token":"bridge-secret-token"
  }'
```

注意：

- `baseUrl` 必须是裸桥接源；不要包含凭据、查询字符串、片段或路径
- 编排器在注册之前执行健康检查
- `security.attach.allowHosts` 必须允许桥接主机
- `allowHosts: ["*"]` 是记录的、非默认的、降低安全性的覆盖。它完全禁用主机允许列表，并允许任何具有允许方案的可访问桥接主机。仅在隔离的、操作员控制的网络上使用。