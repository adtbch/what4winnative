package fyneui

import (
	"context"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/lugvitc/whats4linux/api"
)

type FyneApp struct {
	app    fyne.App
	win    fyne.Window
	api    *api.Api
	ctx    context.Context
	cancel context.CancelFunc
	state  *AppState
	stack  *fyne.Container
}

func New(a *api.Api) *FyneApp {
	ctx, cancel := context.WithCancel(context.Background())
	fa := &FyneApp{
		app:    fyneapp.New(),
		api:    a,
		ctx:    ctx,
		cancel: cancel,
		state:  &AppState{api: a, Screen: "login"},
	}
	fa.win = fa.app.NewWindow("what4winnative")
	fa.stack = container.NewStack(widget.NewLabel("Loading..."))
	fa.win.SetContent(fa.stack)
	fa.win.Resize(fyne.NewSize(1100, 700))
	return fa
}

func (fa *FyneApp) ShowLogin() {
	fa.state.Screen = "login"
	login := NewLoginScreen(fa)
	fyne.Do(func() {
		fa.stack.Objects = []fyne.CanvasObject{login.Content()}
		fa.stack.Refresh()
	})
}

func (fa *FyneApp) SwitchToChats() {
	fa.state.Screen = "chats"
	fyne.Do(func() {
		fa.stack.Objects = []fyne.CanvasObject{widget.NewLabel("Chats (WIP)")}
		fa.stack.Refresh()
	})
}

func (fa *FyneApp) Run() {
	defer fa.cancel()
	go func() { fa.api.Startup(fa.ctx) }()
	fa.ShowLogin()
	fa.win.ShowAndRun()
}
