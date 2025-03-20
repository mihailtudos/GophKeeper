package viewsecretcreator

import (
	"context"
	"fmt"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mihailtudos/gophkeeper/internal/client/dto"
	"strings"
)

// renderCardSecretView renders the card secret creation form.
func (m Model) renderCardSecretView() string {
	s := strings.Builder{}
	subtitle := "Create new card secret:"

	fmt.Fprintf(&s, "%s\n\n%s\n\n", headerStyle.Render(m.AppName), subtitle)

	if len(m.Inputs) == 0 {
		s.WriteString("Loading form...\n")
		return s.String()
	}

	for i, input := range m.Inputs {
		fieldName := ""
		switch i {
		case 0:
			fieldName = "Cardholder name:"
		case 1:
			fieldName = "Card number:"
		case 2:
			fieldName = "Expiration date:"
		case 3:
			fieldName = "CVV:"
		case 4:
			fieldName = "Secret name:"
		}

		if i == m.focusIndex {
			s.WriteString(focusedStyle.Render(fieldName) + " ")
		} else {
			s.WriteString(blurredStyle.Render(fieldName) + " ")
		}

		s.WriteString(input.View())
		s.WriteString("\n\n")
	}

	if m.ErrorMsg != "" {
		s.WriteString(errorStyle.Render(m.ErrorMsg) + "\n")
	}

	s.WriteString("\nPress tab to cycle through fields, enter to submit\n")
	return s.String()
}

// initializeCardInputs sets up the text inputs for the card form
func (m *Model) initializeCardInputs() {
	// Create inputs for cardholder name, Card number, Expiration date, CVV, and secret name
	var inputs []textinput.Model = make([]textinput.Model, 3)

	// Set up website input
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Cardholder name"
	inputs[0].CharLimit = 128
	inputs[0].Width = 40
	inputs[0].Focus()

	// Set up username input
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Card number"
	inputs[1].CharLimit = 16
	inputs[1].Width = 40

	// Set up password input
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Expiration date"
	inputs[2].CharLimit = 64
	inputs[2].Width = 40
	inputs[2].EchoMode = textinput.EchoPassword
	inputs[2].EchoMode = textinput.EchoPassword

	m.Inputs = inputs
	m.focusIndex = 0
	m.cursorMode = cursor.CursorBlink
}

func handleCardSecretTypeKey(m *Model, msg tea.KeyMsg) tea.Cmd {
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

		return tea.Batch(cmds...)

	case "enter":
		// Submit the form
		if m.focusIndex == len(m.Inputs)-1 {
			// Process the form submission
			// For now, just print the values
			username := m.Inputs[0].Value()
			password := m.Inputs[1].Value()
			secretName := m.Inputs[2].Value()

			err := m.SecretsProvider.CreateSecret(context.Background(), dto.SecretMessage{
				Data: dto.LoginSecret{
					Username: username,
					Password: password,
				},
				Type: LoginSecretTypeKey,
				Name: secretName,
			})

			if err != nil {
				m.ErrorMsg = err.Error()
				return nil
			}

			// Here you would typically save these values
			//m.result = fmt.Sprintf("Saved login: %s, %s, %s", username, password, name)

			// Return to the choice screen or send a message to the parent model
			// For now, just return to the choice screen
			m.State = SelectChoiceStateKey
			m.Choice = NoChoiceMadeKey
			return nil
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

		return tea.Batch(cmds...)

	case "ctrl+b":
		// Go back to the selection screen
		m.State = SelectChoiceStateKey
		m.Choice = NoChoiceMadeKey
		return nil
	}

	return nil
}
