# Choreography Pattern — Tradeoffs & Edge Cases

## The Full Tradeoffs of the Choreography Pattern

### ✅ Benefits (the "why you'd want it")

**1. No single point of failure**
There's no central orchestrator that can go down and take the entire workflow with it. Each service operates independently, subscribing to events and reacting on its own. If one service fails, the others keep processing. The message broker provides durability — messages survive service restarts and are processed when the service recovers.

**2. Services are truly loosely coupled**
Services don't call each other directly. They publish events and react to events. Service A doesn't need to know that Service B exists — it just publishes a `PackageCreated` event. Service B subscribes to that event and does its work. This means you can add, remove, or replace services without touching the others, as long as the event contracts remain stable.

**3. Independent deployment and scaling**
Each service can be deployed, scaled, and versioned independently. The drone scheduler can run 10 instances during peak hours while the package service runs 2. You can rewrite one service in a completely different language without affecting the rest. This autonomy is the core promise of microservices — and choreography delivers it.

**4. Natural fit for event-driven and serverless architectures**
Serverless functions and event-driven compute (Azure Functions, AWS Lambda, Azure Container Apps) are designed to react to events. Choreography aligns perfectly: a function fires when a message arrives, processes it, and publishes the result. No orchestrator to provision, no long-running process to keep alive.

**5. Easy to add new capabilities without modifying existing services**
Want to add a notification service that sends an email when a package ships? Subscribe it to the `PackageShipped` event. No existing service needs to change. The package service doesn't know or care that notifications are now being sent. This extensibility is dramatically cheaper than modifying an orchestrator's workflow definition.

**6. No orchestrator bottleneck**
Centralized orchestrators process requests sequentially or manage complex state machines. Under load, the orchestrator becomes a bottleneck — every request flows through it. Choreography distributes the processing across all participating services, naturally parallelizing work wherever possible.

---

### ❌ Tradeoffs / Problems

**1. The workflow is invisible — no one knows the whole picture**
In an orchestrator, you can look at one piece of code and understand the entire workflow: "Call A, then B, then C, handle errors." In choreography, the workflow is an emergent property of the system — it emerges from the interaction of independent services. There is no single place to see "what happens when a delivery request comes in." You must trace the event flow across multiple services, message buses, and subscriptions to understand it.

**2. Error handling is distributed and complex**
When Service B fails, who compensates for Service A's work? Who retries? Who decides the transaction is unrecoverable? In an orchestrator, these decisions are centralized. In choreography, every service must implement its own compensation logic, and the overall recovery strategy is distributed across the system. A saga pattern helps, but it's significantly harder to implement correctly than centralized error handling.

**3. Event ordering is not guaranteed**
Messages can arrive out of order, especially under retries, scale-out, or partition rebalancing. A `PackageCreated` event might arrive after a `PackageShipped` event. Each consumer must be designed for idempotency and must handle out-of-order events gracefully. Session-based ordering helps but adds complexity and limits parallelism.

**4. Debugging and tracing are significantly harder**
A single business transaction might span 5 services, 3 message queues, and 2 event grids. When something goes wrong, you need correlation IDs, distributed tracing, and centralized logging to reconstruct what happened. Without this observability infrastructure, debugging choreographed workflows is like debugging a distributed system with your eyes closed — because that's exactly what it is.

**5. Event schema evolution breaks consumers silently**
When a producer changes the structure of an event (adding a required field, renaming a property, changing a data type), downstream consumers that depend on the old schema break. Unlike an orchestrator where you can update all call sites in one deployment, choreography means consumers are independently deployed — a producer change might break a consumer the producer doesn't even know about. Schema registries help, but they're an additional layer of governance.

**6. Event storms and feedback loops**
When many services react to each other's events, the system can produce cascading event chains. Service A publishes an event, which triggers Service B, which publishes another event, which triggers Service C, which publishes an event that triggers Service A again. These feedback loops are hard to predict during design and hard to detect in production until the message broker is drowning in events.

**7. Sequential workflows are awkward**
Choreography shines for parallel, independent operations. But when Service D must wait for both Service B and Service C to complete, you need session identifiers, correlation logic, and wait conditions in Service D. This "fan-out, fan-in" pattern is natural in an orchestrator but requires careful coordination in choreography — often re-implementing orchestrator-like logic inside a consumer.

---

### Summary Table

| Tradeoff | Severity | Notes |
| --- | --- | --- |
| Invisible workflow | High | No single place to understand the full process |
| Distributed error handling | High | Compensation logic scattered across services |
| Event ordering | Medium–High | Requires idempotency and session management |
| Debugging complexity | High | Requires distributed tracing infrastructure |
| Schema evolution | Medium | Can silently break unknown consumers |
| Event storms | Medium | Feedback loops emerge at scale |
| Sequential workflows | Medium | Fan-in requires re-implementing coordination |

---

## Edge Case: Compensating a Partially Completed Workflow

### The Core Tension

A business transaction spans four services. Service A completes successfully, publishes an event. Service B completes successfully, publishes an event. Service C fails. Now the system is in an inconsistent state — A and B have committed work, C has not, and D never started. In an orchestrator, the central component detects C's failure and issues compensating transactions for A and B. In choreography, there's no central component. Each service must independently decide whether and how to compensate — and they must coordinate that compensation through events, the same mechanism that created the inconsistency.

---

### Approach 1: Choreographed Saga with Compensation Events

Each service publishes a failure event when its operation fails. Other services subscribe to failure events and execute compensating actions for their own completed work.

```text
Service A: completes → publishes SuccessA
Service B: completes → publishes SuccessB
Service C: fails → publishes FailureC
Service B: receives FailureC → compensates (undo B) → publishes CompensatedB
Service A: receives FailureC → compensates (undo A) → publishes CompensatedA
```

**Pros:** Stays true to the choreography model. No central coordinator. Each service owns its own compensation logic.

**Cons:** Compensation events can arrive out of order or be duplicated. What if Service A receives `FailureC` before it has finished processing `SuccessA`? What if `CompensatedB` triggers another service that already reacted to `SuccessB`? The compensation flow is itself a distributed workflow with its own failure modes. You've traded one hard problem (centralized error handling) for another (distributed error handling).

---

### Approach 2: Dead-Letter Queue with Centralized Remediation

Failed messages go to a dead-letter queue (DLQ). A dedicated remediation service consumes from the DLQ and decides how to compensate, retry, or pivot the transaction. This service has domain knowledge about the overall workflow.

```text
Service C: fails → sends to DLQ with failure reason
Remediation Service: reads DLQ → determines A, B completed → issues compensating commands to A and B
```

**Pros:** Compensation logic is centralized in one service — easier to reason about, test, and debug. The remediation service has visibility into the overall transaction state. DLQ is a standard messaging pattern with built-in support in most brokers.

**Cons:** The remediation service is effectively a lightweight orchestrator. You've reintroduced a central component with domain knowledge about the workflow — one of the things choreography was supposed to eliminate. If the remediation service goes down, failed transactions pile up in the DLQ indefinitely.

---

### Approach 3: Transaction State Table with Event-Driven Guards

Maintain a shared transaction state table (e.g., in a database) that tracks which steps have completed for each business transaction. Each service updates the table on success or failure and checks the table before processing to detect aborted transactions.

```text
Transaction 123: [A: completed, B: completed, C: failed, D: pending]
Service D: checks table → sees C failed → skips processing, publishes AbortD
Service A: periodic check → sees C failed → compensates, marks A: compensated
```

**Pros:** Provides a single source of truth for transaction state without centralizing processing logic. Any service can determine the transaction's health by reading the table. Works well with saga patterns.

**Cons:** The state table is shared infrastructure — a coupling point that all services depend on. It must be highly available and consistent. Concurrent updates require locking or optimistic concurrency, which adds complexity. The table can grow unbounded without cleanup policies.

---

### Approach 4: Hybrid — Choreography for Happy Path, Orchestrator for Failure Path

Use choreography for the normal flow (events, independent processing) but introduce a lightweight orchestrator that activates only when something fails. The orchestrator doesn't handle the happy path — it's exclusively a failure recovery coordinator.

```text
Happy path: Service A → event → Service B → event → Service C → event → Service D
Failure path: Service C fails → publishes FailureC → Orchestrator picks up → coordinates compensation
```

**Pros:** Best of both worlds. No orchestrator bottleneck during normal operation. Structured, centralized recovery when things go wrong.

**Cons:** You now maintain two coordination mechanisms. The orchestrator must understand enough about the workflow to compensate correctly, which means it has domain knowledge — and that knowledge must be kept in sync with the choreographed services as they evolve. The boundary between "happy path" and "failure path" isn't always clean; partial failures can blur it.

---

### The Real Question to Ask for Each Workflow

Before choosing a compensation strategy, for each business workflow ask:

1. **How many services participate in the transaction?**
   → 2–3 services: choreographed saga (Approach 1) is manageable.
   → 5+ services: consider DLQ remediation (Approach 2) or hybrid (Approach 4).

2. **Are the operations reversible (compensatable)?**
   → Fully reversible (create order → cancel order): any approach works.
   → Partially reversible (email sent, payment captured): the compensation logic is inherently complex regardless of approach. Invest in the hybrid model.

3. **How quickly must failures be resolved?**
   → Immediately (real-time user-facing): the compensation must be automatic and fast. Hybrid (Approach 4) or state table (Approach 3).
   → Eventually (within minutes/hours): DLQ remediation (Approach 2) with manual or automated recovery is sufficient.

4. **Can you tolerate a shared state dependency?**
   → Yes: transaction state table (Approach 3) provides the clearest visibility.
   → No: pure choreographed saga (Approach 1) keeps everything event-driven.

5. **Is the team prepared to operate distributed tracing?**
   → Yes: choreographed compensation (Approach 1) becomes debuggable.
   → No: start with DLQ remediation (Approach 2) — centralized failure handling is easier to operate without mature observability.

---

### The Uncomfortable Truth

Choreography eliminates the orchestrator as a bottleneck and a single point of failure. But it doesn't eliminate the need for coordination — it distributes it. Every service must know enough about the overall workflow to participate correctly, including how to fail correctly. The compensation problem reveals the real cost of decentralization: you're trading one complex component (the orchestrator) for N complex components (each service's compensation logic), plus the emergent behavior of their interactions. For simple, parallel, fire-and-forget workflows, choreography is elegant. For workflows with sequential dependencies, complex failure modes, and strong consistency requirements, you end up rebuilding orchestration inside your consumers — at which point you should ask whether a lightweight orchestrator would have been simpler from the start.
