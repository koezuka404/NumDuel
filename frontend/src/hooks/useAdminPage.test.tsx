import { act, renderHook, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { ToastProvider } from './useToast';
import { useAdminPage } from './useAdminPage';

const navigate = vi.fn();
const apiDataMock = vi.fn();
const apiFetchMock = vi.fn();
let authUser: { id: string; username: string; role: 'user' | 'master' } = {
  id: 'admin-1',
  username: 'admin',
  role: 'master',
};

vi.mock('../api/client', async () => {
  const actual = await vi.importActual<typeof import('../api/client')>('../api/client');
  return {
    ...actual,
    apiData: (...args: unknown[]) => apiDataMock(...args),
    apiFetch: (...args: unknown[]) => apiFetchMock(...args),
    downloadUrl: (path: string) => `/api${path}`,
  };
});

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom');
  return { ...actual, useNavigate: () => navigate };
});

vi.mock('./useAuth', () => ({
  useAuth: () => ({ user: authUser }),
}));

function wrapper({ children }: { children: React.ReactNode }) {
  return <ToastProvider>{children}</ToastProvider>;
}

describe('useAdminPage', () => {
  beforeEach(() => {
    authUser = { id: 'admin-1', username: 'admin', role: 'master' };
    navigate.mockReset();
    apiDataMock.mockReset();
    apiFetchMock.mockReset();
    apiDataMock.mockImplementation((path: string) => {
      if (path.startsWith('/admin/users/search')) {
        return Promise.resolve([
          { id: 'u2', username: 'bob', email: 'b@test.local', role: 'user', winCount: 0, deletedAt: null },
        ]);
      }
      if (path.startsWith('/admin/users')) {
        return Promise.resolve({
          items: [
            { id: 'u1', username: 'alice', email: 'a@test.local', role: 'user', winCount: 1, deletedAt: null },
          ],
        });
      }
      if (path.startsWith('/admin/logs/types')) {
        return Promise.resolve({ logTypes: ['guess'] });
      }
      if (path.startsWith('/admin/logs')) {
        return Promise.resolve({
          items: [{ id: 'l1', logType: 'guess', detail: 'd', createdAt: '2024-01-01T00:00:00Z' }],
        });
      }
      if (path.startsWith('/admin/backup/status')) {
        return Promise.resolve({ status: 'ok', lastSyncedAt: '2024-01-01T00:00:00Z' });
      }
      return Promise.reject(new Error(`unexpected ${path}`));
    });
    apiFetchMock.mockResolvedValue(undefined);
    vi.spyOn(window, 'confirm').mockReturnValue(true);
    vi.spyOn(window, 'open').mockImplementation(() => null);
  });

  it('loads users tab by default', async () => {
    const { result } = renderHook(() => useAdminPage(), { wrapper });
    await waitFor(() => expect(result.current.users).toHaveLength(1));
  });

  it('redirects non-master users', async () => {
    authUser = { id: '1', username: 'alice', role: 'user' };
    renderHook(() => useAdminPage(), { wrapper });
    await waitFor(() => expect(navigate).toHaveBeenCalledWith('/matching'));
  });

  it('switches tabs and performs admin actions', async () => {
    const { result } = renderHook(() => useAdminPage(), { wrapper });
    await waitFor(() => expect(result.current.users).toHaveLength(1));

    act(() => {
      result.current.setSearchQ('alice');
    });
    await act(async () => {
      await result.current.searchUsers();
    });
    expect(result.current.users[0]?.username).toBe('bob');

    act(() => {
      result.current.setTab('logs');
    });
    await waitFor(() => expect(result.current.logTypes).toEqual(['guess']));

    act(() => {
      result.current.setLogType('guess');
    });
    await act(async () => {
      await result.current.searchLogs();
    });

    act(() => {
      result.current.downloadLogs();
    });
    expect(window.open).toHaveBeenCalledWith('/api/admin/logs/download?logType=guess', '_blank');

    act(() => {
      result.current.setTab('ranking');
    });
    await act(async () => {
      await result.current.rebuildRanking();
    });

    act(() => {
      result.current.setTab('backup');
    });
    await waitFor(() => expect(result.current.backup?.status).toBe('ok'));

    act(() => {
      result.current.deleteUser('u1');
    });
    await waitFor(() => expect(result.current.users).toHaveLength(0));
  });

  it('cancels delete when confirm is false', async () => {
    vi.spyOn(window, 'confirm').mockReturnValue(false);
    const { result } = renderHook(() => useAdminPage(), { wrapper });
    await waitFor(() => expect(result.current.users).toHaveLength(1));
    act(() => {
      result.current.deleteUser('u1');
    });
    expect(result.current.users).toHaveLength(1);
  });
});
