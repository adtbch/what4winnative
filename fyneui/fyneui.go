package fyneui

import (
    "github.com/adtbch/whats4linux/fyneui/utils"
    "github.com/adtbch/whats4linux/api"
    "github.com/adtbch/whats4linux/fyneui/login"
)

// FyneApp is the exported application type. It is defined in the utils package.
// We alias it here so callers can import "github.com/adtbch/whats4linux/fyneui" and
// use fyneui.FyneApp directly.
type FyneApp = utils.FyneApp

// New creates a new FyneApp instance. Delegates to utils.New.
func New(a *api.Api) *FyneApp {
    return utils.New(a)
}

// Run starts the UI loop (calls utils.FyneApp.Run()).
func Run(app *FyneApp) {
    app.Run()
}

// ShowLogin renders the QR login screen using the wrapper's SetRoot.
func ShowLogin(app *FyneApp) {
    ls := login.NewLoginScreen(app)
    app.SetRoot(ls.Content())
}
