# Claim-Check Pattern — Tradeoffs & Edge Cases

## The Full Tradeoffs of the Claim-Check Pattern

### ✅ Benefits (the "why you'd want it")

**1. Bypasses message size limits in the messaging system**
Most message brokers have hard limits — Azure Service Bus caps messages at 256 KB (premium tier at 100 MB), AWS SQS at 256 KB, Kafka's default at 1 MB. The claim-check pattern sidesteps these entirely. The payload lives in blob storage or an object store; the message bus only carries a lightweight token. You can send a 10 GB video file, a 500 MB batch of records, or a massive ML dataset through the same message bus that limits you to 256 KB.

**2. Keeps the message bus healthy under heavy payloads**
Large messages degrade messaging infrastructure. They consume bandwidth, memory, and disk on the broker. They slow down topic subscriptions, increase replication lag across clusters, and make partition rebalancing expensive. By keeping the bus clean with tiny claim-check messages, you preserve its throughput and latency characteristics for all consumers.

**3. Sensitive data never touches the message bus**
If your payload contains PII, credentials, or regulated data, the message bus operators (and anyone with access to queue monitoring tools) can see it. The claim-check pattern stores the sensitive payload in a secured data store with proper access controls. The message bus only sees an opaque token. This separation is meaningful for compliance (GDPR, HIPAA, PCI-DSS) where data must be isolated from infrastructure layers.

**4. Payload can be retrieved only when needed**
Not every consumer needs the full payload. A routing service might only need the message headers to forward the claim check to the right queue. A logging service only needs metadata. These services skip the payload retrieval entirely, saving bandwidth and processing time. The heavy lifting is deferred to the consumer that actually needs the data.

**5. Payload gets the durability guarantees of the data store**
Message buses optimize for throughput, not durability. They might retain messages for hours or days, not weeks. Object stores (Azure Blob Storage, S3) are designed for durable, long-term storage with replication, versioning, and lifecycle policies. The payload gets better reliability guarantees than the message bus could provide.

**6. Conditional application keeps simple cases simple**
You don't have to use claim-check for every message. Implement conditional logic: if the message fits within the broker's limits, send it directly. Only offload to external storage when the payload exceeds the threshold. Small messages avoid the extra latency and complexity of the round trip to blob storage.

---

### ❌ Tradeoffs / Problems

**1. Extra round trips add latency**
The consumer can't process the message immediately. It must: (1) receive the claim check from the message bus, (2) call the data store to fetch the payload, (3) process the payload. That's an extra network hop with its own latency, authentication, and failure modes. For latency-sensitive workflows, this additional round trip can be unacceptable.

**2. The data store becomes a critical dependency**
The payload is no longer self-contained in the message. If the data store goes down, consumers can't process messages even if the message bus is healthy. You've traded one dependency (the message bus) for two (message bus + data store). The data store must be at least as reliable as the message bus, or it becomes the weakest link.

**3. Payload cleanup is a distributed garbage collection problem**
Every payload stored in the data store must eventually be deleted after consumption. But who deletes it? The consumer? The producer? A background cleanup job? If multiple consumers exist, who decides "everyone is done with this payload"? If cleanup fails, payloads accumulate indefinitely, consuming storage and cost. If cleanup is too aggressive, a slow consumer might find the payload gone before it retrieves it.

**4. Two-phase consistency between message and payload**
The producer must: (1) write the payload to the data store, (2) send the claim-check message to the bus. If step 1 succeeds but step 2 fails, you have an orphaned payload in the data store with no message referencing it. If step 2 could be retried and step 1 runs again, you get duplicate payloads. The two-phase write requires careful ordering — persist the payload first, then send the message — and a strategy for handling partial failures.

**5. Claim-check tokens can become stale or broken references**
If the payload is deleted (by lifecycle policy, manual cleanup, or data store failure), the claim-check message becomes a dangling pointer. The consumer receives the message, tries to fetch the payload, and gets a 404. There's no way to recover — the data is gone. This is especially dangerous with aggressive TTL policies on the data store that don't account for slow consumers or replayed messages.

**6. Increased operational surface area**
You now operate two systems instead of one. Monitoring must cover message bus health AND data store health. Access control must be managed on both systems. Deployment changes must be coordinated — a schema change in the payload format requires updating both the producer's upload logic and the consumer's download logic. The pattern multiplies operational complexity.

**7. Not all intermediaries can be bypassed**
Some middleware inspects message bodies for routing, filtering, or content-based logic. An API gateway that routes based on message content can't route on a claim check — it doesn't have the payload. An event processor that transforms messages can't transform what it can't see. These intermediaries either need access to the data store (coupling them to your storage layer) or the claim-check pattern breaks their functionality.

---

### Summary Table

| Tradeoff | Severity | Notes |
| --- | --- | --- |
| Extra latency from round trip | Medium | One additional data store fetch per message |
| Data store as critical dependency | High | Data store outage blocks all processing |
| Payload cleanup complexity | High | Distributed garbage collection is error-prone |
| Two-phase consistency | Medium | Partial failures create orphaned or duplicate payloads |
| Stale claim-check references | Medium | Deleted payloads break consumers |
| Operational surface area | Medium | Two systems to monitor, secure, and coordinate |
| Intermediary inspection | Medium | Content-based routing can't work with claim checks |

---

## Edge Case: Multiple Consumers with Different Processing Speeds

### The Core Tension

A producer publishes a claim-check message to a topic with three subscribers: a fast real-time processor, a medium-speed batch aggregator, and a slow audit logger. The fast consumer processes and finishes in seconds. The audit logger might not process it for hours — or days if it falls behind. When should the payload be deleted? Delete it immediately after the fast consumer finishes and the slow consumer gets a 404. Wait for all consumers and payloads accumulate forever. The claim-check pattern decouples the message from the payload, but now the payload's lifetime must be coordinated across all consumers with wildly different processing speeds.

---

### Approach 1: Reference Counting with Explicit Acknowledgment

Each payload tracks how many consumers have acknowledged it. When a consumer finishes processing, it sends an acknowledgment (not just a message receipt — actual processing completion). When all consumers have acknowledged, a cleanup job deletes the payload.

```text
Payload P1: consumers=[A, B, C], acknowledged=[]
Consumer A finishes → acknowledged=[A]
Consumer B finishes → acknowledged=[A, B]
Consumer C finishes → acknowledged=[A, B, C] → delete P1
```

**Pros:** No premature deletion. Every consumer is guaranteed access to the payload.

**Cons:** Requires a tracking mechanism (a database table, a counter in metadata) for every payload. If a consumer crashes without acknowledging, the payload is never deleted. You need timeout-based fallbacks ("if not acknowledged in 7 days, force delete") which re-introduce the risk of premature deletion. The acknowledgment system is itself a distributed coordination problem.

---

### Approach 2: Time-Based Expiration Generous Enough for the Slowest Consumer

Set the payload's TTL in the data store to a duration that exceeds the slowest consumer's expected processing time. If the audit logger can be up to 48 hours behind, set the TTL to 72 hours. After that, the data store automatically deletes the payload.

```text
Fast consumer: processes in seconds
Batch aggregator: processes in hours
Audit logger: processes within 48 hours
TTL: 72 hours → automatic deletion
```

**Pros:** Simple. No coordination between consumers. The data store handles cleanup automatically.

**Cons:** You must know the slowest consumer's worst-case processing time in advance — and pad it generously. If the audit logger falls further behind than expected (a weekend with no processing), payloads expire before it catches up. The TTL is a blunt instrument: either too aggressive (breaking slow consumers) or too conservative (wasting storage).

---

### Approach 3: Lazy Copy — Each Consumer Gets Its Own Payload Copy

When a consumer receives the claim-check message, it immediately copies the payload to its own storage namespace. After the copy succeeds, the consumer no longer depends on the original payload. The producer can delete the original payload at any time.

```text
Producer uploads P1 to shared storage, sends claim check
Consumer A copies P1 → P1_a (Consumer A's namespace)
Consumer B copies P1 → P1_b (Consumer B's namespace)
Consumer C copies P1 → P1_c (Consumer C's namespace)
Producer deletes P1 from shared storage
Each consumer deletes its own copy when done
```

**Pros:** No coordination between consumers. Each consumer owns its own copy and its own lifecycle. The producer can clean up the original immediately after all consumers have copied it (or after a timeout).

**Cons:** Multiplies storage costs — each payload is stored N+1 times (original + N consumer copies). For large payloads, the copy operation itself is expensive and slow. Each consumer must implement copy-then-acknowledge logic, and if the copy fails, the consumer must retry before the original is deleted.

---

### Approach 4: Keep the Payload Indefinitely with Tiered Storage

Don't delete payloads at all — move them through storage tiers based on age. Recent payloads stay in hot storage (fast access). Older payloads move to cool storage (cheaper, slower). Very old payloads move to archive storage (cheapest, slowest). Consumers that fall behind pay the latency cost of fetching from lower tiers.

```text
Age 0–24 hours: hot storage (SSD, fast retrieval)
Age 1–30 days: cool storage (standard blob, seconds retrieval)
Age 30+ days: archive storage (glacier, hours retrieval)
```

**Pros:** No consumer is ever broken by a missing payload. Storage costs are controlled through tiering. Simplest operational model — no deletion logic, no coordination.

**Cons:** Storage costs grow indefinitely. Even with tiering, you're paying to store data that most consumers processed long ago. Archive retrieval times (hours) may be unacceptable for consumers that need to replay old messages. You need lifecycle policies and monitoring to ensure costs don't silently grow.

---

### The Real Question to Ask for Each Payload

Before choosing a cleanup strategy, for each payload type ask:

1. **How many consumers subscribe to this message?**
   → Single consumer: delete on acknowledgment. No coordination needed.
   → Multiple consumers: you need one of the coordinated approaches.

2. **How different are the consumers' processing speeds?**
   → Similar speeds: time-based expiration (Approach 2) works.
   → Orders of magnitude different (seconds vs. days): reference counting (Approach 1) or lazy copy (Approach 3).

3. **How large are the payloads?**
   → Small (<1 MB): lazy copy (Approach 3) is cheap. Don't overthink it.
   → Large (>1 GB): copying is expensive. Use time-based expiration (Approach 2) or indefinite storage (Approach 4).

4. **How long must payloads be retained for compliance?**
   → Regulatory retention (7 years): you're keeping them anyway. Use tiered storage (Approach 4).
   → No retention requirement: aggressive cleanup keeps costs down.

5. **Can consumers tolerate a missing payload?**
   → Yes (can recompute, can skip): aggressive deletion is safe.
   → No (data is irreplaceable): never delete prematurely. Use reference counting (Approach 1) or indefinite storage (Approach 4).

---

### The Uncomfortable Truth

The claim-check pattern solves a specific problem well: "my message is too big for my message bus." But in solving it, it introduces a harder problem: "who owns the lifecycle of a shared resource accessed by independent consumers at different speeds?" The message bus solved this implicitly — a message exists until it's consumed, and the broker manages deletion. By moving the payload out of the broker, you've taken on the broker's lifecycle management job yourself. The pattern looks simple in a diagram (upload, send token, retrieve, process) but the operational reality is distributed garbage collection — one of the notoriously hard problems in systems design. If you can avoid splitting the message from the payload (by compressing it, chunking it, or using a broker with higher limits), do that first. The claim-check pattern is worth the complexity only when the payload genuinely doesn't fit in the message bus.
