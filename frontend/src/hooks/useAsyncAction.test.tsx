import { act, renderHook, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { ApiError } from '../api/client';
import { ToastProvider } from './useToast';
import { useAsyncAction } from './useAsyncAction';

const showToast = vi.fn();

vi.mock('./useToast', async () => {
  const actual = await vi.importActual<typeof import('./useToast')>('./useToast');
  return {
    ...actual,
    useToast: () => ({ showToast }),
  };
});

function wrapper({ children }: { children: React.ReactNode }) {
  return <ToastProvider>{children}</ToastProvider>;
}

describe('useAsyncAction', () => {
  it('runs action and clears busy state', async () => {
    const { result } = renderHook(() => useAsyncAction(), { wrapper });
    await act(async () => {
      await result.current.run(async () => undefined);
    });
    expect(result.current.busy).toBe(false);
  });

  it('shows toast for ApiError without handler', async () => {
    showToast.mockReset();
    const { result } = renderHook(() => useAsyncAction(), { wrapper });
    await act(async () => {
      await result.current.run(async () => {
        throw new ApiError('forbidden', 'denied', 403);
      });
    });
    expect(showToast).toHaveBeenCalledWith('denied', 'error');
  });

  it('calls custom error handler', async () => {
    const onError = vi.fn();
    const { result } = renderHook(() => useAsyncAction(), { wrapper });
    await act(async () => {
      await result.current.run(async () => {
        throw new ApiError('forbidden', 'denied', 403);
      }, onError);
    });
    expect(onError).toHaveBeenCalled();
    expect(showToast).not.toHaveBeenCalledWith('denied', 'error');
  });

  it('ignores non-ApiError', async () => {
    const { result } = renderHook(() => useAsyncAction(), { wrapper });
    await act(async () => {
      await result.current.run(async () => {
        throw new Error('boom');
      });
    });
    await waitFor(() => expect(result.current.busy).toBe(false));
  });
});
