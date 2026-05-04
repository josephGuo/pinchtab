# Pinchtab API 参考

所有示例的基础 URL：`http://localhost:9867`

> **命令行界面替代方案：** 所有端点都有命令行界面等价物。使用 `pinchtab help` 获取完整列表。示例中显示为 `# 命令行界面:` 注释。

## 代理归属

如果代理直接调用 HTTP API，请在应该保持归属的请求上包含 `X-Agent-Id: <agent-id>`。

示例：

```bash
curl -X POST /navigate \
  -H 'X-Agent-Id: agent-crawl-01' \
  -H 'Content-Type: application/json' \
  -d '{"url": "https://pinchtab.com"}'
```

注意事项：

- 命令行界面用户应优先使用 `pinchtab --agent-id <agent-id> ...` 而不是手动设置 header
- scheduler 提交的任务在执行时会将其 `agentId` 重用为 `X-Agent-Id`
- 省略的 `tabId` 按调用者身份解析：代理会话使用会话范围的当前标签页，`X-Agent-Id` 在没有会话时使用代理范围的当前标签页，匿名请求使用共享的全局/默认标签页

## 导航

```bash
# 命令行界面: pinchtab nav https://pinchtab.com [--new-tab] [--block-images]
curl -X POST /navigate \
  -H 'Content-Type: application/json' \
  -d '{"url": "https://pinchtab.com"}'

# 带选项：自定义超时、阻止图像、在新标签页中打开
curl -X POST /navigate \
  -H 'Content-Type: application/json' \
  -d '{"url": "https://pinchtab.com", "timeout": 60, "blockImages": true, "newTab": true}'
```

## 快照（可访问性树）

```bash
# 命令行界面: pinchtab snap [-i] [-c] [-d] [-s main] [--max-tokens 2000]
# 完整树
curl /snapshot

# 仅交互元素（按钮、链接、输入）——小得多
curl "/snapshot?filter=interactive"

# 限制深度
curl "/snapshot?depth=5"

# 智能差异——仅自上次快照以来的更改（大量令牌节省）
curl "/snapshot?diff=true"

# 文本格式——缩进树，比 JSON 少约 40-60% 的令牌
curl "/snapshot?format=text"

# 紧凑格式——每行一个节点，比 JSON 少 56-64% 的令牌（推荐）
curl "/snapshot?format=compact"

# YAML 格式
curl "/snapshot?format=yaml"

# 限定到 CSS 选择器（例如仅主内容）
curl "/snapshot?selector=main"

# 截断到约 N 个令牌
curl "/snapshot?maxTokens=2000"

# 组合以获得最大效率
curl "/snapshot?format=compact&selector=main&maxTokens=2000&filter=interactive"

# 捕获前禁用动画
curl "/snapshot?noAnimations=true"

# 写入文件
curl "/snapshot?output=file&path=/tmp/snapshot.json"
```

返回带有 `ref`、`role`、`name`、`depth`、`value`、`nodeId` 的平面 JSON 节点数组。

**令牌优化**：使用 `?format=compact` 获得最佳令牌效率。为面向操作的任务添加 `?filter=interactive`（减少约 75% 的节点）。使用 `?selector=main` 限定到相关内容的范围内。使用 `?maxTokens=2000` 限制输出。使用 `?diff=true` 在多步骤工作流程中仅查看更改。自由组合所有参数。

## 对元素执行操作

```bash
# 命令行界面: pinchtab click e5 / pinchtab type e12 hello / pinchtab press Enter
# 通过引用点击
curl -X POST /action -H 'Content-Type: application/json' \
  -d '{"kind": "click", "ref": "e5"}'

# 在聚焦元素中输入（首先点击，然后输入）
curl -X POST /action -H 'Content-Type: application/json' \
  -d '{"kind": "click", "ref": "e12"}'
curl -X POST /action -H 'Content-Type: application/json' \
  -d '{"kind": "type", "ref": "e12", "text": "hello world"}'

# 按键
curl -X POST /action -H 'Content-Type: application/json' \
  -d '{"kind": "press", "key": "Enter"}'

# 聚焦元素
curl -X POST /action -H 'Content-Type: application/json' \
  -d '{"kind": "focus", "ref": "e3"}'

# 填充（直接设置值，无按键）
curl -X POST /action -H 'Content-Type: application/json' \
  -d '{"kind": "fill", "selector": "#email", "text": "user@pinchtab.com"}'

# 悬停（触发下拉菜单/工具提示）
curl -X POST /action -H 'Content-Type: application/json' \
  -d '{"kind": "hover", "ref": "e8"}'

# 移动指针而不点击
curl -X POST /action -H 'Content-Type: application/json' \
  -d '{"kind": "mouse-move", "ref": "e8"}'

# 按下并释放鼠标按钮
curl -X POST /action -H 'Content-Type: application/json' \
  -d '{"kind": "mouse-down", "button": "left"}'
curl -X POST /action -H 'Content-Type: application/json' \
  -d '{"kind": "mouse-up", "button": "left"}'

# 在元素或坐标处滚动滚轮
curl -X POST /action -H 'Content-Type: application/json' \
  -d '{"kind": "mouse-wheel", "ref": "e8", "deltaY": 240}'
curl -X POST /action -H 'Content-Type: application/json' \
  -d '{"kind": "mouse-wheel", "x": 400, "y": 320, "deltaY": -320}'

# 选择下拉选项（按值或可见文本）
curl -X POST /action -H 'Content-Type: application/json' \
  -d '{"kind": "select", "ref": "e10", "value": "option2"}'

# 滚动到元素
curl -X POST /action -H 'Content-Type: application/json' \
  -d '{"kind": "scroll", "ref": "e20"}'

# 按像素滚动（无限滚动页面）
curl -X POST /action -H 'Content-Type: application/json' \
  -d '{"kind": "scroll", "scrollY": 800}'

# 点击并等待导航（链接点击）
curl -X POST /action -H 'Content-Type: application/json' \
  -d '{"kind": "click", "ref": "e5", "waitNav": true}'
```

注意：

- 基于选择器的点击和双击路径在分派指针事件之前通过后端节点 ID 解析
- 低级指针操作接受 `ref`、`selector`、`nodeId` 或 `x`/`y`
- `mouse-down` 和 `mouse-up` 接受 `button` 参数，值为 `left`、`right` 或 `middle`
- `mouse-wheel` 接受 `deltaX` 和 `deltaY`；省略时，旧版 `scrollX` / `scrollY` 仍然有效
- `mouse-down`、`mouse-up` 和 `mouse-wheel` 在不传递新目标时使用当前指针位置

## 等待页面状态

```bash
# 命令行界面: pinchtab wait 'text:Done' / pinchtab wait --url '**/dashboard'
curl -X POST /wait -H 'Content-Type: application/json' \
  -d '{"selector":"text:Done","timeout":15000}'

curl -X POST /wait -H 'Content-Type: application/json' \
  -d '{"url":"**/dashboard","timeout":15000}'

curl -X POST /wait -H 'Content-Type: application/json' \
  -d '{"load":"networkidle","timeout":15000}'

curl -X POST /wait -H 'Content-Type: application/json' \
  -d '{"fn":"document.readyState === \"complete\"","timeout":15000}'
```

## 批量操作

```bash
# 按顺序执行多个操作
curl -X POST /actions -H 'Content-Type: application/json' \
  -d '{"actions":[{"kind":"click","ref":"e3"},{"kind":"type","ref":"e3","text":"hello"},{"kind":"press","key":"Enter"}]}'

# 第一个错误时停止（默认：false）
curl -X POST /actions -H 'Content-Type: application/json' \
  -d '{"tabId":"TARGET_ID","actions":[...],"stopOnError":true}'
```

## 提取文本

```bash
# 命令行界面: pinchtab text [--raw]
# Readability 模式（默认）——剥离导航/页脚/广告
curl /text

# 原始 innerText
curl "/text?mode=raw"
```

返回 `{url, title, text}`。最便宜的选项（大多数页面约 1K 令牌）。

默认模式选择第一个**可见**的 `<article>` / `[role="main"]` / `<main>`（跳过 `display:none`）并剥离导航/页脚/广告。使用 `mode=raw` 获取完整 `innerText`，或使用 `/snapshot` 获取结构化 UI 文本（如价格和按钮标签）。

## PDF 导出

除非用户明确希望将文件写入磁盘，否则优先返回 base64 或原始字节。写入磁盘时，使用安全的临时或工作区路径。

```bash
# 命令行界面: pinchtab pdf --tab TAB_ID [-o file.pdf] [--landscape] [--scale 0.8]
# 返回 base64 JSON
curl "/tabs/TAB_ID/pdf"

# 原始 PDF 字节
curl "/tabs/TAB_ID/pdf?raw=true" -o page.pdf

# 保存到安全临时位置的磁盘
curl "/tabs/TAB_ID/pdf?output=file&path=/tmp/pinchtab-page.pdf"

# 横向带自定义比例
curl "/tabs/TAB_ID/pdf?landscape=true&scale=0.8&raw=true" -o page.pdf

# 自定义纸张尺寸（Letter: 8.5x11, A4: 8.27x11.69）
curl "/tabs/TAB_ID/pdf?paperWidth=8.5&paperHeight=11&marginTop=0.5&marginLeft=0.5&raw=true" -o custom.pdf

# 导出特定页面
curl "/tabs/TAB_ID/pdf?pageRanges=1-5&raw=true" -o pages.pdf

# 带页眉/页脚
curl "/tabs/TAB_ID/pdf?displayHeaderFooter=true&headerTemplate=%3Cspan%20class=title%3E%3C/span%3E&raw=true" -o header.pdf

# 无障碍 PDF 带文档大纲
curl "/tabs/TAB_ID/pdf?generateTaggedPDF=true&generateDocumentOutline=true&raw=true" -o accessible.pdf

# 尊重 CSS 页面尺寸
curl "/tabs/TAB_ID/pdf?preferCSSPageSize=true&raw=true" -o css-sized.pdf
```

**查询参数：**

| 参数 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| `paperWidth` | float | 8.5 | 纸张宽度（英寸） |
| `paperHeight` | float | 11.0 | 纸张高度（英寸） |
| `landscape` | bool | false | 横向方向 |
| `marginTop` | float | 0.4 | 上边距（英寸） |
| `marginBottom` | float | 0.4 | 下边距（英寸） |
| `marginLeft` | float | 0.4 | 左边距（英寸） |
| `marginRight` | float | 0.4 | 右边距（英寸） |
| `scale` | float | 1.0 | 打印比例（0.1–2.0） |
| `pageRanges` | string | all | 要导出的页面（例如 `1-3,5`） |
| `displayHeaderFooter` | bool | false | 显示页眉和页脚 |
| `headerTemplate` | string | — | 页眉的 HTML 模板 |
| `footerTemplate` | string | — | 页脚的 HTML 模板 |
| `preferCSSPageSize` | bool | false | 尊重 CSS `@page` 尺寸 |
| `generateTaggedPDF` | bool | false | 生成无障碍/标记 PDF |
| `generateDocumentOutline` | bool | false | 嵌入文档大纲 |
| `output` | string | JSON | `file` 保存到磁盘，默认返回 base64 |
| `path` | string | auto | 目标路径（使用 `output=file` 时优先使用临时或工作区路径） |
| `raw` | bool | false | 返回原始 PDF 字节而不是 JSON |

包装 `Page.printToPDF`。默认打印背景图形。

## 下载文件

除非用户明确要求保存文件，否则优先使用原始字节或 base64 响应。

```bash
# 默认返回 base64 JSON（使用浏览器会话/cookies/隐身）
curl "/download?url=https://site.com/report.pdf"

# 原始字节（管道到文件）
curl "/download?url=https://site.com/image.jpg&raw=true" -o image.jpg

# 直接保存到安全临时位置的磁盘
curl "/download?url=https://site.com/export.csv&output=file&path=/tmp/pinchtab-export.csv"
```

## 上传文件

仅上传用户明确提供或批准用于任务的本地文件。

```bash
# 将本地文件上传到文件输入
curl -X POST "/upload?tabId=TAB_ID" -H "Content-Type: application/json" \
  -d '{"selector": "input[type=file]", "paths": ["/tmp/user-approved-photo.jpg"]}'

# 上传 base64 编码数据
curl -X POST /upload -H "Content-Type: application/json" \
  -d '{"selector": "#avatar-input", "files": ["data:image/png;base64,iVBOR..."]}'
```

通过 CDP 在 `<input type=file>` 元素上设置文件。触发 `change` 事件。如果省略，选择器默认为 `input[type=file]`。

## 截图

```bash
# 命令行界面: pinchtab ss [-o file.jpg] [-q 80]
# 返回原始 JPEG（默认）
curl "/screenshot?raw=true" -o screenshot.jpg
curl "/screenshot?raw=true&quality=50" -o screenshot.jpg

# 返回原始 PNG
curl "/screenshot?raw=true&format=png" -o screenshot.png
```

## 执行 JavaScript

谨慎使用。首先优先使用 `text`、`snapshot` 和正常操作。默认使用只读 DOM 检查，除非用户明确要求，否则避免读取 cookie、localStorage 或无关的页面秘密。

```bash
# 命令行界面: pinchtab eval "document.title"
curl -X POST /evaluate -H 'Content-Type: application/json' \
  -d '{"expression": "document.title"}'

# 响应前解析返回的 promise
curl -X POST /evaluate -H 'Content-Type: application/json' \
  -d '{"expression": "Promise.resolve(document.title)", "awaitPromise": true}'
```

当表达式返回 promise 且你想要解析的值时，设置 `awaitPromise: true`。如果省略，行为保持不变。

## 标签页管理

```bash
# 命令行界面: pinchtab tabs / pinchtab nav <url> --new-tab / pinchtab tabs close <id>
# 列出标签页
curl /tabs

# 打开新标签页
curl -X POST /tab -H 'Content-Type: application/json' \
  -d '{"action": "new", "url": "https://pinchtab.com"}'

# 关闭标签页
curl -X POST /close -H 'Content-Type: application/json' \
  -d '{"tabId": "TARGET_ID"}'
# 省略 tabId 以关闭当前/默认标签页。
curl -X POST /close -H 'Content-Type: application/json' -d '{}'

# 或使用标签页范围的路由
curl -X POST /tabs/TARGET_ID/close
```

多标签页：向 snapshot/screenshot/text 传递 `?tabId=TARGET_ID`，或在 POST body 中传递 `"tabId"`。显式标签页 ID 始终覆盖并更新调用者的当前标签页范围。

## 标签页特定端点

所有读取/操作端点都有使用 `/tabs/{id}/...` 的标签页范围变体：

```bash
# 导航特定标签页
curl -X POST /tabs/TARGET_ID/navigate \
  -H 'Content-Type: application/json' \
  -d '{"url": "https://pinchtab.com"}'

# 快照特定标签页
curl "/tabs/TARGET_ID/snapshot"
curl "/tabs/TARGET_ID/snapshot?filter=interactive&format=compact"

# 截图特定标签页
curl "/tabs/TARGET_ID/screenshot?raw=true" -o tab-screenshot.jpg

# 从特定标签页提取文本
curl "/tabs/TARGET_ID/text"

# 在特定标签页上执行操作
curl -X POST /tabs/TARGET_ID/action \
  -H 'Content-Type: application/json' \
  -d '{"kind": "click", "ref": "e5"}'

# 在特定标签页上批量操作
curl -X POST /tabs/TARGET_ID/actions \
  -H 'Content-Type: application/json' \
  -d '{"actions": [{"kind": "click", "ref": "e3"}, {"kind": "type", "ref": "e3", "text": "hello"}]}'

# 在特定标签页上等待
curl -X POST /tabs/TARGET_ID/wait \
  -H 'Content-Type: application/json' \
  -d '{"selector":"text:Done","timeout":15000}'

# 暂停自动化以进行人工干预
curl -X POST /tabs/TARGET_ID/handoff \
  -H 'Content-Type: application/json' \
  -d '{"reason":"captcha","timeoutMs":120000}'

# 检查或恢复 handoff 状态
curl /tabs/TARGET_ID/handoff
curl -X POST /tabs/TARGET_ID/resume \
  -H 'Content-Type: application/json' \
  -d '{"status":"completed"}'
```

这些等效于在顶级端点上使用 `?tabId=TARGET_ID`，但遵循 REST 约定。标签页 ID 来自 `/tabs` 或导航/标签页创建响应中的 `tabId` 字段。

`GET /tabs/{id}/handoff` 返回该标签页的当前状态。`POST /tabs/{id}/resume` 清除 `paused_handoff` 并可以携带恢复元数据，如 `status` 或 `resolvedData`。

## 标签页锁定（多代理）

```bash
# 锁定标签页（默认 30 秒超时，最大 5 分钟）
curl -X POST /lock -H 'Content-Type: application/json' \
  -d '{"tabId": "TARGET_ID", "owner": "agent-1", "timeoutSec": 60}'

# 解锁
curl -X POST /unlock -H 'Content-Type: application/json' \
  -d '{"tabId": "TARGET_ID", "owner": "agent-1"}'
```

锁定的标签页在 `/tabs` 中显示 `owner` 和 `lockedUntil`。冲突时返回 409。

## Cookies

```bash
# 获取当前页面的 cookies
curl /cookies

# 设置 cookies
curl -X POST /cookies -H 'Content-Type: application/json' \
  -d '{"url":"https://pinchtab.com","cookies":[{"name":"session","value":"abc123"}]}'
```

## 解决挑战

PinchTab 包含一个可插拔的求解器框架，用于解决浏览器挑战（Cloudflare Turnstile、CAPTCHA、插页式广告）。求解器自动检测挑战类型并使用类人交互解决它。

```bash
# 列出可用求解器
curl /solvers

# 自动检测并解决（按顺序尝试每个求解器）
curl -X POST /solve -H 'Content-Type: application/json' \
  -d '{"maxAttempts": 3, "timeout": 30000}'

# 使用特定求解器按名称
curl -X POST /solve/cloudflare -H 'Content-Type: application/json' \
  -d '{"maxAttempts": 3}'

# 在特定标签页上解决
curl -X POST /tabs/TAB_ID/solve -H 'Content-Type: application/json' \
  -d '{"solver": "cloudflare"}'

# 在特定标签页上使用基于路径的求解器解决
curl -X POST /tabs/TAB_ID/solve/cloudflare -H 'Content-Type: application/json' \
  -d '{}'
```

**请求字段：**

| 字段 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| `solver` | string | — | 求解器名称（省略以自动检测） |
| `tabId` | string | — | 目标标签页（省略以使用默认标签页） |
| `maxAttempts` | int | 3 | 最大解决尝试次数 |
| `timeout` | float | 30000 | 总超时（毫秒） |

**响应：**

```json
{
  "tabId": "DEADBEEF",
  "solver": "cloudflare",
  "solved": true,
  "challengeType": "managed",
  "attempts": 1,
  "title": "Example Site"
}
```

未检测到挑战时返回 `solved: true, attempts: 0`——可以安全地推测性调用。

**内置求解器：** `cloudflare`（Turnstile/插页式广告——通过页面标题检测，使用类人输入点击复选框）。

**隐身要求：** 求解器在 `stealthLevel: "full"` 下效果最佳。Cloudflare 在复选框点击前后检查浏览器指纹。使用 `GET /stealth/status` 验证隐身是否激活。

## 网络导出

```bash
# 导出为 HAR 1.2（流式传输到响应）
curl /network/export?format=har

# 导出为 NDJSON（每行一个 JSON）
curl /network/export?format=ndjson

# 保存到服务器端文件
curl "/network/export?format=har&output=file&path=session.har"

# 包含响应正文（每条目 10 MB 上限）
curl "/network/export?format=har&body=true"

# 包含原始敏感头（Cookie、Authorization）
curl "/network/export?format=har&redact=false"

# 实时流式导出（条目到达时写入文件）
curl -N "/network/export/stream?format=ndjson&path=live.ndjson"

# 标签页范围
curl /tabs/TAB_ID/network/export?format=har
```

所有标准网络过滤器都适用：`filter`、`method`、`status`、`type`、`limit`。

格式是可插拔的。`GET /network/export?format=unknown` 返回 `{"available": ["har", "ndjson"]}`。

## 隐身

```bash
# 检查隐身状态和分数
curl /stealth/status

# 轮换浏览器指纹
curl -X POST /fingerprint/rotate -H 'Content-Type: application/json' \
  -d '{"os":"windows"}'
# os: "windows", "mac", 或省略以随机
```

## 健康检查

```bash
curl /health
```

## 会话认证

如果用户已经给你代理会话令牌，按如下方式发送：

```bash
curl -H "Authorization: Session ses_..." /health
```