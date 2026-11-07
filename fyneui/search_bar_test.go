package fyneui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
)

// Values dumped by frontend/src/harness.tsx from the real SearchBar
// (ChatScreen.tsx:78-95) in headless Chrome at 400x700.
const (
	reactSearchBarHeight = 52
	reactSearchPillH     = 36
	reactSearchPillPadX  = 16
	reactSearchIcon      = 20
	reactSearchIconMR    = 16
	reactSearchFontSize  = 14
	reactSearchPillBg    = "#242626"
	reactSearchWidth     = 221
	reactSearchPillX     = 12
	reactSearchPillY     = 114
)

func TestSearchBarGeometryMatchesReact(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	s := newSearchBar("Search or start new chat", nil)
	w := test.NewWindow(container.NewMax(canvas.NewRectangle(waBg), s))
	// Fyne pads window content by 4px on every side; React has no such gutter.
	w.SetPadded(false)
	w.Resize(fyne.NewSize(reactSearchWidth, reactSearchBarHeight))

	if got := s.MinSize().Height; got != reactSearchBarHeight {
		t.Fatalf("search bar height = %v, want %v", got, reactSearchBarHeight)
	}
	if got := s.bg.Size().Height; got != reactSearchPillH {
		t.Fatalf("pill height = %v, want %v", got, reactSearchPillH)
	}
	if got := s.bg.Position().X; got != reactSearchPillX {
		t.Fatalf("pill x = %v, want %v", got, reactSearchPillX)
	}
	if got := s.bg.Size().Width; got != reactSearchWidth-reactSearchPillX*2 {
		t.Fatalf("pill width = %v, want %v", got, reactSearchWidth-reactSearchPillX*2)
	}
	if got := s.bg.CornerRadius; got != reactSearchPillH/2 {
		t.Fatalf("pill radius = %v, want %v (rounded-full)", got, reactSearchPillH/2)
	}
	// icon sits at pill x + px-4, vertically centred in the pill.
	if got, want := s.icon.Position().X, float32(reactSearchPillX+reactSearchPillPadX); got != want {
		t.Fatalf("icon x = %v, want %v", got, want)
	}
	if got, want := s.icon.Size().Width, float32(reactSearchIcon); got != want {
		t.Fatalf("icon size = %v, want %v", got, want)
	}
	// entry starts after icon + mr-4.
	if got, want := s.entry.Position().X, float32(reactSearchPillX+reactSearchPillPadX+reactSearchIcon+reactSearchIconMR); got != want {
		t.Fatalf("entry x = %v, want %v", got, want)
	}
	w.Close()
}

// TestSearchBarPillColour: the pill fill must be #242626 on a #0c0c0c page, and
// no 1px Entry outline may show through (React has no visible border).
func TestSearchBarPillColour(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	// The app theme maps InputBackground to the pill colour; without it Fyne's
	// default grey input rect shows through.
	a.Settings().SetTheme(newWATheme())
	s := newSearchBar("Search or start new chat", nil)
	img := renderPNG(t, "search_bar",
		container.NewMax(canvas.NewRectangle(waBg), s),
		fyne.NewSize(reactSearchWidth, reactSearchBarHeight))
	assertNotBlank(t, img, 0.02)

	// pill centre, away from the icon and text.
	c := img.At(reactSearchWidth/2, reactSearchBarHeight/2)
	r, g, b, _ := c.RGBA()
	if r>>8 != 0x24 || g>>8 != 0x26 || b>>8 != 0x26 {
		t.Fatalf("pill fill = #%02x%02x%02x, want %s", r>>8, g>>8, b>>8, reactSearchPillBg)
	}
	// outer gutter stays page background.
	c = img.At(2, reactSearchBarHeight/2)
	r, g, b, _ = c.RGBA()
	if r>>8 != 0x0c || g>>8 != 0x0c || b>>8 != 0x0c {
		t.Fatalf("outer gutter = #%02x%02x%02x, want #0c0c0c", r>>8, g>>8, b>>8)
	}
}

func TestSearchBarOnChangeAndSetText(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	var got []string
	s := newSearchBar("Search or start new chat", func(v string) { got = append(got, v) })

	s.entry.SetText("bud")
	if len(got) != 1 || got[0] != "bud" {
		t.Fatalf("onChange calls = %v, want [bud]", got)
	}

	// Programmatic update must not echo back, like the controlled input.
	s.SetText("budi")
	if len(got) != 1 {
		t.Fatalf("SetText fired onChange: %v", got)
	}
	if s.Text() != "budi" {
		t.Fatalf("Text() = %q, want budi", s.Text())
	}
}

// TestSearchBarPlaceholderPerView: ChatScreen.tsx swaps the hint per tab.
func TestSearchBarPlaceholderPerView(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	s := newSearchBar("Search or start new chat", nil)
	if got := s.entry.PlaceHolder; got != "Search or start new chat" {
		t.Fatalf("placeholder = %q", got)
	}
	s.SetPlaceholder("Search communities")
	if got := s.entry.PlaceHolder; got != "Search communities" {
		t.Fatalf("placeholder after SetPlaceholder = %q", got)
	}
}
