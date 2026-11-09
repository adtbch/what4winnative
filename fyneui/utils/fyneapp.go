package utils

import (
    "context"

    "fyne.io/fyne/v2"
    fyneapp "fyne.io/fyne/v2/app"
    "fyne.io/fyne/v2/canvas"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"

    "github.com/adtbch/whats4linux/api"
    themePkg "github.com/adtbch/whats4linux/fyneui/theme"
)

// FyneApp is the core application object. It holds the window, API client,
// context and shared state. UI‑specific helpers are defined in the top‑level
// fyneui package to avoid import cycles.
type FyneApp struct {
    app    fyne.App
    win    fyne.Window
    api    *api.Api
    ctx    context.Context
    cancel context.CancelFunc
    state  *AppState
    stack  *fyne.Container
}

// New creates a new FyneApp instance with minimal UI scaffolding. The caller
// (fyneui wrapper) can set the theme and attach UI components.
func New(a *api.Api) *FyneApp {
    ctx, cancel := context.WithCancel(context.Background())
    fa := &FyneApp{
        app:    fyneapp.New(),
        api:    a,
        ctx:    ctx,
        cancel: cancel,
        state:  NewAppState(a),
    }
    fa.win = fa.app.NewWindow("WhatsApp")
    // Theme is set by the wrapper (fyneui) after construction.
    fa.win.Resize(fyne.NewSize(1360, 860))
    fa.stack = container.NewMax(canvas.NewRectangle(themePkg.WaBg), widget.NewLabel("Loading..."))
    fa.win.SetContent(fa.stack)
    return fa
}

func (fa *FyneApp) API() *api.Api { return fa.api }

func (fa *FyneApp) Ctx() context.Context { return fa.ctx }

// SetRoot replaces the current UI content with the provided CanvasObject.
func (fa *FyneApp) SetRoot(obj fyne.CanvasObject) {
    fyne.Do(func() {
        fa.stack.Objects = []fyne.CanvasObject{obj}
        fa.stack.Refresh()
    })
}

// SwitchToChats swaps UI to the chats view. Placeholder for now.
func (fa *FyneApp) SwitchToChats() {
    fyne.Do(func() {
        fa.SetRoot(widget.NewLabel("Chats"))
    })
}

// Run starts the backend API and runs the window.
func (fa *FyneApp) Run() {
    defer fa.cancel()
    ctx := context.WithValue(fa.ctx, "ui_mode", "fyne")
    go func() { fa.api.Startup(ctx) }()
    // Show window (wrapper may replace content later)
    fa.win.Show()
    // Keep window alive until closed.
    fa.win.ShowAndRun()
}


