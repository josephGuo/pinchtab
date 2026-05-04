# 选择

通过选择器或引用在原生 `<select>` 元素中选择选项。

```bash
curl -X POST http://localhost:9867/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"select","ref":"e12","value":"it"}'
# 命令行界面 替代方案
pinchtab select e12 it
# 响应（使用 --json 获取完整 JSON）
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

所有这些都可以根据页面工作：

```bash
pinchtab select e12 uk
pinchtab select e12 "United Kingdom"
pinchtab select e12 "united kingdom"
pinchtab select e12 "Kingdom"
```

当需要消除歧义时，首选规范选项值或完整可见文本。

选择器查找仅限于当前框架范围（默认：`main`）。在 iframe 选择前使用 [`/frame`](./frame.md)。

## 相关页面

- [框架](./frame.md)
- [快照](./snapshot.md)
- [聚焦](./focus.md)