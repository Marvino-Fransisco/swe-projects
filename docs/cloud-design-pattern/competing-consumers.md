# Competing Consumers Pattern — Tradeoffs & Edge Cases

## The Full Tradeoffs of the Competing Consumers Pattern

### ✅ Benefits (the "why you'd want it")

**1. Natural horizontal scaling under variable load**
Traffic spikes? Add more consumer instances. Quiet period? Scale down to one (or zero, with serverless). The queue acts as a buffer that absorbs burst traffic, and the consumer pool expands and contracts to match. You don't need to predict peak load in advance — you autoscale to match it. This is the simplest and most effective scaling model for asynchronous workloads.

**2. Built-in load balancing with no custom coordination**
The message queue does the distribution. It delivers each message to exactly one consumer — no leader election, no work-stealing algorithm, no custom partitioning logic. Consumers just pull messages from the queue and process them. The queue is the coordinator, and it already exists as infrastructure.

**3. Consumer failure doesn't lose messages**
If a consumer crashes mid-processing, the message isn't lost. With PeekLock (Service Bus) or visibility timeout (SQS), the unacknowledged message becomes visible again and another consumer picks it up. No single consumer is a single point of failure. The system self-heals by redistributing orphaned work to healthy instances.

**4. Decouples producers from consumers**
The producer doesn't know or care how many consumers exist, whether they're healthy, or how long processing takes. It posts a message and moves on. The consumer doesn't know or care who sent the message. This temporal and structural decoupling means producer and consumer can be deployed, scaled, and updated independently.

**5. Cost-efficient with serverless compute**
When backed by serverless consumers (Azure Functions, AWS Lambda), the system scales to zero during idle periods. You pay only for actual message processing, not for standing infrastructure. Combined with queue-based load leveling, this makes competing consumers one of the most cost-effective patterns for bursty workloads.

**6. Works with heterogeneous consumer capabilities**
Not all consumer instances need to be identical. You can run a mix of consumer types (e.g., lightweight instances for simple messages, high-memory instances for heavy processing) against the same queue. As long as each consumer can handle any message it receives, the heterogeneous pool works fine.

---

### ❌ Tradeoffs / Problems

**1. Message ordering is destroyed**
With multiple consumers pulling messages concurrently, the processing order is unpredictable. Message 3 might finish before Message 1. Message 2 might start processing before Message 1 even begins. If your workflow depends on ordering (e.g., process account creation before processing a deposit), the competing consumers pattern breaks that assumption entirely.

**2. Idempotency is mandatory, not optional**
Messages can be delivered more than once. A consumer crashes after processing but before acknowledging — the message is redelivered. A network blip causes a duplicate delivery. Two consumers receive the same message due to a queue bug. Every consumer must be able to process the same message multiple times without side effects. If your processing isn't idempotent, competing consumers will produce incorrect results.

**3. Poison messages block the system**
A malformed message or a message that triggers a bug in the consumer will cause every consumer that touches it to crash or fail. The message gets retried, fails again, returns to the queue, and the cycle repeats. Without a dead-letter queue with a maximum delivery count, a single poison message can consume all your consumer capacity in an endless retry loop.

**4. Results must be communicated through a side channel**
The consumer that processes a message is decoupled from the producer. The producer doesn't know which consumer handled the message or when. If the producer needs the result, it must use a separate mechanism: a reply queue, a shared database, a callback URL, or polling. This result-handling infrastructure is additional complexity that the pattern doesn't address.

**5. The queue itself can become a bottleneck**
At very high throughput, a single queue becomes the limiting factor. Message brokers have per-queue throughput limits (Service Bus: ~2000 messages/second per queue, SQS: ~3000 with batching). If your producers generate more than the queue can handle, messages back up, latency increases, and consumers starve waiting for the queue to deliver.

**6. Autoscaling lag creates temporary backlogs**
Autoscaling isn't instantaneous. When a traffic spike arrives, it takes time for new consumer instances to start (30–60 seconds for containers, potentially longer for VMs). During that ramp-up period, messages accumulate in the queue and processing latency spikes. By the time consumers are fully scaled out, the spike may be over — and now you're over-provisioned.

**7. Complex message types reduce parallelism**
If messages vary widely in processing cost (most take 10ms, but some take 10 minutes), a single slow message can occupy a consumer while fast messages pile up. With a fixed number of consumers, one "heavy" message effectively reduces your consumer pool. The queue doesn't know which messages are expensive — it distributes them blindly.

---

### Summary Table

| Tradeoff | Severity | Notes |
| --- | --- | --- |
| Message ordering | High | Concurrent processing destroys FIFO guarantees |
| Idempotency requirement | High | At-least-once delivery means duplicate processing |
| Poison messages | High | Can exhaust all consumers in a retry loop |
| Result communication | Medium | Requires a side channel (reply queue, database) |
| Queue throughput bottleneck | Medium | Single queue limits at high volume |
| Autoscaling lag | Medium | Traffic spikes create temporary backlogs |
| Uneven message cost | Medium | Heavy messages reduce effective parallelism |

---

## Edge Case: Messages That Must Be Processed in Order Within a Group

### The Core Tension

The competing consumers pattern maximizes throughput by distributing messages across consumers freely. But some workloads require ordering within a subset of messages — all messages for a given order ID must be processed sequentially, while messages for different order IDs can be processed in parallel. You need both competing consumers (for throughput) and per-group ordering (for correctness). The pattern gives you one but breaks the other.

---

### Approach 1: Message Sessions (Session-Affinity Routing)

Use a messaging system that supports sessions (e.g., Service Bus message sessions). Assign each message a session ID (e.g., the order ID). The queue ensures all messages with the same session ID are delivered to the same consumer, in order. Different sessions are distributed across consumers for parallel processing.

```text
Queue: [OrderA-1, OrderB-1, OrderA-2, OrderC-1, OrderB-2, OrderA-3]
Consumer 1: receives session OrderA → processes OrderA-1, OrderA-2, OrderA-3 in order
Consumer 2: receives session OrderB → processes OrderB-1, OrderB-2 in order
Consumer 3: receives session OrderC → processes OrderC-1
```

**Pros:** Ordering is guaranteed within each session. The message broker handles session routing — no custom logic in consumers. Different sessions are processed in parallel, preserving throughput for independent groups.

**Cons:** Sessions are sticky — once a consumer accepts a session, it holds that session until all messages are processed or the session times out. If OrderA has 1,000 messages, one consumer is tied up processing all of them while other consumers sit idle. The throughput gain is limited by the number of active sessions, not the number of consumers. A "hot" session (a single order with many events) becomes a serial bottleneck.

---

### Approach 2: Partitioned Queues (One Queue Per Group)

Instead of one queue, create multiple queues — one per group (or a hash-based mapping of groups to queues). Each queue has its own set of competing consumers. Messages for the same group always go to the same queue, preserving order. Different groups are processed in parallel across different queues.

```text
Hash(order_id) % 3 → Queue 0, Queue 1, Queue 2
Queue 0: consumers [A, B] → competing for Queue 0 messages
Queue 1: consumers [C, D] → competing for Queue 1 messages
Queue 2: consumers [E, F] → competing for Queue 2 messages
OrderA hashes to Queue 1 → always goes to Queue 1 → ordered processing
OrderB hashes to Queue 0 → always goes to Queue 0 → ordered processing
```

**Pros:** Ordering within each group is guaranteed (one queue per partition). Parallelism is preserved across partitions. Well-understood pattern (consistent hashing).

**Cons:** Requires N queues and N consumer pools — more infrastructure to manage. Load balancing depends on the hash function distributing groups evenly; a skewed distribution (80% of orders hash to one partition) creates an imbalanced system. Adding or removing queues requires rehashing all groups, which disrupts ordering during migration.

---

### Approach 3: Single Consumer with In-Process Sequential Grouping

Use a single consumer (or a small pool) that pulls all messages and maintains an in-memory mapping of which groups are currently being processed. Messages within the same group are processed sequentially; messages across groups are processed concurrently using in-process parallelism.

```text
Single consumer pulls: [OrderA-1, OrderB-1, OrderA-2, OrderC-1]
In-process scheduler:
  - Start OrderA-1 (OrderA lock acquired)
  - Start OrderB-1 (OrderB lock acquired) — parallel with A
  - Wait on OrderA-2 (OrderA lock held) — queued
  - Start OrderC-1 (OrderC lock acquired) — parallel with A and B
  - OrderA-1 completes → start OrderA-2
```

**Pros:** Ordering within groups is guaranteed. No infrastructure changes — just one consumer with smarter scheduling. Groups are processed in parallel up to the consumer's concurrency limit.

**Cons:** Single consumer is a SPOF. If it crashes, all processing stops. Throughput is limited by one instance's capacity. In-process parallelism doesn't scale horizontally — you'd need to shard at the consumer level, which reintroduces the partitioned queues problem (Approach 2). State management (which groups are locked, which messages are pending) adds complexity.

---

### Approach 4: Sequence Numbers with Deferred Processing

Each message carries a sequence number within its group. Consumers pull messages freely (competing consumers), but before processing, they check whether all prior messages in the group have been completed. If not, the message is deferred — returned to the queue or held in a waiting area until its turn arrives.

```text
Consumer 1 pulls OrderA-3 → checks: OrderA-1 done? No. OrderA-2 done? No → defer
Consumer 2 pulls OrderA-1 → checks: no prior messages → process → mark done
Consumer 3 pulls OrderA-2 → checks: OrderA-1 done? Yes → process → mark done
Consumer 1 re-pulls OrderA-3 → checks: OrderA-1 done? Yes. OrderA-2 done? Yes → process
```

**Pros:** Full competing consumers throughput for groups that have only one message (the common case). No session affinity or partitioned queues needed. Works with any message broker.

**Cons:** Requires a shared state store to track which sequence numbers have been completed per group — a database, a cache, or a metadata store that all consumers can query. Every message now requires a read-before-process check, adding latency. Deferred messages consume queue space and consumer cycles. Under heavy load, a backlog of deferred messages can grow faster than it's resolved, creating a death spiral of re-deferred messages.

---

### The Real Question to Ask for Each Message Type

Before choosing an ordering strategy, for each message type ask:

1. **Does this message type actually require ordering?**
   → No: standard competing consumers. Don't over-engineer.
   → Yes, across all messages: you can't use competing consumers. Use a single consumer or a sequential queue.
   → Yes, within groups: you need one of the approaches above.

2. **How many distinct groups exist, and how skewed is the distribution?**
   → Many groups, even distribution: sessions (Approach 1) or partitioned queues (Approach 2) work well.
   → Few groups or highly skewed (a few groups dominate): sessions will create hot consumers. Consider sequence numbers (Approach 4).

3. **How many messages per group?**
   → 1–3 messages per group: ordering is mostly a non-issue. Competing consumers with a tolerance for occasional reordering.
   → 10+ messages per group: you need explicit ordering. Sessions or sequence numbers.

4. **What's the cost of processing a message out of order?**
   → Low (cosmetic, easily corrected): tolerate it. Don't add complexity.
   → High (financial inconsistency, data corruption): enforce ordering. Use sessions or partitioned queues.

5. **Can you tolerate a shared state dependency?**
   → Yes: sequence numbers (Approach 4) provide the most flexibility.
   → No: partitioned queues (Approach 2) keep state within the messaging infrastructure.

---

### The Uncomfortable Truth

The competing consumers pattern optimizes for one thing: throughput through parallelism. The moment you need ordering, you're fighting the pattern's core design. Every approach to preserve ordering — sessions, partitions, sequence numbers — reduces parallelism, adds complexity, or introduces new failure modes. The pattern's strength (any consumer can process any message) is exactly what makes ordering hard. If ordering is a hard requirement for most of your messages, competing consumers is the wrong foundation. If ordering is needed for a small subset of messages or a small subset of groups, the hybrid approaches above can work — but they require you to be honest about the throughput cost. You're no longer getting the full benefit of competing consumers; you're getting a constrained version that's more complex than either pure competing consumers or pure sequential processing.
