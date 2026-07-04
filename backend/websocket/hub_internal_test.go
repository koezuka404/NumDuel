package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	gorillaws "github.com/gorilla/websocket"
	"github.com/google/uuid"
)

func TestHubSendToUser(t *testing.T) {
	hub, userID, serverConn, clientConn := setupHubConnPair(t)

	done := make(chan map[string]any, 1)
	go func() {
		_ = clientConn.SetReadDeadline(time.Now().Add(3 * time.Second))
		_, raw, err := clientConn.ReadMessage()
		if err != nil {
			t.Errorf("read: %v", err)
			return
		}
		var msg map[string]any
		if err := json.Unmarshal(raw, &msg); err != nil {
			t.Errorf("json: %v", err)
			return
		}
		done <- msg
	}()

	if err := hub.SendToUser(context.Background(), userID, "CUSTOM_EVENT", map[string]any{"score": 42}); err != nil {
		t.Fatalf("SendToUser: %v", err)
	}
	select {
	case msg := <-done:
		if msg["type"] != "CUSTOM_EVENT" {
			t.Fatalf("type: %+v", msg)
		}
		data, _ := msg["data"].(map[string]any)
		if data["score"].(float64) != 42 {
			t.Fatalf("payload: %+v", data)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for SendToUser message")
	}
	_ = serverConn
}

func TestHubSendRawAndPong(t *testing.T) {
	hub, userID, _, clientConn := setupHubConnPair(t)

	go drainConn(clientConn)
	if err := hub.SendRaw(userID, map[string]any{"type": "CUSTOM"}); err != nil {
		t.Fatalf("SendRaw: %v", err)
	}
	hub.SendPong(userID)
	hub.SendError(userID, "validation_error", "bad")
	hub.Disconnect(userID)
}

func TestHubSendRawMarshalError(t *testing.T) {
	hub, userID, _, _ := setupHubConnPair(t)
	if err := hub.SendRaw(userID, map[string]any{"bad": make(chan int)}); err == nil {
		t.Fatal("expected marshal error")
	}
}

func TestHubSendToUserMarshalError(t *testing.T) {
	hub, userID, _, _ := setupHubConnPair(t)
	err := hub.SendToUser(context.Background(), userID, "EV", map[string]any{"bad": make(chan int)})
	if err == nil {
		t.Fatal("expected marshal error")
	}
}

type failingJSON struct{}

func (failingJSON) MarshalJSON() ([]byte, error) {
	return nil, errors.New("marshal failed")
}

func TestHubSendRawCustomMarshalError(t *testing.T) {
	hub, userID, _, _ := setupHubConnPair(t)
	if err := hub.SendRaw(userID, failingJSON{}); err == nil {
		t.Fatal("expected marshal error")
	}
}

func TestHubRegisterReplacesOldConnection(t *testing.T) {
	hub := NewHub()
	userID := uuid.New()
	oldServerCh := make(chan *gorillaws.Conn, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		up := gorillaws.Upgrader{}
		c, err := up.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		oldServerCh <- c
		<-r.Context().Done()
	}))
	t.Cleanup(srv.Close)
	url := "ws" + strings.TrimPrefix(srv.URL, "http")
	oldClient, _, err := gorillaws.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial old: %v", err)
	}
	oldServer := <-oldServerCh
	hub.Register(userID, "old", oldServer)

	newServerCh := make(chan *gorillaws.Conn, 1)
	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		up := gorillaws.Upgrader{}
		c, err := up.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		newServerCh <- c
		<-r.Context().Done()
	}))
	t.Cleanup(srv2.Close)
	url2 := "ws" + strings.TrimPrefix(srv2.URL, "http")
	newClient, _, err := gorillaws.DefaultDialer.Dial(url2, nil)
	if err != nil {
		t.Fatalf("dial new: %v", err)
	}
	t.Cleanup(func() { _ = newClient.Close() })
	hub.Register(userID, "new", <-newServerCh)

	if _, _, err := oldClient.ReadMessage(); err == nil {
		t.Fatal("old connection should be closed")
	}
}

func dialTestConn(t *testing.T) *gorillaws.Conn {
	t.Helper()
	_, _, _, clientConn := setupHubConnPair(t)
	return clientConn
}

func setupHubConnPair(t *testing.T) (*Hub, uuid.UUID, *gorillaws.Conn, *gorillaws.Conn) {
	t.Helper()
	hub := NewHub()
	userID := uuid.New()
	serverConnCh := make(chan *gorillaws.Conn, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		up := gorillaws.Upgrader{}
		c, err := up.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		serverConnCh <- c
		<-r.Context().Done()
	}))
	t.Cleanup(srv.Close)
	url := "ws" + strings.TrimPrefix(srv.URL, "http")
	clientConn, _, err := gorillaws.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = clientConn.Close() })
	serverConn := <-serverConnCh
	hub.Register(userID, "c1", serverConn)
	return hub, userID, serverConn, clientConn
}

func drainConn(conn *gorillaws.Conn) {
	for {
		_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}
