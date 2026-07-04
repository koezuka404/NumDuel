import type { ApiDataResponse, ApiErrorBody } from '../types/dto';

export class ApiError extends Error {
  readonly code: string;
  readonly status: number;

  constructor(code: string, message: string, status: number) {
    super(message);
    this.code = code;
    this.status = status;
  }
}

const defaultFetchOptions: RequestInit = {
  credentials: 'include',
  headers: {
    'Content-Type': 'application/json',
  },
};

async function parseError(res: Response): Promise<ApiError> {
  const body = (await res.json().catch(() => ({}))) as ApiErrorBody;
  const code = body.error?.code ?? 'internal_error';
  const message = body.error?.message ?? 'request failed';
  return new ApiError(code, message, res.status);
}

export async function apiFetch<T>(path: string, options: RequestInit = {}, retried = false): Promise<T> {
  const res = await fetch(`/api${path}`, {
    ...defaultFetchOptions,
    ...options,
    headers: {
      ...defaultFetchOptions.headers,
      ...options.headers,
    },
  });

  if (res.status === 404) {
    const body = (await res.clone().json().catch(() => ({}))) as ApiErrorBody;
    if (body.error?.code === 'token_expired' && !retried) {
      await fetch('/api/auth/refresh', { method: 'POST', credentials: 'include' });
      return apiFetch<T>(path, options, true);
    }
  }

  if (res.status === 401) {
    window.location.href = '/login';
    throw new ApiError('unauthorized', 'unauthorized', 401);
  }

  if (res.status === 429) {
    throw new ApiError('rate_limit_exceeded', '操作が多すぎます。しばらく待ってください。', 429);
  }

  if (!res.ok) {
    throw await parseError(res);
  }

  if (res.status === 204) {
    return undefined as T;
  }

  return (await res.json()) as T;
}

export async function apiData<T>(path: string, options: RequestInit = {}): Promise<T> {
  const json = await apiFetch<ApiDataResponse<T>>(path, options);
  return json.data;
}

export function downloadUrl(path: string): string {
  return `/api${path}`;
}
