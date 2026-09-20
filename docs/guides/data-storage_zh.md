# 数据存储指南

PinchTab 在本地磁盘上存储配置、Profile、会话状态和使用日志。本指南说明存储了什么、默认存在哪里，以及哪些路径可以更改。

## PinchTab 存储的内容

| 路径 | 用途 | 如何更改 |
| --- | --- | --- |
| `config.json` | PinchTab 主配置 | `PINCHTAB_CONFIG` 选择该文件 |
| `profiles/<profile>/` | 每个 Profile 的 Chrome 用户数据 | `profiles.baseDir` |
| `sessions.json` | 一个 bridge 实例保存的标签页/会话状态 | `server.stateDir` |
| `activity/events-YYYY-MM-DD.jsonl` | 面向 `/api/activity`、CLI 活动和仪表板活动视图的主要每日请求/活动日志 | `server.stateDir`、`observability.activity.retentionDays` |
| `activity/events-<source>-YYYY-MM-DD.jsonl` | 面向具名来源（如 `dashboard` 或 `orchestrator`）的来源专属每日活动日志 | `server.stateDir`、`observability.activity.retentionDays` |
| `<profile>/.pinchtab-state/config.json` | 由编排器写入的子实例配置 | 对受管实例自动生成 |
| `server.log` | 用 `pinchtab server --background` 启动的服务器的输出 | `server.stateDir` |
| `heapsnapshots/<id>.heapsnapshot` | 来自 `POST /memory/snapshot` 的 V8 堆快照 | `server.stateDir` |
| `logs/daemon.out.log`、`logs/daemon.err.log` | 后台服务（`pinchtab daemon`）的输出，始终位于 `~/.pinchtab/logs/` 下 | 不可配置 |

## CLI 状态

命令行在服务器 `stateDir` 之外维护自己的、按用户存放的临时状态：每个服务器一份当前标签页文件（`current-tab-<host>-<port>`）、ref 词表缓存（`vocab-<host>-<port>`）、活动录制标记（`current-recording`），以及每次运行一次的建议标记（`advisories/`）。它位于：

- `$XDG_STATE_HOME/pinchtab/`（若设置了 `XDG_STATE_HOME`）
- 否则为 `~/.local/state/pinchtab/`

删除它只会重置 CLI 便利性状态；它不保存任何浏览器数据。

## 默认存储位置

PinchTab 把配置和服务器状态放在同一个基础目录下。`server.stateDir` 默认为该目录，`profiles.baseDir` 默认为 `<server.stateDir>/profiles`：

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

## Profiles

Profile 是 PinchTab 在多次启动之间复用的持久浏览器状态。一个 Profile 目录可以包含：

- cookies 和登录会话
- 本地存储和 IndexedDB
- 缓存和历史记录
- Chrome 偏好设置与会话文件

用以下配置 Profile 根目录：

```json
{
  "profiles": {
    "baseDir": "/path/to/profiles",
    "defaultProfile": "default"
  }
}
```

`profiles.defaultProfile` 控制单实例流程使用的默认 Profile 名。在编排器模式下，受管实例仍可用其他 Profile 名启动。

## 主配置文件

主配置文件从以下位置读取：

- 若设置了 `PINCHTAB_CONFIG`，则为其中的路径
- 否则为 `<user-config-dir>/config.json`

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

Bridge 会话恢复数据存为：

```text
<server.stateDir>/sessions.json
```

当启用恢复行为时（`instanceDefaults.tabPolicy.restore`，默认 `false`），此文件用于标签页/会话恢复。

## 活动日志

请求活动按每个 UTC 日存为一个 JSONL 文件：

```text
<server.stateDir>/activity/events-YYYY-MM-DD.jsonl
```

具名来源也有自己的每日文件：

```text
<server.stateDir>/activity/events-<source>-YYYY-MM-DD.jsonl
```

默认情况下，PinchTab 保留 30 天的活动数据，并在记录新活动时清理更早的每日文件。你可以这样更改：

```json
{
  "observability": {
    "activity": {
      "retentionDays": 30,
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

`retentionDays` 控制活动日志在磁盘上的保留时长。`sessionIdleSec` 只控制会话分组。
`events` 控制记录哪些非客户端来源。客户端事件始终被记录。

携带 `X-Agent-Id` 的请求会在活动事件中以该值作为 `agentId` 存储。这正是 `GET /api/activity?agentId=<id>` 这类按代理维度查询以及仪表板代理视图的数据来源。

未过滤的 `GET /api/activity` 读取主 feed。通过传 `source=<name>` 仍可查询来源专属日志。

在编排器模式下，子实例在其 Profile 下获得自己的状态目录：

```text
<profile>/.pinchtab-state/
```

PinchTab 在那里写一个子 `config.json`，以便被启动的实例继承正确的 Profile 路径、状态目录和端口。

受管的子 bridge 会禁用其本地活动记录器。仪表板可见的活动来自处理客户端流量的父服务器，因此编排器管理的子状态目录不应在新运行中累积自己的 `activity/events-*.jsonl` 文件。

Profile 的 `logs` 和 `analytics` 端点是从活动存储派生的，而不是来自单独的分析文件。

## 自定义存储

### 选择不同的配置文件

```bash
export PINCHTAB_CONFIG=/etc/pinchtab/config.json
pinchtab server
```

### 选择不同的 Profile 与状态路径

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

对于 Docker 或其他容器，用一个挂载卷同时持久化配置和 Profile 数据，并把 `PINCHTAB_CONFIG` 指向该卷内的一个文件。通过 `PINCHTAB_CONFIG` 选中的配置必须携带 `server.token`（否则容器必须设置 `PINCHTAB_TOKEN`）；PinchTab 不会向操作员提供的文件里自动生成一个 token。

卷内示例布局：

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

Profile 目录通常包含敏感的浏览器状态：

- cookies
- 会话令牌
- 缓存内容
- 站点数据

推荐做法：

- 不要把 Profile 目录纳入版本控制
- 收紧配置和 Profile 目录的权限
- 为不同的安全上下文使用不同的 Profile

## 清理

删除 PinchTab 数据目录会删除：

- 已保存的 Profile
- 会话恢复数据
- 本地配置

如果你需要保留已登录的浏览器会话，请先备份 Profile 目录。
