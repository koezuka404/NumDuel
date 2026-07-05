import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import {  apiData, apiFetch, downloadUrl, fetchSession, notifyUnauthorized, setOnUnauthorized } from './client';

function mockFetch(response: Partial<Response> & { json?: () => Promise<unknown> }) {
  vi.stubGlobal(
    'fetch',
    vi.fn().mockResolvedValue({
      status: 200,
      ok: true,
      clone: () => ({ json: response.json ?? (async () => ({})) }),
      json: response.json ?? (async () => ({})),
      ...response,
    }),
  );
}

describe('apiFetch', () => {
  beforeEach(() => {
    setOnUnauthorized(null);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('calls notifyUnauthorized on 401 without page reload', async () => {
    const handler = vi.fn();
    setOnUnauthorized(handler);
    mockFetch({ status: 401, ok: false });

    await expect(apiFetch('/me')).rejects.toMatchObject({
      code: 'unauthorized',
      message: '認証に失敗しました',
      status: 401,
    });
    expect(handler).toHaveBeenCalledTimes(1);
  });

  it('retries once after token_expired on 404', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce({
        status: 404,
        ok: false,
        clone: () => ({
          json: async () => ({ error: { code: 'token_expired', message: 'expired' } }),
        }),
        json: async () => ({ error: { code: 'token_expired', message: 'expired' } }),
      })
      .mockResolvedValueOnce({
        status: 200,
        ok: true,
        json: async () => ({ data: { id: '1' } }),
      })
      .mockResolvedValueOnce({
        status: 200,
        ok: true,
        json: async () => ({ data: { id: '1' } }),
      });
    vi.stubGlobal('fetch', fetchMock);

    const data = await apiData<{ id: string }>('/me');
    expect(data.id).toBe('1');
    expect(fetchMock).toHaveBeenCalledWith('/api/auth/refresh', expect.objectContaining({ method: 'POST' }));
  });

  it('maps server validation errors to Japanese messages', async () => {
    mockFetch({
      status: 400,
      ok: false,
      json: async () => ({
        error: { code: 'validation_error', message: 'ユーザー名またはメールアドレスが既に使用されています' },
      }),
    });

    await expect(apiFetch('/auth/register', { method: 'POST' })).rejects.toMatchObject({
      code: 'validation_error',
      message: 'ユーザー名またはメールアドレスが既に使用されています',
    });
  });

  it('throws rate limit error in Japanese', async () => {
    mockFetch({ status: 429, ok: false, json: async () => ({}) });

    await expect(apiFetch('/matching/start', { method: 'POST' })).rejects.toMatchObject({
      code: 'rate_limit_exceeded',
      message: '操作が多すぎます。しばらく待ってください。',
    });
  });

  it('returns undefined for 204 responses', async () => {
    mockFetch({ status: 204, ok: true });
    await expect(apiFetch('/auth/logout', { method: 'POST' })).resolves.toBeUndefined();
  });

  it('unwraps apiData payload', async () => {
    mockFetch({ status: 200, ok: true, json: async () => ({ data: { ok: true } }) });
    await expect(apiData<{ ok: boolean }>('/me')).resolves.toEqual({ ok: true });
  });

  it('builds download urls', () => {
    expect(downloadUrl('/admin/logs/download')).toBe('/api/admin/logs/download');
  });

  it('falls back to internal_error for malformed error bodies', async () => {
    mockFetch({
      status: 500,
      ok: false,
      json: async () => {
        throw new Error('bad json');
      },
    });
    await expect(apiFetch('/broken')).rejects.toMatchObject({
      code: 'internal_error',
      message: 'サーバー内部エラーが発生しました',
    });
  });

  it('does not retry token refresh twice', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      status: 404,
      ok: false,
      clone: () => ({
        json: async () => ({ error: { code: 'token_expired', message: 'expired' } }),
      }),
      json: async () => ({ error: { code: 'token_expired', message: 'expired' } }),
    });
    vi.stubGlobal('fetch', fetchMock);
    await expect(apiFetch('/me', {}, true)).rejects.toMatchObject({ code: 'token_expired' });
  });

  it('notifyUnauthorized is safe without handler', () => {
    setOnUnauthorized(null);
    expect(() => notifyUnauthorized()).not.toThrow();
  });
});

describe('fetchSession', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('returns null on 401 without notifyUnauthorized', async () => {
    const handler = vi.fn();
    setOnUnauthorized(handler);
    mockFetch({ status: 401, ok: false });
    await expect(fetchSession()).resolves.toBeNull();
    expect(handler).not.toHaveBeenCalled();
  });

  it('returns user data on 200', async () => {
    mockFetch({
      status: 200,
      ok: true,
      json: async () => ({ data: { id: '1', username: 'alice', role: 'user' } }),
    });
    await expect(fetchSession()).resolves.toEqual({ id: '1', username: 'alice', role: 'user' });
  });
});
