package service

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey []byte

// InitJWTKey initializes the JWT key from environment variable or generates a secure one
func InitJWTKey() error {
	// Try to get JWT key from environment variable
	secretKey := os.Getenv("JWT_SECRET_KEY")

	if secretKey == "" {
		// Generate a secure random key if not provided
		key := make([]byte, 64) // 512 bits
		if _, err := rand.Read(key); err != nil {
			return fmt.Errorf("failed to generate JWT key: %v", err)
		}
		secretKey = base64.URLEncoding.EncodeToString(key)
		fmt.Printf("Generated JWT key: %s\n", secretKey)
		fmt.Println("⚠️  WARNING: Set JWT_SECRET_KEY environment variable for production!")
	}

	jwtKey = []byte(secretKey)
	return nil
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func GenerateJWT(username string) (string, error) {
	if jwtKey == nil {
		return "", fmt.Errorf("JWT key not initialized")
	}

	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "susi-auth-service",
			Subject:   username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func ValidateJWT(tokenString string) (*Claims, error) {
	if jwtKey == nil {
		return nil, fmt.Errorf("JWT key not initialized")
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %v", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid JWT token")
	}

	return claims, nil
}

// GetJWTKey returns the current JWT key (for testing purposes only)
func GetJWTKey() []byte {
	return jwtKey
}

func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func GenerateResetToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
