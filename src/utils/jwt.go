package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5" // Use jwt/v5
)

// Claims defines the JWT claims structure.
// You can add more custom claims as needed (e.g., user role, school ID).
type Claims struct {
	UserID   uint   `json:"user_id"`
	SchoolID uint   `json:"school_id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// Claims defines the JWT claims structure.
// You can add more custom claims as needed (e.g., user role, school ID).
type RefreshClaims struct {
	// Registered Claims (Standard JWT claims)
	jwt.RegisteredClaims
	// 'sub' (Subject) is generally sufficient for identifying the user.
	// However, if you have a specific internal user ID, you can include that.
	UserID uint `json:"user_id"`

	// Crucial for identifying the specific refresh token instance
	// and enabling server-side revocation and rotation.
	// Jti string `json:"jti"` // JWT ID - unique identifier for the token
}

// GenerateToken generates a new JWT token for a given user.
func GenerateToken(userID uint, schoolID uint, email, role, jwtSecret string, expirationMinutes int) (string, error) {
	expirationTime := time.Now().Add(time.Duration(expirationMinutes) * time.Minute)
	claims := &Claims{
		UserID:   userID,
		Email:    email,
		Role:     role,
		SchoolID: schoolID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return tokenString, nil
}

// ValidateToken validates a JWT token and returns its claims.
func ValidateToken(tokenString, jwtSecret string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// GenerateToken generates a new JWT token for a given user.
func GenerateRefreshToken(userID uint, jwtSecret string, expirationMinutes int) (string, error) {
	expirationTime := time.Now().Add(time.Duration(expirationMinutes) * time.Minute)
	claims := &RefreshClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return tokenString, nil
}

// ValidateToken validates a JWT token and returns its claims.
func ValidateRefreshToken(tokenString, jwtSecret string) (*RefreshClaims, error) {
	claims := &RefreshClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	// TODO: make refresh a one-time token

	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
