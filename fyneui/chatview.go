package fyneui

import (
	"go.mau.fi/whatsmeow/types"
	"fmt"
	"image/color"
	"log"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/lugvitc/whats4linux/api"
	"github.com/lugvitc/whats4linux/internal/store"
)

type ChatView struct {
	fa       *FyneApp
	chatID   string
	chatName string
	isGroup  bool
	messages []store.DecodedMessage
	list     *widget.List
	input    *widget.Entry
	header   *canvas.Text
	sub      *canvas.Text
}

func NewChatView(fa *FyneApp, chatID string) *ChatView {
	cv := &ChatView{
		fa:       fa,
		chatID:   chatID,
		messages: []store.DecodedMessage{},
		input:    widget.NewEntry(),
		header:   canvas.NewText(chatID, waText),
		sub:      canvas.NewText("", waMuted),
	}
	cv.header.TextSize = 15
	cv.header.TextStyle = fyne.TextStyle{Bold: true}
	cv.sub.TextSize = 12
	cv.input.SetPlaceHolder("Type a message")
	cv.isGroup = strings.Contains(chatID, "@g.us") || strings.Contains(chatID, "@newsletter")

	cv.list = widget.NewList(
		func() int { return len(cv.messages) },
		func() fyne.CanvasObject { return cv.bubbleTemplate() },
		func(id widget.ListItemID, obj fyne.CanvasObject) { cv.updateBubble(id, obj) },
	)

	cv.input.OnSubmitted = func(text string) {
		if strings.TrimSpace(text) == "" || cv.chatID == "" {
			return
		}
		_, err := cv.fa.api.SendMessage(cv.chatID, api.MessageContent{Type: "text", Text: text})
		if err != nil {
			log.Println("chatview: SendMessage:", err)
			return
		}
		fyne.Do(func() { cv.input.SetText("") })
	}

	go cv.loadHeader()
	go cv.loadMessages()
	go cv.listenEvents()
	return cv
}

func (cv *ChatView) bubbleTemplate() fyne.CanvasObject {
	bg := canvas.NewRectangle(waPanel2)
	bg.CornerRadius = 8
	sender := canvas.NewText("", waAccent)
	sender.TextSize = 11
	body := widget.NewLabel("")
	body.Wrapping = fyne.TextWrapWord
	meta := canvas.NewText("", waMuted)
	meta.TextSize = 10
	meta.Alignment = fyne.TextAlignTrailing
	inner := container.NewVBox(sender, body, meta)
	box := container.NewMax(bg, container.NewPadded(inner))
	return container.NewPadded(container.NewMax(box)) // updated dynamically
}

func msgText(m *store.DecodedMessage) string {
	if m == nil || m.Content == nil {
		return ""
	}
	c := m.Content
	switch {
	case c.Conversation != "":
		return c.Conversation
	case c.ExtendedTextMessage != nil:
		return c.ExtendedTextMessage.Text
	case c.ImageMessage != nil:
		t := "[Photo]"
		if c.ImageMessage.Caption != "" {
			t += " " + c.ImageMessage.Caption
		}
		return t
	case c.VideoMessage != nil:
		t := "[Video]"
		if c.VideoMessage.Caption != "" {
			t += " " + c.VideoMessage.Caption
		}
		return t
	case c.AudioMessage != nil:
		return "[Audio]"
	case c.DocumentMessage != nil:
		t := "[Document]"
		if c.DocumentMessage.FileName != "" {
			t += " " + c.DocumentMessage.FileName
		}
		return t
	case c.StickerMessage != nil:
		return "[Sticker]"
	default:
		return "[Message]"
	}
}

func receipt(m store.DecodedMessage) string {
	// best-effort; fields vary
	if !m.Info.IsFromMe {
		return ""
	}
	// common status patterns in store; fallback
	return "✓"
}

func (cv *ChatView) updateBubble(id widget.ListItemID, obj fyne.CanvasObject) {
	if id < 0 || id >= len(cv.messages) {
		return
	}
	m := cv.messages[id]
	text := msgText(&m)
	fromMe := m.Info.IsFromMe
	sentBg := color.NRGBA{R: 0x00, G: 0x5c, B: 0x4b, A: 255}
	recvBg := color.NRGBA{R: 0x20, G: 0x2c, B: 0x33, A: 255}
	bubbleBg := recvBg
	if fromMe {
		bubbleBg = sentBg
	}

	senderName := ""
	if !fromMe && cv.isGroup {
		senderName = m.Info.PushName
		if senderName == "" {
			senderName = m.Info.Sender
		}
		if i := strings.Index(senderName, "@"); i > 0 {
			senderName = senderName[:i]
		}
	}

	ts := ""
	if m.Info.Timestamp != "" {
		if t, err := time.Parse(time.RFC3339, m.Info.Timestamp); err == nil {
			ts = t.Local().Format("15:04")
		} else {
			ts = m.Info.Timestamp
			if len(ts) > 16 { ts = ts[11:16] }
		}
	}
	if fromMe && ts != "" {
		ts = ts + " " + receipt(m)
	}

	bg := canvas.NewRectangle(bubbleBg)
	bg.CornerRadius = 8
	sender := canvas.NewText(senderName, waAccent)
	sender.TextSize = 11
	if senderName == "" {
		sender.Hide()
	} else {
		sender.Show()
	}
	body := widget.NewLabel(text)
	body.Wrapping = fyne.TextWrapWord
	meta := canvas.NewText(ts, waMuted)
	meta.TextSize = 10
	meta.Alignment = fyne.TextAlignTrailing
	inner := container.NewVBox(sender, body, meta)
	bubble := container.NewMax(bg, container.NewPadded(inner))
	// constrain width-ish
	wrap := container.NewGridWrap(fyne.NewSize(520, 0), bubble)
	var row fyne.CanvasObject
	if fromMe {
		row = container.NewBorder(nil, nil, nil, wrap, canvas.NewRectangle(color.Transparent))
	} else {
		row = container.NewBorder(nil, nil, wrap, nil, canvas.NewRectangle(color.Transparent))
	}
	pad := obj.(*fyne.Container)
	pad.Objects = []fyne.CanvasObject{container.NewPadded(row)}
	pad.Refresh()
}

func (cv *ChatView) loadHeader() {
	name := cv.chatID
	if cv.fa != nil && cv.fa.api != nil {
		if cv.isGroup {
			if gi, err := cv.fa.api.GetGroupInfo(cv.chatID); err == nil {
				if gi.GroupName != "" { name = gi.GroupName }
			}
		} else {
			if jid, err := types.ParseJID(cv.chatID); err == nil {
			if c, err := cv.fa.api.GetContact(jid); err == nil && c != nil {
				if c.FullName != "" {
					name = c.FullName
				} else if c.PushName != "" {
					name = c.PushName
				} else if c.Short != "" {
					name = c.Short
				}
			}
			}
		}
	}
	sub := ""
	if cv.isGroup {
		sub = "group"
	}
	fyne.Do(func() {
		cv.header.Text = name
		cv.header.Refresh()
		cv.sub.Text = sub
		cv.sub.Refresh()
	})
}

func (cv *ChatView) loadMessages() {
	if cv.chatID == "" || cv.fa == nil || cv.fa.api == nil {
		return
	}
	msgs, err := cv.fa.api.FetchMessagesPaged(cv.chatID, 50, 0, "")
	if err != nil {
		log.Println("chatview: FetchMessagesPaged:", err)
		return
	}
	if msgs == nil {
		msgs = []store.DecodedMessage{}
	}
	// ensure chronological for display
	// keep as-is if already fine; try simple reverse detection
	cv.messages = msgs
	fyne.Do(func() {
		cv.list.Refresh()
		if len(cv.messages) > 0 {
			cv.list.ScrollToBottom()
		}
	})
}

func (cv *ChatView) listenEvents() {
	ch := SubscribeEvent("wa:new_message")
	for {
		select {
		case <-cv.fa.ctx.Done():
			return
		case data := <-ch:
			if data == nil {
				continue
			}
			m, ok := data.(map[string]any)
			if !ok {
				cv.loadMessages()
				continue
			}
			chatID, _ := m["chatId"].(string)
			if chatID == "" {
				chatID, _ = m["chat_id"].(string)
			}
			if chatID != "" && chatID != cv.chatID {
				continue
			}
			if msg, ok := m["message"]; ok {
				switch v := msg.(type) {
				case store.DecodedMessage:
					cv.messages = append(cv.messages, v)
				case *store.DecodedMessage:
					if v != nil {
						cv.messages = append(cv.messages, *v)
					}
				default:
					cv.loadMessages()
					continue
				}
				fyne.Do(func() {
					cv.list.Refresh()
					if len(cv.messages) > 0 {
						cv.list.ScrollToBottom()
					}
				})
			} else {
				cv.loadMessages()
			}
		}
	}
}

func (cv *ChatView) Content() fyne.CanvasObject {
	av := canvas.NewCircle(waHover)
	avWrap := container.NewGridWrap(fyne.NewSize(40, 40), av)
	headerText := container.NewVBox(cv.header, cv.sub)
	header := container.NewMax(canvas.NewRectangle(waPanel2), container.NewPadded(container.NewBorder(nil, nil, avWrap, nil, headerText)))

	sendBtn := widget.NewButton("Send", func() { cv.input.OnSubmitted(cv.input.Text) })
	sendBtn.Importance = widget.SuccessImportance
	inputBg := canvas.NewRectangle(waPanel2)
	bottom := container.NewMax(inputBg, container.NewPadded(container.NewBorder(nil, nil, nil, sendBtn, cv.input)))

	bg := canvas.NewRectangle(waBg)
	return container.NewMax(bg, container.NewBorder(header, bottom, nil, nil, cv.list))
}

var _ = fmt.Sprint
