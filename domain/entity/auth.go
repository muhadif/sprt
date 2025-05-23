// Package entity contains all the domain entities for the application.
package entity

import "time"

// SpotifyAuth represents the authentication data for Spotify API.
type SpotifyAuth struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresAt    int64  `json:"expires_at"`
}

// IsExpired checks if the access token is expired.
func (a *SpotifyAuth) IsExpired() bool {
	if a.ExpiresAt == 0 {
		return true
	}
	return a.ExpiresAt <= getCurrentUnixTime()
}

// getCurrentUnixTime returns the current Unix timestamp.
func getCurrentUnixTime() int64 {
	return time.Now().Unix()
}
