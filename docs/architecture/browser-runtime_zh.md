# 浏览器运行时架构（Browser Runtime Architecture）

负责人：bridge、browsers
相关：[browser-abstraction.md](browser-abstraction.md)、[routing-contract.md](routing-contract.md)

## 概览

处理器零 chromedp/cdproto 导入。所有浏览器操作都经由 BridgeAPI（约 75 个方法）。CDP 的使用被约束在 bridge 层和 cdptk 共享工具包内。启动后的行为归属 Bridge；provider 通过 `Capabilities()` 以声明式方式塑造它。

## 分层图

```
┌─────────────────────────────────────────────────────────┐
│  Handlers (internal/handlers)                           │
│                                                         │
│  Zero chromedp imports. Zero cdproto imports.            │
│  All operations via bridge.BridgeAPI.                    │
│  TabContext() returns *TabHandle (opaque).               │
│                                                         │
│  screenshot → Bridge.CaptureScreenshot(ctx, ...)        │
│  screencast → Bridge.StartScreencast(ctx, opts)         │
│  record     → captureFrame closure (injected at init)   │
│  evaluate   → Bridge.Evaluate(ctx, expr, &out, opts)    │
│  cookies    → Bridge.GetCookies(ctx) / SetCookie(ctx)   │
│  DOM        → Bridge.CallFunctionOnNode(ctx, ...)       │
│  download   → Bridge.DownloadURL(ctx, url, opts)        │
│  emulation  → Bridge.SetViewport / SetGeolocation / ... │
│  navigation → Bridge.CurrentURL / CurrentTitle          │
│  actions    → Bridge.ExecuteAction(ctx, kind, req)      │
└──────────────────────────┬──────────────────────────────┘
                           │ BridgeAPI (domain types, no CDP types)
                           ▼
┌─────────────────────────────────────────────────────────┐
│  Bridge (BridgeAPI — ~75 methods)                       │
│                                                         │
│  Owns: lifecycle, tab routing, locks, auto-close,       │
│        network monitoring, CDP connection               │
│                                                         │
│  Visual:     CaptureScreenshot, StartScreencast         │
│  Evaluate:   Evaluate, CallFunctionOnNode,              │
│              EvaluateInFrame                             │
│  DOM:        DescribeNode, ResolveSelectorToNodeID,      │
│              SetFileInputFiles                           │
│  Cookies:    GetCookies, SetCookie                       │
│  Emulation:  SetViewport, SetGeolocation,                │
│              SetEmulatedMedia                            │
│  Network:    SetNetworkConditions, SetExtraHTTPHeaders,  │
│              EnableNetwork, ListenNetworkEvents           │
│  Navigation: CurrentURL, CurrentTitle, GoBack,           │
│              GoForward, Reload                           │
│  Download:   DownloadURL (Fetch interception + Network)  │
│  Auth:       EnableFetchWithAuth                         │
│  PDF:        PrintToPDF                                  │
│  Stealth:    SetUserAgentOverride,                       │
│              AddScriptToEvaluateOnNewDocument            │
│  Tabs:       ListTargets → []TabTarget (bridge type)     │
│                                                         │
│  BridgeAPI signatures use domain types, not CDP types.   │
│  CDP types never appear in BridgeAPI signatures.         │
└──────────────────────────┬──────────────────────────────┘
                           │ chromedp (internal)
                           ▼
┌─────────────────────────────────────────────────────────┐
│  internal/cdptk/ (shared CDP toolkit)                   │
│                                                         │
│  Pure functions. No state. No browser ownership.        │
│  Takes a chromedp context, returns data.                │
│                                                         │
│  cdptk.CaptureWithSurfaceFallback(fromSurface, capture) │
│  cdptk.ClipForNode(ctx, nodeID, css1x) → *ScreenshotClip│
│  cdptk.StartRepaintLoop(ctx) → stop func                │
│  cdptk.InjectInteractiveOverlay / AnnotationRectForNode │
│                                                         │
│  Used by the bridge; handlers call it only for the      │
│  screenshot/annotate overlay helpers.                   │
└─────────────────────────────────────────────────────────┘
```

## 浏览器接口（启动前）

`Browser` 接口仅供启动前使用：ID、DisplayName、BuildLaunchArgs、CanHandle、DiscoverBinary、DoctorChecks、GeoAlignment、Capabilities、ValidateTarget、SupportsRemoteCDP、ClassifyLaunchError。

```go
// internal/browsers/chrome/chrome.go
type Browser struct{}           // implements browsers.Browser (pre-launch)
```

不存在按 provider 区分的启动后运行时类型。需要不同运行时行为的 provider 声明一个 capability，由 Bridge 据此分支——见[基于能力的路由](#基于能力的路由capability-based-routing)。早期设计曾让每个 provider 各自实现一套与 Bridge 已实现的约 20 个操作相同的运行时实现；那份第二份实现与实际运行的那份发生漂移，已被删除。在伸手去做平行实现之前，先试试用一个 capability。

## TabHandle

`TabContext()` 返回 `*TabHandle` 而非 `context.Context`。`TabHandle` 实现了 `context.Context`（Deadline/Done/Err/Value 全部委托给底层 CDP context），但这个返回类型向处理器表明：这个 context 只应传给 Bridge 方法。

## 基于能力的路由（Capability-based routing）

浏览器能力（`CapabilitySet`）驱动运行时行为。例如，`CapEventScreencast` 控制录屏（screencast）策略：

- Chrome 声明 `CapEventScreencast` → 事件驱动的录屏
- Cloak 省略它 → 基于轮询的录屏（与无头路径相同）

`shouldUsePollingScreencast()` 同时检查 `Config.Headless` 和浏览器的能力集。

## 关键模式

- **BridgeAPI 签名使用领域类型，而非 CDP 类型。** 截图返回 `[]byte`，而非 `*page.CaptureScreenshotReturns`。

- **CallFunctionOnNode** 集中处理 DOM.resolveNode → Runtime.callFunctionOn 这一模式。被 attr、box、checked、enabled、visible、value、inspect 和 text 处理器使用。

- **ScreencastStream** 返回 `*ScreencastStream`，其 `Frames <-chan []byte`。两种策略由能力选择。

- **DownloadURL** 封装了完整的 Fetch 拦截 + 网络监控状态机。

- **captureFrame 闭包**——录屏器接收一个在构造时注入的 `captureFrame` 函数，绑定到 `Bridge.CaptureScreenshot`。

- **TabTarget**——bridge 级类型取代 `*target.Info`，以避免 cdproto 导入泄漏到处理器中。

## 各层的 CDP 使用情况

| 层 | 使用的 CDP 域 | 用途 |
|---|---|---|
| bridge/ | Page, DOM, Runtime, Network, Fetch, Emulation, Input, Target | 处理器委托下来的所有浏览器操作 |
| cdptk/ | Page, DOM, Runtime | 共享的纯函数 CDP 包装 |
| browsers/chrome/ | Runtime（经由 `chromedp.Evaluate`） | 仅用于启动探测诊断；无启动后操作 |
| handlers/ | 不直接使用 | 所有操作经由 BridgeAPI；标注覆盖层经由 cdptk 辅助函数 |

## 非目标

- 替换 chromedp 作为 CDP 客户端库。
- 在 v1 中支持非 Chromium 浏览器（但架构本身不阻止这一点）。
