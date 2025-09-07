package auth

import (
	"golang.org/x/crypto/bcrypt"
	"log"
	"time"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"fmt"
)

func HashPassword(password string) (string, error){

	hashed_pw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error generating password: %v", err)
		return "", err
	}

	str_hashed_pw := string(hashed_pw)
	return str_hashed_pw, nil
}

func CheckPasswordHash(password, hash string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		log.Printf("Error comparing hash and password")
		return err
	}

	return nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error){
	/*
	//get secretkey
	secretKey := os.Getenv("SECRET_KEY")
	*/

	//create custom claims for token creation
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:		"chirpy",
		IssuedAt:	jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt:	jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:	userID.String(),
		})
	
	//sign the token
	tokenString, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		log.Printf("Error signing token: %v", err)
		return "", err
	}
	return tokenString, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error){
	token, err := jwt.ParseWithClaims(
		tokenString,
		&jwt.RegisteredClaims{},
		func(t *jwt.Token) (any, error){
			//ensure HMAC (HS256 family)
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				log.Printf("Error ensuring HMAC signing method")
				return nil, fmt.Errorf("unexpected signing method: %T", t.Method)
			}
			return []byte(tokenSecret), nil
		})
	//check if theres any error or token invalid
	if err != nil || !token.Valid {
		log.Printf("Error parsing token claim: %v", err)
		return uuid.Nil, fmt.Errorf("invalid token")
	}
	//check the claims
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid claims type")
	}

	id, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}