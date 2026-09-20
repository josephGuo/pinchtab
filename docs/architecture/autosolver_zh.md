# AutoSolver

## 概述

AutoSolver 系统为 PinchTab 提供模块化、语义优先的浏览器自动化。它脱胎于早期一个单用途的挑战求解框架，演变为一个通用的自动化代理，能够处理 CAPTCHA、登录流程、注册流程、多步导航和引导流程。

### 设计原则

1. **隔离优先** — autosolver 模块（`internal/autosolver/`）与 chromedp 或桥接运行时零耦合。所有浏览器交互都通过 `Page` 和 `ActionExecutor` 接口进行。

2. **语义优先** — `pinchtab/semantic` 包是主要的智能层。LLM 仅作为最后的后备方案使用。

3. **可插拔架构** — 求解器通过 `Registry` 在运行时注册。外部求解器（Capsolver、2Captcha）是通过配置启用的可选插件。

4. **行为 > 欺骗** — 求解器通过合法的浏览器操作（点击、输入）与页面交互，而不是 API 取巧手段或假令牌。

5. **超越 CAPTCHA 的可扩展性** — `IntentType` 系统除 CAPTCHA 求解外，还支持登录、注册、引导和导航流程。

---

## 架构图

```
┌─────────────────────────────────────────────────────┐
│                    AutoSolver                        │
│                                                      │
│  ┌──────────┐    ┌──────────┐    ┌──────────────┐   │
│  │ Registry │───▶│Core Loop │───▶│ Fallback     │   │
│  │ (solvers)│    │(detect + │    │ Chain:       │   │
│  └──────────┘    │ dispatch)│    │ semantic →   │   │
│                  └────┬─────┘    │ rule-based → │   │
│                       │          │ external →   │   │
│                       ▼          │ LLM          │   │
│              ┌────────────────┐  └──────────────┘   │
│              │  Interfaces    │                      │
│              │  Page          │                      │
│              │  ActionExecutor│                      │
│              │  SemanticEngine│                      │
│              │  LLMProvider   │                      │
│              └────────┬───────┘                      │
└───────────────────────┼──────────────────────────────┘
                        │ (interface boundary)
        ┌───────────────┼───────────────┐
        ▼               ▼               ▼
┌──────────────┐ ┌─────────────┐ ┌──────────────┐
│adapters/     │ │semantic/    │ │external/     │
│pinchtab.go   │ │adapter.go   │ │capsolver.go  │
│(chromedp)    │ │(semantic pkg)│ │twocaptcha.go │
└──────────────┘ └─────────────┘ └──────────────┘
```

## 模块结构

```
internal/autosolver/
├── interfaces.go          # Page, ActionExecutor, Solver, SemanticEngine, LLMProvider
├── types.go               # Result, Intent, Config, enums
├── autosolver.go          # Core orchestrator with fallback chain
├── challenge_detection.go # Shared challenge classification (title/URL/HTML)
├── heuristics.go          # Title-based intent detection fallback
├── keygated.go            # Solvers that register only once their API key is set
├── registry.go            # Instance-level solver registry with priority ordering
├── adapters/
│   └── pinchtab.go        # Bridge adapter (ONLY chromedp import)
├── catalog/
│   └── catalog.go         # Single owner of the names autoSolver.solvers accepts
├── semantic/
│   └── adapter.go         # Wraps pinchtab/semantic ElementMatcher
├── external/
│   ├── external.go        # Shared external-solver wrapper
│   ├── capsolver.go       # Capsolver API skeleton
│   └── twocaptcha.go      # 2Captcha API skeleton
├── llm/
│   └── llm.go             # LLM provider skeleton with structured prompts
└── solvers/
    ├── cloudflare.go      # Cloudflare Turnstile solver
    └── jschallenge.go     # Generic JavaScript challenge/interstitial solver
```

测试文件与每个文件并列放置（`*_test.go`）。供 LLM 提示使用的 HTML 裁剪位于共享的 `internal/htmltrim` 包中。

## 核心接口

### Page

当前浏览器页面的只读视图：

```go
type Page interface {
    URL() string
    Title() string
    HTML() (string, error)
    HTMLWithin(timeout time.Duration) (string, error)
    Screenshot() ([]byte, error)
}
```

### ActionExecutor

执行类似人类行为的浏览器操作：

```go
type ActionExecutor interface {
    Click(ctx context.Context, x, y float64) error
    Type(ctx context.Context, text string) error
    WaitFor(ctx context.Context, selector string, timeout time.Duration) error
    Evaluate(ctx context.Context, expr string, result interface{}) error
    Navigate(ctx context.Context, url string) error
}
```

### Solver

处理特定类别的挑战：

```go
type Solver interface {
    Name() string
    Priority() int  // Lower = tried first
    CanHandle(ctx context.Context, page Page) (bool, error)
    Solve(ctx context.Context, page Page, executor ActionExecutor) (*Result, error)
}
```

**优先级范围：**
| 范围 | 类别 |
|-------|----------|
| 0–99 | 内置求解器（Cloudflare 等） |
| 100–199 | 基于语义的求解器 |
| 200–299 | 外部 API 求解器（Capsolver, 2Captcha） |
| 900+ | LLM 后备 |

## 后备链

意图在第一次尝试之前只检测一次；核心循环随后在每次尝试中执行本链的其余部分：

```
1. Detect intent (semantic engine, or title heuristics when none is configured)
2. If intent = normal → return solved
3. Try semantic-first action planning (`/find` + self-healing)
4. If still unresolved, find matching solvers (CanHandle = true)
5. Execute solvers in configured order (`autoSolver.solvers`), falling back to
    priority order when configuration does not match available solvers
6. If all fail AND LLM enabled:
   a. Trim HTML to ~4KB (`htmltrim.TrimHTML`, 4000-byte cap)
   b. Build structured prompt with attempt history
   c. Execute LLM-suggested action
7. Retry with exponential backoff (500ms → 10s cap)
8. Stop after MaxAttempts (default: 8)
```

## 配置

### 配置文件（`config.json`）

```json
{
  "autoSolver": {
    "enabled": true,
    "autoTrigger": true,
    "triggerOnNavigate": true,
    "triggerOnAction": true,
    "maxAttempts": 8,
    "solverTimeoutSec": 30,
    "retryBaseDelayMs": 500,
    "retryMaxDelayMs": 10000,
    "solvers": ["cloudflare", "semantic"],
    "llmProvider": "openai",
    "llmFallback": false,
    "external": {
      "capsolverKey": "CAP-xxx",
      "twoCaptchaKey": "xxx"
    }
  }
}
```

外部 provider 的 API 密钥仅在配置文件的 `autoSolver.external` 中配置。

## 扩展指南

### 添加新求解器

1. 在 `internal/autosolver/solvers/` 中创建新文件：

```go
package solvers

type MySolver struct{}

func (s *MySolver) Name() string  { return "myservice" }
func (s *MySolver) Priority() int { return 150 }

func (s *MySolver) CanHandle(ctx context.Context, page autosolver.Page) (bool, error) {
    // Check if this solver can handle the current page
    return strings.Contains(page.Title(), "my-challenge"), nil
}

func (s *MySolver) Solve(ctx context.Context, page autosolver.Page, executor autosolver.ActionExecutor) (*autosolver.Result, error) {
    // Implement solving logic using Page + ActionExecutor
    result := &autosolver.Result{SolverUsed: "myservice"}
    // ...
    return result, nil
}
```

2. 向 AutoSolver 注册：

```go
as := autosolver.New(cfg, semanticEngine, nil)
as.Registry().Register(&solvers.MySolver{})
```

### 与 PinchTab Bridge 一起使用

```go
// Create Page + Executor from a bridge tab
page, executor, err := adapters.NewFromBridge(bridge, tabID)

// Run the autosolver
as := autosolver.New(autosolver.DefaultConfig(), semanticAdapter, nil)
as.Registry().MustRegister(&solvers.Cloudflare{})

result, err := as.Solve(ctx, page, executor)
if result.Solved {
    log.Printf("Solved by %s in %d attempts", result.SolverUsed, result.Attempts)
}
```

## 与 browser-use 的比较

| 方面 | browser-use | PinchTab AutoSolver |
|--------|-------------|-------------------|
| 决策引擎 | 每步 LLM | 语义优先，LLM 后备 |
| DOM 处理 | 每步完整 DOM/截图 | 裁剪的 HTML，a11y 树 |
| 成本 | 高（每步 LLM） | 低（仅失败时 LLM） |
| 速度 | 慢（LLM 延迟） | 快（本地语义匹配） |
| 确定性 | 低（LLM 非确定性） | 高（基于规则 + 语义） |
| 模块化 | 单片式 | 接口驱动，可插拔 |

## 向后兼容性

`internal/autosolver` 是唯一的求解器框架。前身包、`bridge/cloudflare.go` 中重复的 Cloudflare 求解器、以及本要在两者间搭桥的 `LegacyAdapter` 垫片，全部已删除：从来没有任何代码把该适配器接起来，因此第二个注册表没有读取者，而每一次求解器级的重构仍要去编辑它那份每个求解器的副本。
