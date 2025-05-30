package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/muhadif/sprt/domain/usecase"
	"github.com/spf13/cobra"
)

var lyricCmd = &cobra.Command{
	Use:   "lyric",
	Short: "Lyric commands",
	Long:  `Commands for displaying lyrics for the currently playing track.`,
}

var pipeLyricCmd = &cobra.Command{
	Use:   "pipe-lyric",
	Short: "Display synchronized lyrics for the currently playing track",
	Long:  `Display synchronized lyrics for the currently playing track from lrclib.net.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return displaySyncedLyrics()
	},
}

// init function is no longer needed as commands are initialized in root.go
// through the InitializeCommands function

// displaySyncedLyrics displays synchronized lyrics for the currently playing track.
func displaySyncedLyrics() error {
	// Create the player use case
	playerUseCase := usecase.NewPlayerUseCase(authUseCase)

	// Create the lyric use case
	lyricUseCase := usecase.NewLyricUseCase()

	// Get the currently playing track
	track, err := playerUseCase.GetCurrentlyPlayingDetails(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get currently playing track: %w", err)
	}

	// Get the lyrics
	lyrics, err := lyricUseCase.GetLyrics(context.Background(), track.ArtistNames[0], track.Title, track.Album)
	if err != nil {
		fmt.Printf("No lyric found for %s by %s\n", track.Title, track.Artist)
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

	// Display the lyrics synchronized with the music
	lyricUseCase.DisplaySyncedLyrics(ctx, lyrics, track.ProgressMs, playerUseCase)
	return nil
}
