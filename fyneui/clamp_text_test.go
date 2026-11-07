package fyneui

import (
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

func TestIsLongTextPlain(t *testing.T) {
	if isLongTextPlain("hi") {
		t.Fatal("short")
	}
	if !isLongTextPlain(strings.Repeat("a", 281)) {
		t.Fatal("chars")
	}
	if !isLongTextPlain("1\n2\n3\n4\n5\n6") {
		t.Fatal("lines")
	}
}

func TestTruncatePlain(t *testing.T) {
	s := truncatePlain(strings.Repeat("x", 400), 10, 6)
	if !strings.HasSuffix(s, "…") || len([]rune(s)) != 11 {
		t.Fatalf("got %q len=%d", s, len([]rune(s)))
	}
}

func TestClampTextGolden(t *testing.T) {
	long := strings.Repeat("lorem ipsum ", 40)
	box, w, h := htmlTextBlock(long, 300, 14)
	col := truncatePlain(strings.TrimSpace(htmlToText(long)), longTextChars, longTextLines)
	cb, cw, ch := htmlTextBlock(col, 300, 14)
	wid := newClampText(container.NewGridWrap(fyne.NewSize(w+16, h+16), box), container.NewGridWrap(fyne.NewSize(cw+16, ch+16), cb))
	img := renderPNG(t, "clamp_text", wid, fyne.NewSize(340, 220))
	assertNotBlank(t, img, 0.002)
	if !hasColorNear(img, clampLinkColor(), 20, 20) {
		t.Fatal("missing Read more link color")
	}
}
