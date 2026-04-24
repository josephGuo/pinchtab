# 多实例

PinchTab 可以同时运行多个隔离的 Chrome 实例。每个运行中的实例都有自己的浏览器进程、端口、标签页和基于配置文件的状态。

## 心智模型

- 配置文件是存储在磁盘上的浏览器状态
- 实例是运行中的 Chrome 进程
- 一个配置文件一次最多可以有一个活动的管理实例
- 标签页属于实例，标签页 ID 应被视为 API 返回的不透明值

## 启动编排器

```bash
pinchtab server
```

默认情况下，编排器监听 `http://localhost:9867`。

## 启动实例

当您需要可预测的多实例行为时，使用显式实例 API：

```bash
curl -X POST http://localhost:9867/instances/start \
  -H "Content-Type: application/json" \
  -d '{"mode":"headed","port":"9999"}'
# CLI 替代方案
pinchtab instance start --mode headed --port 9999
# 响应
{
  "id": "inst_0a89a5bb",
  "profileId": "prof_278be873",
  "profileName": "instance-1741410000000",
  "port": "9999",
  "mode": "headed",
  "headless": false,
  "status": "starting"
}
```

注意：

- `POST /instances/launch` 仍然作为兼容性端点存在，但现在遵循与 `POST /instances/start` 相同的语义。
- 如果您省略 `profileId`，PinchTab 会创建一个带有自动生成的配置文件名称的管理实例。
- `securityPolicy.allowedDomains` 允许您仅为该实例扩大 IDPI/域信任。这是在服务器基线之上的添加，因此一个实例可以使用 `["*"]`，而其他实例保持默认的允许列表。
- 启动实例在使用带有自动启动行为的简写路由的工作流中是可选的，例如 `simple` 策略。在 `explicit` 中，您应该假设需要自己启动一个实例。

## 在特定实例中打开标签页

```bash
curl -X POST http://localhost:9867/instances/inst_0a89a5bb/tabs/open \
  -H "Content-Type: application/json" \
  -d '{"url":"https://pinchtab.com"}'
# 响应
{
  "tabId": "8f9c7d4e1234567890abcdef12345678",
  "url": "https://pinchtab.com",
  "title": "PinchTab"
}
```

对于后续操作，继续使用返回的 `tabId`：

```bash
curl "http://localhost:9867/tabs/<tabId>/snapshot"
curl "http://localhost:9867/tabs/<tabId>/text"
curl "http://localhost:9867/tabs/<tabId>/metrics"
```

## 重用持久配置文件

首先列出现有配置文件：

```bash
curl http://localhost:9867/profiles
```

然后为已知配置文件启动实例：

```bash
curl -X POST http://localhost:9867/instances/start \
  -H "Content-Type: application/json" \
  -d '{"profileId":"278be873adeb","mode":"headless"}'
# CLI 替代方案
pinchtab instance start --profile 278be873adeb --mode headless
```

由于一个配置文件只能有一个活动的管理实例，因此在它已经活动时再次启动同一个配置文件会返回错误，而不是创建重复的浏览器。

## 监控运行实例

```bash
curl http://localhost:9867/instances
curl http://localhost:9867/instances/inst_0a89a5bb
curl http://localhost:9867/instances/inst_0a89a5bb/tabs
curl http://localhost:9867/instances/metrics
```

有用的字段：

- `id`：稳定的实例标识符
- `profileId` 和 `profileName`：支持该实例的配置文件
- `port`：实例的 HTTP 端口
- `mode`：用于请求/响应对称性的显式 `"headless"` 或 `"headed"` 字符串
- `headless`：Chrome 是否以无头模式启动
- `status`：通常是 `starting`、`running`、`stopping` 或 `stopped`

## 停止实例

```bash
curl -X POST http://localhost:9867/instances/inst_0a89a5bb/stop
# CLI 替代方案
pinchtab instance stop inst_0a89a5bb
# 响应
{
  "id": "inst_0a89a5bb",
  "status": "stopped"
}
```

停止实例会释放其端口。如果配置文件是持久的，其浏览器状态会保留在磁盘上。

## 端口分配

如果您不传递端口，PinchTab 会从配置的范围中分配一个：

```json
{
  "multiInstance": {
    "instancePortStart": 9868,
    "instancePortEnd": 9968
  }
}
```

当实例停止时，其端口变为可重用状态。

## 何时使用显式多实例 API

当以下情况时，优先使用显式实例 API：

- 多个浏览器会话必须保持隔离
- 您希望同时使用单独的有头和无头浏览器
- 您需要稳定的配置文件到实例的所有权规则
- 您正在构建永远不应该依赖于隐式自动启动的工具