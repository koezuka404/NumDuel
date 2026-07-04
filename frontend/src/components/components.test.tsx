import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import GuessHistory from './GuessHistory';
import GameResultModal from './GameResultModal';
import GameTimer from './GameTimer';
import GameTurnBanner from './GameTurnBanner';
import NumericKeypad from './NumericKeypad';
import RankingFooter from './RankingFooter';
import RankingTable from './RankingTable';

describe('game components', () => {
  afterEach(() => cleanup());

  it('GuessHistory renders empty and filled states', () => {
    const { rerender } = render(<GuessHistory guesses={[]} emptyLabel="empty" />);
    expect(screen.getByText('empty')).toBeInTheDocument();
    rerender(
      <GuessHistory
        guesses={[{ turn: 1, guessNumber: '1234', digitResults: [1, 0], hitCount: 1, isAuto: false }]}
      />,
    );
    expect(screen.getByText('1234')).toBeInTheDocument();
  });

  it('GameResultModal shows win/lose/timeout titles', () => {
    const { rerender } = render(
      <GameResultModal
        gameOver={{ gameId: 'g1', reason: 'guess_win', winnerId: 'u1' }}
        userId="u1"
        onClose={vi.fn()}
      />,
    );
    expect(screen.getByText('勝利！')).toBeInTheDocument();
    rerender(
      <GameResultModal
        gameOver={{ gameId: 'g1', reason: 'guess_win', winnerId: 'u2' }}
        userId="u1"
        onClose={vi.fn()}
      />,
    );
    expect(screen.getByText('敗北…')).toBeInTheDocument();
  });

  it('GameTimer and GameTurnBanner render labels', () => {
    render(<GameTimer remainingSeconds={10} maxSeconds={30} />);
    expect(screen.getByText('残り10秒')).toBeInTheDocument();
    render(<GameTurnBanner isMyTurn />);
    expect(screen.getByText('あなたのターンです')).toBeInTheDocument();
  });

  it('NumericKeypad submits 4 digits', () => {
    const onChange = vi.fn();
    const onSubmit = vi.fn();
    const { rerender } = render(
      <NumericKeypad value="" onChange={onChange} onSubmit={onSubmit} submitLabel="送信" />,
    );
    fireEvent.click(screen.getByRole('button', { name: '1' }));
    fireEvent.click(screen.getByRole('button', { name: '2' }));
    fireEvent.click(screen.getByRole('button', { name: '3' }));
    rerender(<NumericKeypad value="123" onChange={onChange} onSubmit={onSubmit} submitLabel="送信" />);
    fireEvent.click(screen.getByRole('button', { name: '4' }));
    rerender(<NumericKeypad value="1234" onChange={onChange} onSubmit={onSubmit} submitLabel="送信" />);
    fireEvent.click(screen.getByRole('button', { name: '送信' }));
    expect(onSubmit).toHaveBeenCalled();
  });

  it('RankingTable and RankingFooter render', () => {
    render(<RankingTable items={[]} />);
    expect(screen.getByText('ランキングデータがありません')).toBeInTheDocument();
    render(<RankingFooter updatedAt="2024/01/01" />);
    expect(screen.getByText(/最終更新:/)).toBeInTheDocument();
  });
});
