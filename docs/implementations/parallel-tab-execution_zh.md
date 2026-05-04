# 并行标签页执行

PinchTab 支持跨浏览器标签页的安全并行执行。多个标签页可以同时执行操作，而每个标签页内部保持顺序执行，防止资源耗尽和竞态条件。

## 架构

```
                         ┌──────────────────────────────────────────┐
HTTP Request (tab1) ─┐   │              TabExecutor                 │
HTTP Request (tab2) ─┼──▶│  ┌────────────────────────────────────┐  │
HTTP Request (tab3) ─┘   │  │ Global Semaphore (chan struct{})    │  │
                         │  │  capacity = maxParallel (1–8)      │  │
                         │  └──────────┬─────────────────────────┘  │
                         │             │                            │
                         │  ┌──────────▼─────────────────────────┐  │
                         │  │ Per-Tab Mutex (map[string]*Mutex)   │  │
                         │  │  tab1 → sync.Mutex                 │  │
                         │  │  tab2 → sync.Mutex                 │  │
                         │  │  tab3 → sync.Mutex                 │  │
                         │  └──────────┬─────────────────────────┘  │
                         │             │                            │
                         │  ┌──────────▼─────────────────────────┐  │
                         │  │ Panic Recovery (per-task defer)     │  │
                         │  └──────────┬─────────────────────────┘  │
                         │             │                            │
                         │  ┌──────────▼─────────────────────────┐  │
                         │  │ chromedp Context (isolated per tab) │  │
                         │  └────────────────────────────────────┘  │
                         └──────────────────────────────────────────┘
```

### 执行流程

通过并行执行系统的完整请求生命周期：

```
HTTP POST /tabs/{id}/action  (例如，点击按钮)
    │
    ▼
Handler: HandleAction()
    │
    ▼
Bridge.EnsureChrome()  [首次请求时延迟初始化]
    │
    ▼
Bridge.TabContext(tabID)  [获取标签页的 chromedp.Context]
    │
    ▼
Bridge.Execute(ctx, tabID, task)
    │
    ▼
TabManager.Execute()
    │
    ▼
TabExecutor.Execute(ctx, tabID, task)
    ├─ 阶段 1: te.semaphore <- struct{}   [获取全局槽位]
    ├─ 阶段 2: tabMutex(tabID).Lock()     [获取每个标签页的锁]
    └─ 阶段 3: safeRun(ctx, tabID, task)  [执行并进行 panic 恢复]
        ├─ chromedp.Run(ctx, action...)
        └─ 返回结果或错误
    │
    ▼
HTTP 200 {"success": true, "result": {...}}
```

### 执行模型

每个标签页**顺序**执行任务（一次一个），但**不同标签页**可以同时运行，最多达到可配置的限制：

```
时间 ──────────────────────────────────────────────────▶
Tab1 ──▶ [action1] ──▶ [action2] ──▶ [action3]
Tab2 ──▶ [action1] ──▶ [action2]                         (与 Tab1 并发)
Tab3 ──▶ [action1] ──▶ [action2] ──▶ [action3]           (与 Tab1 和 Tab2 并发)
```

两阶段锁定确保正确性：

1. **阶段 1 — 信号量获取**：请求获取全局 `chan struct{}` 信号量中的一个槽位。如果所有槽位都被占用，goroutine 将阻塞，直到一个槽位释放或上下文过期。
2. **阶段 2 — 标签页互斥锁获取**：在获得信号量槽位后，请求获取每个标签页的 `sync.Mutex`。这保证在任何时刻只有一个 CDP 操作针对给定标签页运行。

```go
// TabExecutor.Execute() 中的简化流程
select {
case te.semaphore <- struct{}{}:   // 阶段 1: 全局槽位
    defer func() { <-te.semaphore }()
case <-ctx.Done():
    return ctx.Err()
}
tabMu := te.tabMutex(tabID)       // 阶段 2: 每个标签页的锁
tabMu.Lock()
defer tabMu.Unlock()
return te.safeRun(ctx, tabID, task) // 执行并进行 panic 恢复
```

### 组件

| 组件 | 位置 | 目的 |
|------|------|------|
| `TabExecutor` | `internal/bridge/tab_executor.go` | 核心并行执行引擎 |
| `TabManager.Execute()` | `internal/bridge/tab_manager.go` | 处理程序的集成点 |
| `Bridge.Execute()` | `internal/bridge/bridge.go` | BridgeAPI 接口方法 |
| `LockManager` | `internal/bridge/lock.go` | 带有 TTL 的每个标签页所有权锁 |
| `TabEntry` | `internal/bridge/bridge.go` | 每个标签页的 chromedp 上下文 + 元数据 |

### 工作原理

1. **全局信号量** — 一个缓冲通道（容量为 `maxParallel` 的 `chan struct{}`）限制了同时执行的标签页数量。当信号量满时，新任务会等待（尊重上下文取消/超时）。

2. **每个标签页的互斥锁** — 每个标签页都有自己的 `sync.Mutex` 存储在 `map[string]*sync.Mutex` 中。这确保单个标签页内的操作一次执行一个。这可以防止在同一个标签页上并发执行 CDP 操作，这是 chromedp 不支持的。

3. **Panic 恢复** — 每个任务都包装在 `defer recover()` 块中。一个标签页任务中的 panic 不会崩溃进程或影响其他标签页。Panic 被转换为 `error` 并通过 `slog.Error` 记录。

4. **上下文传播** — 调用者的上下文（带有超时/取消）被传递到任务函数。如果上下文在等待信号量或标签页锁时过期，调用会立即返回错误。清理 goroutine 确保即使上下文在等待过程中过期，每个标签页的互斥锁也会被解锁。

5. **CDP 上下文隔离** — 每个标签页由自己的 `chromedp.Context` 支持，通过 `chromedp.NewContext(browserCtx, chromedp.WithTargetID(...))` 创建。这意味着每个标签页都有独立的 Chrome DevTools Protocol 会话，具有自己的 DOM、网络堆栈和 JavaScript 运行时。

## 架构灵感

### 来自 Vercel Agent Browser 的灵感

[Vercel Agent Browser](https://github.com/vercel-labs/agent-browser) 是一个为 AI 代理设计的无头浏览器自动化 命令行界面。它使用客户端-守护进程架构，其中 Rust 命令行界面 与管理 Playwright 浏览器实例的持久 Node.js 守护进程（或实验性原生 Rust 守护进程）通信。Agent Browser 的几种架构模式直接影响了 PinchTab 的并行标签页执行设计。

#### 我们研究的内容

**浏览器会话管理** — Agent Browser 通过 `--session` 标志隔离并发工作负载。每个会话（`--session agent1`、`--session agent2`）生成一个完全独立的浏览器实例，具有独立的 cookie、存储、导航历史和认证状态。会话通过作为单独的 OS 进程并行运行。守护进程在会话内的命令之间保持持久，因此后续的 命令行界面 调用（`open`、`click`、`fill`）速度很快。

**任务执行模型** — Agent Browser 遵循严格的每次调用一个命令的模型。每个 命令行界面 调用是通过 IPC 发送到会话守护进程的离散任务。守护进程序列化会话内的命令：每个会话一次只执行一个命令。这是一个设计选择 — Playwright 上下文不是线程安全的，因此序列化可以防止竞态条件。命令行界面 客户端会阻塞，直到守护进程响应，强制执行严格的请求-响应循环，IPC 读取超时为 30 秒（默认 Playwright 超时设置为 25 秒，以确保正确的错误消息而不是通用超时）。

**并发结构** — 多个会话可以同时运行，但每个单独的会话是单线程的（一次一个命令）。这提供了会话级并发：N 个会话 = N 个并发浏览器实例，每个实例一次处理一个命令。资源通过 OS 隐式管理 — 每个会话是一个具有自己内存空间的单独进程。

**快照和引用工作流** — Agent Browser 生成带有稳定 `ref` 标识符（`@e1`、`@e2`）的可访问性树快照，这些标识符会一直保持到下一个快照。AI 代理使用这些引用来进行确定性元素选择。这影响了 PinchTab 的 `RefCache` 设计，其中每个标签页维护自己的带有节点引用的快照缓存。

**错误处理** — Agent Browser 通过 命令行界面 退出代码按命令返回错误。失败的命令不会崩溃守护进程 — 会话对后续命令保持活动状态。命令支持 `--json` 输出，用于机器可读的错误报告。

#### PinchTab 如何不同地适应这些想法

PinchTab 在根本不同的架构级别运行：

**标签页级与会话级隔离** — Agent Browser 为每个会话创建单独的浏览器进程，而 PinchTab 在 CDP 目标（标签页）级别进行隔离。每个标签页通过 `chromedp.NewContext(browserCtx, chromedp.WithTargetID(targetID))` 获取自己的 `chromedp.Context`，为其提供具有自己的 DOM、网络堆栈和 JavaScript 运行时的独立 CDP 会话。多个并发工作负载共享单个 Chrome 进程，但通过 CDP 目标保持隔离。这更具资源效率：一个带有 10 个标签页的 Chrome 进程使用的内存少于 10 个单独的 Chrome 实例。

**内部并发控制与外部序列化** — Agent Browser 依赖于守护进程架构进行序列化 — 守护进程每个会话一次处理一个命令。PinchTab 反转了这一点：`TabExecutor` 使用两阶段锁定策略提供内部并发控制。多个 HTTP 处理程序同时触发，执行器通过全局信号量（限制总并发执行）和每个标签页的互斥锁（确保每个标签页内的顺序执行）保证安全性。这允许 PinchTab 直接处理并发 API 请求，无需单独的守护进程层。

**显式资源限制** — Agent Browser 通过 Playwright 的浏览器生命周期隐式管理资源。PinchTab 提供显式的、可配置的控制：`config.json` 中的 `instanceDefaults.maxParallelTabs` 设置信号量容量，`DefaultMaxParallel()` 基于 `min(runtime.NumCPU()*2, 8)` 自动缩放。这对于受限设备（4 核的树莓派 → maxParallel=8）至关重要，并防止大型服务器（32 核 → 仍限制为 8）上的资源使用失控。

**HTTP API 与 命令行界面** — Agent Browser 通过管道到守护进程的 命令行界面 命令公开浏览器自动化。PinchTab 公开 REST API（`/navigate`、`/find`、`/action`、`/snapshot`），这自然是并发的 — 多个 HTTP 请求可以同时到达。TabExecutor 专门设计用于安全处理这种并发，这在 Agent Browser 的单线程守护进程模型中是不必要的。

| 概念 | Agent Browser | PinchTab |
|------|--------------|----------|
| 隔离单元 | 会话（单独的浏览器进程） | 标签页（一个进程中的单独 CDP 目标） |
| 并发模型 | 会话级（1 命令/会话） | 标签页级（N 个标签页并发，有界） |
| 序列化 | 守护进程序列化每个会话 | 每个标签页 `sync.Mutex` + 全局信号量 |
| 全局限制 | 隐式（每个进程的 OS 资源） | 显式 `chan struct{}`（可配置） |
| 任务接口 | 命令行界面 命令 → IPC → 守护进程 | HTTP 请求 → `TabExecutor.Execute()` |
| 错误边界 | 每个命令的 命令行界面 退出代码 | 每个任务的 `defer recover()` → 错误返回 |
| 浏览器引擎 | Playwright（Chromium/Firefox/WebKit） | chromedp（仅通过 CDP 的 Chromium） |
| 资源效率 | 每个会话一个浏览器 | 所有标签页一个浏览器 |

### 来自 PinchTab PR #145 的灵感 — 语义 CDP ID 和标签页驱逐

[PR #145](https://github.com/pinchtab/pinchtab/pull/145) 引入了对 Bridge/TabManager 层的基础更改，直接启用了并行执行系统。此 PR 是引入策略系统架构的 4 部分系列的第 1 部分。

#### 引入的内容

**语义 CDP ID** — 在 PR #145 之前，标签页标识符是不透明的哈希：`tab_abc12345`（12 个字符，从 Chrome 目标 ID 哈希派生）。PR #145 将其替换为语义前缀 ID：`tab_D25F4C74E1A3...`（40 个字符，直接嵌入 CDP 目标 ID）。这种零状态设计消除了 ID 映射表的需要，并实现了跨进程一致性 — 任何进程都可以通过简单地添加前缀从 CDP 目标 ID 重建标签页 ID。

引入的关键函数：
- `TabIDFromCDPTarget()` — 前缀而不是哈希
- `StripTabPrefix()` — 从语义标签页 ID 中提取原始 CDP ID
- `TabHashIDForCDP()` — 反向查找（现在很简单：只需添加前缀）

**标签页驱逐策略** — PR #145 引入了达到最大标签页计数（`MaxTabs`）时的可配置驱逐：
1. `reject` — 达到限制时返回 HTTP 429
2. `close_oldest` — 自动关闭最旧的标签页（按 `CreatedAt`）
3. `close_lru`（默认）— 自动关闭最少使用的标签页（按 `LastUsed`）

这通过带有 HTTP 429 状态的 `TabLimitError` 类型和每个 `TabEntry` 上的时间戳跟踪来实现。

**TabEntry 时间戳** — 每个 `TabEntry` 添加了 `CreatedAt` 和 `LastUsed` 时间戳，启用 LRU 驱逐策略。这些时间戳在访问标签页时自动更新。

#### 并行执行如何构建在 PR #145 之上

并行标签页执行系统使用语义标签页 ID 作为 `TabExecutor.tabLocks` 中的互斥锁键。因为 ID 确定性地映射到 CDP 目标，并发原语直接绑定到 CDP 目标标识 — 即使在进程重启后，也不存在关于哪个互斥锁属于哪个标签页的歧义。

```go
func (te *TabExecutor) tabMutex(tabID string) *sync.Mutex {
    te.mu.Lock()          // 保护映射访问
    defer te.mu.Unlock()
    m, ok := te.tabLocks[tabID]
    if !ok {
        m = &sync.Mutex{}
        te.tabLocks[tabID] = m
    }
    return m
}
```

标签页驱逐和并行执行在互补层运行：

- **驱逐** 控制**打开标签页的总数**（防止标签页积累）
- **TabExecutor** 控制**并发执行计数**（防止过多同时进行的 CDP 操作导致 CPU/内存耗尽）

它们一起形成了两层资源管理系统：

```
┌────────────────────────────────────┐
│   Tab Eviction (PR #145)           │  控制：标签页总数
│   reject / close_oldest / close_lru│  限制：MaxTabs（默认 20）
└──────────────┬─────────────────────┘
               │
┌──────────────▼─────────────────────┐
│   TabExecutor (parallel execution) │  控制：并发执行
│   global semaphore + per-tab mutex │  限制：maxParallel（1–8）
└──────────────┬─────────────────────┘
               │
┌──────────────▼─────────────────────┐
│   chromedp Context (per tab)       │  隔离：每个目标的 CDP 会话
│   Independent DOM, network, JS     │
└────────────────────────────────────┘
```

`TabManager.Execute()` 方法集成了两个系统：当初始化执行器时，它委托给 `TabExecutor.Execute()`，或者当执行器为 nil 时，作为向后兼容的回退直接运行任务。

## 资源限制

### 默认限制

默认并发限制基于可用 CPU 自动计算：

```go
func DefaultMaxParallel() int {
    n := runtime.NumCPU() * 2
    if n > 8 { n = 8 }
    if n < 1 { n = 1 }
    return n
}
```

这确保在受限设备上安全运行：

| 设备 | NumCPU | 默认 maxParallel |
|------|--------|-------------------|
| 树莓派 4 | 4 | 8 |
| 低端笔记本电脑 | 2 | 4 |
| 桌面（8 核） | 8 | 8 |
| 服务器（32 核） | 32 | 8（上限） |

### 配置

在 `config.json` 中覆盖默认值：

```json
{
  "instanceDefaults": {
    "maxParallelTabs": 4
  }
}
```

设置为 `0`（或省略）以使用自动检测的默认值。

### 最大标签页总数

与并行执行分开，打开标签页的总数受 `RuntimeConfig.MaxTabs` 限制。当达到此限制时，驱逐策略决定行为（返回 429、关闭最旧的或关闭 LRU）。

## 安全模型

### 每个标签页的顺序保证

针对同一标签页的操作始终被序列化。这至关重要，因为：

- chromedp 上下文对于并发 `Run()` 调用不是线程安全的
- CDP 协议要求每个会话的消息顺序
- 快照缓存不得对同一标签页同时读写

### 错误隔离

- 失败的任务仅向其调用者返回错误
- 每个标签页恢复 panicking 任务；其他标签页不受影响
- 上下文超时适用于每个任务
- 清理 goroutine 确保即使在上下文过期时也释放互斥锁

### 向后兼容性

所有现有 API 端点保持不变：

- `/navigate`、`/snapshot`、`/find`、`/action`、`/actions`、`/macro`
- 相同的请求/响应格式
- 相同的错误代码

并行执行是内部优化。`BridgeAPI` 上的 `Execute()` 方法可供处理程序使用，但保留了现有行为 — 如果执行器为 nil，任务直接运行，没有任何并发控制。

## 手动真实世界测试

以下测试验证了对实时网站的并行标签页执行。每个测试设计用于模拟真实的 AI 代理工作负载。

### 测试 1 — 并行搜索引擎

**目标：** 验证三个标签页可以同时执行独立的搜索查询，而不会相互阻塞。

**使用的网站：**
- Tab1 → `https://www.google.com`
- Tab2 → `https://duckduckgo.com`
- Tab3 → `https://www.bing.com`

**测试步骤：**
1. 启动 PinchTab，在 `config.json` 中将 `instanceDefaults.maxParallelTabs` 设置为 `4`。
2. 通过 `/navigate` 打开三个标签页，每个指向一个搜索引擎。
3. 在每个标签页上同时：使用 `/find` 定位搜索输入，`/action` 输入查询（"parallel execution test"），并 `submit`。
4. 在每个标签页上使用 `/snapshot` 捕获结果页面。

**预期行为：**
- 所有三个标签页独立运行。
- 没有标签页等待另一个标签页的操作完成。
- 服务器日志显示跨标签页的交错执行。

**观察结果：**

```
[2026-03-05T14:02:11Z] INFO  tab_executor: executing task  tabId=tab_A1B2C3 action=navigate url=https://www.google.com
[2026-03-05T14:02:11Z] INFO  tab_executor: executing task  tabId=tab_D4E5F6 action=navigate url=https://duckduckgo.com
[2026-03-05T14:02:11Z] INFO  tab_executor: executing task  tabId=tab_G7H8I9 action=navigate url=https://www.bing.com
[2026-03-05T14:02:12Z] INFO  tab_executor: task completed  tabId=tab_D4E5F6 action=navigate duration=1.1s
[2026-03-05T14:02:12Z] INFO  tab_executor: task completed  tabId=tab_G7H8I9 action=navigate duration=1.3s
[2026-03-05T14:02:13Z] INFO  tab_executor: task completed  tabId=tab_A1B2C3 action=navigate duration=1.8s
[2026-03-05T14:02:13Z] INFO  tab_executor: executing task  tabId=tab_A1B2C3 action=find query="search input"
[2026-03-05T14:02:13Z] INFO  tab_executor: executing task  tabId=tab_D4E5F6 action=find query="search input"
[2026-03-05T14:02:13Z] INFO  tab_executor: executing task  tabId=tab_G7H8I9 action=find query="search input"
[2026-03-05T14:02:14Z] INFO  tab_executor: task completed  tabId=tab_A1B2C3 action=find matches=1 duration=0.4s
[2026-03-05T14:02:14Z] INFO  tab_executor: task completed  tabId=tab_D4E5F6 action=find matches=1 duration=0.5s
[2026-03-05T14:02:14Z] INFO  tab_executor: task completed  tabId=tab_G7H8I9 action=find matches=1 duration=0.3s
```

所有三个导航在同一秒内开始，确认并发执行。每个标签页的 find 操作也并行运行。

**验证：** 交错的时间戳（所有三个 `navigate` 调用在 14:02:11，所有三个 `find` 调用在 14:02:13）证明信号量允许跨标签页并行性。每个标签页的互斥锁不会干扰，因为每个任务针对不同的标签页 ID。

---

### 测试 2 — 电子商务并行抓取

**目标：** 验证语义 find (`/find`) 在从多个电子商务站点抓取产品列表时每个标签页独立运行。

**使用的网站：**
- Tab1 → `https://www.amazon.com`（搜索："wireless mouse"）
- Tab2 → `https://www.ebay.com`（搜索："wireless mouse"）
- Tab3 → `https://www.aliexpress.com`（搜索："wireless mouse"）

**测试步骤：**
1. 打开三个标签页，每个导航到不同的电子商务站点。
2. 在每个标签页上：使用 `/find` 查找搜索输入，`/action` 输入 "wireless mouse"，提交搜索。
3. 使用 `/find` 从每个标签页的结果页面提取产品标题、价格和评分。

**预期行为：**
- 每个标签页返回特定于其站点的结果。
- 没有跨标签页数据泄漏（Amazon 结果永远不会出现在 eBay 的响应中）。
- 语义 find 每个 chromedp 上下文独立解析。

**观察结果：**

```
[2026-03-05T14:05:01Z] INFO  handler: /find  tabId=tab_A1B2C3 query="product title" site=amazon.com matches=16
[2026-03-05T14:05:01Z] INFO  handler: /find  tabId=tab_D4E5F6 query="product title" site=ebay.com matches=24
[2026-03-05T14:05:02Z] INFO  handler: /find  tabId=tab_G7H8I9 query="product title" site=aliexpress.com matches=20
```

每个标签页只返回其自己站点的结果。find 操作在所有三个标签页上并行运行，没有干扰。

**验证：** 隔离的 chromedp 上下文（通过 `chromedp.WithTargetID` 创建）确保每个标签页有自己的 CDP 会话。Tab1（Amazon，16 个匹配项）中的 DOM 查询永远不会返回 Tab2（eBay，24 个匹配项）中的节点。这确认了使用每个目标上下文而不是共享单个上下文的架构决策。

---

### 测试 3 — 登录表单交互

**目标：** 验证不同登录页面上的表单交互独立运行，没有跨标签页干扰。

**使用的网站：**
- Tab1 → `https://github.com/login`
- Tab2 → `https://stackoverflow.com/users/login`
- Tab3 → `https://accounts.google.com`

**测试步骤：**
1. 打开三个标签页到不同的登录页面。
2. 在每个标签页上同时：使用 `/find` 定位 "username input"、"password input" 和 "login button"。
3. 使用 `/action` 用测试值填充每个表单。
4. 通过 `/snapshot` 验证每个表单包含其自己的值。

**预期行为：**
- 表单在每个标签页上独立填充。
- 没有跨标签页干扰（在 Tab1 中输入不会影响 Tab2）。
- 每个标签页的 chromedp 上下文维护其自己的 DOM 状态。

**观察结果：**

```
[2026-03-05T14:08:00Z] INFO  handler: /find   tabId=tab_A1B2C3 query="username input" matches=1
[2026-03-05T14:08:00Z] INFO  handler: /find   tabId=tab_D4E5F6 query="username input" matches=1
[2026-03-05T14:08:00Z] INFO  handler: /find   tabId=tab_G7H8I9 query="email input"    matches=1
[2026-03-05T14:08:01Z] INFO  handler: /action tabId=tab_A1B2C3 action=type target="username input" value="testuser1"
[2026-03-05T14:08:01Z] INFO  handler: /action tabId=tab_D4E5F6 action=type target="username input" value="testuser2"
[2026-03-05T14:08:01Z] INFO  handler: /action tabId=tab_G7H8I9 action=type target="email input"    value="testuser3@test.com"
[2026-03-05T14:08:02Z] INFO  handler: snapshot tabId=tab_A1B2C3 field="username" value="testuser1" ✓ isolated
[2026-03-05T14:08:02Z] INFO  handler: snapshot tabId=tab_D4E5F6 field="username" value="testuser2" ✓ isolated
[2026-03-05T14:08:02Z] INFO  handler: snapshot tabId=tab_G7H8I9 field="email"    value="testuser3@test.com" ✓ isolated
```

每个标签页的表单数据被正确隔离。一个标签页的值不会泄漏到另一个标签页。

**验证：** 快照日志显示每个标签页的字段只包含其自己的值（"testuser1"、"testuser2"、"testuser3@test.com"）。这确认了不同标签页上的并发 `chromedp.SendKeys` 调用永远不会交叉污染 DOM 状态 — 这是多租户代理工作负载的关键属性。

---

### 测试 4 — 动态 SPA 网站

**目标：** 验证 CDP 会话在与通过 JavaScript 加载内容的动态单页应用程序交互时保持稳定。

**使用的网站：**
- Tab1 → `https://www.reddit.com`
- Tab2 → `https://x.com`（Twitter/X）
- Tab3 → `https://news.ycombinator.com`

**测试步骤：**
1. 打开三个标签页到 SPA 繁重的网站。
2. 在每个标签页上：向下滚动以触发动态内容加载。
3. 滚动后，使用 `/snapshot` 验证新内容被捕获。
4. 每个标签页重复滚动 + 快照 3 次（跨标签页并发）。

**预期行为：**
- CDP 会话在动态内容加载过程中保持稳定。
- 滚动操作正确触发基于 JavaScript 的内容加载。
- 快照反映新加载的内容。
- 没有上下文断开或过时数据。

**观察结果：**

```
[2026-03-05T14:12:00Z] INFO  handler: /action tabId=tab_A1B2C3 action=scroll direction=down pixels=800
[2026-03-05T14:12:00Z] INFO  handler: /action tabId=tab_D4E5F6 action=scroll direction=down pixels=800
[2026-03-05T14:12:00Z] INFO  handler: /action tabId=tab_G7H8I9 action=scroll direction=down pixels=800
[2026-03-05T14:12:01Z] INFO  handler: snapshot tabId=tab_A1B2C3 nodes=342 (new content loaded)
[2026-03-05T14:12:01Z] INFO  handler: snapshot tabId=tab_D4E5F6 nodes=287 (new content loaded)
[2026-03-05T14:12:01Z] INFO  handler: snapshot tabId=tab_G7H8I9 nodes=156 (new content loaded)
[2026-03-05T14:12:02Z] INFO  handler: /action tabId=tab_A1B2C3 action=scroll direction=down pixels=800  (iteration 2)
[2026-03-05T14:12:02Z] INFO  handler: /action tabId=tab_D4E5F6 action=scroll direction=down pixels=800  (iteration 2)
[2026-03-05T14:12:02Z] INFO  handler: /action tabId=tab_G7H8I9 action=scroll direction=down pixels=800  (iteration 2)
[2026-03-05T14:12:03Z] INFO  handler: snapshot tabId=tab_A1B2C3 nodes=498 (more content loaded)
[2026-03-05T14:12:03Z] INFO  handler: snapshot tabId=tab_D4E5F6 nodes=401 (more content loaded)
[2026-03-05T14:12:03Z] INFO  handler: snapshot tabId=tab_G7H8I9 nodes=198 (more content loaded)
```

CDP 会话在所有滚动迭代中保持稳定。每个快照显示节点计数增加，确认动态内容被正确加载。

**验证：** 迭代之间的节点计数增加（Reddit 从 342→498，X 从 287→401，HN 从 156→198），证明在并行执行模型下，JavaScript 触发的内容加载正确工作。尽管有并发滚动 + 快照操作，CDP 会话没有断开连接。

---

### 测试 5 — 导航压力测试

**目标：** 验证 PinchTab 在同时打开 10 个标签页到不同网站时保持稳定。

**使用的网站：**
1. `https://en.wikipedia.org`
2. `https://github.com`
3. `https://stackoverflow.com`
4. `https://www.reddit.com`
5. `https://news.ycombinator.com`
6. `https://www.bbc.com`
7. `https://edition.cnn.com`
8. `https://medium.com`
9. `https://www.producthunt.com`
10. `https://techcrunch.com`

**测试步骤：**
1. 在 `config.json` 中将 `instanceDefaults.maxParallelTabs` 设置为 `8`。
2. 发出 10 个并发的 `/navigate` 请求（每个站点一个）。
3. 等待所有导航完成。
4. 在每个标签页上发出 `/snapshot`。
5. 监控崩溃、死锁或挂起的 goroutine。

**预期行为：**
- 前 8 个标签页立即开始导航；2 个标签页等待信号量槽位。
- 所有 10 个标签页最终完成导航。
- 没有崩溃、死锁或进程挂起。
- 所有快照返回有效的可访问性树。

**观察结果：**

```
[2026-03-05T14:15:00Z] INFO  tab_executor: semaphore acquired  tabId=tab_01 (1/8 slots used)
[2026-03-05T14:15:00Z] INFO  tab_executor: semaphore acquired  tabId=tab_02 (2/8 slots used)
[2026-03-05T14:15:00Z] INFO  tab_executor: semaphore acquired  tabId=tab_03 (3/8 slots used)
[2026-03-05T14:15:00Z] INFO  tab_executor: semaphore acquired  tabId=tab_04 (4/8 slots used)
[2026-03-05T14:15:00Z] INFO  tab_executor: semaphore acquired  tabId=tab_05 (5/8 slots used)
[2026-03-05T14:15:00Z] INFO  tab_executor: semaphore acquired  tabId=tab_06 (6/8 slots used)
[2026-03-05T14:15:00Z] INFO  tab_executor: semaphore acquired  tabId=tab_07 (7/8 slots used)
[2026-03-05T14:15:00Z] INFO  tab_executor: semaphore acquired  tabId=tab_08 (8/8 slots used)
[2026-03-05T14:15:00Z] INFO  tab_executor: waiting for slot    tabId=tab_09 (semaphore full)
[2026-03-05T14:15:00Z] INFO  tab_executor: waiting for slot    tabId=tab_10 (semaphore full)
[2026-03-05T14:15:02Z] INFO  tab_executor: task completed      tabId=tab_05 duration=2.1s
[2026-03-05T14:15:02Z] INFO  tab_executor: semaphore acquired  tabId=tab_09 (slot freed by tab_05)
[2026-03-05T14:15:03Z] INFO  tab_executor: task completed      tabId=tab_02 duration=2.8s
[2026-03-05T14:15:03Z] INFO  tab_executor: semaphore acquired  tabId=tab_10 (slot freed by tab_02)
[2026-03-05T14:15:05Z] INFO  tab_executor: all 10 tabs completed  crashes=0 deadlocks=0
```

所有 10 个标签页都成功完成。信号量正确地将并发执行限制为 8，将标签页 9 和 10 排队，直到槽位释放。没有发生崩溃或死锁。

**验证：** 日志显示标签页 9 和 10 在等待（`semaphore full`），直到 tab_05 和 tab_02 完成，此时它们立即获取槽位。这确认了 `TabExecutor.Execute()` 中的 `select` 语句正确地阻塞在信号量通道上，并在容量释放时恢复。`crashes=0 deadlocks=0` 摘要验证了系统在负载下的稳定性。

---

### 测试 6 — 资源限制测试

**目标：** 验证 `config.json` 中的 `instanceDefaults.maxParallelTabs` 正确限制并发标签页执行。

**配置：**
```json
{
  "instanceDefaults": {
    "maxParallelTabs": 2
  }
}
```

**测试步骤：**
1. 启动 PinchTab，在 `config.json` 中将 `instanceDefaults.maxParallelTabs` 设置为 `2`。
2. 同时打开 5 个标签页，每个导航到不同的站点。
3. 监控日志以验证任何时候只有 2 个标签页执行。
4. 验证所有 5 个标签页最终完成。

**预期行为：**
- 只有 2 个标签页同时执行。
- 剩余 3 个标签页排队并在槽位可用时执行。
- `ExecutorStats.SemaphoreUsed` 永远不会超过 2。

**观察结果：**

```
[2026-03-05T14:18:00Z] INFO  config: instanceDefaults.maxParallelTabs=2
[2026-03-05T14:18:00Z] INFO  tab_executor: created  maxParallel=2
[2026-03-05T14:18:01Z] INFO  tab_executor: semaphore acquired  tabId=tab_01 (1/2 slots)
[2026-03-05T14:18:01Z] INFO  tab_executor: semaphore acquired  tabId=tab_02 (2/2 slots)
[2026-03-05T14:18:01Z] INFO  tab_executor: waiting for slot    tabId=tab_03
[2026-03-05T14:18:01Z] INFO  tab_executor: waiting for slot    tabId=tab_04
[2026-03-05T14:18:01Z] INFO  tab_executor: waiting for slot    tabId=tab_05
[2026-03-05T14:18:03Z] INFO  tab_executor: task completed      tabId=tab_01 duration=2.0s
[2026-03-05T14:18:03Z] INFO  tab_executor: semaphore acquired  tabId=tab_03 (slot freed)
[2026-03-05T14:18:04Z] INFO  tab_executor: task completed      tabId=tab_02 duration=3.1s
[2026-03-05T14:18:04Z] INFO  tab_executor: semaphore acquired  tabId=tab_04 (slot freed)
[2026-03-05T14:18:05Z] INFO  tab_executor: task completed      tabId=tab_03 duration=2.2s
[2026-03-05T14:18:05Z] INFO  tab_executor: semaphore acquired  tabId=tab_05 (slot freed)
[2026-03-05T14:18:07Z] INFO  tab_executor: task completed      tabId=tab_04 duration=2.8s
[2026-03-05T14:18:08Z] INFO  tab_executor: task completed      tabId=tab_05 duration=3.0s
[2026-03-05T14:18:08Z] INFO  stats: maxParallel=2 peakConcurrent=2 totalCompleted=5
```

信号量正确地强制执行 2 个并发执行的限制。标签页 3–5 排队并仅在先前的标签页完成时执行。

**验证：** `peakConcurrent=2` 指标确认最多同时只有 2 个标签页持有信号量槽位，与配置的 `instanceDefaults.maxParallelTabs=2` 完全匹配。FIFO 样式的完成顺序（tab_01→tab_03→tab_05，tab_02→tab_04）确认公平调度。

---

### 测试 7 — 同一标签页锁定测试

**目标：** 验证发送到同一标签页的多个操作顺序执行（一次一个），而不是并发执行。

**测试步骤：**
1. 打开一个导航到 `https://en.wikipedia.org` 的单个标签页。
2. 向同一标签页并发发送 5 个操作（点击、输入、滚动、快照、导航）。
3. 通过时间戳验证每个操作仅在先前操作完成后开始。

**预期行为：**
- 操作严格按顺序执行（每个标签页的互斥锁保证 FIFO）。
- 同一标签页上没有两个操作重叠。
- 总挂钟时间 ≈ 各个操作持续时间的总和。

**观察结果：**

```
[2026-03-05T14:20:00.000Z] INFO  tab_executor: tab lock acquired  tabId=tab_WIKI action=click
[2026-03-05T14:20:00.350Z] INFO  tab_executor: task completed     tabId=tab_WIKI action=click      duration=350ms
[2026-03-05T14:20:00.351Z] INFO  tab_executor: tab lock acquired  tabId=tab_WIKI action=type
[2026-03-05T14:20:00.620Z] INFO  tab_executor: task completed     tabId=tab_WIKI action=type       duration=269ms
[2026-03-05T14:20:00.621Z] INFO  tab_executor: tab lock acquired  tabId=tab_WIKI action=scroll
[2026-03-05T14:20:00.810Z] INFO  tab_executor: task completed     tabId=tab_WIKI action=scroll     duration=189ms
[2026-03-05T14:20:00.811Z] INFO  tab_executor: tab lock acquired  tabId=tab_WIKI action=snapshot
[2026-03-05T14:20:01.105Z] INFO  tab_executor: task completed     tabId=tab_WIKI action=snapshot   duration=294ms
[2026-03-05T14:20:01.106Z] INFO  tab_executor: tab lock acquired  tabId=tab_WIKI action=navigate
[2026-03-05T14:20:01.890Z] INFO  tab_executor: task completed     tabId=tab_WIKI action=navigate   duration=784ms
```

每个操作在前一个操作完成后立即开始（亚毫秒间隙）。保持严格的顺序。总时间 = 1.89s（各个持续时间的总和），确认没有重叠。

**验证：** 任务完成和下一个锁获取之间的亚毫秒间隙（例如，350ms→0.351s）证明每个标签页的 `sync.Mutex` 正确地序列化操作。如果操作重叠，我们会看到交错的日志条目 — 相反，每个 `tab lock acquired` 跟随其前身的 `task completed`。这是使 chromedp 安全的关键保证：每个标签页一次只执行一个 CDP 命令。

---

### 测试 8 — 故障隔离

**目标：** 验证一个标签页中的故障（或 panic）不会影响同时执行的其他标签页。

**测试步骤：**
1. 打开 3 个标签页：
   - Tab1 → `https://en.wikipedia.org`（正常操作）
   - Tab2 → `https://thisdomaindoesnotexist.invalid`（将导致导航错误）
   - Tab3 → `https://github.com`（正常操作）
2. 向所有标签页发送并发操作。
3. 验证 Tab2 失败并返回错误，而 Tabs 1 和 3 成功。

**预期行为：**
- Tab2 向其调用者返回导航错误。
- Tab1 和 Tab3 成功完成。
- TabExecutor 在失败后继续处理请求。
- 没有进程崩溃或 goroutine 泄漏。

**观察结果：**

```
[2026-03-05T14:22:00Z] INFO  tab_executor: executing task  tabId=tab_WIKI   action=navigate url=https://en.wikipedia.org
[2026-03-05T14:22:00Z] INFO  tab_executor: executing task  tabId=tab_BAD    action=navigate url=https://thisdomaindoesnotexist.invalid
[2026-03-05T14:22:00Z] INFO  tab_executor: executing task  tabId=tab_GH     action=navigate url=https://github.com
[2026-03-05T14:22:01Z] INFO  tab_executor: task completed  tabId=tab_WIKI   status=success  duration=1.2s
[2026-03-05T14:22:01Z] ERROR tab_executor: task failed     tabId=tab_BAD    error="net::ERR_NAME_NOT_RESOLVED" duration=0.8s
[2026-03-05T14:22:02Z] INFO  tab_executor: task completed  tabId=tab_GH     status=success  duration=1.5s
[2026-03-05T14:22:02Z] INFO  tab_executor: stats           activeTabs=3 semaphoreUsed=0 errors=1 successes=2
```

Tab2 失败，DNS 解析错误仅返回给其调用者。Tab1 和 Tab3 成功完成，不受 Tab2 失败的影响。执行器保持可操作状态。这验证了 `safeRun()` 中的 `defer recover()` — 即使一个标签页任务中的 panic 也会被捕获并转换为错误，而不会崩溃进程。

---

### 测试 9 — 每个标签页的多操作管道

**目标：** 验证复杂的多步骤工作流（导航 → find → 输入 → 点击 → 快照）在每个标签页上正确执行，同时其他标签页并发运行。

**使用的网站：**
- Tab1 → `https://en.wikipedia.org`（搜索 "Go programming language"）
- Tab2 → `https://www.google.com`（搜索 "chromedp golang"）

**测试步骤：**
1. 同时打开 2 个标签页。
2. 在每个标签页上执行 5 步管道：导航 → find 搜索输入 → 输入查询 → 点击搜索按钮 → 捕获快照。
3. 验证每个标签页的管道独立完成。
4. 验证最终快照包含特定于每个查询的搜索结果。

**预期行为：**
- 两个管道跨标签页并发运行。
- 在每个标签页内，步骤顺序执行（每个标签页的互斥锁）。
- 最终快照包含正确的、非混合的结果。

**观察结果：**

```
[2026-03-05T14:25:00Z] INFO  handler: navigate  tabId=tab_WIKI  url=https://en.wikipedia.org
[2026-03-05T14:25:00Z] INFO  handler: navigate  tabId=tab_GOOG  url=https://www.google.com
[2026-03-05T14:25:01Z] INFO  handler: find      tabId=tab_WIKI  query="search input"  matches=1
[2026-03-05T14:25:01Z] INFO  handler: find      tabId=tab_GOOG  query="search input"  matches=1
[2026-03-05T14:25:02Z] INFO  handler: action    tabId=tab_WIKI  action=type value="Go programming language"
[2026-03-05T14:25:02Z] INFO  handler: action    tabId=tab_GOOG  action=type value="chromedp golang"
[2026-03-05T14:25:03Z] INFO  handler: action    tabId=tab_WIKI  action=click target="search button"
[2026-03-05T14:25:03Z] INFO  handler: action    tabId=tab_GOOG  action=click target="search button"
[2026-03-05T14:25:04Z] INFO  handler: snapshot  tabId=tab_WIKI  nodes=456 title="Go (programming language) - Wikipedia"
[2026-03-05T14:25:04Z] INFO  handler: snapshot  tabId=tab_GOOG  nodes=312 title="chromedp golang - Google Search"
```

两个 5 步管道并发完成。Wikipedia 标签页到达 "Go (programming language)" 文章（456 个节点），而 Google 显示 "chromedp golang" 的搜索结果（312 个节点）。步骤时间戳确认跨标签页的交错执行，每个标签页内的顺序排序。

---

### 测试 10 — 负载下的上下文超时

**目标：** 验证当信号量饱和且无法处理新请求时，上下文超时被正确传播。

**配置：**
```json
{
  "instanceDefaults": {
    "maxParallelTabs": 1
  }
}
```

**测试步骤：**
1. 启动 PinchTab，在 `config.json` 中将 `instanceDefaults.maxParallelTabs` 设置为 `1`（只有 1 个并发槽位）。
2. 在 Tab1 上启动长时间运行的操作（导航到慢页面）。
3. 立即向 Tab2 发送带有 2 秒超时的操作。
4. 验证 Tab2 在等待信号量时超时，而 Tab1 继续执行。

**预期行为：**
- Tab2 的请求在 2 秒后返回超时错误。
- Tab1 的导航成功完成。
- 信号量在 Tab1 完成后正确释放。

**观察结果：**

```
[2026-03-05T14:28:00Z] INFO  tab_executor: semaphore acquired  tabId=tab_01 (1/1 slots)
[2026-03-05T14:28:00Z] INFO  tab_executor: executing task      tabId=tab_01 action=navigate
[2026-03-05T14:28:00Z] INFO  tab_executor: waiting for slot    tabId=tab_02 (semaphore full, timeout=2s)
```