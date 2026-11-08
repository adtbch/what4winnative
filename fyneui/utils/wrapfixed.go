package utils

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

// WrapFixed pins an object to an exact size; Fyne's GridWrap would otherwise
// let the surrounding layout stretch it. A zero w or h means "let it stretch
// on that axis".
func WrapFixed(obj fyne.CanvasObject, w, h float32) fyne.CanvasObject {
	return container.New(&fixedSizeLayout{w: w, h: h}, obj)
}

type fixedSizeLayout struct{ w, h float32 }

func (l *fixedSizeLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	s := fyne.NewSize(l.w, l.h)
	if len(objects) > 0 {
		m := objects[0].MinSize()
		if l.w == 0 {
			s.Width = m.Width
		}
		if l.h == 0 {
			s.Height = m.Height
		}
	}
	return s
}

func (l *fixedSizeLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	w, h := l.w, l.h
	if w == 0 {
		w = size.Width
	}
	if h == 0 {
		h = size.Height
	}
	for _, o := range objects {
		o.Move(fyne.NewPos(0, 0))
		o.Resize(fyne.NewSize(w, h))
	}
}
