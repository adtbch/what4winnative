package fyneui

import (
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/lugvitc/whats4linux/api"
)

type Sidebar struct {
	fa    *FyneApp
	list  *widget.List
	chats []api.ChatElement
	root  fyne.CanvasObject
}

func pill(text string, active bool) *fyne.Container {
	col := waMuted
	if active { col = waText }
	lbl := canvas.NewText(text, col)
	lbl.TextSize = 13
	bg := canvas.NewRectangle(waPill)
	if active { bg.FillColor = waPillOn }
	bg.CornerRadius = 999
	return container.NewMax(bg, container.NewPadded(container.NewHBox(lbl)))
}

func NewSidebar(fa *FyneApp) *Sidebar {
	s := &Sidebar{fa: fa}

	s.list = widget.NewList(
		func() int { return len(s.chats) },
		func() fyne.CanvasObject {
			av := canvas.NewCircle(waHover)
			
			name := canvas.NewText("name", waText)
			name.TextSize = 15
			name.TextStyle = fyne.TextStyle{Bold:true}
			t := canvas.NewText("00:00", waMuted)
			t.TextSize = 12
			top := container.NewBorder(nil,nil,nil,t,name)
			prev := canvas.NewText("preview", waMuted)
			prev.TextSize = 13
			left := container.NewVBox(top, prev)
			avWrap := container.NewGridWrap(fyne.NewSize(48,48), av)
			row := container.NewBorder(nil,nil,container.NewPadded(avWrap),container.NewPadded(left))
			return container.NewPadded(row)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < 0 || id >= len(s.chats) { return }
			c := s.chats[id]
			name := c.FullName
			if strings.TrimSpace(name)=="" { name=c.PushName }
			if strings.TrimSpace(name)=="" { name=c.Short }
			if strings.TrimSpace(name)=="" { name=c.JID }
			// crude update
			pad := obj.(*fyne.Container)
			row := pad.Objects[0].(*fyne.Container)
			left := row.Objects[1].(*fyne.Container)
			top := left.Objects[0].(*fyne.Container)
			nameTxt := top.Objects[0].(*canvas.Text)
			timeTxt := top.Objects[2].(*canvas.Text)
			prevTxt := left.Objects[1].(*canvas.Text)
			nameTxt.Text = name
			nameTxt.Refresh()
			prevTxt.Text = c.LatestMessage
			prevTxt.Refresh()
			timeTxt.Text = ""
			timeTxt.Refresh()
		},
	)
	s.list.OnSelected = func(id widget.ListItemID) {
		if id < 0 || id >= len(s.chats) { return }
		jid := s.chats[id].JID
		s.fa.OpenChat(jid)
	}

	title := canvas.NewText("Chats", waText)
	title.TextSize = 18
	title.TextStyle = fyne.TextStyle{Bold:true}
	header := container.NewBorder(nil,nil,nil,nil,container.NewPadded(title))

	search := widget.NewEntry()
	search.SetPlaceHolder("Search or start new chat")
	searchBg := canvas.NewRectangle(waPanel2)
	searchBg.CornerRadius = 8
	searchWrap := container.NewMax(searchBg, container.NewPadded(search))

	filters := container.NewHBox(
		pill("All", true), pill("Unread 79", false), pill("Favourites", false), pill("Groups 49", false), pill("Archived", false),
	)
	filtersWrap := container.NewHBox(filters, container.NewHBox()) // allow scroll later

	sync := canvas.NewText("Syncing chats…", waMuted)
	sync.TextSize = 12
	syncBar := container.NewCenter(container.NewPadded(sync))

	top := container.NewVBox(header, searchWrap, container.NewPadded(filtersWrap))
	s.root = container.NewBorder(top, syncBar, nil, nil, s.list)

	go s.loadChats()
	go s.listenRefresh()
	return s
}

func (s *Sidebar) loadChats() {
	chats, err := s.fa.api.GetChatList()
	if err != nil { log.Println("sidebar:", err); return }
	fyne.Do(func() { s.chats = chats; s.list.Refresh() })
}

func (s *Sidebar) listenRefresh() {
	ch := SubscribeEvent("wa:chat_list_refresh")
	for {
		select {
		case <-s.fa.ctx.Done(): return
		case <-ch: s.loadChats()
		}
	}
}

func (s *Sidebar) Content() fyne.CanvasObject { return s.root }
