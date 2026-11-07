package fyneui

import (
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	qrcode "github.com/skip2/go-qrcode"
)

var waGreen = color.NRGBA{R: 37, G: 211, B: 102, A: 255}
var bgDark = color.NRGBA{R: 17, G: 27, B: 33, A: 255}
var cardDark = color.NRGBA{R: 31, G: 44, B: 52, A: 255}
var textLight = color.NRGBA{R: 233, G: 237, B: 239, A: 255}
var textMuted = color.NRGBA{R: 134, G: 150, B: 160, A: 255}

type LoginScreen struct {
	fa     *FyneApp
	status *canvas.Text
	qrImg  *canvas.Image
	qrCard *fyne.Container
}

func NewLoginScreen(fa *FyneApp) *LoginScreen {
	ls := &LoginScreen{
		fa:     fa,
		status: canvas.NewText("Menghubungkan...", textMuted),
		qrImg:  canvas.NewImageFromImage(nil),
	}
	ls.status.TextSize = 14
	ls.status.Alignment = fyne.TextAlignCenter
	ls.qrImg.FillMode = canvas.ImageFillContain
	ls.qrImg.SetMinSize(fyne.NewSize(260, 260))

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
				fyne.Do(func() {
					ls.status.Text = statusLabel(s)
					ls.status.Refresh()
				})
				if s == "logged_in" || s == "success" || s == "connected" {
					fyne.Do(func() { ls.fa.SwitchToChats() })
				}
			}
		}
	}
}

func statusLabel(s string) string {
	switch s {
	case "connecting":
		return "Menghubungkan..."
	case "qr", "qr_code":
		return "Scan QR dengan WhatsApp di HP"
	case "logged_in", "success", "connected":
		return "Berhasil masuk"
	default:
		if s == "" {
			return "Menunggu QR..."
		}
		return s
	}
}

func (ls *LoginScreen) setQR(code string) {
	png, err := qrcode.Encode(code, qrcode.Medium, 280)
	if err != nil {
		return
	}
	res := fyne.NewStaticResource("qr.png", png)
	fyne.Do(func() {
		ls.qrImg.Resource = res
		ls.qrImg.Refresh()
		ls.status.Text = "Scan QR dengan WhatsApp di HP"
		ls.status.Refresh()
	})
}

func step(n, text string) fyne.CanvasObject {
	num := canvas.NewText(n+".", waGreen)
	num.TextStyle = fyne.TextStyle{Bold: true}
	num.TextSize = 15
	t := canvas.NewText(text, textLight)
	t.TextSize = 15
	return container.NewHBox(num, t)
}

func (ls *LoginScreen) Content() fyne.CanvasObject {
	title := canvas.NewText("what4winnative", textLight)
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.TextSize = 28
	title.Alignment = fyne.TextAlignCenter

	subtitle := canvas.NewText("WhatsApp desktop client (Fyne)", textMuted)
	subtitle.TextSize = 14
	subtitle.Alignment = fyne.TextAlignCenter

	instrTitle := canvas.NewText("Cara login", textLight)
	instrTitle.TextStyle = fyne.TextStyle{Bold: true}
	instrTitle.TextSize = 18

	instructions := container.NewVBox(
		instrTitle,
		layout.NewSpacer(),
		step("1", "Buka WhatsApp di HP"),
		step("2", "Ketuk Menu / Setelan"),
		step("3", "Pilih Perangkat tertaut"),
		step("4", "Arahkan kamera ke QR ini"),
	)

	qrBg := canvas.NewRectangle(color.White)
	qrBg.CornerRadius = 12
	qrWrap := container.NewMax(qrBg, container.NewPadded(ls.qrImg))

	right := container.NewVBox(
		layout.NewSpacer(),
		container.NewCenter(qrWrap),
		container.NewPadded(ls.status),
		layout.NewSpacer(),
	)

	leftPad := container.NewPadded(container.NewVBox(
		title, subtitle,
		widget.NewSeparator(),
		instructions,
	))

	cardBg := canvas.NewRectangle(cardDark)
	cardBg.CornerRadius = 16
	split := container.NewGridWithColumns(2, container.NewPadded(leftPad), container.NewPadded(right))
	card := container.NewMax(cardBg, container.NewPadded(split))

	rootBg := canvas.NewRectangle(bgDark)
	return container.NewMax(rootBg, container.NewCenter(container.NewGridWrap(fyne.NewSize(920, 520), card)))
}
