# Lite 引擎

PinchTab 包含一个**Lite 引擎**，可以执行 DOM 捕获 — 导航、快照、文本提取、点击和输入 — 而不需要 Chrome 或 Chromium。它由 [Gost-DOM](https://github.com/gost-dom/browser)（v0.11.0，MIT 许可证）提供支持，这是一个用纯 Go 编写的无头浏览器。

**问题：** [#201](https://github.com/pinchtab/pinchtab/issues/201)

---

## 为什么需要 Lite 引擎？

Chrome 是 PinchTab 的默认执行后端。真实的浏览器会话处理 JavaScript 渲染、机器人检测绕过、屏幕截图和 PDF 生成。对于许多工作负载 — 静态站点、维基、新闻文章、API — 这些都不需要。

| 驱动程序 | Chrome | Lite |
|---------|--------|------|
| 每个实例内存 | ~200 MB | ~10 MB |
| 冷启动延迟 | 1–6 秒 | <100 ms |
| JavaScript 渲染 | 是 | 否 |
| 屏幕截图 / PDF | 是 | 否 |
| 无需安装 Chrome | 否 | **是** |

Lite 在仅 DOM 工作负载上表现出色（导航快 3–4 倍，快照快 3 倍），是容器、CI 管道和 Chrome 不可用的边缘环境的正确选择。

---

## 架构

### 引擎接口

所有引擎都实现了在 `internal/engine/engine.go` 中定义的通用接口：

```go
type Engine interface {
    Name() string
    Navigate(ctx context.Context, url string) (*NavigateResult, error)
    Snapshot(ctx context.Context, filter string) ([]SnapshotNode, error)
    Text(ctx context.Context) (string, error)
    Click(ctx context.Context, ref string) error
    Type(ctx context.Context, ref, text string) error
    Capabilities() []Capability
    Close() error
}
```

Chrome 引擎包装了现有的 CDP/chromedp 管道。`internal/engine/lite.go` 中的 `LiteEngine` 使用 Gost-DOM 实现了相同的接口。

### 路由器（策略模式）

```
请求 → 路由器 → [规则 1] → [规则 2] → … → [回退规则] → 引擎
```

`internal/engine/router.go` 中的 `Router` 评估 `RouteRule` 实现的有序链。第一个返回非 `Undecided` 裁决的规则获胜。规则在启动时注册，可通过 `AddRule()` / `RemoveRule()` 热交换。

添加新的路由逻辑时不需要修改处理程序、桥接或配置 — 只需要一个 `RouteRule` 实现和一个 `router.AddRule(myRule)` 调用。

### 内置规则

| 规则 | 文件 | 行为 |
|------|------|------|
| `CapabilityRule` | `rules.go` | 将 `screenshot`、`pdf`、`evaluate`、`cookies` 路由到 Chrome |
| `ContentHintRule` | `rules.go` | 将以 `.html/.htm/.xml/.txt/.md` 结尾的 URL 路由到 Lite |
| `DefaultLiteRule` | `rules.go` | 捕获所有：所有剩余的 DOM 操作 → Lite（在 `lite` 模式下使用） |
| `DefaultChromeRule` | `rules.go` | 最终回退 → Chrome（在 `chrome` 和 `auto` 模式下使用） |

### 三种模式

| 模式 | 行为 |
|------|------|
| `chrome` | 所有请求都通过 Chrome。向后兼容的默认值。 |
| `lite` | DOM 操作（导航、快照、文本、点击、输入）使用 Gost-DOM。屏幕截图 / PDF / 评估 / Cookie 回退到 Chrome（如果 Chrome 不可用则返回 501）。 |
| `auto` | 通过规则进行每个请求的路由：首先评估能力和内容提示规则；未知 URL 回退到 Chrome。 |

---

## 请求流程（Lite 模式）

```
POST /navigate   (server.engine=lite)
    │
    ▼
handlers/navigation.go — HandleNavigate()
    │
    ├─ useLite() == true
    │       │
    │       ▼
    │   LiteEngine.Navigate(ctx, url)
    │       ├─ HTTP GET url
    │       ├─ 剥离 <script> 标签（x/net/html 分词器）
    │       ├─ browser.NewWindowReader(reader)  [Gost-DOM]
    │       └─ 返回 NavigateResult{TabID, URL, Title}
    │
    └─ w.Header().Set("X-Engine", "lite")
       JSON {"tabId": "lp-1", "url": "…", "title": "…"}
```

快照然后遍历 Gost-DOM 文档树，并将 HTML 语义映射到可访问性角色（标题、链接、按钮、文本框等）。文本遍历相同的树并折叠空白运行。

---

## 能力边界

| 操作 | Lite | Chrome |
|------|------|--------|
| 导航 | ✅（HTTP 获取 + DOM 解析） | ✅ |
| 快照 | ✅ | ✅ |
| 文本提取 | ✅ | ✅ |
| 点击 | ✅（DOM 事件分发） | ✅ |
| 输入 | ✅（DOM 输入事件） | ✅ |
| 屏幕截图 | ❌ → `501 Not Implemented` | ✅ |
| PDF | ❌ → `501 Not Implemented` | ✅ |
| 评估（JS） | ❌ → `501 Not Implemented` | ✅ |
| Cookies | ❌ → `501 Not Implemented` | ✅ |
| JavaScript 渲染的 SPA | ❌ | ✅ |
| 机器人检测绕过 | ❌ | ✅ |

`CapabilityRule` 确保即使在 `lite` 模式下，屏幕截图/ PDF/ 评估/ Cookie 也始终路由到 Chrome。

---

## 已知限制

| 限制 | 详细信息 |
|------|----------|
| `<script>` 标签 | Gost-DOM 在未初始化的 `ScriptHost` 上会 panic。脚本在解析前通过 `x/net/html` 分词器剥离。 |
| `<a href>` 点击 | Gost-DOM 在锚点点击时导航，可能会遇到脚本。`Click()` 将执行包装在 `defer recover()` 中，并返回错误而不是 panic。 |
| CSS `display:none` | Lite 没有 CSS 引擎，因此隐藏元素仍然出现在快照中。 |
| JavaScript 渲染的内容 | 只捕获初始 HTML。SPA（React、Next.js 等）应使用 Chrome。 |
| 阻止 HTTP 机器人的站点 | Stack Overflow 和类似站点向普通 HTTP 客户端返回 4xx/5xx。Chrome 通过真实浏览器会话绕过这一点。 |

---

## 配置

在配置文件中设置引擎：

```json
{
  "server": {
    "engine": "lite"
  }
}
```

`engine` 字段也会转发给子桥接实例，因此多实例部署中的每个托管实例都使用相同的模式。

### 响应头

由 Lite 引擎提供的响应包括：

```
X-Engine: lite
```

当采用 lite 路径时，此头出现在 `navigate`、`snapshot` 和 `text` 响应中，对可观测性和调试很有用。

---

## 性能

8 个真实世界网站的基准测试（导航 → 快照 → 文本管道，两个引擎都成功完成的 7 个站点）：

| 指标 | Lite | Chrome | 速度提升 |
|------|------|--------|----------|
| 导航总计 | 4,580 ms | 17,981 ms | **3.9×** 更快 |
| 快照总计 | 1,739 ms | 5,155 ms | **3.0×** 更快 |
| 文本总计 | 925 ms | 500 ms | 0.5×（Chrome 更快） |
| **总计** | **7,244 ms** | **23,636 ms** | **3.3× 更快** |

Chrome 在文本提取方面更快，因为它在浏览器中运行 Mozilla Readability.js。Lite 执行原始 DOM 文本遍历，对于非常大的页面较慢（例如 Wikipedia CS：687 ms vs 130 ms）。

### 何时使用每个引擎

| 工作负载 | 推荐 |
|----------|--------|
| 静态站点、维基、新闻、博客 | **Lite** — 3–12× 更快，无 Chrome 开销 |
| JavaScript 渲染的 SPA | **Chrome** — Lite 仅捕获 JS 前的 HTML |
| 阻止 HTTP 客户端的站点 | **Chrome** — 真实浏览器绕过机器人检测 |
| 大页面快照 / 遍历 | **Lite** — 3× 更快的快照 |
| 大文章的文本提取 | **Chrome** — Readability.js 更准确 |
| 屏幕截图、PDF、评估、Cookie | **Chrome** — Lite 不支持 |

---

## 代码布局

| 文件 | 目的 |
|------|------|
| `internal/engine/engine.go` | `Engine` 接口、`Capability` 常量、`Mode` 枚举、`NavigateResult` / `SnapshotNode` 类型 |
| `internal/engine/lite.go` | `LiteEngine` — HTTP 获取、脚本剥离、Gost-DOM 解析、角色映射 |
| `internal/engine/router.go` | `Router` — 有序规则链、`AddRule` / `RemoveRule` |
| `internal/engine/rules.go` | `CapabilityRule`、`ContentHintRule`、`DefaultLiteRule`、`DefaultChromeRule` |
| `internal/handlers/navigation.go` | `useLite()` 快速路径、`X-Engine` 头 |
| `internal/handlers/snapshot.go` | Lite 路径的 `SnapshotNode → A11yNode` 转换 |
| `internal/handlers/text.go` | Lite 文本快速路径 |
| `cmd/pinchtab/cmd_bridge.go` | 启动时从 `config.Engine` 进行路由器接线 |

---

## 依赖

| 包 | 版本 | 许可证 | 目的 |
|------|------|--------|------|
| `github.com/gost-dom/browser` | v0.11.0 | MIT | 无头浏览器：HTML 解析、DOM 遍历、事件分发 |
| `github.com/gost-dom/css` | v0.1.0 | MIT | CSS 选择器评估 |
| `golang.org/x/net` | 现有 | BSD-3 | 用于脚本剥离的 HTML 分词器 |