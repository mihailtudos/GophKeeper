package viewsecretcreator

import (
	"fmt"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	"strings"
)

func (m Model) renderLoginSecretView() string {
	s := strings.Builder{}
	subtitle := "Create new login secret:"

	fmt.Fprintf(&s, "%s\n\n%s\n\n", headerStyle.Render(m.AppName), subtitle)

	if len(m.Inputs) == 0 {
		s.WriteString("Loading form...\n")
		return s.String()
	}

	for i, input := range m.Inputs {
		fieldName := ""
		switch i {
		case 0:
			fieldName = "Username:"
		case 1:
			fieldName = "Password:"
		case 2:
			fieldName = "Name:"
		}

		if i == m.focusIndex {
			s.WriteString(focusedStyle.Render(fieldName) + " ")
		} else {
			s.WriteString(blurredStyle.Render(fieldName) + " ")
		}

		if m.ErrorMsg != "" {
			s.WriteString(errorStyle.Render(m.ErrorMsg) + "\n")
		}

		s.WriteString(input.View())
		s.WriteString("\n\n")
	}

	s.WriteString("\nPress tab to cycle through fields, enter to submit\n")
	return s.String()
}

// initializeLoginInputs sets up the text inputs for the login form
func (m *Model) initializeLoginInputs() {
	// Create inputs for username, password, and website
	var inputs []textinput.Model = make([]textinput.Model, 3)

	// Set up website input
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "(e.g. www.example.com)"
	inputs[0].CharLimit = 128
	inputs[0].Width = 40

	// Set up username input
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Username"
	inputs[1].Focus()
	inputs[1].CharLimit = 64
	inputs[1].Width = 40

	// Set up password input
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Password"
	inputs[2].CharLimit = 64
	inputs[2].Width = 40
	inputs[2].EchoMode = textinput.EchoPassword
	inputs[2].EchoMode = textinput.EchoPassword

	m.Inputs = inputs
	m.focusIndex = 0
	m.cursorMode = cursor.CursorBlink
}
