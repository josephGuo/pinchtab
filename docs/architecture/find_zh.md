# Find 架构

本页介绍 PinchTab 语义 `find` 管道背后的实现细节。

## 概述

`find` 系统将可访问性快照节点转换为轻量级描述符，根据自然语言查询对它们进行评分，并返回最佳匹配的 `ref`。

该实现旨在保持：

- 本地
- 快速
- 依赖轻量
- 页面重新渲染后可恢复

## 管道

```text
accessibility snapshot
  -> element descriptors
  -> lexical matcher
  -> embedding matcher
  -> combined score
  -> best ref
  -> intent cache / recovery hooks
```

## 元素描述符

每个可访问性节点都被转换为带有以下内容的描述符：

- `ref`
- `role`
- `name`
- `value`

这些字段也组合成一个用于匹配的复合字符串。

## 匹配器

PinchTab 当前使用由以下部分构建的组合匹配器：

- 词汇匹配器
- 基于哈希嵌入器的嵌入匹配器

默认权重为：

```text
0.6 lexical + 0.4 embedding
```

通过 `lexicalWeight` 和 `embeddingWeight` 可以进行每个请求的覆盖。

## 词汇侧

词汇匹配器专注于精确和近似精确的令牌重叠，包括角色感知的匹配行为。

有用的特性：

- 对精确单词表现强
- 易于推理
- 对 `submit button` 等明确查询的精度高

## 嵌入侧

嵌入匹配器使用特征哈希方法，而不是外部 ML 模型。

有用的特性：

- 捕获模糊相似性
- 更好地处理部分和子词重叠
- 没有模型下载或网络依赖

## 组合匹配

组合匹配器并发运行词汇和嵌入评分，按元素引用合并结果，并应用加权最终评分。

它在最终合并之前也使用较低的内部阈值，以便不会过早丢弃仅在一侧表现强的候选者。

## 快照依赖

`find` 依赖于快照驱动交互使用的相同可访问性快照/引用缓存基础结构。

如果缺少缓存的快照，处理程序会尝试自动刷新它，然后再放弃。

## 意图缓存和恢复

成功匹配后，PinchTab 记录：

- 原始查询
- 匹配的描述符
- 评分/置信度元数据

这允许恢复逻辑在后续操作因页面更新后旧引用变得过时而失败时尝试语义重新匹配。

## 编排器路由

编排器暴露 `POST /tabs/{id}/find` 并将其代理到正确的运行实例。实际的匹配实现仍然在共享处理程序层中。

## 设计约束

当前设计有意避免：

- 外部嵌入服务
- 重量级模型依赖
- 选择器优先耦合

这使系统保持可移植性和快速性，但也意味着质量上限受限于进程内匹配器设计和可访问性快照的质量。

## 性能

在 Intel i5-4300U @ 1.90GHz 上的基准测试：

| 操作 | 元素 | 延迟 | 分配 |
| --- | --- | --- | --- |
| Lexical Find | 16 | ~71 us | 134 allocs |
| HashingEmbedder (single) | 1 | ~11 us | 3 allocs |
| HashingEmbedder (batch) | 16 | ~171 us | 49 allocs |
| Embedding Find | 16 | ~180 us | 98 allocs |
| **Combined Find** | **16** | **~233 us** | **263 allocs** |
| Combined Find | 100 | ~1.5 ms | 1685 allocs |