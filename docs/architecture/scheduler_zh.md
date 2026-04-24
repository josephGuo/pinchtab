# 调度器架构

本页描述了 PinchTab 可选调度器的内部工作原理。

## 范围

调度器是排队浏览器任务的可选分发层。它位于现有的标签页操作执行路径之前，并添加：

- 队列准入限制
- 每个代理的公平性
- 有界并发
- 协作取消
- 带 TTL 的结果保留

它不会替换直接操作端点。

## 运行时位置

在仪表板模式下，调度器仅在 `scheduler.enabled` 为 true 时创建。它直接注册在主多路复用器上，并暴露：

- `POST /tasks`
- `GET /tasks`
- `GET /tasks/{id}`
- `POST /tasks/{id}/cancel`
- `GET /scheduler/stats`
- `POST /tasks/batch`

## 高级流程

```text
client
  -> POST /tasks
  -> scheduler admission
  -> in-memory task queue
  -> worker dispatch
  -> resolve tab -> instance port
  -> POST /tabs/{tabId}/action on the owning instance
  -> result store
```

## 核心组件

### Scheduler

调度器拥有：

- 配置
- 任务队列
- 结果存储
- 活动任务的任务查找
- 取消映射
- 工作器生命周期

### TaskQueue

队列是：

- 内存中的
- 全局限制感知
- 每个代理限制感知
- 代理内按优先级排序
- 跨代理公平感知

在代理队列内：

- 较低的 `priority` 值获胜
- 相等优先级回退到按 `CreatedAt` 的 FIFO

跨代理：

- 选择正在执行任务最少的代理

### ResultStore

结果存储保存任务快照，并在配置的 TTL 后驱逐终端任务。

### ManagerResolver

解析器通过 `instance.Manager.FindInstanceByTabID` 将 `tabId` 映射到拥有实例端口。

这是调度器知道将执行转发到哪里的方式。

## 分发生命周期

任务通过这些内部状态移动：

```text
queued -> assigned -> running -> done
                           -> failed
queued -> cancelled
queued -> rejected
```

实现细节：

- `assigned` 在工作器执行开始前设置
- `running` 在代理操作请求前立即设置
- `done`、`failed` 和 `cancelled` 是终端状态
- 被拒绝的任务被存储为终端结果，即使它们从未进入活动执行

## 准入和限制

准入检查包括：

- 全局队列大小
- 每个代理队列大小

执行检查包括：

- 最大全局进行中计数
- 最大每个代理进行中计数

这些在队列和工作器分发路径中强制执行，而不是由外部基础设施强制执行。

## 取消模型

每个运行中的任务都有自己的 `context.WithDeadline(...)`。

调度器存储相应的取消函数，以便：

- `POST /tasks/{id}/cancel` 可以停止运行中的任务
- 关闭可以取消进行中的工作
- 截止日期自然传播到代理请求

## 截止日期处理

存在两条截止日期路径：

- 排队任务过期：后台收割器每秒扫描排队任务并将过期任务标记为失败
- 运行任务截止日期：每个任务的上下文截止日期由对执行器的 HTTP 请求强制执行

排队过期当前记录：

```text
deadline exceeded while queued
```

## 执行契约

调度器不会发明单独的执行协议。它将每个任务转换为正常的操作请求体：

```json
{
  "kind": "<action>",
  "ref": "<ref>",
  "...params": "..."
}
```

并将其转发到：

```text
POST /tabs/{tabId}/action
```

这使即时路径和计划路径保持一致。

## 错误模型

常见失败源：

- 队列已满导致准入拒绝
- 标签页到实例的解析失败
- 执行器 HTTP 失败
- 浏览器端操作失败
- 截止日期过期

调度器任务快照为这些情况保留最终的 `error` 字符串。

## 结果保留

终端任务快照存储在内存中，并在配置的 TTL 后驱逐。这使得 `GET /tasks/{id}` 和 `GET /tasks` 对短期检查有用，而不是用于长期审计存储。

## 设计权衡

当前调度器倾向于：

- 简单的内存操作
- 低依赖计数
- 重用现有的操作执行器

这意味着它目前不提供：

- 持久队列存储
- 进程重启后的持久恢复
- 单独的任务执行 DSL

---

## 第 2 阶段 -- 可观察性层

### 指标

调度器通过 `Metrics` 结构体维护全局和每个代理的计数器。全局计数器使用 `atomic.Uint64` 进行无锁递增。每个代理的计数器使用分片互斥模式 (`agentMetricEntry`) 以避免全局锁争用。

跟踪的计数器：

- `TasksSubmitted`、`TasksCompleted`、`TasksFailed`
- `TasksCancelled`、`TasksRejected`、`TasksExpired`
- `DispatchTotal` 和累积 `DispatchLatency`（用于平均调度延迟计算）

`Metrics.Snapshot()` 返回 `MetricsSnapshot` -- 一个包含每个代理细分的时间点可序列化副本。

### 统计端点

`GET /scheduler/stats` 暴露三个部分：

- **queue** -- 来自 `QueueStats()` 的当前排队/进行中计数
- **metrics** -- 来自 `Metrics.Snapshot()` 的快照
- **config** -- 当前调度器配置值

### 生命周期日志记录

关键调度器事件通过 `slog` 记录：

- 任务提交、分发、完成、失败、取消
- 队列准入拒绝
- 配置重新加载事件
- webhook 传递结果

### Webhook 传递

当任务具有 `callbackUrl` 并达到终端状态时，调度器发送带有任务快照作为 JSON 的 POST。这在 `finishTask()` 的 goroutine 中异步运行。

安全约束：

- 只接受 `http` 和 `https` 方案（SSRF 缓解）
- 具有 10 秒超时的专用 `http.Client` 防止挂起连接
- 传递失败会被记录，但不影响任务状态

发送自定义头：`X-PinchTab-Event: task.completed` 和 `X-PinchTab-Task-ID`。

---

## 第 3 阶段 -- 强化层

### 批量提交

`POST /tasks/batch` 接受一组任务定义（最多 50 个），共享单个 `agentId` 和可选的 `callbackUrl`。每个任务通过 `Submit()` 单独提交，因此队列准入限制按任务应用。

批处理端点支持部分失败：如果某些任务被拒绝（队列已满），接受的任务仍会被提交，并且响应包含每个任务的状态。

### 配置热重载

`ReloadConfig(cfg Config)` 在运行时更新调优旋钮：

- 通过 `queue.SetLimits(maxQueue, maxPerAgent)` 设置队列限制
- 通过 `cfgMu` 保护的配置字段设置进行中限制
- 通过 `results.SetTTL(ttl)` 设置结果 TTL

重载配置中的零值被忽略，保留现有设置。

`ConfigWatcher` 是一个后台 goroutine，定期调用用户提供的 `loadFn() (Config, error)` 并通过 `ReloadConfig` 应用更改。它可以干净地启动和停止。

### SetLimits 和 SetTTL

队列和结果存储暴露安全的运行时变更器：

- `TaskQueue.SetLimits(maxQueue, maxPerAgent)` -- 原子更新准入阈值
- `ResultStore.SetTTL(ttl)` -- 更新终端任务快照的驱逐窗口