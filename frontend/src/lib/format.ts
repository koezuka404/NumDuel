export function formatDateTime(iso: string): string {
  return new Date(iso).toLocaleString('ja-JP');
}

export function truncateId(id: string, length = 8): string {
  return `${id.slice(0, length)}…`;
}
