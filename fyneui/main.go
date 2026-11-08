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

	"github.com/lugvitc/whats4linux/fyneui/chat"
	"github.com/lugvitc/whats4linux/fyneui/login"
	"github.com/lugvitc/whats4linux/fyneui/settings"
	"github.com/lugvitc/whats4linux/fyneui/sidebar"
	"github.com/lugvitc/whats4linux/fyneui/theme"
	"github.com/lugvitc/whats4linux/fyneui/utils"
)

type FyneApp struct {
	currentCV *chat.ChatView
	app       fyne.App
	win       fyne.Window
	api       *api.Api
	ctx       context.Context
	cancel    context.CancelFunc
	state     *utils.AppState
	stack     *fyne.Container

	sidebar  *sidebar.Sidebar
	chatPane *fyne.Container
	shell    fyne.CanvasObject
}

func (fa *FyneApp) API() *api.Api          { return fa.api }
func (fa *FyneApp) Ctx() context.Context   { return fa.ctx }
func (fa *FyneApp) Win() fyne.Window       { return fa.win }
func (fa *FyneApp) State() *utils.AppState { return fa.state }

func New(a *api.Api) *FyneApp {
	ctx, cancel := context.WithCancel(context.Background())
	fa := &FyneApp{
		app:    fyneapp.New(),
		api:    a,
		ctx:    ctx,
		cancel: cancel,
		state:  utils.NewAppState(a),
	}
	fa.win = fa.app.NewWindow("WhatsApp")
	fa.app.Settings().SetTheme(theme.NewWATheme())
	fa.win.Resize(fyne.NewSize(1360, 860))
	fa.stack = container.NewMax(canvas.NewRectangle(theme.WaBg), widget.NewLabel("Loading..."))
	fa.win.SetContent(fa.stack)
	fyne.Do(func() { fa.ShowLogin() })
	go func() {
		for !fa.api.IsWaConnected() {
			time.Sleep(200 * time.Millisecond)
		}
		fyne.Do(func() { fa.ShowLogin() })
	}()
	return fa
}

func (fa *FyneApp) ShowLogin() {
	fa.State().SetScreen("login")
	loginScreen := login.NewLoginScreen(fa)
	fyne.Do(func() {
		fa.stack.Objects = []fyne.CanvasObject{loginScreen.Content()}
		fa.stack.Refresh()
	})
}

func (fa *FyneApp) OpenChat(jid string) {
	if jid == "" {
		return
	}
	fa.State().SetSelectedChatID(jid)
	old := fa.currentCV
	cv := chat.NewChatView(fa, jid)
	fa.currentCV = cv
	if old != nil {
		old.Cancel()
	}
	fyne.Do(func() {
		if fa.chatPane == nil {
			return
		}
		fa.chatPane.Objects = []fyne.CanvasObject{cv.Content()}
		fa.chatPane.Refresh()
	})
}

func (fa *FyneApp) OpenEmptyChat() {
	fa.State().SetSelectedChatID("")
	if fa.currentCV != nil {
		fa.currentCV.Cancel()
		fa.currentCV = nil
	}
	fyne.Do(func() {
		if fa.chatPane == nil {
			return
		}
		fa.chatPane.Objects = []fyne.CanvasObject{fa.emptyPane()}
		fa.chatPane.Refresh()
	})
}

func railItem(path string, name string, active bool, badge int, onTap func()) fyne.CanvasObject {
	fg := theme.WaRailIcon
	if active {
		fg = theme.WaText
	}
	icon := utils.IconImage(name, path, 24, fg)
	cell := container.NewMax(container.NewCenter(icon))
	if active {
		bg := canvas.NewRectangle(theme.WaRailSel)
		bg.CornerRadius = 999
		cell = container.NewMax(bg, container.NewCenter(icon))
	}
	wrap := container.NewGridWrap(fyne.NewSize(44, 44), cell)
	if badge > 0 {
		btxt := strconv.Itoa(badge)
		if badge > 99 {
			btxt = "99+"
		}
		bt := canvas.NewText(btxt, theme.WaWhite)
		bt.TextSize = 9
		bt.TextStyle = fyne.TextStyle{Bold: true}
		bt.Alignment = fyne.TextAlignCenter
		bbg := canvas.NewCircle(theme.WaGreen)
		badgeObj := container.NewGridWrap(fyne.NewSize(18, 18), container.NewMax(bbg, container.NewCenter(bt)))
		wrap = container.NewMax(wrap, container.NewWithoutLayout(badgeObj))
	}
	if onTap != nil {
		ta := chat.NewTapArea(onTap)
		wrap = container.NewMax(wrap, ta)
	}
	return wrap
}

func railAvatar() fyne.CanvasObject {
	bg := canvas.NewCircle(theme.WaHover)
	icon := utils.IconImage("rail-avatar", utils.IconPerson, 26, theme.WaRailIcon)
	return container.NewGridWrap(fyne.NewSize(34, 34), container.NewMax(bg, container.NewCenter(icon)))
}

func (fa *FyneApp) rail() fyne.CanvasObject {
	railBg := canvas.NewRectangle(theme.WaRail)
	top := container.NewVBox(
		utils.Spacer(1, 10),
		container.NewCenter(railItem(utils.IconChats, "rail-chats", true, 0, nil)),
		utils.Spacer(1, 2),
		container.NewCenter(railItem(utils.IconStatus, "rail-status", false, 0, nil)),
		utils.Spacer(1, 2),
		container.NewCenter(railItem(utils.IconChannels, "rail-channels", false, 0, nil)),
		utils.Spacer(1, 2),
		container.NewCenter(railItem(utils.IconCommunities, "rail-communities", false, 0, nil)),
	)
	bottom := container.NewVBox(
		container.NewCenter(railItem(utils.IconSettings, "rail-settings", false, 0, func() { fa.OpenSettings() })),
		utils.Spacer(1, 10),
		container.NewCenter(railAvatar()),
		utils.Spacer(1, 12),
	)
	inner := container.NewBorder(top, bottom, nil, nil, canvas.NewRectangle(color.Transparent))
	return container.NewMax(railBg, inner)
}

func (fa *FyneApp) emptyPane() fyne.CanvasObject {
	bg := chat.ChatBackground(fyne.NewSize(0, 0))
	logo := canvas.NewCircle(theme.WaPanel2)
	logoTxt := canvas.NewText("WA", theme.WaAccent)
	logoTxt.TextSize = 24
	logoTxt.TextStyle = fyne.TextStyle{Bold: true}
	logoTxt.Alignment = fyne.TextAlignCenter
	logoObj := container.NewGridWrap(fyne.NewSize(84, 84), container.NewMax(logo, container.NewCenter(logoTxt)))

	title := canvas.NewText("WhatsApp for Linux", theme.WaText)
	title.TextSize = 32
	title.Alignment = fyne.TextAlignCenter
	desc := canvas.NewText("Send and receive messages without keeping your phone online.", theme.WaMuted)
	desc.TextSize = 14
	desc.Alignment = fyne.TextAlignCenter
	desc2 := canvas.NewText("Use WhatsApp on up to 4 linked devices and 1 phone.", theme.WaMuted)
	desc2.TextSize = 14
	desc2.Alignment = fyne.TextAlignCenter

	download := widget.NewButton("Download", func() {})
	download.Importance = widget.SuccessImportance

	lock := canvas.NewText("🔒  Your personal messages are end-to-end encrypted", theme.WaMuted)
	lock.TextSize = 13
	lock.Alignment = fyne.TextAlignCenter

	accent := canvas.NewRectangle(theme.WaAccent)
	accent.SetMinSize(fyne.NewSize(0, 6))

	content := container.NewVBox(
		container.NewCenter(logoObj),
		utils.Spacer(1, 24),
		container.NewCenter(title),
		utils.Spacer(1, 14),
		container.NewCenter(desc),
		utils.Spacer(1, 4),
		container.NewCenter(desc2),
		utils.Spacer(1, 28),
		container.NewCenter(container.NewGridWrap(fyne.NewSize(132, 38), download)),
		utils.Spacer(1, 70),
		container.NewCenter(lock),
	)
	return container.NewBorder(nil, accent, nil, nil, container.NewMax(bg, container.NewCenter(container.NewPadded(content))))
}

func (fa *FyneApp) buildShell() {
	if fa.sidebar == nil {
		fa.sidebar = sidebar.NewSidebar(fa)
	} else {
		go fa.sidebar.Reload()
	}
	if fa.chatPane == nil {
		fa.chatPane = container.NewMax(fa.emptyPane())
	}

	leftBody := container.NewMax(canvas.NewRectangle(theme.WaSidebarBg), fa.sidebar.Content())
	right := container.NewMax(canvas.NewRectangle(theme.WaChatBg), fa.chatPane)

	split := container.New(&mainSplitLayout{leftRatio: 0.30, minLeft: 320, maxLeft: 600, minRight: 400}, leftBody, right)
	fa.shell = container.NewMax(canvas.NewRectangle(theme.WaChatBg), split)
}

func (fa *FyneApp) SwitchToChats() {
	fa.State().SetScreen("chats")
	fa.buildShell()
	fyne.Do(func() {
		fa.stack.Objects = []fyne.CanvasObject{fa.shell}
		fa.stack.Refresh()
	})
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
	line := canvas.NewRectangle(theme.WaDivider)
	line.SetMinSize(fyne.NewSize(1, 0))
	return line
}

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

func (fa *FyneApp) OpenSettings() {
	fyne.Do(func() {
		scr := settings.NewSettingsScreen(func() { fa.SwitchToChats() })
		fa.stack.Objects = []fyne.CanvasObject{scr}
		fa.stack.Refresh()
	})
}
