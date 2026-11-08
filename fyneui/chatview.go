package fyneui

import (
	"net/url"
	"encoding/base64"
	"context"
	"fmt"
	"image/color"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"

	"github.com/lugvitc/whats4linux/api"
	"github.com/lugvitc/whats4linux/internal/misc"
	"github.com/lugvitc/whats4linux/internal/store"
	"go.mau.fi/whatsmeow/types"
)

// ChatView mirrors frontend/src/screens/ChatDetail.tsx +
// components/chat/{MessageList,MessageItem,ChatInput}.tsx.
type ChatView struct {
	fa       *FyneApp
	jid      string
	chatName string
	messages *fyne.Container
	scroll   *container.Scroll
	entry    *composerEntry
	header   *chatHeader
	cancel   context.CancelFunc

	maxBubble    float32
	rendered     map[string]bool
	lastDay      string
	lastSenderID string
	lastFromMe   bool
	hasLast      bool

	mu              sync.Mutex
	mediaGeneration uint64

	infoPanel *chatInfoPanel
	showInfo  bool
	mainSplit *fyne.Container
}

func NewChatView(fa *FyneApp, jid string) *ChatView {
	ctx, cancel := context.WithCancel(fa.ctx)
	cv := &ChatView{
		fa:        fa,
		jid:       jid,
		chatName:  chatNameFromJID(jid),
		messages:  container.NewVBox(),
		cancel:    cancel,
		maxBubble: 560,
		rendered:  make(map[string]bool),
	}
	cv.loadHeader()
	go cv.loadMessages(ctx)
	go cv.listenEvents(ctx)
	return cv
}

func (cv *ChatView) loadHeader() {
	cv.header = newChatHeader(func() {
		cv.cancel()
		cv.fa.OpenEmptyChat()
	}, func() { cv.toggleInfo() }, chatHeaderData{Name: cv.chatName})

	go func() {
		d := chatHeaderData{Name: cv.chatName}

		if isGroupJID(cv.jid) {
			if gi, err := cv.fa.api.GetGroupInfo(cv.jid); err == nil {
				if n := strings.TrimSpace(gi.GroupName); n != "" {
					d.Name = n
				}
				d.Subtitle = cv.groupSubtitle(gi)
			}
		} else if parsed, err := types.ParseJID(cv.jid); err == nil {
			if contact, cerr := cv.fa.api.GetContact(parsed); cerr == nil && contact != nil {
				if n := strings.TrimSpace(contact.FullName); n != "" {
					d.Name = n
				} else if n := strings.TrimSpace(contact.PushName); n != "" {
					d.Name = n
				}
			}
			// ChatDetail.tsx only passes onCallClick for contacts.
			d.Call = func() {
				if cv.fa == nil || cv.fa.api == nil { return }
				go func() {
					if err := cv.fa.api.PlaceCall(cv.jid); err != nil {
						log.Println("place call:", err)
						fyne.Do(func() { dialog.ShowError(fmt.Errorf("call failed: %w", err), cv.fa.win) })
					}
				}()
			}
		}
		cv.chatName = d.Name

		// avatar photo: same backend call the React list/detail header uses
		if url, err := cv.fa.api.GetCachedAvatar(cv.jid, false); err == nil && url != "" {
			d.Avatar = dataURLResource(cv.jid, url)
		}
		fyne.Do(func() {
			if cv.header != nil {
				cv.header.Set(d)
			}
		})
		// If no cached avatar, fetch once and refresh.
		if d.Avatar == nil && cv.fa != nil && cv.fa.api != nil {
			if url, err := cv.fa.api.GetCachedAvatar(cv.jid, true); err == nil && url != "" {
				if res := dataURLResource(cv.jid, url); res != nil {
					d2 := d
					d2.Avatar = res
					fyne.Do(func() { if cv.header != nil { cv.header.Set(d2) } })
				}
			}
		}
	}()
}

// groupSubtitle mirrors ChatDetail.tsx:135-160 — "You, Alice, ~ Bob, +91 …",
// self filtered out by phone/JID, each participant labelled by the first of
// full name → "~ push name" → short → formatted phone.
func (cv *ChatView) groupSubtitle(gi api.Group) string {
	self, _ := cv.fa.api.GetProfile("")
	parts := []string{"You"}
	for _, p := range gi.Participants {
		c := p.Contact
		if self.Phno != "" && c.Phno == self.Phno {
			continue
		}
		if self.JID != "" && c.JID == self.JID {
			continue
		}
		if label := participantLabel(c); label != "" {
			parts = append(parts, label)
		}
	}
	return strings.Join(parts, ", ")
}

func participantLabel(c api.Contact) string {
	if c.FullName != "" {
		return c.FullName
	}
	if c.PushName != "" {
		return "~ " + c.PushName
	}
	if c.Short != "" {
		return c.Short
	}
	return formatPhone(c.Phno)
}

// formatPhone mirrors frontend/src/lib/utils.ts:12-19.
func formatPhone(digits string) string {
	clean := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, digits)
	if clean == "" {
		return ""
	}
	if len(clean) <= 10 {
		return "+" + clean
	}
	cc := clean[:len(clean)-10]
	rest := clean[len(clean)-10:]
	return "+" + cc + " " + rest[:5] + " " + rest[5:]
}

func (cv *ChatView) loadMessages(ctx context.Context) {
	msgs, err := cv.fa.api.FetchMessagesPaged(cv.jid, 80, 0, "")
	log.Println("chat", cv.jid, "msgs:", len(msgs), err)
	if err != nil {
		log.Println("FetchMessages:", cv.jid, err)
		return
	}
	sort.SliceStable(msgs, func(i, j int) bool { return msgTime(msgs[i]).Before(msgTime(msgs[j])) })
	fyne.Do(func() {
		cv.messages.Objects = nil
		cv.lastDay = ""
		cv.hasLast = false
		cv.lastSenderID = ""
		cv.lastFromMe = false
		if len(msgs) == 0 {
			empty := canvas.NewText("No messages here yet.", waMuted)
			empty.Alignment = fyne.TextAlignCenter
			empty.TextSize = 13
			cv.messages.Add(container.NewCenter(container.NewPadded(empty)))
		}
		for _, m := range msgs {
			cv.addMessageLocked(m)
		}
		cv.messages.Refresh()
		if cv.scroll != nil {
			cv.scroll.ScrollToBottom()
		}
	})
	// mark read once the messages are on screen, like the React view does
	go func() {
		ids := make([]string, 0, len(msgs))
		for _, m := range msgs {
			if !m.Info.IsFromMe && m.Info.ID != "" {
				ids = append(ids, m.Info.ID)
			}
		}
		if len(ids) > 0 {
			if err := cv.fa.api.MarkRead(cv.jid, ids, "read"); err != nil {
				log.Println("MarkRead:", err)
			}
		}
	}()
}

// addMessageLocked appends one message plus its day separator. Caller must
// already be on the Fyne thread.
func (cv *ChatView) addMessageLocked(m store.DecodedMessage) {
	if label := messageDayLabel(msgTime(m)); label != "" && label != cv.lastDay {
		cv.messages.Add(daySeparator(label))
		cv.lastDay = label
		cv.hasLast = false // separator breaks a sender run (MessageList.tsx:150)
	}
	prevSender := cv.lastSenderID
	hadPrev := cv.hasLast
	if m.Info.ID != "" {
		cv.rendered[m.Info.ID] = true
	}
	// grouping per MessageList.tsx:156-159 (skip separators)
	first := !hadPrev || prevSender != senderRunID(m) || m.Info.IsFromMe != cv.lastFromMe
	cv.lastFromMe = m.Info.IsFromMe
	cv.messages.Add(cv.bubble(m, first))
	cv.lastSenderID = senderRunID(m)
	cv.hasLast = true
}

func (cv *ChatView) listenEvents(ctx context.Context) {
	newCh := SubscribeEvent("wa:new_message")
	updCh := SubscribeEvent("wa:message_update")
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-newCh:
			msg, ok := data.(map[string]any)
			if !ok {
				continue
			}
			chatID, _ := msg["chatId"].(string)
			if chatID != cv.jid {
				continue
			}
			go cv.appendLatest()
		case <-updCh:
			go cv.appendLatest()
		}
	}
}

// appendLatest re-reads the newest page so every row is built from the same
// DecodedMessage shape as the initial load.
func (cv *ChatView) appendLatest() {
	msgs, err := cv.fa.api.FetchMessagesPaged(cv.jid, 25, 0, "")
	if err != nil || len(msgs) == 0 {
		return
	}
	sort.SliceStable(msgs, func(i, j int) bool { return msgTime(msgs[i]).Before(msgTime(msgs[j])) })
	var newUnreadIDs []string
	fyne.Do(func() {
		for _, m := range msgs {
			if m.Info.ID != "" && cv.rendered[m.Info.ID] {
				continue
			}
			if !m.Info.IsFromMe && m.Info.ID != "" {
				newUnreadIDs = append(newUnreadIDs, m.Info.ID)
			}
			cv.addMessageLocked(m)
		}
		cv.messages.Refresh()
		if cv.scroll != nil {
			cv.scroll.ScrollToBottom()
		}
	})
	if len(newUnreadIDs) > 0 {
		go func(ids []string) {
			if err := cv.fa.api.MarkRead(cv.jid, ids, "read"); err != nil {
				log.Println("MarkRead appendLatest:", err)
			}
		}(newUnreadIDs)
	}
}

func msgTime(m store.DecodedMessage) time.Time {
	ts := strings.TrimSpace(m.Info.Timestamp)
	if ts == "" {
		return time.Time{}
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05"} {
		if t, err := time.Parse(layout, ts); err == nil {
			return t.Local()
		}
	}
	return time.Time{}
}

func formatMsgTimestamp(ts string) string {
	t := msgTime(store.DecodedMessage{Info: store.DecodedMessageInfo{Timestamp: ts}})
	if t.IsZero() {
		return ""
	}
	return t.Format("15:04")
}

// bubble renders one message like MessageItem.tsx: media block for
// image/video/sticker/audio/document, HTML rich text for text, meta row with
// time + ticks inside the bubble. first controls group/avatar/spacing run.
func (cv *ChatView) bubble(m store.DecodedMessage, first bool) fyne.CanvasObject {
	fromMe := m.Info.IsFromMe
	bg := waRecv
	if fromMe {
		bg = waSent
	}
	card := canvas.NewRectangle(bg)
	card.CornerRadius = 8

	maxW := cv.maxBubble
	if maxW <= 0 {
		maxW = 560
	}

	var body fyne.CanvasObject
	content := m.Content
	switch {
	case content == nil:
		body = cv.textBlock(messageText(m), maxW)
	case content.ImageMessage != nil:
		body = cv.mediaBlock(m, mediaKindImage, maxW)
		if c := strings.TrimSpace(content.ImageMessage.Caption); c != "" {
			body = container.NewVBox(body, cv.textBlock(c, maxW))
		}
	case content.VideoMessage != nil:
		body = cv.mediaBlock(m, mediaKindVideo, maxW)
		if c := strings.TrimSpace(content.VideoMessage.Caption); c != "" {
			body = container.NewVBox(body, cv.textBlock(c, maxW))
		}
	case content.StickerMessage != nil:
		body = cv.mediaBlock(m, mediaKindSticker, maxW)
	case content.AudioMessage != nil:
		body = cv.audioBlock()
	case content.DocumentMessage != nil:
		body = cv.documentBlock(m)
	default:
		body = cv.textBlock(messageText(m), maxW)
	}
	// MessageItem.tsx renders this for every content type it cannot draw.
	// Also catch non-nil content structs with no renderable media and empty text
	// (e.g. SQLite rows with text="" has_media=0 that decode to empty protobuf).
	plain := strings.TrimSpace(htmlToText(messageText(m)))
	hasRenderableMedia := content != nil && (content.ImageMessage != nil || content.VideoMessage != nil ||
		content.StickerMessage != nil || content.AudioMessage != nil || content.DocumentMessage != nil ||
		m.Poll != nil)
	if plain == "" && !hasRenderableMedia {
		body = unsupportedText()
	}

	inner := container.NewVBox(container.NewPadded(body), cv.metaRow(m, fromMe))
	if isGroupJID(cv.jid) && !fromMe && first {
		col := profileColorNRGBA(m.Info.Sender)
		name := canvas.NewText(senderLabel(m), col)
		name.TextSize = 11
		name.TextStyle = fyne.TextStyle{Bold: true}
		inner = container.NewVBox(container.NewPadded(name), inner)
	}
	wrapped := cv.wrapBubble(fromMe, first, content, inner)
	// Copy: avoid overlay on messages with links (would block HyperlinkSegment taps).
	// Pragmatic: plain text without http(s) gets tap-to-copy.
	if plain != "" && !strings.Contains(plain, "http://") && !strings.Contains(plain, "https://") {
		copyText := plain
		copyTap := newTapArea(func() {
			fyne.CurrentApp().Clipboard().SetContent(copyText)
			log.Println("copied message")
		})
		wrapped = container.NewMax(wrapped, copyTap)
	}
	return wrapped
}

// wrapBubble applies bubble background + corner and the 3px/12px/70px outer
// margins (MessageItem.tsx:408-447).
func cvWrapBubble(fromMe, first bool, content *store.DecodedMessageContent, inner fyne.CanvasObject) fyne.CanvasObject {
	var stack fyne.CanvasObject = inner
	if content == nil || content.StickerMessage == nil {
		card := canvas.NewRectangle(bubbleBg(fromMe))
		card.CornerRadius = bubbleCorner(fromMe, first)
		stack = container.NewMax(card, inner)
	}
	// spacing: MessageList.tsx:162 firstInGroup?pt-2 pb-px : py-px (8/1 vs 1/1)
	top := float32(1)
	if first {
		top = 8
	}
	var row fyne.CanvasObject
	if fromMe {
		row = container.NewHBox(layout.NewSpacer(), stack)
	} else {
		row = container.NewHBox(stack, layout.NewSpacer())
	}
	return container.NewBorder(spacer(1, top), spacer(1, 1), spacer(12, 1), spacer(12, 1), row)
}

func (cv *ChatView) wrapBubble(fromMe, first bool, content *store.DecodedMessageContent, inner fyne.CanvasObject) fyne.CanvasObject {
	return cvWrapBubble(fromMe, first, content, inner)
}

func (cv *ChatView) metaRow(m store.DecodedMessage, fromMe bool) fyne.CanvasObject {
	row := container.NewHBox()
	if m.Edited {
		e := canvas.NewText("Edited ", waTime)
		e.TextSize = 11
		row.Add(e)
	}
	ts := canvas.NewText(formatMsgTimestamp(m.Info.Timestamp), waTime)
	ts.TextSize = 11
	row.Add(ts)
	if fromMe {
		ticks := canvas.NewText(" ✓✓", waTick)
		ticks.TextSize = 11
		row.Add(ticks)
	}
	return container.NewHBox(layout.NewSpacer(), row)
}

// textBlock renders message HTML (api.processMessageText output) at maxW.
func (cv *ChatView) textBlock(src string, maxW float32) fyne.CanvasObject {
	inner := maxW - 28
	if inner < 60 {
		inner = 60
	}
	plain := strings.TrimSpace(htmlToText(src))
	if plain == "" {
		return canvas.NewText(" ", waText)
	}
	// MessageItem.tsx:283 — emoji-only messages render at 32px (leading-10),
	// no rich markup. Detected on the stripped plain text.
	if emojiOnlyText(src) {
		t := canvas.NewText(plain, waText)
		t.TextSize = 32
		return wrapFixed(t, minF(fyne.MeasureText(plain, 32, fyne.TextStyle{}).Width+4, inner), 40)
	}
	// Long text clamp/Read more (MessageItem.tsx LONG_TEXT_CHARS/LINES). YAGNI:
	// collapsed renders truncated plain (drops rich markup) to avoid HTML slicing.
	if isLongTextPlain(plain) {
		fullBox, fw, fh := htmlTextBlock(src, inner, 14)
		if fw < 24 {
			fw = 24
		}
		full := container.NewGridWrap(fyne.NewSize(minF(fw+16, inner), fh+16), fullBox)
		short := truncatePlain(plain, longTextChars, longTextLines)
		colBox, cw, ch := htmlTextBlock(short, inner, 14)
		if cw < 24 {
			cw = 24
		}
		collapsed := container.NewGridWrap(fyne.NewSize(minF(cw+16, inner), ch+16), colBox)
		return newClampText(full, collapsed)
	}
	box, w, h := htmlTextBlock(src, inner, 14)
	if w < 24 {
		w = 24
	}
	return container.NewGridWrap(fyne.NewSize(minF(w+16, inner), h+16), box)
}

func (cv *ChatView) mediaBlock(m store.DecodedMessage, kind int, maxW float32) fyne.CanvasObject {
	w, h := mediaDims(m)
	size := mediaBox(w, h)
	if kind == mediaKindSticker {
		size = fyne.NewSize(150, 150)
	}
	if size.Width > maxW {
		size.Width = maxW
	}

	img := canvas.NewImageFromResource(nil)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(size)

	holder := canvas.NewRectangle(waPanel2)
	holder.CornerRadius = 6
	// Placeholder label behind the image so a bubble is never blank while the
	// media is still being fetched (React shows a blurred placeholder box).
	ph := "📷 Photo"
	switch kind {
	case mediaKindVideo:
		ph = "🎥 Video"
	case mediaKindSticker:
		ph = "Sticker"
	}
	phTxt := canvas.NewText(ph, waMuted)
	phTxt.TextSize = 12
	phTxt.Alignment = fyne.TextAlignCenter
	box := container.NewMax(holder, container.NewCenter(phTxt), img)
	box.Resize(size)

	id := m.Info.ID
	chatID := cv.jid
	enqueueMediaFetch(cv, chatID, id, kind, img, box)

	msgID := id
	chatJID := chatID
	onTap := func() {
		go func() {
			if cv.fa == nil || cv.fa.api == nil { return }
			path, err := cv.fa.api.SaveMediaToDownloads(chatJID, msgID)
			if err != nil {
				log.Println("open media:", err)
				fyne.Do(func() { dialog.ShowError(fmt.Errorf("open failed: %w", err), cv.fa.win) })
				return
			}
			if u, err := url.Parse("file:///" + strings.ReplaceAll(path, "\\", "/")); err == nil {
				_ = fyne.CurrentApp().OpenURL(u)
			}
		}()
	}
	if kind == mediaKindVideo {
		play := canvas.NewText("▶", waWhite)
		play.TextSize = 34
		play.Alignment = fyne.TextAlignCenter
		circle := canvas.NewCircle(waPlayOverlay)
		overlay := container.NewCenter(container.NewGridWrap(fyne.NewSize(56, 56), container.NewMax(circle, container.NewCenter(play))))
		return container.NewMax(box, overlay, newTapArea(onTap))
	}
	return container.NewMax(box, newTapArea(onTap))
}

func (cv *ChatView) audioBlock() fyne.CanvasObject {
	return newAudioBubble(mediaBubbleData{Kind: mediaKindAudio})
}

func (cv *ChatView) documentBlock(m store.DecodedMessage) fyne.CanvasObject {
	name := "Document"
	if m.Content != nil && m.Content.DocumentMessage != nil && strings.TrimSpace(m.Content.DocumentMessage.FileName) != "" {
		name = m.Content.DocumentMessage.FileName
	}
	bubble := newDocumentBubble(mediaBubbleData{Kind: mediaKindDocument, FileName: name, SizeLabel: "Document"})
	msgID := m.Info.ID
	chatID := cv.jid
	tap := newTapArea(func() {
		go func() {
			if cv.fa == nil || cv.fa.api == nil { return }
			path, err := cv.fa.api.SaveMediaToDownloads(chatID, msgID)
			if err != nil {
				log.Println("save media:", err)
				fyne.Do(func() { dialog.ShowError(fmt.Errorf("download failed: %w", err), cv.fa.win) })
				return
			}
			if u, err := url.Parse("file:///" + strings.ReplaceAll(path, "\\", "/")); err == nil {
				_ = fyne.CurrentApp().OpenURL(u)
			}
			fyne.Do(func() { dialog.ShowInformation("Downloaded", "Saved: "+path, cv.fa.win) })
		}()
	})
	return container.NewMax(bubble, tap)
}

const (
	mediaKindImage = iota
	mediaKindVideo
	mediaKindSticker
	mediaKindAudio
	mediaKindDocument
)

func mediaDims(m store.DecodedMessage) (int, int) {
	if m.Content == nil {
		return 0, 0
	}
	switch {
	case m.Content.ImageMessage != nil:
		return m.Content.ImageMessage.Width, m.Content.ImageMessage.Height
	case m.Content.VideoMessage != nil:
		return m.Content.VideoMessage.Width, m.Content.VideoMessage.Height
	case m.Content.StickerMessage != nil:
		return m.Content.StickerMessage.Width, m.Content.StickerMessage.Height
	}
	return 0, 0
}

// messageText returns the body the React item renders: HTML for text messages.
func messageText(m store.DecodedMessage) string {
	if m.Content == nil {
		return ""
	}
	if m.Content.Conversation != "" {
		return m.Content.Conversation
	}
	if m.Content.ExtendedTextMessage != nil && m.Content.ExtendedTextMessage.Text != "" {
		return m.Content.ExtendedTextMessage.Text
	}
	if m.Poll != nil && m.Poll.Name != "" {
		return "📊 " + m.Poll.Name
	}
	return ""
}

// unsupportedText matches MessageItem.tsx's fallback for content types the UI
// cannot render (senderKeyDistribution, protocol messages, empty rows).
func unsupportedText() fyne.CanvasObject {
	t := canvas.NewText("Unsupported Message Type", waMuted)
	t.TextSize = 12
	t.TextStyle = fyne.TextStyle{Italic: true}
	return t
}

func senderLabel(m store.DecodedMessage) string {
	name := strings.TrimSpace(m.Info.PushName)
	if name == "" {
		name = shortSender(m.Info.Sender)
	}
	if name == "" {
		name = "Unknown"
	}
	return "~ " + name
}

func daySeparator(label string) fyne.CanvasObject {
	t := canvas.NewText(label, waMuted)
	t.TextSize = 11
	t.TextStyle = fyne.TextStyle{Bold: true}
	bg := canvas.NewRectangle(waPanel2)
	bg.CornerRadius = 8
	box := container.NewMax(bg, container.NewPadded(t))
	return container.NewCenter(container.NewPadded(box))
}

func (cv *ChatView) sendText(text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	go func() {
		if _, err := cv.fa.api.SendMessage(cv.jid, api.MessageContent{Type: "text", Text: text}); err != nil {
			log.Println("SendMessage:", err)
			fyne.Do(func() { dialog.ShowError(fmt.Errorf("send failed: %w", err), cv.fa.win) })
			return
		}
		cv.appendLatest()
	}()
}

func (cv *ChatView) Content() fyne.CanvasObject {
	header := cv.header

	cv.scroll = container.NewVScroll(container.New(&padLayout{h: 0, v: 6}, cv.messages))
	chatBody := container.NewMax(chatBackground(fyne.NewSize(0, 0)), cv.scroll)

	onEmoji := func() {
		if cv.entry != nil {
			cv.entry.SetText(cv.entry.Text + "😊")
		}
	}
	onAttach := func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			path := reader.URI().Path()
			_ = reader.Close()
			if path != "" {
				log.Println("composer attach selected:", path)
				dialog.ShowInformation("Attachment", fmt.Sprintf("Selected: %s\nAttachment sending is coming soon.", path), cv.fa.win)
			}
		}, cv.fa.win)
	}
	onPasteImage := func(pngData []byte) {
		go cv.sendClipboardImage(pngData)
	}
	comp, entry := newComposer(func(s string) { cv.sendText(s) }, onEmoji, onAttach, onPasteImage)
	cv.entry = entry
	footerBg := canvas.NewRectangle(waPanel)
	footer := container.NewMax(footerBg, comp)

	ph := canvas.NewRectangle(color.Transparent)
	ph.SetMinSize(fyne.NewSize(0, 0))
	cv.mainSplit = container.NewBorder(nil, nil, nil, ph, chatBody)
	return container.NewBorder(header, footer, nil, nil, cv.mainSplit)
}

func minF(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func chatNameFromJID(jid string) string {
	j := jid
	if i := strings.Index(j, "@"); i > 0 {
		j = j[:i]
	}
	if j == "" {
		return jid
	}
	return j
}

func messageDayLabel(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	now := time.Now().Local()
	tt := t.Local()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yday := today.AddDate(0, 0, -1)
	d := time.Date(tt.Year(), tt.Month(), tt.Day(), 0, 0, 0, 0, tt.Location())
	switch {
	case !d.Before(today):
		return "Today"
	case !d.Before(yday):
		return "Yesterday"
	default:
		return tt.Format("02/01/2006")
	}
}

// firstInGroup mirrors MessageList.tsx:156-160 — the previous row must come
// from a different sender or side, and date separators break the run.
func (cv *ChatView) firstInGroup(m store.DecodedMessage) bool {
	if !cv.hasLast {
		return true
	}
	return cv.lastSenderID != senderRunID(m) || m.Info.IsFromMe != cv.lastFromMe
}

func senderRunID(m store.DecodedMessage) string {
	side := "r"
	if m.Info.IsFromMe {
		side = "s"
	}
	return side + "|" + m.Info.Sender
}

// profileColorNRGBA parses misc.GetProfileColor's "#RRGGBB" into NRGBA.
func profileColorNRGBA(jid string) color.NRGBA {
	h := misc.GetProfileColor(jid)
	if len(h) == 7 && h[0] == '#' {
		if r, err := strconv.ParseUint(h[1:3], 16, 8); err == nil {
			if g, err := strconv.ParseUint(h[3:5], 16, 8); err == nil {
				if b, err := strconv.ParseUint(h[5:7], 16, 8); err == nil {
					return color.NRGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}
				}
			}
		}
	}
	return waGreen
}

func (cv *ChatView) toggleInfo() {
	cv.showInfo = !cv.showInfo
	if cv.showInfo && cv.infoPanel == nil {
		cv.infoPanel = newChatInfoPanel(func() { cv.toggleInfo() }, func(want bool) {
			go func() {
				if err := cv.fa.api.ToggleChatMute(cv.jid, want); err != nil {
					log.Println("mute:", err)
				}
				cv.loadInfoData()
			}()
		})
		cv.infoPanel.Set(chatInfoData{Name: cv.chatName, Subtitle: cv.jid, IsGroup: isGroupJID(cv.jid)})
		go cv.loadInfoData()
	}
	cv.refreshLayout()
}

func (cv *ChatView) refreshLayout() {
	fyne.Do(func() {
		if cv.mainSplit == nil {
			return
		}
		body := cv.mainSplit.Objects[0]
		var right fyne.CanvasObject
		if cv.showInfo && cv.infoPanel != nil {
			right = cv.infoPanel
		} else {
			right = canvas.NewRectangle(color.Transparent)
			right.(*canvas.Rectangle).SetMinSize(fyne.NewSize(0, 0))
		}
		// keep 0-width placeholder when closed
		cv.mainSplit.Objects = []fyne.CanvasObject{body, right}
		cv.mainSplit.Refresh()
	})
}

func (cv *ChatView) loadInfoData() {
	if cv.infoPanel == nil {
		return
	}
	jid := cv.jid
	d := chatInfoData{Name: cv.chatName, Subtitle: jid, IsGroup: isGroupJID(jid)}
	if m, err := cv.fa.api.IsChatMuted(jid); err == nil {
		d.Muted = m
	}
	if d.IsGroup {
		if gi, err := cv.fa.api.GetGroupInfo(jid); err == nil {
			if strings.TrimSpace(gi.GroupName) != "" {
				d.Name = gi.GroupName
			}
			if strings.TrimSpace(gi.GroupTopic) != "" {
				d.About = gi.GroupTopic
			}
			d.MemberCount = gi.ParticipantCount
			for _, p := range gi.Participants {
				n := strings.TrimSpace(p.Contact.FullName)
				if n == "" {
					n = strings.TrimSpace(p.Contact.PushName)
				}
				if n == "" {
					n = p.Contact.Phno
				}
				d.Participants = append(d.Participants, chatInfoParticipant{Name: n, Phone: p.Contact.Phno, IsAdmin: p.IsAdmin})
			}
			if d.Subtitle == jid {
				d.Subtitle = fmt.Sprintf("Group · %d participants", gi.ParticipantCount)
			}
		}
	} else if parsed, err := types.ParseJID(jid); err == nil {
		if c, err := cv.fa.api.GetContact(parsed); err == nil && c != nil {
			if n := strings.TrimSpace(c.FullName); n != "" {
				d.Name = n
			} else if n := strings.TrimSpace(c.PushName); n != "" {
				d.Name = n
			}
			if c.Phno != "" {
				d.Subtitle = c.Phno
			}
		}
	}
	if url, err := cv.fa.api.GetCachedAvatar(jid, false); err == nil && url != "" {
		d.Avatar = dataURLResource(jid, url)
	}
	fyne.Do(func() {
		if cv.infoPanel != nil {
			cv.infoPanel.Set(d)
		}
	})
}

func (cv *ChatView) sendClipboardImage(pngData []byte) {
	if cv.fa == nil || cv.fa.api == nil || len(pngData) == 0 {
		return
	}
	b64 := base64.StdEncoding.EncodeToString(pngData)
	_, err := cv.fa.api.SendMessage(cv.jid, api.MessageContent{
		Type:       "image",
		Base64Data: b64,
		Mimetype:   "image/png",
	})
	if err != nil {
		log.Println("send clipboard image:", err)
		fyne.Do(func() { dialog.ShowError(fmt.Errorf("send image failed: %w", err), cv.fa.win) })
		return
	}
	cv.appendLatest()
}
