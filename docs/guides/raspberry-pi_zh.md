# 树莓派

只要 Chromium 或 Chrome 可用，PinchTab 就可以在树莓派上运行。当前实现不需要特定于树莓派的功能标志，但由于内存有限，它确实受益于保守的默认值。

## 推荐基线

- 树莓派 OS 或 ARM64 上的 Ubuntu
- 尽可能使用 64 位用户空间
- 本地安装 Chromium
- 默认无头模式
- 较小板上的低标签页计数

## 安装 Chromium

在树莓派 OS 上：

```bash
sudo apt update
sudo apt install -y chromium-browser
```

验证二进制文件：

```bash
which chromium-browser
which chromium
```

如果自动检测错过它，请在配置中设置二进制路径：

```bash
pinchtab config set browser.binary /usr/bin/chromium-browser
pinchtab server
```

## 安装 PinchTab

使用平台的正常 PinchTab 安装路径，或从该存储库构建二进制文件：

```bash
go build -o pinchtab ./cmd/pinchtab
```

然后启动它：

```bash
./pinchtab server
```

## 适合树莓派的配置

创建配置文件并将大多数设置保存在那里，而不是依赖旧的环境变量。

示例：

```json
{
  "browser": {
    "binary": "/usr/bin/chromium-browser",
    "extraFlags": "--disable-gpu --disable-dev-shm-usage"
  },
  "instanceDefaults": {
    "mode": "headless",
    "maxTabs": 5,
    "blockImages": true,
    "blockAds": true
  },
  "profiles": {
    "baseDir": "/home/pi/.pinchtab/profiles",
    "defaultProfile": "default"
  }
}
```

使用它运行：

```bash
PINCHTAB_CONFIG=/home/pi/.pinchtab/config.json ./pinchtab
```

## 无头模式与有头模式

对于大多数树莓派工作负载，保持默认值：

```json
{
  "instanceDefaults": {
    "mode": "headless"
  }
}
```

如果您使用桌面会话并想要可见的浏览器，请切换到：

```json
{
  "instanceDefaults": {
    "mode": "headed"
  }
}
```

有头模式消耗更多 RAM，通常最好仅用于调试。

## 存储

如果 SD 卡很小或很慢，将配置文件存储移动到更大的驱动器：

```json
{
  "profiles": {
    "baseDir": "/mnt/usb/pinchtab-profiles",
    "defaultProfile": "default"
  },
  "server": {
    "stateDir": "/mnt/usb/pinchtab-state"
  }
}
```

这是当前支持的重定位数据的方法。继续使用嵌套的配置键，而不是较旧的扁平配置文件。

## 作为服务运行

示例 `systemd` 单元：

```ini
[Unit]
Description=PinchTab Browser Service
After=network.target

[Service]
Type=simple
User=pi
WorkingDirectory=/home/pi
ExecStart=/home/pi/pinchtab
Environment=PINCHTAB_CONFIG=/home/pi/.pinchtab/config.json
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

然后：

```bash
sudo systemctl daemon-reload
sudo systemctl enable pinchtab
sudo systemctl start pinchtab
sudo systemctl status pinchtab
```

## 性能提示

- 在 1 GB 和 2 GB 板上保持 `instanceDefaults.maxTabs` 较低
- 优先使用无头模式
- 对于大量抓取的工作负载，阻止图像和广告
- 如果 SD 卡是瓶颈，将配置文件移动到更快的外部存储
- 如果经常遇到 OOM 情况，请谨慎添加交换空间

## 故障排除

### Chrome 二进制文件未找到

在配置中设置 `browser.binary`：

```bash
pinchtab config set browser.binary /usr/bin/chromium-browser
```

### 内存不足

在配置中减少工作负载：

```json
{
  "instanceDefaults": {
    "maxTabs": 3,
    "blockImages": true,
    "mode": "headless"
  }
}
```

### 端口已被使用

在配置中更改端口：

```bash
pinchtab config set server.port 9868
./pinchtab server
```