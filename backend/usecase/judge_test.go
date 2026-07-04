package usecase

import (
	"testing"

	"github.com/numduel/numduel/model"
)

// §18.4.1 JudgeDigits
func TestJudgeDigits(t *testing.T) {
	tests := []struct {
		name     string
		secret   [4]int
		guess    [4]int
		expected [4]model.DigitResult
		hits     int
	}{
		{"all hit", [4]int{1, 2, 3, 4}, [4]int{1, 2, 3, 4}, [4]model.DigitResult{model.DigitHit, model.DigitHit, model.DigitHit, model.DigitHit}, 4},
		{"all miss", [4]int{1, 2, 3, 4}, [4]int{5, 6, 7, 8}, [4]model.DigitResult{model.DigitMiss, model.DigitMiss, model.DigitMiss, model.DigitMiss}, 0},
		{"one hit", [4]int{1, 2, 3, 4}, [4]int{1, 4, 5, 6}, [4]model.DigitResult{model.DigitHit, model.DigitMiss, model.DigitMiss, model.DigitMiss}, 1},
		{"position mismatch all miss", [4]int{1, 2, 3, 4}, [4]int{4, 3, 2, 1}, [4]model.DigitResult{model.DigitMiss, model.DigitMiss, model.DigitMiss, model.DigitMiss}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := JudgeDigits(tt.secret, tt.guess)
			if got != tt.expected {
				t.Fatalf("JudgeDigits() = %v, want %v", got, tt.expected)
			}
			if n := HitCount(got); n != tt.hits {
				t.Fatalf("HitCount() = %d, want %d", n, tt.hits)
			}
		})
	}
}

// §18.4.2 IsWin
func TestIsWin(t *testing.T) {
	tests := []struct {
		hits int
		win  bool
	}{
		{4, true},
		{3, false},
		{0, false},
	}
	for _, tt := range tests {
		results := [4]model.DigitResult{}
		for i := 0; i < tt.hits; i++ {
			results[i] = model.DigitHit
		}
		if got := IsWin(results); got != tt.win {
			t.Fatalf("IsWin(hits=%d) = %v, want %v", tt.hits, got, tt.win)
		}
	}
}
