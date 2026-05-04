# 快照

获取当前页面的可访问性快照，包括可被操作命令重用的元素引用。

在快照捕获期间会自动检测 iframe 内容。同源 iframe 后代包含在 iframe 所有者元素下方，它们的引用可以直接与操作命令重用。跨域 iframes 目前仅作为所有者节点存在。

选择器作用域是显式的。`selector=...` 仅搜索当前框架范围，默认为 `main`。要将基于选择器的快照作用域到 iframe 中，请先使用 [`/frame`](./frame.md) 或 `pinchtab frame` 设置框架。

```bash
curl "http://localhost:9867/snapshot?filter=interactive"
# 命令行界面 替代方案（默认为紧凑文本输出）
pinchtab snap -i
# 输出
[e5] link "More information..."

# 使用 --full 或 --compact=false 获取 JSON
pinchtab snap --full
```

## 命令行界面 标志

| 标志 | 描述 |
|------|-------------|
| `-i`, `--interactive` | 过滤到交互式元素 + 标题（默认：true） |
| `-c`, `--compact` | 紧凑文本输出（默认：true） |
| `-d`, `--diff` | 显示与前一个快照的差异 |
| `--full` | 完整 JSON 输出（`--interactive=false --compact=false` 的简写） |
| `--text` | 文本输出格式 |
| `-s`, `--selector` | 用于限定快照的 CSS 选择器 |
| `--max-tokens` | 最大令牌预算 |
| `--depth` | 树深度限制 |
| `--tab` | 目标特定标签页 |

## 示例

```bash
pinchtab snap                           # 交互式紧凑（默认）
pinchtab snap -i -c                     # 同上
pinchtab snap --full                    # 包含所有节点的完整 JSON
pinchtab snap -d                        # 显示自上次快照以来的变化
pinchtab snap --selector "#main"        # 限定到元素
pinchtab snap --max-tokens 2000         # 限制输出大小
```

## API 参数

| 参数 | 描述 |
|-----------|-------------|
| `filter` | `interactive` 表示交互式 + 标题 |
| `format` | `compact`、`text`、`yaml` 或默认 JSON |
| `diff` | `true` 表示差异模式 |
| `selector` | 用于限定的 CSS 选择器 |
| `maxTokens` | 令牌预算限制 |
| `depth` | 树深度限制 |

## 相关页面

- [点击](./click.md)
- [框架](./frame.md)
- [标签页](./tabs.md)