# 实现

本节涵盖以实现为重点的文档：特定子系统在实践中如何工作、它们做了哪些权衡，以及当前代码如何组织。

当你需要比架构概览更底层的细节、但不需要完整 API 参考材料时，使用这些页面。

- [Lite Engine（历史）](./lite-engine.md)——已移除的 `chrome`/`lite`/`auto` 引擎路由；它的 Gost-DOM 路径作为 `ghost-chrome` 提供者使用的静态 fetch 保留下来（`internal/browsers/ghostchrome/staticfetch`）——见[术语](../architecture/terminology.md)
- [Managed Bridge vs Managed Direct CDP](./managed-bridge-vs-managed-direct-cdp.md)
- [Chrome Profile 锁恢复](./chrome-profile-lock-recovery.md)
- [Chrome 文件](./chrome-files.md)
- [并行标签页执行（设计文档）](./parallel-tab-execution.md)
- [Docker 本地测试](./docker-local-testing.md)
