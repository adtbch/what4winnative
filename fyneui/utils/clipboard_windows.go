//go:build windows

package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"syscall"
	"unsafe"
)

var (
	user32Clip           = syscall.NewLazyDLL("user32.dll")
	kernel32Clip         = syscall.NewLazyDLL("kernel32.dll")
	procOpenClipboard    = user32Clip.NewProc("OpenClipboard")
	procCloseClipboard   = user32Clip.NewProc("CloseClipboard")
	procGetClipboardData = user32Clip.NewProc("GetClipboardData")
	procGlobalLock       = kernel32Clip.NewProc("GlobalLock")
	procGlobalUnlock     = kernel32Clip.NewProc("GlobalUnlock")
	procGlobalSize       = kernel32Clip.NewProc("GlobalSize")
)

const cfDIB = 8

// readClipboardImagePNG reads a DIB/bitmap image from the Windows clipboard
// and returns PNG-encoded bytes. Returns (nil, nil) when clipboard has no image.
func ReadClipboardImagePNG() ([]byte, error) {
	r, _, _ := procOpenClipboard.Call(0)
	if r == 0 {
		return nil, fmt.Errorf("open clipboard failed")
	}
	defer procCloseClipboard.Call()

	hMem, _, _ := procGetClipboardData.Call(uintptr(cfDIB))
	if hMem == 0 {
		return nil, nil // no image in clipboard
	}

	pData, _, _ := procGlobalLock.Call(hMem)
	if pData == 0 {
		return nil, fmt.Errorf("global lock failed")
	}
	defer procGlobalUnlock.Call(hMem)

	size, _, _ := procGlobalSize.Call(hMem)
	if size == 0 || size < 40 {
		return nil, nil
	}

	data := unsafe.Slice((*byte)(unsafe.Pointer(pData)), int(size))
	img, err := decodeDIB(data)
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		return nil, fmt.Errorf("png encode: %w", err)
	}
	return out.Bytes(), nil
}

func decodeDIB(data []byte) (image.Image, error) {
	if len(data) < 40 {
		return nil, fmt.Errorf("DIB header too short")
	}
	biSize := binary.LittleEndian.Uint32(data[0:4])
	biWidth := int32(binary.LittleEndian.Uint32(data[4:8]))
	biHeight := int32(binary.LittleEndian.Uint32(data[8:12]))
	biBitCount := binary.LittleEndian.Uint16(data[14:16])
	biCompression := binary.LittleEndian.Uint32(data[16:20])
	if biWidth == 0 || biHeight == 0 {
		return nil, fmt.Errorf("invalid DIB dimensions")
	}

	topDown := false
	h := biHeight
	if h < 0 {
		topDown = true
		h = -h
	}
	w := int(biWidth)
	hh := int(h)

	offset := biSize
	var rMask, gMask, bMask uint32 = 0x00FF0000, 0x0000FF00, 0x000000FF
	if biCompression == 3 { // BI_BITFIELDS
		if len(data) >= int(biSize)+12 {
			rMask = binary.LittleEndian.Uint32(data[biSize : biSize+4])
			gMask = binary.LittleEndian.Uint32(data[biSize+4 : biSize+8])
			bMask = binary.LittleEndian.Uint32(data[biSize+8 : biSize+12])
			offset += 12
		}
	}
	if int(offset) > len(data) {
		return nil, fmt.Errorf("DIB offset out of range")
	}
	pixels := data[offset:]

	stride := ((w*int(biBitCount) + 31) / 32) * 4
	img := image.NewNRGBA(image.Rect(0, 0, w, hh))

	for y := 0; y < hh; y++ {
		srcY := y
		if !topDown {
			srcY = hh - 1 - y
		}
		rowStart := srcY * stride
		if rowStart+stride > len(pixels) {
			break
		}
		row := pixels[rowStart : rowStart+stride]
		for x := 0; x < w; x++ {
			switch biBitCount {
			case 32:
				idx := x * 4
				if idx+4 > len(row) {
					continue
				}
				px := binary.LittleEndian.Uint32(row[idx : idx+4])
				var rr, gg, bb byte
				if biCompression == 3 {
					rr = maskByte(px, rMask)
					gg = maskByte(px, gMask)
					bb = maskByte(px, bMask)
				} else {
					bb = row[idx]
					gg = row[idx+1]
					rr = row[idx+2]
				}
				aa := row[idx+3]
				if aa == 0 && rr == 0 && gg == 0 && bb == 0 {
					aa = 255 // many DIBs leave alpha 0 for opaque
				}
				img.SetNRGBA(x, y, color.NRGBA{R: rr, G: gg, B: bb, A: aa})
			case 24:
				idx := x * 3
				if idx+3 > len(row) {
					continue
				}
				img.SetNRGBA(x, y, color.NRGBA{R: row[idx+2], G: row[idx+1], B: row[idx], A: 255})
			}
		}
	}
	return img, nil
}

func maskByte(v, mask uint32) byte {
	if mask == 0 {
		return 0
	}
	m := mask
	shift := 0
	for m&1 == 0 {
		m >>= 1
		shift++
	}
	return byte((v & mask) >> shift)
}
