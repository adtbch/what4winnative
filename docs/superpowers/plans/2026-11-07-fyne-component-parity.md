# Fyne UI Component-by-Component Parity Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking. One task per session; do not start Task N+1 before Task N's golden PNG has been visually inspected.

**Goal:** Rebuild the Fyne UI of `whats4linux` one component at a time until it matches the React/Wails frontend, where every component is independently renderable and visually verifiable without launching the app.

**Architecture:** `fyneui/` is split by component, not by layer. Each component is a plain constructor that takes data as parameters (never `*api.Api`), returns a `fyne.CanvasObject`, and is covered by a golden-render test that writes a PNG to `fyneui/testdata/out/`. `*api.Api` is only touched by the thin container files (`sidebar.go`, `chatview.go`, `app.go`) that compose components and feed them backend data. This keeps every visual decision testable in isolation and lets a human (or an agent with vision) diff one component at a time against the React source.

**Tech Stack:** Go 1.26.5, Fyne v2.8.2, `fyne.io/fyne/v2/test` for headless capture, existing `api/` + `internal/` packages unchanged.

**Spec:** `frontend/src/` is the spec. Every task names the exact React file it mirrors; that file wins over any assumption in this plan.

## Global Constraints

- Branch `fyne-migration` only. Commit after every task. Never force-push.
- `api/`, `internal/store/`, `internal/wa/`, `internal/voip/` must not change. The only allowed backend edit is `internal/eventbus` / `api/api.go` `emitEvent` (already done).
- All colors come from `fyneui/theme_wa.go`. No literal `color.NRGBA{...}` in component files except `color.Transparent`.
- Every mutation of a widget from a goroutine goes through `fyne.Do`.
- Component constructors must not call `*api.Api`. Data in via struct/params; events out via callbacks.
- Each component file ≤ 300 lines. If it grows past that, split it.
- Measurements are copied from Tailwind classes in the React source, not eyeballed:
  - `h-16` header = 64px, `w-12 h-12` sidebar avatar = 48px, `w-10 h-10` chat header avatar = 40px
  - `rounded-xl` bubble = 12px, `rounded-tl-sm`/`rounded-tr-sm` first-in-group corner = 2px
  - `rounded-full` composer pill = height/2, `h-11 w-11` send button = 44px
  - bubble padding `px-2 pt-1 pb-1.5` = left/right 8, top 4, bottom 6
  - `max-w-[85%] lg:max-w-[65%]` bubble width cap
- Never claim a component is done without the PNG from its test existing and being looked at.

## Review Focus

Failure modes the React source implies but no single component test would catch. Each one gets a test in the task that owns the code.

1. A long unbroken URL must wrap inside the bubble instead of pushing the bubble past `max-w-[65%]` (Task 10).
2. An emoji-only message renders at 32px (`text-[32px] leading-10`), not 14px (Task 10).
3. A message whose `Content` is nil must render "Empty Message" in italic, not an empty colored rectangle (Task 10).
4. The sidebar must stay usable at the 320px minimum and the 600px maximum width, including long chat names that must truncate, not overlap the timestamp (Task 7).
5. Group chats must draw the sender name only on the first message of a run and sharpen only that bubble's corner (Task 12).

---

## Component Checklist (crosscheck table)

Tick a row only when the PNG exists, has been looked at, and matches the named React source. This table is the single progress record for the migration.

| Task | Component | React source (spec) | Go file | Golden PNG | Status |
|---|-----------|--------------------|---------|-----------|--------|
| 1 | Baseline build + tokens | `style.css` `@theme` | `theme_wa.go` | — (asserted by test) | ☐ |
| 2 | Render harness + helpers | — | `golden_test.go`, `layout.go` | `harness.png` | ☐ |
| 3 | Token parity test | `style.css` | `tokens_test.go` | — | ☐ |
| 4 | Sidebar header | `ChatScreen.tsx:50-70` | `sidebar_header.go` | `sidebar_header.png` | ☐ |
| 5 | View tabs | `ChatScreen.tsx:828-849` | `view_tabs.go` | `view_tabs.png` | ☐ |
| 6 | Search bar | `ChatScreen.tsx:84-95` | `search_bar.go` | `search_bar.png` | ☐ |
| 7 | Chat row | `ChatScreen.tsx:247-320` | `chat_row.go` | `chat_row.png`, `chat_row_long.png` | ☐ |
| 8 | Archived + empty | `ChatScreen.tsx:362-375, 861-874` | `sidebar_states.go` | `archived_entry.png`, `sidebar_empty.png` | ☐ |
| 9 | Chat header | `ChatHeader.tsx` | `chat_header.go` | `chat_header.png`, `chat_header_group.png` | ☐ |
| 10 | Text bubble | `MessageItem.tsx:406-560` | `bubble.go` | `bubble_text.png`, `bubble_empty.png`, `bubble_long_url.png` | ☐ |
| 11 | Media bubble | `MediaContent.tsx` | `bubble_media.go` | `media_placeholder.png`, `media_sticker.png`, `media_document.png` | ☐ |
| 12 | Message list | `MessageList.tsx:140-175` | `message_list.go` | `day_separator.png` | ☐ |
| 13 | Composer | `ChatInput.tsx:419-566` | `composer.go` | `composer.png` | ☐ |
| 14 | Chat info + settings | `ChatInfo.tsx`, `SettingsScreen.tsx` | `chat_info.go`, `settings.go` | `chat_info.png`, `settings.png` | ☐ |

Each row's task below ends with the same three gates: **test passes**, **PNG inspected**, **build clean**. A row without all three stays unticked.

---

## File Structure

```
fyneui/
  theme_wa.go            design tokens (colors) — single source, asserted against style.css
  theme_app.go           fyne.Theme adapter mapping tokens to ColorName*
  icons.go               SVG path constants + iconImage/iconButton
  htmltext.go            HTML -> styled text runs (message bodies, previews)
  golden_test.go         render/capture/assert helpers (test-only)
  tokens_test.go         parity test: theme_wa.go vs frontend/src/style.css
  sidebar_header.go      ChatScreen.tsx <Header>
  view_tabs.go           ChatScreen.tsx tab row (chats/communities/channels/status)
  search_bar.go          ChatScreen.tsx <SearchBar>
  chat_row.go            ChatScreen.tsx <ChatListItem>
  sidebar_states.go      ChatScreen.tsx archived entry + <EmptyState>
  sidebar.go             container: composes the above, owns api.GetChatList
  chat_header.go         ChatHeader.tsx
  bubble.go              MessageItem.tsx text body + meta row
  bubble_media.go        MediaContent.tsx image/video/sticker/audio/document
  message_list.go        MessageList.tsx day separators + grouping rules
  composer.go            ChatInput.tsx
  chatview.go            container: composes header/list/composer, owns api.FetchMessagesPaged
  chat_info.go           ChatInfo.tsx (phase 4)
  settings.go            SettingsScreen.tsx (phase 4)
  app.go                 window, screen switching, split layout
  login.go               LoginScreen.tsx (already done, out of scope)
  testdata/out/          PNGs written by every golden test (committed, reviewed)
```

---

### Task 1: Baseline — green build, clean tree, tokens synced

**Files:**
- Modify: `.gitignore`
- Modify: `fyneui/theme_wa.go`
- Delete: `fyneui/chatview.go.bak`, `fyneui/sidebar.go.bak2`, `fyneui/theme_wa.go.bak`, `fyne_test.exe`, `fyne.exe`, `nul`

**Interfaces:**
- Consumes: nothing
- Produces: a `fyne-migration` HEAD where `go build ./...` and `go test ./fyneui/` both pass, with `theme_wa.go` values verified against `frontend/src/style.css`, so every later task starts from a known-good, color-correct point.

- [ ] **Step 1: Confirm the tree builds before touching anything**

Run: `cd /c/whats4linux && go build ./... 2>&1 | head -20`

Known pre-existing failure, unrelated to the Fyne work: `cmd/run.go:15:2: "github.com/wailsapp/wails/v2/pkg/options/windows" imported and not used` (Wails entrypoint, untouched since `7ee2754`). Fix it by removing that import line — the file never references `windows.` — then re-run. `go build ./cmd/fyne/` must also stay at exit 0.

Expected after the one-line fix: no output from either command. If anything else fails, fix that too and stop — do not proceed with a broken baseline.

- [ ] **Step 2: Add build artifacts to .gitignore**

Append to `.gitignore`:

```
/fyne_test.exe
/fyne.exe
/nul
fyneui/testdata/out/*.png
```

- [ ] **Step 3: Delete backup and binary artifacts**

```bash
cd /c/whats4linux
rm -f fyneui/chatview.go.bak fyneui/sidebar.go.bak2 fyneui/theme_wa.go.bak fyne_test.exe fyne.exe nul
```

- [ ] **Step 4: Verify tests still run**

Run: `cd /c/whats4linux && go test ./fyneui/ 2>&1 | tail -20`
Expected: `ok  	github.com/lugvitc/whats4linux/fyneui`

- [ ] **Step 5: Commit the baseline**

```bash
git add -A
git commit -m "chore(fyne): baseline before component-by-component parity pass"
```

---

### Task 2: Shared helpers + golden render harness

**Files:**
- Create: `fyneui/golden_test.go`
- Create: `fyneui/layout.go` (shared layout + text-fitting helpers, non-test)
- Modify: `fyneui/sidebar.go` (delete `padLayout`, `avatarColor`, `initials` — moved)

**Interfaces:**
- Consumes: `fyne.io/fyne/v2/test`, `fyne.MeasureText`
- Produces (test-only, used by every later task):
  - `renderPNG(t *testing.T, name string, obj fyne.CanvasObject, size fyne.Size) image.Image` — resizes a headless window to `size`, captures, writes `fyneui/testdata/out/<name>.png`, returns the image
  - `blankImage(size fyne.Size, bg color.Color) image.Image`
  - `diffRatio(a, b image.Image) float64` — fraction of pixels differing by more than 8 per channel
  - `assertNotBlank(t *testing.T, img image.Image, min float64)`
  - `assertPixel(t *testing.T, img image.Image, x, y int, want color.Color, tol uint8)`
  - `crop(img image.Image, x, y, w, h int) image.Image`
- Produces (production, used by every later task):
  - `type padLayout struct{ h, v float32 }` — `h` horizontal, `v` vertical inner padding
  - `truncateToWidth(s string, maxW, size float32) string` — binary-search runes with `fyne.MeasureText`, append `…`
  - `avatarColor(jid string) color.Color` — moved verbatim from `sidebar.go:821`
  - `initials(name string) string` — moved verbatim from `sidebar.go:848`

- [ ] **Step 1: Write the failing test**

```go
// fyneui/golden_test.go
package fyneui

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
)

func TestHarnessRendersColor(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	red := color.NRGBA{R: 0xff, A: 255}
	root := container.NewMax(canvas.NewRectangle(waBg), canvas.NewRectangle(red))
	img := renderPNG(t, "harness", root, fyne.NewSize(40, 40))

	assertPixel(t, img, 20, 20, red, 2)
	assertNotBlank(t, img, 0.5)
}
```

- [ ] **Step 2: Run it to confirm it fails**

Run: `cd /c/whats4linux && go test ./fyneui/ -run TestHarnessRendersColor -v`
Expected: FAIL — `undefined: renderPNG`

- [ ] **Step 3: Implement the helpers in `fyneui/golden_test.go`**

`renderPNG` uses `test.NewWindow(obj)`, `w.Resize(size)`, `w.Canvas().Capture()`, then `os.MkdirAll("testdata/out", 0o755)` and `png.Encode`. `blankImage` allocates an `image.RGBA` filled with `bg`. `diffRatio` walks both images pixel by pixel, counts pixels where any channel differs by more than 8, divides by total pixels. `assertPixel` compares `color.NRGBAModel.Convert(img.At(x,y))` against `want` with per-channel tolerance. `assertNotBlank` fails when `diffRatio(img, blankImage(size, waBg)) < min`.

- [ ] **Step 4: Run it to confirm it passes**

Run: `cd /c/whats4linux && go test ./fyneui/ -run TestHarnessRendersColor -v`
Expected: PASS, and `fyneui/testdata/out/harness.png` exists.

- [ ] **Step 5: Inspect the PNG**

Open `fyneui/testdata/out/harness.png`. Expected: a solid red 40x40 square. If it is blank, `test.NewWindow` has no software painter in this Fyne build — switch to `test.NewCanvasWithPainter(softpainter.NewPainter())` and re-run.

- [ ] **Step 6: Commit**

```bash
git add fyneui/golden_test.go
git commit -m "test(fyne): headless golden-render harness for component parity"
```

---

### Task 3: Token parity against style.css

**Files:**
- Create: `fyneui/tokens_test.go`
- Modify: `fyneui/theme_wa.go`

**Interfaces:**
- Consumes: `../frontend/src/style.css`
- Produces: a test that fails whenever a Go token drifts from the CSS variable it mirrors. Later tasks may add tokens but must add them to this map.

- [ ] **Step 1: Write the failing test**

```go
// fyneui/tokens_test.go
package fyneui

import (
	"image/color"
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestTokensMatchStyleCSS(t *testing.T) {
	raw, err := os.ReadFile("../frontend/src/style.css")
	if err != nil {
		t.Fatalf("read style.css: %v", err)
	}
	re := regexp.MustCompile(`--color-([a-z0-9-]+):\s*(#[0-9a-fA-F]{6})`)
	vars := map[string]string{}
	for _, m := range re.FindAllStringSubmatch(string(raw), -1) {
		vars[m[1]] = strings.ToLower(m[2])
	}

	cases := []struct {
		css  string
		got  color.Color
	}{
		{"dark-text", waText},
		{"dark-muted", waMuted},
		{"dark-bg", waPanel},
		{"dark-tertiary", waBorder},
		{"green", waAccent},
		{"sent-bubble-dark-bg", waSent},
		{"received-bubble-dark-bg", waRecv},
	}
	for _, c := range cases {
		want, ok := vars[c.css]
		if !ok {
			t.Fatalf("style.css has no --color-%s", c.css)
		}
		r, g, b, _ := c.got.RGBA()
		got := "#" + hex2(uint8(r>>8)) + hex2(uint8(g>>8)) + hex2(uint8(b>>8))
		if got != want {
			t.Errorf("wa token for --color-%s = %s, want %s", c.css, got, want)
		}
	}
}

func hex2(v uint8) string {
	const h = "0123456789abcdef"
	return string([]byte{h[v>>4], h[v&0xf]})
}
```

- [ ] **Step 2: Run it to see which tokens drift**

Run: `cd /c/whats4linux && go test ./fyneui/ -run TestTokensMatchStyleCSS -v`
Expected: FAIL on at least one token (the palette was hand-copied, not parsed).

- [ ] **Step 3: Fix `theme_wa.go` to the exact CSS values**

For each failing line, read the variable from `frontend/src/style.css` and set the Go token to that literal hex. Do not change the CSS.

While here, add the two tokens the later tasks need and register them in the test map:

```go
waDayChip = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 26} // dark:bg-white/10 date chip
waTime    = color.NRGBA{R: 0x86, G: 0x96, B: 0xa0, A: 255} // dark:text-[#8696a0] meta row
```

`waDayChip` is translucent, so assert it as a *blended* pixel over `waChatBg`, not as a raw token — add a case to the test only for opaque tokens; verify the chip visually in Task 12 instead.

- [ ] **Step 4: Run until it passes**

Run: `cd /c/whats4linux && go test ./fyneui/ -run TestTokensMatchStyleCSS -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add fyneui/theme_wa.go fyneui/tokens_test.go
git commit -m "test(fyne): assert palette tokens match style.css @theme"
```

---

### Task 4: Sidebar header

**Files:**
- Create: `fyneui/sidebar_header.go`
- Create: `fyneui/sidebar_header_test.go`
- Modify: `fyneui/sidebar.go` (delete `topBar`, call the new constructor)

**Interfaces:**
- Consumes: `iconImage`, `newIconButton`, `iconNewChat`, `iconMore`, `iconPerson`, `waPanel`, `waText`, `waMuted`, `waDivider`
- Produces: `newSidebarHeader(onNewChat, onMenu func()) *sidebarHeader` with `func (h *sidebarHeader) Content() fyne.CanvasObject` and `func (h *sidebarHeader) SetSelfAvatar(res fyne.Resource)`

Mirrors `ChatScreen.tsx:50-70`: `h-16`, title `text-xl font-bold`, right side two `p-2 rounded-full` icon buttons and a `w-8 h-8` avatar, bottom border `dark:border-white/5`.

- [ ] **Step 1: Write the failing test**

```go
// fyneui/sidebar_header_test.go
package fyneui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

func TestSidebarHeaderGolden(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	h := newSidebarHeader(func() {}, func() {})
	img := renderPNG(t, "sidebar_header", h.Content(), fyne.NewSize(400, 64))

	assertNotBlank(t, img, 0.01)
	// title ink occupies the left third
	assertNotBlank(t, crop(img, 8, 16, 200, 48), 0.02)
	// bottom border is the divider color
	assertPixel(t, img, 200, 63, waDivider, 6)
	// header height is exactly 64: the divider is the last row, nothing below
	if h.Content().MinSize().Height < 64 {
		t.Fatalf("header min height = %v, want >= 64", h.Content().MinSize().Height)
	}
}
```

Add `crop(img image.Image, x, y, w, h int) image.Image` to `golden_test.go` (sub-image copy).

- [ ] **Step 2: Run it to confirm it fails**

Run: `cd /c/whats4linux && go test ./fyneui/ -run TestSidebarHeaderGolden -v`
Expected: FAIL — `undefined: newSidebarHeader`

- [ ] **Step 3: Implement `fyneui/sidebar_header.go`**

Struct `sidebarHeader` holds `title *canvas.Text`, `selfAv *canvas.Image`, `selfBg *canvas.Circle`, `selfIcon fyne.CanvasObject`, `root fyne.CanvasObject`. `Content()` builds: background `canvas.NewRectangle(waPanel)`, 1px `waDivider` bottom, a `container.NewBorder` with title left and `container.NewHBox(newChatBtn, menuBtn, avatarWrap)` right, all inside `container.NewGridWrap(fyne.NewSize(0, 64), ...)`. Avatar is a 32x32 circle `waHover` with `iconPerson` at 18px behind `selfAv`, which starts hidden.

- [ ] **Step 4: Wire it into `sidebar.go`**

Replace the `topBar()` call in `NewSidebar` with `newSidebarHeader(...)`, store the returned pointer on `Sidebar` as `header *sidebarHeader`, and change `loadSelfAvatar` to call `s.header.SetSelfAvatar(res)` instead of touching `s.selfAv`.

- [ ] **Step 5: Run the test and the build**

Run: `cd /c/whats4linux && go test ./fyneui/ -run TestSidebarHeaderGolden -v && go build ./...`
Expected: PASS, no build output.

- [ ] **Step 6: Inspect the PNG against the React header**

Open `fyneui/testdata/out/sidebar_header.png` side by side with `ChatScreen.tsx:50-70`. Check: title left-aligned with 16px padding, icons right-aligned, avatar 32px, no extra vertical space.

- [ ] **Step 7: Commit**

```bash
git add fyneui/sidebar_header.go fyneui/sidebar_header_test.go fyneui/sidebar.go fyneui/golden_test.go
git commit -m "feat(fyne): sidebar header component with golden test"
```

---

### Task 5: View tabs (chats / communities / channels / status)

**Files:**
- Create: `fyneui/view_tabs.go`
- Create: `fyneui/view_tabs_test.go`
- Modify: `fyneui/sidebar.go` (delete `tabPill` + `rebuildTabs`, use the new widget)

**Interfaces:**
- Consumes: `waPillOn`, `waPillOnText`, `waPillBorder`, `waMuted`, `waHover`
- Produces: `newViewTabs(current string, onChange func(view string)) *viewTabs`, `func (v *viewTabs) SetCurrent(view string)`, `func (v *viewTabs) Content() fyne.CanvasObject`

Mirrors `ChatScreen.tsx:828-849`: `flex gap-2 px-3 pb-2 pt-1 overflow-x-auto`, each pill `rounded-full border px-3 py-1 text-sm capitalize`; active = `bg-[#e9edef] text-[#0b141a] font-medium border-transparent`, idle = `border-white/10 text-[#8696a0]`.

- [ ] **Step 1: Write the failing test**

```go
// fyneui/view_tabs_test.go
package fyneui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

func TestViewTabsGolden(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	var picked string
	v := newViewTabs("chats", func(s string) { picked = s })
	img := renderPNG(t, "view_tabs", v.Content(), fyne.NewSize(400, 40))

	assertNotBlank(t, img, 0.01)
	// active pill is a light chip: sample its interior, not the text glyph
	assertPixel(t, img, 18, 20, waPillOn, 8)
	// idle pill is transparent with a faint border; interior stays background
	assertPixel(t, img, 120, 20, waPanel, 10)

	// tapping an idle pill switches the view and reports it upward
	test.Tap(v.pills[3])
	if picked != "status" {
		t.Fatalf("onChange not called, picked=%q", picked)
	}
	if v.current != "status" {
		t.Fatalf("current = %q, want status", v.current)
	}
}
```

Note: `SetCurrent` is programmatic and must **not** fire `onChange` (it would loop back into the container). Only a real tap fires it, matching `onClick={() => setView(v)}`.

- [ ] **Step 2: Run it to confirm it fails**

Run: `cd /c/whats4linux && go test ./fyneui/ -run TestViewTabsGolden -v`
Expected: FAIL — `undefined: newViewTabs`

- [ ] **Step 3: Implement `fyneui/view_tabs.go`**

`viewTabs` is a `widget.BaseWidget` holding `current string`, `onChange func(string)`, `pills []*tabChip`. `tabChip` renders a `canvas.Rectangle` (CornerRadius = half height), an optional 1px border rectangle, and a `canvas.Text` at 13px, `capitalize` applied to the label. `CreateRenderer` returns `widget.NewSimpleRenderer(container.NewHBox(pills...))` wrapped in `container.New(&padLayout{h: 12, v: 4}, ...)`.

- [ ] **Step 4: Run the test**

Run: `cd /c/whats4linux && go test ./fyneui/ -run TestViewTabsGolden -v`
Expected: PASS. If the active-pill pixel assertion fails, the chip is being laid out at a different x — print `v.Content().MinSize()` and adjust the sample coordinate, do not loosen the tolerance.

- [ ] **Step 5: Wire into `sidebar.go`**

Replace `s.tabBox`/`rebuildTabs` with `s.tabs = newViewTabs("chats", func(v string){ ... s.syncChrome(); s.refreshVisible() })`. `syncChrome` keeps hiding the search bar when `view == "status"`.

- [ ] **Step 6: Build, inspect, commit**

Run: `go build ./... && go test ./fyneui/ -run TestViewTabsGolden -v`
Inspect `fyneui/testdata/out/view_tabs.png` against `ChatScreen.tsx:828-849`.

```bash
git add fyneui/view_tabs.go fyneui/view_tabs_test.go fyneui/sidebar.go
git commit -m "feat(fyne): view tabs component with golden test"
```

---

### Task 6: Search bar

**Files:**
- Create: `fyneui/search_bar.go`
- Create: `fyneui/search_bar_test.go`
- Modify: `fyneui/sidebar.go` (delete `searchBar`)

**Interfaces:**
- Consumes: `iconSearch`, `waPanel2`, `waMuted`, `waText`
- Produces: `newSearchBar(placeholder string, onChange func(string)) *searchBar`, `func (s *searchBar) Content() fyne.CanvasObject`, `func (s *searchBar) SetPlaceholder(p string)`

Mirrors `ChatScreen.tsx:84-95`: outer `px-3 py-2`, inner `rounded-full bg-[#242626] flex items-center px-4 py-2`, icon `mr-4`, input `bg-transparent text-sm`.

- [ ] **Step 1: Write the failing test**

```go
// fyneui/search_bar_test.go
package fyneui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

func TestSearchBarGolden(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	var got string
	s := newSearchBar("Search or start new chat", func(v string) { got = v })
	img := renderPNG(t, "search_bar", s.Content(), fyne.NewSize(400, 48))

	// pill background fills the middle, inset by px-3
	assertPixel(t, img, 200, 24, waPanel2, 8)
	// outer margin is the panel background
	assertPixel(t, img, 2, 24, waPanel, 6)
	if h := s.Content().MinSize().Height; h < 40 {
		t.Fatalf("search bar min height = %v, want >= 40", h)
	}
	_ = got
}
```

- [ ] **Step 2: Run it to confirm it fails**

Run: `cd /c/whats4linux && go test ./fyneui/ -run TestSearchBarGolden -v`
Expected: FAIL — `undefined: newSearchBar`

- [ ] **Step 3: Implement `fyneui/search_bar.go`**

Entry with no border (theme already maps `ColorNameInputBorder` to the pill color), `CornerRadius` on the background rectangle = half the pill height, icon in a left `container.NewBorder` slot with `mr-4` = 16px padding, outer `padLayout{h: 12, v: 6}`.

- [ ] **Step 4: Run the test, wire into `sidebar.go`, build**

Run: `go test ./fyneui/ -run TestSearchBarGolden -v && go build ./...`
Expected: PASS, no build output.

- [ ] **Step 5: Inspect and commit**

Inspect `fyneui/testdata/out/search_bar.png`: pill fully rounded, no square outline around the Entry, placeholder visible.

```bash
git add fyneui/search_bar.go fyneui/search_bar_test.go fyneui/sidebar.go
git commit -m "feat(fyne): search bar component with golden test"
```

---

### Task 7: Chat row

**Files:**
- Create: `fyneui/chat_row.go`
- Create: `fyneui/chat_row_test.go`
- Modify: `fyneui/sidebar.go` (delete `newRow`/`updateRow`/`chatRowRefs`, call `newChatRow()`)

**Interfaces:**
- Consumes: `iconPin`, `waText`, `waMuted`, `waGreen`, `waHover`, `waBadgeText`, `avatarColor`, `initials`
- Produces:
  ```go
  type chatRowData struct {
      JID, Name, Preview, Time string
      Pinned                   bool
      Unread                   int
      Avatar                   fyne.Resource // nil => initials
      AvatarBg                 color.Color
  }
  func newChatRow() *chatRow
  func (r *chatRow) Set(d chatRowData)
  func (r *chatRow) Content() fyne.CanvasObject
  ```

Mirrors `ChatScreen.tsx:247-320`: `w-12 h-12 rounded-full mr-4` avatar, `text-[13px]` preview truncated, `text-sm` timestamp, pin icon `shrink-0`, badge `min-w-5 h-5 px-1.5 rounded-full bg-[#21c063] text-[#0a1014] text-xs font-semibold`.

- [ ] **Step 1: Write the failing test**

```go
// fyneui/chat_row_test.go
package fyneui

import (
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

func TestChatRowGolden(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	r := newChatRow()
	r.Set(chatRowData{
		JID:      "6282134544728@s.whatsapp.net",
		Name:     "Cintakuuu",
		Preview:  "tawar mass murah banget",
		Time:     "03:22",
		Pinned:   true,
		Unread:   3,
		AvatarBg: avatarColor("6282134544728@s.whatsapp.net"),
	})
	img := renderPNG(t, "chat_row", r.Content(), fyne.NewSize(400, 72))

	assertNotBlank(t, img, 0.02)
	// badge sits in the bottom-right corner and uses the unread green
	assertPixel(t, img, 386, 56, waGreen, 10)
	if h := r.Content().MinSize().Height; h < 72 {
		t.Fatalf("row min height = %v, want >= 72", h)
	}
}

func TestChatRowTruncatesLongName(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	r := newChatRow()
	long := strings.Repeat("Vapers semarangan 18+ ", 6)
	r.Set(chatRowData{Name: long, Preview: strings.Repeat("x", 200), Time: "12/10/2026", AvatarBg: waPanel2})
	img := renderPNG(t, "chat_row_long", r.Content(), fyne.NewSize(320, 72))

	// the timestamp column must still be reachable at the 320px minimum
	assertNotBlank(t, crop(img, 280, 8, 40, 20), 0.02)
}
```

- [ ] **Step 2: Run both to confirm they fail**

Run: `cd /c/whats4linux && go test ./fyneui/ -run TestChatRow -v`
Expected: FAIL — `undefined: newChatRow`

- [ ] **Step 3: Implement `fyneui/chat_row.go`**

`chatRow` is a `widget.BaseWidget` with fields for every mutable leaf (`name *canvas.Text`, `time *canvas.Text`, `preview *canvas.Text`, `pin fyne.CanvasObject`, `badgeBg *canvas.Rectangle`, `badgeTxt *canvas.Text`, `avatarImg *canvas.Image`, `avatarTxt *canvas.Text`, `avatarBg *canvas.Circle`, `hover *canvas.Rectangle`). `Set` writes each field then calls `Refresh()`. Name and preview are `canvas.Text` inside `container.NewMax` clipped by `container.NewGridWrap` — **never** index `container.Objects` to find a child (that caused the earlier `index out of range [2]` panic).

Truncation: `canvas.Text` does not ellipsize, so `Set` runs `truncateToWidth(name, availableWidth, 15)` — a helper that binary-searches runes using `fyne.MeasureText` and appends `…`.

- [ ] **Step 4: Run both tests**

Run: `cd /c/whats4linux && go test ./fyneui/ -run TestChatRow -v`
Expected: PASS both. If `TestChatRowTruncatesLongName` fails, the timestamp is being pushed out — reduce the name's available width by the measured time width, do not hardcode a pixel budget.

- [ ] **Step 5: Wire into `sidebar.go`**

`Sidebar.newRow` returns `newChatRow().Content()`; `updateRow` calls `refs.row.Set(chatRowData{...})`. Delete `chatRowRefs` and the `rowRefs` map; the `widget.List` template already hands back the same object per item.

- [ ] **Step 6: Build, run the app, inspect**

Run: `go build ./... && go test ./fyneui/ -run TestChatRow -v`
Then launch with `WHATS4LINUX_OPEN_JID` unset and confirm the chat list still renders 300+ rows without panic.

- [ ] **Step 7: Commit**

```bash
git add fyneui/chat_row.go fyneui/chat_row_test.go fyneui/sidebar.go
git commit -m "feat(fyne): chat row component with truncation and golden tests"
```

---

### Task 8: Archived entry and empty state

**Files:**
- Create: `fyneui/sidebar_states.go`
- Create: `fyneui/sidebar_states_test.go`
- Modify: `fyneui/sidebar.go` (delete `archivedRow`/placeholder, use the new constructors)

**Interfaces:**
- Consumes: `waGreen`, `waText`, `waMuted`, `waHover`, `iconArchive` (new: the archive-box path copied verbatim from `ChatScreen.tsx:869`)
- Produces:
  - `newArchivedEntry(count int, onOpen func()) fyne.CanvasObject`
  - `newSidebarEmptyState(hasChats bool, onRetry func()) fyne.CanvasObject`

Mirrors `ChatScreen.tsx:861-874` (archived row) and `ChatScreen.tsx:362-375` (`EmptyState`). Exact copy: "No chats match your search." / "No chats available. Start a conversation!" and the retry button label "Refresh Chats" / "Loading...".

- [ ] **Step 1: Write the failing test**

```go
// fyneui/sidebar_states_test.go
package fyneui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

func TestArchivedEntryGolden(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	img := renderPNG(t, "archived_entry", newArchivedEntry(7, func() {}), fyne.NewSize(400, 48))
	assertNotBlank(t, img, 0.01)
	// the archive glyph is green, like dark:text-[#21c063]
	assertPixel(t, img, 28, 24, waGreen, 40)
}

func TestEmptyStateGolden(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	img := renderPNG(t, "sidebar_empty", newSidebarEmptyState(false, func() {}), fyne.NewSize(320, 200))
	assertNotBlank(t, img, 0.01)
}
```

- [ ] **Step 2: Run to confirm failure**

Run: `cd /c/whats4linux && go test ./fyneui/ -run 'TestArchived|TestEmptyState' -v`
Expected: FAIL — undefined constructors

- [ ] **Step 3: Implement both constructors**

Archived: 48px row, green archive glyph in a `w-12` slot, label "Archived" `flex-1 font-medium`, count `text-xs` muted on the right, hover background `waHover`.

Empty state: centered muted text. Two variants — `hasChats == false` → "No chats yet" plus a retry button; `hasChats == true` → "No chats match your search".

- [ ] **Step 4: Run tests and wire**

Run: `go test ./fyneui/ -run 'TestArchived|TestEmptyState' -v && go build ./...`
In `sidebar.go`, replace `s.archivedR` with a lazily built `newArchivedEntry(count, func(){ s.setArchived(true) })` shown only when `view == "chats" && archivedCount > 0 && !s.archived`.

- [ ] **Step 5: Inspect and commit**

```bash
git add fyneui/sidebar_states.go fyneui/sidebar_states_test.go fyneui/sidebar.go
git commit -m "feat(fyne): archived entry and empty state components"
```

---

### Task 9: Chat header

**Files:**
- Create: `fyneui/chat_header.go`
- Create: `fyneui/chat_header_test.go`
- Modify: `fyneui/chatview.go` (delete header construction from `Content()`)

**Interfaces:**
- Consumes: `iconBack`, `iconMore`, `iconPerson`, `waPanel`, `waDivider`, `waText`, `waMuted`, `avatarColor`
- Produces:
  ```go
  type chatHeaderData struct {
      Name, Subtitle string
      Avatar         fyne.Resource
      AvatarBg       color.Color
      ShowCall       bool
  }
  func newChatHeader(onBack, onInfo, onCall func()) *chatHeader
  func (h *chatHeader) Set(d chatHeaderData)
  func (h *chatHeader) Content() fyne.CanvasObject
  ```

Mirrors `ChatHeader.tsx`: `p-3` container, `gap-3`, `w-10 h-10 rounded-full` avatar, name `text-[16px] font-medium truncate`, subtitle `text-xs text-[#8696a0] truncate`, right side optional phone button then info button, `border-b border-white/5`.

- [ ] **Step 1: Write the failing test**

```go
// fyneui/chat_header_test.go
package fyneui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

func TestChatHeaderGolden(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	h := newChatHeader(func() {}, func() {}, func() {})
	h.Set(chatHeaderData{
		Name:     "Cintakuuu",
		Subtitle: "click here for contact info",
		AvatarBg: avatarColor("6282134544728@s.whatsapp.net"),
	})
	img := renderPNG(t, "chat_header", h.Content(), fyne.NewSize(800, 64))

	assertNotBlank(t, img, 0.02)
	// name/subtitle block sits right of the 40px avatar
	assertNotBlank(t, crop(img, 72, 12, 300, 40), 0.02)
	// bottom border
	assertPixel(t, img, 400, 63, waDivider, 6)
}

func TestChatHeaderGroupHidesCall(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	h := newChatHeader(func() {}, func() {}, nil)
	h.Set(chatHeaderData{Name: "Vapers semarangan 18+", Subtitle: "group info", ShowCall: false})
	img := renderPNG(t, "chat_header_group", h.Content(), fyne.NewSize(800, 64))
	assertNotBlank(t, img, 0.02)
}
```

- [ ] **Step 2: Run to confirm failure**

Run: `cd /c/whats4linux && go test ./fyneui/ -run TestChatHeader -v`
Expected: FAIL — undefined

- [ ] **Step 3: Implement `fyneui/chat_header.go`**

Same structure as `sidebar_header.go` but with a 40px avatar, a two-line text column, and no fixed-width title. `Set` truncates name and subtitle via `truncateToWidth` at 16px and 12px respectively. `ShowCall` toggles the phone button.

- [ ] **Step 4: Wire into `chatview.go`**

`ChatView` gains `header *chatHeader`. `loadHeader` fills `chatHeaderData` instead of touching `canvas.Text` fields directly; delete `headerName`, `headerSub`, `avatarTxt`, `avatarBg`, `avatarImg` from `ChatView`.

- [ ] **Step 5: Run, build, inspect, commit**

```bash
go test ./fyneui/ -run TestChatHeader -v && go build ./...
git add fyneui/chat_header.go fyneui/chat_header_test.go fyneui/chatview.go
git commit -m "feat(fyne): chat header component with golden tests"
```

---

### Task 10: Message bubble — text body, meta, and the blank-bubble bug

**Files:**
- Create: `fyneui/bubble.go`
- Create: `fyneui/bubble_test.go`
- Modify: `fyneui/chatview.go` (delete `bubble`, `metaRow`, `textBlock`, `unsupportedText`)

**Interfaces:**
- Consumes: `htmlTextBlock`, `htmlToText`, `waSent`, `waRecv`, `waText`, `waTime`, `waTick`, `waGreen`
- Produces:
  ```go
  type bubbleData struct {
      HTML      string
      FromMe    bool
      Edited    bool
      Time      string
      Sender    string // group chats, received only
      FirstInGroup bool
      EmojiOnly bool
      MaxWidth  float32
  }
  func newBubble(d bubbleData) fyne.CanvasObject
  func bubbleCorner(firstInGroup bool) float32 // 2 when true, 12 otherwise
  ```

Mirrors `MessageItem.tsx:406-560`: `max-w-[85%] lg:max-w-[65%] rounded-xl px-2 pt-1 pb-1.5`, sent `bg-sent-bubble-dark-bg`, received `bg-received-bubble-dark-bg`, meta `text-[11px] opacity-60` with `BlueTickIcon` when `isFromMe`, sender name `text-[11px] font-semibold` in the group color.

**Known bug this task fixes:** current bubbles render empty because `textBlock` wraps the `RichText` in `container.NewGridWrap` sized from `box.MinSize().Height`, which is `0` for a `RichText` whose `Segments` were assigned after construction without a layout pass. The fix is to stop guessing the height: give the `RichText` an explicit `MinSize` and let the bubble's `VBox` derive its own height.

- [ ] **Step 1: Write the failing tests**

```go
// fyneui/bubble_test.go
package fyneui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

func TestBubbleTextVisible(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	b := newBubble(bubbleData{HTML: "<p>oke lah</p><p>iya</p>", FromMe: true, Time: "22:36", MaxWidth: 560})
	img := renderPNG(t, "bubble_text", b, fyne.NewSize(600, 120))

	// text ink must exist inside the sent bubble, not just the bubble fill
	assertNotBlank(t, crop(img, 60, 10, 460, 80), 0.02)
}

func TestBubbleEmptyContent(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	b := newBubble(bubbleData{HTML: "", FromMe: false, Time: "22:36", MaxWidth: 560})
	img := renderPNG(t, "bubble_empty", b, fyne.NewSize(600, 80))
	assertNotBlank(t, img, 0.01)
}

func TestBubbleLongURLWraps(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	long := "https://uc-share.com/s/2afb0d24aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	b := newBubble(bubbleData{HTML: "<p><a href=\"" + long + "\">" + long + "</a></p>", FromMe: false, Time: "22:36", MaxWidth: 360})
	img := renderPNG(t, "bubble_long_url", b, fyne.NewSize(600, 160))

	// nothing may be painted in the 80px reserved gutter on the right
	assertPixel(t, img, 560, 40, waBg, 6)
}

func TestBubbleEmojiOnlyIsLarge(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	small := newBubble(bubbleData{HTML: "<p>😊</p>", FromMe: true, Time: "22:36", MaxWidth: 560})
	big := newBubble(bubbleData{HTML: "<p>😊</p>", FromMe: true, Time: "22:36", MaxWidth: 560, EmojiOnly: true})
	if big.MinSize().Height <= small.MinSize().Height {
		t.Fatalf("emoji-only bubble height %v not larger than normal %v", big.MinSize().Height, small.MinSize().Height)
	}
}

func TestBubbleCorner(t *testing.T) {
	if bubbleCorner(true) != 2 {
		t.Errorf("first-in-group corner = %v, want 2", bubbleCorner(true))
	}
	if bubbleCorner(false) != 12 {
		t.Errorf("normal corner = %v, want 12", bubbleCorner(false))
	}
}
```

Add `EmojiOnly bool` to `bubbleData`; `newBubble` sets the text size to 32 when true (`text-[32px] leading-10`).

- [ ] **Step 2: Run to confirm all five fail**

Run: `cd /c/whats4linux && go test ./fyneui/ -run TestBubble -v`
Expected: FAIL — `undefined: newBubble`, `bubbleData`, `bubbleCorner`

- [ ] **Step 3: Implement `fyneui/bubble.go`**

Body: `htmlTextBlock(d.HTML, maxInner, 14)` → a `*widget.RichText` with `Wrapping = fyne.TextWrapOff`. Set `rt.Resize(fyne.NewSize(w, h))` **and** wrap it in `container.NewGridWrap(fyne.NewSize(w, h), rt)` where `h` comes from `htmlRichText`'s third return value, not from `MinSize()`. If `w < 24`, fall back to a `canvas.Text` of `htmlToText(d.HTML)` so a single short word is never zero-width.

Empty body (`htmlToText == ""`): italic `"Empty Message"` at 13px in `waMuted`, matching `MessageItem.tsx:254`.

Meta row: `HBox(spacer, edited?, time, ticks?)` at 11px in `waTime`; ticks are `✓✓` in `waTick` when `FromMe`.

Group sender label: 11px bold in `waGreen` above the body, only when `Sender != ""`.

Background: `canvas.NewRectangle(waSent)` / `waRecv` with `CornerRadius = bubbleCorner(d.FirstInGroup)`.

Layout: `container.NewBorder(spacer(1,3), spacer(1,3), leftGutter, rightGutter, stack)` where the gutter on the sending side is 12px and on the opposite side 70px, matching `mx-5` plus the `max-w-[65%]` reserve.

- [ ] **Step 4: Run until all five pass**

Run: `cd /c/whats4linux && go test ./fyneui/ -run TestBubble -v`
Expected: PASS ×5. If `TestBubbleTextVisible` still fails, the RichText is being sized before its segments settle — call `rt.Refresh()` then re-measure with `fyne.MeasureText` per line instead of trusting `Segments`.

- [ ] **Step 5: Inspect all five PNGs**

Open `fyneui/testdata/out/bubble_text.png`, `bubble_empty.png`, `bubble_long_url.png`. Confirm: text visible, ticks blue, long URL wrapped inside the bubble, no clipping on the right.

- [ ] **Step 6: Wire into `chatview.go` and verify against the live app**

`ChatView.bubble` becomes `newBubble(bubbleData{...})` built from `store.DecodedMessage`. Then run:

```bash
go build -o fyne_test.exe ./cmd/fyne
WHATS4LINUX_OPEN_JID="6282134544728@s.whatsapp.net" ./fyne_test.exe > "$LOCALAPPDATA/Temp/fyne_run27.log" 2>&1
```

Capture the window and confirm the previously blank bubbles now show their text.

- [ ] **Step 7: Commit**

```bash
git add fyneui/bubble.go fyneui/bubble_test.go fyneui/chatview.go
git commit -m "fix(fyne): bubble text rendered blank; add bubble component + golden tests"
```

---

### Task 11: Media bubbles — image, video, sticker, audio, document

**Files:**
- Create: `fyneui/bubble_media.go`
- Create: `fyneui/bubble_media_test.go`
- Modify: `fyneui/chatview.go` (delete `mediaBlock`, `audioBlock`, `documentBlock`, `mediaDims`)

**Interfaces:**
- Consumes: `mediaBox`, `dataURLResource`, `waPanel2`, `waMuted`, `waPlayOverlay`, `iconPlay`, `iconDocument`
- Produces:
  ```go
  type mediaBubbleData struct {
      Kind      mediaKind // mediaKindImage | mediaKindVideo | mediaKindSticker | mediaKindAudio | mediaKindDocument
      Thumb     fyne.Resource // nil => placeholder
      Caption   string
      FileName  string
      SizeLabel string
      Width, Height int
      MaxWidth  float32
  }
  func newMediaBubble(d mediaBubbleData) fyne.CanvasObject
  ```

Mirrors `MediaContent.tsx`: image/video get a rounded thumbnail with a play overlay for video, stickers have **no** bubble background, audio is a waveform pill with a play button, documents are a card with a red file-type chip and the filename.

- [ ] **Step 1: Write the failing tests**

```go
// fyneui/bubble_media_test.go
package fyneui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

func TestMediaPlaceholderGolden(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	b := newMediaBubble(mediaBubbleData{Kind: mediaKindImage, Width: 1200, Height: 900, MaxWidth: 560})
	img := renderPNG(t, "media_placeholder", b, fyne.NewSize(600, 320))

	// the placeholder box is waPanel2, visible even with no thumbnail bytes
	assertPixel(t, img, 300, 160, waPanel2, 8)
	assertNotBlank(t, img, 0.02)
}

func TestStickerHasNoBubbleBackground(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	b := newMediaBubble(mediaBubbleData{Kind: mediaKindSticker, MaxWidth: 560})
	img := renderPNG(t, "media_sticker", b, fyne.NewSize(600, 200))

	// far corners stay on the chat wallpaper, not on a sent/received bubble color
	assertPixel(t, img, 20, 20, waBg, 6)
	assertPixel(t, img, 580, 180, waBg, 6)
}

func TestDocumentBubbleGolden(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	b := newMediaBubble(mediaBubbleData{
		Kind: mediaKindDocument, FileName: "new.ipynb", SizeLabel: "12 KB", MaxWidth: 560,
	})
	img := renderPNG(t, "media_document", b, fyne.NewSize(600, 120))
	assertNotBlank(t, img, 0.02)
}
```

- [ ] **Step 2: Run to confirm failure**

Run: `cd /c/whats4linux && go test ./fyneui/ -run 'TestMedia|TestSticker|TestDocument' -v`
Expected: FAIL — undefined

- [ ] **Step 3: Implement `fyneui/bubble_media.go`**

Image/video: `holder := canvas.NewRectangle(waPanel2)` with `CornerRadius = 8`, a centered muted label (`📷 Photo` / `🎥 Video`) behind a `canvas.Image` that starts hidden; video adds a 48px translucent circle with `iconPlay`. Sticker: 150x150 image with no rectangle behind it. Audio: pill with `iconPlay` and a fixed 120px waveform placeholder. Document: card with a red 40x48 chip, filename truncated, size label muted.

- [ ] **Step 4: Run tests, wire into `chatview.go`, build**

`ChatView.mediaBlock` is replaced by `newMediaBubble(...)` with `Thumb` filled from the existing fetch chain (`GetImageThumbnail` → `GetCachedImage` → `DownloadMedia`, video via `GetVideoThumbnail`). Keep the existing goroutine + `fyne.Do` structure; only the view construction changes.

Run: `go test ./fyneui/ -run 'TestMedia|TestSticker|TestDocument' -v && go build ./...`
Expected: PASS ×3.

- [ ] **Step 5: Inspect and commit**

```bash
git add fyneui/bubble_media.go fyneui/bubble_media_test.go fyneui/chatview.go
git commit -m "feat(fyne): media bubble component with placeholder and sticker variants"
```

---

### Task 12: Message list — day separators and run grouping

**Files:**
- Create: `fyneui/message_list.go`
- Create: `fyneui/message_list_test.go`
- Modify: `fyneui/chatview.go` (delete `daySeparator`, `lastDay`, inline grouping in `addMessageLocked`)

**Interfaces:**
- Consumes: `newBubble`, `bubbleData`, `waPanel2`, `waText`
- Produces:
  ```go
  type listItem struct {
      Kind    int // itemDay | itemMessage
      Day     string
      Msg     store.DecodedMessage
      FirstInGroup bool
  }
  func buildListItems(msgs []store.DecodedMessage) []listItem
  func newDaySeparator(label string) fyne.CanvasObject
  ```

Mirrors `MessageList.tsx:156-166` + `MessageItem.tsx` `firstInGroup`: a run breaks **only** when `prev.Info.IsFromMe` changes or `prev.Info.Sender` changes (React has no time-gap rule — do not invent one). Day separators come from `dayKey(ts)` changing, labelled with `formatDateSeparator`.

- [ ] **Step 1: Write the failing tests**

```go
// fyneui/message_list_test.go
package fyneui

import (
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	"github.com/lugvitc/whats4linux/internal/store"
)

func mkMsg(id string, fromMe bool, sender string, ts time.Time) store.DecodedMessage {
	m := store.DecodedMessage{}
	m.Info.ID = id
	m.Info.Sender = sender
	m.Info.IsFromMe = fromMe
	m.Info.Timestamp = ts.Format(time.RFC3339)
	return m
}

func TestBuildListItemsGroupsRuns(t *testing.T) {
	base := time.Date(2026, 11, 7, 22, 30, 0, 0, time.Local)
	msgs := []store.DecodedMessage{
		mkMsg("a", false, "u1", base),
		mkMsg("b", false, "u1", base.Add(30*time.Second)),
		mkMsg("c", false, "u1", base.Add(20*time.Minute)), // same sender: still the same run
		mkMsg("d", false, "u2", base.Add(21*time.Minute)), // sender change breaks it
		mkMsg("e", true, "me", base.Add(22*time.Minute)),  // direction change breaks it
	}
	items := buildListItems(msgs)
	got := []bool{}
	for _, it := range items {
		if it.Kind == itemMessage {
			got = append(got, it.FirstInGroup)
		}
	}
	want := []bool{true, false, false, true, true}
	if len(got) != len(want) {
		t.Fatalf("got %d messages, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("message %d FirstInGroup = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestBuildListItemsInsertsDaySeparator(t *testing.T) {
	d1 := time.Date(2026, 11, 6, 9, 0, 0, 0, time.Local)
	d2 := time.Date(2026, 11, 7, 9, 0, 0, 0, time.Local)
	items := buildListItems([]store.DecodedMessage{mkMsg("a", false, "u1", d1), mkMsg("b", false, "u1", d2)})
	if len(items) != 4 {
		t.Fatalf("got %d items, want 4 (day, msg, day, msg)", len(items))
	}
	if items[0].Kind != itemDay || items[2].Kind != itemDay {
		t.Fatalf("day separators missing: %+v", items)
	}
}

func TestDaySeparatorGolden(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	img := renderPNG(t, "day_separator", newDaySeparator("27/08/2026"), fyne.NewSize(600, 40))
	assertNotBlank(t, img, 0.005)
	// pill is centered and uses the white/10 chip over the chat wallpaper
	assertPixel(t, img, 300, 20, waDayChip, 12)
}
```

- [ ] **Step 2: Run to confirm failure**

Run: `cd /c/whats4linux && go test ./fyneui/ -run 'TestBuildList|TestDaySeparator' -v`
Expected: FAIL — undefined

- [ ] **Step 3: Implement `fyneui/message_list.go`**

`buildListItems` walks messages oldest→newest, emits an `itemDay` whenever `dayKey` of the local calendar date changes, and sets `FirstInGroup` when there is no previous message, `prev.Info.IsFromMe != cur.Info.IsFromMe`, or `prev.Info.Sender != cur.Info.Sender`. `newDaySeparator` centers a `rounded-full` chip (`waDayChip` = white at 10% over `waChatBg`, `CornerRadius = 8`, 11px text in `waMuted`) inside a 40px row — matching `MessageList.tsx:57-65`.

- [ ] **Step 4: Run tests, wire into `chatview.go`, build**

`ChatView.addMessageLocked` is replaced by a full rebuild path: `cv.messages.Objects = nil`, then for each `listItem` append either `newDaySeparator` or `newBubble(bubbleData{FirstInGroup: it.FirstInGroup, ...})`. Rebuilding the whole `VBox` on every new message is acceptable at 80-message page size; if profiling shows it is not, switch to appending only the tail.

Run: `go test ./fyneui/ -run 'TestBuildList|TestDaySeparator' -v && go build ./...`
Expected: PASS ×3.

- [ ] **Step 5: Inspect and commit**

```bash
git add fyneui/message_list.go fyneui/message_list_test.go fyneui/chatview.go
git commit -m "feat(fyne): message list grouping and day separators with tests"
```

---

### Task 13: Composer

**Files:**
- Create: `fyneui/composer.go`
- Create: `fyneui/composer_test.go`
- Modify: `fyneui/chatview.go` (delete composer construction from `Content()`)

**Interfaces:**
- Consumes: `iconEmoji`, `iconAttach`, `iconSend`, `newIconButton`, `waPanel`, `waPanel2`, `waAccent`, `waWhite`, `waMuted`
- Produces: `newComposer(onSend func(text string), onEmoji, onAttach func()) *composer`, `func (c *composer) Content() fyne.CanvasObject`, `func (c *composer) Text() string`, `func (c *composer) Clear()`

Mirrors `ChatInput.tsx:419-566`: outer `relative mx-3 mb-3 mt-1.5` (12px sides, 12px bottom, 6px top), pill `rounded-full border border-transparent bg-[#242626] px-1.5`, emoji button inside left, attach inside right, send as a separate `h-11 w-11 rounded-full bg-green` circle.

- [ ] **Step 1: Write the failing test**

```go
// fyneui/composer_test.go
package fyneui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

func TestComposerGolden(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	c := newComposer(func(string) {}, func() {}, func() {})
	img := renderPNG(t, "composer", c.Content(), fyne.NewSize(800, 64))

	// pill interior is panel2, inset by 12px on both sides
	assertPixel(t, img, 400, 30, waPanel2, 8)
	assertPixel(t, img, 4, 30, waPanel, 6)
	// send button is a 44px accent circle on the far right
	assertPixel(t, img, 770, 30, waAccent, 8)
}

func TestComposerSendClears(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	var sent string
	c := newComposer(func(s string) { sent = s }, func() {}, func() {})
	c.entry.SetText("halo")
	c.send()
	if sent != "halo" {
		t.Fatalf("onSend got %q, want %q", sent, "halo")
	}
	if c.Text() != "" {
		t.Fatalf("composer not cleared after send: %q", c.Text())
	}
}
```

- [ ] **Step 2: Run to confirm failure**

Run: `cd /c/whats4linux && go test ./fyneui/ -run TestComposer -v`
Expected: FAIL — undefined

- [ ] **Step 3: Implement `fyneui/composer.go`**

`composer` holds `entry *widget.Entry` (multi-line, `TextWrapWord`, `SetMinRowsVisible(1)`, `SetPlaceHolder("Type a message")`), `send *canvas.Circle`. `Content()` builds `container.NewBorder(nil, nil, nil, sendWrap, pill)` inside `padLayout{h:12, v:6}` over `waPanel`. `send()` calls `onSend(strings.TrimSpace(entry.Text))` then `Clear()`. `OnSubmitted` is wired to `send()` so Enter sends (Shift+Enter inserts a newline in Fyne's multi-line Entry by default).

- [ ] **Step 4: Run tests, wire into `chatview.go`, build**

Run: `go test ./fyneui/ -run TestComposer -v && go build ./...`
Expected: PASS ×2.

- [ ] **Step 5: Inspect and commit**

Inspect `fyneui/testdata/out/composer.png`: pill fully rounded, no square outline, send circle flush right with 12px margin.

```bash
git add fyneui/composer.go fyneui/composer_test.go fyneui/chatview.go
git commit -m "feat(fyne): composer component with golden and send-behavior tests"
```

---

### Task 14: Chat info panel and settings screen

**Files:**
- Create: `fyneui/chat_info.go`
- Create: `fyneui/settings.go`
- Create: `fyneui/chat_info_test.go`
- Create: `fyneui/settings_test.go`
- Modify: `fyneui/chatview.go`, `fyneui/app.go`

**Interfaces:**
- Consumes: `api.GetGroupInfo`, `api.GetProfile`, `api.GetSettings`, `api.SaveSettings`
- Produces:
  - `newChatInfoPanel(chatName, chatType string, onClose func()) *chatInfoPanel` with `SetDetails(name, subtitle, about string)` and `Content()`
  - `newSettingsScreen(onBack func()) *settingsScreen` with `Content()`

Mirrors `ChatInfo.tsx` (right-side slide-over, `w-80`, header with close button, avatar, name, about section) and `SettingsScreen.tsx` (list of rows navigating to sub-screens).

- [ ] **Step 1: Write the failing tests**

```go
// fyneui/chat_info_test.go
package fyneui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

func TestChatInfoPanelGolden(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	p := newChatInfoPanel("Cintakuuu", "contact", func() {})
	p.SetDetails("Cintakuuu", "+62 821-3454-4728", "Available")
	img := renderPNG(t, "chat_info", p.Content(), fyne.NewSize(320, 500))
	assertNotBlank(t, img, 0.01)
}
```

```go
// fyneui/settings_test.go
package fyneui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

func TestSettingsScreenGolden(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	s := newSettingsScreen(func() {})
	img := renderPNG(t, "settings", s.Content(), fyne.NewSize(600, 500))
	assertNotBlank(t, img, 0.01)
}
```

- [ ] **Step 2: Run to confirm failure**

Run: `cd /c/whats4linux && go test ./fyneui/ -run 'TestChatInfo|TestSettings' -v`
Expected: FAIL — undefined

- [ ] **Step 3: Implement both screens**

Chat info: fixed 320px wide panel, `waPanel` background, 64px header with title + close icon, 96px avatar, name at 18px, subtitle muted, then a divider and an "About" section. The panel is added as a third child of the chat pane's `container.NewBorder` when open.

Settings: header with back icon and "Settings", then a list of rows (Account, Privacy, Chats, Notifications, Advanced, Log out) each with an icon, label and chevron. Row taps are stubs that log the target screen name — sub-screens are out of scope for this plan.

- [ ] **Step 4: Run tests, wire, build**

Run: `go test ./fyneui/ -run 'TestChatInfo|TestSettings' -v && go build ./...`
Expected: PASS ×2.

- [ ] **Step 5: Inspect and commit**

```bash
git add fyneui/chat_info.go fyneui/settings.go fyneui/chat_info_test.go fyneui/settings_test.go fyneui/chatview.go fyneui/app.go
git commit -m "feat(fyne): chat info panel and settings screen scaffolds"
```

---

## Execution Handoff

Plan saved to `docs/superpowers/plans/2026-11-07-fyne-component-parity.md`.

Work order is strict: Task 1 → Task 14. Task 10 (bubble) fixes the visible blank-bubble bug Adit reported; Tasks 4-8 rebuild the sidebar; Task 13 fixes the composer box; Task 12 fixes day separators and run grouping. Each task leaves the app buildable and adds one PNG under `fyneui/testdata/out/` that can be compared against the named React file.

Recommended execution: **native, one task per turn**, because every task depends on the previous task's file split and the same session needs to look at each PNG before moving on. Tick the Component Checklist row only after test + PNG inspection + build all pass.
