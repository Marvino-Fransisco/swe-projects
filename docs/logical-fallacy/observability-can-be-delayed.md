# Observability implementation can be delayed

- The Illusion: Let's build the features first, and we'll add logging, metrics, and tracing right before launch (or after things break).
- The Reality: You cannot debug a distributed system by just reading the code. Once in production, without pre-existing telemetry, you are entirely blind to how data is flowing.
- The Consequence: The classic "It works on my machine" problem. When a weird bug happens in production, teams spend days trying to reproduce it—wasting time that a simple log statement or trace ID would have solved in five minutes. Retrofitting observability into a complex system is incredibly difficult and expensive.
