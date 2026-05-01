# CQRS Pattern — Tradeoffs & Edge Cases

## The Full Tradeoffs of the CQRS Pattern

### ✅ Benefits (the "why you'd want it")

**1. Read and write models scale independently**
Most systems have asymmetric read-to-write ratios (often 10:1 or higher). CQRS lets you scale the read side horizontally — add replicas, use denormalized views, cache aggressively — without touching the write side. The write side stays small, focused on consistency and business logic. You're not paying for write-level consistency on reads or read-level throughput on writes.

**2. Each model uses the right schema for its job**
The write model uses a normalized schema optimized for transactional integrity, constraints, and business rules. The read model uses denormalized materialized views optimized for query speed — precomputed joins, flat DTOs, aggregated data. No single schema has to compromise between write consistency and read performance.

**3. Separation of concerns produces cleaner code**
Write-side code handles validation, domain logic, invariants, and event publishing. Read-side code returns DTOs — no business logic, no validation, no side effects. This separation means the write model can grow complex (as domain models tend to do) without making queries harder to write or understand. Each side is simpler in isolation than a combined CRUD model would be.

**4. Security is more granular**
You can restrict write access to specific roles, services, or endpoints while broadly allowing read access. The read model never exposes the internal structure used by the write model. Sensitive fields (internal IDs, audit columns, encryption metadata) stay in the write model; the read model only projects what consumers need to see.

**5. Enables task-based UIs instead of CRUD screens**
Instead of a generic "Edit Order" form that maps directly to database columns, commands capture business intent: `CancelOrder`, `ShipOrder`, `ApplyDiscount`. This forces the UI to model real operations, which produces better user experiences and makes the system easier to reason about. The write model validates business rules; the read model shows the current state.

**6. Works naturally with event sourcing**
When combined with event sourcing, the write model stores domain events (the single source of truth), and the read model builds materialized views by projecting those events. You can replay events to rebuild or add new read models at any time. This makes the system auditable (every state change is recorded) and flexible (new views without schema migrations).

---

### ❌ Tradeoffs / Problems

**1. Significant increase in system complexity**
CQRS replaces one model with two (or more). You now have: separate codebases for read and write, separate data stores potentially, synchronization logic between them, event publishing and handling infrastructure, and materialized view management. What was a single CRUD service is now a distributed system. This complexity is the primary reason CQRS is recommended only for complex domains — in simple domains, it's overkill.

**2. Eventual consistency between read and write models**
When a write succeeds, the read model doesn't immediately reflect the change. There's a synchronization delay — the event must be published, consumed, and projected into the read store. During this window, a user who just submitted an update might query and see the old state. This isn't a bug; it's inherent to the pattern. But it requires UI design that accounts for it (optimistic updates, "your changes are being processed" messages, or accepting stale reads).

**3. Data synchronization is a hard distributed systems problem**
Keeping the read model in sync with the write model requires reliable event delivery. If events are lost, the read model diverges permanently. If events are duplicated, the read model must handle idempotency. If events arrive out of order, projections may produce incorrect state. You can't wrap "update write DB + publish event" in a single atomic transaction — the transactional outbox pattern helps, but adds more infrastructure.

**4. ORMs and scaffolding tools don't help you**
Traditional CRUD development leans heavily on ORMs (Entity Framework, Hibernate) and scaffolding tools that generate code from database schemas. CQRS requires custom code for both sides: command handlers, event publishers, projection logic, DTO mapping. You're writing more code with less tooling support. The "write model → event → read model" pipeline is entirely custom.

**5. Debugging requires tracing across multiple systems**
A single user action now flows through: command handler → domain model → event store → message broker → event handler → read model projection → read store. When a bug occurs (user reports stale data, missing data, or incorrect data), you must trace this entire chain to find where the breakdown happened. Without robust distributed tracing and correlation IDs, debugging CQRS is significantly harder than debugging CRUD.

**6. Event schema evolution is a long-term maintenance burden**
Events are stored permanently (especially with event sourcing). When the domain model changes — a field is added, renamed, or a new event type is introduced — old events must still be processable by current projections. You need versioning, upcasting, or transformation logic to handle schema evolution. This is a ongoing cost that grows with the age and complexity of the system.

**7. Not suitable for simple domains**
If your domain has straightforward CRUD operations, minimal business logic, and balanced read/write loads, CQRS adds complexity without proportional benefit. A blog engine, a simple admin panel, or a basic inventory tracker doesn't need separate read and write models. The pattern pays for itself only when the domain complexity, read/write asymmetry, or scaling requirements justify the added infrastructure.

---

### Summary Table

| Tradeoff | Severity | Notes |
| --- | --- | --- |
| System complexity | High | Two models, synchronization, event infrastructure |
| Eventual consistency | High | Read model lags behind write model |
| Data synchronization | High | Event delivery must be reliable and ordered |
| Loss of ORM/tooling support | Medium | Custom code for both sides |
| Debugging complexity | High | Must trace across multiple systems |
| Event schema evolution | Medium | Old events must remain processable |
| Overkill for simple domains | Low–Medium | Only justified for complex domains |

---

## Edge Case: The User Acts on Stale Read Model Data

### The Core Tension

A user views an order on the read model. The order shows status "Processing." Meanwhile, another user (or a background process) has already cancelled the order — the write model has processed the `CancelOrder` command and published the event, but the read model projection hasn't caught up yet. The first user clicks "Ship Order" based on the stale "Processing" status. The command handler receives a `ShipOrder` command for an order that's already cancelled. The system must decide: reject the command (frustrating the user who acted on information the system gave them), silently accept it (creating inconsistency), or handle the conflict gracefully.

---

### Approach 1: Optimistic Concurrency with Version Checking

Every read model projection includes a version number (or an expected event sequence number). When the user submits a command, the command includes the version of the data they were looking at. The command handler compares this version against the current version in the write model. If they differ, the command is rejected with a concurrency conflict.

```text
User reads order → status: Processing, version: 5
Other user cancels → write model version: 6, event published
First user submits ShipOrder(version: 5)
Command handler: current version is 6 ≠ 5 → reject with concurrency conflict
UI shows: "This order has been modified. Please refresh and try again."
```

**Pros:** Prevents incorrect commands from executing. The write model remains the source of truth. Users are informed that the data they acted on has changed.

**Cons:** The user experience is poor — they took an action based on data the system showed them, and the system rejected it. Under high contention (many users acting on the same entity), conflicts become frequent and frustrating. The UI must handle concurrency conflicts gracefully, which adds complexity to every form that submits commands.

---

### Approach 2: Accept the Command, Apply Business Rules, Compensate if Needed

The command handler accepts the `ShipOrder` command and applies business rules to determine if shipping is still valid given the current state. If the order has been cancelled, the command handler rejects it with a domain-specific error (not a generic concurrency conflict). If the order state has changed but shipping is still valid (e.g., status changed from "Processing" to "Packing"), the command succeeds.

```text
User submits ShipOrder for order 123
Command handler: loads current state from write model → status: Cancelled
Business rule: cannot ship a cancelled order → reject with "Order has been cancelled"
User submits ShipOrder for order 456
Command handler: loads current state → status: Packing (changed from Processing)
Business rule: can ship from Packing status → accept → order shipped
```

**Pros:** Domain logic drives the decision, not a generic version mismatch. Some state changes are harmless and the command can proceed, avoiding unnecessary rejection. The error message is meaningful to the user.

**Cons:** The command handler must load the current write-model state for every command, which adds latency. The business rule matrix ("which commands are valid in which states") must be comprehensive and tested. Some state transitions create ambiguity — if the order was "Processing" when the user clicked but is now "Packed," should the ship proceed? The answer depends on business context, not technical rules.

---

### Approach 3: Optimistic UI Updates with Background Reconciliation

When the user submits a command, the UI optimistically updates the displayed state immediately (before the server responds). If the command succeeds, the UI stays updated. If it fails, the UI rolls back and shows an error. The read model update is treated as a separate concern — it'll catch up eventually.

```text
User clicks Ship Order → UI immediately shows "Shipped" (optimistic)
Command sent to server:
  Success → UI stays "Shipped", read model eventually catches up
  Failure → UI rolls back to actual state, shows error message
```

**Pros:** Users perceive zero latency between their action and the UI response. The eventual consistency of the read model is hidden from the user. Works well for actions that are likely to succeed.

**Cons:** If the command fails, the user sees a jarring rollback — "it said Shipped but now it says Cancelled." The optimistic state must be kept locally (client-side state management) and reconciled with server responses. If the user navigates away before the server responds, the optimistic update may be lost or conflicting when they return. This approach is complex to implement correctly in multi-tab or multi-device scenarios.

---

### Approach 4: Notify Users of Pending Changes in Real Time

When the write model processes a command that changes an entity, it pushes a notification (via WebSocket, SSE, or signal) to any user currently viewing that entity. The UI updates in real time, reducing the window during which a user could act on stale data.

```text
User A views order 123 → status: Processing
User B cancels order 123 → write model publishes event → push notification to User A
User A's UI updates: status: Cancelled → "This order has been cancelled"
User A no longer sees "Ship Order" button → stale action prevented
```

**Pros:** Eliminates the stale-data problem for active users. The read model catches up asynchronously, but the UI is already correct. Creates a real-time, collaborative feel.

**Cons:** Requires real-time push infrastructure (WebSocket connections, connection management, authentication). Users who are offline or have disconnected won't receive the notification — the stale-data problem still exists for them. The push notification itself is eventually consistent with the write model; there's still a small window where the notification hasn't arrived but the data has changed. This is an additional system to build, operate, and debug.

---

### The Real Question to Ask for Each Command

Before choosing a staleness-handling strategy, for each command in your system ask:

1. **What's the cost of acting on stale data?**
   → Low (cosmetic change, easily reversible): accept the command with current-state validation (Approach 2).
   → High (financial transaction, irreversible action): enforce optimistic concurrency (Approach 1).

2. **How likely is contention on the same entity?**
   → Rare (single-user workflows): eventual consistency is fine. Minimal mitigation needed.
   → Frequent (collaborative editing, shared resources): real-time notifications (Approach 4) or optimistic UI (Approach 3).

3. **Can the business rule tolerate the state change?**
   → Yes (the command is valid in multiple states): accept with business-rule validation (Approach 2).
   → No (the command is only valid in one specific state): optimistic concurrency (Approach 1).

4. **What UX is acceptable when a conflict occurs?**
   → "Refresh and retry" is acceptable: optimistic concurrency (Approach 1).
   → "Never show a confusing state": optimistic UI or real-time push (Approaches 3 or 4).

5. **How fast is the read model synchronization?**
   → Sub-second: most users won't notice staleness. Minimal mitigation.
   → Seconds to minutes: you need active mitigation for high-value commands.

---

### The Uncomfortable Truth

CQRS introduces eventual consistency as a first-class architectural decision, but users don't think in terms of eventual consistency. They see a screen, they take an action, they expect the system to honor it. The gap between the write model's truth and the read model's projection is a gap the user experiences as "I did something and the system ignored it" or "the system showed me one thing and then contradicted itself." Every CQRS system must grapple with this gap — and the solutions (concurrency checks, optimistic updates, real-time push) are layers of complexity added on top of the already complex CQRS architecture. If your domain doesn't have a genuine read/write asymmetry problem, the eventual-consistency tax you pay for CQRS is a cost you bear without a corresponding benefit. CQRS is a powerful pattern for complex, high-scale domains — but it's an expensive one, and the expense is paid not just in infrastructure but in the user experience complexity that eventual consistency creates.
