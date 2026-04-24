# 滚动

滚动当前标签页或特定元素。

```bash
curl -X POST http://localhost:9867/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"scroll","scrollY":800}'
# 响应: {"success":true,"result":{"success":true}}

# CLI 替代方案（默认人类可读）
pinchtab scroll down
# 输出: OK

pinchtab scroll down --snap        # 滚动并输出快照
pinchtab scroll 800 --snap-diff    # 滚动并输出快照差异
pinchtab scroll 800 --json         # 完整 JSON 响应
```

注意：

- 使用 `--snap` 在滚动后输出交互式快照
- 使用 `--snap-diff` 仅输出与前一个快照的变化
- 顶级 CLI 也接受像素值，例如 `pinchtab scroll 800`
- 原始 API 使用 `scrollY` 和 `scrollX` 进行页面滚动
- 原始 API 也可以使用 `ref` 或 `selector` 目标元素
- 选择器查找仅限于当前框架范围；默认范围是 `main`
- 在基于选择器的 iframe 滚动前使用 [`/frame`](./frame.md) 或 `pinchtab frame`

## 相关页面

- [框架](./frame.md)
- [快照](./snapshot.md)
- [文本](./text.md)