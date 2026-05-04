# Docker 本地测试

本文档是在本地测试当前 Docker 设置的实用清单。

它涵盖了两种路径：

- 默认的托管配置流程，其中容器拥有 `/data/.config/pinchtab/config.json`
- 显式配置流程，其中您挂载自己的 `config.json` 并设置 `PINCHTAB_CONFIG`

## 托管配置流程

构建并启动本地 Compose 服务：

```bash
docker compose up --build -d
docker compose logs -f pinchtab
```

检查有效的配置路径和持久化配置：

```bash
docker exec pinchtab pinchtab config path
docker exec pinchtab sh -lc 'cat /data/.config/pinchtab/config.json'
```

预期结果：

- 配置路径为 `/data/.config/pinchtab/config.json`
- 持久化配置中的 `server.bind` 保持为 `127.0.0.1`
- 如果在首次启动时生成或传入了令牌，则存在令牌

验证配置绑定地址：

```bash
docker exec pinchtab pinchtab config get server.bind
```

预期结果：`0.0.0.0`（由入口点在首次启动时设置）

验证重启后的持久性：

```bash
docker compose down
docker compose up -d
docker exec pinchtab sh -lc 'cat /data/.config/pinchtab/config.json'
```

## 显式 `PINCHTAB_CONFIG` 流程

创建本地配置文件，例如 `./tmp/config.json`：

```json
{
  "server": {
    "bind": "0.0.0.0",
    "port": "9867",
    "token": "local-test-token"
  }
}
```

使用挂载的只读配置运行容器：

```bash
docker run --rm -d \
  --name pinchtab-test \
  -p 127.0.0.1:9867:9867 \
  -e PINCHTAB_CONFIG=/config/config.json \
  -v "$PWD/tmp/config.json:/config/config.json:ro" \
  -v pinchtab-data:/data \
  --shm-size=2g \
  pinchtab/pinchtab
```

验证显式配置路径和认证：

```bash
docker exec pinchtab-test pinchtab config path
docker exec pinchtab-test sh -lc 'cat /config/config.json'
curl -H 'Authorization: Bearer local-test-token' http://127.0.0.1:9867/health
```

预期结果：

- `pinchtab config path` 报告 `/config/config.json`
- 按原样使用挂载的文件
- 容器入口点不会重写自定义配置

## 当出现问题时要检查什么

容器日志：

```bash
docker logs pinchtab
docker logs pinchtab-test
```

配置路径：

```bash
docker exec pinchtab pinchtab config path
docker exec pinchtab-test pinchtab config path
```

持久化配置内容：

```bash
docker exec pinchtab sh -lc 'cat /data/.config/pinchtab/config.json'
```

## 当前注意事项

Docker 运行时路径现在拥有 `--no-sandbox` 兼容性。不要将其放在 `browser.extraFlags` 中。