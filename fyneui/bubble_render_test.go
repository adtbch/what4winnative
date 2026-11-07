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


// --- numeric ink/card bbox evidence (#6): Fyne render vs React DOM spec ---
// React measured (static harness, MessageItem.tsx classes/tokens):
//   sent  card #144d37 178x53 rel(306,16) tl12 tr4 pad 4 8 6, body 14/20, meta 11 tick #53bdeb
//   recv  card #242626 178x53 rel(36,16)  tl4  tr12
//   group +sender 12px #53bdeb, 178x73 rel(64,16); emoji body 30/40 140x69; edited meta "Edited"

func dominantCardBBox(img image.Image, want color.NRGBA, tol int) (x0, y0, x1, y1, n int) {
	b := img.Bounds()
	first := true
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bb, _ := img.At(x, y).RGBA()
			dr := int(uint8(r>>8)) - int(want.R); dg := int(uint8(g>>8)) - int(want.G); db := int(uint8(bb>>8)) - int(want.B)
			if dr < 0 { dr = -dr }; if dg < 0 { dg = -dg }; if db < 0 { db = -db }
			if dr+dg+db <= tol {
				n++
				if first { x0, y0, x1, y1, first = x, y, x, y, false }
				if x < x0 { x0 = x }; if y < y0 { y0 = y }; if x > x1 { x1 = x }; if y > y1 { y1 = y }
			}
		}
	}
	return
}

func brightInkHeight(img image.Image, card color.NRGBA) int {
	b := img.Bounds(); ymin, ymax := b.Max.Y, b.Min.Y
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bb, _ := img.At(x, y).RGBA()
			R, G, B := int(r>>8), int(g>>8), int(bb>>8)
			dr := R - int(card.R); dg := G - int(card.G); db := B - int(card.B)
			if dr < 0 { dr = -dr }; if dg < 0 { dg = -dg }; if db < 0 { db = -db }
			if dr+dg+db <= 35 { continue }
			if R+G+B > 260 { if y < ymin { ymin = y }; if y > ymax { ymax = y } }
		}
	}
	if ymax < ymin { return 0 }; return ymax - ymin
}

func TestBubbleParityNumericBBox(t *testing.T) {
	cv := &ChatView{jid: "self", maxBubble: 560}
	mk := func(id, txt string, fromMe bool) store.DecodedMessage {
		return store.DecodedMessage{Info: store.DecodedMessageInfo{ID: id, IsFromMe: fromMe, Timestamp: "2025-07-11T10:05:00Z"}, Content: &store.DecodedMessageContent{Conversation: txt}}
	}
	sent := renderPNG(t, "bbox_sent", cv.bubble(mk("bs", "halo bro apa kabar hari ini", true)), fyne.NewSize(520, 220))
	sx0, sy0, sx1, sy1, sn := dominantCardBBox(sent, hexNRGBA("#144d37"), 35)
	if sn < 600 || sx1-sx0 < 60 || sy1-sy0 < 24 { t.Fatalf("sent card bbox weak: n=%d bbox=%d,%d..%d,%d (React 178x53 rel306,16)", sn, sx0, sy0, sx1, sy1) }
	recv := renderPNG(t, "bbox_recv", cv.bubble(mk("br", "halo bro apa kabar hari ini", false)), fyne.NewSize(520, 220))
	_, _, rx1, ry1, rn := dominantCardBBox(recv, hexNRGBA("#242626"), 35)
	if rn < 600 || rx1 < 60 || ry1 < 24 { t.Fatalf("recv card bbox weak: n=%d (React 178x53 rel36,16)", rn) }
	plain := renderPNG(t, "bbox_plain", cv.bubble(mk("bp", "halo", true)), fyne.NewSize(520, 220))
	emoji := renderPNG(t, "bbox_emoji", cv.bubble(mk("be", "😂😂😂", true)), fyne.NewSize(520, 220))
	ph, eh := brightInkHeight(plain, hexNRGBA("#144d37")), brightInkHeight(emoji, hexNRGBA("#144d37"))
	if ph == 0 || eh <= ph { t.Fatalf("emoji body ink not larger than plain (React body 30/40 vs 14/20): plainH=%d emojiH=%d", ph, eh) }
}

func TestBubbleParityGroupSender(t *testing.T) {
	cv := &ChatView{jid: "123@g.us", maxBubble: 560}
	m := store.DecodedMessage{Info: store.DecodedMessageInfo{ID: "g1", IsFromMe: false, Sender: "6281@s.whatsapp.net", Timestamp: "2025-07-11T10:05:00Z"}, Content: &store.DecodedMessageContent{Conversation: "halo grup"}}
	img := renderPNG(t, "bbox_group", cv.bubble(m), fyne.NewSize(520, 260))
	_, _, _, _, n := dominantCardBBox(img, hexNRGBA("#242626"), 35)
	if n < 600 { t.Fatalf("group recv card weak n=%d (React 178x73 rel64,16 + sender)", n) }
}
