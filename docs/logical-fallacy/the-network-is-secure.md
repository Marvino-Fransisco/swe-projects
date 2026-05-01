# The network is secure

- The Illusion: Because the application is behind a corporate firewall or lives inside a private cloud (like a VPN or VPC), internal traffic is safe. We don't need to encrypt or validate data internally.
- The Reality: "Zero Trust" is the reality. Threats come from compromised internal servers, malicious insiders, or misconfigurations.
- The Consequence: Sending sensitive data (like passwords or personal info) in plain text between internal microservices. Failing to sanitize inputs because developers assume "only our own frontend talks to this API."
