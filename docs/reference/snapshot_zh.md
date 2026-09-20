# 快照（Snapshot）

获取当前页面的无障碍快照，包含可被动作命令复用的元素 ref。

快照捕获时自动检测 iframe 内容。同源 iframe 后代包含在 iframe 拥有者元素之下，其 ref 可直接与动作命令复用。跨源 iframe 目前仍只作为拥有者节点存在。

selector 作用域是显式的。`selector=...` 只在当前 frame 范围内搜索，默认 `main`。要把基于 selector 的快照作用域到 iframe，先用 [`/frame`](./frame.md) 或 `pinchtab frame` 设置 frame。

```bash
curl "http://localhost:9867/snapshot?filter=interactive"
# CLI Alternative (defaults to compact text output)
pinchtab snap -i
# Output: a "# <title> | <url> | <N> nodes" header line, then one node per line
e5:link "More information..."

# --json (same as --compact=false) keeps the interactive filter; --full returns every node as JSON
pinchtab snap --full
```

## 命令行界面 Flags

| Flag | 说明 |
|------|-------------|
| `-i`, `--interactive` | 过滤到可交互元素 + 标题（默认 true） |
| `-c`, `--compact` | 紧凑文本输出（默认 true） |
| `--json` | JSON 输出，保留交互式过滤（同 `--compact=false`） |
| `-d`, `--diff` | 显示相对上一个快照的 diff |
| `--full` | 完整 JSON 输出（`--interactive=false --json` 的简写） |
| `--text` | 文本输出格式 |
| `-s`, `--selector` | 限定快照的 selector（也接受为位置参数 `[selector]`） |
| `--max-tokens` | 最大 token 预算 |
| `--depth` | 树深度上限 |
| `--tab` | 目标特定标签页 |

## 示例

```bash
pinchtab snap                           # Interactive compact (default)
pinchtab snap -i -c                     # Same as above
pinchtab snap --full                    # Full JSON with all nodes
pinchtab snap -d                        # Show changes since last snapshot
pinchtab snap --selector "#main"        # Scope to element
pinchtab snap --max-tokens 2000         # Limit output size
```

## API 参数

| 参数 | 说明 |
|-----------|-------------|
| `tabId` | 目标标签页（默认当前标签页） |
| `filter` | `interactive` 表示可交互 + 标题；`all`（默认）表示整棵树 |
| `interactive` | `filter` 的布尔别名：`true` 即 `filter=interactive`，`false` 即 `filter=all`。与显式 `filter` 矛盾则 400 |
| `format` | `compact`、`text`、`yaml` 或默认 JSON。命令行界面和 MCP 工具都请求 `compact`；HTTP 默认不变 |
| `diff` | `true` 进入 diff 模式 |
| `selector` | 用于限定的统一 selector（ref、CSS、XPath、text……） |
| `maxTokens` | token 预算上限（正整数；其他值 400） |
| `depth` | 树深度上限（`-1` = 无限制；低于 `-1` 为 400） |
| `noAnimations` | `true` 在捕获前禁用动画一次 |
| `output` | `file` 将快照写入状态目录的 `snapshots/`，返回 `{path, size, format, timestamp}` |
| `path` | 配合 `output=file`，状态目录内的文件路径（目录外则 400） |

`GET /tabs/{id}/snapshot` 是同一处理器，只是标签页放在路径中。

未知查询参数不会被拒绝；它们以 `ignoredParams`（JSON/YAML）或一行 `# ignored params:` 注释（compact/text）回显，使打错的 flag 可见。

## 响应

默认 JSON 主体携带 `url`、`title`、`route`、`nodes`、`count` 和 `vocabularyToken`，预算截断树时还有 `truncated`/`maxTokens`，selector 匹配到无可访问节点的元素时还有 `hint`。`vocabularyToken` 命名本次快照发放的 ref 词表；无论何种格式，每次快照还把它设在 `X-PinchTab-Vocab` 响应头（连同 `X-PinchTab-Tab-Id`）。在基于 ref 的动作上把它作为 `vocab` 回传（命令行界面替你做了）：以非当前标签页 token 下的 ref 为目标的动作以 `409 vocab_superseded` 拒绝——重新快照并用新 ref。

页面打开模态对话框时，快照作用域到最顶层对话框的子树。快照期间最顶层对话框变化两次，请求以 `409` 失败——等页面稳定后重试。挂起的 JavaScript 对话框（alert、confirm、prompt）以 `409 dialog_blocked` 阻断读取，直到用 `pinchtab dialog` 应答。

文档为 `hidden` 的标签页（后台标签页）在无障碍读取前会被渲染：PinchTab 启用 focus 模拟并等待一帧绘制（最多 1s），因此仅在可见时才布局的内容仍会出现在树中。

## 各格式的成本与携带内容

同一 40 个真实可交互节点，经各输出路径渲染：

```
  compact     1644 bytes  ~ 411 tokens   1.0x
  text        2060 bytes  ~ 515 tokens   1.3x
  json        8211 bytes  ~2052 tokens   5.0x
  yaml       17467 bytes  ~4366 tokens  10.6x
```

该 JSON 主体的 55% 是 DOM 派生的描述字段——`tag`、`label`、`placeholder`、`alt`、`title`、`testid`、`text`——它们的存在是为了喂给服务端语义匹配器。`compact` 和 `text` 只渲染 `ref`、`role`、`name`、`value` 和状态标志，别无其他，因此那些字段在这两种格式里根本不会出现。填充它们的 DOM 遍历对 `compact` 和 `text` 因此被跳过，最便宜的路径上每个节点还少一次 CDP 往返。

跳过它对调用者可见的内容没有任何改变：那些字段在两种格式里本来就缺失，且 `find` 和语义选择器在匹配前自行丰富缓存节点，因此 `compact` 快照的能力与 JSON 完全相同。

## `maxTokens` 保证什么

`maxTokens` 是上限，不是提示。返回的节点是渲染输出在你所请求格式下不超预算的最长前缀，因此响应绝不超过你的请求，最多差一个节点才到顶。在真实可交互节点页面上对 `compact`、`text`、`json`、`yaml` 实测，真正起约束作用的预算能交付其 87–100%。

成本是实测而非建模：每个节点按其自身格式发出的字节计费——`compact` 和 `text` 实际渲染，`json` 和 `yaml` 实际序列化——因此格式化器的改动会同步改变预算。token 按每 4 字节估算。

给定预算下各格式不可互换。同样节点下 `yaml` 约为 `json` 的三倍，因为节点结构只带 JSON 字段标签、不带 YAML 标签，所以 YAML 把每个字段（含空字段）都发出。同样 `maxTokens` 因此在 `yaml` 中返回的节点远少于 `json`——这是预算在起作用，不是回归。预算紧张时优先 `compact`：它在同样 token 下能容纳比任一结构化格式多数倍的节点。

## 节点上的控件状态

快照报告控件的状态，而不只是它的身份，使代理能验证自己的动作并读懂非自己搭建的页面：

- `value`——输入框当前文本或 `select` 的选中项
- `focused`、`disabled`、`hidden`——布尔，仅为 true 时出现
- `checked`——复选框、单选、`menuitemcheckbox`、`menuitemradio` 或任何带 `aria-checked` 的元素为 `"true"`、`"false"` 或 `"mixed"`

`checked` 是三值字符串而非布尔，因为 `"mixed"` 是真实状态：原生 indeterminate 复选框和 `aria-checked="mixed"` 都会报它。

**缺失的 `checked` 表示节点没有 checkedness——绝不表示它是 off。** 普通节点根本不带这个 key，因此缺失值绝不能读作未选中。处于 off 的控件会显式以 `"false"` 表明。

渲染格式携带全部三种状态，因此未选中的选项绝不会看起来像该字段不适用的节点：

| 状态 | `format=text` | `format=compact` |
|-------|---------------|------------------|
| checked | `[checked]` | `[x]` |
| unchecked | `[unchecked]` | `[ ]` |
| mixed | `[mixed]` | `[/]` |

因此单选组可从单个快照读懂：

```
e4 radio "Standard shipping" [checked]
e5 radio "Express shipping" [unchecked]
e6 radio "Pickup" [unchecked]
```

diff 模式把 `checked` 的变化视为变化，因此 `check` 之后的 `pinchtab snap -d` 把节点显示为 `[~]`。

## 相关页面

- [Click](./click.md)
- [Frame](./frame.md)
- [Tabs](./tabs.md)
