import type { ActivityLogDTO } from '../../types/dto';
import DataTable from '../ui/DataTable';
import { formatDateTime } from '../../lib/format';

type Props = {
  logs: ActivityLogDTO[];
  logType: string;
  logTypes: string[];
  busy: boolean;
  onLogTypeChange: (value: string) => void;
  onSearch: () => void;
  onDownload: () => void;
};

export default function AdminLogsTab({
  logs,
  logType,
  logTypes,
  busy,
  onLogTypeChange,
  onSearch,
  onDownload,
}: Props) {
  return (
    <section className="admin-section">
      <div className="inline-form">
        <select value={logType} onChange={(event) => onLogTypeChange(event.target.value)}>
          <option value="">すべて</option>
          {logTypes.map((type) => (
            <option key={type} value={type}>
              {type}
            </option>
          ))}
        </select>
        <button type="button" className="btn-secondary" onClick={onSearch} disabled={busy}>
          検索
        </button>
        <button type="button" className="btn-secondary" onClick={onDownload}>
          CSV DL
        </button>
      </div>
      <DataTable
        columns={[
          { key: 'logType', header: '種別', render: (row) => row.logType },
          { key: 'detail', header: '詳細', render: (row) => row.detail },
          { key: 'createdAt', header: '日時', render: (row) => formatDateTime(row.createdAt) },
        ]}
        rows={logs}
        rowKey={(row) => row.id}
      />
    </section>
  );
}
