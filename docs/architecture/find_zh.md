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
  -> DOM metadata enrichment
  -> element descriptors
  -> IDPI scan of the candidate text (when enabled)
  -> lexical matcher
  -> embedding matcher
  -> combined score
  -> best ref
  -> intent cache / recovery hooks
```

## 元素描述符

每个可访问性节点都被转换为一个描述符，包含：

- `ref`
- `role`
- `name`
- `value`
- `label`
- `placeholder`
- `alt`
- `title`
- `testid`
- `text`
- `tag`
- `interactive`
- `parent`
- `section`
- `documentIdx`
- 位置提示：`depth`、`siblingIndex`、`siblingCount`、`labelledBy`

PinchTab 负责从后端节点 id 中提取纯 DOM 元数据。对 `role:`、`text:`、`label:`、`placeholder:`、`alt:`、`title:`、`testid:`、`first:`、`last:` 和 `nth:` 这些形式的结构化定位符解析与匹配，被委托给外部的 `github.com/pinchtab/semantic` Go 模块（一个兄弟包，不属于本仓库）。仓库内的 `internal/autosolver/semantic/adapter.go` 只是一个薄适配器，把该模块接入 autosolver。

CSS、XPath、refs、frame 作用域、把匹配到的 ref 转换回后端节点，以及现有的由 DOM 支撑的动作 `text:` 选择器，仍属于 PinchTab 的职责。

## 匹配器

PinchTab 当前使用由以下部分构建的组合匹配器：

- 词汇匹配器
- 基于哈希嵌入器的嵌入匹配器

默认权重为：

```text
0.6 lexical + 0.4 embedding
```

通过 `lexicalWeight` 和 `embeddingWeight` 可以进行每次请求的覆盖。

## 词汇侧

词汇匹配器专注于精确和近似精确的 token 重叠，包括角色感知的匹配行为。

有用的特性：

- 对精确单词表现强
- 易于推理
- 对 `submit button` 等明确查询精度高

## 嵌入侧

嵌入匹配器使用特征哈希方法，而不是外部 ML 模型。

有用的特性：

- 捕获模糊相似性
- 更好地处理部分和子词重叠
- 没有模型下载或网络依赖

## 组合匹配

组合匹配器并发运行词汇和嵌入评分，按元素 ref 合并结果，并应用加权最终评分。

它在最终合并之前还使用一个较低的内部阈值，以便不会过早丢弃仅在一侧表现强的候选者。

## 快照依赖

`find` 依赖于快照驱动交互所使用的相同可访问性快照/ref 缓存基础结构。

如果缺少缓存的快照，处理器会在放弃之前尝试自动刷新它。

## 意图缓存和恢复

成功匹配后，PinchTab 记录：

- 原始查询
- 匹配的描述符
- 评分/置信度元数据

这使得恢复逻辑可以在后续操作因页面更新后旧 ref 变得过时而失败时，尝试一次语义重新匹配。

## 路由

`POST /tabs/{id}/find`（以及活动标签页的简写 `POST /find`）由 `internal/handlers/handlers.go` 中的共享处理器层注册。编排器通过其路由层把这些请求代理到正确的运行实例；它并不拥有该路由本身。

## 设计约束

当前设计有意避免：

- 外部嵌入服务
- 重量级模型依赖
- 选择器优先耦合

这使系统保持可移植和快速，但也意味着质量上限受限于进程内匹配器设计和可访问性快照的质量。

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
