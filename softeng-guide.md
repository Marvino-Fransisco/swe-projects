# Software Engineering Guide

## Logical Fallacies of Distributed Systems

> An optimistic mindset that cause a software developer doing bad patterns before reality hit them in production

| Fallacy | File | Description |
| --- | --- | --- |
| The network is reliable | [the-network-is-reliable.md](docs/logical-fallacy/the-network-is-reliable.md) | Packets always arrive, and connections never drop |
| Latency is zero | [latency-is-zero.md](docs/logical-fallacy/latency-is-zero.md) | A database query or an API request across the internet takes exactly 0 milliseconds |
| Bandwidth is infinite | [bandwidth-is-infinite.md](docs/logical-fallacy/bandwidth-is-infinite.md) | Developers treat the network like a giant, bottomless pipe |
| The network is secure | [the-network-is-secure.md](docs/logical-fallacy/the-network-is-secure.md) | Internal traffic behind a firewall or private cloud is safe |
| Topology doesn't change | [topology-doesnt-change.md](docs/logical-fallacy/topology-doesnt-change.md) | Server IPs and network paths are fixed and unchanging |
| There's one administrator | [theres-one-administrator.md](docs/logical-fallacy/theres-one-administrator.md) | You have full access and root permissions to debug and fix any issue |
| Component versioning is simple | [component-versioning-is-simple.md](docs/logical-fallacy/component-versioning-is-simple.md) | All services can be instantly updated to the latest version |
| Observability can be delayed | [observability-can-be-delayed.md](docs/logical-fallacy/observability-can-be-delayed.md) | Logging, metrics, and tracing can be added right before launch |

## Cloud Design Patterns

Practical tradeoffs and edge case analyses for cloud architecture patterns, based on [Microsoft Azure Architecture Center](https://learn.microsoft.com/en-us/azure/architecture/patterns/) documentation.

| Pattern | File | Description |
| --- | --- | --- |
| Ambassador | [ambassador.md](docs/cloud-design-pattern/ambassador.md) | Helper services that offload cross-cutting concerns like retry, circuit breaking, and TLS from client applications |
| Anti-Corruption Layer | [anti-corruption-layer.md](docs/cloud-design-pattern/anti-corruption-layer.md) | Translation boundary that preserves domain model integrity during integration with legacy or external systems |
| Asynchronous Request-Reply | [asynchronous-request-reply.md](docs/cloud-design-pattern/asynchronous-request-reply.md) | Decouples long-running back-end processing from front-end hosts using HTTP 202 polling |
| Backends for Frontends | [backends-for-frontends.md](docs/cloud-design-pattern/backends-for-frontends.md) | Separate backend services tailored to the specific needs of each frontend interface |
| Bulkhead | [bulkhead.md](docs/cloud-design-pattern/bulkhead.md) | Isolates application elements into pools so that failures in one don't cascade to others |
| Cache-Aside | [cache-aside.md](docs/cloud-design-pattern/cache-aside.md) | Loads data on demand into a cache from a data store to improve performance and maintain consistency |
| Choreography | [choreography.md](docs/cloud-design-pattern/choreography.md) | Decentralizes workflow logic by letting services independently decide when and how to process business operations |
| Circuit Breaker | [circuit-breaker.md](docs/cloud-design-pattern/circuit-breaker.md) | Temporarily blocks access to a failing service to prevent cascading failures and allow recovery |
| Claim Check | [claim-check.md](docs/cloud-design-pattern/claim-check.md) | Splits large messages into a claim-check token and an external payload to avoid overwhelming the message bus |
| Compensating Transaction | [compensating-transaction.md](docs/cloud-design-pattern/compensating-transaction.md) | Undoes work from completed steps when a multi-step eventually consistent operation fails |
| Competing Consumers | [competing-consumers.md](docs/cloud-design-pattern/competing-consumers.md) | Enables multiple concurrent consumers to process messages from the same channel for throughput and scalability |
| Compute Resource Consolidation | [compute-resource-consolidation.md](docs/cloud-design-pattern/compute-resource-consolidation.md) | Consolidates multiple tasks into a single compute unit to increase utilization and reduce costs |
| CQRS | [cqrs.md](docs/cloud-design-pattern/cqrs.md) | Segregates read and write operations into separate models for independent optimization and scaling |
