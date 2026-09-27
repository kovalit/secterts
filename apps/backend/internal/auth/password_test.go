package auth

import "testing"

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if hash == "correct horse battery staple" {
		t.Fatalf("hash must not equal plaintext")
	}

	ok, err := VerifyPassword("correct horse battery staple", hash)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if !ok {
		t.Fatalf("expected password to verify")
	}

	ok, err = VerifyPassword("wrong password", hash)
	if err != nil {
		t.Fatalf("verify wrong: %v", err)
	}
	if ok {
		t.Fatalf("expected wrong password to fail")
	}
}

func TestVerifyRejectsMalformedHash(t *testing.T) {
	if _, err := VerifyPassword("x", "not-a-real-hash"); err == nil {
		t.Fatalf("expected error for malformed hash")
	}
}

func TestHashProducesUniqueSalts(t *testing.T) {
	h1, _ := HashPassword("same")
	h2, _ := HashPassword("same")
	if h1 == h2 {
		t.Fatalf("expected different hashes due to random salt")
	}
}
