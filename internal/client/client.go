package client

import (
	"context"
	"errors"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mihailtudos/gophkeeper/internal/client/cli/messages"
	viewauth "github.com/mihailtudos/gophkeeper/internal/client/cli/view_auth"
	viewbackup "github.com/mihailtudos/gophkeeper/internal/client/cli/view_backup"
	viewhome "github.com/mihailtudos/gophkeeper/internal/client/cli/view_home"
	viewlogin "github.com/mihailtudos/gophkeeper/internal/client/cli/view_login"
	viewregister "github.com/mihailtudos/gophkeeper/internal/client/cli/view_register"
	"github.com/mihailtudos/gophkeeper/internal/client/config"
	"github.com/mihailtudos/gophkeeper/internal/client/dto"
	"github.com/mihailtudos/gophkeeper/pkg/keyring"
	"log/slog"
	"os"
	"syscall"
)

type Auth interface {
	Login(ctx context.Context, username, password string) (*dto.LoginResponse, error)
	StoreTokens(ctx context.Context, response *dto.LoginResponse) error
	GetAccessToken(ctx context.Context) (string, error)
	StoreBackupKey(ctx context.Context, key string) error
}

type Services interface {
	Auth
}

// ScreenType defines the different screens
type ScreenType int

const (
	AppName               = "GopherKeep"
	AuthScreen ScreenType = iota
	LoginScreen
	RegisterScreen
	BackupScreen
	HomeScreen
)

var (
	ErrViewModel      = errors.New("viewing UI model error")
	ErrRetrieveModel  = errors.New("failed retrieve model")
	ErrUserStoppedApp = errors.New("user stopped execution")
)

// MainModel manages which screen is active
type MainModel struct {
	currentScreen ScreenType
	authModel     viewauth.Model
	registerModel viewregister.Model
	loginModel    viewlogin.Model
	homeModel     viewhome.Model
	backupModel   viewbackup.Model
}

type App struct {
	Logger    *slog.Logger
	Cfg       *config.Config
	ch        chan Message
	MainModel MainModel
	Auth      Auth
}

type Message struct {
	Token string `json:"token"`
	Type  string `json:"type"`
	Value []byte `json:"value"`
}

func NewApp(ctx context.Context, cfg *config.Config, Logger *slog.Logger, s Services) *App {
	startScreen := AuthScreen
	at, err := s.GetAccessToken(ctx)
	if err != nil {
		Logger.Debug("failed to retrieve access token", slog.String("error", err.Error()))
	}

	if at != "" {
		b, err := keyring.GetAuthCreds(cfg.ServiceMasterPasswordKey, cfg.AppName)
		if err != nil {
			Logger.Debug("failed to retrieve the backup key", slog.String("error", err.Error()))
		}

		if b != "" {
			startScreen = HomeScreen
		} else {
			startScreen = BackupScreen
		}
	}

	return &App{
		MainModel: MainModel{
			currentScreen: startScreen,
			authModel:     viewauth.NewModel(AppName, ""),
			registerModel: viewregister.NewModel(AppName, ""),
			loginModel:    viewlogin.NewModel(s, Logger, AppName, ""),
			backupModel:   viewbackup.NewModel(s, Logger, AppName, ""),
			homeModel:     viewhome.NewModel(AppName, ""),
		},
		Logger: Logger,
		Cfg:    cfg,
	}
}

func (a *App) Run(ctx context.Context, stop chan os.Signal) {
	const op = "client.Run"
	Logger := a.Logger.With(
		slog.String("op", op),
	)

	Logger.Info("starting application")

	a.ch = make(chan Message)
	p := tea.NewProgram(a.MainModel, tea.WithAltScreen())
	_, _ = p.Run()

	Logger.Info("user stopped execution (q, ctrl+c, esc)")
	stop <- syscall.SIGTERM
	return
}

func (a *App) Stop() {
	close(a.ch)
}

func (m MainModel) Init() tea.Cmd {
	return m.authModel.Init() // Start with Model's Init
}

// Update function (switches screens)
func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Let the active screen handle key presses first
		var newModel tea.Model
		var cmd tea.Cmd

		switch m.currentScreen {
		case AuthScreen:
			newModel, cmd = m.authModel.Update(msg)
			m.authModel = newModel.(viewauth.Model)
		case RegisterScreen:
			newModel, cmd = m.registerModel.Update(msg)
			m.registerModel = newModel.(viewregister.Model)
		case LoginScreen:
			newModel, cmd = m.loginModel.Update(msg)
			m.loginModel = newModel.(viewlogin.Model)
		case HomeScreen:
			newModel, cmd = m.homeModel.Update(msg)
			m.homeModel = newModel.(viewhome.Model)
		case BackupScreen:
			newModel, cmd = m.backupModel.Update(msg)
			m.backupModel = newModel.(viewbackup.Model)
		}

		// If the screen returned a command (not nil), pass it along
		if cmd != nil {
			return m, cmd
		}

		// Handle global keys (exit app)
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		}

	case messages.ActionMsg:
		switch msg.Value {
		case "Login":
			m.currentScreen = LoginScreen
		case "Register":
			m.currentScreen = RegisterScreen
		case viewbackup.BackUpKeyStoredSuccess:
			m.currentScreen = HomeScreen
		}
		return m, nil
	}

	return m, nil
}

// MainModel View function (renders the active screen)
func (m MainModel) View() string {
	switch m.currentScreen {
	case AuthScreen:
		return m.authModel.View()
	case LoginScreen:
		return m.loginModel.View()
	case RegisterScreen:
		return m.registerModel.View()
	case HomeScreen:
		return m.homeModel.View()
	case BackupScreen:
		return m.backupModel.View()
	default:
		return "Unknown screen"
	}
}
