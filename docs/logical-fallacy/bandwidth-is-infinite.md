# Bandwidth is infinite

- The Illusion: Developers often treat the network like a giant, bottomless pipe. They assume that if they need to send 500 megabytes of data from a database to an API, it will just happen instantly because "we have fast internet."

- The Harsh Reality: Every network link—whether it's a fiber optic cable, a AWS load balancer, or a cellular tower—has a maximum throughput limit. More importantly, bandwidth is often a shared resource.

- The Real-World Consequence:
  - Data Hoarding over the wire: Instead of querying a database for just the 5 columns needed for a user profile, a developer queries all 50 columns and sends them across the network.
  - Ignoring compression: Sending massive, uncompressed JSON or XML payloads instead of using Gzip or efficient formats like Protocol Buffers.
  - Cloud Bills: In cloud environments, bandwidth costs real money (egress fees). Moving massive amounts of data unnecessarily can accidentally double your AWS bill.
  - Bottlenecks: One service spamming the network with huge payloads can starve other critical services of bandwidth, causing cascading timeouts.
