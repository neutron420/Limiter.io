import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export const options = {
  stages: [
    { duration: '30s', target: 1000 },
    { duration: '1m', target: 1000 },
    { duration: '30s', target: 0 },
  ],
};

export default function () {
  const res = http.get(`${BASE_URL}/status`, {
    headers: { 'Content-Type': 'application/json' },
  });

  check(res, {
    'status is 200': (r) => r.status === 200,
  });

  sleep(0.01);
}

export function handleSummary(data) {
  return {
    'stdout': JSON.stringify({
      total_requests: data.metrics.http_reqs.values.count,
      avg_latency: data.metrics.http_req_duration.values.avg,
      p95_latency: data.metrics.http_req_duration.values['p(95)'],
    }),
  };
}
