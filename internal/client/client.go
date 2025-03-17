package client

import (
	"context"
	"errors"
	tea "github.com/charmbracelet/bubbletea"
	viewauth "github.com/mihailtudos/gophkeeper/internal/client/cli/view_auth"
	viewregister "github.com/mihailtudos/gophkeeper/internal/client/cli/view_register"
	"github.com/mihailtudos/gophkeeper/internal/client/config"
	"log/slog"
	"os"
	"syscall"
)

const (
	AppName = "GopherKeep"
)

var (
	ErrViewModel      = errors.New("viewing UI model error")
	ErrRetrieveModel  = errors.New("failed retrieve model")
	ErrUserStoppedApp = errors.New("user stopped execution")
)

type ClientApp struct {
	Logger *slog.Logger
	Cfg    *config.Config
	ch     chan Message
}

type Message struct {
	Token string `json:"token"`
	Type  string `json:"type"`
	Value []byte `json:"value"`
}

func NewApp(ctx context.Context, cfg *config.Config, Logger *slog.Logger) *ClientApp {
	return &ClientApp{
		Logger: Logger,
		Cfg:    cfg,
	}
}

func (a *ClientApp) Run(ctx context.Context, stop chan os.Signal) {
	const op = "client.Run"
	Logger := a.Logger.With(
		slog.String("op", op),
	)

	Logger.Info("starting application")

	a.ch = make(chan Message)
	p := tea.NewProgram(viewauth.Model{
		AppName: AppName,
	}, tea.WithAltScreen())
	m, _ := p.Run()

	modelAuth, _ := m.(viewauth.Model)
	if modelAuth.Choice == "" {
		Logger.Info("user stopped execution (q, ctrl+c, esc)")
		stop <- syscall.SIGTERM
		return
	}

	if modelAuth.Choice == "Register" {
		if err := a.registration(ctx); err != nil {
			Logger.Error("registration failed", slog.String("error", err.Error()))
			stop <- syscall.SIGTERM
			return
		}
	}
}

func (a *ClientApp) Stop() {
	close(a.ch)
}

func (a *ClientApp) registration(ctx context.Context) error {
loop:
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			p := tea.NewProgram(viewregister.InitialModel(), tea.WithAltScreen())
			m, err := p.Run()
			if err != nil {
				return ErrViewModel
			}

			modelRegister, ok := m.(*viewregister.Model)
			if !ok {
				return ErrRetrieveModel
			}

			if modelRegister.State == "" {
				// user stopped execution in UI (q, ctrl+C, esc)
				return ErrUserStoppedApp
			}

			if modelRegister.State == "again" {
				continue
			}

			if modelRegister.State == "error" {
				return errors.New("registration failed")
			}

			break loop
		}
	}
	return nil
}

//
//func (app *AppClient) login(ctx context.Context) (string, error) {
//	for {
//		select {
//		case <-ctx.Done():
//			return "", nil
//		default:
//			p := tea.NewProgram(viewlogin.InitialModel(app.grpcClient))
//			m, err := p.Run()
//			if err != nil {
//				return "", ErrViewModel
//			}
//
//			modelLogin, ok := m.(viewlogin.Model)
//			if !ok {
//				return "", ErrRetrieveModel
//			}
//
//			if modelLogin.State == "" {
//				// user stopped execution in UI (q, ctrl+C, esc)
//				return "", ErrUserStoppedApp
//			}
//
//			if modelLogin.State == "again" {
//				continue
//			}
//
//			if modelLogin.State == "error" {
//				return "", errors.New("failed login user")
//			}
//
//			return modelLogin.Token, nil
//		}
//	}
//}
