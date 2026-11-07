package fyneui

import (
	"image"
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/test"
)

// Measured from frontend/src/harness.tsx `out.chatHeader` plus a numpy scan of
// fyneui/testdata/react/chatheader_dark_600x700.png (600px window):
//
//	box      600x65  padding 12px  bg rgb(12,12,12)  border-b 1px white/5
//	back     24x24   x=12 y=20   ink x 16..31 y 24..39
//	avatar   40x40   x=64 y=12   bg #4a5565  ink x 64..103 y 12..51
//	name     x=116 y=12  16px/24/500  #f3f4f6  ink y 18..29
//	subtitle x=116 y=36  12px/16      rgb(134,150,160)  ink y 40..48
//	call     38x38   x=522 y=13  p-2  icon 22  ink x 532..549
//	info     28x28   x=560 y=18  p-1  icon 20  ink x 565..582
const (
	reactHeaderW      = 600
	reactHeaderH      = 65
	reactHeaderBackX  = 12
	reactHeaderAvX    = 64
	reactHeaderTextX  = 116
	reactHeaderCallX  = 522
	reactHeaderInfoX  = 560
	reactHeaderCallSz = 38
	reactHeaderInfoSz = 28
)

func renderHeader(t *testing.T, d chatHeaderData, w float32) (*chatHeader, image.Image) {
	t.Helper()
	h := newChatHeader(func() {}, func() {}, d)
	return h, renderPNG(t, "chat_header", h, fyne.NewSize(w, reactHeaderH))
}

func TestChatHeaderGolden(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	a.Settings().SetTheme(newWATheme())

	h, img := renderHeader(t, chatHeaderData{
		Name:     "cintakuuu",
		Subtitle: "online",
		Call:     func() {},
	}, reactHeaderW)

	if sz := h.MinSize(); sz.Height != reactHeaderH {
		t.Errorf("header height %.0f, want %d", sz.Height, reactHeaderH)
	}

	// Back arrow: 24px GoBackIcon, ink starts 4/24 into the box → x = 16.
	if x := inkLeft(t, img, 0, 60, 0, reactHeaderH); x < 14 || x > 18 {
		t.Errorf("back ink x=%d, want ~16", x)
	}

	// Avatar: 40px disc at x=64, fill #4a5565.
	assertDisc(t, img, reactHeaderAvX, reactHeaderAvX+40, 12, 52, chatHeaderAvatarBg)

	// Name and subtitle share the text column at x=116.
	if x := inkLeft(t, img, 100, reactHeaderW, 8, 36); x < 114 || x > 118 {
		t.Errorf("name ink x=%d, want 116", x)
	}
	if x := inkLeft(t, img, 100, reactHeaderW, 36, 56); x < 114 || x > 118 {
		t.Errorf("subtitle ink x=%d, want 116", x)
	}
}

// TestChatHeaderGroupHidesCall pins ChatDetail.tsx:709 — onCallClick is only
// passed for contacts, so a group header has just the info button.
func TestChatHeaderGroupHidesCall(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	a.Settings().SetTheme(newWATheme())

	h, _ := renderHeader(t, chatHeaderData{
		Name:     "Vapers Semarangan 18+",
		Subtitle: "You, Budi, ~ Sari",
	}, reactHeaderW)

	if h.call.Visible() {
		t.Error("group header must not render the call button")
	}
	if !h.info.Visible() {
		t.Error("info button must always render")
	}
}

func TestChatHeaderInfoPinnedRight(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	a.Settings().SetTheme(newWATheme())

	h, _ := renderHeader(t, chatHeaderData{Name: "cintakuuu", Call: func() {}}, reactHeaderW)

	// 12px right gutter: the 28px info box ends at 600-12 = 588.
	if x := absX(t, h.info); x != reactHeaderInfoX {
		t.Errorf("info x=%d, want %d", x, reactHeaderInfoX)
	}
	// Call sits immediately left of it: 560-38 = 522.
	if x := absX(t, h.call); x != reactHeaderCallX {
		t.Errorf("call x=%d, want %d", x, reactHeaderCallX)
	}
}

func TestChatHeaderTruncatesName(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	h := newChatHeader(nil, nil, chatHeaderData{
		Name:     "Nama grup yang sangat panjang sekali sampai tidak muat di header",
		Subtitle: "sub",
		Call:     func() {},
	})
	h.Resize(fyne.NewSize(reactHeaderW, reactHeaderH))

	// `truncate` must cut, not overflow: the drawn string is shorter than the
	// source and ends with an ellipsis.
	if h.name.Text == h.nameS {
		t.Error("long name was not truncated")
	}
	if w := fyne.MeasureText(h.name.Text, chatHeaderNameSize, fyne.TextStyle{Bold: true}).Width; w > reactHeaderW {
		t.Errorf("truncated name is %0.f px wide, wider than the header", w)
	}
}

// TestChatHeaderAvatarInitial pins the fallback: `chatName.substring(0,1)`
// uppercased, never a photo.
func TestChatHeaderAvatarInitial(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	h := newChatHeader(nil, nil, chatHeaderData{Name: "cintakuuu"})
	if h.avTxt.Text != "C" {
		t.Errorf("avatar initial = %q, want C", h.avTxt.Text)
	}
	if h.avImg.Visible() {
		t.Error("photo must stay hidden without an avatar resource")
	}
}

func absX(t *testing.T, o fyne.CanvasObject) int {
	t.Helper()
	return int(o.Position().X)
}

// assertDisc finds a run of the given fill inside the box.
func assertDisc(t *testing.T, img image.Image, x0, x1, y0, y1 int, want color.Color) {
	t.Helper()
	w := color.NRGBAModel.Convert(want).(color.NRGBA)
	hits := 0
	for y := y0; y < y1 && y < img.Bounds().Dy(); y++ {
		for x := x0; x < x1 && x < img.Bounds().Dx(); x++ {
			px := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
			if pxDiff(px.R, w.R) < 8 && pxDiff(px.G, w.G) < 8 && pxDiff(px.B, w.B) < 8 {
				hits++
			}
		}
	}
	if hits < 200 {
		t.Errorf("avatar fill only %d px in x %d..%d y %d..%d", hits, x0, x1, y0, y1)
	}
}

func pxDiff(a, b uint8) int {
	if a > b {
		return int(a - b)
	}
	return int(b - a)
}

var _ = canvas.NewText

// TestChatHeaderBackHiddenOnDesktop pins `md:hidden` on the back button: at
// >=768px the sidebar is the navigation, so the arrow disappears and the
// avatar slides left to the 12px gutter.
func TestChatHeaderBackHiddenOnDesktop(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	a.Settings().SetTheme(newWATheme())

	h := newChatHeader(func() {}, nil, chatHeaderData{Name: "cintakuuu"})
	h.Resize(fyne.NewSize(900, reactHeaderH))
	if h.back.Visible() {
		t.Error("back button must be hidden at 900px (md:hidden)")
	}

	h.Resize(fyne.NewSize(600, reactHeaderH))
	if !h.back.Visible() {
		t.Error("back button must show at 600px")
	}
}
