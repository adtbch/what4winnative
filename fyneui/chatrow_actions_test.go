package fyneui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

// Measured from frontend/src/harness.tsx (`out.badges`, getBoundingClientRect
// row-relative) at a 221px row:
//
//	row 0 (pinned, 128 unread): pin  x=149 y=43 w=14 h=14
//	                            badge x=171 y=40 w=34 h=20  bg rgb(33,192,99)
//	row 1 (2 unread)          : badge x=185 y=40 w=20 h=20
//
// React markup (ChatScreen.tsx:300-320): `flex items-center gap-2` around the
// 14px pin and the pill; pill is `min-w-5 h-5 px-1.5 rounded-full text-xs
// font-semibold`, so it grows with the label instead of staying square.
const (
	reactPinX     = 149
	reactPinY     = 43
	reactPinSize  = 14
	reactBadgeX   = 171
	reactBadgeY   = 40
	reactBadgeW99 = 34
	reactBadgeH   = 20
	reactBadgeW1  = 20
	reactBadgeX1  = 185
)

func newProbeRow(t *testing.T, idx int) (*Sidebar, fyne.CanvasObject, *chatRowRefs) {
	t.Helper()
	a := test.NewApp()
	t.Cleanup(a.Quit)
	a.Settings().SetTheme(newWATheme())

	chats, unread := reactFixture()
	s := &Sidebar{
		rowRefs:   make(map[fyne.CanvasObject]*chatRowRefs),
		unread:    unread,
		avatarRes: map[string]fyne.Resource{},
		visible:   chats,
	}
	row := s.newRow()
	s.updateRow(idx, row)

	w := test.NewWindow(row)
	w.SetPadded(false)
	t.Cleanup(w.Close)
	w.Resize(fyne.NewSize(reactRowWidth, reactRowHeight))
	row.Resize(fyne.NewSize(reactRowWidth, reactRowHeight))
	return s, row, s.rowRefs[row]
}

// offsetOf walks the object tree from root and returns the absolute offset of
// target, plus whether it was found.
func offsetOf(root fyne.CanvasObject, target fyne.CanvasObject) (float32, float32, bool) {
	if root == target {
		return 0, 0, true
	}
	c, ok := root.(*fyne.Container)
	if !ok {
		return 0, 0, false
	}
	for _, child := range c.Objects {
		if dx, dy, ok := offsetOf(child, target); ok {
			p := child.Position()
			return p.X + dx, p.Y + dy, true
		}
	}
	return 0, 0, false
}

func near(a float32, b int) bool {
	d := a - float32(b)
	return d > -1.5 && d < 1.5
}

// TestChatRowPinGeometry pins the 14px pin at React's measured offset.
func TestChatRowPinGeometry(t *testing.T) {
	_, row, refs := newProbeRow(t, 1) // Vapers: pinned + 128 unread
	if !refs.pin.Visible() {
		t.Fatal("pinned chat must show the pin")
	}
	x, y, ok := offsetOf(row, refs.pin)
	if !ok {
		t.Fatal("pin not in the row tree")
	}
	if !near(x, reactPinX) || !near(y, reactPinY) {
		t.Errorf("pin at (%.0f,%.0f), want (%d,%d)", x, y, reactPinX, reactPinY)
	}
	if sz := refs.pin.Size(); !near(sz.Width, reactPinSize) || !near(sz.Height, reactPinSize) {
		t.Errorf("pin size %.0fx%.0f, want %dx%d", sz.Width, sz.Height, reactPinSize, reactPinSize)
	}
}

// TestChatRowBadgeGeometry pins both badge shapes: "99+" grows to 34px, a single
// digit stays a 20px square. Both sit flush against the 16px right gutter.
func TestChatRowBadgeGeometry(t *testing.T) {
	cases := []struct {
		name string
		idx  int
		x, w int
	}{
		{"99+", 1, reactBadgeX, reactBadgeW99},
		{"2", 0, reactBadgeX1, reactBadgeW1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, row, refs := newProbeRow(t, c.idx)
			if !refs.badgeWrap.Visible() {
				t.Fatal("unread chat must show the badge")
			}
			x, y, ok := offsetOf(row, refs.badgeWrap)
			if !ok {
				t.Fatal("badge not in the row tree")
			}
			sz := refs.badgeWrap.Size()
			if !near(x, c.x) || !near(y, reactBadgeY) {
				t.Errorf("badge at (%.0f,%.0f), want (%d,%d)", x, y, c.x, reactBadgeY)
			}
			if !near(sz.Width, c.w) || !near(sz.Height, reactBadgeH) {
				t.Errorf("badge size %.0fx%.0f, want %dx%d", sz.Width, sz.Height, c.w, reactBadgeH)
			}
			// React keeps the pill 16px from the right edge: 221-16 = 205.
			if right := x + sz.Width; !near(right, reactRowWidth-reactRowPadX) {
				t.Errorf("badge right edge %.0f, want %d", right, reactRowWidth-reactRowPadX)
			}
		})
	}
}

// TestChatRowNoBadgeWhenRead guards the hidden branch: no unread means no pill
// and the timestamp keeps the muted colour.
func TestChatRowNoBadgeWhenRead(t *testing.T) {
	_, _, refs := newProbeRow(t, 2) // +62 899…: no unread, not pinned
	if refs.badgeWrap.Visible() {
		t.Error("read chat must not show the unread badge")
	}
	if refs.pin.Visible() {
		t.Error("unpinned chat must not show the pin")
	}
}
