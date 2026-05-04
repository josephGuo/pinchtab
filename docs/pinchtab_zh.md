# PinchTab

欢迎使用 PinchTab：为 AI 代理、脚本和自动化工作流提供浏览器控制。

## PinchTab 是什么

PinchTab 是一个独立的 HTTP 服务器，通过 命令行界面 和 HTTP API 为您提供对 Chrome 的直接控制。

PinchTab 有两个运行时：

- `pinchtab server`：服务器
- `pinchtab bridge`：单实例桥接运行时

服务器是正常的入口点。它管理配置文件、实例、路由、安全策略和仪表板。
桥接是用于管理子实例背后的轻量级每个实例 HTTP 运行时。

基本模型是：

- 启动服务器
- 启动或附加实例
- 操作标签页

## 主要使用模式

启动 `pinchtab server` 或更好的 `pinchtab daemon install` 并让它运行：

- 将其用作代理的浏览器
- 将其用作本地自动化端点
- 在需要时附加现有的调试浏览器

## 最小工作流程

### 1. 启动服务器

```bash
pinchtab server
```

### 2. 启动实例

默认情况下，我们使用始终开启策略。现在这是可选的，不是必需的。

```bash
curl -X POST http://localhost:9867/instances/start \
  -H "Content-Type: application/json" \
  -d '{"mode":"headless"}'
# 命令行界面 替代方案
pinchtab instance start
# 响应
{
  "id": "inst_0a89a5bb",
  "profileId": "prof_278be873",
  "profileName": "instance-1741400000000000000",
  "port": "9868",
  "mode": "headless",
  "headless": true,
  "status": "starting"
}
```

### 3. 导航

```bash
curl -s -X POST http://localhost:9867/navigate \
  -H "Content-Type: application/json" \
  -d '{"url":"https://pinchtab.com"}' | jq .
# 命令行界面 替代方案
pinchtab nav https://pinchtab.com
# 响应
{
  "tabId": "CDP_TARGET_ID",
  "title": "PinchTab",
  "url": "https://pinchtab.com"
}
```

### 4. 检查交互元素

```bash
curl -s "http://localhost:9867/snapshot?filter=interactive" | jq .
# 命令行界面 替代方案
pinchtab snap -i -c
# 响应
{
  "nodes": [
    { "ref": "e0", "role": "link", "name": "Docs" },
    { "ref": "e1", "role": "button", "name": "Get started" }
  ]
}
```

### 5. 通过引用点击

```bash
curl -s -X POST http://localhost:9867/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"click","ref":"e1"}' | jq .
# 命令行界面 替代方案
pinchtab click e1
# 响应
{
  "success": true,
  "result": {
    "clicked": true
  }
}
```

## 特性

- 服务器优先：主进程是控制平面服务器
- 桥接支持的实例：管理的实例在隔离的桥接运行时后面运行
- 面向标签页：交互发生在标签页级别
- 有状态：配置文件持久化 cookies 和浏览器状态
- 令牌高效：快照和文本端点比截图驱动的工作流更便宜
- 灵活：无头、有头、基于配置文件或附加的 Chrome
- 受控：健康、指标、认证和标签页锁定内置于系统中

## 常见功能

- 带有 `e0`、`e1` 和类似引用的可访问性树快照
- 文本提取
- 直接操作，如点击、输入、填充、按下、聚焦、悬停、选择和滚动
- 截图和 PDF 导出
- 多实例编排
- 外部 Chrome 附加
- 可选的 JavaScript 评估

## 支持

- [GitHub Issues](https://github.com/pinchtab/pinchtab/issues)
- [GitHub Discussions](https://github.com/pinchtab/pinchtab/discussions)
- [@pinchtabdev](https://x.com/pinchtabdev)

## 许可证

[MIT](https://github.com/pinchtab/pinchtab?tab=MIT-1-ov-file#readme)