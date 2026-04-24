# 截图

将当前页面捕获为图像。默认为 **JPEG** 格式。

```bash
# 获取原始 PNG 字节
curl "http://localhost:9867/screenshot?format=png&raw=true" > page.png

# 捕获特定元素（选择器支持引用/CSS/XPath/文本）
curl "http://localhost:9867/screenshot?selector=%23checkout-button&raw=true" > button.jpg

# 以 CSS 1x 大小捕获元素（而不是设备像素）
curl "http://localhost:9867/screenshot?selector=%23checkout-button&css1x=true&raw=true" > button-1x.jpg

# 获取带 base64 JPEG 的 JSON（默认）
curl "http://localhost:9867/screenshot"

# 保存到服务器状态目录
curl "http://localhost:9867/screenshot?output=file"
```

## 响应 (JSON)

```json
{
  "path": "/path/to/state/screenshots/screenshot-20260308-120001.jpg",
  "size": 34567,
  "format": "jpeg",
  "timestamp": "20260308-120001"
}
```

## 有用的标志

### API 查询参数

- `format`: `jpeg`（默认）或 `png`。
- `quality`: JPEG 质量 `0-100`（默认：`80`）。PNG 忽略。
- `selector`: 捕获一个元素的统一选择器（例如 `e5`、`#id`、`xpath://...`、`text:Submit`）。
- `css1x`: `true` 以 CSS 像素大小（1x）输出选择器截图。当省略 `selector` 时忽略。
- `raw`: `true` 直接返回图像字节而不是 JSON。
- `output`: `file` 保存到状态目录。
- `tabId`: 目标特定标签页。

### CLI

- `-o <path>`: 保存到特定路径。
- `-q <0-100>`: 设置 JPEG 质量。
- `-s <selector>`: 捕获特定元素。
- `--css-1x`: 与 `-s/--selector` 一起使用，以 CSS 1x 大小导出。
- `--tab <id>`: 目标特定标签页。

## 相关页面

- [快照](./snapshot.md)
- [PDF](./pdf.md)