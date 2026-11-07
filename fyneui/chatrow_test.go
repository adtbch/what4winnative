package fyneui

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

// Chat-row golden test. Every number asserted here was MEASURED from the real
// React component, not guessed: frontend/src/harness.tsx renders
// ChatListScreen with fixture data and reports getBoundingClientRect():
//
//	rows[].h       = 72        (Tailwind px-4 py-3 + 48px avatar)
//	rows[].padding = 12px 16px
//	rows[].w       = 319       (sidebar at 400px window, minus borders)
//
// Reference screenshot: fyneui/testdata/react/chatrow_dark_400x700.png
const (
	reactRowHeight  = 72
	reactRowPadX    = 16
	reactRowPadY    = 12
	reactAvatarSize = 48
	reactRowWidth   = 319
)

// newTestRow builds one row through the same constructor the list uses, without
// touching *api.Api (only rowRefs must be initialised).
func newTestRow(t *testing.T) fyne.CanvasObject {
	t.Helper()
	s := &Sidebar{rowRefs: make(map[fyne.CanvasObject]*chatRowRefs)}
	return s.newRow()
}

// TestChatRowHeightMatchesReact is the core parity gate for the list row.
func TestChatRowHeightMatchesReact(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	row := newTestRow(t)
	if got := row.MinSize().Height; got != reactRowHeight {
		t.Fatalf("row min height = %v, want %d (measured from React)", got, reactRowHeight)
	}

	img := renderPNG(t, "chatrow", row, fyne.NewSize(reactRowWidth, reactRowHeight))
	assertNotBlank(t, img, 0.02)

	// React rows have a transparent background (`rows[].bg = rgba(0,0,0,0)`) —
	// the list container paints behind them. The row must therefore NOT fill its
	// own box; only the avatar and text paint. The avatar circle is filled with
	// #37444c until updateRow assigns the JID-derived pastel.
	assertPixel(t, img, reactRowPadX+reactAvatarSize/2, reactRowHeight/2, color.NRGBA{R: 0x37, G: 0x44, B: 0x4c, A: 255}, 6)
}

// TestChatRowAvatarGeometry pins the 48px avatar and the 16px left inset that
// React gets from `w-12 h-12` + `px-4`.
func TestChatRowAvatarGeometry(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	row := newTestRow(t)
	img := renderPNG(t, "chatrow_geom", row, fyne.NewSize(reactRowWidth, reactRowHeight))

	avatarFill := color.NRGBA{R: 0x37, G: 0x44, B: 0x4c, A: 255}

	// Horizontal centre of the avatar must be the avatar fill.
	assertPixel(t, img, reactRowPadX+reactAvatarSize/2, reactRowHeight/2, avatarFill, 8)

	// Before the px-4 inset (x < 16) and after the 48px diameter (x > 64) the
	// canvas must NOT be avatar-coloured, which pins both the inset and the size.
	for _, x := range []int{reactRowPadX - 4, reactRowPadX + reactAvatarSize + 4} {
		px := color.NRGBAModel.Convert(img.At(x, reactRowHeight/2)).(color.NRGBA)
		if channelDiff(px.R, avatarFill.R) < 10 && channelDiff(px.G, avatarFill.G) < 10 && channelDiff(px.B, avatarFill.B) < 10 {
			t.Fatalf("pixel x=%d is still avatar fill %v; avatar must span %d..%d",
				x, px, reactRowPadX, reactRowPadX+reactAvatarSize)
		}
	}
}

// TestChatRowTextFits guards truncation: a very long name must not widen the row
// past the sidebar width, which is what happens when canvas.Text is left to
// overflow.
func TestChatRowTextFits(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	const long = "Vapers Semarangan 18+ dan teman-teman semua yang ikut"
	got := truncateToWidth(long, 180, 15)
	if w := fyne.MeasureText(got, 15, fyne.TextStyle{}).Width; w > 180 {
		t.Fatalf("name width %v exceeds 180", w)
	}
	if got == long {
		t.Fatalf("long name was not truncated")
	}
}
