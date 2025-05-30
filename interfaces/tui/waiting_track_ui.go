package tui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muhadif/sprt/domain/usecase"
)

// WaitingTrackModel is the model for the waiting track UI
type WaitingTrackModel struct {
	authUseCase usecase.AuthUseCase
	status      string
	dots        int
	maxDots     int
	ticker      *time.Ticker
	quitting    bool
	windowWidth int
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewWaitingTrackModel creates a new waiting track model
func NewWaitingTrackModel(authUseCase usecase.AuthUseCase) *WaitingTrackModel {
	ctx, cancel := context.WithCancel(context.Background())
	return &WaitingTrackModel{
		authUseCase: authUseCase,
		status:      "No track currently playing",
		dots:        0,
		maxDots:     3,
		windowWidth: 80,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Init initializes the model
func (m *WaitingTrackModel) Init() tea.Cmd {
	m.ticker = time.NewTicker(500 * time.Millisecond)
	return m.tick
}

// Update updates the model
func (m *WaitingTrackModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			m.ticker.Stop()
			m.cancel()
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
	case tickMsg:
		// Update dots animation
		m.dots = (m.dots + 1) % (m.maxDots + 1)

		// Check if a track is playing
		trackInfo, err := m.authUseCase.GetCurrentlyPlaying(m.ctx)
		if err == nil && trackInfo != "No track currently playing" {
			// Track is now playing, return it
			m.ticker.Stop()
			m.cancel()

			// Parse the track information
			title, artist, album := parseTrackInfo(trackInfo)

			// Create and return the current track model
			return NewCurrentTrackModel(artist, title, album, "Unknown", "Unknown", true), nil
		}

		return m, m.tick
	}

	return m, nil
}

// View renders the model
func (m *WaitingTrackModel) View() string {
	if m.quitting {
		return ""
	}

	// Get styles from the shared styles
	titleStyle := GetTitleStyle(m.windowWidth)
	headerStyle := GetHeaderStyle()
	valueStyle := GetValueStyle()
	border := GetBorderStyle(m.windowWidth)

	// Build the view
	s := titleStyle.Render("Waiting for Track") + "\n\n"

	// Create dots animation
	dots := ""
	for i := 0; i < m.dots; i++ {
		dots += "."
	}

	// Content
	content := headerStyle.Render(m.status) + valueStyle.Render(dots) + "\n\n"
	content += valueStyle.Render("Waiting for a track to play on Spotify") + "\n\n"
	content += valueStyle.Render("Press q to return to menu")

	s += border.Render(content)

	return s
}

// tickMsg is a message sent when the ticker ticks
type tickMsg struct{}

// tick is a command that waits for the ticker to tick
func (m *WaitingTrackModel) tick() tea.Msg {
	select {
	case <-m.ticker.C:
		return tickMsg{}
	case <-m.ctx.Done():
		return nil
	}
}

// RunWaitingTrackUI runs the waiting track UI
func RunWaitingTrackUI(authUseCase usecase.AuthUseCase) error {
	p := tea.NewProgram(NewWaitingTrackModel(authUseCase), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
