package crypto

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "testpassword123"

	// Test hashing
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if hash == "" {
		t.Fatal("Hash should not be empty")
	}

	if hash == password {
		t.Fatal("Hash should not be the same as password")
	}
}

func TestCheckPassword(t *testing.T) {
	password := "testpassword123"
	wrongPassword := "wrongpassword"

	// Hash the password
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Test correct password
	if !CheckPassword(password, hash) {
		t.Fatal("CheckPassword should return true for correct password")
	}

	// Test wrong password
	if CheckPassword(wrongPassword, hash) {
		t.Fatal("CheckPassword should return false for wrong password")
	}
}

func TestHashPasswordDifferentResults(t *testing.T) {
	password := "testpassword123"

	// Hash the same password twice
	hash1, err1 := HashPassword(password)
	if err1 != nil {
		t.Fatalf("Failed to hash password: %v", err1)
	}

	hash2, err2 := HashPassword(password)
	if err2 != nil {
		t.Fatalf("Failed to hash password: %v", err2)
	}

	// Hashes should be different due to salt
	if hash1 == hash2 {
		t.Fatal("Two hashes of the same password should be different")
	}

	// But both should validate correctly
	if !CheckPassword(password, hash1) {
		t.Fatal("First hash should validate correctly")
	}

	if !CheckPassword(password, hash2) {
		t.Fatal("Second hash should validate correctly")
	}
}