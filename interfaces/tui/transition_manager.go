package tui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TransitionManager handles transitions between different UI screens
type TransitionManager struct {
	currentView   string
	nextView      string
	transitioning bool
	progress      float64
}

// TransitionUpdateMsg is a message sent when the transition animation should be updated
type TransitionUpdateMsg time.Time

// NewTransitionManager creates a new transition manager
func NewTransitionManager() *TransitionManager {
	return &TransitionManager{
		transitioning: false,
		progress:      0,
	}
}

// StartTransition starts a transition to a new view
func (m *TransitionManager) StartTransition(currentView, nextView string) tea.Cmd {
	m.currentView = currentView
	m.nextView = nextView
	m.transitioning = true
	m.progress = 0

	// Start the transition animation
	return m.tickTransition()
}

// UpdateTransition updates the transition progress
func (m *TransitionManager) UpdateTransition() (bool, tea.Cmd) {
	if !m.transitioning {
		return false, nil
	}

	// Update the transition progress
	m.progress += 0.1

	if m.progress >= 1.0 {
		// Transition complete
		m.transitioning = false
		return true, nil
	}

	// Continue the transition animation
	return false, m.tickTransition()
}

// RenderTransition renders the transition between two views
func (m *TransitionManager) RenderTransition() string {
	if !m.transitioning {
		return m.nextView
	}

	return renderViewTransition(m.currentView, m.nextView, m.progress)
}

// IsTransitioning returns whether a transition is in progress
func (m *TransitionManager) IsTransitioning() bool {
	return m.transitioning
}

// tickTransition returns a command that will tick the transition animation
func (m *TransitionManager) tickTransition() tea.Cmd {
	return tea.Tick(time.Millisecond*10, func(t time.Time) tea.Msg {
		return TransitionUpdateMsg(t)
	})
}

// renderViewTransition renders a transition between two screens
func renderViewTransition(currentView, nextView string, progress float64) string {
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
			if slidePos > 0 && slidePos < len(nextLine) {
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
