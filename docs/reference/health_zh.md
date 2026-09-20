# 健康状态（Health）

检查服务器状态与可用性。

`/health` 刻意**不**是跨模式比较的端点：在服务端模式下它是编排器自己的封装，而编排器拥有 bridge 所没有的事实（`instances`、`profiles`、`defaultInstance`）。失败与崩溃遥测由两种模式在 [`/metrics`](./metrics.md) 上以相同形态提供——用 bridge 复现对照服务端模式问题时请用它。

## Bridge 模式

```bash
curl http://localhost:9867/health
# Response: {"status":"ok","tabs":1}

# CLI Alternative (human-readable by default)
pinchtab health
# Output: ok (followed by next-step hints)

pinchtab health --json              # Full JSON response
```

`tabs` 统计每个浏览器目标，包括 `GET /tabs` 隐藏的那些（如 `about:blank`）。Bridge 模式健康状态还报告 `version`，并可能包含：

- `crashLogs`
- `failures`
- `crashes`

错误情况下返回 `503`，带 `status: "error"`、一个 `reason` 和错误 `code`（`bridge_unavailable`、`browser_init_failed`、`list_targets_failed`）。浏览器重启期间返回 `503` `browser_draining`，带 `status: "draining"`、`retryAfterSeconds` 和 `Retry-After` 头。

## 服务端模式（仪表板）

完整服务端模式下，`/health` 返回仪表板健康封装：

```bash
curl http://localhost:9867/health
# Response
{
  "status": "ok",
  "mode": "dashboard",
  "version": "0.8.0",
  "uptime": 12345,
  "authRequired": true,
  "profiles": 1,
  "temporaryProfiles": 0,
  "quarantinedProfiles": 0,
  "instances": 1,
  "defaultInstance": {
    "id": "inst_abc12345",
    "status": "running",
    "responsiveness": "responsive"
  },
  "agents": 0,
  "restartRequired": false
}
```

| 字段 | 描述 |
| --- | --- |
| `status` | 服务器健康时为 `ok`；至少一个实例无响应时为 `degraded` |
| `mode` | 服务端模式下为 `dashboard` |
| `version` | PinchTab 版本 |
| `uptime` | 服务器启动至今的毫秒数 |
| `authRequired` | 配置了服务器 token 时为 `true` |
| `profiles` | `GET /profiles` 列出的 profile 数（不含临时和隔离的） |
| `temporaryProfiles` | 临时自动生成的实例 profile 数 |
| `quarantinedProfiles` | 隔离的 profile 数 |
| `instances` | 受管实例数 |
| `defaultInstance` | 简写（非 `/instances`）路由使用的实例：`id`、`status`、`responsiveness`。见下方说明 |
| `agents` | 已连接的代理数 |
| `restartRequired` | 基于文件的配置变更需要重启时为 `true` |
| `restartReasons` | 需要重启时的重启原因列表 |
| `security` | 仅对 bearer token 或仪表板 cookie 认证的请求：`level`、`bind`、`allowedDomains`、`idpiEnabled`、`enabledSensitiveEndpoints`、`guardsDown` |
| `crashes` | 任一实例的浏览器崩溃后出现：`total` 和 `recent`，与 bridge `/health` 携带的同一个块，每个事件标出其 `instanceId` |
| `unresponsiveInstances` | 任一实例能应答自己的 `/health` 但不应答其浏览器路由时出现：那些实例的 id |

说明：

- `defaultInstance` 在至少存在一个实例时出现。它是默认浏览器目标中最早启动的运行中实例；若没有，则回退到任意状态中最早启动的实例
- 想确认 Chrome 已就绪时用 `defaultInstance.status == "running"`
- `always-on` 等策略可在启动时自动创建实例
- 浏览器崩溃时 `status` 不变为降级：实例会被重新拉起并继续服务。崩溃已成历史，因此它作为 `crashes` 排在 `status` 旁边；一个崩溃后又重启的实例与从未崩溃的实例仅靠那个 key 区分。死掉的浏览器持有的每个标签页都没了，调用其中之一以 `404` 应答，代码 `browser_crashed` 并带 `hint` 说明
- 实例无响应时 `status` 会降级为 `degraded`，因为该状况是当前的：每次 `/health` 都在其崩溃抓取的同时，用很短的预算启动一次对每个运行中实例 `/tabs` 的探测，或并入已在途的那一次，最多等它 250ms；耗时更长的探测在后台完成，因此本次 `/health` 报告上次记录的结果，下一次报告新结果。某个实例的 `/health` 应答而 `/tabs` 超时则被列入 `unresponsiveInstances`。它自己的 `status` 保持 `running`，没有任何东西重启它；该字段是报告而非补救。见 `GET /instances` 上的 `responsiveness`

## 相关页面

- [Tabs](./tabs.md)
- [Navigate](./navigate.md)
- [Strategies](./strategies.md)
