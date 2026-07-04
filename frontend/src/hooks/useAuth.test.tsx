import { act, renderHook, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { apiData, apiFetch, setOnUnauthorized } from '../api/client';
import { AuthProvider, useAuth } from './useAuth';

vi.mock('../api/client', async () => {
  const actual = await vi.importActual<typeof import('../api/client')>('../api/client');
  return {
    ...actual,
    apiData: vi.fn(),
    apiFetch: vi.fn(),
    setOnUnauthorized: vi.fn(),
  };
});

const mockedApiData = vi.mocked(apiData);
const mockedApiFetch = vi.mocked(apiFetch);
const mockedSetOnUnauthorized = vi.mocked(setOnUnauthorized);

function wrapper({ children }: { children: React.ReactNode }) {
  return <AuthProvider>{children}</AuthProvider>;
}

describe('useAuth', () => {
  beforeEach(() => {
    mockedApiData.mockReset();
    mockedApiFetch.mockReset();
    mockedSetOnUnauthorized.mockReset();
  });

  afterEach(() => {
    setOnUnauthorized(null);
  });

  it('loads user on mount', async () => {
    mockedApiData.mockResolvedValueOnce({ id: '1', username: 'alice', role: 'user' });
    const { result } = renderHook(() => useAuth(), { wrapper });
    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.user?.username).toBe('alice');
    expect(result.current.isAuthenticated).toBe(true);
  });

  it('clears user when refresh fails', async () => {
    mockedApiData.mockRejectedValueOnce(new Error('fail'));
    const { result } = renderHook(() => useAuth(), { wrapper });
    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.user).toBeNull();
  });

  it('registers unauthorized handler and login/logout work', async () => {
    mockedApiData.mockResolvedValueOnce(null as never);
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
