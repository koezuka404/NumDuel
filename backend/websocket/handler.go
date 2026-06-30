package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	gorillaws "github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"

	"github.com/numduel/numduel/middleware"
	"github.com/numduel/numduel/usecase"
)

type Handler struct {
	Hub     *Hub
	WSAuth  usecase.IWSAuthUsecase
	Game    usecase.IGameUsecase
	Allowed map[string]struct{}
	Redis   usecase.IWSSessionStore
	JWTMin  int
}

type clientMsg struct {
	Type         string `json:"type"`
	GameID       string `json:"gameId"`
	SecretNumber string `json:"secretNumber"`
	GuessNumber  string `json:"guessNumber"`
}

func (h *Handler) Handle(c echo.Context) error {
	upgrader := gorillaws.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			if len(h.Allowed) == 0 {
				return true
			}
			origin := r.Header.Get("Origin")
			if origin == "" {
				return true
			}
			_, ok := h.Allowed[origin]
			return ok
		},
	}
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	ctx := c.Request().Context()
	authCh := make(chan struct{}, 1)
	var userID, wsLogID uuid.UUID
	connID := uuid.New().String()

	go func() {
		time.Sleep(5 * time.Second)
		select {
		case <-authCh:
		default:
			conn.Close()
		}
	}()

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			if userID != uuid.Nil {
				h.onDisconnect(ctx, userID, wsLogID)
			}
			return nil
		}
		var msg clientMsg
		if err := json.Unmarshal(raw, &msg); err != nil {
			continue
		}
		msgType := strings.ToUpper(strings.TrimSpace(msg.Type))

		if userID == uuid.Nil {
			if msgType != "AUTH" {
				h.writeError(conn, "unauthorized", "authentication required")
				continue
			}
			token := ""
			if ck, err := c.Request().Cookie(middleware.AccessCookieName); err == nil {
				token = ck.Value
			}
			out, err := h.WSAuth.Authenticate(ctx, token)
			if err != nil {
				h.writeUseCaseError(conn, err)
				continue
			}
			userID = out.UserID
			wsLogID, err = h.WSAuth.RecordConnection(ctx, userID, connID)
			if err != nil {
				h.writeUseCaseError(conn, err)
				continue
			}
			h.Hub.Register(userID, connID, conn)
			if h.Redis != nil {
				ttl := time.Duration(h.JWTMin) * time.Minute
				_ = h.Redis.SetUser(ctx, userID, connID, ttl)
			}
			authCh <- struct{}{}
			_ = h.writeJSON(conn, map[string]any{
				"type": "AUTH_OK", "data": map[string]string{"userId": userID.String()},
			})
			h.WSAuth.NotifyOpponentConnected(ctx, userID)
			continue
		}

		switch msgType {
		case "PING":
			h.WSAuth.TouchActivity(ctx, userID)
			_ = h.writeJSON(conn, map[string]any{"type": "PONG"})
		case "SET_SECRET":
			h.WSAuth.TouchActivity(ctx, userID)
			gameID, err := uuid.Parse(msg.GameID)
			if err != nil {
				h.Hub.SendError(userID, "validation_error", "invalid gameId")
				continue
			}
			if err := h.Game.SetSecretNumber(ctx, userID, gameID, msg.SecretNumber); err != nil {
				h.sendUseCaseError(userID, err)
			}
		case "GUESS":
			h.WSAuth.TouchActivity(ctx, userID)
			gameID, err := uuid.Parse(msg.GameID)
			if err != nil {
				h.Hub.SendError(userID, "validation_error", "invalid gameId")
				continue
			}
			if err := h.Game.SubmitGuess(ctx, userID, gameID, msg.GuessNumber, false); err != nil {
				h.sendUseCaseError(userID, err)
			}
		case "SYNC_REQUEST":
			h.WSAuth.TouchActivity(ctx, userID)
			gameID, err := uuid.Parse(msg.GameID)
			if err != nil {
				h.Hub.SendError(userID, "validation_error", "invalid gameId")
				continue
			}
			if _, err := h.Game.SyncGameState(ctx, userID, gameID); err != nil {
				h.sendUseCaseError(userID, err)
			}
		default:
			h.Hub.SendError(userID, "validation_error", "unknown event type")
		}
	}
}

func (h *Handler) onDisconnect(ctx context.Context, userID, wsLogID uuid.UUID) {
	h.Hub.Disconnect(userID)
	if h.Redis != nil {
		_ = h.Redis.DeleteUser(ctx, userID)
	}
	h.WSAuth.CloseConnectionLog(ctx, wsLogID)
	h.WSAuth.NotifyOpponentDisconnected(ctx, userID)
}

func (h *Handler) sendUseCaseError(userID uuid.UUID, err error) {
	code, msg := wsErrorCode(err)
	h.Hub.SendError(userID, code, msg)
}

func (h *Handler) writeUseCaseError(conn *gorillaws.Conn, err error) {
	code, msg := wsErrorCode(err)
	h.writeError(conn, code, msg)
}

func (h *Handler) writeError(conn *gorillaws.Conn, code, message string) {
	_ = h.writeJSON(conn, map[string]any{
		"type": "ERROR", "data": map[string]string{"code": code, "message": message},
	})
}

func (h *Handler) writeJSON(conn *gorillaws.Conn, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return conn.WriteMessage(gorillaws.TextMessage, b)
}
