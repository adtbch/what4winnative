package fyneui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"

	"image/color"
)

// TestAvatarGlyphVisible guards the fallback avatar glyphs. iconPerson is
// authored in a 48-unit viewBox while every other icon uses 24; scaling it by
// size/24 drew the silhouette at double size and clipped it out of the frame,
// leaving a blank circle. The test composites each glyph over the avatar
// background and requires real ink in the icon colour.
func TestAvatarGlyphVisible(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	bg := canvas.NewRectangle(color.NRGBA{R: 0xA8, G: 0xE0, B: 0xB8, A: 255})

	for _, c := range []struct {
		name  string
		glyph fyne.CanvasObject
	}{
		{"person", iconImageVB("g-person", iconPerson, 36, 48, waAvatarIcon)},
		{"people", iconImage("g-people", iconPeople, 26, waAvatarIcon)},
	} {
		t.Run(c.name, func(t *testing.T) {
			img := renderPNG(t, "avatar_glyph_"+c.name,
				container.NewMax(bg, container.NewCenter(c.glyph)), fyne.NewSize(48, 48))

			ink := 0
			b := img.Bounds()
			for y := b.Min.Y; y < b.Max.Y; y++ {
				for x := b.Min.X; x < b.Max.X; x++ {
					r, g, bl, _ := img.At(x, y).RGBA()
					// waAvatarIcon is #111b21 — anything this dark is glyph ink,
					// not the #A8E0B8 background.
					if r>>8 < 0x40 && g>>8 < 0x40 && bl>>8 < 0x40 {
						ink++
					}
				}
			}
			if ink < 40 {
				t.Fatalf("%s glyph drew only %d ink pixels — it is not visible", c.name, ink)
			}
			t.Logf("%s: %d ink px", c.name, ink)
		})
	}
}
