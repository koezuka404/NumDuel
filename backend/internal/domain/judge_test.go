package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJudgeDigits(t *testing.T) {
	tests := []struct {
		name     string
		secret   [4]int
		guess    [4]int
		want     [4]DigitResult
		hitCount int
	}{
		{"all hit", [4]int{1, 2, 3, 4}, [4]int{1, 2, 3, 4}, [4]DigitResult{DigitHit, DigitHit, DigitHit, DigitHit}, 4},
		{"all miss", [4]int{1, 2, 3, 4}, [4]int{5, 6, 7, 8}, [4]DigitResult{DigitMiss, DigitMiss, DigitMiss, DigitMiss}, 0},
		{"one hit", [4]int{1, 2, 3, 4}, [4]int{1, 4, 5, 6}, [4]DigitResult{DigitHit, DigitMiss, DigitMiss, DigitMiss}, 1},
		{"wrong position only", [4]int{1, 2, 3, 4}, [4]int{4, 3, 2, 1}, [4]DigitResult{DigitMiss, DigitMiss, DigitMiss, DigitMiss}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := JudgeDigits(tt.secret, tt.guess)
			if got != tt.want {
				t.Fatalf("JudgeDigits() = %v, want %v", got, tt.want)
			}
			if HitCount(got) != tt.hitCount {
				t.Fatalf("HitCount() = %d, want %d", HitCount(got), tt.hitCount)
			}
		})
	}
}

func TestIsWin(t *testing.T) {
	if !IsWin([4]DigitResult{DigitHit, DigitHit, DigitHit, DigitHit}) {
		t.Fatal("expected win")
	}
	if IsWin([4]DigitResult{DigitHit, DigitHit, DigitHit, DigitMiss}) {
		t.Fatal("expected not win")
	}
}

func TestNewSecretNumber(t *testing.T) {
	if _, err := NewSecretNumberFromString("123"); err == nil {
		t.Fatal("expected invalid length")
	}
	_, err := NewSecretNumberFromString("123")
	if de, ok := IsDomainError(err); !ok || de.Code != CodeInvalidDigitLength {
		t.Fatalf("got err=%v", err)
	}
	if _, err := NewSecretNumberFromString("12a4"); err == nil {
		t.Fatal("expected invalid digit")
	}
	if _, err := NewSecretNumberFromString("1123"); err == nil {
		t.Fatal("expected duplicate digit")
	}
	if _, err := NewSecretNumberFromString("1234"); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestUserMethods(t *testing.T) {
	u := &User{Role: RoleUser}
	if u.IsDeleted() || u.IsMaster() || !u.CanMatch() {
		t.Fatal("active user should match")
	}
	now := time.Now()
	u.DeletedAt = &now
	if !u.IsDeleted() || u.CanMatch() {
		t.Fatal("deleted user should not match")
	}
	m := &User{Role: RoleMaster}
	if !m.IsMaster() || m.CanMatch() {
		t.Fatal("master should not match")
	}
}

func TestGameSetSecretAndStart(t *testing.T) {
	p1, p2 := uuid.New(), uuid.New()
	now := time.Now()
	g := &Game{
		ID: uuid.New(), Status: GameStatusWaitingSecret,
		Player1ID: p1, Player2ID: p2, CurrentTurn: 1,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := g.SetSecretHash(p1, "hash1"); err != nil {
		t.Fatal(err)
	}
	if g.BothSecretsSet() {
		t.Fatal("only one secret set")
	}
	if err := g.SetSecretHash(p2, "hash2"); err != nil {
		t.Fatal(err)
	}
	if !g.BothSecretsSet() {
		t.Fatal("both secrets should be set")
	}
	if err := g.Start(now); err != nil {
		t.Fatal(err)
	}
	if g.Status != GameStatusInProgress || !g.IsCurrentTurn(p1) || !g.CanGuess(p1) || g.CanGuess(p2) {
		t.Fatal("player1 should have first turn")
	}
}

func TestGameAddGuessAdvancesTurn(t *testing.T) {
	p1, p2 := uuid.New(), uuid.New()
	now := time.Now()
	g := &Game{
		ID: uuid.New(), Status: GameStatusInProgress,
		Player1ID: p1, Player2ID: p2, CurrentTurn: 1,
		CurrentTurnPlayerID: &p1, CreatedAt: now, UpdatedAt: now,
	}
	gn, _ := NewGuessNumberFromString("5678")
	results := JudgeDigits([4]int{1, 2, 3, 4}, gn.Digits())
	guess, err := g.AddGuess(p1, gn, results, false, now)
	if err != nil {
		t.Fatal(err)
	}
	if guess.Turn != 1 || guess.HitCount != 0 {
		t.Fatal("unexpected guess")
	}
	if g.CurrentTurn != 2 || !g.IsCurrentTurn(p2) {
		t.Fatal("turn should advance to player2")
	}
}

func TestGameCancelBySecretTimeout(t *testing.T) {
	now := time.Now()
	g := &Game{ID: uuid.New(), Status: GameStatusWaitingSecret, CreatedAt: now, UpdatedAt: now}
	if err := g.CancelBySecretTimeout(now); err != nil {
		t.Fatal(err)
	}
	if g.Status != GameStatusFinished || g.WinnerID != nil {
		t.Fatal("timeout finish should have no winner")
	}
}

func TestRefreshTokenRevoke(t *testing.T) {
	now := time.Now()
	tok := NewRefreshToken(uuid.New(), "hash", uuid.New(), now.Add(time.Hour), now)
	if !tok.IsActive(now) {
		t.Fatal("expected active")
	}
	tok.Revoke(now)
	if tok.IsActive(now) {
		t.Fatal("expected revoked")
	}
}
