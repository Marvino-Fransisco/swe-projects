# Topology doesn't change

- The Illusion: You assume that Server A will always have the IP address 10.0.0.5, and it will always be able to talk directly to Server B at 10.0.0.6. You assume the path data takes between two points is a fixed, straight line.

- The Harsh Reality: In modern cloud and containerized environments (like Kubernetes or Docker Swarm), the network is a living, breathing thing. Servers are destroyed and recreated constantly. Auto-scaling adds new nodes; crashes remove them. Network cables get cut, routers fail, and IP addresses are recycled dynamically.

- The Real-World Consequence:
  - Hardcoded IPs: A developer hardcodes an IP address into a config file. The server restarts, gets a new IP, and the entire system goes down.
  - No retry logic: If a network route temporarily changes (a blip), the app crashes instead of pausing and retrying the connection.
  - DNS caching issues: Assuming that a domain name always points to the same server, and caching that IP locally forever, only to find out the server behind it changed hours ago.
