package auth

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 0)
	if err != nil {
		return "", fmt.Errorf("error hashing password")
	}
	return string(hash), nil
}

func CheckPasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return fmt.Errorf("incorrect password")
	}
	return nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	registeredClaims := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userID.String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, registeredClaims)
	tokenString, err := token.SignedString([]byte(tokenSecret))

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return []byte{}, fmt.Errorf("incorrect signing method")
		}
		return []byte(tokenSecret), nil
	}

	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, keyFunc)

	if err != nil {
		return uuid.UUID{}, err
	}

	uuid_string, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.UUID{}, err
	}

	extracted_uuid, err := uuid.Parse(uuid_string)
	if err != nil {
		return uuid.UUID{}, err
	}

	return extracted_uuid, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	auth_header := headers.Get("Authorization")

	if auth_header == "" {
		return "", fmt.Errorf("no 'authorization' header present")
	}

	auth_header = strings.Trim(auth_header, "Bearer")
	auth_header = strings.TrimSpace(auth_header)

	return auth_header, nil
}
