# Docker 本地测试

本页是一份在本地测试当前 Docker 配置的实用清单。

它覆盖两条路径：

- 默认的托管配置流程，即容器拥有 `/data/.pinchtab/config.json`（`HOME=/data`；在 Linux 上默认配置位于 `~/.pinchtab` 下）
- 显式配置流程，即你挂载自己的 `config.json` 并设置 `PINCHTAB_CONFIG`

## 托管配置流程

构建并启动本地 Compose 服务：

```bash
docker compose up --build -d
docker compose logs -f pinchtab
```

检查生效的配置路径和持久化配置：

```bash
docker exec pinchtab pinchtab config path
docker exec pinchtab sh -lc 'cat /data/.pinchtab/config.json'
```

预期结果：

- 配置路径是 `/data/.pinchtab/config.json`
- 持久化配置中的 `server.bind` 为 `0.0.0.0`（入口点设置它，以便端口发布生效）
- 如果首次启动时生成了令牌或传入了令牌，则存在令牌

验证配置的绑定地址：

```bash
docker exec pinchtab pinchtab config get server.bind
```

预期结果：`0.0.0.0`（由入口点在首次启动时设置）

验证跨重启的持久性：

```bash
docker compose down
docker compose up -d
docker exec pinchtab sh -lc 'cat /data/.pinchtab/config.json'
```

## 显式 `PINCHTAB_CONFIG` 流程

创建一个本地配置文件，例如 `./tmp/config.json`：

```json
{
  "server": {
    "bind": "0.0.0.0",
    "port": "9867",
    "token": "local-test-token"
  }
}
```

用只读挂载该配置的方式运行容器（要测试本地构建而不是已发布镜像，先跑 `docker build -t pinchtab/pinchtab .`）：

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
- 挂载的文件按原样使用
- 容器入口点不会重写自定义配置

## 出问题时检查什么

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
docker exec pinchtab sh -lc 'cat /data/.pinchtab/config.json'
```

## 自动化 Docker E2E

E2E 套件（`tests/e2e/docker-compose.yml`、`tests/e2e/docker-compose-multi.yml`）构建的是同一个 `Dockerfile` 并运行 `docker-entrypoint.sh`，但始终设置 `PINCHTAB_CONFIG`——因此它只覆盖显式配置流程。上面的托管配置流程仅由本手动清单覆盖。

```bash
./dev e2e                     # extended suite (go run ./tests/tools/runner e2e --suite extended)
./dev e2e smoke               # smoke tier
./dev e2e api <text>          # one suite, filtered by scenario file name (--suite api --filter <text>)
./dev e2e test "<name>"       # runs only the first start_test whose name contains <name>
```

## 当前注意事项

Docker 运行时路径现在负责 `--no-sandbox` 兼容性（在检测到容器时于启动时加入）。不要把它放进 `browser.extraFlags`；配置校验会拒绝它。

`docker-entrypoint.sh` 会在 `$XDG_CONFIG_HOME/pinchtab/config.json`（`/data/.config/...`）检查已有配置，但二进制读写的是 `/data/.pinchtab/config.json`，因此"首次启动"块每次启动都会跑：`server.bind` 被重新设为 `0.0.0.0`，并且当设置了 `PINCHTAB_TOKEN` 时，`server.token` 会被它覆盖。
