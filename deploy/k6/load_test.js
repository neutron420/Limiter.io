import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const API_KEY = __ENV.API_KEY || 'rl_test_key';

const blockRate = new Rate('blocked_requests');
const latencyTrend = new Trend('request_latency');

export const options = {
  stages: [
    { duration: '30s', target: 50 },
    { duration: '1m', target: 100 },
    { duration: '30s', target: 200 },
    { duration: '1m', target: 200 },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'],
    blocked_requests: ['rate<0.1'],
  },
};

export default function () {
  const projectId = __ENV.PROJECT_ID || 'test-project';

  const payload = JSON.stringify({
    route: '/api/v1/test',
    method: 'GET',
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'X-API-Key': API_KEY,
    },
  };

  const res = http.post(`${BASE_URL}/api/v1/gateway/${projectId}`, payload, params);

  check(res, {
    'status is 200 or 429': (r) => r.status === 200 || r.status === 429,
    'response time < 1s': (r) => r.timings.duration < 1000,
  });

  blockRate.add(res.status === 429);
  latencyTrend.add(res.timings.duration);

  sleep(Math.random() * 0.1);
}

export function handleSummary(data) {
  return {
    'stdout': JSON.stringify({
      total_requests: data.metrics.http_reqs.values.count,
      blocked_rate: data.metrics.blocked_requests.values.rate,
      p95_latency: data.metrics.http_req_duration.values['p(95)'],
      p99_latency: data.metrics.http_req_duration.values['p(99)'],
    }),
  };
}
