# 代理会话

代理会话为自动化代理提供持久、可撤销的认证。不再共享服务器 bearer token，而是每个代理都获得自己的会话 token，映射到特定 `agentId`。

## 概览

- **会话 token**：`ses_<48 个十六进制字符>`（24 个随机字节）——从不原始存储（只持久化 SHA-256 哈希）
- **会话 ID**：`ses_<16 个十六进制字符>`——用于管理的公开标识
- **认证头**：`Authorization: Session <token>`
- **环境变量**：`PINCHTAB_SESSION`——命令行界面自动检测并使用会话认证

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
| `off` | 代理会话禁用，与 `enabled: false` 完全相同 |
| `preferred` | bearer 和会话认证都接受（默认） |
| `required` | **未实现——在配置加载时拒绝。** 它本意为只接受代理的会话认证，但 bearer token 和仪表板 cookie 仍能认证，因此该值被拒绝而非接受后忽略 |

模式值不区分大小写，周围空格被忽略：`"Off"`、`"OFF"` 和 `" off "` 都表示 off。此开关上的一个大小写笔误会按字面意思关闭代理会话，而不是被读作别的什么。

**服务器无法解释的模式会让它停止。** `"required"`、拼写错误或表外任何其他值在加载时被拒绝：进程报告该字段并退出，而不是代你选定一个姿态后启动——仅警告式加载会掩盖的失败正是代理会话继续在提供服务。daemon 单元和自动启动运行的是裸 `pinchtab server`，因此配置文件就是全部输入，你只能在进程退出处看到它。`pinchtab config set sessions.agent.mode off` 在服务器拒绝加载的配置上仍可用，所以修复无需手工编辑。

## 生命周期

1. **创建**——`pinchtab session create --agent-id <id>`（或直接 `POST /sessions`）
2. **使用**——代理每个请求发送 `Authorization: Session ses_...`，或设置 `PINCHTAB_SESSION`
3. **撤销**——`pinchtab session revoke <session-id>`（或 `POST /sessions/{id}/revoke`）
4. **结束**——撤销、过期或修剪一个会话会关闭它在每个实例上创建的标签页，但不包括：此后已被其他调用者使用过的标签页、为人工交接暂停的标签页、以及已锁定的标签页

## 安全

- token 从不以明文记录或持久化
- 用 `crypto/subtle.ConstantTimeCompare` 做 SHA-256 哈希比较
- 空闲超时（默认 30m）和最大生命周期（默认 24h）
- 会话持久化到 `<server.stateDir>/sessions.json`（原子写入）
- 每个会话绑定到特定 agentId 用于活动跟踪

> **⚠️ 仅限受信任、受控环境。** 代理会话面向你已经信任的操作员和自动化：本地机器、专用网络、CI 或其他受控系统。它们不是多租户隔离边界，不应被视为对不受信任用户、不受信任代理或公网暴露是安全的。
>
> 会话管理 API（`/sessions`）对创建、列出、检查操作仍有管理员级权限。任何用服务器 bearer token 或有效仪表板 cookie 认证的调用者都能管理任意代理的会话。会话认证的调用者被挡在仪表板/管理员端点家族之外，但没有显式 grant 的会话默认仍可访问正常的非管理员自动化面。
>
> 在不需要代理会话的不受信任或共享环境中，通过在配置中设 `"enabled": false` 或 `"mode": "off"` 完全禁用它们，以缩小认证面。二者是同一个开关：任一都会在前门拒绝现有会话 token，并对整个 `/sessions` 家族应答 `sessions_disabled`。
>
> `"mode": "required"` 未实现，在配置加载时被拒绝而非接受后忽略：服务器 bearer token 和仪表板 cookie 仍能认证，设它只会让你误以为会话认证是唯一入口。

### 会话 Grants

当会话记录包含显式 `grants` 时，PinchTab 在中间件中强制执行，只允许这些 grant 组覆盖的路由。会话没有显式 grants 时，PinchTab 默认允许正常的非管理员自动化路由，但仍阻止仪表板/管理员端点家族，如 config、仪表板事件流、会话管理、profile 管理、实例管理和缓存管理。

内置 grant 组为：`browse`、`network`、`media`、`cookies`、`clipboard`、`evaluate`、`storage`、`console`、`solve`、`tasks` 和 `activity`。

在创建会话时设置它们，可在 API 或命令行界面上：

```bash
curl -X POST "$BASE/sessions" -H "Authorization: Bearer $PINCHTAB_TOKEN" \
  -d '{"agentId":"reader","grants":["browse","network"]}'

pinchtab session create --agent-id reader --grant browse --grant network
```

不认识的 grant 以 `invalid_grant` 拒绝，且不创建会话，因此一个笔误不会让你拿着一个你以为受限的凭据。`"*"` 是"不受限"的显式写法，与省略该字段同义。grants 由 `pinchtab session list` 和 `pinchtab session info` 报告，且跨重启保留。没有路由可改已有会话的 grants：用你想要的范围新建一个并撤销旧的。

因范围被拒的请求应答 `session_scope_forbidden`，提示中会命名该会话持有的 grants 以及本应覆盖该路由的 grant。同一代码也应答管理员动词（此时根本没有 grant 适用）——提示会说明是二者中的哪一个触发，因为补救不同：换一个会话，或换服务器 token。

**Grants 只收窄，从不放宽。** 一个 grant 授权会话访问一组路由，但每个服务器级门控仍叠加生效：`evaluate` 不会重新开启 `security.allowEvaluate`，`cookies` 不会重新开启 `security.allowCookies`，没有 grant 能到达管理员路由。

该默认值是给受信任自动化的便利，不是沙箱。如果你需要代理或租户间的硬隔离，请使用单独的 PinchTab 实例。

## 命令行界面用法

```bash
# Create a new session (prints the session token to stdout; use --json for a full JSON object)
pinchtab session create --agent-id agent-1

# Create one limited to a capability group (repeatable; see Session Grants)
pinchtab session create --agent-id reader --grant browse
export PINCHTAB_SESSION=$(pinchtab session create --agent-id agent-1)

# CLI automatically uses session auth when PINCHTAB_SESSION is set
pinchtab snap

# Inspect the current session (uses PINCHTAB_SESSION)
pinchtab session info

# List all sessions on the server (server bearer token required)
pinchtab session list

# Revoke a session by id
pinchtab session revoke ses_abc123def456
```

## API 端点

| 路由 | 用途 |
| --- | --- |
| `POST /sessions` | 创建会话；主体 `agentId`（必填），可选 `label`、`grants` 和 `browser`（未知时 `invalid_browser`） |
| `GET /sessions` | 列出会话 |
| `GET /sessions/me` | 认证本请求的会话 |
| `GET /sessions/{id}` | 单个会话 |
| `POST /sessions/{id}/revoke` | 撤销会话 |

`pinchtab session create` 也接受 `--label`。完整 API 参考见 [endpoints.md](../endpoints.md)。
