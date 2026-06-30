package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
)

func recordActivityLog(ctx context.Context, repo repository.Repos, userID *uuid.UUID, logType string, detail any, now time.Time) error {
	raw, err := json.Marshal(detail)
	if err != nil {
		return fmt.Errorf("failed to build activity log")
	}
	if err := repo.ActivityLog.Create(ctx, &model.ActivityLog{
		ID: uuid.New(), UserID: userID, LogType: logType,
		Detail: raw, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		return fmt.Errorf("failed to save activity log")
	}
	return nil
}

func recordGuessActivityLog(ctx context.Context, repo repository.Repos, gameID, playerID uuid.UUID, turn, hitCount int, isWin, isAuto bool, now time.Time) error {
	uid := playerID
	return recordActivityLog(ctx, repo, &uid, "guess", map[string]any{
		"gameId": gameID.String(), "playerId": playerID.String(),
		"turn": turn, "hitCount": hitCount, "isWin": isWin, "isAuto": isAuto,
	}, now)
}

func recordGameOverActivityLog(ctx context.Context, repo repository.Repos, gameID uuid.UUID, reason string, winnerID *uuid.UUID, now time.Time) error {
	detail := map[string]string{
		"gameId": gameID.String(), "reason": reason,
	}
	if winnerID != nil {
		detail["winnerId"] = winnerID.String()
	}
	return recordActivityLog(ctx, repo, nil, "game_over", detail, now)
}

func recordTimeoutActivityLog(ctx context.Context, repo repository.Repos, gameID, playerID uuid.UUID, now time.Time) error {
	uid := playerID
	return recordActivityLog(ctx, repo, &uid, "timeout", map[string]string{
		"gameId": gameID.String(), "playerId": playerID.String(),
	}, now)
}

func recordRecoverActivityLog(ctx context.Context, repo repository.Repos, gameID uuid.UUID, now time.Time) error {
	return recordActivityLog(ctx, repo, nil, "recover", map[string]string{
		"gameId": gameID.String(),
	}, now)
}

func recordAdminDeleteUserLog(ctx context.Context, repo repository.Repos, adminID, targetID uuid.UUID, now time.Time) error {
	uid := adminID
	return recordActivityLog(ctx, repo, &uid, "admin_delete_user", map[string]string{
		"adminId": adminID.String(), "targetUserId": targetID.String(),
	}, now)
}

func recordAdminRebuildRankingLog(ctx context.Context, repo repository.Repos, adminID uuid.UUID, now time.Time) error {
	uid := adminID
	return recordActivityLog(ctx, repo, &uid, "admin_rebuild_ranking", map[string]string{
		"adminId": adminID.String(),
	}, now)
}
