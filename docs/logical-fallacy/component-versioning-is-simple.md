# Component versioning is simple

- The Illusion: You assume that if you update Microservice A to version 2.0, you can just flip a switch and every other microservice in your system will instantly be running version 2.0 as well. You assume everything is always on the latest version.

- The Harsh Reality: In distributed systems, different versions of the same software will exist at the same time. Deployments take time (rolling updates). Mobile apps (like iOS/Android) rely on users actually clicking "update." Breaking changes are inevitable, and coordinating updates across dozens of independent services is a nightmare.

- The Real-World Consequence:
  - Breaking changes: You remove a field from an API response because "nobody uses it anymore." But a mobile app from two years ago still relies on that field, and suddenly millions of older app versions crash.
  - Shared library hell: Two microservices rely on the same internal library, but Service A updates to the new version while Service B stays on the old one, causing weird, hard-to-trace data corruption.
  - The need for backward compatibility: This fallacy forces teams to implement complex patterns like "Feature Flags" and strict API versioning (e.g., /api/v1/ vs /api/v2/) to ensure old and new components can talk to each other safely during transitions.
