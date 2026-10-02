package passwords

import (
	"strings"
	"unicode"
)

// WeakStrengthThreshold is the inclusive score at/below which a password is
// considered weak by the secret-health check.
const WeakStrengthThreshold = 1

// EstimateStrength returns a 0..4 strength score for a plaintext password,
// loosely following the common "very weak … very strong" scale. The score is
// stored on write so the health check never needs to decrypt passwords.
//
//	0 - very weak    1 - weak    2 - fair    3 - strong    4 - very strong
func EstimateStrength(password string) int {
	if password == "" {
		return 0
	}

	var (
		lower, upper, digit, symbol bool
	)
	for _, r := range password {
		switch {
		case unicode.IsLower(r):
			lower = true
		case unicode.IsUpper(r):
			upper = true
		case unicode.IsDigit(r):
			digit = true
		default:
			symbol = true
		}
	}

	classes := 0
	for _, ok := range []bool{lower, upper, digit, symbol} {
		if ok {
			classes++
		}
	}

	length := len([]rune(password))

	// Obvious weak passwords are capped regardless of composition.
	if length < 8 || isCommonPassword(password) {
		return 0
	}

	score := 0
	switch {
	case length >= 16:
		score += 2
	case length >= 12:
		score++
	}
	score += classes - 1 // 0..3 for the extra character classes

	if score < 0 {
		score = 0
	}
	if score > 4 {
		score = 4
	}
	return score
}

// commonPasswords is a small blocklist of obviously weak values; the health
// check treats any of these (case-insensitive) as very weak.
var commonPasswords = map[string]struct{}{
	"password": {}, "passw0rd": {}, "123456": {}, "12345678": {}, "123456789": {},
	"qwerty": {}, "qwerty123": {}, "111111": {}, "123123": {}, "abc123": {},
	"letmein": {}, "iloveyou": {}, "admin": {}, "welcome": {}, "monkey": {},
	"dragon": {}, "secret": {}, "master": {}, "000000": {}, "football": {},
}

func isCommonPassword(password string) bool {
	_, ok := commonPasswords[strings.ToLower(strings.TrimSpace(password))]
	return ok
}
