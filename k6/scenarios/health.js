import http from 'k6/http';
import { check } from 'k6';

const apiBase = (__ENV.K6_API_BASE_URL || 'http://127.0.0.1:8090/api').replace(/\/+$/, '');
const origin = apiBase.replace(/\/api$/, '');
const vus = Number(__ENV.K6_VUS || 50);
const duration = __ENV.K6_DURATION || '2m';

export const options = {
  vus,
  duration,
  thresholds: {
    http_req_failed: ['rate<0.20'],
    http_req_duration: ['p(95)<3000'],
  },
};

export default function () {
  const res = http.get(`${origin}/health`);
  check(res, {
    'health status is 200': (r) => r.status === 200,
  });
}