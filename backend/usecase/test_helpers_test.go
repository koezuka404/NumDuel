package usecase_test

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/usecase"
)

type captureNotifier struct {
	mu    sync.Mutex
	calls []notifyCall
}

type notifyCall struct {
	UserID    uuid.UUID
	EventType string
	Payload   map[string]any
}

func (n *captureNotifier) SendToUser(_ context.Context, userID uuid.UUID, eventType string, payload map[string]any) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.calls = append(n.calls, notifyCall{UserID: userID, EventType: eventType, Payload: payload})
	return nil
}

func (n *captureNotifier) last() *notifyCall {
	n.mu.Lock()
	defer n.mu.Unlock()
	if len(n.calls) == 0 {
		return nil
	}
	c := n.calls[len(n.calls)-1]
	return &c
}

type memBackupStatusStore struct {
	mu     sync.Mutex
	status usecase.BackupStatus
	err    error
}

func (m *memBackupStatusStore) GetBackupStatus(_ context.Context) (*usecase.BackupStatus, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return nil, m.err
	}
	st := m.status
	return &st, nil
}

func (m *memBackupStatusStore) SetBackupStatus(_ context.Context, status string, lastSyncedAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.status.Status = status
	t := lastSyncedAt
	m.status.LastSyncedAt = &t
	return nil
}

type memLockStore struct {
	mu      sync.Mutex
	locked  map[string]bool
	failKey string
	errKey  string
}

func newMemLockStore() *memLockStore {
	return &memLockStore{locked: make(map[string]bool)}
}

func (m *memLockStore) AcquireLock(_ context.Context, key string, _ time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.errKey != "" && key == m.errKey {
		return false, errLockAcquire
	}
	if m.failKey != "" && key == m.failKey {
		return false, nil
	}
	if m.locked[key] {
		return false, nil
	}
	m.locked[key] = true
	return true, nil
}

var errLockAcquire = context.Canceled

type memWSSessionStore struct {
	mu      sync.Mutex
	deleted []uuid.UUID
}

func (m *memWSSessionStore) SetUser(_ context.Context, _ uuid.UUID, _ string, _ time.Duration) error {
	return nil
}

func (m *memWSSessionStore) DeleteUser(_ context.Context, userID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deleted = append(m.deleted, userID)
	return nil
}

func (m *memWSSessionStore) wasDeleted(userID uuid.UUID) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, id := range m.deleted {
		if id == userID {
			return true
		}
	}
	return false
}

type failingJWTRevoker struct{}

func (failingJWTRevoker) Revoke(_ context.Context, _ string, _ time.Duration) error {
	return context.Canceled
}

func (failingJWTRevoker) IsRevoked(_ context.Context, _ string) (bool, error) {
	return false, nil
}
