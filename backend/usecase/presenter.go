package usecase

import "time"

func RegisterUserResponse(out *RegisterUserOutput) map[string]any {
	return map[string]any{
		"id": out.ID.String(), "username": out.Username, "role": out.Role, "winCount": out.WinCount,
	}
}

func LoginResponse(out *LoginOutput) map[string]any {
	return map[string]any{
		"id": out.ID.String(), "username": out.Username, "role": string(out.Role),
	}
}

func GetMeResponse(out *GetMeOutput) map[string]any {
	return map[string]any{
		"id": out.ID.String(), "username": out.Username, "role": out.Role, "winCount": out.WinCount,
	}
}

func GetProfileResponse(out *GetProfileOutput) map[string]any {
	data := map[string]any{"username": out.Username, "winCount": out.WinCount, "rank": nil}
	if out.Rank != nil {
		data["rank"] = *out.Rank
	}
	return data
}

func MatchHistoryResponse(items []MatchHistoryItem) []map[string]any {
	rows := make([]map[string]any, len(items))
	for i, item := range items {
		rows[i] = map[string]any{
			"gameId": item.GameID.String(), "winnerUsername": item.WinnerUsername,
			"loserUsername": item.LoserUsername, "finishedAt": item.FinishedAt.UTC().Format(time.RFC3339),
		}
	}
	return rows
}

func LoginHistoryResponse(items []LoginHistoryItem) []map[string]any {
	rows := make([]map[string]any, len(items))
	for i, item := range items {
		rows[i] = map[string]any{
			"action": string(item.Action), "createdAt": item.CreatedAt.UTC().Format(time.RFC3339),
		}
	}
	return rows
}

func WSHistoryResponse(items []WSConnectionHistoryItem) []map[string]any {
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

func RankingResponse(items []RankingItem) []map[string]any {
	rows := make([]map[string]any, len(items))
	for i, item := range items {
		rows[i] = map[string]any{"rank": item.Rank, "username": item.Username, "winCount": item.WinCount}
	}
	return rows
}

func AdminUsersResponse(items []AdminUserItem) []map[string]any {
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

func ActivityLogsResponse(items []ActivityLogItem) []map[string]any {
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

func LogTypesResponse(types []string) map[string]any {
	return map[string]any{"logTypes": types}
}

func BackupStatusResponse(out *BackupStatusOutput) map[string]any {
	data := map[string]any{"status": out.Status, "lastSyncedAt": nil}
	if out.LastSyncedAt != nil {
		data["lastSyncedAt"] = out.LastSyncedAt.UTC().Format(time.RFC3339)
	}
	return data
}

func StartMatchingResponse(out *StartMatchingOutput) map[string]string {
	return map[string]string{"status": out.Status}
}

func CancelMatchingResponse(out *CancelMatchingOutput) map[string]string {
	return map[string]string{"status": out.Status}
}

func MatchingStatusResponse(out *GetMatchingStatusOutput) map[string]any {
	data := map[string]any{"status": out.Status, "gameId": nil}
	if out.GameID != nil {
		data["gameId"] = out.GameID.String()
	}
	return data
}
