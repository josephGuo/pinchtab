# 聚焦

通过选择器或引用来移动焦点到元素。

```bash
curl -X POST http://localhost:9867/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"focus","ref":"e8"}'
# CLI Alternative
pinchtab focus e8
# Response (use --json for full JSON)
OK
```

## 命令行界面 标志

| 标志 | 描述 |
|------|-------------|
| `--css` | 使用 CSS 选择器而不是引用 |
| `--json` | 完整 JSON 响应 |
| `--tab` | 目标特定标签页 |

这在仅键盘流程（如 `press Enter` 或 `type`）之前很有用。

命令行界面 接受统一选择器形式：`e8`, `#input`, `xpath://input`, `text:Email`。

选择器查找仅限于当前框架范围（默认：`main`）。在 iframe 聚焦调用前使用 [`/frame`](./frame.md)。

## 相关页面

- [框架](./frame.md)
- [按键](./press.md)
- [输入](./type.md)