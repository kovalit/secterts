package crypto

import (
	"encoding/base64"
	"fmt"
)

// keyLength is the required AES-256 key size in bytes.
const keyLength = 32

// Keyring maps a key version to its raw 32-byte key. Keys never touch the
// database — they are loaded from environment variables at startup.
type Keyring struct {
	keys          map[int][]byte
	activeVersion int
}

// NewKeyringFromBase64 builds a keyring with a single active key (version 1)
// decoded from a base64 string, as described in docs/01 (APP_MASTER_KEY_V1_BASE64).
func NewKeyringFromBase64(masterKeyV1Base64 string) (*Keyring, error) {
	key, err := base64.StdEncoding.DecodeString(masterKeyV1Base64)
	if err != nil {
		return nil, fmt.Errorf("decode master key base64: %w", err)
	}
	if len(key) != keyLength {
		return nil, fmt.Errorf("master key must be %d bytes, got %d", keyLength, len(key))
	}

	return &Keyring{
		keys:          map[int][]byte{1: key},
		activeVersion: 1,
	}, nil
}

// ActiveVersion returns the key version used for new encryptions.
func (k *Keyring) ActiveVersion() int {
	return k.activeVersion
}

// keyFor returns the raw key for a version, or an error if it is unknown.
func (k *Keyring) keyFor(version int) ([]byte, error) {
	key, ok := k.keys[version]
	if !ok {
		return nil, fmt.Errorf("unknown key version %d", version)
	}
	return key, nil
}
