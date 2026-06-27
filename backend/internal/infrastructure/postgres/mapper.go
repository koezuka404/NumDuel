package postgres

import (
	"encoding/json"
	"fmt"

	"gorm.io/datatypes"

	"github.com/numduel/numduel/internal/domain"
)

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func strVal(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func toUser(m *userModel) *domain.User {
	return &domain.User{
		ID:             m.ID,
		Username:       m.Username,
		Email:          m.Email,
		PasswordHash:   m.Password,
		Role:           domain.Role(m.Role),
		WinCount:       m.WinCount,
		DeletedAt:      m.DeletedAt,
		DeletedBy:      m.DeletedBy,
		LastActivityAt: m.LastActivityAt,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

func fromUser(u *domain.User) *userModel {
	return &userModel{
		ID:             u.ID,
		Username:       u.Username,
		Email:          u.Email,
		Password:       u.PasswordHash,
		Role:           string(u.Role),
		WinCount:       u.WinCount,
		DeletedAt:      u.DeletedAt,
		DeletedBy:      u.DeletedBy,
		LastActivityAt: u.LastActivityAt,
		CreatedAt:      u.CreatedAt,
		UpdatedAt:      u.UpdatedAt,
	}
}

func toGame(m *gameModel) *domain.Game {
	return &domain.Game{
		ID:                  m.ID,
		Status:              domain.GameStatus(m.Status),
		Player1ID:           m.Player1ID,
		Player2ID:           m.Player2ID,
		Player1Secret:       strVal(m.Player1Secret),
		Player2Secret:       strVal(m.Player2Secret),
		CurrentTurnPlayerID: m.CurrentTurnPlayerID,
		CurrentTurn:         m.CurrentTurn,
		WinnerID:            m.WinnerID,
		StartedAt:           m.StartedAt,
		FinishedAt:          m.FinishedAt,
		CreatedAt:           m.CreatedAt,
		UpdatedAt:           m.UpdatedAt,
	}
}

func fromGame(g *domain.Game) *gameModel {
	return &gameModel{
		ID:                  g.ID,
		Status:              string(g.Status),
		Player1ID:           g.Player1ID,
		Player2ID:           g.Player2ID,
		Player1Secret:       strPtr(g.Player1Secret),
		Player2Secret:       strPtr(g.Player2Secret),
		CurrentTurnPlayerID: g.CurrentTurnPlayerID,
		CurrentTurn:         g.CurrentTurn,
		WinnerID:            g.WinnerID,
		StartedAt:           g.StartedAt,
		FinishedAt:          g.FinishedAt,
		CreatedAt:           g.CreatedAt,
		UpdatedAt:           g.UpdatedAt,
	}
}

func digitResultsToJSON(results []domain.DigitResult) (datatypes.JSON, error) {
	ints := make([]int, len(results))
	for i, r := range results {
		ints[i] = int(r)
	}
	b, err := json.Marshal(ints)
	if err != nil {
		return nil, err
	}
	return datatypes.JSON(b), nil
}

func digitResultsFromJSON(raw datatypes.JSON) ([]domain.DigitResult, error) {
	var ints []int
	if err := json.Unmarshal(raw, &ints); err != nil {
		return nil, err
	}
	out := make([]domain.DigitResult, len(ints))
	for i, v := range ints {
		out[i] = domain.DigitResult(v)
	}
	return out, nil
}

func toGuess(m *guessModel) (domain.Guess, error) {
	results, err := digitResultsFromJSON(m.DigitResults)
	if err != nil {
		return domain.Guess{}, fmt.Errorf("decode digit_results: %w", err)
	}
	return domain.Guess{
		ID:           m.ID,
		GameID:       m.GameID,
		PlayerID:     m.PlayerID,
		Turn:         m.Turn,
		GuessNumber:  m.GuessNumber,
		DigitResults: results,
		HitCount:     m.HitCount,
		IsAuto:       m.IsAuto,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}, nil
}

func fromGuess(g *domain.Guess) (*guessModel, error) {
	raw, err := digitResultsToJSON(g.DigitResults)
	if err != nil {
		return nil, err
	}
	return &guessModel{
		ID:           g.ID,
		GameID:       g.GameID,
		PlayerID:     g.PlayerID,
		Turn:         g.Turn,
		GuessNumber:  g.GuessNumber,
		DigitResults: raw,
		HitCount:     g.HitCount,
		IsAuto:       g.IsAuto,
		CreatedAt:    g.CreatedAt,
		UpdatedAt:    g.UpdatedAt,
	}, nil
}

func toMatchHistory(m *matchHistoryModel) domain.MatchHistory {
	return domain.MatchHistory{
		ID:             m.ID,
		GameID:         m.GameID,
		WinnerID:       m.WinnerID,
		LoserID:        m.LoserID,
		WinnerUsername: m.WinnerUsername,
		LoserUsername:  m.LoserUsername,
		FinishedAt:     m.FinishedAt,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

func fromMatchHistory(h *domain.MatchHistory) *matchHistoryModel {
	return &matchHistoryModel{
		ID:             h.ID,
		GameID:         h.GameID,
		WinnerID:       h.WinnerID,
		LoserID:        h.LoserID,
		WinnerUsername: h.WinnerUsername,
		LoserUsername:  h.LoserUsername,
		FinishedAt:     h.FinishedAt,
		CreatedAt:      h.CreatedAt,
		UpdatedAt:      h.UpdatedAt,
	}
}

func toRanking(m *rankingModel) domain.Ranking {
	return domain.Ranking{
		UserID:    m.UserID,
		Rank:      m.Rank,
		Username:  m.Username,
		WinCount:  m.WinCount,
		UpdatedAt: m.UpdatedAt,
	}
}

func fromRanking(r domain.Ranking) rankingModel {
	return rankingModel{
		UserID:    r.UserID,
		Rank:      r.Rank,
		Username:  r.Username,
		WinCount:  r.WinCount,
		UpdatedAt: r.UpdatedAt,
	}
}

func toMatchingQueueEntry(m *matchingQueueModel) domain.MatchingQueueEntry {
	return domain.MatchingQueueEntry{
		ID:        m.ID,
		UserID:    m.UserID,
		Status:    domain.MatchingQueueStatus(m.Status),
		CreatedAt: m.CreatedAt,
	}
}

func fromMatchingQueueEntry(e *domain.MatchingQueueEntry) *matchingQueueModel {
	return &matchingQueueModel{
		ID:        e.ID,
		UserID:    e.UserID,
		Status:    string(e.Status),
		CreatedAt: e.CreatedAt,
	}
}

func toRefreshToken(m *refreshTokenModel) *domain.RefreshToken {
	return &domain.RefreshToken{
		ID:        m.ID,
		UserID:    m.UserID,
		TokenHash: m.TokenHash,
		FamilyID:  m.FamilyID,
		Status:    domain.RefreshTokenStatus(m.Status),
		ExpiresAt: m.ExpiresAt,
		RevokedAt: m.RevokedAt,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func fromRefreshToken(t *domain.RefreshToken) *refreshTokenModel {
	return &refreshTokenModel{
		ID:        t.ID,
		UserID:    t.UserID,
		TokenHash: t.TokenHash,
		FamilyID:  t.FamilyID,
		Status:    string(t.Status),
		ExpiresAt: t.ExpiresAt,
		RevokedAt: t.RevokedAt,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

func toActivityLog(m *activityLogModel) domain.ActivityLog {
	return domain.ActivityLog{
		ID:        m.ID,
		UserID:    m.UserID,
		LogType:   m.LogType,
		Detail:    json.RawMessage(m.Detail),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func fromActivityLog(l *domain.ActivityLog) *activityLogModel {
	return &activityLogModel{
		ID:        l.ID,
		UserID:    l.UserID,
		LogType:   l.LogType,
		Detail:    datatypes.JSON(l.Detail),
		CreatedAt: l.CreatedAt,
		UpdatedAt: l.UpdatedAt,
	}
}

func toLoginLog(m *loginLogModel) domain.LoginLog {
	return domain.LoginLog{
		ID:        m.ID,
		UserID:    m.UserID,
		Action:    domain.LoginAction(m.Action),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func fromLoginLog(l *domain.LoginLog) *loginLogModel {
	return &loginLogModel{
		ID:        l.ID,
		UserID:    l.UserID,
		Action:    string(l.Action),
		CreatedAt: l.CreatedAt,
		UpdatedAt: l.UpdatedAt,
	}
}

func toWSConnectionLog(m *wsConnectionLogModel) domain.WSConnectionLog {
	return domain.WSConnectionLog{
		ID:             m.ID,
		UserID:         m.UserID,
		ConnectionID:   m.ConnectionID,
		ConnectedAt:    m.ConnectedAt,
		DisconnectedAt: m.DisconnectedAt,
	}
}

func fromWSConnectionLog(l *domain.WSConnectionLog) *wsConnectionLogModel {
	return &wsConnectionLogModel{
		ID:             l.ID,
		UserID:         l.UserID,
		ConnectionID:   l.ConnectionID,
		ConnectedAt:    l.ConnectedAt,
		DisconnectedAt: l.DisconnectedAt,
	}
}