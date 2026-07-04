import { describe, expect, it } from 'vitest';
import { createInitialGameState, gameReducer } from '../hooks/useGameState';
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
});
