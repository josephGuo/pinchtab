# 策略和分配

PinchTab 有两个独立的多实例控制：

- `multiInstance.strategy`
- `multiInstance.allocationPolicy`

它们解决不同的问题：

```text
strategy          = PinchTab 暴露哪些路由以及简写请求的行为方式
allocationPolicy  = 当 PinchTab 必须选择一个实例时，选择哪个运行中的实例
```

## 策略

当前实现中的有效策略：

- `always-on` - 默认
- `simple`
- `explicit`
- `simple-autorestart`

### `simple`

行为：

- 注册完整的编排器 API
- 保留简写路由，如 `/snapshot`、`/text`、`/navigate` 和 `/tabs`
- 如果简写请求到达且没有实例运行，PinchTab 会自动启动一个管理实例并等待其变得健康

最佳适用：

- 本地开发
- 单用户自动化
- “只要让浏览器服务可用”的设置

### `explicit`

`explicit` 也暴露编排器 API 和简写路由，但它不会在简写请求时自动启动。

行为：

- 你通过 `/instances/start`、`/instances/launch` 或 `/profiles/{id}/start` 显式启动实例
- 简写路由仅在已存在一个运行实例时代理到第一个运行实例
- 如果没有运行实例，简写路由会返回错误而不是为你启动浏览器

最佳适用：

- 受控的多实例环境
- 应该有意命名实例的代理
- 隐藏的自动启动会令人惊讶的部署

### `always-on`

`always-on` 的行为类似于一个管理的单实例服务，应该在 PinchTab 进程的整个生命周期内保持运行。

行为：

- 在策略启动时启动一个管理实例
- 暴露与 `simple` 相同的简写路由
- 监视该管理实例，并在意外退出后不断重启它，直到达到配置的重启限制
- 暴露 `GET /always-on/status` 以获取当前管理实例状态

最佳适用：

- 守护进程式的本地服务
- 期望始终存在一个默认浏览器的代理主机
- 启动可用性很重要，但你仍然希望有一个有限的故障策略的设置

### `simple-autorestart`

`simple-autorestart` 的行为类似于一个具有恢复功能的管理单实例服务。

行为：

- 在策略启动时启动一个管理实例
- 暴露与 `simple` 相同的简写路由
- 监视该管理实例，并在意外退出后根据配置的重启策略尝试重启它
- 暴露 `GET /autorestart/status` 以获取重启状态

最佳适用：

- 信息亭或设备式设置
- 无人值守的本地服务
- 一个浏览器应该在崩溃后恢复的环境

## 分配策略

当前实现中的有效策略：

- `fcfs`
- `round_robin`
- `random`

分配策略仅在 PinchTab 有多个符合条件的运行实例并需要选择一个时才重要。如果你的请求已经针对 `/instances/{id}/...`，则该请求不涉及分配策略。

### `fcfs`

第一个运行的候选者获胜。

最佳适用：

- 可预测的行为
- 最简单的操作模型
- “始终使用最早运行的实例”工作流

### `round_robin`

候选者按轮换方式选择。

最佳适用：

- 在稳定池中进行轻量级平衡
- 你希望随着时间均匀分布的重复简写式流量

### `random`

PinchTab 选择一个随机的符合条件的候选者。

最佳适用：

- 更宽松的平衡
- 确定性排序不重要的实验

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

### 始终开启服务

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

当默认管理浏览器应该立即启动并通过有限的重启策略保持可用时使用此配置。

### 简单本地服务

```json
{
  "multiInstance": {
    "strategy": "simple",
    "allocationPolicy": "fcfs"
  }
}
```

当你希望简写路由感觉像一个单一的本地浏览器服务时使用此配置。

### 显式编排

```json
{
  "multiInstance": {
    "strategy": "explicit",
    "allocationPolicy": "round_robin"
  }
}
```

当你的客户端了解实例并且你想直接控制生命周期时使用此配置。

### 自我修复单一服务

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

当一个管理浏览器应该保持可用并在崩溃后恢复时使用此配置。

## 决策规则

```text
always-on           = 默认，启动时启动，根据重启策略重启
simple              = 按需简写自动启动
explicit            = 最大控制，无简写自动启动
simple-autorestart  = 一个带有崩溃恢复的管理浏览器

fcfs                = 确定性
round_robin         = 平衡轮换
random              = 宽松分布
```