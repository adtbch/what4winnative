package fyneui

import (
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const (
	longTextChars = 280
	longTextLines = 6
	clampLinkCol  = "#53bdeb"
)

func clampLinkColor() color.NRGBA { return color.NRGBA{R: 0x53, G: 0xbd, B: 0xeb, A: 255} }

// isLongTextPlain mirrors MessageItem.tsx isLongText (stripped, non-empty).
func isLongTextPlain(plain string) bool {
	p := strings.TrimSpace(plain)
	if p == "" {
		return false
	}
	r := []rune(p)
	if len(r) > longTextChars {
		return true
	}
	return len(strings.Split(p, "\n")) >= longTextLines
}

func truncatePlain(plain string, maxChars, maxLines int) string {
	p := strings.TrimSpace(plain)
	lines := strings.Split(p, "\n")
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		p = strings.Join(lines, "\n")
	}
	r := []rune(p)
	if len(r) > maxChars {
		p = string(r[:maxChars])
	}
	return strings.TrimRight(p, " \n") + "…"
}

type clampTextWidget struct {
	widget.BaseWidget
	full      fyne.CanvasObject
	collapsed fyne.CanvasObject
	expanded  bool
	box       *fyne.Container
}

func newClampText(full, collapsed fyne.CanvasObject) *clampTextWidget {
	w := &clampTextWidget{full: full, collapsed: collapsed}
	w.ExtendBaseWidget(w)
	link := canvas.NewText("Read more", clampLinkColor())
	link.TextSize = 12
	link.TextStyle = fyne.TextStyle{Bold: true}
	tap := &tapArea{onTap: func() {
		w.expanded = !w.expanded
		if w.expanded {
			link.Text = "Read less"
			w.box.Objects = []fyne.CanvasObject{container.NewVBox(w.full, container.NewHBox(link))}
		} else {
			link.Text = "Read more"
			w.box.Objects = []fyne.CanvasObject{container.NewVBox(w.collapsed, container.NewHBox(link))}
		}
		link.Refresh()
		w.box.Refresh()
	}}
	linkWrap := container.NewMax(tap, container.NewHBox(link))
	w.box = container.NewMax(container.NewVBox(w.collapsed, linkWrap))
	return w
}

func (w *clampTextWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.box)
}
