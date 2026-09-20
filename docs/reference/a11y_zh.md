# 无障碍审计

`/a11y/audit` 对页面进行无障碍问题评分。它有两个引擎：

- **native**（默认）—— PinchTab 基于无障碍快照自研的规则集。速度快，不注入页面脚本。
- **axe** —— 内嵌的 [axe-core](https://github.com/dequelabs/axe-core)，在页面内运行，返回业界已知的规则 id、WCAG 标签和帮助 URL，并为每个失败元素附带一个快照 `ref`，让代理可以直接对其执行操作。

## 端点

- `GET /a11y/audit`
- `GET /tabs/{id}/a11y/audit`

## 查询参数

| 参数 | 引擎 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `engine` | 两者 | `native` | `native` 或 `axe`；其他取值返回 `400` |
| `tabId` | 两者 | 当前活动标签页 | 使用 `GET /a11y/audit` 时指定标签页 ID |
| `tags` | axe | 不含 best-practice 的 WCAG 标签 | 逗号分隔的 WCAG 标签列表，例如 `wcag2a,wcag2aa,best-practice` |
| `rules` | axe | 标签内全部规则 | 逗号分隔的规则 id 白名单；优先级高于 `tags` |
| `includeIncomplete` | axe | `false` | 包含 axe 无法判定的规则（`incomplete` 集合） |
| `selector` | axe | 整个文档 | 将运行范围限定到某个 CSS 选择器 |

只有当 `tags` 包含 `best-practice` 时才会出现 `best-practice` 规则；默认标签集仅包含 WCAG 合规性标签。

## axe 响应

```json
{
  "tabId": "...",
  "vocabularyToken": "...",
  "engine": "axe",
  "version": "4.13.0",
  "url": "https://example.com/",
  "violations": [
    {
      "id": "image-alt",
      "impact": "critical",
      "tags": ["wcag2a", "wcag111"],
      "help": "Images must have alternate text",
      "helpUrl": "https://dequeuniversity.com/rules/axe/4.13/image-alt",
      "nodes": [
        {
          "target": ["#no-alt"],
          "html": "<img id=\"no-alt\">",
          "ref": "e5",
          "failureSummary": "Fix any of the following: ..."
        }
      ]
    }
  ],
  "passes": 9,
  "incomplete": 0,
  "inapplicable": 51,
  "score": 90
}
```

- `version` 回显内嵌的 axe-core 版本（见 `THIRD_PARTY_LICENSES.md`）。
- 当失败元素能映射到某个快照 ref 时，`ref` 才会被填充，这样代理可以直接对它调用 `/action`。只有单跳目标（顶层框架元素）才会获得 ref；iframe 内部的节点不会附带 ref。
- axe 运行会发布该标签页的 ref 缓存，因此会返回一个新的 `vocabularyToken`（同时也通过 `X-PinchTab-Vocab` 头发送）；在后续操作中需将其作为 `vocab` 回传，否则来自旧读取的 ref 会被拒绝并返回 `409`。
- 只有当 `includeIncomplete=true` 时才会包含 `incompleteViolations`（详情）；`incomplete` 计数始终存在。
- `score` 沿用 native 的 0–100 分制（`100` 减去按严重程度加权的违规计数——每个节点 serious/critical 计 `10`、moderate 计 `5`、minor 计 `2`——下限为 `0`），因此 axe 和 native 的分数在仪表板上保持可比。

该脚本运行在 PinchTab 的**隔离世界**中，因此页面覆盖 `window.axe` 或 DOM 原型都无法改变结果。

## 示例

```bash
# axe engine, JSON
curl "http://localhost:9867/a11y/audit?engine=axe" -H "Authorization: Bearer $TOKEN"

# WCAG 2.0 A only (drops the AA-only color-contrast rule)
curl "http://localhost:9867/a11y/audit?engine=axe&tags=wcag2a" -H "Authorization: Bearer $TOKEN"

# CLI
pinchtab a11y audit --axe
pinchtab a11y audit --axe --tags wcag2a,wcag2aa --json
```

命令行界面 flags：`--axe`（等同于 `--engine axe`）、`--engine`、`--tags`、`--rules`、`--include-incomplete`、`--selector`、`--tab`、`--json`。

MCP：`pinchtab_a11y_audit` 接受 `tabId`、`engine`、`tags`、`rules`、`includeIncomplete` 和 `browser`（无 `selector`）。

覆盖场景见 `tests/e2e/scenarios/api/a11y-axe-basic.sh`、`tests/e2e/scenarios/api/a11y-audit-basic.sh` 和 `tests/e2e/scenarios/cli/a11y-axe-basic.sh`。

## Native 响应

省略 `engine` 时，响应为未改动的 native 报告：

```json
{ "tabId": "...", "score": 88, "findings": [ { "rule": "missing-alt", "severity": "serious", "count": 1, "samples": ["image e5"] } ] }
```
