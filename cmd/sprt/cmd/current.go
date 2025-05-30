package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/muhadif/sprt/domain/usecase"
	"github.com/muhadif/sprt/interfaces/tui"
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

	trackInfo, err := authUseCase.GetCurrentlyPlaying(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get currently playing track: %w", err)
	}

	// Check if no track is playing
	if trackInfo == "No track currently playing" {
		// Show waiting UI instead of just printing the message
		return tui.RunWaitingTrackUI(authUseCase)
	}

	// Parse the track information from the string
	// Format is: "Currently playing: {title} by {artist} from the album {album}"
	title, artist, album := parseTrackInfo(trackInfo)

	// Use the TUI to display the track
	return tui.RunCurrentTrackUI(artist, title, album, "Unknown", "Unknown", true)
}

// parseTrackInfo parses the track information from the formatted string
func parseTrackInfo(trackInfo string) (title, artist, album string) {
	// Remove the "Currently playing: " prefix
	trackInfo = strings.TrimPrefix(trackInfo, "Currently playing: ")

	// Split by " by " to get title and the rest
	parts := strings.Split(trackInfo, " by ")
	if len(parts) < 2 {
		return trackInfo, "", ""
	}

	title = parts[0]

	// Split the rest by " from the album " to get artist and album
	parts = strings.Split(parts[1], " from the album ")
	if len(parts) < 2 {
		return title, parts[0], ""
	}

	return title, parts[0], parts[1]
}
