package fyneui

import (
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

// mediaBubbleData is the render input for T11 (MediaContent.tsx / MessageItem).
// Thumb is optional; nil renders the placeholder state.
type mediaBubbleData struct {
	Kind      int // mediaKindImage/Video/Sticker (chatview consts)
	Thumb     fyne.Resource
	Caption   string
	FileName  string
	SizeLabel string // document/audio secondary label
	Width     int
	Height    int
	MaxWidth  float32
	IsGif     bool
}

// newMediaBubble renders a media body (no outer bubble chrome; chatview/wrapBubble
// supplies margins and, except stickers, the bubble card).
func newMediaBubble(d mediaBubbleData) fyne.CanvasObject {
	switch d.Kind {
	case mediaKindSticker:
		return newStickerBubble(d)
	case mediaKindVideo:
		return newImageVideoBubble(d, true)
	case mediaKindImage:
		return newImageVideoBubble(d, false)
	case mediaKindAudio:
		return newAudioBubble(d)
	case mediaKindDocument:
		return newDocumentBubble(d)
	default:
		return newImageVideoBubble(d, false)
	}
}

func newImageVideoBubble(d mediaBubbleData, video bool) fyne.CanvasObject {
	maxW := d.MaxWidth
	if maxW <= 0 {
		maxW = 330
	}
	size := mediaBox(d.Width, d.Height)
	if size.Width > maxW {
		r := maxW / size.Width
		size = fyne.NewSize(maxW, maxF(size.Height*r, 80))
	}
	holder := canvas.NewRectangle(waPanel2)
	holder.CornerRadius = 8
	ph := "📷 Photo"
	if video {
		ph = "🎥 Video"
		if d.IsGif {
			ph = "GIF"
		}
	}
	phTxt := canvas.NewText(ph+" (offline/cache)", waMuted)
	phTxt.TextSize = 13
	phTxt.Alignment = fyne.TextAlignCenter

	img := canvas.NewImageFromResource(d.Thumb)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(size)
	box := container.NewMax(holder, container.NewCenter(phTxt), img)

	if video {
		play := canvas.NewText("▶", waWhite)
		play.TextSize = 34
		play.Alignment = fyne.TextAlignCenter
		circle := canvas.NewCircle(waPlayOverlay)
		overlay := container.NewCenter(container.NewGridWrap(fyne.NewSize(56, 56), container.NewMax(circle, container.NewCenter(play))))
		return container.NewMax(container.NewGridWrap(size, box), overlay)
	}
	return container.NewGridWrap(size, box)
}

func newStickerBubble(d mediaBubbleData) fyne.CanvasObject {
	size := fyne.NewSize(150, 150)
	if d.Thumb == nil {
		// no bubble background for stickers (MessageItem: bg-transparent shadow-none)
		t := canvas.NewText("Sticker", waMuted)
		t.TextSize = 12
		t.Alignment = fyne.TextAlignCenter
		return container.NewGridWrap(size, container.NewCenter(t))
	}
	img := canvas.NewImageFromResource(d.Thumb)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(size)
	return container.NewGridWrap(size, img)
}

func newAudioBubble(d mediaBubbleData) fyne.CanvasObject {
	icon := canvas.NewText("▶", waText)
	icon.TextSize = 18
	circle := canvas.NewCircle(color.NRGBA{R: 0x1b, G: 0x9a, B: 0x58, A: 255})
	btn := container.NewGridWrap(fyne.NewSize(36, 36), container.NewMax(circle, container.NewCenter(icon)))
	label := d.SizeLabel
	if strings.TrimSpace(label) == "" {
		label = "Audio message"
	}
	txt := canvas.NewText(label, waMuted)
	txt.TextSize = 13
	bar := canvas.NewRectangle(color.NRGBA{R: 255, G: 255, B: 255, A: 40})
	bar.SetMinSize(fyne.NewSize(140, 3))
	row := container.NewHBox(btn, spacer(8, 1), container.NewVBox(spacer(1, 8), txt, spacer(1, 4), bar))
	pill := canvas.NewRectangle(waPanel2)
	pill.CornerRadius = 10
	return container.NewMax(pill, container.NewPadded(row))
}

func newDocumentBubble(d mediaBubbleData) fyne.CanvasObject {
	name := strings.TrimSpace(d.FileName)
	if name == "" {
		name = "Document"
	}
	ext := "FILE"
	if i := strings.LastIndex(name, "."); i >= 0 && i < len(name)-1 {
		ext = strings.ToUpper(name[i+1:])
		if len([]rune(ext)) > 4 {
			ext = string([]rune(ext)[:4])
		}
	}
	// red chip for file type (MessageItem: w-10 h-12 bg-red-500)
	chip := canvas.NewRectangle(color.NRGBA{R: 239, G: 68, B: 68, A: 255}) // red-500
	chip.CornerRadius = 4
	et := canvas.NewText(ext, waWhite)
	et.TextSize = 10
	et.TextStyle = fyne.TextStyle{Bold: true}
	et.Alignment = fyne.TextAlignCenter
	badge := container.NewGridWrap(fyne.NewSize(40, 48), container.NewMax(chip, container.NewCenter(et)))

	title := canvas.NewText(truncateRunes(name, 38), waText)
	title.TextSize = 13
	sub := strings.TrimSpace(d.SizeLabel)
	if sub == "" {
		sub = "Document"
	}
	st := canvas.NewText(sub, waMuted)
	st.TextSize = 11
	col := container.NewVBox(title, st)
	dl := canvas.NewText("↓", waMuted)
	dl.TextSize = 18
	row := container.NewBorder(nil, nil, badge, container.NewPadded(dl), container.NewPadded(col))
	card := canvas.NewRectangle(color.NRGBA{R: 0, G: 0, B: 0, A: 28})
	card.CornerRadius = 8
	return container.NewMax(card, container.NewPadded(row))
}

// stickerNoBubble reports whether this content should render without the outer
// bubble rectangle (MessageItem.tsx:442 bg-transparent for stickers).
func stickerNoBubble(contentKind int) bool { return contentKind == mediaKindSticker }

// formatDocLabel is a tiny helper kept for tests/callers.
func formatDocLabel(name string, size int64) string {
	if size > 0 {
		return fmt.Sprintf("%s", name)
	}
	return "Document"
}
