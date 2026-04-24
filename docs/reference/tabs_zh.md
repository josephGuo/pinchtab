# 标签页

标签页是浏览、提取、交互和诊断的主要执行表面。

一旦你已有标签页 ID，就可以使用标签页范围的 HTTP 路由。在 CLI 中，使用带有 `--tab <id>` 的正常顶级浏览器命令。

`pinchtab tab` 本身仅用于：

- 列出标签页
- 聚焦标签页
- 打开新标签页
- 关闭标签页

没有 `pinchtab tab navigate` 或 `pinchtab tab click` 这样的子命令。

## 顶级浏览器命令

这些页面涵盖了简写路由和匹配的 CLI 命令：

- [健康](./health.md)
- [导航](./navigate.md)
- [快照](./snapshot.md)
- [文本](./text.md)
- [点击](./click.md)
- [输入](./type.md)
- [填充](./fill.md)
- [截图](./screenshot.md)
- [PDF](./pdf.md)
- [评估](./eval.md)
- [按键](./press.md)
- [悬停](./hover.md)
- [滚动](./scroll.md)
- [选择](./select.md)
- [聚焦](./focus.md)
- [查找](./find.md)

## 在特定实例中打开标签页

```bash
curl -X POST http://localhost:9867/instances/inst_ea2e747f/tabs/open \
  -H "Content-Type: application/json" \
  -d '{"url":"https://pinchtab.com"}'
# 响应
{
  "tabId": "8f9c7d4e1234567890abcdef12345678",
  "url": "https://pinchtab.com",
  "title": "PinchTab"
}
```

仍然没有专用的实例范围标签页打开 CLI 命令。CLI 快捷方式是：

```bash
pinchtab instance navigate inst_ea2e747f https://pinchtab.com
```

该命令为实例打开标签页，然后导航它。

## 列出标签页

### 活动桥接或简写上下文

```bash
curl http://localhost:9867/tabs
# 响应（API 始终返回 JSON）
{
  "tabs": [
    {
      "id": "8f9c7d4e1234567890abcdef12345678",
      "url": "https://pinchtab.com",
      "title": "PinchTab",
      "type": "page"
    }
  ]
}

# CLI 替代方案（默认人类可读）
pinchtab tab
# 输出: *8f9c7d4e...  https://pinchtab.com  PinchTab

pinchtab tab --json                    # 完整 JSON 响应
```

注意：

- `GET /tabs` 不是全舰队库存
- 在桥接模式或简写模式下，它列出活动浏览器上下文中的标签页
- `pinchtab tab` 遵循该简写行为

### 一个实例的标签页

```bash
curl http://localhost:9867/instances/inst_ea2e747f/tabs
```

### 所有运行实例的标签页

```bash
curl http://localhost:9867/instances/tabs
```

当你需要编排器范围的视图时，使用 `GET /instances/tabs`。

## 从 CLI 聚焦、创建和关闭

```bash
pinchtab tab                           # 列出标签页
pinchtab tab 2                         # 按 1 基索引聚焦标签页
pinchtab tab 8f9c7d4e1234...           # 按标签页 ID 聚焦标签页
pinchtab tab new                       # 打开空白标签页
pinchtab tab new https://pinchtab.com   # 打开并导航
pinchtab tab close 8f9c7d4e1234...     # 关闭标签页
```

数字参数被解析为相对于 `GET /tabs` 的 1 基索引。非数字参数被视为标签页 ID。

## 操作现有标签页

使用标签页范围的 HTTP 路由或带有 `--tab` 的顶级 CLI 命令。

### 导航

```bash
curl -X POST http://localhost:9867/tabs/<tabId>/navigate \
  -H "Content-Type: application/json" \
  -d '{"url":"https://pinchtab.com"}'
# CLI 替代方案
pinchtab nav https://pinchtab.com --tab <tabId>
```

### 快照

```bash
curl "http://localhost:9867/tabs/<tabId>/snapshot?interactive=true&compact=true"
# CLI 替代方案
pinchtab snap --tab <tabId> -i -c
```

### 文本

```bash
curl "http://localhost:9867/tabs/<tabId>/text?mode=raw"
# CLI 替代方案
pinchtab text --tab <tabId> --raw
```

### 查找

```bash
curl -X POST http://localhost:9867/tabs/<tabId>/find \
  -H "Content-Type: application/json" \
  -d '{"query":"login button"}'
# CLI 替代方案
pinchtab find --tab <tabId> "login button"
```

### 操作

```bash
curl -X POST http://localhost:9867/tabs/<tabId>/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"click","ref":"e5"}'
# CLI 替代方案
pinchtab click --tab <tabId> e5
pinchtab fill --tab <tabId> '#email' 'ada@example.com'
pinchtab wait --tab <tabId> 'text:Done'
pinchtab network --tab <tabId> --limit 20
```

低级指针控制使用相同的操作表面：

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

# CLI 替代方案
pinchtab mouse move --tab <tabId> e5
pinchtab mouse down --tab <tabId> --button left
pinchtab mouse wheel --tab <tabId> 240 --dx 40
```

### 切换状态

人工切换是标签页范围的，可通过 CLI 或 API 使用。

```bash
pinchtab tab handoff <tabId> --reason captcha --timeout-ms 120000
pinchtab tab handoff-status <tabId>
pinchtab tab resume <tabId> --status completed
```

API 等价物：

当标签页被标记为 `paused_handoff` 时，操作执行路由会拒绝并返回 `409 tab_paused_handoff`，直到标签页被恢复或可选的超时过期。

```bash
curl -X POST http://localhost:9867/tabs/<tabId>/handoff \
  -H "Content-Type: application/json" \
  -d '{"reason":"captcha","timeoutMs":120000}'

curl http://localhost:9867/tabs/<tabId>/handoff

curl -X POST http://localhost:9867/tabs/<tabId>/resume \
  -H "Content-Type: application/json" \
  -d '{"status":"completed","resolvedData":{"operator":"human"}}'
```

当自动化必须为 CAPTCHA、2FA 提示、登录批准或其他仅人工步骤暂停时使用此功能。
当提供超时时，切换状态包括 `expiresAt` 和 `timeoutMs`。

### 截图

```bash
curl "http://localhost:9867/tabs/<tabId>/screenshot?raw=true" > out.jpg
# CLI 替代方案
pinchtab screenshot --tab <tabId> -o out.jpg
```

### PDF

```bash
curl "http://localhost:9867/tabs/<tabId>/pdf?raw=true" > page.pdf
# CLI 替代方案
pinchtab pdf --tab <tabId> -o page.pdf
```

##  Cookies

```bash
curl http://localhost:9867/tabs/<tabId>/cookies
curl -X POST http://localhost:9867/tabs/<tabId>/cookies \
  -H "Content-Type: application/json" \
  -d '{"cookies":[{"name":"session","value":"abc"}]}'
```

目前没有专用的顶级 cookies CLI 命令。

## 指标

```bash
curl http://localhost:9867/tabs/<tabId>/metrics
```

这通过桥接报告标签页的内存指标，而不是完整的每个标签页性能配置文件。

## 锁定和解锁

标签页锁定仅通过 API 可用。

```bash
curl -X POST http://localhost:9867/tabs/<tabId>/lock \
  -H "Content-Type: application/json" \
  -d '{"owner":"my-agent","ttl":60}'

curl -X POST http://localhost:9867/tabs/<tabId>/unlock \
  -H "Content-Type: application/json" \
  -d '{"owner":"my-agent"}'
```

在 `POST /lock` 和 `POST /unlock` 也有活动标签页形式。

## 重要限制

- 没有用于获取单个标签页元数据的 `GET /tabs/{id}` 端点。
- `GET /tabs` 和 `GET /instances/tabs` 服务于不同的目的，不可互换。
- 在 CLI 中，标签页范围的工作通过带有 `--tab` 的顶级命令进行，而不是通过 `pinchtab tab <subcommand>` 变体。
- 目前没有专用的 CLI `handoff` 或 `resume` 命令。