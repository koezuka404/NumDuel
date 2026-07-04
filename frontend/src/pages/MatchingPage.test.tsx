import { act, cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import MatchingPage from './MatchingPage';
import { ApiError } from '../api/client';

const navigate = vi.fn();
const apiData = vi.fn();
const showToast = vi.fn();

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom');
  return { ...actual, useNavigate: () => navigate };
});

vi.mock('../api/client', async () => {
  const actual = await vi.importActual<typeof import('../api/client')>('../api/client');
  return { ...actual, apiData: (...args: Parameters<typeof actual.apiData>) => apiData(...args) };
});

vi.mock('../hooks/useAuth', () => ({
  useAuth: () => ({ user: { id: 'user-1', username: 'alice', role: 'user' } }),
}));

vi.mock('../hooks/useToast', () => ({
  useToast: () => ({ showToast }),
}));

vi.mock('../hooks/useWebSocket', () => ({
  useWebSocket: () => ({
    subscribe: () => () => undefined,
    connecting: false,
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
      if (path === '/matching/status') return Promise.resolve({ status: 'idle' });
      if (path === '/matching/start' && options?.method === 'POST') {
        return Promise.resolve({ status: 'matched', gameId: 'game-123' });
      }
      return Promise.reject(new Error(`unexpected: ${path}`));
    });

    renderMatchingPage();
    await waitFor(() => expect(screen.getByText('マッチングを開始できます')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'マッチング開始' }));
    await waitFor(() => expect(navigate).toHaveBeenCalledWith('/game/game-123'));
  });

  it('polls status while waiting', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    let statusCalls = 0;

    apiData.mockImplementation((path: string, options?: RequestInit) => {
      if (path === '/matching/status') {
        statusCalls += 1;
        return Promise.resolve(
          statusCalls === 1 ? { status: 'idle' } : { status: 'matched', gameId: 'game-poll' },
        );
      }
      if (path === '/matching/start' && options?.method === 'POST') {
        return Promise.resolve({ status: 'waiting' });
      }
      return Promise.reject(new Error(`unexpected: ${path}`));
    });

    renderMatchingPage();
    await waitFor(() => expect(screen.getByText('マッチングを開始できます')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'マッチング開始' }));
    await waitFor(() => expect(screen.getByText('対戦相手を探しています…')).toBeInTheDocument());

    await act(async () => {
      await vi.advanceTimersByTimeAsync(2000);
    });
    await waitFor(() => expect(navigate).toHaveBeenCalledWith('/game/game-poll'));
  });

  it('handles already_in_matching error', async () => {
    apiData.mockImplementation((path: string, options?: RequestInit) => {
      if (path === '/matching/status') return Promise.resolve({ status: 'idle' });
      if (path === '/matching/start' && options?.method === 'POST') {
        return Promise.reject(new ApiError('already_in_matching', 'waiting', 409));
      }
      return Promise.reject(new Error(`unexpected: ${path}`));
    });

    renderMatchingPage();
    fireEvent.click(screen.getByRole('button', { name: 'マッチング開始' }));
    await waitFor(() => expect(showToast).toHaveBeenCalledWith('既に待機中です', 'info'));
  });
});
