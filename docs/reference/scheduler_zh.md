# 调度器和任务

调度器是一个可选的内存任务队列，用于多代理协调。它通过 `/tasks` 接受任务，应用准入和公平规则，然后将工作分派到与即时浏览路由使用的相同标签页操作执行器。

它不会替换正常的直接路径。像 `POST /tabs/{id}/action` 这样的路由仍然独立工作。

目前没有 命令行界面 调度器命令。

## 启用调度器

调度器默认关闭。仅当 `scheduler.enabled` 为 true 时，仪表板模式才会注册任务路由。

```json
{
  "scheduler": {
    "enabled": true
  }
}
```

## 调度器配置

```json
{
  "scheduler": {
    "enabled": true,
    "strategy": "fair-fifo",
    "maxQueueSize": 1000,
    "maxPerAgent": 100,
    "maxInflight": 20,
    "maxPerAgentInflight": 10,
    "resultTTLSec": 300,
    "workerCount": 4
  }
}
```

| 字段 | 默认值 | 含义 |
| --- | --- | --- |
| `enabled` | `false` | 在仪表板模式中启用任务路由 |
| `strategy` | `fair-fifo` | 调度器策略标签 |
| `maxQueueSize` | `1000` | 全局排队任务限制 |
| `maxPerAgent` | `100` | 每个代理的排队任务限制 |
| `maxInflight` | `20` | 总体并发执行任务的最大值 |
| `maxPerAgentInflight` | `10` | 每个代理并发执行任务的最大值 |
| `resultTTLSec` | `300` | 终端任务快照的保留时间 |
| `workerCount` | `4` | 工作 goroutine 的数量 |

## 任务对象

任务是调度器拥有的记录，具有以下主要字段：

| 字段 | 含义 |
| --- | --- |
| `taskId` | 生成的任务 ID |
| `agentId` | 提交代理标识符 |
| `action` | 要运行的操作类型 |
| `tabId` | 目标标签页 ID |
| `ref` | 可选的元素引用 |
| `params` | 可选的操作特定请求字段 |
| `priority` | 数字越小优先级越高 |
| `state` | 当前任务状态 |
| `deadline` | 执行截止时间 |
| `createdAt` | 提交时间 |
| `startedAt` | 首次执行时间戳 |
| `completedAt` | 终端时间戳 |
| `latencyMs` | 从开始到完成的经过时间 |
| `result` | 执行器响应有效载荷 |
| `error` | 终端错误消息 |
| `position` | 提交时的队列位置 |
| `callbackUrl` | 终端状态通知的可选 webhook URL |

任务 ID 当前生成为 `tsk_XXXXXXXX`，但调用者仍应将它们视为不透明 ID。

## 提交任务

```bash
curl -X POST http://localhost:9867/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "agentId": "agent-crawl-01",
    "action": "click",
    "tabId": "8f9c7d4e1234567890abcdef12345678",
    "ref": "e14",
    "priority": 5,
    "deadline": "2026-03-08T12:05:00Z"
  }'
# 响应
{
  "taskId": "tsk_a1b2c3d4",
  "state": "queued",
  "position": 1,
  "createdAt": "2026-03-08T12:00:01Z"
}
```

此端点在成功队列提交时返回 `202 Accepted`。

请求字段：

| 字段 | 必需 | 说明 |
| --- | --- | --- |
| `agentId` | 是 | 在请求时验证 |
| `action` | 是 | 成为执行器 `kind` |
| `tabId` | 实际是 | 执行路径需要 |
| `ref` | 否 | 元素目标操作的顶级元素引用 |
| `params` | 否 | 合并到执行器请求体中的操作特定字段 |
| `priority` | 否 | 数字越小优先级越高 |
| `deadline` | 否 | RFC3339 时间戳；默认为 `now + 60s` |
| `callbackUrl` | 否 | webhook URL；在终端状态时接收带有任务快照的 POST |

重要：

- 请求验证仅强制 `agentId` 和 `action`
- 缺少 `tabId` 在执行期间会被拒绝，错误为 `tabId is required for task execution`
- 过去的截止时间在提交时被拒绝
- `agentId` 也作为 `X-Agent-Id` 转发给执行器，因此生成的浏览器操作在 `/api/activity` 和仪表板代理视图中归因于同一代理

## 队列已满响应

如果由于全局队列或代理队列已满而导致准入失败，调度器返回 `429 Too Many Requests`。

```bash
curl -X POST http://localhost:9867/tasks \
  -H "Content-Type: application/json" \
  -d '{"agentId":"agent-crawl-01","action":"click","tabId":"8f9c7d4e1234567890abcdef12345678"}'
# 响应
{
  "code": "queue_full",
  "error": "rejected: global queue full",
  "retryable": true,
  "details": {
    "agentId": "agent-crawl-01",
    "queued": 1000,
    "maxQueue": 1000,
    "maxPerAgent": 100
  }
}
```

## 列出任务

`GET /tasks` 返回调度器的内存任务快照，包括排队、运行和仍在 TTL 窗口内的最近完成的任务。

```bash
curl http://localhost:9867/tasks
# 响应
{
  "tasks": [
    {
      "taskId": "tsk_a1b2c3d4",
      "state": "done",
      "agentId": "agent-crawl-01",
      "action": "click",
      "latencyMs": 842
    }
  ],
  "count": 1
}
```

支持的查询过滤器：

- `agentId`
- `state`

示例：

```bash
curl 'http://localhost:9867/tasks?agentId=agent-crawl-01&state=done,failed'
```

## 获取单个任务

```bash
curl http://localhost:9867/tasks/tsk_a1b2c3d4
# 响应
{
  "taskId": "tsk_a1b2c3d4",
  "agentId": "agent-crawl-01",
  "action": "click",
  "tabId": "8f9c7d4e1234567890abcdef12345678",
  "ref": "e14",
  "priority": 5,
  "state": "done",
  "createdAt": "2026-03-08T12:00:01Z",
  "startedAt": "2026-03-08T12:00:01Z",
  "completedAt": "2026-03-08T12:00:02Z",
  "latencyMs": 842,
  "result": {
    "success": true
  }
}
```

如果未找到任务，调度器返回：

```json
{
  "code": "not_found",
  "error": "task not found"
}
```

## 取消任务

```bash
curl -X POST http://localhost:9867/tasks/tsk_a1b2c3d4/cancel
# 响应
{
  "status": "cancelled",
  "taskId": "tsk_a1b2c3d4"
}
```

行为：

- 排队任务从队列中移除
- 运行任务的执行上下文被取消
- 终端任务返回 `409 Conflict`

## 任务状态

已实现的状态：

- `queued`
- `assigned`
- `running`
- `done`
- `failed`
- `cancelled`
- `rejected`

终端状态：

- `done`
- `failed`
- `cancelled`
- `rejected`

## 任务如何执行

调度器将每个任务转发到正常的标签页操作端点：

```text
POST /tabs/{tabId}/action
```

它构建操作体如下：

```json
{
  "kind": "<action>",
  "ref": "<ref>",
  "...params": "..."
}
```

这意味着：

- `action` 成为 `kind`
- 顶级 `ref` 在存在时被转发
- `params` 中的每个键都被合并到顶级操作体中
- `agentId` 作为 `X-Agent-Id` 传播

示例：

```bash
curl -X POST http://localhost:9867/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "agentId": "my-agent",
    "action": "type",
    "tabId": "8f9c7d4e1234567890abcdef12345678",
    "ref": "e12",
    "params": {
      "text": "Alan Turing"
    }
  }'
```

实际上，任务有效载荷应使用与即时 `/tabs/{id}/action` 路由期望的相同操作字段。

## 公平性、截止时间和保留

- 在一个代理队列中，较低的 `priority` 值先运行
- 同一代理的相同优先级任务回退到 FIFO 顺序
- 跨代理，调度器优先选择正在执行任务最少的代理
- 如果排队任务在执行开始前超过其截止时间，它会被标记为失败，错误为 `deadline exceeded while queued`
- 终端任务快照在内存中保留 `resultTTLSec`

---

## 第二阶段 - 可观察性

### 调度器统计信息

`GET /scheduler/stats` 返回队列状态、运行时指标和配置的快照。

```bash
curl http://localhost:9867/scheduler/stats
# 响应
{
  "queue": {
    "totalQueued": 5,
    "totalInflight": 2,
    "agentCounts": {
      "agent-crawl-01": 3,
      "agent-scrape-02": 2
    }
  },
  "metrics": {
    "tasksSubmitted": 42,
    "tasksCompleted": 35,
    "tasksFailed": 3,
    "tasksCancelled": 2,
    "tasksRejected": 1,
    "tasksExpired": 1,
    "dispatchCount": 38,
    "avgDispatchLatencyMs": 12.5,
    "agents": {
      "agent-crawl-01": {
        "submitted": 25,
        "completed": 22,
        "failed": 2,
        "cancelled": 1,
        "rejected": 0
      }
    }
  },
  "config": {
    "strategy": "fair-fifo",
    "maxQueueSize": 1000,
    "maxPerAgent": 100,
    "maxInflight": 20,
    "maxPerAgentFlight": 10,
    "workerCount": 4,
    "resultTTL": "5m0s"
  }
}
```

#### 指标字段

| 字段 | 类型 | 含义 |
| --- | --- | --- |
| `tasksSubmitted` | uint64 | 启动以来接受的总任务数 |
| `tasksCompleted` | uint64 | 成功完成的任务 |
| `tasksFailed` | uint64 | 出错完成的任务 |
| `tasksCancelled` | uint64 | 通过 `POST /tasks/{id}/cancel` 取消的任务 |
| `tasksRejected` | uint64 | 准入时被拒绝的任务（队列已满） |
| `tasksExpired` | uint64 | 超过截止时间的排队任务 |
| `dispatchCount` | uint64 | 分派给工作者的任务数 |
| `avgDispatchLatencyMs` | float64 | 从队列进入到分派开始的平均时间 |
| `agents` | object | 每个代理的细分（已提交、已完成、失败、已取消、已拒绝） |

### Webhook 回调

任务可以包含 `callbackUrl` 字段。当任务达到终端状态（`done`、`failed` 或 `cancelled`）时，调度器会向该 URL 发送带有任务快照的 POST。

```bash
curl -X POST http://localhost:9867/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "agentId": "my-agent",
    "action": "click",
    "tabId": "8f9c7d4e1234567890abcdef12345678",
    "callbackUrl": "https://pinchtab.com/hooks/task-done"
  }'
```

Webhook 行为：

- 交付是尽力而为的：失败会被记录但不影响任务状态
- 只允许 `http` 和 `https` 方案（SSRF 保护）
- 使用具有 10 秒超时的专用 HTTP 客户端
- 发送自定义标头：`X-PinchTab-Event: task.completed` 和 `X-PinchTab-Task-ID: <taskId>`

`callbackUrl` 字段存储在任务上，并在 `GET /tasks/{id}` 中返回。

---

## 第三阶段 - 强化

### 批处理任务提交

`POST /tasks/batch` 在单个请求中提交多个任务。批处理中的所有任务共享相同的 `agentId` 和可选的 `callbackUrl`。

```bash
curl -X POST http://localhost:9867/tasks/batch \
  -H "Content-Type: application/json" \
  -d '{
    "agentId": "agent-crawl-01",
    "callbackUrl": "https://pinchtab.com/hooks/batch",
    "tasks": [
      { "action": "click", "tabId": "TAB_ID", "params": { "selector": "#btn" } },
      { "action": "scroll", "tabId": "TAB_ID", "params": { "scrollY": 400 } },
      { "action": "hover", "tabId": "TAB_ID", "params": { "selector": "h1" }, "priority": 1 }
    ]
  }'
# 响应 (202 Accepted)
{
  "tasks": [
    { "taskId": "tsk_aaaa1111", "state": "queued", "position": 1 },
    { "taskId": "tsk_bbbb2222", "state": "queued", "position": 2 },
    { "taskId": "tsk_cccc3333", "state": "queued", "position": 3 }
  ],
  "submitted": 3
}
```

#### 批处理请求字段

| 字段 | 必需 | 说明 |
| --- | --- | --- |
| `agentId` | 是 | 共享给批处理中的所有任务 |
| `callbackUrl` | 否 | 应用于每个任务的 webhook URL |
| `tasks` | 是 | 任务定义数组 (1–50) |

每个任务定义支持与单个任务提交相同的字段（`action`、`tabId`、`ref`、`params`、`priority`、`deadline`），除了从批处理继承的 `agentId` 和 `callbackUrl`。

#### 批处理验证

| 条件 | 响应 |
| --- | --- |
| 缺少 `agentId` | `400 Bad Request` |
| 空 `tasks` 数组 | `400 Bad Request` |
| 超过 50 个任务 | `400 Bad Request` 并返回 `batch_too_large` 代码 |
| 无效的 JSON 体 | `400 Bad Request` |

部分失败：如果某些任务被准入拒绝（队列已满），已接受的任务仍会被提交。响应单独包含每个任务的状态。

### 配置热重载

`ReloadConfig(cfg)` 在运行时更新队列限制、进行中限制和结果 TTL，而无需重启调度器。

可重载字段：

| 字段 | 更改内容 |
| --- | --- |
| `maxQueueSize`, `maxPerAgent` | 通过 `SetLimits()` 限制队列准入 |
| `maxInflight`, `maxPerAgentFlight` | 并发限制（受 `cfgMu` 保护） |
| `resultTTL` | 通过 `SetTTL()` 限制结果存储驱逐窗口 |

零值被忽略（保留现有设置）。

#### ConfigWatcher

`ConfigWatcher` 运行后台 goroutine，定期重新读取配置并调用 `ReloadConfig`。创建方式：

```go
cw := scheduler.NewConfigWatcher(30*time.Second, loadFn, sched)
cw.Start()
defer cw.Stop()
```

`loadFn` 是一个 `func() (Config, error)`，用于从磁盘或环境读取当前配置。