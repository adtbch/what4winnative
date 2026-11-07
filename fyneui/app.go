package fyneui

import (
	"strconv"
	"context"
	"image/color"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/lugvitc/whats4linux/api"
)

type FyneApp struct {
	app     fyne.App
	win     fyne.Window
	api     *api.Api
	ctx     context.Context
	cancel  context.CancelFunc
	state   *AppState
	stack   *fyne.Container

	sidebar *Sidebar
	chatBox *fyne.Container
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
	fa.win = fa.app.NewWindow("WhatsApp")
	fa.stack = container.NewStack(canvas.NewRectangle(waBg), widget.NewLabel("Loading..."))
	fa.win.SetContent(fa.stack)
	fa.win.Resize(fyne.NewSize(1200, 760))
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

func (fa *FyneApp) OpenChat(jid string) {
	fa.state.SelectedChatID = jid
	cv := NewChatView(fa, jid)
	fyne.Do(func() {
		if fa.chatBox == nil { return }
		fa.chatBox.Objects = []fyne.CanvasObject{cv.Content()}
		fa.chatBox.Refresh()
	})
}

func railBtn(icon string, badge int) fyne.CanvasObject {
	lbl := canvas.NewText(icon, waMuted)
	lbl.Alignment = fyne.TextAlignCenter
	lbl.TextSize = 16
	box := container.NewCenter(lbl)
	if badge > 0 {
		b := canvas.NewText(strconv.Itoa(badge), waWhite)
		b.TextSize = 10
		b.TextStyle = fyne.TextStyle{Bold:true}
		bg := canvas.NewCircle(waGreen)
		bg.Resize(fyne.NewSize(18,18))
		badgeObj := container.NewMax(bg, container.NewCenter(b))
		box = container.NewMax(box, container.NewWithoutLayout(container.NewGridWrap(fyne.NewSize(18,18), badgeObj)))
	}
	return container.NewPadded(box)
}

func (fa *FyneApp) emptyPane() fyne.CanvasObject {
	title := canvas.NewText("Download WhatsApp for Windows", waText)
	title.TextSize = 22
	title.Alignment = fyne.TextAlignCenter
	desc := canvas.NewText("Make calls, share your screen and get a faster experience.", waMuted)
	desc.TextSize = 13
	desc.Alignment = fyne.TextAlignCenter
	btn := widget.NewButton("Download", func() {})
	// style via importance not exact, ok
	btn.Importance = widget.SuccessImportance
	content := container.NewVBox(
		container.NewCenter(title),
		container.NewGridWrap(fyne.NewSize(0,8), canvas.NewRectangle(color.Transparent)),
		container.NewCenter(desc),
		container.NewGridWrap(fyne.NewSize(0,18), canvas.NewRectangle(color.Transparent)),
		container.NewCenter(container.NewGridWrap(fyne.NewSize(120,36), btn)),
		container.NewGridWrap(fyne.NewSize(0,24), canvas.NewRectangle(color.Transparent)),
		container.NewCenter(canvas.NewText("🔒 Your personal messages are end-to-end encrypted", waMuted)),
	)
	return container.NewMax(canvas.NewRectangle(waPanel2), container.NewCenter(container.NewPadded(content)))
}

func (fa *FyneApp) SwitchToChats() {
	fa.state.Screen = "chats"
	if fa.sidebar == nil { fa.sidebar = NewSidebar(fa) } else { go fa.sidebar.loadChats() }
	if fa.chatBox == nil { fa.chatBox = container.NewStack(fa.emptyPane()) }
	fyne.Do(func() {
		railBg := canvas.NewRectangle(waRail)
		rail := container.NewVBox(
			container.NewGridWrap(fyne.NewSize(48,12), canvas.NewRectangle(color.Transparent)),
			railBtn("☰", 0), railBtn("💬", 0), railBtn("○", 0), railBtn("◎", 0), railBtn("⚙", 1),
		)
		railCol := container.NewMax(railBg, container.NewGridWrap(fyne.NewSize(52,0), container.NewPadded(rail)))

		left := container.NewMax(canvas.NewRectangle(waPanel), fa.sidebar.Content())
		leftWrap := container.NewHBox(railCol, container.NewGridWrap(fyne.NewSize(360,0), left))
		split := container.NewHSplit(leftWrap, fa.chatBox)
		split.Offset = 0.36
		fa.stack.Objects = []fyne.CanvasObject{container.NewMax(canvas.NewRectangle(waBg), split)}
		fa.stack.Refresh()
	})
}

func (fa *FyneApp) Run() {
	defer fa.cancel()
	ctx := context.WithValue(fa.ctx, "ui_mode", "fyne")
	go func() { fa.api.Startup(ctx) }()
	fa.ShowLogin()
	fa.win.ShowAndRun()
}
