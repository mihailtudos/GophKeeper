package main

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	loginView uint = iota
	registerView
	homeView
	createRecordView
	viewRecordView
)

type LoginForm struct {
	Email    string
	Password string
}

type RegisterForm struct {
	FullName        string
	Email           string
	Password        string
	ConfirmPassword string
}

type loginInputMsg struct {
	emailInput    textinput.Model
	passwordInput textinput.Model
}

type registerInputMsg struct {
	fullNameInput        textinput.Model
	emailInput           textinput.Model
	passwordInput        textinput.Model
	confirmPasswordInput textinput.Model
}

type model struct {
	state               uint
	store               *Store
	records             []Record
	currentRecord       Record
	selectedRecordIndex int
	textinput           textinput.Model
	textarea            textarea.Model
	loginForm           LoginForm
	loginInputMsg       loginInputMsg

	registerForm     RegisterForm
	registerInputMsg registerInputMsg

	formError string
}

func NewModel(store *Store) model {
	records, err := store.GetRecords()
	if err != nil {
		panic(err)
	}

	return model{
		state:         loginView,
		store:         store,
		records:       records,
		textinput:     textinput.New(),
		textarea:      textarea.New(),
		loginForm:     LoginForm{},
		registerForm:  RegisterForm{},
		loginInputMsg: setupLoginView(),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch k := msg.(type) {
	case tea.KeyMsg:
		key := k.String()
		switch m.state {
		case loginView:
			return handleLoginView(msg, k.String(), &m)
		case registerView:
			return handleRegisterView(msg, k.String(), &m)
		case homeView:
			switch key {
			case "esc":
				return m, tea.Quit
			case "up":
				if m.selectedRecordIndex > 0 {
					m.selectedRecordIndex--
				}
			case "down":
				if m.selectedRecordIndex < len(m.records)-1 {
					m.selectedRecordIndex--
				}
			case "enter":
				m.currentRecord = m.records[m.selectedRecordIndex]
				m.state = viewRecordView
			}
		}
	}

	return m, nil
}

func handleLoginView(msg tea.Msg, key string, m *model) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	var cmd tea.Cmd
	m.loginInputMsg.emailInput, cmd = m.loginInputMsg.emailInput.Update(msg)
	cmds = append(cmds, cmd)

	m.loginInputMsg.passwordInput, cmd = m.loginInputMsg.passwordInput.Update(msg)
	cmds = append(cmds, cmd)
	m.formError = ""

	switch key {
	case "esc":
		return m, tea.Quit
	case "down", "up", "tab":
		m.loginForm.Email = m.loginInputMsg.emailInput.Value()
		m.loginForm.Password = m.loginInputMsg.passwordInput.Value()

		if m.loginInputMsg.passwordInput.Focused() {
			m.loginInputMsg.passwordInput.Blur()
			m.loginInputMsg.emailInput.Focus()
		} else {
			m.loginInputMsg.passwordInput.Focus()
			m.loginInputMsg.emailInput.Blur()
		}
	case "enter":
		// run validation and show error
		m.loginForm.Email = m.loginInputMsg.emailInput.Value()
		m.loginForm.Password = m.loginInputMsg.passwordInput.Value()

		// Validate inputs
		if m.loginForm.Email == "" || m.loginForm.Password == "" {
			m.formError = "Email and password are required"
			return m, tea.Batch(cmds...)
		}
		// Try to login
		err := m.store.Login(m.loginForm)
		if err != nil {
			m.formError = err.Error()
			return m, tea.Batch(cmds...)
		}

		// Clear error on successful login
		m.formError = ""
		m.state = homeView
	case "ctrl+r":
		renderRegisterView(m)
	}

	return m, tea.Batch(cmds...)
}

func handleRegisterView(msg tea.Msg, key string, m *model) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	m.registerInputMsg.fullNameInput, cmd = m.registerInputMsg.fullNameInput.Update(msg)
	cmds = append(cmds, cmd)

	m.registerInputMsg.emailInput, cmd = m.registerInputMsg.emailInput.Update(msg)
	cmds = append(cmds, cmd)

	m.registerInputMsg.passwordInput, cmd = m.registerInputMsg.passwordInput.Update(msg)
	cmds = append(cmds, cmd)

	m.registerInputMsg.confirmPasswordInput, cmd = m.registerInputMsg.confirmPasswordInput.Update(msg)
	cmds = append(cmds, cmd)

	switch key {
	case "esc":
		return m, tea.Quit
	case "ctrl+l":
		renderLoginView(m)
	case "enter":

	}

	return m, tea.Batch(cmds...)
}

func setupLoginView() loginInputMsg {
	emailInput := textinput.New()
	emailInput.Placeholder = "Enter email"
	emailInput.Focus()
	emailInput.CharLimit = 50
	emailInput.Width = 30
	passwordInput := textinput.New()
	passwordInput.Placeholder = "Enter password"
	passwordInput.EchoMode = textinput.EchoPassword
	passwordInput.CharLimit = 50
	passwordInput.Width = 30

	return loginInputMsg{
		emailInput:    emailInput,
		passwordInput: passwordInput,
	}
}

func setupRegisterView() registerInputMsg {
	fullNameInput := textinput.New()
	fullNameInput.Placeholder = "Enter full name"
	fullNameInput.Focus()
	fullNameInput.CharLimit = 50
	fullNameInput.Width = 30

	emailInput := textinput.New()
	emailInput.Placeholder = "Enter email"
	emailInput.Focus()
	emailInput.CharLimit = 50
	emailInput.Width = 30

	passwordInput := textinput.New()
	passwordInput.Placeholder = "Enter password"
	passwordInput.EchoMode = textinput.EchoPassword
	passwordInput.CharLimit = 50
	passwordInput.Width = 30

	confirmPasswordInput := textinput.New()
	confirmPasswordInput.Placeholder = "Enter confirm password"
	confirmPasswordInput.EchoMode = textinput.EchoPassword
	confirmPasswordInput.CharLimit = 50
	confirmPasswordInput.Width = 30

	return registerInputMsg{
		fullNameInput:        fullNameInput,
		emailInput:           emailInput,
		passwordInput:        passwordInput,
		confirmPasswordInput: confirmPasswordInput,
	}
}

func isHotKey(key string) bool {
	return key == "tab" || key == "enter" || key == "up" || key == "down"
}

func renderRegisterView(m *model) {
	m.registerForm = RegisterForm{}
	m.registerInputMsg = setupRegisterView()
	m.state = registerView
}

func renderLoginView(m *model) {
	m.loginForm = LoginForm{}
	m.loginInputMsg = setupLoginView()
	m.state = loginView
}
