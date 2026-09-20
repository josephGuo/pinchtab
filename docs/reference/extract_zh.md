# Extract（按 schema 抽取）

`/extract` 根据一个 JSON schema 从当前页面读取带类型的值：价格以 JSON 数字返回，复选框以布尔值返回，商品网格以对象数组返回。它基于该标签页的无障碍快照工作，因此每个字段还携带其读取来源的 `ref`，可直接用于 `/action`。

## 端点

PinchTab 提供两种形式：

- `POST /extract`
- `POST /tabs/{id}/extract`

在请求体中带 `tabId` 使用 `POST /extract`（或不带，针对当前活动标签页）。当你已知标签页 ID 并希望编排器将请求路由到拥有它的实例时，使用 `POST /tabs/{id}/extract`。

## 请求体

| 字段 | 类型 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- | --- |
| `schema` | object 或 string | 是 | - | 描述数据的 JSON schema；见 [Schema 子集](#schema-子集) |
| `tabId` | string | 否 | 当前活动标签页 | 使用 `POST /extract` 时指定标签页 ID |
| `scope` | string | 否 | 整页 | 将每个字段限定到某个元素的子树：裸 ref（`e12`）、`role:`、`text:` 或普通查询。CSS 和 XPath 被拒绝（`400`） |
| `threshold` | float | 否 | `0.3` | 每个字段的最低匹配分数 |
| `maxItems` | int | 否 | `100` | 每个数组的条目上限 |

## 主示例

将 schema 保存为 `product.schema.json`：

```json
{
  "type": "object",
  "required": ["name", "price", "inStock"],
  "properties": {
    "name": {"type": "string", "description": "product name", "x-pinchtab-hint": "role:heading"},
    "price": {"type": "number", "description": "product price"},
    "rating": {"type": "number", "description": "product rating out of 5"},
    "inStock": {"type": "boolean", "description": "in stock availability", "x-pinchtab-hint": "role:checkbox"}
  }
}
```

```bash
curl -X POST http://localhost:9867/extract \
  -H "Content-Type: application/json" \
  -d "{\"schema\":$(cat product.schema.json)}"
# CLI Alternative
pinchtab extract --schema product.schema.json
```

在商品页面上，命令行界面仅打印数据：

```json
{
  "inStock": true,
  "name": "Sony WH-1000XM5 Wireless Headphones",
  "price": 1299,
  "rating": 4.7
}
```

## 命令行界面

```bash
pinchtab extract --schema product.schema.json
cat product.schema.json | pinchtab extract --schema -
pinchtab extract --schema product.schema.json --fields
pinchtab extract --schema products.schema.json --max-items 2
pinchtab extract --schema amounts.schema.json --scope role:table
```

| Flag | 说明 |
| --- | --- |
| `--schema <file\|->` | Schema 文件，或 `-` 从 stdin 读取（必填） |
| `--scope <selector>` | 作为 `scope` 发送 |
| `--max-items <n>` | 作为 `maxItems` 发送 |
| `--fields` | 在数据之后，每个字段打印一行 `field<TAB>ref<TAB>confidence`；数组条目显示为 `products[0]` 和 `products[0].name` |
| `--explain` | 同一张表，附带 `score`、`source` 和 `reason` 列 |
| `--json` | 打印完整响应封装而非 `data` |
| `--tab <id>` | 通过 `POST /tabs/{id}/extract` 定位标签页 |

缺失的必填字段和被截断的数组会报告到 stderr；命令仍以 0 退出，因为页面已应答。服务器拒绝的 schema 以非零退出码退出并附带服务器的 `400` 消息，消息中会指出出错的路径。

命令行界面会保留 extract 生成的词表，因此一个字段的 ref 可以紧接着被点击，无需中间再做一次快照：

```bash
pinchtab extract --schema product.schema.json --fields
pinchtab click e7
```

## 响应字段

| 字段 | 说明 |
| --- | --- |
| `data` | 按 schema 类型强制转换后的值；数组中每个重复组对应一个对象 |
| `fields` | 每个属性：`ref`、`score`、`confidence`（`high`、`medium`、`low`）、`source`（`value`、`text`、`name`、`checked`、`hint`），以及未解析时的 `reason`；数组额外提供 `items`（每个含自己的 `ref` 和 `fields`）和 `truncated` |
| `missing` | 未解析的必填属性 |
| `truncated` | 任一数组达到条目上限时为 `true` |
| `vocabularyToken` | 返回的 ref 所属的 ref 词表（也在 `X-PinchTab-Vocab` 中） |
| `latency_ms` | 解析耗时，毫秒 |
| `element_count` | 被考虑的快照节点数 |
| `idpiWarning` | IDPI 处于 warn 模式时的提示性警告 |

## Schema 子集

- 根是一个带 `properties` 的对象；`required` 列出在 `missing` 中报告的字段。
- 属性为 `string`、`number`、`integer` 或 `boolean`，或其 `items` 为上述属性对象的 `array`。数组条目内部的嵌套对象和数组被拒绝。
- 匹配器同时搜索属性名及其 `description`，因此描述性的 `description`（"product price"）比 key 本身更重要。
- `maxItems` 和 `minItems` 约束数组；低于 `minItems` 的数组无法解析。
- 不支持的构造返回 `400` 并指出路径，例如 `properties.price.type: object is not supported`。

## 提示与范围

- 属性上的 `x-pinchtab-hint` 固定其读取来源：ref（`e12`），或 `role:`、`text:` 等语义选择器。当某字段返回 `low` 时添加一个。
- 数组上的 `x-pinchtab-scope` 固定其容器，例如 `role:table`，适用于页面上有多个重复组的情况。
- 请求级的 `scope` 将整个 schema 限定到一个子树；匹配不到任何内容的 scope 会使每个字段都无法解析，reason 为 `scope_not_found`。

## 何时使用它

- `extract` 用于你本需要从快照中解析出来的带类型的值：价格、标志位、表格行、搜索结果。
- `find` 用于找到一个要操作的元素。
- `text --markdown` 用于要阅读的正文。

## 错误情况

| 状态码 | 条件 |
| --- | --- |
| `400` | 无效 JSON、缺少 `schema`，或不支持的 schema 或 `scope`（消息指出路径） |
| `403` | 被严格模式下的 IDPI 拦截 |
| `404` | 标签页未找到 |
| `409` | `dialog_blocked`：JavaScript 对话框正阻塞该标签页（用 `pinchtab dialog` 回答它） |
| `500` | Chrome 未初始化、快照不可用，或快照无元素 |
