# 站点审计与对比（Site Audit & Compare）

`pinchtab audit` 会在浏览器层面对一个或多个页面运行审计，并生成一份带版本号的报告：
截图、控制台日志、网络请求与失效资源、可交互元素、带发现项的无障碍评分、Core Web
Vitals 计时，以及基于规则的安全发现。`pinchtab compare` 会在两个站点版本（线上与预发布）
上审计同一批页面，并逐页报告视觉与数据差异——可作为 CI 门禁使用。

两个命令背后的处理管线：

```
seaportal (HTTP discovery & extraction)  →  pinchtab (browser enrichment)  →  report
   sitemap flattening, page profiles,        screenshots, console, network,     json / md /
   browserRecommended routing                a11y, timing, security rules       html / pdf
```

SeaPortal 负责廉价的 HTTP 工作；浏览器——那个昂贵的资源——只在能产生信号的地方运行。
报告是确定性的：相同输入产生相同报告（e2e 测试套件用一份黄金文件对此加以强制校验）。

## 快速开始（Quick start）

```bash
pinchtab audit https://example.com --output-dir ./audit
# → ./audit/report.json + ./audit/screenshots/*.png
```

从站点地图审计整站，并按页面模板限定采样：

```bash
pinchtab audit https://example.com/sitemap.xml --sitemap --sample-size 2 --output-dir ./audit
```

## `pinchtab audit`

```
pinchtab audit [url] [flags]
```

参数是一个页面 URL；若加上 `--sitemap`，则是一个站点地图 URL。使用
`--seaportal-report` 时，页面改由 SeaPortal 结果文件提供，此时无需 URL 参数。

| Flag | 默认值 | 含义 |
|---|---|---|
| `--sitemap` | false | 将该 URL 视为 sitemap.xml，并审计从中发现的页面 |
| `--seaportal-report <file>` | 空 | 从一个 SeaPortal 结果 JSON 文件（`Result` 对象数组）审计页面 |
| `--enrich-all` | false | 用浏览器增强每一个 seaportal 页面，忽略 `browserRecommended` 路由 |
| `--sample-size <n>` | 0 | 每个模板分组（如 `/products/p1..pN`）审计的页面数（0 = 全部页面；采用确定性挑选） |
| `--screenshot` | true | 捕获页面截图 |
| `--network-monitor` | true | 收集网络请求与失效资源 |
| `--concurrency <n>` | 2 | 并行审计的页面数（最多 8） |
| `--output-dir <dir>` | 空 | 将 `report.json` 和 `screenshots/` 写入此目录 |
| `--format <f>` | json | 报告格式：`json`、`md`、`html` 或 `pdf` |
| `--json` | false | 将完整报告 JSON 打印到标准输出 |
| `--cookie name=value` | 空 | 在运行前把一个 cookie 注入隔离的临时浏览器实例（可重复） |
| `--cookies-file <file>` | 空 | 从一个由 `{name, value, domain, ...}` 对象组成的 JSON 数组，把 cookies 注入隔离的临时浏览器实例 |
| `--profile <name>` | 空 | 针对该浏览器配置文件所属实例运行（不能与 `--cookie` 或 `--cookies-file` 同时使用） |

失败契约：一个加载失败的页面**不会**导致本次运行失败——命令以退出码 0 结束，该页面的
报告条目会带一个 `error` 字段。只有在无内容可审计时（例如空的站点地图），运行本身才会
报错。

### 输入与路由（Inputs and routing）

- **URL** —— 审计那一个页面。
- **`--sitemap`** —— 页面通过 seaportal 的站点地图扁平化（支持递归的站点地图索引）发现，
  随后去重、按 URL 模板分组并采样（`--sample-size`）。入口 URL 始终最先审计；模板页面在
  未分组页面之后审计。
- **`--seaportal-report`** —— 该文件是一个由 seaportal `Result` 对象组成的 JSON 数组
  （即过渡性的 `seaportal-results/v0` 格式，版本记录在报告的 `input.seaportalFormat`）。
  只有 seaportal 标记为 `profile.browserRecommended` 的页面才会被浏览器增强；其余页面在
  报告中保留其 HTTP 提取摘要。`--enrich-all` 可覆盖此行为。

### 报告结构（Report anatomy）

`report.json` 是一份带版本号的 `AuditReport`（`schemaVersion`）：

```
schemaVersion, generatedAt, input, options
summaryScore                     # mean accessibility score of enriched pages;
                                 # broken assets, failed requests and uncaught
                                 # JS errors do not move it
pages[]:
  url, title, statusCode?, error?  # error set when the page failed to load
  seaportal?                     # HTTP-extraction summary when ingested
  securityFindings[]?            # ruleId, severity, detail, url
  browser:
    screenshotPath               # relative path under the output dir
    consoleLogs[], jsErrors[], networkRequests[], brokenAssets[]
    interactiveElements[], accessibilityScore
    timingMetrics: ttfbMs, fcpMs, lcpMs, cls, domContentLoadedMs, loadMs
securityFindings[]               # page findings aggregated site-level
recommendations[]
```

`jsErrors[]` 是未捕获的 JavaScript 异常（message、stack、line、column）——走的是
`GET /errors` 通道，与 `consoleLogs[]` 分开，因此一次 `console.error` 调用与一个抛出的异常
绝不会被合并。抛出过异常的页面绝不会被报告为 `ok`。

`--format md` / `--format html` 会在 `report.json` 旁边写出 `report.md` / `report.html`
（若没有 `--output-dir` 则打印到标准输出）。HTML 是自包含的——内联 CSS、相对截图链接。
`--format pdf` 通过浏览器打印 HTML 报告；它需要 `--output-dir` 和 `evaluate` 能力，并且在
打印失败时仍会写出 `report.json`、弹出一条警告，且退出码非零。

安全发现基于规则、由收集到的数据离线计算：混合内容（mixed content）、不安全的表单
action、通过 http 提交的密码表单、暴露的敏感路径（`.env`、`.git` 等），以及目录列表页面。

## `pinchtab compare`

```
pinchtab compare <live-url> <staging-url> [flags]
```

在两个基础 URL 上审计同一批页面，按路径配对，对截图配对做像素级 diff，并对数据做 diff
（未捕获的 JS 错误、控制台错误数、失效资源、无障碍评分、带噪声阈值的加载时间）。只存在于
一侧的页面会被报告为 `added`（新增）/ `removed`（移除）。

未捕获的 JS 错误（`jsErrors`）和失效资源（`brokenAssets`）按标识（identity）而非按数量
比较，并且比较时会把页面自身的 `scheme://host` 掩码掉，以便两个基础 URL 上的同一失败能够
匹配：匹配依据是异常的首条消息行，以及失效资源的 `url` 加上其 HTTP 状态。仅在一侧出现的
失败——或在数量相同但换成了另一个——即为漂移（drift）；两侧都出现的同一失败则不算漂移，
因为这个门禁比较的是两次部署，不对页面的绝对健康度做判断。其余字段（包括
`consoleErrors`）按数量比较，因此用一条控制台消息替换另一条仍然是不可见的。

没有 HTTP 状态的失效资源——例如连接重置这类传输错误，或在收集窗口内始终未完成的请求——
不参与比较。`--fail-on-diff` 是对部署的门禁，而一次运行恰好丢失了哪个请求是运行本身的
属性：把这些计入会让两个相同站点大约每五次运行就失败一次。它们仍会被审计本身以及
`/network` 完整报告出来（在那里，一次加载失败是真实发现）；排除只适用于本次对比。

| Flag | 默认值 | 含义 |
|---|---|---|
| `--pages <p1,p2>` | 基础 URL | 要对比的逗号分隔相对路径 |
| `--visual-diff` | true | 捕获截图并计算视觉 diff |
| `--concurrency <n>` | 2 | 每一侧并行审计的页面数（最多 8） |
| `--output-dir <dir>` | 空 | 将 `report.json` 和 `diffs/` 写入此目录 |
| `--format <f>` | json | 报告格式：`json`、`md` 或 `html` |
| `--json` | false | 将对比报告 JSON 打印到标准输出 |
| `--fail-on-diff` | false | 当存在任何视觉或数据差异时以非零退出码结束 |
| `--cookie`、`--cookies-file`、`--profile` | 空 | 与 `audit` 相同的鉴权 flags |

差异是数据，而非失败：若不加 `--fail-on-diff`，即使页面有差异，命令也以退出码 0 结束。
发生变化的配对会在 `diffs/` 下得到一张带标注的 diff 图片，由该页面的 `diffImagePath` 引用。

## 鉴权（Authentication）

```bash
pinchtab audit https://example.com/account --cookie session=abc123
pinchtab audit https://example.com/account --cookies-file cookies.json
pinchtab audit https://example.com --profile work
```

cookies 在首次导航之前注入一个隔离的临时浏览器实例。该实例及其 cookie jar 在运行结束后
（包括审计请求失败时）都会被丢弃，因此不会改动任何持久化配置文件。`--cookie` 与
`--cookies-file` 不能与 `--profile` 同时使用。在没有注入 cookie 时，`--profile` 会把本次
运行路由到拥有该浏览器配置文件的实例（用 `pinchtab instance start --profile <name>`
启动一个这样的实例）。

## 库模式（Go）

`pkg/pinchtabaudit` 是供在 Go 程序中嵌入增强能力使用的公开客户端——可参见
`docs/examples/enrich` 中可运行的示例：

```go
client := pinchtabaudit.New("http://localhost:9867", token)
page, err := client.EnrichPage(ctx, "https://example.com", nil)
report, err := client.EnrichWithBrowser(ctx,
    pinchtabaudit.AuditInput{SitemapURL: "https://example.com/sitemap.xml"}, nil)
```

该模块处于 1.0 之前阶段。`pkg/pinchtabaudit` 中导出的 Go 类型可能在任何版本中不经通知
就改变名称、结构或 JSON tag，也不提供弃用过渡期。如果这对你重要，请固定模块版本。下面的
HTTP API 才是稳定契约：如果你需要一个稳定契约，请基于它构建。

## HTTP API

命令行界面是这些端点之上的一层薄客户端（`pinchtab audit` 调用的是 `POST /audit`）：

- `POST /audit/page {"url", "options"}` → 单页 `PageAudit`（`url`、`title`、
  `error`、`screenshot`、`a11yFindings`、`securityFindings`，外加内联的
  `BrowserPageData` 字段）
- `POST /audit {"urls" | "sitemapUrl" | "seaportalResults", "options",
  "seaportalFile", "concurrency", "sampleSize", "enrichAll"}` → `AuditReport`

## Docker / CI

审计命令行界面是运行中的 pinchtab 服务器之上的一层薄客户端，因此在 Docker 中要先运行
服务器容器，再对其 exec 执行审计（服务器搭建见
[guides/docker.md](guides/docker.md)）：

```bash
docker run -d --name pinchtab -v pinchtab-data:/data --shm-size=2g pinchtab/pinchtab
docker exec pinchtab pinchtab audit https://example.com --output-dir /data/audit
docker cp pinchtab:/data/audit ./audit
```

用「预发布与生产是否一致」作为部署门禁——`--fail-on-diff` 在站点一致时退出 0，存在任何
视觉或数据差异时退出非零（两种退出码都被 compose 栈中的 e2e 测试套件覆盖）：

```bash
pinchtab compare https://example.com https://staging.example.com \
  --pages /,pricing,docs --fail-on-diff --output-dir ./compare-artifacts
# exit 0 → ship; exit 1 → inspect ./compare-artifacts/diffs/
```

## 尚未交付（Not yet shipped）

- `--llm-refine`（对报告做 LLM 后处理）——待定。
- 外部扫描器集成（例如 Nuclei）以做更深层安全检查——待定；当前的安全发现即上述内置规则。
