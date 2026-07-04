import { act, cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import MatchingPage from './MatchingPage';
import { ApiError } from '../api/client';

const navigate = vi.fn();
const apiData = vi.fn();
const showToast = vi.fn();
let wsHandler: ((msg: { type: string; data?: Record<string, unknown> }) => void) | null = null;
let connecting = false;
let authUser = { id: 'user-1', username: 'alice', role: 'user' as const };

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom');
  return { ...actual, useNavigate: () => navigate };
});

vi.mock('../api/client', async () => {
  const actual = await vi.importActual<typeof import('../api/client')>('../api/client');
  return { ...actual, apiData: (...args: Parameters<typeof actual.apiData>) => apiData(...args) };
});

vi.mock('../hooks/useAuth', () => ({
  useAuth: () => ({ user: authUser }),
}));

vi.mock('../hooks/useToast', () => ({
  useToast: () => ({ showToast }),
}));

vi.mock('../hooks/useWebSocket', () => ({
  useWebSocket: () => ({
    subscribe: (handler: typeof wsHandler) => {
      wsHandler = handler;
      return () => undefined;
    },
    connecting,
  }),
}));

vi.mock('../components/layout/AuthenticatedLayout', () => ({
  default: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}));

function renderMatchingPage() {
  return render(
    <MemoryRouter>
      <MatchingPage />
    </MemoryRouter>,
  );
}

describe('MatchingPage', () => {
  beforeEach(() => {
    authUser = { id: 'user-1', username: 'alice', role: 'user' };
    connecting = false;
    wsHandler = null;
    navigate.mockReset();
    apiData.mockReset();
    showToast.mockReset();
  });

  afterEach(() => {
    cleanup();
    vi.useRealTimers();
  });

  it('navigates to game when start returns matched', async () => {
    apiData.mockImplementation((path: string, options?: RequestInit) => {
      if (path === '/matching/status') {
        return Promise.resolve({ status: 'idle' });
      }
      if (path === '/matching/start' && options?.method === 'POST') {
        return Promise.resolve({ status: 'matched', gameId: 'game-123' });
      }
      return Promise.reject(new Error(`unexpected api call: ${path}`));
    });

    renderMatchingPage();
    await waitFor(() => {
      expect(screen.getByText('マッチングを開始できます')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole('button', { name: 'マッチング開始' }));

    await waitFor(() => {
      expect(navigate).toHaveBeenCalledWith('/game/game-123');
    });
  });

  it('polls status while waiting and navigates when matched', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    let statusCalls = 0;

    apiData.mockImplementation((path: string, options?: RequestInit) => {
      if (path === '/matching/status') {
        statusCalls += 1;
        if (statusCalls === 1) {
          return Promise.resolve({ status: 'idle' });
        }
        return Promise.resolve({ status: 'matched', gameId: 'game-poll' });
      }
      if (path === '/matching/start' && options?.method === 'POST') {
        return Promise.resolve({ status: 'waiting' });
      }
      return Promise.reject(new Error(`unexpected api call: ${path}`));
    });

    renderMatchingPage();
    await waitFor(() => {
      expect(screen.getByText('マッチングを開始できます')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole('button', { name: 'マッチング開始' }));

    await waitFor(() => {
      expect(screen.getByText('対戦相手を探しています…')).toBeInTheDocument();
    });

    await act(async () => {
      await vi.advanceTimersByTimeAsync(2000);
    });

    await waitFor(() => {
      expect(navigate).toHaveBeenCalledWith('/game/game-poll');
    });
  });

  it('handles websocket events, cancel, connecting message, and errors', async () => {
    connecting = true;
    apiData.mockImplementation((path: string, options?: RequestInit) => {
      if (path === '/matching/status') {
        return Promise.resolve({ status: 'idle' });
      }
      if (path === '/matching/cancel' && options?.method === 'POST') {
        return Promise.resolve({ status: 'idle' });
      }
      if (path === '/matching/start' && options?.method === 'POST') {
        return Promise.reject(new ApiError('already_in_matching', 'waiting', 409));
      }
      return Promise.reject(new Error(`unexpected api call: ${path}`));
    });

    renderMatchingPage();
    await waitFor(() => expect(screen.getByText('リアルタイム接続中…')).toBeInTheDocument());

    act(() => {
      wsHandler?.({ type: 'MATCHED', data: { gameId: 'ws-game' } });
    });
    expect(navigate).toHaveBeenCalledWith('/game/ws-game');

    act(() => {
      wsHandler?.({ type: 'RECONNECT_FAILED' });
    });
    expect(showToast).toHaveBeenCalledWith('リアルタイム接続の再接続に失敗しました', 'error');

    fireEvent.click(screen.getByRole('button', { name: 'マッチング開始' }));
    await waitFor(() => expect(showToast).toHaveBeenCalledWith('既に待機中です', 'info'));

    apiData.mockImplementation((path: string, options?: RequestInit) => {
      if (path === '/matching/status') {
        return Promise.resolve({ status: 'waiting' });
      }
      if (path === '/matching/cancel' && options?.method === 'POST') {
        return Promise.resolve({ status: 'idle' });
      }
      return Promise.reject(new Error(`unexpected api call: ${path}`));
    });
    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    await waitFor(() => expect(screen.getByText('マッチングを開始できます')).toBeInTheDocument());
  });

  it('handles matching error codes and master warning', async () => {
    authUser = { id: 'master-1', username: 'admin', role: 'master' };
    apiData.mockImplementation((path: string, options?: RequestInit) => {
      if (path === '/matching/status') {
        return Promise.resolve({ status: 'idle' });
      }
      if (path === '/matching/start' && options?.method === 'POST') {
        return Promise.reject(new ApiError('forbidden', 'denied', 403));
      }
      return Promise.reject(new Error(`unexpected api call: ${path}`));
    });

    renderMatchingPage();
    expect(screen.getByText('管理者アカウントはマッチングできません')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'マッチング開始' }));
    await waitFor(() => expect(navigate).toHaveBeenCalledWith('/admin'));

    apiData.mockImplementation((path: string, options?: RequestInit) => {
      if (path === '/matching/status') {
        return Promise.resolve({ status: 'idle' });
      }
      if (path === '/matching/start' && options?.method === 'POST') {
        return Promise.reject(new ApiError('user_in_active_game', 'active game', 409));
      }
      return Promise.reject(new Error(`unexpected api call: ${path}`));
    });
    authUser = { id: 'user-1', username: 'alice', role: 'user' };
    renderMatchingPage();
    fireEvent.click(screen.getAllByRole('button', { name: 'マッチング開始' })[0]!);
    await waitFor(() => expect(showToast).toHaveBeenCalledWith('active game', 'error'));
  });
});
