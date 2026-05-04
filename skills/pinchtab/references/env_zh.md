# PinchTab 环境变量

此参考故意较窄。

对于代理工作流程，大多数运行时行为应该通过 `config.json` 或 `pinchtab config` 命令进行配置，而不是环境变量。

## 代理相关变量

| 变量 | 典型用法 | 注意 |
|---|---|---|
| `PINCHTAB_TOKEN` | 向受保护的服务器验证命令行界面或 MCP 请求 | 作为 `Authorization: Bearer ...` 发送 |
| `PINCHTAB_CONFIG` | 覆盖配置文件路径 | 自动化时优先于此而不是临时环境变量覆盖 |

## 定位远程服务器

使用 `--server` 命令行界面标志而不是环境变量：

```bash
pinchtab --server http://192.168.1.50:9867 snap
pinchtab --server https://pinchtab.com snap
```

## 故意不列出的内容

- 浏览器调优通常应放在 `config.json` 中，而不是临时环境变量中。
- 内部进程接线和继承的环境变量直通是实现细节，不是技能契约的一部分。

## 推荐默认值

对于大多数代理任务，你唯一需要的变量是：

```bash
PINCHTAB_TOKEN=...
```

对于在同一标签页上的多步骤流程，运行一次 `pinchtab nav URL`，然后使用不限定的命令。匿名命令行界面调用在共享的本地状态文件中记住当前标签页。识别的调用者使用服务器端当前标签页状态：代理会话按会话限定当前标签页，`--agent-id` / `PINCHTAB_AGENT_ID` 在没有会话时按代理 ID 限定当前标签页。仅在你需要显式定位特定标签页时使用 `--tab <id>`。

或者使用代理会话来获取每个代理的身份和可撤销性：

```bash
PINCHTAB_SESSION=ses_...
```

设置 `PINCHTAB_SESSION` 后，命令行界面使用 `Authorization: Session <token>` 而不是 bearer auth。会话在服务器端映射到特定 agentId，可以独立撤销。

其他所有内容都应通过配置、配置文件、实例和 `--server` 标志处理。