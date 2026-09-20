# 交接（Handoff）

将某标签页标记为人工介入，检查交接状态，然后在手动步骤完成后恢复自动化。

可用的命令行界面包装器：

```bash
pinchtab tab handoff <tabId> --reason captcha --timeout-ms 120000
pinchtab tab handoff-status <tabId>
pinchtab tab resume <tabId> --status completed
```

同样的命令也存在于顶层（`pinchtab handoff`、`pinchtab handoff-status`、`pinchtab resume`）；省略 `<tabId>` 时各自默认针对当前标签页。`--reason` 默认为 `manual_handoff`。

MCP：`pinchtab_handoff`（`tabId` 必填，`reason`、`timeoutMs`）、`pinchtab_handoff_status`（`tabId`）、`pinchtab_resume`（`tabId`、`status`）。

API 等价物：

当标签页处于 `paused_handoff` 时，动作执行路由以 `409 tab_paused_handoff` 拒绝，直到标签页被恢复或可选的交接超时到期。

```bash
curl -X POST http://localhost:9867/tabs/<tabId>/handoff \
  -H "Content-Type: application/json" \
  -d '{"reason":"captcha","timeoutMs":120000}'

curl http://localhost:9867/tabs/<tabId>/handoff

curl -X POST http://localhost:9867/tabs/<tabId>/resume \
  -H "Content-Type: application/json" \
  -d '{"status":"completed","resolvedData":{"operator":"human"}}'
```

注意：

- `POST /tabs/{id}/handoff` 将标签页状态设为 `paused_handoff`
- `GET /tabs/{id}/handoff` 返回当前交接状态，未设置交接时返回 `active`
- 设置了超时时，状态还包含 `expiresAt` 和 `timeoutMs`
- `POST /tabs/{id}/resume` 清除交接状态，并可携带 `status` 或 `resolvedData` 等恢复元数据
- 暂停的标签页以 `409 tab_paused_handoff` 拒绝变更类路由——`/action`、`/navigate`、`/back`、`/forward`、`/reload`、`/evaluate`、`/dialog`、`/cookies`、`/storage`、`/upload`、`/download`、`/solve`、`/emulation/*`、`/network/route`、`/fingerprint/rotate`、`/state/load`——`/actions` 和 `/macro` 应答 `200` 并在每个步骤上报告 `tab_paused_handoff`
- `POST /tabs/{id}/handoff` 返回 `{tabId, status, reason, timeoutMs, hint, remedy}`（带超时时另加 `expiresAt`）；负的 `timeoutMs` 返回 `400`，租借给其他所有者的标签页返回 `423 tab_locked`
- `POST /tabs/{id}/resume` 返回 `{tabId, status:"active", resumeStatus, resolvedData}`
- 用于 CAPTCHA、2FA、登录批准或其他仅人工步骤

## 相关页面

- [Tabs](./tabs.md)
- [CLI Overview](./cli.md)
