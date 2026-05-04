# 配置文件

配置文件是浏览器用户数据目录。它们存储 cookie、本地存储、历史记录和其他持久浏览器状态。

在 PinchTab 中：

- 即使没有实例运行，配置文件也存在
- 一个配置文件一次最多可以有一个活动的管理实例
- 配置文件 ID 和名称都有用，但某些端点特别需要配置文件 ID

## 列出配置文件

```bash
curl http://localhost:9867/profiles
# 响应: JSON 数组（见下文）

# 命令行界面 替代方案（默认人类可读）
pinchtab profiles
# 输出: prof_278be873  work

pinchtab profiles --json              # 完整 JSON 响应
```

`pinchtab profiles` 是从 命令行界面 查看可用配置文件的最简单方法。

响应形状：

```json
[
  {
    "id": "prof_278be873",
    "name": "work",
    "created": "2026-02-27T20:37:13.599055326Z",
    "diskUsage": 534952089,
    "sizeMB": 510.17,
    "running": false,
    "source": "created",
    "useWhen": "Use for work accounts",
    "description": ""
  }
]
```

注意：

- `GET /profiles` 默认排除临时自动生成的实例配置文件
- 使用 `GET /profiles?all=true` 包含临时配置文件

## 获取单个配置文件

```bash
curl http://localhost:9867/profiles/prof_278be873
# 响应
{
  "id": "prof_278be873",
  "name": "work",
  "path": "/path/to/profiles/work",
  "pathExists": true,
  "created": "2026-02-27T20:37:13.599055326Z",
  "diskUsage": 534952089,
  "sizeMB": 510.17,
  "source": "created",
  "chromeProfileName": "Your Chrome",
  "accountEmail": "admin@pinchtab.com",
  "accountName": "Luigi",
  "hasAccount": true,
  "useWhen": "Use for work accounts",
  "description": ""
}
```

`GET /profiles/{id}` 接受配置文件 ID 或配置文件名称。

## 创建配置文件

```bash
curl -X POST http://localhost:9867/profiles \
  -H "Content-Type: application/json" \
  -d '{"name":"scraping-profile","description":"Used for scraping","useWhen":"Use for ecommerce scraping"}'
# 响应
{
  "status": "created",
  "id": "prof_0f32ae81",
  "name": "scraping-profile"
}
```

注意：

- 没有 `pinchtab profile create` 命令行界面 命令
- `POST /profiles` 和 `POST /profiles/create` 都可以用于创建配置文件

## 更新配置文件

```bash
curl -X PATCH http://localhost:9867/profiles/prof_278be873 \
  -H "Content-Type: application/json" \
  -d '{"description":"Updated description","useWhen":"Updated usage note"}'
# 响应
{
  "status": "updated",
  "id": "prof_278be873",
  "name": "work"
}
```

你也可以重命名配置文件：

```bash
curl -X PATCH http://localhost:9867/profiles/prof_278be873 \
  -H "Content-Type: application/json" \
  -d '{"name":"work-renamed"}'
```

重要：

- `PATCH /profiles/{id}` 需要配置文件 ID
- 在该路径中使用配置文件名称会返回错误
- 重命名会更改生成的配置文件 ID，因为 ID 是从名称派生的

## 删除配置文件

```bash
curl -X DELETE http://localhost:9867/profiles/prof_278be873
# 响应
{
  "status": "deleted",
  "id": "prof_278be873",
  "name": "work"
}
```

`DELETE /profiles/{id}` 也需要配置文件 ID。

## 按配置文件启动或停止

启动配置文件的活动实例：

```bash
curl -X POST http://localhost:9867/profiles/prof_278be873/start \
  -H "Content-Type: application/json" \
  -d '{"headless":true}'
# 响应
{
  "id": "inst_ea2e747f",
  "profileId": "prof_278be873",
  "profileName": "work",
  "port": "9868",
  "mode": "headless",
  "headless": true,
  "status": "starting"
}
```

停止配置文件的活动实例：

```bash
curl -X POST http://localhost:9867/profiles/prof_278be873/stop
# 响应
{
  "status": "stopped",
  "id": "prof_278be873",
  "name": "work"
}
```

对于这些编排器路由，路径可以是配置文件 ID 或配置文件名称。返回的实例对象现在包含 `mode` 和 `headless`。

## 检查配置文件是否有运行中的实例

```bash
curl http://localhost:9867/profiles/prof_278be873/instance
# 响应
{
  "name": "work",
  "running": true,
  "status": "running",
  "port": "9868",
  "id": "inst_ea2e747f"
}
```

## 其他配置文件操作

### 重置配置文件

```bash
curl -X POST http://localhost:9867/profiles/prof_278be873/reset
```

此路由需要配置文件 ID。

### 导入配置文件

```bash
curl -X POST http://localhost:9867/profiles/import \
  -H "Content-Type: application/json" \
  -d '{"name":"imported-profile","sourcePath":"/path/to/existing/profile"}'
```

### 获取日志

```bash
curl http://localhost:9867/profiles/prof_278be873/logs
curl 'http://localhost:9867/profiles/work/logs?limit=50'
```

`logs` 接受配置文件 ID 或配置文件名称。结果从该配置文件的活动存储中派生。

### 获取分析

```bash
curl http://localhost:9867/profiles/prof_278be873/analytics
curl http://localhost:9867/profiles/work/analytics
```

`analytics` 也接受配置文件 ID 或配置文件名称。它是根据 `/api/activity` 使用的相同活动数据计算的。

## 相关页面

- [实例](./instances.md)
- [标签页](./tabs.md)
- [配置](./config.md)