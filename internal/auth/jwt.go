package auth

import (
	"time"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims defines the payload
type JWTClaims struct {
	UserID       uint   `json:"user_id"`
	Email        string `json:"email"`
	AuthProvider string `json:"auth_provider"`
	jwt.RegisteredClaims
}

// GenerateJWT creates a signed token
func GenerateJWT(userID uint, email, authProvider string) (string, error) {
	secret := os.Getenv("JWT_SECRET") 
	if secret == "" {
		secret = "dev-secret" // fallback for development
	}

	claims := JWTClaims{
		UserID:       userID,
		Email:        email,
		AuthProvider: authProvider,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)), // 1 day expiry
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}