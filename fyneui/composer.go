package fyneui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// newComposer mirrors ChatInput.tsx:419-566 (pill + separate send circle).
// Minimal parity surface (YAGNI): emoji/attach are callbacks; advanced features
// (mention, paste image, attach menu, emoji picker) are out of this component.
func newComposer(onSend func(string), onEmoji, onAttach func()) (fyne.CanvasObject, *widget.Entry) {
	entry := widget.NewMultiLineEntry()
	entry.SetPlaceHolder("Message")
	entry.Wrapping = fyne.TextWrapWord
	entry.SetMinRowsVisible(1)

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

	emojiBtn := newIconButton("cv-emoji", iconEmoji, 20, waMuted, onEmoji)
	attachBtn := newIconButton("cv-attach", iconAttach, 20, waMuted, onAttach)

	pillBg := canvas.NewRectangle(waPanel2)
	pillBg.CornerRadius = 22
	pill := container.NewMax(pillBg, container.New(&padLayout{h: 6, v: 4},
		container.NewBorder(nil, nil,
			container.NewHBox(container.NewPadded(emojiBtn)), nil,
			container.NewBorder(nil, nil, nil, container.NewPadded(attachBtn), entry))))

	sendBtn := newSendCircle(send)
	// ChatInput.tsx: outer mx-3 mb-3 mt-1.5, row gap-2, send mb-0.5 h-11 w-11
	row := container.NewBorder(nil, nil, nil, container.New(&padLayout{h: 4, v: 2}, sendBtn), pill)
	outer := container.New(&padLayout{h: 12, v: 0}, container.NewBorder(spacer(1, 6), spacer(1, 12), nil, nil, row))
	return outer, entry
}

func newSendCircle(onTap func()) fyne.CanvasObject {
	circle := canvas.NewCircle(waAccent) // #1b9a58
	icon := iconImage("cv-send", iconSend, 20, waWhite)
	btn := newTapArea(onTap)
	inner := container.NewMax(circle, container.NewCenter(icon), btn)
	return container.NewGridWrap(fyne.NewSize(44, 44), inner) // 44px ~ h-11 w-11
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
