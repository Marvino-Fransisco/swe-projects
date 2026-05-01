# From Intuition to Architecture: The Self-Taught Engineer's Roadmap

## Overview

You already know how to build things. Your goal now is to learn the "mechanic's manual"—the formal names, edge cases, and theoretical limits of the systems you build.

Do not read these resources cover-to-cover like a novel. Use them to solve problems you've seen before. When you read about a concept, say to yourself: "Oh, that's what I was doing when X happened."

---

## Phase 1: The Pattern Dictionary (Plugging the Vocabulary Gaps)

*Goal: Learn the industry standard names for the architectural solutions you already intuitively use.*

**What to Learn:**

- [ ] Idempotency & Idempotency Keys
- [ ] Circuit Breaker Pattern
- [ ] Retry & Backoff strategies (Exponential backoff, Jitter)
- [ ] Cache Invalidation Strategies (Cache-Aside, Write-Through, Write-Behind)
- [ ] Polling vs Long-Polling vs WebSockets vs Server-Sent Events (SSE)
- [ ] Heartbeat / Dead-letter Queues

**Resources:**

- **Website:** [Microsoft Azure Cloud Design Patterns](https://learn.microsoft.com/en-us/azure/architecture/patterns/)
  - *Why:* This is the best catalog of architectural patterns on the internet. It's technology-agnostic. Read the patterns for: Circuit Breaker, Cache-Aside, Competing Consumers, Priority Queue.
- **Website:** [Martin Fowler - Enterprise Architecture Patterns](https://martinfowler.com/eaaCatalog/)
  - *Why:* The godfather of software architecture. Skim this for the vocabulary.

---

## Phase 2: Database Internals & Concurrency (Eliminating the Black Box)

*Goal: Understand how databases actually store data, handle concurrent users, and prevent data corruption.*

**What to Learn:**

- [ ] B-Trees vs Hash Indexes (Why your Q2 guess was right)
- [ ] ACID Properties (Atomicity, Consistency, Isolation, Durability)
- [ ] Transaction Isolation Levels (Read Committed, Repeatable Read, Serializable)
- [ ] Concurrency Control: Pessimistic Locking vs Optimistic Locking (The solution to the Ticket Double-Spend problem)
- [ ] MVCC (Multi-Version Concurrency Control) - How Postgres avoids locking readers out when writers are writing.

**Resources:**

- **Book:** *Designing Data-Intensive Applications* by Martin Kleppmann
  - *Why:* THE bible. Read Chapter 3 (Storage and Retrieval - B-Trees) and Chapter 7 (Transactions - Isolation levels & Locking).
- **Documentation:** [PostgreSQL Official Docs - Concurrency Control](https://www.postgresql.org/docs/current/mvcc.html)
  - *Why:* Don't read it all, but search for "Isolation Levels" and "Row Locking". Seeing how a real production DB implements these concepts makes them concrete.
- **Website:** [DB Fiddle](https://www.db-fiddle.com/)
  - *Action:* Open a Postgres fiddle. Create a table, start two separate sessions, and try to replicate a race condition. See how `SELECT ... FOR UPDATE` fixes it.

---

## Phase 3: Distributed Systems & Scale (The Senior Engineer Domain)

*Goal: Learn what breaks when your code runs on 10 servers instead of 1.*

**What to Learn:**

- [ ] CAP Theorem (Consistency vs Availability vs Partition Tolerance)
- [ ] Sharding Strategies (Hash-based vs Range-based / Data Skew)
- [ ] Distributed Tracing & Correlation IDs (The solution to the "Black Box" observability problem)
- [ ] Consistency Models: Strong vs Eventual Consistency
- [ ] Distributed Locks (Redis & Redlock algorithm)

**Resources:**

- **Book:** *Designing Data-Intensive Applications* by Martin Kleppmann
  - *Why:* Read Chapter 5 (Replication), Chapter 6 (Partitioning/Sharding), and Chapter 9 (Consistency and Consensus).
- **Website:** [The Fly.io Blog - Distributed Systems](https://fly.io/blog/)
  - *Why:* Extremely practical, real-world engineering posts about building distributed systems.
- **Website:** [Julia Evans' Zines (Wizard Zines](https://wizardzines.com/)
  - *Why:* Brilliant, visual, simple explanations of complex networking and distributed systems topics (TCP, Load Balancing, Observability).

---

## Phase 4: API Design & Networking Edge Cases

*Goal: Build APIs that are resilient to terrible network conditions and bad clients.*

**What to Learn:**

- [ ] HTTP Status Codes (Focus on the 4xx edge cases: 409 Conflict, 422 Unprocessable Entity, 429 Too Many Requests)
- [ ] Rate Limiting (Token Bucket vs Leaky Bucket algorithms)
- [ ] Pagination Deep Dive (Offset vs Cursor-based pagination - why offset breaks on deletes)

**Resources:**

- **Website:** [Stripe API Design](https://stripe.com/docs/api)
  - *Action:* Don't just read it—*study* it. Stripe has arguably the best-designed API in the world. Look at how they handle errors, pagination (cursor-based), and idempotency keys. Copy their patterns.
- **Article:** "How to Design Resilient APIs" (Search for articles on Idempotency keys and retry safety).

---

## Your Daily Action Plan

1. **Pick one concept** from the "What to Learn" lists above.
2. **Read the Wikipedia article or a blog post** about it until you understand the core idea.
3. **Write 50 lines of code** proving you understand it. (e.g., Write a small Go script that deliberately causes a race condition, then fix it with a Mutex. Or implement a simple Rate Limiter).
4. **Move to the next concept.**

Don't rush it. You have the hardest skill (problem-solving). This is just about collecting the map.
