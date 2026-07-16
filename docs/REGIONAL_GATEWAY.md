# Regional Gateway Deployment

## Multi-Region Architecture

Each region has its own gateway instance connected to a local Redis replica:

```
us-east-1:
  gateway.limiter.io/us-east-1
  redis.us-east-1.example.com:6379

eu-west-1:
  gateway.limiter.io/eu-west-1
  redis.eu-west-1.example.com:6379

ap-southeast-1:
  gateway.limiter.io/ap-southeast-1
  redis.ap-southeast-1.example.com:6379
```

## DNS Setup

```
CNAME *.limiter.io → global.limiter.io
CNAME us-east-1.limiter.io → gateway-us-east-1.limiter.io
CNAME eu-west-1.limiter.io → gateway-eu-west-1.limiter.io
```

## Global Rate Limit Sync

Redis pub/sub channels:
- `limiter:global:sync` - cross-region rate limit state
- `limiter:global:blocks` - emergency blocks propagate globally

## Edge Rate Limiting

Cloudflare Worker at `workers/edge-rate-limiter.js`:
```javascript
export default {
  async fetch(request, env) {
    const key = `rl:${request.cf?.asn}:${request.url}`;
    const { success } = await env.LIMITER.limit({ key, limit: 100, window: 60 });
    if (!success) return new Response('Rate limit exceeded', { status: 429 });
    return fetch(request);
  }
};

export const config = { name: 'edge-rate-limiter' };
```

## Deployment

Each region: `kubectl apply -f deploy/k8s/region-us-east.yaml`
