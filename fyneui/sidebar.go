package fyneui

import (
	"log"

	"fyne.io/fyne/v2"
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

func NewSidebar(fa *FyneApp) *Sidebar {
	s := &Sidebar{fa: fa}
	s.list = widget.NewList(
		func() int { return len(s.chats) },
		func() fyne.CanvasObject {
			name := widget.NewLabel("name")
			name.TextStyle = fyne.TextStyle{Bold: true}
			prev := widget.NewLabel("preview")
			prev.Wrapping = fyne.TextTruncate
			return container.NewVBox(name, prev)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < 0 || id >= len(s.chats) {
				return
			}
			c := s.chats[id]
			name := c.FullName
			if name == "" {
				name = c.PushName
			}
			if name == "" {
				name = c.Short
			}
			if name == "" {
				name = c.JID
			}
			box := obj.(*fyne.Container)
			box.Objects[0].(*widget.Label).SetText(name)
			box.Objects[1].(*widget.Label).SetText(c.LatestMessage)
		},
	)
	s.list.OnSelected = func(id widget.ListItemID) {
		if id < 0 || id >= len(s.chats) {
			return
		}
		jid := s.chats[id].JID
		s.fa.state.SelectedChatID = jid
		s.fa.OpenChat(jid)
	}

	search := widget.NewEntry()
	search.SetPlaceHolder("Search (WIP)")
	s.root = container.NewBorder(search, nil, nil, nil, s.list)

	go s.loadChats()
	go s.listenRefresh()
	return s
}

func (s *Sidebar) loadChats() {
	chats, err := s.fa.api.GetChatList()
	if err != nil {
		log.Println("sidebar: GetChatList:", err)
		return
	}
	fyne.Do(func() {
		s.chats = chats
		s.list.Refresh()
	})
}

func (s *Sidebar) listenRefresh() {
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
	return s.root
}
