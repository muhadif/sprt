package tui

import (
	"context"
	"github.com/muhadif/sprt/domain/usecase"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// AppModel is the main application model that manages transitions between screens
type AppModel struct {
	currentScreen tea.Model
	nextScreen    tea.Model
	transitioning bool
	progress      float64
	windowWidth   int
	windowHeight  int

	// Use cases
	authUseCase   usecase.AuthUseCase
	playerUseCase usecase.PlayerUseCase
	lyricUseCase  usecase.LyricUseCase

	// Context for cancellation
	ctx    context.Context
	cancel context.CancelFunc
}

// TransitionMsg is a message sent when a transition to a new screen is requested
type TransitionMsg struct {
	Screen tea.Model
}

// TransitionTickMsg is a message sent when the transition animation should be updated
type TransitionTickMsg time.Time

// NewAppModel creates a new app model with the main menu as the initial screen
func NewAppModel(authUseCase usecase.AuthUseCase, playerUseCase usecase.PlayerUseCase, lyricUseCase usecase.LyricUseCase) *AppModel {
	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	return &AppModel{
		currentScreen: NewMenuModel(),
		transitioning: false,
		progress:      0,
		windowWidth:   80,
		windowHeight:  24,
		authUseCase:   authUseCase,
		playerUseCase: playerUseCase,
		lyricUseCase:  lyricUseCase,
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Init initializes the model
func (m *AppModel) Init() tea.Cmd {
	return tea.Batch(
		m.currentScreen.Init(),
	)
}

// Update updates the model
func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height

		// Pass the window size to the current screen
		if m.currentScreen != nil {
			_, _ = m.currentScreen.Update(msg)
		}

		// Pass the window size to the next screen if we're transitioning
		if m.transitioning && m.nextScreen != nil {
			_, _ = m.nextScreen.Update(msg)
		}

	case TransitionMsg:
		// Start a transition to the new screen
		m.nextScreen = msg.Screen
		m.transitioning = true
		m.progress = 0

		// Initialize the next screen
		cmd := m.nextScreen.Init()

		// Start the transition animation
		return m, tea.Batch(cmd, m.tickTransition())

	case TransitionTickMsg:
		if !m.transitioning {
			return m, nil
		}

		// Update the transition progress
		m.progress += 0.1

		if m.progress >= 1.0 {
			// Transition complete, switch to the next screen
			m.currentScreen = m.nextScreen
			m.nextScreen = nil
			m.transitioning = false
			return m, nil
		}

		// Continue the transition animation
		return m, m.tickTransition()

	default:
		// If we're transitioning, don't pass messages to the current screen
		if m.transitioning {
			return m, nil
		}

		// Pass the message to the current screen
		newModel, cmd := m.currentScreen.Update(msg)

		// Check if the current screen has changed
		if newModel != m.currentScreen {
			m.currentScreen = newModel.(tea.Model)
		}

		// Handle menu selection
		if menuModel, ok := m.currentScreen.(MenuModel); ok && menuModel.choice != "" && menuModel.choice != "quit" {
			// User selected a menu item, transition to the appropriate screen
			var nextScreen tea.Model

			switch menuModel.choice {
			case "current":
				nextScreen = NewCurrentTrackModel("", "", "", "", "", false)
			case "lyric show":
				// We'll need to initialize this properly later
				// For now, just create a placeholder
				nextScreen = &LyricModel{}
			case "lyric pipe":
				// We'll need to initialize this properly later
				// For now, just create a placeholder
				nextScreen = &PipeLyricModel{}
			case "auth init":
				nextScreen = NewAuthModel()
			case "version":
				nextScreen = NewVersionModel("", "", "")
			default:
				// Unknown command, stay on the menu
				return m, cmd
			}

			// Start the transition to the new screen
			return m, tea.Batch(cmd, func() tea.Msg {
				return TransitionMsg{Screen: nextScreen}
			})
		}

		return m, cmd
	}

	return m, nil
}

// View renders the model
func (m *AppModel) View() string {
	if m.transitioning && m.nextScreen != nil {
		// Render a transition between the current and next screens
		// This is a simple fade transition for now
		return renderTransition(
			m.currentScreen.View(),
			m.nextScreen.View(),
			m.progress,
		)
	}

	// Render the current screen
	return m.currentScreen.View()
}

// tickTransition returns a command that will tick the transition animation
func (m *AppModel) tickTransition() tea.Cmd {
	return tea.Tick(time.Millisecond*50, func(t time.Time) tea.Msg {
		return TransitionTickMsg(t)
	})
}

// renderTransition renders a transition between two screens
func renderTransition(currentView, nextView string, progress float64) string {
	// For a simple fade effect, we'll use a crossfade approach
	// As progress increases, we'll show more of the next view and less of the current view

	// If we're at the beginning or end of the transition, just show the appropriate view
	if progress <= 0 {
		return currentView
	}
	if progress >= 1 {
		return nextView
	}

	// For a simple crossfade, we'll use a slide-in effect
	// We'll split both views into lines and blend them based on progress
	currentLines := strings.Split(currentView, "\n")
	nextLines := strings.Split(nextView, "\n")

	// Determine the maximum number of lines
	maxLines := len(currentLines)
	if len(nextLines) > maxLines {
		maxLines = len(nextLines)
	}

	// Build the blended view
	var result strings.Builder

	for i := 0; i < maxLines; i++ {
		// Get the current and next lines, or empty strings if they don't exist
		var currentLine, nextLine string
		if i < len(currentLines) {
			currentLine = currentLines[i]
		}
		if i < len(nextLines) {
			nextLine = nextLines[i]
		}

		// Calculate the position for the slide effect
		// As progress increases, the next line slides in from the right
		slidePos := int(float64(len(nextLine)) * (1.0 - progress))

		// If we're in the first half of the transition, show the current line fading out
		if progress < 0.5 {
			// Fade out the current line by replacing characters with spaces from right to left
			fadePos := int(float64(len(currentLine)) * (progress * 2))
			if fadePos < len(currentLine) {
				// Replace characters from fadePos to the end with spaces
				currentLine = currentLine[:fadePos] + strings.Repeat(" ", len(currentLine)-fadePos)
			}
			result.WriteString(currentLine)
		} else {
			// In the second half, show the next line sliding in
			if slidePos > 0 {
				// Pad the next line with spaces on the left
				nextLine = strings.Repeat(" ", slidePos) + nextLine[:len(nextLine)-slidePos]
			}
			result.WriteString(nextLine)
		}

		// Add a newline unless this is the last line
		if i < maxLines-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

// RunApp runs the main application UI
func RunApp(authUseCase usecase.AuthUseCase, playerUseCase usecase.PlayerUseCase, lyricUseCase usecase.LyricUseCase) error {
	p := tea.NewProgram(NewAppModel(authUseCase, playerUseCase, lyricUseCase), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
