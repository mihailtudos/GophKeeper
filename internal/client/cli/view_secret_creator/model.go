package viewsecretcreator

import (
	"fmt"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/mihailtudos/gophkeeper/internal/client/cli/messages"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	SelectChoiceStateKey = "select"
)

var (
	headerStyle  = lipgloss.NewStyle().Background(lipgloss.Color("99")).Foreground(lipgloss.Color("#ffffff")).Padding(1)
	choices      = []string{"Login", "Card", "Text", "Binary"}
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle  = focusedStyle
	noStyle      = lipgloss.NewStyle()
)

type Model struct {
	AppName    string
	Choice     string
	cursor     int
	Inputs     []textinput.Model
	cursorMode cursor.Mode
	focusIndex int
	ErrorMsg   string
	State      string
	result     string
}

func NewModel(AppName, ErrorMsg string) Model {
	return Model{
		State:    SelectChoiceStateKey,
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
	subtitle := "Create new secrets:"
	if m.State == SelectChoiceStateKey {
		subtitle += "Select secret type:"
	}

	fmt.Fprintf(&s, "%s\n\n%s\n\n", headerStyle.Render(m.AppName), subtitle)

	buttons := make([]string, len(choices))
	for i := 0; i < len(choices); i++ {
		if m.cursor == i {
			buttons[i] = fmt.Sprintf("[ %s ]", focusedStyle.Render(choices[i]))
		} else {
			buttons[i] = fmt.Sprintf("[ %s ]", blurredStyle.Render(choices[i]))
		}
	}

	fmt.Fprintf(&s, "\n\n%s\t%s\n\n", buttons[0], buttons[1])

	s.WriteString("\n(press ctrl+c or esc quit)\n\n")

	return s.String()
}

func (m Model) ActionMsg(action string) func() tea.Msg {
	return func() tea.Msg {
		return messages.ActionMsg{Value: action}
	}
}

// Fake authentication function (simulating API call)
func fakeAuth(username, password string) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(time.Second)
		return messages.LoginSuccessMsg{Token: "fake-jwt-token"}
	}
}
