import { Link } from 'react-router-dom';
import GameResultModal from '../components/GameResultModal';
import GameTimer from '../components/GameTimer';
import GameTurnBanner from '../components/GameTurnBanner';
import GuessHistory from '../components/GuessHistory';
import NumericKeypad from '../components/NumericKeypad';
import { Spinner } from '../components/ui/FormField';
import { GameStateProvider } from '../hooks/useGameState';
import { useGamePage } from '../hooks/useGamePage';
import { useParams } from 'react-router-dom';

function GamePageInner() {
  const {
    gameId,
    user,
    state,
    reconnectBanner,
    timerMax,
    isMyTurn,
    isSecretPhase,
    isPlaying,
    inputValue,
    setInputValue,
    inputError,
    setInputError,
    inputDisabled,
    submitCurrentInput,
    closeResult,
  } = useGamePage();

  if (state.loading) {
    return (
      <main className="game-shell game-shell--loading">
        <Spinner label="ゲーム読み込み中" className="spinner--light" />
      </main>
    );
  }

  return (
    <main className="game-shell">
      <header className="game-shell__header">
        <Link to="/matching" className="game-shell__back">
          ← 戻る
        </Link>
        <span className="game-shell__title">NumDuel</span>
        <span className="game-shell__meta">#{gameId.slice(0, 8)}</span>
      </header>

      {reconnectBanner && <div className="game-banner">{reconnectBanner}</div>}
      {state.opponentDisconnected && <div className="game-banner game-banner--warn">相手が切断しました</div>}

      <section className="game-board">
        <GameTimer remainingSeconds={state.remainingSeconds} maxSeconds={timerMax} />

        {isSecretPhase && (
          <GameTurnBanner
            isMyTurn={!state.secretSubmitted}
            label={state.secretSubmitted ? '相手の登録待ち' : '秘密数字を入力してください'}
          />
        )}

        {isPlaying && <GameTurnBanner isMyTurn={isMyTurn} />}

        <div className="game-board__history">
          <div className="game-board__history-head">
            <span>予想履歴</span>
            {isPlaying && <span className="game-board__opponent-count">相手 {state.opponentGuessCount} 回</span>}
          </div>
          <GuessHistory
            guesses={state.myGuesses}
            emptyLabel={isSecretPhase ? '対戦開始後に表示されます' : 'まだ予想がありません'}
          />
        </div>

        {(isSecretPhase || isPlaying) && (
          <div className="game-board__input">
            {inputError && <p className="game-board__error">{inputError}</p>}
            {state.secretSubmitted && isSecretPhase && (
              <p className="game-board__hint">登録済み — 相手を待っています</p>
            )}
            <NumericKeypad
              value={inputValue}
              onChange={(value) => {
                setInputError('');
                setInputValue(value);
              }}
              onSubmit={submitCurrentInput}
              disabled={inputDisabled}
              submitLabel={isSecretPhase ? '登録' : '送信'}
            />
          </div>
        )}
      </section>

      {state.gameOver && user && (
        <GameResultModal gameOver={state.gameOver} userId={user.id} onClose={closeResult} />
      )}
    </main>
  );
}

export default function GamePage() {
  const { id: gameId = '' } = useParams();
  return (
    <GameStateProvider gameId={gameId}>
      <GamePageInner />
    </GameStateProvider>
  );
}
