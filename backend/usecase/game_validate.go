package usecase

import "github.com/numduel/numduel/model"

func ValidateFourDigits(input string) ([4]int, error) {
	if len(input) != 4 {
		return [4]int{}, model.ErrBadDigitLength
	}
	seen := map[int]struct{}{}
	var digits [4]int
	for i, ch := range input {
		if ch < '0' || ch > '9' {
			return [4]int{}, model.ErrBadDigit
		}
		d := int(ch - '0')
		if _, ok := seen[d]; ok {
			return [4]int{}, model.ErrDuplicateDigit
		}
		seen[d] = struct{}{}
		digits[i] = d
	}
	return digits, nil
}

func ValidateFourDigitsArray(digits [4]int) error {
	seen := map[int]struct{}{}
	for i := 0; i < 4; i++ {
		if digits[i] < 0 || digits[i] > 9 {
			return model.ErrBadDigit
		}
		if _, ok := seen[digits[i]]; ok {
			return model.ErrDuplicateDigit
		}
		seen[digits[i]] = struct{}{}
	}
	return nil
}
