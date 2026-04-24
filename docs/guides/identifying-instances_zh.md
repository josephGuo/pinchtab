# 识别实例

当您在正常浏览器旁边运行 PinchTab 时，区分其 Chrome 进程的最简单方法是结合三个信号：

- 专用的 Chrome 二进制文件名
- 可识别的命令行标志
- PinchTab 仪表板和实例元数据

## 1. 使用不同的 Chrome 二进制文件名

如果您将 Chrome 或 Chromium 复制到自定义文件名，该文件名会出现在进程列表中。

```bash
# macOS 示例
cp "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" /usr/local/bin/pinchtab-chrome
chmod +x /usr/local/bin/pinchtab-chrome

# 在 config.json 中设置
pinchtab config set browser.binary /usr/local/bin/pinchtab-chrome
pinchtab server
```

现在，进程列表如 `ps -axo pid,command | rg pinchtab-chrome` 为您提供了一种快速识别 PinchTab 启动的浏览器的方法。

## 2. 添加可识别的 Chrome 标志

使用 `instanceDefaults.userAgent` 作为可见的进程标记，并为安全的非降低安全性的标志保留 `browser.extraFlags`：

```json
{
  "instanceDefaults": {
    "userAgent": "PinchTab-Automation/1.0"
  },
  "browser": {
    "extraFlags": "--ash-no-nudges --disable-focus-on-load"
  }
}
```

这些标志会出现在 Chrome 命令行中，这使得进程检查更容易：

```bash
ps -axo pid,command | rg 'PinchTab-Automation|user-data-dir'
```

当您想区分角色如“爬虫”、“监控”或“调试”时使用此方法。

不要将降低安全性或 PinchTab 拥有的标志放入 `browser.extraFlags`。例如，`--user-agent=...`、`--no-sandbox` 和隐身/运行时拥有的标志会被拒绝。

## 3. 使用配置文件路径作为标识符

每个管理的配置文件都位于配置的配置文件基目录下。默认情况下，这是 `profiles/` 下的操作系统特定的 PinchTab 配置目录。

PinchTab 启动的 Chrome 进程包含一个 `--user-data-dir=...` 参数，指向该配置文件位置。这通常是确认浏览器进程属于 PinchTab 而不是您的个人 Chrome 配置文件的最快方法。

## 4. 使用仪表板获取最可靠的视图

打开仪表板：

- `http://localhost:9867/`
- 或 `http://localhost:9867/dashboard`

仪表板和实例 API 显示：

- 实例 ID
- 配置文件 ID 和配置文件名称
- 分配的端口
- 无头模式与有头模式
- 当前状态

如果您需要基于 API 的视图而不是 UI：

```bash
curl http://localhost:9867/instances
```

## 实际组合

对于大多数设置，以下组合就足够了：

1. 通过配置中的 `browser.binary` 将 PinchTab 指向重命名的 Chrome 二进制文件
2. 在配置中添加可识别的 `instanceDefaults.userAgent` 标记或安全的 `browser.extraFlags` 标记
3. 在仪表板中验证配置文件路径或实例 ID

## Docker

相同的方法在容器中也有效：

- 如果需要覆盖捆绑的浏览器路径，在配置中设置 `browser.binary`
- 仅在 `browser.extraFlags` 中放置安全的识别标志
- 从 API 或仪表板检查实例列表，而不是仅依赖容器内的进程名称