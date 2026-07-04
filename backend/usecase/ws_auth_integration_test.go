package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	infrcrypto "github.com/numduel/numduel/crypto"
	"github.com/numduel/numduel/testutil"
	"github.com/numduel/numduel/usecase"
)

func TestWSAuthRecordConnectionAndClose(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	jwtSvc, err := infrcrypto.NewJWTService(testutil.TestJWTSecret, 60)
	if err != nil {
		t.Fatalf("jwt: %v", err)
	}
	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	wsAuth := usecase.NewWSAuthUseCase(repos, jwtSvc, nil, nil, nil)
	wsAuth.Now = func() time.Time { return time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC) }

	logID, err := wsAuth.RecordConnection(context.Background(), user.ID, "conn-abc")
	if err != nil || logID == uuid.Nil {
		t.Fatalf("record connection: logID=%v err=%v", logID, err)
	}

	rows, _, err := repos.WSConnectionLog.ListByUserID(context.Background(), user.ID, 1, 10)
	if err != nil || len(rows) != 1 || rows[0].ConnectionID != "conn-abc" {
		t.Fatalf("ws logs: %+v err=%v", rows, err)
	}

	wsAuth.CloseConnectionLog(context.Background(), logID)
	wsAuth.CloseConnectionLog(context.Background(), uuid.Nil)

	rows, _, _ = repos.WSConnectionLog.ListByUserID(context.Background(), user.ID, 1, 10)
	if rows[0].DisconnectedAt == nil {
		t.Fatalf("expected disconnected_at to be set")
	}
}

func TestWSAuthTouchActivity(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	jwtSvc, _ := infrcrypto.NewJWTService(testutil.TestJWTSecret, 60)
	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	wsAuth := usecase.NewWSAuthUseCase(repos, jwtSvc, nil, nil, nil)
	touchTime := time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC)
	wsAuth.Now = func() time.Time { return touchTime }

	wsAuth.TouchActivity(context.Background(), user.ID)

	updated, err := repos.User.FindByID(context.Background(), user.ID)
	if err != nil || !updated.LastActivityAt.Equal(touchTime) {
		t.Fatalf("activity not updated: %+v err=%v", updated.LastActivityAt, err)
	}

	wsAuth.TouchActivity(context.Background(), uuid.New())
}

func TestWSAuthNotifyOpponentStatus(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	notifier := &captureNotifier{}
	match := usecase.NewMatchingUseCase(repos, notifier)
	gameUC := testutil.NewGameUC(t, repos)
	jwtSvc, _ := infrcrypto.NewJWTService(testutil.TestJWTSecret, 60)
	wsAuth := usecase.NewWSAuthUseCase(repos, jwtSvc, nil, nil, notifier)

	a := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, repos, "bob", "bob@test.local", "password123")
	gameID := matchTwo(t, match, a.ID, b.ID)
	setBothSecrets(t, gameUC, gameID, a.ID, b.ID, "1234", "5678")

	wsAuth.NotifyOpponentConnected(context.Background(), a.ID)
	call := notifier.last()
	if call == nil || call.EventType != "OPPONENT_STATUS" || call.UserID != b.ID {
		t.Fatalf("connected notify: %+v", call)
	}
	if connected, ok := call.Payload["connected"].(bool); !ok || !connected {
		t.Fatalf("connected payload: %+v", call.Payload)
	}

	wsAuth.NotifyOpponentDisconnected(context.Background(), a.ID)
	call = notifier.last()
	if call == nil || call.UserID != b.ID {
		t.Fatalf("disconnected notify: %+v", call)
	}
	if connected, ok := call.Payload["connected"].(bool); !ok || connected {
		t.Fatalf("disconnected payload: %+v", call.Payload)
	}

	wsAuth.NotifyOpponentConnected(context.Background(), uuid.New())
}

func TestWSAuthNotifyNoActiveGame(t *testing.T) {
	_, repos := testutil.OpenSQLiteDB(t)
	notifier := &captureNotifier{}
	jwtSvc, _ := infrcrypto.NewJWTService(testutil.TestJWTSecret, 60)
	wsAuth := usecase.NewWSAuthUseCase(repos, jwtSvc, nil, nil, notifier)
	user := testutil.CreateUser(t, repos, "alice", "alice@test.local", "password123")

	wsAuth.NotifyOpponentConnected(context.Background(), user.ID)
	if len(notifier.calls) != 0 {
		t.Fatalf("expected no notifications without active game")
	}
}
