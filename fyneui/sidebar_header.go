package fyneui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Geometry measured from ChatScreen.tsx via the React harness
// (`out.header` in harness.tsx), not hand-tuned:
//
//	row:     h 64, padding 0 16, justify space-between, border-b 1px white/5
//	title:   "WhatsApp" 20px / 28px line-height / weight 700 / #fff
//	button:  40x40 (p-2 around a 24px icon), rounded-full, gap 4
//	avatar:  32x32, rounded-full, ml-2 (8px), bg gray-600
//
// Buttons are 40x40 so the row's 64px height is dominated by the 40px hit area,
// exactly like React's p-2 buttons.
const (
	headerHeight     = 64
	headerPadX       = 16
	headerBtnSize    = 40
	headerIconSize   = 24
	headerBtnGap     = 4
	headerAvatarSize = 32
	headerAvatarML   = 8
	headerTitleSize  = 20
)

// sidebarHeader is ChatScreen.tsx's header: title on the left, then new-chat,
// menu and the self avatar pushed to the right edge.
type sidebarHeader struct {
	widget.BaseWidget
	onNewChat func()
	onMenu    func()

	title  *canvas.Text
	avatar *canvas.Image
	avBg   *canvas.Circle
	avIcon fyne.CanvasObject
	newBtn *iconButton
	menuBt *iconButton
}

func newSidebarHeader(onNewChat, onMenu func()) *sidebarHeader {
	h := &sidebarHeader{onNewChat: onNewChat, onMenu: onMenu}
	h.ExtendBaseWidget(h)
	return h
}

// SetSelfAvatar swaps the placeholder glyph for the downloaded profile picture.
// An empty resource restores the placeholder.
func (h *sidebarHeader) SetSelfAvatar(res fyne.Resource) {
	if h.avatar == nil {
		return
	}
	if res == nil {
		h.avatar.Resource = nil
		h.avatar.Hide()
		if h.avIcon != nil {
			h.avIcon.Show()
		}
	} else {
		h.avatar.Resource = res
		h.avatar.Show()
		if h.avIcon != nil {
			h.avIcon.Hide()
		}
	}
	h.avatar.Refresh()
}

func (h *sidebarHeader) CreateRenderer() fyne.WidgetRenderer {
	h.title = canvas.NewText("WhatsApp", waWhite)
	h.title.TextSize = headerTitleSize
	h.title.TextStyle = fyne.TextStyle{Bold: true}

	h.newBtn = newIconButton("hdr-new", iconNewChat, headerIconSize, waMuted, func() {
		if h.onNewChat != nil {
			h.onNewChat()
		}
	})
	h.menuBt = newIconButton("hdr-menu", iconMore, headerIconSize, waMuted, func() {
		if h.onMenu != nil {
			h.onMenu()
		}
	})

	// bg-gray-600 dark:bg-gray-600 (#4a5565) circle with the UserAvatar glyph.
	h.avBg = canvas.NewCircle(color.NRGBA{R: 0x4a, G: 0x55, B: 0x65, A: 255})
	h.avIcon = iconImage("hdr-self", iconPerson, headerAvatarSize, waWhite)
	h.avatar = canvas.NewImageFromResource(nil)
	h.avatar.FillMode = canvas.ImageFillCover
	h.avatar.Hide()
	avWrap := container.NewMax(h.avBg, container.NewCenter(h.avIcon), h.avatar)

	right := container.New(&headerActionsLayout{},
		wrapFixed(h.newBtn, headerBtnSize, headerBtnSize),
		wrapFixed(h.menuBt, headerBtnSize, headerBtnSize),
		container.New(&padLayout{h: headerAvatarML}, wrapFixed(avWrap, headerAvatarSize, headerAvatarSize)),
	)

	body := container.New(&headerBodyLayout{}, h.title, right)

	// CSS `h-16` is 64px including the 1px border-bottom, so the content box is
	// 63px. container.NewBorder added the divider on top of a 64px child and
	// produced 69px, hence the explicit layout.
	divider := canvas.NewRectangle(waHeaderDivider)

	return widget.NewSimpleRenderer(container.New(&headerLayout{},
		wrapFixed(body, 0, headerHeight-1), divider))
}

// headerBodyLayout is the `flex items-center justify-between px-4` row: title
// left, actions right, both vertically centred.
//
// container.NewBorder cannot be used here: it reserves the right object's width
// out of the centre and the nested padLayout ended up collapsing the row to
// zero width, which pushed the icons off-canvas.
type headerBodyLayout struct{}

func (l *headerBodyLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) < 2 {
		return fyne.Size{}
	}
	t, r := objects[0].MinSize(), objects[1].MinSize()
	return fyne.NewSize(t.Width+r.Width+headerPadX*2, maxF(t.Height, r.Height))
}

func (l *headerBodyLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 2 {
		return
	}
	title, actions := objects[0], objects[1]
	th := title.MinSize().Height
	title.Move(fyne.NewPos(headerPadX, (size.Height-th)/2))
	title.Resize(fyne.NewSize(size.Width-headerPadX*2-actions.MinSize().Width, th))

	ah := actions.MinSize().Height
	aw := actions.MinSize().Width
	actions.Move(fyne.NewPos(size.Width-headerPadX-aw, (size.Height-ah)/2))
	actions.Resize(fyne.NewSize(aw, ah))
}

// headerLayout stacks the 63px content box above a 1px border.
type headerLayout struct{}

func (l *headerLayout) MinSize([]fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(0, headerHeight)
}

func (l *headerLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) < 2 {
		return
	}
	objects[0].Move(fyne.NewPos(0, 0))
	objects[0].Resize(fyne.NewSize(size.Width, headerHeight-1))
	objects[1].Move(fyne.NewPos(0, headerHeight-1))
	objects[1].Resize(fyne.NewSize(size.Width, 1))
}

// headerActionsLayout lays the two 40px buttons out with a 4px gap and the
// avatar 8px after them — React's `gap-1` plus the avatar's `ml-2`.
type headerActionsLayout struct{}

func (l *headerActionsLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	w := float32(0)
	for i, o := range objects {
		if i > 0 {
			if i == len(objects)-1 {
				w += headerAvatarML
			} else {
				w += headerBtnGap
			}
		}
		w += o.MinSize().Width
	}
	return fyne.NewSize(w, headerBtnSize)
}

func (l *headerActionsLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	x := float32(0)
	for i, o := range objects {
		if i > 0 {
			if i == len(objects)-1 {
				x += headerAvatarML
			} else {
				x += headerBtnGap
			}
		}
		s := o.MinSize()
		o.Move(fyne.NewPos(x, 0))
		o.Resize(s)
		x += s.Width
	}
}

// wrapFixed pins an object to an exact size; Fyne's GridWrap would otherwise
// let the surrounding layout stretch it. A zero w or h means "let it stretch on
// that axis" — passing 0 for the width collapsed the header row to 0px.
func wrapFixed(obj fyne.CanvasObject, w, h float32) fyne.CanvasObject {
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
