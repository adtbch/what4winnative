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
	"fyne.io/fyne/v2/widget"

	"github.com/lugvitc/whats4linux/api"
)

// Sidebar mirrors frontend/src/screens/ChatScreen.tsx: a 64px header
// ("WhatsApp" + new-chat/menu/avatar), the chats/communities/channels/status
// tab row, a rounded search pill and the chat list (48px avatar, name + time,
// preview + pin/badge).
type Sidebar struct {
	fa *FyneApp

	mu        sync.RWMutex
	allChats  []api.ChatElement
	visible   []api.ChatElement
	unread    map[string]int
	favourite map[string]bool
	avatarRes map[string]fyne.Resource
	view      string // chats | communities | channels | status
	archived  bool
	search    string

	list      *widget.List
	root      fyne.CanvasObject
	tabs      *viewTabs
	searchBox fyne.CanvasObject
	archivedR fyne.CanvasObject
	statusLbl *canvas.Text
	selfAv    *canvas.Image
	rowRefs   map[fyne.CanvasObject]*chatRowRefs
}

type chatRowRefs struct {
	jid       string
	hover     *canvas.Rectangle
	name      *truncLabel
	time      *canvas.Text
	preview   *truncLabel
	badgeBg   *canvas.Rectangle
	badgeTxt  *canvas.Text
	badgeWrap fyne.CanvasObject
	pin       fyne.CanvasObject
	avatarBg  *canvas.Circle
	avatarImg *canvas.Image
	avatarTxt *canvas.Text
	avPerson  fyne.CanvasObject
	avPeople  fyne.CanvasObject
	catcher   *rowCatcher
}

const chatRowHeight = 72

func NewSidebar(fa *FyneApp) *Sidebar {
	s := &Sidebar{
		fa:        fa,
		unread:    make(map[string]int),
		favourite: make(map[string]bool),
		avatarRes: make(map[string]fyne.Resource),
		view:      "chats",
		rowRefs:   make(map[fyne.CanvasObject]*chatRowRefs),
	}
	s.list = widget.NewList(
		func() int {
			s.mu.RLock()
			defer s.mu.RUnlock()
			if s.view != "chats" {
				return 0
			}
			return len(s.visible)
		},
		func() fyne.CanvasObject { return s.newRow() },
		func(id widget.ListItemID, obj fyne.CanvasObject) { s.updateRow(id, obj) },
	)
	s.list.OnSelected = func(id widget.ListItemID) {
		s.mu.RLock()
		if s.view != "chats" || id < 0 || id >= len(s.visible) {
			s.mu.RUnlock()
			return
		}
		c := s.visible[id]
		s.mu.RUnlock()
		s.selectChat(c)
	}

	s.root = container.NewBorder(s.topBar(), nil, nil, nil, s.body())
	go s.loadLoop()
	go s.listenEvents()
	go s.loadSelfAvatar()
	return s
}

func (s *Sidebar) topBar() fyne.CanvasObject {
	title := canvas.NewText("WhatsApp", waText)
	title.TextSize = 20
	title.TextStyle = fyne.TextStyle{Bold: true}

	newChat := newIconButton("sb-new", iconNewChat, 22, waMuted, func() {})
	menu := newIconButton("sb-menu", iconMore, 22, waMuted, func() {})

	s.selfAv = canvas.NewImageFromResource(nil)
	s.selfAv.FillMode = canvas.ImageFillCover
	s.selfAv.Hide()
	selfBg := canvas.NewCircle(color.NRGBA{R: 0x37, G: 0x44, B: 0x4c, A: 255})
	selfIcon := iconImage("sb-self", iconPerson, 18, waMuted)
	selfWrap := container.NewGridWrap(fyne.NewSize(32, 32),
		container.NewMax(selfBg, container.NewCenter(selfIcon), s.selfAv))

	actions := container.NewHBox(newChat, menu, selfWrap)
	row := container.NewBorder(nil, nil,
		container.NewHBox(container.NewPadded(title)),
		container.NewPadded(actions), nil)

	bg := canvas.NewRectangle(waPanel)
	bottom := canvas.NewRectangle(waDivider)
	bottom.SetMinSize(fyne.NewSize(0, 1))
	header := container.NewMax(bg, container.NewBorder(nil, bottom, nil, nil,
		container.NewGridWrap(fyne.NewSize(0, 64), row)))
	header.Resize(fyne.NewSize(0, 64))

	// tabs + search live below the header, like ChatScreen.tsx
	s.tabs = newViewTabs(s.view, func(v string) {
		s.mu.Lock()
		s.view = v
		s.mu.Unlock()
		s.syncChrome()
		s.refreshVisible()
	})

	s.searchBox = s.searchBar()
	s.archivedR = s.archivedRow()

	return container.NewVBox(header, s.tabs, s.searchBox, s.archivedR)
}

func (s *Sidebar) body() fyne.CanvasObject {
	s.statusLbl = canvas.NewText("", waMuted)
	s.statusLbl.TextSize = 12
	s.statusLbl.Alignment = fyne.TextAlignCenter
	placeholder := container.NewCenter(container.NewPadded(s.statusLbl))
	return container.NewMax(s.list, placeholder)
}

func (s *Sidebar) searchBar() fyne.CanvasObject {
	ent := widget.NewEntry()
	ent.SetPlaceHolder("Search or start new chat")
	ent.OnChanged = func(v string) {
		s.mu.Lock()
		s.search = v
		s.mu.Unlock()
		s.refreshVisible()
	}
	bg := canvas.NewRectangle(waPanel2)
	bg.CornerRadius = 18
	icon := iconImage("sb-search", iconSearch, 18, waMuted)
	inner := container.NewBorder(nil, nil,
		container.NewHBox(container.NewPadded(icon)), nil, ent)
	box := container.NewMax(bg, container.New(&padLayout{h: 10, v: 5}, inner))
	wrap := container.New(&padLayout{h: 12, v: 6}, box)
	return wrap
}

func (s *Sidebar) archivedRow() fyne.CanvasObject {
	return container.NewVBox()
}

// syncChrome hides the search bar for the status tab, like ChatScreen.tsx.
func (s *Sidebar) syncChrome() {
	s.mu.RLock()
	view := s.view
	archivedCount := 0
	for _, c := range s.allChats {
		if c.Archived {
			archivedCount++
		}
	}
	showArchived := s.archived
	s.mu.RUnlock()

	fyne.Do(func() {
		if s.searchBox != nil {
			if view == "status" {
				s.searchBox.Hide()
			} else {
				s.searchBox.Show()
			}
		}
		if s.archivedR != nil {
			if view == "chats" && archivedCount > 0 && !showArchived {
				s.archivedR.Show()
			} else {
				s.archivedR.Hide()
			}
		}
		if s.statusLbl != nil {
			switch view {
			case "chats":
				s.statusLbl.Text = ""
			case "communities":
				s.statusLbl.Text = "No communities yet"
			case "channels":
				s.statusLbl.Text = "No channels yet"
			case "status":
				s.statusLbl.Text = "No status updates"
			}
			s.statusLbl.Refresh()
		}
	})
}

// ---------- row rendering ----------

func (s *Sidebar) newRow() fyne.CanvasObject {
	hover := canvas.NewRectangle(color.Transparent)
	hover.CornerRadius = 0

	avatarBg := canvas.NewCircle(color.NRGBA{R: 0x37, G: 0x44, B: 0x4c, A: 255})
	avatarImg := canvas.NewImageFromResource(nil)
	avatarImg.FillMode = canvas.ImageFillCover
	avatarImg.Hide()
	// React swaps the fallback avatar contents by chat type: a "people" glyph for
	// groups, a person silhouette for contacts (ChatScreen.tsx MemoizedChatAvatar).
	avatarTxt := canvas.NewText("?", waText)
	avatarTxt.Alignment = fyne.TextAlignCenter
	avatarTxt.TextSize = 18
	avatarTxt.TextStyle = fyne.TextStyle{Bold: true}
	avatarPerson := iconImageVB("row-av-person", iconPerson, 36, 48, waAvatarIcon)
	avatarPeople := iconImage("row-av-people", iconPeople, 26, waAvatarIcon)
	avatarPerson.Hide()
	avatarPeople.Hide()
	avatarInner := container.NewMax(avatarBg, avatarImg,
		container.NewCenter(avatarTxt),
		container.NewCenter(avatarPerson),
		container.NewCenter(avatarPeople))
	avatar := container.NewGridWrap(fyne.NewSize(48, 48), avatarInner)

	// Name and preview must ellipsize, not overflow: canvas.Text never clips, so
	// a long name draws straight over the timestamp. React gets this from
	// `truncate` (overflow hidden + text-overflow ellipsis).
	// React: text-base (16px) / leading-6 (24px) / font-medium (500).
	// Fyne's TextStyle is a Bold flag only — no 500 weight — so Bold (700) is the
	// closest match; regular (400) reads visibly thinner than React next to the
	// 14px preview.
	name := newTruncLabel(16, waText, true)
	timeTxt := canvas.NewText("", waMuted)
	timeTxt.TextSize = 12
	timeTxt.TextStyle = fyne.TextStyle{Bold: true} // font-medium
	timeTxt.Alignment = fyne.TextAlignTrailing
	topRow := container.NewBorder(nil, nil, nil,
		container.NewHBox(timeTxt), container.New(&truncLayout{l: name}, name.obj()))

	preview := newTruncLabel(14, waMuted, false)

	badgeTxt := canvas.NewText("", waBadgeText)
	badgeTxt.TextSize = 11
	badgeTxt.TextStyle = fyne.TextStyle{Bold: true}
	badgeTxt.Alignment = fyne.TextAlignCenter
	badgeBg := canvas.NewRectangle(waGreen)
	badgeBg.CornerRadius = 9
	badgeWrap := container.NewGridWrap(fyne.NewSize(20, 20),
		container.NewMax(badgeBg, container.NewCenter(badgeTxt)))
	badgeWrap.Hide()

	pin := iconImage("row-pin", iconPin, 14, waMuted)
	pin.Hide()

	right := container.NewHBox(pin, badgeWrap)
	bottomRow := container.NewBorder(nil, nil, nil, right, container.New(&truncLayout{l: preview}, preview.obj()))

	// React: `flex justify-between items-baseline mb-1` over a `text-sm` row.
	// 24px (name line) + 4px (mb-1) + 20px (preview line) = 48px, the same as
	// the avatar, which is why the row is exactly 72px with py-3.
	mid := container.New(&stackedLayout{heights: []float32{24, 20}, gap: 4}, topRow, bottomRow)

	// px-4 py-3 + avatar mr-4
	body := container.NewBorder(nil, nil,
		container.New(&padLayout{h: 16, v: 0}, avatar), nil,
		container.New(&padLayout{h: 16, v: 12}, mid))

	catcher := &rowCatcher{}
	catcher.ExtendBaseWidget(catcher)
	root := container.NewMax(hover, body, catcher)
	sized := container.New(&minSizeLayout{w: 0, h: chatRowHeight}, root)

	catcher.onIn = func() {
		hover.FillColor = waHover
		hover.Refresh()
	}
	catcher.onOut = func() {
		hover.FillColor = color.Transparent
		hover.Refresh()
	}
	catcher.onTap = func() {
		// resolve the row back to its chat through the template map
		if refs := s.rowRefs[sized]; refs != nil && refs.jid != "" {
			s.mu.RLock()
			var found *api.ChatElement
			for i := range s.visible {
				if s.visible[i].JID == refs.jid {
					c := s.visible[i]
					found = &c
					break
				}
			}
			s.mu.RUnlock()
			if found != nil {
				s.selectChat(*found)
			}
		}
	}

	s.rowRefs[sized] = &chatRowRefs{
		hover: hover, name: name, time: timeTxt, preview: preview,
		badgeBg: badgeBg, badgeTxt: badgeTxt, badgeWrap: badgeWrap, pin: pin,
		avatarBg: avatarBg, avatarImg: avatarImg, avatarTxt: avatarTxt,
		avPerson: avatarPerson, avPeople: avatarPeople, catcher: catcher,
	}
	return sized
}

func (s *Sidebar) updateRow(id widget.ListItemID, obj fyne.CanvasObject) {
	defer func() { _ = recover() }()
	refs := s.rowRefs[obj]
	if refs == nil {
		return
	}
	s.mu.RLock()
	if id < 0 || id >= len(s.visible) {
		s.mu.RUnlock()
		return
	}
	c := s.visible[id]
	unread := s.unread[c.JID]
	avRes := s.avatarRes[c.JID]
	s.mu.RUnlock()
	refs.jid = c.JID

	refs.name.Set(chatDisplayName(c))

	refs.preview.Set(chatPreviewText(c))

	refs.time.Text = formatChatTimestamp(c.LatestTS)
	if unread > 0 {
		refs.time.Color = waGreen
	} else {
		refs.time.Color = waMuted
	}
	refs.time.Refresh()

	if c.Pinned {
		refs.pin.Show()
	} else {
		refs.pin.Hide()
	}
	if unread > 0 {
		if unread > 99 {
			refs.badgeTxt.Text = "99+"
		} else {
			refs.badgeTxt.Text = fmt.Sprintf("%d", unread)
		}
		refs.badgeTxt.Refresh()
		refs.badgeWrap.Show()
	} else {
		refs.badgeWrap.Hide()
	}

	refs.avatarBg.FillColor = avatarColor(c.JID)
	refs.avatarBg.Refresh()
	if avRes != nil {
		refs.avatarImg.Resource = avRes
		refs.avatarImg.Show()
		refs.avatarImg.Refresh()
		refs.avatarTxt.Hide()
	} else {
		refs.avatarImg.Hide()
		// React shows a glyph, never initials: "people" for groups, a person
		// silhouette for contacts. initials() is only a last-resort fallback.
		refs.avatarTxt.Hide()
		if isGroupJID(c.JID) {
			refs.avPeople.Show()
			refs.avPerson.Hide()
		} else {
			refs.avPerson.Show()
			refs.avPeople.Hide()
		}
	}
}

func (s *Sidebar) selectChat(c api.ChatElement) {
	s.mu.Lock()
	delete(s.unread, c.JID)
	s.mu.Unlock()
	s.fa.state.SetSelectedChatID(c.JID)
	s.fa.OpenChat(c.JID)
	s.refreshVisible()
}

// chatPreviewText flattens the stored HTML preview and maps media markers to
// the labels WhatsApp shows, matching ChatScreen.tsx's subtitle handling.
func chatPreviewText(c api.ChatElement) string {
	prev := htmlPreview(c.LatestMessage)
	switch prev {
	case "image":
		prev = "📷 Photo"
	case "video":
		prev = "🎥 Video"
	case "audio":
		prev = "🎵 Audio"
	case "document":
		prev = "📄 Document"
	case "sticker":
		prev = "Sticker"
	}
	if c.Sender != "" && isGroupJID(c.JID) && prev != "" {
		prev = shortSender(c.Sender) + ": " + prev
	}
	if prev == "" {
		prev = " "
	}
	return prev
}

// ---------- data ----------

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
			if s.statusLbl != nil {
				s.statusLbl.Text = "Waiting for WhatsApp…"
				s.statusLbl.Refresh()
			}
		})
		return 0, err
	}
	s.mu.Lock()
	s.allChats = chats
	s.mu.Unlock()
	s.refreshVisible()
	s.syncChrome()
	go s.loadAvatars(chats)
	return len(chats), nil
}

func (s *Sidebar) refreshVisible() {
	s.mu.Lock()
	s.visible = s.filterAndSortLocked()
	s.mu.Unlock()
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
		if s.archived != c.Archived {
			continue
		}
		if s.view == "chats" {
			if term != "" {
				name := strings.ToLower(chatDisplayName(c))
				if !strings.Contains(name, term) &&
					!strings.Contains(strings.ToLower(htmlPreview(c.LatestMessage)), term) {
					continue
				}
			}
		}
		out = append(out, c)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Pinned != out[j].Pinned {
			return out[i].Pinned
		}
		return out[i].LatestTS > out[j].LatestTS
	})
	return out
}

func (s *Sidebar) listenEvents() {
	refreshCh := SubscribeEvent("wa:chat_list_refresh")
	statusCh := SubscribeEvent("wa:status")
	newMsgCh := SubscribeEvent("wa:new_message")
	readCh := SubscribeEvent("wa:chat_marked_read")
	unreadCh := SubscribeEvent("wa:chat_marked_unread")
	picCh := SubscribeEvent("wa:picture_update")

	debounce := time.NewTimer(time.Hour)
	debounce.Stop()
	pending := false

	for {
		select {
		case <-s.fa.ctx.Done():
			return
		case <-refreshCh:
			pending = true
			debounce.Reset(350 * time.Millisecond)
		case <-debounce.C:
			if pending {
				pending = false
				s.fetchChats()
			}
		case data := <-statusCh:
			if v, ok := data.(string); ok {
				switch v {
				case "logged_in", "success", "connected", "synced", "ready":
					s.fetchChats()
				case "connecting":
					fyne.Do(func() {
						if s.statusLbl != nil {
							s.statusLbl.Text = "Connecting…"
							s.statusLbl.Refresh()
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

func (s *Sidebar) loadSelfAvatar() {
	url, err := s.fa.api.GetSelfAvatar(false)
	if err != nil || url == "" || s.selfAv == nil {
		return
	}
	res := resourceFromDataURL("self-avatar", url)
	if res == nil {
		return
	}
	fyne.Do(func() {
		s.selfAv.Resource = res
		s.selfAv.Show()
		s.selfAv.Refresh()
	})
}

func (s *Sidebar) Content() fyne.CanvasObject { return s.root }

func (s *Sidebar) Reload() { go s.fetchChats() }

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

// ---------- shared helpers ----------

func chatDisplayName(c api.ChatElement) string {
	for _, v := range []string{c.FullName, c.PushName, c.Short, c.Phno} {
		if n := strings.TrimSpace(v); n != "" {
			return n
		}
	}
	return c.JID
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
	return string(r[:n-1]) + "…"
}

// formatChatTimestamp mirrors lib/utils.ts formatChatTimestamp: time today,
// "Yesterday", else a numeric date.
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
	default:
		return t.Format("2/1/2006")
	}
}

func sidebarAvatarInitial(name, jid string) string {
	if strings.HasSuffix(jid, "@g.us") {
		return "👥"
	}
	if strings.HasSuffix(jid, "@newsletter") {
		return "📢"
	}
	n := strings.TrimSpace(name)
	if n == "" || n == jid {
		return "#"
	}
	r := []rune(n)
	return strings.ToUpper(string(r[0]))
}

// ---------- layouts & small widgets ----------

// rowCatcher gives a list row hover feedback and a tap target.
type rowCatcher struct {
	widget.BaseWidget
	onIn  func()
	onOut func()
	onTap func()
}

func (r *rowCatcher) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(canvas.NewRectangle(color.Transparent))
}
func (r *rowCatcher) MouseIn(*desktop.MouseEvent) {
	if r.onIn != nil {
		r.onIn()
	}
}
func (r *rowCatcher) MouseOut() {
	if r.onOut != nil {
		r.onOut()
	}
}
func (r *rowCatcher) Tapped(*fyne.PointEvent) {
	if r.onTap != nil {
		r.onTap()
	}
}
func (r *rowCatcher) Cursor() desktop.Cursor { return desktop.PointerCursor }
