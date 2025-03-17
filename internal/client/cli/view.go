package cli

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
)

const (
	appName = "GopherKeep"
)

var (
	headerStyle     = lipgloss.NewStyle().Background(lipgloss.Color("99")).Foreground(lipgloss.Color("#ffffff")).Padding(1)
	enumeratorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).MarginRight(1)
)

func (m Model) View() string {
	var s string
	s = headerStyle.Render(appName) + "\n\n"

	if m.state == loginView {
		s += "Login to GophKeeper\n\n"
		s += "Email:\n"
		s += m.loginInputMsg.emailInput.View() + "\n\n"
		s += "Password:\n"
		s += m.loginInputMsg.passwordInput.View()
		if m.formError != "" {
			s += "\n❌ " + m.formError
		}

		s += "\n\ntab - change input, enter - submit\nctrl + r - register, esc, ctrl + q - quit"
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
		s += "Welcome to GophKeeper\n\n"

		if m.isLoading {
			return fmt.Sprintf("%s\n\nPress Ctrl+Q to quit", m.loadingMessage)
		} else {
			s += "Records:\n"
			for i, record := range m.records {
				prefix := " "
				if i == m.selectedRecordIndex {
					prefix = ">"
				}

				s += enumeratorStyle.Render(prefix) + " " + record.SName + "\n\n"
			}
		}

		s += "\n\nctrl + r - reset\nesc,ctrl + q - quit, ctrl + l - logout\n"
	}

	if m.state == homeMasterPasswordView {
		s += "Welcome to GophKeeper\n\n"

		if mp, err := getAuthCreds(serviceMasterPasswordKey); err != nil || mp == "" {
			s += "\nEnter the backup key to fetch your secrets:\n"
			s += m.masterPasswordMsg.masterPasswordInput.View() + "\n\n"

			if m.formError != "" {
				s += "❌ " + m.formError + "\n\n"
			}

			s += "enter - submit\n"
		} else {
			s += "You already have a backup key set.\n"
			s += "\nctrl + h = return to home, ctrl + r - reset backup key\n"
		}

		s += "esc,ctrl + q - quit, ctrl + l - logout"
	}

	return s
}
