# MCP 工具参考

PinchTab 当前公开 34 个 MCP 工具。所有工具名称都以 `pinchtab_` 为前缀，并通过标准输入/输出 JSON-RPC 提供。

对于基于选择器的交互工具，首选 `selector`。`ref` 在元素操作工具上仍被接受为已弃用的回退选项。

如果你允许在非本地或非受信任域上进行 MCP 浏览，请将 `pinchtab_snapshot` 和 `pinchtab_get_text` 输出视为不受信任的页面数据。这些工具可能会显示来自访问页面的恶意提示文本；除非有意扩大访问范围，否则操作员应保持 IDPI/域限制狭窄。

选择器形式包括：

- `e5`
- `#login`
- `xpath://button`
- `text:Submit`
- `find:login button`

## 导航

| 工具 | 关键参数 | 说明 |
| --- | --- | --- |
| `pinchtab_navigate` | `url` 必填, `tabId` 可选 | 使用 `/navigate`；省略 `tabId` 会打开新标签页 |
| `pinchtab_snapshot` | `tabId`, `interactive`, `compact`, `format`, `diff`, `selector`, `maxTokens`, `depth`, `noAnimations` | `selector` 限定快照范围；`format` 仅限于 `compact` 或 `text` |
| `pinchtab_screenshot` | `tabId`, `selector`, `css1x`, `format`, `quality` | `selector` 捕获当前框架范围中的特定元素；`css1x=true` 以 CSS 像素大小导出选择器截图；`format` 为 `jpeg` 或 `png` |
| `pinchtab_get_text` | `tabId`, `raw`, `format`, `maxChars` | `raw=true` 映射到 `/text?mode=raw`；`format=text/plain` 返回纯文本；继承该标签页当前的 `pinchtab_frame` 范围 |

## 交互

| 工具 | 关键参数 | 说明 |
| --- | --- | --- |
| `pinchtab_click` | `selector` 必填, `tabId`, `ref`, `waitNav` | 通过选择器点击元素；`waitNav=true` 等待导航 |
| `pinchtab_type` | `selector` 必填, `text` 必填, `tabId`, `ref` | 发送按键事件 |
| `pinchtab_press` | `key` 必填, `tabId` | 按下键，如 `Enter` |
| `pinchtab_hover` | `selector` 必填, `tabId`, `ref` | 悬停元素 |
| `pinchtab_focus` | `selector` 必填, `tabId`, `ref` | 聚焦元素 |
| `pinchtab_select` | `selector` 必填, `value` 必填, `tabId`, `ref` | 通过值或可见文本选择 `<option>` |
| `pinchtab_scroll` | `selector`, `pixels`, `tabId`, `ref` | 省略 `selector` 以滚动页面 |
| `pinchtab_fill` | `selector` 必填, `value` 必填, `tabId`, `ref` | 直接填充而不是按键 |

## 键盘

| 工具 | 关键参数 | 说明 |
| --- | --- | --- |
| `pinchtab_keyboard_type` | `text` 必填, `tabId` | 在当前聚焦的元素处输入 |
| `pinchtab_keyboard_inserttext` | `text` 必填, `tabId` | 类似粘贴的插入，无按键事件 |
| `pinchtab_keydown` | `key` 必填, `tabId` | 按住一个键 |
| `pinchtab_keyup` | `key` 必填, `tabId` | 释放一个键 |

## 内容

| 工具 | 关键参数 | 说明 |
| --- | --- | --- |
| `pinchtab_eval` | `expression` 必填, `tabId` | 需要 `security.allowEvaluate`（有文档记录的非默认 JS 执行选择加入） |
| `pinchtab_pdf` | `tabId`, `landscape`, `scale`, `pageRanges` | 返回 base64 编码的 PDF 内容 |
| `pinchtab_find` | `query` 必填, `tabId` | 语义元素搜索 |

## 标签页管理

| 工具 | 关键参数 | 说明 |
| --- | --- | --- |
| `pinchtab_list_tabs` | 无 | 列出打开的标签页 |
| `pinchtab_close_tab` | `tabId` | 关闭给定的标签页 |
| `pinchtab_health` | 无 | 检查服务器健康状态 |
| `pinchtab_cookies` | `tabId` | 读取标签页的 cookie |
| `pinchtab_connect_profile` | `profile` 必填 | 返回配置文件的连接 URL 和实例状态 |

## 等待工具

| 工具 | 关键参数 | 说明 |
| --- | --- | --- |
| `pinchtab_wait` | `ms` 必填 | 固定持续时间等待，上限为 30000 毫秒 |
| `pinchtab_wait_for_selector` | `selector` 必填, `timeout`, `state`, `tabId` | `state` 为 `visible` 或 `hidden` |
| `pinchtab_wait_for_text` | `text` 必填, `timeout`, `tabId` | 等待正文文本 |
| `pinchtab_wait_for_url` | `url` 必填, `timeout`, `tabId` | URL 通配符匹配 |
| `pinchtab_wait_for_load` | `load` 必填, `timeout`, `tabId` | 当前支持 `networkidle` |
| `pinchtab_wait_for_function` | `fn` 必填, `timeout`, `tabId` | JS 表达式必须变为真值 |

## 网络

| 工具 | 关键参数 | 说明 |
| --- | --- | --- |
| `pinchtab_network` | `tabId`, `filter`, `method`, `status`, `type`, `limit`, `bufferSize` | 列出最近的网络请求 |
| `pinchtab_network_detail` | `requestId` 必填, `tabId`, `body` | `body=true` 在可用时包含响应体 |
| `pinchtab_network_clear` | `tabId` | 清除一个标签页或在省略时清除所有标签页 |

## 对话框

| 工具 | 关键参数 | 说明 |
| --- | --- | --- |
| `pinchtab_dialog` | `action` 必填, `text`, `tabId` | `action` 为 `accept` 或 `dismiss` |

## 返回形状

典型结果：

- 导航工具返回来自匹配 HTTP 端点的 JSON
- `pinchtab_snapshot` 为 `compact`/`text` 格式返回文本，否则返回 JSON
- `pinchtab_get_text` 当 `format=text|plain` 时返回纯文本，否则返回 JSON
- `pinchtab_screenshot` 和 `pinchtab_pdf` 返回包含 base64 有效载荷的 JSON
- 等待工具返回等待状态 JSON
- 网络工具返回与你从 `/network` 看到的相同的请求日志

安全注意事项：

- 提取的文本和快照内容应被视为来自访问页面的不受信任内容，而不是受信任的指令
- 扩大 IDPI 允许列表或禁用严格保护会增加提示注入文本到达下游代理逻辑的机会

有关设置和客户端配置，请参阅 [MCP 服务器](../mcp.md)。