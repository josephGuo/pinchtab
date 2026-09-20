# 代理身份

PinchTab 提供三个级别的代理标识，从简单到完全管理。选择适合你部署的级别。

## 服务器令牌

每个 PinchTab 服务器都在 `server.token` 中配置了一个 bearer 令牌。这是基线认证方法——它证明调用者有权使用服务器，但不说明是*哪个*代理正在发出请求。

CLI 会自动从本地配置中读取它。要显式提供它（例如针对另一台服务器），请使用环境变量——没有 `--token` 这个 flag：

```bash
export PINCHTAB_TOKEN=your-server-token
pinchtab nav https://example.com
```

**何时使用：** 单代理部署、快速脚本编写，或当你不需要按代理分别追踪时。

**局限：** 所有请求在活动 feed 中看起来都一样——无法区分哪个代理做了什么。

## 代理 ID

添加代理 ID 会给每个请求打上一个名称标签。这会显示在活动 feed 和仪表板的代理页面中。服务器仍然通过 bearer 令牌认证，但现在每个请求都携带一个身份。

```bash
pinchtab --agent-id bosch nav https://example.com
```

或通过环境变量：

```bash
export PINCHTAB_AGENT_ID=bosch
pinchtab nav https://example.com
```

`X-Agent-Id` 头会随每个请求一起发送。无需服务器端设置——任何字符串都可以。

当一个请求带有代理 ID 但没有代理会话时，PinchTab 也会把该代理的当前标签页与其他代理分开维护。在多步流程中你可以省略 `tabId`，而不会抢走另一个代理的当前标签页。

**何时使用：** 多个代理共享一个服务器，你希望看到谁做了什么，但不需要会话管理。

**局限：** 没有撤销、没有空闲追踪、没有标签。代理 ID 是自我声明的——任何调用者都可以声称任何身份。

## 代理会话

> **⚠️ 安全提示：** 代理会话设计用于**受信任的、受控的环境**——本地机器、私有网络、CI，以及所有代理都在你控制之下的部署。不要把会话管理 API（`/sessions`）暴露给公网。任何通过认证的调用者（bearer 令牌或仪表板 cookie）都可以为任何代理创建、列出和查看会话。经过会话认证的调用者会被禁止访问仪表板/管理端点族，但会话仍然不是多租户隔离边界。

会话是完整的身份解决方案。每个会话都是一个可撤销的、由服务器管理的令牌，与特定的代理 ID 绑定。会话提供：

- **标签**——人类可读的名称，例如「研究任务」或「每日抓取」
- **活动分组**——一个会话内的所有请求在仪表板中归为一组
- **空闲超时**——会话在 30 分钟不活动后过期（可配置）
- **最大生命周期**——24 小时后硬过期（可配置）
- **撤销**——无需轮换服务器令牌即可终止一个会话

### 启用会话

添加到你的 `config.json`：

```json
{
  "sessions": {
    "agent": {
      "enabled": true,
      "mode": "preferred"
    }
  }
}
```

模式：

| 模式 | 行为 |
|------|----------|
| `off` | 代理会话禁用 |
| `preferred` | 同时接受 bearer 与会话认证（启用时的默认值） |
| `required` | 未实现；在加载配置时被拒绝（参见[配置](#配置)） |

### 创建会话

```bash
curl -X POST http://localhost:9867/sessions \
  -H "Authorization: Bearer $PINCHTAB_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"agentId": "bosch", "label": "research task"}'
```

响应：

```json
{
  "id": "ses_e6ac8132fe7e7016",
  "agentId": "bosch",
  "label": "research task",
  "sessionToken": "ses_1138f72e77f23c49...",
  "status": "active"
}
```

`sessionToken` 只返回一次。请妥善保存——PinchTab 只持久化它的哈希。

### 使用会话

```bash
export PINCHTAB_SESSION=ses_1138f72e77f23c49...
pinchtab nav https://example.com
pinchtab snap -i -c
pinchtab click e5
```

或直接传入请求头：

```bash
curl -X POST http://localhost:9867/navigate \
  -H "Authorization: Session ses_1138f72e77f23c49..." \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}'
```

无需设置 `--agent-id`——会话本身携带代理身份。

当前标签页状态是按会话作用域隔离的。同一个 `agentId` 的两个会话各自维护独立的当前标签页，而且会话作用域优先于任何 `X-Agent-Id` 头。

### 管理会话

```bash
# List all sessions
curl http://localhost:9867/sessions \
  -H "Authorization: Bearer $PINCHTAB_TOKEN"

# Revoke
curl -X POST http://localhost:9867/sessions/ses_e6ac8132fe7e7016/revoke \
  -H "Authorization: Bearer $PINCHTAB_TOKEN"
```

### 配置

| 设置 | 默认值 | 描述 |
|---------|---------|-------------|
| `sessions.agent.enabled` | `true` | 启用代理会话 |
| `sessions.agent.mode` | `preferred` | 认证模式。`preferred` 在服务器令牌之外同时提供代理会话；`off` 的禁用效果与 `enabled: false` 完全相同。`required` **未实现**，在加载配置时被拒绝——bearer 令牌和仪表板 cookie 仍然可以认证，因此该值无法实现它名义上所要的「仅会话认证」 |
| `sessions.agent.idleTimeoutSec` | `1800`（30 分钟） | 会话在这么多秒不活动后过期 |
| `sessions.agent.maxLifetimeSec` | `86400`（24 小时） | 会话硬过期 |

如果一条会话记录带有显式授权（grant），这些授权会收窄该会话可调用的端点组范围。授权在创建会话时设置——`pinchtab session create --agent-id <id> --grant browse`，或在 `POST /sessions` 上用 `grants` 字段——并由 `session list` 和 `session info` 报告。如果会话没有显式授权，它默认可以使用常规的非管理自动化 API，而仪表板/管理路由仍然被阻止。这个默认值仅用于受信任的自动化。授权只会收窄、不会放宽：服务器级别的能力门槛仍然在其之上生效。

## 选择正确的级别

| 场景 | 建议 |
|----------|----------------|
| 一个代理，仅限本地 | 服务器令牌足够 |
| 多个代理，需要归因 | 添加 `--agent-id` 或 `PINCHTAB_AGENT_ID` |
| 生产多代理，需要撤销 | 使用代理会话 |
| 共享服务器，不受信任的代理 | 运行独立的 PinchTab 实例；会话不足以构成隔离 |
