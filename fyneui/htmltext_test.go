package fyneui

import (
	"image/color"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

// TestRichTextFlowsInline guards the empty-bubble bug: every segment must be
// Inline, otherwise Fyne starts a new paragraph per segment and a sentence
// renders one word per line.
func TestRichTextFlowsInline(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	box, _, _ := htmlTextBlock("<p>hello world ini pesan biasa</p>", 504, 14)
	rt, ok := box.(*widget.RichText)
	if !ok {
		t.Fatalf("htmlTextBlock returned %T, want *widget.RichText", box)
	}
	for i, seg := range rt.Segments {
		ts := seg.(*widget.TextSegment)
		if !ts.Style.Inline {
			t.Fatalf("segment %d (%q) is not inline; Fyne will break the line there", i, ts.Text)
		}
	}
	// One visual line of 19px, not five lines of ~127px.
	if h := rt.MinSize().Height; h > 60 {
		t.Fatalf("RichText height %.1f suggests one word per line", h)
	}
}

// TestBubbleTextIsOneLine renders a short message and asserts a single ink
// band, i.e. the words really flow on one line.
func TestBubbleTextIsOneLine(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	cv := &ChatView{maxBubble: 560}
	blk := cv.textBlock("<p>hello world ini pesan biasa</p>", 560)
	img := renderPNG(t, "bubble_one_line",
		container.NewMax(canvas.NewRectangle(waRecv), container.NewCenter(blk)),
		fyne.NewSize(620, 100))
	assertNotBlank(t, img, 0.002)

	bg := color.NRGBAModel.Convert(img.At(0, 0)).(color.NRGBA)
	bands := 0
	inBand := false
	for y := 0; y < img.Bounds().Dy(); y++ {
		ink := false
		for x := 0; x < img.Bounds().Dx(); x++ {
			c := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
			if absDiff(int(c.R), int(bg.R))+absDiff(int(c.G), int(bg.G))+absDiff(int(c.B), int(bg.B)) > 25 {
				ink = true
				break
			}
		}
		if ink && !inBand {
			bands++
		}
		inBand = ink
	}
	if bands != 1 {
		t.Fatalf("found %d text bands, want 1 (words must flow on one line)", bands)
	}
}

// TestWrapRespectsMaxWidth: the measured line width must never exceed maxW,
// otherwise RichText re-wraps inside its padding and the bubble looks ragged.
func TestWrapRespectsMaxWidth(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	src := "<p>ini pesan yang cukup panjang supaya harus wrap ke beberapa baris karena melebihi lebar bubble maksimum yang tersedia di chat</p>"
	const maxW float32 = 532
	lines := wrapRuns(parseHTMLRuns(src), maxW, 14)
	if len(lines) < 2 {
		t.Fatalf("expected the sample to wrap, got %d line(s)", len(lines))
	}
	for i, ln := range lines {
		if w := lineWidth(ln, 14); w > maxW {
			t.Fatalf("line %d is %.1fpx wide, over the %.1fpx budget", i, w, maxW)
		}
	}
	// Reassembling must preserve every word, in order.
	var got []string
	for _, ln := range lines {
		for _, r := range ln {
			got = append(got, strings.Fields(r.text)...)
		}
	}
	want := strings.Fields(htmlToText(src))
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Fatalf("wrapped text lost words:\n got %q\nwant %q", strings.Join(got, " "), strings.Join(want, " "))
	}
}

func absDiff(a, b int) int {
	if a > b {
		return a - b
	}
	return b - a
}
