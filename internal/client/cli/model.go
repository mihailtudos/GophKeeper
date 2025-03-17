package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mihailtudos/gophkeeper/internal/client/config"
	store2 "github.com/mihailtudos/gophkeeper/internal/client/infrastructure/store"
	"log/slog"
	"unicode/utf8"
)

const (
	loginView uint = iota
	registerView
	homeMasterPasswordView
	homeView
	createRecordView
	viewRecordView
)

type loginForm struct {
	Email    string `json:"username"`
	Password string `json:"password"`
}

type registerForm struct {
	FullName        string
	Email           string
	Password        string
	ConfirmPassword string
}

type masterPasswordForm struct {
	MasterPassword string
}

type loginInputMsg struct {
	emailInput    textinput.Model
	passwordInput textinput.Model
}

type masterPasswordMsg struct {
	masterPasswordInput textinput.Model
}

type registerInputMsg struct {
	fullNameInput        textinput.Model
	emailInput           textinput.Model
	passwordInput        textinput.Model
	confirmPasswordInput textinput.Model
}

type authTokens struct {
	accessToken  string
	refreshToken string
}
type Model struct {
	state               uint
	store               *store2.Store
	records             []store2.Record
	currentRecord       store2.Record
	selectedRecordIndex int
	textInput           textinput.Model
	textarea            textarea.Model
	loginForm           loginForm
	loginInputMsg       loginInputMsg
	isLoading           bool
	loadingMessage      string
	registerForm        registerForm
	registerInputMsg    registerInputMsg

	masterPasswordForm masterPasswordForm
	masterPasswordMsg  masterPasswordMsg

	formError string

	authTokens authTokens
}

func NewModel(cfg config.Config) Model {
	var viewState uint = homeView

	if err := setNewAccessToken(); err != nil {
		logger.Debug("failed to set new access token redirecting to login")
		viewState = loginView
	}

	mp, err := getAuthCreds(serviceMasterPasswordKey)
	if err != nil {
		mp = ""
	}

	if viewState == homeView && mp == "" {
		viewState = homeMasterPasswordView
	}

	records, err := GetRecords()
	if err != nil {
		panic(err)
	}

	return Model{
		state:             viewState,
		isLoading:         false,
		loadingMessage:    "",
		records:           records,
		textInput:         textinput.New(),
		textarea:          textarea.New(),
		loginForm:         loginForm{},
		registerForm:      registerForm{},
		loginInputMsg:     setupLoginView(),
		masterPasswordMsg: setupHomeMasterPasswordView(),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch k := msg.(type) {
	case fetchRecordsSuccessMsg:
		m.records = k.records
		m.isLoading = false
		return m, nil
	case fetchRecordsErrMsg:
		m.isLoading = false
		m.formError = fmt.Sprintf("Failed to fetch records: %v", k.err)
		return m, nil
	case tea.KeyMsg:
		switch m.state {
		case loginView:
			return handleLoginView(msg, k.String(), &m)
		case registerView:
			return handleRegisterView(msg, k.String(), &m)
		case homeView:
			return handleHomeView(msg, k.String(), &m)
		case homeMasterPasswordView:
			return handleHomeMasterPasswordView(msg, k.String(), &m)
		default:
			panic("unhandled default case")
		}
	}

	return m, nil
}

func handleHomeMasterPasswordView(msg tea.Msg, key string, m *Model) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	var cmd tea.Cmd
	m.masterPasswordMsg.masterPasswordInput, cmd = m.masterPasswordMsg.masterPasswordInput.Update(msg)
	cmds = append(cmds, cmd)
	m.formError = ""

	switch key {
	case "esc", "ctrl+q":
		return m, tea.Quit
	case "ctrl+h":
		m.state = homeView
	case "ctrl+r":
		_ = storeAuthCreds(serviceMasterPasswordKey, "")
		m.masterPasswordMsg.masterPasswordInput.SetValue("")
		m.state = homeMasterPasswordView
	case "enter":
		logger.Debug("master password entered")
		m.masterPasswordForm.MasterPassword = m.masterPasswordMsg.masterPasswordInput.Value()

		if utf8.RuneCountInString(m.masterPasswordForm.MasterPassword) < 8 {
			m.formError = "Master password must be at least 8 characters long"
			return m, tea.Batch(cmds...)
		}

		if err := storeAuthCreds(serviceMasterPasswordKey, m.masterPasswordForm.MasterPassword); err != nil {
			m.formError = err.Error()
			return m, tea.Batch(cmds...)
		}

		accessToken, err := getAuthCreds(serviceAccessTokenKey)
		if err != nil {
			logger.Error("failed to get access token", slog.String("error", err.Error()))
			m.formError = "Failed to get access token"
			return m, tea.Batch(cmds...)
		}

		m.state = homeView
		m.isLoading = true
		m.loadingMessage = "Fetching your secrets..."

		return m, fetchRecordsCmd(m.store, accessToken, m.masterPasswordForm.MasterPassword)
	case "ctrl+l":
		_ = storeAuthCreds(serviceAccessTokenKey, "")
		_ = storeAuthCreds(serviceRefreshTokenKey, "")
		m.state = loginView
	}

	return m, tea.Batch(cmds...)
}

func handleHomeView(msg tea.Msg, key string, m *Model) (tea.Model, tea.Cmd) {
	switch key {
	case "esc", "ctrl+q":
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
		logger.Debug("selected record index", slog.String("index", m.masterPasswordMsg.masterPasswordInput.Value()))
	case "ctrl+l":
		_ = storeAuthCreds(serviceAccessTokenKey, "")
		_ = storeAuthCreds(serviceRefreshTokenKey, "")
		m.state = loginView
	case "ctrl+r":
		m.state = homeMasterPasswordView
	}
	return m, tea.Batch()
}

func handleLoginView(msg tea.Msg, key string, m *Model) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	var cmd tea.Cmd
	m.loginInputMsg.emailInput, cmd = m.loginInputMsg.emailInput.Update(msg)
	cmds = append(cmds, cmd)

	m.loginInputMsg.passwordInput, cmd = m.loginInputMsg.passwordInput.Update(msg)
	cmds = append(cmds, cmd)
	m.formError = ""

	switch key {
	case "esc", "ctrl+q":
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

		// Try to log in
		loginTokens, err := handleLogin(m.loginForm)
		if err != nil {
			m.formError = err.Error()
			return m, tea.Batch(cmds...)
		}

		if err = storeAuthCreds(serviceAccessTokenKey, loginTokens.AccessToken); err != nil {
			m.formError = err.Error()
			return m, tea.Batch(cmds...)
		}

		if err = storeAuthCreds(serviceRefreshTokenKey, loginTokens.RefreshToken); err != nil {
			m.formError = err.Error()
			return m, tea.Batch(cmds...)
		}

		logger.Debug("logged in successfully",
			slog.String("access_token", loginTokens.AccessToken),
			slog.String("refresh_token", loginTokens.RefreshToken))

		// Clear error on successful login
		m.formError = ""
		m.state = homeView
	case "ctrl+r":
		renderRegisterView(m)
	}

	return m, tea.Batch(cmds...)
}

func handleRegisterView(msg tea.Msg, key string, m *Model) (tea.Model, tea.Cmd) {
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
	case "esc", "ctrl+q":
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

func setupHomeMasterPasswordView() masterPasswordMsg {
	masterPasswordInput := textinput.New()
	masterPasswordInput.Placeholder = "Enter master password"
	masterPasswordInput.EchoMode = textinput.EchoPassword
	masterPasswordInput.CharLimit = 50
	masterPasswordInput.Width = 30
	masterPasswordInput.Focus()

	return masterPasswordMsg{
		masterPasswordInput: masterPasswordInput,
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

func renderRegisterView(m *Model) {
	m.registerForm = registerForm{}
	m.registerInputMsg = setupRegisterView()
	m.state = registerView
}

func renderLoginView(m *Model) {
	m.loginForm = loginForm{}
	m.loginInputMsg = setupLoginView()
	m.state = loginView
}

type fetchRecordsSuccessMsg struct {
	records []store2.Record
}

type fetchRecordsErrMsg struct {
	err error
}

func fetchRecordsCmd(store *store2.Store, accessToken, masterPassword string) tea.Cmd {
	return func() tea.Msg {
		secrets, err := fetchSecretsFromServer(accessToken, masterPassword)
		if err != nil {
			logger.Error("Failed to fetch secrets: %v", err)
			return fetchRecordsErrMsg{err}
		}

		if err = store.UploadSecrets(secrets); err != nil {
			return fetchRecordsErrMsg{err: err}
		}

		records, err := store.GetRecords()
		if err != nil {
			return fetchRecordsErrMsg{err: err}
		}
		return fetchRecordsSuccessMsg{records: records}
	}
}
