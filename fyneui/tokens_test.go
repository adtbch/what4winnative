package fyneui

import (
	"bufio"
	"image/color"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// TestTokensMatchStyleCSS reads frontend/src/style.css @theme directly and
// asserts every Fyne token that claims a CSS source is still identical. Pinning
// the hex values in Go would only re-state the copy; parsing the CSS makes the
// test fail the moment either side drifts.
func TestTokensMatchStyleCSS(t *testing.T) {
	css := parseThemeVars(t, "../frontend/src/style.css")

	pairs := []struct {
		token string // CSS custom property
		got   color.NRGBA
	}{
		{"--color-dark-bg", waBg},
		{"--color-dark-tertiary", waBorder},
		{"--color-dark-text", waText},
		{"--color-dark-muted", waMuted},
		{"--color-green", waAccent},
		{"--color-sent-bubble-dark-bg", waSent},
		{"--color-received-bubble-dark-bg", waRecv},
		{"--color-toggle-dark-bg", waGreen},
	}
	for _, p := range pairs {
		want, ok := css[p.token]
		if !ok {
			t.Fatalf("%s not found in style.css @theme", p.token)
		}
		if p.got != want {
			t.Errorf("%s: fyne %s != css %s", p.token, hex(p.got), hex(want))
		}
	}

	// waTime and waMuted share a value in React: message timestamps use the same
	// muted token as secondary text.
	if waTime != waMuted {
		t.Errorf("waTime %s must equal waMuted %s", hex(waTime), hex(waMuted))
	}

	// .msg-link in style.css is the bubble hyperlink colour, NOT
	// --color-light-primary. Grep the literal to keep the two in sync.
	if !strings.Contains(readFile(t, "../frontend/src/style.css"), "#53bdeb") {
		t.Errorf("waLink %s has no #53bdeb source in style.css", hex(waLink))
	}
	if waLink != (color.NRGBA{R: 0x53, G: 0xbd, B: 0xeb, A: 255}) {
		t.Errorf("waLink drifted from #53bdeb: %s", hex(waLink))
	}
}

var themeVarRe = regexp.MustCompile(`^\s*(--color-[a-z0-9-]+)\s*:\s*(#[0-9a-fA-F]{3,8})\s*;`)

// parseThemeVars extracts hex custom properties from the @theme block.
func parseThemeVars(t *testing.T, path string) map[string]color.NRGBA {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()

	out := map[string]color.NRGBA{}
	inTheme := false
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if strings.Contains(line, "@theme") {
			inTheme = true
			continue
		}
		if !inTheme {
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(line), "}") {
			break
		}
		m := themeVarRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		c, ok := parseHex(m[2])
		if !ok {
			t.Fatalf("bad hex %q for %s", m[2], m[1])
		}
		out[m[1]] = c
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("scan %s: %v", path, err)
	}
	if len(out) == 0 {
		t.Fatalf("no @theme variables parsed from %s", path)
	}
	return out
}

func parseHex(s string) (color.NRGBA, bool) {
	s = strings.TrimPrefix(s, "#")
	if len(s) == 3 {
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2]})
	}
	// #rrggbbaa is valid CSS; keep the alpha instead of failing the parse.
	alpha := uint8(255)
	if len(s) == 8 {
		a, err := strconv.ParseUint(s[6:], 16, 8)
		if err != nil {
			return color.NRGBA{}, false
		}
		alpha = uint8(a)
		s = s[:6]
	}
	if len(s) != 6 {
		return color.NRGBA{}, false
	}
	v, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return color.NRGBA{}, false
	}
	return color.NRGBA{R: uint8(v >> 16), G: uint8(v >> 8), B: uint8(v), A: alpha}, true
}

func hex(c color.NRGBA) string {
	const digits = "0123456789abcdef"
	b := []byte{'#', 0, 0, 0, 0, 0, 0}
	for i, v := range []uint8{c.R, c.G, c.B} {
		b[1+i*2] = digits[v>>4]
		b[2+i*2] = digits[v&0xf]
	}
	return string(b)
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}
