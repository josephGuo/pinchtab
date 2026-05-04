# 展示

## 为您的代理提供浏览器

首先启动服务器和一个实例：

```bash
pinchtab server
#或
pinchtab daemon install
```

启动实例可能是可选的，取决于策略/配置。

```bash
curl -s -X POST http://127.0.0.1:9867/instances/start \
  -H "Content-Type: application/json" \
  -d '{"mode":"headless"}' | jq .
# 命令行界面 替代方案
pinchtab instance start
# 响应
{
  "id": "inst_0a89a5bb",
  "profileId": "prof_278be873",
  "profileName": "instance-1741400000000000000",
  "port": "9868",
  "mode": "headless",
  "headless": true,
  "status": "starting"
}
```

### 导航

```bash
curl -s -X POST http://127.0.0.1:9867/navigate \
  -H "Content-Type: application/json" \
  -d '{"url":"https://github.com/pinchtab/pinchtab"}' | jq .
# 命令行界面 替代方案
pinchtab nav https://github.com/pinchtab/pinchtab
# 响应
{
  "tabId": "CDP_TARGET_ID",
  "title": "GitHub - pinchtab/pinchtab",
  "url": "https://github.com/pinchtab/pinchtab"
}
```

### 快照

```bash
curl -s "http://127.0.0.1:9867/snapshot?filter=interactive" | jq .
# 命令行界面 替代方案
pinchtab snap -i -c
# 响应
{
  "nodes": [
    { "ref": "e0", "role": "link", "name": "Skip to content" },
    { "ref": "e1", "role": "link", "name": "GitHub Homepage" },
    { "ref": "e14", "role": "button", "name": "Search or jump to…" }
  ]
}
```

### 提取文本

```bash
curl -s http://127.0.0.1:9867/text | jq .
# 命令行界面 替代方案
pinchtab text
# 响应
{
  "text": "High-performance browser automation bridge and multi-instance orchestrator...",
  "title": "GitHub - pinchtab/pinchtab",
  "url": "https://github.com/pinchtab/pinchtab"
}
```

### 通过引用点击

```bash
curl -s -X POST http://127.0.0.1:9867/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"click","ref":"e14"}' | jq .
# 命令行界面 替代方案
pinchtab click e14
# 响应
{
  "success": true,
  "result": {
    "clicked": true
  }
}
```

### 截图

```bash
curl -s http://127.0.0.1:9867/screenshot > smoke.jpg
ls -lh smoke.jpg
# 命令行界面 替代方案
pinchtab ss -o smoke.jpg
# 响应
Saved smoke.jpg (55876 bytes)
```

### 导出 PDF

```bash
curl -s http://127.0.0.1:9867/pdf > smoke.pdf
ls -lh smoke.pdf
# 命令行界面 替代方案
pinchtab pdf -o smoke.pdf
# 响应
Saved smoke.pdf (1494657 bytes)
```

## 网页自动化工具

将 PinchTab 用作可脚本化的浏览器端点，用于可重复的网页任务。

### 填写表单字段

```bash
curl -s -X POST http://127.0.0.1:9867/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"fill","ref":"e3","text":"user@example.com"}' | jq .
# 命令行界面 替代方案
pinchtab fill e3 "user@example.com"
# 响应
{
  "success": true,
  "result": {
    "filled": "user@example.com"
  }
}
```

### 按键

```bash
curl -s -X POST http://127.0.0.1:9867/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"press","key":"Enter"}' | jq .
# 命令行界面 替代方案
pinchtab press Enter
# 响应
{
  "success": true,
  "result": {
    "pressed": "Enter"
  }
}
```

### 生成工件

```bash
curl -s http://127.0.0.1:9867/pdf > report.pdf
ls -lh report.pdf
# 命令行界面 替代方案
pinchtab pdf -o report.pdf
# 响应
Saved report.pdf (1494657 bytes)
```

```bash
curl -s http://127.0.0.1:9867/screenshot > page.jpg
ls -lh page.jpg
# 命令行界面 替代方案
pinchtab ss -o page.jpg
# 响应
Saved page.jpg (55876 bytes)
```

这适用于：

- 浏览器驱动的脚本
- 内容提取和报告
- 视觉检查和工件
- 需要本地浏览器端点的自动化工具

## 人工-代理开发表面

当 Chrome 已经在远程调试模式下运行时，PinchTab 可以附加到它并通过相同的 API 暴露它。

### 1. 以远程调试模式启动 Chrome

```bash
google-chrome --remote-debugging-port=9222
# 或在某些系统上：
# chromium --remote-debugging-port=9222
```

### 2. 读取浏览器 CDP URL

```bash
curl -s http://127.0.0.1:9222/json/version | jq .
# 响应
{
  "webSocketDebuggerUrl": "ws://127.0.0.1:9222/devtools/browser/abc123"
}
```

### 3. 将该浏览器附加到 PinchTab

```bash
CDP_URL=$(curl -s http://127.0.0.1:9222/json/version | jq -r '.webSocketDebuggerUrl')

curl -s -X POST http://127.0.0.1:9867/instances/attach \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"dev-chrome\",\"cdpUrl\":\"$CDP_URL\"}" | jq .
# 响应
{
  "id": "inst_abc12345",
  "profileId": "prof_def67890",
  "profileName": "dev-chrome",
  "attached": true,
  "cdpUrl": "ws://127.0.0.1:9222/devtools/browser/abc123",
  "status": "running"
}
```

### 4. 通过 PinchTab 检查它

```bash
curl -s http://127.0.0.1:9867/instances | jq .
# 命令行界面 替代方案
pinchtab instances
```

这在以下情况很有用：

- 您在真实的浏览器会话中开发
- 您希望代理检查您已经打开的页面
- 您不希望 PinchTab 启动单独的托管浏览器
- 您希望为托管和附加的浏览器工作使用一个本地 API