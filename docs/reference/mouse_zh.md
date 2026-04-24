# 鼠标

用于拖动句柄、类画布 UI、悬停驱动菜单和 DOM 原生 `click` 或 `hover` 不足够的流程的低级指针控制。

## CLI

```bash
pinchtab mouse move <x> <y>
pinchtab mouse move <selector>

pinchtab mouse down [selector] --button left
pinchtab mouse up [selector] --button left

pinchtab mouse wheel <dy> [--dx <n>]
pinchtab mouse wheel [selector]

pinchtab drag <from> <to>
```

示例：

```bash
# 移动到元素，然后使用当前指针语义
pinchtab mouse move e5
pinchtab mouse down --button left
pinchtab mouse move 400 320
pinchtab mouse up --button left

# 在元素上明确目标按下/释放
pinchtab mouse down e5 --button left
pinchtab mouse up e5 --button left

# 在当前指针处滚动
pinchtab mouse wheel 240 --dx 40

# 在新目标处滚动
pinchtab mouse wheel e5

# 从元素拖动到坐标
pinchtab drag e5 400,320
```

注意：

- `mouse move` 接受坐标或统一选择器。
- `mouse down` 和 `mouse up` 接受可选的选择器。没有选择器时，它们使用当前指针位置。
- `mouse wheel` 接受增量形式 (`<dy> [--dx <n>]`) 或可选的选择器。没有选择器时，它使用当前指针位置。
- `drag <from> <to>` 接受选择器/引用目标或 `x,y` 坐标对。
- `button` 支持 `left`、`right` 和 `middle`。

## HTTP API

规范的操作类型：

- `mouse-move`
- `mouse-down`
- `mouse-up`
- `mouse-wheel`
- `drag`

目标字段：

- `ref`
- `selector`
- `nodeId`
- `x` 和 `y`

滚轮字段：

- `deltaX`
- `deltaY`

示例：

```bash
# 移动到元素
curl -X POST http://localhost:9867/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"mouse-move","ref":"e5"}'

# 移动到坐标
curl -X POST http://localhost:9867/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"mouse-move","x":120,"y":220}'

# 在当前指针处按下/释放
curl -X POST http://localhost:9867/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"mouse-down","button":"left"}'

curl -X POST http://localhost:9867/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"mouse-up","button":"left"}'

# 在明确目标处按下/释放
curl -X POST http://localhost:9867/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"mouse-down","ref":"e5","button":"left"}'

curl -X POST http://localhost:9867/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"mouse-up","ref":"e5","button":"left"}'

# 在当前指针处滚动
curl -X POST http://localhost:9867/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"mouse-wheel","deltaY":240,"deltaX":40}'

# 在明确坐标处滚动
curl -X POST http://localhost:9867/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"mouse-wheel","x":400,"y":320,"deltaY":240}'
```

标签页范围示例：

```bash
curl -X POST http://localhost:9867/tabs/<tabId>/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"mouse-move","ref":"e5"}'
```

## 行为

- POST 坐标体使用普通的 `x` 和 `y`；不需要额外的 `hasXY` 标志。
- 当你省略新目标时，`mouse-down`、`mouse-up` 和 `mouse-wheel` 使用每个标签页的当前指针状态。
- 如果尚未知道当前指针位置，`mouse-down`、`mouse-up` 和 `mouse-wheel` 会失败并显示明确的错误。请先使用 `mouse-move` 或传递明确的目标。
- 当只提供 `deltaY` 时，`mouse-wheel` 默认垂直滚动。

## 相关页面

- [点击](./click.md)
- [悬停](./hover.md)
- [滚动](./scroll.md)
- [CLI](./cli.md)