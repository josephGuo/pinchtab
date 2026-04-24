# 仪表板

PinchTab 包含一个内置的 Web 仪表板，用于监控实例、管理配置文件和编辑配置。

仪表板是完整服务器的一部分：

- `pinchtab` 或 `pinchtab server` 启动完整服务器并提供仪表板
- `pinchtab bridge` 不提供仪表板

您可以在以下地址打开仪表板：

- `http://localhost:9867`
- `http://localhost:9867/dashboard`

> [!WARNING]
> 仪表板是操作员/管理员控制面板，不是公共或多用户应用程序。不要将其暴露给不受信任的用户。任何可以使用仪表板的人都可以管理配置文件、实例、配置和该服务器上启用的其他浏览器控制功能。

---

## 仪表板概览

当前仪表板公开三个主要页面：

1. **监控**
2. **配置文件**
3. **设置**

UI 是一个由 Go 服务器提供服务的 React SPA。

---

## 监控页面

![仪表板实例](media/dashboard-instances.jpeg)

监控页面是默认视图。

它显示：

- 运行和停止的实例
- 选定实例的详细信息
- 选定实例的打开标签页
- 图表化监控数据
- 在设置中启用时的可选内存指标

您可以做什么：

- 选择实例
- 检查其端口、模式和状态
- 查看其打开的标签页
- 停止运行中的实例

操作数据来自：

- `GET /api/events` 上的 SSE 更新
- `GET /instances` 中的实例列表
- `GET /instances/{id}/tabs` 中的标签页数据
- `GET /instances/metrics` 中的可选内存数据

---

## 配置文件页面

![仪表板配置文件](media/dashboard-profiles.jpeg)

配置文件页面管理保存的浏览器配置文件。

它显示：

- 可用配置文件
- 启动和停止操作
- 配置文件元数据，如名称、路径、大小、源和账户详细信息

您可以做什么：

- 创建新配置文件
- 将配置文件启动为管理实例
- 停止配置文件的运行实例
- 编辑配置文件元数据
- 删除配置文件
- 打开配置文件详细信息模态框

启动流程在后台使用服务器 API：

```bash
curl -X POST http://localhost:9867/profiles \
  -H "Content-Type: application/json" \
  -d '{"name":"work","useWhen":"Team account workflows"}'
# 响应
{
  "status": "created",
  "id": "prof_278be873",
  "name": "work"
}
```

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
  "profileName": "work",
  "port": "9868",
  "mode": "headed",
  "headless": false,
  "status": "starting"
}
```

---

## 配置文件详细信息模态框

配置文件详细信息显示在模态框中，而不是作为单独的顶级页面。

模态框当前包含以下标签页：

- **配置文件**
- **实时**
- **日志**

从那里您可以：

- 查看配置文件 ID 和元数据
- 编辑名称和 `useWhen`
- 检查运行实例的实时标签页
- 打开标签页预览的屏幕截图瓦片

---

## 设置页面

![仪表板设置](media/dashboard-settings.jpeg)

设置页面将本地仪表板首选项与后端配置相结合。

它包括以下部分：

- 仪表板
- 实例默认值
- 编排
- 安全
- 安全 IDPI
- 配置文件
- 网络和附加
- 浏览器运行时
- 超时

您可以做什么：

- 更改本地仪表板首选项，如监控和屏幕截图设置
- 从 `GET /api/config` 加载后端配置
- 通过 `PUT /api/config` 保存后端配置
- 查看服务器级更改是否需要重启

安全部分现在包括：

- `security.allowedDomains` 用于 IDPI 域检查使用的网站允许列表
- `security.trustedProxyCIDRs` 用于已知内部代理，其运行时远程 IP 应被信任
- `security.trustedResolveCIDRs` 用于操作员控制的 DNS 或代理设置，其中主机名故意解析为非公共 IP

安全 IDPI 部分专注于内容保护行为：

- `security.idpi.enabled`
- `security.idpi.strictMode`
- `security.idpi.scanContent`
- `security.idpi.wrapContent`
- `security.idpi.customPatterns`

健康负载也会显示摘要信息：

```bash
curl http://localhost:9867/health | jq .
# 响应
{
  "status": "ok",
  "mode": "dashboard",
  "profiles": 3,
  "instances": 1,
  "agents": 0,
  "restartRequired": false
}
```

---

## 事件流

仪表板使用服务器发送事件，而不是 WebSockets。

主要流端点：

```bash
curl http://localhost:9867/api/events
```

此流携带：

- `init`
- `action`
- `system`
- `monitoring`

---

## 构建说明

如果 React 仪表板资产未构建到二进制文件中，服务器会提供一个回退页面，告诉您构建仪表板包。

---

## 故障排除

### 仪表板未加载

```bash
curl http://localhost:9867/health
```

如果服务器已启动，请尝试：

- `http://localhost:9867`
- `http://localhost:9867/dashboard`

### 没有可见实例

启动一个：

```bash
curl -X POST http://localhost:9867/instances/start \
  -H "Content-Type: application/json" \
  -d '{"mode":"headless"}'
# CLI 替代方案
pinchtab instance start
```

### 没有实时配置文件预览

配置文件必须有运行实例，才能在配置文件详细信息模态框的实时标签页中显示实时标签页数据。