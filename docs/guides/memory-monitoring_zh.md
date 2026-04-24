# 内存监控

PinchTab 暴露它启动的 Chrome 进程的内存信息。当前实现在进程级别测量浏览器内存，并报告每个实例的浏览器范围聚合。

## PinchTab 测量的内容

PinchTab 遍历正在运行的实例的 Chrome 进程树：

1. 找到主浏览器 PID
2. 枚举子进程
3. 计算浏览器及其子进程的 RSS 内存总和
4. 计算渲染器进程数

这会为该实例的 Chrome 进程树提供真实的操作系统级内存使用情况。

## 内存字段

| 字段 | 含义 |
| --- | --- |
| `memoryMB` | 浏览器进程树的真实 RSS 内存 |
| `jsHeapUsedMB` | 从 `memoryMB` 派生的估计值 |
| `jsHeapTotalMB` | 从 `memoryMB` 派生的估计值 |
| `renderers` | 浏览器进程树中的渲染器进程数 |
| `documents`、`frames`、`nodes`、`listeners` | 遗留兼容性字段；当前不使用实时 DOM 计数填充 |

重要限制：

- `jsHeapUsedMB` 和 `jsHeapTotalMB` 是估计值，不是真正的每个标签页的 DevTools 堆测量值
- `GET /tabs/{id}/metrics` 返回拥有该标签页的浏览器实例的聚合内存，而不是隔离的每个标签页内存

## 实例指标

对于单个运行中的浏览器：

```bash
curl http://localhost:9867/metrics
```

示例形状：

```json
{
  "metrics": {
    "goHeapAllocMB": 12.5,
    "goHeapSysMB": 24.0,
    "goNumGoroutine": 15
  },
  "memory": {
    "memoryMB": 850.5,
    "jsHeapUsedMB": 340.2,
    "jsHeapTotalMB": 425.25,
    "renderers": 11
  }
}
```

## 每个标签页的指标

```bash
curl http://localhost:9867/tabs/<tabId>/metrics
```

示例形状：

```json
{
  "memoryMB": 850.5,
  "jsHeapUsedMB": 340.2,
  "jsHeapTotalMB": 425.25,
  "renderers": 11,
  "documents": 0,
  "frames": 0,
  "nodes": 0,
  "listeners": 0
}
```

将此视为“拥有此标签页的浏览器实例的内存”，而不是“仅此标签页的内存”。

## 所有运行实例

在编排器模式下：

```bash
curl http://localhost:9867/instances/metrics
```

这会为每个运行中的实例返回一个指标对象，这是比较整个集群内存的最佳 API。

## 仪表板监控

仪表板从以下位置消费监控快照：

```bash
curl http://localhost:9867/api/events?memory=1
```

该流包括：

- 实例列表
- 标签页列表
- 当 `memory=1` 时的每个实例指标
- PinchTab 进程本身的服务器指标

当前的 SSE 监控循环在短间隔内更新，适合实时仪表板视图。

## 故障排除

### 内存显示 `0`

可能的原因：

- Chrome 尚未启动
- 实例已停止
- 浏览器上下文未初始化

### 内存看起来比预期高

请记住，`memoryMB` 包括：

- 浏览器进程
- 渲染器进程
- 如果存在，GPU 和实用程序子进程

这通常更接近“操作系统看到的”，而不是狭窄的 JavaScript 堆数字。

### 数字与活动监视器或任务管理器不完全匹配

不同的工具报告不同的内存定义。PinchTab 当前报告它拥有的 Chrome 进程树的基于 RSS 的总计。