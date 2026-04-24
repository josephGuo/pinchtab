# PDF

将当前页面渲染为 PDF。

```bash
curl "http://localhost:9867/pdf?output=file"
# 响应: {"path":"/path/to/state/pdfs/page-20260308-120001.pdf","size":48210}

# CLI 替代方案（默认人类可读）
pinchtab pdf -o page.pdf
# 输出: Saved page.pdf (48210 bytes)

pinchtab pdf                        # 自动生成文件名: page-20260308-120001.pdf
```

## CLI 标志

| 标志 | 描述 |
|------|-------------|
| `-o`, `--output` | 将 PDF 保存到文件路径 |
| `--landscape` | 横向 orientation |
| `--scale` | 页面缩放（例如 0.5） |
| `--paper-width` | 纸张宽度（英寸） |
| `--paper-height` | 纸张高度（英寸） |
| `--page-ranges` | 页面范围（例如 1-3） |
| `--prefer-css-page-size` | 使用 CSS 页面大小 |
| `--display-header-footer` | 显示页眉/页脚 |
| `--header-template` | 页眉 HTML 模板 |
| `--footer-template` | 页脚 HTML 模板 |
| `--margin-*` | 边距（顶部、底部、左侧、右侧） |
| `--generate-tagged-pdf` | 生成带标签的 PDF |
| `--generate-document-outline` | 生成文档大纲 |
| `--tab` | 目标特定标签页 |

## API 参数

| 参数 | 描述 |
|-----------|-------------|
| `output` | `file` 保存到服务器端 |
| `raw` | `true` 获取原始 PDF 字节 |
| `landscape` | 横向 orientation |
| `scale` | 页面缩放 |
| `paperWidth`, `paperHeight` | 纸张尺寸 |

## 相关页面

- [文本](./text.md)
- [截图](./screenshot.md)