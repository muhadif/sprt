// Package cli provides command-line interface functionality for the application.
package cli

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/muhadif/sprt/domain/usecase"
	httpinterface "github.com/muhadif/sprt/interfaces/http"
)

// AuthCommand handles the authentication-related CLI commands.
type AuthCommand struct {
	authUseCase usecase.AuthUseCase
}

// NewAuthCommand creates a new instance of AuthCommand.
func NewAuthCommand(authUseCase usecase.AuthUseCase) *AuthCommand {
	return &AuthCommand{
		authUseCase: authUseCase,
	}
}

// Init initializes the authentication process.
func (c *AuthCommand) Init() error {
	fmt.Println("Initializing Spotify authentication...")

	// Prompt for client ID
	clientID, err := c.promptInput("Enter your Spotify Client ID: ")
	if err != nil {
		return fmt.Errorf("failed to read client ID: %w", err)
	}

	// Prompt for client secret
	clientSecret, err := c.promptInput("Enter your Spotify Client Secret: ")
	if err != nil {
		return fmt.Errorf("failed to read client secret: %w", err)
	}

	// Initialize authentication with the provided credentials
	authURL, err := c.authUseCase.InitAuth(context.Background(), clientID, clientSecret)
	if err != nil {
		return fmt.Errorf("failed to initialize authentication: %w", err)
	}

	// Display the authorization URL
	fmt.Println("\nPlease open the following URL in your browser to authorize the application:")
	fmt.Println(authURL)

	// Start the callback server
	callbackServer := httpinterface.NewCallbackServer(c.authUseCase)
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

// TestCurrentlyPlaying tests the authentication by retrieving the currently playing track.
func (c *AuthCommand) TestCurrentlyPlaying() error {
	fmt.Println("Testing authentication by retrieving currently playing track...")

	track, err := c.authUseCase.GetCurrentlyPlaying(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get currently playing track: %w", err)
	}

	// Check if no track is playing
	if track == "No track currently playing" {
		fmt.Println("No track is currently playing on Spotify. Please start playing a track and try again.")
		return nil
	}

	fmt.Println(track)
	return nil
}

// promptInput prompts the user for input with the given message.
func (c *AuthCommand) promptInput(message string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(message)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}
