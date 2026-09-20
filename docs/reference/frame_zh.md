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

# CLI Alternative
pinchtab frame                          # Shows: main (or frameId if scoped)
pinchtab frame "#payment-frame"         # Shows: <frameId> (<name>)
pinchtab frame main                     # Shows: main
pinchtab frame --json                   # Full JSON response
```

`POST /frame` 和 `pinchtab frame` 接受的目标：

- `main` 清除框架范围
- iframe 所有者的快照 ref
- iframe 元素的选择器
- 框架名称或框架 URL

响应：`{tabId, scoped, target, current}`——未限定范围时，`target` 和 `current` 均为 `"main"`；限定范围时，`target` 为框架 ID，`current`/`frame` 携带 `{frameId, frameUrl, frameName, ownerRef}`。不带 `target` 的 `POST` 返回 `400`；非 iframe/框架的目标返回 `400`。

MCP：`pinchtab_frame` 接受 `target`（省略则读取当前范围）、`tabId`、`browser`。

典型 iframe 流程（API 形式见 `tests/e2e/scenarios/api/actions-extended.sh`）：

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
- 框架范围适用于 `/snapshot`、`/capture`、基于选择器的 `/action` 调用，以及未显式提供 `frameId` 时的 `/text`
- 从框架范围提供的读取携带一个 `frame` 对象（`frameId`、`frameUrl`、`frameName`、`ownerRef`、`frameTitle`），使后续读取者能判断该内容不是顶层文档
- `/evaluate` 是独立的，不继承框架范围

## 相关页面

- [快照](./snapshot.md)
- [点击](./click.md)
- [填充](./fill.md)