# 健康状态

检查服务器状态和可用性。

## 桥接模式

```bash
curl http://localhost:9867/health
# 响应: {"status":"ok","tabs":1}

# 命令行界面 替代方案（默认人类可读）
pinchtab health
# 输出: ok

pinchtab health --json              # 完整 JSON 响应
```

桥接模式健康状态还可能包括：

- `crashLogs`
- `failures`
- `crashes`

在错误情况下，它会返回 `503` 状态码，`status: "error"` 和一个 `reason`。

## 服务器模式（仪表板）

在完整服务器模式下，`/health` 返回仪表板健康状态包：

```bash
curl http://localhost:9867/health
# 响应
{
  "status": "ok",
  "mode": "dashboard",
  "version": "0.8.0",
  "uptime": 12345,
  "authRequired": true,
  "profiles": 1,
  "instances": 1,
  "defaultInstance": {
    "id": "inst_abc12345",
    "status": "running"
  },
  "agents": 0,
  "restartRequired": false
}
```

| 字段 | 描述 |
| --- | --- |
| `status` | 服务器健康时为 `ok` |
| `mode` | 服务器模式下为 `dashboard` |
| `version` | PinchTab 版本 |
| `uptime` | 服务器启动后的毫秒数 |
| `authRequired` | 配置服务器令牌时为 `true` |
| `profiles` | 已配置的配置文件数量 |
| `instances` | 管理的实例数量 |
| `defaultInstance` | 第一个管理实例信息（如果存在） |
| `agents` | 连接的代理数量 |
| `restartRequired` | 基于文件的配置更改需要重启时为 `true` |
| `restartReasons` | 需要重启时的重启原因列表 |

注意：

- 当至少存在一个实例时，会出现 `defaultInstance`
- 当你想确认 Chrome 已准备就绪时，使用 `defaultInstance.status == "running"`
- 像 `always-on` 这样的策略可以在启动时自动创建实例

## 相关页面

- [标签页](./tabs.md)
- [导航](./navigate.md)
- [策略](./strategies.md)