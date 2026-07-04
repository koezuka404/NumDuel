package usecase

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
)

//管理画面ユースケース。
type IAdminUsecase interface {
	ListUsers(ctx context.Context, page, limit int) ([]AdminUserItem, int64, error)
	SearchUsers(ctx context.Context, query string) ([]AdminUserItem, error)
	DeleteUser(ctx context.Context, adminID, targetID uuid.UUID) error
	SearchActivityLogs(ctx context.Context, logType string, userID *uuid.UUID, from, to *time.Time, page, limit int) ([]ActivityLogItem, int64, error)
	ListActivityLogTypes(ctx context.Context) ([]string, error)
	DownloadActivityLogsCSV(ctx context.Context, adminID uuid.UUID, logType string, userID *uuid.UUID, from, to *time.Time) ([]byte, error)
	RebuildRanking(ctx context.Context, adminID uuid.UUID) error
	GetBackupStatus(ctx context.Context) (*BackupStatusOutput, error)
}

type AdminUseCase struct {
	Repos        repository.Repos
	Ranking      IRankingUsecase
	WSSessions   IWSSessionStore
	ForceLogout  IForceLogoutStore
	BackupStatus IBackupStatusReader
	Locks        IDistributedLockStore
	AdminLockTTL time.Duration
	Now          func() time.Time
}

func (a *AdminUseCase) now() time.Time {
	if a != nil && a.Now != nil {
		return a.Now().UTC()
	}
	return time.Now().UTC()
}

type AdminUserItem struct {
	ID        uuid.UUID
	Username  string
	Email     string
	Role      model.Role
	WinCount  int
	DeletedAt *time.Time
	CreatedAt time.Time
}

func (a *AdminUseCase) ListUsers(ctx context.Context, page, limit int) ([]AdminUserItem, int64, error) {
	users, total, err := a.Repos.User.List(ctx, page, limit)
	if err != nil {
		return nil, 0, err
	}
	return mapAdminUsers(users), total, nil
}

func (a *AdminUseCase) SearchUsers(ctx context.Context, query string) ([]AdminUserItem, error) {
	if query == "" {
		return nil, ErrBadRequest
	}
	users, _, err := a.Repos.User.Search(ctx, query, 1, 100)
	if err != nil {
		return nil, err
	}
	return mapAdminUsers(users), nil
}

func (a *AdminUseCase) DeleteUser(ctx context.Context, adminID, targetID uuid.UUID) error {
	if err := a.acquireLock(ctx, adminUserDeleteLockKey(adminID)); err != nil {
		return err
	}
	if adminID == targetID {
		return ErrCannotDeleteSelf
	}
	target, err := a.Repos.User.FindByID(ctx, targetID)
	if err != nil {
		return err
	}
	if target == nil {
		return ErrNotFound
	}
	if target.IsDeleted() {
		return ErrUserAlreadyDeleted
	}
	if target.IsMaster() {
		return ErrCannotDeleteMaster
	}
	active, err := findActiveGameForUser(ctx, a.Repos, targetID)
	if err != nil {
		return err
	}
	if active != nil {
		return ErrUserInActiveGame
	}
	now := a.now()
	if a.ForceLogout != nil {
		if err := a.ForceLogout.SetForceLogoutBefore(ctx, targetID, now); err != nil {
			return err
		}
	}
	if a.WSSessions != nil {
		_ = a.WSSessions.DeleteUser(ctx, targetID)
	}
	if err := repository.WithTx(ctx, a.Repos.DB, func(ctx context.Context) error {
		if err := revokeRefreshTokensByUserID(ctx, a.Repos.RefreshToken, targetID, now); err != nil {
			return err
		}
		if err := a.Repos.MatchingQueue.DeleteByUserID(ctx, targetID); err != nil {
			return err
		}
		target.DeletedAt = &now
		target.DeletedBy = &adminID
		target.UpdatedAt = now
		return a.Repos.User.Update(ctx, target)
	}); err != nil {
		return err
	}
	return recordAdminDeleteUserLog(ctx, a.Repos, adminID, targetID, now)
}

type ActivityLogItem struct {
	ID        uuid.UUID
	UserID    *uuid.UUID
	LogType   string
	Detail    json.RawMessage
	CreatedAt time.Time
}

func (a *AdminUseCase) SearchActivityLogs(ctx context.Context, logType string, userID *uuid.UUID, from, to *time.Time, page, limit int) ([]ActivityLogItem, int64, error) {
	rows, total, err := a.Repos.ActivityLog.Search(ctx, logType, userID, from, to, page, limit)
	if err != nil {
		return nil, 0, err
	}
	items := make([]ActivityLogItem, len(rows))
	for i, l := range rows {
		items[i] = ActivityLogItem{
			ID: l.ID, UserID: l.UserID, LogType: l.LogType,
			Detail: l.Detail, CreatedAt: l.CreatedAt,
		}
	}
	return items, total, nil
}

func (a *AdminUseCase) ListActivityLogTypes(ctx context.Context) ([]string, error) {
	return a.Repos.ActivityLog.ListDistinctLogTypes(ctx)
}

func (a *AdminUseCase) DownloadActivityLogsCSV(ctx context.Context, adminID uuid.UUID, logType string, userID *uuid.UUID, from, to *time.Time) ([]byte, error) {
	if err := a.acquireLock(ctx, adminLogDownloadLockKey(adminID)); err != nil {
		return nil, err
	}
	rows, _, err := a.Repos.ActivityLog.Search(ctx, logType, userID, from, to, 1, 10000)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"id", "user_id", "log_type", "detail", "created_at"})
	for _, l := range rows {
		userCol := ""
		if l.UserID != nil {
			userCol = l.UserID.String()
		}
		_ = w.Write([]string{
			sanitizeCSVCell(l.ID.String()),
			sanitizeCSVCell(userCol),
			sanitizeCSVCell(l.LogType),
			sanitizeCSVCell(string(l.Detail)),
			sanitizeCSVCell(l.CreatedAt.UTC().Format(time.RFC3339)),
		})
	}
	if err := flushCSVWriter(w); err != nil {
		return nil, fmt.Errorf("failed to generate csv: %w", err)
	}
	return buf.Bytes(), nil
}

func (a *AdminUseCase) RebuildRanking(ctx context.Context, adminID uuid.UUID) error {
	if err := a.acquireLock(ctx, adminRankingRebuildLockKey(adminID)); err != nil {
		return err
	}
	if a.Ranking == nil {
		return ErrBadRequest
	}
	if err := a.Ranking.Rebuild(ctx); err != nil {
		return err
	}
	return recordAdminRebuildRankingLog(ctx, a.Repos, adminID, a.now())
}

func (a *AdminUseCase) GetBackupStatus(ctx context.Context) (*BackupStatusOutput, error) {
	if a.BackupStatus == nil {
		return &BackupStatusOutput{Status: "ok"}, nil
	}
	st, err := a.BackupStatus.GetBackupStatus(ctx)
	if err != nil {
		return nil, err
	}
	return &BackupStatusOutput{LastSyncedAt: st.LastSyncedAt, Status: st.Status}, nil
}

type BackupStatusOutput struct {
	LastSyncedAt *time.Time
	Status       string
}

var flushCSVWriter = func(w *csv.Writer) error {
	w.Flush()
	return w.Error()
}

func sanitizeCSVCell(value string) string {
	if value == "" {
		return value
	}
	switch value[0] {
	case '=', '+', '-', '@', '\t', '\r':
		return "'" + value
	default:
		return value
	}
}

func mapAdminUsers(users []*model.User) []AdminUserItem {
	out := make([]AdminUserItem, len(users))
	for i, u := range users {
		out[i] = AdminUserItem{
			ID: u.ID, Username: u.Username, Email: u.Email, Role: u.Role,
			WinCount: u.WinCount, DeletedAt: u.DeletedAt, CreatedAt: u.CreatedAt,
		}
	}
	return out
}

func NewAdminUseCase(repos repository.Repos, ranking IRankingUsecase, ws IWSSessionStore, forceLogout IForceLogoutStore, backup IBackupStatusReader, locks IDistributedLockStore, adminLockTTL time.Duration) *AdminUseCase {
	return &AdminUseCase{
		Repos:        repos,
		Ranking:      ranking,
		WSSessions:   ws,
		ForceLogout:  forceLogout,
		BackupStatus: backup,
		Locks:        locks,
		AdminLockTTL: adminLockTTL,
	}
}

