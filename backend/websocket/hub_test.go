package websocket_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	infrws "github.com/numduel/numduel/websocket"
)

func TestHubSendToMissingUser(t *testing.T) {
	hub := infrws.NewHub()
	if err := hub.SendToUser(t.Context(), uuid.New(), "X", nil); err != nil {
		t.Fatalf("SendToUser: %v", err)
	}
}

func TestSessionStoreWithNilRedis(t *testing.T) {
	hub := infrws.NewHub()
	store := infrws.NewSessionStore(hub, nil)
	userID := uuid.New()
	if err := store.SetUser(t.Context(), userID, "c1", time.Minute); err != nil {
		t.Fatalf("SetUser: %v", err)
	}
	if err := store.DeleteUser(t.Context(), userID); err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}
	if err := store.DisconnectWithError(t.Context(), userID, "unauthorized", "bye"); err != nil {
		t.Fatalf("DisconnectWithError: %v", err)
	}
}
