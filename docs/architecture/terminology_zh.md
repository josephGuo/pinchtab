# 术语：浏览器选择概念（Terminology: Browser Selection Concepts）

浏览器选择、配置和 provider 路由所用词汇的规范参考。

## 核心术语

| 术语 | 范围 | 定义 |
|-----------------|----------|------------|
| **browser** | 公开 | 面向用户的选择概念。用户在配置、命令行 flag 和 API 调用中选择浏览器。在所有面向用户的语境中取代 "engine"。 |
| **provider** | 公开 | `chrome | cloak | ghost-chrome` 之一。浏览器选择背后的实现。每个 provider 映射到一套独特的启动与路由策略。 |
| **static fetch** | 公开 | `ghost-chrome` 在升级到 Chrome 之前使用的轻量 HTTP+DOM 路径。在面向用户的语言中取代 "lite engine"。 |
| **engine** | 已移除 | 不再是公开概念：配置文件中的 `server.engine` 会校验失败。旧的 `chrome` / `lite` / `auto` 值仅作为内部的 `browsers.LaunchMode`（`internal/browsers/config.go`）保留。 |
| **target** | 配置 | 配置文件中 `browser.targets` 下的一个命名启动画像（provider、二进制路径、代理、flags 等），由 `browser.defaultTarget` 命名默认值。在接受 `browser` 值的地方（例如 `POST /instances/attach`），可以给定一个 target 名而非 provider。 |

## Provider 描述

### `chrome`

经由 CDP 的完整 Chrome。默认 provider。启动一个本地的基于 Chromium 的浏览器，并通过 Chrome DevTools Protocol 连接。

### `cloak`

CloakBrowser——反检测的 Chrome 分支。使用打了补丁的 Chromium 二进制，带指纹随机化、时区/区域伪装和 WebRTC 泄漏防护。

### `ghost-chrome`

静态优先路由。先尝试轻量 HTTP 抓取和 DOM 解析；当内容稀薄、动态或需要 JavaScript 执行时，升级到完整的 Chrome 会话。

## 迁移说明

- `ServerConfig` 中的 `engine` 字段不再受支持：它仍会被解析，仅仅是为了让配置校验能拒绝它。请改用配置文件中的 `browsers.default`。`RuntimeConfig` 没有 engine 字段。
- `browser {}` 配置块内部的 `provider` 字段不再受支持（校验错误）；请使用顶层的 `browsers.default` 键。按 target 的 `browser.targets.<name>.provider` 仍然必需。
- 公开文档、命令行帮助文本和 API 响应应使用 "browser" 和 "provider"——绝不用 "engine" 或 "lite engine"。
