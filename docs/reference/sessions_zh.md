# 代理会话

代理会话为自动化代理提供持久、可撤销的身份验证。每个代理都获得自己的会话令牌，而不是共享服务器承载令牌，该令牌映射到特定的 `agentId`。

## 概览

- **会话令牌**：`ses_<48 个十六进制字符>` — 高熵，从不原始存储（仅持久化 SHA-256 哈希）
- **会话 ID**：`ses_<16 个十六进制字符>` — 用于管理的公共标识符
- **认证头**：`Authorization: Session <token>`
- **环境变量**：`PINCHTAB_SESSION` — 命令行界面 自动检测并使用会话认证

## 配置

在 `config.json` 中：

```json
{
  "sessions": {
    "agent": {
      "enabled": true,
      "mode": "preferred",
      "idleTimeoutSec": 1800,
      "maxLifetimeSec": 86400
    }
  }
}
```

### 模式

| 模式 | 行为 |
|------|----------|
| `off` | 代理会话禁用 |
| `preferred` | 接受承载和会话认证（默认） |
| `required` | 代理仅接受会话认证 |

## 生命周期

1. **创建** — 通过仪表板 API：`POST /sessions`
2. **使用** — 代理在每个请求中发送 `Authorization: Session ses_...`
3. **撤销** — 永久禁用：`POST /sessions/{id}/revoke`

## 安全性

- 令牌从不以明文形式记录或持久化
- 使用 `crypto/subtle.ConstantTimeCompare` 进行 SHA-256 哈希比较
- 空闲超时（默认 30 分钟）和最大生命周期（默认 24 小时）
- 会话持久化到 `agent-sessions.json`（原子写入）
- 每个会话绑定到特定的 agentId 以进行活动跟踪

> **⚠️ 仅在受信任、受控的环境中使用。** 代理会话适用于您已经信任的操作员和自动化：本地机器、专用网络、CI 或其他受控系统。它们不是多租户隔离边界，不应被视为对不受信任的用户、不受信任的代理或公共互联网暴露是安全的。
>
> 会话管理 API (`/sessions`) 仍然具有创建、列出和检查操作的管理员风格权限。任何使用服务器承载令牌或有效仪表板 cookie 认证的调用者都可以管理任何代理的会话。会话认证的调用者被阻止访问仪表板/管理员端点系列，但默认情况下，没有明确授权的会话仍然可以访问正常的非管理员自动化界面。
>
> 在不需要代理会话的不受信任或共享环境中，通过在配置中设置 `"enabled": false` 或 `"mode": "off"` 完全禁用它们，以减少认证表面。

### 会话授权

当会话记录包含明确的 `grants` 时，PinchTab 在中间件中强制执行它们，只允许那些授权组覆盖的路由。当会话没有明确授权时，PinchTab 默认允许正常的非管理员自动化路由，但仍然阻止仪表板/管理员端点系列，如配置、仪表板事件流、会话管理、配置文件管理、实例管理和缓存管理。

内置的授权组有：`browse`、`network`、`media`、`cookies`、`clipboard`、`evaluate`、`storage`、`console`、`solve`、`tasks` 和 `activity`。

该默认值是为受信任的自动化提供的便利，而不是沙盒。如果您需要代理或租户之间的硬隔离，请使用单独的 PinchTab 实例。

## 命令行界面 使用

```bash
# 设置会话令牌
export PINCHTAB_SESSION=ses_abc123...

# 命令行界面 自动使用会话认证
pinchtab snap

# 检查会话信息
pinchtab session info
```

## API 端点

有关完整的 API 参考，请参阅 [endpoints.md](../endpoints.md)。