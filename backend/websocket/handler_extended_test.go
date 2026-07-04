package websocket_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	gorillaws "github.com/gorilla/websocket"

	"github.com/numduel/numduel/testutil"
	infrws "github.com/numduel/numduel/websocket"
)

func TestWSInvalidGameIdReturnsError(t *testing.T) {
	env := setupWSTest(t)
	testutil.CreateUser(t, env.repos, "alice", "alice@test.local", "password123")
	conn := dialWS(t, env, loginAccessToken(t, env, "alice@test.local", "password123"))
	sendWS(t, conn, map[string]string{"type": "AUTH"})
	if msg := readWSMessage(t, conn); msg["type"] != "AUTH_OK" {
		t.Fatalf("auth: %+v", msg)
	}
	sendWS(t, conn, map[string]string{"type": "SET_SECRET", "gameId": "not-uuid", "secretNumber": "1234"})
	msg := readWSMessage(t, conn)
	data, _ := msg["data"].(map[string]any)
	if msg["type"] != "ERROR" || data["code"] != "validation_error" {
		t.Fatalf("message: %+v", msg)
	}
}

func TestWSUnknownEventType(t *testing.T) {
	env := setupWSTest(t)
	testutil.CreateUser(t, env.repos, "alice", "alice@test.local", "password123")
	conn := dialWS(t, env, loginAccessToken(t, env, "alice@test.local", "password123"))
	sendWS(t, conn, map[string]string{"type": "AUTH"})
	if msg := readWSMessage(t, conn); msg["type"] != "AUTH_OK" {
		t.Fatalf("auth: %+v", msg)
	}
	sendWS(t, conn, map[string]string{"type": "NOPE"})
	msg := readWSMessage(t, conn)
	data, _ := msg["data"].(map[string]any)
	if msg["type"] != "ERROR" || data["code"] != "validation_error" {
		t.Fatalf("message: %+v", msg)
	}
}

func TestWSSetSecretUseCaseError(t *testing.T) {
	env := setupWSTest(t)
	testutil.CreateUser(t, env.repos, "alice", "alice@test.local", "password123")
	conn := dialWS(t, env, loginAccessToken(t, env, "alice@test.local", "password123"))
	sendWS(t, conn, map[string]string{"type": "AUTH"})
	if msg := readWSMessage(t, conn); msg["type"] != "AUTH_OK" {
		t.Fatalf("auth: %+v", msg)
	}
	sendWS(t, conn, map[string]string{
		"type": "SET_SECRET", "gameId": "00000000-0000-0000-0000-000000000001", "secretNumber": "1234",
	})
	msg := readWSMessage(t, conn)
	data, _ := msg["data"].(map[string]any)
	if msg["type"] != "ERROR" {
		t.Fatalf("message: %+v", msg)
	}
	if data["code"] != "not_found" && data["code"] != "conflict" {
		t.Fatalf("unexpected code: %+v", data)
	}
}

func TestWSGuessInvalidGameId(t *testing.T) {
	conn := wsAuthConn(t, setupWSTest(t), "alice@test.local", "password123")
	sendWS(t, conn, map[string]string{"type": "GUESS", "gameId": "not-uuid", "guessNumber": "9012"})
	msg := readWSMessage(t, conn)
	data, _ := msg["data"].(map[string]any)
	if msg["type"] != "ERROR" || data["code"] != "validation_error" {
		t.Fatalf("message: %+v", msg)
	}
}

func TestWSGuessUseCaseError(t *testing.T) {
	conn := wsAuthConn(t, setupWSTest(t), "alice@test.local", "password123")
	sendWS(t, conn, map[string]string{
		"type": "GUESS", "gameId": "00000000-0000-0000-0000-000000000001", "guessNumber": "9012",
	})
	msg := readWSMessage(t, conn)
	data, _ := msg["data"].(map[string]any)
	if msg["type"] != "ERROR" || data["code"] != "not_found" {
		t.Fatalf("message: %+v", msg)
	}
}

func TestWSGuessSuccess(t *testing.T) {
	env := setupWSTest(t)
	a := testutil.CreateUser(t, env.repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, env.repos, "bob", "bob@test.local", "password123")
	gameID := matchTwoUsers(t, env, a.ID, b.ID)
	setBothSecrets(t, env, gameID, a.ID, b.ID, "1234", "5678")

	conn := dialWS(t, env, loginAccessToken(t, env, "alice@test.local", "password123"))
	sendWS(t, conn, map[string]string{"type": "AUTH"})
	if msg := readWSMessage(t, conn); msg["type"] != "AUTH_OK" {
		t.Fatalf("auth: %+v", msg)
	}
	sendWS(t, conn, map[string]string{
		"type": "GUESS", "gameId": gameID.String(), "guessNumber": "9012",
	})
	msg := readWSMessage(t, conn)
	if msg["type"] != "GUESS_RESULT" {
		t.Fatalf("guess result: %+v", msg)
	}
}

func TestWSSyncRequestInvalidGameId(t *testing.T) {
	conn := wsAuthConn(t, setupWSTest(t), "alice@test.local", "password123")
	sendWS(t, conn, map[string]string{"type": "SYNC_REQUEST", "gameId": "bad"})
	msg := readWSMessage(t, conn)
	data, _ := msg["data"].(map[string]any)
	if msg["type"] != "ERROR" || data["code"] != "validation_error" {
		t.Fatalf("message: %+v", msg)
	}
}

func TestWSSyncRequestSuccess(t *testing.T) {
	env := setupWSTest(t)
	a := testutil.CreateUser(t, env.repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, env.repos, "bob", "bob@test.local", "password123")
	gameID := matchTwoUsers(t, env, a.ID, b.ID)
	setBothSecrets(t, env, gameID, a.ID, b.ID, "1234", "5678")

	conn := dialWS(t, env, loginAccessToken(t, env, "alice@test.local", "password123"))
	sendWS(t, conn, map[string]string{"type": "AUTH"})
	if msg := readWSMessage(t, conn); msg["type"] != "AUTH_OK" {
		t.Fatalf("auth: %+v", msg)
	}
	sendWS(t, conn, map[string]string{"type": "SYNC_REQUEST", "gameId": gameID.String()})
	msg := readWSMessage(t, conn)
	if msg["type"] != "GAME_STATE_SYNC" {
		t.Fatalf("sync: %+v", msg)
	}
}

func TestWSSyncRequestForbidden(t *testing.T) {
	env := setupWSTest(t)
	a := testutil.CreateUser(t, env.repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, env.repos, "bob", "bob@test.local", "password123")
	testutil.CreateUser(t, env.repos, "carol", "carol@test.local", "password123")
	gameID := matchTwoUsers(t, env, a.ID, b.ID)

	conn := dialWS(t, env, loginAccessToken(t, env, "carol@test.local", "password123"))
	sendWS(t, conn, map[string]string{"type": "AUTH"})
	if msg := readWSMessage(t, conn); msg["type"] != "AUTH_OK" {
		t.Fatalf("auth: %+v", msg)
	}
	sendWS(t, conn, map[string]string{"type": "SYNC_REQUEST", "gameId": gameID.String()})
	msg := readWSMessage(t, conn)
	data, _ := msg["data"].(map[string]any)
	if msg["type"] != "ERROR" || data["code"] != "forbidden" {
		t.Fatalf("message: %+v", msg)
	}
}

func TestWSOriginCheckRejected(t *testing.T) {
	env := setupWSTestWithOpts(t, wsTestOpts{
		allowed: map[string]struct{}{"https://allowed.example": {}},
	})
	testutil.CreateUser(t, env.repos, "alice", "alice@test.local", "password123")
	token := loginAccessToken(t, env, "alice@test.local", "password123")
	_, err := dialWSWithOrigin(t, env, token, "https://evil.example")
	if err == nil {
		t.Fatal("expected origin rejection")
	}
	if !strings.Contains(err.Error(), "bad handshake") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWSOriginCheckAllowed(t *testing.T) {
	env := setupWSTestWithOpts(t, wsTestOpts{
		allowed: map[string]struct{}{"https://allowed.example": {}},
	})
	conn, err := dialWSWithOrigin(t, env, "", "https://allowed.example")
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	sendWS(t, conn, map[string]string{"type": "PING"})
	msg := readWSMessage(t, conn)
	data, _ := msg["data"].(map[string]any)
	if msg["type"] != "ERROR" || data["message"] != "authentication required" {
		t.Fatalf("message: %+v", msg)
	}
}

func TestWSOriginEmptyAllowedWhenRestricted(t *testing.T) {
	env := setupWSTestWithOpts(t, wsTestOpts{
		allowed: map[string]struct{}{"https://allowed.example": {}},
	})
	conn, err := dialWSWithOrigin(t, env, "", "")
	if err != nil {
		t.Fatalf("dial without origin: %v", err)
	}
	sendWS(t, conn, map[string]string{"type": "PING"})
	msg := readWSMessage(t, conn)
	data, _ := msg["data"].(map[string]any)
	if msg["type"] != "ERROR" || data["message"] != "authentication required" {
		t.Fatalf("message: %+v", msg)
	}
}

func TestWSAuthStoresAndClearsRedisOnDisconnect(t *testing.T) {
	redisStore, mr := newMiniredisStore(t)
	env := setupWSTestWithOpts(t, wsTestOpts{redis: redisStore})
	a := testutil.CreateUser(t, env.repos, "alice", "alice@test.local", "password123")
	token := loginAccessToken(t, env, "alice@test.local", "password123")

	conn, err := dialWSWithOrigin(t, env, token, "")
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	sendWS(t, conn, map[string]string{"type": "AUTH"})
	if msg := readWSMessage(t, conn); msg["type"] != "AUTH_OK" {
		t.Fatalf("auth: %+v", msg)
	}
	if !mr.Exists(wsUserRedisKey(a.ID)) {
		t.Fatal("expected redis session key after auth")
	}

	if err := conn.WriteMessage(gorillaws.CloseMessage, gorillaws.FormatCloseMessage(gorillaws.CloseNormalClosure, "")); err != nil {
		t.Fatalf("close: %v", err)
	}
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if !mr.Exists(wsUserRedisKey(a.ID)) {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("expected redis session key removed after disconnect")
}

func TestWSIgnoresInvalidJSON(t *testing.T) {
	conn := wsAuthConn(t, setupWSTest(t), "alice@test.local", "password123")
	if err := conn.WriteMessage(gorillaws.TextMessage, []byte("{not-json")); err != nil {
		t.Fatalf("write: %v", err)
	}
	sendWS(t, conn, map[string]string{"type": "PING"})
	msg := readWSMessage(t, conn)
	if msg["type"] != "PONG" {
		t.Fatalf("pong after bad json: %+v", msg)
	}
}

func TestWSAuthTimeoutClosesConnection(t *testing.T) {
	env := setupWSTest(t)
	testutil.CreateUser(t, env.repos, "alice", "alice@test.local", "password123")
	conn := dialWS(t, env, loginAccessToken(t, env, "alice@test.local", "password123"))
	time.Sleep(5100 * time.Millisecond)
	sendWS(t, conn, map[string]string{"type": "AUTH"})
	_ = conn.SetReadDeadline(time.Now().Add(time.Second))
	if _, _, err := conn.ReadMessage(); err == nil {
		t.Fatal("expected closed connection after auth timeout")
	}
}

func TestWSGuessNotYourTurn(t *testing.T) {
	env := setupWSTest(t)
	a := testutil.CreateUser(t, env.repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, env.repos, "bob", "bob@test.local", "password123")
	gameID := matchTwoUsers(t, env, a.ID, b.ID)
	setBothSecrets(t, env, gameID, a.ID, b.ID, "1234", "5678")

	conn := dialWS(t, env, loginAccessToken(t, env, "bob@test.local", "password123"))
	sendWS(t, conn, map[string]string{"type": "AUTH"})
	if msg := readWSMessage(t, conn); msg["type"] != "AUTH_OK" {
		t.Fatalf("auth: %+v", msg)
	}
	sendWS(t, conn, map[string]string{
		"type": "GUESS", "gameId": gameID.String(), "guessNumber": "9012",
	})
	msg := readWSMessage(t, conn)
	data, _ := msg["data"].(map[string]any)
	if msg["type"] != "ERROR" || data["code"] != "conflict" {
		t.Fatalf("message: %+v", msg)
	}
}

func TestWSSetSecretSuccess(t *testing.T) {
	env := setupWSTest(t)
	a := testutil.CreateUser(t, env.repos, "alice", "alice@test.local", "password123")
	b := testutil.CreateUser(t, env.repos, "bob", "bob@test.local", "password123")
	gameID := matchTwoUsers(t, env, a.ID, b.ID)

	conn := dialWS(t, env, loginAccessToken(t, env, "alice@test.local", "password123"))
	sendWS(t, conn, map[string]string{"type": "AUTH"})
	if msg := readWSMessage(t, conn); msg["type"] != "AUTH_OK" {
		t.Fatalf("auth: %+v", msg)
	}
	sendWS(t, conn, map[string]string{
		"type": "SET_SECRET", "gameId": gameID.String(), "secretNumber": "1234",
	})
	_ = conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	if _, _, err := conn.ReadMessage(); err == nil {
		t.Fatal("set secret success should not push error to same connection")
	}
	game, err := env.repos.Game.FindByID(t.Context(), gameID)
	if err != nil || game.Player1Secret == "" {
		t.Fatalf("secret saved: %+v err=%v", game, err)
	}
}

func TestWSUpgradeError(t *testing.T) {
	env := setupWSTest(t)
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	rec := httptest.NewRecorder()
	env.echo.ServeHTTP(rec, req)
	if rec.Code == http.StatusSwitchingProtocols {
		t.Fatalf("expected upgrade failure, got %d", rec.Code)
	}
}

func TestWSRecordConnectionError(t *testing.T) {
	env := setupWSTestWithOpts(t, wsTestOpts{recordConnErr: errors.New("record failed")})
	testutil.CreateUser(t, env.repos, "alice", "alice@test.local", "password123")
	conn := dialWS(t, env, loginAccessToken(t, env, "alice@test.local", "password123"))
	sendWS(t, conn, map[string]string{"type": "AUTH"})
	msg := readWSMessage(t, conn)
	data, _ := msg["data"].(map[string]any)
	if msg["type"] != "ERROR" || data["code"] != "internal_error" {
		t.Fatalf("message: %+v", msg)
	}
}

func TestSessionStoreSetUserWithRedis(t *testing.T) {
	redisStore, mr := newMiniredisStore(t)
	hub := infrws.NewHub()
	store := infrws.NewSessionStore(hub, redisStore)
	userID := uuid.New()
	if err := store.SetUser(t.Context(), userID, "conn-1", time.Minute); err != nil {
		t.Fatalf("SetUser: %v", err)
	}
	if !mr.Exists(wsUserRedisKey(userID)) {
		t.Fatal("expected redis key")
	}
}

func TestSessionStoreDeleteUserWithRedis(t *testing.T) {
	redisStore, mr := newMiniredisStore(t)
	hub := infrws.NewHub()
	store := infrws.NewSessionStore(hub, redisStore)
	userID := uuid.New()
	if err := store.SetUser(t.Context(), userID, "conn-1", time.Minute); err != nil {
		t.Fatalf("SetUser: %v", err)
	}
	if err := store.DeleteUser(t.Context(), userID); err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}
	if mr.Exists(wsUserRedisKey(userID)) {
		t.Fatal("expected redis key removed")
	}
}
