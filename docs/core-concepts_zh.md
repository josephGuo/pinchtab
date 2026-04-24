# 核心概念

本文档描述了 PinchTab 中当前实现的概念。

## 服务器

**服务器**是 PinchTab 的主进程。

使用以下命令启动：

```bash
pinchtab
# 或显式启动
pinchtab server
```

服务器的功能：

- 默认在端口 `9867` 上暴露主 HTTP API 和仪表板
- 管理配置文件和实例
- 将标签页范围的请求代理到正确的管理实例
- 可以暴露简写路由，如 `/navigate`、`/snapshot` 和 `/action`

重要说明：

- 服务器是公共入口点
- 对于管理实例，服务器通常**不**直接与 Chrome 通信
- 相反，它会为每个实例生成或路由到一个**桥接**进程

## 桥接

**桥接**是单实例运行时。

仅当您需要一个独立的浏览器运行时时直接启动它：

```bash
pinchtab bridge
```

桥接的功能：

- 拥有恰好一个 Chrome 浏览器进程
- 暴露浏览器和标签页端点，如 `/navigate`、`/snapshot`、`/action` 和 `/tabs/{id}/...`
- 是服务器为每个管理实例启动的进程

在正常的多实例使用中，您通常与服务器交互，而不是直接与桥接进程交互。

## 配置文件

**配置文件**是 Chrome 用户数据目录。

它存储持久的浏览器状态，例如：

- Cookie
- 本地存储
- 缓存
- 浏览历史
- 扩展
- 保存的账户状态

与当前实现匹配的配置文件事实：

- 配置文件在磁盘上是持久的
- 配置文件可以在没有任何运行实例的情况下存在
- 一个给定的配置文件一次最多只能被一个活动的管理实例使用
- 配置文件 ID 使用 `prof_XXXXXXXX` 格式
- `GET /profiles` 隐藏临时自动生成的配置文件，除非您传递 `?all=true`

使用 API 创建配置文件：

```bash
curl -X POST http://localhost:9867/profiles \
  -H "Content-Type: application/json" \
  -d '{
    "name": "work",
    "description": "Main logged-in work profile"
  }'
# 响应
{
  "status": "created",
  "id": "prof_278be873",
  "name": "work"
}
```

## 实例

**实例**是管理的浏览器运行时。

实际上，一个实例意味着：

- 一个桥接进程
- 一个 Chrome 进程
- 零个或一个配置文件
- 一个专用端口
- 多个标签页

与当前实现匹配的实例事实：

- 实例 ID 使用 `inst_XXXXXXXX` 格式
- 端口默认从 `9868-9968` 自动分配
- 实例状态被跟踪为 `starting`、`running`、`stopping`、`stopped` 或 `error`
- 一个配置文件不能同时附加到多个活动的管理实例

### 持久实例与临时实例

启动实例有两种常见方式：

1. 使用命名配置文件
2. 不使用配置文件 ID

如果您使用配置文件 ID 启动实例，实例会使用该持久配置文件。

如果您不使用配置文件 ID 启动实例，PinchTab 会创建一个名为 `instance-...` 的自动生成配置文件。
当实例停止时，该临时配置文件会被删除。

因此，这是正确的心智模型：

- 没有显式配置文件的实例是**临时的**
- 实现仍然在后台创建临时配置文件目录
- 该临时配置文件是清理状态，不是可重用的长期配置文件

### 启动实例

首选端点：

```bash
curl -X POST http://localhost:9867/instances/start \
  -H "Content-Type: application/json" \
  -d '{
    "profileId": "prof_278be873",
    "mode": "headed"
  }'
# CLI 替代方案
pinchtab instance start --profile prof_278be873 --mode headed
# 响应
{
  "id": "inst_0a89a5bb",
  "profileId": "prof_278be873",
  "profileName": "work",
  "port": "9868",
  "mode": "headed",
  "headless": false,
  "status": "starting"
}
```

## 标签页

**标签页**是实例内的单个页面。

标签页属于一个实例，因此继承该实例的配置文件状态。

标签页为您提供：

- 自己的 URL 和页面状态
- 可访问性树的快照
- 操作执行，如点击、输入、填充、悬停和按键
- 文本提取、屏幕截图、PDF 导出、Cookie 访问和评估

在特定实例中打开标签页：

```bash
INST=inst_0a89a5bb

curl -X POST http://localhost:9867/instances/$INST/tabs/open \
  -H "Content-Type: application/json" \
  -d '{"url":"https://pinchtab.com"}'
# 响应
{
  "tabId": "CDP_TARGET_ID"
}
```

然后使用标签页范围的端点：

```bash
TAB=CDP_TARGET_ID

curl http://localhost:9867/tabs/$TAB/snapshot

curl -X POST http://localhost:9867/tabs/$TAB/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"click","ref":"e5"}'

curl -X POST http://localhost:9867/tabs/$TAB/close
```

### 标签页是持久的吗？

通常不是。

对于由服务器启动的管理实例：

- 标签页是运行时对象
- 当实例停止时，标签页会消失
- 配置文件会持久存在，但打开的标签页不会

这意味着持久的部分是**配置文件状态**，而不是标签页列表。

## 元素引用

快照返回元素引用，如 `e0`、`e1`、`e2` 等。

这些引用很有用，因为它们允许您与元素交互，而无需为常见流程编写 CSS 选择器。

## 关系

通过以下规则可以最容易地理解实现：

| 关系 | 当前的实际情况 |
|---|---|
| 服务器 -> 实例 | 一个服务器可以管理多个实例 |
| 桥接 -> Chrome | 一个桥接拥有一个 Chrome 进程 |
| 实例 -> 配置文件 | 一个实例有零个或一个配置文件 |
| 配置文件 -> 实例 | 一个配置文件一次最多可以有一个活动的管理实例 |
| 实例 -> 标签页 | 一个实例可以有多个标签页 |
| 标签页 -> 实例 | 每个标签页恰好属于一个实例 |
| 标签页 -> 配置文件 | 标签页继承实例的配置文件（如果存在） |

配置文件是可重用的持久状态。实例是可能使用配置文件的临时运行时。

## 简写路由与显式路由

PinchTab 暴露两种交互风格：

### 显式路由

这些始终命名您想要的资源：

- `POST /instances/start`
- `POST /instances/{id}/tabs/open`
- `GET /tabs/{id}/snapshot`
- `POST /tabs/{id}/action`

这是多实例工作的最清晰模型。

### 简写路由

这些省略了实例，有时也省略了标签页：

- `POST /navigate`
- `GET /snapshot`
- `POST /action`
- `GET /text`

这些路由到"当前"或第一个运行的实例。

## 推荐的心智模型

对于大多数用户，以下是正确的顺序：

1. 使用 `pinchtab` 启动服务器
2. 如果需要持久性，创建配置文件
3. 从该配置文件启动实例
4. 在该实例中打开一个或多个标签页
5. 对标签页进行快照
6. 对该快照中的引用执行操作

如果您不需要持久性：

1. 启动没有 `profileId` 的实例
2. 正常使用它
3. 完成后停止它
4. 让 PinchTab 自动删除临时配置文件

## 示例工作流

### 工作流 1：持久登录浏览器

```bash
PROFILE_ID=$(curl -s -X POST http://localhost:9867/profiles \
  -H "Content-Type: application/json" \
  -d '{"name":"work"}' | jq -r '.id')

INST=$(curl -s -X POST http://localhost:9867/instances/start \
  -H "Content-Type: application/json" \
  -d "{\"profileId\":\"$PROFILE_ID\",\"mode\":\"headed\"}" | jq -r '.id')

TAB=$(curl -s -X POST http://localhost:9867/instances/$INST/tabs/open \
  -H "Content-Type: application/json" \
  -d '{"url":"https://pinchtab.com/login"}' | jq -r '.tabId')

curl http://localhost:9867/tabs/$TAB/snapshot
```

当您希望 Cookie 和账户状态在实例重启后仍然存在时使用此方法。

### 工作流 2：一次性运行

```bash
INST=$(curl -s -X POST http://localhost:9867/instances/start \
  -H "Content-Type: application/json" \
  -d '{"mode":"headless"}' | jq -r '.id')

TAB=$(curl -s -X POST http://localhost:9867/instances/$INST/tabs/open \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com"}' | jq -r '.tabId')

curl http://localhost:9867/tabs/$TAB/text

curl -X POST http://localhost:9867/instances/$INST/stop
```

当您需要一个干净的、一次性的会话时使用此方法。

## 总结

PinchTab 中的持久对象是**配置文件**。
运行时对象是**实例**。
页面对象是**标签页**。
**服务器**管理它们，**桥接**执行它们。