package oauth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// TokenStorage represents a token storage interface
type TokenStorage interface {
	StoreToken(userID string, token string, user *GitHubUser) error
	GetToken(userID string) (string, error)
	GetUser(userID string) (*GitHubUser, error)
	DeleteToken(userID string) error
	CleanupExpired() error
}

// FileTokenStorage implements file-based token storage
type FileTokenStorage struct {
	filePath string
	mu       sync.RWMutex
}

// TokenData represents stored token information
type TokenData struct {
	Token     string      `json:"token"`
	User      *GitHubUser `json:"user"`
	CreatedAt time.Time   `json:"created_at"`
	ExpiresAt time.Time   `json:"expires_at"`
}

// NewFileTokenStorage creates a new file-based token storage
func NewFileTokenStorage(dataDir string) (*FileTokenStorage, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}
	
	return &FileTokenStorage{
		filePath: filepath.Join(dataDir, "oauth_tokens.json"),
	}, nil
}

// StoreToken stores a token for a user
func (s *FileTokenStorage) StoreToken(userID string, token string, user *GitHubUser) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Load existing tokens
	tokens, err := s.loadTokens()
	if err != nil {
		return fmt.Errorf("failed to load existing tokens: %w", err)
	}
	
	// Add or update token
	tokens[userID] = TokenData{
		Token:     token,
		User:      user,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hours
	}
	
	// Save tokens
	return s.saveTokens(tokens)
}

// GetToken retrieves a token for a user
func (s *FileTokenStorage) GetToken(userID string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	tokens, err := s.loadTokens()
	if err != nil {
		return "", fmt.Errorf("failed to load tokens: %w", err)
	}
	
	tokenData, exists := tokens[userID]
	if !exists {
		return "", fmt.Errorf("token not found for user: %s", userID)
	}
	
	// Check if token is expired
	if time.Now().After(tokenData.ExpiresAt) {
		delete(tokens, userID)
		s.saveTokens(tokens) // Clean up expired token
		return "", fmt.Errorf("token expired for user: %s", userID)
	}
	
	return tokenData.Token, nil
}

// GetUser retrieves user information
func (s *FileTokenStorage) GetUser(userID string) (*GitHubUser, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	tokens, err := s.loadTokens()
	if err != nil {
		return nil, fmt.Errorf("failed to load tokens: %w", err)
	}
	
	tokenData, exists := tokens[userID]
	if !exists {
		return nil, fmt.Errorf("user not found: %s", userID)
	}
	
	// Check if token is expired
	if time.Now().After(tokenData.ExpiresAt) {
		delete(tokens, userID)
		s.saveTokens(tokens) // Clean up expired token
		return nil, fmt.Errorf("user session expired: %s", userID)
	}
	
	return tokenData.User, nil
}

// DeleteToken removes a token for a user
func (s *FileTokenStorage) DeleteToken(userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	tokens, err := s.loadTokens()
	if err != nil {
		return fmt.Errorf("failed to load tokens: %w", err)
	}
	
	delete(tokens, userID)
	return s.saveTokens(tokens)
}

// CleanupExpired removes all expired tokens
func (s *FileTokenStorage) CleanupExpired() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	tokens, err := s.loadTokens()
	if err != nil {
		return fmt.Errorf("failed to load tokens: %w", err)
	}
	
	now := time.Now()
	for userID, tokenData := range tokens {
		if now.After(tokenData.ExpiresAt) {
			delete(tokens, userID)
		}
	}
	
	return s.saveTokens(tokens)
}

// loadTokens loads tokens from file
func (s *FileTokenStorage) loadTokens() (map[string]TokenData, error) {
	tokens := make(map[string]TokenData)
	
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return tokens, nil // File doesn't exist, return empty map
		}
		return nil, err
	}
	
	if err := json.Unmarshal(data, &tokens); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tokens: %w", err)
	}
	
	return tokens, nil
}

// saveTokens saves tokens to file
func (s *FileTokenStorage) saveTokens(tokens map[string]TokenData) error {
	data, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tokens: %w", err)
	}
	
	return os.WriteFile(s.filePath, data, 0600)
}

// MemoryTokenStorage implements in-memory token storage (for testing)
type MemoryTokenStorage struct {
	tokens map[string]TokenData
	mu     sync.RWMutex
}

// NewMemoryTokenStorage creates a new memory-based token storage
func NewMemoryTokenStorage() *MemoryTokenStorage {
	return &MemoryTokenStorage{
		tokens: make(map[string]TokenData),
	}
}

// StoreToken stores a token for a user
func (s *MemoryTokenStorage) StoreToken(userID string, token string, user *GitHubUser) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.tokens[userID] = TokenData{
		Token:     token,
		User:      user,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	
	return nil
}

// GetToken retrieves a token for a user
func (s *MemoryTokenStorage) GetToken(userID string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	tokenData, exists := s.tokens[userID]
	if !exists {
		return "", fmt.Errorf("token not found for user: %s", userID)
	}
	
	if time.Now().After(tokenData.ExpiresAt) {
		delete(s.tokens, userID)
		return "", fmt.Errorf("token expired for user: %s", userID)
	}
	
	return tokenData.Token, nil
}

// GetUser retrieves user information
func (s *MemoryTokenStorage) GetUser(userID string) (*GitHubUser, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	tokenData, exists := s.tokens[userID]
	if !exists {
		return nil, fmt.Errorf("user not found: %s", userID)
	}
	
	if time.Now().After(tokenData.ExpiresAt) {
		delete(s.tokens, userID)
		return nil, fmt.Errorf("user session expired: %s", userID)
	}
	
	return tokenData.User, nil
}

// DeleteToken removes a token for a user
func (s *MemoryTokenStorage) DeleteToken(userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	delete(s.tokens, userID)
	return nil
}

// CleanupExpired removes all expired tokens
func (s *MemoryTokenStorage) CleanupExpired() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	now := time.Now()
	for userID, tokenData := range s.tokens {
		if now.After(tokenData.ExpiresAt) {
			delete(s.tokens, userID)
		}
	}
	
	return nil
}
