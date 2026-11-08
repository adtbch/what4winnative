package theme

import "image/color"

// Palette copied 1:1 from frontend/src/style.css @theme + the dark: classes in
// ChatScreen.tsx / ChatInput.tsx, so the Fyne shell uses the exact same tokens
// as the React/Wails build instead of a generic WhatsApp palette.
var (
	WaBg            = color.NRGBA{R: 0x0c, G: 0x0c, B: 0x0c, A: 255} // --color-dark-bg
	WaPanel         = color.NRGBA{R: 0x0c, G: 0x0c, B: 0x0c, A: 255} // header/sidebar (bg-dark-bg)
	WaPanel2        = color.NRGBA{R: 0x24, G: 0x26, B: 0x26, A: 255} // #242626 search pill / composer / selected row
	WaHover         = color.NRGBA{R: 0x1a, G: 0x1a, B: 0x1a, A: 255} // dark:hover:bg-[#1a1a1a]
	WaBorder        = color.NRGBA{R: 0x2e, G: 0x2f, B: 0x2f, A: 255} // --color-dark-tertiary
	WaDivider       = color.NRGBA{R: 0x2e, G: 0x2f, B: 0x2f, A: 255} // dark:border-dark-tertiary
	WaAccent        = color.NRGBA{R: 0x1b, G: 0x9a, B: 0x58, A: 255} // --color-green (send button)
	WaGreen         = color.NRGBA{R: 0x21, G: 0xc0, B: 0x63, A: 255} // --color-toggle-dark-bg (badge, unread time)
	WaSent          = color.NRGBA{R: 0x14, G: 0x4d, B: 0x37, A: 255} // --color-sent-bubble-dark-bg
	WaRecv          = color.NRGBA{R: 0x24, G: 0x26, B: 0x26, A: 255} // --color-received-bubble-dark-bg
	WaText          = color.NRGBA{R: 0xe9, G: 0xed, B: 0xef, A: 255} // --color-dark-text
	WaAvatarIcon    = color.NRGBA{R: 0x11, G: 0x1b, B: 0x21, A: 255} // AVATAR_ICON_COLOR #111b21
	WaMuted         = color.NRGBA{R: 0x86, G: 0x96, B: 0xa0, A: 255} // --color-dark-muted
	WaTime          = color.NRGBA{R: 0x86, G: 0x96, B: 0xa0, A: 255}
	WaWhite         = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	WaLink          = color.NRGBA{R: 0x53, G: 0xbd, B: 0xeb, A: 255} // #53bdeb link/read tick
	WaTick          = color.NRGBA{R: 0x53, G: 0xbd, B: 0xeb, A: 255}
	WaPlayOverlay   = color.NRGBA{R: 0x0b, G: 0x14, B: 0x1a, A: 160}
	WaPill          = color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0}   // unselected: transparent + border
	WaPillBorder    = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 26}  // dark:border-white/10
	WaPillOn        = color.NRGBA{R: 0xe9, G: 0xed, B: 0xef, A: 255} // dark:bg-[#e9edef]
	WaPillOnText    = color.NRGBA{R: 0x0b, G: 0x14, B: 0x1a, A: 255} // dark:text-[#0b141a]
	WaRail          = color.NRGBA{R: 0x20, G: 0x2c, B: 0x33, A: 255}
	WaRailIcon      = color.NRGBA{R: 0xae, G: 0xba, B: 0xc0, A: 255}
	WaRailSel       = color.NRGBA{R: 0x37, G: 0x44, B: 0x4c, A: 255}
	WaDoodle        = color.NRGBA{R: 255, G: 255, B: 255, A: 23}     // 0.09 opacity doodle tile
	WaBadgeText     = color.NRGBA{R: 0x0a, G: 0x10, B: 0x14, A: 255} // badge text-[#0a1014]
	WaDayChip       = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 26}  // date separator dark:bg-white/10
	WaHeaderDivider = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 13}  // header border-b dark:border-white/5

	// aliases used by the shell builder
	WaSidebarBg = WaPanel
	WaChatBg    = color.NRGBA{R: 0x0a, G: 0x0a, B: 0x0a, A: 255} // .dark .chat-wallpaper
)
