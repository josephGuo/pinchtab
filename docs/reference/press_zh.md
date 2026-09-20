# 按键（Press）

向当前标签页发送一个键盘按键或组合键，可选择先聚焦某个元素。

```bash
curl -X POST http://localhost:9867/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"press","key":"Enter"}'
# CLI Alternative
pinchtab press Enter
# Response (use --json for full JSON)
OK
```

## 命令行界面 Flags

| Flag | 说明 |
|------|-------------|
| `--snap` | 按键后输出交互式快照 |
| `--snap-diff` | 按键后输出快照差异 |
| `--text` | 按键后输出页面文本 |
| `--json` | 完整 JSON 响应 |
| `--tab` | 目标特定标签页 |

常用按键包括 `Enter`、`Tab`、`Escape`、`ArrowDown`、`ArrowUp`、`Backspace`、`Delete`。

命令行界面用法为 `pinchtab press [ref] <key|chord>`：带两个参数时，第一个是在发送按键前要聚焦的选择器或 ref（API：`ref` 或 `selector`）。

组合键用 `+` 连接修饰键和一个键，例如 `Ctrl+A` 或 `Shift+ArrowLeft`。修饰键为 `Ctrl`/`Control`、`Alt`/`Option`、`Shift` 以及 `Meta`/`Cmd`/`Command`/`Super`/`Win`（不区分大小写）；未知或重复的修饰键被拒绝。API 也可改用 CDP 位掩码 `modifiers`（Alt=1、Ctrl=2、Meta=4、Shift=8）配一个普通 `key`，但两种形式不能同时使用。

## 相关页面

- [Click](./click.md)
- [Focus](./focus.md)
- [Keyboard](./keyboard.md)
