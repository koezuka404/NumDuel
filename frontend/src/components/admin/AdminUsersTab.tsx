import type { FormEvent } from 'react';
import type { AdminUserDTO } from '../../types/dto';
import { userRoleLabel } from '../../lib/labels';
import DataTable from '../ui/DataTable';

type Props = {
  users: AdminUserDTO[];
  searchQ: string;
  busy: boolean;
  onSearchQChange: (value: string) => void;
  onSearch: () => void;
  onDelete: (id: string) => void;
};

export default function AdminUsersTab({ users, searchQ, busy, onSearchQChange, onSearch, onDelete }: Props) {
  const handleSubmit = (event: FormEvent) => {
    event.preventDefault();
    onSearch();
  };

  return (
    <section className="admin-section">
      <form className="inline-form" onSubmit={handleSubmit}>
        <input
          value={searchQ}
          onChange={(event) => onSearchQChange(event.target.value)}
          placeholder="ユーザー名・メールで検索"
        />
        <button type="submit" className="btn-secondary" disabled={busy}>
          検索
        </button>
      </form>
      <DataTable
        columns={[
          { key: 'username', header: 'ユーザー名', render: (row) => row.username },
          { key: 'email', header: 'メール', render: (row) => row.email },
          { key: 'role', header: 'ロール', render: (row) => userRoleLabel(row.role) },
          { key: 'winCount', header: '勝利数', render: (row) => row.winCount },
          {
            key: 'actions',
            header: '操作',
            render: (row) => (
              <button
                type="button"
                className="btn-danger"
                disabled={busy || row.deletedAt !== null}
                onClick={() => onDelete(row.id)}
              >
                削除
              </button>
            ),
          },
        ]}
        rows={users}
        rowKey={(row) => row.id}
      />
    </section>
  );
}
