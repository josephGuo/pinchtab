# 滚动（Scroll）

滚动当前标签页或特定元素。

```bash
curl -X POST http://localhost:9867/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"scroll","scrollY":800}'
# Response: {"success":true,"result":{"scrolled":true,...},"route":{...}}
# result also carries deltaX/deltaY (the scroll applied), targetX/targetY (the point scrolled at)
# and legacy x/y (same as deltaX/deltaY); an element scroll returns only scrolled:true

# CLI Alternative (human-readable by default)
pinchtab scroll down
# Output: OK

pinchtab scroll down --snap        # scroll and output snapshot
pinchtab scroll 800 --snap-diff    # scroll and output snapshot diff
pinchtab scroll 800 --json         # Full JSON response

pinchtab scroll -300               # scroll up 300px
pinchtab scroll --dy -300          # the same, as a flag
pinchtab scroll --dx -120          # scroll left 120px
```

注意：

- 两种写法都支持负像素数——`pinchtab scroll -300` 与 `pinchtab scroll --dy -300` 是同一次滚动，且 `--tab` 在两种写法中都可放在任意位置
- 要么给 flags，要么给一个位置参数，二者不能同时用；且只接受一个位置参数——因此 `--tab` 必须保持为 flag，不能放在 `--` 之后，否则会被当作位置参数
- 两轴 delta 都为零会被拒绝：命令行界面在发送前拒绝，服务器也拒绝（两轴都显式 `scrollX`/`scrollY` 为零是错误）。单轴显式为零（`--dy 0 --dx 500`）是真实滚动并放行。只有既无 delta 又无目标的请求才向下滚动默认 120px
- 方向关键字（`up`、`down`、`left`、`right`）每步滚动 800px
- 用 `--snap` 在滚动后输出交互式快照
- 用 `--snap-diff` 只输出相对上一个快照的变化
- 顶级命令行界面也接受像素值，如 `pinchtab scroll 800`
- 原始 API 用 `scrollY` 和 `scrollX` 做页面滚动
- 原始 API 也可用 `ref` 或 `selector` 定位元素
- selector 查找限于当前 frame 范围；默认范围是 `main`
- 在基于 selector 的 iframe 滚动前，使用 [`/frame`](./frame.md) 或 `pinchtab frame`

## 相关页面

- [Frame](./frame.md)
- [Snapshot](./snapshot.md)
- [Text](./text.md)
