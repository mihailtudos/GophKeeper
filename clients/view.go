package main

import (
	"github.com/charmbracelet/lipgloss"
)

const (
	appName = "GopherKeep"
)

var (
	headerStyle     = lipgloss.NewStyle().Background(lipgloss.Color("99")).Foreground(lipgloss.Color("#ffffff")).Padding(1)
	enumeratorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).MarginRight(1)
)

func (m model) View() string {
	s := headerStyle.Render(appName) + "\n\n"

	if m.state == loginView {
		s += "Login to GophKeeper\n\n"
		s += "Email:\n"
		s += m.loginInputMsg.emailInput.View() + "\n\n"
		s += "Password:\n"
		s += m.loginInputMsg.passwordInput.View()
		if m.formError != "" {
			s += "\n❌ " + m.formError
		}

		s += "\n\ntab - change input, enter - submit\nctrl + r - register, esc - quit"
	}

	if m.state == registerView {
		s += "Register to GophKeeper\n\n"
		
		s += "Full name"
		s += m.registerInputMsg.fullNameInput.View() + "\n\n"
		s += "Email:\n"
		s += m.registerInputMsg.emailInput.View() + "\n\n"
		s += "Password:\n"
		s += m.registerInputMsg.passwordInput.View() + "\n\n"
		s += "Confirm password:\n"
		s += m.registerInputMsg.confirmPasswordInput.View()

		if m.formError != "" {
			s += "\n❌ " + m.formError
		}

		s += "enter - submit input, ctrl + l - login, esc - quit"
	}

	if m.state == homeView {
		for i, record := range m.records {
			prefix := " "
			if i == m.selectedRecordIndex {
				prefix = ">"
			}

			s += enumeratorStyle.Render(prefix) + " " + record.Title + "\n\n"
		}

		s += "n - new record, esc - quit"
	}

	return s
}
