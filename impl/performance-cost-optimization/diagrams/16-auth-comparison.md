# Web vs Mobile Auth Comparison

Side-by-side comparison of how authentication tokens are delivered and
validated in each BFF.

```mermaid
sequenceDiagram
    actor Web as Web Browser
    participant WebBFF as web-bff
    actor Mobile as Mobile App
    participant MobileBFF as mobile-bff

    Note over Web, WebBFF: Cookie-based Auth

    Web->>WebBFF: POST /auth/login
    WebBFF-->>Web: Set-Cookie: access_token (15m)<br/>Set-Cookie: refresh_token (7d)
    Web->>WebBFF: GET /cart (cookies sent automatically)
    WebBFF->>WebBFF: Read cookie
    WebBFF-->>Web: 200 Cart

    Web->>WebBFF: POST /auth/logout
    WebBFF-->>Web: Clear-Cookie: access_token<br/>Clear-Cookie: refresh_token

    Note over Mobile, MobileBFF: Bearer Token Auth

    Mobile->>MobileBFF: POST /auth/login
    MobileBFF-->>Mobile: { access_token, refresh_token } (JSON body)
    Mobile->>MobileBFF: GET /cart<br/>Authorization: Bearer &lt;token&gt;
    MobileBFF->>MobileBFF: Parse Authorization header
    MobileBFF-->>Mobile: 200 Cart

    Mobile->>MobileBFF: POST /auth/logout
    MobileBFF-->>Mobile: 200 (client discards tokens)
```

## Token Differences

| Aspect | web-bff | mobile-bff |
|---|---|---|
| Access token delivery | HTTP-only cookie (15 min) | JSON body (15 min) |
| Refresh token delivery | HTTP-only cookie (7 days) | JSON body (7 days) |
| Token sent on requests | Automatic via cookies | Manual `Authorization` header |
| Logout | Server clears cookies | Client discards tokens |
| Refresh flow | Cookie or body | Body required |
