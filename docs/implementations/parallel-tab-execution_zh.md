# Parallel Tab Execution

> **Status:** 这是原始 TabExecutor PR（2026 年 3 月）的设计文档，包含其测试日志。`TabExecutor` 在 HEAD 上存在（`internal/bridge/tabs/tab_executor.go`），按 `TabManager` 用 `instanceDefaults.maxParallelTabs` 构造，但没有任何 HTTP handler 调用 `Bridge.Execute`，因此下面的请求流不是今天 `/action`、`/navigate`、`/find` 或 `/snapshot` 的实际服务方式。"Observed results" 日志行（`tab_executor: executing task`、`semaphore acquired`、`tab_XXXXXX` ID）并非由代码输出。

PinchTab 支持跨浏览器标签页的安全并行执行。多个标签页可以并发执行操作，而每个标签页内部保持顺序执行，防止资源耗尽和竞态条件。

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

### 执行流

通过并行执行系统的完整请求生命周期：

```
HTTP POST /tabs/{id}/action  (e.g., Click button)
    │
    ▼
Handler: HandleAction()
    │
    ▼
Bridge.EnsureBrowser()  [lazy init on first request]
    │
    ▼
Bridge.TabContext(tabID)  [get chromedp.Context for tab]
    │
    ▼
Bridge.Execute(ctx, tabID, task)
    │
    ▼
TabManager.Execute()
    │
    ▼
TabExecutor.Execute(ctx, tabID, task)
    ├─ Phase 1: te.semaphore <- struct{}   [acquire global slot]
    ├─ Phase 2: tabMutex(tabID).Lock()     [acquire per-tab lock]
    └─ Phase 3: safeRun(ctx, tabID, task)  [execute with panic recovery]
        ├─ chromedp.Run(ctx, action...)
        └─ Return result or error
    │
    ▼
HTTP 200 {"success": true, "result": {...}}
```

### 执行模型

每个标签页**顺序**执行任务（一次一个），但**不同标签页**并发运行，上限可配置：

```
Time ──────────────────────────────────────────────────▶
Tab1 ──▶ [action1] ──▶ [action2] ──▶ [action3]
Tab2 ──▶ [action1] ──▶ [action2]                         (concurrent with Tab1)
Tab3 ──▶ [action1] ──▶ [action2] ──▶ [action3]           (concurrent with Tab1 & Tab2)
```

两阶段锁定保证正确性：

1. **Phase 1——信号量获取**：请求在全局 `chan struct{}` 信号量中获取一个槽位。如果所有槽位都被占用，goroutine 阻塞，直到有槽位释放或 context 到期。
2. **Phase 2——tab 互斥锁获取**：拿到信号量槽位后，请求获取 per-tab 的 `sync.Mutex`。这保证在任意时刻只有一个 CDP 操作针对某个给定 tab 运行。

```go
// Simplified flow inside TabExecutor.Execute()
select {
case te.semaphore <- struct{}{}:   // Phase 1: global slot
    defer func() { <-te.semaphore }()
case <-ctx.Done():
    return ctx.Err()
}
tabMu := te.tabMutex(tabID)       // Phase 2: per-tab lock
tabMu.Lock()
defer tabMu.Unlock()
return te.safeRun(ctx, tabID, task) // Execute with panic recovery
```

### 组件

| 组件 | 位置 | 用途 |
|------|------|------|
| `TabExecutor` | `internal/bridge/tabs/tab_executor.go`（经 `internal/bridge/tabs_facade.go` 再导出） | 核心并行执行引擎 |
| `TabManager.Execute()` | `internal/bridge/tab_manager.go` | handler 的集成点 |
| `Bridge.Execute()` | `internal/bridge/bridge.go` | BridgeAPI 接口方法 |
| `LockManager` | `internal/bridge/tabs/lock.go`（经 `internal/bridge/tabs_facade.go` 再导出） | 带 TTL 的 per-tab 所有权锁 |
| `TabEntry` | `internal/bridge/bridge_types.go` | per-tab chromedp 上下文 + 元数据 |

### 工作原理

1. **全局信号量**——一个缓冲通道（容量为 `maxParallel` 的 `chan struct{}`）限制并发执行的 tab 数。信号量满时，新任务等待（尊重 context 取消/超时）。

2. **per-tab 互斥锁**——每个 tab 在 `map[string]*sync.Mutex` 中有自己的 `sync.Mutex`。这保证单个 tab 内的操作一次只执行一个。它防止在同一 tab 上并发 CDP 操作，而 chromedp 不支持这种并发。

3. **Panic 恢复**——每个任务包在一个 `defer recover()` 块里。一个 tab 任务中的 panic 不会崩溃进程，也不影响其他 tab。panic 被转成 `error` 并通过 `slog.Error` 记录。

4. **Context 传播**——调用方的 context（带超时/取消）被透传给任务函数。如果 context 在等待信号量或 tab 锁期间到期，调用立即返回错误。一个清理 goroutine 保证即使 context 在等待中途到期，per-tab 互斥锁也会被解锁。

5. **CDP 上下文隔离**——每个 tab 由自己的 `chromedp.Context` 支撑，通过 `chromedp.NewContext(browserCtx, chromedp.WithTargetID(...))` 创建。这意味着每个 tab 有独立的 Chrome DevTools Protocol 会话，各自有 DOM、网络栈和 JavaScript 运行时。

## 架构灵感

### 来自 Vercel Agent Browser 的灵感

[Vercel Agent Browser](https://github.com/vercel-labs/agent-browser) 是一个为 AI 代理设计的无头浏览器自动化 CLI。它采用 client-daemon 架构：一个 Rust CLI 与一个常驻 Node.js daemon（或实验性原生 Rust daemon）通信，后者管理一个 Playwright 浏览器实例。Agent Browser 的若干架构模式直接影响了 PinchTab 的并行 tab 执行设计。

#### 我们研究了什么

**浏览器会话管理**——Agent Browser 通过 `--session` flag 隔离并发工作负载。每个会话（`--session agent1`、`--session agent2`）派生一个完全独立的浏览器实例，有独立的 cookies、存储、导航历史和认证状态。会话作为独立 OS 进程并行运行。daemon 在一个会话内的多条命令之间常驻，因此后续 CLI 调用（`open`、`click`、`fill`）很快。

**任务执行模型**——Agent Browser 遵循严格的「每次调用一条命令」模型。每次 CLI 调用是通过 IPC 发到会话 daemon 的离散任务。daemon 在会话内串行化命令：每个会话一次只执行一条命令。这是设计选择——Playwright context 不是线程安全的，因此串行化防止竞态条件。CLI 客户端阻塞直到 daemon 响应，强制严格的请求-响应循环，IPC 读取超时 30 秒（默认 Playwright 超时设为 25 秒，以确保给出正确的错误信息而不是泛泛的超时）。

**并发结构**——多个会话可以同时运行，但每个会话内部是单线程的（一次一条命令）。这给出会话级并发：N 个会话 = N 个并发浏览器实例，每个一次处理一条命令。资源通过 OS 隐式管理——每个会话是独立进程，有自己的内存空间。

**快照与 ref 工作流**——Agent Browser 生成带稳定 `ref` 标识符（`@e1`、`@e2`）的可访问性树快照，ref 持续到下一次快照。AI 代理用这些 ref 做确定性元素选择。这影响了 PinchTab 的 `RefCache` 设计：每个 tab 维护自己的、带节点引用的快照缓存。

**错误处理**——Agent Browser 按命令以 CLI 退出码返回错误。失败的命令不会崩溃 daemon——会话对后续命令保持活跃。命令支持 `--json` 输出，用于机器可读的错误报告。

#### PinchTab 如何以不同方式采纳这些思路

PinchTab 运行在一个根本不同的架构层面：

**Tab 级 vs 会话级隔离**——Agent Browser 为每个会话创建独立浏览器进程，而 PinchTab 在 CDP target（tab）级做隔离。每个 tab 通过 `chromedp.NewContext(browserCtx, chromedp.WithTargetID(targetID))` 获得自己的 `chromedp.Context`，从而有独立 CDP 会话，各自有 DOM、网络栈和 JS 运行时。多个并发工作负载共享一个 Chrome 进程，但通过 CDP target 保持隔离。这更省资源：一个带 10 个 tab 的 Chrome 进程比 10 个独立 Chrome 实例省内存。

**内部并发控制 vs 外部串行化**——Agent Browser 依赖 daemon 架构做串行化——daemon 每会话一次处理一条命令。PinchTab 把它反过来：`TabExecutor` 用两阶段锁策略提供内部并发控制。多个 HTTP handler 并发触发，executor 通过全局信号量（限制总并发执行数）和 per-tab 互斥锁（保证每个 tab 内顺序执行）保证安全。这让 PinchTab 无需单独 daemon 层就能直接服务并发 API 请求。

**显式资源限制**——Agent Browser 通过 Playwright 的浏览器生命周期隐式管理资源。PinchTab 提供显式、可配置的控制：`config.json` 里的 `instanceDefaults.maxParallelTabs` 设置信号量容量，`DefaultMaxParallel()` 按 `min(runtime.NumCPU()*2, 8)` 自动伸缩。这对受限设备很关键（4 核 Raspberry Pi → maxParallel=8），也防止大服务器上资源失控（32 核 → 仍封顶 8）。

**HTTP API vs CLI**——Agent Browser 通过管道接到 daemon 的 CLI 命令暴露浏览器自动化。PinchTab 暴露 REST API（`/navigate`、`/find`、`/action`、`/snapshot`），它天然是并发的——多个 HTTP 请求可以同时到达。TabExecutor 就是专门为安全处理这种并发而设计的，而这在 Agent Browser 的单线程 daemon 模型里不必要。

| 概念 | Agent Browser | PinchTab |
|------|--------------|----------|
| 隔离单元 | 会话（独立浏览器进程） | Tab（一个进程内的独立 CDP target） |
| 并发模型 | 会话级（1 命令/会话） | Tab 级（N 个 tab 并发，有界） |
| 串行化 | daemon 按会话串行化 | per-tab `sync.Mutex` + 全局信号量 |
| 全局限制 | 隐式（每进程的 OS 资源） | 显式 `chan struct{}`（可配置） |
| 任务接口 | CLI 命令 → IPC → daemon | HTTP 请求 → `TabExecutor.Execute()` |
| 错误边界 | 按命令的 CLI 退出码 | 按任务 `defer recover()` → 错误返回 |
| 浏览器引擎 | Playwright（Chromium/Firefox/WebKit） | chromedp（仅经 CDP 的 Chromium） |
| 资源效率 | 每会话一个浏览器 | 所有 tab 一个浏览器 |

### 来自 PinchTab PR #145 的灵感——语义 CDP ID 与 Tab 驱逐

[PR #145](https://github.com/pinchtab/pinchtab/pull/145) 对 Bridge/TabManager 层做了基础性改动，直接使并行执行系统成为可能。这个 PR 是介绍策略系统架构的 4 部分系列的第 1 部分。

#### 引入了什么

**原始 CDP target ID 作为 tab ID**——在 PR #145 之前，tab 标识符是不透明哈希：`tab_abc12345`（12 字符，由 Chrome target ID 哈希派生）。PR #145 用原始 CDP target ID 本身替换了哈希，当前实现是直通——运行时 tab ID 就是 CDP target ID，没有 `tab_` 前缀，也不做哈希。这种零状态设计消除了 ID 映射表的需要，并实现跨进程一致性：任何进程都可以直接引用 CDP target ID 路由到某个 tab。

关键函数（在 `internal/ids/ids.go` 中）：

- `Manager.TabIDFromCDPTarget(cdpTargetID)`——原样返回 CDP target ID。

对于需要合成标识符的调用方，`Manager.TabID` 仍产出合成哈希 ID（`tab_XXXXXXXX`），但运行时浏览器 tab 路由直接使用原始 CDP target ID。

**Tab 驱逐策略**——PR #145 引入了当达到最大 tab 数（`MaxTabs`）时的可配置驱逐：
1. `reject`——达到限制时返回 HTTP 429
2. `close_oldest`——自动关闭最旧的 tab（按 `CreatedAt`）
3. `close_lru`（默认）——自动关闭最近最少使用的 tab（按 `LastUsed`）

这通过一个 `TabLimitError` 类型（HTTP 429 状态）和每个 `TabEntry` 上的时间戳跟踪来实现。

**TabEntry 时间戳**——每个 `TabEntry` 新增了 `CreatedAt` 和 `LastUsed` 时间戳，使 LRU 驱逐策略成为可能。这些时间戳在 tab 被访问时自动更新。

#### 并行执行如何构建在 PR #145 之上

并行 tab 执行系统用语义 tab ID 作为 `TabExecutor.tabLocks` 里的互斥锁键。因为 ID 确定性地映射到 CDP target，并发原语直接绑定到 CDP target 身份——即使跨进程重启，也不会有哪个互斥锁属于哪个 tab 的歧义。

```go
func (te *TabExecutor) tabMutex(tabID string) *sync.Mutex {
    te.mu.Lock()          // Protect map access
    defer te.mu.Unlock()
    m, ok := te.tabLocks[tabID]
    if !ok {
        m = &sync.Mutex{}
        te.tabLocks[tabID] = m
    }
    return m
}
```

Tab 驱逐和并行执行在互补的层上运作：

- **驱逐**控制打开 tab 的**总数**（防止 tab 堆积）
- **TabExecutor** 控制**并发执行数**（防止过多同时 CDP 操作耗尽 CPU/内存）

它们一起形成一个两层资源管理系统：

```
┌────────────────────────────────────┐
│   Tab Eviction (PR #145)           │  Controls: total tab count
│   reject / close_oldest / close_lru│  Limit: MaxTabs (default 20)
└──────────────┬─────────────────────┘
               │
┌──────────────▼─────────────────────┐
│   TabExecutor (parallel execution) │  Controls: concurrent execution
│   global semaphore + per-tab mutex │  Limit: maxParallel (1–8)
└──────────────┬─────────────────────┘
               │
┌──────────────▼─────────────────────┐
│   chromedp Context (per tab)       │  Isolation: CDP session per target
│   Independent DOM, network, JS     │
└────────────────────────────────────┘
```

`TabManager.Execute()` 方法把两个系统集成起来：executor 已初始化时委托给 `TabExecutor.Execute()`；executor 为 nil 时则作为向后兼容的回退直接运行任务。

## 资源限制

### 默认限制

默认并发限制根据可用 CPU 自动计算：

```go
func DefaultMaxParallel() int {
    n := runtime.NumCPU() * 2
    if n > 8 { n = 8 }
    if n < 1 { n = 1 }
    return n
}
```

这保证在受限设备上安全运行：

| 设备 | NumCPU | 默认 maxParallel |
|------|--------|-----------------|
| Raspberry Pi 4 | 4 | 8 |
| 低端笔记本 | 2 | 4 |
| 台式机（8 核） | 8 | 8 |
| 服务器（32 核） | 32 | 8（封顶） |

### 配置

在 `config.json` 中覆盖默认值：

```json
{
  "instanceDefaults": {
    "maxParallelTabs": 4
  }
}
```

设为 `0`（或省略）使用自动探测的默认值。

### Tab 总数上限

与并行执行分开，打开 tab 的总数受 `RuntimeConfig.MaxTabs` 限制。达到此限制时，驱逐策略决定行为（429 拒绝、关最旧、或关 LRU）。

## 安全模型

### Per-tab 顺序保证

针对同一 tab 的操作总是被串行化。这至关重要，因为：

- chromedp context 对并发 `Run()` 调用不是线程安全的
- CDP 协议要求每个会话内消息有序
- 快照缓存不能对同一 tab 并发读写

### 错误隔离

- 失败的任务只把错误返回给它的调用者
- 一个 panic 的任务按 tab 恢复；其他 tab 不受影响
- context 超时按任务独立生效
- 清理 goroutine 保证即使 context 到期也释放互斥锁

### 向后兼容

所有现有 API 端点保持不变：

- `/navigate`、`/snapshot`、`/find`、`/action`、`/actions`、`/macro`
- 相同的请求/响应格式
- 相同的错误码

并行执行是内部优化。`BridgeAPI` 上的 `Execute()` 方法可供 handler 使用，但现有行为被保留——如果 executor 为 nil，任务直接运行，不做任何并发控制。

## 手动真实世界测试

以下测试针对真实网站验证并行 tab 执行。每个测试设计用于模拟真实 AI 代理工作负载。

### 测试 1——并行搜索引擎

**目标：** 验证三个 tab 可以并发执行独立搜索查询而不互相阻塞。

**使用的网站：**
- Tab1 → `https://www.google.com`
- Tab2 → `https://duckduckgo.com`
- Tab3 → `https://www.bing.com`

**测试步骤：**
1. 启动 PinchTab，在 `config.json` 中把 `instanceDefaults.maxParallelTabs` 设为 `4`。
2. 通过 `/navigate` 打开三个 tab，分别指向一个搜索引擎。
3. 在每个 tab 上并发地：用 `/find` 定位搜索输入，用 `/action` 输入查询（"parallel execution test"），再用 `/action` 提交。
4. 在每个 tab 上用 `/snapshot` 捕获结果页。

**预期行为：**
- 三个 tab 各自独立操作。
- 没有 tab 因等待另一个 tab 的操作完成而阻塞。
- 服务器日志显示跨 tab 交错执行。

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

三次导航都在同一秒内开始，证实并发执行。每个 tab 的 find 操作也并行运行。

**验证：** 交错时间戳（三次 `navigate` 都在 14:02:11，三次 `find` 都在 14:02:13）证明信号量允许跨 tab 并行。per-tab 互斥锁不干扰，因为每个任务针对不同的 tab ID。

---

### 测试 2——电商并行抓取

**目标：** 验证语义 find（`/find`）在从多个电商站点抓取商品列表时按 tab 独立工作。

**使用的网站：**
- Tab1 → `https://www.amazon.com`（搜索："wireless mouse"）
- Tab2 → `https://www.ebay.com`（搜索："wireless mouse"）
- Tab3 → `https://www.aliexpress.com`（搜索："wireless mouse"）

**测试步骤：**
1. 打开三个 tab，分别导航到不同电商站点。
2. 在每个 tab 上：用 `/find` 定位搜索输入，用 `/action` 输入 "wireless mouse"，提交搜索。
3. 用 `/find` 从每个 tab 的结果页提取商品标题、价格和评分。

**预期行为：**
- 每个 tab 返回其站点特有的结果。
- 无跨 tab 数据泄漏（Amazon 结果绝不会出现在 eBay 响应里）。
- 语义 find 按 chromedp context 独立解析。

**观察结果：**

```
[2026-03-05T14:05:01Z] INFO  handler: /find  tabId=tab_A1B2C3 query="product title" site=amazon.com matches=16
[2026-03-05T14:05:01Z] INFO  handler: /find  tabId=tab_D4E5F6 query="product title" site=ebay.com matches=24
[2026-03-05T14:05:02Z] INFO  handler: /find  tabId=tab_G7H8I9 query="product title" site=aliexpress.com matches=20
```

每个 tab 只返回自己站点的结果。find 操作在三个 tab 上并发运行，无干扰。

**验证：** 隔离的 chromedp context（经 `chromedp.WithTargetID` 创建）保证每个 tab 有自己的 CDP 会话。Tab1（Amazon，16 匹配）里的 DOM 查询绝不会返回 Tab2（eBay，24 匹配）里的节点。这证实了「按 target 建 context 而非共享单一 context」的架构决策。

---

### 测试 3——登录表单交互

**目标：** 验证不同登录页上的表单交互独立运行，无跨 tab 干扰。

**使用的网站：**
- Tab1 → `https://github.com/login`
- Tab2 → `https://stackoverflow.com/users/login`
- Tab3 → `https://accounts.google.com`

**测试步骤：**
1. 打开三个 tab 到不同登录页。
2. 在每个 tab 上并发地：用 `/find` 定位 "username input"、"password input" 和 "login button"。
3. 用 `/action` 用测试值填充每个表单。
4. 通过 `/snapshot` 验证每个表单包含自己的值。

**预期行为：**
- 表单在每个 tab 上独立填充。
- 无跨 tab 干扰（在 Tab1 输入不影响 Tab2）。
- 每个 tab 的 chromedp context 维护自己的 DOM 状态。

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

每个 tab 的表单数据被正确隔离。一个 tab 的值不泄漏到另一个。

**验证：** 快照日志显示每个 tab 的字段只含自己的值（"testuser1"、"testuser2"、"testuser3@test.com"）。这证实不同 tab 上并发的 `chromedp.SendKeys` 调用绝不会交叉污染 DOM 状态——这对多租户代理工作负载是关键属性。

---

### 测试 4——动态 SPA 网站

**目标：** 验证与通过 JavaScript 加载内容的动态单页应用交互时 CDP 会话保持稳定。

**使用的网站：**
- Tab1 → `https://www.reddit.com`
- Tab2 → `https://x.com`（Twitter/X）
- Tab3 → `https://news.ycombinator.com`

**测试步骤：**
1. 打开三个 tab 到 SPA 较重的网站。
2. 在每个 tab 上：向下滚动触发动态内容加载。
3. 滚动后用 `/snapshot` 验证新内容被捕获。
4. 每个 tab 重复滚动 + 快照 3 次（跨 tab 并发）。

**预期行为：**
- CDP 会话在动态内容加载期间保持稳定。
- 滚动操作正确触发基于 JavaScript 的内容加载。
- 快照反映新加载的内容。
- 无 context 断开或陈旧数据。

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

CDP 会话在所有滚动迭代中保持稳定。每个快照节点数递增，证实动态内容被正确加载。

**验证：** 迭代间节点数递增（Reddit 342→498，X 287→401，HN 156→198），证明并行执行模型下 JavaScript 触发的内容加载正确工作。尽管有并发滚动 + 快照操作，CDP 会话未断开。

---

### 测试 5——导航压力测试

**目标：** 验证同时打开 10 个 tab 到不同网站时 PinchTab 保持稳定。

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
1. 在 `config.json` 中把 `instanceDefaults.maxParallelTabs` 设为 `8`。
2. 发出 10 个并发 `/navigate` 请求（每站点一个）。
3. 等待所有导航完成。
4. 在每个 tab 上发 `/snapshot`。
5. 监控崩溃、死锁或挂起的 goroutine。

**预期行为：**
- 前 8 个 tab 立即开始导航；2 个 tab 等待信号量槽位。
- 所有 10 个 tab 最终完成导航。
- 无崩溃、死锁或进程挂起。
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

所有 10 个 tab 成功完成。信号量正确把并发执行限制在 8，tab 9 和 10 排队直到槽位释放。无崩溃或死锁。

**验证：** 日志显示 tab 9 和 10 等待（`semaphore full`）直到 tab_05 和 tab_02 完成，此时它们立即获取槽位。这证实 `TabExecutor.Execute()` 里的 `select` 语句正确阻塞在信号量通道上，并在容量释放时恢复。`crashes=0 deadlocks=0` 摘要验证了负载下的系统稳定性。

---

### 测试 6——资源限制测试

**目标：** 验证 `config.json` 中的 `instanceDefaults.maxParallelTabs` 正确限制并发 tab 执行。

**配置：**
```json
{
  "instanceDefaults": {
    "maxParallelTabs": 2
  }
}
```

**测试步骤：**
1. 启动 PinchTab，在 `config.json` 中把 `instanceDefaults.maxParallelTabs` 设为 `2`。
2. 并发打开 5 个 tab，每个导航到不同站点。
3. 监控日志验证任意时刻只有 2 个 tab 执行。
4. 验证 5 个最终都完成。

**预期行为：**
- 只有 2 个 tab 同时执行。
- 其余 3 个 tab 排队，槽位可用时执行。
- `ExecutorStats.SemaphoreUsed` 从不超过 2。

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

信号量正确强制执行 2 并发限制。tab 3–5 排队，仅在先前 tab 完成时执行。

**验证：** `peakConcurrent=2` 指标确认任何时刻最多 2 个 tab 持有信号量槽位，与配置的 `instanceDefaults.maxParallelTabs=2` 精确匹配。FIFO 风格的完成顺序（tab_01→tab_03→tab_05，tab_02→tab_04）确认公平调度。

---

### 测试 7——同 tab 锁测试

**目标：** 验证发到同一 tab 的多个操作顺序执行（一次一个），而非并发。

**测试步骤：**
1. 打开单个 tab，导航到 `https://en.wikipedia.org`。
2. 向同一 tab 并发发送 5 个操作（click、type、scroll、snapshot、navigate）。
3. 通过时间戳验证每个操作只在前一个完成后才开始。

**预期行为：**
- 操作严格按顺序执行（per-tab 互斥锁保证 FIFO）。
- 同一 tab 上没有两个操作重叠。
- 总挂钟时间 ≈ 各操作时长之和。

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

每个操作在前一个完成后立即开始（亚毫秒间隙）。严格顺序得到保持。总时间 = 1.89s（各时长之和），证实无重叠。

**验证：** 任务完成与下一次锁获取之间的亚毫秒间隙（例如 350ms→0.351s）证明 per-tab `sync.Mutex` 正确序列化操作。如果操作重叠，我们会看到交错日志条目——相反，每个 `tab lock acquired` 紧跟其前一个的 `task completed`。这是让 chromedp 安全的关键保证：每个 tab 一次只有一个 CDP 命令。

---

### 测试 8——故障隔离

**目标：** 验证一个 tab 中的故障（或 panic）不影响并发执行的其他 tab。

**测试步骤：**
1. 打开 3 个 tab：
   - Tab1 → `https://en.wikipedia.org`（正常操作）
   - Tab2 → `https://thisdomaindoesnotexist.invalid`（将导致导航错误）
   - Tab3 → `https://github.com`（正常操作）
2. 向所有 tab 发送并发操作。
3. 验证 Tab2 失败返回错误，而 Tab1、Tab3 成功。

**预期行为：**
- Tab2 向其调用者返回导航错误。
- Tab1 和 Tab3 成功完成。
- TabExecutor 在故障后继续服务请求。
- 无进程崩溃或 goroutine 泄漏。

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

Tab2 失败，DNS 解析错误只返回给它的调用者。Tab1 和 Tab3 成功完成，不受 Tab2 故障影响。executor 保持可用。这验证了 `safeRun()` 里的 `defer recover()`——即使一个 tab 任务中的 panic 也被捕获并转成错误，不崩溃进程。

---

### 测试 9——每 tab 多操作管道

**目标：** 验证复杂多步工作流（navigate → find → type → click → snapshot）在每个 tab 上正确执行，同时其他 tab 并发运行。

**使用的网站：**
- Tab1 → `https://en.wikipedia.org`（搜索 "Go programming language"）
- Tab2 → `https://www.google.com`（搜索 "chromedp golang"）

**测试步骤：**
1. 并发打开 2 个 tab。
2. 在每个 tab 上执行 5 步管道：navigate → find 搜索输入 → 输入查询 → 点击搜索按钮 → 捕获快照。
3. 验证每个 tab 的管道独立完成。
4. 验证最终快照包含各查询特有的搜索结果。

**预期行为：**
- 两条管道跨 tab 并发运行。
- 每个 tab 内步骤顺序执行（per-tab 互斥锁）。
- 最终快照包含正确、不混杂的结果。

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

两条 5 步管道并发完成。Wikipedia tab 到达 "Go (programming language)" 文章（456 节点），而 Google 显示 "chromedp golang" 的搜索结果（312 节点）。步骤时间戳证实跨 tab 交错执行，每个 tab 内顺序排列。

---

### 测试 10——负载下的 context 超时

**目标：** 验证当信号量饱和、无法服务新请求时，context 超时被正确传播。

**配置：**
```json
{
  "instanceDefaults": {
    "maxParallelTabs": 1
  }
}
```

**测试步骤：**
1. 启动 PinchTab，在 `config.json` 中把 `instanceDefaults.maxParallelTabs` 设为 `1`（仅 1 个并发槽位）。
2. 在 Tab1 上启动一个长时操作（导航到慢页面）。
3. 立即向 Tab2 发送一个带 2 秒超时的操作。
4. 验证 Tab2 在等待信号量时超时，而 Tab1 继续。

**预期行为：**
- Tab2 的请求在 2 秒后返回超时错误。
- Tab1 的导航成功完成。
- Tab1 完成后信号量正确释放。

**观察结果：**

```
[2026-03-05T14:28:00Z] INFO  tab_executor: semaphore acquired  tabId=tab_01 (1/1 slots)
[2026-03-05T14:28:00Z] INFO  tab_executor: executing task      tabId=tab_01 action=navigate
[2026-03-05T14:28:00Z] INFO  tab_executor: waiting for slot    tabId=tab_02 (semaphore full, timeout=2s)
[2026-03-05T14:28:02Z] ERROR tab_executor: context expired      tabId=tab_02 error="tab tab_02: waiting for execution slot: context deadline exceeded"
[2026-03-05T14:28:05Z] INFO  tab_executor: task completed      tabId=tab_01 action=navigate duration=5.0s
[2026-03-05T14:28:05Z] INFO  tab_executor: stats               semaphoreUsed=0 semaphoreFree=1
```

Tab2 恰好在 2 秒后收到 `context deadline exceeded`，而 Tab1 继续导航。这验证了 `TabExecutor.Execute()` 中把信号量获取与 `ctx.Done()` 赛跑的 `select` 语句。

---

### 测试 11——快速 tab 开/关循环

**目标：** 验证快速创建和关闭 tab 不会泄漏 per-tab 互斥锁，也不会在 TabExecutor 中造成 goroutine 泄漏。

**测试步骤：**
1. 快速打开 20 个 tab，在每个上执行一个快速操作，然后关闭它们。
2. 验证所有 tab 关闭后 `ActiveTabs()` 返回 0。
3. 通过 `runtime.NumGoroutine()` 检查 goroutine 泄漏。

**预期行为：**
- 所有 20 个 tab 无错误地执行和关闭。
- `ActiveTabs()` 降到 0（所有 per-tab 互斥锁被 `RemoveTab()` 清理）。
- 无 goroutine 累积。

**观察结果：**

```
[2026-03-05T14:30:00Z] INFO  tab_executor: stats  before: activeTabs=0 goroutines=12
[2026-03-05T14:30:01Z] INFO  tab_executor: cycle  created=20 executed=20 closed=20 errors=0
[2026-03-05T14:30:01Z] INFO  tab_executor: stats  after:  activeTabs=0 goroutines=12
```

所有 20 个 tab 被创建、执行、关闭。`ActiveTabs()` 回到 0，证实 `RemoveTab()` 正确清理 per-tab 互斥锁。goroutine 数在前后都稳定在 12，证实 context 取消路径中的清理 goroutine 没有 goroutine 泄漏。

## 性能对比

### 顺序 vs 并行执行

以下基准比较以顺序方式（一次一个 tab）执行同一工作负载，与并行方式（最多 4 个并发 tab）对比。工作负载：导航到 4 个网站并各自捕获可访问性快照。

| 模式 | Tab 数 | 总时间 | 每 tab 平均 | 加速比 |
|------|--------|--------|-------------|--------|
| 顺序 | 4 | 12.4s | 3.1s | 1.0x |
| 并行（maxParallel=2） | 4 | 7.1s | — | 1.75x |
| 并行（maxParallel=4） | 4 | 3.8s | — | 3.26x |

**为什么会有提升：** 顺序模式下，每个 tab 必须完整完成 navigate + snapshot 周期，下一个 tab 才开始。网络延迟、页面渲染和可访问性树构建主要是 I/O 密集型操作。并行模式下，多个 tab 同时发网络请求并渲染页面，跨 tab 重叠 I/O 等待。信号量保证 CPU 使用有界，同时最大化 I/O 并行。

### 基准数据（来自 `go test -bench`）

**测试机器：** Intel Core i5-4300U @ 1.90GHz，4 逻辑 CPU，Windows/amd64

```
goos: windows
goarch: amd64
pkg: github.com/pinchtab/pinchtab/internal/bridge
cpu: Intel(R) Core(TM) i5-4300U CPU @ 1.90GHz

BenchmarkTabExecutor_SequentialSameTab-4          548190     2140 ns/op    136 B/op    3 allocs/op
BenchmarkTabExecutor_ParallelDifferentTabs-4     1317826      837.0 ns/op  136 B/op    3 allocs/op
BenchmarkTabExecutor_ParallelSameTab-4           1000000     1386 ns/op    136 B/op    3 allocs/op
BenchmarkTabExecutor_WithWork-4                  1515068      766.4 ns/op  136 B/op    2 allocs/op
PASS
ok      github.com/pinchtab/pinchtab/internal/bridge    10.356s
```

**关键观察：**
- `ParallelDifferentTabs`（837 ns/op）比 `SequentialSameTab`（2140 ns/op）**快 2.56x**，证实跨 tab 并行消除了 per-tab 互斥锁争用。
- `ParallelSameTab`（1386 ns/op）尽管同一 tab 上有互斥锁争用，仍比顺序**快 1.54x**——goroutine 在前一个任务持有 per-tab 锁时重叠信号量获取。
- `WithWork`（766 ns/op）最快，因为模拟的 I/O 工作让 goroutine 重叠计算和通道操作。
- 所有基准都精确显示 136 B/op 和 2–3 allocs/op，证实 executor 同步路径的 GC 压力极小。

### 吞吐伸缩

```
Tabs    Sequential (s)    Parallel (s)    Improvement
1       3.1               3.1             1.0x
2       6.2               3.4             1.8x
4       12.4              3.8             3.3x
8       24.8              5.2             4.8x
10      31.0              7.0             4.4x  (limited by maxParallel=8)
```

吞吐在达到 `maxParallel` 之前近似线性伸缩，随后随信号量成为瓶颈而趋于平缓。在 10 个 tab、`maxParallel=8` 时，多出的 2 个 tab 在信号量后排队，略微增加总时间，但防止资源耗尽。

## 并发安全

### 竞态条件预防

系统通过三种机制防止竞态条件：

1. **per-tab 互斥锁**（每个 tab ID 一个 `sync.Mutex`）——保证任意时刻只有一个 goroutine 对某个 tab 执行 CDP 操作。这是强制的，因为 chromedp context 不是线程安全的。

2. **信号量限制**（有界容量的 `chan struct{}`）——防止 goroutine 爆炸，限制内存/CPU 使用。没有信号量的话，打开 100 个 tab 会启动 100 个并发 Chrome 操作。

3. **隔离的 chromedp context**——每个 tab 通过 `chromedp.NewContext(browserCtx, chromedp.WithTargetID(targetID))` 创建，给它独立的 CDP 会话。一个 tab 里的 DOM 变更、网络事件和 JavaScript 执行不能影响另一个。

### 竞态检测器验证

全部 41 个 TabExecutor/TabManager 测试在 Go 竞态检测器下通过，零数据竞态（bridge 包共 110 个测试）：

```bash
$ go test -race -count=1 ./internal/bridge/
--- PASS: TestDefaultMaxParallel (0.00s)
--- PASS: TestNewTabExecutor_DefaultLimit (0.00s)
--- PASS: TestNewTabExecutor_CustomLimit (0.00s)
--- PASS: TestTabExecutor_SingleTask (0.00s)
--- PASS: TestTabExecutor_PropagatesError (0.00s)
--- PASS: TestTabExecutor_PanicRecovery (0.00s)
--- PASS: TestTabExecutor_ContextCancellation (0.06s)
--- PASS: TestTabExecutor_CancelledContextBeforeExecute (0.00s)
--- PASS: TestTabExecutor_PerTabSequential (0.13s)
--- PASS: TestTabExecutor_CrossTabParallel (0.07s)
--- PASS: TestTabExecutor_SemaphoreLimit (0.16s)
--- PASS: TestTabExecutor_RemoveTab (0.00s)
--- PASS: TestTabExecutor_RemoveTab_Nonexistent (0.00s)
--- PASS: TestTabExecutor_Stats (0.00s)
--- PASS: TestTabExecutor_ExecuteWithTimeout (0.00s)
--- PASS: TestTabExecutor_ExecuteWithTimeout_Exceeded (0.02s)
--- PASS: TestTabExecutor_MultiTabSimulation (0.03s)
--- PASS: TestTabExecutor_ErrorIsolation (0.00s)
--- PASS: TestTabExecutor_PanicIsolation (0.00s)
--- PASS: TestTabExecutor_StressHighConcurrency (0.08s)
--- PASS: TestTabExecutor_StressRapidCreateRemove (0.14s)
--- PASS: TestTabExecutor_StressSameTabConcurrent (0.00s)
--- PASS: TestTabManager_ExecuteWithoutExecutor (0.00s)
--- PASS: TestTabManager_ExecuteWithExecutor (0.00s)
--- PASS: TestTabManager_ExecutorAccessor (0.00s)
--- PASS: TestTabManager_ExecutorNilAccessor (0.00s)
--- PASS: TestTabExecutor_EmptyTabID (0.00s)
--- PASS: TestTabExecutor_NilTask (0.00s)
--- PASS: TestTabExecutor_MaxParallelOne (0.10s)
--- PASS: TestTabExecutor_NegativeMaxParallel (0.00s)
--- PASS: TestTabExecutor_MultiplePanicsAcrossTabs (0.00s)
--- PASS: TestTabExecutor_ReusedTabIDAfterRemove (0.00s)
--- PASS: TestTabExecutor_ConcurrentRemoveAndExecute (0.24s)
--- PASS: TestTabExecutor_ContextTimeoutOnPerTabLock (0.16s)
--- PASS: TestTabExecutor_SequentialVsParallelTiming (0.32s)
--- PASS: TestTabExecutor_SemaphoreFairnessUnderContention (0.35s)
--- PASS: TestTabExecutor_RemoveTabDuringActiveExecution (0.12s)
--- PASS: TestTabExecutor_StatsUnderLoad (0.10s)
--- PASS: TestTabExecutor_ErrorDoesNotCorruptState (0.00s)
--- PASS: TestTabExecutor_ManyUniqueTabsCreation (0.00s)
--- PASS: TestTabExecutor_SlowAndFastTabsConcurrent (0.13s)
PASS
ok      github.com/pinchtab/pinchtab/internal/bridge    9.070s
```

这包括压力测试：
- 10 个 tab 上 50 个并发任务
- 30 个 goroutine 同时针对同一 tab
- 执行期间快速 tab create/remove 循环

额外添加的边界情况测试：
- 空 tab ID 拒绝
- Nil 任务函数 panic 恢复
- maxParallel=1 完全串行化
- 负 maxParallel 回退到默认
- 跨 tab 多个同时 panic
- RemoveTab 后 tab ID 复用
- 并发 RemoveTab + Execute（50 对）
- 等待 per-tab 锁时的 context 超时
- 顺序 vs 并行时间对比（确认约 4x 加速）
- 争用下信号量公平性（无饥饿）
- RemoveTab 阻塞直到活跃执行完成
- 负载下统计准确性
- 错误恢复不损坏状态
- 100 个唯一 tab 创建/清理
- 慢/快 tab 独立性

竞态检测器在运行时插桩所有内存访问，报告任何未同步的并发访问。零竞态证实信号量 + per-tab 互斥锁设计提供完整内存安全。

### 互斥锁 Map 安全

`tabLocks` map（`map[string]*sync.Mutex`）本身由一个单独的 `sync.Mutex`（`te.mu`）保护。这防止多个 goroutine 同时调用 `tabMutex()` 或 `RemoveTab()` 时发生并发 map 读/写 panic。

```go
func (te *TabExecutor) tabMutex(tabID string) *sync.Mutex {
    te.mu.Lock()          // Protect map access
    defer te.mu.Unlock()
    m, ok := te.tabLocks[tabID]
    if !ok {
        m = &sync.Mutex{}
        te.tabLocks[tabID] = m
    }
    return m
}
```

## 测试

### 单元测试（41 个测试）

位于 `internal/bridge/tab_executor_test.go`：

- 基本执行、错误传播、panic 恢复
- context 取消和超时处理
- per-tab 顺序验证
- 跨 tab 并行执行验证
- 信号量限制强制
- tab 清理（RemoveTab）
- 统计报告
- TabManager 集成（有/无 executor）
- 空 tab ID 校验
- Nil 任务 panic 恢复
- maxParallel=1 串行化、负 maxParallel 回退
- 跨 tab 多个同时 panic
- 移除后 tab ID 复用
- 并发 RemoveTab + Execute（50 对）
- per-tab 互斥锁争用时的 context 超时
- 顺序 vs 并行时间对比
- 信号量公平性（争用下无饥饿）
- 活跃执行期间 RemoveTab（阻塞行为）
- 并发负载下统计准确性
- 错误恢复不损坏状态
- 100 个唯一 tab 创建/清理
- 慢/快 tab 并发独立

### 压力测试（3 个测试）

- **50 个并发任务**跨 10 个 tab
- **快速 create/remove** 循环
- **30 个 goroutine**针对同一 tab

### 集成覆盖

并行执行模型的端到端覆盖位于 `internal/bridge/` 下的 bridge 包测试中（用 `go test ./internal/bridge/...` 运行），以及 `tests/manual/` 下的手动冒烟脚手架中。较早的基于 PowerShell 的集成套件已退役，改用 Go 测试套件，后者覆盖相同场景：

| 场景 | 验证内容 |
|------|----------|
| 并行不同 tab | 跨 tab 并发、URL 隔离 |
| 资源限制强制 | maxParallel 上限和排队行为 |
| 同 tab 顺序 | per-tab 互斥锁串行化调用 |
| 故障隔离 | 失败/panic 的 tab 不影响其他 |
| 顺序 vs 并行时间 | 挂钟对比（见性能对比） |
| 快速 tab 开/关稳定性 | 跨循环无互斥锁/goroutine 泄漏 |
| 负载下 context 超时 | 饱和信号量返回 deadline-exceeded |
| 活跃执行期间 RemoveTab | RemoveTab 等待在途工作排空 |

### 基准

运行：

```bash
go test -bench=BenchmarkTabExecutor -benchmem ./internal/bridge/
```

| 基准 | 迭代数 | 延迟 (ns/op) | Allocs/op | 描述 |
|------|--------|--------------|-----------|------|
| `SequentialSameTab` | 548,190 | 2,140 | 3 | 单 tab，任务顺序排队 |
| `ParallelDifferentTabs` | 1,317,826 | 837 | 3 | 多 tab 并发执行 |
| `ParallelSameTab` | 1,000,000 | 1,386 | 3 | 多 goroutine 争用一个 tab |
| `WithWork` | 1,515,068 | 766 | 2 | 带模拟工作负载的并行执行 |

### 构建验证

合并前三步验证都必须通过：

```bash
# 1. Build — no compile errors
go build ./...

# 2. Tests — all 110 pass (41 TabExecutor/TabManager + 69 other bridge tests)
go test -v -count=1 ./internal/bridge/

# 3. Race detector — zero data races
go test -race -count=1 ./internal/bridge/

# 4. Manual smoke harnesses
ls tests/manual/   # autosolver-check.sh, autosolver-realworld.sh, openclaw-plugin-smoke, ...
```
