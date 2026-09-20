# 输入

在元素中输入文本，在输入文本时发送按键事件。

```bash
curl -X POST http://localhost:9867/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"type","ref":"e8","text":"Ada Lovelace"}'
# CLI Alternative
pinchtab type e8 "Ada Lovelace"
# Response (use --json for full JSON)
OK
```

## 命令行界面 Flags

| Flag | 说明 |
|------|-------------|
| `--humanize` | 使用拟人化的逐字符按键时序（覆盖实例配置） |
| `--json` | 完整 JSON 响应 |
| `--tab` | 目标特定标签页 |

## 注意事项

- 想更直接地设值时用 `fill`
- 接受统一选择器：`e8`、`#name`、`xpath://input`、`text:Name`
- selector 查找限于当前 frame 范围（默认 `main`）
- iframe 输入前用 [`/frame`](./frame.md)
- 缺失 selector 立即失败；异步字段用 `pinchtab wait`（见 [`commands.md`](../commands.md)）
- 要向聚焦元素输入，用 `keyboard type`
- 原始键盘输入是默认。要让一次 type 动作走较慢的拟人化逐字符路径，在动作 JSON 中传 `humanize:true` 或设 `instanceDefaults.humanize:true`。

## 相关页面

- [Frame](./frame.md)
- [Fill](./fill.md) — 直接设置输入值
- [Keyboard](./keyboard.md) — 低层键盘输入（在聚焦元素处输入）
- [Snapshot](./snapshot.md)