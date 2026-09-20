# 导航（Navigate）

将当前被跟踪标签页导航到某个 URL，或在没有当前标签页时新建一个。

```bash
curl -X POST http://localhost:9867/navigate \
  -H "Content-Type: application/json" \
  -d '{"url":"https://pinchtab.com"}'
# CLI Alternative
pinchtab nav https://pinchtab.com
# Response (terminal: tab ID, then the landed URL; piped: tab ID only; --json for full JSON)
8f9c7d4e1234567890abcdef12345678
https://pinchtab.com/
```

API 响应：`{"tabId":"...","url":"<落点 URL>","title":"...","route":...}`。

## 命令行界面 Flags

`pinchtab nav <url>` 在默认本地服务器尚未运行时自动启动它。当 `--server` 或 `PINCHTAB_SERVER` 指向另一台服务器时，PinchTab 连接那台服务器而不自动启动新进程。默认配置启动无头浏览器，因此成功的 `pinchtab nav` 可能不打开可见窗口。用 `--snap` 查看结果，或在想要可见浏览器时以有头模式运行 PinchTab。隐藏别名：`goto`、`navigate`、`open`。

| Flag | 说明 |
|------|-------------|
| `--tab` | 复用现有标签页 |
| `--new-tab` | 强制新标签页 |
| `--block-images` | 阻止图片加载 |
| `--block-ads` | 阻止广告 |
| `--dismiss-banners` | 落点后，点击可见的 cookie/同意关闭按钮或移除明显的覆盖层 |
| `--timeout` | 导航超时秒数（最大 120）；覆盖新标签页 30s 的上限 |
| `--snap` | 导航后输出快照 |
| `--snap-diff` | 导航后输出快照差异 |
| `--text` | 导航后输出页面文本 |
| `--print-tab-id` | 仅打印标签页 ID（管道时自动）；与 `--snap`/`--text` 一起时标签页 ID 进 stderr |
| `--json` | 完整 JSON 响应 |

## 示例

```bash
pinchtab nav https://example.com              # Navigate current tab, or create one
pinchtab nav https://example.com --snap       # Navigate and snapshot
TAB=$(pinchtab nav https://example.com)       # Capture tab ID for reuse
pinchtab nav https://other.com --tab "$TAB"   # Reuse tab
pinchtab nav https://example.com --new-tab    # Force another tab
pinchtab nav https://example.com --block-images  # Skip images
```

## API 主体字段

| 字段 | 说明 |
|-------|-------------|
| `url` | 目标 URL（必填） |
| `tabId` | 复用现有标签页 |
| `newTab` | 强制新标签页 |
| `blockImages` | 阻止图片加载 |
| `blockMedia` | 阻止媒体加载 |
| `blockAds` | 阻止广告 |
| `dismissBanners` | 落点后关闭 cookie/同意横幅 |
| `timeout` | 导航超时秒数（上限 120） |
| `waitTitle` | 最多等待 N 秒标题出现（上限 30） |
| `waitFor` | 等待条件：`none`（默认）、`dom`、`selector`、`networkidle` |
| `waitSelector` | 等待的选择器；`waitFor` 为 `selector` 时必填 |
| `dispatchOnly` | 一旦导航已派发就返回 `{tabId,url,dispatched:true}`，不等待加载 |
| `browser` | 将请求路由到的浏览器 |

`GET /navigate?url=...` 接受相同字段作为查询参数，但 `blockImages`、`blockMedia`、`blockAds` 除外。

## 行为

- 顶层 `POST /navigate` 在未提供 `tabId` 时打开新标签页，除非调用方已被识别（会话或代理 id）且有当前标签页，则复用它。在严格空指针策略下，已识别但无当前标签页的调用方得到 `409 no_current_tab`。
- `pinchtab nav <url>` 在有可用的被跟踪标签页时用它；否则打开新标签页。
- `POST /tabs/{id}/navigate`、带 `tabId` 的 `POST /navigate`、`pinchtab nav <url> --tab <id>` 复用指定标签页，并使其成为后续未限定操作的当前标签页。
- `--new-tab` 和 `newTab:true` 即使当前已有标签页也强制新标签页。
- 不带 `--tab` 的命令使用当前被跟踪标签页。聚焦或使用某标签页会更新该当前标签页指针；指针过期时 PinchTab 回退到最近使用的被跟踪标签页。
- 当保存的当前标签页指针指向服务器已不再持有的标签页时，`pinchtab nav` 不带标签页 id 重试导航。对会话或代理-id 调用方（通过 `--agent-id` 或 `PINCHTAB_AGENT_ID`，任一），该重试复用该作用域的当前标签页，命令行界面保持静默，因为没有创建任何东西。对两者都没有的调用方，服务器打开一个**新**标签页——文档化的匿名契约——因此命令行界面在 stderr 打印 `HINT`，同时指出消失的那个和新建的那个，而非在让你在同一 URL 上留下两个标签页的情况下报告纯成功。设置 `PINCHTAB_SESSION` 运行可保持单一工作面。显式 `--tab` 绝不重试：它直接暴露 404。

错误：无效 URL 为 `400`，`waitFor` 条件不受支持或不满足时为 `400 bad_wait_for`，目标被拦截时为 `403`（IDPI 域名策略或私有/内部地址），标签页上打开着 JavaScript 对话框时为 `409 dialog_blocked`，标签页因 [交接](./handoff.md) 暂停时为 `409 tab_paused_handoff`，重定向循环时为 `422`。

理由：命令行界面默认保持一个明显的工作面。有意另开标签页时用 `--new-tab`，需要特定标签页时用 `--tab`/`tabId`。

## 相关页面

- [Snapshot](./snapshot.md)
- [Tabs](./tabs.md)
