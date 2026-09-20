# Solve（挑战求解）

在当前页面检测并求解浏览器挑战（Cloudflare Turnstile、CAPTCHA、插页等）。

这些端点由 `internal/autosolver` 流水线驱动。自动模式下，PinchTab 先运行语义意图检测，然后按顺序尝试已配置的求解器，并在启用时可选回退到 LLM。

## 端点

```text
GET  /solvers
POST /solve
POST /solve/{name}
POST /tabs/{id}/solve
POST /tabs/{id}/solve/{name}
```

## 列出求解器

```bash
curl http://localhost:9867/solvers
```

```json
{
  "solvers": ["cloudflare", "semantic", "jschallenge"]
}
```

配置了对应 API key 时才包含 `capsolver` 和 `twocaptcha`。

## 自动检测求解

未提供 `solver` 字段时，PinchTab 按配置顺序（`autoSolver.solvers`）运行 autosolver 链。

```bash
curl -X POST http://localhost:9867/solve \
  -H "Content-Type: application/json" \
  -d '{"maxAttempts": 3, "timeout": 30000}'
```

页面未检测到挑战时，响应立即返回 `solved: true` 和 `attempts: 0`。

## 命名求解器

在主体或路径中按名称指定求解器：

```bash
# Body
curl -X POST http://localhost:9867/solve \
  -H "Content-Type: application/json" \
  -d '{"solver": "cloudflare", "maxAttempts": 3}'

# Path
curl -X POST http://localhost:9867/solve/cloudflare \
  -H "Content-Type: application/json" \
  -d '{"maxAttempts": 3}'
```

## 标签页范围求解

```bash
curl -X POST http://localhost:9867/tabs/{tabId}/solve \
  -H "Content-Type: application/json" \
  -d '{"solver": "cloudflare"}'
```

## 请求主体

| 字段 | 类型 | 默认值 | 说明 |
|--------------|--------|---------|------------------------------------------|
| `tabId` | string | — | 标签页 ID（可选，用默认标签页） |
| `solver` | string | — | 求解器名称（可选，自动检测） |
| `maxAttempts` | int | 配置（`autoSolver.maxAttempts`，默认 8） | 最大求解尝试次数 |
| `timeout` | float | 按求解器链自动估计（绝不低于 30000） | 总超时（毫秒）；显式值按原值使用 |

路径形式（`/solve/{name}`）覆盖主体 `solver`。在 `/tabs/{id}/solve` 上，主体 `tabId` 与路径不一致则拒绝。

## 响应

```json
{
  "tabId": "DEADBEEF",
  "solver": "cloudflare",
  "solved": true,
  "challengeType": "turnstile",
  "attempts": 1,
  "title": "thuisbezorgd.nl"
}
```

| 字段 | 类型 | 说明 |
|-----------------|--------|------------------------------------------------|
| `tabId` | string | 求解运行在哪个标签页 |
| `solver` | string | 处理该挑战的求解器 |
| `solved` | bool | 挑战是否已解决 |
| `challengeType` | string | 挑战变体（`turnstile`、`recaptcha-v2`、`hcaptcha`）或宽泛意图（`captcha`、`blocked`） |
| `attempts` | int | 尝试次数 |
| `title` | string | 最终页面标题 |
| `handoff` | string | 检测到挑战但未解决时为 `paused_handoff`——标签页已暂停等待 [handoff](./handoff.md) |
| `hint` | string | 与 `handoff` 同时出现：把控制权交还给用户，再恢复该标签页 |

## 错误响应

| 代码 | 含义 |
|------|----------------------------------------|
| 400 | 无效主体；未注册名称为 `unknown_solver`；`capsolver`/`twocaptcha` 缺 API key 时为 `solver_key_missing` |
| 404 | 标签页未找到 |
| 409 | `dialog_blocked`（有 JavaScript 对话框打开）或 `tab_paused_handoff` |
| 423 | `tab_locked`：标签页被其他所有者租用 |
| 500 | CDP/Chrome 错误 |

## 内置求解器

### Semantic（`semantic`）

语义优先求解器，使用 `/find` 式匹配和多步动作规划来解决挑战与流程。

### JS Challenge（`jschallenge`）

通用 JavaScript 反爬/插页求解器，等待、探测常见验证控件，并轮询挑战是否解决。

### Cloudflare（`cloudflare`）

处理 Cloudflare Turnstile 和插页挑战。

**检测**：检查页面标题中的已知 Cloudflare 标志（"Just a moment..."、"Attention Required"、"Checking your browser"）。

**挑战类型**：

| 类型 | 处理 |
|-------------------|--------------------------------------------------------|
| `non-interactive` | 等待自动解决（最多 15s） |
| `managed` | 定位 Turnstile iframe，点击复选框 |
| `interactive` | 同 managed |
| `embedded` | 通过 Turnstile script 标签检测，点击复选框 |

**点击策略**：求解器用类人鼠标输入（贝塞尔曲线移动、随机延迟、按下/释放偏移）点击 Turnstile 复选框。点击坐标相对 widget 尺寸计算（而非硬编码像素偏移），并带随机抖动。

**隐身要求**：Cloudflare 求解器在 PinchTab 配置中 `instanceDefaults.stealthLevel: "full"` 时效果最佳。Cloudflare 在复选框交互前后评估浏览器指纹（CDP 检测、WebGL、canvas、navigator 属性）。没有完整隐身，求解器可能点对了，但挑战仍可能在指纹验证中失败。用 `GET /stealth/status` 检查隐身状态。

### 外部求解器

- `capsolver`（需 `autoSolver.external.capsolverKey`）
- `twocaptcha`（需 `autoSolver.external.twoCaptchaKey`）

## 编写自定义求解器

实现 `autosolver.Solver` 接口，并在 autosolver 注册表构造处注册：

```go
package mygateway

import (
    "context"
  "github.com/pinchtab/pinchtab/internal/autosolver"
)

type MyGatewaySolver struct{}

func (s *MyGatewaySolver) Name() string { return "mygateway" }

func (s *MyGatewaySolver) Priority() int { return 150 }

func (s *MyGatewaySolver) CanHandle(ctx context.Context, page autosolver.Page) (bool, error) {
    // Check page markers (title, DOM elements, etc.)
    return false, nil
}

func (s *MyGatewaySolver) Solve(ctx context.Context, page autosolver.Page, exec autosolver.ActionExecutor) (*autosolver.Result, error) {
    // Detect, interact, and resolve the challenge.
  return &autosolver.Result{SolverUsed: "mygateway", Solved: true}, nil
}
```

然后把它加入 `internal/autosolver/catalog/catalog.go` 中的 `buildAll`，处理器注册表设置会通过 `catalog.Registrable` 读取它。
