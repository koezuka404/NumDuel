// WebSocket Hub: 接続管理と Server → Client 通知
package websocket

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	gorillaws "github.com/gorilla/websocket"

	"github.com/numduel/numduel/model"
)

type Hub struct {
	mu      sync.RWMutex
	clients map[uuid.UUID]*clientConn
}

type clientConn struct {
	userID uuid.UUID
	connID string
	conn   *gorillaws.Conn
}

var _ model.EventNotifier = (*Hub)(nil)

func NewHub() *Hub {
	return &Hub{clients: make(map[uuid.UUID]*clientConn)}
}

func (h *Hub) Register(userID uuid.UUID, connectionID string, conn *gorillaws.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if old, ok := h.clients[userID]; ok && old.conn != nil {
		_ = old.conn.Close()
	}
	h.clients[userID] = &clientConn{userID: userID, connID: connectionID, conn: conn}
}

func (h *Hub) Disconnect(userID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if c, ok := h.clients[userID]; ok {
		if c.conn != nil {
			_ = c.conn.Close()
		}
		delete(h.clients, userID)
	}
}

func (h *Hub) SendToUser(_ context.Context, userID uuid.UUID, eventType string, payload map[string]any) error {
	h.mu.RLock()
	c, ok := h.clients[userID]
	h.mu.RUnlock()
	if !ok || c.conn == nil {
		return nil
	}
	b, err := json.Marshal(map[string]any{"type": eventType, "data": payload})
	if err != nil {
		return err
	}
	return c.conn.WriteMessage(gorillaws.TextMessage, b)
}

func (h *Hub) SendRaw(userID uuid.UUID, v any) error {
	h.mu.RLock()
	c, ok := h.clients[userID]
	h.mu.RUnlock()
	if !ok || c.conn == nil {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.conn.WriteMessage(gorillaws.TextMessage, b)
}

func (h *Hub) SendError(userID uuid.UUID, code, message string) {
	_ = h.SendRaw(userID, map[string]any{
		"type": "ERROR",
		"data": map[string]string{"code": code, "message": message},
	})
}

func (h *Hub) SendPong(userID uuid.UUID) {
	_ = h.SendRaw(userID, map[string]any{"type": "PONG"})
}

// SessionStore は Hub 切断 + Redis キー削除をまとめる
type SessionStore struct {
	Hub   *Hub
	Redis model.WSSessionStore
}

var _ model.WSSessionStore = (*SessionStore)(nil)

func NewSessionStore(hub *Hub, redis model.WSSessionStore) *SessionStore {
	return &SessionStore{Hub: hub, Redis: redis}
}

func (s *SessionStore) SetUser(ctx context.Context, userID uuid.UUID, connectionID string, ttl time.Duration) error {
	if s.Redis == nil {
		return nil
	}
	return s.Redis.SetUser(ctx, userID, connectionID, ttl)
}

func (s *SessionStore) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	s.Hub.Disconnect(userID)
	if s.Redis == nil {
		return nil
	}
	return s.Redis.DeleteUser(ctx, userID)
}
