package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// PlayerUseCase defines the interface for player-related use cases.
type PlayerUseCase interface {
	// GetCurrentlyPlayingDetails retrieves detailed information about the user's currently playing track.
	GetCurrentlyPlayingDetails(ctx context.Context) (*CurrentlyPlaying, error)
}

// CurrentlyPlaying represents detailed information about the currently playing track.
type CurrentlyPlaying struct {
	IsPlaying   bool   `json:"is_playing"`
	ProgressMs  int    `json:"progress_ms"`
	Title       string `json:"title"`
	Artist      string `json:"artist"`
	Album       string `json:"album"`
	ArtistNames []string
	DurationMs  int `json:"duration_ms"`
}

// playerUseCase implements the PlayerUseCase interface.
type playerUseCase struct {
	authUseCase AuthUseCase
}

// NewPlayerUseCase creates a new instance of PlayerUseCase.
func NewPlayerUseCase(authUseCase AuthUseCase) PlayerUseCase {
	return &playerUseCase{
		authUseCase: authUseCase,
	}
}

// GetCurrentlyPlayingDetails retrieves detailed information about the user's currently playing track.
func (p *playerUseCase) GetCurrentlyPlayingDetails(ctx context.Context) (*CurrentlyPlaying, error) {
	// Get the token
	auth, err := p.authUseCase.GetToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	// Check if the token is expired and attempt to refresh it
	if auth.IsExpired() {
		// Try to refresh the token
		auth, err = p.authUseCase.RefreshToken(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to refresh token: %w", err)
		}
	}

	// Make a request to Spotify's API
	apiURL := "https://api.spotify.com/v1/me/player/currently-playing"
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create API request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("%s %s", auth.TokenType, auth.AccessToken))

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get currently playing track: %w", err)
	}
	defer resp.Body.Close()

	// Check for error response
	if resp.StatusCode == http.StatusNoContent {
		return nil, fmt.Errorf("no track currently playing")
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Read the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read API response: %w", err)
	}

	// Parse the response
	var trackResponse struct {
		IsPlaying  bool `json:"is_playing"`
		ProgressMs int  `json:"progress_ms"`
		Item       struct {
			Name       string `json:"name"`
			DurationMs int    `json:"duration_ms"`
			Album      struct {
				Name string `json:"name"`
			} `json:"album"`
			Artists []struct {
				Name string `json:"name"`
			} `json:"artists"`
		} `json:"item"`
	}
	if err := json.Unmarshal(body, &trackResponse); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	// Extract artist names
	artistNames := make([]string, len(trackResponse.Item.Artists))
	for i, artist := range trackResponse.Item.Artists {
		artistNames[i] = artist.Name
	}

	// Create the result
	result := &CurrentlyPlaying{
		IsPlaying:   trackResponse.IsPlaying,
		ProgressMs:  trackResponse.ProgressMs,
		Title:       trackResponse.Item.Name,
		Artist:      strings.Join(artistNames, ", "),
		Album:       trackResponse.Item.Album.Name,
		ArtistNames: artistNames,
		DurationMs:  trackResponse.Item.DurationMs,
	}

	return result, nil
}
