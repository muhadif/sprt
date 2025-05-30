package tui

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AuthModel is the model for the authentication UI
type AuthModel struct {
	clientID     string
	clientSecret string
	authURL      string
	status       string
	step         int // 0: input clientID, 1: input clientSecret, 2: waiting for auth, 3: completed
	cursor       int
	input        string
	quitting     bool
	windowWidth  int
}

// NewAuthModel creates a new authentication model
func NewAuthModel() *AuthModel {
	return &AuthModel{
		step:        0,
		status:      "Please enter your Spotify Client ID",
		windowWidth: 80,
	}
}

func NewAuthModelWithStep(step int) *AuthModel {
	return &AuthModel{
		step:   step,
		status: "Waiting for authorization",
	}
}

// Init initializes the model
func (m *AuthModel) Init() tea.Cmd {
	return nil
}

// Update updates the model
func (m *AuthModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			if m.step == 0 {
				m.clientID = m.input
				m.input = ""
				m.step = 1
				m.status = "Please enter your Spotify Client Secret"
			} else if m.step == 1 {
				m.clientSecret = m.input
				m.input = ""
				m.step = 2
				m.status = "Initializing authentication..."
				return m, tea.Quit
			} else if m.step == 2 {
				m.step = 3
				m.status = "Authentication completed"
				return m, tea.Quit
			}
		case "ctrl+y", "cmd+y":
			// Handle copy operation for the auth URL
			if m.step == 2 && m.authURL != "" {
				err := clipboard.WriteAll(m.authURL)
				if err == nil {
					m.status = "URL copied to clipboard!"
				}
			}
		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		case "ctrl+v", "cmd+v":
			// Handle paste operation
			text, err := clipboard.ReadAll()
			if err == nil {
				m.input += text
			}
		default:
			// Only add printable characters
			if len(msg.String()) == 1 {
				m.input += msg.String()
			}
		}
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
	}

	return m, nil
}

// View renders the model
func (m *AuthModel) View() string {
	if m.quitting {
		return ""
	}

	// Get styles from the shared styles
	titleStyle := GetTitleStyle(m.windowWidth)
	promptStyle := GetHeaderStyle()
	inputStyle := GetInputStyle()
	infoStyle := GetInfoStyle()
	urlStyle := GetLinkStyle()
	border := GetBorderStyle(m.windowWidth)

	// Build the view
	s := titleStyle.Render("Spotify Authentication") + "\n\n"

	content := ""

	if m.step == 0 || m.step == 1 {
		// Input prompt
		content += promptStyle.Render(m.status) + "\n\n"

		// Input field
		displayInput := m.input
		if m.step == 1 {
			// Mask the client secret
			displayInput = strings.Repeat("*", len(m.input))
		}
		content += inputStyle.Render(displayInput) + "\n\n"

		// Instructions
		content += infoStyle.Render("Press Enter to continue, Esc to cancel, Ctrl+V/Cmd+V to paste")
	} else if m.step == 2 {
		// Waiting for authorization
		content += promptStyle.Render("Please open the following URL in your browser:") + "\n\n"
		content += urlStyle.Render(m.authURL) + "\n\n"

		// Show status message if URL was copied
		if m.status == "URL copied to clipboard!" {
			content += GetHeaderStyle().Foreground(lipgloss.Color("#00FF00")).Render(m.status) + "\n\n"
		}

		content += infoStyle.Render("After authorizing, you will be redirected to a local callback URL.") + "\n"
		content += infoStyle.Render("Press Enter after you have completed the authorization process...") + "\n"
		content += infoStyle.Render("Press Ctrl+Y (or Cmd+Y on Mac) to copy the URL to clipboard")
	} else if m.step == 3 {
		// Completed
		content += promptStyle.Render("Authentication completed successfully!") + "\n\n"
		content += infoStyle.Render("You can now use the sprt.")
	}

	s += border.Render(content)

	return s
}

// GetCredentials returns the client ID and client secret
func (m *AuthModel) GetCredentials() (string, string) {
	return m.clientID, m.clientSecret
}

// SetAuthURL sets the authorization URL
func (m *AuthModel) SetAuthURL(url string) {
	m.authURL = url
}

// RunAuthUI runs the authentication UI for input
func RunAuthUI() (string, string, error) {
	p := tea.NewProgram(NewAuthModel(), tea.WithAltScreen())
	model, err := p.Run()
	if err != nil {
		return "", "", err
	}

	authModel, ok := model.(*AuthModel)
	if !ok {
		return "", "", fmt.Errorf("could not cast model to AuthModel")
	}

	clientID, clientSecret := authModel.GetCredentials()
	return clientID, clientSecret, nil
}

// RunAuthWaitingUI runs the authentication UI for waiting for authorization
func RunAuthWaitingUI(authURL string, clientID string, clientSecret string) error {
	model := NewAuthModelWithStep(2)
	model.clientID = clientID
	model.clientSecret = clientSecret
	model.authURL = authURL
	model.status = "Waiting for authorization"

	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
