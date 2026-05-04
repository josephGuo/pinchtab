# 命令行界面命令参考 — PinchTab

> **快速提示：** 使用 `pinchtab help` 或 `pinchtab <command> --help` 获取完整标志列表。

---

## 控制平面

### `pinchtab server`
启动 PinchTab 服务器（默认端口 9867）。

```bash
pinchtab server
pinchtab server -y              # 关闭保护（启用 evaluate、宏、下载）
pinchtab server -H              # 可见浏览器用于调试
pinchtab server -yH             # 两者组合
pinchtab server -e ./ext        # 加载浏览器扩展
```

| 标志 | 缩写 | 描述 |
|------|-------|-------------|
| `--yolo` | `-y` | 应用关闭保护预设（启用 evaluate、宏、下载） |
| `--headed` | `-H` | 以有头（可见）模式启动浏览器 |
| `--extension <path>` | `-e` | 加载浏览器扩展（可重复） |

> **注意：** 仅在你需要视觉反馈（调试、观看自动化）时使用 `--headed`。无头模式资源效率更高。

### `pinchtab daemon`
管理用户级后台服务。

```bash
pinchtab daemon
pinchtab daemon install
pinchtab daemon start
pinchtab daemon stop
pinchtab daemon restart
```

### `pinchtab health`
检查服务器是否运行且健康。

---

## 浏览器命令

### `pinchtab nav <url>`
将当前跟踪的标签页导航到 URL，或者在没有当前标签页时创建一个。这是当默认本地服务器未运行时自动启动它的浏览器命令。如果没有会话，`nav` 使用共享的当前标签页——首先设置 `PINCHTAB_SESSION` 以获取隔离的标签页。

```bash
pinchtab nav https://pinchtab.com
pinchtab nav https://pinchtab.com --new-tab
pinchtab nav https://pinchtab.com --snap
pinchtab nav https://pinchtab.com --block-images
pinchtab nav https://pinchtab.com --tab <tabId>
```

| 标志 | 描述 |
|------|-------------|
| `--new-tab` | 显式强制新标签页 |
| `--tab <id>` | 重用特定标签页 |
| `--snap` | 导航并打印交互式紧凑快照 |
| `--block-images` | 阻止图像加载（更快、更少令牌） |
| `--block-ads` | 为此次导航阻止广告 |
| `--print-tab-id` | 仅打印标签页 ID |

### `pinchtab tab`（不是 `tabs`）
管理浏览器标签页。

```bash
pinchtab tab                 # 列出所有打开的标签页
pinchtab tab <tabId>         # 按 ID 或 1 基索引聚焦标签页
pinchtab nav <url> --new-tab # 打开新标签页并导航
pinchtab tab close <tabId>   # 关闭特定标签页
```

未限定命令按调用者身份解析当前标签页。经过会话身份验证的调用者使用该会话范围的当前标签页；`--agent-id` / `PINCHTAB_AGENT_ID` 调用者在没有会话时使用该代理 ID 范围的当前标签页；匿名命令行界面调用使用共享的本地当前标签页状态文件。

---

## 交互命令

### `pinchtab click <ref>`
通过其可访问性引用（来自 `snap`）点击元素。

```bash
pinchtab click e5
pinchtab click e5 --snap-diff    # 点击 + 仅返回更改的元素
pinchtab click e5 --snap         # 点击 + 返回完整快照
pinchtab click e5 --tab <tabId>
```

### `pinchtab type <ref> <text>`
在输入元素中输入文本。

```bash
pinchtab type e12 "hello world"
```

### `pinchtab fill <ref> <value>`
使用 JS 事件分发填写表单字段。对于 React/Vue/Angular 表单，优先于 `type`。

```bash
pinchtab fill e12 "hello world"
pinchtab fill e12 "hello" --snap-diff    # 填写 + 仅返回更改的元素
```

### `pinchtab press <key>`
按下命名的键盘键。

```bash
pinchtab press Enter
pinchtab press Tab
pinchtab press Escape
```

### `pinchtab hover <ref>`
悬停在元素上以触发工具提示或悬停样式。

### `pinchtab mouse move|down|up|wheel [ref]`
低级指针控制，用于 DOM 原生点击或悬停行为不足的情况。

```bash
pinchtab mouse move e5
pinchtab mouse move 120 220
pinchtab mouse down e5 --button left
pinchtab mouse down --button left
pinchtab mouse up e5 --button left
pinchtab mouse up --button left
pinchtab mouse wheel 240 --dx 40
pinchtab mouse move --x 400 --y 320
pinchtab drag e5 400,320
```

将这些用于拖动手柄、画布控件、精确悬停编排或需要精确指针序列的站点。

### `pinchtab scroll [ref]`
滚动页面或特定元素。

```bash
pinchtab scroll            # 向下滚动页面 300px
pinchtab scroll --pixels -300   # 向上滚动
pinchtab scroll e20 --pixels 500
```

### `pinchtab select <ref> <value>`
从 `<select>` 下拉菜单中选择选项。

```bash
pinchtab select e8 "option-value"
pinchtab select e8 "value" --snap-diff    # 选择 + 仅返回更改的元素
```

---

## 输出命令

### `pinchtab snap`（快照）
获取当前页面的可访问性树。**理解页面状态的主要工具。**

```bash
pinchtab snap                   # 紧凑交互式快照（默认）
pinchtab snap "#main"           # 限定的位置选择器
pinchtab snap -s main           # 用 --selector 限定
pinchtab snap --full            # 完整 JSON 树
pinchtab snap -d                # 差异：仅自上次 snap 以来的更改（优先在操作上使用 --snap-diff）
pinchtab snap --max-tokens 2000 # 令牌预算上限
```

> ⚠️ **怪癖：** 使用 `snap`，而不是 `snapshot`。别名 `snap` 是预期的短形式。

### `pinchtab screenshot`
捕获当前页面的截图。

```bash
pinchtab screenshot
pinchtab screenshot --quality 80   # JPEG 80% 质量
```

> ⚠️ **怪癖：** 使用 `screenshot`（全词），而不是 `ss` 或 `shot`。

### `pinchtab text`
从页面提取可读文本。

```bash
pinchtab text
pinchtab text --raw    # 无格式化清理
pinchtab text "#main"  # 一个元素的文本
```

### `pinchtab find <query>`
通过文本内容或 CSS 选择器查找元素。

```bash
pinchtab find "Submit"
pinchtab find ".btn-primary"
```

### `pinchtab eval <expression>`
在浏览器上下文中运行 JavaScript。

```bash
pinchtab eval "document.title"
pinchtab eval "document.querySelectorAll('a').length"
```

> 需要配置中 `security.allowEvaluate: true`。默认返回 403。

### `pinchtab network`
检查当前标签页捕获的网络请求。

```bash
pinchtab network
pinchtab network --limit 20
pinchtab network --filter api
pinchtab network <requestId> --body
```

---

## 舰队/多配置文件命令

### `pinchtab profiles`
列出可用配置文件。

```bash
pinchtab profiles
pinchtab instance start --profile work
```

### `pinchtab instances`
列出跨配置文件运行的 PinchTab 实例。

---

## 已知怪癖摘要

| 错误 | 正确 | 说明 |
|------|------|------|
| `pinchtab ss` | `pinchtab screenshot` | 没有 `ss` 别名 |
| `pinchtab snapshot` | `pinchtab snap` | 使用短形式 |