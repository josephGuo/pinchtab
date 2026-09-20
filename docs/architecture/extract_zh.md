# Extract 架构

`internal/extract` 针对一份抓取到的可访问性（accessibility）快照，填充一个 JSON schema——由扁平字段加对象数组组成。它位于 `find`（单个元素）与 `scrape`（整个站点）之间：代理交出一个 schema，拿回带类型的数据，并附带逐字段的置信度，以便在某个字段不确定时回退到 `find`。

它不依赖模型（model-free），也不用浏览器（browserless）。它复用了支撑 `find` 的同一套语义匹配器（`github.com/pinchtab/semantic`，词法 + 特征哈希嵌入，无网络、无模型下载），作用于交给它的节点列表上。它从不触碰 CDP。

## 处理流水线

```text
JSON schema ──ParseSchema──► Schema (ordered properties, resolved hints)
node list   ──canonical order (by ref)──► descriptors (semdesc.Build)
                                    │
        per property: build query ──► matcher.Find ──► best ref + score
                                    │         (score desc, then document order)
                read value (Value|Text|Name, Checked for bool) ──► coerce to type
                                    │
                    data + per-field {ref, score, confidence, source, reason}

array property: choose container (scope | group detection) ──► items
                                    │
        per item: the same field resolution over the item's subtree only
                                    │
     data[prop] = [ {…}, … ]  fields[prop] = {ref: container, items: [{ref, fields}], truncated}
```

## Schema 子集

接受的 schema 是一个带 `properties` 的 `object`。每个属性是 `string`、`number`、`integer`、`boolean` 或 `array` 之一。对每个属性，提取器会识别：

- `description`——与属性名一起作为查询文本参与匹配。
- `required`——一个无法填充的必填字段会被列入 `missing`，绝不凭空捏造。
- `x-pinchtab-hint`——一个显式查询或选择器，覆盖基于名字的查询（见下）。

一个 `array` 属性必须携带类型为 `object` 的 `items`，并拥有自己的扁平 `properties`；对每个数组，提取器识别 `minItems`、`maxItems` 和 `x-pinchtab-scope`（一个命名容器的选择器，语法与 hint 相同）。

对于超出该子集的构造，`ParseSchema` 返回一个带类型的 `*UnsupportedError`，指明出问题的路径：

- 非 `object` 的根（`type: array is not supported`），
- 类型为 `object` 的属性（`properties.price.type: object is not supported`），
- 未知类型（`properties.price.type: decimal is not supported`），
- 某个属性上嵌套的 `properties`（`properties.price.properties: ...`），
- 需要浏览器的 `css:` 或 `xpath:` hint 或 scope（`properties.price.x-pinchtab-hint: ...`），
- 不带对象 `items` 的数组（`properties.tags.items: ...`）、嵌套在 `items` 内部的数组（`properties.rows.items.properties.tags.type: ...`），或负数的 `minItems`/`maxItems`。

属性顺序按字母排列，因此 `missing` 与迭代是确定的。

## 字段解析

对每个属性，提取器构建一个查询，并取阈值（默认 `0.3`，与 `find` 相同）以上的唯一最佳节点：

1. **查询。** 一个已解析的 `x-pinchtab-hint` 优先。否则查询为属性名与其描述拼接而成。
2. **匹配。** 查询通过共享的组合匹配器对每个描述符运行。胜者在此处选出，而非取自匹配器自身的 `best_ref`：分数四舍五入到六位小数，最高者胜出，平局按文档顺序打破，因为组合匹配器是并发合并其两半的，其平局排序不稳定。
3. **读取。** 匹配节点的值按优先级 `Value`、`Text`、`Name` 读取；布尔值先参考可访问性 `Checked` 状态。
4. **强制转换（coerce）。** 原始字符串被转换为 schema 类型。转换失败会让该字段带原因留空——绝不给出类型错误的值。

### Hint

`x-pinchtab-hint` 接受裸查询或选择器，被限制为节点列表无需浏览器即可回答的种类：

- 裸字符串 → 原样用作查询（裸 `eN` 是一个 ref，裸 `//…` XPath 会被拒绝，两者例外），
- `find:` → 自然语言查询，
- `role:`、`label:`、`placeholder:`、`alt:`、`title:`、`testid:`，以及包裹上述之一的 `first:` / `last:` / `nth:` → 经由 `selector.SemanticQuery` 路由，使语法与 action 和 `find` 路径保持单一来源，
- `text:` → 文本作为查询，
- `ref:`（或裸 `eN`）→ 原样选中该节点；不在节点列表中的 ref 会让该字段带原因 `ref_not_found` 留空，
- `css:` / `xpath:` → 被 `ParseSchema` 拒绝（它们需要实时 DOM）。

一个能解析出结果的 hint 胜过基于名字的查询，因此代理可以钉住一个歧义字段（例如多个价格中的促销价），而无需重命名 schema。

### 请求 scope

`POST /extract` 还接受一个可选的顶层 `scope`，用与 hint 相同的语法解析（`Schema.WithScope`）。设置后，它在整个节点列表上解析一次，然后整个 schema 只在匹配节点的子树内解析。一个什么都不匹配的 scope 会把每个字段都报告为原因 `scope_not_found`，并把必填字段列入 `missing`。

## 对象数组

一个 `array` 属性为快照中每个重复组解析出一个对象。

### 树推导

`A11yNode` 不带子节点链接；树由先序（pre-order）节点列表和 `Depth` 推导：

- 一个节点的**子树**是其后跟随节点中一段连续区间，这些节点的 `Depth` 大于该节点，
- 它的**直接子节点**是该区间中恰为 `Depth + 1` 的节点，
- 它的**祖先**通过反向行走找到，即每个 `Depth` 小于上一个被取节点的较早节点。

下面的每一步（候选、条目、逐条目 scope、表头查找）都建立在这三种推导之上，别无其他。

### 容器

1. **Scope。** 如果该属性携带 `x-pinchtab-scope`，该选择器在整个节点列表上完全像 hint 一样解析（`ref:` 原样，语义种类经由匹配器），检测只在匹配节点的子树内运行，节点本身也作为候选被接纳，即使只有一个条目，因此对 `table` 的 scope 会得到其行，对 `list` 的 scope 会得到其项。一个什么都不匹配的 scope 会让该属性带原因 `scope_not_found` 留空；不在它之外运行任何检测。
2. **检测。** 否则，每个直接子节点共享同一主导 role（它们中最频繁的 role）的节点都是候选，且该 role 至少出现**三**次；当节点自身的 role 是 `list`、`table`、`rowgroup`、`grid` 或 `feed` 时，至少出现**两**次即可。
3. **评分。** 每个候选通过在其前三个条目（不足则按实际数量）内解析条目 schema 来打分，取被填充字段的比例：`score = filled / (sampled items × fields)`。候选按分数、再按条目数、再按文档顺序排序。若最佳分数为 `0`——任何地方都没有条目字段能解析——则报告 `no_repeated_group`，一个完全没有候选的页面同样如此。带 scope 的容器跳过 `0` 检查：那是代理自己选的。

这个排序正是把一个含多个重复组的页面消歧的关键：一个链接导航菜单，面对一个要求标题和价格的商品条目 schema，得分为 `0`，而商品区域得分为 `1`。

### 条目

条目是携带主导 role 的容器直接子节点（`listitem`、`article`、`row`……）。其子节点包含 `columnheader` 的 `row` 是表头，绝不是条目。

存在两种逐条目模式：

- **子树模式。** 扁平解析器只在条目自身的子树上运行（含条目节点本身，因为剪枝后的树常常把菜单或结果的文本留在条目节点上），因此条目 3 中的 `price` 永远不可能匹配条目 1，某个条目中缺失的字段在那里保持缺失，即使其他每个条目都有它。
- **列式模式。** 当条目是 `row`，且最近的外层 `table` 或 `grid`（或容器自身）持有一行表头时，每个字段先针对表头单元格解析一次，得到一个列索引，然后每一行读取该索引处的单元格。表格单元格是按其值而非按其列命名的，因此逐行匹配它们将一无所获。短于该索引的行会让该字段带 `no_match` 留空。

所有字段都缺失的条目会被丢弃。输出上限为 schema 的 `maxItems` 与 `Options.MaxItems`（默认 `100`）中的较小者；该上限在丢弃空条目之后才测试，因此只有当确实有一个非空条目被丢弃时才报告 `truncated: true`。

条目数少于 `minItems` 会让该属性带原因 `too_few_items` 留空。

### 结果结构

`data[prop]` 是对象数组。`fields[prop]` 携带容器 `ref`、组 `score` 及其置信区间、`items`——每个返回条目一个 `{ref, fields}`，带与扁平字段相同的逐字段诊断——以及 `truncated`。

## 强制转换规则

- **string**——去除首尾空白后的原始值。
- **number / integer**——只解析**第一个数字 token**：剥离前导货币符号和正负号，去掉该 token 的千位分隔符，保留符号和小数点，并在 token 之后的第一个字符处停止解析，因此尾部单词中的数字绝不会被黏附进来。ASCII `-` 与 Unicode 减号 `−`（U+2212）均被识别。`integer` 向零截断。示例：`"$1,299.00"` → `1299`，`"−3.5 kg"` → `-3.5`，`"4.7 out of 5"` → `4.7`（而非 `4.75`），`"2 of 3"` → `2`，`"call for price"` → 带原因 `not_numeric` 留空。
- **boolean**——先看可访问性 `Checked` 状态（`true`/`false`），再看读取值中的 `yes`/`no`/`true`/`false` 字样。`mixed` 复选框或不相关字符串会让该字段带原因 `not_boolean` 留空。

## 置信度

每个字段携带 `ref`、`score` 和 `confidence`。置信区间来自 `semantic.CalibrateConfidence`——`high`（score ≥ 0.8）、`medium`（≥ 0.6）、`low`（低于）——与 `find` 报告的区间相同，因此两者不会漂移。代理可对 `low` 字段回退到 `find`。

## 确定性

`Resolve` 是确定性的。匹配之前，节点列表被复制并按 ref 稳定排序为文档顺序（快照 ref `e1`、`e2`……按先序分配，因此数字 ref 顺序就是文档顺序）。同一 schema 作用于同一批节点——无论输入顺序如何——都产生完全相同的输出；平局按文档顺序打破。数组解析继承这一点：候选和条目按文档顺序枚举，平局按文档顺序排序。

## 描述符

节点通过 `internal/semdesc` 转换为匹配器描述符，这是与 `find` 处理器共享的节点到描述符映射的唯一来源。提取器没有浏览器，因此它跳过 `find` 应用于实时快照的 DOM 元数据增强，而是按交给它的节点列表进行匹配。
