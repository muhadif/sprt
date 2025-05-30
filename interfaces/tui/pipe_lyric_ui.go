package tui

import (
	"context"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muhadif/sprt/domain/usecase"
)

// PipeLyricModel is the model for the pipe lyric UI
type PipeLyricModel struct {
	lyrics         *usecase.Lyrics
	currentLine    string
	currentLineIdx int
	width          int
	height         int
	updateCh       <-chan *usecase.LyricUpdate
	ctx            context.Context
	cancel         context.CancelFunc
	err            error
	quitting       bool
	windowWidth    int
}

// NewPipeLyricModel creates a new pipe lyric model
func NewPipeLyricModel(ctx context.Context, startTimeMs int, playerUseCase usecase.PlayerUseCase) (*PipeLyricModel, error) {
	// Create the lyric use case
	lyricUseCase := usecase.NewLyricUseCase()

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(ctx)

	// Get the lyric updates channel
	updateCh := lyricUseCase.GetLyricChannel(ctx, startTimeMs, playerUseCase)

	return &PipeLyricModel{
		currentLine:    "Loading lyrics...",
		currentLineIdx: -1,
		width:          80,
		height:         20,
		updateCh:       updateCh,
		ctx:            ctx,
		cancel:         cancel,
		windowWidth:    80,
	}, nil
}

// Init initializes the model
func (m *PipeLyricModel) Init() tea.Cmd {
	return m.waitForUpdate
}

// Update updates the model
func (m *PipeLyricModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			m.cancel()
			return m, tea.Quit
		}

	case *usecase.LyricUpdate:
		if msg.IsError {
			m.err = fmt.Errorf(msg.ErrorMsg)
			m.currentLine = fmt.Sprintf("Error: %s", msg.ErrorMsg)
		} else if msg.Lyrics != nil {
			m.lyrics = msg.Lyrics
			m.currentLineIdx = msg.LineIndex
			m.currentLine = msg.Text

			// Write the current line to a file for external use
			if msg.Text != "" {
				err := os.WriteFile("/tmp/current-lyric.txt", []byte(msg.Text), 0644)
				if err != nil {
					m.err = fmt.Errorf("error writing to file: %v", err)
				}
			}
		}

		return m, m.waitForUpdate
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
	}

	return m, nil
}

// View renders the model
func (m *PipeLyricModel) View() string {
	if m.quitting {
		return ""
	}

	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress q to quit.", m.err)
	}

	// Get styles from the shared styles
	titleStyle := GetTitleStyle(m.windowWidth)
	currentLineStyle := GetCurrentLineStyle(m.windowWidth)
	infoStyle := GetInfoStyle()

	// Build the view
	var sb strings.Builder

	// Add a title
	if m.lyrics != nil {
		title := fmt.Sprintf("%s - %s", m.lyrics.Artist, m.lyrics.Name)
		sb.WriteString(titleStyle.Render(title))
		sb.WriteString("\n\n")
	} else {
		sb.WriteString(titleStyle.Render("Spotify Lyrics"))
		sb.WriteString("\n\n")
	}

	// Display the current line
	sb.WriteString(currentLineStyle.Render(m.currentLine))
	sb.WriteString("\n\n")

	// Add a footer
	sb.WriteString(infoStyle.Render("Press q to quit"))

	return sb.String()
}

// waitForUpdate waits for an update from the lyric channel
func (m *PipeLyricModel) waitForUpdate() tea.Msg {
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

// RunPipeLyricUI runs the pipe lyric UI
func RunPipeLyricUI(ctx context.Context, startTimeMs int, playerUseCase usecase.PlayerUseCase) error {
	model, err := NewPipeLyricModel(ctx, startTimeMs, playerUseCase)
	if err != nil {
		return err
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}
