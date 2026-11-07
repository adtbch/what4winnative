package fyneui

import (
	"fmt"
	"image/color"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	"github.com/lugvitc/whats4linux/api"
)

type chatFilter string

const (
	filterAll        chatFilter = "all"
	filterUnread     chatFilter = "unread"
	filterFavourites chatFilter = "favourites"
	filterGroups     chatFilter = "groups"
	filterArchived   chatFilter = "archived"
)

type tappablePill struct {
	widget.BaseWidget
	bg    *canvas.Rectangle
	label *canvas.Text
	onTap func()
}

func newTappablePill(text string, active bool, onTap func()) *tappablePill {
	p := &tappablePill{onTap: onTap}
	p.ExtendBaseWidget(p)
	p.bg = canvas.NewRectangle(pillBg(active))
	p.bg.CornerRadius = 999
	p.label = canvas.NewText(text, pillFg(active))
	p.label.TextSize = 13
	p.label.TextStyle = fyne.TextStyle{Bold: active}
	return p
}

func pillBg(active bool) color.Color {
	if active {
		return waPillOn
	}
	return waPill
}
func pillFg(active bool) color.Color {
	if active {
		return waAccent
	}
	return waMuted
}

func (p *tappablePill) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewMax(p.bg, container.NewPadded(container.NewCenter(p.label))))
}
func (p *tappablePill) Tapped(*fyne.PointEvent) {
	if p.onTap != nil {
		p.onTap()
	}
}
func (p *tappablePill) Cursor() desktop.Cursor { return desktop.PointerCursor }
func (p *tappablePill) SetState(text string, active bool) {
	p.label.Text = text
	p.label.Color = pillFg(active)
	p.label.TextStyle = fyne.TextStyle{Bold: active}
	p.bg.FillColor = pillBg(active)
	p.label.Refresh()
	p.bg.Refresh()
}

type Sidebar struct {
	fa        *FyneApp
	mu        sync.RWMutex
	allChats  []api.ChatElement
	visible   []api.ChatElement
	unread    map[string]int
	favourite map[string]bool // real data source not in backend; kept for real-count/filter (currently none)
	filter    chatFilter
	search    string
	list      *widget.List
	root      fyne.CanvasObject
	syncLbl   *canvas.Text
	pills     map[chatFilter]*tappablePill
	pillBox   *fyne.Container
	searchEnt *widget.Entry
}

func NewSidebar(fa *FyneApp) *Sidebar {
	s := &Sidebar{
		fa:        fa,
		unread:    make(map[string]int),
		favourite: make(map[string]bool),
		filter:    filterAll,
		pills:     make(map[chatFilter]*tappablePill),
	}

	s.list = widget.NewList(
		func() int {
			s.mu.RLock()
			defer s.mu.RUnlock()
			return len(s.visible)
		},
		func() fyne.CanvasObject { return s.newRow() },
		func(id widget.ListItemID, obj fyne.CanvasObject) { s.updateRow(id, obj) },
	)
	s.list.OnSelected = func(id widget.ListItemID) {
		s.mu.RLock()
		if id < 0 || id >= len(s.visible) {
			s.mu.RUnlock()
			return
		}
		jid := s.visible[id].JID
		s.mu.RUnlock()
		s.mu.Lock()
		delete(s.unread, jid)
		s.mu.Unlock()
		s.fa.OpenChat(jid)
		s.refreshVisible()
	}

	title := canvas.NewText("WhatsApp", waText)
	title.TextSize = 22
	title.TextStyle = fyne.TextStyle{Bold: true}
	newChat := canvas.NewText("\u2295", waMuted)
	newChat.TextSize = 18
	menu := canvas.NewText("\u22EE", waMuted)
	menu.TextSize = 18
	header := container.NewPadded(container.NewBorder(nil, nil, nil, container.NewHBox(newChat, spacer(16, 0), menu), title))

	s.searchEnt = widget.NewEntry()
	s.searchEnt.SetPlaceHolder("Search or start a new chat")
	s.searchEnt.OnChanged = func(v string) {
		s.mu.Lock()
		s.search = v
		s.mu.Unlock()
		s.refreshVisible()
	}
	searchBg := canvas.NewRectangle(waPanel2)
	searchBg.CornerRadius = 8
	searchIcon := canvas.NewText("\U0001F50D", waMuted)
	searchIcon.TextSize = 13
	searchWrap := container.NewMax(searchBg, container.NewPadded(container.NewBorder(nil, nil, container.NewPadded(searchIcon), nil, s.searchEnt)))

	s.pillBox = container.NewHBox()
	s.rebuildPills()
	filtersWrap := container.NewHScroll(container.NewPadded(s.pillBox))
	filtersWrap.SetMinSize(fyne.NewSize(0, 38))

	s.syncLbl = canvas.NewText("Syncing chats\u2026", waMuted)
	s.syncLbl.TextSize = 12
	syncBar := container.NewCenter(container.NewPadded(s.syncLbl))

	top := container.NewVBox(header, searchWrap, filtersWrap)
	s.root = container.NewBorder(top, syncBar, nil, nil, s.list)

	go s.loadChats()
	go s.listenRefresh()
	go func() {
		for _, d := range []int{1, 3, 6, 12} {
			select { case <-s.fa.ctx.Done(): return; case <-time.After(time.Duration(d)*time.Second): }
			s.loadChats()
		}
	}()
	go s.listenIncoming()
	go s.listenReadState()
	return s
}

func (s *Sidebar) rebuildPills() {
	fyne.Do(func() { s.rebuildPillsUI() })
}
func (s *Sidebar) rebuildPillsUI() {
	type def struct {
		f chatFilter
		l string
	}
	defs := []def{
		{filterAll, "All"},
		{filterUnread, "Unread"},
		{filterFavourites, "Favourites"},
		{filterGroups, "Groups"},
		{filterArchived, "Archived"},
	}
	objs := make([]fyne.CanvasObject, 0, len(defs))
	for _, d := range defs {
		df := d.f
		p := newTappablePill(s.pillText(df), s.filter == df, func() {
			s.mu.Lock()
			s.filter = df
			s.mu.Unlock()
			s.rebuildPills()
			s.refreshVisible()
		})
		s.pills[df] = p
		objs = append(objs, p)
	}
	s.pillBox.Objects = objs
	s.pillBox.Refresh()
}

func (s *Sidebar) pillText(f chatFilter) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	unread, groups, archived, fav := s.countsLocked()
	switch f {
	case filterUnread:
		if unread > 0 {
			return fmt.Sprintf("Unread %d", unread)
		}
		return "Unread"
	case filterFavourites:
		if fav > 0 {
			return fmt.Sprintf("Favourites %d", fav)
		}
		return "Favourites"
	case filterGroups:
		if groups > 0 {
			return fmt.Sprintf("Groups %d", groups)
		}
		return "Groups"
	case filterArchived:
		if archived > 0 {
			return fmt.Sprintf("Archived %d", archived)
		}
		return "Archived"
	default:
		return "All"
	}
}

func (s *Sidebar) countsLocked() (unread, groups, archived, fav int) {
	for _, c := range s.allChats {
		if c.Archived {
			archived++
		}
		if isGroupJID(c.JID) && !c.Archived {
			groups++
		}
		if s.favourite[c.JID] && !c.Archived {
			fav++
		}
		if n := s.unread[c.JID]; n > 0 && !c.Archived {
			unread += n
		}
	}
	return
}

func (s *Sidebar) newRow() fyne.CanvasObject {
	bg := canvas.NewRectangle(color.Transparent)
	avatarBg := canvas.NewCircle(waHover)
	avatarTxt := canvas.NewText("?", waText)
	avatarTxt.Alignment = fyne.TextAlignCenter
	avatarTxt.TextSize = 14
	avatarTxt.TextStyle = fyne.TextStyle{Bold: true}
	avatar := container.NewMax(avatarBg, container.NewCenter(avatarTxt))

	name := canvas.NewText("name", waText)
	name.TextSize = 15
	name.TextStyle = fyne.TextStyle{Bold: true}
	timeTxt := canvas.NewText("", waMuted)
	timeTxt.TextSize = 12
	timeTxt.Alignment = fyne.TextAlignTrailing
	top := container.NewBorder(nil, nil, nil, timeTxt, name)

	preview := canvas.NewText("preview", waMuted)
	preview.TextSize = 13
	pin := canvas.NewText("", waMuted)
	pin.TextSize = 12
	badgeBg := canvas.NewRectangle(waGreen)
	badgeBg.CornerRadius = 999
	badgeTxt := canvas.NewText("", waPanel)
	badgeTxt.TextSize = 11
	badgeTxt.TextStyle = fyne.TextStyle{Bold: true}
	badgeTxt.Alignment = fyne.TextAlignCenter
	badgeWrap := container.NewMax(badgeBg, container.NewPadded(container.NewCenter(badgeTxt)))
	badgeWrap.Hide()
	right := container.NewHBox(pin, spacer(6, 0), badgeWrap)
	bottom := container.NewBorder(nil, nil, nil, right, preview)

	mid := container.NewVBox(top, bottom)
	row := container.NewBorder(nil, nil, container.NewPadded(container.NewGridWrap(fyne.NewSize(49, 49), avatar)), nil, container.NewPadded(mid))
	padded := container.NewPadded(row)
	return container.NewMax(bg, padded)
}

func (s *Sidebar) updateRow(id widget.ListItemID, obj fyne.CanvasObject) {
	s.mu.RLock()
	if id < 0 || id >= len(s.visible) {
		s.mu.RUnlock()
		return
	}
	c := s.visible[id]
	unread := s.unread[c.JID]
	s.mu.RUnlock()

	root := obj.(*fyne.Container)
	padded := root.Objects[1].(*fyne.Container)
	row := padded.Objects[0].(*fyne.Container)
	avatarMax := row.Objects[1].(*fyne.Container).Objects[0].(*fyne.Container)
	avatarStack := avatarMax.Objects[0].(*fyne.Container)
	avatarBg := avatarStack.Objects[0].(*canvas.Circle)
	avatarCenter := avatarStack.Objects[1].(*fyne.Container)
	avatarTxt := avatarCenter.Objects[0].(*canvas.Text)

	midPadded := row.Objects[0].(*fyne.Container)
	mid := midPadded.Objects[0].(*fyne.Container)
	top := mid.Objects[0].(*fyne.Container)
	nameTxt := top.Objects[0].(*canvas.Text)
	timeTxt := top.Objects[2].(*canvas.Text)
	bottom := mid.Objects[1].(*fyne.Container)
	prevTxt := bottom.Objects[0].(*canvas.Text)
	rightBox := bottom.Objects[2].(*fyne.Container)
	pinTxt := rightBox.Objects[0].(*canvas.Text)
	badgeWrap := rightBox.Objects[2].(*fyne.Container)
	badgeBg := badgeWrap.Objects[0].(*canvas.Rectangle)
	badgePad := badgeWrap.Objects[1].(*fyne.Container)
	badgeCenter := badgePad.Objects[0].(*fyne.Container)
	badgeTxt := badgeCenter.Objects[0].(*canvas.Text)

	name := chatDisplayName(c)
	nameTxt.Text = truncateRunes(name, 28)
	nameTxt.Refresh()

	prev := c.LatestMessage
	if c.Sender != "" && isGroupJID(c.JID) {
		prev = shortSender(c.Sender) + ": " + prev
	}
	prevTxt.Text = truncateRunes(prev, 42)
	prevTxt.Refresh()

	timeTxt.Text = formatChatTimestamp(c.LatestTS)
	if unread > 0 {
		timeTxt.Color = waGreen
	} else {
		timeTxt.Color = waMuted
	}
	timeTxt.Refresh()

	if c.Pinned {
		pinTxt.Text = "\U0001F4CC"
	} else if c.Archived && s.filter == filterArchived {
		pinTxt.Text = "\U0001F5C4"
	} else {
		pinTxt.Text = ""
	}
	pinTxt.Refresh()

	if unread > 0 {
		if unread > 99 {
			badgeTxt.Text = "99+"
		} else {
			badgeTxt.Text = fmt.Sprintf("%d", unread)
		}
		badgeBg.FillColor = waGreen
		badgeWrap.Show()
		badgeTxt.Refresh()
		badgeBg.Refresh()
		badgeWrap.Refresh()
	} else {
		badgeWrap.Hide()
	}

	avatarBg.FillColor = avatarColor(c.JID)
	avatarTxt.Text = sidebarAvatarInitial(name, c.JID)
	avatarBg.Refresh()
	avatarTxt.Refresh()
}

func (s *Sidebar) loadChats() {
	chats, err := s.fa.api.GetChatList()
	if err != nil {
		log.Println("sidebar GetChatList:", err)
		return
	}
	s.mu.Lock()
	s.allChats = chats
	s.mu.Unlock()
	s.refreshVisible()
	fyne.Do(func() {
		if s.syncLbl != nil {
			s.syncLbl.Text = ""
			s.syncLbl.Refresh()
		}
	})
}

func (s *Sidebar) refreshVisible() {
	s.mu.Lock()
	s.visible = s.filterAndSortLocked()
	s.mu.Unlock()
	s.rebuildPills()
	fyne.Do(func() {
		if s.list != nil {
			s.list.Refresh()
		}
	})
}

func (s *Sidebar) filterAndSortLocked() []api.ChatElement {
	term := strings.ToLower(strings.TrimSpace(s.search))
	out := make([]api.ChatElement, 0, len(s.allChats))
	for _, c := range s.allChats {
		if !s.matchesFilterLocked(c) {
			continue
		}
		if term != "" {
			name := strings.ToLower(chatDisplayName(c))
			if !strings.Contains(name, term) && !strings.Contains(strings.ToLower(c.LatestMessage), term) {
				continue
			}
		}
		out = append(out, c)
	}
	// Sort: pinned first, then LatestTS desc (matches chatSort.ts)
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.Pinned != b.Pinned {
			return a.Pinned
		}
		return a.LatestTS > b.LatestTS
	})
	return out
}

func (s *Sidebar) matchesFilterLocked(c api.ChatElement) bool {
	switch s.filter {
	case filterArchived:
		return c.Archived
	case filterUnread:
		if c.Archived {
			return false
		}
		return s.unread[c.JID] > 0
	case filterGroups:
		if c.Archived {
			return false
		}
		return isGroupJID(c.JID)
	case filterFavourites:
		if c.Archived {
			return false
		}
		return s.favourite[c.JID]
	default: // all
		return !c.Archived
	}
}

func (s *Sidebar) listenRefresh() {
	ch := SubscribeEvent("wa:chat_list_refresh")
	st := SubscribeEvent("wa:status")
	for {
		select {
		case <-s.fa.ctx.Done():
			return
		case <-ch:
			s.loadChats()
		case data := <-st:
			if v, ok := data.(string); ok && (v == "logged_in" || v == "success" || v == "connected" || v == "synced") {
				s.loadChats()
			}
		}
	}
}

func (s *Sidebar) listenIncoming() {
	ch := SubscribeEvent("wa:new_message")
	for {
		select {
		case <-s.fa.ctx.Done():
			return
		case data := <-ch:
			m, ok := data.(map[string]any)
			if !ok {
				continue
			}
			chatID, _ := m["chatId"].(string)
			isFromMe, _ := m["isFromMe"].(bool)
			if chatID == "" || isFromMe {
				continue
			}
			s.mu.RLock()
			selected := s.fa.state.SelectedChatID
			s.mu.RUnlock()
			if selected == chatID {
				continue
			}
			s.mu.Lock()
			s.unread[chatID]++
			s.mu.Unlock()
			s.refreshVisible()
		}
	}
}

func (s *Sidebar) listenReadState() {
	readCh := SubscribeEvent("wa:chat_marked_read")
	unreadCh := SubscribeEvent("wa:chat_marked_unread")
	for {
		select {
		case <-s.fa.ctx.Done():
			return
		case data := <-readCh:
			if m, ok := data.(map[string]any); ok {
				if chatID, _ := m["chatId"].(string); chatID != "" {
					s.mu.Lock()
					delete(s.unread, chatID)
					s.mu.Unlock()
					s.refreshVisible()
				}
			}
		case data := <-unreadCh:
			if m, ok := data.(map[string]any); ok {
				if chatID, _ := m["chatId"].(string); chatID != "" {
					s.mu.Lock()
					if s.unread[chatID] == 0 {
						s.unread[chatID] = 1
					}
					s.mu.Unlock()
					s.refreshVisible()
				}
			}
		}
	}
}

func (s *Sidebar) Content() fyne.CanvasObject { return s.root }

func chatDisplayName(c api.ChatElement) string {
	name := strings.TrimSpace(c.FullName)
	if name == "" {
		name = strings.TrimSpace(c.PushName)
	}
	if name == "" {
		name = strings.TrimSpace(c.Short)
	}
	if name == "" {
		name = c.JID
	}
	return name
}

func isGroupJID(jid string) bool {
	return strings.HasSuffix(jid, "@g.us") || strings.HasSuffix(jid, "@newsletter")
}

func shortSender(sender string) string {
	s := sender
	if i := strings.Index(s, "@"); i > 0 {
		s = s[:i]
	}
	if len(s) > 12 {
		s = s[:12]
	}
	return s
}

func truncateRunes(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	if n <= 1 {
		return string(r[:n])
	}
	return string(r[:n-1]) + "\u2026"
}

func formatChatTimestamp(ts int64) string {
	if ts <= 0 {
		return ""
	}
	t := time.Unix(ts, 0).Local()
	now := time.Now().Local()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterday := today.AddDate(0, 0, -1)
	msgDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	switch {
	case !msgDay.Before(today):
		return t.Format("15:04")
	case !msgDay.Before(yesterday):
		return "Yesterday"
	case now.Sub(t) < 7*24*time.Hour:
		return t.Format("Mon")
	default:
		return t.Format("02/01/2006")
	}
}

func sidebarAvatarInitial(name, jid string) string {
	if isGroupJID(jid) {
		return "\U0001F465"
	}
	n := strings.TrimSpace(name)
	if n == "" || n == jid {
		return "#"
	}
	r := []rune(n)
	return strings.ToUpper(string(r[0]))
}

func avatarColor(jid string) color.Color {
	palette := []color.NRGBA{
		{R: 0x00, G: 0xa8, B: 0x84, A: 255},
		{R: 0x53, G: 0xbd, B: 0xeb, A: 255},
		{R: 0xd4, G: 0x76, B: 0xe8, A: 255},
		{R: 0xf7, G: 0x9c, B: 0x4c, A: 255},
		{R: 0xe5, G: 0x42, B: 0xa3, A: 255},
		{R: 0x06, G: 0xcf, B: 0x9c, A: 255},
	}
	var h uint32
	for _, c := range jid {
		h = h*31 + uint32(c)
	}
	return palette[h%uint32(len(palette))]
}

func spacer(w, h float32) fyne.CanvasObject {
	if h <= 0 {
		h = 1
	}
	if w <= 0 {
		w = 1
	}
	return container.NewGridWrap(fyne.NewSize(w, h), canvas.NewRectangle(color.Transparent))
}
