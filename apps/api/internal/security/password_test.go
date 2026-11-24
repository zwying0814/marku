package security

import "testing"

func TestPasswordHashAndCheck(t *testing.T) {
	hash, err := HashPassword("correct-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if hash == "correct-password" {
		t.Fatal("password hash must not equal plaintext")
	}
	if !CheckPassword(hash, "correct-password") {
		t.Fatal("expected password to match")
	}
	if CheckPassword(hash, "wrong-password") {
		t.Fatal("unexpected password match")
	}
}
