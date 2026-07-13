# Limiter.io — Node.js SDK

Official Node.js client and Express middleware for [Limiter.io](https://limiter.io) rate limiting gateway.

## Installation

```bash
npm install @limiter/sdk-node
```

## Quick Start

```javascript
const LimiterClient = require('@limiter/sdk-node');

const client = new LimiterClient('http://localhost:8080', 'lim_live_your_api_key');

// Simple boolean check
const allowed = await client.verify('/v1/users');
console.log('Allowed:', allowed);

// Full result with remaining quota
const result = await client.check('/v1/users');
console.log(`Allowed: ${result.allowed} | Remaining: ${result.remaining}/${result.limit} | Resets in: ${result.resetIn}s`);
```

## Express Middleware

```javascript
const LimiterClient = require('@limiter/sdk-node');
const expressRateLimit = require('@limiter/sdk-node/middleware');

const client = new LimiterClient('http://localhost:8080', 'lim_live_your_api_key');

app.get('/api/resource', expressRateLimit(client), handler);
```

The middleware automatically returns `429 Too Many Requests` when quota is exceeded and **fails open** if the gateway is unreachable.
