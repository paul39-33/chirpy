package auth

import (
	"testing"
)

//check success path
func TestPassHash(t *testing.T) {
	pwHash, err := HashPassword("Test12345")
	if err != nil {
		t.Fatalf("Error hashing password: %v", err)
	}
	err = CheckPasswordHash("Test12345", pwHash)
	if err != nil {
		t.Errorf("Error comparing password and hashed password: %v", err)
	}
}

//check failure path
func TestPassHashFail(t *testing.T){
	pwHash, err := HashPassword("Test2")
	if err != nil {
		t.Fatalf("Error hashing password: %v", err)
	}
	err = CheckPasswordHash("Test1", pwHash)
	if err == nil {
		t.Errorf("Expected error for wrong password, got nil")
	}
}

//check for determinism of the resulting hash by hashing it twice
func TestPassHashTwice(t *testing.T){
	pwHash1, err := HashPassword("Test3")
	if err != nil {
		t.Fatalf("Error hashing password: %v", err)
	}
	pwHash2, err := HashPassword("Test3")
	if err != nil {
		t.Fatalf("Error hashing password: %v", err)
	}
	err = CheckPasswordHash("Test3", pwHash1)
	if err != nil {
		t.Errorf("Error comparing password and hashed password: %v", err)
	}
	err = CheckPasswordHash("Test3", pwHash2)
	if err != nil {
		t.Errorf("Error comparing password and hashed password: %v", err)
	}
}