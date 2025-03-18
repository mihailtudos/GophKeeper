package viewregister

import (
	"fmt"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	headerStyle  = lipgloss.NewStyle().Background(lipgloss.Color("99")).Foreground(lipgloss.Color("#ffffff")).Padding(1)
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle  = focusedStyle
	noStyle      = lipgloss.NewStyle()

	focusedButton = focusedStyle.Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

type Model struct {
	State      string
	result     string
	focusIndex int
	Inputs     []textinput.Model
	cursorMode cursor.Mode
	AppName    string
	ErrorMsg   string
}

func NewModel(AppName, ErrorMsg string) Model {
	m := Model{
		AppName:  AppName,
		ErrorMsg: ErrorMsg,
		State:    "register",
		Inputs:   make([]textinput.Model, 3),
	}

	var t textinput.Model
	for i := range m.Inputs {
		t = textinput.New()
		t.Cursor.Style = cursorStyle
		t.CharLimit = 32

		switch i {
		case 0:
			t.Placeholder = "Email"
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
			t.CharLimit = 64
		case 1:
			t.Placeholder = "Password"
			t.EchoMode = textinput.EchoPassword
			t.EchoMode = textinput.EchoPassword
		case 2:
			t.Placeholder = "Password"
			t.EchoMode = textinput.EchoPassword
			t.EchoMode = textinput.EchoPassword
		}

		m.Inputs[i] = t
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// Change cursor mode
		case "ctrl+r":
			m.cursorMode++
			if m.cursorMode > cursor.CursorHide {
				m.cursorMode = cursor.CursorBlink
			}
			cmds := make([]tea.Cmd, len(m.Inputs))
			for i := range m.Inputs {
				cmds[i] = m.Inputs[i].Cursor.SetMode(m.cursorMode)
			}
			return m, tea.Batch(cmds...)

		// Set focus to next input
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			// Did the user press enter while the submit button was focused?
			// If so, exit.
			if s == "enter" && m.focusIndex == len(m.Inputs) {
				//!!!  здесь начинается процесс регистрации
				if m.Inputs[1].Value() != m.Inputs[2].Value() {
					m.State = "again"
					m.result = "invalid password, try again"
					m.focusIndex = -1
					return m, nil
				}
				//err := m.grpcClient.Register(context.Background(), m.Inputs[0].Value(), m.Inputs[1].Value())
				//if err != nil {
				//	if err.Error() == fmt.Sprintf("user with email %s already registered", m.Inputs[0].Value()) {
				//		m.State = "completed"
				//	} else {
				//		m.State = "error"
				//	}
				//	m.result = err.Error()
				//	m.focusIndex = -1
				//	return m, nil
				//}
				m.State = "completed"
				m.result = "registration completed successfully"
				m.focusIndex = -1
				return m, nil
			}
			if s == "enter" && m.focusIndex == -1 {
				return m, tea.Quit
			}
			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.Inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.Inputs)
			}

			cmds := make([]tea.Cmd, len(m.Inputs))
			for i := 0; i <= len(m.Inputs)-1; i++ {
				if i == m.focusIndex {
					// Set focused state
					cmds[i] = m.Inputs[i].Focus()
					m.Inputs[i].PromptStyle = focusedStyle
					m.Inputs[i].TextStyle = focusedStyle
					continue
				}
				// Remove focused state
				m.Inputs[i].Blur()
				m.Inputs[i].PromptStyle = noStyle
				m.Inputs[i].TextStyle = noStyle
			}

			return m, tea.Batch(cmds...)
		}
	}

	// Handle character input and blinking
	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m Model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.Inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.Inputs {
		m.Inputs[i], cmds[i] = m.Inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m Model) View() string {
	var s strings.Builder
	fmt.Fprintf(&s, "%s\n\n%s\n\n", headerStyle.Render(m.AppName), "Register:")

	for i := range m.Inputs {
		s.WriteString(m.Inputs[i].View())
		if i < len(m.Inputs)-1 {
			s.WriteRune('\n')
		}
	}

	button := &blurredButton
	if m.focusIndex == len(m.Inputs) {
		button = &focusedButton
	}

	fmt.Fprintf(&s, "\n\n%s\n", *button)

	if m.State != "" {
		s.WriteString(fmt.Sprintf("\n%s\n", m.result))
	}

	s.WriteString("\nctrl+c, esc - quit\n\n")
	return s.String()
}
