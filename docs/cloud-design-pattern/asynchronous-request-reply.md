# Asynchronous Request-Reply Pattern — Tradeoffs & Edge Cases

## The Full Tradeoffs of the Asynchronous Request-Reply Pattern

### ✅ Benefits (the "why you'd want it")

**1. Decouples back-end processing from front-end hosts**
The front end never blocks on a long-running operation. It fires a request, gets an immediate HTTP 202 (Accepted), and moves on. The back end processes the work at its own pace, scaling independently. This separation is critical when processing times range from seconds to minutes — a synchronous call would either timeout or hold resources hostage.

**2. Works with any HTTP client, anywhere**
No WebSockets, no callbacks, no open inbound ports, no special libraries. The client just makes standard HTTP GET requests to poll for status. This works behind corporate firewalls, in browser applications, in mobile apps with flaky connectivity, and in server-to-server calls between restricted networks. If you can make an HTTP request, you can use this pattern.

**3. Avoids timeout failures for long-running operations**
Most HTTP clients, load balancers, and proxies have default timeouts (30s, 60s, 120s). A video transcoding job, a large report generation, or a bulk data import will blow past those limits. The async pattern sidesteps this entirely — the initial request returns instantly, and subsequent polling calls are short-lived status checks.

**4. Simple client-side implementation**
The client doesn't need to manage persistent connections, handle reconnection logic, or run a webhook receiver. It just polls a URL on a timer. This simplicity matters when you don't control the client (third-party integrations, mobile apps, low-code platforms like Logic Apps) or when callback infrastructure isn't worth building for a single use case.

**5. Natural cancellation support**
By exposing a DELETE endpoint on the status resource, clients can cancel long-running operations. The back end receives the cancellation instruction and cleans up. This is harder to implement with fire-and-forget messaging patterns where there's no feedback loop.

---

### ❌ Tradeoffs / Problems

**1. Polling overhead — repeated HTTP requests waste resources**
Every poll is a full HTTP round trip: TCP handshake (or TLS if reused), request parsing, database query to check status, response serialization. If 1,000 clients poll every 5 seconds for a 2-minute operation, that's 24,000 unnecessary requests — most of which return "still pending." The back end must handle this load on top of actual work.

**2. Latency between completion and detection**
The client only learns about completion on its next poll. If `Retry-After` is 10 seconds and the operation finishes 1 second after a poll, the client waits 9 unnecessary seconds. There's an inherent tradeoff: poll more often to reduce latency, or poll less often to reduce load. You can't optimize both simultaneously.

**3. State management complexity**
Every operation needs persistent state storage. The back end must write status updates (pending → running → completed/failed) to a durable store that the status endpoint can read. This means a database, blob storage, or cache — all of which add infrastructure, cost, and failure modes. The state must survive restarts, deployments, and crashes.

**4. Storage and cleanup — orphaned results accumulate**
Status resources and results consume storage indefinitely if no one cleans them up. Clients crash, lose interest, or never follow the redirect. You need a retention policy (e.g., delete results after 24 hours) and a background cleanup job. The `Expires` header can hint at this, but clients may ignore it.

**5. Inconsistent implementations across services**
There is no universal standard. Some services poll the target resource URL directly (returning 404 until ready). Others use a dedicated status endpoint. Some redirect with HTTP 303, others with 302. Azure Resource Manager has its own variant. When you integrate with multiple APIs, you must handle each one's quirks — the pattern is the same, but the details differ.

**6. Ambiguous failure modes**
If the initial POST returns 202 but the client never receives it (network failure), the client retries — potentially creating a duplicate operation. Without an idempotency key (`Idempotency-Key` header), the back end can't distinguish "I already accepted this" from "this is a new request." The client also can't distinguish "the operation was submitted but I lost the response" from "the operation was never submitted."

**7. Not suitable for real-time or high-frequency results**
If the client needs results streamed as they become available, or needs sub-second latency on completion, polling is the wrong tool. Server-Sent Events (SSE), WebSockets, or a message broker are better fits. The async request-reply pattern optimizes for simplicity, not responsiveness.

---

### Summary Table

| Tradeoff | Severity | Notes |
| --- | --- | --- |
| Polling overhead | Medium | Scales with concurrent operations × poll frequency |
| Completion detection latency | Low–Medium | Depends on poll interval; tunable |
| State management | Medium | Requires durable storage for every operation |
| Storage cleanup | Medium | Orphaned results accumulate without a retention policy |
| Implementation inconsistency | Medium | Same pattern, different HTTP semantics per service |
| Ambiguous failure modes | High | Duplicate operations and lost responses are hard to detect |
| Not real-time | High | Hard limit — polling can't match push-based alternatives |

---

## Edge Case: Unpredictable Operation Duration

### The Core Tension

The pattern assumes you can give the client a reasonable `Retry-After` hint. But what if some operations finish in 500ms and others take 45 minutes? A fixed poll interval is either too aggressive for fast operations (wasting requests) or too slow for long ones (the client gives up). Worse, the back end might not know which kind of operation it received until it's halfway through processing it.

---

### Approach 1: Adaptive `Retry-After` (Server-Driven Backoff)

The status endpoint returns a different `Retry-After` value on each poll, based on the operation's current state and estimated remaining time.

```text
Poll 1: Retry-After: 2   (just started, check soon)
Poll 2: Retry-After: 10  (still early, back off)
Poll 3: Retry-After: 30  (long operation, slow down)
Poll 4: HTTP 303        (done)
```

**Pros:** Reduces unnecessary polling. The server knows more about the operation's trajectory than the client does.

**Cons:** Requires the back end to estimate remaining time — which is often a guess. If the estimate is wrong, the client either polls too aggressively or waits too long. The status endpoint now has slightly more complex logic.

---

### Approach 2: Client-Side Exponential Backoff

The client manages its own backoff strategy, ignoring or supplementing the server's `Retry-After`. It starts polling frequently and gradually increases the interval.

```text
Poll 1: wait 1s
Poll 2: wait 2s
Poll 3: wait 4s
Poll 4: wait 8s
...cap at 60s
```

**Pros:** No server-side changes needed. Simple to implement. Works reasonably well across a range of operation durations.

**Cons:** The client has no visibility into actual progress. It might overshoot (waiting 60s when the result was ready after 5s) or undershoot (polling aggressively for a 45-minute job). Different clients may implement different strategies, leading to inconsistent behavior.

---

### Approach 3: Long Polling for Fast Operations, Regular Polling for Slow Ones

Expose a query parameter on the status endpoint that lets the client request long-polling behavior. The server holds the connection open if the result is expected soon, and falls back to immediate 200 responses for long-running operations.

```text
GET /api/status/123?mode=longpoll&timeout=30
→ Server holds connection up to 30s, returns immediately if result is ready
→ If timeout expires, returns 200 with status "pending" and Retry-After
```

**Pros:** Best of both worlds. Fast operations are detected immediately. Slow operations degrade gracefully to regular polling.

**Cons:** Long-held connections consume server resources (thread, memory, connection pool slot). Requires the server to track which operations are "almost done" vs. "just started." More complex to implement and test than simple polling.

---

### Approach 4: Hybrid — Switch to Push Notification When Available

Start with polling (the universal baseline). If the client supports it, also register for a push notification (webhook, SSE, or WebSocket) that fires on completion. The client stops polling once it receives the push notification, or falls back to polling if the push channel drops.

```text
POST /api/work         → 202 + Location header
     + Prefer: respond-async, webhook="https://client/callback"
         ↓
Client polls as backup, but also listens for webhook callback.
First signal wins; the other is ignored.
```

**Pros:** Polling becomes a fallback, not the primary mechanism. Fast notification when push works; guaranteed delivery via polling when it doesn't.

**Cons:** Two notification channels to implement, test, and debug. Requires the client to expose a callback endpoint — which may not be possible behind firewalls or in browser apps. Adds server-side webhook delivery infrastructure (retries, failure tracking).

---

### The Real Question to Ask for Each Operation

Before picking a polling strategy, for each long-running operation ask:

1. **How predictable is the duration?**
   → Highly predictable (e.g., "always ~3 minutes"): fixed `Retry-After` works fine.
   → Unpredictable (e.g., "500ms to 45 minutes"): use adaptive backoff or hybrid.

2. **How latency-sensitive is the client?**
   → "I need the result as soon as it's ready": long polling or push notification.
   → "Within 10–30 seconds is fine": regular polling with a reasonable interval.

3. **How many concurrent operations will poll simultaneously?**
   → Low volume (<100): any approach works. Don't over-engineer.
   → High volume (>10,000): adaptive or server-driven strategies reduce load meaningfully.

4. **Can the client accept push notifications?**
   → Yes: use hybrid (polling + push) for best reliability.
   → No: you're limited to polling. Invest in adaptive `Retry-After`.

5. **What happens if the client polls too aggressively?**
   → If your status endpoint is cheap (cached in-memory lookup): less concern.
   → If it's expensive (database query, blob existence check): rate-limit or return aggressive `Retry-After` values.

---

### The Uncomfortable Truth

The Asynchronous Request-Reply pattern is a pragmatic compromise. It trades real-time responsiveness for universal compatibility. When operation durations are predictable and clients are tolerant of a few seconds of delay, it works beautifully. But when you start optimizing for latency, throughput, and resource efficiency simultaneously, you end up reinventing push notification infrastructure — at which point you should have used WebSockets, SSE, or a message broker from the start. The pattern's strength (simplicity) is also its ceiling.
