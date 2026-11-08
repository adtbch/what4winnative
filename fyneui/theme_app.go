package fyneui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// waTheme forces Fyne's widget palette onto the WhatsApp dark scheme. Without
// it RichText/Entry resolve foreground against the default (light) theme and
// dark text lands on the dark bubbles.
type waTheme struct {
	base fyne.Theme
}

func newWATheme() fyne.Theme {
	return &waTheme{base: theme.DefaultTheme()}
}

func (t *waTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameForeground, theme.ColorNamePrimary:
		return waText
	case theme.ColorNameHyperlink:
		return waLink
	case theme.ColorNameBackground:
		return waPanel
	case theme.ColorNameOverlayBackground, theme.ColorNameMenuBackground:
		return waPanel
	case theme.ColorNameInputBackground:
		return waPanel2
	case theme.ColorNameInputBorder:
		// WhatsApp has no visible outline around the composer field; Fyne draws
		// a 1px rect here, so match the fill exactly to make it disappear.
		return waPanel2
	case theme.ColorNamePlaceHolder, theme.ColorNameDisabled:
		return waMuted
	case theme.ColorNameButton:
		return waPanel2
	case theme.ColorNameSeparator:
		return waDivider
	case theme.ColorNameHover, theme.ColorNameFocus, theme.ColorNameSelection:
		return waHover
	case theme.ColorNamePressed:
		return waRailSel
	case theme.ColorNameForegroundOnPrimary:
		return waWhite
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 0x37, G: 0x44, B: 0x4c, A: 200}
	}
	return t.base.Color(name, theme.VariantDark)
}

func (t *waTheme) Font(style fyne.TextStyle) fyne.Resource { return t.base.Font(style) }

func (t *waTheme) Icon(name fyne.ThemeIconName) fyne.Resource { return t.base.Icon(name) }

func (t *waTheme) Size(name fyne.ThemeSizeName) float32 { return t.base.Size(name) }

// borderlessEntryTheme makes widget.Entry blend into a custom pill/card.
type borderlessEntryTheme struct{ fyne.Theme; bg color.Color }

func (t *borderlessEntryTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	switch n {
	case theme.ColorNameInputBorder, theme.ColorNameInputBackground:
		return t.bg
	case theme.ColorNameFocus:
		return color.Transparent
	}
	return t.Theme.Color(n, v)
}
