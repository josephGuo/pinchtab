# 配置文件管理

运行 `pinchtab` 时，配置文件通过端口 9867 上的 HTTP API 进行管理。

## 列出配置文件

```bash
curl http://localhost:9867/profiles
```

返回包含 `id`、`name`、`accountEmail`、`useWhen` 等的配置文件数组。

## 启动配置文件

```bash
# 自动分配端口（推荐）
curl -X POST http://localhost:9867/profiles/<ID>/start

# 使用特定端口和无头模式
curl -X POST http://localhost:9867/profiles/<ID>/start \
  -H 'Content-Type: application/json' \
  -d '{"port": "9868", "headless": true}'

# 简短别名
curl -X POST http://localhost:9867/start/<ID>
```

返回包含分配的 `port` 的实例信息。使用该端口进行所有后续 API 调用。

## 停止配置文件

```bash
curl -X POST http://localhost:9867/profiles/<ID>/stop

# 简短别名
curl -X POST http://localhost:9867/stop/<ID>
```

## 检查实例状态

```bash
# 通过配置文件 ID（推荐）
curl http://localhost:9867/profiles/<ID>/instance

# 通过配置文件名称
curl http://localhost:9867/profiles/My%20Profile/instance
```

## 通过现有配置文件启动

```bash
curl -X POST http://localhost:9867/profiles \
  -H 'Content-Type: application/json' \
  -d '{"name": "work"}'

curl -X POST http://localhost:9867/instances/start \
  -H 'Content-Type: application/json' \
  -d '{"profileId": "work", "port": "9868"}'
```

## 使用配置文件的命令行界面用法

命令行界面还没有配置文件子命令 — 使用 `curl` 进行配置文件管理。配置文件实例运行后，使用 `--server` 标志将命令行界面指向它：

```bash
# 获取实例端口，然后使用命令行界面
pinchtab --server http://localhost:9868 snap -i
```

## 典型代理流程

```bash
# 1. 列出配置文件
PROFILES=$(curl -s http://localhost:9867/profiles)

# 2. 启动配置文件（自动分配端口）
INSTANCE=$(curl -s -X POST http://localhost:9867/profiles/$PROFILE_ID/start)
PORT=$(echo $INSTANCE | jq -r .port)

# 3. 使用实例
curl -X POST http://localhost:$PORT/navigate -H 'Content-Type: application/json' \
  -d '{"url": "https://mail.google.com"}'
curl http://localhost:$PORT/snapshot?maxTokens=4000

# 4. 完成后停止
curl -s -X POST http://localhost:9867/profiles/$PROFILE_ID/stop
```

## 配置文件 ID

每个配置文件获得一个稳定的 12 字符十六进制 ID（名称的 SHA-256，截断），存储在 `profile.json` 中。ID 是 URL 安全的，永远不会改变 — 在自动化中使用它们而不是名称。

## 有头模式

有头模式 = Pinchtab 管理的真实可见 Chrome 窗口。

- 人类可以登录、通过 2FA/captcha、验证状态
- 代理对同一运行实例调用 HTTP API
- 会话状态保存在配置文件目录中（cookie/存储会保留）

推荐的人类 + 代理流程：

```bash
# 人类启动 PinchTab 并设置配置文件
pinchtab

# 代理解析配置文件端点
PINCHTAB_BASE_URL="$(pinchtab connect <profile-name>)"
curl "$PINCHTAB_BASE_URL/health"
```