# Latency is zero

- The Illusion: A database query or an API request across the internet takes exactly 0 milliseconds.
- The Reality: Data cannot travel faster than the speed of light. Add in routing hops, processing time, and network congestion, and cross-continent calls can take hundreds of milliseconds.
- The Consequence: Developers build "chatty" applications that make 100 sequential API calls instead of 1 batched call. User interfaces freeze because the app performs synchronous tasks without loading states.
