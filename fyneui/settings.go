package fyneui

import (
	"image/color"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type settingsRow struct {
	Icon   string
	Label  string
	Danger bool
}

func defaultSettingsRows() []settingsRow {
	return []settingsRow{
		{Icon: iconPerson, Label: "Account"},
		{Icon: iconMuted, Label: "Privacy"},
		{Icon: iconChats, Label: "Chats"},
		{Icon: iconStatus, Label: "Notifications"},
		{Icon: iconSettings, Label: "Advanced"},
		{Icon: iconBack, Label: "Log out", Danger: true},
	}
}

type settingsScreen struct {
	widget.BaseWidget
	onBack func()
	rows   []settingsRow
}

func newSettingsScreen(onBack func()) *settingsScreen {
	s := &settingsScreen{onBack: onBack, rows: defaultSettingsRows()}
	s.ExtendBaseWidget(s)
	return s
}

func (s *settingsScreen) CreateRenderer() fyne.WidgetRenderer {
	bg := canvas.NewRectangle(waBg)
	back := newIconHit("set-back", iconBack, 20, 40, waMuted, s.onBack)
	title := canvas.NewText("Settings", waText)
	title.TextSize = 20
	title.TextStyle = fyne.TextStyle{Bold: true}
	header := container.New(&padLayout{h: 12, v: 10}, container.NewHBox(back, container.NewPadded(title)))
	hLine := canvas.NewRectangle(color.NRGBA{R: 255, G: 255, B: 255, A: 13})
	hLine.SetMinSize(fyne.NewSize(0, 1))

	items := make([]fyne.CanvasObject, 0, len(s.rows)*2)
	for _, r := range s.rows {
		row := r
		fg := waText
		if row.Danger {
			fg = color.NRGBA{R: 239, G: 68, B: 68, A: 255}
		}
		ic := iconImage("set-"+row.Label, row.Icon, 20, fg)
		lbl := canvas.NewText(row.Label, fg)
		lbl.TextSize = 15
		chev := canvas.NewText("›", waMuted)
		chev.TextSize = 18
		tap := &tapArea{onTap: func() { log.Println("settings stub:", row.Label) }}
		line := container.NewBorder(nil, nil, container.NewPadded(ic), container.NewPadded(chev),
			container.New(&padLayout{h: 4, v: 10}, lbl))
		items = append(items, container.NewMax(tap, line))
		sep := canvas.NewRectangle(color.NRGBA{R: 255, G: 255, B: 255, A: 8})
		sep.SetMinSize(fyne.NewSize(0, 1))
		items = append(items, sep)
	}
	list := container.NewVScroll(container.NewVBox(items...))
	content := container.NewBorder(container.NewVBox(header, hLine), nil, nil, nil, container.NewPadded(list))
	return widget.NewSimpleRenderer(container.NewMax(bg, content))
}

func (s *settingsScreen) MinSize() fyne.Size { return fyne.NewSize(420, 400) }
