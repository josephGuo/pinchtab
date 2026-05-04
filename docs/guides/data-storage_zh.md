# 数据存储指南

PinchTab 在本地磁盘上存储配置、配置文件、会话状态和使用日志。本指南描述了存储的内容、默认存储位置以及可以更改的路径。

## PinchTab 存储的内容

| 路径 | 用途 | 如何更改 |
| --- | --- | --- |
| `config.json` | PinchTab 主配置 | `PINCHTAB_CONFIG` 选择文件 |
| `profiles/<profile>/` | 每个配置文件的 Chrome 用户数据 | `profiles.baseDir` |
| `sessions.json` | 桥接实例的保存标签页/会话状态 | `server.stateDir` |
| `activity/events-YYYY-MM-DD.jsonl` | 用于 `/api/activity`、命令行界面 活动和仪表板活动视图的主要每日请求/活动日志 | `server.stateDir`、`observability.activity.retentionDays` |
| `activity/events-<source>-YYYY-MM-DD.jsonl` | 特定来源的每日活动日志，用于命名来源如 `dashboard` 或 `orchestrator` | `server.stateDir`、`observability.activity.retentionDays` |
| `<profile>/.pinchtab-state/config.json` | 由编排器写入的子实例配置 | 为管理实例自动生成 |

## 默认存储位置

PinchTab 使用操作系统配置目录：

| 操作系统 | 默认基础目录 |
| --- | --- |
| Linux | `~/.pinchtab/` |
| macOS | `~/.pinchtab/` |
| Windows | `%APPDATA%\pinchtab\` |

典型布局：

```text
pinchtab/
├── config.json
├── activity/
│   └── events-2026-03-16.jsonl
├── sessions.json
└── profiles/
    └── default/
```

## 平台默认值

在 macOS 和 Linux 上，`~/.pinchtab/` 是默认基础目录。

在 Windows 上，PinchTab 使用 `%APPDATA%\pinchtab\` 下的操作系统原生配置目录。

## 配置文件

配置文件是 PinchTab 在启动之间重用的持久浏览器状态。配置文件目录可以包含：

- cookies 和登录会话
- 本地存储和 IndexedDB
- 缓存和历史记录
- Chrome 首选项和会话文件

使用以下配置文件根目录：

```json
{
  "profiles": {
    "baseDir": "/path/to/profiles",
    "defaultProfile": "default"
  }
}
```

`profiles.defaultProfile` 控制单实例流程使用的默认配置文件名称。在编排器模式下，管理实例仍然可以使用其他配置文件名称启动。

## 配置文件

主配置文件从以下位置读取：

- 如果设置了 `PINCHTAB_CONFIG`，则从该路径读取
- 否则从 `<user-config-dir>/config.json` 读取

示例：

```json
{
  "server": {
    "port": "9867",
    "stateDir": "/var/lib/pinchtab/state"
  },
  "profiles": {
    "baseDir": "/var/lib/pinchtab/profiles",
    "defaultProfile": "default"
  }
}
```

## 会话状态

桥接会话恢复数据存储为：

```text
<server.stateDir>/sessions.json
```

当启用恢复行为时，此文件用于标签页/会话恢复。

## 活动日志

请求活动存储为每个 UTC 日一个 JSONL 文件：

```text
<server.stateDir>/activity/events-YYYY-MM-DD.jsonl
```

命名来源也有自己的每日文件：

```text
<server.stateDir>/activity/events-<source>-YYYY-MM-DD.jsonl
```

默认情况下，PinchTab 保留 1 天的活动数据，并在记录新活动时修剪较旧的每日文件。您可以通过以下方式更改：

```json
{
  "observability": {
    "activity": {
      "retentionDays": 1,
      "sessionIdleSec": 1800,
      "events": {
        "dashboard": false,
        "server": false,
        "bridge": false,
        "orchestrator": false,
        "scheduler": false,
        "mcp": false,
        "other": false
      }
    }
  }
}
```

`retentionDays` 控制活动日志的磁盘保留。`sessionIdleSec` 仅控制会话分组。
`events` 控制记录哪些非客户端来源。客户端事件始终被记录。

携带 `X-Agent-Id` 的请求在活动事件中以该值作为 `agentId` 存储。这是支持代理范围查询（如 `GET /api/activity?agentId=<id>` 和仪表板代理视图）的原因。

未过滤的 `GET /api/activity` 读取主要 feed。通过传递 `source=<name>` 仍可查询特定来源的日志。

在编排器模式下，子实例在配置文件下获得自己的状态目录：

```text
<profile>/.pinchtab-state/
```

PinchTab 在那里写入子 `config.json`，以便启动的实例可以继承正确的配置文件路径、状态目录和端口。

管理的子桥接禁用其本地活动记录器。仪表板可见的活动来自处理客户端流量的父服务器，因此编排器管理的子状态目录不应为新运行累积自己的 `activity/events-*.jsonl` 文件。

配置文件 `logs` 和 `analytics` 端点从活动存储派生，而不是从单独的分析文件派生。

## 自定义存储

### 选择不同的配置文件

```bash
export PINCHTAB_CONFIG=/etc/pinchtab/config.json
pinchtab
```

### 选择不同的配置文件和状态路径

```json
{
  "server": {
    "stateDir": "/srv/pinchtab/state"
  },
  "profiles": {
    "baseDir": "/srv/pinchtab/profiles",
    "defaultProfile": "default"
  }
}
```

## 容器使用

对于 Docker 或其他容器，使用挂载卷持久化配置和配置文件数据，并将 `PINCHTAB_CONFIG` 指向该卷内的文件。

卷内的示例布局：

```text
/data/
├── config.json
├── state/
└── profiles/
```

然后设置：

```json
{
  "server": {
    "stateDir": "/data/state"
  },
  "profiles": {
    "baseDir": "/data/profiles"
  }
}
```

## 安全注意事项

配置文件目录通常包含敏感的浏览器状态：

- cookies
- 会话令牌
- 缓存内容
- 站点数据

推荐做法：

- 保持配置文件目录不在版本控制中
- 限制配置和配置文件目录的权限
- 为不同的安全上下文使用单独的配置文件

## 清理

删除 PinchTab 数据目录会删除：

- 保存的配置文件
- 会话恢复数据
- 本地配置

如果需要保留登录的浏览器会话，请先备份配置文件目录。