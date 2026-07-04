type Props = {
  remainingSeconds: number;
  maxSeconds: number;
};

export default function GameTimer({ remainingSeconds, maxSeconds }: Props) {
  const safeMax = Math.max(maxSeconds, 1);
  const ratio = Math.min(Math.max(remainingSeconds / safeMax, 0), 1);
  const degrees = ratio * 360;

  return (
    <div className="game-timer" aria-live="polite">
      <div
        className="game-timer__ring"
        style={{ background: `conic-gradient(#3b82f6 ${degrees}deg, #334155 ${degrees}deg)` }}
      >
        <div className="game-timer__inner">
          <span className="game-timer__icon" aria-hidden>
            ⏱
          </span>
          <span className="game-timer__label">残り{remainingSeconds}秒</span>
        </div>
      </div>
    </div>
  );
}
