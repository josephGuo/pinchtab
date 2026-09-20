# 展示

## 为您的代理提供浏览器

首先启动服务器和一个实例：

```bash
pinchtab server
#or
pinchtab daemon install
```

启动实例可能是可选的，取决于策略/配置。

HTTP API 需要服务器 token。本页的 `curl` 示例为简洁起见省略了它；在
`export PINCHTAB_TOKEN=$(pinchtab config token --stdout)` 之后，再补上
`-H "Authorization: Bearer $PINCHTAB_TOKEN"`。命令行界面替代方案会从你的配置读取 token。

```bash
curl -s -X POST http://127.0.0.1:9867/instances/start \
  -H "Content-Type: application/json" \
  -d '{"mode":"headless"}' | jq .
# CLI Alternative
pinchtab instance start
# Response
{
  "id": "inst_0a89a5bb",
  "profileId": "prof_278be873",
  "profileName": "instance-1741400000000000000-9f3c2a1b",
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
# CLI Alternative
pinchtab nav https://github.com/pinchtab/pinchtab
# Response
{
  "tabId": "CDP_TARGET_ID",
  "title": "GitHub - pinchtab/pinchtab",
  "url": "https://github.com/pinchtab/pinchtab"
}
```

### 快照

```bash
curl -s "http://127.0.0.1:9867/snapshot?filter=interactive" | jq .
# CLI Alternative
pinchtab snap -i -c
# Response
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
# CLI Alternative
pinchtab text
# Response
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
# CLI Alternative
pinchtab click e14
# Response
{
  "success": true,
  "result": {
    "clicked": true
  }
}
```

### 截图

```bash
curl -s "http://127.0.0.1:9867/screenshot?raw=true" > smoke.jpg
ls -lh smoke.jpg
# CLI Alternative
pinchtab screenshot -o smoke.jpg
# Response
Saved smoke.jpg (55876 bytes)
```

### 导出 PDF

```bash
curl -s "http://127.0.0.1:9867/pdf?raw=true" > smoke.pdf
ls -lh smoke.pdf
# CLI Alternative
pinchtab pdf -o smoke.pdf
# Response
Saved smoke.pdf (1494657 bytes)
```

## 网页自动化工具

将 PinchTab 用作可脚本化的浏览器端点，用于可重复的网页任务。

### 填写表单字段

```bash
curl -s -X POST http://127.0.0.1:9867/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"fill","ref":"e3","text":"user@example.com"}' | jq .
# CLI Alternative
pinchtab fill e3 "user@example.com"
# Response
{
  "success": true,
  "result": {
    "filled": true,
    "len": 16
  }
}
```

### 按键

```bash
curl -s -X POST http://127.0.0.1:9867/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"press","key":"Enter"}' | jq .
# CLI Alternative
pinchtab press Enter
# Response
{
  "success": true,
  "result": {
    "pressed": "Enter"
  }
}
```

### 生成产物

```bash
curl -s "http://127.0.0.1:9867/pdf?raw=true" > report.pdf
ls -lh report.pdf
# CLI Alternative
pinchtab pdf -o report.pdf
# Response
Saved report.pdf (1494657 bytes)
```

```bash
curl -s "http://127.0.0.1:9867/screenshot?raw=true" > page.jpg
ls -lh page.jpg
# CLI Alternative
pinchtab screenshot -o page.jpg
# Response
Saved page.jpg (55876 bytes)
```

这适用于：

- 浏览器驱动的脚本
- 内容提取和报告
- 视觉检查和产物
- 需要本地浏览器端点的自动化工具

## 人机协作开发接口

当 Chrome 已经在远程调试模式下运行时，PinchTab 可以附加到它并通过相同的 API 暴露它。

附加默认禁用；先用 `pinchtab config set security.attach.enabled true` 启用它并重启服务器。

### 1. 以远程调试模式启动 Chrome

```bash
google-chrome --remote-debugging-port=9222
# Or on some systems:
# chromium --remote-debugging-port=9222
```

### 2. 读取浏览器 CDP URL

```bash
curl -s http://127.0.0.1:9222/json/version | jq .
# Response
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
# Response
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
# CLI Alternative
pinchtab instance list
```

这在以下情况很有用：

- 您在真实的浏览器会话中开发
- 您希望代理检查您已经打开的页面
- 您不希望 PinchTab 启动单独的托管浏览器
- 您希望为托管和附加的浏览器工作使用一个本地 API