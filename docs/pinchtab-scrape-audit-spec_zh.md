# PinchTab 抓取与审计规格（PinchTab Scraping & Audit Spec）

> **HEAD 处的状态：** 这是最初的设计规格。其中大部分已作为 `pinchtab audit`、
> `pinchtab compare` 和 `pinchtab scrape` 交付——已交付的契约见 [audit.md](audit.md) 与
> [scrape.md](scrape.md)。下文标有 *(not shipped)* 的条目描述的是代码尚未实现的行为。

## 概述（Overview）

PinchTab 充当 SeaPortal 的发现与 HTTP 提取之上的**深度浏览器增强层**。它补充了纯 HTTP
抓取无法提供的视觉、交互、性能与安全能力。

## 核心命令（Core Command）

```bash
pinchtab audit <url> [flags]
pinchtab compare <live-url> <staging-url> [flags]
```

## 职责（Responsibilities）

1. **消费 SeaPortal 输出**作为基础数据
2. **为采样页面增强**浏览器专属数据
3. **运行视觉回归**（在比较不同版本时）
4. **生成最终富报告**（或把结构化数据发给 LLM）

## 输入（Input）

- SeaPortal `SiteReport` JSON（或直接给 URL + 站点地图数据）——*已交付为一个由 SeaPortal
  `Result` 对象组成的 JSON 数组（`--seaportal-report`）、一个 URL，或一个站点地图
  （`--sitemap`）*
- 可选：鉴权细节（cookies、登录流程）——*cookies 与 `--profile` 已交付；登录流程未交付*
- 对比基线（用于 diff 模式）

## 输出（Output）

增强后的 `SiteReport`，每个页面带上浏览器增强字段。*已交付为一份带版本号的
`AuditReport`（`internal/audit/types.go`），而非一个扩展版 `SiteReport`。*

## 页面增强字段（PinchTab）

```go
type BrowserPageData struct {
    ScreenshotPath      string
    FullPageScreenshot  bool
    ConsoleLogs         []ConsoleLogEntry
    NetworkRequests     []NetworkRequest
    BrokenAssets        []BrokenAsset      // especially 404 images
    InteractiveElements []InteractiveElement
    AccessibilityScore  int
    VisualDiff          *VisualDiffResult   // only in compare mode
    TimingMetrics       BrowserTimingMetrics
}
```

*已交付时额外增加了一个 `JSErrors []JSError` 字段。`FullPageScreenshot` 存在但从不被
设置：审计截图只覆盖视口。*

## 关键特性（Key Features）

### 1. 视觉分析
- 整页截图（*未交付：审计截图只覆盖视口*）
- 图片 diff（用于 compare 模式）
- 带标注的高亮 diff 图片

### 2. 控制台与 JS 监控
- 捕获所有控制台日志、错误、警告
- 过滤关键 JS 错误

### 3. 网络监控
- 跟踪所有资源请求
- 检测失效图片、脚本、样式表（404、500 等）
- 监控 API 失败

### 4. 性能（浏览器层）
- Core Web Vitals（经由 CDP）
- 导航计时
- 资源加载分解（*未作为分解项交付；逐请求的 `duration` 在 `networkRequests` 中*）

### 5. 可用性与无障碍
- 基础 a11y 检查
- 缺失的表单 label、alt 文本
- 导航流程校验（*未交付*）

### 6. 安全面（可选）
- 供 Nuclei 或类似工具接入的集成点（*未交付；仅有内置规则*）
- 暴露端点检测
- 混合内容警告

## 采样与可扩展性（Sampling & Scalability）

- 尊重 SeaPortal 的页面分组与采样
- 允许覆盖采样大小
- 并行处理，并发可配置
- 智能优先级排序（首页、关键流程优先）——*已交付为：入口 URL 最先，未分组页面先于模板分组
  采样；无关键流程检测*

### 预览 → 展开（大型站点）（Preview → Expand）

以完整保真度抓取一个大型站点，成本会双倍叠加：传输每个页面的 markdown 正文主宰了 token
成本，而对每个被路由页面做浏览器渲染则主宰了挂钟时间成本。为了让模型先概览、再下钻，抓取
拆成「先廉价、后深入」的两轮：

1. **预览（Preview）**（`scrape <url> --preview`，`preview: true`）：做 HTTP 爬取和逐页
   路由判定，但**不做浏览器渲染**、**不抓完整页面正文**。每个页面的 markdown 被扣留，替换为：
   - `charCount`——提取内容的大小，便于调用方估量完整展开会有多重
   - `snippet`——内容开头一段、折叠空白后的切片
   - 该页面已有的 `meta`/描述与 `contentType`

   其结果是一份模型可廉价阅读、用以判断何者重要的大纲。

2. **展开（Expand）**（`scrape <url> --only <url> [--only <url> …]`，`only: [...]`）：
   以完整保真度抓取**精确选定的那些 URL**（按正常路由做 HTTP 提取 + 浏览器渲染），而不是
   重新发现站点。它是无状态的——模型把它在预览中看到的那些确切 URL 交回来即可；不在服务端
   保留任何会话或缓存。

两轮复用同一套导航安全栈：每一次预览爬取 fetch、每一次展开 fetch（带 `CrawlGuard`
策略的 `FetchBytes`）都走与浏览器导航相同的 SSRF/重定向审查。

## 报告结构（Report Structure）

（参见此前合并版报告模板，其中包含以下章节：）
- 汇总评分
- SEO 与元数据
- 内容与功能
- 视觉差异
- 性能
- 控制台与 JS 错误
- 失效资源 / 图片
- 可用性问题
- 安全发现
- 建议

## 集成点（Integration Points）

- 库模式：`pinchtab.EnrichWithBrowser(seaPortalReport)`——*已交付为
  `pkg/pinchtabaudit` 中的
  `pinchtabaudit.New(baseURL, token).EnrichWithBrowser(ctx, AuditInput, opts)`*
- 供独立使用的命令行界面
- 对 Docker 友好，便于 CI/CD
- 输出格式：JSON（供 LLM）、Markdown/HTML 报告、PDF

## 鉴权支持（Authentication Support）

- Cookie 注入
- 登录流程录制/回放（*未交付*）
- 无头 + 隐身选项（*无审计专属 flags*）

## Flags（示例）

- `--sample-size int`
- `--screenshot`
- `--network-monitor`
- `--visual-diff`
- `--llm-refine`（*未交付*）
- `--output-dir`
- `--concurrency int`

---

**状态**：可供进入实现规划。（*已被取代：见顶部状态说明。*）

这与 SeaPortal 规格完美互补。
