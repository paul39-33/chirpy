package auth

import (
	"testing"
	"time"
	"github.com/google/uuid"
	"net/http"
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

func TestMakeAndValidateJWT_Succeeds(t *testing.T){
	userID := uuid.New()
	secret := "test-secret"

	tok, err := MakeJWT(userID, secret, time.Minute)
	if err != nil {
		t.Fatalf("MakeJWT err: %v", err)
	}

	gotID, err := ValidateJWT(tok, secret)
	if err != nil {
		t.Fatalf("ValidateJWT err: %v", err)
	}
	if gotID != userID{
		t.Fatalf("want %v, got %v", userID, gotID)
	}
}

func TestValidateJWT_Expired(t *testing.T){
	userID := uuid.New()
	secret := "test-secret2"

	tok, err := MakeJWT(userID, secret, -time.Minute)
	if err != nil {
		t.Fatalf("MakeJWT err: %v", err)
	}

	if _, err := ValidateJWT(tok, secret); err == nil {
		t.Fatalf("expected error for expired token")
	}
}

func TestValidateJWT_WrongSecret(t *testing.T){
	userID := uuid.New()

	tok, err := MakeJWT(userID, "secret1", time.Minute)
	if err != nil {
		t.Fatalf("MakeJWT err: %v", err)
	}

	if _, err := ValidateJWT(tok, "secret2"); err == nil {
		t.Fatalf("expected error for wrong secret")
	}
}

func TestGetBearerToken_success (t *testing.T){
	headers := make(http.Header)

	headers.Add("Content-Type", "application/json")
	headers.Add("X-Custom-Header", "value1")
	headers.Add("Authorization", "Bearer SOME_SECRET_TOKEN")

	token, err := GetBearerToken(headers)
	if err != nil {
		t.Fatalf("GetBearerToken error: %v", err)
	}
	if token != "SOME_SECRET_TOKEN" {
		t.Fatalf("error expected %v, got %v", "SOME_SECRET_TOKEN", token)
	}
}

func TestGetBearerToken_empty (t *testing.T){
	headers := make(http.Header)

	headers.Add("Content-Type", "application/json")
	headers.Add("X-Custom-Header", "value1")

	_, err := GetBearerToken(headers)
	if err == nil {
		t.Fatalf("expected error for empty bearer token")
	}
}