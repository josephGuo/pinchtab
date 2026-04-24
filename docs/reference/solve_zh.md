# 解决

检测并解决当前页面上的浏览器挑战（Cloudflare Turnstile、CAPTCHAs、 interstitial 等）。

PinchTab 附带一个可插拔的**解决器框架**。每个解决器针对特定的提供商（例如 Cloudflare）。解决器在启动时注册，可以通过名称显式调用或自动发现。

## 端点

```text
GET  /solvers
POST /solve
POST /solve/{name}
POST /tabs/{id}/solve
POST /tabs/{id}/solve/{name}
```

## 列出解决器

```bash
curl http://localhost:9867/solvers
```

```json
{
  "solvers": ["cloudflare"]
}
```

## 自动检测解决

当未提供 `solver` 字段时，PinchTab 按顺序尝试每个注册的解决器。使用第一个 `CanHandle` 返回 true 的解决器。

```bash
curl -X POST http://localhost:9867/solve \
  -H "Content-Type: application/json" \
  -d '{"maxAttempts": 3, "timeout": 30000}'
```

如果在页面上未检测到挑战，响应会立即返回 `solved: true` 和 `attempts: 0`。

## 命名解决器

在请求体或路径中通过名称指定解决器：

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

## 标签页范围解决

```bash
curl -X POST http://localhost:9867/tabs/{tabId}/solve \
  -H "Content-Type: application/json" \
  -d '{"solver": "cloudflare"}'
```

## 请求体

| 字段        | 类型   | 默认值 | 描述                              |
|--------------|--------|---------|------------------------------------------|
| `tabId`      | string | —       | 标签页 ID（可选，使用默认标签页）      |
| `solver`     | string | —       | 解决器名称（可选，自动检测）      |
| `maxAttempts`| int    | 3       | 最大解决尝试次数                   |
| `timeout`    | float  | 30000   | 总超时时间（毫秒）          |

## 响应

```json
{
  "tabId": "DEADBEEF",
  "solver": "cloudflare",
  "solved": false,
  "challengeType": "managed",
  "attempts": 1,
  "title": "Just a moment...",
  "needsHumanHandoff": true,
  "handoffReason": "challenge_requires_manual_intervention"
}
```

| 字段           | 类型   | 描述                                    |
|-----------------|--------|------------------------------------------------|
| `tabId`         | string | 解决运行的标签页                           |
| `solver`        | string | 处理挑战的解决器             |
| `solved`        | bool   | 挑战是否已解决             |
| `challengeType` | string | 挑战变体（例如 `managed`、`embedded`） |
| `attempts`      | int    | 尝试次数                        |
| `title`         | string | 最终页面标题                               |
| `needsHumanHandoff` | bool | 是否需要手动干预才能继续 |
| `handoffReason` | string | 切换原因（`challenge_requires_manual_intervention`、`credentials_required`，或不需要时为空） |

## 错误响应

| 代码 | 含义                                |
|------|----------------------------------------|
| 400  | 无效的请求体或未知的解决器名称    |
| 404  | 标签页未找到                          |
| 423  | 标签页被另一个所有者锁定            |
| 500  | CDP/Chrome 错误                       |

## 内置解决器

### Cloudflare (`cloudflare`)

处理 Cloudflare Turnstile 和 interstitial 挑战。

**检测**：检查页面标题是否有已知的 Cloudflare 指标（"Just a moment..."、"Attention Required"、"Checking your browser"）。

**挑战类型**：

| 类型              | 处理                                               |
|-------------------|--------------------------------------------------------|
| `non-interactive` | 等待自动解决（最多 15 秒）                  |
| `managed`         | 定位 Turnstile iframe，点击复选框              |
| `interactive`     | 与 managed 相同                                        |
| `embedded`        | 通过 Turnstile 脚本标签检测，点击复选框      |

**点击策略**：解决器使用类人鼠标输入（贝塞尔曲线移动、随机延迟、按下/释放偏移）点击 Turnstile 复选框。点击坐标相对于小部件尺寸计算（不是硬编码的像素偏移），并带有随机抖动。

**隐身要求**：Cloudflare 解决器在 PinchTab 配置中使用 `stealthLevel: "full"` 效果最佳。Cloudflare 在复选框交互前后评估浏览器指纹（CDP 检测、WebGL、画布、导航器属性）。没有完全隐身，解决器可能正确点击但挑战仍可能在指纹验证中失败。使用 `GET /stealth/status` 检查隐身状态。

## 编写自定义解决器

实现 `solver.Solver` 接口并在 `init()` 期间注册它：

```go
package mygateway

import (
    "context"
    "github.com/pinchtab/pinchtab/internal/solver"
)

func init() {
    solver.MustRegister("mygateway", &MyGatewaySolver{})
}

type MyGatewaySolver struct{}

func (s *MyGatewaySolver) Name() string { return "mygateway" }

func (s *MyGatewaySolver) CanHandle(ctx context.Context) (bool, error) {
    // 检查页面标记（标题、DOM 元素等）
    return false, nil
}

func (s *MyGatewaySolver) Solve(ctx context.Context, opts solver.Options) (*solver.Result, error) {
    // 检测、交互和解决挑战。
    return &solver.Result{Solver: "mygateway", Solved: true}, nil
}
```

解决器可以访问完整的 chromedp 上下文以进行 CDP 操作。