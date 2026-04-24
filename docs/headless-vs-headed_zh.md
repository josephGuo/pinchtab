# 无头模式 vs 有头模式

PinchTab 实例可以在两种模式下运行 Chrome：

- **无头模式**：无可见浏览器窗口
- **有头模式**：可见浏览器窗口

您通常使用 `pinchtab` 运行一个服务器，然后通过 API 或 CLI 以任一模式启动实例。

---

## 无头模式

无头模式是管理实例的默认模式。

```bash
curl -X POST http://localhost:9867/instances/start \
  -H "Content-Type: application/json" \
  -d '{"mode":"headless"}'
# CLI 替代方案
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

### 适合场景

- 代理和自动化
- CI 和批处理作业
- 容器和远程服务器
- 更高吞吐量的工作负载

### 权衡

- 无可见浏览器窗口
- 调试通常通过快照、屏幕截图、文本提取和日志进行

---

## 有头模式

有头模式启动可见的 Chrome 窗口。

```bash
curl -X POST http://localhost:9867/instances/start \
  -H "Content-Type: application/json" \
  -d '{"mode":"headed"}'
# CLI 替代方案
pinchtab instance start --mode headed
# 响应
{
  "id": "inst_1b9a5dcc",
  "profileId": "prof_278be873",
  "profileName": "instance-1741400000000000001",
  "port": "9869",
  "mode": "headed",
  "headless": false,
  "status": "starting"
}
```

### 适合场景

- 开发
- 调试
- 本地测试
- 视觉验证
- 人工干预工作流

### 权衡

- 需要显示环境
- 通常比无头模式使用更多的 CPU 和内存

---

## 并排比较

| 方面 | 无头模式 | 有头模式 |
|---|---|---|
| 可见性 | 无窗口 | 可见窗口 |
| 调试 | 基于快照和日志 | 直接视觉反馈 |
| 需要显示 | 否 | 是 |
| CI 使用 | 非常适合 | 通常不适合 |
| 本地开发 | 可用 | 通常更简单 |
| 资源使用 | 较低 | 较高 |

---

## 推荐工作流

## 开发工作流

在构建和验证流程时使用可见浏览器：

```bash
curl -X POST http://localhost:9867/instances/start \
  -H "Content-Type: application/json" \
  -d '{"mode":"headed"}'
# CLI 替代方案
pinchtab instance start --mode headed
```

当您需要持久性时，先创建一个配置文件：

```bash
curl -X POST http://localhost:9867/profiles \
  -H "Content-Type: application/json" \
  -d '{"name":"dev"}'
# 响应
{
  "status": "created",
  "id": "prof_278be873",
  "name": "dev"
}
```

然后在有头模式下启动配置文件：

```bash
curl -X POST http://localhost:9867/instances/start \
  -H "Content-Type: application/json" \
  -d '{"profileId":"prof_278be873","mode":"headed"}'
# CLI 替代方案
pinchtab instance start --profile prof_278be873 --mode headed
# 响应
{
  "id": "inst_ea2e747f",
  "profileId": "prof_278be873",
  "profileName": "dev",
  "port": "9868",
  "mode": "headed",
  "headless": false,
  "status": "starting"
}
```

## 生产工作流

使用无头模式进行可重复的无人值守工作：

```bash
for i in 1 2 3; do
  curl -s -X POST http://localhost:9867/instances/start \
    -H "Content-Type: application/json" \
    -d '{"mode":"headless"}' | jq .
done
# CLI 替代方案
for i in 1 2 3; do
  pinchtab instance start
done
```

---

## 检查无头实例

您可以通过标签页 API 调试无头实例。

列出实例标签页：

```bash
curl http://localhost:9867/instances/inst_0a89a5bb/tabs | jq .
# 响应
[
  {
    "id": "CDP_TARGET_ID",
    "instanceId": "inst_0a89a5bb",
    "url": "https://pinchtab.com",
    "title": "PinchTab"
  }
]
```

然后检查标签页：

```bash
curl http://localhost:9867/tabs/CDP_TARGET_ID/snapshot | jq .
```

```bash
curl http://localhost:9867/tabs/CDP_TARGET_ID/text | jq .
```

```bash
curl http://localhost:9867/tabs/CDP_TARGET_ID/screenshot > page.jpg
```

---

## 显示要求

有头实例需要显示环境。

### macOS

有头模式适用于原生桌面会话。

### Linux

无头模式在任何地方都可以工作。
有头模式需要 X11 或 Wayland。

```bash
ssh -X user@server 'pinchtab instance start --mode headed'
```

### Windows

Windows 构建可用，但 Windows 支持目前有限且尽力而为。
有头模式针对原生桌面会话。
首选使用 `pinchtab server` 或 `pinchtab bridge` 直接本地运行；守护进程工作流不是主要的 Windows 路径。

### Docker

无头模式是容器中的正常选择：

```bash
docker run -d -p 9867:9867 pinchtab/pinchtab
curl -X POST http://localhost:9867/instances/start \
  -H "Content-Type: application/json" \
  -d '{"mode":"headless"}'
```

---

## 仪表板

仪表板允许您在使用任一模式时监控运行中的实例和配置文件。

有用的视图：

- 监控：运行中的实例、状态、标签页和可选内存指标
- 配置文件：保存的配置文件、启动操作和实时详细信息
- 设置：运行时和仪表板首选项

---

## 总结

- 使用**无头模式**进行无人值守自动化和扩展。
- 使用**有头模式**进行本地调试和人工可见工作流。
- 按实例选择模式，而不是按服务器。