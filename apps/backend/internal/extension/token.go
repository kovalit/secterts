// Package extension implements the Chrome extension API: token management
// (behind a web session) and bearer-authenticated lookup / reveal by domain.
package extension

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

// tokenBytes is the entropy of an extension token.
const tokenBytes = 32

// generateToken returns a random URL-safe token.
func generateToken() (string, error) {
	b := make([]byte, tokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "sce_" + base64.RawURLEncoding.EncodeToString(b), nil
}

// hashToken returns the hex SHA-256 of a token. Only hashes are stored.
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
