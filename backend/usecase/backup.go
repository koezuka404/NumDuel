package usecase

import (
	"context"
	"time"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
)

const defaultBackupMaxRetries = 3

// BackupDeps は BackupWorker / RunBackupSync の依存
type BackupDeps struct {
	Syncer       *repository.BackupSyncer
	BackupStatus model.BackupStatusStore
	MaxRetries   int
	Now          func() time.Time
}

// RunBackupSync は primary → backup への差分 UPSERT を実行する（§12.8）
func RunBackupSync(ctx context.Context, d BackupDeps) error {
	if d.Syncer == nil {
		return nil
	}
	maxRetries := d.MaxRetries
	if maxRetries <= 0 {
		maxRetries = defaultBackupMaxRetries
	}

	var lastSyncedAt *time.Time
	if d.BackupStatus != nil {
		st, err := d.BackupStatus.GetBackupStatus(ctx)
		if err != nil {
			return err
		}
		if st != nil {
			lastSyncedAt = st.LastSyncedAt
		}
	}

	var syncErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if _, syncErr = d.Syncer.Sync(ctx, lastSyncedAt); syncErr == nil {
			if d.BackupStatus != nil {
				return d.BackupStatus.SetBackupStatus(ctx, "ok", d.now())
			}
			return nil
		}
	}

	if d.BackupStatus != nil {
		preserved := time.Time{}
		if st, err := d.BackupStatus.GetBackupStatus(ctx); err == nil && st != nil && st.LastSyncedAt != nil {
			preserved = *st.LastSyncedAt
		}
		_ = d.BackupStatus.SetBackupStatus(ctx, "error", preserved)
	}
	return syncErr
}

func (d BackupDeps) now() time.Time {
	if d.Now != nil {
		return d.Now().UTC()
	}
	return time.Now().UTC()
}
