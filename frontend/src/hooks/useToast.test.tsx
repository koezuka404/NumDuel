import { act, render, renderHook, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { ToastProvider, useToast } from './useToast';

describe('useToast', () => {
  afterEach(() => {
    vi.useRealTimers();
  });

  it('shows and auto-removes toast', () => {
    vi.useFakeTimers();

    function TestComponent() {
      const { showToast } = useToast();
      return (
        <button type="button" onClick={() => showToast('hello', 'success')}>
          toast
        </button>
      );
    }

    render(
      <ToastProvider>
        <TestComponent />
      </ToastProvider>,
    );

    act(() => {
      screen.getByRole('button', { name: 'toast' }).click();
    });
    expect(screen.getByText('hello')).toHaveClass('toast-success');

    act(() => {
      vi.advanceTimersByTime(3000);
    });
    expect(screen.queryByText('hello')).not.toBeInTheDocument();
  });

  it('throws outside provider', () => {
    expect(() => renderHook(() => useToast())).toThrow('useToast must be used within ToastProvider');
  });
});
