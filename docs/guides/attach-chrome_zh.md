# 附加 Chrome

当以下情况时使用本指南：

- Chrome 已经存在于 PinchTab 外部
- 您希望 PinchTab 服务器将该浏览器注册为实例
- 您已经有浏览器级别的 DevTools WebSocket URL

如果您的目标只是：

- 为您的代理启动浏览器
- 运行正常的本地 PinchTab 工作流

请不要使用本指南。

对于这种情况，请使用带有 `pinchtab` 和 `POST /instances/start` 的管理实例。

---

## 启动 vs 附加

心智模型是：

```text
启动 = PinchTab 启动并拥有浏览器
附加 = PinchTab 注册已经运行的浏览器
```

使用附加：

- Chrome 在其他地方启动
- PinchTab 接收 `cdpUrl`
- 服务器将该浏览器注册为附加实例

---

## 当前实现

当前代码库实现：

- `POST /instances/attach`
- 配置中 `security.attach` 下的附加策略
- `GET /instances` 中的附加实例元数据

附加请求体是：

```json
{
  "name": "shared-chrome",
  "cdpUrl": "ws://127.0.0.1:9222/devtools/browser/..."
}
```

目前没有 命令行界面 附加命令。

---

## 步骤 1：启用附加策略

除非您在配置中允许，否则附加是禁用的。

示例：

```json
{
  "security": {
    "attach": {
      "enabled": true,
      "allowHosts": ["127.0.0.1", "localhost", "::1"],
      "allowSchemes": ["ws", "wss"]
    }
  }
}
```

这会：

- 启用附加端点
- 限制接受哪些主机
- 限制接受哪些 URL 方案

这不会：

- 启动 Chrome
- 定义全局远程浏览器
- 替换管理实例

---

## 步骤 2：使用远程调试启动 Chrome

示例：

```bash
google-chrome --remote-debugging-port=9222
# 或在某些系统上：
# chromium --remote-debugging-port=9222
```

这使 Chrome 暴露浏览器级别的 DevTools 端点。

---

## 步骤 3：获取浏览器 WebSocket URL

查询 Chrome：

```bash
curl -s http://127.0.0.1:9222/json/version | jq .
# 响应
{
  "webSocketDebuggerUrl": "ws://127.0.0.1:9222/devtools/browser/abc123"
}
```

`webSocketDebuggerUrl` 的值是您传递给 PinchTab 的 `cdpUrl`。

---

## 步骤 4：将其附加到 PinchTab

```bash
curl -X POST http://localhost:9867/instances/attach \
  -H "Content-Type: application/json" \
  -d '{
    "name": "shared-chrome",
    "cdpUrl": "ws://127.0.0.1:9222/devtools/browser/abc123"
  }'
# 响应
{
  "id": "inst_0a89a5bb",
  "profileId": "prof_278be873",
  "profileName": "shared-chrome",
  "port": "",
  "mode": "headed",
  "headless": false,
  "status": "running",
  "attached": true,
  "cdpUrl": "ws://127.0.0.1:9222/devtools/browser/abc123"
}
```

注意：

- `name` 是可选的；如果省略，服务器会生成一个类似 `attached-...` 的名称
- 服务器根据 `security.attach.allowHosts` 和 `security.attach.allowSchemes` 验证 URL

---

## 步骤 5：确认它已注册

```bash
curl -s http://localhost:9867/instances | jq .
# 命令行界面 替代方案
pinchtab instances
```

附加实例出现在正常实例列表中，带有：

- `attached: true`
- `cdpUrl: ...`
- `status: "running"`

---

## 所有权和生命周期

附加实例是外部拥有的。

这意味着：

- PinchTab 没有启动浏览器
- PinchTab 将该浏览器的元数据存储为实例
- 外部 Chrome 进程保持在 PinchTab 生命周期所有权之外

在实际操作中：

- 在 PinchTab 中停止附加实例会将其从服务器注销
- 这并不意味着 PinchTab 启动或可以完全管理外部 Chrome 进程

---

## 附加何时有意义

当以下情况时使用附加：

- Chrome 由另一个系统管理
- Chrome 已经在单独的服务或容器中运行
- 您希望服务器知道外部管理的浏览器
- 您希望将浏览器所有权保持在 PinchTab 之外

---

## 安全

附加扩大了信任边界，因此请保持其锁定。

推荐规则：

- 除非需要，否则保持附加禁用
- 保持 `allowHosts` 范围狭窄
- 保持 `allowSchemes` 范围狭窄
- 当服务器可从 localhost 外部访问时设置 `PINCHTAB_TOKEN`
- 仅附加到您信任的 CDP 端点

如果您将 `allowHosts` 设置为 `["*"]`，PinchTab 会接受任何具有允许方案的可访问附加主机。这是一个记录的、非默认的、降低安全性的覆盖：它完全移除了主机允许列表，只应在隔离的、操作员控制的网络上使用。

还要记住：

- Chrome DevTools 提供强大的浏览器控制
- 可访问的 CDP 端点应被视为敏感基础设施

如果 Chrome 是远程的，首选隧道而不是广泛暴露调试端口。

---

## 操作模型

预期模型是：

```text
代理 -> PinchTab 服务器 -> 附加的外部 Chrome
```

这是专家路径，不是默认用户路径。

默认路径仍然是：

```bash
pinchtab
```

然后通过以下方式启动管理实例：

```bash
curl -X POST http://localhost:9867/instances/start \
  -H "Content-Type: application/json" \
  -d '{"mode":"headless"}'
# 命令行界面 替代方案
pinchtab instance start
```