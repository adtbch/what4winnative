package fyneui

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/lugvitc/whats4linux/api"
	"github.com/lugvitc/whats4linux/internal/store"
	"go.mau.fi/whatsmeow/types"
)

type ChatView struct {
	fa         *FyneApp
	jid        string
	chatName   string
	messages   *fyne.Container
	scroll     *container.Scroll
	entry      *widget.Entry
	headerName *canvas.Text
	headerSub  *canvas.Text
	avatarTxt  *canvas.Text
	avatarBg   *canvas.Circle
	cancel     context.CancelFunc
}

func NewChatView(fa *FyneApp, jid string) *ChatView {
	ctx, cancel := context.WithCancel(fa.ctx)
	cv := &ChatView{
		fa:       fa,
		jid:      jid,
		chatName: chatNameFromJID(jid),
		messages: container.NewVBox(),
		cancel:   cancel,
	}
	cv.loadHeader()
	go cv.loadMessages(ctx)
	go cv.listenNewMessages(ctx)
	return cv
}

func (cv *ChatView) loadHeader() {
	cv.headerName = canvas.NewText(cv.chatName, waText)
	cv.headerName.TextSize = 15
	cv.headerName.TextStyle = fyne.TextStyle{Bold: true}
	cv.headerSub = canvas.NewText("click here for contact info", waMuted)
	cv.headerSub.TextSize = 12
	cv.avatarTxt = canvas.NewText(sidebarAvatarInitial(cv.chatName, cv.jid), waText)
	cv.avatarTxt.Alignment = fyne.TextAlignCenter
	cv.avatarTxt.TextSize = 14
	cv.avatarTxt.TextStyle = fyne.TextStyle{Bold: true}
	cv.avatarBg = canvas.NewCircle(avatarColor(cv.jid))

	go func() {
		name := cv.chatName
		sub := "click here for contact info"
		if parsed, perr := types.ParseJID(cv.jid); perr == nil {
			if contact, err := cv.fa.api.GetContact(parsed); err == nil && contact != nil {
			if n := strings.TrimSpace(contact.FullName); n != "" {
				name = n
			} else if n := strings.TrimSpace(contact.PushName); n != "" {
				name = n
			}
		}
		}
		if isGroupJID(cv.jid) {
			sub = "group info"
		}
		cv.chatName = name
		fyne.Do(func() {
			if cv.headerName != nil {
				cv.headerName.Text = truncateRunes(name, 42)
				cv.headerName.Refresh()
			}
			if cv.headerSub != nil {
				cv.headerSub.Text = sub
				cv.headerSub.Refresh()
			}
			if cv.avatarTxt != nil {
				cv.avatarTxt.Text = sidebarAvatarInitial(name, cv.jid)
				cv.avatarTxt.Refresh()
			}
			if cv.avatarBg != nil {
				cv.avatarBg.FillColor = avatarColor(cv.jid)
				cv.avatarBg.Refresh()
			}
		})
	}()
}

func (cv *ChatView) loadMessages(ctx context.Context) {
	msgs, err := cv.fa.api.FetchMessagesPaged(cv.jid, 30, 0, "")
	log.Println("chat", cv.jid, "msgs:", len(msgs), err)
	if err != nil {
		log.Println("FetchMessages:", cv.jid, err)
		return
	}
	fyne.Do(func() {
		cv.messages.Objects = nil
		if len(msgs) == 0 {
			empty := canvas.NewText("No messages here yet.", waMuted)
			empty.Alignment = fyne.TextAlignCenter
			empty.TextSize = 13
			cv.messages.Add(container.NewCenter(container.NewPadded(empty)))
		}
		var prevDay string
		for _, m := range msgs {
			if m.Info.IsFromMe {
				// read mark: best-effort omitted in Fyne view
			}
			day := ""
			if day != "" && day != prevDay {
				cv.messages.Add(daySeparator(day))
				prevDay = day
			}
			cv.messages.Add(cv.bubble(m))
		}
		cv.messages.Refresh()
		if cv.scroll != nil {
			cv.scroll.ScrollToBottom()
		}
	})
}

func (cv *ChatView) listenNewMessages(ctx context.Context) {
	ch := SubscribeEvent("wa:new_message")
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-ch:
			msg, ok := data.(map[string]any)
			if !ok {
				continue
			}
			chatID, _ := msg["chatId"].(string)
			if chatID != cv.jid {
				continue
			}
			text, _ := msg["text"].(string)
			isFromMe, _ := msg["isFromMe"].(bool)
			ts, _ := msg["timestamp"].(string)
			if ts == "" {
				ts = time.Now().Format("15:04")
			}
			mtype, _ := msg["messageType"].(string)
			if strings.TrimSpace(text) == "" {
				text = messageTypeFallback(mtype)
			}
			go cv.loadHeader()
			fyne.Do(func() {
				cv.messages.Add(cv.textBubble(text, isFromMe, ts))
				cv.messages.Refresh()
				if cv.scroll != nil {
					cv.scroll.ScrollToBottom()
				}
			})
		}
	}
}

func (cv *ChatView) bubble(msg store.DecodedMessage) fyne.CanvasObject {
	body := extractDecodedBody(msg)
	ts := formatMsgTimestamp(msg.Info.Timestamp)
	if body == "" {
		body = messageTypeFallback(string(msg.Type))
	}
	return cv.textBubble(body, msg.Info.IsFromMe, ts)
}

func (cv *ChatView) textBubble(text string, isFromMe bool, ts string) fyne.CanvasObject {
	bg := waRecv
	align := fyne.TextAlignLeading
	outerPadLeft := float32(14)
	outerPadRight := float32(90)
	if isFromMe {
		bg = waSent
		align = fyne.TextAlignTrailing
		outerPadLeft = 90
		outerPadRight = 14
	}
	txt := widget.NewLabel(text)
	txt.Wrapping = fyne.TextWrapWord
	txt.Alignment = align

	meta := canvas.NewText(ts, waTime)
	meta.TextSize = 11
	meta.Alignment = fyne.TextAlignTrailing
	metaRow := container.NewHBox(meta)
	if isFromMe {
		ticks := canvas.NewText(" ✓✓", waMuted)
		ticks.TextSize = 11
		metaRow = container.NewHBox(meta, ticks)
	}

	card := canvas.NewRectangle(bg)
	card.CornerRadius = 9
	content := container.NewVBox(container.NewPadded(txt), container.NewPadded(container.NewBorder(nil, nil, nil, metaRow, canvas.NewRectangle(waBg))))
	bubble := container.NewMax(card, content)

	side := container.NewMax(canvas.NewRectangle(waBg))
	wrapped := container.NewBorder(spacer(1, 2), spacer(1, 2), spacer(outerPadLeft, 1), spacer(outerPadRight, 1), bubble)
	return container.NewMax(side, wrapped)
}

func daySeparator(label string) fyne.CanvasObject {
	t := canvas.NewText(label, waMuted)
	t.TextSize = 12
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
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("send failed: %w", err), cv.fa.win)
			})
			return
		}
		fyne.Do(func() {
			cv.messages.Add(cv.textBubble(text, true, time.Now().Format("15:04")))
			cv.messages.Refresh()
			if cv.scroll != nil {
				cv.scroll.ScrollToBottom()
			}
		})
	}()
}

func (cv *ChatView) Content() fyne.CanvasObject {
	back := widget.NewButton("←", func() { cv.fa.state.SetSelectedChatID(""); fyne.Do(func(){ if cv.fa.chatBox != nil { cv.fa.chatBox.Objects = []fyne.CanvasObject{cv.fa.emptyPane()}; cv.fa.chatBox.Refresh() } }) })
	back.Importance = widget.LowImportance
	avatar := container.NewGridWrap(fyne.NewSize(40, 40), container.NewMax(cv.avatarBg, container.NewCenter(cv.avatarTxt)))
	nameCol := container.NewVBox(spacer(1, 2), cv.headerName, cv.headerSub)
	left := container.NewHBox(container.NewGridWrap(fyne.NewSize(36, 36), back), spacer(4, 1), avatar, spacer(8, 1), nameCol)
	actions := container.NewHBox(
		widget.NewButton("🔍", func() {}),
		widget.NewButton("⋮", func() {}),
	)
	headerBg := canvas.NewRectangle(waPanel)
	header := container.NewMax(headerBg, container.NewPadded(container.NewBorder(nil, nil, left, actions, canvas.NewRectangle(waPanel))))

	cv.scroll = container.NewVScroll(container.NewPadded(cv.messages))
	bg := chatBackground(fyne.NewSize(0, 0))
	body := container.NewMax(bg, cv.scroll)

	cv.entry = widget.NewMultiLineEntry()
	cv.entry.SetPlaceHolder("Type a message")
	cv.entry.Wrapping = fyne.TextWrapWord
	cv.entry.OnSubmitted = func(string) { cv.sendMessage() }

	emoji := widget.NewButton("😊", func() {})
	emoji.Importance = widget.LowImportance
	plus := widget.NewButton("+", func() {})
	plus.Importance = widget.LowImportance
	send := widget.NewButton("➤", func() { cv.sendMessage() })
	send.Importance = widget.HighImportance

	inputBg := canvas.NewRectangle(waPanel)
	entryBg := canvas.NewRectangle(waPanel2)
	entryBg.CornerRadius = 8
	entryWrap := container.NewMax(entryBg, container.NewPadded(cv.entry))
	controls := container.NewBorder(nil, nil, container.NewHBox(emoji, plus), send, entryWrap)
	footer := container.NewMax(inputBg, container.NewPadded(controls))

	view := container.NewBorder(header, footer, nil, nil, body)
	cv.fa.win.Canvas().SetOnTypedKey(func(ev *fyne.KeyEvent) {
		if ev.Name == fyne.KeyReturn || ev.Name == fyne.KeyEnter {
			// keep no-op; send button/submit handles
		}
	})
	return view
}

func extractDecodedBody(msg store.DecodedMessage) string {
	if msg.Content != nil {
		if msg.Content.Conversation != "" {
			return strings.TrimSpace(msg.Content.Conversation)
		}
		if msg.Content.ExtendedTextMessage != nil && strings.TrimSpace(msg.Content.ExtendedTextMessage.Text) != "" {
			return strings.TrimSpace(msg.Content.ExtendedTextMessage.Text)
		}
		if msg.Content.ImageMessage != nil {
			if c := strings.TrimSpace(msg.Content.ImageMessage.Caption); c != "" {
				return "📷 " + c
			}
			return "📷 Photo"
		}
		if msg.Content.VideoMessage != nil {
			if c := strings.TrimSpace(msg.Content.VideoMessage.Caption); c != "" {
				return "🎥 " + c
			}
			return "🎥 Video"
		}
		if msg.Content.AudioMessage != nil {
			return "🎵 Audio"
		}
		if msg.Content.DocumentMessage != nil {
			return "📄 Document"
		}
		if msg.Content.StickerMessage != nil {
			return "Sticker"
		}
	}
	if msg.Poll != nil && msg.Poll.Name != "" {
		return "📊 " + msg.Poll.Name
	}
	return ""
}

func parseMsgTime(ts string) time.Time {
	ts = strings.TrimSpace(ts)
	if ts == "" {
		return time.Time{}
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02T15:04:05Z07:00"} {
		if t, err := time.Parse(layout, ts); err == nil {
			return t.Local()
		}
	}
	// try unix seconds string
	if n, err := fmt.Sscan(ts, new(int64)); err == nil && n == 1 {
		// ignore; not worth pulling strconv in this helper
	}
	return time.Time{}
}

func formatMsgTimestamp(ts string) string {
	t := parseMsgTime(ts)
	if t.IsZero() {
		if strings.Contains(ts, ":") {
			return ts
		}
		return ""
	}
	return t.Format("15:04")
}

var _ context.Context


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

func messageTypeFallback(t string) string {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case "image", "imagemessage":
		return "📷 Photo"
	case "video", "videomessage":
		return "🎥 Video"
	case "audio", "audiomessage", "ptt":
		return "🎵 Audio"
	case "document", "documentmessage":
		return "📄 Document"
	case "sticker", "stickermessage":
		return "Sticker"
	default:
		if strings.TrimSpace(t) == "" {
			return "Message"
		}
		return t
	}
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
