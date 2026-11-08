package fyneui

import (
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// Measured from frontend/src/harness.tsx `out.chatHeader` (ChatHeader.tsx at a
// 584px window):
//
//	box      h=65  padding=12px  bg rgb(12,12,12)  border-b 1px white/5
//	back     24x24 at x=12, vertically centred
//	avatar   40x40 at x=64 (12 pad + 24 back + 16 mr-4 + 12 gap-3), radius full,
//	         bg oklch(.446 .03 256.802) = #4a5565 (gray-600)
//	name     16px/24/500 at x=116, colour oklch(.967 .003 264.542) = #f3f4f6
//	subtitle 12px/16     at x=116 y=36, rgb(134,150,160)
//	call     38x38 at x=506, p-2, icon 22, colour #d1d5dc (gray-300)
//	info     28x28 at x=544, p-1, icon 20
const (
	chatHeaderPad      float32 = 12
	chatHeaderAvatar   float32 = 40
	chatHeaderGap      float32 = 12
	chatHeaderBackSize float32 = 24
	chatHeaderBackMR   float32 = 16 // mr-4 on the back button
	chatHeaderNameSize float32 = 16
	chatHeaderNameLH   float32 = 24
	chatHeaderSubSize  float32 = 12
	chatHeaderSubLH    float32 = 16
	chatHeaderCallBox  float32 = 38
	chatHeaderInfoBox  float32 = 28
	chatHeaderBorderH  float32 = 1
)

var (
	chatHeaderBg       = waPanel
	chatHeaderNameFg   = color.NRGBA{R: 0xf3, G: 0xf4, B: 0xf6, A: 0xff} // gray-100
	chatHeaderAvatarBg = color.NRGBA{R: 0x4a, G: 0x55, B: 0x65, A: 0xff} // gray-600
	chatHeaderCallFg   = color.NRGBA{R: 0xd1, G: 0xd5, B: 0xdc, A: 0xff} // gray-300
)

// chatHeaderData is everything the header renders. Subtitle is optional; Call
// is nil for groups (ChatDetail.tsx:709 only passes onCallClick for contacts).
type chatHeaderData struct {
	Name     string
	Subtitle string
	Avatar   fyne.Resource
	Call     func()
}

// chatHeader mirrors components/chat/ChatHeader.tsx: a 12px padded row with a
// back arrow, a 40px avatar, name + optional subtitle, and the call/info
// actions pinned right.
type chatHeader struct {
	widget.BaseWidget

	bg    *canvas.Rectangle
	line  *canvas.Rectangle
	back  fyne.CanvasObject
	avBg  *canvas.Rectangle
	avImg *canvas.Image
	avTxt *canvas.Text
	name  *canvas.Text
	sub   *canvas.Text
	call  *iconHit
	info  *iconHit

	nameS  string
	onBack func()
	callCb func()
}

// iconHit is a bare icon with a fixed square hit box — React's <button> carries
// only a hover circle (`p-2 rounded-full hover:bg-...`), no border of its own.
type iconHit struct {
	widget.BaseWidget
	img    fyne.CanvasObject
	hover  *canvas.Rectangle
	box    float32
	onTap  func()
	hoverC color.Color
}

func newIconHit(name, path string, icon, box float32, c color.Color, onTap func()) *iconHit {
	h := &iconHit{box: box, onTap: onTap, hoverC: waHover}
	h.ExtendBaseWidget(h)
	h.img = iconImage(name, path, icon, c)
	h.hover = canvas.NewRectangle(color.Transparent)
	h.hover.CornerRadius = 999
	return h
}

func (h *iconHit) MinSize() fyne.Size { return fyne.NewSize(h.box, h.box) }

func (h *iconHit) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewMax(h.hover, container.NewCenter(h.img)))
}

func (h *iconHit) Tapped(*fyne.PointEvent) {
	if h.onTap != nil {
		h.onTap()
	}
}

func (h *iconHit) MouseIn(*desktop.MouseEvent) {
	h.hover.FillColor = h.hoverC
	h.hover.Refresh()
}

func (h *iconHit) MouseOut() {
	h.hover.FillColor = color.Transparent
	h.hover.Refresh()
}

func (h *iconHit) Cursor() desktop.Cursor { return desktop.PointerCursor }

// newChatHeader builds the header. onBack may be nil (desktop keeps the sidebar
// visible), onInfo opens the chat-info panel, d.Call is nil for groups.
func newChatHeader(onBack, onInfo func(), d chatHeaderData) *chatHeader {
	h := &chatHeader{onBack: onBack}
	h.ExtendBaseWidget(h)

	h.bg = canvas.NewRectangle(chatHeaderBg)
	h.line = canvas.NewRectangle(waHeaderDivider)

	// GoBackIcon is a 24px svg inside a bare <button> (`mr-4 md:hidden`).
	backImg := iconImage("ch-back", iconBack, chatHeaderBackSize, chatHeaderCallFg)
	h.back = newIconBox(backImg, onBack, chatHeaderBackSize, chatHeaderBackSize)

	h.avBg = func() *canvas.Rectangle{r:=canvas.NewRectangle(chatHeaderAvatarBg);r.CornerRadius=6;return r}()
	h.avImg = canvas.NewImageFromResource(nil)
	h.avImg.FillMode = canvas.ImageFillCover
	h.avTxt = canvas.NewText("", color.White)
	h.avTxt.TextSize = 16
	h.avTxt.TextStyle = fyne.TextStyle{Bold: true}
	h.avTxt.Alignment = fyne.TextAlignCenter

	h.name = canvas.NewText("", chatHeaderNameFg)
	h.name.TextSize = chatHeaderNameSize
	h.name.TextStyle = fyne.TextStyle{Bold: true}

	h.sub = canvas.NewText("", waMuted)
	h.sub.TextSize = chatHeaderSubSize

	h.call = newIconHit("ch-call", iconPhone, 22, chatHeaderCallBox, chatHeaderCallFg, func() {
		if h.callCb != nil {
			h.callCb()
		}
	})
	h.info = newIconHit("ch-info", iconInfo, 20, chatHeaderInfoBox, color.White, onInfo)

	h.Set(d)
	return h
}

// Set applies new header data. The call button only renders when d.Call is set,
// matching `{onCallClick && <button/>}`.
func (h *chatHeader) Set(d chatHeaderData) {
	h.nameS = d.Name
	h.name.Text = d.Name
	h.name.Refresh()

	if strings.TrimSpace(d.Subtitle) == "" {
		h.sub.Hide()
	} else {
		h.sub.Text = d.Subtitle
		h.sub.Show()
		h.sub.Refresh()
	}

	if d.Avatar != nil {
		h.avImg.Resource = d.Avatar
		h.avImg.Show()
		h.avTxt.Hide()
	} else {
		h.avImg.Hide()
		h.avTxt.Text = firstRuneUpper(d.Name)
		h.avTxt.Show()
	}

	h.callCb = d.Call
	if d.Call != nil {
		h.call.Show()
	} else {
		h.call.Hide()
	}
	h.Refresh()
}

func firstRuneUpper(s string) string {
	for _, r := range s {
		return strings.ToUpper(string(r))
	}
	return ""
}

func (h *chatHeader) CreateRenderer() fyne.WidgetRenderer {
	av := container.NewMax(h.avBg, h.avImg, container.NewCenter(h.avTxt))
	l := &chatHeaderLayout{back: h.back, av: av, name: h.name, nameS: &h.nameS,
		sub: h.sub, call: h.call, info: h.info, line: h.line}
	return widget.NewSimpleRenderer(container.New(l, h.bg, l.line, h.back, av, h.name, h.sub, h.call, h.info))
}

// chatHeaderLayout places the slots at their measured offsets instead of nesting
// containers — NewBorder/NewHBox add theme padding and break the 12px grid that
// `p-3` + `gap-3` produces in ChatHeader.tsx.
type chatHeaderLayout struct {
	bg, line, back, av, sub, call, info fyne.CanvasObject
	name                                *canvas.Text
	nameS                               *string
}

func (l *chatHeaderLayout) MinSize([]fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(chatHeaderPad*2+chatHeaderBackSize+chatHeaderBackMR+chatHeaderGap+
		chatHeaderAvatar+chatHeaderGap+40+chatHeaderCallBox+chatHeaderInfoBox,
		chatHeaderAvatar+chatHeaderPad*2+chatHeaderBorderH)
}

func (l *chatHeaderLayout) Layout(_ []fyne.CanvasObject, size fyne.Size) {
	l.line.Resize(fyne.NewSize(size.Width, chatHeaderBorderH))
	l.line.Move(fyne.NewPos(0, size.Height-chatHeaderBorderH))

	contentH := size.Height - chatHeaderBorderH
	cy := (contentH - chatHeaderAvatar) / 2
	if cy < 0 {
		cy = 0
	}

	// `mr-4 md:hidden`: the back arrow only exists below Tailwind's md (768px)
	// breakpoint; on desktop the sidebar is the navigation.
	showBack := l.back != nil && size.Width < 768
	if l.back != nil {
		if showBack {
			l.back.Show()
		} else {
			l.back.Hide()
		}
	}

	x := chatHeaderPad
	if showBack {
		l.back.Resize(fyne.NewSize(chatHeaderBackSize, chatHeaderBackSize))
		l.back.Move(fyne.NewPos(x, (contentH-chatHeaderBackSize)/2))
		x += chatHeaderBackSize + chatHeaderBackMR
	}

	x += chatHeaderGap
	l.av.Resize(fyne.NewSize(chatHeaderAvatar, chatHeaderAvatar))
	l.av.Move(fyne.NewPos(x, cy))
	x += chatHeaderAvatar + chatHeaderGap

	// Actions are pinned right in DOM order: call, then info.
	rx := size.Width - chatHeaderPad
	rx -= chatHeaderInfoBox
	l.info.Resize(fyne.NewSize(chatHeaderInfoBox, chatHeaderInfoBox))
	l.info.Move(fyne.NewPos(rx, (contentH-chatHeaderInfoBox)/2))
	if l.call.Visible() {
		rx -= chatHeaderCallBox
		l.call.Resize(fyne.NewSize(chatHeaderCallBox, chatHeaderCallBox))
		l.call.Move(fyne.NewPos(rx, (contentH-chatHeaderCallBox)/2))
	}

	textW := rx - x
	if textW < 0 {
		textW = 0
	}
	// `truncate` on the h2: cut to the available width, never overflow.
	if l.nameS != nil {
		if clipped := truncateToWidth(*l.nameS, textW, chatHeaderNameSize); clipped != l.name.Text {
			l.name.Text = clipped
			l.name.Refresh()
		}
	}
	l.name.Resize(fyne.NewSize(textW, chatHeaderNameLH))
	l.name.Move(fyne.NewPos(x, cy))
	if l.sub.Visible() {
		l.sub.Resize(fyne.NewSize(textW, chatHeaderSubLH))
		l.sub.Move(fyne.NewPos(x, cy+chatHeaderNameLH))
	}
}

// iconBox is a fixed-size hit box around a bare icon.
type iconBox struct {
	widget.BaseWidget
	inner fyne.CanvasObject
	onTap func()
	w, h  float32
}

// newIconBox must ExtendBaseWidget or Resize/Visible never propagate and the
// icon silently never paints.
func newIconBox(inner fyne.CanvasObject, onTap func(), w, h float32) *iconBox {
	b := &iconBox{inner: inner, onTap: onTap, w: w, h: h}
	b.ExtendBaseWidget(b)
	return b
}

func (t *iconBox) MinSize() fyne.Size { return fyne.NewSize(t.w, t.h) }

func (t *iconBox) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewCenter(t.inner))
}

func (t *iconBox) Tapped(*fyne.PointEvent) {
	if t.onTap != nil {
		t.onTap()
	}
}

func (t *iconBox) Cursor() desktop.Cursor { return desktop.PointerCursor }
