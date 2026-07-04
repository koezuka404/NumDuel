import { fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
import AdminBackupTab from './admin/AdminBackupTab';
import AdminLogsTab from './admin/AdminLogsTab';
import AdminRankingTab from './admin/AdminRankingTab';
import AdminUsersTab from './admin/AdminUsersTab';
import AppHeader from './layout/AppHeader';
import AppShell from './layout/AppShell';
import AuthenticatedLayout from './layout/AuthenticatedLayout';
import DataTable from './ui/DataTable';
import FormField, { FormError, PageHeader, ProfileStat, Spinner } from './ui/FormField';
import TabBar from './ui/TabBar';

vi.mock('../hooks/useAuth', () => ({
  useAuth: () => ({ user: { id: '1', username: 'alice', role: 'user' } }),
}));

vi.mock('../hooks/useLogout', () => ({
  useLogout: () => vi.fn(),
}));

describe('FormField components', () => {
  it('renders FormField and handles change', () => {
    const onChange = vi.fn();
    render(<FormField label="Name" value="" onChange={onChange} required />);
    fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'bob' } });
    expect(onChange).toHaveBeenCalledWith('bob');
  });

  it('renders ProfileStat with and without valueClassName', () => {
    const { rerender } = render(<ProfileStat label="Rank" value="1" />);
    expect(screen.getByText('Rank')).toBeInTheDocument();
    rerender(<ProfileStat label="Rank" value="1" valueClassName="rank" />);
    expect(document.querySelector('.rank')).toBeTruthy();
  });

  it('renders FormError, Spinner, and PageHeader variants', () => {
    render(
      <>
        <FormError message="error" />
        <Spinner label="loading" className="spinner--light" />
        <PageHeader title="Title" action={<button type="button">Action</button>} />
      </>,
    );
    expect(screen.getByText('error')).toBeInTheDocument();
    expect(screen.getByLabelText('loading')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Action' })).toBeInTheDocument();

    render(<PageHeader title="Only Title" />);
    expect(screen.getByRole('heading', { name: 'Only Title' })).toBeInTheDocument();
  });
});

describe('DataTable', () => {
  it('renders headers and rows', () => {
    render(
      <DataTable
        columns={[{ key: 'name', header: 'Name', render: (row: { name: string }) => row.name }]}
        rows={[{ name: 'alice' }]}
        rowKey={(row) => row.name}
      />,
    );
    expect(screen.getByText('Name')).toBeInTheDocument();
    expect(screen.getByText('alice')).toBeInTheDocument();
  });
});

describe('TabBar', () => {
  it('calls onChange when tab clicked', () => {
    const onChange = vi.fn();
    render(
      <TabBar
        tabs={[
          { id: 'a', label: 'A' },
          { id: 'b', label: 'B' },
        ]}
        active="a"
        onChange={onChange}
      />,
    );
    fireEvent.click(screen.getByRole('button', { name: 'B' }));
    expect(onChange).toHaveBeenCalledWith('b');
  });
});

describe('AppShell', () => {
  it('renders default brand header and narrow main', () => {
    render(
      <AppShell narrow>
        <p>content</p>
      </AppShell>,
    );
    expect(screen.getByText('NumDuel')).toBeInTheDocument();
    expect(screen.getByText('content')).toBeInTheDocument();
    expect(document.querySelector('.app-shell__main--narrow')).toBeTruthy();
  });

  it('renders custom header', () => {
    render(<AppShell header={<header>Custom</header>}>content</AppShell>);
    expect(screen.getByText('Custom')).toBeInTheDocument();
  });
});

describe('AppHeader', () => {
  it('renders links, username, and logout', () => {
    const onLogout = vi.fn();
    render(
      <MemoryRouter>
        <AppHeader links={[{ to: '/ranking', label: 'Ranking' }]} onLogout={onLogout} />
      </MemoryRouter>,
    );
    expect(screen.getByText('Ranking')).toBeInTheDocument();
    expect(screen.getByText('alice')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'ログアウト' }));
    expect(onLogout).toHaveBeenCalled();
  });
});

describe('AuthenticatedLayout', () => {
  it('renders children with header', () => {
    render(
      <MemoryRouter>
        <AuthenticatedLayout links={[{ to: '/profile', label: 'Profile' }]}>
          <p>child</p>
        </AuthenticatedLayout>
      </MemoryRouter>,
    );
    expect(screen.getByText('child')).toBeInTheDocument();
    expect(screen.getByText('Profile')).toBeInTheDocument();
  });
});

describe('Admin tabs', () => {
  it('AdminUsersTab searches and deletes', () => {
    const onSearch = vi.fn();
    const onDelete = vi.fn();
    render(
      <AdminUsersTab
        users={[
          {
            id: 'u1',
            username: 'alice',
            email: 'a@test.local',
            role: 'user',
            winCount: 1,
            deletedAt: null,
          },
          {
            id: 'u2',
            username: 'deleted',
            email: 'd@test.local',
            role: 'user',
            winCount: 0,
            deletedAt: '2024-01-01T00:00:00Z',
          },
        ]}
        searchQ="alice"
        busy={false}
        onSearchQChange={vi.fn()}
        onSearch={onSearch}
        onDelete={onDelete}
      />,
    );
    fireEvent.submit(screen.getByRole('button', { name: '検索' }).closest('form')!);
    expect(onSearch).toHaveBeenCalled();
    fireEvent.click(screen.getByRole('button', { name: '削除' }));
    expect(onDelete).toHaveBeenCalledWith('u1');
    expect(screen.getByRole('button', { name: '削除' })).not.toBeDisabled();
  });

  it('AdminLogsTab filters and downloads', () => {
    const onSearch = vi.fn();
    const onDownload = vi.fn();
    render(
      <AdminLogsTab
        logs={[{ id: 'l1', logType: 'guess', detail: 'detail', createdAt: '2024-01-01T00:00:00Z' }]}
        logType=""
        logTypes={['guess']}
        busy={false}
        onLogTypeChange={vi.fn()}
        onSearch={onSearch}
        onDownload={onDownload}
      />,
    );
    fireEvent.click(screen.getByRole('button', { name: '検索' }));
    fireEvent.click(screen.getByRole('button', { name: 'CSVダウンロード' }));
    expect(onSearch).toHaveBeenCalled();
    expect(onDownload).toHaveBeenCalled();
  });

  it('AdminRankingTab rebuilds ranking', () => {
    const onRebuild = vi.fn();
    render(<AdminRankingTab busy={false} onRebuild={onRebuild} />);
    fireEvent.click(screen.getByRole('button', { name: 'ランキング再集計' }));
    expect(onRebuild).toHaveBeenCalled();
  });

  it('AdminBackupTab shows ok and error statuses', () => {
    const { rerender } = render(
      <AdminBackupTab backup={{ status: 'ok', lastSyncedAt: '2024-01-01T00:00:00Z' }} />,
    );
    expect(screen.getByText('正常')).toBeInTheDocument();

    rerender(<AdminBackupTab backup={{ status: 'error', lastSyncedAt: null }} />);
    expect(screen.getByText('エラー')).toBeInTheDocument();
    expect(screen.getByText('-')).toBeInTheDocument();
  });
});
