import { useCallback, useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ApiError, apiData } from '../api/client';
import AuthenticatedLayout from '../components/layout/AuthenticatedLayout';
import { FormError, Spinner } from '../components/ui/FormField';
import { NAV } from '../constants/navigation';
import { useAuth } from '../hooks/useAuth';
import { useAsyncAction } from '../hooks/useAsyncAction';
import { useToast } from '../hooks/useToast';
import { useWebSocket } from '../hooks/useWebSocket';
import type { MatchingStatusDTO } from '../types/dto';

function matchingStatusMessage(
  connecting: boolean,
  status: MatchingStatusDTO['status'],
): string {
  if (status === 'waiting') {
    return '対戦相手を探しています…';
  }
  if (status === 'idle') {
    if (connecting) {
      return 'リアルタイム接続中…';
    }
    return 'マッチングを開始できます';
  }
  return '';
}

function applyMatchingStatus(
  res: MatchingStatusDTO,
  goToGame: (gameId: string) => void,
  setStatus: (status: MatchingStatusDTO['status']) => void,
) {
  setStatus(res.status);
  if (res.status === 'matched' && res.gameId) {
    goToGame(res.gameId);
  }
}

export default function MatchingPage() {
  const navigate = useNavigate();
  const { user } = useAuth();
  const { showToast } = useToast();
  const { subscribe, connecting } = useWebSocket();
  const { busy, run } = useAsyncAction();
  const [status, setStatus] = useState<MatchingStatusDTO['status']>('idle');

  const goToGame = useCallback(
    (gameId: string) => {
      setStatus('matched');
      navigate(`/game/${gameId}`);
    },
    [navigate],
  );

  const handleMatchingError = useCallback(
    (err: ApiError) => {
      if (err.code === 'user_in_active_game') {
        showToast(err.message, 'error');
        return;
      }
      if (err.code === 'already_in_matching') {
        showToast('既に待機中です', 'info');
        setStatus('waiting');
        return;
      }
      if (err.code === 'forbidden') {
        showToast(err.message, 'error');
        navigate('/admin');
        return;
      }
      showToast(err.message, 'error');
    },
    [navigate, showToast],
  );

  useEffect(() => {
    return subscribe((msg) => {
      if (msg.type === 'MATCHED') {
        const gameId = String(msg.data?.gameId ?? '');
        if (gameId) {
          goToGame(gameId);
        }
      }
      if (msg.type === 'RECONNECT_FAILED') {
        showToast('リアルタイム接続の再接続に失敗しました', 'error');
      }
    });
  }, [subscribe, goToGame, showToast]);

  useEffect(() => {
    apiData<MatchingStatusDTO>('/matching/status')
      .then((res) => applyMatchingStatus(res, goToGame, setStatus))
      .catch(() => undefined);
  }, [goToGame]);

  useEffect(() => {
    if (status !== 'waiting') {
      return;
    }
    const timer = window.setInterval(() => {
      void apiData<MatchingStatusDTO>('/matching/status')
        .then((res) => applyMatchingStatus(res, goToGame, setStatus))
        .catch(() => undefined);
    }, 2000);
    return () => window.clearInterval(timer);
  }, [status, goToGame]);

  const startMatching = () =>
    run(async () => {
      const res = await apiData<MatchingStatusDTO>('/matching/start', { method: 'POST' });
      applyMatchingStatus(res, goToGame, setStatus);
    }, handleMatchingError);

  const cancelMatching = () =>
    run(async () => {
      const res = await apiData<{ status: string }>('/matching/cancel', { method: 'POST' });
      setStatus(res.status as MatchingStatusDTO['status']);
    });

  const statusMessage = matchingStatusMessage(connecting, status);

  return (
    <AuthenticatedLayout links={NAV.matching}>
      <h1>マッチング</h1>
      <section className="status-panel">
        {statusMessage && <p className="status-panel__message">{statusMessage}</p>}
        {status === 'waiting' && <Spinner />}
        <div className="button-row button-row--center">
          <button type="button" className="btn-primary" onClick={() => void startMatching()} disabled={busy || status === 'waiting'}>
            {busy ? '処理中…' : 'マッチング開始'}
          </button>
          <button type="button" className="btn-secondary" onClick={() => void cancelMatching()} disabled={busy || status !== 'waiting'}>
            キャンセル
          </button>
        </div>
      </section>
      {user?.role === 'master' && <FormError message="管理者アカウントはマッチングできません" />}
    </AuthenticatedLayout>
  );
}
