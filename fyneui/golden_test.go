package fyneui

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
)

// Golden-render harness. Every component test renders itself headlessly, writes
// a PNG under testdata/out/ and asserts a few pixels so a visual regression
// fails the build instead of only being noticed by eye.

const goldenDir = "testdata/out"

// renderPNG lays obj out at size, captures the canvas and writes the PNG.
// The returned image is what the assertions run against, so a component is
// verified against exactly the pixels a human would review.
func renderPNG(t *testing.T, name string, obj fyne.CanvasObject, size fyne.Size) image.Image {
	t.Helper()
	if size.Width <= 0 || size.Height <= 0 {
		t.Fatalf("renderPNG(%s): non-positive size %v", name, size)
	}

	w := test.NewWindow(obj)
	w.Resize(size)
	// One layout pass so containers settle before capture.
	obj.Resize(size)
	obj.Move(fyne.NewPos(0, 0))

	img := w.Canvas().Capture()
	w.Close()

	if img == nil {
		t.Fatalf("renderPNG(%s): capture returned nil", name)
	}

	dir := filepath.Join(".", goldenDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	f, err := os.Create(filepath.Join(dir, name+".png"))
	if err != nil {
		t.Fatalf("create %s: %v", name, err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("encode %s: %v", name, err)
	}
	return img
}

// blankImage is the reference for assertNotBlank: a canvas painted with bg only.
func blankImage(size fyne.Size, bg color.Color) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, int(size.Width), int(size.Height)))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: bg}, image.Point{}, draw.Src)
	return img
}

// diffRatio is the fraction of pixels whose channels differ by more than tol.
func diffRatio(a, b image.Image) float64 {
	ab, bb := a.Bounds(), b.Bounds()
	w := minInt(ab.Dx(), bb.Dx())
	h := minInt(ab.Dy(), bb.Dy())
	if w == 0 || h == 0 {
		return 1
	}
	const tol = 8
	diff := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r1, g1, b1, _ := a.At(ab.Min.X+x, ab.Min.Y+y).RGBA()
			r2, g2, b2, _ := b.At(bb.Min.X+x, bb.Min.Y+y).RGBA()
			if channelDiff(uint8(r1>>8), uint8(r2>>8)) > tol ||
				channelDiff(uint8(g1>>8), uint8(g2>>8)) > tol ||
				channelDiff(uint8(b1>>8), uint8(b2>>8)) > tol {
				diff++
			}
		}
	}
	return float64(diff) / float64(w*h)
}

func channelDiff(a, b uint8) int {
	if a > b {
		return int(a - b)
	}
	return int(b - a)
}

// assertNotBlank fails when the image is indistinguishable from an empty
// background, which is the exact failure mode of a component that rendered
// nothing (the blank-bubble bug).
func assertNotBlank(t *testing.T, img image.Image, minRatio float64) {
	t.Helper()
	b := img.Bounds()
	r := diffRatio(img, blankImage(fyne.NewSize(float32(b.Dx()), float32(b.Dy())), waBg))
	if r < minRatio {
		t.Fatalf("image looks blank: %.4f of pixels differ from waBg, want >= %.4f", r, minRatio)
	}
}

// assertPixel compares one pixel against want with a per-channel tolerance.
func assertPixel(t *testing.T, img image.Image, x, y int, want color.Color, tol uint8) {
	t.Helper()
	b := img.Bounds()
	if x < b.Min.X || y < b.Min.Y || x >= b.Max.X || y >= b.Max.Y {
		t.Fatalf("assertPixel(%d,%d): outside image bounds %v", x, y, b)
	}
	got := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
	w := color.NRGBAModel.Convert(want).(color.NRGBA)
	if channelDiff(got.R, w.R) > int(tol) ||
		channelDiff(got.G, w.G) > int(tol) ||
		channelDiff(got.B, w.B) > int(tol) {
		t.Fatalf("pixel (%d,%d) = #%02x%02x%02x, want #%02x%02x%02x (tol %d)",
			x, y, got.R, got.G, got.B, w.R, w.G, w.B, tol)
	}
}

// crop returns a copy of the sub-rectangle so assertions can target one region
// (a text block, a badge corner) without scanning the whole component.
func crop(img image.Image, x, y, w, h int) image.Image {
	out := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(out, out.Bounds(), img, image.Point{X: x, Y: y}, draw.Src)
	return out
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestHarnessRendersColor guards the harness itself: if capture stops working
// every other golden test would silently pass on a blank image.
func TestHarnessRendersColor(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	red := color.NRGBA{R: 0xff, A: 255}
	root := container.NewMax(canvas.NewRectangle(red))
	img := renderPNG(t, "harness", root, fyne.NewSize(40, 40))

	assertPixel(t, img, 20, 20, red, 2)
	assertNotBlank(t, img, 0.5)
}

// TestTruncateToWidth covers the helper every truncating label depends on.
func TestTruncateToWidth(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	const size float32 = 14
	long := "Vapers semarangan 18+ dan teman-teman semua"
	got := truncateToWidth(long, 120, size)

	if got == long {
		t.Fatalf("long string was not truncated")
	}
	if w := fyne.MeasureText(got, size, fyne.TextStyle{}).Width; w > 120 {
		t.Fatalf("truncated width %v exceeds 120", w)
	}
	if got[len(got)-3:] != "…" && got[len(got)-1:] != "…" {
		t.Fatalf("truncated string has no ellipsis: %q", got)
	}
	if short := truncateToWidth("halo", 200, size); short != "halo" {
		t.Fatalf("short string changed: %q", short)
	}
}

// TestAvatarHelpers covers the deterministic fallback avatar values.
func TestAvatarHelpers(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	if c1, c2 := avatarColor("6282134544728@s.whatsapp.net"), avatarColor("6282134544728@s.whatsapp.net"); c1 != c2 {
		t.Fatalf("avatarColor is not deterministic: %v vs %v", c1, c2)
	}
	if got := initials("  cintakuuu"); got != "C" {
		t.Fatalf("initials = %q, want C", got)
	}
	if got := initials(""); got != "?" {
		t.Fatalf("initials(empty) = %q, want ?", got)
	}
}
