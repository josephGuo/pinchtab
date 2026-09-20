# 多实例

PinchTab 可以同时运行多个隔离的 Chrome 实例。每个运行中的实例都有自己的浏览器进程、端口、标签页和基于 Profile 的状态。

## 心智模型

- Profile 是存储在磁盘上的浏览器状态
- 实例是运行中的 Chrome 进程
- 一个 Profile 一次最多可以有一个活动的受管实例
- 标签页属于实例，标签页 ID 应被视为 API 返回的不透明值

## 启动编排器

```bash
pinchtab server
```

默认情况下，编排器监听 `http://localhost:9867`。下面的每个 API 调用都需要服务器令牌；示例假定已执行 `export PINCHTAB_TOKEN=$(pinchtab config token --stdout)`，并以 `-H "Authorization: Bearer $PINCHTAB_TOKEN"` 发送（为简洁起见在下文省略）。CLI 替代方案会自动取到令牌。

## 启动实例

当你需要可预测的多实例行为时，使用显式实例 API：

```bash
curl -X POST http://localhost:9867/instances/start \
  -H "Content-Type: application/json" \
  -d '{"mode":"headed","port":"9999"}'
# CLI Alternative
pinchtab instance start --mode headed --port 9999
# Response
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
- 如果你省略 `profileId`，PinchTab 会创建一个带有自动生成的 Profile 名称的受管实例。
- `securityPolicy.allowedDomains` 让你可以仅为该实例放宽 IDPI/域名信任。它与服务器的 `security.allowedDomains` 合并，因此在一个仅本地的服务器允许列表下，某个实例可以用 `["*"]`，而其余实例仍保持受限。服务器默认不设置 `allowedDomains`，所以在默认服务器上给某个实例传一个列表会把该实例限制在该列表内。
- 启动实例只在使用带自动启动行为的简写路由的工作流中才是可选的，例如默认的 `always-on` 策略或 `simple`。在 `explicit` 中，你应当假设需要自己启动一个实例。

## 在特定实例中打开标签页

```bash
curl -X POST http://localhost:9867/instances/inst_0a89a5bb/tabs/open \
  -H "Content-Type: application/json" \
  -d '{"url":"https://pinchtab.com"}'
# Response
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

`/tabs/<tabId>/metrics` 返回的是所属实例的聚合内存，而不是该标签页自己的——标签页 id 只用来选定要询问哪个实例。见[内存监控](memory-monitoring.md)。

## 重用持久 Profile

首先列出现有 Profile：

```bash
curl http://localhost:9867/profiles
```

然后为已知 Profile 启动实例：

```bash
curl -X POST http://localhost:9867/instances/start \
  -H "Content-Type: application/json" \
  -d '{"profileId":"prof_278be873","mode":"headless"}'
# CLI Alternative
pinchtab instance start --profile prof_278be873 --mode headless
```

由于一个 Profile 只能有一个活动的受管实例，因此在它已经活动时再次启动同一个 Profile 会返回错误，而不是创建重复的浏览器。

## 监控运行实例

```bash
curl http://localhost:9867/instances
curl http://localhost:9867/instances/inst_0a89a5bb
curl http://localhost:9867/instances/inst_0a89a5bb/tabs
curl http://localhost:9867/instances/metrics
```

有用的字段：

- `id`：稳定的实例标识符
- `profileId` 和 `profileName`：支持该实例的 Profile
- `port`：实例的 HTTP 端口
- `mode`：用于请求/响应对称性的显式 `"headless"` 或 `"headed"` 字符串
- `headless`：Chrome 是否以无头模式启动
- `status`：通常是 `starting`、`running`、`stopping` 或 `stopped`

## 停止实例

```bash
curl -X POST http://localhost:9867/instances/inst_0a89a5bb/stop
# CLI Alternative
pinchtab instance stop inst_0a89a5bb
# Response
{
  "id": "inst_0a89a5bb",
  "status": "stopped"
}
```

停止实例会释放其端口。如果 Profile 是持久的，其浏览器状态会保留在磁盘上。

## 端口分配

如果你不传递端口，PinchTab 会从配置的范围中分配一个：

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
- 你希望同时使用单独的有头和无头浏览器
- 你需要稳定的 Profile 到实例的所有权规则
- 你正在构建永远不应该依赖于隐式自动启动的工具