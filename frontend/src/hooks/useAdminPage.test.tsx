import { act, renderHook, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { ToastProvider } from './useToast';
import { useAdminPage } from './useAdminPage';

const navigate = vi.fn();
const apiDataMock = vi.fn();
const apiFetchMock = vi.fn();

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
  useAuth: () => ({ user: { id: 'admin-1', username: 'admin', role: 'master' } }),
}));

function wrapper({ children }: { children: React.ReactNode }) {
  return <ToastProvider>{children}</ToastProvider>;
}

describe('useAdminPage', () => {
  beforeEach(() => {
    navigate.mockReset();
    apiDataMock.mockReset();
    apiFetchMock.mockReset();
    apiDataMock.mockResolvedValue({
      items: [{ id: 'u1', username: 'alice', email: 'a@test.local', role: 'user', winCount: 1, deletedAt: null }],
    });
    apiFetchMock.mockResolvedValue(undefined);
    vi.spyOn(window, 'confirm').mockReturnValue(true);
    vi.spyOn(window, 'open').mockImplementation(() => null);
  });

  it('loads users on mount', async () => {
    const { result } = renderHook(() => useAdminPage(), { wrapper });
    await waitFor(() => expect(result.current.users).toHaveLength(1));
  });

  it('searches users and deletes after confirm', async () => {
    apiDataMock.mockImplementation((path: string) => {
      if (path.startsWith('/admin/users/search')) {
        return Promise.resolve([
          { id: 'u2', username: 'bob', email: 'b@test.local', role: 'user', winCount: 0, deletedAt: null },
        ]);
      }
      return Promise.resolve({
        items: [{ id: 'u1', username: 'alice', email: 'a@test.local', role: 'user', winCount: 1, deletedAt: null }],
      });
    });

    const { result } = renderHook(() => useAdminPage(), { wrapper });
    await waitFor(() => expect(result.current.users).toHaveLength(1));
    await act(async () => {
      await result.current.searchUsers();
    });
    expect(result.current.users[0]?.username).toBe('bob');
    act(() => {
      result.current.deleteUser('u2');
    });
    await waitFor(() => expect(result.current.users).toHaveLength(0));
  });
});
