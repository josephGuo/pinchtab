# Memory（内存）

PinchTab 无需人工打开 DevTools 即可回答"这个页面是否在泄漏？"：读取标签页的堆内存用量、将 V8 堆快照写入服务端文件、对单个快照做摘要、以及对比两个快照看哪些构造函数增长了。

`GET /memory` 始终可用。所有触及堆快照的操作都需要 `security.allowMemory`（默认关闭；关闭时返回代码 `memory_disabled`），因为堆快照包含页面上的每一个字符串，包括 token。

## 端点

| 方法 | 路径 | 能力 | 用途 |
| --- | --- | --- | --- |
| `GET` | `/memory`、`/tabs/{id}/memory` | 无 | 堆用量与 DOM 计数器 |
| `POST` | `/memory/snapshot`、`/tabs/{id}/memory/snapshot` | `allowMemory` | 将堆快照写入文件 |
| `GET` | `/memory/snapshot/{snapshotId}/summary` | `allowMemory` | 单个快照中排名靠前的构造函数与重复字符串 |
| `GET` | `/memory/compare` | `allowMemory` | 两个快照之间的构造函数增长 |

## 用法

`GET /memory?tabId=<id>&gc=true` 返回 `tabId`、`usedJSHeapSize`、`totalJSHeapSize`、`jsHeapSizeLimit`、`documents`、`nodes`、`listeners`、`frames` 和 `gc`。`gc=true` 先执行一次垃圾回收，使两次读取只比较存活内存；`gc` 值非布尔时返回 `400 bad_gc`。

```bash
pinchtab memory --gc            # --tab <id>, --json
```

MCP `pinchtab_memory` 接受 `tabId`、`gc` 和 `browser`。

## 快照

`POST /memory/snapshot` 带可选的 `{"tabId": "..."}`，将快照流式写入 `<stateDir>/heapsnapshots/<id>.heapsnapshot`，并返回 `{id, path, bytes, nodeCount, durationMs, tabId}`。该文件可在 Chrome DevTools 的 Memory 面板中加载。超过 `security.memorySnapshotMaxBytes`（默认 512 MB，上限 4 GB）的快照会被丢弃并应答 `413 memory_snapshot_too_large`（details 携带 `maxBytes`）。

```bash
pinchtab memory snapshot                          # prints the id and path
pinchtab memory snapshot --out app.heapsnapshot   # also copy it locally
```

MCP `pinchtab_memory_snapshot` 接受 `tabId`、`top` 和 `browser`；它在一次调用中完成快照并返回 `{id, path, bytes, summary}`。

## 摘要

`GET /memory/snapshot/{snapshotId}/summary?top=20` 返回 `id`、`path`、`top`、`nodeCount`、`edgeCount`、`totalSelfSize`、`constructors`（不同构造函数的数量）、`topBySize`、`topByCount`（构造函数行，含 `name`、`count`、`selfSize`）以及 `duplicateStrings`（`value`、`length`、`count`、`selfSize`）。

```bash
pinchtab memory summary heap_20260913_101010 --top 5
```

对象按构造函数名分组；其他 V8 节点类型归为 `(array)`、`(string)`、`(closure)`、`(compiled code)`、`(system)` 等，与 DevTools 中一致。JavaScript `Array` 的元素存放在单独的 `(array)` 后备存储中，因此增长的数组其字节数显示在 `(array)` 下，而 `Array` 本身保持很小。

## 对比

`GET /memory/compare?base=<id>&head=<id>&top=20&retained=false`

| 参数 | 默认值 | 说明 |
| --- | --- | --- |
| `base` | 必填 | 较早的快照 id |
| `head` | 必填 | 较晚的快照 id |
| `top` | `20` | `constructors` 和 `newDuplicateStrings` 中的行数；超过 200 的值被截断为 200，非正数或非整数返回 `400 bad_top` |
| `retained` | `false` | 为每个返回行添加 `retainedSize`，来自 `head` 的支配树 |

```json
{
  "top": 20,
  "retained": false,
  "base": {"id": "heap_20260913_101010", "nodeCount": 51234, "edgeCount": 210876, "totalSelfSize": 2811904},
  "head": {"id": "heap_20260913_101042", "nodeCount": 51310, "edgeCount": 211002, "totalSelfSize": 18543616},
  "nodeDelta": 76,
  "sizeDelta": 15731712,
  "changed": 41,
  "constructors": [
    {"name": "(array)", "baseCount": 812, "headCount": 815, "countDelta": 3,
     "baseSelfSize": 190440, "headSelfSize": 15919128, "sizeDelta": 15728688}
  ],
  "newDuplicateStrings": [
    {"value": "session-expired", "length": 15, "count": 3, "selfSize": 96}
  ]
}
```

- `constructors` 只列出计数或自身大小发生变化的构造函数，按绝对 `sizeDelta` 从大到小排列，因此增长和释放都会浮到顶部。`changed` 统计所有发生变化的构造函数，包括被 `top` 截断的那些。
- `newDuplicateStrings` 是在 `head` 中被持有超过一次、但在 `base` 中并未重复的字符串。
- `retainedSize` 是：若该构造函数的每个对象都被回收可释放的内存——即每个对象支配树子树的总和，嵌套在同一构造函数另一对象下的对象只计一次。它仅在 `retained=true` 时出现。与 DevTools 一样，弱边被忽略。计算它会将整个 head 图驻留在内存中，耗时与边数成正比，因此它是可选开启的。
- 已解析的快照按快照文件在进程生命周期内缓存在内存中，因此再次对比相同 id、或对已对比过的快照做摘要，都不会重新解析。

快照 id 匹配 `^[A-Za-z0-9][A-Za-z0-9_-]{0,127}$`；服务器命名为 `heap_<YYYYMMDD_HHMMSS>`。错误：`400 bad_snapshot_id`（id 缺失或格式错误）、`400 bad_top`、`400 bad_retained`、`404 memory_snapshot_not_found`（details 指出 `id`）、`422 memory_snapshot_invalid`（details 指出 `id` 和损坏的 `section`）、`403 memory_disabled`。

命令行界面与 MCP：

```bash
pinchtab memory compare <base> <head> --top 5
pinchtab memory compare <base> <head> --retained --json
```

MCP `pinchtab_memory_compare` 接受 `base` 和 `head`（均必填）、`top`、`retained` 和 `browser`。

覆盖场景见 `tests/e2e/scenarios/api/memory-basic.sh`、`tests/e2e/scenarios/api/memory-extended.sh` 和 `tests/e2e/scenarios/cli/memory-basic.sh`。

## 实操：定位一次泄漏

1. 加载页面并让其稳定，然后拍基准快照：`pinchtab memory snapshot` 打印出 `heap_A`。
2. 重复可疑操作若干次（打开并关闭对话框、翻页、来回路由跳转）。重复操作会让泄漏线性增长，而一次性缓存保持平稳。
3. 拍第二个快照：`heap_B`。拍快照会执行一次完整垃圾回收，因此仍被计入的对象都是可达的。
4. 对比：`pinchtab memory compare heap_A heap_B --top 10`。泄漏的构造函数会排在顶部附近，具有正的 `sizeDelta`，且 `countDelta` 与重复次数对应。
5. 询问谁持有它：`pinchtab memory compare heap_A heap_B --retained` 显示每个列出的构造函数维持存活的内存量。一个 retained size 很大的小对象（某个组件、闭包、`Map`）通常就是要修复的持有者。
6. 撤销该操作（关闭、释放、导航离开），拍 `heap_C`，对比 `heap_B heap_C`：该构造函数应显示负的 `sizeDelta`。若没有，则仍有某处引用它。
7. 在 DevTools 中打开这些文件（Memory 面板，Comparison 视图）查看单个对象的 retainer 路径。
