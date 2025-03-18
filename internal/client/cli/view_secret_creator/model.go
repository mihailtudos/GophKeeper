package viewsecretcreator

import (
	"context"
	"fmt"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mihailtudos/gophkeeper/internal/client/cli/messages"
	"github.com/mihailtudos/gophkeeper/internal/client/dto"
	"strings"
)

const (
	SelectChoiceStateKey = "select"
	LoginSecretTypeKey   = "login"
	CardSecretTypeKey    = "card"
	TextSecretTypeKey    = "text"
	BinarySecretTypeKey  = "binary"
	NoChoiceMadeKey      = ""
)

var (
	headerStyle  = lipgloss.NewStyle().Background(lipgloss.Color("99")).Foreground(lipgloss.Color("#ffffff")).Padding(1)
	choices      = []string{LoginSecretTypeKey, CardSecretTypeKey, TextSecretTypeKey, BinarySecretTypeKey}
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	cursorStyle  = focusedStyle
	noStyle      = lipgloss.NewStyle()
)

type SecretsProvider interface {
	CreateSecret(ctx context.Context, message dto.SecretMessage) error
}

type Model struct {
	AppName         string
	Choice          string
	cursor          int
	Inputs          []textinput.Model
	SecretsProvider SecretsProvider
	cursorMode      cursor.Mode
	focusIndex      int
	ErrorMsg        string
	State           string
	result          string
}

func NewModel(sp SecretsProvider, AppName, ErrorMsg string) Model {
	return Model{
		SecretsProvider: sp,
		State:           SelectChoiceStateKey,
		AppName:         AppName,
		ErrorMsg:        ErrorMsg,
		Choice:          NoChoiceMadeKey,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle different key presses based on the current state
		if m.State == SelectChoiceStateKey {
			// Handle selection state
			switch msg.String() {
			case "enter":
				m.Choice = choices[m.cursor]
				m.State = m.Choice

				// Initialize inputs for the specific form type
				if m.Choice == LoginSecretTypeKey {
					m.initializeLoginInputs()
				}

				return m, nil
			case "down", "left", "j":
				m.cursor++
				if m.cursor >= len(choices) {
					m.cursor = 0
				}
			case "up", "right", "k":
				m.cursor--
				if m.cursor < 0 {
					m.cursor = len(choices) - 1
				}
			case "ctrl+b":
				return m, m.ActionMsg("back_to_home")
			}
		} else if m.State == LoginSecretTypeKey {
			// Handle login form state
			switch msg.String() {
			case "tab", "shift+tab", "up", "down":
				// Cycle through inputs
				if msg.String() == "up" || msg.String() == "shift+tab" {
					m.focusIndex--
					if m.focusIndex < 0 {
						m.focusIndex = len(m.Inputs) - 1
					}
				} else {
					m.focusIndex++
					if m.focusIndex >= len(m.Inputs) {
						m.focusIndex = 0
					}
				}

				// Set focus
				cmds := make([]tea.Cmd, len(m.Inputs))
				for i := 0; i < len(m.Inputs); i++ {
					if i == m.focusIndex {
						cmds[i] = m.Inputs[i].Focus()
						m.Inputs[i].PromptStyle = focusedStyle
						m.Inputs[i].TextStyle = focusedStyle
						continue
					}

					m.Inputs[i].Blur()
					m.Inputs[i].PromptStyle = noStyle
					m.Inputs[i].TextStyle = noStyle
				}

				return m, tea.Batch(cmds...)

			case "enter":
				// Submit the form
				if m.focusIndex == len(m.Inputs)-1 {
					// Process the form submission
					// For now, just print the values
					username := m.Inputs[0].Value()
					password := m.Inputs[1].Value()
					secretName := m.Inputs[2].Value()

					err := m.SecretsProvider.CreateSecret(context.Background(), dto.SecretMessage{
						Value: dto.LoginSecret{
							Username: username,
							Password: password,
						},
						SType: LoginSecretTypeKey,
						SName: secretName,
					})

					if err != nil {
						m.ErrorMsg = err.Error()
						return m, nil
					}

					// Here you would typically save these values
					//m.result = fmt.Sprintf("Saved login: %s, %s, %s", username, password, name)

					// Return to the choice screen or send a message to the parent model
					// For now, just return to the choice screen
					m.State = SelectChoiceStateKey
					m.Choice = NoChoiceMadeKey
					return m, nil
				}

				// Move to the next input
				m.focusIndex++
				if m.focusIndex >= len(m.Inputs) {
					m.focusIndex = 0
				}

				// Set focus
				cmds := make([]tea.Cmd, len(m.Inputs))
				for i := 0; i < len(m.Inputs); i++ {
					if i == m.focusIndex {
						cmds[i] = m.Inputs[i].Focus()
						m.Inputs[i].PromptStyle = focusedStyle
						m.Inputs[i].TextStyle = focusedStyle
						continue
					}

					m.Inputs[i].Blur()
					m.Inputs[i].PromptStyle = noStyle
					m.Inputs[i].TextStyle = noStyle
				}

				return m, tea.Batch(cmds...)

			case "ctrl+b":
				// Go back to the selection screen
				m.State = SelectChoiceStateKey
				m.Choice = NoChoiceMadeKey
				return m, nil
			}

			// Handle input for the focused textinput
			var cmd tea.Cmd
			m.Inputs[m.focusIndex], cmd = m.Inputs[m.focusIndex].Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m Model) View() string {
	s := strings.Builder{}
	switch m.State {
	case SelectChoiceStateKey:
		s.WriteString(m.renderSelectChoiceView())
	case LoginSecretTypeKey:
		s.WriteString(m.renderLoginSecretView())
	default:
		s.WriteString(m.renderSelectChoiceView())
	}

	s.WriteString("\nctrl+b - back | ctrl+c, esc - quit\n\n")
	if m.result != "" {
		s.WriteString(fmt.Sprintf("\n%s\n", m.result))
	}
	return s.String()
}

func (m Model) ActionMsg(action string) func() tea.Msg {
	return func() tea.Msg {
		return messages.ActionMsg{Value: action}
	}
}

func (m Model) renderSelectChoiceView() string {
	s := strings.Builder{}
	subtitle := "Select secret type:"

	fmt.Fprintf(&s, "%s\n\n%s\n\n", headerStyle.Render(m.AppName), subtitle)

	for i := 0; i < len(choices); i++ {
		if m.cursor == i {
			s.WriteString("--> ")
		} else {
			s.WriteString("    ")
		}
		s.WriteString(choices[i])
		s.WriteString("\n")
	}

	return s.String()
}
