---
name: pinchtab-docs-zh-review
overview: 对 PinchTab 文档仓库内全部 93 组英文 Markdown 及其 _zh 中文版本进行逐篇对照核对与修正，覆盖顶层与 architecture/guides/reference/implementations/deep-dive 全部子目录，按目录分批推进。
todos:
  - id: verify-top-level
    content: 核对并修正顶层 docs/ 下 13 组 *_zh.md，仅改中文不改英文源
    status: completed
  - id: verify-reference-a
    content: 用 [skill:dispatching-parallel-agents] 并行核对 reference/ 前 20 组（a11y–mcp-tools）
    status: completed
  - id: verify-reference-b
    content: 核对 reference/ 后 20 组（memory–type）中文文档准确性
    status: completed
  - id: verify-guides
    content: 核对并修正 guides/ 下 17 组 *_zh.md 内容与术语
    status: completed
  - id: verify-architecture
    content: 核对并修正 architecture/ 下 14 组 *_zh.md 结构一致性
    status: completed
  - id: verify-impl-deepdive
    content: 核对 implementations/ 7 组与 deep-dive/ 2 组 *_zh.md
    status: completed
  - id: final-consistency
    content: 用 [skill:verification-before-completion] 全站术语一致性与结构终检
    status: completed
    dependencies:
      - verify-top-level
      - verify-reference-a
      - verify-reference-b
      - verify-guides
      - verify-architecture
      - verify-impl-deepdive
---

## 产品概述

对 `docs` 目录下的英文 Markdown 文档进行中文化，产出与原文件名对应、以 `_zh` 结尾的中文文档；由于各英文文档当前均已存在对应 `_zh.md`，本次工作核心为**逐篇对照英文源重新核对并修正中文译文**，确保中文文档与英文源在结构、内容与术语上准确一致。

## 核心功能

- 全量核对 6 个目录共 93 组英文/中文文档配对（顶层 13、architecture 14、guides 17、reference 40、implementations 7、deep-dive 2）。
- 修正中文文档中与英文源不一致、遗漏、误译或多余的内容。
- 统一专有名词译法与中文排版（中英文空格、全角标点）。
- 严格保留代码块、命令、CLI 参数、API 路径、配置键与链接 URL 原样，确保可执行性。
- 保持 Markdown 结构（标题层级、列表、表格、代码围栏语言标记）与英文源一一对应，文档可正常渲染。

## 技术方案概述

- **任务性质**：纯文档语言处理任务，无代码实现，不引入任何运行时依赖。
- **产物**：仅修改 `<name>_zh.md` 文件，**绝不改动英文源文档**，不新增/删除文档文件（除非发现英文文档确实缺失 `_zh` 对应版本）。

## 翻译与校对规范

- **保留英文不译**：代码块内容（含注释）、命令行、CLI 参数、API 路径、配置键（如 `browser.proxy.geo.*`）、环境变量、URL、文件路径、标识符/类型名/专有术语（如 provider、token、snapshot）。
- **译为简体中文**：标题（采用「英文标题 + 中文」或纯中文惯例）、说明性正文、表格文本、列表描述、引用提示。
- **中文排版**：中英文之间加空格；使用全角标点；引号统一为弯引号。
- **链接沿用现有约定**：中文文档中的站内链接保持指向英文 `.md`（如 `[安全指南](guides/security.md)`），不改写为 `_zh.md`。
- **代码围栏语言标记不变**：如 ```bash、```go、```json 等保持原样。

## 逐篇核对方法

1. **结构对齐**：比对标题层级与数量、列表项数、表格行列数、代码块数量及语言标记，找出漏译/多译/结构错位。
2. **内容比对**：逐段核对译文是否覆盖英文全部信息，修正误译、错译与语义偏差。
3. **术语一致性**：抽查 provider、token、dashboard、instance、tab、snapshot、scheduler 等术语译法是否统一（现有文档存在 dashboard 时而保留时而译为「仪表板/仪表盘」等不一致，需统一）。
4. **执行性校验**：确认命令、代码、路径、配置键未被翻译或破坏。
5. **渲染校验**：确认 Markdown 语法可正常渲染，无破坏的锚点、围栏或表格。

## 分批推进策略（按目录，规模降序，批次间相互独立可并行）

- 批次 A：顶层 `docs/*.md`（13 组）
- 批次 B：`reference/`（40 组，拆为两个子批：前 20 组、后 20 组）
- 批次 C：`guides/`（17 组）
- 批次 D：`architecture/`（14 组）
- 批次 E：`implementations/`（7 组）+ `deep-dive/`（2 组）
- 收尾：全站术语一致性、结构完整性终检

## 目录结构（仅列出将被修改的 `_zh.md` 文件，均标注 [MODIFY]）

```
docs/
├── *_zh.md                                  # [MODIFY] 13 个：audit, benchmark, commands, core-concepts, dashboard, endpoints, get-started, headless-vs-headed, mcp, pinchtab-scrape-audit-spec, pinchtab, scrape, showcase
├── architecture/*_zh.md                     # [MODIFY] 14 个：autosolver, browser-abstraction, browser-runtime, extract, find, geo-provider, index, instance-charts, mcp, orchestration, routing-contract, scheduler, system-charts, terminology
├── guides/*_zh.md                           # [MODIFY] 17 个：agent-identity, annotate-for-llm-fixes, attach-chrome, cloakbrowser, contributing, daemon, data-storage, docker, headed-mode, identifying-instances, mcp-agents, memory-monitoring, multi-instance, raspberry-pi, remote-bridge-orchestrator, security, tailscale-bridge-orchestrator
├── reference/*_zh.md                        # [MODIFY] 40 个：a11y, cache, capture, cli, click, config, dialog, eval, extract, fill, find, focus, frame, handoff, health, hover, index, instances, keyboard, mcp-tools, memory, metrics, mouse, navigate, pdf, press, profiles, record, scheduler, screenshot, scroll, select, sessions, snapshot, solve, state, strategies, tabs, text, type
├── implementations/*_zh.md                  # [MODIFY] 7 个：chrome-files, chrome-profile-lock-recovery, docker-local-testing, index, lite-engine, managed-bridge-vs-managed-direct-cdp, parallel-tab-execution
└── deep-dive/*_zh.md                        # [MODIFY] 2 个：benchmark, lite-engin-prototype
```

说明：每个 `_zh.md` 以同目录同名英文 `.md` 为唯一基准进行核对；修改内容限定为译文的准确性、完整性、术语统一与中文排版，绝不触碰英文源与代码/命令/链接。

## Agent Extensions

### Skill

- **dispatching-parallel-agents**
- Purpose: 将相互独立、无共享状态的目录批次（顶层、reference 子批、guides、architecture、implementations/deep-dive）并行推进，提升核对效率。
- Expected outcome: 各批次并行完成核对与修正，产出符合规范的中文文档。
- **verification-before-completion**
- Purpose: 在最终声明完成前，对全部 93 组文档执行结构化验证（结构与术语一致性、代码块/链接未破坏、可渲染）。
- Expected outcome: 以证据确认全部配对准确一致，无遗漏或破坏性改动。