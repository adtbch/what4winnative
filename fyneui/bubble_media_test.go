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

// --- React spec parity (static harness, MessageItem.tsx/MediaContent.tsx tokens) ---
// image/video: recv bubble bg rgb(36,38,38)=#242626, placeholder 300x200 bg #242626 r8, play 56x56 rgba(11,20,26,.63) r=full, play glyph 34px
// sticker: bg transparent 150x150 (MessageItem:442 bg-transparent)
// audio: bubble #242626 156x71, play 36x36 bg #1b9a58 full-round
// document: bubble #242626 256x93, chip red-500 40x48 r4, name 14px
func TestMediaParityNumeric(t *testing.T) {
	// Image placeholder: dominant panel #242626, compact (React ph 300x200)
	img := renderPNG(t, "parity_media_img", newMediaBubble(mediaBubbleData{Kind: mediaKindImage, Width: 300, Height: 200, MaxWidth: 330}), fyne.NewSize(360, 260))
	_, _, ix1, iy1, in := dominantCardBBox(img, hexNRGBA("#242626"), 25)
	if in < 8000 || ix1 < 150 || iy1 < 100 {
		t.Fatalf("image panel weak n=%d max=%d,%d (React ph 300x200 bg #242626)", in, ix1, iy1)
	}
	// Video: has overlay dark pixels near center (React play 56x56 rgba(11,20,26,.63))
	vid := renderPNG(t, "parity_media_vid", newMediaBubble(mediaBubbleData{Kind: mediaKindVideo, Width: 300, Height: 200, MaxWidth: 330}), fyne.NewSize(360, 260))
	b := vid.Bounds()
	dark := 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bb, _ := vid.At(x, y).RGBA()
			R, G, B := int(r>>8), int(g>>8), int(bb>>8)
			if R < 30 && G < 40 && B < 45 && R+G+B > 10 && R+G+B < 90 {
				dark++
			}
		}
	}
	if dark < 40 {
		t.Fatalf("video play overlay weak dark=%d (React 56x56 rgba(11,20,26,.63))", dark)
	}
	// Sticker: no bubble/card fill (React transparent 150x150)
	st := renderPNG(t, "parity_media_sticker", newStickerBubble(mediaBubbleData{Kind: mediaKindSticker}), fyne.NewSize(170, 170))
	_, _, _, _, sn := dominantCardBBox(st, hexNRGBA("#144d37"), 30)
	_, _, _, _, rn := dominantCardBBox(st, hexNRGBA("#242626"), 30)
	if sn > 1500 || rn > 1500 {
		t.Fatalf("sticker not transparent (React bg transparent): sent=%d recv=%d", sn, rn)
	}
	// Audio: green play present (React 36x36 #1b9a58)
	au := renderPNG(t, "parity_media_audio", newAudioBubble(mediaBubbleData{Kind: mediaKindAudio}), fyne.NewSize(220, 100))
	if !hasColorNear(au, color.NRGBA{R: 0x1b, G: 0x9a, B: 0x58, A: 255}, 25, 120) {
		t.Fatalf("audio green play missing (React #1b9a58 36x36)")
	}
	// Document: red chip present (React 40x48 red-500)
	dc := renderPNG(t, "parity_media_doc", newDocumentBubble(mediaBubbleData{Kind: mediaKindDocument, FileName: "invoice-report.pdf", SizeLabel: "Document"}), fyne.NewSize(360, 120))
	if !hasColorNear(dc, color.NRGBA{R: 239, G: 68, B: 68, A: 255}, 25, 120) {
		t.Fatalf("doc red chip missing (React 40x48 red-500)")
	}
}
