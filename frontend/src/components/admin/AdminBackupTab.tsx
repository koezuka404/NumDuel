import { backupStatusLabel } from '../../lib/labels';
import { ProfileStat } from '../ui/FormField';
import { formatDateTime } from '../../lib/format';
import type { BackupStatusDTO } from '../../types/dto';

type Props = {
  backup: BackupStatusDTO;
};

export default function AdminBackupTab({ backup }: Props) {
  return (
    <section className="form-card">
      <ProfileStat
        label="ステータス"
        value={backupStatusLabel(backup.status)}
        valueClassName={backup.status === 'ok' ? 'backup-status-ok' : 'backup-status-error'}
      />
      <ProfileStat label="最終同期" value={backup.lastSyncedAt ? formatDateTime(backup.lastSyncedAt) : '-'} />
    </section>
  );
}
