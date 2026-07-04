package controller

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/usecase"
)

func TestPresenterMappers(t *testing.T) {
	now := time.Now().UTC()
	rank := 3
	gameID := uuid.New()
	userID := uuid.New()
	deleted := now.Add(-time.Hour)
	disconnected := now.Add(time.Minute)
	synced := now.Add(-2 * time.Hour)

	_ = registerUserResponse(&usecase.RegisterResult{
		ID: userID.String(), Username: "alice", Role: string(model.RoleUser), WinCount: 1,
	})
	_ = loginResponse(&usecase.LoginResult{ID: userID.String(), Username: "alice", Role: string(model.RoleUser)})
	_ = getMeResponse(&usecase.MeResult{ID: userID.String(), Username: "alice", Role: string(model.RoleUser), WinCount: 2})

	_ = getProfileResponse(&usecase.GetProfileOutput{Username: "alice", WinCount: 2, Rank: nil})
	_ = getProfileResponse(&usecase.GetProfileOutput{Username: "alice", WinCount: 2, Rank: &rank})

	_ = matchHistoryResponse([]usecase.MatchHistoryItem{{
		GameID: gameID, WinnerUsername: "a", LoserUsername: "b", FinishedAt: now,
	}})
	_ = loginHistoryResponse([]usecase.LoginHistoryItem{{
		Action: model.LoginActionLogin, CreatedAt: now,
	}})
	_ = wsHistoryResponse([]usecase.WSConnectionHistoryItem{{
		ConnectionID: "c1", ConnectedAt: now, DisconnectedAt: nil,
	}})
	_ = wsHistoryResponse([]usecase.WSConnectionHistoryItem{{
		ConnectionID: "c2", ConnectedAt: now, DisconnectedAt: &disconnected,
	}})

	_ = rankingResponse([]usecase.RankingItem{{Rank: 1, Username: "alice", WinCount: 5}})

	_ = adminUsersResponse([]usecase.AdminUserItem{{
		ID: userID, Username: "alice", Email: "a@test.local", Role: model.RoleUser,
		WinCount: 0, DeletedAt: nil, CreatedAt: now,
	}})
	_ = adminUsersResponse([]usecase.AdminUserItem{{
		ID: userID, Username: "bob", Email: "b@test.local", Role: model.RoleUser,
		WinCount: 1, DeletedAt: &deleted, CreatedAt: now,
	}})

	uid := userID
	_ = activityLogsResponse([]usecase.ActivityLogItem{{
		ID: uuid.New(), UserID: nil, LogType: "x", Detail: json.RawMessage(`{}`), CreatedAt: now,
	}})
	_ = activityLogsResponse([]usecase.ActivityLogItem{{
		ID: uuid.New(), UserID: &uid, LogType: "y", Detail: json.RawMessage(`{}`), CreatedAt: now,
	}})
	_ = logTypesResponse([]string{"login", "game"})
	_ = backupStatusResponse(&usecase.BackupStatusOutput{Status: "idle", LastSyncedAt: nil})
	_ = backupStatusResponse(&usecase.BackupStatusOutput{Status: "ok", LastSyncedAt: &synced})

	_ = startMatchingResponse(&usecase.StartMatchingOutput{Status: "waiting"})
	_ = cancelMatchingResponse(&usecase.CancelMatchingOutput{Status: "cancelled"})
	_ = matchingStatusResponse(&usecase.GetMatchingStatusOutput{Status: "waiting", GameID: nil})
	_ = matchingStatusResponse(&usecase.GetMatchingStatusOutput{Status: "matched", GameID: &gameID})

	_ = gameStateResponse(&usecase.GameStateOutput{
		GameID: gameID, Status: model.GameStatusInProgress, CurrentTurn: 1,
		CurrentTurnPlayerID: userID.String(), RemainingSeconds: 30,
		MyGuesses: []usecase.GuessSummary{{
			Turn: 1, GuessNumber: "1234", DigitResults: []int{1, 0, 1, 0}, HitCount: 2, IsAuto: false,
		}},
		OpponentGuessCount: 1,
	})
}
