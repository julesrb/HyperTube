package auth

import "testing"

func TestHashPasswordAndCheckPassword(t *testing.T) {
	password := "correct-horse-battery"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if hash == password {
		t.Fatal("password hash must not contain the raw password")
	}
	if !CheckPassword(hash, password) {
		t.Fatal("expected password check to accept the original password")
	}
	if CheckPassword(hash, "wrong-password") {
		t.Fatal("expected password check to reject the wrong password")
	}
}
