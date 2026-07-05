import type { ApiDataResponse, ApiErrorBody, AuthUser } from '../types/dto';
import { resolveApiBaseURL } from '../lib/apiBase';
import { apiErrorMessage } from '../lib/labels';

const API_BASE_URL = resolveApiBaseURL();

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

type UnauthorizedHandler = () => void;

let onUnauthorized: UnauthorizedHandler | null = null;

export function setOnUnauthorized(handler: UnauthorizedHandler | null) {
  onUnauthorized = handler;
}

export function notifyUnauthorized() {
  onUnauthorized?.();
}

async function parseError(res: Response): Promise<ApiError> {
  const body = (await res.json().catch(() => ({}))) as ApiErrorBody;
  const code = body.error?.code ?? 'internal_error';
  const rawMessage = body.error?.message ?? '';
  const message = apiErrorMessage(code, rawMessage);
  return new ApiError(code, message, res.status);
}

export async function fetchSession(): Promise<AuthUser | null> {
  const res = await fetch(`${API_BASE_URL}/auth/session`, {
    ...defaultFetchOptions,
    method: 'GET',
  });
  if (res.status === 401 || res.status === 404) {
    return null;
  }
  if (!res.ok) {
    return null;
  }
  const json = (await res.json()) as ApiDataResponse<AuthUser | null>;
  return json.data ?? null;
}

export async function apiFetch<T>(path: string, options: RequestInit = {}, retried = false): Promise<T> {
  const res = await fetch(`${API_BASE_URL}${path}`, {
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
      await fetch(`${API_BASE_URL}/auth/refresh`, { method: 'POST', credentials: 'include' });
      return apiFetch<T>(path, options, true);
    }
  }

  if (res.status === 401) {
    notifyUnauthorized();
    throw new ApiError('unauthorized', apiErrorMessage('unauthorized', ''), 401);
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
  return `${API_BASE_URL}${path}`;
}
