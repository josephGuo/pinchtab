# 文本

从当前页面或特定元素中提取文本。

默认情况下，PinchTab 对当前文档运行 Readability 风格的提取。当你想要 `document.body.innerText` 时，使用 full/raw 模式。

## 元素选择

使用选择器从特定元素提取文本：

```bash
# 位置选择器参数
pinchtab text "#article-body"
pinchtab text "text:Welcome"

# 或使用 --selector 标志
pinchtab text --selector "#article-body"
pinchtab text -s "xpath://div[@class='content']"
```

支持的选择器类型：引用 (`e5`)、CSS (`#id`)、XPath (`xpath://...`)、文本 (`text:...`)。

## 框架范围

`/text` 是框架感知的：

- `--frame <id>` 或 `frameId=<id>` 针对特定 iframe 进行一次性读取
- 否则，`/text` 从 [`/frame`](./frame.md) 继承标签页的当前框架范围
- 如果未选择框架，`/text` 从顶级文档读取

## 输出格式

默认输出是人类可读的文本。使用 `--json` 获取结构化输出：

```bash
pinchtab text                           # 纯文本输出
pinchtab text --json                    # JSON: {"url":"...","title":"...","text":"..."}
```

## 示例

```bash
# 默认 Readability 提取
pinchtab text

# 完整页面文本 (document.body.innerText)
pinchtab text --full
pinchtab text --raw                     # --full 的别名

# 从特定元素提取文本
pinchtab text "#main-content"
pinchtab text --selector ".article-body"

# 通过 frame id 一次性 iframe 读取
pinchtab text --frame FRAME123

# API 等价物
curl "http://localhost:9867/text?mode=raw"
curl "http://localhost:9867/text?selector=%23article-body"
curl "http://localhost:9867/text?frameId=FRAME123&format=text"
```

## 标志

| 标志 | 描述 |
|------|-------------|
| `--selector`, `-s` | 元素选择器（引用/CSS/XPath/文本） |
| `--frame` | 从特定 iframe 提取（通过 frameId） |
| `--full` | 完整页面 innerText 而不是 Readability |
| `--raw` | --full 的别名 |
| `--json` | 输出 JSON 而不是纯文本 |
| `--tab` | 目标特定标签页 |

## API 参数

| 参数 | 描述 |
|-----------|-------------|
| `selector` | 用于文本提取的元素选择器 |
| `ref` | 快照引用（例如 `e5`） |
| `frameId` | 目标 iframe ID |
| `mode` | `raw` 表示 innerText，默认为 Readability |
| `maxChars` | 截断输出 |
| `format` | `text` 表示纯文本响应 |

对于类文章页面，使用默认模式。对于 UI 繁重的页面，如仪表板、SERP、网格、价格表或 Readability 可能会修剪掉的短日志窗格，使用 `--full` / `mode=raw`。

## 相关页面

- [快照](./snapshot.md)
- [框架](./frame.md)
- [PDF](./pdf.md)