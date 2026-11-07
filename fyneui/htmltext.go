package fyneui

import (
	"encoding/base64"
	"html"
	"regexp"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Messages in this app are stored as HTML (internal/markdown renders WhatsApp
// markdown into <p>/<b>/<i>/<a>/<ul> etc). Fyne has no HTML widget, so we
// convert that HTML into styled text runs and wrap them ourselves — the same
// visual result the React frontend gets from dangerouslySetInnerHTML.

type htmlRun struct {
	text  string
	bold  bool
	ital  bool
	strike bool
	code  bool
	link  bool
}

var (
	tagRE     = regexp.MustCompile(`<[^>]*>`)
	brkRE     = regexp.MustCompile(`(?i)<br\s*/?>`)
	blockEndRE = regexp.MustCompile(`(?i)</(p|div|li|blockquote|h[1-6])>`)
)

// htmlToText flattens message HTML to the plain text WhatsApp shows in chat
// list previews (no tags, entities decoded, block boundaries become newlines).
func htmlToText(s string) string {
	if s == "" {
		return ""
	}
	s = brkRE.ReplaceAllString(s, "\n")
	s = blockEndRE.ReplaceAllString(s, "\n")
	s = tagRE.ReplaceAllString(s, "")
	s = html.UnescapeString(s)
	s = strings.ReplaceAll(s, "\u00a0", " ")
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		out = append(out, l)
	}
	return strings.Join(out, "\n")
}

// htmlPreview flattens to a single line for the chat list.
func htmlPreview(s string) string {
	t := htmlToText(s)
	t = strings.ReplaceAll(t, "\n", " ")
	return strings.TrimSpace(t)
}

func parseHTMLRuns(s string) []htmlRun {
	var runs []htmlRun
	var cur strings.Builder
	state := htmlRun{}
	flush := func() {
		if cur.Len() == 0 {
			return
		}
		r := state
		r.text = html.UnescapeString(cur.String())
		runs = append(runs, r)
		cur.Reset()
	}
	i := 0
	for i < len(s) {
		lt := strings.IndexByte(s[i:], '<')
		if lt < 0 {
			cur.WriteString(s[i:])
			break
		}
		cur.WriteString(s[i : i+lt])
		gt := strings.IndexByte(s[i+lt:], '>')
		if gt < 0 {
			cur.WriteString(s[i+lt:])
			break
		}
		tag := s[i+lt : i+lt+gt+1]
		low := strings.ToLower(tag)
		flush()
		switch {
		case strings.HasPrefix(low, "<br"):
			runs = append(runs, htmlRun{text: "\n"})
		case strings.HasPrefix(low, "</p"), strings.HasPrefix(low, "</div"),
			strings.HasPrefix(low, "</li"), strings.HasPrefix(low, "</blockquote"):
			runs = append(runs, htmlRun{text: "\n"})
		case strings.HasPrefix(low, "<li"):
			runs = append(runs, htmlRun{text: "• "})
		case strings.HasPrefix(low, "<b"), strings.HasPrefix(low, "<strong"):
			state.bold = true
		case strings.HasPrefix(low, "</b"), strings.HasPrefix(low, "</strong"):
			state.bold = false
		case strings.HasPrefix(low, "<i"), strings.HasPrefix(low, "<em"):
			state.ital = true
		case strings.HasPrefix(low, "</i"), strings.HasPrefix(low, "</em"):
			state.ital = false
		case strings.HasPrefix(low, "<s>"), strings.HasPrefix(low, "<strike"), strings.HasPrefix(low, "<del"):
			state.strike = true
		case strings.HasPrefix(low, "</s>"), strings.HasPrefix(low, "</strike"), strings.HasPrefix(low, "</del"):
			state.strike = false
		case strings.HasPrefix(low, "<span"):
			state.code = true
		case strings.HasPrefix(low, "</span"):
			state.code = false
		case strings.HasPrefix(low, "<a "):
			state.link = true
		case strings.HasPrefix(low, "</a"):
			state.link = false
		}
		i += lt + gt + 1
	}
	flush()
	return runs
}

func runStyle(r htmlRun) widget.RichTextStyle {
	st := widget.RichTextStyle{
		TextStyle: fyne.TextStyle{Bold: r.bold, Italic: r.ital},
		ColorName: theme.ColorNameForeground,
	}
	switch {
	case r.link:
		st.ColorName = theme.ColorNameHyperlink
	case r.code:
		st.ColorName = theme.ColorNameForeground
	}
	return st
}

func runWidth(r htmlRun, size float32) float32 {
	return fyne.MeasureText(r.text, size, fyne.TextStyle{Bold: r.bold, Italic: r.ital}).Width
}

// wrapRuns breaks runs into lines no wider than maxW, preserving styles.
func wrapRuns(runs []htmlRun, maxW, size float32) [][]htmlRun {
	var lines [][]htmlRun
	var line []htmlRun
	var lineW float32

	pushLine := func() {
		lines = append(lines, line)
		line = nil
		lineW = 0
	}

	for _, r := range runs {
		if r.text == "\n" {
			pushLine()
			continue
		}
		words := strings.SplitAfter(r.text, " ")
		for _, w := range words {
			if w == "" {
				continue
			}
			ww := runWidth(htmlRun{text: strings.TrimRight(w, " "), bold: r.bold, ital: r.ital}, size)
			space := float32(0)
			if lineW > 0 {
				space = fyne.MeasureText(" ", size, fyne.TextStyle{}).Width
			}
			if lineW > 0 && lineW+space+ww > maxW {
				pushLine()
				w = strings.TrimLeft(w, " ")
				if w == "" {
					continue
				}
				ww = runWidth(htmlRun{text: strings.TrimRight(w, " "), bold: r.bold, ital: r.ital}, size)
			}
			nr := r
			nr.text = w
			// w already carries its trailing space (SplitAfter); injecting another
			// one here doubled every inter-word gap.
			line = append(line, nr)
			lineW += ww
		}
	}
	if len(line) > 0 || len(lines) == 0 {
		lines = append(lines, line)
	}
	return lines
}

// htmlRichText builds a non-wrapping RichText already wrapped at maxW so the
// bubble can be sized exactly (Fyne cannot constrain a wrapped RichText width).
func htmlRichText(src string, maxW, size float32) (*widget.RichText, float32, float32) {
	runs := parseHTMLRuns(src)
	lines := wrapRuns(runs, maxW, size)
	rt := widget.NewRichText()
	rt.Wrapping = fyne.TextWrapOff
	var segs []widget.RichTextSegment
	var widest float32
	lineH := fyne.MeasureText("Ag", size, fyne.TextStyle{}).Height
	curW := float32(0)
	flushLine := func() {
		if curW > widest {
			widest = curW
		}
		curW = 0
	}
	for li, line := range lines {
		if li > 0 {
			segs = append(segs, &widget.TextSegment{Text: "\n", Style: widget.RichTextStyle{TextStyle: fyne.TextStyle{}}})
			flushLine()
		}
		for _, r := range line {
			segs = append(segs, &widget.TextSegment{Text: r.text, Style: runStyle(r)})
			curW += runWidth(r, size)
		}
	}
	flushLine()
	rt.Segments = segs
	rt.Refresh()
	h := float32(len(lines)) * lineH
	if h <= 0 {
		h = lineH
	}
	return rt, widest, h
}

// htmlTextHeight reports how tall the wrapped HTML will be at maxW.
func htmlTextSize(src string, maxW, size float32) (float32, float32) {
	runs := parseHTMLRuns(src)
	lines := wrapRuns(runs, maxW, size)
	var widest float32
	for _, line := range lines {
		var w float32
		for _, r := range line {
			w += runWidth(r, size)
		}
		if w > widest {
			widest = w
		}
	}
	lineH := fyne.MeasureText("Ag", size, fyne.TextStyle{}).Height
	return widest, float32(len(lines)) * lineH
}

// htmlTextBlock builds the bubble body: a non-wrapping RichText pre-wrapped at
// maxW plus the measured width/height so the caller can size the container.
func htmlTextBlock(src string, maxW, size float32) (fyne.CanvasObject, float32, float32) {
	rt, w, h := htmlRichText(src, maxW, size)
	return rt, w, h
}

// dataURLResource decodes a "data:<mime>;base64,<payload>" string into a Fyne
// resource usable by canvas.Image.
func dataURLResource(name, dataURL string) fyne.Resource {
	i := strings.IndexByte(dataURL, ',')
	if i < 0 || i+1 >= len(dataURL) {
		return nil
	}
	raw, err := base64.StdEncoding.DecodeString(dataURL[i+1:])
	if err != nil || len(raw) == 0 {
		return nil
	}
	return fyne.NewStaticResource(name, raw)
}

// mediaBox fits a media thumbnail into WhatsApp's chat bounds (max 300x400,
// min width 300 when the source is wide enough) and returns the display size.
func mediaBox(w, h int) fyne.Size {
	const maxW, maxH, minW = 300.0, 400.0, 240.0
	if w <= 0 || h <= 0 {
		return fyne.NewSize(minW, 180)
	}
	scale := maxW / float64(w)
	if s := maxH / float64(h); s < scale {
		scale = s
	}
	if scale > 1 {
		scale = 1
	}
	dw := float64(w) * scale
	dh := float64(h) * scale
	if dw < minW {
		dh = dh * minW / dw
		dw = minW
		if dh > maxH {
			dh = maxH
		}
	}
	return fyne.NewSize(float32(dw), float32(dh))
}

var _ = canvas.NewRectangle
