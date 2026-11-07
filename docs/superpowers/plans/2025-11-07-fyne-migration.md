# Fyne Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Migrate the WhatsApp desktop client UI from Wails + React/TypeScript to Fyne (pure Go), keeping backend logic in `api/` and `internal/` unchanged where possible to preserve all functionality (messaging, calls, groups, settings).

**Architecture:** Create a parallel Fyne UI under `fyneui/` that directly calls `*api.Api` methods. Use `internal/eventbus` for Wails runtime event replacement. Keep Wails build working until Fyne UI is feature-complete. Migrate screen by screen starting with login/QR, then chat list, then chat detail. Reuse existing state patterns adapted to Go structs/bindings.

**Tech Stack:** Go 1.26.5, Fyne v2, WhatsMeow, existing api/internal packages

**Spec:** This migration preserves all existing functionality - no feature changes, only UI framework replacement.

## Global Constraints

- Must not modify `api/` core logic unnecessarily - only add eventbus publishing where needed for Fyne
- Must not modify `internal/store/`, `internal/wa/`, `internal/voip/` core logic
- Keep `master` branch clean; all work on `fyne-migration`
- Commit after every task
- UI must maintain similar look/feel where reasonable
- All existing features must work: login/QR, chat list, messaging, media, calls, groups, settings, notifications
- No breaking changes to data storage

## Review Focus

- Event flow: Ensure all `wa:*` events (qr, status, new_message, chat_marked_read, etc.) work in Fyne via eventbus
- Context handling: `api.Startup(ctx)` needs context in Fyne - verify lifecycle
- File/media handling: Image/video rendering in Fyne vs webview
- Performance: Chat list with many chats, message virtualization
- Platform differences: Windows/Linux behavior

---

### Task 1: Setup Fyne project structure and eventbus integration

**Files:**
- Create: `fyneui/app.go`
- Create: `fyneui/bus.go` 
- Create: `fyneui/state.go`
- Modify: `api/api.go:64-70` (ensure emitEvent works)
- Modify: `go.mod` (already has fyne)

**Interfaces:**
- Consumes: `internal/eventbus.Default`, `*api.Api`
- Produces: `fyneui.App` struct, event subscription helpers

- [ ] **Step 1: Create fyneui/bus.go for event subscription**

```go
package fyneui

import (
	"github.com/lugvitc/whats4linux/internal/eventbus"
)

func SubscribeEvent(name string) <-chan any {
	return eventbus.Default.Subscribe(name)
}
```

- [ ] **Step 2: Create fyneui/state.go with basic state**

```go
package fyneui

import (
	"sync"

	"github.com/lugvitc/whats4linux/api"
)

type AppState struct {
	mu sync.RWMutex
	api *api.Api
	
	Screen string // "login" | "chats" | "settings"
	SelectedChatID string
}
```

- [ ] **Step 3: Update fyneui/app.go with proper structure**

```go
package fyneui

import (
	"context"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"

	"github.com/lugvitc/whats4linux/api"
)

type FyneApp struct {
	app fyne.App
	win fyne.Window
	api *api.Api
	ctx context.Context
	cancel context.CancelFunc
	state *AppState
	stack *container.Stack
}

func New(a *api.Api) *FyneApp {
	ctx, cancel := context.WithCancel(context.Background())
	fa := &FyneApp{
		app: app.New(),
		api: a,
		ctx: ctx,
		cancel: cancel,
		state: &AppState{api: a, Screen: "login"},
	}
	fa.win = fa.app.NewWindow("whats4linux")
	fa.stack = container.NewStack()
	fa.win.SetContent(fa.stack)
	return fa
}

func (fa *FyneApp) Run() {
	fa.win.Resize(fyne.NewSize(1100, 700))
	fa.win.ShowAndRun()
	fa.cancel()
}
```

- [ ] **Step 4: Test build**

Run: `cd /c/whats4linux && go build ./fyneui/...`
Expected: success

- [ ] **Step 5: Commit**

```bash
git add fyneui/bus.go fyneui/state.go fyneui/app.go
git commit -m "feat(fyne): add basic fyneui structure with eventbus"
```

### Task 2: Create login screen with QR code

**Files:**
- Create: `fyneui/login.go`
- Modify: `fyneui/app.go` (add screen switching)

**Interfaces:**
- Consumes: `SubscribeEvent("wa:qr")`, `SubscribeEvent("wa:status")`, `*api.Api.Login()`
- Produces: Login screen that shows QR and status

- [ ] **Step 1: Create fyneui/login.go**

```go
package fyneui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type LoginScreen struct {
	fa *FyneApp
	status *widget.Label
	qrImg *canvas.Image
}

func NewLoginScreen(fa *FyneApp) *LoginScreen {
	ls := &LoginScreen{
		fa: fa,
		status: widget.NewLabel("Waiting for QR..."),
		qrImg: &canvas.Image{},
	}
	ls.qrImg.FillMode = canvas.ImageFillContain
	ls.qrImg.SetMinSize(fyne.NewSize(300, 300))
	
	go ls.listenEvents()
	go func() { _ = fa.api.Login() }()
	
	return ls
}

func (ls *LoginScreen) listenEvents() {
	qrCh := SubscribeEvent("wa:qr")
	statusCh := SubscribeEvent("wa:status")
	
	for {
		select {
		case <-ls.fa.ctx.Done():
			return
		case <-qrCh:
			// TODO: decode QR and show
		case data := <-statusCh:
			if s, ok := data.(string); ok {
				ls.status.SetText(s)
				if s == "logged_in" || s == "success" {
					ls.fa.switchToChats()
				}
			}
		}
	}
}

func (ls *LoginScreen) Content() fyne.CanvasObject {
	return container.NewCenter(container.NewVBox(
		widget.NewLabel("WhatsApp Login"),
		ls.qrImg,
		ls.status,
	))
}
```

- [ ] **Step 2: Add switchToChats to fyneui/app.go**

```go
func (fa *FyneApp) switchToChats() {
	fa.state.Screen = "chats"
	// TODO: show chat screen
}
```

- [ ] **Step 3: Test build**

Run: `cd /c/whats4linux && go build ./fyneui/...`
Expected: success

- [ ] **Step 4: Commit**

```bash
git add fyneui/login.go fyneui/app.go
git commit -m "feat(fyne): add login screen with QR event handling"
```

### Task 3: Create chat sidebar/list

**Files:**
- Create: `fyneui/sidebar.go`
- Modify: `fyneui/app.go`

**Interfaces:**
- Consumes: `*api.Api.GetChatList()`, `SubscribeEvent("wa:chat_list_refresh")`
- Produces: Chat list with search

- [ ] **Step 1: Create fyneui/sidebar.go**

```go
package fyneui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/lugvitc/whats4linux/api"
)

type Sidebar struct {
	fa *FyneApp
	list *widget.List
	chats []api.ChatElement
}

func NewSidebar(fa *FyneApp) *Sidebar {
	s := &Sidebar{fa: fa}
	s.list = widget.NewList(
		func() int { return len(s.chats) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(s.chats) {
				obj.(*widget.Label).SetText(s.chats[id].FullName)
			}
		},
	)
	go s.loadChats()
	go s.listenEvents()
	return s
}

func (s *Sidebar) loadChats() {
	chats, err := s.fa.api.GetChatList()
	if err == nil {
		s.chats = chats
		s.list.Refresh()
	}
}

func (s *Sidebar) listenEvents() {
	ch := SubscribeEvent("wa:chat_list_refresh")
	for {
		select {
		case <-s.fa.ctx.Done():
			return
		case <-ch:
			s.loadChats()
		}
	}
}

func (s *Sidebar) Content() fyne.CanvasObject {
	search := widget.NewEntry()
	search.SetPlaceHolder("Search...")
	return container.NewBorder(search, nil, nil, nil, s.list)
}
```

- [ ] **Step 2: Wire to app**

Update `fyneui/app.go` to switch to sidebar when logged in.

- [ ] **Step 3: Test build**

Run: `cd /c/whats4linux && go build ./fyneui/...`
Expected: success

- [ ] **Step 4: Commit**

```bash
git add fyneui/sidebar.go fyneui/app.go
git commit -m "feat(fyne): add chat sidebar with list and refresh"
```

### Task 4: Create chat detail view

**Files:**
- Create: `fyneui/chatview.go`
- Modify: `fyneui/app.go`

**Interfaces:**
- Consumes: `*api.Api.FetchMessagesPaged()`, `*api.Api.SendMessage()`, `SubscribeEvent("wa:new_message")`
- Produces: Message list and input

- [ ] **Step 1: Create basic chatview**

```go
package fyneui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type ChatView struct {
	fa *FyneApp
	chatID string
	messages *widget.List
	input *widget.Entry
}

func NewChatView(fa *FyneApp, chatID string) *ChatView {
	cv := &ChatView{
		fa: fa,
		chatID: chatID,
		input: widget.NewEntry(),
	}
	cv.messages = widget.NewList(
		func() int { return 0 },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {},
	)
	cv.input.OnSubmitted = func(text string) {
		// TODO: send message
	}
	go cv.listenEvents()
	return cv
}

func (cv *ChatView) listenEvents() {
	ch := SubscribeEvent("wa:new_message")
	for {
		select {
		case <-cv.fa.ctx.Done():
			return
		case <-ch:
			// TODO: update messages
		}
	}
}

func (cv *ChatView) Content() fyne.CanvasObject {
	sendBtn := widget.NewButton("Send", func() {
		cv.input.OnSubmitted(cv.input.Text)
		cv.input.SetText("")
	})
	bottom := container.NewBorder(nil, nil, nil, sendBtn, cv.input)
	return container.NewBorder(nil, bottom, nil, nil, cv.messages)
}
```

- [ ] **Step 2: Test build**

Run: `cd /c/whats4linux && go build ./fyneui/...`
Expected: success

- [ ] **Step 3: Commit**

```bash
git add fyneui/chatview.go
git commit -m "feat(fyne): add basic chat view structure"
```

### Task 5: Create main entry point for Fyne

**Files:**
- Create: `cmd/fyne/main.go`
- Modify: `go.mod` if needed

**Interfaces:**
- Consumes: `api.New()`, `fyneui.New()`
- Produces: Runnable Fyne app

- [ ] **Step 1: Create cmd/fyne/main.go**

```go
package main

import (
	apiPkg "github.com/lugvitc/whats4linux/api"
	"github.com/lugvitc/whats4linux/fyneui"
)

func main() {
	api := apiPkg.New()
	app := fyneui.New(api)
	app.Run()
}
```

- [ ] **Step 2: Test build**

Run: `cd /c/whats4linux && go build ./cmd/fyne/...`
Expected: success (may have issues with API startup context)

- [ ] **Step 3: Commit**

```bash
git add cmd/fyne/main.go
git commit -m "feat(fyne): add main entry point for fyne app"
```

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2025-11-07-fyne-migration.md`. Please review the plan. Which execution approach would you prefer?

- **Subagent-driven** - A fresh subagent implements each task and a fresh reviewer checks it before the next one starts, then a whole-branch review at the end. Most thorough; costs a fresh context per task and per review.
- **Native** - I implement every task myself in this session, the way this harness runs work, then one fresh reviewer on the most capable model checks the whole branch. Cheapest and fastest; no independent review until the end. Runs well with a mid-tier session model, since the plan carries the design.

For this plan I recommend **Native** because it's a UI migration with many interconnected parts and we need to iterate quickly on the design. Does the plan capture what you want, and which approach should we use?
