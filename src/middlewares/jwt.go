package middlewares

import (
	"net/http"
	"strings"

	"sama/sama-backend-2025/src/utils" // Adjust import path

	"github.com/gin-gonic/gin"
)

const (
	// UserContextKey is the key to store user claims in the Gin context.
	UserContextKey = "userClaims"
)

// AuthMiddleware validates JWT tokens and injects user claims into the Gin context.
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Authorization header required"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid Authorization header format"})
			return
		}

		tokenString := parts[1]
		claims, err := utils.ValidateToken(tokenString, jwtSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid or expired token: " + err.Error()})
			return
		}

		// Store claims in Gin context
		c.Set(UserContextKey, claims)
		c.Next() // Proceed to the next handler
	}
}

// GetUserClaimsFromContext retrieves user claims from the Gin context.
func GetUserClaimsFromContext(c *gin.Context) (*utils.Claims, bool) {
	claims, ok := c.Get(UserContextKey)
	if !ok {
		return nil, false
	}
	userClaims, ok := claims.(*utils.Claims)
	return userClaims, ok
}
