# Chrome 用户数据目录与 Profile

PinchTab 使用专用的**用户数据目录**（Profile）来管理 Chrome 实例。本文说明这些目录如何解析、锁如何管理，以及如何处理并行浏览器实例。

## Profile 解析

PinchTab 把 Chrome 的 `--user-data-dir`（即 Profile）推导为 `profiles.baseDir` 拼接 `profiles.defaultProfile`（见 `internal/config/config_load.go` 中的 `finalizeProfileConfig`）：

1. **命名 Profile**：当为某个命名 Profile 启动实例时（仪表板、`pinchtab instance start --profile <name>` 或 API），编排器会写入子进程配置，让 `profiles.baseDir` / `profiles.defaultProfile` 指向该 Profile 的目录（见 `internal/orchestrator/child_config.go` 中的 `buildChildFileConfig`）。
2. **默认 Profile**：否则 `profiles.baseDir` 默认为 `<server.stateDir>/profiles`，`profiles.defaultProfile` 默认为 `default`，在 Linux/macOS 上即 `~/.pinchtab/profiles/default`。

没有用于任意 Profile 路径的配置项，而且配置校验会拒绝 `browser.extraFlags` 中的 `--user-data-dir`。

## 单例模型

Chrome 对每个用户数据目录强制实行**单例**模型。这意味着：

* 在任意给定时刻，只有**一个** Chrome 进程可以使用某个特定 Profile 目录。
* 如果第二个进程尝试使用同一目录，Chrome 要么启动失败，要么（在某些配置下）尝试在已存在的进程里打开一个新窗口。

PinchTab 额外加了一层保护：在 Profile 目录内放一个 `pinchtab.pid` 文件，确保只有一个 PinchTab 实例管理某个特定 Profile。

## 新无头模式与并行实例

随着**新无头模式**（`--headless=new`）的引入，Chrome 的 Profile 共享行为变得更严格：

* **不可共享**：不能在多个并发实例之间复用同一个 `--user-data-dir`。
* **必须用独立目录**：并行浏览器**必须**使用独立目录，以避免随机启动错误和锁冲突。

### 无头自动回退

为支持并行自动化任务，PinchTab 为无头实例实现了**自动回退**：
1. 如果一个无头实例试图使用已被另一个 PinchTab 进程锁定的 Profile，它会**自动创建一个唯一的临时目录**（在系统临时目录下，例如 `/tmp`，名为 `pinchtab-profile-*`）。
2. 这让你可以并行跑多个无头任务，而无需手动管理 Profile 路径。

### 手动并行（有头模式）

在**有头模式**下，PinchTab *不会*自动回退到临时目录（以避免意外丢失用户会话数据）。如果你需要并行跑多个有头浏览器，必须：
* 每个实例用一个不同的命名 Profile，或
* 为每台服务器把 `profiles.baseDir` / `profiles.defaultProfile` 指向各自独立的目录。

## 面向 AI 代理的最佳实践

构建使用 PinchTab 的代理时，请遵循以下指引：

* **持久化**：如果你需要浏览器在多次会话之间记住登录、cookies 或历史，使用命名 Profile（例如 `agent-alpha`、`agent-beta`）。
* **隔离**：对于一次性任务或高并发抓取，依赖默认的无头模式——发生冲突时它会自动处理目录隔离。
* **清理**：如果你手动创建临时目录，确保任务完成后清理它们，以免填满磁盘。

## 故障排查

如果你看到错误 `"The profile appears to be in use by another Chromium process"`：
1. **检查活动实例**：确保没有另一个 PinchTab 或 Chrome 进程已经在使用该 Profile。
2. **陈旧锁**：如果没有活动进程，当浏览器启动因该错误失败时，PinchTab 会自动清除陈旧的 `SingletonLock` / `SingletonSocket` / `SingletonCookie` 文件，然后重试一次。
3. **手动修复**：在极少数情况下，你可能需要手动从 Profile 目录删除 `SingletonLock` 文件。

PinchTab 如何从崩溃中恢复的更多细节，见 [Chrome Profile 锁恢复](./chrome-profile-lock-recovery.md)。
