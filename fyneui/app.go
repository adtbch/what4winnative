package fyneui

import (
	"context"
	"image/color"
	"strconv"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
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

	sidebar  *Sidebar
	chatPane *fyne.Container
	shell    fyne.CanvasObject
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
	fa.win.Resize(fyne.NewSize(1360, 860))
	fa.stack = container.NewMax(canvas.NewRectangle(waBg), widget.NewLabel("Loading..."))
	fa.win.SetContent(fa.stack)
	return fa
}

func (fa *FyneApp) ShowLogin() {
	fa.state.SetScreen("login")
	login := NewLoginScreen(fa)
	fyne.Do(func() {
		fa.stack.Objects = []fyne.CanvasObject{login.Content()}
		fa.stack.Refresh()
	})
}

func (fa *FyneApp) OpenChat(jid string) {
	if jid == "" {
		return
	}
	fa.state.SetSelectedChatID(jid)
	cv := NewChatView(fa, jid)
	fyne.Do(func() {
		if fa.chatPane == nil {
			return
		}
		fa.chatPane.Objects = []fyne.CanvasObject{cv.Content()}
		fa.chatPane.Refresh()
	})
}

func (fa *FyneApp) OpenEmptyChat() {
	fa.state.SetSelectedChatID("")
	fyne.Do(func() {
		if fa.chatPane == nil {
			return
		}
		fa.chatPane.Objects = []fyne.CanvasObject{fa.emptyPane()}
		fa.chatPane.Refresh()
	})
}

func railItem(label string, active bool, badge int) fyne.CanvasObject {
	fg := waRailIcon
	if active {
		fg = waText
	}
	t := canvas.NewText(label, fg)
	t.TextSize = 18
	t.Alignment = fyne.TextAlignCenter
	t.TextStyle = fyne.TextStyle{Bold: active}
	cell := container.NewMax(canvas.NewRectangle(color.Transparent), container.NewCenter(t))
	if active {
		bg := canvas.NewRectangle(waRailSel)
		bg.CornerRadius = 999
		cell = container.NewMax(bg, container.NewCenter(t))
	}
	wrap := container.NewGridWrap(fyne.NewSize(40, 40), container.NewCenter(cell))
	if badge > 0 {
		btxt := strconv.Itoa(badge)
		if badge > 99 {
			btxt = "99+"
		}
		bt := canvas.NewText(btxt, waWhite)
		bt.TextSize = 9
		bt.TextStyle = fyne.TextStyle{Bold: true}
		bbg := canvas.NewCircle(waGreen)
		badgeObj := container.NewGridWrap(fyne.NewSize(18, 18), container.NewMax(bbg, container.NewCenter(bt)))
		wrap = container.NewMax(wrap, container.NewWithoutLayout(badgeObj))
	}
	return wrap
}

func railAvatar() fyne.CanvasObject {
	bg := canvas.NewCircle(waHover)
	txt := canvas.NewText("•", waMuted)
	txt.Alignment = fyne.TextAlignCenter
	txt.TextSize = 16
	return container.NewGridWrap(fyne.NewSize(32, 32), container.NewMax(bg, container.NewCenter(txt)))
}

func (fa *FyneApp) rail() fyne.CanvasObject {
	railBg := canvas.NewRectangle(waRail)
	top := container.NewVBox(
		spacer(1, 10),
		container.NewCenter(railItem("≡", false, 0)),
		spacer(1, 4),
		container.NewCenter(railItem("💬", true, 0)),
		spacer(1, 4),
		container.NewCenter(railItem("⭘", false, 0)),
		spacer(1, 4),
		container.NewCenter(railItem("◎", false, 0)),
		spacer(1, 4),
		container.NewCenter(railItem("👥", false, 0)),
		spacer(1, 4),
		container.NewCenter(railItem("🤖", false, 0)), // bot/morty slot compatible
	)
	bottom := container.NewVBox(
		container.NewCenter(railItem("⚙", false, 0)),
		spacer(1, 10),
		container.NewCenter(railAvatar()),
		spacer(1, 12),
	)
	rail := container.NewBorder(top, bottom, nil, nil, canvas.NewRectangle(color.Transparent))
	return container.NewMax(railBg, container.NewGridWrap(fyne.NewSize(64, 0), rail))
}

func (fa *FyneApp) emptyPane() fyne.CanvasObject {
	bg := chatBackground(fyne.NewSize(0, 0))
	logo := canvas.NewCircle(waPanel2)
	logoTxt := canvas.NewText("WA", waAccent)
	logoTxt.TextSize = 24
	logoTxt.TextStyle = fyne.TextStyle{Bold: true}
	logoTxt.Alignment = fyne.TextAlignCenter
	logoObj := container.NewGridWrap(fyne.NewSize(84, 84), container.NewMax(logo, container.NewCenter(logoTxt)))

	title := canvas.NewText("WhatsApp for Linux", waText)
	title.TextSize = 32
	title.Alignment = fyne.TextAlignCenter
	desc := canvas.NewText("Send and receive messages without keeping your phone online.", waMuted)
	desc.TextSize = 14
	desc.Alignment = fyne.TextAlignCenter
	desc2 := canvas.NewText("Use WhatsApp on up to 4 linked devices and 1 phone.", waMuted)
	desc2.TextSize = 14
	desc2.Alignment = fyne.TextAlignCenter

	download := widget.NewButton("Download", func() {})
	download.Importance = widget.SuccessImportance

	lock := canvas.NewText("🔒  Your personal messages are end-to-end encrypted", waMuted)
	lock.TextSize = 13
	lock.Alignment = fyne.TextAlignCenter

	accent := canvas.NewRectangle(waAccent)
	accent.SetMinSize(fyne.NewSize(0, 6))

	content := container.NewVBox(
		container.NewCenter(logoObj),
		spacer(1, 24),
		container.NewCenter(title),
		spacer(1, 14),
		container.NewCenter(desc),
		spacer(1, 4),
		container.NewCenter(desc2),
		spacer(1, 28),
		container.NewCenter(container.NewGridWrap(fyne.NewSize(132, 38), download)),
		spacer(1, 70),
		container.NewCenter(lock),
	)
	return container.NewBorder(nil, accent, nil, nil, container.NewMax(bg, container.NewCenter(container.NewPadded(content))))
}

func (fa *FyneApp) buildShell() {
	if fa.sidebar == nil {
		fa.sidebar = NewSidebar(fa)
	} else {
		go fa.sidebar.Reload()
	}
	if fa.chatPane == nil {
		fa.chatPane = container.NewMax(fa.emptyPane())
	}

	rail := fa.rail()
	leftBody := container.NewMax(canvas.NewRectangle(waPanel), fa.sidebar.Content())
	left := container.NewHBox(rail, verticalDivider(), container.New(&sidebarWidthLayout{width: 420}, leftBody))
	right := container.NewMax(canvas.NewRectangle(waBg), fa.chatPane)

	split := container.New(&mainSplitLayout{leftRatio: 0.34, minLeft: 460, minRight: 420}, left, right)
	fa.shell = container.NewMax(canvas.NewRectangle(waBg), split)
}

func (fa *FyneApp) SwitchToChats() {
	fa.state.SetScreen("chats")
	fa.buildShell()
	fyne.Do(func() {
		fa.stack.Objects = []fyne.CanvasObject{fa.shell}
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

func verticalDivider() fyne.CanvasObject {
	line := canvas.NewRectangle(waDivider)
	line.SetMinSize(fyne.NewSize(1, 0))
	return line
}

// sidebarWidthLayout forces sidebar body width while keeping rail natural.
type sidebarWidthLayout struct {
	width float32
}

func (l *sidebarWidthLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	h := float32(0)
	for _, o := range objects {
		if o.MinSize().Height > h {
			h = o.MinSize().Height
		}
	}
	w := l.width
	if w <= 0 {
		w = 420
	}
	return fyne.NewSize(w, h)
}

func (l *sidebarWidthLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objects {
		o.Resize(size)
		o.Move(fyne.NewPos(0, 0))
	}
}

// mainSplitLayout emulates WA desktop split (~30-35% left).
type mainSplitLayout struct {
	leftRatio float32
	minLeft   float32
	minRight  float32
}

func (l *mainSplitLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(l.minLeft+l.minRight, 600)
}

func (l *mainSplitLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 2 {
		if len(objects) == 1 {
			objects[0].Resize(size)
			objects[0].Move(fyne.NewPos(0, 0))
		}
		return
	}
	r := l.leftRatio
	if r <= 0 || r >= 1 {
		r = 0.34
	}
	leftW := size.Width * r
	if leftW < l.minLeft {
		leftW = l.minLeft
	}
	if leftW > size.Width-l.minRight {
		leftW = size.Width - l.minRight
	}
	if leftW < 0 {
		leftW = 0
	}
	objects[0].Move(fyne.NewPos(0, 0))
	objects[0].Resize(fyne.NewSize(leftW, size.Height))
	objects[1].Move(fyne.NewPos(leftW, 0))
	objects[1].Resize(fyne.NewSize(size.Width-leftW, size.Height))
}

var _ = color.Transparent
