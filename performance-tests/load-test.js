import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 20 },
    { duration: '1m', target: 20 },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],
    http_req_failed: ['rate<0.01'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'https://staging-api.gosol.com';

export default function () {
  const responses = http.batch([
    ['GET', `${BASE_URL}/api/health`],
    ['GET', `${BASE_URL}/api/metrics`],
  ]);

  check(responses[0], {
    'health check status is 200': (r) => r.status === 200,
  });

  check(responses[1], {
    'metrics status is 200': (r) => r.status === 200,
  });

  sleep(1);
} 