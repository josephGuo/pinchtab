# 浏览器抽象层（Browser Abstraction）

状态：已实现（BridgeAPI 封装与浏览器注册表抽取已完成）
负责人：bridge

## 问题（历史背景）

在抽取注册表之前，新增一个浏览器需要修改至少五个互不相干的文件：provider 特有的逻辑散落在 `config/`、`bridge/runtime/`、`browserprobe/` 和 `doctor/` 中，靠字符串开关和按 provider 区分的谓词函数耦合在一起。这些分支点现在各自收敛到一个抽象上：

| 关注点 | 位置 | 现今机制 |
|---|---|---|
| 启动参数 | `internal/browsers/<id>/` + `internal/browsers/runtimekit/` | 通过 `ResolveProviderLaunchPlan` 调用各 provider 的 `BuildLaunchArgs` |
| 地理对齐 | `internal/browsers/` 配置 | provider 能力 + 启动计划上的 `GeoConfig` |
| 二进制发现 | `internal/browsers/<id>/` | 各 provider 的 `DiscoverBinary()` |
| Doctor 检查 | `internal/doctor/browsers.go` | 来自注册表的各 provider `DoctorChecks()` |
| 配置校验 | `internal/config/browser_targets.go` | 由注册表驱动（`browsers.Get`/`browsers.IDs`） |
| 生命周期清理 | `internal/browsers/providerhooks/` | 已注册的 `Hooks`（装饰、清理、关闭） |

`Engine`（`chrome` / `lite` / `auto`）——已从配置中移除（`server.engine` 现在会校验失败）；由浏览器 provider 模型（`chrome` / `cloak` / `ghost-chrome`）取代。关于规范的 provider 定义，见 [routing-contract.md](routing-contract.md) 与 [terminology.md](terminology.md)。`Capabilities` 是叠加在其上的额外概念，已部分接入。

## 目标

新增一个浏览器实现应当只需：

1. 在 `internal/browsers/<id>/` 下创建一个子包。
2. 通过 `browsers.Register(...)` 注册它。

新增一个 provider 不应再需要修改运行时分发逻辑（bridge、doctor、校验器、地理分发）。对新 provider 而言，公开配置 schema、文档和 UI 可能仍需更新。

## 设计

### 包布局

```
internal/browsers/
  browser.go           // Browser interface + registry
  config.go            // LaunchConfig, GeoConfig, TargetConfig (provider-neutral)
  capabilities.go      // CapabilitySet (moved from config/)
  runtimekit/          // shared launch-plan resolution (ResolveProviderLaunchPlan)
  providerhooks/       // optional bridge decoration + cleanup/shutdown hooks
  all/, builtin/       // barrel packages that blank-import every built-in provider
  chrome/              // chrome implementation, registers itself
  cloak/               // cloak implementation, composes chrome
  ghostchrome/         // static-first provider (staticfetch + bridgekit adapter)
  lightpanda/          // future
  brave/               // future
```

`browserprobe/`、`bridge/runtime/init.go` 中的 provider 分支、以及 `doctor` 的 provider 门控均已移除，改为委托给已注册的 `Browser`。

### 接口

```go
package browsers

type Browser interface {
    ID() string                              // "chrome", "cloak", "lightpanda"
    DisplayName() string                     // human-readable
    Capabilities() CapabilitySet

    // Discovery & doctor
    DiscoverBinary() BinaryDiscovery
    DoctorChecks(cfg TargetConfig) []DoctorCheck

    // Launch
    BuildLaunchArgs(cfg LaunchConfig) (args []string, env []string, err error)
    SupportsRemoteCDP() bool

    // Strategy hooks
    GeoAlignment(geo GeoConfig) GeoStrategy
    ValidateTarget(cfg TargetConfig) error
    ClassifyLaunchError(f LaunchFailure) LaunchErrorKind

    // Request routing (see routing-contract.md)
    CanHandle(intent RequestIntent) HandleDecision
}

type GeoStrategy struct {
    Flags          []string
    Env            []string
    OperatorWins   bool  // explicit user config overrides geo-derived values
}

var registry = map[string]Browser{}

func Register(b Browser)               { registry[b.ID()] = b } // panics on duplicate ID
func Get(id string) (Browser, bool)    { b, ok := registry[id]; return b, ok }
func IDs() []string                    { /* sorted list */ }
```

每个实现都在其包的 `init()` 中注册自身：

```go
// internal/browsers/chrome/chrome.go
func init() { browsers.Register(&Browser{}) }
```

导入方通过一个 barrel 文件（例如 `internal/browsers/all/all.go`）来接线 provider，这样调用方只需一次空导入即可获得所有内置 provider。

### 启动模式

`LaunchMode`（`internal/browsers/config.go`）仅供内部使用：`chrome`、`lite` 或 `auto`。`runtimekit` 依据配置的默认浏览器推导它，而 `ResolveLaunchMode` 将 `auto`（以及未知值）映射为 `chrome`。它不是公开配置——公开选择使用浏览器 provider 名（`chrome` / `cloak` / `ghost-chrome`）；旧的公开 `Engine` 字段已移除。见 [terminology.md](terminology.md)。

有头（headed）与无头（headless）是 `LaunchConfig` 上一个独立的 `Headless` 布尔字段：

```go
type LaunchConfig struct {
    Mode       LaunchMode // chrome | lite | auto (internal only)
    Binary     string
    ProfileDir string
    Proxy      ProxyConfig
    ExtraFlags []string
    Headless   bool       // true → --headless=new + swiftshader flags
    // ...
}
```

provider 可以拒绝它无法服务的模式：`chrome` 和 `cloak` 对 `lite` 都会返回错误。

### Cloak 作为对 Chrome 的组合

Cloak = Chrome + 隐身（stealth）flags + 更严格的地理优先级。不要复制——请嵌入：

```go
// internal/browsers/cloak/cloak.go
type Browser struct{ chrome.Browser }

func (b Browser) ID() string          { return "cloak" }
func (b Browser) DisplayName() string { return "CloakBrowser" }

func (b Browser) BuildLaunchArgs(cfg browsers.LaunchConfig) ([]string, []string, error) {
    args, env, err := b.Browser.BuildLaunchArgs(cfg)
    if err != nil { return nil, nil, err }
    return append(args, cloakFlagArgs(cfg)...), env, nil
}

func (b Browser) GeoAlignment(geo browsers.GeoConfig) browsers.GeoStrategy {
    s := cloakGeoFlags(geo)
    s.OperatorWins = true  // explicit user config wins
    return s
}

func (b Browser) DiscoverBinary() browsers.BinaryDiscovery {
    return discoverCloakBinary()  // separate paths from chrome
}
```

### Doctor

`doctor/runner.go` 是一个薄编排器，仅从所配置的浏览器拉取检查项：

```go
func Registry(cfg *config.RuntimeConfig) []CheckEntry {
    entries := []CheckEntry{{Name: "config_file", Fn: checkConfigFile}}
    browserID := config.NormalizeBrowser(browserFromCfg(cfg))
    if b, ok := browsers.Get(browserID); ok {
        for _, dc := range b.DoctorChecks(browsers.TargetConfig{Provider: browserID}) {
            entries = append(entries, /* CheckEntry wrapping dc */)
        }
    }
    // then binary_exists, binary_executable, binary_starts
    return entries
}
```

`internal/doctor/browsers.go`（`ReportBrowsers`）则单独遍历所有已注册浏览器，用于多浏览器报告。

provider 特有的检查（`cdp_reachable`、`fingerprint_flags_accepted`、`linux_fonts_present`）移入 `browsers/cloak/doctor.go`。

### 配置校验

`browser_targets.go` 用以下逻辑取代其硬编码的白名单：

```go
if _, ok := browsers.Get(target.Provider); !ok {
    return fmt.Errorf("unknown browser provider %q (known: %v)", target.Provider, browsers.IDs())
}
return browsers.MustGet(target.Provider).ValidateTarget(target)
```

### Bridge 集成

`bridge/runtime/init.go` 不再按 provider 分支。它只解析一次 `Browser`，然后委托：

```go
b, ok := browsers.Get(cfg.Provider)
if !ok { return fmt.Errorf("unsupported provider %q", cfg.Provider) }

args, env, err := b.BuildLaunchArgs(launchCfg)
geo := b.GeoAlignment(geoCfg)
// apply args + env + geo.Flags + geo.Env
```

`InitRemoteCDP` 在 attach 之前会检查 `b.SupportsRemoteCDP()`。

## 迁移计划

先抽取，再重新接线。每一步独立落地，不改变行为。

1. **骨架。** 添加 `internal/browsers/`，包含接口、注册表和空的 `chrome` 包。在一处（`cmd/pinchtab/root.go`）接好 barrel 导入。此时尚不改任何调用点。
2. **Chrome 抽取。** 将 Chrome 特有的启动辅助、地理逻辑和 provider 发现移到 `browsers/chrome/` 之后，外加共享的 `runtimekit` 入口点。旧调用点通过 `browsers.Get("chrome")` / `FindBrowserBinary(...)` 委托。测试保持绿色。
3. **Cloak 抽取。** 将隐身 flag 构建器、cloak 地理、cloak 二进制发现移入 `browsers/cloak/`，作为对 chrome 的组合。逐个移除 `IsCloakBrowserProvider()` 的调用方。
4. **Doctor 迁移。** 将 provider 特有检查移入 `browsers/{chrome,cloak}/doctor.go`。`doctor/runner.go` 坍缩为一次注册表遍历。
5. **地理迁移。** 用 `b.GeoAlignment(...)` 调用取代 `geo_align.go` 的 switch。删除 `geo_align.go` 的函数体。
6. **校验器迁移。** 用 `browsers.Get(...)` 检查取代 `browser_targets.go` 白名单。
7. **启动模式集成。** 将 `LaunchMode` 贯穿接入 `LaunchConfig`（仅供内部；面向公开的 engine 概念由浏览器 provider 路由取代）。`auto` 解析放在注册表或一个小助手函数中。
8. **通过桩（stub）验证。** 添加 `browsers/lightpanda/`，它完成注册但只把 `BuildLaunchArgs` 做成桩（返回 "not implemented"）。如果接好它需要修改 `browsers/lightpanda/` 之外的任何东西，说明抽象已经泄漏——在合并前修复。

## 处理器层的路由泄漏

### 问题

上面的 `browsers.Browser` 接口解决了启动/发现/doctor 这条轴——新增浏览器不再需要修改五个文件。但还有第二条抽象边界仍在泄漏：**处理器层的请求路由**。

Ghost-chrome 的内部路由（静态浏览器 vs Chrome）散布在处理器、server 接线层和一个路由包中。处理器代码里有显式的 `if browser == "ghost-chrome" && h.StaticBrowser != nil` 分支，`Handlers` 结构体携带一个 `StaticBrowser` 字段，路由包则对 `*ghostchrome.Browser` 做类型断言。Chrome 和 Cloak 很简单（一切都走 CDP）。Ghost-chrome 有内部路由（静态优先，升级到 Chrome）。这种区分应当对调用方不可见。

### 目标架构

```
┌──────────────────────────────────────────────────────┐
│                     Handlers                         │
│                                                      │
│  Only knows about BridgeAPI.                         │
│  No StaticBrowser field. No shouldUseStaticAction.   │
│  No ghost-chrome conditionals. No routing package.   │
│                                                      │
│  navigate() → Bridge.Navigate(url)                   │
│  snapshot() → Bridge.Snapshot(tabID, filter)          │
│  text()     → Bridge.Text(tabID)                     │
│  action()   → Bridge.TabContext(tabID)               │
│              → Bridge.ExecuteAction(ctx, kind, req)   │
│                                                      │
└──────────────────┬───────────────────────────────────┘
                   │ BridgeAPI interface
                   ▼
┌──────────────────────────────────────────────────────┐
│              BridgeAPI implementations               │
│                                                      │
│  ┌──────────────┐  ┌─────────────┐  ┌─────────────┐ │
│  │ Chrome       │  │ Cloak       │  │ GhostChrome  │ │
│  │ (bridge.New) │  │ (bridge.New │  │ Adapter      │ │
│  │              │  │  + stealth) │  │              │ │
│  │ All calls go │  │ All calls   │  │ Internally   │ │
│  │ to CDP       │  │ go to CDP   │  │ routes each  │ │
│  │              │  │ + stealth   │  │ call to      │ │
│  │              │  │ injection   │  │ static or    │ │
│  │              │  │             │  │ Chrome       │ │
│  └──────────────┘  └─────────────┘  └──────┬───────┘ │
└──────────────────────────────────────────────┼───────┘
                                               │
                          ┌────────────────────┴──────────────────┐
                          │         GhostChrome Adapter           │
                          │         (internal detail)             │
                          │                                      │
                          │  ┌─────────────┐  ┌───────────────┐  │
                          │  │ staticfetch  │  │ Chrome bridge │  │
                          │  │ (gost-dom)   │  │ (real CDP)    │  │
                          │  └─────────────┘  └───────────────┘  │
                          │                                      │
                          │  Navigate → static first, Chrome if  │
                          │             quality too low           │
                          │  Snapshot → static if tab is static,  │
                          │             Chrome if escalated       │
                          │  Text     → static if tab is static   │
                          │  Action   → static for click/type     │
                          │             with ref, Chrome otherwise │
                          │  Evaluate → always Chrome (escalate)  │
                          │  Screenshot → always Chrome (escalate)│
                          │                                      │
                          │  Ref mapping, quality gates, tab      │
                          │  escalation all happen HERE, not in   │
                          │  the handler layer.                   │
                          └──────────────────────────────────────┘
```

#### 当前状态（Phase 6 之后）

目标架构已实现至 Phase 6。关键组件：

- **`bridge.BridgeAPI`**（`internal/bridge/api.go`）——一个接口，除已有的标签页/动作操作外，还有 `Navigate`、`Snapshot`、`Text` 方法。
- **`bridgekit.BridgeAdapter`**（`internal/browsers/ghostchrome/bridgekit/bridge_adapter.go`）——为 ghost-chrome 路由包装 `BridgeAPI`。它嵌入 Chrome bridge，并将静态优先路由委托给 `ghostchrome.BridgeProxy`。它把 `NavigateParams` 作为网络策略强制实施，并在静态路径上把 `ContentParams.ContentGuard` 作为 IDPI 扫描强制实施。
- **处理器**——只使用 `BridgeAPI`。Navigate 与 Snapshot 处理器分别调用 `Bridge.Navigate()` 和 `Bridge.Snapshot()`。没有 `StaticBrowser` 字段，没有 ghost-chrome 条件分支，没有路由包的导入。
- **延迟启动的两阶段 navigate**——对于全新标签页的导航，处理器会探测 `StaticFirstNavigate()`，并以 `NavigateParams.NoEscalate` 运行第一阶段：适配器只尝试静态浏览器，并发出 `*bridge.StaticEscalateError` 信号，而不是在内部升级，从而仅在确实需要时才启动 Chrome。第二阶段以 `SkipStatic` 重新运行。超时预算是按阶段计算的（每个阶段都有 `NavigateTimeout`，默认 60 秒），因此一次发生升级的导航最多可能耗费所配置超时的两倍。
- **路由元数据**——对所有适配器路径，`usedProvider` 统一为 `"ghost-chrome"`。静态 vs Chrome 的路由记录在 `Attempts[]` 中。
- **`internal/routing/`**——已删除。
- **`internal/handlers/routing.go`**——已删除。
- **`bridgeProxyAdapter`**——已从 `server/bridge.go` 删除。

#### 关键原则

1. **处理器只知道 BridgeAPI。** 没有 `StaticBrowser` 字段，没有 `shouldUseStaticAction`，没有 `useStaticBrowser`，没有 `if browser == "ghost-chrome"` 条件分支。

2. **BridgeAPI 增加读操作。** `Navigate`、`Snapshot`、`Text` 成为 BridgeAPI 上的方法，这样 ghost-chrome 适配器就能像拦截 `TabContext` 和 `ExecuteAction` 一样拦截它们。

3. **ghost-chrome 适配器是一个 BridgeAPI 包装器。** 它包装真正的 Chrome bridge 和一个 `staticfetch.Browser`。每个 BridgeAPI 方法都被拦截：只读操作先尝试静态，仅 Chrome 的操作会升级，动作按 kind+ref 路由。所有路由逻辑都住在适配器内部。

4. **不为浏览器单独建路由包。** `routing.Route()` 函数及其 `ghostchrome.StaticFetcher` 适配器之所以存在，只是因为处理器无法通过 BridgeAPI 调用 `Browser.Route()`。一旦 BridgeAPI 有了 Navigate/Snapshot/Text，路由决策就发生在适配器内部。

5. **server 层是薄接线。** `configureBridgeRouter` 向 `providerhooks.DecorateBridge` 索取所配置浏览器的结果，并把 `h.Bridge` 设为该结果。仅此而已。没有 `h.StaticBrowser = ...`。

### 违规项

#### 已解决（Phase 1-5）

| ID | 描述 | 解决方式 |
|----|-------------|-------------|
| ~~V1~~ | `Handlers.StaticBrowser` 字段 | Phase 3——字段已删；处理器只使用 `BridgeAPI` |
| ~~V2~~ | snapshot 处理器中的 ghost-chrome 快速路径 | Phase 1+3——`BridgeAPI.Snapshot()` 封装了路由 |
| ~~V3~~ | text 处理器中的 ghost-chrome 快速路径 | Phase 1+3——`BridgeAPI.Text()` 封装了路由 |
| ~~V4~~ | 重复的 `useStaticBrowser` 快速路径 | Phase 3——辅助函数及所有调用点已删 |
| ~~V5~~ | `shouldUseStaticAction` / `executeStaticAction` | Phase 3——已删；适配器在内部路由动作 |
| ~~V6~~ | 静态动作跳过了标签页解析 | Phase 3——处理器始终调用 `tabContext()`；适配器处理静态标签页 |
| ~~V7~~ | 处理器中的 `staticFetcher` 适配器 | Phase 4——`internal/handlers/routing.go` 已删 |
| ~~V8~~ | `routing.Route()` 对 ghostchrome 做类型断言 | Phase 4——`internal/routing/` 包已删 |
| ~~V9~~ | 健康检查处理器对 StaticBrowser 短路 | Phase 3——`StaticBrowser` 字段已移除；健康检查使用 BridgeAPI |
| ~~V10~~ | server 层中的 `populateEscalatedRefCache` | Phase 5——ref 缓存逻辑移入 `bridgekit.BridgeAdapter` |
| ~~V11~~ | `bridgeProxyAdapter` 过厚 | Phase 5——由 `bridgekit.BridgeAdapter` 取代；server 适配器已删 |

#### 未解决

**~~V12: BridgeAPI 读方法不拥有安全策略的强制执行。~~** 已解决
ghost-chrome 适配器现在把 `NavigateParams` 作为 `staticfetch.NavigateNetworkPolicy` 强制实施（SSRF、重定向限制、可信 CIDR/IP），并通过 `ContentParams.ContentGuard` 扫描静态 Snapshot/Text 内容（IDPI 阻断/警告）。处理器对 Chrome 路径保留它们自己的预检；适配器确保静态路径有等价的强制执行。

**~~V13: 路由元数据由处理器构造。~~** 已解决
`NavigateResult`、`SnapshotResult` 和 `TextResult` 携带来自适配器的路由元数据。对所有适配器路径（静态被接受、已升级或回退），`usedProvider` 统一为 `"ghost-chrome"`。静态 vs Chrome 的路由细节记录在 `Attempts` 数组中。处理器在可用时传播适配器提供的路由元数据。

**V14: 目标/provider 记账被割裂（Phase 6 事项）。**
目标解析与 provider 选择仍涉及分散的配置查找。浏览器注册表抽取（迁移计划第 1-8 步）部分完成，但尚未端到端接好。

### 迁移路径（BridgeAPI 封装）

增量式。每个阶段独立落地。

**Phase 1——把读操作移入 BridgeAPI。** 已完成
`Navigate`、`Snapshot`、`Text` 被加入 `BridgeAPI` 接口（`internal/bridge/api.go`）。`bridge.Bridge` 委托给 Chrome CDP。`NavigateParams` 与 `ContentParams` 携带每次请求的安全策略。

**Phase 2——ghost-chrome 适配器直接实现 BridgeAPI。** 已完成
`bridgekit.BridgeAdapter`（`internal/browsers/ghostchrome/bridgekit/bridge_adapter.go`）包装 `bridge.BridgeAPI` 和一个 `ghostchrome.BridgeProxy`。所有 ghost-chrome 静态 vs Chrome 的路由决策都封装在此。`bridgekit` 的 `init()` 为 `ghost-chrome` 注册一个 `providerhooks.Hooks.DecorateBridge`，它调用 `bridgekit.NewBridgeAdapter(chromeBridge, cfg)`。

**Phase 3——从 Handlers 移除 StaticBrowser。** 已完成
`StaticBrowser` 字段、`useStaticBrowser()`、`shouldUseStaticAction()`、`executeStaticAction()`、snapshot/text 处理器中的 ghost-chrome 快速路径，以及 `staticFetcher` 适配器全部删除。处理器对 ghost-chrome 零感知。

**Phase 4——移除路由包。** 已完成
`internal/routing/` 包和 `internal/handlers/routing.go` 已删除。

**Phase 5——清理 BridgeProxy API 表面。** 已完成
`populateEscalatedRefCache` 和 `bridgeProxyAdapter` 从 `server/bridge.go` 移除。ref 缓存逻辑移入适配器。`configureBridgeRouter` 现在是薄接线：

```go
func configureBridgeRouter(h *handlers.Handlers, cfg *config.RuntimeConfig) {
    decorated := providerhooks.DecorateBridge(config.NormalizeBrowser(cfg.DefaultBrowser), h.Bridge, cfg)
    if decorated == h.Bridge {
        return
    }
    h.Bridge = decorated
}
```

**Phase 6——安全策略归属 + 路由元数据。** 基本完成
- ghost-chrome 适配器把 `NavigateParams` 作为网络策略强制实施，并把 `ContentParams.ContentGuard` 作为静态路径上的 IDPI 扫描强制实施（V12 已解决）。
- 路由元数据由适配器返回，由处理器消费（V13 已解决）。
- 剩余：目标/provider 记账的抽取尚未端到端接好（V14）。

## 非目标

- 替换 chromedp / CDP 客户端选择。
- 在动作层做按目标的能力门控。Capabilities 保持为建议性。
- 沙箱化或进程隔离的改动。
- 跨 provider 的 profile 互操作。

## 悬而未决的问题

1. **共享的 chrome 家族代码。** Cloak、Brave 以及带 chromium 模式的 lightpanda 都共享大部分 flag 构建。放在 `browsers/common/`，还是作为 `chrome.Browser` 上导出的辅助函数？倾向于从 `chrome` 导出，以避免第三个包。
2. **Doctor 检查的身份。** 现今检查有稳定的字符串 ID，供 CI 使用。把它们移入 provider 包必须保留这些 ID。在第 4 步之前做一次审计。
3. **地理 `OperatorWins` 语义。** 目前隐式编码在 cloak 地理里。是按提议提升为显式字段，还是保留在策略的 flag 列表内部？
4. **Engine `auto`。** 由 provider 解析，还是由一个使用 `Capabilities()` 的注册表辅助函数解析？大概是后者，以便策略集中在一处。

## v1 范围之外

- Brave、Firefox、lightpanda 的实际实现。
- 非 CDP 浏览器的 WebDriver 回退。
- 按 provider 的遥测命名空间。
