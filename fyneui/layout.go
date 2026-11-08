package fyneui

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

// padLayout insets its single child by h horizontally and v vertically,
// mirroring Tailwind's px-*/py-* pairs.
type padLayout struct{ h, v float32 }

func (l *padLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.Size{}
	}
	s := objects[0].MinSize()
	return fyne.NewSize(s.Width+l.h*2, s.Height+l.v*2)
}

func (l *padLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objects {
		o.Move(fyne.NewPos(l.h, l.v))
		o.Resize(fyne.NewSize(maxF(size.Width-l.h*2, 0), maxF(size.Height-l.v*2, 0)))
	}
}

// minSizeLayout forces a minimum box while stretching its child to fill.
type minSizeLayout struct{ w, h float32 }

func (l *minSizeLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	s := fyne.NewSize(l.w, l.h)
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

func (l *minSizeLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objects {
		o.Move(fyne.NewPos(0, 0))
		o.Resize(size)
	}
}

// spacer is an invisible fixed-size gap.
func spacer(w, h float32) fyne.CanvasObject {
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

// avatarColor is the dark-theme variant (the Fyne UI only ships dark).
func avatarColor(jid string) color.Color {
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

// stackedLayout places children top-to-bottom at explicit heights separated by
// a fixed gap. Use it where the React source sets line heights through Tailwind
// classes, so the box height is deterministic instead of following whatever the
// current font metrics happen to measure.
type stackedLayout struct {
	heights []float32
	gap     float32
}

func (l *stackedLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
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
			h += l.gap
		}
	}
	return fyne.NewSize(w, h)
}

func (l *stackedLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	y := float32(0)
	for i, o := range objects {
		if i > 0 {
			y += l.gap
		}
		h := l.heightAt(i)
		o.Move(fyne.NewPos(0, y))
		o.Resize(fyne.NewSize(size.Width, h))
		y += h
	}
}

func (l *stackedLayout) heightAt(i int) float32 {
	if i < len(l.heights) {
		return l.heights[i]
	}
	return 0
}

// truncLabel is a single-line text that ellipsizes instead of overflowing.
// canvas.Text never clips, so a long chat name draws straight over the
// timestamp; React avoids this with `truncate`. The layout re-truncates the full
// string every time the widget is resized, and Set keeps the source text.
type truncLabel struct {
	txt  *canvas.Text
	full string
}

func newTruncLabel(textSize float32, col color.Color, bold bool) *truncLabel {
	t := canvas.NewText("", col)
	t.TextSize = textSize
	t.TextStyle = fyne.TextStyle{Bold: bold}
	return &truncLabel{txt: t}
}

func (l *truncLabel) Set(s string) {
	l.full = s
	l.txt.Text = s
	l.txt.Refresh()
}

// obj is the canvas object to place in a container.
func (l *truncLabel) obj() fyne.CanvasObject { return l.txt }

// truncLayout drives one truncLabel.
type truncLayout struct{ l *truncLabel }

func (t *truncLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if t.l.txt.Text == "" {
		return fyne.NewSize(0, t.l.txt.MinSize().Height)
	}
	return t.l.txt.MinSize()
}

func (t *truncLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objects {
		o.Move(fyne.NewPos(0, 0))
		o.Resize(size)
	}
	if t.l.full != "" {
		clipped := truncateToWidth(t.l.full, size.Width, t.l.txt.TextSize)
		if clipped != t.l.txt.Text {
			t.l.txt.Text = clipped
			t.l.txt.Refresh()
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

// truncateToWidth shortens s with an ellipsis until it fits maxW at the given
// text size. canvas.Text does not ellipsize on its own, so every truncating
// label must go through this instead of letting text overflow its column.
func truncateToWidth(s string, maxW, size float32) string {
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

func maxF(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}
