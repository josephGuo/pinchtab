# 导航

打开新标签页并导航到 URL，或者在提供标签页 ID 时重用标签页。

```bash
curl -X POST http://localhost:9867/navigate \
  -H "Content-Type: application/json" \
  -d '{"url":"https://pinchtab.com"}'
# CLI 替代方案
pinchtab nav https://pinchtab.com
# 响应（默认为标签页 ID；使用 --json 获取完整 JSON）
8f9c7d4e1234567890abcdef12345678
```

## CLI 标志

| 标志 | 描述 |
|------|-------------|
| `--tab` | 重用现有标签页 |
| `--new-tab` | 强制新标签页 |
| `--block-images` | 阻止图像加载 |
| `--block-ads` | 阻止广告 |
| `--snap` | 导航后输出快照 |
| `--snap-diff` | 导航后输出快照差异 |
| `--print-tab-id` | 仅打印标签页 ID（管道时自动） |
| `--json` | 完整 JSON 响应 |

## 示例

```bash
pinchtab nav https://example.com              # 导航，打印标签页 ID
pinchtab nav https://example.com --snap       # 导航并快照
TAB=$(pinchtab nav https://example.com)       # 捕获标签页 ID 以供重用
pinchtab nav https://other.com --tab "$TAB"   # 重用标签页
pinchtab nav https://example.com --block-images  # 跳过图像
```

## API 体字段

| 字段 | 描述 |
|-------|-------------|
| `url` | 目标 URL（必需） |
| `tabId` | 重用现有标签页 |
| `newTab` | 强制新标签页 |
| `blockImages` | 阻止图像加载 |
| `blockAds` | 阻止广告 |
| `timeout` | 导航超时 |
| `waitFor` | 等待条件 |
| `waitSelector` | 等待选择器 |

## 相关页面

- [快照](./snapshot.md)
- [标签页](./tabs.md)