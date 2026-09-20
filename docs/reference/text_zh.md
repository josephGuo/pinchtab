# 文本（Text）

从当前页面或特定元素提取文本。

默认情况下，PinchTab 对当前文档运行 Readability 风格提取。需要 `document.body.innerText` 时用 full/raw 模式。

## 元素选择

用选择器从特定元素提取文本：

```bash
# Positional selector argument
pinchtab text "#article-body"
pinchtab text "text:Welcome"

# Or use --selector flag
pinchtab text --selector "#article-body"
pinchtab text -s "xpath://div[@class='content']"
```

支持的选择器类型：ref（`e5`）、CSS（`#id`）、XPath（`xpath://...`）、text（`text:...`）。

## Frame 范围

`/text` 是 frame 感知的：

- `--frame <id>` 或 `frameId=<id>` 针对特定 iframe 做一次性读取
- 否则 `/text` 从 [`/frame`](./frame.md) 继承标签页当前 frame 范围
- 未选 frame 时，`/text` 从顶级文档读取

## 输出格式

默认输出人类可读文本。用 `--json` 取结构化输出：

```bash
pinchtab text                           # Plain text output
pinchtab text --json                    # JSON envelope: url, title, text, truncated, extraction, textLength, rawLength, ...
```

元素读取（`selector`/`ref`）只返回 `{url, title, text}`——该元素的 `innerText`——并忽略 `mode`。

页面打开模态对话框时，整页读取返回最顶层对话框的文本（以 `raw` 报告），元素 selector 在其内部解析。读取期间最顶层对话框变化两次，请求以 `409` 失败；挂起的 JavaScript 对话框（alert、confirm、prompt）以 `409 dialog_blocked` 阻断。

## 提取模式

Readability 是文章启发式，因此在落地页、仪表板或文档索引上它可能只返回一个区块而非整页。`/text` 把其输出与 `document.body.innerText` 比较，覆盖率过低时改回返回原始文本。JSON 信封总是说明是哪个提取器产出的文本：

| 字段 | 说明 |
|-------|-------------|
| `extraction` | `readability`、`raw`（显式请求）、`markdown`（页面转 Markdown）、`markdown_fallback`（转换器无产出，返回原始文本）或 `readability_fallback`（Readability 坍缩，返回原始文本） |
| `textLength` | 返回文本的长度 |
| `rawLength` | `document.body.innerText` 的长度，因此覆盖率可计算 |

`format=text` 时主体保持裸文本，模式在 `X-PT-Text-Extraction` 响应头中报告（Markdown 模式为 `text/markdown` content type）。`mode=raw` 不做比较。`truncated` 含义不变——被 `maxChars` 截断的文本——不受回退影响。发生回退时，命令行界面在 stderr 打印一行说明；stdout 保持纯文本。

`--markdown`（`mode=markdown`）把渲染后的页面转成 Markdown，保留标题、行内链接和表格——对想保存或重读的文章形页面最佳。它与 `--full`/`--raw` 互斥。配 `--output <file>`（`-o`）把 Markdown 写入磁盘；此时命令只打印一行确认，使长页面不会淹没代理上下文。

## 示例

```bash
# Default Readability extraction
pinchtab text

# Full page text (document.body.innerText)
pinchtab text --full
pinchtab text --raw                     # Alias of --full

# Markdown (preserves headings, links, tables)
pinchtab text --markdown
pinchtab text --markdown --output page.md   # writes the file, prints a one-line confirmation

# Extract text from specific element
pinchtab text "#main-content"
pinchtab text --selector ".article-body"

# One-shot iframe read by frame id
pinchtab text --frame FRAME123

# API equivalent
curl "http://localhost:9867/text?mode=raw"
curl "http://localhost:9867/text?mode=markdown"
curl "http://localhost:9867/text?selector=%23article-body"
curl "http://localhost:9867/text?frameId=FRAME123&format=text"
```

## Flags

| Flag | 说明 |
|------|-------------|
| `--selector`, `-s` | 元素选择器（ref/CSS/XPath/text） |
| `--frame` | 按 frameId 从特定 iframe 提取 |
| `--full` | 整页 innerText 而非 Readability |
| `--raw` | --full 的别名 |
| `--markdown` | 把页面转 Markdown（标题、链接、表格）；与 `--full`/`--raw` 互斥 |
| `--output`, `-o` | 把提取文本写入此文件并打印一行确认 |
| `--json` | 输出 JSON 而非纯文本 |
| `--tab` | 目标特定标签页 |

## API 参数

| 参数 | 说明 |
|-----------|-------------|
| `selector` | 用于文本提取的元素选择器 |
| `ref` | 快照 ref（如 `e5`） |
| `frameId` | 目标 iframe ID |
| `tabId` | 目标标签页（默认当前标签页） |
| `mode` | `raw` 或其别名 `full` 取 innerText，`markdown` 取 Markdown，省略则 Readability；其他值为 400 |
| `maxChars` | 截断输出到此字符数（正整数；Markdown 按行边界截断） |
| `format` | `text` 或 `plain` 取裸文本响应；默认 JSON 信封 |

`GET /tabs/{id}/text` 是同一处理器，标签页放在路径中。

类文章页面用默认模式。UI 密集页面如仪表板、SERP、网格、价格表或 Readability 可能裁掉的短日志窗格，用 `--full` / `mode=raw`。

## 相关页面

- [Snapshot](./snapshot.md)
- [Frame](./frame.md)
- [PDF](./pdf.md)
