package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/muhadif/sprt/domain/usecase"
	"github.com/muhadif/sprt/interfaces/tui"
	"github.com/spf13/cobra"
)

// Version information
var (
	version string
	commit  string
	date    string
)

// Use cases
var (
	authUseCase   usecase.AuthUseCase
	playerUseCase usecase.PlayerUseCase
	lyricUseCase  usecase.LyricUseCase
)

var rootCmd = &cobra.Command{
	Use:   "sprt",
	Short: "sprt - A command-line interface for Spotify",
	Long: `sprt is a command-line interface for interacting with Spotify.
It allows you to authenticate with Spotify, get information about your currently playing track,
and display synchronized lyrics for the current track.`,
}

// InitializeCommands initializes all commands with the provided use cases and version information.
// This is called by main.main() to set up dependency injection.
func InitializeCommands(auth usecase.AuthUseCase, player usecase.PlayerUseCase, lyric usecase.LyricUseCase, ver, com, dt string) {
	// Set use cases
	authUseCase = auth
	playerUseCase = player
	lyricUseCase = lyric

	// Set version information
	version = ver
	commit = com
	date = dt

	// Initialize all commands
	initAuthCommand()
	initCurrentCommand()
	initLyricCommand()
	initVersionCommand()
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// Check if any arguments were provided
	if len(os.Args) > 1 {
		// If arguments were provided, use the standard Cobra command execution
		if err := rootCmd.Execute(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return
	}

	// If no arguments were provided, show the TUI menu
	showTUIMenu()
}

// showTUIMenu displays the TUI menu and executes the selected command
func showTUIMenu() {
	// Run the main menu with transitions
	choice, err := tui.RunMenuWithTransition(authUseCase, playerUseCase, lyricUseCase)
	if err != nil {
		fmt.Printf("Error running menu: %v\n", err)
		os.Exit(1)
	}

	// If the user chose to quit, exit
	if choice == "quit" || choice == "" {
		return
	}

	// Execute the selected command
	// Split the choice into separate arguments
	args := strings.Split(choice, " ")
	os.Args = append(os.Args, args...)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Helper functions to initialize each command
func initAuthCommand() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(authInitCmd)
	authCmd.AddCommand(authTestCmd)
}

func initCurrentCommand() {
	rootCmd.AddCommand(currentCmd)
}

func initLyricCommand() {
	rootCmd.AddCommand(lyricCmd)
	lyricCmd.AddCommand(pipeLyricCmd)
	lyricCmd.AddCommand(showLyricCmd)
}

// Version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Long:  `Print the version, build date, and commit hash of the application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return tui.RunVersionUI(version, date, commit)
	},
}

func initVersionCommand() {
	rootCmd.AddCommand(versionCmd)
}
