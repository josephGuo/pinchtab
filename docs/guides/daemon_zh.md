# 后台服务（守护进程）

PinchTab 可以在 macOS（`launchd`）和 Linux（`systemd`）上作为用户级后台服务（守护进程）运行。这确保 PinchTab 服务器始终可供您的代理使用，而无需打开终端窗口。

此工作流目前在 Windows 上不可用。Windows 二进制文件可用，但 Windows 支持有限且尽力而为；在 Windows 上，建议直接运行 `pinchtab server` 或 `pinchtab bridge`。

![守护进程状态和选择器](../media/daemon-status.png)

## 快速开始

正常入口点是：

```bash
pinchtab
```

然后从菜单中选择 `Daemon`，或直接管理服务：

```bash
pinchtab daemon
```

在交互式终端中不带参数运行时，此命令会显示当前状态并打开一个常用操作的选择器。

## 守护进程命令

| 命令 | 描述 |
|---------|-------------|
| `pinchtab daemon` | 显示状态摘要、最近日志，并打开交互式选择器。 |
| `pinchtab daemon install` | 创建并启用后台服务文件。 |
| `pinchtab daemon start` | 如果服务已停止，则启动后台服务。 |
| `pinchtab daemon stop` | 停止后台服务。 |
| `pinchtab daemon restart` | 重启服务（配置更改后有用）。 |
| `pinchtab daemon uninstall` | 禁用并移除后台服务文件。 |

## 状态和诊断

`pinchtab daemon` 命令提供服务的综合概览：

- **服务状态**：显示 `.plist`（macOS）或 `.service`（Linux）文件是否已安装。
- **状态**：指示进程是 `active (running)` 还是 `stopped`。
- **PID**：运行中服务器的进程 ID。
- **路径**：服务配置文件在系统上的确切位置。
- **最近日志**：服务器输出的最后几行，以帮助诊断问题。

## 手动安装

如果由于权限问题或系统限制导致自动命令失败，PinchTab 会提供针对您的操作系统的手动说明。

当当前会话无法管理用户服务时，PinchTab 现在会在安装前快速失败。

典型情况：

- 没有工作 `systemctl --user` 会话的 Linux shell
- 没有活动 GUI `launchd` 域的 macOS shell

在这些情况下，请使用下面的手动步骤，或在前台运行 `pinchtab server`。

### macOS（launchd）
服务文件：`~/Library/LaunchAgents/com.pinchtab.pinchtab.plist`

1. 创建 plist 文件（PinchTab 会在错误时提供内容）。
2. 注册并启动：
   ```bash
   launchctl bootstrap gui/$(id -u) ~/Library/LaunchAgents/com.pinchtab.pinchtab.plist
   ```

### Linux（systemd）
服务文件：`~/.config/systemd/user/pinchtab.service`

1. 创建单元文件。
2. 重新加载并启用：
   ```bash
   systemctl --user daemon-reload
   systemctl --user enable --now pinchtab.service
   ```

## 冲突检测

如果您尝试在前台启动 PinchTab 服务器（`pinchtab server`），而守护进程已经在同一端口上运行，PinchTab 会检测到冲突，警告您并退出以防止端口绑定错误。