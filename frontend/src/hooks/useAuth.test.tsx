import { act, renderHook, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { apiFetch, fetchSession, setOnUnauthorized } from '../api/client';
import { AuthProvider, useAuth } from './useAuth';

vi.mock('../api/client', async () => {
  const actual = await vi.importActual<typeof import('../api/client')>('../api/client');
  return {
    ...actual,
    fetchSession: vi.fn(),
    apiFetch: vi.fn(),
    setOnUnauthorized: vi.fn(),
  };
});

const mockedFetchSession = vi.mocked(fetchSession);
const mockedApiFetch = vi.mocked(apiFetch);
const mockedSetOnUnauthorized = vi.mocked(setOnUnauthorized);

function wrapper({ children }: { children: React.ReactNode }) {
  return <AuthProvider>{children}</AuthProvider>;
}

describe('useAuth', () => {
  beforeEach(() => {
    mockedFetchSession.mockReset();
    mockedApiFetch.mockReset();
    mockedSetOnUnauthorized.mockReset();
  });

  afterEach(() => {
    setOnUnauthorized(null);
  });

  it('loads user on mount', async () => {
    mockedFetchSession.mockResolvedValueOnce({ id: '1', username: 'alice', role: 'user' });
    const { result } = renderHook(() => useAuth(), { wrapper });
    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.user?.username).toBe('alice');
    expect(result.current.isAuthenticated).toBe(true);
    expect(mockedFetchSession).toHaveBeenCalledTimes(1);
  });

  it('clears user when session is empty', async () => {
    mockedFetchSession.mockResolvedValueOnce(null);
    const { result } = renderHook(() => useAuth(), { wrapper });
    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.user).toBeNull();
  });

  it('clears user when refresh fails', async () => {
    mockedFetchSession.mockRejectedValueOnce(new Error('fail'));
    const { result } = renderHook(() => useAuth(), { wrapper });
    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.user).toBeNull();
  });

  it('registers unauthorized handler and login/logout work', async () => {
    mockedFetchSession.mockResolvedValueOnce(null);
    const { result } = renderHook(() => useAuth(), { wrapper });
    await waitFor(() => expect(result.current.isLoading).toBe(false));

    expect(mockedSetOnUnauthorized).toHaveBeenCalled();
    const handler = mockedSetOnUnauthorized.mock.calls[0]?.[0];
    act(() => handler?.());

    mockedApiFetch.mockResolvedValueOnce({
      data: { id: '2', username: 'bob', role: 'master' },
    });
    await act(async () => {
      const user = await result.current.login('bob@test.local', 'password123');
      expect(user.username).toBe('bob');
    });

    mockedApiFetch.mockResolvedValueOnce(undefined as never);
    await act(async () => {
      await result.current.logout();
    });
    expect(result.current.user).toBeNull();
  });

  it('throws outside provider', () => {
    expect(() => renderHook(() => useAuth())).toThrow('useAuth must be used within AuthProvider');
  });
});
