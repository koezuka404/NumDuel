package usecase

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
)

type AdminDeps struct {
	Repo          repository.IRepository
	Tx            repository.TxManager
	WSSessions    model.WSSessionStore
	ForceLogout   model.ForceLogoutStore
	BackupStatus  model.BackupStatusStore
	Locks         model.GameLockStore // admin:{adminId}:*_lock（§13.10.2）
	AdminLockTTL  time.Duration
	Now           func() time.Time
}

func (d AdminDeps) now() time.Time {
	if d.Now != nil {
		return d.Now()
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

func GetAdminUsers(ctx context.Context, d AdminDeps, page, limit int) ([]AdminUserItem, int64, error) {
	users, total, err := d.Repo.Users().List(ctx, page, limit)
	if err != nil {
		return nil, 0, model.ErrInternal("failed to list users")
	}
	return mapAdminUsers(users), total, nil
}

func SearchAdminUsers(ctx context.Context, d AdminDeps, query string) ([]AdminUserItem, error) {
	if query == "" {
		return nil, model.ErrValidation("query is required")
	}
	users, _, err := d.Repo.Users().Search(ctx, query, 1, 100)
	if err != nil {
		return nil, model.ErrInternal("failed to search users")
	}
	return mapAdminUsers(users), nil
}

// DeleteUser は master によるユーザーの論理削除
// force_logout_before SET → WS 切断 → refresh 失効 → matching キュー削除 → users.deleted_at 更新 → activity_logs
func DeleteUser(ctx context.Context, d AdminDeps, adminID, targetID uuid.UUID) error {
	if err := acquireAdminLock(ctx, d, adminUserDeleteLockKey(adminID)); err != nil {
		return err
	}
	if adminID == targetID {
		return model.ErrCannotDeleteSelf()
	}
	target, err := d.Repo.Users().FindByID(ctx, targetID)
	if err != nil {
		return model.ErrInternal("failed to find user")
	}
	if target == nil {
		return model.ErrNotFound("user not found")
	}
	if target.IsDeleted() {
		return model.ErrUserAlreadyDeleted()
	}
	if target.IsMaster() {
		return model.ErrCannotDeleteMaster()
	}
	active, err := FindActiveGameForUser(ctx, d.Repo, targetID)
	if err != nil {
		return model.ErrInternal("failed to check active game")
	}
	if active != nil {
		return model.ErrUserInActiveGame()
	}
	now := d.now()
	if d.ForceLogout != nil {
		if err := d.ForceLogout.SetForceLogoutBefore(ctx, targetID, now); err != nil {
			return model.ErrInternal("failed to set force logout")
		}
	}
	if d.WSSessions != nil {
		_ = d.WSSessions.DeleteUser(ctx, targetID)
	}
	if err := d.Tx.WithinTx(ctx, func(ctx context.Context, tx repository.ITxRepos) error {
		if err := revokeRefreshTokensByUserID(ctx, tx, targetID, now); err != nil {
			return model.ErrInternal("failed to revoke refresh tokens")
		}
		if err := tx.MatchingQueue().DeleteByUserID(ctx, targetID); err != nil {
			return model.ErrInternal("failed to remove matching queue")
		}
		target.DeletedAt = &now
		target.DeletedBy = &adminID
		target.UpdatedAt = now
		if err := tx.Users().Update(ctx, target); err != nil {
			return model.ErrInternal("failed to delete user")
		}
		return nil
	}); err != nil {
		return err
	}
	return recordAdminDeleteUserLog(ctx, d.Repo, adminID, targetID, now)
}

type ActivityLogItem struct {
	ID        uuid.UUID
	UserID    *uuid.UUID
	LogType   string
	Detail    json.RawMessage
	CreatedAt time.Time
}

func SearchActivityLogs(ctx context.Context, d AdminDeps, logType string, userID *uuid.UUID, from, to *time.Time, page, limit int) ([]ActivityLogItem, int64, error) {
	rows, total, err := d.Repo.ActivityLogs().Search(ctx, logType, userID, from, to, page, limit)
	if err != nil {
		return nil, 0, model.ErrInternal("failed to search activity logs")
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

func ListActivityLogTypes(ctx context.Context, d AdminDeps) ([]string, error) {
	types, err := d.Repo.ActivityLogs().ListDistinctLogTypes(ctx)
	if err != nil {
		return nil, model.ErrInternal("failed to list log types")
	}
	return types, nil
}

func DownloadActivityLogsCSV(ctx context.Context, d AdminDeps, adminID uuid.UUID, logType string, userID *uuid.UUID, from, to *time.Time) ([]byte, error) {
	if err := acquireAdminLock(ctx, d, adminLogDownloadLockKey(adminID)); err != nil {
		return nil, err
	}
	rows, _, err := d.Repo.ActivityLogs().Search(ctx, logType, userID, from, to, 1, 10000)
	if err != nil {
		return nil, model.ErrInternal("failed to search activity logs")
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
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, model.ErrInternal("failed to generate csv")
	}
	return buf.Bytes(), nil
}

func RebuildRankingAsAdmin(ctx context.Context, d AdminDeps, adminID uuid.UUID) error {
	if err := acquireAdminLock(ctx, d, adminRankingRebuildLockKey(adminID)); err != nil {
		return err
	}
	if err := RebuildRanking(ctx, RankingDeps{Repo: d.Repo, Tx: d.Tx, Now: d.Now}); err != nil {
		return err
	}
	return recordAdminRebuildRankingLog(ctx, d.Repo, adminID, d.now())
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

type BackupStatusOutput struct {
	LastSyncedAt *time.Time
	Status       string
}

func GetBackupStatus(ctx context.Context, d AdminDeps) (*BackupStatusOutput, error) {
	if d.BackupStatus == nil {
		return &BackupStatusOutput{Status: "ok"}, nil
	}
	st, err := d.BackupStatus.GetBackupStatus(ctx)
	if err != nil {
		return nil, model.ErrInternal("failed to read backup status")
	}
	return &BackupStatusOutput{LastSyncedAt: st.LastSyncedAt, Status: st.Status}, nil
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

func ParseOptionalUUID(raw string) (*uuid.UUID, error) {
	if raw == "" {
		return nil, nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil, model.ErrValidation("invalid user id")
	}
	return &id, nil
}

func ParseOptionalTime(raw string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, model.ErrValidation("invalid time format")
	}
	return &t, nil
}
