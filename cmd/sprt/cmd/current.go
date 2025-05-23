package cmd

import (
	"context"
	"fmt"

	"github.com/muhadif/sprt/domain/usecase"
	"github.com/spf13/cobra"
)

var currentCmd = &cobra.Command{
	Use:   "current",
	Short: "Get currently playing track",
	Long:  `Get information about your currently playing track on Spotify.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return getCurrentlyPlaying(authUseCase)
	},
}

// init function is no longer needed as commands are initialized in root.go
// through the InitializeCommands function

// getCurrentlyPlaying retrieves the user's currently playing track.
func getCurrentlyPlaying(authUseCase usecase.AuthUseCase) error {
	fmt.Println("Retrieving currently playing track...")

	track, err := authUseCase.GetCurrentlyPlaying(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get currently playing track: %w", err)
	}

	fmt.Println(track)
	return nil
}
