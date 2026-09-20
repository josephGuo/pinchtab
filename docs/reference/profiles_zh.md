# Profiles（配置文件）

Profiles 是浏览器用户数据目录。它们存放 cookies、localStorage、历史记录和其他持久浏览器状态。

在 PinchTab 中：

- 即使没有实例运行，profile 也存在
- 一个 profile 同一时间最多只能有一个活动的受管实例
- profile ID 和名称都有用，但某些端点特别要求 profile ID

## 列出 Profiles

```bash
curl http://localhost:9867/profiles
# Response: JSON array (see below)

# CLI Alternative (human-readable by default)
pinchtab profiles
# Output (tab-separated): prof_278be873  work

pinchtab profiles --json              # Full JSON response
```

`pinchtab profiles` 是从命令行界面查看可用 profile 最简单的方式。被隔离的 profile 单独列在活动 profile 之后，并附其大小。

响应形状（每个字段始终存在；`GET /profiles/{id}` 返回同一对象）：

```json
[
  {
    "id": "prof_278be873",
    "name": "work",
    "path": "/path/to/profiles/work",
    "pathExists": true,
    "created": "2026-02-27T20:37:13.599055326Z",
    "lastUsed": "2026-03-01T09:12:44Z",
    "diskUsage": 534952089,
    "sizeMB": 510.17,
    "running": false,
    "quarantined": false,
    "temporary": false,
    "source": "created",
    "chromeProfileName": "",
    "accountEmail": "",
    "accountName": "",
    "hasAccount": false,
    "useWhen": "Use for work accounts",
    "description": ""
  }
]
```

注意：

- `GET /profiles` 默认排除临时自动生成的实例 profile
- 用 `GET /profiles?all=true` 包含临时 profile（`"temporary": true`）

## 获取单个 Profile

```bash
curl http://localhost:9867/profiles/prof_278be873
# Response
{
  "id": "prof_278be873",
  "name": "work",
  "path": "/path/to/profiles/work",
  "pathExists": true,
  "created": "2026-02-27T20:37:13.599055326Z",
  "lastUsed": "2026-03-01T09:12:44Z",
  "diskUsage": 534952089,
  "sizeMB": 510.17,
  "running": false,
  "quarantined": false,
  "temporary": false,
  "source": "created",
  "chromeProfileName": "Your Chrome",
  "accountEmail": "admin@pinchtab.com",
  "accountName": "Luigi",
  "hasAccount": true,
  "useWhen": "Use for work accounts",
  "description": ""
}
```

`GET /profiles/{id}` 接受 profile ID 或 profile 名。

## 创建 Profile

```bash
curl -X POST http://localhost:9867/profiles \
  -H "Content-Type: application/json" \
  -d '{"name":"scraping-profile","description":"Used for scraping","useWhen":"Use for ecommerce scraping"}'
# Response
{
  "status": "created",
  "id": "prof_0f32ae81",
  "name": "scraping-profile"
}
```

注意：

- `name` 必填；`description` 和 `useWhen` 可选
- `POST /profiles` 和 `POST /profiles/create` 都可用于创建 profile
- 命令行界面形式为 `pinchtab profiles create <name>`，打印新的 `id` 和 `name`

## 更新 Profile

```bash
curl -X PATCH http://localhost:9867/profiles/prof_278be873 \
  -H "Content-Type: application/json" \
  -d '{"description":"Updated description","useWhen":"Updated usage note"}'
# Response
{
  "status": "updated",
  "id": "prof_278be873",
  "name": "work"
}
```

你也可以重命名 profile：

```bash
curl -X PATCH http://localhost:9867/profiles/prof_278be873 \
  -H "Content-Type: application/json" \
  -d '{"name":"work-renamed"}'
```

重要：

- `PATCH /profiles/{id}` 需要 profile ID
- 在该路径中用 profile 名返回错误
- 重命名会改变生成的 profile ID，因为 ID 从名称派生

`PATCH /profiles/meta` 更新主体中 `name` 指定的 profile 的 `description` 和/或 `useWhen`，应答 `{"status":"updated","name":"..."}`。

## 删除 Profile

```bash
curl -X DELETE http://localhost:9867/profiles/prof_278be873
# Response
{
  "status": "deleted",
  "id": "prof_278be873",
  "name": "work"
}
```

`DELETE /profiles/{id}` 也需要 profile ID。加 `?force=true` 可删除仍有实例的 profile；此时响应把该实例命名为 `orphanedInstance`。

## 按 Profile 启动或停止

启动某 profile 的活动实例：

```bash
curl -X POST http://localhost:9867/profiles/prof_278be873/start \
  -H "Content-Type: application/json" \
  -d '{"headless":true}'
# Response
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

停止某 profile 的活动实例：

```bash
curl -X POST http://localhost:9867/profiles/prof_278be873/stop
# Response
{
  "status": "stopped",
  "id": "prof_278be873",
  "name": "work"
}
```

对这些编排器路由，路径可以是 profile ID 或 profile 名。start 应答 `201` 带实例对象（见 [Instances](./instances.md)），其中同时含 `mode` 和 `headless`。其主体接受 `headless`（默认 `false`，因此省略它启动有头浏览器）、`port`、`securityPolicy`、`browser` 和 `fallbackTargets`。

## 检查 Profile 是否有运行中的实例

```bash
curl http://localhost:9867/profiles/prof_278be873/instance
# Response
{
  "name": "work",
  "exists": true,
  "running": true,
  "status": "running",
  "port": "9868",
  "id": "inst_ea2e747f"
}
```

实例存在时（此时 `id` 已设置）`status` 为 `running` 或 `starting`，不存在时为 `stopped`，profile 不存在时为 `missing`（带 `exists: false` 和 `message`）。该路由始终应答 `200`。

## 其他 Profile 操作

### 重置 Profile

```bash
curl -X POST http://localhost:9867/profiles/prof_278be873/reset
```

此路由需要 profile ID，应答 `{"status":"reset","id":"...","name":"..."}`。

### 导入 Profile

```bash
curl -X POST http://localhost:9867/profiles/import \
  -H "Content-Type: application/json" \
  -d '{"name":"imported-profile","sourcePath":"/path/to/existing/profile"}'
```

`name` 和 `sourcePath` 必填；`description` 和 `useWhen` 可选。响应为 `{"status":"imported","name":"..."}`。

### 修剪被隔离的 Profiles

```bash
curl -X POST http://localhost:9867/profiles/prune
curl -X POST http://localhost:9867/profiles/prune \
  -H "Content-Type: application/json" \
  -d '{"confirm":true}'
# CLI Alternative
pinchtab profiles prune              # list what would be removed
pinchtab profiles prune --confirm    # remove it
```

不带 `confirm`（主体或 `?confirm=true`）时不删除任何内容，响应列出将被移除的。`profile`（主体或查询）将其限定到单个隔离目录。响应为 `{"removed":<bool>,"count":N,"totalBytes":N,"profiles":[{"name","path","bytes"}]}`。

### 获取日志

```bash
curl http://localhost:9867/profiles/prof_278be873/logs
curl 'http://localhost:9867/profiles/work/logs?limit=50'
```

`logs` 接受 profile ID 或 profile 名。结果从该 profile 的活动存储派生。

### 获取分析

```bash
curl http://localhost:9867/profiles/prof_278be873/analytics
curl http://localhost:9867/profiles/work/analytics
```

`analytics` 同样接受 profile ID 或 profile 名。它从 `/api/activity` 使用的相同活动数据计算。

## 相关页面

- [Instances](./instances.md)
- [Tabs](./tabs.md)
- [Config](./config.md)
