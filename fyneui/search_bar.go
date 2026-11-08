package fyneui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// SearchBar from ChatScreen.tsx:78-95. Measured in headless Chrome:
//
//	outer  px-3 py-2            → h 52, padding 8px 12px
//	pill   rounded-full px-4 py-2 → h 36, padding 8px 16px, bg #242626
//	icon   20x20, mr-4          → 16px gap to the input
//	input  text-sm              → 14px / line-height 20
const (
	searchOuterPadX float32 = 12
	searchOuterPadY float32 = 8
	searchPillH     float32 = 36
	searchPillPadX  float32 = 16
	searchIconSize  float32 = 20
	searchIconMR    float32 = 16
	searchFontSize  float32 = 14
	searchBarHeight float32 = searchPillH + searchOuterPadY*2 // 52
	// searchTextLift corrects for widget.Entry's internal top inset; measured
	// so the placeholder ink lands on rows 13..24 like React.
	searchTextLift float32 = -7
)

// searchBar is the sidebar's search field. Pure constructor: text leaves via
// onChange, never by reading Sidebar state.
type searchBar struct {
	widget.BaseWidget

	entry       *widget.Entry
	entryChrome fyne.CanvasObject
	bg          *canvas.Rectangle
	icon        fyne.CanvasObject
	placeholder string
	onChange    func(string)
	silent      bool // true while SetText runs, so it does not echo onChange
}

func newSearchBar(placeholder string, onChange func(string)) *searchBar {
	s := &searchBar{placeholder: placeholder, onChange: onChange}

	s.bg = canvas.NewRectangle(waPanel2)
	s.bg.CornerRadius = searchPillH / 2
	s.icon = iconImage("search-bar-icon", iconSearch, searchIconSize, iconSearchMuted)

	s.entry = widget.NewEntry()
	s.entry.SetPlaceHolder(placeholder)
	s.entry.TextStyle = fyne.TextStyle{}
	s.entry.OnChanged = func(v string) {
		if s.silent {
			return
		}
		if s.onChange != nil {
			s.onChange(v)
		}
	}
	// The pill (s.bg) owns the chrome. Strip Entry's own InputBackground/border/focus
	// so the typed text and hit area stay in sync with the pill, not a second box.
	s.entryChrome = container.NewThemeOverride(s.entry, &borderlessEntryTheme{Theme: fyne.CurrentApp().Settings().Theme(), bg: color.Transparent})
	s.ExtendBaseWidget(s)
	return s
}

// SetPlaceholder swaps the hint text without firing onChange — ChatScreen.tsx
// does this per view (communities / channels / chats).
func (s *searchBar) SetPlaceholder(p string) {
	if p == s.placeholder {
		return
	}
	s.placeholder = p
	fyne.Do(func() {
		s.entry.SetPlaceHolder(p)
		s.entry.Refresh()
	})
}

// SetText updates the field programmatically, mirroring the controlled
// <input value={searchTerm}>: no onChange echo.
func (s *searchBar) SetText(v string) {
	// widget.Entry.SetText fires OnChanged; the React input is controlled, so a
	// programmatic value must not bounce back into the store.
	s.silent = true
	s.entry.SetText(v)
	s.silent = false
}

func (s *searchBar) Text() string { return s.entry.Text }

func (s *searchBar) CreateRenderer() fyne.WidgetRenderer {
	return &searchBarRenderer{s: s, objs: []fyne.CanvasObject{s.bg, s.icon, s.entryChrome}}
}

type searchBarRenderer struct {
	s    *searchBar
	objs []fyne.CanvasObject
}

func (r *searchBarRenderer) Layout(size fyne.Size) {
	pillW := size.Width - searchOuterPadX*2
	r.s.bg.Resize(fyne.NewSize(pillW, searchPillH))
	r.s.bg.Move(fyne.NewPos(searchOuterPadX, searchOuterPadY))

	iconY := searchOuterPadY + (searchPillH-searchIconSize)/2
	r.s.icon.Resize(fyne.NewSize(searchIconSize, searchIconSize))
	r.s.icon.Move(fyne.NewPos(searchOuterPadX+searchPillPadX, iconY))

	// widget.Entry draws its text with a fixed ~12px inset from the top of its
	// box, so centring the box would push the glyphs to the bottom of the pill.
	// React centres a 20px line box (ink at y 13..24); lifting the entry by
	// entryInkInset-searchOuterPadY puts Fyne's ink on the same rows.
	entryX := searchOuterPadX + searchPillPadX + searchIconSize + searchIconMR
	entryW := pillW - searchPillPadX*2 - searchIconSize - searchIconMR
	r.s.entryChrome.Resize(fyne.NewSize(entryW, searchPillH))
	r.s.entryChrome.Move(fyne.NewPos(entryX, searchOuterPadY+searchTextLift))
}

func (r *searchBarRenderer) MinSize() fyne.Size {
	return fyne.NewSize(0, searchBarHeight)
}

func (r *searchBarRenderer) Refresh() {
	r.s.bg.FillColor = waPanel2
	r.s.bg.Refresh()
	canvas.Refresh(r.s)
}

func (r *searchBarRenderer) Objects() []fyne.CanvasObject { return r.objs }

func (r *searchBarRenderer) Destroy() {}

// iconSearchMuted is Tailwind's gray-400 in the dark palette — the search glyph
// sits one step brighter than waMuted in the React markup
// (text-gray-500 dark:text-gray-400).
var iconSearchMuted = color.NRGBA{R: 0x99, G: 0xa1, B: 0xaf, A: 255}
