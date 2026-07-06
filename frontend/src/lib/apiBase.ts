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

  // WebSocketはVercelのrewriteでプロキシできないため、
  // 環境変数が未設定でも同一オリジンにはせず、常にバックエンドへ直接接続する
  if (import.meta.env.PROD) {
    return 'wss://numduel-backend.onrender.com/ws';
  }

  return 'ws://localhost:8090/ws';
}