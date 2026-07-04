import { useCallback, useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { apiData, apiFetch, downloadUrl } from '../api/client';
import { useAuth } from './useAuth';
import { useAsyncAction } from './useAsyncAction';
import { useToast } from './useToast';
import type { ActivityLogDTO, AdminUserDTO, BackupStatusDTO } from '../types/dto';

export type AdminTab = 'users' | 'logs' | 'ranking' | 'backup';

export const ADMIN_TABS: { id: AdminTab; label: string }[] = [
  { id: 'users', label: 'ユーザー' },
  { id: 'logs', label: 'ログ' },
  { id: 'ranking', label: 'ランキング' },
  { id: 'backup', label: 'バックアップ' },
];

export function useAdminPage() {
  const navigate = useNavigate();
  const { user } = useAuth();
  const { showToast } = useToast();
  const { busy, run } = useAsyncAction();

  const [tab, setTab] = useState<AdminTab>('users');
  const [users, setUsers] = useState<AdminUserDTO[]>([]);
  const [searchQ, setSearchQ] = useState('');
  const [logs, setLogs] = useState<ActivityLogDTO[]>([]);
  const [logType, setLogType] = useState('');
  const [logTypes, setLogTypes] = useState<string[]>([]);
  const [backup, setBackup] = useState<BackupStatusDTO | null>(null);

  useEffect(() => {
    if (user && user.role !== 'master') {
      navigate('/matching');
    }
  }, [user, navigate]);

  const searchLogs = useCallback(async () => {
    await run(async () => {
      const params = new URLSearchParams({ page: '1', limit: '20' });
      if (logType) {
        params.set('logType', logType);
      }
      const res = await apiData<{ items: ActivityLogDTO[] }>(`/admin/logs?${params.toString()}`);
      setLogs(res.items);
    });
  }, [logType, run]);

  useEffect(() => {
    if (tab === 'users') {
      apiData<{ items: AdminUserDTO[] }>('/admin/users?page=1&limit=20')
        .then((res) => setUsers(res.items))
        .catch(() => undefined);
    }
    if (tab === 'logs') {
      apiData<{ logTypes: string[] }>('/admin/logs/types')
        .then((res) => setLogTypes(res.logTypes))
        .catch(() => undefined);
      void searchLogs();
    }
    if (tab === 'backup') {
      apiData<BackupStatusDTO>('/admin/backup/status')
        .then(setBackup)
        .catch(() => undefined);
    }
  }, [tab, searchLogs]);

  const searchUsers = () =>
    run(async () => {
      const data = await apiData<AdminUserDTO[]>(`/admin/users/search?q=${encodeURIComponent(searchQ)}`);
      setUsers(data);
    });

  const deleteUser = (id: string) => {
    if (!window.confirm('このユーザーを削除しますか？')) {
      return;
    }
    void run(async () => {
      await apiFetch(`/admin/users/${id}`, { method: 'DELETE' });
      setUsers((prev) => prev.filter((userRow) => userRow.id !== id));
      showToast('ユーザーを削除しました', 'success');
    });
  };

  const downloadLogs = () => {
    const params = new URLSearchParams();
    if (logType) {
      params.set('logType', logType);
    }
    window.open(downloadUrl(`/admin/logs/download?${params.toString()}`), '_blank');
  };

  const rebuildRanking = () =>
    run(async () => {
      await apiFetch('/admin/ranking/rebuild', { method: 'POST' });
      showToast('再集計しました', 'success');
    });

  return {
    tab,
    setTab,
    busy,
    users,
    searchQ,
    setSearchQ,
    logs,
    logType,
    setLogType,
    logTypes,
    backup,
    searchUsers,
    searchLogs,
    deleteUser,
    downloadLogs,
    rebuildRanking,
  };
}
