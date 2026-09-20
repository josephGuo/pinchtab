# 后台服务（守护进程）

PinchTab 可以在 macOS（`launchd`）和 Linux（`systemd`）上作为用户级后台服务（守护进程）运行。这确保 PinchTab 服务器始终可供你的代理使用，而无需打开一个终端窗口。

该工作流目前在 Windows 上不提供。Windows 二进制可用，但 Windows 支持有限且为尽力而为；在 Windows 上，更倾向于直接运行 `pinchtab server` 或 `pinchtab bridge`。

## 快速开始

先检查服务，再安装它：

```bash
pinchtab daemon
pinchtab daemon install
```

不带 action 时，`pinchtab daemon` 会打印当前状态、适用于该状态的命令以及最近日志。它不会打开选择器；请直接运行你想要的 action。`pinchtab daemon --json` 以 JSON 打印状态。

## 守护进程命令

| 命令 | 描述 |
|---------|-------------|
| `pinchtab daemon` | 显示状态摘要、接下来可用的命令以及最近日志（`status` 是别名）。 |
| `pinchtab daemon install` | 创建并启用后台服务文件。 |
| `pinchtab daemon start` | 若服务已停止，则启动后台服务。 |
| `pinchtab daemon stop` | 停止后台服务。 |
| `pinchtab daemon restart` | 重启服务（配置更改后很有用）。 |
| `pinchtab daemon uninstall` | 禁用并移除后台服务文件。 |

## 状态与诊断

`pinchtab daemon` 命令提供服务的综合概览：

- **服务状态**：显示 `.plist`（macOS）或 `.service`（Linux）文件是否已安装。
- **状态**：指示进程是 `running` 还是 `stopped`。
- **PID**：运行中服务器的进程 ID。
- **路径**：服务配置文件在你系统上的确切位置。
- **最近日志**：服务器输出的最后几行，帮助诊断问题（完整日志位于 `~/.pinchtab/logs/daemon.out.log` 和 `daemon.err.log`）。

## 手动安装

如果自动命令因权限问题或系统限制而失败，PinchTab 会提供针对你操作系统的手动说明。

当当前会话无法管理用户服务时，PinchTab 现在会在安装前快速失败。

典型情形：

- 没有可用 `systemctl --user` 会话的 Linux shell
- 没有活动 GUI `launchd` 域的 macOS shell

在这些情况下，请使用下面的手动步骤，或者改为在前台运行 `pinchtab server`。

### macOS（launchd）
服务文件：`~/Library/LaunchAgents/com.pinchtab.pinchtab.plist`

1. 创建 plist 文件（错误输出会指明要创建的路径）。
2. 注册并启动：
   ```bash
   launchctl bootstrap gui/$(id -u) ~/Library/LaunchAgents/com.pinchtab.pinchtab.plist
   ```

### Linux（systemd）
服务文件：`~/.config/systemd/user/pinchtab.service`

1. 创建 unit 文件。
2. 重新加载并启用：
   ```bash
   systemctl --user daemon-reload
   systemctl --user enable --now pinchtab.service
   ```

## 冲突检测

如果你在守护进程已在同一端口上运行时，尝试在前台启动 PinchTab 服务器（`pinchtab server`），PinchTab 会检测到冲突、警告你并退出，以防止端口绑定错误。
