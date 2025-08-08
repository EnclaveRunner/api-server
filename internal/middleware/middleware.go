package middleware

import (
	"log"
	"net/http"
	"time"

	"crypto/rand"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/sirupsen/logrus"
)

// TODO: Store JWT secret in a secure way, e.g., environment variable or config file
var signatureSecret []byte

func StartSecretLifecycle(logger *logrus.Logger, ttl time.Duration) {
	go func() {
		cycle := 0
		for {
			time.Sleep(ttl)
			rotateSecret()
			logger.WithFields(map[string]interface{}{
				"action ": "Secret rotation",
				"cycle ":  cycle,
			}).Info("Secret rotated successfully")
			cycle++
		}
	}()

}

func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			return
		}

		// TODO: Validate the token
		if token != "Bearer secret" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Next()
	}
}

func rotateSecret() {
	secretValue, err := generateSecret()
	if err != nil {
		log.Fatalf("Failed to generate secret: %v", err)
	}
	signatureSecret = secretValue
}

// generateSecret generates a secure random 32-byte secret.
func generateSecret() ([]byte, error) {
	secret := make([]byte, 32)
	_, err := rand.Read(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random secret: %w", err)
	}
	return secret, nil
}

func IssueJWT(c *gin.Context, username string, password string) (string, error) {
	// TODO: Only mocking, validate username and password against meta-database
	if username != "john" || password != "doe" {
		return "", fmt.Errorf("invalid username or password")
	}

	// Create a new token object
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"firstname": "john",
		"lastname":  "doe",
		"iat":       time.Now().Add(time.Minute * 30).Unix(), //
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(signatureSecret)

	return tokenString, err
}
