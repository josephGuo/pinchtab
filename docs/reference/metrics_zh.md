# Metrics（指标）

运行时计数器、近期失败与崩溃诊断。

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:9867/metrics
```

## 两个层次，绝不相加

服务端模式运行**两个进程**，各自维护自己的计数器：

| 层次 | 进程 | 能看到 |
|-------|---------|------|
| `frontDoor` | 拥有端口的编排器 | 认证拒绝、未路由路径、请求被转发之前的一切 |
| `instance` | 浏览器控制子进程（也是 `pinchtab bridge` 独立运行的那个） | 每一个到达浏览器的请求 |

每个响应都说明其数字描述的是哪一层：

```json
{
  "layer": "frontDoor",
  "metrics": { "requestsTotal": 41, "requestsFailed": 7, "avgLatencyMs": 3.1, "...": "..." },
  "failures": {
    "layer": "frontDoor",
    "requestsFailed": 7,
    "recent": [
      { "time": "...", "requestId": "a1b2c3", "method": "GET", "path": "/tabs",
        "status": 401, "type": "http_error", "layer": "frontDoor" }
    ]
  },
  "crashes": { "total": 0, "recent": [] }
}
```

`metrics` 携带 `requestsTotal`、`requestsFailed`、`avgLatencyMs`、`rateLimited`、`staleRefRetries`、`rateBucketHosts`，以及 Go 运行时指标 `goHeapAllocMB`、`goHeapSysMB`、`goNumGoroutine` 和 `goHeapObjects`。失败事件在有原因时携带 `code` 和 `message`。

失败并不总是 4xx。`POST /actions` 和 `POST /macro` 按设计返回 200 并附带逐项结果，因此步骤失败的一次运行会发布一个失败原因而非状态码：它计入 `requestsFailed`，以该运行的路径、`status: 200` 出现在 `failures.recent` 中，消息形如 `N of M steps failed: <第一个步骤的错误>`，并以 `WARN` 级别记录日志。因此观察 `requestsFailed` 能看到一个批次全部失败的代理，而这正是这些端点所要服务的流量形态。

这在**两个**层次都成立。原因通过它自己的响应头跨过代理这一跳传递，因此由实例服务、由 front door 代理的批次也会移动 front door 的 `requestsFailed`——而这正是运行 `pinchtab server` 的运维人员 curl 的数字，因为那里的 `GET /metrics` 报告的是 front door 自己的计数器。一个不携带原因的代理响应无论状态码为何都不记录任何内容。

**两个层次绝不能相加。** 一个同时覆盖两者的 `requestsTotal` 会同时指代两件事，运维人员无法据此采取行动：一次飙升既不能说明客户在门口被拒，也不能说明浏览器在出故障。而应在各自的端点读取每一层：

| 端点 | 由谁应答 | 报告 |
|----------|-------------|---------|
| `GET /metrics` | front door（服务端模式）/ bridge 自身（bridge 模式） | 该进程自己的计数器 |
| `GET /instances/{id}/metrics` | 代理到该实例 | 该实例的计数器 |
| `GET /instances/metrics` | front door | 每个实例的**浏览器内存**，而非请求计数器 |
| `GET /tabs/{id}/metrics` | 代理到所属实例 | 该实例的**浏览器内存**——整个进程树，而非该标签页 |

最后一行中的标签页 id 只用于选择要询问的实例，绝不用于测量。其背后没有按标签页的读数：同一实例的两个标签页给出相同答案，且该数字就是该实例为自己报告的那个。见 [内存监控](../guides/memory-monitoring.md)。

`failures.recent` 在每个事件上、以及在整个块上都重复 `layer`，因此把事件粘贴进 bug 报告后仍能说明它来自哪里。

`failures` 和 `crashes` 始终存在，无内容时为空。一个只在首次失败后才出现的 key，任何监听失败的程序都无法依赖它。

## 两种模式共享什么、不共享什么

共享的，由结构决定——两种模式服务相同的诊断负载：

- `metrics`（请求计数器、延迟、Go 运行时指标）
- 带 `requestsFailed` 和 `recent` 的 `failures`
- 带 `total` 和 `recent` 的 `crashes`
- `/health` 上的 `version`

刻意不同的：

- **`/metrics` 在两种模式下都本地应答；`/health` 不是。** 在服务端模式下 `/health` 是编排器自己的封装（`instances`、`profiles`、`defaultInstance`、`restartRequired`），因为这些事实只存在于 front door。bridge 没有实例可报告。因此两个 `/health` 主体有意不同，跨模式比较应使用 `/metrics`。
- **`memory` 只出现在实例的 `/metrics` 上**，因为只有持有浏览器的进程才能测量它。
- **浏览器崩溃事件由拥有浏览器的进程记录**，因此在服务端模式下崩溃出现在实例层之下，而非 front door——front door 的 `crashes` 块按设计保持为空。

## 相关页面

- [Health](./health.md)
- [Instances](./instances.md)
