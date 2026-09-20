# 内存监控

PinchTab 暴露它启动的 Chrome 进程的内存信息。每个实例并排上报两组测量：操作系统视角下其进程树的视图，以及 Chrome 自身为该实例跟踪的各标签页上报的页面计数器。

## PinchTab 测量什么

对于进程树，PinchTab 遍历运行中实例的 Chrome 进程：

1. 找到主浏览器 PID
2. 枚举子进程
3. 把浏览器及其子进程的 RSS 内存求和
4. 统计 renderer 进程数

对于页面，PinchTab 通过 CDP 向该实例跟踪的每个标签页询问其各自的 `Performance.getMetrics` 读数，并把结果求和。

## 内存字段

| 字段 | 含义 |
| --- | --- |
| `memoryMB` | 整个浏览器进程树的真实 RSS 内存 |
| `renderers` | 浏览器进程树中的 renderer 进程数 |
| `page.targets` | 读数被计入 `page` 汇总的标签页数 |
| `page.jsHeapUsedMB` / `page.jsHeapTotalMB` | 这些标签页上的 JavaScript 堆已用与已预留，求和 |
| `page.documents` / `page.frames` / `page.nodes` / `page.jsEventListeners` | 这些标签页上的 DOM 计数器，求和 |
| `unreadableTargets` | 在读取超时内未应答的标签页；它们不贡献任何数值 |

每个字段都是实测的；负载中没有哪个字段是从另一个字段推导出来的。

## 聚合规则

- **哪些 target 贡献：** 该实例跟踪的每个有活动上下文的标签页。每个都用 `Performance.getMetrics` 单独读取，`page` 块是所有已到达读数之和。
- **范围：** 实例。`memoryMB` 是整个进程树，其中还包含 GPU 与工具进程以及共享浏览器进程；`page` 只覆盖标签页。二者描述的是不同总体，从不做算术合并或比较。
- **缺失 vs 不可读：** 没有标签页贡献时，`page` 被省略。`unreadableTargets` 表示有多少标签页被询问却未应答（采集中途关闭或崩溃，或读取超时）。绝不会因为某个标签页读不到就把它报成堆 `0` 或节点 `0`：`{"unreadableTargets":0}` 且无 `page` 表示没有标签页；`{"unreadableTargets":2}` 且无 `page` 表示有两个标签页不肯应答。
- **开销（实测，开 5 个标签页）：** 每次读标签页约 1 ms，进程树遍历约 30 ms，因此一次轮询每个实例几十毫秒，每多开一个标签页大约多 1 ms。仪表板的 **Memory metrics** 开关并不专门用于控制 CDP 读取：它控制的是仪表板在每次监控 tick 时到底是否轮询每个运行中实例的 `/metrics`。对一个实例执行 `GET /metrics` 总是同时采集两者。

重要限制：

- `GET /tabs/{id}/metrics` 返回的是所属浏览器实例的聚合，包括其全部标签页上的 `page` 汇总，而不是隔离的单标签页数字

## 实例指标

对于单个运行中的浏览器，读该实例自己的 `/metrics`。在 `pinchtab bridge` 下就是 bridge 端口；在 `pinchtab server` 后面，要走实例路由，因为服务器自己的 `GET /metrics` 返回的是前门计数器（`"layer": "frontDoor"`），没有 `memory` 块：

```bash
curl http://localhost:9867/metrics                      # pinchtab bridge
curl http://localhost:9867/instances/<instanceId>/metrics   # pinchtab server
```

示例形状（节选；省略了 `failures` 和 `crashes`）：

```json
{
  "layer": "instance",
  "metrics": {
    "goHeapAllocMB": 12.5,
    "goHeapSysMB": 24.0,
    "goNumGoroutine": 15
  },
  "memory": {
    "memoryMB": 850.5,
    "renderers": 11,
    "page": {
      "targets": 3,
      "jsHeapUsedMB": 41.2,
      "jsHeapTotalMB": 64.0,
      "documents": 4,
      "frames": 5,
      "nodes": 9120,
      "jsEventListeners": 212
    },
    "unreadableTargets": 0
  }
}
```

## 单标签页指标

```bash
curl http://localhost:9867/tabs/<tabId>/metrics
```

示例形状：

```json
{
  "memoryMB": 850.5,
  "renderers": 11,
  "page": { "targets": 3, "jsHeapUsedMB": 41.2, "jsHeapTotalMB": 64.0, "documents": 4, "frames": 5, "nodes": 9120, "jsEventListeners": 212 },
  "unreadableTargets": 0
}
```

把它理解为「拥有此标签页的浏览器实例的内存」，而不是「仅此标签页的内存」：`page` 块是对该实例所有标签页求和的。

## 所有运行中实例

在编排器模式下：

```bash
curl http://localhost:9867/instances/metrics
```

它为每个运行中实例返回一个指标对象——`instanceId`、`profileName`、`memoryMB`、`renderers`、`page` 和 `unreadableTargets`——这是在多台实例之间比较内存的最佳 API。

## 页面堆与泄漏排查

上面的数字是实例级的。要看单个标签页的 JavaScript 堆、堆快照和快照对比，请改用 `/memory` 端点和 `pinchtab memory`：

```bash
pinchtab memory --gc                 # GET /memory?gc=true: heap usage and DOM counters
pinchtab memory snapshot             # POST /memory/snapshot (needs security.allowMemory)
pinchtab memory compare <base> <head>
```

`GET /memory` 始终可用；snapshot、summary 和 compare 需要 `security.allowMemory`（默认关）。端点、响应形状和泄漏排查演练见 [Memory](../reference/memory.md)。

## 仪表板监控

仪表板从以下流消费监控快照：

```bash
curl http://localhost:9867/api/events?memory=1
```

该流包含：

- 实例列表
- 标签页列表
- 当 `memory=1` 时的逐实例指标
- PinchTab 进程自身的服务器指标

当前 SSE 监控循环以较短间隔更新，适合实时仪表板视图。

## 故障排查

### 内存显示 `0`

可能原因：

- Chrome 还没启动
- 实例已停止
- 浏览器上下文尚未初始化

### 内存看起来比预期高

记住 `memoryMB` 包含：

- 浏览器进程
- renderer 进程
- GPU 与工具子进程（若存在）

这通常更接近「操作系统看到的」，而不是一个狭窄的 JavaScript 堆数字。

### 数字与活动监视器或任务管理器不完全一致

不同工具报告的内存定义不同。PinchTab 目前对它所拥有的 Chrome 进程树报告基于 RSS 的总量。
