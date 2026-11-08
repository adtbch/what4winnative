package theme

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// WaTheme forces Fyne's widget palette onto the WhatsApp dark scheme. Without
// it RichText/Entry resolve foreground against the default (light) theme and
// dark text lands on the dark bubbles.
type WaTheme struct {
	base fyne.Theme
}

func NewWATheme() fyne.Theme {
	return &WaTheme{base: theme.DefaultTheme()}
}

func (t *WaTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameForeground, theme.ColorNamePrimary:
		return WaText
	case theme.ColorNameHyperlink:
		return WaLink
	case theme.ColorNameBackground:
		return WaPanel
	case theme.ColorNameOverlayBackground, theme.ColorNameMenuBackground:
		return WaPanel
	case theme.ColorNameInputBackground:
		return WaPanel2
	case theme.ColorNameInputBorder:
		// WhatsApp has no visible outline around the composer field; Fyne draws
		// a 1px rect here, so match the fill exactly to make it disappear.
		return WaPanel2
	case theme.ColorNamePlaceHolder, theme.ColorNameDisabled:
		return WaMuted
	case theme.ColorNameButton:
		return WaPanel2
	case theme.ColorNameSeparator:
		return WaDivider
	case theme.ColorNameHover, theme.ColorNameFocus, theme.ColorNameSelection:
		return WaHover
	case theme.ColorNamePressed:
		return WaRailSel
	case theme.ColorNameForegroundOnPrimary:
		return WaWhite
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 0x37, G: 0x44, B: 0x4c, A: 200}
	}
	return t.base.Color(name, theme.VariantDark)
}

func (t *WaTheme) Font(style fyne.TextStyle) fyne.Resource { return t.base.Font(style) }

func (t *WaTheme) Icon(name fyne.ThemeIconName) fyne.Resource { return t.base.Icon(name) }

func (t *WaTheme) Size(name fyne.ThemeSizeName) float32 { return t.base.Size(name) }

// BorderlessEntryTheme makes widget.Entry blend into a custom pill/card.
type BorderlessEntryTheme struct {
	fyne.Theme
	bg color.Color
}

func (t *BorderlessEntryTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	switch n {
	case theme.ColorNameInputBorder, theme.ColorNameInputBackground:
		return t.bg
	case theme.ColorNameFocus:
		return color.Transparent
	}
	return t.Theme.Color(n, v)
}
