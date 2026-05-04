# 远程桥接与编排器

当以下情况时使用本指南：

- PinchTab 编排器在一台机器上运行
- PinchTab 桥接服务器在另一台机器上运行
- 您希望代理在浏览器工作在远程进行时保持与编排器的通信

这是一种高级部署模式。仅当您了解安全模型、将桥接保持在私有或其他封闭网络上，并且避免将桥接或编排器广泛暴露到超出需要访问它们的系统之外时使用。高风险端点系列应保持禁用，除非明确需要，如果启用，它们应仅可由部署中涉及的最少受信任系统访问。

现在这是通过以下方式支持的编排模式：

```text
POST /instances/attach-bridge
```

编排器不会在远程机器上启动进程。它附加到已经运行的桥接并将请求路由到其注册的源。

---

## 心智模型

现在有三种不同的模型：

- 本地管理实例：编排器启动并拥有本地桥接进程
- 附加的 Chrome：编排器使用 `POST /instances/attach` 注册外部 CDP 浏览器
- 附加的桥接：编排器使用 `POST /instances/attach-bridge` 注册外部 PinchTab 桥接

对于远程桥接附加，控制流程是：

```text
代理 -> 编排器 -> 附加的远程桥接 -> Chrome
```

主要的实用规则是：

- 客户端与编排器通信
- 编排器与远程桥接通信

---

## 良好的用例

### 共享有头浏览器主机

- 机器 A 运行编排器和仪表板
- 机器 B 运行一个或多个有头桥接服务器
- 代理在机器 A 上保持单一控制平面
- 实际的浏览器渲染在机器 B 上进行

当您想要一个编排表面但不想在每台开发机器上运行有头 Chrome 时，这很有用。

### 区域本地浏览器工作器

- 机器 A 运行编排器
- 机器 B 在与目标站点相同的 LAN、VPC 或区域中运行桥接
- 编排器附加该桥接并将工作路由到那里

当延迟、出口位置或网络拓扑很重要时，这很有用。

---

## 功能的作用

`POST /instances/attach-bridge`：

- 根据 `security.attach` 验证桥接 URL
- 在注册前检查远程桥接健康端点
- 将桥接源存储为规范实例 URL
- 可选地存储编排器使用的每个桥接的承载令牌
- 允许正常的编排器路由代理到该桥接

此功能还更新了路由行为：

- 代理不再假设每个实例都使用 `localhost`
- 编排器现在路由到注册的实例源
- 代理目标仅限于属于注册实例的源

最后一点对于安全很重要：代理不会对任意目标开放。

---

## 配置

远程桥接附加重用现有的附加策略：

```json
{
  "security": {
    "attach": {
      "enabled": true,
      "allowHosts": ["10.0.12.24", "bridge.internal"],
      "allowSchemes": ["ws", "wss", "http", "https"]
    }
  }
}
```

重要说明：

- `security.attach.enabled` 必须为 `true`
- `allowHosts` 必须包含远程桥接主机
- `allowSchemes` 必须包含 `http` 或 `https` 以进行桥接附加
- `ws` 和 `wss` 仍用于 CDP 附加
- `baseUrl` 必须是裸桥接源；不要包含凭据、查询字符串、片段或路径

如果您使用 `allowHosts: ["*"]`，编排器将接受任何具有允许方案的可访问桥接主机。这是一个有文档记录的、非默认的、降低安全性的覆盖：它完全移除了主机允许列表，只应在隔离的、操作员控制的网络上使用。

如果您将 `allowSchemes` 仅保留为 `ws,wss`，`attach-bridge` 将被拒绝。

---

## 步骤 1：启动远程桥接

在远程机器上，配置并启动桥接：

```bash
# 设置网络访问的绑定地址
pinchtab config set server.bind 0.0.0.0
pinchtab config set server.port 9868
pinchtab config set server.token bridge-secret-token

# 启动桥接
pinchtab bridge
```

这种非环回绑定是有文档记录的、非默认的、降低安全性的部署更改。这里是适当的，因为桥接必须可从编排器访问。保持桥接令牌设置，并仅在受控网络边界上暴露端口。

示例桥接源：

```text
http://10.0.12.24:9868
```

桥接在附加之前应该已经健康且可访问。

---

## 步骤 2：直接验证桥接

从编排器机器：

```bash
curl -H "Authorization: Bearer bridge-secret-token" \
  http://10.0.12.24:9868/health
```

您应该得到 `200 OK`。

`attach-bridge` 也执行自己的健康探测，但首先直接检查使网络和身份验证问题更容易调试。

---

## 步骤 3：将桥接附加到编排器

针对编排器：

```bash
curl -X POST http://127.0.0.1:9867/instances/attach-bridge \
  -H "Authorization: Bearer orchestrator-token" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "bridge-eu-west-1",
    "baseUrl": "http://10.0.12.24:9868",
    "token": "bridge-secret-token"
  }'
```

响应形状：

```json
{
  "id": "inst_0a89a5bb",
  "profileId": "prof_278be873",
  "profileName": "bridge-eu-west-1",
  "port": "",
  "url": "http://10.0.12.24:9868",
  "mode": "headed",
  "headless": false,
  "status": "running",
  "attached": true,
  "attachType": "bridge"
}
```

需要注意的字段：

- `attached: true`
- `attachType: "bridge"`
- `url` 是注册的桥接源

---

## 步骤 4：确认它已注册

```bash
curl -H "Authorization: Bearer orchestrator-token" \
  http://127.0.0.1:9867/instances
```

附加的桥接出现在正常实例列表中。编排器将其视为运行中的实例，用于路由和集群操作。

---

## 步骤 5：使用正常的编排器路由

附加后，客户端继续与编排器通信。

示例：

```bash
curl -H "Authorization: Bearer orchestrator-token" \
  http://127.0.0.1:9867/instances/<instanceId>/tabs
```

```bash
curl -X POST http://127.0.0.1:9867/instances/<instanceId>/tabs/open \
  -H "Authorization: Bearer orchestrator-token" \
  -H "Content-Type: application/json" \
  -d '{"url":"https://pinchtab.com"}'
```

```bash
curl -H "Authorization: Bearer orchestrator-token" \
  http://127.0.0.1:9867/tabs/<tabId>/snapshot
```

如果您的活动策略使用简写路由，这些路由也可以落在附加的桥接上，因为实例选择现在从规范实例 URL 而不仅仅是本地端口工作。

---

## 身份验证模型

有两个单独的身份验证跳：

1. 客户端到编排器
2. 编排器到附加的桥接

这意味着您可以使用不同的令牌：

- 客户端发送 `Authorization: Bearer orchestrator-token`
- 编排器向远程桥接发送 `Authorization: Bearer bridge-secret-token`

客户端不需要知道桥接令牌。编排器将其存储为实例附加元数据，并在出站桥接请求上注入它。

---

## 生命周期语义

附加的桥接是外部拥有的。

这意味着：

- 编排器没有启动远程桥接进程
- `attach-bridge` 注册路由元数据，而不是远程进程所有权

当前行为：

- 本地 `Launch()` 保持仅本地
- 停止附加的桥接会将其从编排器中移除
- 对于附加的桥接，编排器还会在注销前进行尽力而为的 `POST /shutdown` 调用
- 不支持通过编排器启动附加的非桥接实例

如果您需要真正的远程进程启动，那是一个不同的问题，需要传输，如 SSH、代理或调度器支持的工作器系统。

---

## 安全约束

远程桥接支持很有用，但它扩大了信任边界。

推荐做法：

- 保持 `allowHosts` 狭窄
- 只允许您实际需要的方案
- 使用专用的桥接令牌
- 当桥接穿越不受信任的网络时，优先使用 `https`
- 尽可能将桥接本身放在网络 ACL 或隧道后面

编排器代理被有意限制：

- 它仅代理到注册的实例源
- 它不接受任意的调用者控制的目标

这可以防止远程桥接功能变成通用的 SSRF 机制。

---

## 限制

此功能不执行以下操作：

- 它不会在远程机器上启动桥接进程
- 它不会在主机之间同步配置文件目录
- 它不会跨机器迁移标签页或浏览器状态
- 它不会自动发现工作器

支持的模型是：

- 远程启动桥接
- 明确附加它
- 通过编排器路由流量

---

## 仅中心模式

如果您只想要远程桥接，永远不想要本地 Chrome，请使用 `no-instance` 策略：

```json
{
  "multiInstance": {
    "strategy": "no-instance"
  }
}
```

这会阻止所有本地启动端点，并将服务器作为纯中心启动。远程桥接通过 `POST /instances/attach-bridge` 附加，简写路由代理到第一个连接的桥接。

---

## 总结

当您需要以下情况时使用 `POST /instances/attach-bridge`：

- 机器 A 上的编排器
- 机器 B 上的桥接
- 代理仍只与机器 A 通信
- 远程浏览器工作，无需远程进程管理复杂性

当您想要一个永远不启动本地 Chrome 的专用中心时，使用 `no-instance` 策略。

当您想要分布式执行与单一控制平面时，这是正确的功能。