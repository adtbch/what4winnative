package fyneui

import (
	"context"
	"fmt"
	"image/color"
	"log"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/lugvitc/whats4linux/api"
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
	entry    *widget.Entry
	header   *chatHeader
	cancel   context.CancelFunc

	maxBubble float32
	rendered  map[string]bool
	lastDay   string
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
	}, func() {}, chatHeaderData{Name: cv.chatName})

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
			d.Call = func() {}
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
	}
	if m.Info.ID != "" {
		cv.rendered[m.Info.ID] = true
	}
	cv.messages.Add(cv.bubble(m))
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
	fyne.Do(func() {
		for _, m := range msgs {
			if m.Info.ID != "" && cv.rendered[m.Info.ID] {
				continue
			}
			cv.addMessageLocked(m)
		}
		cv.messages.Refresh()
		if cv.scroll != nil {
			cv.scroll.ScrollToBottom()
		}
	})
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
// time + ticks inside the bubble.
func (cv *ChatView) bubble(m store.DecodedMessage) fyne.CanvasObject {
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
	if strings.TrimSpace(htmlToText(messageText(m))) == "" && content == nil {
		body = unsupportedText()
	}

	inner := container.NewVBox(container.NewPadded(body), cv.metaRow(m, fromMe))
	if isGroupJID(cv.jid) && !fromMe {
		name := canvas.NewText(senderLabel(m), waGreen)
		name.TextSize = 12
		name.TextStyle = fyne.TextStyle{Bold: true}
		inner = container.NewVBox(container.NewPadded(name), inner)
	}
	// MessageItem.tsx: stickers float with no bubble background.
	var stack fyne.CanvasObject = inner
	if content == nil || content.StickerMessage == nil {
		stack = container.NewMax(card, inner)
	}

	if fromMe {
		return container.NewBorder(spacer(1, 3), spacer(1, 3), spacer(70, 1), spacer(12, 1), stack)
	}
	return container.NewBorder(spacer(1, 3), spacer(1, 3), spacer(12, 1), spacer(70, 1), stack)
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
	return container.NewHBox(spacer(1, 1), row)
}

// textBlock renders message HTML (api.processMessageText output) at maxW.
func (cv *ChatView) textBlock(src string, maxW float32) fyne.CanvasObject {
	inner := maxW - 28
	if inner < 60 {
		inner = 60
	}
	if strings.TrimSpace(htmlToText(src)) == "" {
		return canvas.NewText(" ", waText)
	}
	box, w, h := htmlTextBlock(src, inner, 14)
	if w < 24 {
		w = 24
	}
	// widget.RichText reports MinSize as text + 2*innerPadding on both axes and
	// draws the text inset by that padding. Sizing the cell to the bare text
	// width therefore made RichText wrap every line ~16px early, and using
	// MinSize().Height left a blank band under the text. Size from the measured
	// wrap instead and give back the padding it will consume.
	pad := theme.Current().Size(theme.SizeNameInnerPadding)
	return container.NewGridWrap(fyne.NewSize(minF(w+pad*2, inner), h+pad*2), box)
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
	go func() {
		var dataURL string
		if kind == mediaKindVideo {
			dataURL = cv.fa.api.GetVideoThumbnail(id)
		} else {
			dataURL = cv.fa.api.GetImageThumbnail(id)
			// MediaContent.tsx auto-downloads images/stickers as soon as they are
			// on screen: the embedded thumbnail is only a placeholder.
			if dataURL == "" {
				if u, err := cv.fa.api.GetCachedImage(id); err == nil {
					dataURL = u
				}
			}
			if dataURL == "" {
				if u, err := cv.fa.api.DownloadMedia(chatID, id); err == nil {
					dataURL = u
				} else {
					log.Println("DownloadMedia:", id, err)
				}
			}
		}
		if dataURL == "" {
			return
		}
		res := dataURLResource(id, dataURL)
		if res == nil {
			return
		}
		fyne.Do(func() {
			img.Resource = res
			img.Refresh()
			box.Refresh()
		})
	}()

	if kind == mediaKindVideo {
		play := canvas.NewText("▶", waWhite)
		play.TextSize = 34
		play.Alignment = fyne.TextAlignCenter
		circle := canvas.NewCircle(waPlayOverlay)
		overlay := container.NewCenter(container.NewGridWrap(fyne.NewSize(56, 56), container.NewMax(circle, container.NewCenter(play))))
		return container.NewMax(box, overlay)
	}
	return box
}

func (cv *ChatView) audioBlock() fyne.CanvasObject {
	icon := canvas.NewText("▶", waText)
	icon.TextSize = 20
	label := canvas.NewText("Audio message", waMuted)
	label.TextSize = 13
	row := container.NewHBox(container.NewGridWrap(fyne.NewSize(36, 36), container.NewCenter(icon)), label)
	holder := canvas.NewRectangle(waPanel2)
	holder.CornerRadius = 6
	return container.NewMax(holder, container.NewPadded(row))
}

func (cv *ChatView) documentBlock(m store.DecodedMessage) fyne.CanvasObject {
	name := "Document"
	if m.Content != nil && m.Content.DocumentMessage != nil && strings.TrimSpace(m.Content.DocumentMessage.FileName) != "" {
		name = m.Content.DocumentMessage.FileName
	}
	badge := canvas.NewText("📄", waText)
	badge.TextSize = 22
	title := canvas.NewText(truncateRunes(name, 38), waText)
	title.TextSize = 13
	sub := canvas.NewText("Document", waMuted)
	sub.TextSize = 11
	col := container.NewVBox(title, sub)
	row := container.NewBorder(nil, nil, container.NewGridWrap(fyne.NewSize(40, 40), container.NewCenter(badge)), nil, col)
	holder := canvas.NewRectangle(waPanel2)
	holder.CornerRadius = 6
	return container.NewMax(holder, container.NewPadded(row))
}

const (
	mediaKindImage = iota
	mediaKindVideo
	mediaKindSticker
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

func (cv *ChatView) sendMessage() {
	text := strings.TrimSpace(cv.entry.Text)
	if text == "" {
		return
	}
	cv.entry.SetText("")
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
	body := container.NewMax(chatBackground(fyne.NewSize(0, 0)), cv.scroll)

	// Composer (ChatInput.tsx): a rounded pill (emoji left, attach right) with
	// the send button as a separate green circle outside it.
	cv.entry = widget.NewMultiLineEntry()
	cv.entry.SetPlaceHolder("Type a message")
	cv.entry.Wrapping = fyne.TextWrapWord
	cv.entry.SetMinRowsVisible(1)
	cv.entry.OnSubmitted = func(string) { cv.sendMessage() }

	emojiBtn := newIconButton("cv-emoji", iconEmoji, 20, waMuted, func() {})
	attachBtn := newIconButton("cv-attach", iconAttach, 20, waMuted, func() {})

	pillBg := canvas.NewRectangle(waPanel2)
	pillBg.CornerRadius = 22
	pill := container.NewMax(pillBg, container.New(&padLayout{h: 6, v: 4},
		container.NewBorder(nil, nil,
			container.NewHBox(container.NewPadded(emojiBtn)), nil,
			container.NewBorder(nil, nil, nil, container.NewPadded(attachBtn), cv.entry))))

	send := newSendButton(func() { cv.sendMessage() })

	row := container.NewBorder(nil, nil, nil, send, pill)
	footerBg := canvas.NewRectangle(waPanel)
	footer := container.NewMax(footerBg, container.New(&padLayout{h: 12, v: 6},
		container.NewBorder(nil, spacer(1, 4), nil, nil, row)))

	return container.NewBorder(header, footer, nil, nil, body)
}

// newSendButton is ChatInput.tsx's 44px green circle with the paper-plane.
func newSendButton(onTap func()) fyne.CanvasObject {
	circle := canvas.NewCircle(waAccent)
	icon := iconImage("cv-send", iconSend, 20, waWhite)
	btn := newTapArea(onTap)
	inner := container.NewMax(circle, container.NewCenter(icon), btn)
	return container.NewGridWrap(fyne.NewSize(44, 44), inner)
}

type tapArea struct {
	widget.BaseWidget
	onTap func()
}

func newTapArea(onTap func()) *tapArea {
	t := &tapArea{onTap: onTap}
	t.ExtendBaseWidget(t)
	return t
}

func (t *tapArea) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(canvas.NewRectangle(color.Transparent))
}
func (t *tapArea) Tapped(*fyne.PointEvent) {
	if t.onTap != nil {
		t.onTap()
	}
}
func (t *tapArea) Cursor() desktop.Cursor { return desktop.PointerCursor }

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
