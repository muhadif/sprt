package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/muhadif/sprt/domain/usecase"
	"github.com/muhadif/sprt/interfaces/tui"
	"github.com/spf13/cobra"
)

var lyricCmd = &cobra.Command{
	Use:   "lyric",
	Short: "Lyric commands",
	Long:  `Commands for displaying lyrics for the currently playing track.`,
}

var pipeLyricCmd = &cobra.Command{
	Use:   "pipe",
	Short: "Display synchronized lyrics for the currently playing track",
	Long:  `Display synchronized lyrics for the currently playing track from lrclib.net.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return displaySyncedLyrics()
	},
}

var showLyricCmd = &cobra.Command{
	Use:   "show",
	Short: "Display lyrics for the currently playing track with a nice UI",
	Long:  `Display lyrics for the currently playing track from lrclib.net with a nice UI.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return displayLyricsWithUI()
	},
}

// init function is no longer needed as commands are initialized in root.go
// through the InitializeCommands function

// displayLyricsWithUI displays lyrics for the currently playing track with a nice UI.
func displayLyricsWithUI() error {
	// Create the player use case
	playerUseCase := usecase.NewPlayerUseCase(authUseCase)

	// Get the currently playing track
	track, err := playerUseCase.GetCurrentlyPlayingDetails(context.Background())
	if err != nil {
		// Check if the error is "no track currently playing"
		if err.Error() == "no track currently playing" {
			// Show waiting UI instead of returning an error
			return tui.RunWaitingTrackUI(authUseCase)
		}
		return fmt.Errorf("failed to get currently playing track: %w", err)
	}

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle Ctrl+C to gracefully exit
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cancel()
		fmt.Println("\nStopping lyrics display...")
		os.Exit(0)
	}()

	// Run the lyric UI
	return tui.RunLyricUI(ctx, track.ProgressMs, playerUseCase)
}

// displaySyncedLyrics displays synchronized lyrics for the currently playing track.
func displaySyncedLyrics() error {
	// Create the player use case
	playerUseCase := usecase.NewPlayerUseCase(authUseCase)

	// Get the currently playing track
	track, err := playerUseCase.GetCurrentlyPlayingDetails(context.Background())
	if err != nil {
		// Check if the error is "no track currently playing"
		if err.Error() == "no track currently playing" {
			// Show waiting UI instead of returning an error
			return tui.RunWaitingTrackUI(authUseCase)
		}
		return fmt.Errorf("failed to get currently playing track: %w", err)
	}

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle Ctrl+C to gracefully exit
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cancel()
		fmt.Println("\nStopping lyrics display...")
		os.Exit(0)
	}()

	// Run the pipe lyric UI
	return tui.RunPipeLyricUI(ctx, track.ProgressMs, playerUseCase)
}
