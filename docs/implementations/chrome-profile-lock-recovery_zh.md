# Chrome 配置文件锁恢复

Chrome 在其用户数据目录中使用 `SingletonLock` 文件来防止多个实例同时共享同一个配置文件。如果 PinchTab 或 Chrome 崩溃，这个锁文件（以及相关的 `SingletonSocket` 和 `SingletonCookie` 文件）可能会被遗留下来，导致下一次 PinchTab 启动失败，错误信息为：

> "The profile appears to be in use by another Chromium process"

本文档解释了 PinchTab 如何识别、验证并从这些过时的锁中恢复，同时确保多实例安全。

## 恢复机制

恢复过程采用多层方法来区分真正活动的配置文件和崩溃实例遗留的过时配置文件。

### 1. 检测

在 `internal/bridge/init.go` 中的初始化过程中，如果 Chrome 启动失败，PinchTab 会检查错误消息中是否有配置文件锁的特征：
- `The profile appears to be in use by another Chromium process`
- `The profile appears to be in use by another Chrome process`
- `process_singleton_posix.cc`（表示 ProcessSingleton 逻辑中的失败）

### 2. 验证和 PinchTab PID 锁

为了安全地清除锁，PinchTab 必须确定没有*其他*活动的 PinchTab 实例当前正在使用该配置文件。

- **`pinchtab.pid`**：当桥接启动时，它会将自己的 PID 写入 `$PROFILE_DIR/pinchtab.pid`。
- **所有权检查**：在清除任何 Chrome 锁之前，PinchTab 会读取此文件。
  - 如果文件中的 PID 仍在运行**并且**被验证为 `pinchtab` 进程（通过检查其命令行参数），它会假设另一个 PinchTab 实例处于活动状态，**不会**触及锁。
  - 此验证可防止 PID 重用的问题，其中死亡的 PinchTab 实例的 PID 被重新分配给不同的进程。
  - 如果 PID 未运行或不是 PinchTab 进程，则认为之前的实例已"死亡"，配置文件有资格恢复。

### 3. 无头回退

如果无头 PinchTab 实例无法获取请求的配置文件目录的锁（因为另一个 PinchTab 实例确实在使用它），它会自动回退到创建唯一的临时配置文件目录。这允许多个无头桥接并发运行，即使它们都默认为相同的配置文件路径，同时仍保持隔离和安全。

### 4. 过时进程终止

即使之前的 PinchTab 实例已死亡，孤立的 Chrome 进程仍可能持有配置文件锁。

- **进程列表**：PinchTab 扫描系统进程列表，查找使用相同 `--user-data-dir` 启动的任何进程。
- **积极清理**：如果 `pinchtab.pid` 检查确认没有活动所有者，PinchTab 会向与该配置文件关联的任何孤立 Chrome 进程发送 `SIGKILL`。这是必要的，因为如果 Chrome 认为另一个进程甚至部分存活，其内部的"单例"逻辑可能会非常顽固。

### 5. 锁文件删除

过时进程终止后，PinchTab 会从配置文件目录中删除以下文件：
- `SingletonLock`
- `SingletonSocket`
- `SingletonCookie`

### 6. 自动重试

清除过时状态后，`InitChrome` 会自动重试启动序列一次。这使得恢复对用户和 API 调用者是透明的（例如，第一次 `/health` 检查会在短暂的内部恢复延迟后成功）。

## 实现细节

逻辑分布在以下组件中：

- **`internal/bridge/profile_lock.go`**：核心逻辑，用于检测、PID 锁管理（`AcquireProfileLock`）和过时文件删除。
- **`internal/bridge/profile_lock_pid_*.go`**：平台特定的实现，用于 PID 探测和进程终止（支持类 Unix 系统和 Windows）。
- **`internal/bridge/init.go`**：在 `startChromeWithRecovery` 中编排重试逻辑。
- **`internal/server/bridge.go`**：通过信号处理确保干净关闭，以防止锁被遗留。

## 多实例安全

通过将 Chrome 级别的 `SingletonLock` 与应用程序级别的 `pinchtab.pid` 相结合，PinchTab 实现了：
1. **安全性**：它永远不会终止被健康的 PinchTab 实例使用的浏览器。
2. **弹性**：它在崩溃或电源故障后自动"自我修复"。
3. **透明度**：用户不需要手动 `rm -rf` 配置文件目录来修复"正在使用"的错误。