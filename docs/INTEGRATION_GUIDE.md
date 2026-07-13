# Integration Guide (How to consume Limiter.io)

This document shows how developers can integrate **Limiter.io** to protect their APIs. It details both using the built-in Go SDK and implementing custom HTTP requests in other languages (Node.js, Python, or cURL) without needing a dedicated SDK.

---

## 🐹 1. Using the Go SDK (Built-in)

The repository provides a pre-packaged Go client and Gin middleware inside the [sdk/](file:///c:/Users/R.K%20Singh/Desktop/rate-limiter/sdk/) directory.

### Express Go Middleware Integration:
```go
package main

import (
	"github.com/gin-gonic/gin"
	"limiter.io/sdk"
)

func main() {
	r := gin.Default()

	// 1. Initialize Limiter Client with API key
	limiterClient := sdk.NewClient("http://localhost:8888", "lim_live_your_api_key_here")

	// 2. Wrap routes with Gin middleware
	r.GET("/v1/billing/invoices", sdk.GinRateLimit(limiterClient), func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "invoices loaded"})
	})

	r.Run(":8081")
}
```

---

## ⚡ 2. Integrating Without an SDK (Raw HTTP API)

Since Limiter.io is built on standard HTTP REST architecture, developers in other languages can easily consume the gateway using local HTTP clients (e.g. `axios`, `requests`, `fetch`).

The gateway URL is: `http://localhost:8888/api/v1/gateway`

### A. Node.js & Express Custom Middleware

You can write a simple 15-line Express middleware to evaluate rate limits:

```javascript
const axios = require("axios");

const LIMITER_URL = "http://localhost:8888/api/v1/gateway";
const API_KEY = "lim_live_your_api_key_here";

const rateLimitMiddleware = async (req, res, next) => {
  try {
    // Intercept path and check rate limits
    const response = await axios.get(`${LIMITER_URL}${req.path}`, {
      headers: { "X-API-Key": API_KEY }
    });
    
    // Status 200 = request is within limits
    next();
  } catch (error) {
    if (error.response && error.response.status === 429) {
      return res.status(429).json({ error: "Too Many Requests. Rate limit exceeded." });
    }
    // Fail-open strategy: allow request if rate limiter backend is down
    console.warn("Rate limiter failed, allowing request:", error.message);
    next();
  }
};

// Application routes
app.get("/v1/payments", rateLimitMiddleware, (req, res) => {
  res.json({ status: "processed" });
});
```

### B. Python & Flask Decorator

In Python, write a custom Flask decorator to enforce rate limits:

```python
import requests
from functools import wraps
from flask import Flask, jsonify, request

app = Flask(__name__)

LIMITER_URL = "http://localhost:8888/api/v1/gateway"
API_KEY = "lim_live_your_api_key_here"

def rate_limit_required(f):
    @wraps(f)
    def decorated_function(*args, **kwargs):
        try:
            url = f"{LIMITER_URL}{request.path}"
            res = requests.get(url, headers={"X-API-Key": API_KEY}, timeout=2)
            if res.status_code == 429:
                return jsonify({"error": "Rate limit exceeded"}), 429
        except requests.RequestException as e:
            # Fail-open: allow request on network errors
            app.logger.warning(f"Rate limiter request failed: {e}")
            
        return f(*args, **kwargs)
    return decorated_function

@app.route("/v1/data")
@rate_limit_required
def get_data():
    return jsonify({"status": "data loaded"})
```

### C. Raw Shell Testing (cURL)

To manually verify if an API key is active or check remaining quota:

```bash
curl -i -X GET http://localhost:8888/api/v1/gateway/v1/data \
  -H "X-API-Key: lim_live_your_api_key_here"
```

Response headers returned by the server:
```http
HTTP/1.1 200 OK
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 99
X-RateLimit-Reset: 59
```
If quota is exhausted, you receive:
```http
HTTP/1.1 429 Too Many Requests
```
