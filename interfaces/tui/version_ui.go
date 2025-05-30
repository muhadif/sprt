package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// VersionModel is the model for the version UI
type VersionModel struct {
	version     string
	buildDate   string
	commitHash  string
	quitting    bool
	windowWidth int
}

// NewVersionModel creates a new version model
func NewVersionModel(version, buildDate, commitHash string) *VersionModel {
	return &VersionModel{
		version:     version,
		buildDate:   buildDate,
		commitHash:  commitHash,
		windowWidth: 80,
	}
}

// Init initializes the model
func (m VersionModel) Init() tea.Cmd {
	return nil
}

// Update updates the model
func (m VersionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
func (m VersionModel) View() string {
	if m.quitting {
		return ""
	}

	// Get styles from the shared styles
	titleStyle := GetTitleStyle(m.windowWidth)
	headerStyle := GetHeaderStyle()
	valueStyle := GetValueStyle()
	border := GetBorderStyle(m.windowWidth)

	// Build the view
	s := titleStyle.Render("Spotify CLI Version Information") + "\n\n"

	// Version info
	versionInfo := ""
	versionInfo += headerStyle.Render("Version: ") + valueStyle.Render(m.version) + "\n"
	versionInfo += headerStyle.Render("Build Date: ") + valueStyle.Render(m.buildDate) + "\n"
	versionInfo += headerStyle.Render("Commit Hash: ") + valueStyle.Render(m.commitHash) + "\n"

	s += border.Render(versionInfo)
	s += "\n\n" + valueStyle.Render("Press q to return to menu")

	return s
}

// RunVersionUI runs the version UI
func RunVersionUI(version, buildDate, commitHash string) error {
	p := tea.NewProgram(NewVersionModel(version, buildDate, commitHash), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
