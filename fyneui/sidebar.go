package fyneui

import (
	"encoding/base64"
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
	"fyne.io/fyne/v2/theme"
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
	p.bg.CornerRadius = 16
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

type layoutHPad struct{ h, v float32 }

func (l layoutHPad) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(0, 0)
	}
	s := objects[0].MinSize()
	return fyne.NewSize(s.Width+l.h*2, s.Height+l.v*2)
}
func (l layoutHPad) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objects {
		o.Resize(fyne.NewSize(size.Width-l.h*2, size.Height-l.v*2))
		o.Move(fyne.NewPos(l.h, l.v))
	}
}

func (p *tappablePill) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewMax(p.bg, container.New(layoutHPad{h: 12, v: 6}, p.label)))
}
func (p *tappablePill) Tapped(*fyne.PointEvent) {
	if p.onTap != nil {
		p.onTap()
	}
}
func (p *tappablePill) Cursor() desktop.Cursor { return desktop.PointerCursor }

type chatRowRefs struct {
	selBg     *canvas.Rectangle
	name      *canvas.Text
	time      *canvas.Text
	preview   *canvas.Text
	pin       *canvas.Text
	badge     *fyne.Container
	badgeTxt  *canvas.Text
	avatarBg  *canvas.Circle
	avatarImg *canvas.Image
	avatarTxt *canvas.Text
}

type Sidebar struct {
	fa        *FyneApp
	mu        sync.RWMutex
	allChats  []api.ChatElement
	visible   []api.ChatElement
	unread    map[string]int
	favourite map[string]bool
	avatarRes map[string]fyne.Resource
	filter    chatFilter
	search    string
	list      *widget.List
	root      fyne.CanvasObject
	syncLbl   *canvas.Text
	pillBox   *fyne.Container
	searchEnt *widget.Entry
	rowRefs   map[fyne.CanvasObject]*chatRowRefs
}

func NewSidebar(fa *FyneApp) *Sidebar {
	s := &Sidebar{
		fa:        fa,
		unread:    make(map[string]int),
		favourite: make(map[string]bool),
		avatarRes: make(map[string]fyne.Resource),
		filter:    filterAll,
		rowRefs:   make(map[fyne.CanvasObject]*chatRowRefs),
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
		s.fa.state.mu.Lock()
		s.fa.state.SelectedChatID = jid
		s.fa.state.mu.Unlock()
		s.fa.OpenChat(jid)
		s.refreshVisible()
	}

	title := canvas.NewText("WhatsApp", waText)
	title.TextSize = 20
	title.TextStyle = fyne.TextStyle{Bold: true}
	iconNew := widget.NewButtonWithIcon("", theme.ContentAddIcon(), nil)
	iconNew.Importance = widget.LowImportance
	iconMenu := widget.NewButtonWithIcon("", theme.MoreVerticalIcon(), nil)
	iconMenu.Importance = widget.LowImportance
	header := container.NewPadded(container.NewBorder(nil, nil, title, container.NewHBox(iconNew, iconMenu), nil))

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
	searchWrap := container.NewMax(searchBg, container.NewPadded(container.NewBorder(nil, nil, widget.NewIcon(theme.SearchIcon()), nil, s.searchEnt)))

	s.pillBox = container.NewHBox()
	s.rebuildPills()
	filtersWrap := container.NewHScroll(container.NewPadded(s.pillBox))
	filtersWrap.SetMinSize(fyne.NewSize(0, 40))

	s.syncLbl = canvas.NewText("Syncing chats\u2026", waMuted)
	s.syncLbl.TextSize = 12
	syncBar := container.NewPadded(container.NewCenter(s.syncLbl))

	top := container.NewVBox(header, searchWrap, filtersWrap)
	s.root = container.NewBorder(top, syncBar, nil, nil, s.list)

	go s.loadLoop()
	go s.listenEvents()
	return s
}

func (s *Sidebar) rebuildPills() {
	s.mu.RLock()
	filter := s.filter
	texts := map[chatFilter]string{
		filterAll:        "All",
		filterUnread:     s.pillTextLocked(filterUnread),
		filterFavourites: s.pillTextLocked(filterFavourites),
		filterGroups:     s.pillTextLocked(filterGroups),
		filterArchived:   s.pillTextLocked(filterArchived),
	}
	s.mu.RUnlock()
	fyne.Do(func() {
		objs := make([]fyne.CanvasObject, 0, 5)
		for _, df := range []chatFilter{filterAll, filterUnread, filterFavourites, filterGroups, filterArchived} {
			dff := df
			p := newTappablePill(texts[dff], filter == dff, func() {
				s.mu.Lock()
				s.filter = dff
				s.mu.Unlock()
				s.rebuildPills()
				s.refreshVisible()
			})
			objs = append(objs, p)
		}
		s.pillBox.Objects = objs
		s.pillBox.Refresh()
	})
}

func (s *Sidebar) pillTextLocked(f chatFilter) string {
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
	selBg := canvas.NewRectangle(color.Transparent)
	avatarBg := canvas.NewCircle(waHover)
	avatarImg := canvas.NewImageFromResource(nil)
	avatarImg.FillMode = canvas.ImageFillContain
	avatarImg.Hide()
	avatarTxt := canvas.NewText("?", waText)
	avatarTxt.Alignment = fyne.TextAlignCenter
	avatarTxt.TextSize = 15
	avatarTxt.TextStyle = fyne.TextStyle{Bold: true}
	avatarInner := container.NewMax(avatarBg, avatarImg, container.NewCenter(avatarTxt))
	avatar := container.NewGridWrap(fyne.NewSize(49, 49), avatarInner)

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
	pin.TextSize = 11
	badgeBg := canvas.NewCircle(waGreen)
	badgeTxt := canvas.NewText("", color.NRGBA{R: 0x11, G: 0x1b, B: 0x21, A: 255})
	badgeTxt.TextSize = 11
	badgeTxt.TextStyle = fyne.TextStyle{Bold: true}
	badgeTxt.Alignment = fyne.TextAlignCenter
	badge := container.NewGridWrap(fyne.NewSize(20, 20), container.NewMax(badgeBg, container.NewCenter(badgeTxt)))
	badge.Hide()
	right := container.NewHBox(pin, badge)
	bottom := container.NewBorder(nil, nil, nil, right, preview)

	mid := container.NewVBox(spacer(0, 2), top, spacer(0, 1), bottom)
	row := container.NewBorder(nil, nil, container.NewPadded(avatar), nil, container.NewPadded(mid))
	divider := canvas.NewRectangle(waDivider)
	divider.SetMinSize(fyne.NewSize(0, 1))
	inner := container.NewBorder(nil, divider, nil, nil, row)
	root := container.NewMax(selBg, inner)
	if s.rowRefs == nil { s.rowRefs = make(map[fyne.CanvasObject]*chatRowRefs) }
	s.rowRefs[root] = &chatRowRefs{selBg: selBg, name: name, time: timeTxt, preview: preview, pin: pin, badge: badge, badgeTxt: badgeTxt, avatarBg: avatarBg, avatarImg: avatarImg, avatarTxt: avatarTxt}
	return root
}

func (s *Sidebar) updateRow(id widget.ListItemID, obj fyne.CanvasObject) {
	defer func(){ recover() }()
	s.mu.RLock()
	if id < 0 || id >= len(s.visible) {
		s.mu.RUnlock()
		return
	}
	c := s.visible[id]
	unread := s.unread[c.JID]
	avRes := s.avatarRes[c.JID]
	selID := ""
	if s.fa != nil && s.fa.state != nil { selID = s.fa.state.SelectedChatID }
	selected := selID == c.JID
	s.mu.RUnlock()

	root, ok := obj.(*fyne.Container)
	if !ok || s.rowRefs == nil { return }
	refs := s.rowRefs[root]
	if refs == nil { return }
	if selected { refs.selBg.FillColor = waHover } else { refs.selBg.FillColor = color.Transparent }
	refs.selBg.Refresh()
	name := chatDisplayName(c)
	refs.name.Text = truncateRunes(name, 30); refs.name.Refresh()
	prev := strings.TrimSpace(c.LatestMessage)
	if c.Sender != "" && isGroupJID(c.JID) && prev != "" { prev = shortSender(c.Sender) + ": " + prev }
	if prev == "" { prev = " " }
	refs.preview.Text = truncateRunes(prev, 44); refs.preview.Refresh()
	refs.time.Text = formatChatTimestamp(c.LatestTS)
	if unread > 0 { refs.time.Color = waGreen } else { refs.time.Color = waMuted }
	refs.time.Refresh()
	if c.Pinned { refs.pin.Text = "📌" } else { refs.pin.Text = "" }
	refs.pin.Refresh()
	if unread > 0 {
		if unread > 99 { refs.badgeTxt.Text = "99+" } else { refs.badgeTxt.Text = fmt.Sprintf("%d", unread) }
		refs.badgeTxt.Refresh(); refs.badge.Show()
	} else { refs.badge.Hide() }
	refs.avatarBg.FillColor = avatarColor(c.JID); refs.avatarBg.Refresh()
	if avRes != nil { refs.avatarImg.Resource = avRes; refs.avatarImg.Show(); refs.avatarImg.Refresh(); refs.avatarTxt.Hide()
	} else { refs.avatarImg.Hide(); refs.avatarTxt.Text = initials(name); refs.avatarTxt.Refresh(); refs.avatarTxt.Show() }
}

func (s *Sidebar) loadLoop() {
	delays := []time.Duration{200 * time.Millisecond, 800 * time.Millisecond, 2 * time.Second, 4 * time.Second, 6 * time.Second, 10 * time.Second}
	for i, d := range delays {
		select {
		case <-s.fa.ctx.Done():
			return
		case <-time.After(d):
		}
		n, err := s.fetchChats()
		log.Println("sidebar load attempt", i+1, "count", n, "err", err)
		if n > 0 {
			return
		}
	}
	for {
		select {
		case <-s.fa.ctx.Done():
			return
		case <-time.After(15 * time.Second):
		}
		s.mu.RLock()
		empty := len(s.allChats) == 0
		s.mu.RUnlock()
		if empty {
			s.fetchChats()
		} else {
			return
		}
	}
}

func (s *Sidebar) fetchChats() (int, error) {
	chats, err := s.fa.api.GetChatList()
	if err != nil {
		log.Println("sidebar GetChatList error:", err)
		fyne.Do(func() {
			if s.syncLbl != nil {
				s.syncLbl.Text = "Waiting for WhatsApp\u2026"
				fyne.Do(func(){ s.syncLbl.Refresh() })
			}
		})
		return 0, err
	}
	s.mu.Lock()
	s.allChats = chats
	s.mu.Unlock()
	s.refreshVisible()
	fyne.Do(func() {
		if s.syncLbl == nil {
			return
		}
		if len(chats) == 0 {
			s.syncLbl.Text = "No chats yet \u2014 syncing history\u2026"
		} else {
			s.syncLbl.Text = ""
		}
		fyne.Do(func(){ s.syncLbl.Refresh() })
	})
	go s.loadAvatars(chats)
	return len(chats), nil
}

func (s *Sidebar) refreshVisible() {
	s.mu.Lock()
	s.visible = s.filterAndSortLocked()
	s.mu.Unlock()
	s.rebuildPills()
	fyne.Do(func() {
		if s.list != nil {
			s.list.UnselectAll()
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
			prev := strings.ToLower(c.LatestMessage)
			if !strings.Contains(name, term) && !strings.Contains(prev, term) {
				continue
			}
		}
		out = append(out, c)
	}
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
		return !c.Archived && s.unread[c.JID] > 0
	case filterGroups:
		return !c.Archived && isGroupJID(c.JID)
	case filterFavourites:
		return !c.Archived && s.favourite[c.JID]
	default:
		return !c.Archived
	}
}

func (s *Sidebar) listenEvents() {
	refreshCh := SubscribeEvent("wa:chat_list_refresh")
	statusCh := SubscribeEvent("wa:status")
	newMsgCh := SubscribeEvent("wa:new_message")
	readCh := SubscribeEvent("wa:chat_marked_read")
	unreadCh := SubscribeEvent("wa:chat_marked_unread")
	picCh := SubscribeEvent("wa:picture_update")

	refreshDebounce := time.NewTimer(0)
	if !refreshDebounce.Stop() {
		select {
		case <-refreshDebounce.C:
		default:
		}
	}
	pendingRefresh := false

	for {
		select {
		case <-s.fa.ctx.Done():
			return
		case <-refreshCh:
			pendingRefresh = true
			refreshDebounce.Reset(350 * time.Millisecond)
		case <-refreshDebounce.C:
			if pendingRefresh {
				pendingRefresh = false
				s.fetchChats()
			}
		case data := <-statusCh:
			if v, ok := data.(string); ok {
				log.Println("sidebar status event:", v)
				switch v {
				case "logged_in", "success", "connected", "synced", "ready":
					s.fetchChats()
				case "connecting":
					fyne.Do(func() {
						if s.syncLbl != nil {
							s.syncLbl.Text = "Connecting\u2026"
							fyne.Do(func(){ s.syncLbl.Refresh() })
						}
					})
				}
			}
		case data := <-newMsgCh:
			s.handleNewMessage(data)
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
		case data := <-picCh:
			if jid, ok := data.(string); ok && jid != "" {
				go s.loadAvatar(jid, true)
			}
		}
	}
}

func (s *Sidebar) handleNewMessage(data any) {
	m, ok := data.(map[string]any)
	if !ok {
		return
	}
	chatID, _ := m["chatId"].(string)
	if chatID == "" {
		return
	}
	isFromMe, _ := m["isFromMe"].(bool)
	text, _ := m["messageText"].(string)
	sender, _ := m["sender"].(string)
	reaction, _ := m["reaction"].(string)
	var ts int64
	switch v := m["timestamp"].(type) {
	case int64:
		ts = v
	case float64:
		ts = int64(v)
	case int:
		ts = int64(v)
	}
	if reaction != "" {
		if text != "" {
			text = sender + " reacted " + reaction + " to: \"" + truncateRunes(text, 40) + "\""
		} else {
			text = sender + " reacted " + reaction
		}
		sender = ""
	}

	s.fa.state.mu.RLock()
	selected := s.fa.state.SelectedChatID
	s.fa.state.mu.RUnlock()

	s.mu.Lock()
	idx := -1
	for i := range s.allChats {
		if s.allChats[i].JID == chatID {
			idx = i
			break
		}
	}
	if idx >= 0 {
		if text != "" {
			s.allChats[idx].LatestMessage = text
		}
		if ts > 0 {
			s.allChats[idx].LatestTS = ts
		}
		s.allChats[idx].Sender = sender
		if !isFromMe && selected != chatID && reaction == "" {
			s.unread[chatID]++
		}
		s.mu.Unlock()
		s.refreshVisible()
		return
	}
	s.mu.Unlock()
	go func() {
		s.fetchChats()
		if !isFromMe && selected != chatID && reaction == "" {
			s.mu.Lock()
			s.unread[chatID]++
			s.mu.Unlock()
			s.refreshVisible()
		}
	}()
}

func (s *Sidebar) loadAvatars(chats []api.ChatElement) {
	const workers = 4
	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup
	for _, c := range chats {
		jid := c.JID
		if jid == "" {
			continue
		}
		s.mu.RLock()
		_, has := s.avatarRes[jid]
		s.mu.RUnlock()
		if has {
			continue
		}
		wg.Add(1)
		sem <- struct{}{}
		go func(j string) {
			defer wg.Done()
			defer func() { <-sem }()
			s.loadAvatar(j, false)
		}(jid)
	}
	wg.Wait()
}

func (s *Sidebar) loadAvatar(jid string, recache bool) {
	dataURL, err := s.fa.api.GetCachedAvatar(jid, recache)
	if err != nil || dataURL == "" {
		return
	}
	res := resourceFromDataURL(jid, dataURL)
	if res == nil {
		return
	}
	s.mu.Lock()
	s.avatarRes[jid] = res
	s.mu.Unlock()
	fyne.Do(func() {
		if s.list != nil {
			s.list.Refresh()
		}
	})
}

func resourceFromDataURL(name, dataURL string) fyne.Resource {
	i := strings.Index(dataURL, ",")
	if i < 0 {
		return nil
	}
	raw, err := base64.StdEncoding.DecodeString(dataURL[i+1:])
	if err != nil || len(raw) == 0 {
		return nil
	}
	return fyne.NewStaticResource(name, raw)
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
		name = strings.TrimSpace(c.Phno)
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
	if s == "" {
		return sender
	}
	if len([]rune(s)) > 14 {
		s = string([]rune(s)[:14])
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
		return t.Format("3:04 PM")
	case !msgDay.Before(yesterday):
		return "Yesterday"
	default:
		return t.Format("02/01/2006")
	}
}

func sidebarAvatarInitial(name, jid string) string {
	if strings.HasSuffix(jid, "@g.us") {
		return "\U0001F465"
	}
	if strings.HasSuffix(jid, "@newsletter") {
		return "\U0001F4E2"
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
		{R: 0x7f, G: 0x66, B: 0xff, A: 255},
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


func (s *Sidebar) Reload() {
	go s.fetchChats()
}


func initials(name string) string {
	name = strings.TrimSpace(name)
	if name == "" { return "?" }
	r := []rune(name)
	if len(r) == 1 { return strings.ToUpper(string(r)) }
	return strings.ToUpper(string(r[0]))
}
