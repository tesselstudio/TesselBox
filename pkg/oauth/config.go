package oauth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Config holds OAuth configuration
type Config struct {
	GitHubClientID     string
	GitHubClientSecret string
	GitHubRedirectURI  string
	JWTSecret         string
	JWTExpiryHours    int
	ServerHost        string
	ServerPort        string
}

// LoadConfig loads OAuth configuration from environment variables
func LoadConfig() (*Config, error) {
	config := &Config{
		GitHubClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		GitHubClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		GitHubRedirectURI:  os.Getenv("GITHUB_REDIRECT_URI"),
		JWTSecret:         os.Getenv("JWT_SECRET"),
		JWTExpiryHours:    24, // Default
		ServerHost:        os.Getenv("SERVER_HOST"),
		ServerPort:        os.Getenv("SERVER_PORT"),
	}

	// Set defaults
	if config.ServerHost == "" {
		config.ServerHost = "localhost"
	}
	if config.ServerPort == "" {
		config.ServerPort = "8080"
	}
	if config.GitHubRedirectURI == "" {
		config.GitHubRedirectURI = fmt.Sprintf("http://%s:%s/auth/github/callback", config.ServerHost, config.ServerPort)
	}

	// Validate required fields
	if config.GitHubClientID == "" {
		return nil, fmt.Errorf("GITHUB_CLIENT_ID environment variable is required")
	}
	if config.GitHubClientSecret == "" {
		return nil, fmt.Errorf("GITHUB_CLIENT_SECRET environment variable is required")
	}
	if config.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET environment variable is required")
	}

	return config, nil
}

// GenerateState generates a random state string for OAuth CSRF protection
func GenerateState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GenerateJWT creates a JWT token for the authenticated user
func GenerateJWT(userID, username string, config *Config) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      time.Now().Add(time.Hour * time.Duration(config.JWTExpiryHours)).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.JWTSecret))
}

// ValidateJWT validates a JWT token and returns the claims
func ValidateJWT(tokenString string, config *Config) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}
