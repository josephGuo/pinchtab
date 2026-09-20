# 调度器与任务（Scheduler And Tasks）

调度器是一个可选的内存任务队列，用于多代理协调。它通过 `/tasks` 接受任务，应用准入和公平规则，然后将工作分派到与即时浏览路由相同的标签页动作执行器。

它不替代正常的直接路径。`POST /tabs/{id}/action` 等路由仍独立工作。

目前没有命令行界面调度器命令。

## 启用调度器

调度器默认关闭。仅当 `scheduler.enabled` 为 true 时，仪表板模式才注册任务路由。

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
    "workerCount": 4,
    "maxBatchSize": 50
  }
}
```

| 字段 | 默认值 | 含义 |
| --- | --- | --- |
| `enabled` | `false` | 在仪表板模式启用任务路由 |
| `strategy` | `fair-fifo` | 调度器策略标签 |
| `maxQueueSize` | `1000` | 全局排队任务上限 |
| `maxPerAgent` | `100` | 每个代理的排队任务上限 |
| `maxInflight` | `20` | 整体并发执行任务数上限 |
| `maxPerAgentInflight` | `10` | 每个代理并发执行任务数上限 |
| `resultTTLSec` | `300` | 终端任务快照的保留时间 |
| `workerCount` | `4` | worker goroutine 数量 |
| `maxBatchSize` | `50` | 一次 `POST /tasks/batch` 最多可提交的任务数；更大的批次以 `batch_too_large` 拒绝 |

数值型旋钮设为 `0` 或以下即表示取其默认值，绝不表示无限制。调度器在启动时读取这些值一次；改动后需重启服务器才生效。

## 任务对象

任务是调度器拥有的记录，主要字段如下：

| 字段 | 含义 |
| --- | --- |
| `taskId` | 生成的任务 ID |
| `agentId` | 提交代理标识 |
| `action` | 要运行的动作种类 |
| `tabId` | 目标标签页 ID |
| `selector` | 可选元素选择器（ref、CSS、XPath 或文本） |
| `ref` | 可选元素 ref（已弃用；用 `selector`） |
| `params` | 可选的动作特定请求字段 |
| `priority` | 数字越小优先级越高 |
| `state` | 当前任务状态 |
| `deadline` | 执行截止时间 |
| `createdAt` | 提交时间 |
| `startedAt` | 首次执行时间戳 |
| `completedAt` | 终端时间戳 |
| `latencyMs` | 从开始到完成的耗时 |
| `result` | 执行器响应负载 |
| `error` | 终端错误消息 |
| `position` | 提交时的队列位置 |
| `callbackUrl` | 终端状态通知的可选 webhook URL |

任务 ID 当前生成为 `tsk_XXXXXXXX`，但调用方仍应将其视为不透明 ID。

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
# Response
{
  "taskId": "tsk_a1b2c3d4",
  "state": "queued",
  "position": 1,
  "createdAt": "2026-03-08T12:00:01Z"
}
```

该端点在成功入队时返回 `202 Accepted`。

请求字段：

| 字段 | 必填 | 说明 |
| --- | --- | --- |
| `agentId` | 是 | 请求时校验 |
| `action` | 是 | 成为执行器 `kind` |
| `tabId` | 实际上必填 | 执行路径需要 |
| `selector` | 否 | 面向元素动作的顶层元素选择器 |
| `ref` | 否 | 已弃用的顶层元素 ref；用 `selector` |
| `params` | 否 | 合并进执行器请求主体的动作特定字段（`params` 内的 `kind`、`ref`、`selector` 和 `tabId` 被忽略） |
| `priority` | 否 | 数字越小优先级越高 |
| `deadline` | 否 | RFC3339 时间戳；默认为 `now + 60s` |
| `callbackUrl` | 否 | webhook URL；终端状态时接收带任务快照的 POST |

重要：

- 请求校验只强制 `agentId` 和 `action`（若给了 `callbackUrl` 则还要有效）；其他失败应答 `400`
- 缺少 `tabId` 在执行期间稍后被拒绝，错误为 `tabId is required for task execution`
- 已过去的截止时间在提交时被拒绝
- `agentId` 也作为 `X-Agent-Id` 转发给执行器，因此产生的浏览器动作在 `/api/activity` 和仪表板 Agents 视图中归因于同一代理

## 队列已满响应

若因全局队列或某代理队列已满导致准入失败，调度器返回 `429 Too Many Requests`。

```bash
curl -X POST http://localhost:9867/tasks \
  -H "Content-Type: application/json" \
  -d '{"agentId":"agent-crawl-01","action":"click","tabId":"8f9c7d4e1234567890abcdef12345678"}'
# Response
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

`GET /tasks` 返回调度器内存中的任务快照，包括排队中、运行中以及仍在 TTL 窗口内的最近完成任务。

```bash
curl http://localhost:9867/tasks
# Response
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
# Response
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

任务未找到时，调度器返回：

```json
{
  "code": "not_found",
  "error": "task not found"
}
```

## 取消任务

```bash
curl -X POST http://localhost:9867/tasks/tsk_a1b2c3d4/cancel
# Response
{
  "status": "cancelled",
  "taskId": "tsk_a1b2c3d4"
}
```

行为：

- 排队任务从队列移除
- 运行中任务的执行上下文被取消
- 终端任务返回 `409 Conflict`

## 任务状态

已实现状态：

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

调度器将每个任务转发到正常的标签页动作端点：

```text
POST /tabs/{tabId}/action
```

它这样构建动作主体：

```json
{
  "kind": "<action>",
  "ref": "<ref>",
  "selector": "<selector>",
  "...params": "..."
}
```

这意味着：

- `action` 成为 `kind`
- 顶层 `ref` 和 `selector` 存在时被转发
- `params` 中每个其他 key 被合并进顶层动作主体；`params` 不能覆盖 `kind`、`ref`、`selector` 或 `tabId`
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

实践中，任务负载应使用即时 `/tabs/{id}/action` 路由所期望的相同动作字段。

## 公平性、截止时间与保留

- 同一代理队列内，`priority` 值越低越先运行
- 同一代理的相同优先级任务回退到 FIFO 顺序
- 跨代理时，调度器优先选择在途任务最少的代理
- 排队任务若在执行开始前超过截止时间，会被标记失败，错误为 `deadline exceeded while queued`
- 终端任务快照在内存中保留 `resultTTLSec`

---

## 第二阶段 — 可观测性

### 调度器统计

`GET /scheduler/stats` 返回队列状态、运行时指标和配置的快照。

```bash
curl http://localhost:9867/scheduler/stats
# Response
{
  "queue": {
    "totalQueued": 5,
    "totalInflight": 2,
    "agents": {
      "agent-crawl-01": { "queued": 3, "inflight": 1 },
      "agent-scrape-02": { "queued": 2, "inflight": 1 }
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
| `dispatchCount` | uint64 | 分派给 worker 的任务数 |
| `avgDispatchLatencyMs` | float64 | 从入队到分派开始的平均时间 |
| `agents` | object | 每个代理的细分（submitted、completed、failed、cancelled、rejected） |

### Webhook 回调

任务可带 `callbackUrl` 字段。任务到达终端状态（`done`、`failed` 或 `cancelled`）时，调度器向该 URL POST 任务快照。

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

- 交付为尽力而为：失败会被记录但不影响任务状态
- 仅允许 `http` 和 `https` scheme，URL 中不得含凭据，且主机必须解析到公网地址（环回和私有/内部主机在提交时以 `400` 拒绝）
- 使用带 10 秒超时的专用 HTTP 客户端
- 发送自定义头：`X-PinchTab-Event: task.completed` 和 `X-PinchTab-Task-ID: <taskId>`

`callbackUrl` 字段存储在任务上，并在 `GET /tasks/{id}` 中返回。

---

## 第三阶段 — 强化

### 批量任务提交

`POST /tasks/batch` 在单个请求中提交多个任务。批次中所有任务共享相同的 `agentId` 和可选 `callbackUrl`。

```bash
curl -X POST http://localhost:9867/tasks/batch \
  -H "Content-Type: application/json" \
  -d '{
    "agentId": "agent-crawl-01",
    "callbackUrl": "https://pinchtab.com/hooks/batch",
    "tasks": [
      { "action": "click", "tabId": "TAB_ID", "ref": "e5" },
      { "action": "scroll", "tabId": "TAB_ID", "params": { "scrollY": 400 } },
      { "action": "hover", "tabId": "TAB_ID", "ref": "e9", "priority": 1 }
    ]
  }'
# Response (202 Accepted)
{
  "tasks": [
    { "taskId": "tsk_aaaa1111", "state": "queued", "position": 1 },
    { "taskId": "tsk_bbbb2222", "state": "queued", "position": 2 },
    { "taskId": "tsk_cccc3333", "state": "queued", "position": 3 }
  ],
  "submitted": 3
}
```

#### 批量请求字段

| 字段 | 必填 | 说明 |
| --- | --- | --- |
| `agentId` | 是 | 批次中所有任务共享 |
| `callbackUrl` | 否 | 应用于每个任务的 webhook URL |
| `tasks` | 是 | 任务定义数组（1 到 `scheduler.maxBatchSize`） |

每个任务定义支持 `action`、`tabId`、`ref`、`params`、`priority` 和 `deadline`；`agentId` 和 `callbackUrl` 从批次继承。批量任务没有 `selector` 字段，`params` 内的 `selector` 也被丢弃，因此在批次中用 `ref` 定位元素。

#### 批量校验

| 条件 | 响应 |
| --- | --- |
| 缺少 `agentId` | `400 Bad Request` |
| 空 `tasks` 数组 | `400 Bad Request` |
| 任务数超过 `scheduler.maxBatchSize`（默认 50） | `400 Bad Request`，代码 `batch_too_large` |
| 无效 JSON 主体 | `400 Bad Request` |

部分失败：若某些任务被准入（队列已满）或校验拒绝，已接受的任务仍会提交。每个被拒项带回 `state: "rejected"` 和一个 `error`；`submitted` 是响应中的项数，含被拒项。

### 配置热重载（仅 Go API）

这些是 `internal/scheduler` 中面向嵌入者的 Go API。`pinchtab server` 不会启动 `ConfigWatcher` 也不调用 `ReloadConfig`，因此编辑配置文件中的 `scheduler.*` 仍需重启。

`ReloadConfig(cfg)` 在运行时更新队列上限、在途上限和结果 TTL，无需重启调度器。

可重载字段：

| 字段 | 变更内容 |
| --- | --- |
| `maxQueueSize`, `maxPerAgent` | 通过 `SetLimits()` 调整队列准入上限 |
| `maxInflight`, `maxPerAgentFlight` | 并发上限（受 `cfgMu` 保护） |
| `resultTTL` | 通过 `SetTTL()` 调整结果存储驱逐窗口 |

零值被忽略（保留现有设置）。

#### ConfigWatcher

`ConfigWatcher` 运行一个后台 goroutine，周期性重读配置并调用 `ReloadConfig`。创建方式：

```go
cw := scheduler.NewConfigWatcher(30*time.Second, loadFn, sched)
cw.Start()
defer cw.Stop()
```

`loadFn` 是一个 `func() (Config, error)`，从磁盘或环境读取当前配置。
