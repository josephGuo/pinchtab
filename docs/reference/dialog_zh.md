# Dialog（对话框）

接受或关闭某标签页上打开的 JavaScript 对话框（`alert`、`confirm`、`prompt`）。

```bash
curl -X POST http://localhost:9867/dialog \
  -H "Content-Type: application/json" \
  -d '{"action":"accept"}'
# CLI Alternative
pinchtab dialog accept
# Response (use --json for full JSON)
OK
```

JSON 响应：

```json
{"type": "alert", "message": "Hello", "handled": true}
```

当 PinchTab 错过了对话框打开事件但仍回答了对话框时，`type` 为 `unknown`，`message` 为空。

## 命令行界面

| 命令 | 说明 |
|---------|-------------|
| `pinchtab dialog accept [text]` | 接受（确定）；`text` 为 prompt 的回复 |
| `pinchtab dialog dismiss` | 关闭（取消） |

两者均接受 `--tab <id>` 和 `--json`。

## HTTP

`POST /dialog` 或 `POST /tabs/{id}/dialog`：

| 字段 | 说明 |
|-------|-------------|
| `action` | 必填：`accept` 或 `dismiss` |
| `text` | prompt 的回复，与 `accept` 一起使用 |
| `tabId` | 目标标签页（默认：当前标签页） |

错误：缺少或未知的 `action`，或标签页上无对话框打开时返回 `400`（`no dialog open on tab <id>`）；标签页因交接而暂停时返回 `409 tab_paused_handoff`。

## 被对话框阻塞的标签页

对话框打开期间，驱动页面的路由——navigate、后退/前进/刷新、wait、find、evaluate、元素读取（如 `attr` 和 `count`）、`/action` 和 `/actions`、存储以及状态保存——会立即返回 `409 dialog_blocked` 而非挂起。错误 `details` 携带 `hint`、`remedy`、`tabId`、`dialogType` 和 `dialogMessage`。批次中某一步骤打开了对话框时，即使 `stopOnError:false`，该批次也会因携带代码 `dialog_blocked` 的失败步骤而停止。用 `pinchtab dialog accept|dismiss` 回答对话框，或在打开对话框的点击操作上传入 `--dialog-action accept|dismiss`（API 为 `dialogAction`）——见 [Click](./click.md)。

由 `tests/e2e/scenarios/api/dialog-guard-basic.sh`、`tests/e2e/scenarios/cli/dialog-guard-basic.sh` 和 `tests/e2e/scenarios/api/actions-extended.sh` 测试覆盖。

MCP：`pinchtab_dialog`，参数 `action`（必填）、`text`、`tabId`。
