# 填充

直接设置输入值，不依赖与 `type` 相同的事件序列。

```bash
curl -X POST http://localhost:9867/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"fill","ref":"e8","text":"ada@pinchtab.com"}'
# CLI Alternative
pinchtab fill e8 "ada@pinchtab.com"
# Response (use --json for full JSON)
OK
```

## 命令行界面 标志

| 标志 | 描述 |
|------|-------------|
| `--submit` | 填充字段后按回车（动作 JSON 中为 `submit:true`） |
| `--snap` | 填充后输出交互式快照 |
| `--snap-diff` | 填充后输出快照差异 |
| `--text` | 填充后输出页面文本 |
| `--json` | 完整 JSON 响应 |
| `--tab` | 目标特定标签页 |

## 示例

```bash
pinchtab fill e8 "ada@pinchtab.com"     # Fill by ref
pinchtab fill "#email" "user@example.com"  # Fill by CSS
pinchtab fill "text:Email" "test@test.com" # Fill by text selector
pinchtab fill e8 "value" --snap         # Fill and show snapshot
```

## 注意事项

- 接受统一选择器：`e8`, `#email`, `xpath://...`, `text:Email`
- iframe 后代的引用可以直接填充，无需切换框架
- 选择器查找仅限于当前框架范围（默认：`main`）
- 在基于选择器的 iframe 填充前使用 [`/frame`](./frame.md)
- 缺失的选择器会立即失败；对于异步字段，先使用 `pinchtab wait`（见 [`commands.md`](../commands.md)）
- 对于 API，使用 `selector` 字段进行 CSS/XPath/文本选择器
- API 读取 `text`，回退到 `value`。两者都没有的填充以 `400 missing_fill_text` 拒绝；发送 `"text":""` 可清空字段

## 相关页面

- [框架](./frame.md)
- [输入](./type.md)
- [快照](./snapshot.md)