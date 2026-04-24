# 点击

使用快照引用、CSS 选择器、XPath 选择器、文本选择器或语义选择器点击元素。

```bash
curl -X POST http://localhost:9867/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"click","ref":"e5"}'
# CLI 替代方案
pinchtab click e5
# 响应（使用 --json 获取完整 JSON）
OK
```

## CLI 标志

| 标志 | 描述 |
|------|-------------|
| `--css` | 使用 CSS 选择器而不是引用 |
| `--wait-nav` | 点击后等待导航完成 |
| `--snap` | 点击后输出交互式快照 |
| `--snap-diff` | 点击后输出快照差异 |
| `--text` | 点击后输出页面文本 |
| `--dialog-action` | 自动处理 JS 对话框：`accept` 或 `dismiss` |
| `--dialog-text` | 提示响应文本（配合 `--dialog-action accept`） |
| `--x`, `--y` | 在特定坐标处点击 |
| `--json` | 完整 JSON 响应 |
| `--tab` | 目标特定标签页 |

## 示例

```bash
pinchtab click e5                       # 通过引用点击
pinchtab click "#login"                 # 通过 CSS 点击
pinchtab click "text:Submit"            # 通过文本点击
pinchtab click e5 --snap                # 点击并显示新快照
pinchtab click e5 --wait-nav            # 点击并等待导航完成
pinchtab click e5 --dialog-action accept  # 自动接受警告/确认
pinchtab click --x 100 --y 200          # 在坐标处点击
```

## 注意事项

- 元素引用来自 `/snapshot`
- iframe 后代的引用可以直接点击，无需切换框架
- 选择器查找仅限于当前框架范围（默认：`main`）
- 在基于选择器的 iframe 操作前使用 [`/frame`](./frame.md)
- 缺失的选择器会立即失败；对于动态 UI，先使用 [`pinchtab wait`](./wait.md)
- API 也接受 `selector` 字段：`{"kind":"click","selector":"#login"}`

## 相关页面

- [框架](./frame.md)
- [快照](./snapshot.md)
- [导航](./navigate.md)