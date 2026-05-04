# 框架

获取或设置基于选择器的快照和操作的当前框架范围。

默认情况下，选择器查找保持在主文档中。要使用 CSS、XPath 或文本选择器定位 iframe 内容，请先设置框架。

来自 `/snapshot` 的引用不同：如果快照包含同源 iframe 后代，这些引用仍然可以直接使用，无需设置框架范围。

```bash
curl http://localhost:9867/frame

curl -X POST http://localhost:9867/frame \
  -H "Content-Type: application/json" \
  -d '{"target":"#payment-frame"}'

curl -X POST http://localhost:9867/frame \
  -H "Content-Type: application/json" \
  -d '{"target":"main"}'

# 命令行界面 替代方案
pinchtab frame                          # 显示：main（如果有范围则显示 frameId）
pinchtab frame "#payment-frame"         # 显示：<frameId> (<name>)
pinchtab frame main                     # 显示：main
pinchtab frame --json                   # 完整 JSON 响应
```

`POST /frame` 和 `pinchtab frame` 接受的目标：

- `main` 清除框架范围
- iframe 所有者的快照引用
- iframe 元素的选择器
- 框架名称或框架 URL

典型的 iframe 流程：

```bash
pinchtab snap -i
pinchtab frame "#payment-frame"
pinchtab snap -i
pinchtab fill "#card-number" "4111111111111111"
pinchtab click "#pay-button"
pinchtab frame main
```

注意：

- 选择器范围是显式的；未限定范围的选择器不会自动穿透到 iframes 中
- 支持同源 iframe 内容；目前不将跨域 iframe 后代暴露为框架范围
- 嵌套 iframes 通常需要多次 `frame` 跳转
- 相同的框架范围适用于基于选择器的 `/snapshot` 和 `/action` 调用，以及当未明确提供 `frameId` 时的 `/text`
- `/evaluate` 是独立的，不继承框架范围

## 相关页面

- [快照](./snapshot.md)
- [点击](./click.md)
- [填充](./fill.md)