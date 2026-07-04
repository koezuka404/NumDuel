import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
import GamePage from './GamePage';
import RankingPage from './RankingPage';
import ProfilePage from './ProfilePage';
import AdminPage from './AdminPage';
import { ApiError } from '../api/client';

const apiData = vi.fn();
const useGamePage = vi.fn();
const useAdminPage = vi.fn();

vi.mock('../api/client', async () => {
  const actual = await vi.importActual<typeof import('../api/client')>('../api/client');
  return { ...actual, apiData: (...args: Parameters<typeof actual.apiData>) => apiData(...args) };
});

vi.mock('../hooks/useGamePage', () => ({
  useGamePage: () => useGamePage(),
}));

vi.mock('../hooks/useAdminPage', () => ({
  useAdminPage: () => useAdminPage(),
  ADMIN_TABS: [
    { id: 'users', label: 'ユーザー' },
    { id: 'logs', label: 'ログ' },
    { id: 'ranking', label: 'ランキング' },
    { id: 'backup', label: 'バックアップ' },
  ],
}));

vi.mock('../hooks/useAuth', () => ({
  useAuth: () => ({ user: { id: '1', username: 'alice', role: 'user' } }),
}));

vi.mock('../hooks/useLogout', () => ({
  useLogout: () => vi.fn(),
}));

describe('GamePage', () => {
  it('shows loading spinner', () => {
    useGamePage.mockReturnValue({
      gameId: 'game-1',
      user: { id: '1', username: 'alice', role: 'user' },
      state: { loading: true },
    });
    render(
      <MemoryRouter>
        <GamePage />
      </MemoryRouter>,
    );
    expect(screen.getByLabelText('ゲーム読み込み中')).toBeInTheDocument();
  });

  it('renders playing UI and result modal', () => {
    const closeResult = vi.fn();
    useGamePage.mockReturnValue({
      gameId: 'game-12345678',
      user: { id: '1', username: 'alice', role: 'user' },
      state: {
        loading: false,
        status: 'IN_PROGRESS',
        remainingSeconds: 10,
        myGuesses: [],
        opponentGuessCount: 1,
        opponentDisconnected: true,
        secretSubmitted: false,
        gameOver: { gameId: 'game-1', reason: 'guess_win', winnerId: '1' },
      },
      reconnectBanner: '再接続中…',
      timerMax: 30,
      isMyTurn: true,
      isSecretPhase: false,
      isPlaying: true,
      inputValue: '',
      setInputValue: vi.fn(),
      inputError: '',
      setInputError: vi.fn(),
      inputDisabled: false,
      submitCurrentInput: vi.fn(),
      closeResult,
    });

    render(
      <MemoryRouter>
        <GamePage />
      </MemoryRouter>,
    );

    expect(screen.getByText('再接続中…')).toBeInTheDocument();
    expect(screen.getByText('相手が切断しました')).toBeInTheDocument();
    expect(screen.getByText('勝利！')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '閉じる' }));
    expect(closeResult).toHaveBeenCalled();
  });

  it('renders secret phase UI', () => {
    useGamePage.mockReturnValue({
      gameId: 'game-12345678',
      user: { id: '1', username: 'alice', role: 'user' },
      state: {
        loading: false,
        status: 'WAITING_SECRET',
        remainingSeconds: 30,
        myGuesses: [],
        opponentGuessCount: 0,
        opponentDisconnected: false,
        secretSubmitted: true,
        gameOver: null,
      },
      reconnectBanner: '',
      timerMax: 60,
      isMyTurn: false,
      isSecretPhase: true,
      isPlaying: false,
      inputValue: '',
      setInputValue: vi.fn(),
      inputError: 'bad',
      setInputError: vi.fn(),
      inputDisabled: true,
      submitCurrentInput: vi.fn(),
      closeResult: vi.fn(),
    });

    render(
      <MemoryRouter>
        <GamePage />
      </MemoryRouter>,
    );

    expect(screen.getByText('bad')).toBeInTheDocument();
    expect(screen.getByText('登録済み — 相手を待っています')).toBeInTheDocument();
  });
});

describe('RankingPage', () => {
  it('loads ranking and reloads on button click', async () => {
    apiData.mockResolvedValue([{ rank: 1, username: 'alice', winCount: 2 }]);
    render(
      <MemoryRouter>
        <RankingPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByText('alice')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: '再読み込み' }));
    await waitFor(() => expect(apiData).toHaveBeenCalledTimes(2));
  });

  it('shows error message', async () => {
    apiData.mockRejectedValue(new ApiError('internal_error', 'failed', 500));
    render(
      <MemoryRouter>
        <RankingPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByText('failed')).toBeInTheDocument());
  });
});

describe('ProfilePage', () => {
  it('loads profile and tab histories', async () => {
    apiData.mockImplementation((path: string) => {
      if (path === '/me/profile') {
        return Promise.resolve({ username: 'alice', winCount: 2, rank: null });
      }
      if (path.startsWith('/me/match-history')) {
        return Promise.resolve({
          items: [
            {
              gameId: 'game-1',
              winnerUsername: 'alice',
              loserUsername: 'bob',
              finishedAt: '2024-01-01T00:00:00Z',
            },
          ],
        });
      }
      if (path.startsWith('/me/login-history')) {
        return Promise.resolve({
          items: [{ action: 'login', createdAt: '2024-01-01T00:00:00Z' }],
        });
      }
      return Promise.reject(new Error(path));
    });

    render(
      <MemoryRouter>
        <ProfilePage />
      </MemoryRouter>,
    );

    await waitFor(() => expect(screen.getByText('圏外')).toBeInTheDocument());
    expect(screen.getByText('alice')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'ログイン履歴' }));
    await waitFor(() => expect(screen.getByText('login')).toBeInTheDocument());
  });
});

describe('AdminPage', () => {
  it('renders admin tabs from hook state', () => {
    useAdminPage.mockReturnValue({
      tab: 'users',
      setTab: vi.fn(),
      busy: false,
      users: [],
      searchQ: '',
      setSearchQ: vi.fn(),
      logs: [],
      logType: '',
      setLogType: vi.fn(),
      logTypes: [],
      backup: null,
      searchUsers: vi.fn(),
      searchLogs: vi.fn(),
      deleteUser: vi.fn(),
      downloadLogs: vi.fn(),
      rebuildRanking: vi.fn(),
    });

    const { rerender } = render(
      <MemoryRouter>
        <AdminPage />
      </MemoryRouter>,
    );
    expect(screen.getByText('管理画面')).toBeInTheDocument();

    useAdminPage.mockReturnValue({
      tab: 'backup',
      setTab: vi.fn(),
      busy: false,
      users: [],
      searchQ: '',
      setSearchQ: vi.fn(),
      logs: [],
      logType: '',
      setLogType: vi.fn(),
      logTypes: [],
      backup: { status: 'ok', lastSyncedAt: '2024-01-01T00:00:00Z' },
      searchUsers: vi.fn(),
      searchLogs: vi.fn(),
      deleteUser: vi.fn(),
      downloadLogs: vi.fn(),
      rebuildRanking: vi.fn(),
    });
    rerender(
      <MemoryRouter>
        <AdminPage />
      </MemoryRouter>,
    );
    expect(screen.getByText('正常')).toBeInTheDocument();
  });
});
