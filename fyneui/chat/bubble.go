package chat

import (
	"image/color"
	"strings"
	"unicode/utf8"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/lugvitc/whats4linux/fyneui/theme"
	"github.com/lugvitc/whats4linux/fyneui/utils"
)

// bubbleData carries everything MessageItem.tsx needs to render a text bubble.
// FromMe/FirstInGroup/Edited drive layout and corners; EmojiOnly drives the
// 32px/leading-10 treatment (MessageItem.tsx:283).
type bubbleData struct {
	Text         string // preprocessed HTML (api.processMessageText output)
	Time         string
	FromMe       bool
	FirstInGroup bool
	Edited       bool
	EmojiOnly    bool
	Sender       string      // group label; only rendered when !FromMe && FirstInGroup
	SenderColor  color.Color // GetProfileColor / misc.GetProfileColor
	MaxWidth     float32
}

// bubbleBg returns theme.WaSent/theme.WaRecv; bubbleCorner returns the 2px sharp corner
// that faces the sender on the first bubble of a run (rounded-tl-sm /
// rounded-tr-sm in MessageItem.tsx:445-446; every other corner stays 12px).
func bubbleBg(fromMe bool) color.Color {
	if fromMe {
		return theme.WaSent
	}
	return theme.WaRecv
}

func bubbleCorner(fromMe, firstInGroup bool) float32 { return 4 }

// emojiOnlyText mirrors the stripped+EMOJI_ONLY_RE test in MessageItem.tsx:272-277.
func emojiOnlyText(text string) bool {
	stripped := strings.TrimSpace(htmlToText(text))
	if stripped == "" || utf8.RuneCountInString(stripped) > 16 {
		return false
	}
	for _, r := range stripped {
		switch {
		case r == 0x20 || r == 0xFE0F || r == 0x200D:
		case r >= 0x2000 && r <= 0x2BFF: // misc symbols, dingbats, arrows
		case r >= 0x1F000 && r <= 0x1FAFF: // extended pictographic
		case r >= 0x1F1E6 && r <= 0x1F1FF: // regional indicators
		default:
			return false
		}
	}
	return true
}

type bubble struct {
	widget.BaseWidget
	d   bubbleData
	bg  *canvas.Rectangle
	blk fyne.CanvasObject
}

func newBubble(d bubbleData) *bubble {
	b := &bubble{d: d}
	b.ExtendBaseWidget(b)
	b.bg = canvas.NewRectangle(bubbleBg(d.FromMe))
	r := bubbleCorner(d.FromMe, d.FirstInGroup)
	// Canvas rectangle corners are uniform, so we model the sharp corner as
	// 2px overall and rely on the chat's tight spacing to read the same way
	// React does. 0 would make the whole bubble square.
	b.bg.CornerRadius = r
	if r < 4 {
		b.bg.CornerRadius = 4
	}

	inner := d.MaxWidth - 28
	if inner < 60 {
		inner = 60
	}

	src := strings.TrimSpace(d.Text)
	switch {
	case src == "":
		t := canvas.NewText("Empty Message", theme.WaMuted)
		t.TextSize = 14
		t.TextStyle = fyne.TextStyle{Italic: true}
		b.blk = utils.WrapFixed(t, inner, 19)
	case d.EmojiOnly:
		t := canvas.NewText(htmlToText(src), theme.WaText)
		t.TextSize = 32
		b.blk = utils.WrapFixed(t, inner, 40)
	default:
		b.blk = htmlBubble(src, inner, 14)
	}
	return b
}

func (b *bubble) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(nil)
}

// htmlBubble renders message HTML at maxW using the same pre-wrap + RichText
// fix as ChatView.textBlock (htmltext.go).
func htmlBubble(src string, maxW, size float32) fyne.CanvasObject {
	return htmlBubbleSized(src, maxW, size)
}

// htmlBubbleSized is the shared RichText path for text bubbles and captions:
// pre-wrap at maxW, give RichText back its 2*innerPadding so it never
// re-wraps early, then pin the cell with utils.WrapFixed/GridWrap.
func htmlBubbleSized(src string, maxW, size float32) fyne.CanvasObject {
	inner := maxW
	if inner < 60 {
		inner = 60
	}
	if strings.TrimSpace(htmlToText(src)) == "" {
		return canvas.NewText(" ", theme.WaText)
	}
	box, w, h := htmlTextBlock(src, inner, size)
	if w < 24 {
		w = 24
	}
	return container.NewGridWrap(fyne.NewSize(minF(w+16, inner), h+16), box)
}
