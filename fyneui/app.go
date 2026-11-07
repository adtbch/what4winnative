package fyneui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/lugvitc/whats4linux/api"
)

func Run(a *api.Api) fyne.App {
	myApp := app.New()
	w := myApp.NewWindow("whats4linux (fyne)")
	w.SetMainMenu(fyne.NewMainMenu(fyne.NewMenu("File", fyne.NewMenuItem("Quit", func() { myApp.Quit() }))))
	w.SetContent(container.NewCenter(widget.NewLabel("Migrating to Fyne - WIP")))
	w.Resize(fyne.NewSize(1100, 700))
	w.ShowAndRun()
	return myApp
}
