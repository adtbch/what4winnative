package fyneui

import (
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	qrcode "github.com/skip2/go-qrcode"
)

var (
	waGreen    = color.NRGBA{R: 37, G: 211, B: 102, A: 255}
	bgPage     = color.NRGBA{R: 240, G: 242, B: 245, A: 255}
	cardWhite  = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	borderGray = color.NRGBA{R: 229, G: 231, B: 235, A: 255}
	textDark   = color.NRGBA{R: 55, G: 65, B: 81, A: 255}
	textBody   = color.NRGBA{R: 75, G: 85, B: 99, A: 255}
	textMuted  = color.NRGBA{R: 107, G: 114, B: 128, A: 255}
)

type LoginScreen struct {
	fa     *FyneApp
	status *canvas.Text
	qrImg  *canvas.Image
}

func NewLoginScreen(fa *FyneApp) *LoginScreen {
	ls := &LoginScreen{
		fa:     fa,
		status: canvas.NewText("Waiting for QR code...", textMuted),
		qrImg:  canvas.NewImageFromImage(nil),
	}
	ls.status.TextSize = 13
	ls.status.Alignment = fyne.TextAlignCenter
	ls.qrImg.FillMode = canvas.ImageFillContain
	ls.qrImg.SetMinSize(fyne.NewSize(264, 264))

	go ls.listenEvents()
	go func() {
		if fa.api.IsLoggedIn() {
			fyne.Do(func() { fa.SwitchToChats() })
			return
		}
		_ = fa.api.Login()
	}()
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
			if s, ok := data.(string); ok && strings.TrimSpace(s) != "" {
				ls.setQR(s)
			}
		case data := <-statusCh:
			if s, ok := data.(string); ok {
				fyne.Do(func() { ls.status.Text = statusLabel(s); ls.status.Refresh() })
				if s == "logged_in" || s == "success" || s == "connected" {
					fyne.Do(func() { ls.fa.SwitchToChats() })
				}
			}
		}
	}
}

func statusLabel(s string) string {
	switch s {
	case "connecting", "refreshing":
		return "Connecting..."
	case "qr", "qr_code":
		return "Scan this code with WhatsApp"
	case "logged_in", "success", "connected":
		return "Logged in"
	default:
		if s == "" {
			return "Waiting for QR code..."
		}
		return s
	}
}

func (ls *LoginScreen) setQR(code string) {
	png, err := qrcode.Encode(code, qrcode.Medium, 264)
	if err != nil {
		return
	}
	res := fyne.NewStaticResource("qr.png", png)
	fyne.Do(func() {
		ls.qrImg.Resource = res
		ls.qrImg.Refresh()
		ls.status.Text = "Scan this code with WhatsApp"
		ls.status.Refresh()
	})
}

func waLogo(size float32) fyne.CanvasObject {
	img := canvas.NewImageFromResource(resourceWALogo)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(size, size))
	return container.NewCenter(container.NewGridWrap(fyne.NewSize(size, size), img))
}

func instructionLine(prefix, bold, suffix string) fyne.CanvasObject {
	a := canvas.NewText(prefix, textBody)
	a.TextSize = 16
	b := canvas.NewText(bold, textDark)
	b.TextSize = 16
	b.TextStyle = fyne.TextStyle{Bold: true}
	c := canvas.NewText(suffix, textBody)
	c.TextSize = 16
	return container.NewHBox(a, b, c)
}

func (ls *LoginScreen) Content() fyne.CanvasObject {
	title := canvas.NewText("Log in to WhatsApp", textDark)
	title.TextSize = 24

	header := container.NewHBox(container.NewCenter(container.NewGridWrap(fyne.NewSize(40, 40), waLogo(40))), widget.NewLabel("  "), container.NewCenter(title))

	steps := container.NewVBox(
		canvas.NewText("1. Open WhatsApp on your phone", textBody),
		instructionLine("2. Tap ", "Menu", " or "),
		func() fyne.CanvasObject {
			t := instructionLine("   ", "Settings", "")
			return t
		}(),
		instructionLine("3. Select ", "Linked Devices", ""),
		canvas.NewText("4. Point your phone at this screen", textBody),
	)
	for _, o := range steps.Objects {
		if t, ok := o.(*canvas.Text); ok {
			t.TextSize = 16
		}
	}

	left := container.NewVBox(
		container.NewPadded(header),
		widget.NewLabel(""),
		container.NewPadded(steps),
	)

	qrBg := canvas.NewRectangle(color.White)
	qrBg.StrokeColor = borderGray
	qrBg.StrokeWidth = 1
	qrBg.CornerRadius = 8
	qrInner := container.NewPadded(ls.qrImg)
	qrCard := container.NewMax(qrBg, qrInner)

	right := container.NewVBox(
		container.NewCenter(container.NewGridWrap(fyne.NewSize(280, 280), qrCard)),
		widget.NewLabel(""),
		container.NewCenter(ls.status),
	)

	divider := canvas.NewRectangle(borderGray)
	divider.SetMinSize(fyne.NewSize(1, 0))

	body := container.NewBorder(nil, nil, nil, nil,
		container.NewHBox(
			container.NewGridWrap(fyne.NewSize(430, 420), container.NewPadded(left)),
			divider,
			container.NewGridWrap(fyne.NewSize(430, 420), container.NewPadded(container.NewCenter(right))),
		),
	)

	cardBg := canvas.NewRectangle(cardWhite)
	cardBg.StrokeColor = borderGray
	cardBg.StrokeWidth = 1
	cardBg.CornerRadius = 10
	card := container.NewMax(cardBg, container.NewPadded(body))

	pageBg := canvas.NewRectangle(bgPage)
	return container.NewMax(pageBg, container.NewCenter(container.NewGridWrap(fyne.NewSize(900, 480), card)))
}
