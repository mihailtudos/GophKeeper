package viewhome

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mihailtudos/gophkeeper/internal/client/cli/messages"
	"strings"
)

var (
	headerStyle  = lipgloss.NewStyle().Background(lipgloss.Color("99")).Foreground(lipgloss.Color("#ffffff")).Padding(1)
	choices      = []string{"Login", "Register"}
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle  = focusedStyle
	noStyle      = lipgloss.NewStyle()
)

type Model struct {
	AppName  string
	cursor   int
	Choice   string
	ErrorMsg string
}

func NewModel(AppName, ErrorMsg string) Model {
	return Model{
		AppName:  AppName,
		ErrorMsg: ErrorMsg,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.Choice = choices[m.cursor]

			return m, m.ActionMsg(m.Choice)
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
		}

	}
	return m, nil
}

func (m Model) View() string {
	s := strings.Builder{}
	s.WriteString(headerStyle.Render(m.AppName) + "\n\n")

	s.WriteString("Your secrets:\n")

	//buttons := make([]string, len(choices))
	//for i := 0; i < len(choices); i++ {
	//	if m.cursor == i {
	//		buttons[i] = fmt.Sprintf("[ %s ]", focusedStyle.Render(choices[i]))
	//	} else {
	//		buttons[i] = fmt.Sprintf("[ %s ]", blurredStyle.Render(choices[i]))
	//	}
	//}

	//fmt.Fprintf(&s, "\n\n%s\t%s\n\n", buttons[0], buttons[1])

	s.WriteString("\n(press ctrl+c or esc quit)\n\n")

	return s.String()
}

func (m Model) ActionMsg(action string) func() tea.Msg {
	return func() tea.Msg {
		return messages.ActionMsg{Value: action}
	}
}
