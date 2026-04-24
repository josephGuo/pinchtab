# AutoSolver 架构

## 概述

AutoSolver 系统为 Pinchtab 提供模块化、语义优先的浏览器自动化。它将现有的 `internal/solver` 框架（PR #395）演变为一个通用的自动化代理，能够处理 CAPTCHA、登录流程、注册流程、多步导航和入职序列。

### 设计原则

1. **隔离优先** — autosolver 模块 (`internal/autosolver/`) 与 chromedp 或桥接运行时零耦合。所有浏览器交互都通过 `Page` 和 `ActionExecutor` 接口进行。

2. **语义优先** — `pinchtab/semantic` 包是主要的智能层。LLM 仅用作最后的后备方案。

3. **可插拔架构** — 求解器通过 `Registry` 在运行时注册。外部求解器（Capsolver、2Captcha）是通过配置启用的可选插件。

4. **行为 > 欺骗** — 求解器通过合法的浏览器操作（点击、输入）与页面交互，而不是 API 黑客或假令牌。

5. **超越 CAPTCHA 的可扩展性** — `IntentType` 系统支持登录、注册、入职和导航流程以及 CAPTCHA 求解。

---

## 架构图

```
┌─────────────────────────────────────────────────────┐
│                    AutoSolver                        │
│                                                      │
│  ┌──────────┐    ┌──────────┐    ┌──────────────┐   │
│  │ Registry │───▶│Core Loop │───▶│ Fallback     │   │
│  │ (solvers)│    │(detect + │    │ Chain:       │   │
│  └──────────┘    │ dispatch)│    │ built-in →   │   │
│                  └────┬─────┘    │ semantic →   │   │
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
├── heuristics.go          # Title-based intent detection fallback
├── registry.go            # Instance-level solver registry with priority ordering
├── autosolver_test.go     # Core loop tests (7 test cases)
├── registry_test.go       # Registry tests (8 test cases)
├── adapters/
│   └── pinchtab.go        # Bridge adapter (ONLY chromedp import)
├── semantic/
│   └── adapter.go         # Wraps pinchtab/semantic ElementMatcher
├── external/
│   ├── capsolver.go       # Capsolver API skeleton
│   └── twocaptcha.go      # 2Captcha API skeleton
├── llm/
│   ├── llm.go             # LLM provider skeleton with structured prompts
│   └── trim.go            # HTML trimming for token efficiency
└── solvers/
    ├── cloudflare.go      # Cloudflare Turnstile (new interface, no chromedp)
    └── legacy.go          # Compatibility shim for existing solver.Solver
```

## 核心接口

### Page

当前浏览器页面的只读视图：

```go
type Page interface {
    URL() string
    Title() string
    HTML() (string, error)
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

核心循环每次尝试执行以下链：

```
1. 检测意图（语义引擎 → 标题启发式）
2. 如果意图 = 正常 → 返回已解决
3. 找到匹配的求解器（CanHandle = true，按优先级排序）
4. 尝试每个求解器：
   a. 内置（cloudflare，优先级 10）
   b. 外部（capsolver 优先级 200，twocaptcha 优先级 210）
5. 如果全部失败且启用 LLM：
   a. 将 HTML 裁剪到 ~4KB
   b. 使用尝试历史构建结构化提示
   c. 执行 LLM 建议的操作
6. 以指数退避重试（500ms → 10s 上限）
7. 在 MaxAttempts 后停止（默认：8）
```

## 配置

### 配置文件 (`config.json`)

```json
{
  "autoSolver": {
    "enabled": true,
    "maxAttempts": 8,
    "solvers": ["cloudflare", "semantic", "capsolver", "twocaptcha"],
    "llmProvider": "openai",
    "llmFallback": false,
    "external": {
      "capsolverKey": "CAP-xxx",
      "twoCaptchaKey": "xxx"
    }
  }
}
```

外部提供商 API 密钥仅在配置文件的 `autoSolver.external` 中配置。

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

### 与 Pinchtab Bridge 一起使用

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

| 方面 | browser-use | Pinchtab AutoSolver |
|--------|-------------|-------------------|
| 决策引擎 | 每步 LLM | 语义优先，LLM 后备 |
| DOM 处理 | 每步完整 DOM/截图 | 裁剪的 HTML，a11y 树 |
| 成本 | 高（每步 LLM） | 低（仅失败时 LLM） |
| 速度 | 慢（LLM 延迟） | 快（本地语义匹配） |
| 确定性 | 低（LLM 非确定性） | 高（基于规则 + 语义） |
| 模块化 | 单片式 | 接口驱动，可插拔 |

## 向后兼容性

现有的 `internal/solver` 包（PR #395）**未修改**。`bridge/cloudflare.go` 中的 `CloudflareSolver` 继续按原样工作。`LegacyAdapter` 垫片 (`solvers/legacy.go`) 包装旧的 `solver.Solver` 实现，使其与新的 `autosolver.Solver` 接口一起工作。