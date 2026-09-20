# 选择

通过选择器或引用在原生 `<select>` 元素中选择选项。

```bash
curl -X POST http://localhost:9867/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"select","ref":"e12","value":"it"}'
# CLI Alternative
pinchtab select e12 it
# Response (use --json for full JSON)
OK
```

## 命令行界面 标志

| 标志 | 描述 |
|------|-------------|
| `--snap` | 选择后输出快照 |
| `--snap-diff` | 选择后输出快照差异 |
| `--text` | 选择后输出页面文本 |
| `--json` | 完整 JSON 响应 |
| `--tab` | 目标特定标签页 |

## 选项匹配

匹配是宽容的。PinchTab 按顺序尝试这些策略：

1. 精确的 `<option value="...">`
2. 精确的可见文本
3. 不区分大小写的可见文本
4. 不区分大小写的可见文本子字符串

具体哪条策略有效取决于页面：

```bash
pinchtab select e12 uk
pinchtab select e12 "United Kingdom"
pinchtab select e12 "united kingdom"
pinchtab select e12 "Kingdom"
```

当需要消除歧义时，首选规范选项值或完整可见文本。

API 读取 `value`，回退到 `text`。无任何匹配时，该动作应答 `422 option_not_found`；`details.available` 以 `{value, text}` 列出各选项，`details.hint` 将其渲染。

selector 查找限于当前 frame 范围（默认 `main`）。在 iframe 选择前使用 [`/frame`](./frame.md)。

## 相关页面

- [Frame](./frame.md)
- [Snapshot](./snapshot.md)
- [Focus](./focus.md)