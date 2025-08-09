package middleware

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/sirupsen/logrus"
)

const (
	secretLength           = 32
	tokenExpirationMinutes = 30
)

// ErrInvalidCredentials is returned when authentication fails
var ErrInvalidCredentials = errors.New("invalid username or password")

// SecretManager manages JWT signing secrets with rotation
type SecretManager struct {
	mu     sync.RWMutex
	secret []byte
}

// NewSecretManager creates a new secret manager
func NewSecretManager() (*SecretManager, error) {
	manager := &SecretManager{
		mu:     sync.RWMutex{},
		secret: nil,
	}

	err := manager.rotateSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize secret manager: %w", err)
	}

	return manager, nil
}

// GetSecret safely returns the current secret
func (sm *SecretManager) GetSecret() []byte {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	secret := make([]byte, len(sm.secret))
	copy(secret, sm.secret)

	return secret
}

// rotateSecret generates and sets a new secret
func (sm *SecretManager) rotateSecret() error {
	secretValue, err := generateSecret()
	if err != nil {
		return fmt.Errorf("failed to generate secret: %w", err)
	}

	sm.mu.Lock()
	sm.secret = secretValue
	sm.mu.Unlock()

	return nil
}

// globalSecretManager is the package-level secret manager
var globalSecretManager *SecretManager

func init() {
	var err error
	globalSecretManager, err = NewSecretManager()
	if err != nil {
		log.Fatalf("Failed to initialize secret manager: %v", err)
	}
}

// StartSecretLifecycle starts the secret rotation lifecycle
func StartSecretLifecycle(logger *logrus.Logger, ttl time.Duration) {
	go func() {
		cycle := 0
		for {
			time.Sleep(ttl)
			err := globalSecretManager.rotateSecret()
			if err != nil {
				logger.WithError(err).Error("Failed to rotate secret")

				continue
			}

			logger.WithFields(map[string]interface{}{
				"action": "Secret rotation",
				"cycle":  cycle,
			}).Info("Secret rotated successfully")
			cycle++
		}
	}()
}

// JWTMiddleware returns a middleware that validates JWT tokens
func JWTMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.GetHeader("Authorization")
		if token == "" {
			ctx.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{"error": "authorization header required"},
			)

			return
		}

		// Simple token validation - in production, validate JWT properly
		if token != "Bearer secret" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})

			return
		}

		ctx.Next()
	}
}

// generateSecret generates a secure random secret
func generateSecret() ([]byte, error) {
	secret := make([]byte, secretLength)
	_, err := rand.Read(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random secret: %w", err)
	}

	return secret, nil
}

// IssueJWT issues a JWT token for the given credentials
func IssueJWT(ctx *gin.Context, username, password string) (string, error) {
	// Mock authentication - in production, validate against proper user store
	if username != "john" || password != "doe" {
		return "", ErrInvalidCredentials
	}

	// Create a new token object
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"firstname": "john",
		"lastname":  "doe",
		"iat":       time.Now().Add(time.Minute * tokenExpirationMinutes).Unix(),
	})

	// Get the current secret
	secret := globalSecretManager.GetSecret()

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT token: %w", err)
	}

	return tokenString, nil
}
