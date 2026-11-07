package fyneui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

// badgeLayout draws the unread pill: fixed 20px height, at least 20px wide,
// growing with the label plus 6px horizontal padding each side
// (ChatScreen.tsx:314 `min-w-5 h-5 px-1.5 rounded-full text-xs font-semibold`).
// "99+" measures 34x20 in React, a single digit 20x20.
type badgeLayout struct {
	bg  *canvas.Rectangle
	txt *canvas.Text
}

func (l *badgeLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	// The pill grows with its label, so MinSize must report the label width —
	// otherwise the parent border reserves a 20px square and "99+" spills left.
	return fyne.NewSize(badgeWidthFor(l.txt.Text, chatRowBadgeFont), chatRowBadgeH)
}

func (l *badgeLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	w := size.Width
	if w < chatRowBadgeH {
		w = chatRowBadgeH
	}
	l.bg.Resize(fyne.NewSize(w, chatRowBadgeH))
	l.bg.Move(fyne.NewPos(0, 0))
	l.bg.CornerRadius = chatRowBadgeH / 2
	l.bg.Refresh()

	// Fyne centres text with a couple of pixels of leading slack; nudge the
	// label so its ink sits on the pill's optical centre like React's line box.
	l.txt.Resize(fyne.NewSize(w, chatRowBadgeH))
	l.txt.Move(fyne.NewPos(0, chatRowBadgeTextLift))
	l.txt.Refresh()
}

// badgeWidthFor reports the pill width React would use for a label: the text
// width plus px-1.5 on both sides, floored at the 20px square.
func badgeWidthFor(label string, textSize float32) float32 {
	if label == "" {
		return chatRowBadgeH
	}
	w := fyne.MeasureText(label, textSize, fyne.TextStyle{Bold: true}).Width + chatRowBadgePadX*2
	if w < chatRowBadgeH {
		return chatRowBadgeH
	}
	return w
}

// rowActionsLayout is `flex items-center gap-2 shrink-0`: children keep their
// own width, right-aligned, vertically centred.
type rowActionsLayout struct{ gap float32 }

func (l *rowActionsLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	var w float32
	var h float32
	first := true
	for _, o := range objects {
		if !o.Visible() {
			continue
		}
		m := o.MinSize()
		if !first {
			w += l.gap
		}
		first = false
		w += actionWidth(o, m.Width)
		if m.Height > h {
			h = m.Height
		}
	}
	return fyne.NewSize(w, h)
}

// actionWidth returns the width an action should occupy. The badge pill's
// width follows its label, and a label change does not re-run the parent
// layout, so both MinSize and Layout derive it from the live text instead of
// trusting the child's cached MinSize.
func actionWidth(o fyne.CanvasObject, fallback float32) float32 {
	if c, ok := o.(*fyne.Container); ok {
		if _, isBadge := c.Layout.(*badgeLayout); isBadge {
			return badgeWidthFor(badgeLabel(c), chatRowBadgeFont)
		}
	}
	return fallback
}

func (l *rowActionsLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	visible := make([]fyne.CanvasObject, 0, len(objects))
	for _, o := range objects {
		if o.Visible() {
			visible = append(visible, o)
		}
	}
	// right-align: walk backwards from size.Width
	x := size.Width
	for i := len(visible) - 1; i >= 0; i-- {
		o := visible[i]
		m := o.MinSize()
		w := actionWidth(o, m.Width)
		x -= w
		o.Resize(fyne.NewSize(w, m.Height))
		o.Move(fyne.NewPos(x, (size.Height-m.Height)/2))
		x -= l.gap
	}
}

func badgeLabel(c *fyne.Container) string {
	for _, o := range c.Objects {
		if t, ok := o.(*canvas.Text); ok {
			return t.Text
		}
	}
	return ""
}

// rowBodyLayout is the ChatListItem body: `flex items-center px-4 py-3` with a
// 48px avatar, mr-4, and a text column that fills the rest minus the 16px right
// gutter. Spelled out because container.NewBorder would add 4px of theme
// padding on every side.
type rowBodyLayout struct{}

func (l *rowBodyLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) < 2 {
		return fyne.Size{}
	}
	h := objects[0].MinSize().Height + chatRowPadY*2
	if m := objects[1].MinSize().Height + chatRowPadY*2; m > h {
		h = m
	}
	return fyne.NewSize(chatRowPadX+chatAvatarSize+chatAvatarGap+objects[1].MinSize().Width+chatRowPadX, h)
}

func (l *rowBodyLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 2 {
		return
	}
	avatar, mid := objects[0], objects[1]
	avatar.Resize(fyne.NewSize(chatAvatarSize, chatAvatarSize))
	avatar.Move(fyne.NewPos(chatRowPadX, (size.Height-chatAvatarSize)/2))

	midX := chatRowPadX + chatAvatarSize + chatAvatarGap
	midW := size.Width - midX - chatRowPadX
	if midW < 0 {
		midW = 0
	}
	mid.Resize(fyne.NewSize(midW, size.Height-chatRowPadY*2))
	mid.Move(fyne.NewPos(midX, chatRowPadY))
}

// rowLineLayout is one line of the text column: `flex justify-between` with a
// growing, truncating left item and a shrink-0 right cluster. container.NewBorder
// is unusable here because it adds the theme's 4px padding on every edge, which
// stole 8px from the truncating label and made the preview drop a character.
type rowLineLayout struct{ gap float32 }

func (l *rowLineLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) < 2 {
		return fyne.Size{}
	}
	h := objects[0].MinSize().Height
	if r := objects[1].MinSize().Height; r > h {
		h = r
	}
	return fyne.NewSize(objects[0].MinSize().Width+objects[1].MinSize().Width+l.gap, h)
}

func (l *rowLineLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 2 {
		return
	}
	left, right := objects[0], objects[1]
	rw := right.MinSize().Width
	if rw > size.Width {
		rw = size.Width
	}
	rh := right.MinSize().Height
	right.Resize(fyne.NewSize(rw, rh))
	right.Move(fyne.NewPos(size.Width-rw, (size.Height-rh)/2))

	lw := size.Width - rw - l.gap
	if lw < 0 {
		lw = 0
	}
	left.Resize(fyne.NewSize(lw, size.Height))
	left.Move(fyne.NewPos(0, 0))
}
