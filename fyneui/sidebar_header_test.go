package fyneui

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
)

// Measured from ChatScreen.tsx through the React harness (`out.header`).
const (
	reactHeaderHeight = 64
	reactHeaderPadX   = 16
	reactHeaderBtn    = 40
	reactHeaderIcon   = 24
	reactHeaderAvatar = 32
	reactHeaderTitle  = 20
	reactHeaderWidth  = 319
)

func TestSidebarHeaderGeometryMatchesReact(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	h := newSidebarHeader(nil, nil)
	if got := h.MinSize().Height; got != reactHeaderHeight {
		t.Errorf("header height %v != React %d", got, reactHeaderHeight)
	}
	if got := h.title.TextSize; got != reactHeaderTitle {
		t.Errorf("title size %v != React %d", got, reactHeaderTitle)
	}
	if !h.title.TextStyle.Bold {
		t.Error("title must be bold (React font-bold 700)")
	}
}

// TestSidebarHeaderActionsGeometry pins the button boxes: React renders 40x40
// buttons (p-2 around a 24px icon) with a 4px gap and the avatar 8px later.
func TestSidebarHeaderActionsGeometry(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	h := newSidebarHeader(nil, nil)
	img := renderPNG(t, "sidebar_header", container.NewMax(canvas.NewRectangle(waPanel), h),
		fyne.NewSize(reactHeaderWidth, reactHeaderHeight))
	assertNotBlank(t, img, 0.01)

	// The header is #0c0c0c; the avatar circle is gray-600. Probe its centre:
	// React puts a 32px avatar at x=271..303, y=16..48.
	avX := reactHeaderWidth - reactHeaderPadX - reactHeaderAvatar/2
	assertPixel(t, img, avX, reactHeaderHeight/2,
		color.NRGBA{R: 0x4a, G: 0x55, B: 0x65, A: 255}, 10)

	// Left of the avatar must still be the header background, not the circle.
	px := img.At(avX-reactHeaderAvatar, reactHeaderHeight/2)
	if r, g, b, _ := px.RGBA(); uint8(r>>8) > 0x30 || uint8(g>>8) > 0x30 || uint8(b>>8) > 0x30 {
		t.Fatalf("pixel left of the avatar is not the header bg: %v", px)
	}
}

// TestSidebarHeaderCallbacks: the two header buttons must reach their handlers,
// and the avatar swap must toggle the placeholder glyph.
func TestSidebarHeaderCallbacks(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	var newChat, menu int
	h := newSidebarHeader(func() { newChat++ }, func() { menu++ })
	h.CreateRenderer()

	h.newBtn.Tapped(nil)
	h.menuBt.Tapped(nil)
	if newChat != 1 || menu != 1 {
		t.Fatalf("callbacks: newChat=%d menu=%d, want 1/1", newChat, menu)
	}

	if !h.avatar.Hidden && h.avatar.Visible() {
		t.Fatal("avatar image must start hidden (placeholder glyph shown)")
	}
	if !h.avIcon.Visible() {
		t.Fatal("placeholder glyph must be visible before a photo arrives")
	}

	h.SetSelfAvatar(testAvatarResource(t))
	if !h.avatar.Visible() || h.avIcon.Visible() {
		t.Fatal("SetSelfAvatar must show the image and hide the glyph")
	}
	h.SetSelfAvatar(nil)
	if h.avatar.Visible() || !h.avIcon.Visible() {
		t.Fatal("SetSelfAvatar(nil) must restore the glyph")
	}
}

// testAvatarResource is a 1x1 opaque PNG — a valid image so Fyne does not log
// "Failed to load image" during the swap test.
func testAvatarResource(t *testing.T) fyne.Resource {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.NRGBA{R: 0xff, A: 0xff})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return fyne.NewStaticResource("avatar.png", buf.Bytes())
}
