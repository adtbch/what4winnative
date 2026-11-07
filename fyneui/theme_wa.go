package fyneui

import "image/color"

// WhatsApp desktop dark palette (Windows/React reference-aligned)
var (
	waBg       = color.NRGBA{R: 0x0b, G: 0x14, B: 0x1a, A: 255} // #0b141a app/chat bg
	waPanel    = color.NRGBA{R: 0x11, G: 0x1b, B: 0x21, A: 255} // #111b21 sidebar/panels
	waPanel2   = color.NRGBA{R: 0x20, G: 0x2c, B: 0x33, A: 255} // #202c33 search/input/header cards
	waHover    = color.NRGBA{R: 0x2a, G: 0x39, B: 0x42, A: 255} // #2a3942 hover/selected
	waBorder   = color.NRGBA{R: 0x22, G: 0x2d, B: 0x34, A: 255} // #222d34 borders
	waDivider  = color.NRGBA{R: 0x1f, G: 0x2c, B: 0x34, A: 255} // list/input divider
	waAccent   = color.NRGBA{R: 0x00, G: 0xa8, B: 0x84, A: 255} // #00a884 teal accent
	waGreen    = color.NRGBA{R: 0x25, G: 0xd3, B: 0x66, A: 255} // #25d366 badge/WA green
	waSent     = color.NRGBA{R: 0x00, G: 0x5c, B: 0x4b, A: 255} // #005c4b outgoing bubble
	waRecv     = color.NRGBA{R: 0x20, G: 0x2c, B: 0x33, A: 255} // #202c33 incoming bubble
	waText     = color.NRGBA{R: 0xe9, G: 0xed, B: 0xef, A: 255} // #e9edef primary text
	waMuted    = color.NRGBA{R: 0x86, G: 0x96, B: 0xa0, A: 255} // #8696a0 secondary text
	waTime     = color.NRGBA{R: 0x86, G: 0x96, B: 0xa0, A: 255} // timestamp
	waWhite    = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	waPill     = color.NRGBA{R: 0x20, G: 0x2c, B: 0x33, A: 255} // filter pill
	waPillOn   = color.NRGBA{R: 0x0a, G: 0x33, B: 0x2c, A: 255} // selected pill
	waRail     = color.NRGBA{R: 0x20, G: 0x2c, B: 0x33, A: 255} // left rail
	waRailIcon = color.NRGBA{R: 0xae, G: 0xba, B: 0xc0, A: 255} // rail icon
	waRailSel  = color.NRGBA{R: 0x37, G: 0x44, B: 0x4c, A: 255} // rail selected pill
	waDoodle   = color.NRGBA{R: 255, G: 255, B: 255, A: 10}     // subtle pattern
)
