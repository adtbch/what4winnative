package fyneui

import (
	"image"
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	"github.com/lugvitc/whats4linux/api"
)

// Measured from frontend/src/harness.tsx `out.archived` / `out.empty`
// (getBoundingClientRect + getComputedStyle, 221px sidebar):
//
//	archived: box 221x48, padding 12px 16px, gap 16px
//	          icon 20x20 at x=30, rgb(33,192,99)
//	          label 16px/500 rgb(233,237,239)
//	          count 12px rgb(134,150,160) at x=198
//	empty:    padding 32px, text 16px/24px, colour #99a1af
//	          button 128x40, padding 8px 16px, radius 8px, bg #00c950,
//	          white 16px/400, margin-top 16px
const (
	reactArchivedH      = 48
	reactArchivedIconX  = 30
	reactArchivedIconSz = 20
	reactArchivedCountX = 198
	reactEmptyBtnW      = 128
	reactEmptyBtnH      = 40
	reactEmptyGap       = 16
)

func TestArchivedEntryGolden(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	a.Settings().SetTheme(newWATheme())

	e := newArchivedEntry(1, nil)
	img := renderPNG(t, "archived_entry", e, fyne.NewSize(reactRowWidth, reactArchivedH))
	assertNotBlank(t, img, 0.01)

	if got := e.MinSize().Height; got != reactArchivedH {
		t.Errorf("row height %.0f, want %d", got, reactArchivedH)
	}

	// The archive glyph must be #21c063 and sit at x=30..49, vertically centred.
	assertGreenBox(t, img, reactArchivedIconX, reactArchivedIconX+reactArchivedIconSz,
		14, 34, waArchiveGreen)

	// Label ink starts after the 48px icon box + 16px gap: x = 16+48+16 = 80.
	if x := inkLeft(t, img, 70, 190, 0, reactArchivedH); x < 78 || x > 82 {
		t.Errorf("label ink starts at x=%d, want 80", x)
	}

	// The count is flush against the 16px right gutter: its ink must end at 205.
	if x := inkRight(t, img, 190, 221, 0, reactArchivedH); x < 200 || x > 206 {
		t.Errorf("count ink ends at x=%d, want ~205", x)
	}
}

// TestArchivedEntryVisibility pins the ChatScreen.tsx:859 guard:
// `!showArchived && view === "chats" && archivedCount > 0`.
func TestArchivedEntryVisibility(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	s := &Sidebar{view: "chats", allChats: []api.ChatElement{
		{Contact: api.Contact{JID: "1@s.whatsapp.net"}},
		{Contact: api.Contact{JID: "2@s.whatsapp.net"}, Archived: true},
	}}
	s.archivedE = newArchivedEntry(0, nil)
	s.syncChrome()
	if !s.archivedE.Visible() {
		t.Error("archived entry must show for view=chats with one archived chat")
	}
	if s.archivedE.count.Text != "1" {
		t.Errorf("count = %q, want 1", s.archivedE.count.Text)
	}

	// Entering the archive hides the entry (the back header takes over).
	s.archived = true
	s.syncChrome()
	if s.archivedE.Visible() {
		t.Error("archived entry must hide inside the archive view")
	}

	// Another view has no archived entry.
	s.archived = false
	s.view = "communities"
	s.syncChrome()
	if s.archivedE.Visible() {
		t.Error("archived entry must hide outside the chats view")
	}
}

func TestArchivedEntryHiddenAtZero(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	s := &Sidebar{}
	row := s.archivedRow()
	if row.Visible() {
		t.Error("archived entry must start hidden until a count arrives")
	}
	s.archivedE.SetCount(3)
	if s.archivedE.count.Text != "3" {
		t.Errorf("count = %q, want 3", s.archivedE.count.Text)
	}
}

func TestEmptyStateGolden(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	a.Settings().SetTheme(newWATheme())

	e := newSidebarEmptyState(false, nil)
	img := renderPNG(t, "empty_state", e, fyne.NewSize(reactRowWidth, 447))
	assertNotBlank(t, img, 0.005)

	// The CTA is `px-4 py-2 bg-green-500 rounded-lg`: 128x40, fill #00c950.
	b := e.btn
	if sz := b.MinSize(); sz.Width != reactEmptyBtnW || sz.Height != reactEmptyBtnH {
		t.Errorf("button MinSize %.0fx%.0f, want %dx%d", sz.Width, sz.Height, reactEmptyBtnW, reactEmptyBtnH)
	}
	// The button sits in the vertical centre of the 447px panel: 128x40 at
	// x=46..173, y=236..275 (React: same centre, `mt-4` below a 48px paragraph).
	assertFillBox(t, img, waBtnGreen, 230, 285)
}

// TestArchivedHeaderGeometry pins ChatScreen.tsx:876-887: 48px tall, back arrow
// at x=16, label after a 16px gap, hairline along the bottom edge.
func TestArchivedHeaderGeometry(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	a.Settings().SetTheme(newWATheme())

	h := newArchivedHeader(nil)
	img := renderPNG(t, "archived_header", h, fyne.NewSize(reactRowWidth, reactArchivedH))

	if sz := h.MinSize(); sz.Height != reactArchivedH {
		t.Errorf("header height %.0f, want %d", sz.Height, reactArchivedH)
	}
	// GoBackIcon path starts at x=4 of a 24px box placed at x=16 → ink at 20.
	if x := inkLeft(t, img, 0, 60, 0, reactArchivedH); x < 18 || x > 22 {
		t.Errorf("back arrow ink starts at x=%d, want 20", x)
	}
	// gap-4 after the 24px button: 16+24+16 = 56.
	if x := inkLeft(t, img, 50, reactRowWidth, 0, reactArchivedH); x < 54 || x > 58 {
		t.Errorf("label ink starts at x=%d, want 56", x)
	}
}

func TestEmptyStateStrings(t *testing.T) {
	cases := []struct {
		hasChats bool
		want     string
	}{
		{true, "No chats match your search."},
		{false, "No chats available. Start a conversation!"},
	}
	for _, c := range cases {
		if got := emptyMessage(c.hasChats); got != c.want {
			t.Errorf("emptyMessage(%v) = %q, want %q", c.hasChats, got, c.want)
		}
	}
}

func TestEmptyStateWraps(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	e := newSidebarEmptyState(false, nil)
	e.Resize(fyne.NewSize(reactRowWidth, 447))

	// p-8 at 221px leaves 157px, so the long string must break over two lines
	// exactly like React's 157x48 paragraph.
	visible := 0
	for _, t2 := range e.msg {
		if t2.Visible() && t2.Text != "" {
			visible++
		}
	}
	if visible != 2 {
		t.Errorf("wrapped lines = %d, want 2", visible)
	}
}

// assertGreenBox checks a solid colour occupies the given band.
func assertGreenBox(t *testing.T, img image.Image, x0, x1, y0, y1 int, want color.Color) {
	t.Helper()
	found := false
	for y := y0; y < y1 && !found; y++ {
		for x := x0; x < x1; x++ {
			px := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
			w := color.NRGBAModel.Convert(want).(color.NRGBA)
			if channelDiff(px.R, w.R) < 12 && channelDiff(px.G, w.G) < 12 && channelDiff(px.B, w.B) < 12 {
				found = true
				break
			}
		}
	}
	if !found {
		t.Errorf("no pixel of the archive green in x %d..%d y %d..%d", x0, x1, y0, y1)
	}
}

func assertFillBox(t *testing.T, img image.Image, want color.Color, y0, y1 int) {
	t.Helper()
	assertGreenBox(t, img, 40, reactRowWidth-40, y0, y1, want)
}

func inkLeft(t *testing.T, img image.Image, x0, x1, y0, y1 int) int {
	t.Helper()
	for x := x0; x < x1; x++ {
		for y := y0; y < y1; y++ {
			px := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
			if px.R > 90 && px.G > 90 && px.B > 90 {
				return x
			}
		}
	}
	return -1
}

func inkRight(t *testing.T, img image.Image, x0, x1, y0, y1 int) int {
	t.Helper()
	for x := x1 - 1; x >= x0; x-- {
		for y := y0; y < y1; y++ {
			px := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
			if px.R > 90 && px.G > 90 && px.B > 90 {
				return x
			}
		}
	}
	return -1
}
