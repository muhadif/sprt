package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muhadif/sprt/domain/usecase"
)

// MenuWithTransitionModel is the model for the menu UI with transitions
type MenuWithTransitionModel struct {
	menuModel     *MenuModel
	transitionMgr *TransitionManager
	nextScreen    tea.Model
	authUseCase   usecase.AuthUseCase
	playerUseCase usecase.PlayerUseCase
	lyricUseCase  usecase.LyricUseCase
	ctx           context.Context
	cancel        context.CancelFunc
	windowWidth   int
	windowHeight  int
	version       string
	buildDate     string
	commitHash    string
}

func NewMenuWithTransitionModel(authUseCase usecase.AuthUseCase, playerUseCase usecase.PlayerUseCase, lyricUseCase usecase.LyricUseCase, version, buildDate, commitHash string) *MenuWithTransitionModel {
	ctx, cancel := context.WithCancel(context.Background())

	return &MenuWithTransitionModel{
		menuModel:     NewMenuModel(),
		transitionMgr: NewTransitionManager(),
		authUseCase:   authUseCase,
		playerUseCase: playerUseCase,
		lyricUseCase:  lyricUseCase,
		ctx:           ctx,
		cancel:        cancel,
		windowWidth:   80,
		windowHeight:  24,
		version:       version,
		buildDate:     buildDate,
		commitHash:    commitHash,
	}
}

// Init initializes the model
func (m *MenuWithTransitionModel) Init() tea.Cmd {
	return m.menuModel.Init()
}

// Update updates the model
func (m *MenuWithTransitionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height

		// Pass the window size to the menu model
		_, _ = m.menuModel.Update(msg)

		// Pass the window size to the next screen if we're transitioning
		if m.transitionMgr.IsTransitioning() && m.nextScreen != nil {
			_, _ = m.nextScreen.Update(msg)
		}

	case TransitionUpdateMsg:
		// Update the transition
		done, cmd := m.transitionMgr.UpdateTransition()
		if done {
			// Transition complete, switch to the next screen
			return m.nextScreen, nil
		}
		return m, cmd

	default:
		// If we're transitioning, don't pass messages to the menu model
		if m.transitionMgr.IsTransitioning() {
			return m, nil
		}

		// Pass the message to the menu model
		newModel, cmd := m.menuModel.Update(msg)

		// Check if the menu model has changed
		if menuModel, ok := newModel.(MenuModel); ok {
			m.menuModel = &menuModel

			// Check if a menu item was selected
			if m.menuModel.choice != "" && m.menuModel.choice != "quit" {
				// User selected a menu item, transition to the appropriate screen
				var nextScreen tea.Model

				switch m.menuModel.choice {
				case "current":
					// Get the currently playing track
					trackInfo, err := m.authUseCase.GetCurrentlyPlaying(m.ctx)
					if err != nil {
						// Handle error
						return m, cmd
					}

					// Check if no track is playing
					if trackInfo == "No track currently playing" {
						// Show waiting screen instead of returning to menu
						nextScreen = NewWaitingTrackModel(m.authUseCase)
					} else {
						// Parse the track information
						title, artist, album := parseTrackInfo(trackInfo)

						// Create the current track model
						nextScreen = NewCurrentTrackModel(artist, title, album, "Unknown", "Unknown", true)
					}

				case "lyric show":
					// This requires more complex initialization, so we'll just return to the menu for now
					return m, cmd

				case "lyric pipe":
					// This requires more complex initialization, so we'll just return to the menu for now
					return m, cmd

				case "auth init":
					// Return to the menu and let the showTUIMenu function handle the execution of the command
					return m, cmd

				case "version":
					// Use the version information from the struct
					nextScreen = NewVersionModel(m.version, m.buildDate, m.commitHash)

				default:
					// Unknown command, stay on the menu
					return m, cmd
				}

				// Start the transition to the new screen
				m.nextScreen = nextScreen

				// Initialize the next screen
				initCmd := m.nextScreen.Init()

				// Start the transition
				transitionCmd := m.transitionMgr.StartTransition(m.View(), m.nextScreen.View())

				return m, tea.Batch(initCmd, transitionCmd)
			}
		}

		return m, cmd
	}

	return m, nil
}

// View renders the model
func (m *MenuWithTransitionModel) View() string {
	if m.transitionMgr.IsTransitioning() {
		return m.transitionMgr.RenderTransition()
	}

	return m.menuModel.View()
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

// RunMenuWithTransition runs the menu UI with transitions
func RunMenuWithTransition(authUseCase usecase.AuthUseCase, playerUseCase usecase.PlayerUseCase, lyricUseCase usecase.LyricUseCase, version, buildDate, commitHash string) (string, error) {
	model := NewMenuWithTransitionModel(authUseCase, playerUseCase, lyricUseCase, version, buildDate, commitHash)
	p := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	// Check if we ended up with a menu model
	if menuModel, ok := finalModel.(*MenuWithTransitionModel); ok {
		if menuModel.menuModel.choice == "quit" {
			return "quit", nil
		}
		return menuModel.menuModel.choice, nil
	}

	// If we ended up with a different model, try to get its choice
	// This is a bit of a hack, but it should work for most cases
	switch finalModel.(type) {
	case *CurrentTrackModel:
		return "current", nil
	case *LyricModel:
		return "lyric show", nil
	case *PipeLyricModel:
		return "lyric pipe", nil
	case *AuthModel:
		return "auth init", nil
	case *VersionModel:
		return "version", nil
	}

	return "", fmt.Errorf("unknown model type")
}
