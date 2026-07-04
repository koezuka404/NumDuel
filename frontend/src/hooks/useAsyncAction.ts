import { useCallback, useState } from 'react';
import { ApiError } from '../api/client';
import { useToast } from './useToast';

export function useAsyncAction() {
  const { showToast } = useToast();
  const [busy, setBusy] = useState(false);

  const run = useCallback(
    async (action: () => Promise<void>, onError?: (err: ApiError) => void) => {
      setBusy(true);
      try {
        await action();
      } catch (err) {
        if (err instanceof ApiError) {
          if (onError) {
            onError(err);
          } else {
            showToast(err.message, 'error');
          }
        }
      } finally {
        setBusy(false);
      }
    },
    [showToast],
  );

  return { busy, run };
}
