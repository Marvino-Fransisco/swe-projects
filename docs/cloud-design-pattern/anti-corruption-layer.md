# Anti-Corruption Layer Pattern — Tradeoffs & Edge Cases

## The Full Tradeoffs of the Anti-Corruption Layer Pattern

### ✅ Benefits (the "why you'd want it")

**1. Preserves the integrity of your domain model**
Without an ACL, your new system is forced to absorb the legacy system's data models, protocols, and API semantics. Over time, this "corrupts" the design of the new system — it becomes a patchwork of old and new concepts. The ACL acts as a translation boundary, keeping your modern domain model clean and coherent.

**2. Enables incremental migration**
You don't have to rewrite everything at once. The ACL lets you migrate subsystem by subsystem, maintaining integration with legacy resources throughout. New features ship against the clean API while the ACL handles the dirty work of talking to old systems behind the scenes.

**3. Decouples teams and release cycles**
The team building the new system doesn't need to understand the legacy system's internals. They code against the ACL's interface. The team maintaining the ACL deals with the legacy complexity. These teams can release independently — the ACL absorbs changes on either side without forcing a coordinated deployment.

**4. Applicable beyond legacy migration**
The same problem exists with any external system your team doesn't control: third-party SaaS APIs, partner systems, shared infrastructure services. Any external dependency with different semantics is a corruption risk. The ACL pattern works for all of these.

---

### ❌ Tradeoffs / Problems

**1. Latency overhead**
Every call between subsystems now passes through an extra layer. For most applications this is negligible, but for latency-sensitive systems (real-time processing, high-frequency trading), the translation step adds measurable delay. The ACL must deserialize, transform, and reserialize every request.

**2. Another service to own and operate**
The ACL is a full-fledged service in your architecture. It needs deployment pipelines, monitoring, alerting, scaling policies, and on-call rotation. If it goes down, communication between subsystems stops entirely. You've traded coupling for an additional operational dependency.

**3. Scaling is your problem now**
The ACL sits in the critical path between subsystems. Under load, it must scale to handle the throughput of the busier side. If subsystem A generates 10x the traffic of subsystem B, the ACL must scale to match A — not B. This can be expensive if the translation logic is computationally heavy.

**4. One ACL might not be enough**
You may need multiple ACLs for different subsystem boundaries, or even decompose a single ACL into multiple services using different technologies. A translation between a REST API and a gRPC service is a different problem space than translating between an event stream and a SOAP endpoint. Each boundary might warrant its own ACL with its own tech stack.

**5. Transaction and data consistency become harder**
The ACL sits between two systems that may have fundamentally different consistency models. The legacy system might use pessimistic locking; the new system might use eventual consistency. The ACL must handle translation between these models, and if something fails mid-translation, you need compensating logic to avoid leaving both systems in an inconsistent state.

**6. Permanent or temporary? — a strategic decision**
If the ACL is part of a migration strategy, it's supposed to be temporary — retired once all legacy functionality is migrated. But temporary infrastructure has a way of becoming permanent. If you don't plan for the ACL's retirement from the start, it becomes just another layer of indirection that future developers must navigate.

**7. Not all communication needs an ACL**
If the semantic differences between subsystems are minor — say, slightly different field names or date formats — a full ACL is overkill. A simple adapter or mapper at the call site is sufficient. The ACL pattern pays for itself when the semantic gap is wide and ongoing.

---

### Summary Table

| Tradeoff | Severity | Notes |
| --- | --- | --- |
| Latency overhead | Low–Medium | Usually negligible; critical for real-time systems |
| Operational burden | Medium | Full service lifecycle to manage |
| Scaling responsibility | Medium | Must match the busier subsystem's throughput |
| Multiple ACLs needed | Medium | Depends on number of subsystem boundaries |
| Data consistency | High | Different consistency models are hard to reconcile |
| Temporary vs. permanent | Medium | Easy to accidentally make permanent |
| Overkill for small gaps | Low | Simple adapters suffice for minor differences |

---

## Edge Case: The ACL Becomes the Most Complex Service in Your System

### The Core Tension

The ACL is supposed to be a thin translation layer. But when the legacy system has convoluted data schemas, inconsistent APIs, and undocumented business rules baked into its responses, the ACL absorbs all of that complexity. Over time, it stops being a simple adapter and becomes the most complex service in your architecture — the very thing it was meant to prevent.

---

### Approach 1: Split the ACL Into Bounded-Context Translators

Instead of one monolithic ACL, create separate ACL components per bounded context. Each handles translation for one domain area (e.g., `BillingACL`, `InventoryACL`, `UserACL`).

```text
Subsystem A → BillingACL → Subsystem B
Subsystem A → InventoryACL → Subsystem B
Subsystem A → UserACL → Subsystem B
```

**Pros:** Each ACL stays focused and small. Teams can own individual ACLs independently.

**Cons:** More services to deploy. Shared translation logic (e.g., common data format mapping) may get duplicated across ACLs unless you introduce a shared library — which reintroduces coupling.

---

### Approach 2: Push Complexity Back Into the Legacy System

Instead of the ACL compensating for every legacy quirk, create a thin API layer on the legacy side that exposes a cleaner interface. The ACL then translates between two reasonably-structured APIs instead of translating between clean and chaos.

```text
New System → ACL → Legacy Facade → Legacy System
```

**Pros:** Reduces ACL complexity significantly. The legacy facade is simpler to write because it runs within the legacy system's own context.

**Cons:** Requires modifying the legacy system — which might be exactly what you're trying to avoid. Not always politically or technically feasible.

---

### Approach 3: Canonical Intermediate Model

Introduce a canonical data model that both sides translate to and from. Instead of the ACL knowing about both subsystems' models directly, each side has its own translator to/from the canonical model.

```text
Subsystem A → Translator A → Canonical Model ← Translator B ← Subsystem B
```

**Pros:** Adding a third subsystem only requires one new translator, not a new ACL. The canonical model acts as a contract.

**Cons:** Designing a good canonical model is hard. If it's too abstract, it can't express important domain concepts. If it's too specific, it biases toward one subsystem. The canonical model can become its own corruption risk if it tries to serve too many masters.

---

### Approach 4: Event-Driven Translation (Choreography Over Orchestration)

Instead of synchronous request-response translation, use an event-driven approach. The legacy system emits events. The ACL (or a set of event processors) consumes those events, transforms them, and publishes them in the new system's format.

```text
Legacy System → emits event → ACL (event processor) → publishes to New System's event bus
```

**Pros:** Decouples the systems temporally. The new system doesn't wait on the legacy system. Easier to scale and retry failed translations independently.

**Cons:** Eventual consistency by default. Harder to debug. Requires event infrastructure (message broker, event store). The translation logic is now distributed across event processors, which can be harder to reason about than a single synchronous ACL.

---

### The Real Question to Ask Before Building an ACL

1. **How wide is the semantic gap?**
   → Narrow gap (field renaming, format conversion): use a simple adapter.
   → Wide gap (different domain models, business rules, protocols): use an ACL.

2. **Is the legacy system modifiable?**
   → Yes: consider Approach 2 (push a facade onto the legacy side).
   → No: the ACL absorbs all complexity. Plan for it.

3. **How many subsystem boundaries exist?**
   → One boundary: a single ACL is fine.
   → Multiple boundaries: consider bounded-context ACLs (Approach 1) or a canonical model (Approach 3).

4. **Is synchronous or asynchronous communication appropriate?**
   → Synchronous: traditional ACL (façade/adapter).
   → Asynchronous: event-driven translation (Approach 4).

5. **Is this a migration or a long-term integration?**
   → Migration: plan the ACL's retirement from day one.
   → Long-term: invest in the ACL's reliability, monitoring, and team ownership as you would any critical service.

---

### The Uncomfortable Truth

The ACL pattern is often described as a way to "protect" your new system from legacy complexity. But the complexity doesn't disappear — it moves into the ACL. If you don't actively manage that complexity (by splitting, simplifying, or retiring the ACL), you've just built a new legacy system: one whose sole purpose is to talk to the old legacy system. The ACL is a useful boundary, but boundaries require maintenance.
