# Limiter.io — Go SDK

Official Go client and Gin middleware for [Limiter.io](https://limiter.io) rate limiting gateway.

## Installation

```bash
go get limiter.io/sdk/go
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    sdk "limiter.io/sdk/go"
)

func main() {
    client := sdk.NewClient("http://localhost:8080", "lim_live_your_api_key")

    // Simple boolean check
    allowed, _ := client.Verify(context.Background(), "/v1/users")
    fmt.Println("Allowed:", allowed)

    // Full result with remaining quota
    result, _ := client.Check(context.Background(), "/v1/users")
    fmt.Printf("Allowed: %v | Remaining: %d/%d | Resets in: %s\n",
        result.Allowed, result.Remaining, result.Limit, result.Reset)
}
```

## Gin Middleware

```go
limiterClient := sdk.NewClient("http://localhost:8080", "lim_live_your_api_key")

r := gin.Default()
r.GET("/api/resource", sdk.GinRateLimit(limiterClient), handler)
```

The middleware automatically returns `429 Too Many Requests` when quota is exceeded and **fails open** if the gateway is unreachable.
