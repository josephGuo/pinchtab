# 悬停

通过选择器或引用来将指针移动到元素上方。

```bash
curl -X POST http://localhost:9867/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"hover","ref":"e5"}'
# CLI Alternative
pinchtab hover e5
# Response (use --json for full JSON)
OK
```

当菜单或工具提示仅在悬停后出现时使用此功能。

## 命令行界面 标志

| 标志 | 描述 |
|------|-------------|
| `--css` | 使用 CSS 选择器而不是引用 |
| `--x`, `--y` | 在特定坐标处悬停 |
| `--humanize` | 启用人性化贝塞尔曲线 + 抖动输入路径（覆盖实例配置） |
| `--json` | 完整 JSON 响应 |
| `--tab` | 目标特定标签页 |

命令行界面 接受统一选择器形式：`e5`, `#menu`, `xpath://button`, `text:Menu`, `find:account menu`。

选择器查找仅限于当前框架范围（默认：`main`）。在 iframe 悬停调用前使用 [`/frame`](./frame.md)。

## 相关页面

- [点击](./click.md)
- [框架](./frame.md)
- [快照](./snapshot.md)