import { describe, expect, it } from 'vitest';
import { renderHook } from '@testing-library/react';
import { createInitialGameState, gameReducer, GameStateProvider, useGameState } from './useGameState';
import type { GameStateDTO } from '../types/dto';

describe('gameReducer', () => {
  it('does not store secret numbers in state', () => {
    const initial = createInitialGameState('game-1');
    expect(Object.keys(initial)).not.toContain('secretNumber');
  });

  it('applies GUESS_RESULT for own player', () => {
    const initial = createInitialGameState('game-1');
    const next = gameReducer(initial, {
      type: 'GUESS_RESULT',
      userId: 'user-1',
      pendingGuess: '1234',
      data: {
        playerId: 'user-1',
        digitResults: [0, 1, 0, 0],
        hitCount: 1,
        isAuto: false,
        nextTurnPlayerID: 'user-2',
      },
    });
    expect(next.myGuesses).toHaveLength(1);
    expect(next.myGuesses[0].guessNumber).toBe('1234');
  });

  it('syncs full game state from server payload', () => {
    const initial = createInitialGameState('game-1');
    const dto: GameStateDTO = {
      gameId: 'game-1',
      status: 'IN_PROGRESS',
      currentTurn: 2,
      currentTurnPlayerID: 'user-2',
      remainingSeconds: 15,
      myGuesses: [],
      opponentGuessCount: 1,
    };
    const next = gameReducer(initial, { type: 'SET_STATE', state: dto, userId: 'user-1' });
    expect(next.status).toBe('IN_PROGRESS');
    expect(next.opponentGuessCount).toBe(1);
  });

  it('applies TURN_CHANGED payload', () => {
    const initial = createInitialGameState('game-1');
    const next = gameReducer(initial, {
      type: 'TURN_CHANGED',
      data: { currentTurn: 3, currentTurnPlayerID: 'user-2', remainingSeconds: 20 },
    });
    expect(next.status).toBe('IN_PROGRESS');
    expect(next.currentTurn).toBe(3);
    expect(next.currentTurnPlayerID).toBe('user-2');
    expect(next.remainingSeconds).toBe(20);
  });

  it('stores GAME_OVER payload and clears it', () => {
    const initial = createInitialGameState('game-1');
    const gameOver = { gameId: 'game-1', reason: 'guess_win' as const, winnerId: 'user-1' };
    const finished = gameReducer(initial, { type: 'GAME_OVER', data: gameOver });
    expect(finished.status).toBe('FINISHED');
    expect(finished.gameOver).toEqual(gameOver);

    const cleared = gameReducer(finished, { type: 'CLEAR_GAME_OVER' });
    expect(cleared.gameOver).toBeNull();
  });

  it('increments opponent guess count for other player', () => {
    const initial = createInitialGameState('game-1');
    const next = gameReducer(initial, {
      type: 'GUESS_RESULT',
      userId: 'user-1',
      data: { playerId: 'user-2', hitCount: 1 },
    });
    expect(next.myGuesses).toHaveLength(0);
    expect(next.opponentGuessCount).toBe(1);
  });

  it('tracks opponent disconnect status', () => {
    const initial = createInitialGameState('game-1');
    const disconnected = gameReducer(initial, { type: 'OPPONENT_STATUS', connected: false });
    expect(disconnected.opponentDisconnected).toBe(true);
    const reconnected = gameReducer(disconnected, { type: 'OPPONENT_STATUS', connected: true });
    expect(reconnected.opponentDisconnected).toBe(false);
  });

  it('handles LOADING, ERROR, SECRET_SUBMITTED, and unknown actions', () => {
    const initial = createInitialGameState('game-1');
    expect(gameReducer(initial, { type: 'LOADING' }).loading).toBe(true);
    expect(gameReducer(initial, { type: 'ERROR', message: 'fail' }).error).toBe('fail');
    expect(gameReducer(initial, { type: 'SECRET_SUBMITTED' }).secretSubmitted).toBe(true);
    expect(gameReducer(initial, { type: 'UNKNOWN' as never })).toEqual(initial);
  });

  it('preserves gameOver when syncing FINISHED state', () => {
    const initial = createInitialGameState('game-1');
    const withGameOver = gameReducer(initial, {
      type: 'GAME_OVER',
      data: { gameId: 'game-1', reason: 'guess_win', winnerId: 'user-1' },
    });
    const synced = gameReducer(withGameOver, {
      type: 'SET_STATE',
      userId: 'user-1',
      state: {
        gameId: 'game-1',
        status: 'FINISHED',
        currentTurn: 1,
        currentTurnPlayerID: 'user-1',
        remainingSeconds: 0,
        myGuesses: [],
        opponentGuessCount: 0,
      },
    });
    expect(synced.gameOver?.winnerId).toBe('user-1');
  });

  it('uses placeholder guess when pendingGuess is missing', () => {
    const initial = createInitialGameState('game-1');
    const next = gameReducer(initial, {
      type: 'GUESS_RESULT',
      userId: 'user-1',
      data: { playerId: 'user-1', hitCount: 0, digitResults: [0, 0, 0, 0] },
    });
    expect(next.myGuesses[0]?.guessNumber).toBe('????');
  });
});

describe('GameStateProvider', () => {
  it('provides state and throws outside provider', () => {
    const { result } = renderHook(() => useGameState(), {
      wrapper: ({ children }) => <GameStateProvider gameId="game-1">{children}</GameStateProvider>,
    });
    expect(result.current.state.gameId).toBe('game-1');
    expect(() => renderHook(() => useGameState())).toThrow(
      'useGameState must be used within GameStateProvider',
    );
  });
});
