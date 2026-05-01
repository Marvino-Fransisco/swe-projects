# Backends for Frontends Pattern — Tradeoffs & Edge Cases

## The Full Tradeoffs of the Backends for Frontends Pattern

### ✅ Benefits (the "why you'd want it")

**1. Tailored backend per client interface**
Each frontend gets a backend optimized for its specific needs. A mobile BFF returns lightweight, bandwidth-efficient responses with pagination. A desktop BFF returns rich, aggregated data for an immersive experience. No single backend is forced into awkward compromises to serve all clients.

**2. Decoupled frontend and backend teams**
Frontend teams independently own their BFF service — choosing language, release cadence, workload prioritization, and feature integration. They don't need to coordinate with a centralized backend team or wait for consensus across other frontend teams. This autonomy accelerates development.

**3. Isolated failures per client**
When one BFF fails, only its associated frontend is affected. A malfunction in the mobile BFF doesn't take down the desktop experience. This containment model makes the overall system more resilient than a single shared backend where a failure impacts everyone.

**4. Independent scaling and optimization**
Each BFF can be scaled and optimized independently. The mobile BFF might run on lightweight serverless functions with aggressive caching. The desktop BFF might run on beefier compute with complex aggregation logic. You're not forced to scale a monolithic backend to serve the most demanding client.

**5. Security surface area reduction**
Each BFF exposes only the endpoints its client needs. A mobile app never sees desktop-only admin endpoints, and vice versa. This segmentation reduces the API surface area and limits lateral movement between backends — a meaningful security improvement over a shared backend.

**6. Easier migration and evolution**
When you need to replace or refactor one client's backend, you do it without touching the others. The mobile team can rewrite their BFF in a completely different language while the desktop BFF stays untouched. This is dramatically safer than refactoring a shared backend that everyone depends on.

---

### ❌ Tradeoffs / Problems

**1. Operational overhead — more services to manage**
Each BFF is a full service with its own lifecycle: deployment pipelines, monitoring, alerting, scaling policies, and on-call rotation. Two frontends means two BFFs to operate. Five frontends means five. This multiplies operational cost and complexity, especially for smaller teams.

**2. Code duplication across BFFs**
Multiple BFFs often need similar logic: authentication validation, data transformation, error handling, and microservice orchestration. This code gets duplicated across BFFs unless you extract it into shared libraries — which reintroduces coupling between teams.

**3. Latency overhead from extra network hop**
Clients no longer call backend services directly. Every request passes through the BFF, adding a network hop. For most applications this is negligible, but for latency-sensitive systems, the extra hop is measurable. The BFF must deserialize, transform, and reserialize every request.

**4. When to consolidate vs. split is a hard judgment call**
Different interfaces often make similar requests. At what point do you merge two BFFs into one? Sharing a single BFF between interfaces introduces conflicting requirements that complicate growth. Splitting too aggressively creates unnecessary duplication. There's no formula — it's a judgment call that must be revisited as requirements evolve.

**5. BFF can become a dumping ground**
Without discipline, the BFF accumulates business logic that belongs in the underlying microservices. What starts as "just reshaping the response for mobile" becomes "mobile-specific discount calculation logic." The BFF becomes a thick, hard-to-test layer that duplicates business rules.

**6. GraphQL may eliminate the need entirely**
If your organization uses GraphQL with frontend-specific resolvers, clients can request exactly the data they need through a single API. GraphQL's querying mechanism makes BFF services redundant in many cases. Building BFFs on top of GraphQL adds a layer that provides no additional value.

**7. An API gateway + microservices architecture might suffice**
If you already have an API gateway handling routing, rate limiting, and authentication, and your microservices are well-designed, the BFF layer may be unnecessary overhead. The gateway can handle cross-cutting concerns while clients call microservices directly for client-specific needs.

---

### Summary Table

| Tradeoff | Severity | Notes |
| --- | --- | --- |
| Operational overhead | Medium–High | Multiplies with each new frontend |
| Code duplication | Medium | Shared libraries reintroduce coupling |
| Latency overhead | Low–Medium | Extra network hop per request |
| Consolidation vs. split decision | Medium | No formula; must revisit over time |
| BFF becomes business logic dump | High | Requires active discipline to prevent |
| GraphQL makes BFF redundant | Medium | Evaluate before building BFFs |
| API gateway may be sufficient | Medium | Assess existing architecture first |

---

## Edge Case: Shared Microservice Calls Create Tight Coupling Between BFFs

### The Core Tension

Each BFF is supposed to be independent, but they often call the same underlying microservices. When one BFF team requests a change to a shared microservice's API (e.g., adding a new field, changing a response format), it can break the other BFF's integration. The BFFs are decoupled from each other but coupled through the shared services they depend on — reintroducing the coordination problem the pattern was meant to solve.

---

### Approach 1: Versioned APIs on Shared Microservices

Each microservice exposes versioned APIs. When BFF A needs a change, the microservice team creates a new version (`/v2/orders`) while keeping the old version (`/v1/orders`) intact for BFF B.

```text
Mobile BFF  → /v1/orders → Orders Microservice
Desktop BFF → /v2/orders → Orders Microservice
```

**Pros:** No coordination needed between BFF teams. Each BFF migrates at its own pace.

**Cons:** API versioning is its own operational burden. Maintaining multiple versions increases the microservice's code complexity and testing surface. Old versions must eventually be retired, which requires tracking and coordinating with the BFFs still using them.

---

### Approach 2: BFF-Specific Facade on Shared Microservices

Instead of versioning the entire API, the microservice exposes dedicated facade endpoints per BFF. Each facade returns data shaped for its specific consumer.

```text
Mobile BFF  → /orders/mobile-summary → Orders Microservice
Desktop BFF → /orders/desktop-detail  → Orders Microservice
```

**Pros:** Each endpoint is optimized for its consumer. No risk of one BFF breaking another.

**Cons:** The microservice now has BFF-specific knowledge, which inverts the dependency. The microservice team must understand and maintain endpoints shaped for specific frontends — which is exactly what the BFF was supposed to handle. This defeats the purpose.

---

### Approach 3: Contract Testing Between BFFs and Microservices

BFF teams write contract tests (e.g., using Pact) that define the expected API shape from each microservice. The microservice team runs these contracts in their CI pipeline. If a change breaks a BFF's contract, the build fails before it deploys.

```text
BFF A writes contract → Orders Microservice CI verifies contract
BFF B writes contract → Orders Microservice CI verifies contract
```

**Pros:** No API versioning complexity. Changes are safe by default. BFF teams don't coordinate with each other — they just maintain their own contracts.

**Cons:** Contract testing infrastructure adds CI complexity. Contracts must be kept up-to-date — stale contracts give false confidence. Doesn't prevent breaking changes, only detects them earlier.

---

### Approach 4: BFF Teams Own Their Own Microservice Slices

Instead of sharing microservices, each BFF team owns a vertical slice of the backend — including the microservices it calls. The mobile BFF team owns the mobile-oriented order service; the desktop BFF team owns the desktop-oriented order service. They share a database but have separate service code.

```text
Mobile BFF → Mobile Order Service → Shared Database
Desktop BFF → Desktop Order Service → Shared Database
```

**Pros:** Complete autonomy. No shared service dependency. Each team can optimize their slice independently.

**Cons:** Massive duplication. Same business logic implemented twice (or more). Database becomes the coupling point — schema changes still require coordination. Only viable for large organizations with enough engineers to sustain the duplication.

---

### The Real Question to Ask for Each Shared Microservice

Before picking an approach, for each microservice that multiple BFFs depend on, ask:

1. **How often does its API change?**
   → Rarely stable: simple coordination or contract testing suffices.
   → Frequently changing: versioned APIs or dedicated slices are necessary.

2. **How many BFFs depend on it?**
   → 2 BFFs: informal coordination might work.
   → 5+ BFFs: you need automated contract testing or versioned APIs.

3. **Is the microservice team resourced to support multiple consumers?**
   → Yes: versioned APIs (Approach 1) are straightforward.
   → No: push BFF-specific logic into the BFF itself, keeping the microservice generic.

4. **Can the database tolerate being the coupling point?**
   → If schema changes are rare: vertical slices (Approach 4) can work.
   → If schema changes frequently: avoid duplicated services sharing a database.

5. **Is the BFF truly independent, or is it a thin proxy?**
   → Thin proxy: reconsider whether you need BFFs at all. An API gateway might suffice.
   → Rich client-specific logic: BFFs are justified. Invest in contract testing.

---

### The Uncomfortable Truth

The Backends for Frontends pattern trades one coupling problem for another. You decouple the frontends from each other, but you couple them through shared backend services. The BFF is not a silver bullet — it's a boundary that shifts coordination costs from the frontend layer to the backend integration layer. If your underlying microservices are stable and well-designed, BFFs add real value. If they're not, BFFs become a fragile orchestration layer that breaks every time a shared service changes. The pattern works best when you accept that some coordination is inevitable and invest in making that coordination cheap (contract tests, versioned APIs) rather than pretending it doesn't exist.
