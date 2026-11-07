package fyneui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// Geometry measured from the React harness (getComputedStyle + rect):
//
//	container: padding 4px 12px 8px, gap 8px
//	pill:      h 30 (14px text + py-1), padding 4px 12px, fully rounded,
//	           border 1px, fontSize 14px
//	active:    bg #e9edef, text #0b141a, weight 500, border transparent
//	idle:      bg transparent, border rgba(255,255,255,.1), text #8696a0, weight 400
const (
	viewTabHeight   = 30
	viewTabPadX     = 12
	viewTabPadY     = 4
	viewTabGap      = 8
	viewTabBarPadT  = 4
	viewTabBarPadX  = 12
	viewTabBarPadB  = 8
	viewTabFontSize = 14
)

var viewTabLabels = []string{"chats", "communities", "channels", "status"}

// capitalizeWords mirrors Tailwind's `capitalize`: the button text is lowercase
// in the DOM ("chats") but renders as "Chats".
func capitalizeWords(s string) string {
	out := []rune(s)
	up := true
	for i, r := range out {
		if up && r >= 'a' && r <= 'z' {
			out[i] = r - 32
		}
		up = r == ' '
	}
	return string(out)
}

// viewTabs is ChatScreen.tsx's `flex gap-2 px-3 pb-2 pt-1 overflow-x-auto` row.
// SetCurrent is programmatic and must NOT fire onChange — only a real tap does,
// matching React where setView is only called from the button's onClick.
type viewTabs struct {
	widget.BaseWidget
	current  string
	onChange func(string)
	pills    []*viewTabPill
}

func newViewTabs(current string, onChange func(string)) *viewTabs {
	t := &viewTabs{current: current, onChange: onChange}
	t.ExtendBaseWidget(t)
	for _, label := range viewTabLabels {
		l := label
		p := newViewTabPill(l, l == current, func() { t.tap(l) })
		t.pills = append(t.pills, p)
	}
	return t
}

func (t *viewTabs) tap(label string) {
	if t.current == label {
		return
	}
	t.current = label
	t.refresh()
	if t.onChange != nil {
		t.onChange(label)
	}
}

// SetCurrent updates the highlight without firing onChange.
func (t *viewTabs) SetCurrent(label string) {
	if t.current == label {
		return
	}
	t.current = label
	t.refresh()
}

func (t *viewTabs) refresh() {
	for _, p := range t.pills {
		p.setActive(p.label == t.current)
	}
}

func (t *viewTabs) CreateRenderer() fyne.WidgetRenderer {
	objs := make([]fyne.CanvasObject, len(t.pills))
	for i, p := range t.pills {
		objs[i] = p
	}
	row := container.New(&viewTabsLayout{}, objs...)
	return widget.NewSimpleRenderer(container.New(&viewTabsBarLayout{}, row))
}

// viewTabsLayout places pills left to right with a fixed gap.
type viewTabsLayout struct{}

func (l *viewTabsLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	w, h := float32(0), float32(0)
	for i, o := range objects {
		s := o.MinSize()
		if i > 0 {
			w += viewTabGap
		}
		w += s.Width
		if s.Height > h {
			h = s.Height
		}
	}
	return fyne.NewSize(w, h)
}

func (l *viewTabsLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	x := float32(0)
	for _, o := range objects {
		s := o.MinSize()
		o.Move(fyne.NewPos(x, 0))
		// Pills keep their CSS height (30) even when the bar is given more room;
		// stretching them is what made them look "gendut" next to React.
		o.Resize(fyne.NewSize(s.Width, viewTabHeight))
		x += s.Width + viewTabGap
	}
}

// viewTabsBarLayout applies the container's px-3 pb-2 pt-1 padding.
type viewTabsBarLayout struct{}

func (l *viewTabsBarLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.Size{}
	}
	s := objects[0].MinSize()
	return fyne.NewSize(s.Width+viewTabBarPadX*2, s.Height+viewTabBarPadT+viewTabBarPadB)
}

func (l *viewTabsBarLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objects {
		o.Move(fyne.NewPos(viewTabBarPadX, viewTabBarPadT))
		o.Resize(fyne.NewSize(maxF(size.Width-viewTabBarPadX*2, 0), viewTabHeight))
	}
}

// viewTabPill is one `rounded-full border px-3 py-1 text-sm capitalize` button.
type viewTabPill struct {
	widget.BaseWidget
	label  string
	active bool
	onTap  func()

	bg     *canvas.Rectangle
	border *canvas.Rectangle
	text   *canvas.Text
}

func newViewTabPill(label string, active bool, onTap func()) *viewTabPill {
	p := &viewTabPill{label: label, active: active, onTap: onTap}
	p.ExtendBaseWidget(p)
	return p
}

func (p *viewTabPill) setActive(active bool) {
	if p.active == active {
		return
	}
	p.active = active
	p.applyStyle()
}

func (p *viewTabPill) applyStyle() {
	if p.bg == nil {
		return
	}
	if p.active {
		p.bg.FillColor = waPillOn
		p.bg.Show()
		p.border.Hide()
		p.text.Color = waPillOnText
		p.text.TextStyle = fyne.TextStyle{Bold: true}
	} else {
		p.bg.FillColor = color.Transparent
		p.border.StrokeColor = waPillBorder
		// Fyne centres the stroke on the rect edge, so a 1px stroke lands on a
		// half-pixel boundary and antialiases into two rows at half alpha. A 2px
		// stroke reproduces React's solid 1px border at this DPR.
		p.border.StrokeWidth = 2
		p.border.Show()
		p.text.Color = waMuted
		p.text.TextStyle = fyne.TextStyle{}
	}
	p.bg.Refresh()
	p.border.Refresh()
	p.text.Refresh()
}

func (p *viewTabPill) CreateRenderer() fyne.WidgetRenderer {
	p.bg = canvas.NewRectangle(waPillOn)
	p.bg.CornerRadius = viewTabHeight / 2
	p.border = canvas.NewRectangle(color.Transparent)
	p.border.CornerRadius = viewTabHeight / 2
	p.text = canvas.NewText(capitalizeWords(p.label), waPillOnText)
	p.text.TextSize = viewTabFontSize
	p.text.Alignment = fyne.TextAlignCenter
	p.applyStyle()

	inner := container.New(&padLayout{h: viewTabPadX, v: viewTabPadY},
		container.NewCenter(p.text))
	return widget.NewSimpleRenderer(container.NewMax(p.bg, p.border, inner))
}

func (p *viewTabPill) MinSize() fyne.Size {
	w := fyne.MeasureText(capitalizeWords(p.label), viewTabFontSize, fyne.TextStyle{Bold: p.active}).Width
	// border is 1px on each side and sits outside the padding box, like CSS.
	return fyne.NewSize(w+viewTabPadX*2+2, viewTabHeight)
}

func (p *viewTabPill) Tapped(*fyne.PointEvent) {
	if p.onTap != nil {
		p.onTap()
	}
}

func (p *viewTabPill) Cursor() desktop.Cursor { return desktop.PointerCursor }
