package fyneui

import (
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// WhatsApp icon set, traced from frontend/src/assets/svgs/*.tsx so the Fyne
// build uses the same glyphs as the React app instead of emoji stand-ins.
// Every icon is a 24x24 (or 48x48) path filled with a caller-supplied color.

const iconChats = `<path d="M12 2C6.5 2 2 6.03 2 11c0 2.42 1.09 4.62 2.87 6.25L4 22l4.35-1.4C9.5 20.86 10.72 21 12 21c5.5 0 10-4.03 10-9s-4.5-10-10-10z"/>`

const iconStatus = `<path d="M12 2a10 10 0 1 0 0 20 10 10 0 0 0 0-20zm0 2a8 8 0 0 1 7.75 6H14a2 2 0 0 0-1.9 1.37L9.5 18.9A8 8 0 0 1 12 4zm0 16a8 8 0 0 1-4.9-1.68l2.6-5.5A2 2 0 0 0 11.6 12H18a8 8 0 0 1-6 8z"/>`

const iconChannels = `<path d="M3 10v4a1 1 0 0 0 1 1h2.5l5.2 4.3A1 1 0 0 0 13 18.5v-13a1 1 0 0 0-1.3-.8L5.5 9H4a1 1 0 0 0-1 1zm13.5 1a4.5 4.5 0 0 1-1.4 3.3l1.4 1.4A6.5 6.5 0 0 0 18.5 11a6.5 6.5 0 0 0-2-4.7l-1.4 1.4A4.5 4.5 0 0 1 16.5 11zm3.5 0a8 8 0 0 1-2.4 5.7l1.4 1.4A10 10 0 0 0 22 11a10 10 0 0 0-2.9-7.1L17.7 5.3A8 8 0 0 1 20 11z"/>`

const iconCommunities = `<path d="M12 2 2 7l10 5 10-5-10-5zM4 10.2V15l8 4 8-4v-4.8l-8 4-8-4z"/>`

const iconSettings = `<path d="M19.4 13a7.7 7.7 0 0 0 0-2l2-1.6a.5.5 0 0 0 .1-.6l-1.9-3.3a.5.5 0 0 0-.6-.2l-2.4 1a7.4 7.4 0 0 0-1.7-1l-.4-2.5a.5.5 0 0 0-.5-.4h-3.8a.5.5 0 0 0-.5.4L9 5.3a7.4 7.4 0 0 0-1.7 1l-2.4-1a.5.5 0 0 0-.6.2L2.4 8.8a.5.5 0 0 0 .1.6l2 1.6a7.7 7.7 0 0 0 0 2l-2 1.6a.5.5 0 0 0-.1.6l1.9 3.3a.5.5 0 0 0 .6.2l2.4-1c.5.4 1.1.8 1.7 1l.4 2.5a.5.5 0 0 0 .5.4h3.8a.5.5 0 0 0 .5-.4l.4-2.5c.6-.2 1.2-.6 1.7-1l2.4 1a.5.5 0 0 0 .6-.2l1.9-3.3a.5.5 0 0 0-.1-.6l-2-1.6zM12 15.5A3.5 3.5 0 1 1 12 8.5a3.5 3.5 0 0 1 0 7z"/>`

const iconSearch = `<path d="M15.5 14h-.8l-.3-.3a6.5 6.5 0 1 0-.7.7l.3.3v.8l5 5 1.5-1.5-5-5zm-6 0a4.5 4.5 0 1 1 0-9 4.5 4.5 0 0 1 0 9z"/>`

// iconNewChat is NewChatIcon from frontend/src/assets/svgs/chat_icons.tsx — a
// chat bubble with two lines, not a plus sign.
const iconNewChat = `<path d="M19.005 3.175H4.674C3.642 3.175 3 3.789 3 4.821V21.02l3.544-3.514h12.461c1.033 0 2.064-1.06 2.064-2.093V4.821c-.001-1.032-1.032-1.646-2.064-1.646zm-4.989 9.869H6.666V11.5h7.35v1.544zm3.35-4.135H6.666V7.36h10.7v1.55z"/>`

const iconMore = `<path d="M12 8a2 2 0 1 0 0-4 2 2 0 0 0 0 4zm0 6a2 2 0 1 0 0-4 2 2 0 0 0 0 4zm0 6a2 2 0 1 0 0-4 2 2 0 0 0 0 4z"/>`

const iconBack = `<path d="M20 11H7.8l5.6-5.6L12 4l-8 8 8 8 1.4-1.4L7.8 13H20v-2z"/>`

const iconEmoji = `<path d="M12 2a10 10 0 1 0 0 20 10 10 0 0 0 0-20zm-3.5 7a1.5 1.5 0 1 1 0 3 1.5 1.5 0 0 1 0-3zm7 0a1.5 1.5 0 1 1 0 3 1.5 1.5 0 0 1 0-3zM12 18a5.5 5.5 0 0 1-4.7-2.6l1.5-1A3.8 3.8 0 0 0 12 16a3.8 3.8 0 0 0 3.2-1.6l1.5 1A5.5 5.5 0 0 1 12 18z"/>`

const iconAttach = `<path d="M16.5 6.5v8a4.5 4.5 0 1 1-9 0V7a3 3 0 0 1 6 0v7.5a1.5 1.5 0 0 1-3 0V7h-1.6v7.5a3 3 0 0 0 6 0V7a4.6 4.6 0 0 0-9.2 0v7.5a6.1 6.1 0 0 0 12.2 0v-8h-1.4z"/>`

const iconSend = `<path d="M2 21 23 12 2 3v7l15 2-15 2v7z"/>`

const iconPin = `<path d="M16 3a1 1 0 0 1 .95 1.31l-.9 2.72 3.42 3.42a1 1 0 0 1-.21 1.57l-3.62 2.07-1.9 4.75a1 1 0 0 1-1.64.33L9 16.07l-4.29 4.3-1.42-1.42 4.3-4.29-3.1-3.1a1 1 0 0 1 .33-1.64l4.75-1.9 2.07-3.62A1 1 0 0 1 12.5 4z"/>`

// iconArchive is the Material "archive" glyph from ChatScreen.tsx:865.
const iconArchive = `<path d="M20.54 5.23 19.15 3.55A1.5 1.5 0 0 0 18 3H6a1.5 1.5 0 0 0-1.16.55L3.46 5.23A2 2 0 0 0 3 6.5V19a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V6.5a2 2 0 0 0-.46-1.27ZM6.24 5h11.52l.81.97H5.44ZM5 19V8h14v11Zm8-5.5V11h-2v2.5H8.5L12 17l3.5-3.5Z"/>`

// iconChatBubble is EmptyStateIcon from chat_icons.tsx (120px in React).
const iconChatBubble = `<path d="M20 2H4c-1.1 0-2 .9-2 2v18l4-4h14c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2z"/>`

const iconMuted = `<path d="M20 18.7 7.8 6.1 5.3 3.5 4 4.8l2.8 2.8c-.5 1-.8 2.2-.8 3.4v5l-2 2v1h13.7l2 2L21 19.7l-1-1zM12 22a2 2 0 0 0 2-2h-4a2 2 0 0 0 2 2zm6-7.3V11c0-3.1-1.6-5.6-4.5-6.3V4a1.5 1.5 0 0 0-3 0v.7c-.2 0-.3.1-.4.1L18 14.7z"/>`

const iconPhoto = `<path d="M21 5H3a1 1 0 0 0-1 1v12a1 1 0 0 0 1 1h18a1 1 0 0 0 1-1V6a1 1 0 0 0-1-1zm-1 12H4l4.5-6 3 3.8L14 11l6 6z"/>`

const iconVideo = `<path d="M17 10.5V7a1 1 0 0 0-1-1H4a1 1 0 0 0-1 1v10a1 1 0 0 0 1 1h12a1 1 0 0 0 1-1v-3.5l4 3.5V7l-4 3.5z"/>`

const iconPlay = `<path d="M8 5v14l11-7z"/>`

const iconDocument = `<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z"/>`

const iconPerson = `<path d="M24 23q-1.857 0-3.178-1.322Q19.5 20.357 19.5 18.5t1.322-3.178T24 14t3.178 1.322Q28.5 16.643 28.5 18.5t-1.322 3.178T24 23m-6.75 10q-.928 0-1.59-.66-.66-.662-.66-1.59v-.9q0-.956.492-1.758A3.3 3.3 0 0 1 16.8 26.87a16.7 16.7 0 0 1 3.544-1.308q1.8-.435 3.656-.436 1.856 0 3.656.436T31.2 26.87q.816.422 1.308 1.223T33 29.85v.9q0 .928-.66 1.59-.662.66-1.59.66z"/>`

const iconPeople = `<path d="M16 11c1.66 0 2.99-1.34 2.99-3S17.66 5 16 5c-1.66 0-3 1.34-3 3s1.34 3 3 3zm-8 0c1.66 0 2.99-1.34 2.99-3S9.66 5 8 5C6.34 5 5 6.34 5 8s1.34 3 3 3zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-2.5c0-2.33-4.67-3.5-7-3.5zm8 0c-.29 0-.62.02-.97.05 1.16.84 1.97 1.97 1.97 3.45V19h6v-2.5c0-2.33-4.67-3.5-7-3.5z"/>`

// svgResource wraps a path into a standalone SVG resource rendered at `size`
// with a flat fill so Fyne does not need currentColor support.
func svgResource(name, path string, size float32, c color.Color) fyne.Resource {
	return svgResourceVB(name, path, size, 24, c)
}

// svgResourceVB is svgResource for paths authored in a non-24 unit viewBox.
// React's contact avatar silhouette uses viewBox="0 0 48 48", so hardcoding a
// /24 scale draws it at double size and clips it out of the frame entirely.
func svgResourceVB(name, path string, size, viewBox float32, c color.Color) fyne.Resource {
	r, g, b, a := c.RGBA()
	hex := fmt.Sprintf("#%02x%02x%02x", r>>8, g>>8, b>>8)
	op := float64(a) / 65535.0
	if op <= 0 {
		op = 1
	}
	svg := strings.Join([]string{
		fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %g %g" width="%g" height="%g">`, viewBox, viewBox, size, size),
		fmt.Sprintf(`<g fill="%s" fill-opacity="%.3f">`, hex, op),
		path,
		`</g></svg>`,
	}, "")
	return fyne.NewStaticResource(name, []byte(svg))
}

// iconImageVB renders a path authored in a viewBox-unit space at size px.
func iconImageVB(name, path string, size, viewBox float32, c color.Color) fyne.CanvasObject {
	img := canvas.NewImageFromResource(svgResourceVB(name, path, size, viewBox, c))
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(size, size))
	return img
}

// iconImage renders a path at size px tinted with c.
func iconImage(name, path string, size float32, c color.Color) fyne.CanvasObject {
	img := canvas.NewImageFromResource(svgResource(name, path, size, c))
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(size, size))
	return img
}

// iconButton is a flat, tappable icon with a circular hover target — the shape
// WhatsApp uses for header/composer icon actions.
type iconButton struct {
	widget.BaseWidget
	img    *canvas.Image
	path   string
	size   float32
	onTap  func()
	hover  *canvas.Rectangle
	bgCol  color.Color
	hoverC color.Color
}

func newIconButton(name, path string, size float32, c color.Color, onTap func()) *iconButton {
	b := &iconButton{path: path, size: size, onTap: onTap, bgCol: color.Transparent, hoverC: waHover}
	b.ExtendBaseWidget(b)
	b.hover = canvas.NewRectangle(color.Transparent)
	b.hover.CornerRadius = 999
	b.img = canvas.NewImageFromResource(svgResource(name, path, size, c))
	b.img.FillMode = canvas.ImageFillContain
	b.img.SetMinSize(fyne.NewSize(size, size))
	return b
}

func (b *iconButton) CreateRenderer() fyne.WidgetRenderer {
	pad := (b.size * 1.6)
	inner := container.NewCenter(b.img)
	body := container.NewMax(b.hover, container.NewGridWrap(fyne.NewSize(pad, pad), inner))
	return widget.NewSimpleRenderer(body)
}

func (b *iconButton) Tapped(*fyne.PointEvent) {
	if b.onTap != nil {
		b.onTap()
	}
}

func (b *iconButton) MouseIn(*desktop.MouseEvent) {
	b.hover.FillColor = b.hoverC
	b.hover.Refresh()
}

func (b *iconButton) MouseOut() {
	b.hover.FillColor = b.bgCol
	b.hover.Refresh()
}

func (b *iconButton) Cursor() desktop.Cursor { return desktop.PointerCursor }
