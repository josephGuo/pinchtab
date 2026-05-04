# 评估

在当前标签页中运行 JavaScript。除非在配置中明确启用评估，否则此端点会被禁用。

启用 `security.allowEvaluate` 是一个有文档记录的、非默认的、降低安全性的配置更改。它允许在页面上下文中执行任意 JavaScript，仅应在经过明确审查的身份验证和网络暴露的可信系统上使用。

```bash
curl -X POST http://localhost:9867/evaluate \
  -H "Content-Type: application/json" \
  -d '{"expression":"document.title"}'
# 命令行界面 替代方案
pinchtab eval "document.title"
# 响应（默认为结果值；使用 --json 获取完整 JSON）
Example Domain
```

## 命令行界面 标志

| 标志 | 描述 |
|------|-------------|
| `--await-promise` | 响应前解析返回的 Promise |
| `--json` | 完整 JSON 响应 |
| `--tab` | 目标特定标签页 |

## 示例

```bash
pinchtab eval "document.title"
pinchtab eval "document.querySelectorAll('a').length"
pinchtab eval "fetch('/api/data').then(r => r.json())" --await-promise
pinchtab eval "document.title" --json    # {"result":"Example Domain"}
```

## 注意事项

- 需要 `security.allowEvaluate: true`
- 标签页范围变体：`POST /tabs/{id}/evaluate`
- `/evaluate` 故意**不**限于框架范围
- 当前 `/frame` 状态不影响 `pinchtab eval`
- 对于 iframe 访问，你的表达式必须显式处理

## 相关页面

- [配置](./config.md)
- [框架](./frame.md)
- [标签页](./tabs.md)