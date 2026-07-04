import { useEffect, useState } from 'react';
import { apiData } from '../api/client';
import { NAV } from '../constants/navigation';
import AuthenticatedLayout from '../components/layout/AuthenticatedLayout';
import DataTable, { type TableColumn } from '../components/ui/DataTable';
import { ProfileStat, Spinner } from '../components/ui/FormField';
import TabBar from '../components/ui/TabBar';
import { formatDateTime, truncateId } from '../lib/format';
import type {
  LoginHistoryItemDTO,
  MatchHistoryItemDTO,
  ProfileDTO,
  WSHistoryItemDTO,
} from '../types/dto';

type ProfileTab = 'matches' | 'logins' | 'ws';

const PROFILE_TABS: { id: ProfileTab; label: string }[] = [
  { id: 'matches', label: '勝敗履歴' },
  { id: 'logins', label: 'ログイン履歴' },
  { id: 'ws', label: 'WS接続履歴' },
];

const MATCH_COLUMNS: TableColumn<MatchHistoryItemDTO>[] = [
  { key: 'gameId', header: 'ゲームID', render: (row) => truncateId(row.gameId) },
  { key: 'winner', header: '勝者', render: (row) => row.winnerUsername },
  { key: 'loser', header: '敗者', render: (row) => row.loserUsername },
  { key: 'finishedAt', header: '終了日時', render: (row) => formatDateTime(row.finishedAt) },
];

const LOGIN_COLUMNS: TableColumn<LoginHistoryItemDTO>[] = [
  { key: 'action', header: '操作', render: (row) => row.action },
  { key: 'createdAt', header: '日時', render: (row) => formatDateTime(row.createdAt) },
];

const WS_COLUMNS: TableColumn<WSHistoryItemDTO>[] = [
  { key: 'connectionId', header: '接続ID', render: (row) => truncateId(row.connectionId) },
  { key: 'connectedAt', header: '接続', render: (row) => formatDateTime(row.connectedAt) },
  {
    key: 'disconnectedAt',
    header: '切断',
    render: (row) => (row.disconnectedAt ? formatDateTime(row.disconnectedAt) : '-'),
  },
];

export default function ProfilePage() {
  const [profile, setProfile] = useState<ProfileDTO | null>(null);
  const [tab, setTab] = useState<ProfileTab>('matches');
  const [matches, setMatches] = useState<MatchHistoryItemDTO[]>([]);
  const [logins, setLogins] = useState<LoginHistoryItemDTO[]>([]);
  const [wsLogs, setWsLogs] = useState<WSHistoryItemDTO[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    apiData<ProfileDTO>('/me/profile')
      .then(setProfile)
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    if (tab === 'matches') {
      apiData<{ items: MatchHistoryItemDTO[] }>('/me/match-history?page=1&limit=20').then((res) =>
        setMatches(res.items),
      );
    }
    if (tab === 'logins') {
      apiData<{ items: LoginHistoryItemDTO[] }>('/me/login-history?page=1&limit=20').then((res) =>
        setLogins(res.items),
      );
    }
    if (tab === 'ws') {
      apiData<{ items: WSHistoryItemDTO[] }>('/me/ws-history?page=1&limit=20').then((res) => setWsLogs(res.items));
    }
  }, [tab]);

  return (
    <AuthenticatedLayout links={NAV.profile}>
      <h1>プロフィール</h1>
      {loading && <Spinner />}
      {profile && (
        <section className="profile-summary">
          <ProfileStat label="ユーザー名" value={profile.username} />
          <ProfileStat label="勝利数" value={profile.winCount} />
          <ProfileStat label="ランキング" value={profile.rank ?? '圏外'} />
        </section>
      )}

      <TabBar tabs={PROFILE_TABS} active={tab} onChange={setTab} />

      {tab === 'matches' && (
        <DataTable columns={MATCH_COLUMNS} rows={matches} rowKey={(row) => row.gameId} />
      )}
      {tab === 'logins' && (
        <DataTable columns={LOGIN_COLUMNS} rows={logins} rowKey={(row, index) => `${row.action}-${row.createdAt}-${index}`} />
      )}
      {tab === 'ws' && (
        <DataTable columns={WS_COLUMNS} rows={wsLogs} rowKey={(row) => row.connectionId} />
      )}
    </AuthenticatedLayout>
  );
}
