# Ambassador Pattern — Tradeoffs & Edge Cases

## The Full Tradeoffs of the Ambassador Pattern

### ✅ Benefits (the "why you'd want it")

**1. Centralized cross-cutting concerns**
Retry logic, circuit breaking, TLS, auth, logging, and metrics all live in one place. Without an ambassador, every service in a polyglot system (Go, Python, Java) would need to re-implement this logic independently — or worse, inconsistently.

**2. Language/framework agnostic**
Because the ambassador is a separate process (a sidecar), it doesn't care what language your app is written in. Your app just makes a plain HTTP call to the ambassador, and the ambassador handles the complexity. This is huge in microservice environments where teams use different stacks.

**3. Legacy application support**
It can be used with legacy applications or other applications that are difficult to modify, in order to extend their networking capabilities. You don't need to touch the old codebase to add retry logic or mTLS — just deploy an ambassador beside it.

**4. Team ownership separation**
A specialized team can implement and maintain security, networking, or authentication features that have been moved to the ambassador. Your platform/infra team owns the ambassador; your product teams don't need to think about it.

---

### ❌ Tradeoffs / Problems

**1. Latency overhead**
Your app calls the ambassador, the ambassador calls the actual service. That's an extra network hop, even if it's localhost. For most apps this is microseconds — negligible. But for high-frequency trading, real-time games, or anything where sub-millisecond latency matters, this overhead is real and meaningful.

**2. Retry safety — idempotency is now your problem**
The ambassador could handle retries, but that might not be safe unless all operations are idempotent. An idempotent operation is one where doing it twice has the same result as doing it once (e.g., a `GET` request, or `SET x = 5`). But if your operation is "charge the customer $100", a blind retry from the ambassador could charge them twice. The ambassador doesn't know the business semantics — it just sees a failed HTTP call.

**3. Single-language systems don't benefit much**
When client connectivity features are consumed by a single language, a better option might be a client library distributed to development teams as a package. If your entire backend is in Go, just write a shared Go library with your retry/circuit-breaker logic. The ambassador shines in polyglot environments — in homogenous ones, it adds complexity for little gain.

**4. Context passing becomes awkward**
Sometimes your application needs to tell the ambassador "don't retry this one" or "use a timeout of 200ms, not the default 2s." Since the ambassador is a separate process, you can't just pass a function argument — you have to design a protocol (usually custom HTTP headers) for this. This adds a layer of convention that every team must know about.

**5. Operational complexity — now you have two things to deploy**
Every application now has a companion process. You need to think about: How do you deploy and version them together? What happens if the ambassador crashes but the app is fine? Who restarts it? In Kubernetes this is somewhat solved (sidecar container in the same Pod), but in other environments it's a real operational burden.

**6. Shared vs. per-client ambassador — a sharp edge**
A shared ambassador is more resource-efficient but becomes a single point of failure — if it goes down, everything it proxies goes down with it. Per-client ambassadors are more resilient but multiply your resource usage. There's no universally right answer; it depends on your load and reliability requirements.

**7. Deeper integration is impossible**
When connectivity features can't be generalized and require deeper integration with the client application, the ambassador pattern is not suitable. For example, if your retry strategy needs to inspect the application's internal state (like "only retry if the local cache is stale"), the ambassador can't do that — it's an external process with no visibility into your app's memory or state.

---

### Summary Table

| Tradeoff | Severity | Notes |
| --- | --- | --- |
| Latency | Low–High | Usually negligible; critical for real-time systems |
| Retry idempotency | High | Can cause data corruption if ignored |
| Overkill for single-language | Medium | A shared library is simpler |
| Context passing awkwardness | Medium | Requires custom header conventions |
| Operational complexity | Medium | Extra process to deploy, monitor, restart |
| Shared vs. per-client decision | Medium | Wrong choice → SPOF or resource bloat |
| Can't do deep app integration | High | Hard limit on what the ambassador can know |

---

## Edge Case: Some Endpoints Need Custom Integration Depth

### The Core Tension

The ambassador assumes networking concerns are separable from business logic. This edge case breaks that assumption for *some* endpoints — not all. The goal is to avoid throwing away the ambassador entirely just because a few endpoints don't fit.

---

### Approach 1: Selective Bypass (Direct Call for Special Endpoints)

Your app calls the ambassador for most endpoints, but for the ones needing deep integration, it calls the downstream service **directly**.

```text
Normal endpoints:   App → Ambassador → Service
Special endpoints:  App → Service (direct)
```

**Pros:** Simple. No changes to the ambassador.

**Cons:** Two patterns in the codebase. New developers won't know which endpoints bypass the ambassador and why. Special endpoints also lose *all* ambassador benefits (logging, TLS, etc.), not just retry logic.

---

### Approach 2: Pass Context Hints to the Ambassador (Opt-Out Per Feature)

Instead of bypassing the ambassador entirely, your app sends hints via headers telling the ambassador **what not to do** for this specific request.

```text
X-Ambassador-Retry: false
X-Ambassador-Timeout: 50ms
X-Ambassador-CircuitBreaker: false
```

The ambassador reads these and adjusts its behavior per-request. The special endpoint still goes through the ambassador, but with its problematic features disabled.

**Pros:** Single traffic path. Logging, TLS, and other safe features still apply.

**Cons:** Requires a convention every team must know and follow. Grows messy as opt-out combinations increase. Can't help if the special endpoint needs to read app-internal state.

---

### Approach 3: Hybrid — Ambassador + In-App Library for Special Endpoints

Acknowledge that the ambassador is the right tool for infrastructure-level concerns, and a shared in-app library is the right tool for business-logic-coupled concerns.

```text
Normal endpoints:   App → Ambassador → Service
Special endpoints:  App (uses internal lib) → Service
```

The in-app library handles *only* the deep integration cases. The ambassador handles everything else. Both share the same underlying config (timeouts, endpoints, credentials) to stay consistent.

**Pros:** Each tool is used where it fits. No awkward workarounds.

**Cons:** Maintaining two things. If the library is language-specific, you lose the polyglot benefit for those endpoints.

---

### Approach 4: Push the Custom Logic Into the Ambassador Itself

Extend the ambassador with endpoint-specific configuration or plugins for special cases. This is essentially what **Envoy proxy** does with its filter chains — you can attach custom WASM plugins or Lua scripts to specific routes.

```text
Route /normal  → default ambassador behavior
Route /special → ambassador + custom filter (your deep logic here)
```

**Pros:** Single traffic path for everything. Custom logic is still centralized.

**Cons:** You're writing business logic inside your infrastructure layer. If the custom logic needs to access app-internal state, this still doesn't solve it.

---

### The Real Question to Ask for Each Special Endpoint

Before picking an approach, for each "special" endpoint ask:

1. **Does it need app-internal state?** (memory, local cache, auth tokens held in-process)
   → Ambassador *cannot* help here, no matter what. Use bypass or in-app library.

2. **Does it just need different retry/timeout behavior?**
   → Context hints (Approach 2) are sufficient.

3. **Is the custom logic infrastructure-level or business-logic-level?**
   → Infrastructure → extend the ambassador (Approach 4). Business logic → in-app library (Approach 3).

4. **How many special endpoints are there?**
   → If it's 2–3, selective bypass is pragmatic. If it's 20+, you need a systematic approach or the ambassador pattern may not be the right fit for your system.

---

### The Uncomfortable Truth

When "several endpoints need custom integration depth," it's often a signal that your system's networking concerns are more coupled to business logic than the ambassador pattern assumes. That's not a failure — it just means the ambassador works well as a **partial** solution in your architecture, not a total one.
