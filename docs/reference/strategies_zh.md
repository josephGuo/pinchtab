# 策略与分配（Strategies And Allocation）

PinchTab 有两个独立的多实例控制：

- `multiInstance.strategy`
- `multiInstance.allocationPolicy`

它们解决不同的问题：

```text
strategy          = what routes PinchTab exposes and how shorthand requests behave
allocationPolicy  = which running instance gets picked when PinchTab must choose one
```

## Strategy

当前实现中的有效 strategy：

- `always-on` - 默认
- `simple`
- `explicit`
- `simple-autorestart`
- `no-instance`

### `simple`

行为：

- 注册完整的编排器 API
- 保留简写路由，如 `/snapshot`、`/text`、`/navigate` 和 `/tabs`
- 简写请求到达且无实例运行时，PinchTab 自动启动一个受管实例并等待其变健康

最适合：

- 本地开发
- 单用户自动化
- "只要让浏览器服务可用"的设置

### `explicit`

`explicit` 也暴露编排器 API 和简写路由，但它不在简写请求时自动启动。

行为：

- 你用 `/instances/start`、`/instances/launch` 或 `/profiles/{id}/start` 显式启动实例
- 简写路由仅当已存在运行实例时代理到（默认浏览器目标）最早启动的运行实例
- 没有运行实例时，简写路由返回 `503 no running instances — launch one from the Profiles tab`，而不是替你启动浏览器；`GET /tabs` 返回空列表

最适合：

- 受控的多实例环境
- 应刻意指定实例的代理
- 隐藏的自动启动会令人意外的部署

### `always-on`

`always-on` 行为像一个受管单实例服务，应在 PinchTab 进程整个生命周期内保持在线。

行为：

- strategy 启动时启动一个受管实例
- 暴露与 `simple` 相同的简写路由
- 监视该受管实例，意外退出后持续重启，直到达到配置的重启上限
- 暴露 `GET /always-on/status` 给出当前受管实例状态：`instanceId`、`restartCount`、`maxRestarts`、`lastCrash`、`lastStart` 和 `status`（`running`、`restarting`、`crashed` 或 `stopped`）

最适合：

- daemon 式本地服务
- 期望始终有一个默认浏览器的代理宿主
- 启动可用性重要但仍要有界故障策略的设置

### `simple-autorestart`

`simple-autorestart` 行为像带恢复的受管单实例服务。

行为：

- strategy 启动时启动一个受管实例
- 暴露与 `simple` 相同的简写路由
- 监视该受管实例，意外退出后按配置的重启策略尝试重启
- 暴露 `GET /autorestart/status` 给出重启状态（形状同 `/always-on/status`）

两个受管 strategy 的上限都取自 `multiInstance.restart`；`maxRestarts: -1` 表示无限制。

最适合：

- 信息亭或设备式设置
- 无人值守本地服务
- 一个浏览器崩溃后应自动回来的环境

### `no-instance`

`no-instance` 把 PinchTab 运行成一个不启动任何本地 Chrome 进程的 hub。它只通过 `POST /instances/attach-bridge` 接受远程桥接，并把简写请求代理到第一个接入的桥接。

行为：

- 绝不启动本地实例
- 以不启动模式注册编排器 API
- 简写路由（`/snapshot`、`/text`、`/navigate`、`/tabs` 等）代理到第一个接入的远程桥接；未接入时返回 `503 no remote instances connected — attach a bridge first`
- 未接入桥接时 `GET /tabs` 返回空列表，而非报错

最适合：

- 浏览器跑在其他主机上的桥接-编排器部署（Tailscale、远程 worker）
- 绝不应自行拉起 Chrome 的集中路由层
- 本地 PinchTab 进程必须保持纯代理的环境

## Allocation Policy

当前实现中的有效 policy：

- `fcfs`
- `round_robin`
- `random`

分配策略用于 PinchTab 有多个合格运行实例、需要选一个的情况。若你的请求已针对 `/instances/{id}/...`，该请求不涉及分配策略。

**当前行为：** 该值被校验并载入编排器的分配器，但尚无请求路径查询该分配器。无论配置哪个 policy，简写路由总是去往所请求（或默认）浏览器目标最早启动的运行实例——也就是 `fcfs` 描述的行为。

### `fcfs`

第一个运行的候选获胜。

最适合：

- 可预测行为
- 最简单的操作模型
- "始终用最早运行实例"的工作流

### `round_robin`

候选按轮换选取。

最适合：

- 稳定池上的轻量负载均衡
- 希望随时间均匀分布的重复简写式流量

### `random`

PinchTab 随机挑一个合格候选。

最适合：

- 更宽松的均衡
- 确定性顺序不重要的实验

## 示例配置

```json
{
  "multiInstance": {
    "strategy": "explicit",
    "allocationPolicy": "round_robin",
    "instancePortStart": 9868,
    "instancePortEnd": 9968
  }
}
```

## 推荐默认值

### Always-On 服务

```json
{
  "multiInstance": {
    "strategy": "always-on",
    "allocationPolicy": "fcfs",
    "restart": {
      "maxRestarts": 20,
      "initBackoffSec": 2,
      "maxBackoffSec": 60,
      "stableAfterSec": 300
    }
  }
}
```

当默认受管浏览器应立即启动并以有界重启策略保持可用时使用。

### Simple 本地服务

```json
{
  "multiInstance": {
    "strategy": "simple",
    "allocationPolicy": "fcfs"
  }
}
```

当你希望简写路由感觉像单一本地浏览器服务时使用。

### Explicit 编排

```json
{
  "multiInstance": {
    "strategy": "explicit",
    "allocationPolicy": "round_robin"
  }
}
```

当你的客户端感知实例、你想直接控制生命周期时使用。

### 自愈单服务

```json
{
  "multiInstance": {
    "strategy": "simple-autorestart",
    "allocationPolicy": "fcfs",
    "restart": {
      "maxRestarts": 3,
      "initBackoffSec": 2,
      "maxBackoffSec": 60,
      "stableAfterSec": 300
    }
  }
}
```

当一个受管浏览器应保持可用并在崩溃后恢复时使用。

## 决策规则

```text
always-on           = default, launched at startup, restarted under restart policy
simple              = on-demand shorthand auto-launch
explicit            = most control, no shorthand auto-launch
simple-autorestart  = one managed browser with crash recovery
no-instance         = pure proxy/hub for remote bridges, never launches Chrome

fcfs                = deterministic
round_robin         = balanced rotation
random              = loose distribution
```
