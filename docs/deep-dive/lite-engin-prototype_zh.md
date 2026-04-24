# Lite Engine：使用 Gost-DOM 的无 Chrome DOM 捕获

**分支：** `feat/lite-engine-gostdom`
**问题：** [#201](https://github.com/pinchtab/pinchtab/issues/201)
**相关草稿 PR：** [#200](https://github.com/pinchtab/pinchtab/pull/200)
**依赖：** [gost-dom/browser v0.11.0](https://github.com/gost-dom/browser)（MIT，~255 stars，Go 78.4%）

---

## 概述

此实现添加了一个 **Lite Engine**，可以执行 DOM 捕获（导航、快照、文本提取、点击、输入），而不需要 Chrome/Chromium。它使用 [Gost-DOM](https://github.com/gost-dom/browser)，一个用纯 Go 编写的无头浏览器，来解析和遍历 HTML 文档。

该架构遵循维护者的指导，实现了 **"无需触摸其余代码即可扩展的智能路由"** — 通过带有可插拔规则的策略模式路由器实现。

## 架构

### 引擎接口 (`internal/engine/engine.go`)

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

### 路由器 (`internal/engine/router.go`)

路由器评估 `RouteRule` 实现的有序链。第一个返回非 `Undecided` 裁决的规则获胜。

```
请求 → 路由器 → [规则 1] → [规则 2] → ... → [回退规则] → 引擎
```

规则可通过 `AddRule()` / `RemoveRule()` 在运行时热交换 — 无需更改处理程序代码。

### 三种模式

| 模式 | 行为 | 默认规则 |
|------|----------|---------------|
| `chrome` | 所有请求 → Chrome（默认，向后兼容） | DefaultChromeRule |
| `lite` | DOM 操作 → Gost-DOM，截图/PDF/评估 → Chrome | CapabilityRule → DefaultLiteRule |
| `auto` | 基于 URL 模式的每请求路由 | CapabilityRule → ContentHintRule → DefaultChromeRule |

### 内置规则 (`internal/engine/rules.go`)

| 规则 | 目的 |
|------|---------|
| `CapabilityRule` | 将截图/pdf/评估/cookies 路由到 Chrome（lite 无法执行这些） |
| `ContentHintRule` | 将 `.html/.htm/.xml/.txt/.md` URL 路由到 Lite（用于导航/快照/文本） |
| `DefaultLiteRule` | 捕获所有：将所有 DOM 操作路由到 Lite |
| `DefaultChromeRule` | 最终回退：将所有内容路由到 Chrome |

### 可扩展性

添加新的路由逻辑只需：
1. 实现 `RouteRule` 接口（2 个方法：`Name()`，`Decide()`）
2. 调用 `router.AddRule(myRule)` — 插入到回退规则之前

无需更改处理程序、配置或 CMD。

## 更改的文件

### 新文件（8 个）
| 文件 | 目的 | 行数 |
|------|---------|-------|
| `internal/engine/engine.go` | 引擎接口、类型、能力 | ~70 |
| `internal/engine/lite.go` | 使用 Gost-DOM 的 LiteEngine 实现 | ~430 |
| `internal/engine/router.go` | 带有 AddRule/RemoveRule 的路由器 | ~120 |
| `internal/engine/rules.go` | 4 个内置 RouteRule 实现 | ~95 |
| `internal/engine/lite_test.go` | LiteEngine 单元测试 | ~280 |
| `internal/engine/router_test.go` | 路由器单元测试 | ~130 |
| `internal/engine/rules_test.go` | 规则单元测试 | ~115 |
| `internal/engine/realworld_test.go` | 真实网站比较测试 | ~570 |

### 修改的文件（8 个）
| 文件 | 更改 |
|------|--------|
| `internal/config/config.go` | 向 RuntimeConfig + ServerConfig 添加 `Engine` 字段 |
| `internal/handlers/handlers.go` | 添加 `Router *engine.Router` 字段，`useLite()` 辅助函数 |
| `internal/handlers/navigation.go` | ensureChrome 之前的 Lite 快速路径 |
| `internal/handlers/snapshot.go` | 带有 SnapshotNode → A11yNode 转换的 Lite 快速路径 |
| `internal/handlers/text.go` | 返回纯文本的 Lite 快速路径 |
| `cmd/pinchtab/cmd_bridge.go` | 基于配置模式的引擎路由器连接 |
| `go.mod` | 添加 gost-dom/browser v0.11.0，gost-dom/css v0.1.0 |
| `go.sum` | 更新校验和 |

## 相比 PR #200 草稿的改进

| 领域 | PR #200 | 此实现 |
|------|---------|-------------------|
| 标签页管理 | 单个窗口 | 带有序列 ID 的多标签页 |
| HTML 解析 | `browser.Open()` 双重获取 | HTTP 获取 → 剥离脚本 → `html.NewWindowReader` |
| 脚本处理 | 在 `<script>` 标签上 panic | 通过 `x/net/html` 分词器预解析剥离 |
| 点击安全性 | 无 panic 保护 | Click 方法中的 `defer recover()` |
| 文本输出 | 原始 DOM 文本 | `normalizeWhitespace()` — 折叠空白运行 |
| 角色映射 | 基本（a, button, input 等） | 扩展：section→region, details→group, summary→button, dialog, article |
| 交互检测 | 基本标签 | 添加 summary，ARIA 角色（tab, menuitem, switch） |
| 路由 | 无（始终 lite） | 带有可插拔规则的策略模式路由器 |
| 配置 | 无 | 配置文件支持（`server.engine`） |

## 测试结果

### 引擎包测试（40+ 测试，全部通过）

```
=== 单元测试 ===
TestLiteEngine_Navigate          PASS
TestLiteEngine_Snapshot_All      PASS
TestLiteEngine_Snapshot_Interactive  PASS
TestLiteEngine_Text              PASS
TestLiteEngine_Click             PASS
TestLiteEngine_Type              PASS
TestLiteEngine_RefNotFound       PASS
TestLiteEngine_ScriptStyleSkipped  PASS
TestLiteEngine_AriaAttributes    PASS
TestLiteEngine_MultiTab          PASS
TestLiteEngine_Close             PASS
TestLiteEngine_Capabilities      PASS
TestLiteEngine_Name              PASS
TestNormalizeWhitespace          PASS

=== 路由器测试 ===
TestRouterChromeMode             PASS
TestRouterLiteMode               PASS
TestRouterAutoModeStaticContent  PASS
TestRouterAutoModeLiteNil        PASS
TestRouterAddRemoveRule          PASS
TestRouterRulesSnapshot          PASS

=== 规则测试 ===
TestCapabilityRule (9 cases)     PASS
TestContentHintRule (9 cases)    PASS
TestDefaultLiteRule (7 cases)    PASS
TestDefaultChromeRule (4 cases)  PASS
```

### 真实网站比较测试（16 个套件，63+ 子测试）

| 套件 | 模拟 | 子测试 | 结果 |
|-------|-----------|----------|--------|
| WikipediaStyle | Wikipedia 文章页面 | 9 | PASS |
| HackerNewsStyle | HN 首页 | 4 | PASS |
| EcommerceStyle | 带表单的产品页面 | 9 | PASS |
| FormHeavy | 注册表单 | 7 | PASS |
| AriaHeavy | 带 ARIA 角色的仪表板 | 11 | PASS |
| DeeplyNested | 5+ 级 div 嵌套 | 4 | PASS |
| SpecialCharacters | Unicode，HTML 实体，CJK | 3 | PASS |
| EmptyPage | 空 HTML 主体 | 1 | PASS |
| NonHTMLContentType | JSON 响应 | 1 | PASS |
| HTTP404 | 404 错误页面 | 1 | PASS |
| LargePagePerformance | 200 个部分，800+ 节点 | 1 | PASS |
| MultipleScriptTags | head+body 中的 5 个脚本标签 | 1 | PASS |
| InlineStyles | head+body 中的样式标签 | 1 | PASS |
| ClickWorkflow | 按钮点击 | 1 | PASS |
| ClickLinkRecovery | 锚点点击 panic 恢复 | 1 | PASS |
| TypeWorkflow | 输入到所有文本框 | 1 | PASS |

### 完整项目测试套件

```
ok   cmd/pinchtab           2.8s
ok   internal/allocation    2.0s
ok   internal/config        1.6s
ok   internal/dashboard     3.1s
ok   internal/engine        1.4s   ← 新包
ok   internal/handlers      6.8s
ok   internal/human         10.7s
ok   internal/idpi          2.0s
ok   internal/idutil        1.8s
ok   internal/instance      2.6s
ok   internal/orchestrator  3.2s
ok   internal/profiles      2.8s
ok   internal/proxy         2.8s
ok   internal/scheduler     4.0s
ok   internal/semantic      1.6s
ok   internal/strategy      1.7s
ok   internal/uameta        1.1s
ok   internal/web           1.5s
```

## 已知边缘情况和限制

| 边缘情况 | 行为 | 缓解措施 |
|-----------|----------|------------|
| HTML 中的 `<script>` 标签 | Gost-DOM panic（nil ScriptHost） | 通过 x/net/html 分词器预解析剥离 |
| 点击 `<a href>` | Gost-DOM 导航，可能遇到脚本 | Click 中的 `defer recover()`，返回错误 |
| CSS `display:none` | 元素仍在快照中出现 | Lite 引擎没有 CSS 引擎 |
| JavaScript 渲染的内容 | 未捕获（SPA，动态 DOM） | 在自动模式下回退到 Chrome |
| 截图 / PDF | lite 不支持 | CapabilityRule 路由到 Chrome |
| Cookies / Evaluate | lite 不支持 | CapabilityRule 路由到 Chrome |
| `<noscript>` 内容 | 从快照中剥离 | 与禁用脚本的行为一致 |

## 配置

在配置文件中设置引擎：
```json
{
  "server": {
    "engine": "lite"
  }
}
```

### 响应头
Lite 提供的响应包含 `X-Engine: lite` 头，用于可观测性。

## 依赖分析

| 包 | 大小 | 许可证 | 目的 |
|---------|------|---------|---------|
| gost-dom/browser v0.11.0 | ~2.5MB 源代码 | MIT | 无头浏览器（HTML 解析，DOM 遍历） |
| gost-dom/css v0.1.0 | ~200KB | MIT | CSS 选择器支持 |
| golang.org/x/net（现有） | 已在 go.mod 中 | BSD-3 | 用于脚本剥离的 HTML 分词器 |

## 性能基准：Lite vs Chrome

**Lite 运行：** 2026-03-09 | **Chrome 运行：** 2026-03-09
**方法：** 8 个真实网站 × 每个 4 个操作（导航 → 快照全部 → 快照交互 → 文本）

### 响应时间（ms）

| 网站 | Lite 导航 | Lite 快照（全部） | Lite 文本 | Chrome 导航 | Chrome 快照（全部） | Chrome 文本 | 获胜者 |
|---------|:------------:|:--------------:|:---------:|:--------------:|:----------------:|:-----------:|:------:|
| Example.com | 38ms | 23ms | 29ms | 396ms | 46ms | 34ms | **LITE** |
| Wikipedia (Go) | 657ms | 775ms | 120ms | 1310ms | 2703ms | 201ms | **LITE** |
| Hacker News | 1032ms | 188ms | 21ms | 1218ms | 247ms | 27ms | **LITE** |
| httpbin.org | 1117ms | 31ms | 24ms | 4745ms | 187ms | 47ms | **LITE** |
| GitHub Explore | 1402ms | 161ms | 24ms | 6156ms | 329ms | 20ms | **LITE** |
| DuckDuckGo | 119ms | 26ms | 20ms | 1488ms | 394ms | 41ms | **LITE** |
| Wikipedia (CS) | 215ms | 535ms | 687ms | 2668ms | 1249ms | 130ms | **LITE** |
| Stack Overflow | ❌ 502 | 694ms | 111ms | 6433ms | 376ms | 61ms | **CHROME** |

> Stack Overflow 阻止机器人 HTTP 请求 — Lite 引擎的 `Navigate` 获得 502。Chrome 通过真实浏览器会话处理此问题。

### 总计（两个引擎都成功的 7 个站点）

| 指标 | Lite | Chrome | 速度提升 |
|--------|-----:|-------:|--------:|
| 导航总计 | 4,580ms | 17,981ms | **3.9×** 更快 |
| 快照总计 | 1,739ms | 5,155ms | **3.0×** 更快 |
| 文本总计 | 925ms | 500ms | 0.5×（Chrome 更快） |
| **总计** | **7,244ms** | **23,636ms** | **3.3× 更快** |

> Lite 在 **7/8 个站点** 上总体获胜。Chrome 在文本提取方面更快，因为它在浏览器中运行 Mozilla Readability.js。Lite 执行原始 DOM 文本遍历，对于非常大的文章较慢（例如 Wikipedia CS：687ms vs 130ms）。

### 节点计数比较

| 网站 | Lite 节点 | Chrome 节点 | Lite 交互 | Chrome 交互 | Lite 文本（字符） | Chrome 文本（字符） |
|---------|:----------:|:------------:|:----------------:|:-----------------:|:----------------:|:-----------------:|
| Example.com | 6 | 8 | 1 | 1 | 125 | 209 |
| Wikipedia (Go) | 6,074 | 7,110 | 1,276 | 1,063 | 75,659 | 62,859 |
| Hacker News | 805 | 975 | 229 | 229 | 4,025 | 4,169 |
| httpbin.org | 62 | 113 | 5 | 29 | 274 | 1,179 |
| GitHub Explore | 1,533 | 830 | 331 | 240 | 8,340 | 368 |
| DuckDuckGo | 143 | 655 | 20 | 102 | 123 | 7,231 |
| Wikipedia (CS) | 4,941 | 4,653 | 1,627 | 1,061 | 79,799 | 58,071 |
| Stack Overflow | — | 779 | — | 192 | — | 23,671 |

> **节点计数为何不同：** Lite 在解析前剥离 `<script>` 标签，并且没有 CSS 引擎，因此隐藏元素仍然出现。Chrome 的可访问性树会修剪隐藏/不可见元素。DuckDuckGo 和 GitHub Explore 显示 Chrome 文本较少，因为 Chrome 的 Readability.js 剥离导航/侧边栏内容，而 Lite 捕获所有可见文本。

### 关键结论

| 场景 | 推荐 |
|----------|---------------|
| 静态站点、维基、新闻、博客 | **Lite** — 3–12× 更快，无 Chrome 开销 |
| JavaScript 渲染的 SPA（React、Next.js 等） | **Chrome** — Lite 仅捕获 JS 前的 HTML |
| 阻止 HTTP 机器人的站点（Stack Overflow、一些社交网站） | **Chrome** — 真实浏览器绕过机器人检测 |
| 大页面上的快照 / DOM 遍历 | **Lite** — Wikipedia 上 3× 更快的快照 |
| 大文章上的文本提取 | **Chrome** — Readability.js 更准确、更快 |
| 需要截图 / PDF / 评估的管道 | **Chrome** — Lite 不支持这些 |

*基准测试于 2026-03-09 从 `tests/lite_engine_benchmark.ps1` 运行*