# Tabs（标签页）

标签页是浏览、提取、交互和诊断的主要执行面。

一旦已有标签页 ID，就用标签页范围的 HTTP 路由。命令行界面中，用带 `--tab <id>` 的常规顶级浏览器命令。

`pinchtab tab` 本身仅用于：

- 列出标签页
- 聚焦标签页
- 关闭标签页
- 人工交接（`handoff`、`handoff-status`、`resume`）

用 `pinchtab nav <url> --new-tab` 打开新标签页。没有 `pinchtab tab navigate` 或 `pinchtab tab click` 之类子命令。

## 顶级浏览器命令

下列页面涵盖简写路由及对应的命令行界面命令：

- [Health](./health.md)
- [Navigate](./navigate.md)
- [Snapshot](./snapshot.md)
- [Text](./text.md)
- [Click](./click.md)
- [Type](./type.md)
- [Fill](./fill.md)
- [Screenshot](./screenshot.md)
- [PDF](./pdf.md)
- [Eval](./eval.md)
- [Press](./press.md)
- [Hover](./hover.md)
- [Scroll](./scroll.md)
- [Select](./select.md)
- [Focus](./focus.md)
- [Find](./find.md)

## 在特定实例中打开标签页

```bash
curl -X POST http://localhost:9867/instances/inst_ea2e747f/tabs/open \
  -H "Content-Type: application/json" \
  -d '{"url":"https://pinchtab.com"}'
# Response
{
  "tabId": "8f9c7d4e1234567890abcdef12345678",
  "url": "https://pinchtab.com",
  "title": "PinchTab"
}
```

仍无专用的实例范围打开标签页命令行界面命令。命令行界面快捷方式为：

```bash
pinchtab instance navigate inst_ea2e747f https://pinchtab.com
```

该命令在一次 `tabs/open` 调用中就为已在该 URL 上的实例打开标签页。

## 列出标签页

### 活动桥接或简写上下文

```bash
curl http://localhost:9867/tabs
# Response (API always returns JSON)
{
  "tabs": [
    {
      "id": "8f9c7d4e1234567890abcdef12345678",
      "url": "https://pinchtab.com",
      "title": "PinchTab",
      "type": "page",
      "status": "active"
    }
  ]
}

# CLI Alternative (human-readable by default)
pinchtab tab
# Output: *8f9c7d4e...  https://pinchtab.com  PinchTab

pinchtab tab --json                    # Full JSON response
```

注意：

- `GET /tabs` 不是全机队清单
- 桥接模式或简写模式下列出活动浏览器上下文的标签页
- `pinchtab tab` 遵循该简写行为
- 当前标签页列在最前
- `about:blank`、`chrome://`、`chrome-extension://`、`devtools://`、`file://` 或服务器自身端口上的标签页不列出
- `status` 为 `active`，或带 `handoffReason` 和 `pausedAt` 的 `paused_handoff`；锁定标签页还带 `owner` 和 `lockedUntil`，标签页有 `browserContextId` 时也会出现

### 单个实例的标签页

```bash
curl http://localhost:9867/instances/inst_ea2e747f/tabs
```

### 所有运行实例的标签页

```bash
curl http://localhost:9867/instances/tabs
```

需要编排器范围视图时用 `GET /instances/tabs`。

两个实例路由都返回 `{"id","instanceId","url","title"}` 对象的裸 JSON 数组。结果按实例缓存；加 `?fresh=1` 重新拉取。

## 从命令行界面聚焦与关闭

```bash
pinchtab tab                           # list tabs
pinchtab tab 2                         # focus tab by 1-based index
pinchtab tab 8f9c7d4e1234...           # focus tab by tab ID
pinchtab nav https://pinchtab.com --new-tab  # open a new tab and navigate it
pinchtab tab close 8f9c7d4e1234...     # close tab
```

数字参数按相对 `GET /tabs` 的 1 基索引解析。非数字参数视为标签页 ID。

聚焦、导航或以其他方式访问受跟踪标签页会把它标记为当前标签页。未限定作用域的命令使用该当前标签页；若记录的当前标签页已过期，PinchTab 回退到最近使用的受跟踪标签页。

顶级导航在有当前标签页时使用它。显式想要另一个标签页时用 `pinchtab nav <url> --new-tab`。

## 操作现有标签页

用标签页范围的 HTTP 路由或带 `--tab` 的顶级命令行界面命令。

### 导航

```bash
curl -X POST http://localhost:9867/tabs/<tabId>/navigate \
  -H "Content-Type: application/json" \
  -d '{"url":"https://pinchtab.com"}'
# CLI Alternative
pinchtab nav https://pinchtab.com --tab <tabId>
```

### 快照

```bash
curl "http://localhost:9867/tabs/<tabId>/snapshot?filter=interactive&format=compact"
# CLI Alternative
pinchtab snap --tab <tabId> -i -c
```

### 文本

```bash
curl "http://localhost:9867/tabs/<tabId>/text?mode=raw"
# CLI Alternative
pinchtab text --tab <tabId> --raw
```

### 查找

```bash
curl -X POST http://localhost:9867/tabs/<tabId>/find \
  -H "Content-Type: application/json" \
  -d '{"query":"login button"}'
# CLI Alternative
pinchtab find --tab <tabId> "login button"
```

### 动作

```bash
curl -X POST http://localhost:9867/tabs/<tabId>/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"click","ref":"e5"}'
# CLI Alternative
pinchtab click --tab <tabId> e5
pinchtab fill --tab <tabId> '#email' 'ada@example.com'
pinchtab wait --tab <tabId> 'text:Done'
pinchtab network --tab <tabId> --limit 20
```

低层指针控制使用同一动作面：

```bash
curl -X POST http://localhost:9867/tabs/<tabId>/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"mouse-move","ref":"e5"}'

curl -X POST http://localhost:9867/tabs/<tabId>/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"mouse-down","button":"left"}'

curl -X POST http://localhost:9867/tabs/<tabId>/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"mouse-wheel","x":400,"y":320,"deltaY":240}'

# CLI Alternatives
pinchtab mouse move --tab <tabId> e5
pinchtab mouse down --tab <tabId> --button left
pinchtab mouse wheel --tab <tabId> 240 --dx 40
```

### 交接状态

人工交接是标签页范围的，可通过命令行界面或 API 使用。

```bash
pinchtab tab handoff <tabId> --reason captcha --timeout-ms 120000
pinchtab tab handoff-status <tabId>
pinchtab tab resume <tabId> --status completed
```

API 等价物：

标签页被标记为 `paused_handoff` 时，动作执行路由以 `409 tab_paused_handoff` 拒绝，直到该标签页被恢复或可选超时到期。

```bash
curl -X POST http://localhost:9867/tabs/<tabId>/handoff \
  -H "Content-Type: application/json" \
  -d '{"reason":"captcha","timeoutMs":120000}'

curl http://localhost:9867/tabs/<tabId>/handoff

curl -X POST http://localhost:9867/tabs/<tabId>/resume \
  -H "Content-Type: application/json" \
  -d '{"status":"completed","resolvedData":{"operator":"human"}}'
```

当自动化必须为 CAPTCHA、2FA 提示、登录批准或其他仅人工步骤暂停时使用。提供超时后，交接状态包含 `expiresAt` 和 `timeoutMs`。

### 截图

```bash
curl "http://localhost:9867/tabs/<tabId>/screenshot?raw=true" > out.jpg
# CLI Alternative
pinchtab screenshot --tab <tabId> -o out.jpg
```

### PDF

```bash
curl "http://localhost:9867/tabs/<tabId>/pdf?raw=true" > page.pdf
# CLI Alternative
pinchtab pdf --tab <tabId> -o page.pdf
```

## Cookies

```bash
curl http://localhost:9867/tabs/<tabId>/cookies
curl -X POST http://localhost:9867/tabs/<tabId>/cookies \
  -H "Content-Type: application/json" \
  -d '{"cookies":[{"name":"session","value":"abc"}]}'
```

`POST` 把 `url` 默认设为标签页当前页面，因此注入会话 cookie 无需 URL 查找。空 `value` 的 cookie 被设置（清空但不删除）；无 `name` 的 cookie 以 400 拒绝而非跳过。`DELETE /tabs/<tabId>/cookies` 仅寻址上是标签页范围的——它清空每个源的每个 cookie，与 `DELETE /cookies` 完全一样。

命令行界面中同样的操作：

```bash
pinchtab cookies get --tab <tabId>                  # read cookies, with values
pinchtab cookies get --tab <tabId> --name session   # one cookie
pinchtab cookies set session abc123 --tab <tabId>   # set on the tab's current URL
pinchtab cookies clear                              # every cookie, every origin
```

`cookies set` 接受 `--url`、`--domain`、`--path`、`--same-site`、`--secure` 和 `--http-only`；服务器报告 cookie 未设置时它以非零退出。没有按单个 cookie 删除的动词——`clear` 是浏览器范围的，命令行界面里没有东西能恢复它擦除的内容，因此重新设置你需要的，或用 `state load`。

读取、写入和清空 cookies 需要 `security.allowCookies=true`。

## 指标

```bash
curl http://localhost:9867/tabs/<tabId>/metrics
```

这里报告的是所属浏览器实例的聚合内存——整个进程树——而非孤立的单标签页读数：标签页 id 只是选定要询问的实例。同一实例的两个标签页返回相同数字。见 [Memory monitoring](../guides/memory-monitoring.md)。

## 锁定与解锁

标签页锁定仅 API 可用。

```bash
curl -X POST http://localhost:9867/tabs/<tabId>/lock \
  -H "Content-Type: application/json" \
  -d '{"owner":"my-agent","timeoutSec":60}'

curl -X POST http://localhost:9867/tabs/<tabId>/unlock \
  -H "Content-Type: application/json" \
  -d '{"owner":"my-agent"}'
```

`owner` 必填。`timeoutSec` 可选，默认 10 分钟。lock 应答 `{"locked":true,"owner":"...","expiresAt":"..."}`，unlock 应答 `{"unlocked":true}`，冲突的 owner 得到 `409`。

根路径形式 `POST /lock` 和 `POST /unlock` 也存在；它们把标签页放在主体里的 `tabId`（必填）。

## 重要限制

- 没有 `GET /tabs/{id}` 端点。`GET /tabs/{id}/state` 报告单个标签页的 `tabId`、`url`、`title`、`dialogPresent`/`dialog`、`load`（`readyState`、`navigationInProgress`、`networkIdle`、`state`）和 `actionability`。
- `GET /tabs` 和 `GET /instances/tabs` 用途不同，不可互换。
- 命令行界面中，标签页范围工作通过带 `--tab` 的顶级命令进行，而非 `pinchtab tab <subcommand>` 变体——`handoff`、`resume`、`handoff-status` 除外，它们既暴露为顶级命令，也暴露为 `pinchtab tab handoff|resume|handoff-status` 子命令。
