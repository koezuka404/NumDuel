import AdminBackupTab from '../components/admin/AdminBackupTab';
import AdminLogsTab from '../components/admin/AdminLogsTab';
import AdminRankingTab from '../components/admin/AdminRankingTab';
import AdminUsersTab from '../components/admin/AdminUsersTab';
import AuthenticatedLayout from '../components/layout/AuthenticatedLayout';
import TabBar from '../components/ui/TabBar';
import { ADMIN_TABS, useAdminPage } from '../hooks/useAdminPage';

export default function AdminPage() {
  const admin = useAdminPage();

  return (
    <AuthenticatedLayout>
      <h1>管理画面</h1>
      <TabBar tabs={ADMIN_TABS} active={admin.tab} onChange={admin.setTab} />

      {admin.tab === 'users' && (
        <AdminUsersTab
          users={admin.users}
          searchQ={admin.searchQ}
          busy={admin.busy}
          onSearchQChange={admin.setSearchQ}
          onSearch={() => void admin.searchUsers()}
          onDelete={admin.deleteUser}
        />
      )}

      {admin.tab === 'logs' && (
        <AdminLogsTab
          logs={admin.logs}
          logType={admin.logType}
          logTypes={admin.logTypes}
          busy={admin.busy}
          onLogTypeChange={admin.setLogType}
          onSearch={() => void admin.searchLogs()}
          onDownload={admin.downloadLogs}
        />
      )}

      {admin.tab === 'ranking' && (
        <AdminRankingTab busy={admin.busy} onRebuild={() => void admin.rebuildRanking()} />
      )}

      {admin.tab === 'backup' && admin.backup && <AdminBackupTab backup={admin.backup} />}
    </AuthenticatedLayout>
  );
}
