# 缓存

清除浏览器的HTTP磁盘缓存。

## 清除缓存

```bash
curl -X POST http://localhost:9867/cache/clear
# 响应
{
  "status": "cleared"
}

# 命令行界面 替代方案（默认人类可读）
pinchtab cache clear
# 输出: OK

pinchtab cache clear --json              # 完整 JSON 响应
```

## 检查状态

```bash
curl http://localhost:9867/cache/status
# 响应
{
  "canClear": true
}

# 命令行界面 替代方案（默认人类可读）
pinchtab cache status
# 输出: can-clear (或 cache-empty)

pinchtab cache status --json             # 完整 JSON 响应
```

## 注意事项

- 清除所有来源的HTTP磁盘缓存
- 不影响 cookies、localStorage 或 sessionStorage
- 在应用重新部署后使用，确保获取新鲜的 JS/CSS 捆绑包
- 可以在没有活动标签页的情况下调用

## 使用场景

**应用重新部署后：** 当 Vite/webpack 应用使用新的 JS 捆绑包哈希重新构建时，过时的缓存捆绑包可能会导致问题。清除缓存以确保获取新鲜资源：

```bash
pinchtab cache clear
pinchtab nav http://localhost:3000
```

**调试缓存问题：** 如果怀疑缓存资源导致问题：

```bash
pinchtab cache clear
pinchtab reload
```

## 相关页面

- [导航](./navigate.md)
- [配置文件](./profiles.md)