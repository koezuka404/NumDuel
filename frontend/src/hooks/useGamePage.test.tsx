import { act, renderHook, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { ApiError } from '../api/client';
import { GameStateProvider } from './useGameState';
import { useGamePage } from './useGamePage';

const navigate = vi.fn();
const showToast = vi.fn();
const send = vi.fn();
const subscribe = vi.fn();
const apiDataMock = vi.fn();
let connected = true;
let routeGameId = 'game-1';

vi.mock('../api/client', async () => {
  const actual = await vi.importActual<typeof import('../api/client')>('../api/client');
  return { ...actual, apiData: (...args: Parameters<typeof actual.apiData>) => apiDataMock(...args) };
});

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom');
  return {
    ...actual,
    useNavigate: () => navigate,
    useParams: () => ({ id: routeGameId }),
  };
});

vi.mock('./useAuth', () => ({
  useAuth: () => ({ user: { id: 'user-1', username: 'alice', role: 'user' } }),
}));

vi.mock('./useToast', () => ({
  useToast: () => ({ showToast }),
}));

vi.mock('./useWebSocket', () => ({
  useWebSocket: () => ({
    send,
    subscribe,
    get connected() {
      return connected;
    },
  }),
}));

function wrapper({ children }: { children: React.ReactNode }) {
  return <GameStateProvider gameId="game-1">{children}</GameStateProvider>;
}

describe('useGamePage', () => {
  let messageHandler: ((msg: { type: string; data?: Record<string, unknown> }) => void) | null = null;

  beforeEach(() => {
    routeGameId = 'game-1';
    connected = true;
    navigate.mockReset();
    showToast.mockReset();
    send.mockReset();
    messageHandler = null;
    subscribe.mockImplementation((handler: typeof messageHandler) => {
      messageHandler = handler;
      return () => undefined;
    });
    apiDataMock.mockResolvedValue({
      gameId: 'game-1',
      status: 'IN_PROGRESS',
      currentTurn: 1,
      currentTurnPlayerID: 'user-1',
      remainingSeconds: 30,
      myGuesses: [],
      opponentGuessCount: 0,
    });
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('loads game state and sends sync request', async () => {
    const { result } = renderHook(() => useGamePage(), { wrapper });
    await waitFor(() => expect(result.current.state.loading).toBe(false));
    expect(send).toHaveBeenCalledWith({ type: 'SYNC_REQUEST', gameId: 'game-1' });
  });

  it('navigates away on not_found fetch error', async () => {
    apiDataMock.mockRejectedValueOnce(new ApiError('not_found', 'missing', 404));
    renderHook(() => useGamePage(), { wrapper });
    await waitFor(() => expect(navigate).toHaveBeenCalledWith('/matching'));
  });

  it('handles websocket messages and submits secret/guess', async () => {
    apiDataMock.mockResolvedValueOnce({
      gameId: 'game-1',
      status: 'WAITING_SECRET',
      currentTurn: 0,
      currentTurnPlayerID: 'user-1',
      remainingSeconds: 60,
      myGuesses: [],
      opponentGuessCount: 0,
    });

    const { result } = renderHook(() => useGamePage(), { wrapper });
    await waitFor(() => expect(result.current.isSecretPhase).toBe(true));

    act(() => {
      result.current.setInputValue('1234');
      result.current.submitCurrentInput();
    });
    expect(send).toHaveBeenCalledWith({ type: 'SET_SECRET', gameId: 'game-1', secretNumber: '1234' });

    act(() => {
      messageHandler?.({
        type: 'GAME_STATE_SYNC',
        data: {
          gameId: 'game-1',
          status: 'IN_PROGRESS',
          currentTurn: 1,
          currentTurnPlayerID: 'user-1',
          remainingSeconds: 25,
          myGuesses: [],
          opponentGuessCount: 0,
        },
      });
      messageHandler?.({ type: 'TURN_CHANGED', data: { currentTurn: 2, remainingSeconds: 20 } });
      messageHandler?.({
        type: 'GUESS_RESULT',
        data: { playerId: 'user-1', hitCount: 1, digitResults: [1, 0, 0, 0] },
      });
      messageHandler?.({ type: 'OPPONENT_STATUS', data: { connected: false } });
      messageHandler?.({
        type: 'GAME_OVER',
        data: { gameId: 'game-1', reason: 'guess_win', winnerId: 'user-1' },
      });
    });

    act(() => {
      result.current.setInputValue('5678');
      result.current.submitCurrentInput();
    });
    expect(send).toHaveBeenCalledWith({ type: 'GUESS', gameId: 'game-1', guessNumber: '5678' });

    act(() => {
      result.current.setInputValue('1123');
      result.current.submitCurrentInput();
    });
    expect(result.current.inputError).toBeTruthy();

    act(() => {
      result.current.closeResult();
    });
    expect(navigate).toHaveBeenCalledWith('/matching');
  });

  it('handles websocket errors and reconnect banner', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    apiDataMock.mockResolvedValueOnce({
      gameId: 'game-1',
      status: 'FINISHED',
      currentTurn: 1,
      currentTurnPlayerID: 'user-1',
      remainingSeconds: 0,
      myGuesses: [],
      opponentGuessCount: 0,
    });

    const { result, rerender } = renderHook(() => useGamePage(), { wrapper });
    await waitFor(() => expect(result.current.state.status).toBe('FINISHED'));

    act(() => {
      messageHandler?.({ type: 'ERROR', data: { code: 'forbidden', message: 'forbidden' } });
      messageHandler?.({ type: 'ERROR', data: { code: 'game_already_finished', message: 'done' } });
      messageHandler?.({ type: 'ERROR', data: { code: 'not_your_turn', message: 'wait' } });
    });
    expect(navigate).toHaveBeenCalledWith('/matching');
    expect(showToast).toHaveBeenCalled();

    connected = false;
    rerender();
    await waitFor(() => expect(result.current.reconnectBanner).toBe('再接続中…'));

    connected = true;
    rerender();
    await waitFor(() => expect(result.current.reconnectBanner).toBe('接続が復旧しました'));

    await act(async () => {
      await vi.advanceTimersByTimeAsync(3000);
    });
    await waitFor(() => expect(result.current.reconnectBanner).toBe(''));
  });

  it('skips fetch when gameId is missing', async () => {
    routeGameId = '';
    apiDataMock.mockClear();
    renderHook(() => useGamePage(), { wrapper });
    expect(apiDataMock).not.toHaveBeenCalled();
  });
});
