import type { RankingItemDTO } from '../types/dto';
import DataTable from './ui/DataTable';

type Props = {
  items: RankingItemDTO[];
};

export default function RankingTable({ items }: Props) {
  if (items.length === 0) {
    return <p className="muted">ランキングデータがありません</p>;
  }

  return (
    <DataTable
      columns={[
        { key: 'rank', header: '順位', render: (row) => row.rank },
        { key: 'username', header: 'ユーザー名', render: (row) => row.username },
        { key: 'winCount', header: '勝利数', render: (row) => row.winCount },
      ]}
      rows={items}
      rowKey={(row) => String(row.rank)}
    />
  );
}
