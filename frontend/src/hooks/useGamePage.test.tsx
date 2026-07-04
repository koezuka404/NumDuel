import { act, renderHook, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { ApiError } from '../api/client';
import { createInitialGameState } from './useGameState';
import { useGamePage } from './useGamePage';

const navigate = vi.fn();
const showToast = vi.fn();
const send = vi.fn();
const subscribe = vi.fn();
const dispatch = vi.fn();
const apiDataMock = vi.fn();
let connected = true;
let routeGameId = 'game-1';
let gameState = createInitialGameState('game-1');

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

vi.mock('./useGameState', async () => {
  const actual = await vi.importActual<typeof import('./useGameState')>('./useGameState');
  return {
    ...actual,
    useGameState: () => ({ state: gameState, dispatch }),
  };
});

describe('useGamePage', () => {
  let messageHandler: ((msg: { type: string; data?: Record<string, unknown> }) => void) | null = null;

  beforeEach(() => {
    routeGameId = 'game-1';
    connected = true;
    gameState = { ...createInitialGameState('game-1'), loading: false, status: 'IN_PROGRESS', currentTurnPlayerID: 'user-1' };
    navigate.mockReset();
    showToast.mockReset();
    send.mockReset();
    dispatch.mockReset();
    messageHandler = null;
    subscribe.mockImplementation((handler: typeof messageHandler) => {
      messageHandler = handler;
      return () => undefined;
    });
    dispatch.mockImplementation((action: { type: string; state?: { status: string; currentTurnPlayerID: string } }) => {
      if (action.type === 'SET_STATE' && action.state) {
        Object.assign(gameState, {
          loading: false,
          status: action.state.status,
          currentTurnPlayerID: action.state.currentTurnPlayerID,
        });
      }
      if (action.type === 'GAME_OVER') {
        Object.assign(gameState, { status: 'FINISHED', gameOver: action });
      }
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

  it('loads game state and sends sync request', async () => {
    renderHook(() => useGamePage());
    await waitFor(() => expect(apiDataMock).toHaveBeenCalled());
    expect(send).toHaveBeenCalledWith({ type: 'SYNC_REQUEST', gameId: 'game-1' });
  });

  it('navigates away on not_found fetch error', async () => {
    apiDataMock.mockRejectedValueOnce(new ApiError('not_found', 'missing', 404));
    renderHook(() => useGamePage());
    await waitFor(() => expect(navigate).toHaveBeenCalledWith('/matching'));
  });

  it('handles websocket errors', async () => {
    renderHook(() => useGamePage());
    await waitFor(() => expect(messageHandler).toBeTruthy());
    act(() => {
      messageHandler?.({ type: 'ERROR', data: { code: 'not_your_turn', message: 'wait' } });
    });
    expect(showToast).toHaveBeenCalled();
  });

  it('shows reconnect banner when disconnected', async () => {
    connected = false;
    const { result } = renderHook(() => useGamePage());
    await waitFor(() => expect(result.current.reconnectBanner).toBe('再接続中…'));
  });

  it('skips fetch when gameId is missing', () => {
    routeGameId = '';
    apiDataMock.mockClear();
    renderHook(() => useGamePage());
    expect(apiDataMock).not.toHaveBeenCalled();
  });
});
