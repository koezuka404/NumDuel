package crypto

import "testing"

func TestPasswordServiceRejectsWrongPassword(t *testing.T) {
	svc := NewPasswordService()
	hash, err := svc.Hash("password123")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if svc.Verify(hash, "wrongpass1") {
		t.Fatalf("wrong password should not verify")
	}
	if !svc.Verify(hash, "password123") {
		t.Fatalf("correct password should verify")
	}
}

func TestPasswordServiceDifferentHashesForSamePassword(t *testing.T) {
	svc := NewPasswordService()
	h1, err := svc.Hash("password123")
	if err != nil {
		t.Fatalf("hash1: %v", err)
	}
	h2, err := svc.Hash("password123")
	if err != nil {
		t.Fatalf("hash2: %v", err)
	}
	if h1 == h2 {
		t.Fatalf("bcrypt hashes should differ due to salt")
	}
}
