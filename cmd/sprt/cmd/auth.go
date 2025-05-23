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

	// Prompt for client ID
	clientID, err := promptInput("Enter your Spotify Client ID: ")
	if err != nil {
		return fmt.Errorf("failed to read client ID: %w", err)
	}

	// Prompt for client secret
	clientSecret, err := promptInput("Enter your Spotify Client Secret: ")
	if err != nil {
		return fmt.Errorf("failed to read client secret: %w", err)
	}

	// Initialize authentication with the provided credentials
	authURL, err := authUseCase.InitAuth(context.Background(), clientID, clientSecret)
	if err != nil {
		return fmt.Errorf("failed to initialize authentication: %w", err)
	}

	// Display the authorization URL
	fmt.Println("\nPlease open the following URL in your browser to authorize the application:")
	fmt.Println(authURL)

	// Start the callback server
	callbackServer := httpinterface.NewCallbackServer(authUseCase)
	go func() {
		err := callbackServer.Start(8080)
		if err != nil && err != http.ErrServerClosed {
			fmt.Printf("Error starting callback server: %v\n", err)
		}
	}()

	fmt.Println("\nWaiting for authorization...")
	fmt.Println("After authorizing, you will be redirected to a local callback URL.")
	fmt.Println("Press Enter after you have completed the authorization process...")

	// Wait for user to press Enter
	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')

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
