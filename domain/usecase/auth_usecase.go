// Package usecase contains the application business rules.
package usecase

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/muhadif/sprt/domain/entity"
	"github.com/muhadif/sprt/domain/repository"
)

// AuthUseCase defines the interface for authentication use cases.
type AuthUseCase interface {
	// InitAuth initializes the authentication process with client credentials.
	InitAuth(ctx context.Context, clientID, clientSecret string) (string, error)

	// HandleCallback processes the callback from Spotify with the authorization code.
	HandleCallback(ctx context.Context, code string) error

	// ExchangeCodeForToken exchanges the authorization code for an access token.
	ExchangeCodeForToken(ctx context.Context) error

	// GetCurrentlyPlaying retrieves the user's currently playing track.
	GetCurrentlyPlaying(ctx context.Context) (string, error)

	// GetToken retrieves the stored authentication data.
	GetToken(ctx context.Context) (*entity.SpotifyAuth, error)

	// RefreshToken refreshes the access token using the refresh token.
	RefreshToken(ctx context.Context) (*entity.SpotifyAuth, error)
}

// authUseCase implements the AuthUseCase interface.
type authUseCase struct {
	authRepo repository.AuthRepository
}

// NewAuthUseCase creates a new instance of AuthUseCase.
func NewAuthUseCase(authRepo repository.AuthRepository) AuthUseCase {
	return &authUseCase{
		authRepo: authRepo,
	}
}

// InitAuth initializes the authentication process with client credentials.
func (a *authUseCase) InitAuth(ctx context.Context, clientID, clientSecret string) (string, error) {
	// Store client credentials
	if err := a.authRepo.StoreClientCredentials(ctx, clientID, clientSecret); err != nil {
		return "", fmt.Errorf("failed to store client credentials: %w", err)
	}

	// Generate the authorization URL
	authURL := generateAuthURL(clientID)
	return authURL, nil
}

// HandleCallback processes the callback from Spotify with the authorization code.
func (a *authUseCase) HandleCallback(ctx context.Context, code string) error {
	// Store the authorization code
	if err := a.authRepo.StoreAuthCode(ctx, code); err != nil {
		return fmt.Errorf("failed to store authorization code: %w", err)
	}
	return nil
}

// ExchangeCodeForToken exchanges the authorization code for an access token.
func (a *authUseCase) ExchangeCodeForToken(ctx context.Context) error {
	// Get the authorization code
	code, err := a.authRepo.GetAuthCode(ctx)
	if err != nil {
		return fmt.Errorf("failed to get authorization code: %w", err)
	}

	// Get client credentials
	auth, err := a.authRepo.GetToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get client credentials: %w", err)
	}

	// Prepare the request to exchange the code for a token
	tokenURL := "https://accounts.spotify.com/api/token"
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", "http://127.0.0.1:8080/callback")

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create token request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Set basic auth with client ID and secret
	authHeader := base64.StdEncoding.EncodeToString([]byte(auth.ClientID + ":" + auth.ClientSecret))
	req.Header.Set("Authorization", "Basic "+authHeader)

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to exchange code for token: %w", err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read token response: %w", err)
	}

	// Check for error response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse the response
	var tokenResponse struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		Scope        string `json:"scope"`
	}
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return fmt.Errorf("failed to parse token response: %w", err)
	}

	// Update the auth object
	auth = &entity.SpotifyAuth{
		ClientID:     auth.ClientID,
		ClientSecret: auth.ClientSecret,
		AccessToken:  tokenResponse.AccessToken,
		RefreshToken: tokenResponse.RefreshToken,
		ExpiresIn:    tokenResponse.ExpiresIn,
		TokenType:    tokenResponse.TokenType,
		Scope:        tokenResponse.Scope,
		ExpiresAt:    time.Now().Unix() + int64(tokenResponse.ExpiresIn),
	}

	// Store the token
	if err := a.authRepo.StoreToken(ctx, auth); err != nil {
		return fmt.Errorf("failed to store token: %w", err)
	}

	return nil
}

// GetCurrentlyPlaying retrieves the user's currently playing track.
func (a *authUseCase) GetCurrentlyPlaying(ctx context.Context) (string, error) {
	// Get the token
	auth, err := a.authRepo.GetToken(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get token: %w", err)
	}

	// Check if the token is expired and attempt to refresh it
	if auth.IsExpired() {
		// Try to refresh the token
		auth, err = a.RefreshToken(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to refresh token: %w", err)
		}
	}

	// Make a request to Spotify's API
	apiURL := "https://api.spotify.com/v1/me/player/currently-playing"
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create API request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("%s %s", auth.TokenType, auth.AccessToken))

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get currently playing track: %w", err)
	}
	defer resp.Body.Close()

	// Check for error response
	if resp.StatusCode == http.StatusNoContent {
		return "No track currently playing", nil
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Read the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read API response: %w", err)
	}

	// Parse the response
	var trackResponse struct {
		Item struct {
			Name  string `json:"name"`
			Album struct {
				Name string `json:"name"`
			} `json:"album"`
			Artists []struct {
				Name string `json:"name"`
			} `json:"artists"`
		} `json:"item"`
	}
	if err := json.Unmarshal(body, &trackResponse); err != nil {
		return "", fmt.Errorf("failed to parse API response: %w", err)
	}

	// Format the track information
	artistNames := make([]string, len(trackResponse.Item.Artists))
	for i, artist := range trackResponse.Item.Artists {
		artistNames[i] = artist.Name
	}

	return fmt.Sprintf("Currently playing: %s by %s from the album %s",
		trackResponse.Item.Name,
		strings.Join(artistNames, ", "),
		trackResponse.Item.Album.Name), nil
}

// GetToken retrieves the stored authentication data.
func (a *authUseCase) GetToken(ctx context.Context) (*entity.SpotifyAuth, error) {
	return a.authRepo.GetToken(ctx)
}

// RefreshToken refreshes the access token using the refresh token.
func (a *authUseCase) RefreshToken(ctx context.Context) (*entity.SpotifyAuth, error) {
	// Get the current auth data
	auth, err := a.authRepo.GetToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	// Check if we have a refresh token
	if auth.RefreshToken == "" {
		return nil, fmt.Errorf("no refresh token available")
	}

	// Prepare the request to refresh the token
	tokenURL := "https://accounts.spotify.com/api/token"
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", auth.RefreshToken)

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token refresh request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Set basic auth with client ID and secret
	authHeader := base64.StdEncoding.EncodeToString([]byte(auth.ClientID + ":" + auth.ClientSecret))
	req.Header.Set("Authorization", "Basic "+authHeader)

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	// Check for error response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse the response
	var tokenResponse struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		Scope        string `json:"scope"`
	}
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	// Update the auth object
	newAuth := &entity.SpotifyAuth{
		ClientID:     auth.ClientID,
		ClientSecret: auth.ClientSecret,
		AccessToken:  tokenResponse.AccessToken,
		TokenType:    tokenResponse.TokenType,
		ExpiresIn:    tokenResponse.ExpiresIn,
		Scope:        tokenResponse.Scope,
		ExpiresAt:    time.Now().Unix() + int64(tokenResponse.ExpiresIn),
	}

	// Keep the existing refresh token if a new one wasn't provided
	if tokenResponse.RefreshToken != "" {
		newAuth.RefreshToken = tokenResponse.RefreshToken
	} else {
		newAuth.RefreshToken = auth.RefreshToken
	}

	// Store the updated token
	if err := a.authRepo.StoreToken(ctx, newAuth); err != nil {
		return nil, fmt.Errorf("failed to store refreshed token: %w", err)
	}

	return newAuth, nil
}

// generateAuthURL generates the authorization URL for Spotify.
func generateAuthURL(clientID string) string {
	baseURL := "https://accounts.spotify.com/authorize"
	redirectURI := "http://127.0.0.1:8080/callback"
	scope := "user-read-currently-playing"

	params := url.Values{}
	params.Add("client_id", clientID)
	params.Add("response_type", "code")
	params.Add("redirect_uri", redirectURI)
	params.Add("scope", scope)

	return fmt.Sprintf("%s?%s", baseURL, params.Encode())
}
