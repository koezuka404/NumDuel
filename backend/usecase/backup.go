package usecase

import (
	"context"
	"time"

	"github.com/numduel/numduel/repository"
)

type BackupStatus struct {
	Status       string
	LastSyncedAt *time.Time
}

// バックアップ状態の読み取り。
type IBackupStatusReader interface {
	GetBackupStatus(ctx context.Context) (*BackupStatus, error)
}

// バックアップ状態の Redis 管理。
type IBackupStatusStore interface {
	IBackupStatusReader
	SetBackupStatus(ctx context.Context, status string, lastSyncedAt time.Time) error
}

// DB バックアップ同期ユースケース。
type IBackupUsecase interface {
	RunSync(ctx context.Context) error
}

const defaultBackupMaxRetries = 3

type BackupUseCase struct {
	Syncer       *repository.BackupSyncer
	BackupStatus IBackupStatusStore
	MaxRetries   int
	Now          func() time.Time
}

func NewBackupUseCase(syncer *repository.BackupSyncer, backup IBackupStatusStore, maxRetries int) *BackupUseCase {
	return &BackupUseCase{Syncer: syncer, BackupStatus: backup, MaxRetries: maxRetries}
}

func (b *BackupUseCase) now() time.Time {
	if b != nil && b.Now != nil {
		return b.Now().UTC()
	}
	return time.Now().UTC()
}

func (b *BackupUseCase) RunSync(ctx context.Context) error {
	if b.Syncer == nil {
		return nil
	}
	maxRetries := b.MaxRetries
	if maxRetries <= 0 {
		maxRetries = defaultBackupMaxRetries
	}

	var lastSyncedAt *time.Time
	if b.BackupStatus != nil {
		st, err := b.BackupStatus.GetBackupStatus(ctx)
		if err != nil {
			return err
		}
		if st != nil {
			lastSyncedAt = st.LastSyncedAt
		}
	}

	var syncErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if _, syncErr = b.Syncer.Sync(ctx, lastSyncedAt); syncErr == nil {
			if b.BackupStatus != nil {
				return b.BackupStatus.SetBackupStatus(ctx, "ok", b.now())
			}
			return nil
		}
	}

	if b.BackupStatus != nil {
		preserved := time.Time{}
		if st, err := b.BackupStatus.GetBackupStatus(ctx); err == nil && st != nil && st.LastSyncedAt != nil {
			preserved = *st.LastSyncedAt
		}
		_ = b.BackupStatus.SetBackupStatus(ctx, "error", preserved)
	}
	return syncErr
}
