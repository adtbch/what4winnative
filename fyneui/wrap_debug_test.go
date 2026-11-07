package fyneui

import (
	"testing"

	"fyne.io/fyne/v2/test"
)

func TestWrapDebug(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	src := `<p>oke lah</p><p>iya</p>`
	runs := parseHTMLRuns(src)
	for _, r := range runs {
		t.Logf("run %q bold=%v ital=%v link=%v w=%v", r.text, r.bold, r.ital, r.link, runWidth(r, 14))
	}
	lines := wrapRuns(runs, 532, 14)
	t.Logf("lines=%d", len(lines))
	for i, l := range lines {
		var s string
		for _, r := range l {
			s += r.text
		}
		t.Logf("line %d = %q", i, s)
	}
	_, w, h := htmlTextBlock(src, 532, 14)
	t.Logf("block w=%v h=%v", w, h)
}
