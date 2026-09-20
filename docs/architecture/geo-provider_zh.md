# 地理 Provider（Geo Provider）

## 概览

地理 provider（`internal/config/geo`）为代理出口 IP 解析地理信息，让用户无需手工填写 `browser.proxy.geo.*`，即可配置"自动"地理对齐。

该包目前只提供 `Noop` 和 `Static` 两个 provider。本说明记录了一个未来基于 HTTP 的 provider 的拟议契约，使设计留在文档中，而不是作为已编译包中的累赘代码。

## 未来：基于 HTTP 的 GeoProvider

未来的阶段将增加一个 `HTTPGeoProvider`，针对外部服务（ipinfo.io、Maxmind GeoLite2、ip-api 等）解析代理出口 IP。

### 拟议契约

```go
type HTTPGeoProvider struct {
    Endpoint string       // e.g. "https://ipinfo.io/{ip}/json"
    Token    string       // optional bearer/api key, redacted in logs
    Client   *http.Client // injected for testability; nil → http.DefaultClient with a short timeout
    Cache    GeoCache     // in-memory TTL cache keyed by IP
}

func (h HTTPGeoProvider) Lookup(ctx context.Context, ip string) (Info, error) {
    // 1. ip == "" → return Info{}, nil (best-effort, never fail)
    // 2. cache hit → return cached Info, nil
    // 3. ctx-aware GET to Endpoint, parse provider-specific JSON, map to
    //    Info, cache, return.
    // 4. on any HTTP/parse error → return Info{}, nil and log at Warn.
    //    Geo alignment is a hint, not a hard requirement; a bad lookup
    //    must not block launch.
}
```

### 悬而未决的问题

- 配置面（`browser.proxy.geo.http.{endpoint,token,ttl}`？）。
- 是否支持多 provider 故障转移（failover）。
- 缓存淘汰策略与 TTL 默认值。
- 如何在 `/stealth/status` 或仪表板上呈现查询状态。

在这些问题确定之前，不发布任何可执行代码，以免引入一条只接了一半的网络路径。
