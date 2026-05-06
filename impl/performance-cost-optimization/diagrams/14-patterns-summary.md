# Architecture Patterns Summary

```mermaid
graph TB
    subgraph "Patterns Demonstrated"
        BFF["Backend For Frontends<br/>-------------------------<br/>web-bff: cookies, full payloads<br/>mobile-bff: bearer, slim payloads"]
        CRC["Compute Resource Consolidation<br/>-------------------------<br/>5 processes in 1 container<br/>Avoids idle resource waste"]
        CA1["Cache-Aside: Write DB First<br/>-------------------------<br/>Applied to: User Profile<br/>Strong consistency"]
        CA2["Cache-Aside: Prefill<br/>-------------------------<br/>Applied to: Product Catalog<br/>Maximum read speed"]
        WB["Write-Behind: Cache First<br/>-------------------------<br/>Applied to: Cart, View Counter<br/>Low write latency"]
    end

    BFF --- CRC
    CRC --- CA1
    CRC --- CA2
    CRC --- WB
```

## Trade-offs

| Strategy | Latency | Freshness | Complexity | Use When |
|---|---|---|---|---|
| Cache-Aside (write DB first) | Higher writes | Immediate | Medium | Data integrity matters |
| Write-Behind (cache first) | Lowest writes | Eventual | High | Frequent writes, ok with delay |
| Cache-Aside (prefill) | Lowest reads | Periodic | Low | Read-heavy, infrequent changes |

## Consistency Spectrum

```mermaid
graph LR
    Strong["Strong Consistency<br/>(Profile)"] --> CA["Cache-Aside<br/>Write DB First"]
    CA --> WB["Write-Behind<br/>Write Cache First"]
    WB --> EP["Eventual Consistency<br/>(Cart, Views)"]

    style Strong fill:#ffcdd2
    style EP fill:#c8e6c9
```
