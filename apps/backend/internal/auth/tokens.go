package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"math/big"
)

// generateToken returns a cryptographically random URL-safe token string.
func generateToken(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// hashToken returns the hex SHA-256 of a token. Only the hash is stored.
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// generateNumericCode returns a zero-padded numeric code of n digits.
func generateNumericCode(n int) (string, error) {
	const digits = "0123456789"
	out := make([]byte, n)
	for i := range out {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		out[i] = digits[idx.Int64()]
	}
	return string(out), nil
}
