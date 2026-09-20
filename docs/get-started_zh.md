# 入门指南

在几分钟内让 PinchTab 运行起来，从零到浏览器自动化。

本指南涵盖默认的本地设置。如果您计划在 localhost 之外发布端口、绑定到非环回接口或运行远程或分布式拓扑，请将其视为高级部署，并先阅读 [安全指南](guides/security.md)。

---

## 安装

### 选项 1：一键安装

**macOS / Linux**

```bash
curl -fsSL https://pinchtab.com/install.sh | bash
```

然后验证：

```bash
pinchtab --version
```

### 选项 2：npm

**要求：** Node.js 18+

```bash
npm install -g pinchtab
pinchtab --version
```

### 选项 3：Docker

**要求：** Docker

```bash
docker run -d --name pinchtab -p 127.0.0.1:9867:9867 pinchtab/pinchtab
docker exec pinchtab pinchtab health
```

### 选项 4：从源代码构建

**要求：** Go 1.26+、Git、Chrome/Chromium

```bash
git clone https://github.com/pinchtab/pinchtab.git
cd pinchtab
./dev doctor
go build -o pinchtab ./cmd/pinchtab
./pinchtab --version
```

**[完整构建与贡献指南 ->](guides/contributing.md)**

> **macOS 升级提示：** 当前 PinchTab 使用 `~/.pinchtab/config.json` 作为默认配置路径。如果你在
> `~/Library/Application Support/pinchtab/config.json` 仍留有较旧的配置，请把它移动或合并到
> `~/.pinchtab/config.json`。

## 平台支持

PinchTab 的主要测试工作流是本地 macOS 和 Linux。

Windows 二进制文件可用，但 Windows 支持目前有限且尽力而为，因为该项目在那里没有相同级别的测试覆盖率。在 Windows 上，首选使用 `pinchtab server` 直接运行，而不是期望完整的守护进程工作流。

## Shell 自动完成

安装后，您可以从 命令行界面 生成 shell 自动完成：

```bash
# Generate and install zsh completions
pinchtab completion zsh > "${fpath[1]}/_pinchtab"

# Generate bash completions
pinchtab completion bash > /etc/bash_completion.d/pinchtab

# Generate fish completions
pinchtab completion fish > ~/.config/fish/completions/pinchtab.fish
```

---

## 快速开始

正常流程是：

1. 启动服务器
2. 启动实例
3. 导航
4. 检查或操作

### 步骤 1：启动服务器

```bash
pinchtab server
# Response (log lines)
time=... level=INFO msg=orchestration strategy=always-on allocation=fcfs
time=... level=INFO msg="dashboard started" port=9867
```

服务器运行在 `http://127.0.0.1:9867`。
您可以在 `http://127.0.0.1:9867` 或 `http://127.0.0.1:9867/dashboard` 打开仪表板。

HTTP API 需要服务器 token（`server.token`，首次运行时生成）。为下面的 `curl` 示例导出一次：

```bash
export PINCHTAB_TOKEN=$(pinchtab config token --stdout)
```

### 步骤 2：启动您的第一个实例

```bash
curl -s -X POST http://127.0.0.1:9867/instances/start \
  -H "Authorization: Bearer $PINCHTAB_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"mode":"headless"}' | jq .
# CLI Alternative
pinchtab instance start
# Response
{
  "id": "inst_0a89a5bb",
  "profileId": "prof_278be873",
  "profileName": "instance-1741400000000000000-9f3c2a1b",
  "port": "9868",
  "mode": "headless",
  "headless": true,
  "status": "starting"
}
```

### 步骤 3：导航

```bash
curl -s -X POST http://127.0.0.1:9867/navigate \
  -H "Authorization: Bearer $PINCHTAB_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"url":"https://github.com/pinchtab/pinchtab"}' | jq .
# CLI Alternative
pinchtab nav https://github.com/pinchtab/pinchtab
# Response
{
  "tabId": "CDP_TARGET_ID",
  "title": "GitHub - pinchtab/pinchtab",
  "url": "https://github.com/pinchtab/pinchtab"
}
```

### 步骤 4：检查页面

```bash
curl -s -H "Authorization: Bearer $PINCHTAB_TOKEN" "http://127.0.0.1:9867/snapshot?filter=interactive" | jq .
# CLI Alternative
pinchtab snap -i -c
# Response
{
  "nodes": [
    { "ref": "e0", "role": "link", "name": "Skip to content" },
    { "ref": "e14", "role": "button", "name": "Search or jump to…" }
  ]
}
```

现在您有一个工作的 PinchTab 服务器、一个运行中的浏览器实例和一个已导航的标签页。

---

## 故障排除

### 连接被拒绝

```bash
pinchtab health
```

如果失败，启动服务器：

```bash
pinchtab server
```

### 端口已被使用

```bash
pinchtab config set server.port 9868
pinchtab server
```

### 未找到 Chrome

```bash
# macOS
brew install chromium

# Linux (Ubuntu/Debian)
sudo apt install chromium-browser

# Custom Chrome binary (set in config)
pinchtab config set browser.binary /path/to/chrome
```

---

## 获取帮助

- [GitHub Issues](https://github.com/pinchtab/pinchtab/issues)
- [GitHub Discussions](https://github.com/pinchtab/pinchtab/discussions)