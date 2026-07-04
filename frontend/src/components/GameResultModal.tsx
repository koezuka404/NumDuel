import type { GameOverData } from '../types/dto';

type Props = {
  gameOver: GameOverData;
  userId: string;
  onClose: () => void;
};

export default function GameResultModal({ gameOver, userId, onClose }: Props) {
  let title = 'ゲーム終了';
  if (gameOver.reason === 'secret_setup_timeout') {
    title = '登録時間切れのためゲーム終了';
  } else if (gameOver.reason === 'guess_win') {
    title = gameOver.winnerId === userId ? '勝利！' : '敗北…';
  }

  return (
    <div className="game-modal-backdrop" role="dialog" aria-modal="true">
      <div className="game-modal">
        <h2>{title}</h2>
        <button type="button" className="game-keypad__key game-keypad__key--submit game-modal__button" onClick={onClose}>
          閉じる
        </button>
      </div>
    </div>
  );
}
