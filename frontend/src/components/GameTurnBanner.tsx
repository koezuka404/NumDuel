type Props = {
  isMyTurn: boolean;
  label?: string;
};

export default function GameTurnBanner({ isMyTurn, label }: Props) {
  const text = label ?? (isMyTurn ? 'あなたのターンです' : '相手のターンです');

  return (
    <div className={`game-turn-banner ${isMyTurn ? 'game-turn-banner--mine' : 'game-turn-banner--opponent'}`}>
      <span className="game-turn-banner__icon" aria-hidden>
        👤
      </span>
      <span>{text}</span>
    </div>
  );
}
