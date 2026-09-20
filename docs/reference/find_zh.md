# 查找（Find）

`/find` 让 PinchTab 通过自然语言描述而非 CSS 选择器或 XPath 来定位元素。

它基于某标签页的无障碍快照工作，返回最佳匹配的 `ref`，你可将其传给 `/action`。

## 端点

PinchTab 提供两种形式：

- `POST /find`
- `POST /tabs/{id}/find`

当你直接与 bridge 风格运行时或简写路由通信、并希望在请求体中传 `tabId` 时，使用 `POST /find`。

当你已知标签页 ID 并希望编排器将请求路由到正确实例时，使用 `POST /tabs/{id}/find`。

## 请求体

| 字段 | 类型 | 必填 | 默认值 | 描述 |
| --- | --- | --- | --- | --- |
| `query` | string | 是 | - | 目标元素的自然语言描述 |
| `tabId` | string | 否 | 当前活动标签页 | 使用 `POST /find` 时的标签页 ID |
| `threshold` | float | 否 | `0.3` | 最低相似度分数 |
| `topK` | int | 否 | `3` | 返回的最大匹配数 |
| `lexicalWeight` | float | 否 | 匹配器默认 | 覆盖词汇分数权重 |
| `embeddingWeight` | float | 否 | 匹配器默认 | 覆盖嵌入分数权重 |
| `explain` | bool | 否 | `false` | 包含每个匹配的分数细分 |

## 主示例

```bash
curl -X POST http://localhost:9867/tabs/<tabId>/find \
  -H "Content-Type: application/json" \
  -d '{"query":"login button","threshold":0.3,"topK":3}'
# CLI Alternative
pinchtab find --tab <tabId> "login button"
```

有一个专用的命令行界面 `find` 命令：

```bash
pinchtab find "login button"
pinchtab find --threshold 0.5 --explain "primary submit button"
pinchtab find --ref-only "search input"
```

没有匹配时，`find` 打印 `No elements matched "<query>"` 并以 0 退出——空结果就是一个答案，与 `console` 和 `errors` 报告"无可显示内容"同理。`--ref-only` 则刻意将该行打印到 stderr 并以非零退出码退出：它是为 `REF=$(pinchtab find … --ref-only)` 设计的，因为把空 `REF` 花在一次点击上比命令失败更糟。

## 使用 `POST /find`

```bash
curl -X POST http://localhost:9867/find \
  -H "Content-Type: application/json" \
  -d '{"tabId":"<tabId>","query":"search input"}'
```

省略 `tabId` 时，PinchTab 使用当前 bridge 上下文中的活动标签页。

## 响应字段

| 字段 | 描述 |
| --- | --- |
| `best_ref` | 用于 `/action` 的最高评分元素 ref |
| `confidence` | `high`、`medium` 或 `low` |
| `score` | 最佳匹配的分数 |
| `matches` | 高于阈值的靠前匹配 |
| `strategy` | 使用的匹配策略 |
| `threshold` | 请求所用的阈值 |
| `latency_ms` | 匹配耗时，毫秒 |
| `element_count` | 评估的元素数 |
| `idpiWarning` | IDPI 处于 warn 模式时的提示性警告 |

启用 `explain` 时，每个匹配还可能附带词汇和嵌入分数详情。

## 查询语法

除普通自然语言描述外，匹配器还理解三种查询修饰：

### 序数查询

用序数词从本来看起来相似的匹配中挑出位置：

```bash
pinchtab find "first button"
pinchtab find "second search result"
pinchtab find "last link"
```

序数匹配在语义评分之后应用。快照无坐标时，以文档顺序作为稳定顺序。

### 否定查询

用 `not`、`without`、`exclude`、`excluding`、`except`、`no` 或 `ignore` 把元素从匹配中推开：

```bash
# Picks Cancel over Submit
pinchtab find "button not submit"

# Compose multiple exclusions
pinchtab find "input no password no username"

# Exclude a phrase
pinchtab find "button without sign in"
```

触发词之前的 token 为肯定；其后到下一个触发词或查询结尾之间的一切为否定。

### 视觉/位置查询

方向性和相对短语将匹配偏向页面上对应位置的元素：

```bash
# Directional (top / bottom / left / right / corner)
pinchtab find "bottom button"
pinchtab find "button in top right corner"
pinchtab find "sidebar on the left"

# Anchor-relative (above / below / under / over)
pinchtab find "link below the search box"
pinchtab find "button above the footer"
```

无障碍快照无坐标时，文档顺序作为垂直位置的回退——因此 `"bottom button"` 选中快照中最后一个匹配的按钮。元素边界框可用时优先用它。

视觉提示由组合匹配器（默认）应用。它们不影响纯词汇匹配器。

## 置信度级别

| 级别 | 分数区间 | 含义 |
| --- | --- | --- |
| `high` | `>= 0.80` | 通常可直接操作 |
| `medium` | `0.60 - 0.79` | 合理匹配，但关键操作需验证 |
| `low` | `< 0.60` | 弱匹配；改写查询或谨慎降低阈值 |

## 常见流程

标准模式是：

```text
navigate -> find -> action
```

示例：

```bash
curl -X POST http://localhost:9867/tabs/<tabId>/find \
  -H "Content-Type: application/json" \
  -d '{"query":"username input"}'
```

然后使用返回的 ref：

```bash
curl -X POST http://localhost:9867/tabs/<tabId>/action \
  -H "Content-Type: application/json" \
  -d '{"ref":"e14","kind":"type","text":"user@pinchtab.com"}'
```

## 操作说明

- `/find` 使用标签页的无障碍快照，而非原始 DOM 选择器。
- 查询是自然语言，不是选择器：CSS、XPath 和 ref 不会被解析，只作为文本评分。在查询前加 `find:` 或 `semantic:` 前缀可强制自然语言匹配。
- 结构化 `/find` 查询如 `role:button Save`、`text:Submit`、`label:Email`、`placeholder:Search`、`alt:Logo`、`title:Close`、`testid:submit`、`first:role:button`、`last:text:Submit` 和 `nth:2:label:Email` 由语义引擎对照富化描述符匹配。`/find` 原样把查询交给匹配器，因此其 `nth:` 索引从 1 开始计数（`nth:1:label:Email` 是第一个匹配）；`nth:0` 或越过最后一个匹配的索引返回 `200` 且 `best_ref` 为空。从零开始的索引和越界拒绝适用于动作选择器中的 `nth:`，而非 `/find`。
- 在动作命令中，`role:`、`label:`、`placeholder:`、`alt:`、`title:`、`testid:` 以及围绕这些形式的包装使用语义匹配。CSS、XPath、ref、既有的 `text:` 动作选择器，以及 `first:button` 这类裸 CSS/文本包装仍走浏览器端选择器解析。
- 若无缓存快照，PinchTab 在匹配前会尝试自动刷新它。
- 每个成功响应都在 `X-PinchTab-Vocab` 头中携带该标签页的 ref 词表 token；命令行界面和 MCP 会捕获它，因此返回的 ref 可在中间无需再快照的情况下直接操作。
- 成功匹配是 `/action`、`/actions` 和更高级恢复逻辑的有用输入。
- 若无任何内容达到阈值，`200` 响应仍可能返回空的 `best_ref`。

## 错误情况

| 状态码 | 条件 |
| --- | --- |
| `400` | 无效 JSON 或缺少 `query` |
| `403` | 被严格模式下的 IDPI 拦截 |
| `404` | 标签页未找到 |
| `409` | `dialog_blocked`：JavaScript 对话框正阻塞该标签页 |
| `500` | Chrome 未初始化、快照无元素，或匹配器失败 |
