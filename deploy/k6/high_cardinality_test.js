import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export const options = {
  stages: [
    { duration: '1m', target: 500 },
    { duration: '2m', target: 500 },
    { duration: '1m', target: 0 },
  ],
};

export default function () {
  const uniqueKey = `rl_highcard_${__VU}_${Date.now()}`;
  const projectId = `project_${__VU % 1000}`;

  const res = http.post(`${BASE_URL}/api/v1/gateway/${projectId}`, JSON.stringify({
    route: '/api/v1/test',
    method: 'GET',
  }), {
    headers: {
      'Content-Type': 'application/json',
      'X-API-Key': uniqueKey,
    },
  });

  check(res, { 'status ok': (r) => r.status < 500 });
  sleep(0.05);
}
