package service

import (
	"encoding/hex"
	"testing"
)

func TestAccessTokenIsRandomAndHashed(t *testing.T) {
	a, err := newAccessToken()
	if err != nil {
		t.Fatal(err)
	}
	b, err := newAccessToken()
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Fatal("independent access tokens must differ")
	}
	if len(hashToken(a)) != 64 {
		t.Fatalf("expected SHA-256 hex digest, got %d characters", len(hashToken(a)))
	}
	if _, err := hex.DecodeString(hashToken(a)); err != nil {
		t.Fatalf("token hash is not hex: %v", err)
	}
	if hashToken(a) == a {
		t.Fatal("database token representation must not equal the raw token")
	}
}

func TestPasswordHashUsesBcrypt(t *testing.T) {
	hash, err := HashPassword("phase-seven-unique-test-password")
	if err != nil {
		t.Fatal(err)
	}
	if hash == "phase-seven-unique-test-password" || len(hash) < 50 {
		t.Fatal("password should be stored as a bcrypt hash")
	}
}
