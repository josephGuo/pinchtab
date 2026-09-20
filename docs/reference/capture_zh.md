# Capture（成对抓取）

一次 HTTP 调用，从**同一个 DOM 纪元**返回成对的截图与无障碍快照。当模型需要在同一回合中既读取像素又基于 ref 执行操作时使用此接口——非成对的 `/screenshot` + `/snapshot` 序列在两次调用之间页面发生变动时会产生漂移。

```bash
# Default: file output, wait for page quiescence, bounds included
curl "http://localhost:9867/capture"

# CLI alternative — writes the image locally and prints a summary
pinchtab capture -o /tmp/cap.jpg

# Half-size image (snapshot/bounds unchanged)
pinchtab capture --scale 0.5

# Fail with 409 if the main frame navigates mid-capture
pinchtab capture --require-pair

# Full-document image; bounding boxes in page coords
pinchtab capture --beyond-viewport

# Scope to one element: image clips to it, snapshot subtree filters to it.
# Bounding boxes are relative to the clipped image origin.
pinchtab capture -s "#checkout-form"
```

## 响应（JSON）

```json
{
  "status": "ok",
  "tabId": "tab_abc",
  "url": "https://example.com/checkout",
  "title": "Checkout",
  "capturedAt": "2026-05-29T15:44:12.431Z",
  "epoch": {
    "frameId": "8E2F...A1",
    "loaderId": "5C9D...0B",
    "domEpoch": "ep_..."
  },
  "pairing": {
    "navigated": false,
    "captureDurationMs": 312
  },
  "image": {
    "format": "jpeg",
    "path": "/.../state/captures/cap-20260529-154412.jpg",
    "bytes": 184223,
    "coordinateSpace": "viewport",
    "devicePixelRatio": 2,
    "viewport": { "w": 1440, "h": 900, "scrollX": 0, "scrollY": 0 }
  },
  "snapshot": {
    "filter": "interactive",
    "nodeCount": 14,
    "nodes": [
      {
        "ref": "e4", "role": "textbox", "name": "Email",
        "boundingBox": { "x": 520, "y": 312, "w": 280, "h": 36 },
        "visible": true
      },
      {
        "ref": "e9", "role": "button", "name": "Submit",
        "boundingBox": { "x": 520, "y": 1480, "w": 96, "h": 40 },
        "visible": false
      }
    ]
  }
}
```

## 成对保证了什么

原子性契约是**"两次 CDP 调用之间主框架不发生导航"**——当主框架的 `loaderId` 在抓取窗口内发生变化时，`pairing.navigated` 翻转为 `true`。同一文档内部的漂移（React 重新渲染、`IntersectionObserver` 变化）不会被检测到；`wait=stable` 能减少但不能消除它。

`epoch.domEpoch` 是该标签页的 ref 词表 token——即 `/snapshot` 作为 `vocabularyToken` 返回的同一个不透明值——同时也会设置在 `X-PinchTab-Vocab` 响应头中。当新节点与上一个词表共享 ref 时它保持不变，不共享时则重新生成。在基于 ref 的操作中需将其作为 `vocab` 回传：使用与该标签页当前 token 不同的 token 进行 ref 操作会被拒绝，返回 `409 vocab_superseded`。

## 边界框与坐标系

当 `withBounds=true`（默认值）时，每个具有非零后端节点 id 的快照节点都会获得一个 `boundingBox` 和一个 `visible` 标志；无法测量的节点两者都没有。

`boundingBox` 是**边框盒（border box）**——元素绘制出的边缘，与 `pinchtab box`、`screenshot?annotate=true` 和 `getBoundingClientRect` 报告的是同一个矩形。从它绘制的覆盖层和裁剪矩形正好覆盖查看者所见的控件，来自 `/capture` 的框可以直接与来自上述任一接口的框比较。

坐标系取决于 `selector` 和 `beyondViewport`：

- **`viewport`**（默认）：框为相对于视口的 CSS 像素。图片即可见视口。`image.devicePixelRatio` 告诉你图片像素与 CSS 像素之比。
- **`clip`**（设置了 `selector` 时）：框相对于裁剪后图片的原点。响应还包含 `image.clip`，即原始文档相对的裁剪矩形。
- **`document`**（`beyondViewport=true` 时）：框使用页面坐标（`box.x` 和 `box.y` 包含滚动偏移）。图片为完整文档。

在这三种模式下，图片尺寸都恰好等于所报告空间乘以 `image.devicePixelRatio` 再乘以你请求的 `scale`（默认 `1`）：将 `boundingBox` 按 `devicePixelRatio × scale` 缩放，并从该空间所命名的原点出发，即可落在图片像素上。模式从不进入计算，`pinchtab set viewport` 也不——除了所报告的比率之外，唯一的额外因子就是你自己传入的那个，而响应报告的是页面比率而非乘积。在默认 `scale` 下，单凭该比率就是完整的映射关系。

保持这一等价关系，让默认抓取放弃了更快的"读取视图"路径——那条路径返回真实窗口表面，其缩放因子是屏幕的而非页面的，且视口模拟不会触及它——因此一个空闲的有头浏览器可能让 `/capture` 比 `/screenshot` 更慢。

当框具有正面积且与视口相交时，`visible` 为 true——这是一个廉价的启发式判断，而非严格的遮挡检查。向任一方向滚出的节点仍会被测量，并报告 `"visible": false`。

**"缺失即未测量"，绝非"否"。** `visible` 仅在 `boundingBox` 出现时出现，因此该 key 仅在未进行测量的位置缺失：`withBounds=false`、无后端节点 id 的节点，或框查询失败的节点。将缺失的 `visible` 视为未知，将存在的 `false` 视为在屏幕外；这是两种不同的答案。

### 此处的 `visible` 不是 `GET /visible`

`GET /visible`（以及 `pinchtab visible <ref>`）在同一个词下回答的是另一个问题：**CSS 渲染状态**——`display`、`visibility`、`opacity`，以及一个已布局或已定位且尺寸非零的框。**滚动位置不是输入**，因此远在首屏下方的元素在那里是 `"visible": true`，而本快照在同一时刻对同一节点报告 `"visible": false`。两个答案都正确且都有用：一个说元素是否被渲染出来，另一个说它此刻是否在屏幕上。

`GET /visible` 还返回 `onScreen`，即为该单个元素计算的本快照判定，因此一次调用即可回答两个问题。`onScreen` 遵循与上述 `visible` 相同的"缺失即未测量"规则。

## 常用 flags

### API 查询参数

| 参数 | 说明 |
|-----------|-------------|
| `tabId` | 指定目标标签页 |
| `selector` | 范围：将图片裁剪并将快照子树过滤到同一元素 |
| `filter` | `interactive`（默认）或 `all` |
| `format` | `jpeg`（默认）或 `png` |
| `quality` | JPEG 质量 0-100（默认 `80`） |
| `depth` | 快照树深度限制（默认 `-1`，完整树） |
| `output` | `file`（默认）、`inline`（JSON 中 base64）或 `raw`（仅字节——丢弃快照） |
| `wait` | `stable`（默认）等待 `Page.lifecycleEvent` 静止（250ms 无活动 / 750ms 上限）；`load` 轮询 `document.readyState` 直到 `complete`（2s 上限）；`none` 跳过等待 |
| `withBounds` | `true`（默认）——为每个可测量的快照节点填充 `boundingBox`（边框盒）+ `visible`；`false` 则在所有位置省略这两个 key |
| `beyondViewport` | `true`——抓取完整可滚动文档；坐标系变为 `document` |
| `scale` | 重新缩放输出位图。默认 `1`。`0.5` 将每轴减半（像素数为四分之一）；超出范围的值会被钳制到范围内 |
| `requirePair` | `true`：若 `pairing.navigated` 将为 true 则返回 409 |
| `noAnimations` | `true`——在抓取窗口内注入 `prefers-reduced-motion` CSS |

### 命令行界面

| Flag | 说明 |
|------|-------------|
| `-o <path>` | 将抓取的图片保存到本地（默认：`capture-<ts>.jpg`） |
| `-s <selector>` | 范围：裁剪图片并过滤快照子树 |
| `--filter <name>` | 快照过滤器 |
| `--format <fmt>` | `jpeg` 或 `png`（默认：当 `-o` 以 `.png` 结尾时为 `png`，否则为 `jpeg`） |
| `-q <0-100>` | JPEG 质量 |
| `--depth <n>` | 快照深度限制 |
| `--wait <mode>` | `stable`（默认）/ `load` / `none` |
| `--with-bounds` | 布尔值（默认 true） |
| `--beyond-viewport` | 抓取完整文档 |
| `--scale <f>` | 位图重新缩放（例如 `0.5`） |
| `--require-pair` | 抓取中途导航则以 409 失败 |
| `--tab <id>` | 指定目标标签页 |
| `--json` | 打印完整 JSON 响应而非简略摘要 |

`GET /tabs/{id}/capture` 是同一个处理器，标签页在路径中。未知的 `output` 值返回 400；`requirePair` 失败返回 409，待处理的 JavaScript 对话框也一样（`409 dialog_blocked`）。

## 相关页面

- [Screenshot](./screenshot.md)
- [Snapshot](./snapshot.md)
- [Frame](./frame.md)
