package fyneui

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/lugvitc/whats4linux/api"
	"github.com/lugvitc/whats4linux/internal/store"
)

type ChatView struct {
	fa       *FyneApp
	chatID   string
	messages []store.DecodedMessage
	list     *widget.List
	input    *widget.Entry
}

func NewChatView(fa *FyneApp, chatID string) *ChatView {
	cv := &ChatView{
		fa:       fa,
		chatID:   chatID,
		messages: []store.DecodedMessage{},
		input:    widget.NewEntry(),
	}

	cv.list = widget.NewList(
		func() int { return len(cv.messages) },
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < 0 || id >= len(cv.messages) {
				return
			}
			m := cv.messages[id]
			label := obj.(*widget.Label)
			text := ""
			if m.Content != nil {
				if m.Content.Conversation != "" {
					text = m.Content.Conversation
				} else if m.Content.ExtendedTextMessage != nil {
					text = m.Content.ExtendedTextMessage.Text
				} else if m.Content.ImageMessage != nil {
					text = "[image] " + m.Content.ImageMessage.Caption
				} else if m.Content.VideoMessage != nil {
					text = "[video] " + m.Content.VideoMessage.Caption
				} else if m.Content.AudioMessage != nil {
					text = "[audio]"
				} else if m.Content.DocumentMessage != nil {
					text = "[document] " + m.Content.DocumentMessage.FileName
				} else if m.Content.StickerMessage != nil {
					text = "[sticker]"
				}
			}
			if m.Info.Sender != "" && m.Info.IsFromMe {
				label.SetText("Me: " + text)
			} else if m.Info.Sender != "" {
				label.SetText(m.Info.Sender + ": " + text)
			} else {
				label.SetText(text)
			}
		},
	)

	cv.input.OnSubmitted = func(text string) {
		if text == "" || cv.chatID == "" {
			return
		}
		_, err := cv.fa.api.SendMessage(cv.chatID, api.MessageContent{
			Type: "text",
			Text: text,
		})
		if err != nil {
			log.Println("chatview: SendMessage:", err)
			return
		}
		fyne.Do(func() {
			cv.input.SetText("")
		})
	}

	go cv.loadMessages()
	go cv.listenEvents()
	return cv
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
	if len(msgs) > 0 {
		cv.messages = msgs
	} else {
		cv.messages = []store.DecodedMessage{}
	}
	fyne.Do(func() {
		cv.list.Refresh()
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
				continue
			}
			chatID, _ := m["chatId"].(string)
			if chatID != cv.chatID {
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
				}
				fyne.Do(func() {
					cv.list.Refresh()
				})
			} else {
				cv.loadMessages()
			}
		}
	}
}

func (cv *ChatView) Content() fyne.CanvasObject {
	sendBtn := widget.NewButton("Send", func() {
		cv.input.OnSubmitted(cv.input.Text)
	})
	bottom := container.NewBorder(nil, nil, nil, sendBtn, cv.input)
	return container.NewBorder(nil, bottom, nil, nil, cv.list)
}
