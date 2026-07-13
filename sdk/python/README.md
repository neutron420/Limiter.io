# Limiter.io — Python SDK

Official Python client and Flask decorator for [Limiter.io](https://limiter.io) rate limiting gateway.

## Installation

```bash
pip install limiter-sdk
```

Or install from source:

```bash
cd sdk/python
pip install .
```

## Quick Start

```python
from client import LimiterClient

client = LimiterClient('http://localhost:8080', 'lim_live_your_api_key')

# Simple boolean check
allowed = client.verify('/v1/users')
print('Allowed:', allowed)

# Full result with remaining quota
result = client.check('/v1/users')
print(f'Allowed: {result.allowed} | Remaining: {result.remaining}/{result.limit} | Resets in: {result.reset_in}s')
```

## Flask Decorator

```python
from client import LimiterClient
from decorator import flask_rate_limit

client = LimiterClient('http://localhost:8080', 'lim_live_your_api_key')

@app.route('/api/resource')
@flask_rate_limit(client)
def my_endpoint():
    return jsonify({"status": "ok"})
```

The decorator automatically returns `429 Too Many Requests` when quota is exceeded and **fails open** if the gateway is unreachable.
