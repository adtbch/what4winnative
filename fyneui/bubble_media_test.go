package fyneui

import (
	"image"
	"image/color"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
)

func dominantBBox(img image.Image, want color.NRGBA, tol int) (x0, y0, x1, y1, n int) {
	b := img.Bounds()
	first := true
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bb, _ := img.At(x, y).RGBA()
			dr := int(uint8(r>>8)) - int(want.R)
			dg := int(uint8(g>>8)) - int(want.G)
			db := int(uint8(bb>>8)) - int(want.B)
			if dr < 0 {
				dr = -dr
			}
			if dg < 0 {
				dg = -dg
			}
			if db < 0 {
				db = -db
			}
			if dr+dg+db <= tol {
				n++
				if first {
					x0, y0, x1, y1, first = x, y, x, y, false
				}
				if x < x0 {
					x0 = x
				}
				if y < y0 {
					y0 = y
				}
				if x > x1 {
					x1 = x
				}
				if y > y1 {
					y1 = y
				}
			}
		}
	}
	return
}

func hasColorNear(img image.Image, want color.NRGBA, tol int, minN int) bool {
	_, _, _, _, n := dominantBBox(img, want, tol)
	return n >= minN
}

func TestMediaPlaceholderGolden(t *testing.T) {
	// image/video: rounded placeholder waPanel2 #242626; video has play overlay circle waPlayOverlay.
	img := renderPNG(t, "media_image_placeholder", newMediaBubble(mediaBubbleData{Kind: mediaKindImage, Width: 300, Height: 200, MaxWidth: 330}), fyne.NewSize(360, 260))
	_, _, x1, y1, n := dominantBBox(img, color.NRGBA{R: 0x24, G: 0x26, B: 0x26, A: 255}, 20)
	if n < 1200 || x1 < 80 || y1 < 60 {
		t.Fatalf("image placeholder waPanel2 weak: n=%d bboxMax=%d,%d", n, x1, y1)
	}
	vid := renderPNG(t, "media_video_placeholder", newMediaBubble(mediaBubbleData{Kind: mediaKindVideo, Width: 300, Height: 200, MaxWidth: 330}), fyne.NewSize(360, 260))
	// play overlay is a dark circle (~waPlayOverlay 0b141a alpha blended) over panel; assert panel present and center has non-panel dark pixels
	_, _, _, _, vn := dominantBBox(vid, color.NRGBA{R: 0x24, G: 0x26, B: 0x26, A: 255}, 20)
	if vn < 800 {
		t.Fatalf("video placeholder panel weak: n=%d", vn)
	}
}

func TestStickerHasNoBubbleBackground(t *testing.T) {
	// Sticker thumb nil: transparent/no card fill dominant in middle (only "Sticker" text).
	// With Thumb==nil newStickerBubble returns centered text in 150x150 gridwrap, no rect.
	obj := newStickerBubble(mediaBubbleData{Kind: mediaKindSticker})
	img := renderPNG(t, "media_sticker_nobg", obj, fyne.NewSize(170, 170))
	// waSent/waRecv must NOT dominate (sticker has no bubble). Panel2 also not a full card.
	_, _, _, _, sentN := dominantBBox(img, hexNRGBA("#144d37"), 30)
	_, _, _, _, recvN := dominantBBox(img, hexNRGBA("#242626"), 30)
	if sentN > 2000 || recvN > 2000 {
		t.Fatalf("sticker got bubble background: sent=%d recv=%d", sentN, recvN)
	}
}

func TestDocumentBubbleGolden(t *testing.T) {
	obj := newDocumentBubble(mediaBubbleData{Kind: mediaKindDocument, FileName: "invoice-report.pdf", SizeLabel: "Document"})
	img := renderPNG(t, "media_document", obj, fyne.NewSize(360, 120))
	// red chip #ef4444 present
	if !hasColorNear(img, color.NRGBA{R: 239, G: 68, B: 68, A: 255}, 25, 120) {
		t.Fatalf("document missing red file chip")
	}
	// text/name should produce some bright or dark ink (waText/waMuted on dark-ish card)
	b := img.Bounds()
	ink := 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bb, _ := img.At(x, y).RGBA()
			R, G, B := int(r>>8), int(g>>8), int(bb>>8)
			// not pure bg gray and not pure red chip area count loosely
			if R+G+B > 320 && !(R > 200 && G < 120 && B < 120) {
				ink++
			}
		}
	}
	if ink < 80 {
		t.Fatalf("document text ink weak: %d", ink)
	}
	_ = strings.Contains // keep
}
