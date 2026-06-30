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
	"github.com/numduel/numduel/model"
	"github.com/numduel/numduel/usecase"
)

type Handler struct {
	Hub     *Hub
	WSAuth  usecase.WSAuthDeps
	Game    usecase.GameDeps
	Allowed map[string]struct{}
	Redis   model.IWSSessionStore
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
				h.writeError(conn, model.CodeUnauthorized, "authentication required")
				continue
			}
			token := ""
			if ck, err := c.Request().Cookie(middleware.AccessCookieName); err == nil {
				token = ck.Value
			}
			out, err := usecase.AuthenticateWebSocket(ctx, h.WSAuth, token)
			if err != nil {
				h.writeDomainError(conn, err)
				continue
			}
			userID = out.UserID
			wsLogID, err = usecase.RecordWSConnection(ctx, h.WSAuth, userID, connID)
			if err != nil {
				h.writeDomainError(conn, err)
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
			usecase.NotifyOpponentConnected(ctx, h.WSAuth, userID)
			continue
		}

		switch msgType {
		case "PING":
			usecase.TouchWSActivity(ctx, h.WSAuth, userID)
			_ = h.writeJSON(conn, map[string]any{"type": "PONG"})
		case "SET_SECRET":
			usecase.TouchWSActivity(ctx, h.WSAuth, userID)
			gameID, err := uuid.Parse(msg.GameID)
			if err != nil {
				h.Hub.SendError(userID, model.CodeValidation, "invalid gameId")
				continue
			}
			if err := usecase.SetSecretNumber(ctx, h.Game, userID, gameID, msg.SecretNumber); err != nil {
				h.sendDomainError(userID, err)
			}
		case "GUESS":
			usecase.TouchWSActivity(ctx, h.WSAuth, userID)
			gameID, err := uuid.Parse(msg.GameID)
			if err != nil {
				h.Hub.SendError(userID, model.CodeValidation, "invalid gameId")
				continue
			}
			if err := usecase.SubmitGuess(ctx, h.Game, userID, gameID, msg.GuessNumber, false); err != nil {
				h.sendDomainError(userID, err)
			}
		case "SYNC_REQUEST":
			usecase.TouchWSActivity(ctx, h.WSAuth, userID)
			gameID, err := uuid.Parse(msg.GameID)
			if err != nil {
				h.Hub.SendError(userID, model.CodeValidation, "invalid gameId")
				continue
			}
			if _, err := usecase.SyncGameState(ctx, h.Game, userID, gameID); err != nil {
				h.sendDomainError(userID, err)
			}
		default:
			h.Hub.SendError(userID, model.CodeValidation, "unknown event type")
		}
	}
}

func (h *Handler) onDisconnect(ctx context.Context, userID, wsLogID uuid.UUID) {
	h.Hub.Disconnect(userID)
	if h.Redis != nil {
		_ = h.Redis.DeleteUser(ctx, userID)
	}
	usecase.CloseWSConnectionLog(ctx, h.WSAuth, wsLogID)
	usecase.NotifyOpponentDisconnected(ctx, h.WSAuth, userID)
}

func (h *Handler) sendDomainError(userID uuid.UUID, err error) {
	if de, ok := model.IsDomainError(err); ok {
		h.Hub.SendError(userID, de.Code, de.Error())
		return
	}
	h.Hub.SendError(userID, model.CodeInternalError, "internal server error")
}

func (h *Handler) writeDomainError(conn *gorillaws.Conn, err error) {
	if de, ok := model.IsDomainError(err); ok {
		_ = h.writeJSON(conn, map[string]any{
			"type": "ERROR", "data": map[string]string{"code": de.Code, "message": de.Error()},
		})
		return
	}
	_ = h.writeJSON(conn, map[string]any{
		"type": "ERROR", "data": map[string]string{"code": model.CodeInternalError, "message": "internal server error"},
	})
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
