export function resolveApiBaseURL(): string {
  const raw = import.meta.env.VITE_API_BASE_URL;

  if (typeof raw === 'string' && raw.trim() !== '') {
    const base = raw.trim().replace(/\/+$/, '');
    return base.endsWith('/api') ? base : `${base}/api`;
  }

  return '/api';
}

export function resolveWsBaseURL(): string {
  const raw = import.meta.env.VITE_WS_BASE_URL;

  if (typeof raw === 'string' && raw.trim() !== '') {
    const base = raw.trim().replace(/\/+$/, '');
    return base.endsWith('/ws') ? base : `${base}/ws`;
  }

  if (typeof window !== 'undefined' && window.location) {
    const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    return `${proto}//${window.location.host}/ws`;
  }

  return 'ws://localhost:8090/ws';
}