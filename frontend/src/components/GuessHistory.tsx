import type { GuessDTO } from '../types/dto';

type Props = {
  guesses: GuessDTO[];
  emptyLabel?: string;
};

export default function GuessHistory({ guesses, emptyLabel = 'まだ予想がありません' }: Props) {
  if (guesses.length === 0) {
    return <p className="game-history__empty">{emptyLabel}</p>;
  }

  return (
    <ul className="game-history">
      {guesses.map((guess, index) => (
        <li key={`${guess.turn}-${index}`} className="game-history__row">
          <span className="game-history__number">{guess.guessNumber}</span>
          <div className="game-history__marks">
            {guess.digitResults.map((result, digitIndex) => (
              <span
                key={digitIndex}
                className={`game-history__mark ${result === 1 ? 'game-history__mark--hit' : 'game-history__mark--miss'}`}
              >
                {result === 1 ? '○' : '×'}
              </span>
            ))}
          </div>
          {guess.isAuto && <span className="game-history__auto">自動</span>}
        </li>
      ))}
    </ul>
  );
}
