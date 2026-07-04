import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import GuessHistory from './GuessHistory';
import GameResultModal from './GameResultModal';
import GameTimer from './GameTimer';
import GameTurnBanner from './GameTurnBanner';
import NumericKeypad from './NumericKeypad';
import RankingFooter from './RankingFooter';
import RankingTable from './RankingTable';

describe('GuessHistory', () => {
  it('shows empty label when there are no guesses', () => {
    render(<GuessHistory guesses={[]} emptyLabel="empty" />);
    expect(screen.getByText('empty')).toBeInTheDocument();
  });

  it('renders guess rows with hit and miss marks', () => {
    render(
      <GuessHistory
        guesses={[
          {
            turn: 1,
            guessNumber: '1234',
            digitResults: [1, 0, 1, 0],
            hitCount: 2,
            isAuto: true,
          },
        ]}
      />,
    );
    expect(screen.getByText('1234')).toBeInTheDocument();
    expect(screen.getAllByText('○')).toHaveLength(2);
    expect(screen.getAllByText('×')).toHaveLength(2);
    expect(screen.getByText('自動')).toBeInTheDocument();
  });
});

describe('GameResultModal', () => {
  it('shows win title', () => {
    render(
      <GameResultModal
        gameOver={{ gameId: 'g1', reason: 'guess_win', winnerId: 'u1' }}
        userId="u1"
        onClose={vi.fn()}
      />,
    );
    expect(screen.getByText('勝利！')).toBeInTheDocument();
  });

  it('shows lose title', () => {
    render(
      <GameResultModal
        gameOver={{ gameId: 'g1', reason: 'guess_win', winnerId: 'u2' }}
        userId="u1"
        onClose={vi.fn()}
      />,
    );
    expect(screen.getByText('敗北…')).toBeInTheDocument();
  });

  it('shows default game over title', () => {
    render(
      <GameResultModal
        gameOver={{ gameId: 'g1', reason: 'disconnect' as 'guess_win' }}
        userId="u1"
        onClose={vi.fn()}
      />,
    );
    expect(screen.getByText('ゲーム終了')).toBeInTheDocument();
  });

  it('shows timeout title and closes', () => {
    const onClose = vi.fn();
    render(
      <GameResultModal
        gameOver={{ gameId: 'g1', reason: 'secret_setup_timeout' }}
        userId="u1"
        onClose={onClose}
      />,
    );
    expect(screen.getByText('登録時間切れのためゲーム終了')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '閉じる' }));
    expect(onClose).toHaveBeenCalled();
  });
});

describe('GameTimer', () => {
  it('shows remaining seconds label', () => {
    render(<GameTimer remainingSeconds={15} maxSeconds={30} />);
    expect(screen.getByText('残り15秒')).toBeInTheDocument();
  });

  it('clamps ratio when max is zero', () => {
    render(<GameTimer remainingSeconds={0} maxSeconds={0} />);
    expect(screen.getByText('残り0秒')).toBeInTheDocument();
  });
});

describe('GameTurnBanner', () => {
  it('shows default mine turn text', () => {
    render(<GameTurnBanner isMyTurn />);
    expect(screen.getByText('あなたのターンです')).toBeInTheDocument();
  });

  it('shows default opponent turn text', () => {
    render(<GameTurnBanner isMyTurn={false} />);
    expect(screen.getByText('相手のターンです')).toBeInTheDocument();
  });

  it('shows custom label', () => {
    render(<GameTurnBanner isMyTurn label="custom" />);
    expect(screen.getByText('custom')).toBeInTheDocument();
  });
});

describe('NumericKeypad', () => {
  it('appends unique digits and submits', () => {
    const onChange = vi.fn();
    const onSubmit = vi.fn();
    const { rerender } = render(
      <NumericKeypad value="" onChange={onChange} onSubmit={onSubmit} submitLabel="送信" />,
    );

    fireEvent.click(screen.getByRole('button', { name: '1' }));
    expect(onChange).toHaveBeenCalledWith('1');

    rerender(<NumericKeypad value="123" onChange={onChange} onSubmit={onSubmit} />);
    fireEvent.click(screen.getByRole('button', { name: '4' }));
    expect(onChange).toHaveBeenLastCalledWith('1234');

    rerender(<NumericKeypad value="1234" onChange={onChange} onSubmit={onSubmit} />);
    fireEvent.click(screen.getByRole('button', { name: '送信' }));
    expect(onSubmit).toHaveBeenCalled();
  });

  it('blocks duplicate digits, backspace, and disabled input', () => {
    const onChange = vi.fn();
    render(<NumericKeypad value="12" onChange={onChange} onSubmit={vi.fn()} disabled />);
    fireEvent.click(screen.getByRole('button', { name: '1' }));
    expect(onChange).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole('button', { name: '⌫' }));
    expect(onChange).not.toHaveBeenCalled();

    render(<NumericKeypad value="12" onChange={onChange} onSubmit={vi.fn()} />);
    fireEvent.click(screen.getByRole('button', { name: '2' }));
    expect(onChange).not.toHaveBeenCalledWith('122');
    fireEvent.click(screen.getByRole('button', { name: '⌫' }));
    expect(onChange).toHaveBeenCalledWith('1');
  });
});

describe('RankingTable', () => {
  it('shows empty message', () => {
    render(<RankingTable items={[]} />);
    expect(screen.getByText('ランキングデータがありません')).toBeInTheDocument();
  });

  it('renders ranking rows', () => {
    render(<RankingTable items={[{ rank: 1, username: 'alice', winCount: 3 }]} />);
    expect(screen.getByText('alice')).toBeInTheDocument();
  });
});

describe('RankingFooter', () => {
  it('shows updated time when provided', () => {
    render(<RankingFooter updatedAt="2024/01/01" />);
    expect(screen.getByText(/最終更新:/)).toBeInTheDocument();
  });

  it('hides updated time when null', () => {
    render(<RankingFooter updatedAt={null} />);
    expect(screen.queryByText(/最終更新:/)).not.toBeInTheDocument();
  });
});
