package fyneui

import (
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

// avatarColor picks a deterministic fallback color for a JID, matching the
// palette used by the React avatars.
func avatarColor(jid string) color.Color {
	palette := []color.NRGBA{
		{R: 0x00, G: 0xa8, B: 0x84, A: 255},
		{R: 0x53, G: 0xbd, B: 0xeb, A: 255},
		{R: 0xd4, G: 0x76, B: 0xe8, A: 255},
		{R: 0xf7, G: 0x9c, B: 0x4c, A: 255},
		{R: 0xe5, G: 0x42, B: 0xa3, A: 255},
		{R: 0x06, G: 0xcf, B: 0x9c, A: 255},
		{R: 0x7f, G: 0x66, B: 0xff, A: 255},
	}
	var h uint32
	for _, c := range jid {
		h = h*31 + uint32(c)
	}
	return palette[h%uint32(len(palette))]
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
