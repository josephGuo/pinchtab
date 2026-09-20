# 鼠标（Mouse）

用于拖动句柄、类画布 UI、悬停驱动菜单，以及 DOM 原生 `click` 或 `hover` 不足的流程的低级指针控制。

## 命令行界面

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
# Move to an element, then use current-pointer semantics
pinchtab mouse move e5
pinchtab mouse down --button left
pinchtab mouse move 400 320
pinchtab mouse up --button left

# Explicitly target down/up at an element
pinchtab mouse down e5 --button left
pinchtab mouse up e5 --button left

# Wheel at current pointer
pinchtab mouse wheel 240 --dx 40

# Wheel at a fresh target
pinchtab mouse wheel e5

# Drag from an element to coordinates
pinchtab drag e5 400,320
```

注意：

- 所有 `mouse` 子命令都接受 `--humanize`，让该动作走人性化贝塞尔 + 抖动输入路径（覆盖 `instanceDefaults.humanize`）。`drag` 没有 `--humanize` flag。
- `mouse move` 接受坐标或统一选择器。
- `mouse down` 和 `mouse up` 接受可选选择器。没有时使用当前指针位置。
- `mouse wheel` 接受增量形式（`<dy> [--dx <n>]`）或可选选择器。没有选择器时使用当前指针位置。
- `drag <from> <to>` 两端都接受选择器/ref 目标或 `x,y` 坐标对。
- `drag` 有两种形式，仅在如何表达拖动终点上不同：一个目标（`drag <from> <to>`，`toSelector`/`toX`/`toY`）或一个偏移（`--drag-x/--drag-y`，`dragX`/`dragY`）。两者都是"按钮仍按住时移动指针"的单一动作，这正是让 HTML5 拖放生效的关键——Chrome 只在报告了按下按钮的移动上进入拖动管线。每个 `mouse move` 是独立请求，不知道有按下，因此用 `mouse move`/`mouse down`/`mouse up` 拼出的拖动无论分多少小步都不触发 `dragstart`——步数不是问题。对任何带 `draggable` 源的东西用 `drag`。
- `button` 支持 `left`、`right`、`middle`。其他名字被拒绝——命令行界面在发送前拒绝，HTTP 动作主体以 400 应答并指出这三个。省略 `button` 仍表示 `left`；这是默认值，而服务器不认识的名字过去会被重新解释为 `left` 并报告成功。DOM 的 `primary` 和 `secondary` 被拒绝而非映射，因为 `primary` 恰好是左，而 `secondary` 不是。
- 指针移动先使用有界的 CDP `mouseMoved` 派发。若无头 Chromium 让该派发卡住，PinchTab 回退到 DOM 鼠标事件，使悬停和鼠标移动流程保持响应。

## HTTP API

规范动作种类：

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
# Move to an element
curl -X POST http://localhost:9867/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"mouse-move","ref":"e5"}'

# Move to coordinates
curl -X POST http://localhost:9867/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"mouse-move","x":120,"y":220}'

# Press/release at current pointer
curl -X POST http://localhost:9867/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"mouse-down","button":"left"}'

curl -X POST http://localhost:9867/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"mouse-up","button":"left"}'

# Press/release at an explicit target
curl -X POST http://localhost:9867/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"mouse-down","ref":"e5","button":"left"}'

curl -X POST http://localhost:9867/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"mouse-up","ref":"e5","button":"left"}'

# Wheel at current pointer
curl -X POST http://localhost:9867/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"mouse-wheel","deltaY":240,"deltaX":40}'

# Wheel at explicit coordinates
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

- POST 坐标主体用普通 `x` 和 `y` 即可；不需要额外 `hasXY` 标志。
- 省略新目标时，`mouse-down`、`mouse-up`、`mouse-wheel` 使用每个标签页的当前指针状态。
- 若尚不知道当前指针位置，`mouse-down` 和 `mouse-up` 以明确错误失败。先用 `mouse-move` 或传明确目标。
- 若尚不知道当前指针位置，`mouse-wheel` 用视口中心作为确定性回退。不带 `x`/`y` 的页面 `scroll` 始终以视口中心为目标。
- 仅提供 `deltaY` 时 `mouse-wheel` 默认垂直滚动。
- CDP 到 DOM 的移动回退只处理渲染进程确认超时。其他 CDP 错误和调用方取消都返回给调用方。

理由：低级指针动作常用于悬停菜单、拖动句柄和类画布控件，五秒的无头鼠标移动卡顿会让一次简单检查看起来像整套测试超时。回退让这些流程保持快速，同时仍暴露真实的 CDP 错误。

## 相关页面

- [Click](./click.md)
- [Hover](./hover.md)
- [Scroll](./scroll.md)
- [CLI](./cli.md)
