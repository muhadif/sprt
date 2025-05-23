// Package main is the entry point for the Spotify CLI application.
package main

import (
	"github.com/muhadif/sprt/cmd/sprt/cmd"
	"github.com/muhadif/sprt/domain/usecase"
	"github.com/muhadif/sprt/infrastructure/persistence/jsonfile"
)

// Version information set by GoReleaser at build time
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Initialize repositories
	authRepo := jsonfile.NewAuthRepository()

	// Initialize use cases
	authUseCase := usecase.NewAuthUseCase(authRepo)
	playerUseCase := usecase.NewPlayerUseCase(authUseCase)
	lyricUseCase := usecase.NewLyricUseCase()

	// Initialize commands with version information
	cmd.InitializeCommands(authUseCase, playerUseCase, lyricUseCase, version, commit, date)

	// Execute the root command
	cmd.Execute()
}
