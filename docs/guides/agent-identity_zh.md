# 代理身份

PinchTab 提供三个级别的代理标识，从简单到完全管理。选择适合您设置的级别。

## 服务器令牌

每个 PinchTab 服务器都在 `server.token` 中配置了一个承载令牌。这是基线认证方法 —— 它证明调用者有权使用服务器，但不说明是*哪个*代理正在发出请求。

```bash
pinchtab --token "your-server-token" nav https://example.com
```

或通过环境变量：

```bash
export PINCHTAB_TOKEN=your-server-token
pinchtab nav https://example.com
```

**何时使用：** 单代理设置、快速脚本编写，或当您不需要每个代理的跟踪时。

**限制：** 所有请求在活动 feed 中看起来都一样 —— 无法区分哪个代理做了什么。

## 代理 ID

添加代理 ID 会为每个请求添加一个名称标签。这会显示在活动 feed 和仪表板的代理页面中。服务器仍然通过承载令牌进行认证，但现在每个请求都携带一个身份。

```bash
pinchtab --agent-id bosch nav https://example.com
```

或通过环境变量：

```bash
export PINCHTAB_AGENT_ID=bosch
pinchtab nav https://example.com
```

`X-Agent-Id` 头会随每个请求一起发送。不需要服务器端设置 —— 任何字符串都可以。

**何时使用：** 多个代理共享一个服务器，您希望看到谁做了什么，但不需要会话管理。

**限制：** 没有撤销、没有空闲跟踪、没有标签。代理 ID 是自我声明的 —— 任何调用者都可以声称任何身份。

## 代理会话

> **⚠️ 安全注意：** 代理会话设计用于**受信任的、受控的环境** —— 本地机器、私有网络、CI 以及所有代理都在您控制下的设置。不要将会话管理 API (`/sessions`) 暴露给公共互联网。任何经过认证的调用者（承载令牌或仪表板 cookie）都可以为任何代理创建、列出和检查会话。会话认证的调用者被阻止访问仪表板/管理端点系列，但会话仍然不是多租户隔离边界。

会话是完整的身份解决方案。每个会话是一个可撤销的、服务器管理的令牌，与特定的代理 ID 绑定。会话提供：

- **标签** —— 人类可读的名称，如 "研究任务" 或 "每日抓取"
- **活动分组** —— 会话内的所有请求在仪表板中分组
- **空闲超时** —— 会话在 12 小时不活动后过期（可配置）
- **最大生命周期** —— 24 小时后硬过期（可配置）
- **撤销** —— 无需轮换服务器令牌即可终止会话

### 启用会话

添加到您的 `config.json`：

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
| `preferred` | 接受承载和会话认证（启用时的默认值） |
| `required` | 代理只能使用会话认证 |

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

`sessionToken` 只返回一次。存储它 —— PinchTab 只持久化哈希。

### 使用会话

```bash
export PINCHTAB_SESSION=ses_1138f72e77f23c49...
pinchtab nav https://example.com
pinchtab snap -i -c
pinchtab click e5
```

或直接传递头：

```bash
curl -X POST http://localhost:9867/navigate \
  -H "Authorization: Session ses_1138f72e77f23c49..." \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}'
```

无需设置 `--agent-id` —— 会话携带代理身份。

### 管理会话

```bash
# 列出所有会话
curl http://localhost:9867/sessions \
  -H "Authorization: Bearer $PINCHTAB_TOKEN"

# 撤销
curl -X POST http://localhost:9867/sessions/ses_e6ac8132fe7e7016/revoke \
  -H "Authorization: Bearer $PINCHTAB_TOKEN"
```

### 配置

| 设置 | 默认值 | 描述 |
|---------|---------|-------------|
| `sessions.agent.enabled` | `false` | 启用代理会话 |
| `sessions.agent.mode` | `preferred` | 认证模式：`off`、`preferred`、`required` |
| `sessions.agent.idleTimeoutSec` | `43200` (12h) | 会话在这么多秒不活动后过期 |
| `sessions.agent.maxLifetimeSec` | `86400` (24h) | 会话硬过期 |

如果会话记录携带明确的授权，这些授权会缩小会话可以调用的端点组。如果会话没有明确的授权，默认情况下它可以使用正常的非管理自动化 API，而仪表板/管理路由仍然被阻止。这个默认值仅用于受信任的自动化。

## 选择正确的级别

| 场景 | 建议 |
|----------|----------------|
| 一个代理，仅限本地 | 服务器令牌足够 |
| 多个代理，需要归因 | 添加 `--agent-id` 或 `PINCHTAB_AGENT_ID` |
| 生产多代理，需要撤销 | 使用代理会话 |
| 共享服务器，不受信任的代理 | 运行单独的 PinchTab 实例；会话不足以隔离 |