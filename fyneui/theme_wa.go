package fyneui

import "image/color"

var (
	// WhatsApp dark palette (approx)
	waBg        = color.NRGBA{R: 11, G: 20, B: 26, A: 255}   // #0b141a
	waPanel     = color.NRGBA{R: 17, G: 27, B: 33, A: 255}   // #111b21
	waPanel2    = color.NRGBA{R: 32, G: 44, B: 51, A: 255}   // #202c33
	waHover     = color.NRGBA{R: 42, G: 57, B: 66, A: 255}   // #2a3942
	waBorder    = color.NRGBA{R: 34, G: 46, B: 53, A: 255}   // #222e35
	waAccent    = color.NRGBA{R: 0, G: 168, B: 132, A: 255}   // #00a884
	waAccent2   = color.NRGBA{R: 37, G: 211, B: 102, A: 255}  // #25d366 (badge)
	waText      = color.NRGBA{R: 233, G: 237, B: 239, A: 255} // #e9edef
	waMuted     = color.NRGBA{R: 134, G: 150, B: 160, A: 255} // #8696a0
	waWhite     = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	waPill      = color.NRGBA{R: 32, G: 44, B: 51, A: 255}
	waPillOn    = color.NRGBA{R: 11, G: 72, B: 59, A: 255}   // selected filter bg approx
)

func tcolor(b bool, on, off color.NRGBA) color.NRGBA { if b { return on }; return off }
