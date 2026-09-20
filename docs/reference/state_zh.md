# State（状态）

`pinchtab state` 显示某标签页当前完整的浏览器状态，同时也管理磁盘上已保存的浏览器状态。

有三个相关的状态视图：

- **完整浏览器状态**，通过 `pinchtab state` 或 `GET /state`
- **实时标签页状态**，通过 `GET /tabs/{id}/state`
- **当前源的存储**，通过 `pinchtab storage ...`

完整浏览器状态包括：

- cookies
- 当前源的 `localStorage`
- 当前源的 `sessionStorage`
- 可选元数据

所有完整/已保存状态的操作都需要 `security.allowStateExport=true`；关闭时返回 `403 state_export_disabled`。

## 命令

```bash
pinchtab state [--tab <id>]
pinchtab state list
pinchtab state save [--name <name>] [--encrypt] [--tab <id>]
pinchtab state load --name <name-or-prefix> [--tab <id>]
pinchtab state show --name <name>
pinchtab state delete --name <name>
pinchtab state clean [--older-than <hours>]
```

`save`、`load`、`show` 和 `delete` 也接受位置参数形式的名称（`pinchtab state load work-login`）；`--name` 是同一个值。

## 查看当前浏览器状态

```bash
pinchtab state
pinchtab state --tab <tabId>
```

返回活动标签页或指定标签页的当前完整浏览器状态：

- cookies
- 当前源的存储
- 标签页信息，如 `tabId`、URL 和标题
- 元数据，如 origin 和 user agent

当你需要更丰富的受控状态视图时使用此接口。轻量的运行/就绪视图请使用 `GET /tabs/{id}/state`。

## 列出已保存状态

```bash
pinchtab state list
```

在配置的状态目录中列出已保存的状态文件。

## 保存当前浏览器状态

```bash
pinchtab state save
pinchtab state save --name work-login
pinchtab state save --name checkout --tab <tabId>
pinchtab state save --name work-login --encrypt
```

注意：

- 省略 `--name` 时由 PinchTab 自动生成一个
- `--tab <id>` 从特定标签页而非活动/当前标签页捕获状态
- `--encrypt` 需要 `security.stateEncryptionKey`（或 `PINCHTAB_STATE_KEY` 环境变量，后者优先）；缺少时保存返回 `400`

## 加载已保存状态

```bash
pinchtab state load --name work-login
pinchtab state load --name work-log    # prefix match, newest match wins
pinchtab state load --name checkout --tab <tabId>
```

注意：

- `--name` 接受精确名称或前缀；精确匹配优先
- 前缀匹配解析到最近的匹配已保存状态
- 加载会先恢复已保存的 cookies，然后将已保存的 `localStorage` 和 `sessionStorage` 项写入目标标签页的当前页面，因此请先将该标签页导航到已保存的源
- 响应报告 `cookiesRestored`、`storageItemsRestored` 和 `origins`

## 查看已保存状态详情

```bash
pinchtab state show --name work-login
```

显示完整的已保存浏览器状态记录，包括存储的元数据和源存储负载。与 `load` 不同，`show` 需要精确名称。

## 删除已保存状态

```bash
pinchtab state delete --name work-login
```

从磁盘移除指定名称的状态文件。

## 清理旧的已保存状态文件

```bash
pinchtab state clean
pinchtab state clean --older-than 72
```

移除早于给定小时数的已保存状态文件。默认 `24`（零或负值也表示 `24`）。

## HTTP API

```text
GET    /state
GET    /state/list
GET    /state/show
POST   /state/save
POST   /state/load
DELETE /state
POST   /state/clean
```

| 路由 | 输入 |
| --- | --- |
| `GET /state` | `?tabId=`（可选） |
| `GET /state/show`、`DELETE /state` | `?name=`（必填，否则 `400`） |
| `POST /state/save` | `{"name", "encrypt", "tabId", "metadata"}`，字段可选但需要 JSON 主体（至少 `{}`）；返回 `name`、`path`、`cookies`（计数）、`origins`、`encrypted` |
| `POST /state/load` | `{"name", "tabId"}`，`name` 必填 |
| `POST /state/clean` | `{"olderThanHours"}`（需要 JSON 主体）；返回 `removed`、`olderThanHours`、`sessionsDir` |

`GET /state/list` 返回 `{states, count}`；已保存文件位于 `<stateDir>/sessions/` 下。

简短路由索引见 [Endpoints](../endpoints.md)。完整的受控浏览器状态视图用 `GET /state?tabId=<id>`，实时就绪/阻塞数据用 `GET /tabs/{id}/state`。

## 相关

- [Tabs](./tabs.md) 用于实时标签页操作和 `--tab <id>`
- [Sessions](./sessions.md) 用于代理/会话认证
- [Config](./config.md) 用于安全 flags 和存储目录
