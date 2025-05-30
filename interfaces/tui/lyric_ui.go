package tui

import (
	"context"
	"fmt"
	"strings"

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
	width          int
	height         int
	uiConfig       *config.UIConfig
	updateCh       <-chan *usecase.LyricUpdate
	ctx            context.Context
	cancel         context.CancelFunc
	err            error
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
		width:          uiConfig.Lyric.Width,
		height:         uiConfig.Lyric.Height,
		uiConfig:       uiConfig,
		updateCh:       updateCh,
		ctx:            ctx,
		cancel:         cancel,
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
			return m, tea.Quit
		}

	case *usecase.LyricUpdate:
		if msg.IsError {
			m.err = fmt.Errorf(msg.ErrorMsg)
			m.lines = []string{fmt.Sprintf("Error: %s", msg.ErrorMsg)}
		} else if msg.Lyrics != nil {
			m.lyrics = msg.Lyrics
			m.currentLineIdx = msg.LineIndex

			// Build the lines array with all lyrics
			if len(m.lyrics.Lines) > 0 {
				m.lines = make([]string, len(m.lyrics.Lines))
				for i, line := range m.lyrics.Lines {
					m.lines[i] = line.Text
				}
			}
		}

		return m, m.waitForUpdate
	}

	return m, nil
}

// View renders the model
func (m *LyricModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress q to quit.", m.err)
	}

	// Create styles for current and other lines
	currentStyle := lipgloss.NewStyle()
	otherStyle := lipgloss.NewStyle()

	// Apply current line style from config
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

	// Apply other line style from config
	if m.uiConfig.Lyric.OtherLineStyle.ForegroundColor != "" {
		otherStyle = otherStyle.Foreground(lipgloss.Color(m.uiConfig.Lyric.OtherLineStyle.ForegroundColor))
	}
	if m.uiConfig.Lyric.OtherLineStyle.BackgroundColor != "" {
		otherStyle = otherStyle.Background(lipgloss.Color(m.uiConfig.Lyric.OtherLineStyle.BackgroundColor))
	}
	if m.uiConfig.Lyric.OtherLineStyle.Bold {
		otherStyle = otherStyle.Bold(true)
	}
	if m.uiConfig.Lyric.OtherLineStyle.Italic {
		otherStyle = otherStyle.Italic(true)
	}
	if m.uiConfig.Lyric.OtherLineStyle.Underline {
		otherStyle = otherStyle.Underline(true)
	}

	// Center the text
	currentStyle = currentStyle.Width(m.width).Align(lipgloss.Center)
	otherStyle = otherStyle.Width(m.width).Align(lipgloss.Center)

	// Build the view
	var sb strings.Builder

	// Add a title
	if m.lyrics != nil {
		title := fmt.Sprintf("%s - %s", m.lyrics.Artist, m.lyrics.Name)
		sb.WriteString(lipgloss.NewStyle().Bold(true).Width(m.width).Align(lipgloss.Center).Render(title))
		sb.WriteString("\n\n")
	}

	// Calculate how many lines to show before and after the current line
	linesBeforeAfter := (m.height - 3) / 2 // -3 for title and spacing
	startIdx := max(0, m.currentLineIdx-linesBeforeAfter)
	endIdx := min(len(m.lines), m.currentLineIdx+linesBeforeAfter+1)

	// Show all lyrics with the current line highlighted
	for i := startIdx; i < endIdx; i++ {
		line := m.lines[i]
		if i == m.currentLineIdx {
			sb.WriteString(currentStyle.Render(line))
		} else {
			sb.WriteString(otherStyle.Render(line))
		}
		sb.WriteString("\n")
	}

	// Add a footer
	sb.WriteString("\nPress q to quit")

	return sb.String()
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
