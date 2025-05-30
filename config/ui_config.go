package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// UIConfig holds the configuration for the UI
type UIConfig struct {
	Lyric LyricConfig `json:"lyric"`
}

// LyricConfig holds the configuration for the lyric display
type LyricConfig struct {
	CurrentLineStyle StyleConfig     `json:"currentLineStyle"`
	OtherLineStyle   StyleConfig     `json:"otherLineStyle"`
	Width            int             `json:"width"`
	Height           int             `json:"height"`
	Animation        AnimationConfig `json:"animation"`
}

// AnimationConfig holds the configuration for animations
type AnimationConfig struct {
	Enabled       bool   `json:"enabled"`
	Type          string `json:"type"`          // "fade", "slide", "none"
	DurationMs    int    `json:"durationMs"`    // Duration of the animation in milliseconds
	FadeSteps     int    `json:"fadeSteps"`     // Number of steps for fade animation
	SlideDistance int    `json:"slideDistance"` // Distance to slide in characters
}

// StyleConfig holds the configuration for a style
type StyleConfig struct {
	ForegroundColor string `json:"foregroundColor"`
	BackgroundColor string `json:"backgroundColor"`
	Bold            bool   `json:"bold"`
	Italic          bool   `json:"italic"`
	Underline       bool   `json:"underline"`
}

// DefaultUIConfig returns the default UI configuration
func DefaultUIConfig() *UIConfig {
	return &UIConfig{
		Lyric: LyricConfig{
			CurrentLineStyle: StyleConfig{
				ForegroundColor: "#00FF00", // Green
				BackgroundColor: "",
				Bold:            true,
				Italic:          false,
				Underline:       false,
			},
			OtherLineStyle: StyleConfig{
				ForegroundColor: "#FFFFFF", // White
				BackgroundColor: "",
				Bold:            false,
				Italic:          false,
				Underline:       false,
			},
			Width:  80,
			Height: 20,
			Animation: AnimationConfig{
				Enabled:       true,
				Type:          "fade",
				DurationMs:    300,
				FadeSteps:     5,
				SlideDistance: 3,
			},
		},
	}
}

// LoadUIConfig loads the UI configuration from the config file
func LoadUIConfig() (*UIConfig, error) {
	// Get the home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return DefaultUIConfig(), fmt.Errorf("failed to get home directory: %w", err)
	}

	// Create the config directory path
	configDir := filepath.Join(homeDir, ".sprt")
	configFile := filepath.Join(configDir, "ui_config.json")

	// Check if the config file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// Create the default config
		config := DefaultUIConfig()

		// Create the config directory if it doesn't exist
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return config, fmt.Errorf("failed to create config directory: %w", err)
		}

		// Save the default config
		if err := SaveUIConfig(config); err != nil {
			return config, fmt.Errorf("failed to save default config: %w", err)
		}

		return config, nil
	}

	// Read the config file
	data, err := os.ReadFile(configFile)
	if err != nil {
		return DefaultUIConfig(), fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse the config
	var config UIConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return DefaultUIConfig(), fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// SaveUIConfig saves the UI configuration to the config file
func SaveUIConfig(config *UIConfig) error {
	// Get the home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Create the config directory path
	configDir := filepath.Join(homeDir, ".sprt")
	configFile := filepath.Join(configDir, "ui_config.json")

	// Create the config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal the config to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write the config file
	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
