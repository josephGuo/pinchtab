# 贡献指南

这是 PinchTab 的权威贡献者和开发指南。

## 系统要求

### 最低要求

| 要求 | 版本 | 用途 |
|------------|---------|---------|
| Go | 1.25+ | 构建语言 |
| golangci-lint | 最新版 | 代码检查（预提交钩子必需） |
| Chrome/Chromium | 最新版 | 浏览器自动化 |
| macOS、Linux 或 WSL2 | 当前版本 | 操作系统支持 |

对于仪表板工作，使用 Bun 1.2+。
较旧的 Bun 版本在使用 `--frozen-lockfile` 进行干净安装时，会在已签入的 `dashboard/bun.lock` 上失败。

### 推荐设置

- **macOS**：使用 Homebrew 进行包管理
- **Linux**：apt（Debian/Ubuntu）或 yum（RHEL/CentOS）
- **WSL2**：完整的 Linux 环境（非 WSL1）

---

## 快速开始

**最快的开始方式：**

```bash
# 1. 克隆
 git clone https://github.com/pinchtab/pinchtab.git
cd pinchtab

# 2. 运行医生（验证环境，安装前提示）
./dev doctor

# 3. 构建并运行
go build ./cmd/pinchtab
./pinchtab
```

**示例输出：**
```
  🦀 Pinchtab Doctor
  Verifying and setting up development environment...

Go Backend
  ✓ Go 1.26.0
  ✗ golangci-lint
    Required for pre-commit hooks and CI.
    Install golangci-lint via brew? [y/N] y
    ✓ golangci-lint installed
  ✓ Git hooks
  ✓ Go dependencies

Dashboard (React/TypeScript)
  ✓ Node.js 22.15.1
  · Bun not found
    Optional — used for fast dashboard builds.
    Install Bun? [y/N] n
    curl -fsSL https://bun.sh/install | bash

Summary

  · 1 warning(s)
```

医生在安装任何东西之前会请求确认。
如果您拒绝，它会显示手动安装命令。

---

## 第一部分：先决条件

### 安装 Go

**macOS（Homebrew）：**
```bash
brew install go
go version  # 验证：go version go1.25.0
```

**Linux（Ubuntu/Debian）：**
```bash
sudo apt update
sudo apt install -y golang-go git build-essential
go version
```

**Linux（RHEL/CentOS）：**
```bash
sudo yum install -y golang git
go version
```

**或从以下地址下载：** https://go.dev/dl/

### 安装 golangci-lint（必需）

预提交钩子必需：

**macOS/Linux：**
```bash
brew install golangci-lint
```

**或通过 Go：**
```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

验证：
```bash
golangci-lint --version
```

### 安装 gotestsum（推荐）

推荐用于 `./dev test unit` 使用的更干净的本地单元测试输出：

```bash
go install gotest.tools/gotestsum@latest
```

### 安装 Chrome/Chromium

**macOS（Homebrew）：**
```bash
brew install chromium
```

**Linux（Ubuntu/Debian）：**
```bash
sudo apt install -y chromium-browser
```

**Linux（RHEL/CentOS）：**
```bash
sudo yum install -y chromium
```

### 自动设置

克隆后，运行医生来验证和设置您的环境：

```bash
git clone https://github.com/pinchtab/pinchtab.git
cd pinchtab
./dev doctor
```

医生检查您的环境并**在安装前询问**：
- Go 1.25+ 和 golangci-lint（提供 `brew install` 或 `go install`）
- Git 钩子（复制预提交钩子）
- Go 依赖项（`go mod download`）
- Node.js、Bun 和仪表板依赖项（可选，用于仪表板开发）

随时运行 `./dev doctor` 来验证或修复您的环境。

---

## 第二部分：构建项目

### 简单构建

```bash
go build -o pinchtab ./cmd/pinchtab
```

**它会：**
- 编译 Go 源代码
- 生成二进制文件：`./pinchtab`
- 花费约 30-60 秒

> **注意：** 这只构建 Go 服务器。仪表板会显示一个
> "未构建" 占位符。要包含完整的 React 仪表板，请使用
> `./dev build` 代替 —— 它会构建仪表板，编译 Go，并
> 一步运行服务器。或者在 `go build` 之前运行 `./scripts/build-dashboard.sh`。

**验证：**
```bash
ls -la pinchtab
./pinchtab --version
```

---

## 第三部分：运行服务器

### 启动（无头模式）

```bash
./pinchtab
```

**预期输出：**
```
🦀 PINCH! PINCH! port=9867
auth disabled (set PINCHTAB_TOKEN to enable)
```

### 启动（有头模式）

```bash
BRIDGE_HEADLESS=false ./pinchtab
```

在前台打开 Chrome。

### 后台运行

```bash
nohup ./pinchtab > pinchtab.log 2>&1 &
tail -f pinchtab.log  # 查看日志
```

---

## 第四部分：快速测试

### 健康检查

```bash
curl http://localhost:9867/health
```

### 尝试 命令行界面

```bash
./pinchtab quick https://pinchtab.com
./pinchtab nav https://pinchtab.com
./pinchtab snap
```

---

## 开发

### 运行测试

```bash
go test ./...                              # 仅单元测试
go test ./... -v                           # 详细输出
go test ./... -v -coverprofile=coverage.out
go tool cover -html=coverage.out           # 查看覆盖率
./dev e2e                                 # 运行默认的 E2E 发布套件
./dev e2e docker                          # 构建本地镜像并运行 Docker 冒烟测试
./dev e2e pr                              # 运行 API + 命令行界面 + 基础架构基本测试
./dev e2e api                             # 运行 API 基本测试
./dev e2e 命令行界面                             # 运行 命令行界面 基本测试
./dev e2e infra                           # 运行基础架构基本测试
./dev e2e api-extended                    # 运行 API 扩展测试（多实例）
./dev e2e 命令行界面-extended                    # 运行 命令行界面 扩展测试
./dev e2e infra-extended                  # 运行基础架构扩展测试（多实例）
```

### 开发者工具包（`dev`）

所有开发脚本都可通过 `./dev` 访问：

```bash
./dev              # 交互式选择器（如果安装了 gum，使用 gum，否则使用编号回退）
./dev check        # 直接运行命令
./dev test unit    # 支持子命令
./dev --help       # 列出所有命令
```

![dev 交互式菜单](../media/dev-menu.jpg)

**可用命令：**

| 命令 | 描述 |
|---------|-------------|
| `check` | 所有检查（Go + 仪表板） |
| `check go` | 仅 Go 检查 |
| `check dashboard` | 仅仪表板检查 |
| `check security` | Gosec 安全扫描 |
| `check docs` | 验证文档 JSON |
| `format dashboard` | 在仪表板源上运行 Prettier |
| `test` | 运行所有测试 |
| `test unit` | 仅单元测试 |
| `test dashboard` | 仅仪表板测试 |
| `e2e` | 运行默认的 E2E 发布套件（所有扩展测试） |
| `e2e docker` | 构建本地镜像并运行 Docker 冒烟测试 |
| `e2e pr` | 运行 PR E2E 套件（`api` + `命令行界面` + `infra` 基本测试） |
| `e2e api` | 运行 API 基本测试 |
| `e2e 命令行界面` | 运行 命令行界面 基本测试 |
| `e2e infra` | 运行基础架构基本测试 |
| `e2e api-extended` | 运行 API 扩展测试（多实例） |
| `e2e 命令行界面-extended` | 运行 命令行界面 扩展测试 |
| `e2e infra-extended` | 运行基础架构扩展测试（多实例） |
| `e2e release` | 运行发布 E2E 元套件（所有扩展测试） |
| `build` | 构建应用程序 |
| `dev` | 构建并运行应用程序 |
| `run` | 运行应用程序 |
| `binary` | 构建本地发布风格的二进制文件 |
| `doctor` | 设置开发环境 |

要使用花哨的交互式选择器，安装 [gum](https://github.com/charmbracelet/gum)：`brew install gum`

**提示：** 将此添加到 `~/.zshrc` 以使用 `dev` 而无需 `./`：
```bash
dev() { if [ -x "./dev" ]; then ./dev "$@"; else echo "dev not found in current directory"; return 1; fi }
```

### 代码质量

```bash
./dev check              # 完整的非测试检查（推荐）
./dev format dashboard   # 修复仪表板格式
 gofmt -w .                # 格式化代码
 golangci-lint run         # 代码检查
./dev doctor             # 验证环境
```

### Git 钩子

Git 钩子由 `./dev doctor`（或 `./scripts/install-hooks.sh`）安装。它们在每次提交时运行：
- `gofmt` — 格式检查
- `golangci-lint` — 代码检查
- `prettier` — 仪表板格式化

要手动重新安装钩子：
```bash
./scripts/install-hooks.sh
```

### 开发工作流

```bash
# 1. 设置（首次）
./dev doctor

# 2. 创建功能分支
git checkout -b feat/my-feature

# 3. 进行更改
# ... 编辑文件 ...

# 4. 推送前运行检查
./dev check

# 5. 提交（钩子自动运行）
git commit -m "feat: description"

# 6. 推送
git push origin feat/my-feature
```

**注意：** Git 钩子会在提交时自动格式化和检查您的代码。如果检查失败，提交会被阻止。

---

## 持续集成

工作流遵循命名约定：

| 前缀 | 目的 | 示例 |
|--------|---------|---------|
| `ci-*` | PR/推送时的自动检查 | `ci-go.yml` → **CI / Go** |
| `reusable-*` | 构建块（仅 `workflow_call`） | `reusable-e2e.yml` → **Reusable / E2E** |
| `release-*` | 发布管道 | `release.yml` → **Release** |

### CI 检查

在拉取请求和/或推送到 `main` 时自动运行：

| 工作流 | 触发器 | 检查内容 |
|----------|----------|----------------|
| **CI / Go** | PR + 推送 | gofmt、vet、构建、测试、覆盖率、代码检查、安全 |
| **CI / Dashboard** | PR + 推送（仪表板路径） | TypeScript、ESLint、Prettier、测试、构建 |
| **CI / Docs** | PR + 推送（文档路径） | docs.json 参考验证 |
| **CI / npm** | PR（npm 路径）+ 标签推送 | npm 包验证 |
| **CI / E2E** | PR（快速套件）+ 手动（完整套件） | 基于 Docker 的端到端测试 |
| **CI / Branch Naming** | PR | 分支命名约定执行 |

### 发布管道

| 工作流 | 触发器 | 执行内容 |
|----------|---------|--------------|
| **Release** | 手动 | 运行所有检查 + E2E → 手动批准门 → 创建标签 → 发布二进制文件、npm、Docker 和技能 |
| **Release / Manual Publish** | 手动 | 发布现有标签作为恢复路径 |

在 **Release** 中，E2E 和 Docker 冒烟测试失败是非阻塞的 —— 它们在批准摘要中显示，以便您可以决定是否继续。核心检查（Go、仪表板、文档、npm、发布预演）必须通过才能显示批准门。

---

## 作为 命令行界面 安装

### 从源代码

```bash
go build -o ~/go/bin/pinchtab ./cmd/pinchtab
```

然后在任何地方使用：
```bash
pinchtab help
pinchtab --version
```

### 通过 npm（已发布构建）

```bash
npm install -g pinchtab
pinchtab --version
```

---

## 资源

- **GitHub 仓库：** https://github.com/pinchtab/pinchtab
- **Go 文档：** https://golang.org/doc/
- **Chrome DevTools 协议：** https://chromedevtools.github.io/devtools-protocol/
- **Chromedp 库：** https://github.com/chromedp/chromedp

---

## 故障排除

### 环境问题

**第一步：** 运行医生来验证您的设置：
```bash
./dev doctor
```

这将确切地告诉您缺少什么或配置错误。

### 常见问题

**"Go 版本太旧"**
- 从 https://go.dev/dl/ 安装 Go 1.25+
- 验证：`go version`

**"golangci-lint: command not found"**
- 安装：`brew install golangci-lint`
- 或：`go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`

**"Git 钩子在提交时不运行"**
- 运行：`./scripts/install-hooks.sh`
- 或：`./dev doctor`（提示安装）

**"Chrome not found"**
- 安装 Chromium：`brew install chromium`（macOS）
- 或：`sudo apt install chromium-browser`（Linux）

**"Port 9867 already in use"**
- 检查：`lsof -i :9867`
- 停止其他实例或使用不同端口：`BRIDGE_PORT=9868 ./pinchtab`

**构建失败**
1. 验证依赖项：`go mod download`
2. 清理缓存：`go clean -cache`
3. 重新构建：`go build ./cmd/pinchtab`

---

## 支持

有问题？检查：
1. 首先运行 `./dev doctor`
2. 所有依赖项都已安装且版本正确？
3. 端口 9867 可用？
4. 检查日志：`tail -f pinchtab.log`

有关指南和示例，请参阅 `docs/`。