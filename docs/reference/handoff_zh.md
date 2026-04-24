# 切换

标记标签页以进行人工干预，检查切换状态，然后在手动步骤完成后恢复自动化。

可用的 CLI 包装器：

```bash
pinchtab tab handoff <tabId> --reason captcha --timeout-ms 120000
pinchtab tab handoff-status <tabId>
pinchtab tab resume <tabId> --status completed
```

API 等价物：

当标签页处于 `paused_handoff` 状态时，操作执行路由会拒绝并返回 `409 tab_paused_handoff`，直到标签页被恢复或可选的切换超时过期。

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

- `POST /tabs/{id}/handoff` 将标签页状态设置为 `paused_handoff`
- `GET /tabs/{id}/handoff` 返回当前切换状态，或在未设置切换时返回 `active`
- 当设置超时时，状态还包括 `expiresAt` 和 `timeoutMs`
- `POST /tabs/{id}/resume` 清除切换状态，并可携带恢复元数据，如 `status` 或 `resolvedData`
- 暂停的标签页会以 `tab_paused_handoff` 拒绝 `/action`、`/actions` 和 `/macro` 请求
- 用于 CAPTCHA、2FA、登录批准或其他仅人工步骤

## 相关页面

- [标签页](./tabs.md)
- [CLI 概览](./cli.md)