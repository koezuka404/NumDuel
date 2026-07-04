package websocket_test

import (
	"testing"

	"github.com/numduel/numduel/testutil"
)

// セキュリティ: 未認証 WS 接続
func TestWSAuthFailsWithoutCookie(t *testing.T) {
	env := setupWSTest(t)
	conn := dialWS(t, env, "")

	sendWS(t, conn, map[string]string{"type": "AUTH"})
	msg := readWSMessage(t, conn)
	data, _ := msg["data"].(map[string]any)
	if msg["type"] != "ERROR" || data["code"] != "unauthorized" {
		t.Fatalf("message: %+v", msg)
	}
}

func TestWSAuthFailsWithInvalidToken(t *testing.T) {
	env := setupWSTest(t)
	testutil.CreateUser(t, env.repos, "alice", "alice@test.local", "password123")
	conn := dialWS(t, env, "not-a-valid-jwt")

	sendWS(t, conn, map[string]string{"type": "AUTH"})
	msg := readWSMessage(t, conn)
	data, _ := msg["data"].(map[string]any)
	if msg["type"] != "ERROR" || data["code"] != "unauthorized" {
		t.Fatalf("message: %+v", msg)
	}
}
