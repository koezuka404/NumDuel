package usecase

import (
	"testing"

	"github.com/numduel/numduel/model"
)

func TestValidateFourDigits(t *testing.T) {
	digits, err := ValidateFourDigits("1234")
	if err != nil || digits != [4]int{1, 2, 3, 4} {
		t.Fatalf("valid digits: %v %v", digits, err)
	}
	if _, err = ValidateFourDigits("123"); err == nil {
		t.Fatal("expected error for short input")
	}
}

func TestIsWin(t *testing.T) {
	if !IsWin([4]model.DigitResult{model.DigitHit, model.DigitHit, model.DigitHit, model.DigitHit}) {
		t.Fatal("four hits should win")
	}
}
