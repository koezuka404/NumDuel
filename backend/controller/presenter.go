package controller

import (
	"time"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/usecase"
)

func loginActionLabel(action model.LoginAction) string {
	switch action {
	case model.LoginActionLogin:
		return "ログイン"
	case model.LoginActionLogout:
		return "ログアウト"
	case model.LoginActionAutoLogout:
		return "自動ログアウト"
	default:
		return string(action)
	}
}

func registerUserResponse(out *usecase.RegisterResult) map[string]any {
	return map[string]any{
		"id": out.ID, "username": out.Username, "role": out.Role, "winCount": out.WinCount,
	}
}

func loginResponse(out *usecase.LoginResult) map[string]any {
	return map[string]any{
		"id": out.ID, "username": out.Username, "role": out.Role,
	}
}

func getMeResponse(out *usecase.MeResult) map[string]any {
	return map[string]any{
		"id": out.ID, "username": out.Username, "role": out.Role, "winCount": out.WinCount,
	}
}

func getProfileResponse(out *usecase.GetProfileOutput) map[string]any {
	data := map[string]any{"username": out.Username, "winCount": out.WinCount, "rank": nil}
	if out.Rank != nil {
		data["rank"] = *out.Rank
	}
	return data
}

func matchHistoryResponse(items []usecase.MatchHistoryItem) []map[string]any {
	rows := make([]map[string]any, len(items))
	for i, item := range items {
		rows[i] = map[string]any{
			"gameId": item.GameID.String(), "winnerUsername": item.WinnerUsername,
			"loserUsername": item.LoserUsername, "finishedAt": item.FinishedAt.UTC().Format(time.RFC3339),
		}
	}
	return rows
}

func loginHistoryResponse(items []usecase.LoginHistoryItem) []map[string]any {
	rows := make([]map[string]any, len(items))
	for i, item := range items {
		rows[i] = map[string]any{
			"action": loginActionLabel(item.Action), "createdAt": item.CreatedAt.UTC().Format(time.RFC3339),
		}
	}
	return rows
}

func wsHistoryResponse(items []usecase.WSConnectionHistoryItem) []map[string]any {
	rows := make([]map[string]any, len(items))
	for i, item := range items {
		disconnected := any(nil)
		if item.DisconnectedAt != nil {
			disconnected = item.DisconnectedAt.UTC().Format(time.RFC3339)
		}
		rows[i] = map[string]any{
			"connectionId":   item.ConnectionID,
			"connectedAt":    item.ConnectedAt.UTC().Format(time.RFC3339),
			"disconnectedAt": disconnected,
		}
	}
	return rows
}

func rankingResponse(items []usecase.RankingItem) []map[string]any {
	rows := make([]map[string]any, len(items))
	for i, item := range items {
		rows[i] = map[string]any{"rank": item.Rank, "username": item.Username, "winCount": item.WinCount}
	}
	return rows
}

func adminUsersResponse(items []usecase.AdminUserItem) []map[string]any {
	rows := make([]map[string]any, len(items))
	for i, u := range items {
		var deleted any
		if u.DeletedAt != nil {
			deleted = u.DeletedAt.UTC().Format(time.RFC3339)
		}
		rows[i] = map[string]any{
			"id": u.ID.String(), "username": u.Username, "email": u.Email,
			"role": string(u.Role), "winCount": u.WinCount,
			"deletedAt": deleted, "createdAt": u.CreatedAt.UTC().Format(time.RFC3339),
		}
	}
	return rows
}

func activityLogsResponse(items []usecase.ActivityLogItem) []map[string]any {
	rows := make([]map[string]any, len(items))
	for i, item := range items {
		var uid any
		if item.UserID != nil {
			uid = item.UserID.String()
		}
		rows[i] = map[string]any{
			"id": item.ID.String(), "userId": uid, "logType": item.LogType,
			"detail": item.Detail, "createdAt": item.CreatedAt.UTC().Format(time.RFC3339),
		}
	}
	return rows
}

func logTypesResponse(types []string) map[string]any {
	return map[string]any{"logTypes": types}
}

func backupStatusResponse(out *usecase.BackupStatusOutput) map[string]any {
	data := map[string]any{"status": out.Status, "lastSyncedAt": nil}
	if out.LastSyncedAt != nil {
		data["lastSyncedAt"] = out.LastSyncedAt.UTC().Format(time.RFC3339)
	}
	return data
}

func startMatchingResponse(out *usecase.StartMatchingOutput) map[string]any {
	data := map[string]any{"status": out.Status, "gameId": nil}
	if out.GameID != nil {
		data["gameId"] = out.GameID.String()
	}
	return data
}

func cancelMatchingResponse(out *usecase.CancelMatchingOutput) map[string]string {
	return map[string]string{"status": out.Status}
}

func matchingStatusResponse(out *usecase.GetMatchingStatusOutput) map[string]any {
	data := map[string]any{"status": out.Status, "gameId": nil}
	if out.GameID != nil {
		data["gameId"] = out.GameID.String()
	}
	return data
}

func gameStateResponse(s *usecase.GameStateOutput) map[string]any {
	guesses := make([]map[string]any, len(s.MyGuesses))
	for i, row := range s.MyGuesses {
		guesses[i] = map[string]any{
			"turn": row.Turn, "guessNumber": row.GuessNumber,
			"digitResults": row.DigitResults, "hitCount": row.HitCount, "isAuto": row.IsAuto,
		}
	}
	return map[string]any{
		"gameId": s.GameID.String(), "status": string(s.Status),
		"currentTurn": s.CurrentTurn, "currentTurnPlayerID": s.CurrentTurnPlayerID,
		"remainingSeconds": s.RemainingSeconds, "myGuesses": guesses,
		"opponentGuessCount": s.OpponentGuessCount,
	}
}
