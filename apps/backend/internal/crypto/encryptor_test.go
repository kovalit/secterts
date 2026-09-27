package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"testing"
)

func newTestKeyBase64(t *testing.T) string {
	t.Helper()
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("rand: %v", err)
	}
	return base64.StdEncoding.EncodeToString(key)
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	enc, err := NewEncryptor(newTestKeyBase64(t))
	if err != nil {
		t.Fatalf("new encryptor: %v", err)
	}

	cases := []string{"", "hunter2", "postgres://user:pass@host/db", "юникод пароль 🔐"}
	for _, plain := range cases {
		ciphertext, nonce, version, err := enc.EncryptString(plain)
		if err != nil {
			t.Fatalf("encrypt %q: %v", plain, err)
		}
		if version != 1 {
			t.Fatalf("expected key version 1, got %d", version)
		}
		if plain != "" && string(ciphertext) == plain {
			t.Fatalf("ciphertext must differ from plaintext")
		}
		if len(nonce) != 12 {
			t.Fatalf("expected 12-byte nonce, got %d", len(nonce))
		}

		got, err := enc.DecryptString(ciphertext, nonce, version)
		if err != nil {
			t.Fatalf("decrypt: %v", err)
		}
		if got != plain {
			t.Fatalf("round trip mismatch: got %q want %q", got, plain)
		}
	}
}

func TestEncryptProducesDistinctCiphertext(t *testing.T) {
	enc, err := NewEncryptor(newTestKeyBase64(t))
	if err != nil {
		t.Fatalf("new encryptor: %v", err)
	}

	c1, n1, _, _ := enc.EncryptString("same-value")
	c2, n2, _, _ := enc.EncryptString("same-value")

	if string(n1) == string(n2) {
		t.Fatalf("nonces must be random per encryption")
	}
	if string(c1) == string(c2) {
		t.Fatalf("ciphertext must differ due to random nonce")
	}
}

func TestDecryptWithWrongKeyFails(t *testing.T) {
	enc1, _ := NewEncryptor(newTestKeyBase64(t))
	enc2, _ := NewEncryptor(newTestKeyBase64(t))

	ciphertext, nonce, version, err := enc1.EncryptString("top secret")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	if _, err := enc2.DecryptString(ciphertext, nonce, version); err == nil {
		t.Fatalf("expected decryption with wrong key to fail")
	}
}

func TestDecryptTamperedCiphertextFails(t *testing.T) {
	enc, _ := NewEncryptor(newTestKeyBase64(t))
	ciphertext, nonce, version, _ := enc.EncryptString("tamper me")

	ciphertext[0] ^= 0xFF
	if _, err := enc.DecryptString(ciphertext, nonce, version); err == nil {
		t.Fatalf("expected GCM auth failure on tampered ciphertext")
	}
}

func TestNewEncryptorRejectsBadKey(t *testing.T) {
	if _, err := NewEncryptor("not-base64!!!"); err == nil {
		t.Fatalf("expected error for invalid base64 key")
	}
	shortKey := base64.StdEncoding.EncodeToString([]byte("too-short"))
	if _, err := NewEncryptor(shortKey); err == nil {
		t.Fatalf("expected error for wrong key length")
	}
}

func TestDecryptUnknownKeyVersionFails(t *testing.T) {
	enc, _ := NewEncryptor(newTestKeyBase64(t))
	ciphertext, nonce, _, _ := enc.EncryptString("value")
	if _, err := enc.DecryptString(ciphertext, nonce, 99); err == nil {
		t.Fatalf("expected error for unknown key version")
	}
}
