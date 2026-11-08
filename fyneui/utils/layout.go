package utils

import (
	"crypto/sha1"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

// Shared layout and text-fitting helpers used by every component. Kept in one
// file so component files stay focused on a single visual unit.

// PadLayout insets its single child by h horizontally and v vertically,
// mirroring Tailwind's px-*/py-* pairs.
type PadLayout struct{ H, V float32 }

func (l *PadLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.Size{}
	}
	s := objects[0].MinSize()
	return fyne.NewSize(s.Width+l.H*2, s.Height+l.V*2)
}

func (l *PadLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objects {
		o.Move(fyne.NewPos(l.H, l.V))
		o.Resize(fyne.NewSize(MaxF(size.Width-l.H*2, 0), MaxF(size.Height-l.V*2, 0)))
	}
}

// MinSizeLayout forces a minimum box while stretching its child to fill.
type MinSizeLayout struct{ W, H float32 }

func (l *MinSizeLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	s := fyne.NewSize(l.W, l.H)
	if len(objects) > 0 {
		c := objects[0].MinSize()
		if c.Width > s.Width {
			s.Width = c.Width
		}
		if c.Height > s.Height {
			s.Height = c.Height
		}
	}
	return s
}

func (l *MinSizeLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objects {
		o.Move(fyne.NewPos(0, 0))
		o.Resize(size)
	}
}

// Spacer is an invisible fixed-size gap.
func Spacer(w, h float32) fyne.CanvasObject {
	if h <= 0 {
		h = 1
	}
	if w <= 0 {
		w = 1
	}
	return container.NewGridWrap(fyne.NewSize(w, h), canvas.NewRectangle(color.Transparent))
}

// avatarColorsLight / avatarColorsDark are the WhatsApp placeholder pastels,
// copied verbatim from frontend/src/lib/utils.ts (AVATAR_COLORS_LIGHT/DARK).
// Order is part of the contract: the index comes from a SHA-1 of the JID.
var avatarColorsLight = []color.NRGBA{
	{R: 0xE5, G: 0xCD, B: 0xB0, A: 255}, {R: 0xA8, G: 0xE0, B: 0xB8, A: 255},
	{R: 0x7E, G: 0xC4, B: 0xE8, A: 255}, {R: 0xF0, G: 0xB4, B: 0xB4, A: 255},
	{R: 0xC9, G: 0xB8, B: 0xE8, A: 255}, {R: 0xF5, G: 0xD0, B: 0x8A, A: 255},
	{R: 0x9D, G: 0xD9, B: 0xD0, A: 255}, {R: 0xE8, G: 0xB8, B: 0xD0, A: 255},
	{R: 0xB8, G: 0xD4, B: 0xA8, A: 255}, {R: 0xA8, G: 0xC8, B: 0xE8, A: 255},
	{R: 0xE8, G: 0xC8, B: 0xA0, A: 255}, {R: 0xD0, G: 0xE8, B: 0xA8, A: 255},
	{R: 0xD8, G: 0xB8, B: 0xE0, A: 255}, {R: 0xA8, G: 0xD8, B: 0xE8, A: 255},
	{R: 0xE8, G: 0xD0, B: 0xB8, A: 255}, {R: 0xB8, G: 0xE0, B: 0xD0, A: 255},
}

var avatarColorsDark = []color.NRGBA{
	{R: 0xC4, G: 0xA8, B: 0x82, A: 255}, {R: 0x6F, G: 0xB8, B: 0x8A, A: 255},
	{R: 0x5A, G: 0x9E, B: 0xC4, A: 255}, {R: 0xC4, G: 0x7A, B: 0x7A, A: 255},
	{R: 0x9A, G: 0x88, B: 0xC4, A: 255}, {R: 0xC4, G: 0xA8, B: 0x5A, A: 255},
	{R: 0x6B, G: 0xB0, B: 0xA8, A: 255}, {R: 0xC4, G: 0x88, B: 0xA8, A: 255},
	{R: 0x88, G: 0xB0, B: 0x78, A: 255}, {R: 0x78, G: 0x98, B: 0xC4, A: 255},
	{R: 0xC4, G: 0xA8, B: 0x78, A: 255}, {R: 0xA0, G: 0xB8, B: 0x78, A: 255},
	{R: 0xA8, G: 0x88, B: 0xB8, A: 255}, {R: 0x78, G: 0xB0, B: 0xC4, A: 255},
	{R: 0xC4, G: 0xB0, B: 0x90, A: 255}, {R: 0x88, G: 0xB8, B: 0xA8, A: 255},
}

// AvatarColor is the dark-theme variant (the Fyne UI only ships dark).
func AvatarColor(jid string) color.Color {
	return avatarColorFor(jid, true)
}

// avatarColorFor mirrors getAvatarColor(jid, dark) from lib/utils.ts: SHA-1 of
// the JID, first four bytes big-endian as uint32, modulo the palette length.
func avatarColorFor(jid string, dark bool) color.Color {
	palette := avatarColorsLight
	if dark {
		palette = avatarColorsDark
	}
	return palette[avatarColorIndex(jid, len(palette))]
}

// StackedLayout places children top-to-bottom at explicit heights separated by
// a fixed gap. Use it where the React source sets line heights through Tailwind
// classes, so the box height is deterministic instead of following whatever the
// current font metrics happen to measure.
type StackedLayout struct {
	Heights []float32
	Gap     float32
}

func (l *StackedLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	w := float32(0)
	for _, o := range objects {
		if m := o.MinSize().Width; m > w {
			w = m
		}
	}
	h := float32(0)
	for i := range objects {
		h += l.heightAt(i)
		if i > 0 {
			h += l.Gap
		}
	}
	return fyne.NewSize(w, h)
}

func (l *StackedLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	y := float32(0)
	for i, o := range objects {
		if i > 0 {
			y += l.Gap
		}
		h := l.heightAt(i)
		o.Move(fyne.NewPos(0, y))
		o.Resize(fyne.NewSize(size.Width, h))
		y += h
	}
}

func (l *StackedLayout) heightAt(i int) float32 {
	if i < len(l.Heights) {
		return l.Heights[i]
	}
	return 0
}

// TruncLabel is a single-line text that ellipsizes instead of overflowing.
// canvas.Text never clips, so a long chat name draws straight over the
// timestamp; React avoids this with `truncate`. The layout re-truncates the full
// string every time the widget is resized, and Set keeps the source text.
type TruncLabel struct {
	txt  *canvas.Text
	full string
}

func NewTruncLabel(textSize float32, col color.Color, bold bool) *TruncLabel {
	t := canvas.NewText("", col)
	t.TextSize = textSize
	t.TextStyle = fyne.TextStyle{Bold: bold}
	return &TruncLabel{txt: t}
}

func (l *TruncLabel) Set(s string) {
	l.full = s
	l.txt.Text = s
	l.txt.Refresh()
}

// obj is the canvas object to place in a container.
func (l *TruncLabel) obj() fyne.CanvasObject { return l.txt }

// Obj exposes the canvas object to other packages.
func (l *TruncLabel) Obj() fyne.CanvasObject { return l.txt }

// TruncLayout drives one TruncLabel.
type TruncLayout struct{ L *TruncLabel }

func (t *TruncLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if t.L.txt.Text == "" {
		return fyne.NewSize(0, t.L.txt.MinSize().Height)
	}
	return t.L.txt.MinSize()
}

func (t *TruncLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objects {
		o.Move(fyne.NewPos(0, 0))
		o.Resize(size)
	}
	if t.L.full != "" {
		clipped := TruncateToWidth(t.L.full, size.Width, t.L.txt.TextSize)
		if clipped != t.L.txt.Text {
			t.L.txt.Text = clipped
			t.L.txt.Refresh()
		}
	}
}

// avatarColorIndex reproduces colorIndex() from lib/utils.ts. An empty JID
// yields 0 there, so it must yield 0 here too.
func avatarColorIndex(jid string, n int) int {
	if jid == "" || n == 0 {
		return 0
	}
	sum := sha1.Sum([]byte(jid))
	v := uint32(sum[0])<<24 | uint32(sum[1])<<16 | uint32(sum[2])<<8 | uint32(sum[3])
	return int(v % uint32(n))
}

// initials is the single-letter fallback shown inside an avatar circle.
func initials(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "?"
	}
	r := []rune(name)
	return strings.ToUpper(string(r[0]))
}

// TruncateToWidth shortens s with an ellipsis until it fits maxW at the given
// text size. canvas.Text does not ellipsize on its own, so every truncating
// label must go through this instead of letting text overflow its column.
func TruncateToWidth(s string, maxW, size float32) string {
	if maxW <= 0 || s == "" {
		return s
	}
	if fyne.MeasureText(s, size, fyne.TextStyle{}).Width <= maxW {
		return s
	}
	const ellipsis = "…"
	ellW := fyne.MeasureText(ellipsis, size, fyne.TextStyle{}).Width
	if ellW >= maxW {
		return ellipsis
	}
	r := []rune(s)
	lo, hi := 0, len(r)
	for lo < hi {
		mid := (lo + hi + 1) / 2
		if fyne.MeasureText(string(r[:mid]), size, fyne.TextStyle{}).Width+ellW <= maxW {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	if lo <= 0 {
		return ellipsis
	}
	return strings.TrimRight(string(r[:lo]), " ") + ellipsis
}

func MaxF(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

// NewTruncLayout wraps a TruncLabel in a layout container driver.
func NewTruncLayout(l *TruncLabel) *TruncLayout { return &TruncLayout{L: l} }

// NewStackedLayout stacks children top-to-bottom at explicit heights.
func NewStackedLayout(heights []float32, gap float32) *StackedLayout {
	return &StackedLayout{Heights: heights, Gap: gap}
}

// NewMinSizeLayout enforces a minimum box size while stretching its child.
func NewMinSizeLayout(w, h float32) *MinSizeLayout {
	return &MinSizeLayout{W: w, H: h}
}
