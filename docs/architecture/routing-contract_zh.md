# 浏览器路由契约（Browser Routing Contract）

## 概览

PinchTab 中的路由决定由哪个浏览器 provider 处理给定请求。系统采用 `CanHandle` 模式：每个 provider 针对给定的请求意图报告自己的能力。

## 决策类型

- **`DecisionHandle`**——"我能服务此请求；使用此浏览器继续"
- **`DecisionSkip`**——"我无法服务此形状/意图；若有回退可用，调用方可尝试另一个 provider。绝不可用于安全拒绝。"
- **`DecisionFail`**——"致命的 provider 错误；立即中止请求"

## 请求形状（Request Shapes）

每个形状常量对请求所代表的工作种类进行分类：

| 常量 | 值 | 操作 |
|----------|-------|------------|
| `ShapeStaticRead` | `static-read` | 轻量 DOM 读取（无需渲染） |
| `ShapeStaticSnapshot` | `static-snapshot` | 轻量快照抓取 |
| `ShapeRenderedRead` | `rendered-read` | 需要完整渲染的 DOM 读取（Chrome） |
| `ShapeVisual` | `visual` | 截图、PDF 生成 |
| `ShapeInteraction` | `interaction` | 点击、输入、按键、滚动等 |
| `ShapeSessionState` | `session-state` | 会话/cookie 管理 |
| `ShapeNetworkControl` | `network-control` | 网络拦截、HAR 抓取 |
| `ShapeDownloadUpload` | `download-upload` | 文件下载/上传操作 |

### StateChanging 标志

当 `RequestIntent.StateChanging` 为 true 时，该请求会改变浏览器状态。仅处理只读操作的 provider（例如 ghost-chrome 的静态抓取）应对状态改变的请求返回 `DecisionSkip`，即使该形状在其他情况下可接受。

## 安全不变量

安全拒绝（域名封锁、IDPI 内容封锁、私有/内部 IP 封锁、重定向限制）：

- 在**处理器层**强制实施，先于任何浏览器特有的执行
- **绝不**可回退——403 是终局，无论选中了哪个浏览器
- 与能力决策分离——一个返回 `DecisionHandle` 的浏览器不会覆盖安全策略

## Provider 能力（当前）

| Provider | 处理的形状 | 备注 |
|----------|---------------|-------|
| chrome | 所有形状 | 完整 CDP 浏览器；处理一切 |
| cloak | 所有形状 | 反检测的 Chrome 分支；与 chrome 能力相同 |
| ghost-chrome | StaticRead、StaticSnapshot（非状态改变） | 跳过所有其他形状；内部可升级到 Chrome |

## Ghost-Chrome 升级

- Ghost-chrome 的升级位于 `bridgekit.BridgeAdapter`（`internal/browsers/ghostchrome/bridgekit/`）中，它包装 Chrome `BridgeAPI`，并为 `Navigate`、`Snapshot` 和 `Text` 实现静态优先、再到 Chrome 的策略
- 这**不是**通用的多浏览器回退——它是一个 provider 内部硬编码的 ghost 到 chrome 的升级
- 当 ghost-chrome 的静态尝试未通过质量门（`ghostchrome.AssessContent`：SPA 标记、稀薄内容）时，它会自动升级到 Chrome

## 回退策略（当前 vs 未来）

- **当前**：`DecisionSkip` 将请求降级为 `chrome`（`internal/handlers/browser_routing.go` 中的 `resolveBrowserForRequest`）。不存在有序的多 provider 回退；`chrome` 是唯一的回退目标。
- **未来**：`DecisionSkip` 应触发按优先级顺序回退到下一个可用 provider。安全拒绝必须保持不可回退。

## 路由顺序

1. 解析浏览器 provider（`config.ResolveBrowser`）：请求（`?browser=` 查询或 body 中的 `browser`），然后是会话浏览器，再然后是标签页所属实例的浏览器，再是 `browsers.default`，再是 `browsers.available` 的第一个，否则 `chrome`。未知 provider 返回 400
2. 对解析出的 provider 调用 `CanHandle(intent)`
3. 若 `DecisionSkip`——降级为 `chrome`（未来：尝试下一个 provider）
4. 若 `DecisionFail`——返回 400 及错误
5. 若 `DecisionHandle`——进入安全检查
6. 应用安全策略（域名、IDPI、IP、重定向）
7. 若安全拒绝——返回 403（绝不回退）
8. 用选中的 provider 执行请求
