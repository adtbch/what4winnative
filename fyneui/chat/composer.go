package chat

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/adtbch/whats4linux/fyneui/theme"
	"github.com/adtbch/whats4linux/fyneui/utils"
)

// newComposer mirrors ChatInput.tsx:419-566 (pill + separate send circle).
// Minimal parity surface (YAGNI): emoji/attach are callbacks; advanced features
// (mention, paste image, attach menu, emoji picker) are out of this component.
// composerEntry intercepts paste to support image clipboard (Windows DIB).
type composerEntry struct {
	widget.Entry
	onPasteImage func(pngData []byte)
}

func newComposerEntry(onPasteImage func([]byte)) *composerEntry {
	e := &composerEntry{onPasteImage: onPasteImage}
	e.MultiLine = true
	e.Wrapping = fyne.TextWrapWord
	e.ExtendBaseWidget(e)
	e.SetPlaceHolder("Message")
	e.SetMinRowsVisible(1)
	return e
}

func (e *composerEntry) TypedShortcut(s fyne.Shortcut) {
	if _, ok := s.(*fyne.ShortcutPaste); ok && e.onPasteImage != nil {
		if pngData, err := utils.ReadClipboardImagePNG(); err == nil && len(pngData) > 0 {
			e.onPasteImage(pngData)
			return
		}
		// fall through to default text paste
	}
	e.Entry.TypedShortcut(s)
}

func newComposer(onSend func(string), onEmoji, onAttach func(), onPasteImage func([]byte)) (fyne.CanvasObject, *composerEntry) {
	entry := newComposerEntry(onPasteImage)

	send := func() {
		text := entry.Text
		if text == "" {
			return
		}
		entry.SetText("")
		if onSend != nil {
			onSend(text)
		}
	}
	entry.OnSubmitted = func(string) { send() }

	emojiBtn := utils.NewIconButton("cv-emoji", utils.IconEmoji, 20, theme.WaMuted, onEmoji)
	attachBtn := utils.NewIconButton("cv-attach", utils.IconAttach, 20, theme.WaMuted, onAttach)

	pillBg := canvas.NewRectangle(theme.WaPanel2)
	pillBg.CornerRadius = 22
	pill := container.NewMax(pillBg, container.New(&utils.PadLayout{H: 6, V: 4},
		container.NewBorder(nil, nil,
			container.NewHBox(container.NewPadded(emojiBtn)), nil,
			container.NewBorder(nil, nil, nil, container.NewPadded(attachBtn), entry))))

	sendBtn := newSendCircle(send)
	// ChatInput.tsx: outer mx-3 mb-3 mt-1.5, row gap-2, send mb-0.5 h-11 w-11
	row := container.NewBorder(nil, nil, nil, container.New(&utils.PadLayout{H: 4, V: 2}, sendBtn), pill)
	outer := container.New(&utils.PadLayout{H: 12, V: 0}, container.NewBorder(utils.Spacer(1, 6), utils.Spacer(1, 12), nil, nil, row))
	return outer, entry
}

func newSendCircle(onTap func()) fyne.CanvasObject {
	circle := canvas.NewCircle(theme.WaAccent) // #1b9a58
	icon := utils.IconImage("cv-send", utils.IconSend, 20, theme.WaWhite)
	btn := NewTapArea(onTap)
	inner := container.NewMax(circle, container.NewCenter(icon), btn)
	return container.NewGridWrap(fyne.NewSize(44, 44), inner) // 44px ~ h-11 w-11
}

type tapArea struct {
	widget.BaseWidget
	onTap   func()
	minSize fyne.Size
}

func NewTapArea(onTap func()) *tapArea {
	t := &tapArea{onTap: onTap}
	t.ExtendBaseWidget(t)
	return t
}
func (t *tapArea) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(canvas.NewRectangle(color.Transparent))
}

func (t *tapArea) MinSize() fyne.Size {
	if t.minSize.Width > 0 || t.minSize.Height > 0 {
		return t.minSize
	}
	return fyne.NewSize(44, 44)
}
func (t *tapArea) Tapped(*fyne.PointEvent) {
	if t.onTap != nil {
		t.onTap()
	}
}
func (t *tapArea) Cursor() desktop.Cursor { return desktop.PointerCursor }
