export type UserRole = 'user' | 'master';

export type AuthUser = {
  id: string;
  username: string;
  role: UserRole;
  winCount?: number;
};

export type ApiErrorBody = {
  error?: {
    code?: string;
    message?: string;
  };
};

export type ApiDataResponse<T> = {
  data: T;
};

export type PagedResponse<T> = {
  data: {
    items: T[];
    page: number;
    limit: number;
    total: number;
  };
};

export type GameStatus = 'WAITING_SECRET' | 'IN_PROGRESS' | 'FINISHED';

export type GuessDTO = {
  turn: number;
  guessNumber: string;
  digitResults: number[];
  hitCount: number;
  isAuto: boolean;
};

export type GameStateDTO = {
  gameId: string;
  status: GameStatus;
  currentTurn: number;
  currentTurnPlayerID: string;
  remainingSeconds: number;
  myGuesses: GuessDTO[];
  opponentGuessCount: number;
};

export type RankingItemDTO = {
  rank: number;
  username: string;
  winCount: number;
};

export type ProfileDTO = {
  username: string;
  winCount: number;
  rank: number | null;
};

export type MatchHistoryItemDTO = {
  gameId: string;
  winnerUsername: string;
  loserUsername: string;
  finishedAt: string;
};

export type LoginHistoryItemDTO = {
  action: string;
  createdAt: string;
};

export type WSHistoryItemDTO = {
  connectionId: string;
  connectedAt: string;
  disconnectedAt: string | null;
};

export type AdminUserDTO = {
  id: string;
  username: string;
  email: string;
  role: UserRole;
  winCount: number;
  deletedAt: string | null;
  createdAt: string;
};

export type ActivityLogDTO = {
  id: string;
  userId: string | null;
  logType: string;
  detail: string;
  createdAt: string;
};

export type BackupStatusDTO = {
  status: string;
  lastSyncedAt: string | null;
};

export type MatchingStatusDTO = {
  status: 'idle' | 'waiting' | 'matched' | 'cancelled';
  gameId?: string | null;
};

export type WSMessage = {
  type: string;
  data?: Record<string, unknown>;
};

export type GameOverData = {
  gameId: string;
  reason: 'guess_win' | 'secret_setup_timeout';
  winnerId?: string;
};
