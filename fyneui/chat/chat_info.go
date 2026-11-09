package chat

import (
	"fmt"
	"image/color"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/adtbch/whats4linux/fyneui/theme"
	"github.com/adtbch/whats4linux/fyneui/utils"
)

const (
	chatInfoWidth    float32 = 320
	chatInfoHeaderH  float32 = 64
	chatInfoAvatarSz float32 = 96
	chatInfoNameSize float32 = 18
)

type chatInfoParticipant struct {
	Name    string
	Phone   string
	IsAdmin bool
	Avatar  fyne.Resource
}

type chatInfoData struct {
	Name         string
	Subtitle     string
	About        string
	Avatar       fyne.Resource
	IsGroup      bool
	Muted        bool
	MemberCount  int
	Participants []chatInfoParticipant
}

type chatInfoPanel struct {
	widget.BaseWidget
	bg             *canvas.Rectangle
	lineV          *canvas.Rectangle
	title          *canvas.Text
	avBg           *canvas.Rectangle
	avImg          *canvas.Image
	avTxt          *canvas.Text
	name           *canvas.Text
	sub            *canvas.Text
	aboutH, aboutV *canvas.Text
	muteLabel      *canvas.Text
	partsBox       *fyne.Container
	onClose        func()
	onMute         func(bool)
}

func newChatInfoPanel(onClose func(), onMute func(bool)) *chatInfoPanel {
	p := &chatInfoPanel{onClose: onClose, onMute: onMute}
	p.ExtendBaseWidget(p)
	p.bg = canvas.NewRectangle(theme.WaPanel)
	p.lineV = canvas.NewRectangle(color.NRGBA{R: 255, G: 255, B: 255, A: 13})
	p.title = canvas.NewText("Contact Info", theme.WaText)
	p.title.TextSize = 16
	p.title.TextStyle = fyne.TextStyle{Bold: true}
	p.avBg = func() *canvas.Rectangle { r := canvas.NewRectangle(chatHeaderAvatarBg); r.CornerRadius = 6; return r }()
	p.avImg = &canvas.Image{FillMode: canvas.ImageFillContain}
	p.avImg.SetMinSize(fyne.NewSize(chatInfoAvatarSz, chatInfoAvatarSz))
	p.avTxt = canvas.NewText("?", theme.WaWhite)
	p.avTxt.Alignment = fyne.TextAlignCenter
	p.avTxt.TextSize = 34
	p.name = canvas.NewText("", chatHeaderNameFg)
	p.name.TextSize = chatInfoNameSize
	p.name.TextStyle = fyne.TextStyle{Bold: true}
	p.name.Alignment = fyne.TextAlignCenter
	p.sub = canvas.NewText("", theme.WaMuted)
	p.sub.TextSize = 12
	p.sub.Alignment = fyne.TextAlignCenter
	p.aboutH = canvas.NewText("About", theme.WaMuted)
	p.aboutH.TextSize = 12
	p.aboutV = canvas.NewText("No about info", theme.WaText)
	p.aboutV.TextSize = 14
	p.muteLabel = canvas.NewText("Mute notifications", theme.WaText)
	p.muteLabel.TextSize = 14
	p.partsBox = container.NewVBox()
	return p
}

func (p *chatInfoPanel) Set(d chatInfoData) {
	p.title.Text = map[bool]string{true: "Group Info", false: "Contact Info"}[d.IsGroup]
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
	p.muteLabel.Text = map[bool]string{true: "Unmute notifications", false: "Mute notifications"}[d.Muted]
	p.muteLabel.Color = theme.WaText
	// participants
	p.partsBox.Objects = nil
	if d.IsGroup {
		h := canvas.NewText(fmt.Sprintf("%d members", maxI(d.MemberCount, len(d.Participants))), theme.WaMuted)
		h.TextSize = 12
		p.partsBox.Add(container.New(&utils.PadLayout{H: 16, V: 6}, h))
		lim := len(d.Participants)
		if lim > 12 {
			lim = 12
		}
		for i := 0; i < lim; i++ {
			pt := d.Participants[i]
			lbl := canvas.NewText(pt.Name, theme.WaText)
			lbl.TextSize = 14
			line := container.New(&utils.PadLayout{H: 16, V: 4}, container.NewHBox(lbl))
			if pt.IsAdmin {
				tag := canvas.NewText("admin", theme.WaAccent)
				tag.TextSize = 11
				line = container.New(&utils.PadLayout{H: 16, V: 4}, container.NewBorder(nil, nil, lbl, tag, nil))
			}
			p.partsBox.Add(line)
		}
		if len(d.Participants) > lim {
			more := canvas.NewText(fmt.Sprintf("+%d more", len(d.Participants)-lim), theme.WaMuted)
			more.TextSize = 12
			p.partsBox.Add(container.New(&utils.PadLayout{H: 16, V: 4}, more))
		}
	}
	p.partsBox.Refresh()
	p.Refresh()
}

func maxI(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func stringsTrim(s string) string {
	i, j := 0, len(s)
	for i < j && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n' || s[i] == '\r') {
		i++
	}
	for j > i && (s[j-1] == ' ' || s[j-1] == '\t' || s[j-1] == '\n' || s[j-1] == '\r') {
		j--
	}
	return s[i:j]
}

func actionRow(icon, label string, fg color.NRGBA, onTap func()) fyne.CanvasObject {
	ic := utils.IconImage("act-"+label, icon, 20, fg)
	lbl := canvas.NewText(label, fg)
	lbl.TextSize = 14
	line := container.New(&utils.PadLayout{H: 16, V: 10}, container.NewHBox(ic, container.NewPadded(lbl)))
	return container.NewMax(NewTapArea(onTap), line)
}

func (p *chatInfoPanel) CreateRenderer() fyne.WidgetRenderer {
	closeBtn := NewIconHit("info-close", utils.IconBack, 20, 40, theme.WaMuted, p.onClose)
	header := container.New(&utils.PadLayout{H: 12, V: 0},
		container.NewHBox(container.NewCenter(closeBtn), container.NewPadded(p.title)))
	headerH := container.NewMax(container.NewGridWrap(fyne.NewSize(chatInfoWidth, chatInfoHeaderH), container.NewMax(canvas.NewRectangle(color.Transparent), header)))
	hLine := canvas.NewRectangle(color.NRGBA{R: 255, G: 255, B: 255, A: 13})
	hLine.SetMinSize(fyne.NewSize(chatInfoWidth, 1))

	av := container.NewGridWrap(fyne.NewSize(chatInfoAvatarSz, chatInfoAvatarSz), container.NewMax(p.avBg, container.NewCenter(p.avImg), container.NewCenter(p.avTxt)))
	profile := container.NewVBox(container.NewCenter(container.NewPadded(av)), container.NewCenter(p.name), container.NewCenter(p.sub))
	about := container.New(&utils.PadLayout{H: 16, V: 8}, container.NewVBox(p.aboutH, p.aboutV))

	searchRow := actionRow(utils.IconSearch, "Search", theme.WaText, func() { log.Println("chatinfo: search stub") })
	muteTap := &tapArea{onTap: func() {
		want := p.muteLabel.Text == "Mute notifications"
		if p.onMute != nil {
			p.onMute(want)
		}
	}}
	muteTap.ExtendBaseWidget(muteTap)
	muteIc := utils.IconImage("act-mute", utils.IconMuted, 20, theme.WaText)
	muteLine := container.New(&utils.PadLayout{H: 16, V: 10}, container.NewHBox(muteIc, container.NewPadded(p.muteLabel)))
	muteRow := container.NewMax(muteTap, muteLine)

	actions := container.NewVBox(hLine, searchRow, muteRow, hLine)
	body := container.NewVScroll(container.NewVBox(container.NewPadded(profile), about, actions, p.partsBox))

	content := container.NewBorder(headerH, nil, nil, nil, body)
	return widget.NewSimpleRenderer(container.NewMax(p.bg, container.NewBorder(nil, nil, p.lineV, nil, content)))
}

func (p *chatInfoPanel) MinSize() fyne.Size { return fyne.NewSize(chatInfoWidth, 200) }
