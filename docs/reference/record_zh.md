# Record（录制）

将浏览器活动录制为视频文件。支持 GIF（纯 Go 实现）、WebM 和 MP4（需要 ffmpeg）。

```bash
# Start recording
curl -X POST http://localhost:9867/record/start \
  -H "Content-Type: application/json" \
  -d '{"format":"gif","fps":5,"quality":80}'

# Check status
curl http://localhost:9867/record/status

# Stop: encoding runs in the background into <stateDir>/recordings/
curl -X POST http://localhost:9867/record/stop -d '{}'
# {"status":"encoding","path":"<stateDir>/recordings/rec_<timestamp>.gif","format":"gif","frames":62,"hint":"..."}

# Stop without encoding
curl -X POST http://localhost:9867/record/stop -d '{"discard":true}'
# {"status":"discarded","format":"gif","frames":62}
```

服务器选择输出路径；轮询 `/record/status` 直到 `state` 为 `finished`（API 流程在 `tests/e2e/scenarios/api/recording-smoke.sh` 中验证）。

## 开始响应

```json
{
  "status": "recording",
  "format": "gif",
  "fps": 5,
  "quality": 80,
  "tabId": "tab1"
}
```

## 状态响应

```json
{
  "active": true,
  "state": "recording",
  "format": "gif",
  "durationSeconds": 12.5,
  "frames": 62,
  "tabId": "tab1",
  "fps": 5
}
```

`state` 取值为 `idle`、`recording`、`limit_reached`、`stopping`、`encoding`、`finished`、`aborted` 之一。录制存在期间，主体还可能携带 `stopReason` 和 `outputPath`；一旦 `finished`，它携带 `outputPath`，编码失败或被截断时携带 `error` 或 `warning`。

## 请求体字段（POST /record/start）

- `format`：`gif`（默认）、`webm` 或 `mp4`。在命令行界面中由文件扩展名决定。其他取值返回 `400 invalid_format`。
- `fps`：每秒帧数，上限 30（默认 5）。
- `quality`：JPEG 抓取质量，上限 100（默认 80）。
- `scale`：分辨率倍率，上限 1.0（默认 1.0）。小于 1 的值减小输出尺寸。
- `tabId`：指定目标标签页。

录制进行中再次 `start` 返回 `409 recording_error`；无活动录制时 `stop` 返回 `400 recording_error`。

## 请求体字段（POST /record/stop）

- `discard`：`true` 则丢弃帧而不编码。

## 命令行界面

- `record start <file>`：开始录制。格式由扩展名决定（.gif、.webm、.mp4）。
- `record stop`：停止，等待服务器完成编码，然后将文件移动到开始时给定的路径（未指定时为 `recording-<timestamp>.gif`）。
- `record status`：显示活动录制信息。
- `--fps <n>`：每秒帧数（默认 5）。
- `--quality <n>`：JPEG 抓取质量（默认 80）。
- `--scale <f>`：分辨率倍率（默认 1.0）。
- `--tab <id>`：指定目标标签页。

## 格式依赖

| 格式 | 依赖 | 编码 | 说明 |
| --- | --- | --- | --- |
| `.gif` | 无 | 纯 Go（Floyd-Steinberg 抖动） | 始终可用；长录制对 CPU 压力大 |
| `.webm` | `ffmpeg` | 通过 ffmpeg 管道编码 VP8 | 需要 `$PATH` 中有 ffmpeg |
| `.mp4` | `ffmpeg` | 通过 ffmpeg 管道编码 H.264 | 需要 `$PATH` 中有 ffmpeg |

GIF 编码完全在进程内运行——无需外部二进制。抖动受 CPU 限制。一个 GIF 最多编码 600 帧（5 fps 下约 2 分钟），内存中的帧数据最多 256 MB；帧按请求的 scale 保留（不强制降采样），因此更大的抓取在截断前能容纳的帧更少。抓取本身在 5 分钟或 9000 帧时停止（`state: limit_reached`）。

WebM 和 MP4 在 `record stop` 后将帧流式传给 ffmpeg。若未安装 ffmpeg，`record start` 返回 `400 ffmpeg_required`。通过你的包管理器安装 ffmpeg（`brew install ffmpeg`、`apt install ffmpeg` 等），或使用 `.gif` 作为零依赖替代。

## 注意

- 受 `security.allowScreencast` 控制（默认禁用）。用 `pinchtab config set security.allowScreencast true` 启用并重启服务器。
- 每个 bridge 实例同一时间只有一个活动录制。
- 录制归调用方的会话（或代理 id）所有；匿名录制可被任何调用方停止。
- MCP：`pinchtab_record` 接受 `action`（`start`|`stop`|`status`，必填）、`file`、`fps`、`quality`、`scale`、`tabId`。

## 相关页面

- [Screenshot](./screenshot.md)
- [PDF](./pdf.md)
