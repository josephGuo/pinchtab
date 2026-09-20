# PDF

将当前页面渲染为 PDF。

```bash
curl "http://localhost:9867/pdf?output=file"
# Response: {"path":"/path/to/state/pdfs/page-20260308-120001.pdf","size":48210}

# CLI Alternative (human-readable by default)
pinchtab pdf -o page.pdf
# Output: Saved page.pdf (48210 bytes)

pinchtab pdf                        # Auto-generates filename: page-20260308-120001.pdf
```

## 命令行界面 Flags

| Flag | 说明 |
|------|-------------|
| `-o`, `--output` | 将 PDF 保存到文件路径 |
| `--landscape` | 横向 |
| `--scale` | 页面缩放（如 0.5） |
| `--paper-width` | 纸张宽度（英寸） |
| `--paper-height` | 纸张高度（英寸） |
| `--page-ranges` | 页面范围（如 1-3） |
| `--prefer-css-page-size` | 使用 CSS 页面大小 |
| `--display-header-footer` | 显示页眉/页脚 |
| `--header-template` | 页眉 HTML 模板 |
| `--footer-template` | 页脚 HTML 模板 |
| `--margin-*` | 边距（上、下、左、右） |
| `--generate-tagged-pdf` | 生成 tagged PDF |
| `--generate-document-outline` | 生成文档大纲 |
| `--file-output` | 服务器端保存（`output=file`）而非流式字节 |
| `--path` | 服务器端输出路径（在状态目录下） |
| `--tab` | 目标特定标签页 |

## API 参数

`GET /pdf` 和 `POST /pdf` 读取查询参数；`POST /tabs/{id}/pdf` 也接受它们作为 JSON 主体字段。

| 参数 | 说明 |
|-----------|-------------|
| `output` | `file` 保存到服务器端；响应为 `{path, size}` |
| `path` | `output=file` 时的服务器端路径；必须留在状态目录内（否则 `400`） |
| `raw` | `true` 取原始 PDF 字节；否则响应为 `{"format":"pdf","base64":"..."}` |
| `landscape` | 横向 |
| `scale` | 页面缩放（默认 `1`） |
| `paperWidth`, `paperHeight` | 纸张尺寸，英寸（默认 `8.5` x `11`） |
| `marginTop`, `marginBottom`, `marginLeft`, `marginRight` | 边距，英寸（默认 `0.4`） |
| `pageRanges` | 要导出的页面，如 `1-3,5` |
| `preferCSSPageSize`, `displayHeaderFooter`, `generateTaggedPDF`, `generateDocumentOutline` | 布尔值 |
| `headerTemplate`, `footerTemplate` | HTML 模板；含 `<script`、`javascript:` 或 `on*=` 处理器的模板以 `400` 拒绝 |
| `tabId` | 目标特定标签页 |

启用 IDPI 内容扫描时，触发扫描器的页面返回 `403`。

MCP：`pinchtab_pdf` 接受 `tabId`、`landscape`、`scale`、`pageRanges`，返回 base64 PDF。

## 相关页面

- [Text](./text.md)
- [Screenshot](./screenshot.md)
