# 输入

在元素中输入文本，在输入文本时发送按键事件。

```bash
curl -X POST http://localhost:9867/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"type","ref":"e8","text":"Ada Lovelace"}'
# 命令行界面 替代方案
pinchtab type e8 "Ada Lovelace"
# 响应（使用 --json 获取完整 JSON）
OK
```

## 命令行界面 标志

| 标志 | 描述 |
|------|-------------|
| `--json` | 完整 JSON 响应 |
| `--tab` | 目标特定标签页 |

## 注意事项

- 当你想更直接地设置值时，使用 `fill`
- 接受统一选择器：`e8`、`#name`、`xpath://input`、`text:Name`
- 选择器查找仅限于当前框架范围（默认：`main`）
- 在 iframe 输入前使用 [`/frame`](./frame.md)
- 缺失的选择器会立即失败；对于异步字段，使用 [`pinchtab wait`](./wait.md)
- 要在聚焦元素中输入，使用 `keyboard type`

## 相关页面

- [框架](./frame.md)
- [填充](./fill.md) — 直接设置输入值
- [键盘](./keyboard.md) — 低级键盘输入（在聚焦元素处输入）
- [快照](./snapshot.md)