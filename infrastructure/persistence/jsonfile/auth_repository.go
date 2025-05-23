package jsonfile

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/muhadif/sprt/domain/entity"
	"github.com/muhadif/sprt/domain/repository"
)

// authRepository implements the repository.AuthRepository interface using JSON file storage.
type authRepository struct {
	mu       sync.RWMutex
	filePath string
	authCode string
	auth     *entity.SpotifyAuth
}

// NewAuthRepository creates a new instance of the JSON file-based auth repository.
func NewAuthRepository() repository.AuthRepository {
	// Create the directory if it doesn't exist
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	configDir := filepath.Join(homeDir, ".sprt")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		fmt.Printf("Warning: Failed to create config directory: %v\n", err)
	}

	filePath := filepath.Join(configDir, "auth.json")

	repo := &authRepository{
		filePath: filePath,
		auth:     &entity.SpotifyAuth{},
	}

	// Load existing data if available
	repo.loadFromFile()

	return repo
}

// loadFromFile loads authentication data from the JSON file.
func (r *authRepository) loadFromFile() {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if the file exists
	if _, err := os.Stat(r.filePath); os.IsNotExist(err) {
		return
	}

	// Read the file
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		fmt.Printf("Warning: Failed to read auth file: %v\n", err)
		return
	}

	// Parse the JSON
	var auth entity.SpotifyAuth
	if err := json.Unmarshal(data, &auth); err != nil {
		fmt.Printf("Warning: Failed to parse auth file: %v\n", err)
		return
	}

	r.auth = &auth
}

// saveToFile saves authentication data to the JSON file.
func (r *authRepository) saveToFile() error {
	// Marshal the auth data to JSON
	data, err := json.MarshalIndent(r.auth, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal auth data: %w", err)
	}

	// Write to the file
	if err := os.WriteFile(r.filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write auth file: %w", err)
	}

	return nil
}

// StoreClientCredentials saves the client ID and secret.
func (r *authRepository) StoreClientCredentials(ctx context.Context, clientID, clientSecret string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.auth.ClientID = clientID
	r.auth.ClientSecret = clientSecret

	return r.saveToFile()
}

// StoreAuthCode saves the authorization code received from Spotify.
func (r *authRepository) StoreAuthCode(ctx context.Context, code string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.authCode = code
	return nil
}

// GetAuthCode retrieves the stored authorization code.
func (r *authRepository) GetAuthCode(ctx context.Context) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.authCode == "" {
		return "", fmt.Errorf("authorization code not found")
	}

	return r.authCode, nil
}

// StoreToken saves the access and refresh tokens.
func (r *authRepository) StoreToken(ctx context.Context, auth *entity.SpotifyAuth) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.auth = auth
	return r.saveToFile()
}

// GetToken retrieves the stored authentication data.
func (r *authRepository) GetToken(ctx context.Context) (*entity.SpotifyAuth, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.auth == nil {
		return nil, fmt.Errorf("authentication data not found")
	}

	return r.auth, nil
}
