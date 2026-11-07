package fyneui

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
)

func TestSettingsScreenGolden(t *testing.T) {
	s := newSettingsScreen(func() {})
	img := renderPNG(t, "settings", s, fyne.NewSize(420, 520))
	assertNotBlank(t, img, 0.003)
	if !hasColorNear(img, color.NRGBA{R: 239, G: 68, B: 68, A: 255}, 15, 50) { // Log out row red
		t.Fatal("settings missing danger/logout accent")
	}
}

func TestSettingsScreenRows(t *testing.T) {
	rows := defaultSettingsRows()
	if len(rows) < 5 {
		t.Fatalf("rows=%d", len(rows))
	}
	if !rows[len(rows)-1].Danger {
		t.Fatal("last row should be danger/logout")
	}
}
