package fyneui

import (
	"context"
	"image/color"
	"os"
	"strconv"
	"time"

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
	fa.app.Settings().SetTheme(newWATheme())
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

func railItem(path string, name string, active bool, badge int, onTap func()) fyne.CanvasObject {
	fg := waRailIcon
	if active {
		fg = waText
	}
	icon := iconImage(name, path, 24, fg)
	cell := container.NewMax(container.NewCenter(icon))
	if active {
		bg := canvas.NewRectangle(waRailSel)
		bg.CornerRadius = 999
		cell = container.NewMax(bg, container.NewCenter(icon))
	}
	wrap := container.NewGridWrap(fyne.NewSize(44, 44), cell)
	if badge > 0 {
		btxt := strconv.Itoa(badge)
		if badge > 99 {
			btxt = "99+"
		}
		bt := canvas.NewText(btxt, waWhite)
		bt.TextSize = 9
		bt.TextStyle = fyne.TextStyle{Bold: true}
		bt.Alignment = fyne.TextAlignCenter
		bbg := canvas.NewCircle(waGreen)
		badgeObj := container.NewGridWrap(fyne.NewSize(18, 18), container.NewMax(bbg, container.NewCenter(bt)))
		wrap = container.NewMax(wrap, container.NewWithoutLayout(badgeObj))
	}
	if onTap != nil {
		wrap = container.NewMax(wrap, &tapArea{onTap: onTap})
	}
	return wrap
}

func railAvatar() fyne.CanvasObject {
	bg := canvas.NewCircle(waHover)
	icon := iconImage("rail-avatar", iconPerson, 26, waRailIcon)
	return container.NewGridWrap(fyne.NewSize(34, 34), container.NewMax(bg, container.NewCenter(icon)))
}

func (fa *FyneApp) rail() fyne.CanvasObject {
	railBg := canvas.NewRectangle(waRail)
	top := container.NewVBox(
		spacer(1, 10),
		container.NewCenter(railItem(iconChats, "rail-chats", true, 0, nil)),
		spacer(1, 2),
		container.NewCenter(railItem(iconStatus, "rail-status", false, 0, nil)),
		spacer(1, 2),
		container.NewCenter(railItem(iconChannels, "rail-channels", false, 0, nil)),
		spacer(1, 2),
		container.NewCenter(railItem(iconCommunities, "rail-communities", false, 0, nil)),
	)
	bottom := container.NewVBox(
		container.NewCenter(railItem(iconSettings, "rail-settings", false, 0, func(){ fa.OpenSettings() })),
		spacer(1, 10),
		container.NewCenter(railAvatar()),
		spacer(1, 12),
	)
	inner := container.NewBorder(top, bottom, nil, nil, canvas.NewRectangle(color.Transparent))
	return container.NewMax(railBg, inner)
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

	leftBody := container.NewMax(canvas.NewRectangle(waSidebarBg), fa.sidebar.Content())
	right := container.NewMax(canvas.NewRectangle(waChatBg), fa.chatPane)

	// React: ResizablePanel defaultSize 30%, minSize 320px, maxSize 600px.
	split := container.New(&mainSplitLayout{leftRatio: 0.30, minLeft: 320, maxLeft: 600, minRight: 400}, leftBody, right)
	fa.shell = container.NewMax(canvas.NewRectangle(waChatBg), split)
}

func (fa *FyneApp) SwitchToChats() {
	fa.state.SetScreen("chats")
	fa.buildShell()
	fyne.Do(func() {
		fa.stack.Objects = []fyne.CanvasObject{fa.shell}
		fa.stack.Refresh()
	})
	// dev hook: auto-open a chat so the bubble rendering can be inspected
	// without clicking (WHATS4LINUX_OPEN_JID=<jid or phone>).
	if jid := os.Getenv("WHATS4LINUX_OPEN_JID"); jid != "" {
		go func() {
			time.Sleep(1500 * time.Millisecond)
			fa.OpenChat(jid)
		}()
	}
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

// mainSplitLayout emulates the React resizable split (30% left, 320-600px).
type mainSplitLayout struct {
	leftRatio float32
	minLeft   float32
	maxLeft   float32
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
		r = 0.30
	}
	leftW := size.Width * r
	if leftW < l.minLeft {
		leftW = l.minLeft
	}
	if l.maxLeft > 0 && leftW > l.maxLeft {
		leftW = l.maxLeft
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


func (fa *FyneApp) OpenSettings() {
	fyne.Do(func() {
		scr := newSettingsScreen(func() { fa.SwitchToChats() })
		fa.stack.Objects = []fyne.CanvasObject{scr}
		fa.stack.Refresh()
	})
}
