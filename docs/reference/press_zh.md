# 按键

向当前标签页发送键盘按键。

```bash
curl -X POST http://localhost:9867/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"press","key":"Enter"}'
# CLI 替代方案
pinchtab press Enter
# 响应（使用 --json 获取完整 JSON）
OK
```

## CLI 标志

| 标志 | 描述 |
|------|-------------|
| `--json` | 完整 JSON 响应 |
| `--tab` | 目标特定标签页 |

常用按键包括 `Enter`、`Tab`、`Escape`、`ArrowDown`、`ArrowUp`、`Backspace`、`Delete`。

## 相关页面

- [点击](./click.md)
- [聚焦](./focus.md)
- [键盘](./keyboard.md)