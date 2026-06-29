package usecase

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
)

func findUserByEmailActive(ctx context.Context, repo model.Repository, email string) (*model.User, error) {
	user, err := repo.Users().FindByEmail(ctx, email)
	if err != nil || user == nil || user.IsDeleted() {
		return nil, err
	}
	return user, nil
}

func emailOrUsernameExists(ctx context.Context, repo model.Repository, email, username string) (bool, error) {
	byEmail, err := repo.Users().FindByEmail(ctx, email)
	if err != nil {
		return false, err
	}
	if byEmail != nil {
		return true, nil
	}
	byUsername, err := repo.Users().FindByUsername(ctx, username)
	if err != nil {
		return false, err
	}
	return byUsername != nil, nil
}

func listUsersForRankingRebuild(ctx context.Context, repo model.Repository) ([]model.RankingRebuildRow, error) {
	users, err := repo.Users().ListAll(ctx)
	if err != nil {
		return nil, err
	}
	rows := make([]model.RankingRebuildRow, 0, len(users))
	for _, u := range users {
		if u.IsDeleted() || u.IsMaster() {
			continue
		}
		rows = append(rows, model.RankingRebuildRow{
			UserID:   u.ID,
			Username: u.Username,
			WinCount: u.WinCount,
		})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].WinCount != rows[j].WinCount {
			return rows[i].WinCount > rows[j].WinCount
		}
		return rows[i].Username < rows[j].Username
	})
	return rows, nil
}

func incrementUserWinCount(ctx context.Context, repo model.Repository, tx model.Transaction, userID uuid.UUID, now time.Time) error {
	user, err := repo.Users().FindByID(ctx, userID)
	if err != nil || user == nil {
		return model.ErrInternal("failed to find user")
	}
	user.WinCount++
	user.UpdatedAt = now
	return repo.Users().Update(ctx, tx, user)
}

func revokeRefreshTokensByUserID(ctx context.Context, repo model.Repository, tx model.Transaction, userID uuid.UUID, now time.Time) error {
	return repo.RefreshTokens().RevokeByUserID(ctx, tx, userID, now)
}

func revokeRefreshTokenFamily(ctx context.Context, repo model.Repository, tx model.Transaction, familyID uuid.UUID, now time.Time) error {
	return repo.RefreshTokens().RevokeByFamilyID(ctx, tx, familyID, now)
}

func userWaitingInMatchingQueue(ctx context.Context, repo model.Repository, userID uuid.UUID) (bool, error) {
	entry, err := repo.MatchingQueue().FindByUserID(ctx, userID)
	if err != nil || entry == nil {
		return false, err
	}
	return entry.Status == model.MatchingQueueWaiting, nil
}

// FindActiveGameForUser は waiting_secret / in_progress の対戦中ゲームを返す。
func FindActiveGameForUser(ctx context.Context, repo model.Repository, userID uuid.UUID) (*model.Game, error) {
	games, err := repo.Games().ListByPlayerID(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, game := range games {
		if game.Status == model.GameStatusWaitingSecret || game.Status == model.GameStatusInProgress {
			return game, nil
		}
	}
	return nil, nil
}
