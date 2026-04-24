# 命令参考

## 服务器和运行时

```bash
pinchtab server                         # 启动完整服务器（仪表板 + API）
pinchtab bridge                         # 启动仅桥接运行时
pinchtab mcp                            # 启动 MCP 标准输入/输出服务器
pinchtab daemon                         # 显示守护进程状态
pinchtab daemon install                 # 作为后台服务安装
pinchtab daemon start                   # 启动后台服务
pinchtab daemon stop                    # 停止后台服务
pinchtab daemon restart                 # 重启后台服务
pinchtab daemon uninstall               # 移除后台服务
pinchtab completion <shell>             # 生成 shell 自动完成
```

## 导航

`pinchtab nav <url>` 使用 `/navigate`。当您不传递 `--tab` 时，PinchTab 会打开一个新标签页并导航到它。

```bash
pinchtab nav <url>                      # 打开新标签页并导航
pinchtab nav <url> --tab <id>           # 重用特定标签页
pinchtab nav <url> --new-tab            # 明确强制新标签页
pinchtab nav <url> --block-images       # 为此导航阻止图片
pinchtab nav <url> --block-ads          # 为此导航阻止广告
pinchtab nav <url> --snap               # 导航并输出交互式快照
pinchtab quick <url>                    # 导航并拍摄快照
pinchtab back                           # 在活动标签页中返回
pinchtab back --tab <id>                # 在特定标签页中返回
pinchtab forward                        # 在活动标签页中前进
pinchtab reload                         # 重新加载活动标签页
```

隐藏别名：`goto`、`navigate`、`open`

## 标签页

`tab` 命令仅列出、聚焦、创建和关闭标签页。它不代理其余的浏览器命令集。

```bash
pinchtab tab                            # 列出标签页
pinchtab tab <id>                       # 通过 ID 或基于 1 的索引聚焦标签页
pinchtab tab new                        # 打开空白标签页
pinchtab tab new <url>                  # 打开标签页并导航
pinchtab tab close <id>                 # 关闭标签页
```

使用带有 `--tab` 的顶级命令进行标签页范围的工作：

```bash
pinchtab snap --tab <id>
pinchtab click --tab <id> <selector>
pinchtab pdf --tab <id> -o page.pdf
```

## 交互

大多数元素命令接受统一的选择器：

- 快照引用，如 `e5`
- CSS 选择器，如 `#login`
- XPath，如 `xpath://button`
- 文本选择器，如 `text:Submit`
- 语义选择器，如 `find:login button`

选择器查找按框架显式进行。非作用域选择器仅搜索当前框架作用域，默认为 `main`。在基于选择器的 iframe 工作之前使用 `pinchtab frame ...`。支持同源 iframe 作用域；当前不暴露跨域 iframe 后代。

```bash
pinchtab frame                         # 显示当前框架作用域
pinchtab frame "#payment-frame"        # 将选择器作用域限定到 iframe
pinchtab frame main                    # 将选择器作用域返回到顶层文档
pinchtab click [selector]               # 点击元素或使用 --x/--y 点击坐标
pinchtab click --css <selector>         # 强制 CSS 选择器模式
pinchtab click --wait-nav <selector>    # 点击并等待导航
pinchtab click --snap <selector>        # 点击并输出交互式快照
pinchtab dblclick [selector]            # 双击
pinchtab type <selector> <text>         # 通过键盘事件输入
pinchtab fill <selector> <text>         # 直接填充
pinchtab press <key>                    # 按下键
pinchtab hover [selector]               # 悬停元素
pinchtab mouse move <x> <y>             # 将鼠标移动到坐标
pinchtab mouse move [selector]          # 或移动到元素中心
pinchtab mouse down [selector]          # 按下鼠标按钮
pinchtab mouse up [selector]            # 释放鼠标按钮
pinchtab mouse wheel [dy|selector]      # 分发滚轮增量
pinchtab drag <from> <to>               # 在目标之间拖动（选择器/引用或 x,y）
pinchtab focus [selector]               # 聚焦元素
pinchtab scroll <selector|pixels>       # 滚动元素或页面
pinchtab scroll down --snap             # 滚动并输出快照
pinchtab scroll 800 --snap-diff         # 滚动并输出快照差异
pinchtab select <selector> <value>      # 选择 <select> 选项
pinchtab check <selector>               # 勾选复选框或单选按钮
pinchtab uncheck <selector>             # 取消勾选复选框或单选按钮
pinchtab scrollintoview <selector>      # 滚动元素到视图中
```

低级鼠标命令对于拖动句柄、类似画布的 UI 以及 DOM 原生点击或悬停抽象不足的流程很有用：

```bash
pinchtab mouse move e5
pinchtab mouse down --button left
pinchtab mouse up --button left
pinchtab mouse wheel 240 --dx 40
pinchtab mouse move --x 400 --y 320
pinchtab drag e5 400,320
```

## 页面分析

```bash
pinchtab snap                           # 可访问性快照
pinchtab snap -i -c                     # 交互式 + 紧凑
pinchtab snap -d                        # 与之前快照的差异
pinchtab snap --selector <css>          # 作用域快照
pinchtab snap --max-tokens <n>          # 限制令牌预算
pinchtab snap --depth <n>               # 限制树深度
pinchtab snap --text                    # 文本输出
pinchtab text                           # 提取可读文本
pinchtab text --full                    # 完整页面 innerText
pinchtab text --raw                     # 原始提取
pinchtab text --frame <frameId>         # 从一个 iframe 读取文本
pinchtab find <query>                   # 语义元素搜索
pinchtab find --threshold <0-1>         # 最小相似度分数
pinchtab find --explain                 # 包含分数分解
pinchtab find --ref-only                # 仅打印最佳引用
pinchtab eval <expression>              # 评估 JavaScript
```

`pinchtab eval` 故意不进行框架作用域限定。当前的 `pinchtab frame` 状态影响基于选择器的命令，如 `snap`、`click`、`fill` 和 `type`，当未明确提供 `--frame` 时，它也会影响 `text`。

基于选择器的操作现在在选择器不匹配时快速失败。如果 UI 仍在加载，请先使用 `pinchtab wait`，而不是依赖操作超时。

## 键盘、等待和诊断

```bash
pinchtab keyboard type <text>           # 在聚焦元素处输入
pinchtab keyboard inserttext <text>     # 插入文本而无键盘事件
pinchtab keydown <key>                  # 按住键
pinchtab keyup <key>                    # 释放键
pinchtab wait <selector|ms>             # 等待选择器或固定持续时间
pinchtab wait --text <text>             # 等待页面文本
pinchtab wait --url <glob>              # 等待 URL 匹配
pinchtab wait --load networkidle        # 等待加载状态
pinchtab wait --fn <expression>         # 等待 JS 变为真值
pinchtab network                        # 列出捕获的网络请求
pinchtab network <requestId>            # 详细显示一个请求
pinchtab network --stream               # 流式传输网络条目
pinchtab network --clear                # 清除捕获的网络数据
pinchtab network-export                 # 导出为 HAR 1.2（保存到 exports/）
pinchtab network-export -o session.har  # 导出到特定文件
pinchtab network-export --format ndjson # 导出为 NDJSON（每行一个条目）
pinchtab network-export --body          # 包含响应体
pinchtab network-export --stream -o l.har # 浏览时实时捕获到文件
pinchtab dialog accept [text]           # 接受警告/确认/提示
pinchtab dialog dismiss                 # 取消对话框
pinchtab console                        # 显示控制台日志
pinchtab console --clear                # 清除控制台日志
pinchtab errors                         # 显示浏览器错误日志
pinchtab errors --clear                 # 清除浏览器错误日志
pinchtab clipboard read                 # 读取服务器端剪贴板文本
pinchtab clipboard write <text>         # 写入剪贴板文本
pinchtab clipboard copy <text>          # write 的别名
pinchtab clipboard paste                # read 的别名
pinchtab cache clear                    # 清除浏览器 HTTP 磁盘缓存
pinchtab cache status                   # 检查是否可以清除缓存
```

手动切换和恢复可通过 CLI 和 API 获得：

```bash
pinchtab tab handoff <tabId> --reason captcha --timeout-ms 120000
pinchtab tab handoff-status <tabId>
pinchtab tab resume <tabId> --status completed
```

API 等效项：

暂停的切换状态会阻止动作执行路由（`/action`、`/actions`、`/macro`），返回 `409 tab_paused_handoff`，直到通过超时恢复或过期。

```bash
curl -X POST "$PINCHTAB_SERVER/tabs/<tabId>/handoff"
curl "$PINCHTAB_SERVER/tabs/<tabId>/handoff"
curl -X POST "$PINCHTAB_SERVER/tabs/<tabId>/resume"
```

## 捕获和导出

```bash
pinchtab screenshot                     # 保存屏幕截图到生成的 .jpg 路径
pinchtab screenshot -o <path>           # 保存屏幕截图到选定路径
pinchtab screenshot -q <0-100>          # JPEG 质量
pinchtab screenshot -s <selector>       # 通过选择器捕获特定元素
pinchtab screenshot -s <selector> --css-1x # 以 CSS 像素大小导出选择器屏幕截图
pinchtab pdf                            # 将活动页面导出为 PDF
pinchtab pdf -o <path>                  # 保存 PDF 到选定路径
pinchtab pdf --landscape                # 横向方向
pinchtab pdf --scale <n>                # 打印比例
pinchtab pdf --paper-width <in>         # 纸张宽度（英寸）
pinchtab pdf --paper-height <in>        # 纸张高度（英寸）
pinchtab pdf --page-ranges <r>          # 页面范围，如 1-3
pinchtab pdf --prefer-css-page-size     # 使用 CSS 页面大小
pinchtab pdf --display-header-footer    # 显示页眉/页脚
pinchtab download <url>                 # 通过浏览器会话下载
pinchtab download <url> -o <path>       # 将下载的文件保存到路径
pinchtab upload <file>                  # 上传到默认文件输入
pinchtab upload <file> -s <css>         # 上传到特定文件输入
```

## 实例、配置文件和活动

```bash
pinchtab instances                      # 列出运行中的实例
pinchtab instance start                 # 启动实例
pinchtab instance start --profile <id-or-name>
pinchtab instance start --mode headed
pinchtab instance start --port <n>
pinchtab instance start --extension /path/to/ext
pinchtab instance stop <id>             # 停止实例
pinchtab instance logs <id>             # 显示实例日志
pinchtab instance navigate <id> <url>   # 在实例中打开标签页并导航
pinchtab profiles                       # 列出配置文件
pinchtab activity                       # 列出记录的活动事件
pinchtab activity tab <tab-id>          # 按标签页过滤活动
pinchtab health                         # 检查服务器健康状态
```

## 配置和安全

```bash
pinchtab config                         # 交互式配置概览/编辑器
pinchtab config init                    # 创建默认配置文件
pinchtab config show                    # 打印有效的运行时配置
pinchtab config token                   # 将 server.token 复制到剪贴板而不打印
pinchtab config path                    # 打印配置文件路径
pinchtab config validate                # 验证当前配置文件
pinchtab config get <path>              # 读取一个文件配置值
pinchtab config set <path> <val>        # 设置一个文件配置值
pinchtab config patch <json>            # 将 JSON 合并到配置文件
pinchtab security                       # 交互式安全概览
pinchtab security up                    # 应用更严格的默认值
pinchtab security down                  # 应用文档化的无保护预设
```

## 全局标志

根命令支持：

```bash
pinchtab --server http://host:9867 <command>
pinchtab --help
pinchtab --version
```

带有 `--tab` 的命令当前包括：

- `nav`
- `back`
- `forward`
- `reload`
- `snap`
- `screenshot`
- `pdf`
- `find`
- `text`
- `click`
- `dblclick`
- `hover`
- `mouse move`
- `mouse down`
- `mouse up`
- `mouse wheel`
- `focus`
- `type`
- `press`
- `fill`
- `scroll`
- `select`
- `eval`
- `check`
- `uncheck`
- `keyboard type`
- `keyboard inserttext`
- `keydown`
- `keyup`
- `scrollintoview`
- `network`
- `network-export`
- `wait`
- `dialog accept`
- `dialog dismiss`
- `console`
- `errors`

## 输出格式

大多数命令默认输出人类可读的文本。使用 `--json` 获取机器可解析的 JSON 输出：

```bash
pinchtab tab                            # 人类可读：*abc123  https://...  页面标题
pinchtab tab --json                     # JSON：{"tabs":[...]}  
pinchtab frame                          # 人类可读：main
pinchtab frame --json                   # JSON：{"tabId":"...","scoped":false,...}
pinchtab network                        # 人类可读：GET  200  https://...
pinchtab network --json                 # JSON：{"entries":[...],"count":5}
```

**对于脚本和自动化**：当通过管道传递输出或以编程方式解析时，始终使用 `--json`。人类可读格式可能在版本之间变化，不保证稳定。JSON 模式是稳定的约定。

带有 `--json` 的命令包括：`tab`、`frame`、`network`、`click`、`type`、`scroll`、`nav`、`back`、`forward`、`reload`、`wait`、`find`、`eval` 和大多数动作命令。