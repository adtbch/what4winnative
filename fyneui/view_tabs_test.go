package fyneui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
)

// Measured from the React harness (see harness.tsx `out.tabs`).
//
// The bar's getBoundingClientRect() reports 57px, but that includes a 15px
// horizontal scrollbar: the four pills need 338px and the 319px sidebar only
// leaves 295px, so `overflow-x-auto` scrolls. clientHeight/scrollHeight are both
// 42 — 4px pt + 30px pill + 8px pb — which is the height the CSS actually asks
// for. Use 42; 57 would bake a headless-Chrome scrollbar into the layout.
const (
	reactTabBarHeight = 42
	reactTabHeight    = 30
	reactTabFontSize  = 14
	// Pill widths in CSS px, straight from getBoundingClientRect().
	// Fyne ships its own font, so a few px of drift is unavoidable; the test
	// tolerates it instead of pretending the metrics are identical.
	reactTabWidthTolerance = 6
)

var reactTabWidths = map[string]int{
	"chats":       61,
	"communities": 108,
	"channels":    82,
	"status":      63,
}

// TestViewTabsGeometryMatchesReact pins the pill height, font size and bar
// padding to the measured React values instead of hand-tuned numbers.
func TestViewTabsGeometryMatchesReact(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	tabs := newViewTabs("chats", nil)

	for _, p := range tabs.pills {
		if got := p.MinSize().Height; got != reactTabHeight {
			t.Errorf("%s pill height %v != React %d", p.label, got, reactTabHeight)
		}
		want := float32(reactTabWidths[p.label])
		if got := p.MinSize().Width; got < want-reactTabWidthTolerance || got > want+reactTabWidthTolerance {
			t.Errorf("%s pill width %v outside React %v ±%d",
				p.label, got, want, reactTabWidthTolerance)
		}
	}

	if got := tabs.MinSize().Height; got != reactTabBarHeight {
		t.Errorf("tab bar height %v != React %d", got, reactTabBarHeight)
	}
}

// TestViewTabsCapitalize: React's button has Tailwind `capitalize`, so the DOM
// text "chats" renders as "Chats".
func TestViewTabsCapitalize(t *testing.T) {
	cases := map[string]string{
		"chats":       "Chats",
		"communities": "Communities",
		"channels":    "Channels",
		"status":      "Status",
	}
	for in, want := range cases {
		if got := capitalizeWords(in); got != want {
			t.Errorf("capitalizeWords(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestViewTabsSetCurrentDoesNotFireOnChange mirrors React: setView is only called
// from the button's onClick, so a programmatic update must stay silent.
func TestViewTabsSetCurrentDoesNotFireOnChange(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	var fired []string
	tabs := newViewTabs("chats", func(v string) { fired = append(fired, v) })

	tabs.SetCurrent("channels")
	if len(fired) != 0 {
		t.Fatalf("SetCurrent fired onChange: %v", fired)
	}
	if tabs.current != "channels" {
		t.Fatalf("current = %q, want channels", tabs.current)
	}
	if !tabs.pills[2].active || tabs.pills[0].active {
		t.Fatalf("highlight did not follow SetCurrent: %v %v",
			tabs.pills[0].active, tabs.pills[2].active)
	}

	// A real tap fires exactly once.
	tabs.pills[1].Tapped(nil)
	if len(fired) != 1 || fired[0] != "communities" {
		t.Fatalf("tap fired %v, want [communities]", fired)
	}
}

// TestViewTabsTapActiveIsNoop: tapping the already-selected pill must not
// re-fire onChange, matching React where setView(same) is a no-op render.
func TestViewTabsTapActiveIsNoop(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	n := 0
	tabs := newViewTabs("chats", func(string) { n++ })
	tabs.pills[0].Tapped(nil)
	if n != 0 {
		t.Fatalf("tapping the active pill fired onChange %d times", n)
	}
}

// TestViewTabsGolden renders the bar so the PNG can be compared with the React
// screenshot's tab row.
func TestViewTabsGolden(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	tabs := newViewTabs("chats", nil)
	// Render on the app's real background: the default Fyne theme is #4c4e4f and
	// would make every colour comparison meaningless.
	bg := canvas.NewRectangle(waBg)
	img := renderPNG(t, "view_tabs", container.NewMax(bg, tabs), fyne.NewSize(319, reactTabBarHeight))
	assertNotBlank(t, img, 0.02)

	// The active pill is filled with #e9edef; the idle ones keep the dark canvas.
	assertPixel(t, img, viewTabBarPadX+8, viewTabBarPadT+reactTabHeight/2,
		waPillOn, 6)

	// Right of the active pill the canvas must not be the light chip colour.
	px := img.At(viewTabBarPadX+70, viewTabBarPadT+reactTabHeight/2)
	if r, g, b, _ := px.RGBA(); uint8(r>>8) > 0xd0 && uint8(g>>8) > 0xd0 && uint8(b>>8) > 0xd0 {
		t.Fatalf("pixel past the active pill is still the chip colour: %v", px)
	}
}
