package fyneui

import (
	"time"
	"image/color"
	"strings"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	qrcode "github.com/skip2/go-qrcode"
)


var (
	bgPage       = color.NRGBA{R: 240, G: 242, B: 245, A: 255} // bg-light-secondary approx
	cardWhite    = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	borderLight  = color.NRGBA{R: 243, G: 244, B: 246, A: 255} // gray-100
	borderDiv    = color.NRGBA{R: 243, G: 244, B: 246, A: 255}
	titleGray    = color.NRGBA{R: 55, G: 65, B: 81, A: 255}    // gray-700
	textGray     = color.NRGBA{R: 75, G: 85, B: 99, A: 255}    // gray-600
	boldGray     = color.NRGBA{R: 31, G: 41, B: 55, A: 255}    // gray-800
	mutedGray    = color.NRGBA{R: 107, G: 114, B: 128, A: 255} // gray-500
	statusGreen  = color.NRGBA{R: 74, G: 222, B: 128, A: 255}  // green-400 (light-primary-ish)
	qrShadow     = color.NRGBA{R: 0, G: 0, B: 0, A: 18}
)

type LoginScreen struct {
	fa     *FyneApp
	status *canvas.Text
	qrImg  *canvas.Image
}

func NewLoginScreen(fa *FyneApp) *LoginScreen {
	ls := &LoginScreen{
		fa:     fa,
		status: canvas.NewText("", statusGreen),
		qrImg:  canvas.NewImageFromImage(nil),
	}
	ls.status.TextSize = 13
	ls.status.TextStyle = fyne.TextStyle{Bold: true}
	ls.status.Alignment = fyne.TextAlignCenter
	ls.qrImg.FillMode = canvas.ImageFillContain
	ls.qrImg.SetMinSize(fyne.NewSize(264, 264)) // size-66 approx

	go ls.listenEvents()
	go func() {
		if fa.api.IsLoggedIn() {
			// Wait until the websocket is connected before showing the chat UI.
			for !fa.api.IsWaConnected() {
				fyne.Do(func() {
					ls.status.Text = "CONNECTING…"
					ls.status.Refresh()
				})
				time.Sleep(200 * time.Millisecond)
			}
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
	case "timeout":
		return "QR EXPIRED — REFRESHING…"
	case "refreshing":
		return "REFRESHING QR…"
	case "connecting":
		return "CONNECTING…"
	case "qr", "qr_code":
		return ""
	case "logged_in", "success", "connected":
		return "LOGGED IN"
	default:
		if s == "" {
			return ""
		}
		return strings.ToUpper(s)
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
		ls.status.Text = ""
		ls.status.Refresh()
	})
}

func waLogo(size float32) fyne.CanvasObject {
	img := canvas.NewImageFromResource(resourceWALogo)
	img.FillMode = canvas.ImageFillContain
	return container.NewGridWrap(fyne.NewSize(size, size), img)
}

func txt(s string, size float32, col color.NRGBA, bold bool) *canvas.Text {
	t := canvas.NewText(s, col)
	t.TextSize = size
	t.TextStyle = fyne.TextStyle{Bold: bold}
	return t
}


func rich(parts ...fyne.CanvasObject) fyne.CanvasObject {
	c := container.NewHBox()
	for _, p := range parts {
		c.Add(p)
	}
	return c
}

func (ls *LoginScreen) Content() fyne.CanvasObject {
	header := container.NewHBox(
		waLogo(40),
		canvas.NewRectangle(color.Transparent), // gap
		txt("Log in to WhatsApp", 24, titleGray, false), // font-light approx
	)

	// list-decimal style steps, matching React text
	steps := container.NewVBox(
		rich(txt("1. ", 18, textGray, false), txt("Open WhatsApp on your phone", 18, textGray, false)),
		rich(txt("2. ", 18, textGray, false), txt("Tap ", 18, textGray, false), txt("Menu", 18, boldGray, true), txt(" or ", 18, textGray, false), txt("Settings", 18, boldGray, true)),
		rich(txt("3. ", 18, textGray, false), txt("Select ", 18, textGray, false), txt("Linked Devices", 18, boldGray, true)),
		rich(txt("4. ", 18, textGray, false), txt("Point your phone at this screen", 18, textGray, false)),
	)

	// space-y-4 approx
	leftInner := container.NewVBox(
		container.NewPadded(header),
		container.NewGridWrap(fyne.NewSize(0, 24), canvas.NewRectangle(color.Transparent)),
		container.NewPadded(steps),
	)
	left := container.NewPadded(container.NewPadded(leftInner))

	qrWhite := canvas.NewRectangle(color.White)
	qrWhite.CornerRadius = 8
		qrBox := container.NewMax(qrWhite, container.NewPadded(container.NewPadded(ls.qrImg)))

	scanText := txt("Scan this QR code with the WhatsApp app", 14, mutedGray, false)
	scanText.Alignment = fyne.TextAlignCenter

	rightInner := container.NewVBox(
		container.NewCenter(container.NewGridWrap(fyne.NewSize(280, 280), qrBox)),
		container.NewGridWrap(fyne.NewSize(0, 16), canvas.NewRectangle(color.Transparent)),
		container.NewCenter(scanText),
		container.NewGridWrap(fyne.NewSize(0, 12), canvas.NewRectangle(color.Transparent)),
		container.NewCenter(ls.status),
	)
	right := container.NewPadded(container.NewCenter(container.NewPadded(rightInner)))

	divider := canvas.NewRectangle(borderDiv)
	divider.SetMinSize(fyne.NewSize(1, 0))

	// max-w-[900px] h-[520px] approx
	body := container.NewHBox(
		container.NewGridWrap(fyne.NewSize(449, 520), left),
		divider,
		container.NewGridWrap(fyne.NewSize(450, 520), right),
	)

	cardBg := canvas.NewRectangle(cardWhite)
	cardBg.CornerRadius = 8
	cardBg.StrokeColor = borderLight
	cardBg.StrokeWidth = 1
	card := container.NewMax(cardBg, body)

	pageBg := canvas.NewRectangle(bgPage)
	return container.NewMax(pageBg, container.NewCenter(container.NewGridWrap(fyne.NewSize(900, 520), card)))
}
