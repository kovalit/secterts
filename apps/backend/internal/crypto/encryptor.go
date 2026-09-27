// Package crypto implements the AES-256-GCM encryption layer for Secrets Center.
//
// All password values, app-secret values and comments are stored in PostgreSQL
// only in encrypted form. The master key lives outside the database, in the
// APP_MASTER_KEY_V1_BASE64 environment variable.
//
// Rule: never log plaintext values from this package.
package crypto

// Encryptor encrypts and decrypts short string values with a versioned key.
type Encryptor interface {
	// EncryptString encrypts a plaintext string, returning the ciphertext, the
	// random nonce, and the key version used.
	EncryptString(plain string) (ciphertext []byte, nonce []byte, keyVersion int, err error)
	// DecryptString reverses EncryptString for the given key version.
	DecryptString(ciphertext []byte, nonce []byte, keyVersion int) (string, error)
}

// keyringEncryptor is the default Encryptor backed by a Keyring.
type keyringEncryptor struct {
	ring *Keyring
}

// NewEncryptor creates an Encryptor from a base64-encoded 32-byte master key.
func NewEncryptor(masterKeyV1Base64 string) (Encryptor, error) {
	ring, err := NewKeyringFromBase64(masterKeyV1Base64)
	if err != nil {
		return nil, err
	}
	return &keyringEncryptor{ring: ring}, nil
}

func (e *keyringEncryptor) EncryptString(plain string) ([]byte, []byte, int, error) {
	version := e.ring.ActiveVersion()
	key, err := e.ring.keyFor(version)
	if err != nil {
		return nil, nil, 0, err
	}
	ciphertext, nonce, err := encryptAESGCM(key, []byte(plain))
	if err != nil {
		return nil, nil, 0, err
	}
	return ciphertext, nonce, version, nil
}

func (e *keyringEncryptor) DecryptString(ciphertext, nonce []byte, keyVersion int) (string, error) {
	key, err := e.ring.keyFor(keyVersion)
	if err != nil {
		return "", err
	}
	plaintext, err := decryptAESGCM(key, ciphertext, nonce)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
