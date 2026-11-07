package fyneui

import (
	"encoding/base64"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type LoginScreen struct {
	fa     *FyneApp
	status *widget.Label
	qrImg  *canvas.Image
}

func NewLoginScreen(fa *FyneApp) *LoginScreen {
	ls := &LoginScreen{
		fa:     fa,
		status: widget.NewLabel("Menunggu QR..."),
		qrImg:  canvas.NewImageFromImage(nil),
	}
	ls.qrImg.FillMode = canvas.ImageFillContain
	ls.qrImg.SetMinSize(fyne.NewSize(280, 280))

	go ls.listenEvents()
	go func() { _ = fa.api.Login() }()
	return ls
}

func (ls *LoginScreen) listenEvents() {
	qrCh := SubscribeEvent("wa:qr")
	statusCh := SubscribeEvent("wa:status")
	for {
		select {
		case <-ls.fa.ctx.Done():
			return
		case data := <-qrCh:
			if s, ok := data.(string); ok {
				ls.setQR(s)
			}
		case data := <-statusCh:
			if s, ok := data.(string); ok {
				ls.status.SetText(s)
				if s == "logged_in" || s == "success" || s == "connected" {
					fyne.Do(func() { ls.fa.SwitchToChats() })
				}
			}
		}
	}
}

func (ls *LoginScreen) setQR(payload string) {
	// payload may be data:image/png;base64,... or raw base64
	b64 := payload
	if i := strings.Index(payload, ","); i >= 0 && strings.Contains(payload[:i], "base64") {
		b64 = payload[i+1:]
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(b64))
	if err != nil {
		return
	}
	res := fyne.NewStaticResource("qr.png", raw)
	fyne.Do(func() {
		ls.qrImg.Resource = res
		ls.qrImg.Refresh()
	})
}

func (ls *LoginScreen) Content() fyne.CanvasObject {
	return container.NewCenter(container.NewVBox(
		widget.NewLabelWithStyle("what4winnative", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		ls.qrImg,
		ls.status,
	))
}
