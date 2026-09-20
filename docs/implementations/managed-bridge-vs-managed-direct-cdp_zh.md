# 桥接 vs 直接 CDP

本文档比较了 PinchTab 管理浏览器实例的两种方式：

- `managed + bridge`（托管 + 桥接）
- `managed + direct-cdp`（托管 + 直接 CDP）

两者都是**托管**的，因为 PinchTab 拥有实例的生命周期。区别在于浏览器控制逻辑的位置以及服务器如何到达 Chrome。

## 简版

```text
managed + bridge
  server -> bridge -> Chrome

managed + direct-cdp
  server -> Chrome
```

桥接模型增加了一个额外的进程和一个额外的跳数。直接 CDP 模型移除了这个跳数，并将控制权保留在主服务器中。

## 图表 1：运行时形态

```text
Managed + bridge
  PinchTab server
    └─ PinchTab bridge child
         └─ Chrome
              └─ Tabs

Managed + direct-cdp
  PinchTab server
    └─ Chrome
         └─ Tabs
```

## 托管 + 桥接

### 是什么

PinchTab 为每个托管实例启动一个子 `pinchtab bridge` 进程。该桥接拥有一个浏览器并暴露单实例 HTTP API。主服务器将实例和标签页请求路由到该子进程。

### 通信路径

```text
agent -> server -> bridge -> Chrome
```

### 优点

- 强实例级隔离
- 更清晰的进程边界
- 更容易的崩溃控制
- 更容易的实例级日志和健康检查
- 作为工作进程模型，在运维上更容易理解

### 成本

- 每个实例一个额外进程
- 到达 Chrome 前一个额外的 HTTP 跳
- 更多需要分配和监控的端口
- 更多的启动开销
- 一些配置必须传播到子运行时

### 最佳适用场景

- 多实例编排
- 实例之间的强隔离
- 实例故障应保持局部的情况
- 受益于工作式进程监督的系统

## 托管 + 直接 CDP

### 是什么

PinchTab 自己启动 Chrome，并将 CDP 会话保存在主服务器进程内。没有桥接子进程，也没有额外的实例级 HTTP 服务器。

### 通信路径

```text
agent -> server -> Chrome
```

### 优点

- 更少的移动部件
- 更低的延迟
- 更少的进程和端口开销
- 更简单的网络模型
- 更少的重复 HTTP 处理

### 成本

- 默认情况下进程隔离较弱
- 主服务器内部更复杂
- 更难控制实例特定的故障
- 一个进程内更多的共享内存和状态
- 主服务器直接负责更多的生命周期细节

### 最佳适用场景

- 低开销的单主机部署
- 效率比硬隔离更重要的工作负载
- 额外工作进程不必要的环境
- 希望减少内部跳数的未来架构

## 图表 2：所有权和传输

```text
managed + bridge
  ownership: pinchtab
  transport: http-bridge + cdp

managed + direct-cdp
  ownership: pinchtab
  transport: direct cdp
```

## 图表 3：故障边界

```text
managed + bridge
  one instance crash
    -> bridge worker dies
    -> instance is affected
    -> server survives

managed + direct-cdp
  one instance failure
    -> handled inside server process
    -> isolation depends on server design
```

## 决策框架

使用以下规则：

- 当隔离和操作清晰度更重要时，选择**托管 + 桥接**
- 当运行时路径的简单性和更低的开销更重要时，选择**托管 + 直接 CDP**

或者更简短地说：

```text
bridge      = better isolation
direct-cdp  = better efficiency
```

## 当前状态

目前，预期的架构是：

- PinchTab 启动的实例使用 `managed + bridge`
- 外部受管的浏览器使用 `attached + bridge`：`POST /instances/attach` 会派生一个 `pinchtab bridge --cdp-attach <cdpUrl>` 子进程（`attachType: cdp-bridge`），`POST /instances/attach-bridge` 则注册一个已在运行的 bridge（`attachType: bridge`）——见[编排](../architecture/orchestration.md#attach-flow)

`managed + direct-cdp` 尚未实现；它只是一个架构选项。每个实例，无论是 managed 还是 attached，都通过 bridge 访问。