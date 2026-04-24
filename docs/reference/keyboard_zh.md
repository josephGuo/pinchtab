# 键盘

用于输入文本和分发按键事件的低级键盘输入命令。

## keyboard type

通过为每个字符分发单独的按键事件（keyDown/keyUp）来输入文本。这会触发一些应用程序依赖的键盘事件监听器。

```bash
pinchtab keyboard type "hello world"
```

**性能说明：** 对于超过 20 个字符的字符串，PinchTab 使用混合方法来避免 CDP 超时：前 5 个和后 5 个字符使用真实的按键事件输入，而中间部分使用 `Input.insertText`。这在边界提供真实的按键模拟，同时保持长字符串的性能可接受。

## keyboard inserttext

直接插入文本而不分发按键事件。相当于粘贴文本——更快但不会触发 keydown/keypress/keyup 监听器。

```bash
pinchtab keyboard inserttext "test@pinchtab.com"
```

在以下情况使用 `inserttext`：
- 你需要最大速度
- 目标不依赖于按键事件监听器
- 你正在以编程方式填写表单

在以下情况使用 `type`：
- 应用程序在按键时验证输入
- 你需要触发自动完成或实时搜索
- 你正在模拟真实的用户输入

## keydown / keyup

按住或释放单个键。对于修饰键（Shift、Ctrl、Alt）或测试按键保持行为很有用。

```bash
pinchtab keydown Shift
pinchtab keyboard type "abc"   # 输入 "ABC"（按住 Shift）
pinchtab keyup Shift
```

## API 等价物

```bash
curl -X POST http://localhost:9867/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"keyboard-type","text":"hello world"}'

curl -X POST http://localhost:9867/action \
  -H "Content-Type: application/json" \
  -d '{"kind":"keyboard-inserttext","text":"hello world"}'
```

## 相关页面

- [输入](./type.md) — 通过选择器/引用在特定元素中输入
- [填充](./fill.md) — 直接设置输入值
- [按键](./press.md) — 按单个键