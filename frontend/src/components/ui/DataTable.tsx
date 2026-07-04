import type { ReactNode } from 'react';

export type TableColumn<T> = {
  key: string;
  header: string;
  render: (row: T) => ReactNode;
};

type Props<T> = {
  columns: TableColumn<T>[];
  rows: T[];
  rowKey: (row: T, index: number) => string;
};

export default function DataTable<T>({ columns, rows, rowKey }: Props<T>) {
  return (
    <table className="data-table">
      <thead>
        <tr>
          {columns.map((column) => (
            <th key={column.key}>{column.header}</th>
          ))}
        </tr>
      </thead>
      <tbody>
        {rows.map((row, index) => (
          <tr key={rowKey(row, index)}>
            {columns.map((column) => (
              <td key={column.key}>{column.render(row)}</td>
            ))}
          </tr>
        ))}
      </tbody>
    </table>
  );
}
