package login

import (
	"context"
	"image/color"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	qrcode "github.com/skip2/go-qrcode"

	"github.com/adtbch/whats4linux/api"
	"github.com/adtbch/whats4linux/fyneui/theme"
	"github.com/adtbch/whats4linux/fyneui/utils"
)

var (
	statusGreen = color.NRGBA{R: 0x25, G: 0xD3, B: 0x66, A: 255}
	titleGray   = color.NRGBA{R: 0xea, G: 0xea, B: 0xea, A: 255}
	textGray    = color.NRGBA{R: 0xaa, G: 0xaa, B: 0xaa, A: 255}
	boldGray    = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 255}
	mutedGray   = color.NRGBA{R: 0x88, G: 0x88, B: 0x88, A: 255}
	borderDiv   = color.NRGBA{R: 0x2a, G: 0x2a, B: 0x2a, A: 255}
	cardWhite   = color.NRGBA{R: 0x1a, G: 0x1a, B: 0x1a, A: 255}
	borderLight = color.NRGBA{R: 0x3a, G: 0x3a, B: 0x3a, A: 255}
	bgPage      = color.NRGBA{R: 0x0b, G: 0x14, B: 0x1f, A: 255}
)

// App is the host surface LoginScreen needs: the WhatsApp API, a
// cancellable context and navigation into chats. Implemented by
// fyneui.FyneApp so login never imports the shell package.
type App interface {
	API() *api.Api
	Ctx() context.Context
	SwitchToChats()
}

type LoginScreen struct {
	fa     App
	status *canvas.Text
	qrImg  *canvas.Image
}

func NewLoginScreen(host App) *LoginScreen {
	ls := &LoginScreen{
		fa:     host,
		status: canvas.NewText("", statusGreen),
		qrImg:  canvas.NewImageFromImage(nil),
	}
	ls.status.TextSize = 13
	ls.status.TextStyle = fyne.TextStyle{Bold: true}
	ls.status.Alignment = fyne.TextAlignCenter
	ls.qrImg.FillMode = canvas.ImageFillContain
	ls.qrImg.SetMinSize(fyne.NewSize(264, 264))

	go ls.listenEvents()
	go func() {
		if ls.fa.API().IsLoggedIn() {
			for !ls.fa.API().IsWaConnected() {
				fyne.Do(func() {
					ls.status.Text = "CONNECTING…"
					ls.status.Refresh()
				})
				time.Sleep(200 * time.Millisecond)
			}
			fyne.Do(func() { ls.fa.SwitchToChats() })
			return
		}
		_ = ls.fa.API().Login()
	}()
	return ls
}

func (ls *LoginScreen) listenEvents() {
	qrCh := utils.SubscribeEvent("wa:qr")
	statusCh := utils.SubscribeEvent("wa:status")
	for {
		select {
		case <-ls.fa.Ctx().Done():
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
	img := canvas.NewImageFromResource(theme.ResourceWALogo)
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
		canvas.NewRectangle(color.Transparent),
		txt("Log in to WhatsApp", 24, titleGray, false),
	)

	steps := container.NewVBox(
		rich(txt("1. ", 18, textGray, false), txt("Open WhatsApp on your phone", 18, textGray, false)),
		rich(txt("2. ", 18, textGray, false), txt("Tap ", 18, textGray, false), txt("Menu", 18, boldGray, true), txt(" or ", 18, textGray, false), txt("Settings", 18, boldGray, true)),
		rich(txt("3. ", 18, textGray, false), txt("Select ", 18, textGray, false), txt("Linked Devices", 18, boldGray, true)),
		rich(txt("4. ", 18, textGray, false), txt("Point your phone at this screen", 18, textGray, false)),
	)

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
