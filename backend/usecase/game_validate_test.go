package usecase

import (
	"errors"
	"testing"

	"github.com/numduel/numduel/model"
)

// §18.4.3 NewSecretNumber / 4-digit validation
func TestValidateFourDigits(t *testing.T) {
	tests := []struct {
		input string
		ok    bool
		err   error
	}{
		{"1234", true, nil},
		{"123", false, model.ErrBadDigitLength},
		{"12345", false, model.ErrBadDigitLength},
		{"12a4", false, model.ErrBadDigit},
		{"1123", false, model.ErrDuplicateDigit},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := ValidateFourDigits(tt.input)
			if tt.ok {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if !errors.Is(err, tt.err) {
				t.Fatalf("error = %v, want %v", err, tt.err)
			}
		})
	}
}

func TestValidateFourDigitsArray(t *testing.T) {
	if err := ValidateFourDigitsArray([4]int{1, 2, 3, 4}); err != nil {
		t.Fatalf("valid array: %v", err)
	}
	if err := ValidateFourDigitsArray([4]int{1, 1, 2, 3}); !errors.Is(err, model.ErrDuplicateDigit) {
		t.Fatalf("duplicate: %v", err)
	}
	if err := ValidateFourDigitsArray([4]int{-1, 2, 3, 4}); !errors.Is(err, model.ErrBadDigit) {
		t.Fatalf("bad digit: %v", err)
	}
}
