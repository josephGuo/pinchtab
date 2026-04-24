# 查找

`/find` 允许 PinchTab 通过自然语言描述而不是 CSS 选择器或 XPath 来定位元素。

它针对标签页的可访问性快照工作，并返回最佳匹配的 `ref`，你可以将其传递给 `/action`。

## 端点

PinchTab 公开两种形式：

- `POST /find`
- `POST /tabs/{id}/find`

当你直接与桥接式运行时或简写路由通信并希望在请求体中传递 `tabId` 时，使用 `POST /find`。

当你已经知道标签页 ID 并希望编排器将请求路由到正确的实例时，使用 `POST /tabs/{id}/find`。

## 请求体

| 字段 | 类型 | 必需 | 默认值 | 描述 |
| --- | --- | --- | --- | --- |
| `query` | string | 是 | - | 目标元素的自然语言描述 |
| `tabId` | string | 否 | 活动标签页 | 使用 `POST /find` 时的标签页 ID |
| `threshold` | float | 否 | `0.3` | 最低相似度分数 |
| `topK` | int | 否 | `3` | 返回的最大匹配数 |
| `lexicalWeight` | float | 否 | 匹配器默认值 | 覆盖词汇分数权重 |
| `embeddingWeight` | float | 否 | 匹配器默认值 | 覆盖嵌入分数权重 |
| `explain` | bool | 否 | `false` | 包含每个匹配的分数细分 |

## 主要示例

```bash
curl -X POST http://localhost:9867/tabs/<tabId>/find \
  -H "Content-Type: application/json" \
  -d '{"query":"login button","threshold":0.3,"topK":3}'
# CLI 替代方案
pinchtab find --tab <tabId> "login button"
```

有一个专用的 CLI `find` 命令：

```bash
pinchtab find "login button"
pinchtab find --threshold 0.5 --explain "primary submit button"
pinchtab find --ref-only "search input"
```

## 使用 `POST /find`

```bash
curl -X POST http://localhost:9867/find \
  -H "Content-Type: application/json" \
  -d '{"tabId":"<tabId>","query":"search input"}'
```

如果省略 `tabId`，PinchTab 将使用当前桥接上下文中的活动标签页。

## 响应字段

| 字段 | 描述 |
| --- | --- |
| `best_ref` | 用于 `/action` 的最高评分元素引用 |
| `confidence` | `high`、`medium` 或 `low` |
| `score` | 最佳匹配的分数 |
| `matches` | 高于阈值的顶级匹配 |
| `strategy` | 使用的匹配策略 |
| `threshold` | 请求使用的阈值 |
| `latency_ms` | 匹配时间（毫秒） |
| `element_count` | 评估的元素数量 |
| `idpiWarning` | 当 IDPI 处于警告模式时的咨询警告 |

当启用 `explain` 时，每个匹配还可能包含词汇和嵌入分数详细信息。

## 置信度级别

| 级别 | 分数范围 | 含义 |
| --- | --- | --- |
| `high` | `>= 0.80` | 通常可以直接采取行动 |
| `medium` | `0.60 - 0.79` | 合理匹配，但关键操作需要验证 |
| `low` | `< 0.60` | 弱匹配；重新表述查询或谨慎降低阈值 |

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

然后使用返回的引用：

```bash
curl -X POST http://localhost:9867/tabs/<tabId>/action \
  -H "Content-Type: application/json" \
  -d '{"ref":"e14","kind":"type","text":"user@pinchtab.com"}'
```

## 操作说明

- `/find` 使用标签页的可访问性快照，而不是原始 DOM 选择器。
- 如果没有缓存的快照，PinchTab 会在匹配前尝试自动刷新它。
- 成功的匹配是 `/action`、`/actions` 和更高级别恢复逻辑的有用输入。
- `200` 响应仍然可能返回空的 `best_ref`，如果没有任何内容达到阈值。

## 错误情况

| 状态 | 条件 |
| --- | --- |
| `400` | 无效的 JSON 或缺少 `query` |
| `403` | 被严格模式下的 IDPI 阻止 |
| `404` | 标签页未找到 |
| `500` | Chrome 未初始化、快照不可用或匹配器失败 |