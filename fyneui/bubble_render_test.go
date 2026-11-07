package fyneui

import (
	"image"
	"image/color"
	"image/png"
	"os"
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
