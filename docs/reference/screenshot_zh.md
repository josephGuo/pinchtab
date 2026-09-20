# 截图（Screenshot）

将当前页面捕获为图像。API 默认为 **JPEG**。命令行界面在 `-o` 以 `.png` 结尾时用 PNG，否则用 JPEG；显式的 `--format` 始终优先。

```bash
# Get raw PNG bytes
curl "http://localhost:9867/screenshot?format=png&raw=true" > page.png

# Capture a specific element (selector supports ref/CSS/XPath/text)
curl "http://localhost:9867/screenshot?selector=%23checkout-button&raw=true" > button.jpg

# Half-size output (quarter the pixels)
curl "http://localhost:9867/screenshot?scale=0.5&raw=true" > page-half.jpg

# Capture the entire scrollable document, not just the visible viewport
curl "http://localhost:9867/screenshot?beyondViewport=true&raw=true" > fullpage.jpg

# Get JSON with base64 JPEG (default)
curl "http://localhost:9867/screenshot"

# Save to server state directory
curl "http://localhost:9867/screenshot?output=file"
```

## 响应 (JSON)

默认：`{"format":"jpeg","base64":"..."}`。带 `annotate=true` 时，主体还携带 `annotations: [{ref, role, name, tag, box:{x,y,w,h}}]`。带 `output=file` 时：

```json
{
  "path": "/path/to/state/screenshots/screenshot-20260308-120001.jpg",
  "size": 34567,
  "format": "jpeg",
  "timestamp": "20260308-120001"
}
```

未匹配到任何内容的 `selector` 返回与其他读取端点相同的 selector 失败 404。

## 常用 flags

### API 查询参数

- `format`：`jpeg`（默认）或 `png`。
- `quality`：JPEG 质量 `0-100`（默认 `80`）。PNG 忽略。
- `selector`：捕获单个元素的统一选择器（如 `e5`、`#id`、`xpath://...`、`text:Submit`）。
- `scale`：缩放输出位图。默认 `1`，钳制在 `0.05`–`4`。`0.5` 使每个轴减半（像素数为四分之一）。非法的 `quality`/`scale` 值回退到默认。
- `beyondViewport`：`true` 捕获整个可滚动文档而非仅可见视口。设置 `selector` 时忽略。带 `annotate=true` 时返回的 box 坐标为文档相对坐标。
- `annotate`：`true` 在可交互元素上叠加编号 ref 框（或在 `selector` 匹配项上，此时收窄而非裁剪），并返回 `annotations` 列表。
- `noAnimations`：`true` 在捕获前禁用 CSS 动画。
- `raw`：`true` 直接返回图像字节而非 JSON。
- `output`：`file` 保存到状态目录。
- `tabId`：目标特定标签页。

### 命令行界面

不带 `-o` 时，命令行界面保存到当前目录的 `screenshot-<timestamp>.jpg`（或 `.png`）。

- `-o <path>`：保存到特定路径。省略 `--format` 时，`.png` 扩展名选择 PNG；其他扩展名用 JPEG。
- `--format <jpeg|png>`：显式选择图像格式，覆盖扩展名推断。
- `-q <0-100>`：设置 JPEG 质量。
- `-s <selector>`：捕获特定元素。
- `--scale <f>`：位图缩放（如 `0.5`）。
- `--beyond-viewport`：捕获整个可滚动文档。设置 `--selector` 时忽略。
- `--annotate`：叠加编号 ref 框，并向 stdout 打印 `[n] ref role "name"` 图例。
- `--tab <id>`：目标特定标签页。

## 何时改用 `pinchtab capture`

`screenshot` 只返回图像字节。当模型需要在读取像素的同一轮中基于 ref 进行操作时，使用 [`capture`](./capture.md)——来自同一 DOM epoch 的图像 + 无障碍快照配对。

## 相关页面

- [Capture](./capture.md)
- [Snapshot](./snapshot.md)
- [PDF](./pdf.md)
