package client

import (
	"context"
	"errors"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mihailtudos/gophkeeper/internal/client/application/security"
	"github.com/mihailtudos/gophkeeper/internal/client/application/services"
	"github.com/mihailtudos/gophkeeper/internal/client/cli/messages"
	viewauth "github.com/mihailtudos/gophkeeper/internal/client/cli/view_auth"
	viewbackup "github.com/mihailtudos/gophkeeper/internal/client/cli/view_backup"
	viewhome "github.com/mihailtudos/gophkeeper/internal/client/cli/view_home"
	viewlogin "github.com/mihailtudos/gophkeeper/internal/client/cli/view_login"
	viewregister "github.com/mihailtudos/gophkeeper/internal/client/cli/view_register"
	viewsecretcreator "github.com/mihailtudos/gophkeeper/internal/client/cli/view_secret_creator"
	"github.com/mihailtudos/gophkeeper/internal/client/config"
	"log/slog"
	"os"
	"syscall"
	"time"
)

// ScreenType defines the different screens
type ScreenType int

const (
	AppName                     = "GopherKeep"
	backToHomeKey               = "back_to_home"
	logoutMessageKey            = "logout"
	AuthScreen       ScreenType = iota
	LoginScreen
	RegisterScreen
	BackupScreen
	HomeScreen
	SecretCreatorScreen
)

var (
	ErrViewModel      = errors.New("viewing UI model error")
	ErrRetrieveModel  = errors.New("failed retrieve model")
	ErrUserStoppedApp = errors.New("user stopped execution")
)

// MainModel manages which screen is active
type MainModel struct {
	currentScreen             ScreenType
	authModel                 viewauth.Model
	registerModel             viewregister.Model
	loginModel                viewlogin.Model
	homeModel                 viewhome.Model
	backupModel               viewbackup.Model
	secretCreator             viewsecretcreator.Model
	keyManager                security.KeyManagerProvider
	cfg                       *config.Config
	isNotAuthenticatedHandled bool
}

type App struct {
	Logger    *slog.Logger
	Cfg       *config.Config
	ch        chan Message
	MainModel MainModel
	Services  *services.Services
}

type Message struct {
	Token string `json:"token"`
	Type  string `json:"type"`
	Value []byte `json:"value"`
}

func NewApp(ctx context.Context, cfg *config.Config, Logger *slog.Logger, s *services.Services) *App {
	startScreen := AuthScreen
	at, err := s.AuthService.GetAccessToken(ctx)
	if err != nil {
		Logger.Debug("failed to retrieve access token", slog.String("error", err.Error()))
	}

	if at != "" {
		b, err := s.KeyManager.GetKey(cfg.ServiceMasterPasswordKey, cfg.AppName)
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
			keyManager:    s.KeyManager,
			cfg:           cfg,
			authModel:     viewauth.NewModel(AppName, ""),
			registerModel: viewregister.NewModel(AppName, ""),
			loginModel:    viewlogin.NewModel(s.AuthService, Logger, AppName, ""),
			backupModel:   viewbackup.NewModel(s.AuthService, Logger, AppName, ""),
			homeModel:     viewhome.NewModel(AppName, ""),
			secretCreator: viewsecretcreator.NewModel(s.SecretsService, AppName, ""),
		},
		Logger: Logger,
		Cfg:    cfg,
	}
}

func (a *App) Run(ctx context.Context, stop chan os.Signal) {
	const op = "client.Run"
	Logger := a.Logger.With(slog.String("op", op))

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
		case SecretCreatorScreen:
			newModel, cmd = m.secretCreator.Update(msg)
			m.secretCreator = newModel.(viewsecretcreator.Model)
		}

		// If the screen returned a command (not nil), pass it along
		if cmd != nil {
			return m, cmd
		}

		// Handle global keys (exit app)
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "ctrl+l":
			return m, func() tea.Msg {
				return messages.ActionMsg{Value: logoutMessageKey}
			}
		}

	case messages.ActionMsg:
		switch msg.Value {
		case "Login":
			m.currentScreen = LoginScreen
		case "Register":
			m.currentScreen = RegisterScreen
		case viewbackup.BackUpKeyStoredSuccess:
			m.currentScreen = HomeScreen
		case viewhome.CreateNewSecretMessageKey:
			m.currentScreen = SecretCreatorScreen
		case viewlogin.LoggedInSuccessMsgKey:
			bk, _ := m.keyManager.GetKey(m.cfg.ServiceMasterPasswordKey, m.cfg.AppName)
			if bk == "" {
				m.currentScreen = BackupScreen
			}
			m.currentScreen = HomeScreen
		case backToHomeKey:
			m.currentScreen = HomeScreen
		case logoutMessageKey:
			_ = m.keyManager.RemoveKey(m.cfg.ServiceMasterPasswordKey, m.cfg.AppName)
			_ = m.keyManager.RemoveKey(m.cfg.ServiceAccessTokenKey, m.cfg.AppName)
			_ = m.keyManager.RemoveKey(m.cfg.ServiceRefreshTokenKey, m.cfg.AppName)
			m.currentScreen = AuthScreen
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
	case SecretCreatorScreen:
		return m.secretCreator.View()
	default:
		return "Unknown screen"
	}
}

func isTokenExpired(tokenString string) bool {
	// Parse the token without validating the signature
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		// If we can't parse the token, consider it expired
		return true
	}

	// Get the claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		// If we can't get the claims, consider it expired
		return true
	}

	// Check the exp claim
	exp, ok := claims["exp"].(float64)
	if !ok {
		// If there's no exp claim or it's not a number, consider it expired
		return true
	}

	// Convert exp to time.Time and check if it's in the past
	expTime := time.Unix(int64(exp), 0)
	return time.Now().After(expTime)
}
