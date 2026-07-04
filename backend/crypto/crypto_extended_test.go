package crypto

import (
	"testing"

	"github.com/google/uuid"

	"github.com/numduel/numduel/model"
)

func TestRandomNumberServiceGenerate(t *testing.T) {
	svc := NewRandomNumberService()
	for i := 0; i < 20; i++ {
		n, err := svc.GenerateGuessNumber()
		if err != nil || len(n) != 4 {
			t.Fatalf("generate: %q err=%v", n, err)
		}
		seen := map[rune]bool{}
		for _, ch := range n {
			if ch < '0' || ch > '9' || seen[ch] {
				t.Fatalf("invalid digit string %q", n)
			}
			seen[ch] = true
		}
	}
}

func TestRandIntInvalidRange(t *testing.T) {
	_, err := randInt(5, 4)
	if err == nil {
		t.Fatal("expected range error")
	}
}

func TestSecretHashServiceInvalidPepper(t *testing.T) {
	_, err := NewSecretHashService("short")
	if err == nil {
		t.Fatal("expected pepper length error")
	}
}

func TestSecretHashWrongGameOrSlot(t *testing.T) {
	svc, err := NewSecretHashService("abcdefghijklmnopqrstuvwxyz1234567890abcd")
	if err != nil {
		t.Fatalf("NewSecretHashService: %v", err)
	}
	gameID := uuid.New()
	hash, err := svc.Hash([4]int{1, 2, 3, 4}, gameID, 1)
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	results, err := svc.Verify(hash, [4]int{1, 2, 3, 4}, uuid.New(), 1)
	if err != nil || hitCount(results) != 0 {
		t.Fatalf("wrong game id should miss: %v results=%v", err, results)
	}
	results, err = svc.Verify(hash, [4]int{1, 2, 3, 4}, gameID, 2)
	if err != nil || hitCount(results) != 0 {
		t.Fatalf("wrong slot should miss: %v results=%v", err, results)
	}
	_, err = svc.Verify("bad", [4]int{1, 2, 3, 4}, gameID, 1)
	if err == nil {
		t.Fatal("invalid hash format")
	}
}

func hitCount(results []model.DigitResult) int {
	n := 0
	for _, r := range results {
		if r == model.DigitHit {
			n++
		}
	}
	return n
}

func TestRefreshTokenServiceGenerate(t *testing.T) {
	svc := NewRefreshTokenService()
	pair, err := svc.Generate()
	if err != nil || pair.Plaintext == "" || pair.Hash == "" || pair.Hash != svc.Hash(pair.Plaintext) {
		t.Fatalf("Generate: %+v err=%v", pair, err)
	}
}
