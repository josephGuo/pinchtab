# 有头模式（仅手动本地）

PinchTab 支持有头 Chrome（可见的浏览器窗口），用于本地开发、调试和人在回路（human-in-the-loop）的工作流。各模式的取舍详见[无头与有头](../headless-vs-headed.md)。

本页是关于*运行*有头模式的操作指南。有头模式**仅支持手动本地**：PinchTab 不提供在容器或 CI 中使用有头模式的受支持路径。

## 为什么捆绑的 Docker 镜像仅支持无头

标准 PinchTab Docker 镜像（从仓库 `Dockerfile` 构建）在安装 Chromium 时不带任何 X11 / Wayland / Xvfb 包。这是有意为之：

- 该镜像面向无头自动化、代理和 CI。
- 让运行时不包含显示栈，可以保持镜像小巧、攻击面狭窄、启动开销低。
- 从容器内部驱动 X 服务器既脆弱、又依赖具体主机，PinchTab 不会把它作为一等部署目标来支持。

如果你尝试针对捆绑镜像启动一个有头实例，它会失败（没有显示器、没有 X 库）。这是预期行为。

如果你需要有头 Chrome，必须在镜像之外运行 PinchTab，或者自行构建一个带 X 栈和显示服务器的镜像——这两者都被视为手动本地工作流，而非官方支持的配置。

## 运行有头模式

### Linux 原生（推荐）

直接在 Linux 桌面会话上运行 PinchTab。

```bash
export DISPLAY=:0
pinchtab server &
pinchtab instance start --mode headed
```

这是最简单、最可靠的有头设置。Wayland 会话通常也能工作，因为 Chrome 会回退到 XWayland。

### 通过 SSH 进行 X 转发的 Linux

```bash
ssh -X user@workstation
# in that SSH session, DISPLAY already points at the forwarded display
pinchtab server --headed
```

Chrome 从 `pinchtab server` 进程继承 `DISPLAY`，而不是从调用 `pinchtab instance start` 的客户端继承，因此服务器必须在被转发的会话内部启动。转发可用，但交互使用时会有延迟；更倾向于直接在工作站本机上运行 PinchTab。

### 带 X11 转发的 Docker（手动，仅限 Linux 主机）

开箱即用不受支持，但在已经运行 X 服务器的 Linux 主机上可以实现。你必须自行构建一个镜像，加入 Chromium 所需的 X 客户端库，然后转发主机的 X 套接字：

```bash
# On the host, allow local containers to use the X server
xhost +local:

docker run --rm \
  -e DISPLAY="$DISPLAY" \
  -v /tmp/.X11-unix:/tmp/.X11-unix \
  -p 9867:9867 \
  your-pinchtab-image-with-x11
```

注意事项：

- 仅限 Linux 主机——`xhost` 和 `/tmp/.X11-unix` 在 macOS 或 Windows 主机上并不存在。
- `xhost +local:` 会削弱 X 访问控制；请将其限制为受信任用户。
- GPU 加速、字体、输入设备和剪贴板集成都与主机相关，行为可能与原生会话不同。
- 这**不是**捆绑镜像。你有责任维护自己叠加的 X 栈。

### 通过 XQuartz 的 macOS

XQuartz 可以在 macOS 上托管 X 客户端，包括运行在 Linux 容器中的 Chromium，但体验比较脆弱：

```bash
# One-time: install XQuartz from https://www.xquartz.org/
# In XQuartz preferences → Security: enable "Allow connections from network clients"
open -a XQuartz
xhost + 127.0.0.1

docker run --rm \
  -e DISPLAY=host.docker.internal:0 \
  -p 9867:9867 \
  your-pinchtab-image-with-x11
```

注意事项：

- 在启动容器之前 XQuartz 必须已经在运行。
- 性能明显差于原生 macOS 应用。
- 某些 Chrome 功能（音频、GPU、特定输入事件）可能无法工作。
- 做有头工作时，更倾向于在 macOS 上原生运行 PinchTab——主机上的 `pinchtab server` 加有头实例才是受支持的路径。

### 通过 WSLg 或 VcXsrv 的 Windows

- **WSLg**（Windows 11 / 较新版本的 Windows 10）：WSL2 内置了一个 Wayland/X 合成器。在你的 WSL 发行版内运行 PinchTab，有头 Chrome 无需额外配置即可出现在 Windows 桌面上。
- **VcXsrv / Xming**：在 Windows 上安装一个 X 服务器，针对回环接口关闭访问控制运行它，然后把 WSL 或容器的 `DISPLAY` 指向 Windows 主机 IP。较脆弱；更推荐 WSLg。

`pinchtab` 的原生 Windows 构建是尽力而为——更完整的 Windows 支持情况请参见[无头与有头](../headless-vs-headed.md#windows)。

## CI 中的有头模式

不受支持。无头镜像是 PinchTab 在 CI 中唯一测试的配置，捆绑的 Docker 镜像无法运行有头 Chrome。如果你需要验证仅在有头模式下的行为，请在一台带真实显示器的开发工作站上运行这些测试。

## 指纹差异（CloakBrowser 用户）

有头与无头 Chrome 的指纹可观察地不同——`navigator.webdriver` 处理方式不同、GPU 字符串不同、特性 flag 不同，默认窗口/视口尺寸也不同。CloakBrowser 能缩小差距，但无法让无头模式与同一台机器上的有头会话像素级一致。如果某个目标站点把你在笔记本上有头模式的流量当成人，而把容器里无头模式的流量当成机器人，那是指纹差异，不是 CloakBrowser 的 bug——请在你打算部署的模式下验证你的流程。

## 相关指南

- [cloakbrowser.md](cloakbrowser.md)——CloakBrowser 配置与故障排查
- [docker.md](docker.md)——捆绑的无头镜像与 CloakBrowser 冒烟镜像（均仅无头）
- [attach-chrome.md](attach-chrome.md)——通过 CDP 附加到单独启动的有头 Chrome 或 CloakBrowser
- [security.md](security.md)——本地与远程部署的安全模型
