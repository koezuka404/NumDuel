package usecase

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/repository"
)

func findUserByEmailActive(ctx context.Context, repo repository.Repos, email string) (*model.User, error) {
	user, err := repo.User.FindByEmail(ctx, email)
	if err != nil || user == nil || user.IsDeleted() {
		return nil, err
	}
	return user, nil
}

func emailOrUsernameExists(ctx context.Context, repo repository.Repos, email, username string) (bool, error) {
	byEmail, err := repo.User.FindByEmail(ctx, email)
	if err != nil {
		return false, err
	}
	if byEmail != nil {
		return true, nil
	}
	byUsername, err := repo.User.FindByUsername(ctx, username)
	if err != nil {
		return false, err
	}
	return byUsername != nil, nil
}

type rankingRebuildRow struct {
	UserID   uuid.UUID
	Username string
	WinCount int
}

func listUsersForRankingRebuild(ctx context.Context, repo repository.Repos) ([]rankingRebuildRow, error) {
	users, err := repo.User.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	rows := make([]rankingRebuildRow, 0, len(users))
	for _, u := range users {
		if u.IsDeleted() || u.IsMaster() {
			continue
		}
		rows = append(rows, rankingRebuildRow{
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

func incrementUserWinCount(ctx context.Context, repos repository.Repos, userID uuid.UUID, now time.Time) error {
	user, err := repos.User.FindByID(ctx, userID)
	if err != nil || user == nil {
		return model.ErrInternal("failed to find user")
	}
	user.WinCount++
	user.UpdatedAt = now
	return repos.User.Update(ctx, user)
}

func revokeRefreshTokensByUserID(ctx context.Context, repos repository.Repos, userID uuid.UUID, now time.Time) error {
	return repos.RefreshToken.RevokeByUserID(ctx, userID, now)
}

func revokeRefreshTokenFamily(ctx context.Context, repos repository.Repos, familyID uuid.UUID, now time.Time) error {
	return repos.RefreshToken.RevokeByFamilyID(ctx, familyID, now)
}

func userWaitingInMatchingQueue(ctx context.Context, repo repository.Repos, userID uuid.UUID) (bool, error) {
	entry, err := repo.MatchingQueue.FindByUserID(ctx, userID)
	if err != nil || entry == nil {
		return false, err
	}
	return entry.Status == model.MatchingQueueWaiting, nil
}

// FindActiveGameForUser は waiting_secret / in_progress の対戦中ゲームを返す
func FindActiveGameForUser(ctx context.Context, repo repository.Repos, userID uuid.UUID) (*model.Game, error) {
	games, err := repo.Game.ListByPlayerID(ctx, userID)
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
