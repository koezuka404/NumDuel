import { useCallback, useEffect, useState } from 'react';
import { ApiError, apiData } from '../api/client';
import AuthenticatedLayout from '../components/layout/AuthenticatedLayout';
import RankingFooter from '../components/RankingFooter';
import RankingTable from '../components/RankingTable';
import { FormError, PageHeader, Spinner } from '../components/ui/FormField';
import { NAV } from '../constants/navigation';
import type { RankingItemDTO } from '../types/dto';

export default function RankingPage() {
  const [items, setItems] = useState<RankingItemDTO[]>([]);
  const [updatedAt, setUpdatedAt] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const fetchRanking = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const data = await apiData<RankingItemDTO[]>('/ranking');
      setItems(data);
      setUpdatedAt(new Date().toLocaleString('ja-JP'));
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      }
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void fetchRanking();
  }, [fetchRanking]);

  return (
    <AuthenticatedLayout links={NAV.ranking}>
      <PageHeader
        title="ランキング"
        action={
          <button type="button" className="btn-secondary" onClick={() => void fetchRanking()} disabled={loading}>
            再読み込み
          </button>
        }
      />
      {loading && <Spinner />}
      {error && <FormError message={error} />}
      {!loading && !error && (
        <>
          <RankingTable items={items} />
          <RankingFooter updatedAt={updatedAt} />
        </>
      )}
    </AuthenticatedLayout>
  );
}
