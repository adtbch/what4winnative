package fyneui

import (
	"image/color"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

func TestComposerSendClears(t *testing.T) {
	var got []string
	obj, entry := newComposer(func(s string) { got = append(got, s) }, func() {}, func() {})
	_ = obj
	entry.SetText("halo")
	entry.OnSubmitted("halo")
	if len(got) != 1 || got[0] != "halo" {
		t.Fatalf("send got=%v want [halo]", got)
	}
	if strings.TrimSpace(entry.Text) != "" {
		t.Fatalf("entry not cleared: %q", entry.Text)
	}
	// empty/whitespace no-op
	entry.OnSubmitted("")
	if len(got) != 1 {
		t.Fatalf("empty should not send: got=%v", got)
	}
}

func TestComposerGolden(t *testing.T) {
	obj, _ := newComposer(func(string) {}, func() {}, func() {})
	img := renderPNG(t, "composer", obj, fyne.NewSize(520, 90))
	// pill bg waPanel2 #242626 dominant
	_, _, _, _, n := dominantCardBBox(img, color.NRGBA{R: 0x24, G: 0x26, B: 0x26, A: 255}, 25)
	if n < 5000 {
		t.Fatalf("composer pill waPanel2 weak n=%d", n)
	}
	// send circle waAccent #1b9a58 present
	if !hasColorNear(img, color.NRGBA{R: 0x1b, G: 0x9a, B: 0x58, A: 255}, 25, 200) {
		t.Fatalf("composer send accent #1b9a58 missing")
	}
}

func TestComposerPlaceholder(t *testing.T) {
	_, entry := newComposer(func(string) {}, nil, nil)
	var _ *widget.Entry = entry
	if entry.PlaceHolder != "Message" {
		t.Fatalf("placeholder=%q want Message", entry.PlaceHolder)
	}
}
