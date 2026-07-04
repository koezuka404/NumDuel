package crypto

import (
	"testing"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
)

func TestSecretHashServiceVerify(t *testing.T) {
	svc, err := NewSecretHashService("abcdefghijklmnopqrstuvwxyz1234567890abcd")
	if err != nil {
		t.Fatalf("NewSecretHashService: %v", err)
	}
	gameID := uuid.New()
	hash, err := svc.Hash([4]int{1, 2, 3, 4}, gameID, 1)
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}

	results, err := svc.Verify(hash, [4]int{1, 2, 3, 4}, gameID, 1)
	if err != nil {
		t.Fatalf("Verify exact: %v", err)
	}
	if hitAll(results) != 4 {
		t.Fatalf("expected 4 hits, got %d", hitAll(results))
	}

	results, err = svc.Verify(hash, [4]int{5, 6, 7, 8}, gameID, 1)
	if err != nil {
		t.Fatalf("Verify miss: %v", err)
	}
	if hitAll(results) != 0 {
		t.Fatalf("expected 0 hits, got %d", hitAll(results))
	}
}

func hitAll(results []model.DigitResult) int {
	n := 0
	for _, r := range results {
		if r == model.DigitHit {
			n++
		}
	}
	return n
}
