package cmd

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/muhadif/sprt/domain/usecase"
	httpinterface "github.com/muhadif/sprt/interfaces/http"
	"github.com/muhadif/sprt/interfaces/tui"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication commands",
	Long:  `Commands for authenticating with Spotify.`,
}

var authInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize authentication with Spotify",
	Long:  `Initialize authentication with Spotify by providing your client ID and secret.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return initAuth(authUseCase)
	},
}

var authTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test authentication by retrieving currently playing track",
	Long:  `Test authentication by retrieving information about your currently playing track.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return testCurrentlyPlaying(authUseCase)
	},
}

// init function is no longer needed as commands are initialized in root.go
// through the InitializeCommands function

// initAuth initializes the authentication process.
func initAuth(authUseCase usecase.AuthUseCase) error {
	fmt.Println("Initializing Spotify authentication...")

	// Use the TUI to get client ID and client secret
	clientID, clientSecret, err := tui.RunAuthUI()
	if err != nil {
		return fmt.Errorf("failed to get authentication credentials: %w", err)
	}

	// Initialize authentication with the provided credentials
	authURL, err := authUseCase.InitAuth(context.Background(), clientID, clientSecret)
	if err != nil {
		return fmt.Errorf("failed to initialize authentication: %w", err)
	}

	// Start the callback server
	callbackServer := httpinterface.NewCallbackServer(authUseCase)
	go func() {
		err := callbackServer.Start(8080)
		if err != nil && err != http.ErrServerClosed {
			fmt.Printf("Error starting callback server: %v\n", err)
		}
	}()

	// Use the TUI to display the authorization URL and wait for completion
	err = tui.RunAuthWaitingUI(authURL, clientID, clientSecret)
	if err != nil {
		return fmt.Errorf("error in authentication UI: %w", err)
	}

	// Stop the callback server
	ctx := context.Background()
	if err := callbackServer.Stop(ctx); err != nil {
		fmt.Printf("Error stopping callback server: %v\n", err)
	}

	fmt.Println("Authentication process completed.")
	return nil
}

// testCurrentlyPlaying tests the authentication by retrieving the currently playing track.
func testCurrentlyPlaying(authUseCase usecase.AuthUseCase) error {
	fmt.Println("Testing authentication by retrieving currently playing track...")

	track, err := authUseCase.GetCurrentlyPlaying(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get currently playing track: %w", err)
	}

	// Check if no track is playing
	if track == "No track currently playing" {
		// Show waiting UI instead of just printing the message
		return tui.RunWaitingTrackUI(authUseCase)
	}

	fmt.Println(track)
	return nil
}

// promptInput prompts the user for input with the given message.
func promptInput(message string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(message)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}
