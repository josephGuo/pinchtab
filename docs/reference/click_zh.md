# 点击（Click）

使用快照 ref、CSS 选择器、XPath 选择器、文本选择器或语义选择器点击元素。

```bash
curl -X POST http://localhost:9867/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"click","ref":"e5"}'
# CLI Alternative
pinchtab click e5
# Response (use --json for full JSON)
OK
```

## 当点击触发导航时

使页面发生移动的点击**会成功**。无需提前声明任何内容，也无需选择开启：结果会报告标签页落到了哪里、以及你的 refs 已失效。

```bash
pinchtab nav https://example.com --snap
# e1:link "Learn more"
pinchtab click e1
# OK navigated https://www.iana.org/help/example-domains
# HINT: every ref from your last snapshot is dead — run `pinchtab snap -i` before the next action
echo $?   # 0
```

带 `--json` 时，同一个点击会在 `result` 中携带结果：

```json
{
  "success": true,
  "result": {
    "clicked": true,
    "navigated": true,
    "url": "https://www.iana.org/help/example-domains",
    "previousUrl": "https://example.com/",
    "refsStale": true
  }
}
```

`navigated` 的判断基于实际发生的事——动作前的 URL 与动作后的 URL 比较——而非元素的 role 或标签，因此对链接、路由 `<button>` 以及会跳转的表单控件，答案都一样。仅片段变化（`#section`）不算导航：文档相同，你的 refs 仍然有效。

**`refsStale: true` 意味着你上一次 `/snapshot` 的每个 ref 都已失效。** ref 是按快照生成的，因此在下一次以 ref 为目标的操作前要重新拍一张：

```bash
pinchtab click e1 --snap    # click, then print the new snapshot in one call
```

无论你用哪种形式——普通点击、带 `--wait-nav` 的点击、还是会跳转的 `--submit` 点击——这些字段都会出现，都会报告落点和 refs 已失效。`--submit` 点击以其自身的 `postState` 为主，同时附带落点信息。

`--wait-nav` 不是"允许导航"——它让点击在返回前**等待**导航完成，这正是当下一操作依赖新页面已加载时你想要的。

## 命令行界面 Flags

| Flag | 说明 |
|------|-------------|
| `--css` | 用 CSS 选择器代替 ref |
| `--wait-nav` | 点击后等待导航 |
| `--snap` | 点击后输出交互式快照 |
| `--snap-diff` | 点击后输出快照差异 |
| `--text` | 点击后输出页面文本 |
| `--dialog-action` | 自动处理 JS 对话框：`accept` 或 `dismiss` |
| `--dialog-text` | prompt 回复文本（配合 `--dialog-action accept`） |
| `--dismiss-banners` | 在 `--wait-nav` 点击后关闭 cookie/同意横幅（无 `--wait-nav` 时无效） |
| `--dismiss-known-interstitials` | 在解析点击目标前关闭已识别的门户 interstitial（无法关闭时以 `known_interstitial_not_dismissed` 拒绝） |
| `--x`, `--y` | 在指定坐标点击 |
| `--humanize` | 启用人性化贝塞尔曲线 + 抖动输入路径（覆盖实例配置） |
| `--submit` | 走一次性 submit-click 路径，并在响应中带上有界的提交后状态 |
| `--mode dom\|dispatch` | 点击投递的底层应急通道。省略 `--mode` 走正常点击路径，`dom` 走 `element.click()`，或 `dispatch` 对目标派发合成点击事件 |
| `--json` | 完整 JSON 响应 |
| `--tab` | 目标特定标签页 |

## 示例

```bash
pinchtab click e5                       # Click by ref
pinchtab click "#login"                 # Click by CSS
pinchtab click "text:Submit"            # Click by text
pinchtab click e5 --snap                # Click and show new snapshot
pinchtab click e5 --wait-nav            # Click and wait for navigation
pinchtab click e5 --dialog-action accept  # Auto-accept alert/confirm
pinchtab click "#sign-in" --submit        # Submit once and report the observed outcome
pinchtab click e5 --mode dom             # Activate target directly despite occlusion
pinchtab click e5 --mode dispatch        # Dispatch click events on target despite occlusion
pinchtab click --x 100 --y 200           # Click at coordinates
```

## 注意事项

- 元素 ref 来自 `/snapshot`，一次导航会使它们全部失效——见 [当点击触发导航时](#当点击触发导航时)
- iframe 后代的 ref 可直接点击，无需切换框架
- 选择器查找限于当前框架范围（默认：`main`）
- 在基于选择器的 iframe 操作前使用 [`/frame`](./frame.md)
- 缺失的选择器立即失败；对动态 UI 先用 `pinchtab wait`（见 [`commands.md`](../commands.md)）
- API 也接受 `selector` 字段：`{"kind":"click","selector":"#login"}`
- 点击行为如下：省略 `mode` 走正常点击路径，`mode:"dom"` 走 `element.click()`，或 `mode:"dispatch"` 走合成点击事件。
- 将 `mode` 视为点击投递的宽泛底层逃生舱。绕过遮挡是常见情形，但它也能帮助需要非默认点击路径的页面。
- `mode` 与 humanize 互斥——无论 humanize 来自请求上的 `humanize:true` 还是 `instanceDefaults.humanize:true`。
- 要让某个需要它的页面走较慢的人性化路径，在动作 JSON 中传 `humanize:true`，或设置 `instanceDefaults.humanize:true`。
- `submit:true` 用于终态表单操作，重试可能导致重复提交。对点击而言它只发送恰好一次 DOM 点击，禁用恢复/重试投递，并报告有界的 `postState` 结果（URL 变化或打开的 modal 关闭时为 `succeeded`；否则为 `pending`）。它需要元素目标，不能与坐标、`waitNav`、`mode` 或 `humanize:true` 组合。它只在单个 `/action` 请求上接受，不用于批次或 macro。
- Ref 词表：以 ref 为目标的动作可在主体中以 `vocab` 回传快照的 `X-PinchTab-Vocab` token（或通过 `X-PinchTab-Vocab` 请求头），并用 `vocabTab` 指明其标签页。若更新的快照已对该标签页的 ref 重新编号，`/action` 会以 `vocab_superseded` 冲突拒绝，而非作用在错误节点上——重新快照并使用新 ref。命令行界面会自动附上你上一次快照的 token。当某个动作使标签页的 ref 重新纪元化时，响应携带新的 `X-PinchTab-Vocab` 头。
- 标签页上打开着 JavaScript 对话框时，`/action` 以 `dialog_blocked` 冲突拒绝（details 携带 `dialogType` 和 `dialogMessage`）；用 `pinchtab dialog accept|dismiss` 回答它，或在打开它的点击上传入 `--dialog-action`。见 [Dialog](./dialog.md)。
- 无法解析的 ref 应答 `404 ref_not_found`，带 `details.dispatched: false`。

## 相关页面

- [Frame](./frame.md)
- [Snapshot](./snapshot.md)
- [Navigate](./navigate.md)
