import {
  createContext,
  useContext,
  useMemo,
  useReducer,
  type Dispatch,
  type ReactNode,
} from 'react';
import type { GameOverData, GameStateDTO, GuessDTO } from '../types/dto';

export type GameViewState = {
  gameId: string;
  status: GameStateDTO['status'];
  currentTurn: number;
  currentTurnPlayerID: string;
  remainingSeconds: number;
  myGuesses: GuessDTO[];
  opponentGuessCount: number;
  secretSubmitted: boolean;
  gameOver: GameOverData | null;
  opponentDisconnected: boolean;
  loading: boolean;
  error: string | null;
};

type GameAction =
  | { type: 'LOADING' }
  | { type: 'ERROR'; message: string }
  | { type: 'SET_STATE'; state: GameStateDTO; userId: string }
  | { type: 'SECRET_SUBMITTED' }
  | { type: 'TURN_CHANGED'; data: Record<string, unknown> }
  | { type: 'GUESS_RESULT'; data: Record<string, unknown>; userId: string; pendingGuess?: string }
  | { type: 'GAME_OVER'; data: GameOverData }
  | { type: 'OPPONENT_STATUS'; connected: boolean }
  | { type: 'CLEAR_GAME_OVER' };

export function createInitialGameState(gameId: string): GameViewState {
  return {
    gameId,
    status: 'WAITING_SECRET',
    currentTurn: 0,
    currentTurnPlayerID: '',
    remainingSeconds: 0,
    myGuesses: [],
    opponentGuessCount: 0,
    secretSubmitted: false,
    gameOver: null,
    opponentDisconnected: false,
    loading: true,
    error: null,
  };
}

function mapGameState(dto: GameStateDTO, prev: GameViewState): GameViewState {
  return {
    ...prev,
    gameId: dto.gameId,
    status: dto.status,
    currentTurn: dto.currentTurn,
    currentTurnPlayerID: dto.currentTurnPlayerID,
    remainingSeconds: dto.remainingSeconds,
    myGuesses: dto.myGuesses,
    opponentGuessCount: dto.opponentGuessCount,
    loading: false,
    error: null,
    gameOver: dto.status === 'FINISHED' ? prev.gameOver : null,
  };
}

export function gameReducer(state: GameViewState, action: GameAction): GameViewState {
  switch (action.type) {
    case 'LOADING':
      return { ...state, loading: true, error: null };
    case 'ERROR':
      return { ...state, loading: false, error: action.message };
    case 'SET_STATE':
      return mapGameState(action.state, state);
    case 'SECRET_SUBMITTED':
      return { ...state, secretSubmitted: true };
    case 'TURN_CHANGED':
      return {
        ...state,
        status: 'IN_PROGRESS',
        currentTurn: Number(action.data.currentTurn ?? state.currentTurn),
        currentTurnPlayerID: String(action.data.currentTurnPlayerID ?? state.currentTurnPlayerID),
        remainingSeconds: Number(action.data.remainingSeconds ?? state.remainingSeconds),
      };
    case 'GUESS_RESULT': {
      const playerId = String(action.data.playerId ?? '');
      if (playerId !== action.userId) {
        return {
          ...state,
          opponentGuessCount: state.opponentGuessCount + 1,
        };
      }
      const guessNumber = action.pendingGuess ?? '????';
      const nextGuess: GuessDTO = {
        turn: state.myGuesses.length + 1,
        guessNumber,
        digitResults: (action.data.digitResults as number[]) ?? [],
        hitCount: Number(action.data.hitCount ?? 0),
        isAuto: Boolean(action.data.isAuto),
      };
      return {
        ...state,
        myGuesses: [...state.myGuesses, nextGuess],
        currentTurnPlayerID: String(action.data.nextTurnPlayerID ?? state.currentTurnPlayerID),
      };
    }
    case 'GAME_OVER':
      return {
        ...state,
        status: 'FINISHED',
        gameOver: action.data,
      };
    case 'OPPONENT_STATUS':
      return { ...state, opponentDisconnected: !action.connected };
    case 'CLEAR_GAME_OVER':
      return { ...state, gameOver: null };
    default:
      return state;
  }
}

type GameStateContextValue = {
  state: GameViewState;
  dispatch: Dispatch<GameAction>;
};

const GameStateContext = createContext<GameStateContextValue | null>(null);

export function GameStateProvider({
  gameId,
  children,
}: {
  gameId: string;
  children: ReactNode;
}) {
  const [state, dispatch] = useReducer(gameReducer, gameId, createInitialGameState);
  const value = useMemo(() => ({ state, dispatch }), [state]);
  return <GameStateContext.Provider value={value}>{children}</GameStateContext.Provider>;
}

export function useGameState(): GameStateContextValue {
  const ctx = useContext(GameStateContext);
  if (!ctx) {
    throw new Error('useGameState must be used within GameStateProvider');
  }
  return ctx;
}
