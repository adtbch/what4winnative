package fyneui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ChatInfo panel — T14a parity scaffold (TASKS spec; React ChatInfo.tsx is richer).
// Panel 320px (right), header 64px + close, avatar 96px, name 18px, About section.
const (
	chatInfoWidth     float32 = 320
	chatInfoHeaderH   float32 = 64
	chatInfoAvatarSz  float32 = 96
	chatInfoNameSize  float32 = 18
)

type chatInfoData struct {
	Name     string
	Subtitle string // phone / "Group · N participants"
	About    string
	Avatar   fyne.Resource
	IsGroup  bool
}

type chatInfoPanel struct {
	widget.BaseWidget
	bg      *canvas.Rectangle
	lineV   *canvas.Rectangle
	title   *canvas.Text
	avBg    *canvas.Circle
	avImg   *canvas.Image
	avTxt   *canvas.Text
	name    *canvas.Text
	sub     *canvas.Text
	aboutH  *canvas.Text
	aboutV  *canvas.Text
	onClose func()
}

func newChatInfoPanel(onClose func()) *chatInfoPanel {
	p := &chatInfoPanel{onClose: onClose}
	p.ExtendBaseWidget(p)
	p.bg = canvas.NewRectangle(waPanel) // dark secondary
	p.lineV = canvas.NewRectangle(color.NRGBA{R:255,G:255,B:255,A:13})
	p.title = canvas.NewText("Contact Info", waText)
	p.title.TextSize = 16
	p.title.TextStyle = fyne.TextStyle{Bold: true}
	p.avBg = canvas.NewCircle(chatHeaderAvatarBg)
	p.avImg = &canvas.Image{FillMode: canvas.ImageFillContain}
	p.avImg.SetMinSize(fyne.NewSize(chatInfoAvatarSz, chatInfoAvatarSz))
	p.avTxt = canvas.NewText("?", waWhite)
	p.avTxt.Alignment = fyne.TextAlignCenter
	p.avTxt.TextSize = 34
	p.name = canvas.NewText("", chatHeaderNameFg)
	p.name.TextSize = chatInfoNameSize
	p.name.TextStyle = fyne.TextStyle{Bold: true}
	p.name.Alignment = fyne.TextAlignCenter
	p.sub = canvas.NewText("", waMuted)
	p.sub.TextSize = 12
	p.sub.Alignment = fyne.TextAlignCenter
	p.aboutH = canvas.NewText("About", waMuted)
	p.aboutH.TextSize = 12
	p.aboutV = canvas.NewText("No about info", waText)
	p.aboutV.TextSize = 14
	return p
}

func (p *chatInfoPanel) Set(d chatInfoData) {
	title := "Contact Info"
	if d.IsGroup {
		title = "Group Info"
	}
	p.title.Text = title
	p.name.Text = stringsTrim(d.Name)
	p.sub.Text = stringsTrim(d.Subtitle)
	if stringsTrim(d.About) != "" {
		p.aboutV.Text = d.About
	} else {
		p.aboutV.Text = "No about info"
	}
	if d.Avatar != nil {
		p.avImg.Resource = d.Avatar
		p.avImg.Show()
		p.avTxt.Hide()
	} else {
		p.avImg.Hide()
		ini := "?"
		if r := []rune(stringsTrim(d.Name)); len(r) > 0 {
			ini = string(r[0])
		}
		p.avTxt.Text = ini
		p.avTxt.Show()
	}
	p.Refresh()
}

func stringsTrim(s string) string {
	// local to avoid import churn in small file; strings imported if needed
	i, j := 0, len(s)
	for i < j && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n' || s[i] == '\r') { i++ }
	for j > i && (s[j-1] == ' ' || s[j-1] == '\t' || s[j-1] == '\n' || s[j-1] == '\r') { j-- }
	return s[i:j]
}

func (p *chatInfoPanel) CreateRenderer() fyne.WidgetRenderer {
	closeBtn := newIconHit("info-close", iconBack, 20, 40, waMuted, p.onClose) // back/close affordance
	header := container.New(&padLayout{h: 12, v: 0},
		container.NewHBox(container.NewCenter(closeBtn), container.NewPadded(p.title)))
	headerH := container.NewMax(container.NewGridWrap(fyne.NewSize(chatInfoWidth, chatInfoHeaderH), container.NewMax(canvas.NewRectangle(color.Transparent), header)))
	hLine := canvas.NewRectangle(color.NRGBA{R:255,G:255,B:255,A:13})
	hLine.SetMinSize(fyne.NewSize(chatInfoWidth, 1))

	av := container.NewGridWrap(fyne.NewSize(chatInfoAvatarSz, chatInfoAvatarSz), container.NewMax(p.avBg, container.NewCenter(p.avImg), container.NewCenter(p.avTxt)))
	profile := container.NewVBox(container.NewCenter(container.NewPadded(av)), container.NewCenter(p.name), container.NewCenter(p.sub))
	about := container.New(&padLayout{h: 16, v: 8}, container.NewVBox(p.aboutH, p.aboutV))
	body := container.NewVScroll(container.NewVBox(container.NewPadded(profile), hLine, about))

	content := container.NewBorder(headerH, nil, nil, nil, body)
	return widget.NewSimpleRenderer(container.NewMax(p.bg, container.NewBorder(nil, nil, p.lineV, nil, content)))
}

func (p *chatInfoPanel) MinSize() fyne.Size { return fyne.NewSize(chatInfoWidth, 200) }
