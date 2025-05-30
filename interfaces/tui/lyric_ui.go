package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muhadif/sprt/config"
	"github.com/muhadif/sprt/domain/usecase"
)

// LyricModel is the model for the lyric UI
type LyricModel struct {
	lyrics         *usecase.Lyrics
	lines          []string
	currentLineIdx int
	prevLineIdx    int
	width          int
	height         int
	uiConfig       *config.UIConfig
	updateCh       <-chan *usecase.LyricUpdate
	ctx            context.Context
	cancel         context.CancelFunc
	err            error

	// Animation state
	animating       bool
	animationStep   int
	animationSteps  int
	animationType   string
	animationTicker *time.Ticker
}

// NewLyricModel creates a new lyric model
func NewLyricModel(ctx context.Context, startTimeMs int, playerUseCase usecase.PlayerUseCase) (*LyricModel, error) {
	// Load UI config
	uiConfig, err := config.LoadUIConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load UI config: %w", err)
	}

	// Create the lyric use case
	lyricUseCase := usecase.NewLyricUseCase()

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(ctx)

	// Get the lyric updates channel
	updateCh := lyricUseCase.GetLyricChannel(ctx, startTimeMs, playerUseCase)

	return &LyricModel{
		lines:          []string{"Loading lyrics..."},
		currentLineIdx: -1,
		prevLineIdx:    -1,
		width:          uiConfig.Lyric.Width,
		height:         uiConfig.Lyric.Height,
		uiConfig:       uiConfig,
		updateCh:       updateCh,
		ctx:            ctx,
		cancel:         cancel,
		animating:      false,
		animationType:  uiConfig.Lyric.Animation.Type,
		animationSteps: uiConfig.Lyric.Animation.FadeSteps,
	}, nil
}

// Init initializes the model
func (m *LyricModel) Init() tea.Cmd {
	return m.waitForUpdate
}

// Update updates the model
func (m *LyricModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.cancel()
			if m.animationTicker != nil {
				m.animationTicker.Stop()
			}
			return m, tea.Quit
		}

	case *usecase.LyricUpdate:
		if msg.IsError {
			m.err = fmt.Errorf(msg.ErrorMsg)
			m.lines = []string{fmt.Sprintf("Error: %s", msg.ErrorMsg)}
		} else if msg.Lyrics != nil {
			m.lyrics = msg.Lyrics

			// Store previous line index for animation
			if m.currentLineIdx != msg.LineIndex {
				m.prevLineIdx = m.currentLineIdx
				m.currentLineIdx = msg.LineIndex

				// Start animation if enabled
				if m.uiConfig.Lyric.Animation.Enabled && m.prevLineIdx != -1 {
					m.startAnimation()
				}
			}

			// Build the lines array with all lyrics
			if len(m.lyrics.Lines) > 0 {
				m.lines = make([]string, len(m.lyrics.Lines))
				for i, line := range m.lyrics.Lines {
					m.lines[i] = line.Text
				}
			}
		}

		return m, m.waitForUpdate

	case animationTickMsg:
		if m.animating {
			m.animationStep++
			if m.animationStep >= m.animationSteps {
				m.animating = false
				if m.animationTicker != nil {
					m.animationTicker.Stop()
					m.animationTicker = nil
				}
			}
			return m, nil
		}
	}

	return m, nil
}

// animationTickMsg is a message sent when the animation ticker ticks
type animationTickMsg struct{}

// startAnimation starts the animation for transitioning between lyric lines
func (m *LyricModel) startAnimation() {
	if m.animationTicker != nil {
		m.animationTicker.Stop()
	}

	m.animating = true
	m.animationStep = 0

	// Calculate tick duration based on total animation duration and steps
	tickDuration := time.Duration(m.uiConfig.Lyric.Animation.DurationMs) * time.Millisecond / time.Duration(m.animationSteps)
	m.animationTicker = time.NewTicker(tickDuration)

	// Send animation tick messages
	go func() {
		for range m.animationTicker.C {
			if !m.animating {
				return
			}
			cmd := func() tea.Msg {
				return animationTickMsg{}
			}
			tea.NewProgram(m).Send(cmd())
		}
	}()
}

// View renders the model
func (m *LyricModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress q to quit.", m.err)
	}

	// Get base styles from the shared styles
	titleStyle := GetTitleStyle(m.width)

	// Create styles for current and other lines based on config
	currentStyle := GetCurrentLineStyle(m.width)
	otherStyle := GetOtherLineStyle(m.width)
	prevStyle := GetOtherLineStyle(m.width)

	// Apply custom styling from config if available
	if m.uiConfig.Lyric.CurrentLineStyle.ForegroundColor != "" {
		currentStyle = currentStyle.Foreground(lipgloss.Color(m.uiConfig.Lyric.CurrentLineStyle.ForegroundColor))
	}
	if m.uiConfig.Lyric.CurrentLineStyle.BackgroundColor != "" {
		currentStyle = currentStyle.Background(lipgloss.Color(m.uiConfig.Lyric.CurrentLineStyle.BackgroundColor))
	}
	if m.uiConfig.Lyric.CurrentLineStyle.Bold {
		currentStyle = currentStyle.Bold(true)
	}
	if m.uiConfig.Lyric.CurrentLineStyle.Italic {
		currentStyle = currentStyle.Italic(true)
	}
	if m.uiConfig.Lyric.CurrentLineStyle.Underline {
		currentStyle = currentStyle.Underline(true)
	}

	// Apply custom styling for other lines from config if available
	if m.uiConfig.Lyric.OtherLineStyle.ForegroundColor != "" {
		otherStyle = otherStyle.Foreground(lipgloss.Color(m.uiConfig.Lyric.OtherLineStyle.ForegroundColor))
		prevStyle = prevStyle.Foreground(lipgloss.Color(m.uiConfig.Lyric.OtherLineStyle.ForegroundColor))
	}
	if m.uiConfig.Lyric.OtherLineStyle.BackgroundColor != "" {
		otherStyle = otherStyle.Background(lipgloss.Color(m.uiConfig.Lyric.OtherLineStyle.BackgroundColor))
		prevStyle = prevStyle.Background(lipgloss.Color(m.uiConfig.Lyric.OtherLineStyle.BackgroundColor))
	}
	if m.uiConfig.Lyric.OtherLineStyle.Bold {
		otherStyle = otherStyle.Bold(true)
		prevStyle = prevStyle.Bold(true)
	}
	if m.uiConfig.Lyric.OtherLineStyle.Italic {
		otherStyle = otherStyle.Italic(true)
		prevStyle = prevStyle.Italic(true)
	}
	if m.uiConfig.Lyric.OtherLineStyle.Underline {
		otherStyle = otherStyle.Underline(true)
		prevStyle = prevStyle.Underline(true)
	}

	// Build the view
	var sb strings.Builder

	// Add a title
	if m.lyrics != nil {
		title := fmt.Sprintf("%s - %s", m.lyrics.Artist, m.lyrics.Name)
		sb.WriteString(titleStyle.Render(title))
		sb.WriteString("\n\n")
	}

	// Calculate how many lines to show before and after the current line
	linesBeforeAfter := (m.height - 3) / 2 // -3 for title and spacing
	startIdx := max(0, m.currentLineIdx-linesBeforeAfter)
	endIdx := min(len(m.lines), m.currentLineIdx+linesBeforeAfter+1)

	// Show all lyrics with the current line highlighted
	for i := startIdx; i < endIdx; i++ {
		line := m.lines[i]

		// Apply animation if enabled and currently animating
		if m.animating && m.uiConfig.Lyric.Animation.Enabled {
			if i == m.currentLineIdx {
				// Current line is fading in
				if m.animationType == "fade" {
					// Calculate fade-in progress (0.0 to 1.0)
					progress := float64(m.animationStep) / float64(m.animationSteps)

					// Interpolate between other style and current style
					fgColor := interpolateColor(
						m.uiConfig.Lyric.OtherLineStyle.ForegroundColor,
						m.uiConfig.Lyric.CurrentLineStyle.ForegroundColor,
						progress,
					)

					// Create a style with the interpolated color
					fadeStyle := lipgloss.NewStyle().
						Foreground(lipgloss.Color(fgColor)).
						Width(m.width).
						Align(lipgloss.Center)

					if m.uiConfig.Lyric.CurrentLineStyle.Bold {
						fadeStyle = fadeStyle.Bold(progress > 0.5)
					}

					sb.WriteString(fadeStyle.Render(line))
				} else if m.animationType == "slide" {
					// Slide animation
					slideDistance := m.uiConfig.Lyric.Animation.SlideDistance
					progress := float64(m.animationStep) / float64(m.animationSteps)

					// Calculate padding based on progress
					padding := int(float64(slideDistance) * (1.0 - progress))
					paddedLine := strings.Repeat(" ", padding) + line

					sb.WriteString(currentStyle.Render(paddedLine))
				} else {
					// No animation or unknown type
					sb.WriteString(currentStyle.Render(line))
				}
			} else if i == m.prevLineIdx {
				// Previous line is fading out
				if m.animationType == "fade" {
					// Calculate fade-out progress (0.0 to 1.0)
					progress := float64(m.animationStep) / float64(m.animationSteps)

					// Interpolate between current style and other style
					fgColor := interpolateColor(
						m.uiConfig.Lyric.CurrentLineStyle.ForegroundColor,
						m.uiConfig.Lyric.OtherLineStyle.ForegroundColor,
						progress,
					)

					// Create a style with the interpolated color
					fadeStyle := lipgloss.NewStyle().
						Foreground(lipgloss.Color(fgColor)).
						Width(m.width).
						Align(lipgloss.Center)

					if m.uiConfig.Lyric.CurrentLineStyle.Bold {
						fadeStyle = fadeStyle.Bold(progress < 0.5)
					}

					sb.WriteString(fadeStyle.Render(line))
				} else if m.animationType == "slide" {
					// Slide animation
					slideDistance := m.uiConfig.Lyric.Animation.SlideDistance
					progress := float64(m.animationStep) / float64(m.animationSteps)

					// Calculate padding based on progress
					padding := int(float64(slideDistance) * progress)
					paddedLine := strings.Repeat(" ", padding) + line

					sb.WriteString(prevStyle.Render(paddedLine))
				} else {
					// No animation or unknown type
					sb.WriteString(otherStyle.Render(line))
				}
			} else {
				sb.WriteString(otherStyle.Render(line))
			}
		} else {
			// No animation
			if i == m.currentLineIdx {
				sb.WriteString(currentStyle.Render(line))
			} else {
				sb.WriteString(otherStyle.Render(line))
			}
		}

		sb.WriteString("\n")
	}

	// Add a footer
	sb.WriteString("\nPress q to quit")

	return sb.String()
}

// interpolateColor interpolates between two hex colors
func interpolateColor(startColor, endColor string, progress float64) string {
	// Parse hex colors
	var startR, startG, startB, endR, endG, endB int
	fmt.Sscanf(startColor, "#%02x%02x%02x", &startR, &startG, &startB)
	fmt.Sscanf(endColor, "#%02x%02x%02x", &endR, &endG, &endB)

	// Interpolate
	r := int(float64(startR) + progress*float64(endR-startR))
	g := int(float64(startG) + progress*float64(endG-startG))
	b := int(float64(startB) + progress*float64(endB-startB))

	// Clamp values
	r = max(0, min(255, r))
	g = max(0, min(255, g))
	b = max(0, min(255, b))

	// Return hex color
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// waitForUpdate waits for an update from the lyric channel
func (m *LyricModel) waitForUpdate() tea.Msg {
	select {
	case update, ok := <-m.updateCh:
		if !ok {
			return tea.Quit
		}
		return update
	case <-m.ctx.Done():
		return tea.Quit
	}
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// RunLyricUI runs the lyric UI
func RunLyricUI(ctx context.Context, startTimeMs int, playerUseCase usecase.PlayerUseCase) error {
	model, err := NewLyricModel(ctx, startTimeMs, playerUseCase)
	if err != nil {
		return err
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}
