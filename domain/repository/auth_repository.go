// Package repository defines the interfaces for data access.
package repository

import (
	"context"

	"github.com/muhadif/sprt/domain/entity"
)

// AuthRepository defines the interface for authentication data storage.
type AuthRepository interface {
	// StoreClientCredentials saves the client ID and secret.
	StoreClientCredentials(ctx context.Context, clientID, clientSecret string) error

	// StoreAuthCode saves the authorization code received from Spotify.
	StoreAuthCode(ctx context.Context, code string) error

	// GetAuthCode retrieves the stored authorization code.
	GetAuthCode(ctx context.Context) (string, error)

	// StoreToken saves the access and refresh tokens.
	StoreToken(ctx context.Context, auth *entity.SpotifyAuth) error

	// GetToken retrieves the stored authentication data.
	GetToken(ctx context.Context) (*entity.SpotifyAuth, error)
}
