package testutil

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/numduel/numduel/usecase"
)

// MemJWTRevoker is an in-memory JWT revocation store for security tests.
type MemJWTRevoker struct {
	mu      sync.Mutex
	revoked map[string]struct{}
}

var _ usecase.IJWTRevoker = (*MemJWTRevoker)(nil)

func NewMemJWTRevoker() *MemJWTRevoker {
	return &MemJWTRevoker{revoked: make(map[string]struct{})}
}

func (m *MemJWTRevoker) Revoke(_ context.Context, jti string, _ time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.revoked[jti] = struct{}{}
	return nil
}

func (m *MemJWTRevoker) IsRevoked(_ context.Context, jti string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.revoked[jti]
	return ok, nil
}

// MemForceLogout is an in-memory force-logout store for security tests.
type MemForceLogout struct {
	mu     sync.Mutex
	before map[uuid.UUID]time.Time
}

var _ usecase.IForceLogoutStore = (*MemForceLogout)(nil)

func NewMemForceLogout() *MemForceLogout {
	return &MemForceLogout{before: make(map[uuid.UUID]time.Time)}
}

func (m *MemForceLogout) SetForceLogoutBefore(_ context.Context, userID uuid.UUID, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.before[userID] = at.UTC()
	return nil
}

func (m *MemForceLogout) GetForceLogoutBefore(_ context.Context, userID uuid.UUID) (time.Time, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.before[userID], nil
}
