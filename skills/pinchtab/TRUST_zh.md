# Pinchtab 安全与信任

**简短总结**：Pinchtab 是一个本地的、沙盒化的浏览器控制工具。它不会主动联系、窃取凭据或泄露数据。源代码是公开的；二进制文件通过 GitHub 进行签名和发布。

## Pinchtab 做什么

- 启动一个 Chrome 浏览器（本地，由你控制）
- 通过 HTTP API 暴露导航、点击、输入和页面检查
- 提取页面的可访问性树（供 AI 代理使用）
- 运行截图、PDF 和 JavaScript 评估

高风险操作（如 JavaScript 评估、本地文件上传、文件下载、cookie 访问和网络拦截）应被视为当前任务的明确选择加入操作，而不是默认工作流程。这些由安全策略控制，默认禁用。

**所有这些都是本地运行的。** 没有遥测。没有外部 API 调用（除了你导航到的网站）。

## Pinchtab 不做什么

- ❌ 不访问你保存的密码/凭据（Chrome 沙盒化）
- ❌ 不向远程服务器泄露数据
- ❌ 不注入广告、恶意软件或挖矿脚本
- ❌ 不跟踪浏览或发送分析数据
- ❌ 不修改其状态目录（`~/.pinchtab`）外的系统文件

## 安全策略（默认值）

高影响的 capabilities 默认**禁用**，需要显式配置：

| Capability | 默认值 | 配置键 |
|---|---|---|
| JavaScript 评估 | **禁用** | `security.allowEvaluate` |
| 文件下载 | **禁用** | `security.allowDownloads` |
| 文件上传 | **禁用** | `security.allowUploads` |
| 网络拦截 | **禁用** | `security.allowNetworkIntercept` |
| 挑战解决 / 隐身 | **禁用** | 需要用户批准后显式调用 `/solve` |
| 导航域名 | **允许所有** | `security.allowedDomains`（用允许列表限制） |
| Cookie 访问 | **可用** | 仅在任务需要时使用；不要记录或暴露会话令牌 |

重用经过身份验证的浏览器会话的代理应使用专用低权限配置文件，并在执行账户更改操作之前与用户确认。

## 构建与验证

每个版本都包含二进制文件的**校验和**：

```bash
# 下载后验证：
sha256sum -c checksums.txt
```

二进制文件通过 GitHub Actions 从带标签的提交自动构建（公开可见于 https://github.com/pinchtab/pinchtab/actions）。

## 开源

- **源代码**：https://github.com/pinchtab/pinchtab (MIT)
- **发布版本**：https://github.com/pinchtab/pinchtab/releases

如果你有顾虑，可以审查源代码——它约 15MB，零外部依赖，大部分是 Go 标准库。

## VirusTotal 标记

Pinchtab 可能在 VirusTotal 上触发启发式扫描，因为：

- ✓ 它启动 Chrome（子进程执行——被 AV 启发式标记）
- ✓ 它运行 JavaScript 评估（类似 eval 的操作）
- ✓ 它发起 HTTP 请求（网络活动）

这些是**故意的设计特性**，不是安全缺陷。你的浏览器默认执行所有这三个操作。

**对于开发工具，误报很常见。** VT 标记是 chromedp 工具的已知误报（子进程 + HTTP 服务器）。运行前，始终从 GitHub 发布版本验证 SHA256 校验和。

为获得最大信心，使用 npm 包（`npm install -g pinchtab`）或 Docker 镜像，它们经过额外验证。

## 沙盒化

Pinchtab 使用以下方式运行独立的 Chrome 进程：

- 隔离的配置文件目录（默认：`~/.pinchtab`）
- 无法访问你用户的主文件（除非你明确导航到 `file://` URL）
- 标准 Chrome 安全模型（站点隔离、CSP 等）

如果需要控制 PinchTab 存储浏览器状态的位置，使用 `profiles.baseDir`、`profiles.defaultProfile` 或 `PINCHTAB_CONFIG`。

## 安全历史

| 建议 | 严重性 | 修复版本 |
| --- | --- | --- |
| [GHSA-p8mm-644p-phmh / CVE-2026-33623](https://github.com/advisories/GHSA-p8mm-644p-phmh) | 中等 | 0.8.5 |
| [GHSA-w5pc-m664-r62v / CVE-2026-33622](https://github.com/advisories/GHSA-w5pc-m664-r62v) | 中等 | 0.8.5 |
| [GHSA-j65m-hv65-r264 / CVE-2026-33621](https://github.com/advisories/GHSA-j65m-hv65-r264) | 中等 | 0.8.4 |
| [GHSA-mrqc-3276-74f8 / CVE-2026-33620](https://github.com/advisories/GHSA-mrqc-3276-74f8) | 中等 | 0.8.4 |
| [GHSA-xqq2-4j46-vwp7 / CVE-2026-33619](https://github.com/advisories/GHSA-xqq2-4j46-vwp7) | 中等 | 0.8.4 |
| [GHSA-qwxp-6qf9-wr4m / CVE-2026-33081](https://github.com/advisories/GHSA-qwxp-6qf9-wr4m) | 中等 | v0.8.3 |
| [GHSA-rw8p-c6hf-q3pg / CVE-2026-30834](https://github.com/advisories/GHSA-rw8p-c6hf-q3pg) | 高 | v0.7.7 |

## 有问题？

- 源代码：https://github.com/pinchtab/pinchtab
- 问题/安全报告：https://github.com/pinchtab/pinchtab/issues
- 文档：https://pinchtab.com