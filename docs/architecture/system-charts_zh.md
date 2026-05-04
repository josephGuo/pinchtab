# 系统图表

本页收集了当前 PinchTab 架构的主要高级图表。

## 图表 1：产品形状

```mermaid
flowchart TD
    U["Agent / 命令行界面 / Tool"] --> S["PinchTab Server"]

    S --> D["Dashboard + Config + Profiles API"]
    S --> O["Orchestrator + Strategy Layer"]

    O --> M1["Managed Instance"]
    O --> M2["Managed Instance"]

    M1 --> B1["pinchtab bridge"]
    M2 --> B2["pinchtab bridge"]

    B1 --> C1["Chrome"]
    B2 --> C2["Chrome"]

    C1 --> T1["Tabs"]
    C2 --> T2["Tabs"]

    S -. "advanced attach path" .-> E["Registered External Chrome"]
```

这是今天的默认系统形状：

- 代理通过 HTTP 与服务器通信
- 服务器管理配置文件、实例和路由
- 管理实例由桥接支持
- 附加作为高级外部浏览器注册路径存在

## 图表 2：主要使用路径

```mermaid
flowchart LR
    I["Install PinchTab"] --> R["Run: pinchtab server"]
    R --> L["Local server on localhost:9867"]
    L --> A["Agent / 命令行界面 sends HTTP requests"]
    A --> W["Browser work happens through PinchTab"]
```

这是用户的正常心智模型。大多数用户应该考虑 `pinchtab server`，而不是 `pinchtab bridge`。

## 图表 3：运行时形状

```mermaid
flowchart LR
    subgraph S1["Server Mode"]
        C1["Client"] --> P1["pinchtab server"]
        P1 --> B1["pinchtab bridge"]
        B1 --> CH1["Chrome"]
        CH1 --> T1["Tabs"]
    end

    subgraph S2["Bridge Mode"]
        C2["Client"] --> B2["pinchtab bridge"]
        B2 --> CH2["Chrome"]
        CH2 --> T2["Tabs"]
    end
```

含义：

- **服务器模式** 是多实例控制平面路径
- **桥接模式** 是单实例浏览器运行时

## 图表 4：当前请求路径

```mermaid
flowchart TD
    R["HTTP Request"] --> M["Auth + Middleware"]
    M --> T{"Route Type"}

    T -->|Direct browser route| X["Tab / Instance Resolution"]
    T -->|Task route, when enabled| Q["Scheduler"]
    T -->|Attach route| A["Attach Policy Check"]

    Q --> X
    X --> H["Bridge Handler"]
    H --> P["Handler-Level Policy Checks"]
    P --> C["Chrome via CDP"]
    C --> O["JSON / Text / PDF / Image Response"]

    A --> AR["Register External Instance"]
```

重要细节：

- 身份验证和共享中间件在 HTTP 层运行
- 附加策略在服务器的附加路由上强制执行
- IDPI 和类似的面向浏览器的检查在导航、文本和快照等处理程序中运行
- 标签页范围的路由在执行前解析到拥有实例
- 调度器是可选的，仅服务器端，适用于 `/tasks`
- 桥接处理程序执行实际的浏览器工作