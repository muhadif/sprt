package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// CurrentTrackModel is the model for the current track UI
type CurrentTrackModel struct {
	artist      string
	title       string
	album       string
	duration    string
	progress    string
	isPlaying   bool
	albumArt    string
	quitting    bool
	windowWidth int
}

// NewCurrentTrackModel creates a new current track model
func NewCurrentTrackModel(artist, title, album, duration, progress string, isPlaying bool) *CurrentTrackModel {
	return &CurrentTrackModel{
		artist:      artist,
		title:       title,
		album:       album,
		duration:    duration,
		progress:    progress,
		isPlaying:   isPlaying,
		windowWidth: 80,
	}
}

// Init initializes the model
func (m CurrentTrackModel) Init() tea.Cmd {
	return nil
}

// Update updates the model
func (m CurrentTrackModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
	}

	return m, nil
}

// View renders the model
func (m CurrentTrackModel) View() string {
	if m.quitting {
		return ""
	}

	// Get styles from the shared styles
	titleStyle := GetTitleStyle(m.windowWidth)
	headerStyle := GetHeaderStyle()
	valueStyle := GetValueStyle()
	border := GetBorderStyle(m.windowWidth)

	// Build the view
	s := titleStyle.Render("Currently Playing") + "\n\n"

	// Track info
	trackInfo := ""
	trackInfo += headerStyle.Render("Title: ") + valueStyle.Render(m.title) + "\n"
	trackInfo += headerStyle.Render("Artist: ") + valueStyle.Render(m.artist) + "\n"
	trackInfo += headerStyle.Render("Album: ") + valueStyle.Render(m.album) + "\n"
	trackInfo += headerStyle.Render("Duration: ") + valueStyle.Render(m.duration) + "\n"

	// Status
	status := "Paused"
	if m.isPlaying {
		status = "Playing"
	}
	trackInfo += headerStyle.Render("Status: ") + valueStyle.Render(status) + "\n"

	// Progress bar
	if m.progress != "" && m.duration != "" {
		progressPercent := 0.0
		if m.progress != "0:00" && m.duration != "0:00" {
			// Extract minutes and seconds from progress and duration
			var progressMin, progressSec, durationMin, durationSec int
			fmt.Sscanf(m.progress, "%d:%d", &progressMin, &progressSec)
			fmt.Sscanf(m.duration, "%d:%d", &durationMin, &durationSec)

			progressTotal := progressMin*60 + progressSec
			durationTotal := durationMin*60 + durationSec

			if durationTotal > 0 {
				progressPercent = float64(progressTotal) / float64(durationTotal)
			}
		}

		// Create a progress bar
		barWidth := m.windowWidth - 20
		completedWidth := int(float64(barWidth) * progressPercent)
		remainingWidth := barWidth - completedWidth

		progressBar := "["
		progressBar += strings.Repeat("=", completedWidth)
		if remainingWidth > 0 {
			progressBar += ">"
			progressBar += strings.Repeat(" ", remainingWidth-1)
		}
		progressBar += "]"

		trackInfo += headerStyle.Render("Progress: ") + valueStyle.Render(m.progress+" / "+m.duration) + "\n"
		trackInfo += valueStyle.Render(progressBar) + "\n"
	}

	s += border.Render(trackInfo)
	s += "\n\n" + valueStyle.Render("Press q to return to menu")

	return s
}

// RunCurrentTrackUI runs the current track UI
func RunCurrentTrackUI(artist, title, album, duration, progress string, isPlaying bool) error {
	p := tea.NewProgram(NewCurrentTrackModel(artist, title, album, duration, progress, isPlaying), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
