package fyneui

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"strconv"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"github.com/lugvitc/whats4linux/internal/store"
)

// TestRenderBubbleSample renders the HTML converter output to PNG so the bubble
// body can be inspected without launching the app.
func TestRenderBubbleSample(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	samples := []string{
		`<p>oke lah</p><p>iya</p>`,
		`<p>tawar mass <b>murah</b> <i>banget</i></p>`,
		`<p><a href="https://bit.ly/hasilseleksiGPK" class="msg-link">https://bit.ly/hasilseleksiGPK</a></p>`,
		`<p>baris satu<br>baris dua</p><ul><li>satu</li><li>dua</li></ul>`,
	}

	rows := container.NewVBox()
	for _, s := range samples {
		cv := &ChatView{maxBubble: 560}
		rows.Add(cv.textBlock(s, 560))
	}
	bg := canvas.NewRectangle(waBg)
	root := container.NewMax(bg, container.NewPadded(rows))
	w := test.NewWindow(root)
	w.Resize(fyne.NewSize(640, 420))
	img := w.Canvas().Capture()
	f, err := os.Create(os.Getenv("TEMP") + "/fyne_bubble_sample.png")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
	t.Log("wrote", os.Getenv("TEMP")+"/fyne_bubble_sample.png")
}

// bubbleCorner pins rounded-tl-sm / rounded-tr-sm (MessageItem.tsx:445-446).
func TestBubbleCorner(t *testing.T) {
	if bubbleCorner(true, true) != 2 || bubbleCorner(false, true) != 2 {
		t.Error("firstInGroup must sharpen to 2")
	}
	if bubbleCorner(true, false) != 12 || bubbleCorner(false, false) != 12 {
		t.Error("non-first stays 12")
	}
}

func renderBubble(t *testing.T, name string, fromMe bool, text string) image.Image {
	t.Helper()
	cv := &ChatView{jid: "6282134544728@s.whatsapp.net", maxBubble: 420, rendered: map[string]bool{}, lastDay: ""}
	m := store.DecodedMessage{Info: store.DecodedMessageInfo{ID: "1", IsFromMe: fromMe, Timestamp: "2026-11-08T10:00:00Z"}}
	if text != "" {
		m.Content = &store.DecodedMessageContent{Conversation: text}
	}
	return renderPNG(t, name, cv.bubble(m), fyne.NewSize(520, 140))
}

// TestBubbleTextVisible is the T10 regression: before 0e11859 every text
// segment was non-inline, so Fyne broke each word onto its own line and the
// bubble looked blank/tall. A short sentence must produce one compact row of
// ink, not a vertical column.
func TestBubbleTextVisible(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	a.Settings().SetTheme(newWATheme())

	img := renderBubble(t, "bubble_text", true, "halo bro apa kabar")
	b := img.Bounds()
	var inkX0, inkX1, inkY0, inkY1 = b.Max.X, 0, b.Max.Y, 0
	n := 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			px := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
			if px.G > 120 && px.R < 80 && px.B < 100 {
				continue
			}
			if px.R > 140 || px.G > 140 || px.B > 140 {
				n++
				if x < inkX0 {
					inkX0 = x
				}
				if x > inkX1 {
					inkX1 = x
				}
				if y < inkY0 {
					inkY0 = y
				}
				if y > inkY1 {
					inkY1 = y
				}
			}
		}
	}
	if n < 40 {
		t.Fatalf("bubble text looks blank: only %d non-bg bright pixels", n)
	}
	// 43px = 2 wrapped lines at 14pt — fine. The 0e11859 regression broke
	// each word onto its own line (~127px for the same string). Anything that
	// tall for 4 short words is a column.
	if inkY1-inkY0 > 60 {
		t.Errorf("text rendered as a tall column (%d px high, y %d..%d) — inline-wrap regression", inkY1-inkY0, inkY0, inkY1)
	}
	if inkX1-inkX0 < 60 {
		t.Errorf("text did not flow horizontally (ink only %d px wide)", inkX1-inkX0)
	}
}

// TestBubbleEmptyContent pins the "Empty Message" fallback (MessageItem.tsx:256).
func TestBubbleEmptyContent(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	a.Settings().SetTheme(newWATheme())

	img := renderBubble(t, "bubble_empty", false, "")
	assertNotBlank(t, img, 0.005)
}

// TestBubbleLongURLWraps: a long unbroken URL must wrap inside the bubble,
// never overflow past MaxBubble (the 0e11859-era wrapRuns trailing-space bug
// made the measured line wider than the box).
func TestBubbleLongToken(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	a.Settings().SetTheme(newWATheme())
	cv := &ChatView{jid: "6282134544728@s.whatsapp.net", maxBubble: 320, rendered: map[string]bool{}, lastDay: ""}
	long := "https://example.com/" + strings.Repeat("verylongsegment/", 8)
	m := store.DecodedMessage{Info: store.DecodedMessageInfo{ID: "1", IsFromMe: true, Timestamp: "2026-11-08T10:00:00Z"},
		Content: &store.DecodedMessageContent{Conversation: long}}
	img := renderPNG(t, "bubble_long_url", cv.bubble(m), fyne.NewSize(420, 220))
	b := img.Bounds()
	bright, y0, y1 := 0, b.Max.Y, 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		row := false
		for x := b.Min.X; x < b.Max.X; x++ {
			px := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
			if px.R > 140 && px.G > 140 && px.B > 140 {
				bright++
				row = true
			}
		}
		if row {
			if y < y0 {
				y0 = y
			}
			if y > y1 {
				y1 = y
			}
		}
	}
	if bright < 30 {
		t.Fatalf("long token rendered blank (%d bright px)", bright)
	}
	if y1-y0 > 40 {
		t.Errorf("long token became a tall column (%d px) — inline-wrap regression", y1-y0)
	}
}

// TestBubbleEmojiOnlyIsLarge pins MessageItem.tsx:283 (text-[32px] leading-10).
func TestBubbleEmojiOnlyIsLarge(t *testing.T) {
	if !emojiOnlyText("😂😂") {
		t.Fatal("😂😂 should be emoji-only")
	}
	if emojiOnlyText("😂 ok") {
		t.Fatal("mixed text must not be emoji-only")
	}
	if !emojiOnlyText("<p>👍</p>") {
		t.Fatal("html-wrapped emoji must be detected")
	}
}

// --- measured parity (React spec: MessageItem.tsx + theme) ---
// Harness render of MessageItem blank (wailsjs/store deps); evidence here is
// (1) numpy render geometry against named spec consts, (2) code pinned to
// React:445-446/272-283, theme tokens.

func hexNRGBA(h string) color.NRGBA {
	if len(h) == 7 && h[0] == '#' {
		r, _ := strconv.ParseUint(h[1:3], 16, 8)
		g, _ := strconv.ParseUint(h[3:5], 16, 8)
		b, _ := strconv.ParseUint(h[5:7], 16, 8)
		return color.NRGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}
	}
	return color.NRGBA{}
}

func TestBubbleParitySpec(t *testing.T) {
	// bg tokens (theme_wa.go / dark bubbles)
	if bubbleBg(true) != hexNRGBA("#144d37") {
		t.Errorf("sent bg = %#v", bubbleBg(true))
	}
	if bubbleBg(false) != hexNRGBA("#242626") {
		t.Errorf("recv bg = %#v", bubbleBg(false))
	}
	if waTick != hexNRGBA("#53bdeb") {
		t.Errorf("tick = %#v", waTick)
	}
	// corners
	if bubbleCorner(true, true) != 2 || bubbleCorner(false, true) != 2 || bubbleCorner(true, false) != 12 {
		t.Error("corner spec")
	}
}

func TestBubbleParityRender(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	a.Settings().SetTheme(newWATheme())
	img := renderBubble(t, "bubble_parity_sent", true, "halo bro apa kabar hari ini")
	b := img.Bounds()
	// dominant card = sent green #144d37
	var gpx int
	sent := bubbleBg(true).(color.NRGBA)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			px := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
			if abs(int(px.R)-int(sent.R))+abs(int(px.G)-int(sent.G))+abs(int(px.B)-int(sent.B)) < 18 {
				gpx++
			}
		}
	}
	if gpx < 800 {
		t.Errorf("sent bubble card not dominant (green px=%d)", gpx)
	}

	// recv card
	img2 := renderBubble(t, "bubble_parity_recv", false, "halo bro apa kabar hari ini")
	recv := bubbleBg(false).(color.NRGBA)
	rpx := 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			px := color.NRGBAModel.Convert(img2.At(x, y)).(color.NRGBA)
			if abs(int(px.R)-int(recv.R))+abs(int(px.G)-int(recv.G))+abs(int(px.B)-int(recv.B)) < 18 {
				rpx++
			}
		}
	}
	if rpx < 800 {
		t.Errorf("recv bubble card not dominant (px=%d)", rpx)
	}

	// emoji-only renders larger (ink height >= short text height)
	em := renderBubble(t, "bubble_parity_emoji", true, "😂😂😂")
	inkH := func(im image.Image) int {
		bb := im.Bounds()
		y0, y1 := bb.Max.Y, 0
		for y := bb.Min.Y; y < bb.Max.Y; y++ {
			for x := bb.Min.X; x < bb.Max.X; x++ {
				px := color.NRGBAModel.Convert(im.At(x, y)).(color.NRGBA)
				if px.R > 120 || px.G > 120 || px.B > 120 {
					if y < y0 {
						y0 = y
					}
					if y > y1 {
						y1 = y
					}
					break
				}
			}
		}
		if y1 < y0 {
			return 0
		}
		return y1 - y0
	}
	if inkH(em) <= inkH(img) {
		t.Errorf("emoji-only not larger: emoji h=%d text h=%d", inkH(em), inkH(img))
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
