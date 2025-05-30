package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// MenuItem represents an item in the menu
type MenuItem struct {
	title       string
	description string
	command     string
}

// MenuModel is the model for the menu UI
type MenuModel struct {
	items       []MenuItem
	cursor      int
	choice      string
	quitting    bool
	windowWidth int
}

// NewMenuModel creates a new menu model
func NewMenuModel() *MenuModel {
	return &MenuModel{
		items: []MenuItem{
			{title: "Current Track", description: "Display information about the currently playing track", command: "current"},
			{title: "Show Lyrics", description: "Display lyrics with a nice UI", command: "lyric show"},
			{title: "Pipe Lyrics", description: "Display lyrics in the terminal", command: "lyric pipe"},
			{title: "Authenticate", description: "Initialize authentication with Spotify", command: "auth init"},
			{title: "Version", description: "Display version information", command: "version"},
			{title: "Quit", description: "Exit the application", command: "quit"},
		},
		cursor:      0,
		windowWidth: 80,
	}
}

// Init initializes the model
func (m MenuModel) Init() tea.Cmd {
	return nil
}

// Update updates the model
func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "enter":
			m.choice = m.items[m.cursor].command
			if m.choice == "quit" {
				m.quitting = true
			}
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
	}

	return m, nil
}

// View renders the model
func (m MenuModel) View() string {
	if m.quitting {
		return ""
	}

	// Get styles from the shared styles
	titleStyle := GetTitleStyle(m.windowWidth)
	selectedStyle := GetSelectedStyle()
	normalStyle := GetNormalStyle()
	descriptionStyle := GetInfoStyle()

	// Build the view
	s := titleStyle.Render("Spotify CLI") + "\n\n"

	for i, item := range m.items {
		cursor := " "
		style := normalStyle
		if i == m.cursor {
			cursor = ">"
			style = selectedStyle
		}

		s += fmt.Sprintf("%s %s\n", cursor, style.Render(item.title))
		if i == m.cursor {
			s += "  " + descriptionStyle.Render(item.description) + "\n"
		}
	}

	s += "\n" + normalStyle.Render("Press q to quit, up/down to navigate, enter to select")

	return s
}

// GetChoice returns the selected command
func (m MenuModel) GetChoice() string {
	return m.choice
}

// RunMainMenu runs the main menu UI and returns the selected command
func RunMainMenu() (string, error) {
	p := tea.NewProgram(NewMenuModel(), tea.WithAltScreen())
	model, err := p.Run()
	if err != nil {
		return "", err
	}

	menuModel, ok := model.(MenuModel)
	if !ok {
		return "", fmt.Errorf("could not cast model to MenuModel")
	}

	return menuModel.GetChoice(), nil
}
