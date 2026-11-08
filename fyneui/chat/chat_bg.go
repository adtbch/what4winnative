package chat

import (
	"image"
	"image/color"
	"image/draw"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"github.com/lugvitc/whats4linux/fyneui/theme"
)

// chatBackground returns WA-like dark chat bg with subtle doodle pattern.
func ChatBackground(_ fyne.Size) fyne.CanvasObject {
	base := canvas.NewRectangle(theme.WaBg)
	pat := canvas.NewImageFromImage(makeDoodleTile(240, 240))
	pat.FillMode = canvas.ImageFillStretch
	pat.Translucency = 0
	return container.NewMax(base, pat)
}

func makeDoodleTile(w, h int) image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.NRGBA{0, 0, 0, 0}}, image.Point{}, draw.Src)
	dot := color.NRGBA{R: 255, G: 255, B: 255, A: 9}
	line := color.NRGBA{R: 255, G: 255, B: 255, A: 7}
	for y := 18; y < h; y += 44 {
		for x := 14; x < w; x += 52 {
			setPx(img, x, y, dot)
			setPx(img, x+1, y, dot)
			setPx(img, x, y+1, dot)
			for i := 0; i < 5; i++ {
				setPx(img, x+20+i, y+8, line)
				setPx(img, x+22, y+6+i, line)
			}
		}
	}
	for y := 40; y < h; y += 66 {
		for x := 34; x < w; x += 74 {
			for i := 0; i < 6; i++ {
				setPx(img, x+i, y, line)
				if y+1 < h {
					setPx(img, x+5-i, y+1, line)
				}
			}
		}
	}
	return img
}

func setPx(img *image.NRGBA, x, y int, c color.NRGBA) {
	if x < 0 || y < 0 || x >= img.Bounds().Dx() || y >= img.Bounds().Dy() {
		return
	}
	i := img.PixOffset(x, y)
	sa := float64(c.A) / 255.0
	da := float64(img.Pix[i+3]) / 255.0
	oa := sa + da*(1-sa)
	if oa <= 0 {
		return
	}
	for ch, sv := range []uint8{c.R, c.G, c.B} {
		dv := img.Pix[i+ch]
		img.Pix[i+ch] = uint8((float64(sv)*sa + float64(dv)*da*(1-sa)) / oa)
	}
	img.Pix[i+3] = uint8(oa * 255)
}
