package fyneui

import (
	"image/png"
	"os"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
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
