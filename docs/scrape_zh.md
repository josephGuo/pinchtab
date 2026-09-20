# 站点抓取（Site Scrape）

`pinchtab scrape` 把整站变成一棵由 markdown 内容组成的页面树。页面先通过纯 HTTP 被发现并
提取（经由 SeaPortal：站点地图或链接爬取、URL 模式采样）。只有当 HTTP 提取结果单薄、被
拦截或失败的页面，才会放到真实浏览器中重新渲染，并从渲染后的 DOM 重新提取——这样一来，
纯 JavaScript 渲染的内容仍能进入报告，而静态页面则永远不必为浏览器付出成本。

处理管线：

```
seaportal (HTTP crawl & extraction)  →  routing  →  pinchtab (browser)  →  report
   sitemap/link crawl, URL-pattern       thin /        render + re-extract   json / md
   sampling, per-page markdown           blocked?       JS-only pages         page tree
```

SeaPortal 负责廉价的 HTTP 工作；浏览器——那个昂贵的资源——只在 HTTP 不够用时运行。每个
页面都会记录其内容来源（`http` 或 `browser`）以及把它路由到该来源的判定结果。

## 快速开始（Quick start）

```bash
pinchtab scrape https://example.com --output-dir ./scrape
# → ./scrape/report.json
```

同时写一份人类可读的摘要并打印汇总：

```bash
pinchtab scrape https://example.com --format md --output-dir ./scrape
# → ./scrape/report.json + ./scrape/report.md
```

## 预览 → 展开（大型站点）（Preview → Expand）

以完整保真度抓取一个大型站点，成本会双倍叠加：传输每个页面的 markdown 主宰了 token 成本，
而对每个被路由页面做浏览器渲染则主宰了挂钟时间（wall-clock）成本。为了先概览、再下钻，
把抓取拆成「先廉价、后深入」的两轮。

**1. 预览（Preview）**——做 HTTP 爬取和逐页路由判定，但不做浏览器渲染，也不抓完整页面
正文。每个页面的 markdown 被扣留，替换为一个 `charCount`（完整展开会有多重）和一段开头的
`snippet`：

```bash
pinchtab scrape https://example.com --preview
#   https://example.com/            · 4229 chars
#       Browser control for AI agents. 12MB Go binary…
#   https://example.com/docs/app    · 87 chars · needs browser: thin-content
#       Loading…
```

**2. 展开（Expand）**——以完整保真度抓取你选定的那些 URL（按正常路由做 HTTP 提取 + 浏览器
渲染），而不是重新爬取：

```bash
pinchtab scrape https://example.com \
  --only https://example.com/docs/app \
  --only https://example.com/pricing \
  --output-dir ./scrape
```

展开是无状态的：把预览里显示的那些确切 URL 交回来即可。不在服务端保留任何会话或缓存。

## `pinchtab scrape`

```
pinchtab scrape <url> [flags]
```

参数是站点 URL。发现工作从该主机的根开始播种：SeaPortal 会在主机根目录查找
`robots.txt` 和 `sitemap.xml`，若找不到则从根页面开始链接爬取。

| Flag | 默认值 | 含义 |
|---|---|---|
| `--preview` | false | 仅出大纲：页面树、每页的 `charCount` 与 `snippet`——不做浏览器渲染或完整正文 |
| `--only <url>` | 空 | 以完整保真度精确展开这些 URL，而非爬取（可重复） |
| `--max-pages <n>` | 50 | 整站采样的最大页面数 |
| `--max-per-pattern <n>` | 8 | 每个 URL 模式分组（如 `/blog/*`）采样的最大页面数 |
| `--include <regex>` | 空 | 只爬取匹配此正则的 URL（可重复） |
| `--exclude <regex>` | 空 | 跳过匹配此正则的 URL（可重复） |
| `--enrich-all` | false | 对每个可达页面做浏览器渲染，忽略路由 |
| `--no-browser` | false | 仅 HTTP 爬取；记录路由判定但不做浏览器渲染 |
| `--concurrency <n>` | 2 | 并行做浏览器渲染的页面数（最多 8） |
| `--timeout <s>` | 60 | 整个 HTTP 爬取超时时间（秒） |
| `--output-dir <dir>` | 空 | 将 `report.json`（加 `--format md` 时还有 `report.md`）写入此目录 |
| `--format <f>` | json | 报告格式：`json` 或 `md` |
| `--json` | false | 将完整报告 JSON 打印到标准输出 |
| `--cookie name=value` | 空 | 在运行前把一个 cookie 注入隔离的临时浏览器实例（可重复） |
| `--cookies-file <file>` | 空 | 从一个由 `{name, value, domain, ...}` 对象组成的 JSON 数组，把 cookies 注入隔离的临时浏览器实例 |
| `--profile <name>` | 空 | 针对该浏览器配置文件所属实例运行（不能与 `--cookie` 或 `--cookies-file` 同时使用） |

失败契约：一个在**两个**引擎中都失败的页面不会导致本次运行失败——命令以退出码 0 结束，
该页面的报告条目会带一个 `error` 字段。浏览器增强失败时会保留 HTTP 提取结果并记录
`browserError`。只有在爬取彻底失败或未发现任何页面时，运行本身才会报错。

### 路由（Routing）

每个 HTTP 提取出来的页面都会被打分，判断浏览器是否应重新渲染它：

- **not-found**（404 / 410）永远不路由——重新渲染无法挽回内容。
- **fetch-error** 或 **被拦截状态**（401、403、407、429、503）会路由：一个带隐身与挑战
  处理能力的真实浏览器，可能在普通 HTTP 客户端被拒的地方成功。
- **thin-content**（提取结果短于「静态即可用」阈值）会路由：很可能是个 JavaScript 外壳。

`--enrich-all` 会把除 not-found（404 / 410）之外的每个页面都强制交给浏览器；`--no-browser`
在每个页面上记录判定结果但不做任何渲染。

## 报告结构（Report anatomy）

`report.json` 是一份带版本号的抓取报告（`schemaVersion`）：

```
schemaVersion, generatedAt, input
site:      baseUrl, title, sitemapFound, totalURLsInSitemap, sampledPages
pageGroups[]:  pattern, total, sampled, urls[]        # site tree by URL pattern
pages[]:
  url, title, statusCode, contentType
  markdown                     # withheld in preview mode
  charCount, snippet           # set in preview mode instead of markdown
  meta, schema, internalLinks, externalLinks
  source                       # "http" or "browser"
  browserRecommended, browserReasons   # the routing verdict
  browserError?                # browser enrichment failed; HTTP content kept
  error?                       # page failed in both engines
summary:   contentTypes, httpPages, browserPages, failedPages, recommendations
```

`--format md` 会在 `report.json` 旁边写出 `report.md`（若没有 `--output-dir` 则打印到标准
输出）：一份单一摘要，包含站点概览、按 URL 模式组织的页面树，以及每个页面的内容（预览
模式下则是其 snippet）。

### 汇总计数（Summary counters）

`httpPages`、`browserPages` 与 `failedPages` 三者把页面做了划分：每个页面恰好落入其中之一，
三者之和等于页面总数。

- `failedPages` 统计未返回可用内容的页面——即一个在两个引擎中都失败的页面（一个传输层
  `error`）**或**其 `statusCode` 为 `>= 400` 的页面（4xx/5xx 响应）。一个 `3xx` 重定向以及
  一个从未携带状态的页面（`statusCode` 为 `0`/未设置）**不算**失败，仍留在
  `httpPages`/`browserPages` 中。一个非 2xx、随后被浏览器渲染成了内容的页面
  （`source: "browser"`）也**不**计入失败——这次恢复已经清除了失败，尽管该页面仍保留其
  原始 `statusCode`。当 `failedPages` 非零时，recommendations 中会包含
  `"N of M pages returned errors or 4xx/5xx responses"`，于是失效链接会在汇总中浮现，而
  不是被计为成功页面。
- `contentTypes` 仅是**成功**页面的分类体系：非 2xx 页面被排除在外，因此错误响应体绝不会
  被当作普通 `page` 报告。该页面自身的 `statusCode` 与 `markdown` 保持原样不动。

### 模式历史（Schema history）

`schemaVersion` 会被盖到每份报告上，以便消费方检测格式变化。当前版本为 **`2.0`**。

**`1.0` → `2.0`（破坏性变更）。** `site.totalDiscovered` 被**移除**，由
`site.totalURLsInSitemap` 取代。

数值本身没有变——变的是名字。`totalDiscovered` 当初只统计该站点的站点地图里列出的 URL，
从不统计爬取发现的全部内容，因此在没有站点地图的站点上它读出 `0`，而页面明明正在被抓取
（`N page(s) discovered of 0`）。它被重命名，以如实表达它所承载的内容。

一个 `1.0` 读取方需要改动两处：

- 在原先读取 `totalDiscovered` 的地方改读 `totalURLsInSitemap`；两者是同一项度量。
- 它现在是 `omitempty`，因此当站点没有站点地图时，它是**缺席**的，而不是 `0`。应把它
  与 `sitemapFound` 配对使用，而不是把缺失的键当作零。

如果你真正想要的是爬取触及了多少页面，那是 `site.sampledPages`（或 `pages[]` 的长度）——
它从来就不是 `totalDiscovered`。

## HTTP API

命令行界面是一个端点之上的薄客户端：

- `POST /scrape {"url", "preview", "only", "maxPages", "maxPerPattern",
	"includePatterns", "excludePatterns", "concurrency", "enrichAll",
	"noBrowser", "timeoutSeconds", "browser"}` → 抓取报告

每一次爬取 fetch、每一次展开 fetch，都走与浏览器导航相同的 SSRF/重定向导航守卫。

## MCP

代理可以通过 MCP 用 `pinchtab_scrape` 工具驱动同一流程。先用 `preview=true`
拿到廉价大纲，再用 `only`（逗号分隔的 URL）展开——完整报告可能很大。参见
[MCP Server](mcp.md)。

```
pinchtab_scrape { "url": "https://example.com", "preview": true }
pinchtab_scrape { "url": "https://example.com", "only": "https://example.com/a, https://example.com/b" }
```

## Docker / CI

抓取命令行界面是运行中的 pinchtab 服务器之上的一层薄客户端，因此在 Docker 中要先运行
服务器容器，再对其 exec 执行抓取（服务器搭建见
[guides/docker.md](guides/docker.md)）：

```bash
docker run -d --name pinchtab -v pinchtab-data:/data --shm-size=2g pinchtab/pinchtab
docker exec pinchtab pinchtab scrape https://example.com --output-dir /data/scrape
docker cp pinchtab:/data/scrape ./scrape
```
