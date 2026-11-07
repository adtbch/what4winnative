package fyneui

import (
	"image/color"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// Values measured from frontend/src/harness.tsx `out.archived` and `out.empty`
// (getBoundingClientRect + getComputedStyle) at a 221px sidebar:
//
//	archived row  h=48  padding=12px 16px  gap=16px
//	              icon 20x20 at x=30  rgb(33,192,99)
//	              label 16px/500  rgb(233,237,239)
//	              count 12px      rgb(134,150,160) at x=198
//	empty state   padding=32px  text 16px/24px  oklch(.707 .022 261.325)=#99a1af
//	              button 128x40  padding 8px 16px  radius 8px  bg #00c950
//	                     white 16px/400  margin-top 16px
const (
	archivedRowH     float32 = 48
	archivedPadX     float32 = 16
	archivedPadY     float32 = 12
	archivedGap      float32 = 16
	archivedIconBox  float32 = 48
	archivedIconSize float32 = 20
	archivedFontSize float32 = 16
	archivedCountSz  float32 = 12
	archivedBackSize float32 = 24

	emptyPad       float32 = 32
	emptyFontSize  float32 = 16
	emptyLineH     float32 = 24
	emptyBtnW      float32 = 128
	emptyBtnH      float32 = 40
	emptyBtnPadX   float32 = 16
	emptyBtnRadius float32 = 8
	emptyBtnGap    float32 = 16
)

// React: `text-[#1b9a58] dark:text-[#21c063]` — the dark token is the app's
// default, so the archive glyph uses the same green as the unread badge.
var (
	waArchiveGreen = color.NRGBA{R: 0x21, G: 0xc0, B: 0x63, A: 0xff}
	// tailwind `bg-green-500` resolves to oklch(0.723 0.219 149.579) = #00c950.
	waBtnGreen = color.NRGBA{R: 0x00, G: 0xc9, B: 0x50, A: 0xff}
	waBtnHover = color.NRGBA{R: 0x00, G: 0xa8, B: 0x43, A: 0xff} // green-600
	waEmptyFg  = color.NRGBA{R: 0x99, G: 0xa1, B: 0xaf, A: 0xff} // gray-400
)

// archivedEntry is the "Archived" row above the chat list
// (ChatScreen.tsx:859-875). It is a plain button: `flex w-full items-center
// gap-4 px-4 py-3`, an icon centred in a 48px block, a growing label, and the
// count pinned right.
type archivedEntry struct {
	widget.BaseWidget
	bg    *canvas.Rectangle
	icon  fyne.CanvasObject
	label *canvas.Text
	count *canvas.Text
	onTap func()
}

func newArchivedEntry(count int, onTap func()) *archivedEntry {
	e := &archivedEntry{onTap: onTap}
	e.ExtendBaseWidget(e)

	e.bg = canvas.NewRectangle(color.Transparent)
	e.icon = iconImage("archived-icon", iconArchive, archivedIconSize, waArchiveGreen)

	e.label = canvas.NewText("Archived", waText)
	e.label.TextSize = archivedFontSize
	e.label.TextStyle = fyne.TextStyle{Bold: true} // font-medium (Fyne has no 500)
	e.label.Alignment = fyne.TextAlignLeading

	e.count = canvas.NewText(strconv.Itoa(count), waMuted)
	e.count.TextSize = archivedCountSz
	e.count.Alignment = fyne.TextAlignTrailing

	return e
}

// SetCount updates the right-hand badge; the row hides itself at zero in React
// (`archivedCount > 0` guard), so callers decide visibility.
func (e *archivedEntry) SetCount(n int) {
	e.count.Text = strconv.Itoa(n)
	e.count.Refresh()
}

func (e *archivedEntry) CreateRenderer() fyne.WidgetRenderer {
	inner := container.New(&archivedLayout{}, e.icon, e.label, e.count)
	return widget.NewSimpleRenderer(container.NewMax(e.bg, inner))
}

func (e *archivedEntry) Tapped(*fyne.PointEvent) {
	if e.onTap != nil {
		e.onTap()
	}
}

func (e *archivedEntry) MouseIn(*desktop.MouseEvent) {
	e.bg.FillColor = waHover
	e.bg.Refresh()
}

func (e *archivedEntry) MouseOut() {
	e.bg.FillColor = color.Transparent
	e.bg.Refresh()
}

func (e *archivedEntry) Cursor() desktop.Cursor { return desktop.PointerCursor }

// archivedLayout is `flex items-center gap-4 px-4 py-3`: 16px padding, a 48px
// icon box, 16px gaps, and the count flush against the right gutter.
type archivedLayout struct{}

func (l *archivedLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) < 3 {
		return fyne.Size{}
	}
	return fyne.NewSize(archivedPadX*2+archivedIconBox+archivedGap+objects[1].MinSize().Width,
		archivedRowH)
}

func (l *archivedLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 3 {
		return
	}
	icon, label, count := objects[0], objects[1], objects[2]

	// w-12 flex justify-center: the glyph sits centred in its 48px box.
	icon.Resize(fyne.NewSize(archivedIconSize, archivedIconSize))
	icon.Move(fyne.NewPos(archivedPadX+(archivedIconBox-archivedIconSize)/2,
		(size.Height-archivedIconSize)/2))

	cw := count.MinSize().Width
	count.Resize(fyne.NewSize(cw, size.Height-archivedPadY*2))
	count.Move(fyne.NewPos(size.Width-archivedPadX-cw, archivedPadY))

	lx := archivedPadX + archivedIconBox + archivedGap
	label.Resize(fyne.NewSize(size.Width-archivedPadX-cw-archivedGap-lx, size.Height-archivedPadY*2))
	label.Move(fyne.NewPos(lx, archivedPadY))
}

// archivedHeader replaces the entry once you are inside the archive
// (ChatScreen.tsx:876-887): `flex items-center gap-4 border-b px-4 py-3` with a
// back arrow and the same "Archived" label.
type archivedHeader struct {
	widget.BaseWidget
	back   fyne.CanvasObject
	label  *canvas.Text
	line   *canvas.Rectangle
	onBack func()
}

func newArchivedHeader(onBack func()) *archivedHeader {
	h := &archivedHeader{}
	h.ExtendBaseWidget(h)
	// GoBackIcon is a 24px svg (header_icons.tsx:1-9) inside a plain <button>
	// with no hover chrome, so the layout owns the hit box.
	h.back = iconImage("archived-back", iconBack, archivedBackSize, waMuted)
	h.onBack = onBack

	h.label = canvas.NewText("Archived", waText)
	h.label.TextSize = archivedFontSize
	h.label.TextStyle = fyne.TextStyle{Bold: true}

	h.line = canvas.NewRectangle(waHeaderDivider)
	return h
}

func (h *archivedHeader) Tapped(*fyne.PointEvent) {
	if h.onBack != nil {
		h.onBack()
	}
}

func (h *archivedHeader) Cursor() desktop.Cursor { return desktop.PointerCursor }

func (h *archivedHeader) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.New(&archivedHeaderLayout{},
		h.back, h.label, h.line))
}

type archivedHeaderLayout struct{}

func (l *archivedHeaderLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(archivedPadX*2+archivedBackSize+archivedGap+60, archivedRowH)
}

func (l *archivedHeaderLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	back, label, line := objects[0], objects[1], objects[2]
	back.Resize(fyne.NewSize(archivedBackSize, archivedBackSize))
	back.Move(fyne.NewPos(archivedPadX, (size.Height-archivedBackSize)/2))

	lx := archivedPadX + archivedBackSize + archivedGap
	label.Resize(fyne.NewSize(size.Width-lx-archivedPadX, size.Height))
	label.Move(fyne.NewPos(lx, (size.Height-emptyLineH)/2))

	line.Resize(fyne.NewSize(size.Width, 1))
	line.Move(fyne.NewPos(0, size.Height-1))
}

// emptyState is the "no chats" panel (ChatScreen.tsx:361-374):
// `flex flex-col items-center justify-center h-full p-8`, a centred message,
// then a green refresh button 16px below.
type emptyState struct {
	widget.BaseWidget
	msg    []*canvas.Text // one per wrapped line; canvas.Text cannot wrap
	btn    *button
	hasAny bool
	width  float32
}

// emptyMaxLines bounds the paragraph: React renders two lines at 221px and the
// strings are fixed, so four is already generous.
const emptyMaxLines = 4

func newSidebarEmptyState(hasChats bool, onRefresh func()) *emptyState {
	e := &emptyState{hasAny: hasChats}
	for i := 0; i < emptyMaxLines; i++ {
		t := canvas.NewText("", waEmptyFg)
		t.TextSize = emptyFontSize
		t.Alignment = fyne.TextAlignCenter
		t.Hide()
		e.msg = append(e.msg, t)
	}
	e.btn = newButton("Refresh Chats", waBtnGreen, waBtnHover, onRefresh)
	e.ExtendBaseWidget(e)
	e.rewrap()
	return e
}

// SetHasChats switches between the two React strings: a search with no hits vs
// an account with nothing at all.
func (e *emptyState) SetHasChats(hasChats bool) {
	e.hasAny = hasChats
	e.rewrap()
}

// Resize re-wraps because the paragraph width follows the sidebar width, like
// the CSS line box does.
func (e *emptyState) Resize(size fyne.Size) {
	e.BaseWidget.Resize(size)
	if size.Width != e.width {
		e.width = size.Width
		e.rewrap()
	}
}

// rewrap fills msg with the words of the message, greedily broken at the padded
// content width (p-8 leaves width-64).
func (e *emptyState) rewrap() {
	innerW := e.width - emptyPad*2
	if innerW <= 0 {
		innerW = emptyBtnW // pre-layout: fall back to the button width
	}
	lines := wrapText(emptyMessage(e.hasAny), emptyFontSize, innerW)
	for i, t := range e.msg {
		if i < len(lines) {
			t.Text = lines[i]
			t.Show()
		} else {
			t.Text = ""
			t.Hide()
		}
		t.Refresh()
	}
	e.BaseWidget.Refresh()
}

// wrapText greedily breaks on spaces at the given width.
func wrapText(s string, size, width float32) []string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return nil
	}
	var out []string
	cur := words[0]
	for _, w := range words[1:] {
		try := cur + " " + w
		if fyne.MeasureText(try, size, fyne.TextStyle{}).Width <= width {
			cur = try
			continue
		}
		out = append(out, cur)
		cur = w
	}
	return append(out, cur)
}

// SetLoading mirrors the `disabled` + "Loading..." branch of the React button.
func (e *emptyState) SetLoading(loading bool) {
	if loading {
		e.btn.SetText("Loading...")
	} else {
		e.btn.SetText("Refresh Chats")
	}
	e.btn.SetEnabled(!loading)
}

func emptyMessage(hasChats bool) string {
	if hasChats {
		return "No chats match your search."
	}
	return "No chats available. Start a conversation!"
}

func (e *emptyState) CreateRenderer() fyne.WidgetRenderer {
	objs := make([]fyne.CanvasObject, 0, len(e.msg)+1)
	for _, t := range e.msg {
		objs = append(objs, t)
	}
	objs = append(objs, e.btn)
	return widget.NewSimpleRenderer(container.New(&emptyLayout{}, objs...))
}

// emptyLayout stacks the wrapped message lines and the button in the vertical
// centre — `flex flex-col items-center justify-center p-8` with `mt-4` between
// the paragraph and the CTA.
type emptyLayout struct{}

func (l *emptyLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) < 2 {
		return fyne.Size{}
	}
	btn := objects[len(objects)-1]
	var w, h float32
	for _, o := range objects[:len(objects)-1] {
		if m := o.MinSize(); m.Width > w {
			w = m.Width
		}
	}
	h = float32(len(objects)-1)*emptyLineH + emptyBtnGap
	if bw := btn.MinSize().Width; bw > w {
		w = bw
	}
	return fyne.NewSize(w+emptyPad*2, h+emptyBtnH+emptyPad*2)
}

func (l *emptyLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 2 {
		return
	}
	btn := objects[len(objects)-1]
	lines := objects[:len(objects)-1]

	var textH float32
	for _, o := range lines {
		if o.Visible() {
			textH += emptyLineH
		}
	}
	total := textH + emptyBtnGap + emptyBtnH
	top := (size.Height - total) / 2
	if top < emptyPad {
		top = emptyPad
	}

	innerW := size.Width - emptyPad*2
	if innerW < 0 {
		innerW = 0
	}
	y := top
	for _, o := range lines {
		if !o.Visible() {
			continue
		}
		o.Resize(fyne.NewSize(innerW, emptyLineH))
		o.Move(fyne.NewPos(emptyPad, y))
		y += emptyLineH
	}

	btn.Resize(fyne.NewSize(emptyBtnW, emptyBtnH))
	btn.Move(fyne.NewPos((size.Width-emptyBtnW)/2, y+emptyBtnGap))
}

// button is the `px-4 py-2 bg-green-500 text-white rounded-lg` CTA. Fyne's
// widget.Button paints its own theme chrome, so this draws the rectangle and
// label directly to keep the measured 128x40 / radius 8 / #00c950.
type button struct {
	widget.BaseWidget
	bg      *canvas.Rectangle
	label   *canvas.Text
	bgColor color.Color
	hover   color.Color
	enabled bool
	onTap   func()
}

func newButton(text string, bg, hover color.Color, onTap func()) *button {
	b := &button{bgColor: bg, hover: hover, enabled: true, onTap: onTap}
	b.ExtendBaseWidget(b)

	b.bg = canvas.NewRectangle(bg)
	b.bg.CornerRadius = emptyBtnRadius

	b.label = canvas.NewText(text, color.White)
	b.label.TextSize = emptyFontSize
	b.label.Alignment = fyne.TextAlignCenter

	return b
}

func (b *button) SetText(s string) {
	b.label.Text = s
	b.label.Refresh()
	b.Resize(fyne.NewSize(emptyBtnW, emptyBtnH))
}

func (b *button) SetEnabled(on bool) {
	b.enabled = on
	if on {
		b.bg.FillColor = b.bgColor
		b.label.Color = color.White
	} else {
		// disabled:opacity-50 over the green fill.
		b.bg.FillColor = color.NRGBA{R: 0x80, G: 0xe4, B: 0xa8, A: 0xff}
		b.label.Color = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x80}
	}
	b.bg.Refresh()
	b.label.Refresh()
}

func (b *button) MinSize() fyne.Size { return fyne.NewSize(emptyBtnW, emptyBtnH) }

func (b *button) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewMax(b.bg, container.NewCenter(b.label)))
}

func (b *button) Tapped(*fyne.PointEvent) {
	if b.enabled && b.onTap != nil {
		b.onTap()
	}
}

func (b *button) MouseIn(*desktop.MouseEvent) {
	if b.enabled {
		b.bg.FillColor = b.hover
		b.bg.Refresh()
	}
}

func (b *button) MouseOut() {
	if b.enabled {
		b.bg.FillColor = b.bgColor
		b.bg.Refresh()
	}
}

func (b *button) Cursor() desktop.Cursor { return desktop.PointerCursor }
