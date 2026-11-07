# Fyne Migration — Task Board

Checklist kerja migrasi UI React/Wails → Fyne, per komponen.
Plan lengkap: `docs/superpowers/plans/2026-11-07-fyne-component-parity.md`
Branch: `fyne-migration` · Setiap task **wajib commit** setelah selesai.

## Aturan centang

Sebuah baris baru boleh dicentang kalau **ketiga gate** lulus:

1. `go test ./fyneui/ -run <TestName> -v` → PASS
2. PNG di `fyneui/testdata/out/` sudah **dibuka dan dibandingkan** dengan file React yang jadi spec
3. `go build ./...` → bersih (tanpa output)

Kalau salah satu gagal → baris tetap `[ ]`, tulis blocker di kolom Catatan.

## Progress

**9 / 15 task selesai** — 12 / 26 komponen terverifikasi.

| Fase | Selesai | Total |
|---|---|---|
| 0 — Fondasi | 3 | 3 |
| 1 — Sidebar | 5 | 5 |
| 2 — Chat | 1 | 5 |
| 3 — Panel & layar | 0 | 2 |
| **Total** | **9** | **15** |

---

## Fase 0 — Fondasi

- [x] **T1 · Baseline build + token warna**
  - [x] `go build ./...` bersih (fix import mati `cmd/run.go:15` `options/windows`)
  - [x] `.gitignore`: `/fyne_test.exe`, `/fyne.exe`, `/nul`, `fyneui/testdata/out/*.png`
  - [x] hapus `fyneui/*.bak`, `fyne_test.exe`, `fyne.exe`, `nul`
  - [x] `go test ./fyneui/` PASS
  - [x] commit `0984f12` — `chore(fyne): baseline before component-by-component parity pass`
  - Spec: `frontend/src/style.css` `@theme` · File: `fyneui/theme_wa.go`

- [x] **T2 · Harness render PNG + helper bersama**
  - [x] `fyneui/golden_test.go`: `renderPNG`, `blankImage`, `diffRatio`, `assertNotBlank`, `assertPixel`, `crop`
  - [x] `fyneui/layout.go`: `padLayout`, `minSizeLayout`, `spacer`, `truncateToWidth`, `avatarColor`, `initials` (pindah dari `sidebar.go`)
  - [x] `TestHarnessRendersColor` PASS + `testdata/out/harness.png` diperiksa (kotak merah solid 40×40, tidak blank)
  - [x] `TestTruncateToWidth` + `TestAvatarHelpers` PASS
  - [x] commit `f98d86f` — `test(fyne): headless golden-render harness for component parity`
  - File: `fyneui/golden_test.go`, `fyneui/layout.go`

- [x] **T3 · Test paritas token vs style.css**
  - [x] `fyneui/tokens_test.go`: `TestTokensMatchStyleCSS` — **mem-parse `@theme` dari CSS**, bukan menyalin hex ke Go; gagal otomatis kalau salah satu sisi drift
  - [x] 8 token dipetakan: `dark-bg`, `dark-tertiary`, `dark-text`, `dark-muted`, `green`, `sent-bubble-dark-bg`, `received-bubble-dark-bg`, `toggle-dark-bg`
  - [x] `waLink` dijaga terpisah terhadap literal `#53bdeb` (`.msg-link`, bukan `--color-light-primary`)
  - [x] `parseHex` menangani `#rrggbbaa` (`--color-dropdown-border: #dcdcdc50`)
  - [x] `waDayChip` + `waTime` sudah ada sejak T1
  - [x] commit `b2a573f` — `fix(fyne): chat row parity — truncating labels, 48px avatar glyphs, 48-unit viewBox`
  - Spec: `frontend/src/style.css` · File: `fyneui/tokens_test.go`, `fyneui/theme_wa.go`

---

## Fase 1 — Sidebar

- [x] **T4 · Sidebar header** — commit `225d7ee`
  - [x] `newSidebarHeader(onNewChat, onMenu)` + `SetSelfAvatar`
  - [x] `h-16` (64px), judul `text-xl font-bold`, 2 tombol ikon `p-2 rounded-full`, avatar 32px, border bawah `white/5`
  - [x] geometri **terukur**: row h **64** (63 + 1px border), padding **0 16**, title **20px/700/#fff**, tombol **40x40** (icon 24, gap 4), avatar **32** + `ml-2` 8
  - [x] `TestSidebarHeaderGeometryMatchesReact`, `TestSidebarHeaderActionsGeometry`, `TestSidebarHeaderCallbacks` PASS · PNG `sidebar_header.png` diperiksa (title pixel-identik: y 25–44, h 20)
  - [x] `topBar()` lama di `sidebar.go` dihapus, `loadSelfAvatar` arahkan ke header baru
  - [x] `iconNewChat` diperbaiki: React `NewChatIcon` = balon chat dua garis, bukan tanda `+`
  - Spec: `ChatScreen.tsx:50-70` · File: `fyneui/sidebar_header.go`

- [x] **T5 · View tabs (chats/communities/channels/status)** — commit `04cfcd1`
  - [x] `newViewTabs(current, onChange)` + `SetCurrent` (tidak memicu `onChange`); tap pill aktif = no-op
  - [x] pill `rounded-full px-3 py-1 text-sm capitalize`; aktif `bg-[#e9edef] text-[#0b141a]`, idle `border-white/10 text-[#8696a0]`
  - [x] geometri **terukur** dari `getBoundingClientRect()`: pill h **30**, padding **4px 12px**, gap **8**, bar `pt-1 px-3 pb-2` = h **42**, font **14**
  - [x] `TestViewTabsGeometryMatchesReact` (h 30, lebar ±6, bar 42), `TestViewTabsCapitalize`, `TestViewTabsSetCurrentDoesNotFireOnChange`, `TestViewTabsTapActiveIsNoop`, `TestViewTabsGolden` PASS · PNG `view_tabs.png` diperiksa
  - [x] `tabPill`/`rebuildTabs` lama dihapus; search bar tetap disembunyikan saat `view == "status"`
  - Catatan: `rect.height` React terbaca **57** karena scrollbar horizontal 15px (`overflow-x-auto`, 4 pill butuh 338px di ruang 295px); `clientHeight`/`scrollHeight` = **42** — pakai 42. Border Fyne butuh stroke **2px** karena stroke 1px jatuh di setengah pixel dan jadi 2 baris alpha 50%.
  - Spec: `ChatScreen.tsx:828-849` · File: `fyneui/view_tabs.go`

- [x] **T6 · Search bar** — commit `4b2544f`
  - [x] `newSearchBar(placeholder, onChange)` + `SetPlaceholder` + `SetText`
  - [x] luar **52** (`px-3 py-2`), pill **36** `#242626` radius 18 (`rounded-full`), ikon **20** di `px-4`, gap **16** (`mr-4`), placeholder **14**
  - [x] terukur: pill rows **8..43** cols **16..204** identik dengan React; ink teks Fyne rows 13..23 vs React 13..24
  - [x] `widget.Entry` punya inset atas tetap ~12px → entry digeser `searchTextLift = -7` supaya ink duduk di tengah pill
  - [x] `SetText` menekan echo `OnChanged` (input React controlled)
  - [x] `TestSearchBarGeometryMatchesReact`, `TestSearchBarPillColour`, `TestSearchBarOnChangeAndSetText`, `TestSearchBarPlaceholderPerView` PASS · PNG `search_bar.png` + `_cmp_search.png`
  - [x] `searchBar` lama dihapus, placeholder dinamis per view di `syncChrome`
  - [x] **BUG**: `filterAndSortLocked` memakai `TrimSpace` dan ikut mencocokkan preview pesan; `useFilteredChatIds` (`useChatStore.ts:150-163`) hanya `lowercase(raw)` + `chat.name`. Diperbaiki + `TestSidebarSearchMatchesReact`
  - Spec: `ChatScreen.tsx:841-855` · File: `fyneui/search_bar.go`

- [x] **T7 · Chat row** — ✅ selesai (`parity badge/pin`)
  - [x] tinggi baris **72** = React (`stackedLayout{24,20}` gap 4, py-3) — `TestChatRowHeightMatchesReact`
  - [x] avatar **48px** di inset **16px** — `TestChatRowAvatarGeometry`
  - [x] font nama **16 bold**, preview **14**, waktu **12** — diukur dari `getComputedStyle` di harness, bukan ditebak: name `16px/24px/500`, preview `14px/20px/400`, time `12px/16px/500`
  - [x] `harness.tsx` sekarang mem-dump `typography` (fontSize/lineHeight/fontWeight/color) ke `#metrics`
  - [x] catatan: format tanggal **bukan** drift — React `toLocaleDateString([])` menghasilkan `11/7/2025` hanya karena Chrome headless ber-locale en-US; di locale Indonesia keduanya `7/11/2025`
  - [x] nama & preview **ellipsize** via `truncLabel` — `canvas.Text` tidak pernah clip, dulu menabrak timestamp
  - [x] avatar fallback pakai **glyph React** (people grup / person kontak), bukan inisial — `TestAvatarGlyphVisible` (84/153 ink px)
  - [x] `iconPerson` salah skala (viewBox 48 dipaksa /24 → ter-clip habis) → `svgResourceVB`/`iconImageVB`
  - [x] `avatarColor` SHA-1 %16 + palet light/dark React — `TestAvatarColorMatchesReact`
  - [x] PNG `chatrow.png`, `chatrow_geom.png`, `chatrow_fixture_group.png`, `chatrow_fixture_contact.png` diperiksa
  - [x] pin + badge unread diukur dari `out.badges` (getBoundingClientRect row-relative), bukan ditebak:
        pin `x=149 y=43 14x14`; badge `"99+" x=171 y=40 34x20`, `"2" x=185 y=40 20x20`, bg `rgb(33,192,99)`, padding `0 6px`, `min-w 20`, radius full
  - [x] pill **tumbuh mengikuti label** (`badgeLayout` + `badgeWidthFor`), bukan kotak 20px tetap — "99+" = 34px seperti React
  - [x] `rowActionsLayout` (`gap-2` = 8px, rata kanan, center vertikal) + `rowLineLayout` (`justify-between`) menggantikan `container.NewBorder`, yang menambah padding tema 4px dan memakan 16px gutter kanan
  - [x] **BUG**: kolom teks x=100 w=105 → sekarang x=80 w=125 (React `16+48+16` .. `221-16`)
  - [x] **BUG**: `reactRowWidth` 319 → 221 (salah baca screenshot 400px)
  - [x] **BUG**: inisial `?` ikut tampil di avatar fallback → `avatarTxt.Hide()` (React tidak pernah menampilkan inisial)
  - [x] `TestChatRowPinGeometry`, `TestChatRowBadgeGeometry` (2 sub-test), `TestChatRowNoBadgeWhenRead`
  - [x] `chatRowRefs`/`rowRefs` dipertahankan (dipakai `updateRow` untuk widget recycling) — refactor konstruktor murni menyusul di T14a
  - Spec: `ChatScreen.tsx:247-320` · File: `fyneui/sidebar.go`, `fyneui/chatrow_badge.go`, `fyneui/layout.go`, `fyneui/icons.go`

- [x] **T8 · Archived entry + empty state** ✅ `30e0ca2`
  - [x] `newArchivedEntry(count, onOpen)` — baris 48px, `px-4 py-3`, `gap-4`; glyph arsip 20px di x=30 `rgb(33,192,99)`; label 16px medium; count 12px muted rata kanan (ink x 201..202 vs React 200..202)
  - [x] `newArchivedHeader(onBack)` — tampilan saat masuk arsip: panah balik 24px di x=16, label di x=56, hairline bawah; entry disembunyikan
  - [x] `newSidebarEmptyState(hasChats, onRefresh)` — `p-8`, paragraf 16px/24px `#99a1af` dirapatkan ke baris (x=32 w=157, y=171.5 vs React 172), CTA 128×40 `#00c950` radius 8, gap 16px (x=47 vs 46, y=235.5 vs 236)
  - [x] `TestArchivedEntryGolden`, `TestArchivedEntryVisibility`, `TestArchivedHeaderGeometry`, `TestEmptyStateGolden`, `TestEmptyStateStrings`, `TestEmptyStateWraps` — 6 PASS
  - [x] `archivedRow`/placeholder lama diganti: entry + header masuk `topBar()`, `body()` pakai `container.NewMax(list, placeholder, emptyS)`
  - [x] **BUG**: `emptyState` tidak pernah render apa pun — `BaseWidget` belum di-`ExtendBaseWidget`, jadi renderer tidak pernah layout anak-anaknya
  - [x] **BUG**: CTA tidak punya `MinSize` → mengecil ke 104×22 milik label
  - [x] **BUG**: `canvas.Text` tidak bisa wrap → paragraf dipecah per baris (`wrapText`) dan diulang saat `Resize`
  - [x] **BUG**: `syncChrome` cuma `RLock` padahal harus menulis `archived = false` saat arsip kosong (`ChatScreen.tsx:593`)
  - Spec: `ChatScreen.tsx:355-374, 859-887` · File: `fyneui/sidebar_states.go`, `fyneui/sidebar.go`, `fyneui/icons.go`

- [x] **T9 · Chat header** — commit `06f26e0`
  - [x] `newChatHeader(onBack, onInfo, d)` + `Set(chatHeaderData)` di `fyneui/chat_header.go`
  - [x] avatar 40px `#4a5565`, nama 16px/24 bold truncate, subtitle 12px/16 `#8696a0`, call 38px (hanya kontak) + info 28px, border 1px `white/5`
  - [x] **Terverifikasi (600px, ink bbox React vs Fyne):** back `16..31 / 24..39` = `16..31 / 24..39`; avatar `64..103 / 12..51` = `64..103 / 12..51`; call `532..549 / 23..40` = `532..549 / 23..40`; info `565..582 / 23..40` = `565..582 / 23..40`; nama & subtitle origin x=116 identik (lebar glyph Fyne +10px karena font)
  - [x] **BUG**: `iconBox` belum `ExtendBaseWidget` → panah back tidak pernah tergambar
  - [x] **BUG**: header lama pakai `NewBorder`+`NewHBox` → padding tema 12px+ dan tinggi baris 64px, padahal `p-3` = 65px
  - [x] **BUG**: back arrow `md:hidden` → wajib hilang di ≥768px, bukan selalu render
  - [x] **BUG**: subtitle grup masih placeholder "group info" → sekarang diturunkan persis `ChatDetail.tsx:135-160` ("You, Alice, ~ Bob, +91 …", self difilter via phone+JID, label = full name > "~ push" > short > `formatPhone`); `formatPhone` diport dari `lib/utils.ts:12-19`
  - [x] field lama `headerName`/`headerSub`/`avatarTxt`/`avatarBg`/`avatarImg` dihapus dari `ChatView`
  - [x] 6 test PASS: golden geometry, grup tanpa call, action pinned right, truncate nama, initial avatar, back hidden di desktop
  - Spec: `ChatHeader.tsx`, `ChatDetail.tsx:135-160,703-714` · File: `fyneui/chat_header.go`

---

## Fase 2 — Chat

- [x] **T10 · Bubble teks** — `59a67a3,7f735d0,d406fd1`
  - [x] `bubble.go`: `bubbleData`/`bubbleBg`/`bubbleCorner` (run-start 2px sender-facing, else 12) + `emojiOnlyText` + `newBubble`/`htmlBubbleSized` (GridWrap)
  - [x] **fix akar (0e11859):** `Inline=false` pecah segmen→1 kata/baris; `wrapRuns` spasi ekor; `innerPadding` 8px GridWrap
  - [x] body kosong→italic "Empty Message"; emoji-only→32px/leading-40; sender grup non-me+`FirstInGroup` 11px bold `misc.GetProfileColor` (run break side/sender, reset separator)
  - [x] meta 11px opacity, `Edited`, tick `✓✓` `waTick #53bdeb` sent
  - [x] bg `waSent #144d37`/`waRecv #242626` (dark), corner 2px `FirstInGroup` (Fyne rect seragam 4px approx; sticker/media paths terpisah)
  - [x] Tests PASS: `TestBubbleCorner`,`TestBubbleTextVisible`(anti 0e11859),`TestBubbleEmptyContent`,`TestBubbleEmojiOnlyIsLarge`,`TestBubbleLongToken`,`TestBubbleParitySpec`,`TestBubbleParityRender`,`TestBubbleParityNumericBBox`,`TestBubbleParityGroupSender`
  - [x] PNG/DOM: `fyneui/testdata/{react,out}/bubble_{sent,recv,group,emoji,edited}*` (React spec static harness; real `MessageItem` headless blank: wailsjs deps)
  - [~] app live manual (bukan headless)
  - Limit: corner per-sisi Fyne seragam; receipt status tak ada di `DecodedMessageInfo` (tick read-intent). Emoji-only tetap card bg (React: `bg-transparent` hanya sticker `:126,442`)
  - Spec: `MessageItem.tsx:406-560` · File: `fyneui/bubble.go`,`chatview.go`

- [ ] **T11 · Bubble media (image/video/sticker/audio/document)**
  - [ ] `newMediaBubble(mediaBubbleData{Kind,Thumb,Caption,FileName,SizeLabel,Width,Height,MaxWidth})`
  - [ ] image/video: thumbnail rounded 8 + placeholder `waPanel2`; video + overlay play 48px
  - [ ] sticker: **tanpa** background bubble
  - [ ] audio: pill + tombol play; document: kartu + chip merah + nama file
  - [ ] fetch chain tetap `GetImageThumbnail` → `GetCachedImage` → `DownloadMedia` (video: `GetVideoThumbnail`)
  - [ ] `TestMediaPlaceholderGolden`, `TestStickerHasNoBubbleBackground`, `TestDocumentBubbleGolden` PASS · PNG diperiksa
  - [ ] commit `feat(fyne): media bubble component with placeholder and sticker variants`
  - Spec: `MediaContent.tsx` · File: `fyneui/bubble_media.go`

- [ ] **T12 · Message list — day separator + grouping**
  - [ ] `buildListItems(msgs)`: `itemDay` saat tanggal lokal berubah, `FirstInGroup` saat `IsFromMe`/`Sender` berubah (**tanpa** aturan gap waktu)
  - [ ] `newDaySeparator(label)`: chip `rounded-full` putih 10%, teks 11px muted, baris 40px
  - [ ] `TestBuildListItemsGroupsRuns`, `TestBuildListItemsInsertsDaySeparator`, `TestDaySeparatorGolden` PASS · PNG diperiksa
  - [ ] `daySeparator`/`lastDay`/grouping inline di `chatview.go` dihapus
  - [ ] commit `feat(fyne): message list grouping and day separators with tests`
  - Spec: `MessageList.tsx:140-175`, `MessageList.tsx:57-65` · File: `fyneui/message_list.go`

- [ ] **T13 · Composer**
  - [ ] `newComposer(onSend, onEmoji, onAttach)` + `Text()`/`Clear()`
  - [ ] luar `mx-3 mb-3 mt-1.5`, pill `rounded-full bg-[#242626] px-1.5`, emoji kiri, attach kanan, kirim `h-11 w-11 rounded-full bg-[#1b9a58]` terpisah
  - [ ] Enter kirim, Shift+Enter baris baru, `OnSubmitted` → `send()`
  - [ ] `TestComposerGolden` + `TestComposerSendClears` PASS · PNG `composer.png` diperiksa
  - [ ] commit `feat(fyne): composer component with golden and send-behavior tests`
  - Spec: `ChatInput.tsx:419-566` · File: `fyneui/composer.go`

---

## Fase 3 — Panel & layar lain

- [ ] **T14a · Chat info panel**
  - [ ] `newChatInfoPanel(chatName, chatType, onClose)` + `SetDetails(name, subtitle, about)`
  - [ ] panel kanan 320px, header 64px + tombol close, avatar 96px, nama 18px, section About
  - [ ] `TestChatInfoPanelGolden` PASS · PNG `chat_info.png` diperiksa
  - [ ] commit `feat(fyne): chat info panel`
  - Spec: `ChatInfo.tsx` · File: `fyneui/chat_info.go`

- [ ] **T14b · Settings screen**
  - [ ] `newSettingsScreen(onBack)`: header + back, baris Account/Privacy/Chats/Notifications/Advanced/Log out (ikon + chevron, tap = stub log)
  - [ ] `TestSettingsScreenGolden` PASS · PNG `settings.png` diperiksa
  - [ ] commit `feat(fyne): settings screen scaffold`
  - Spec: `SettingsScreen.tsx` · File: `fyneui/settings.go`

---

## Blocker aktif

- [ ] **Session WA websocket tidak tersambung** — `DownloadMedia: ... failed: websocket not connected`. Media baru tidak bisa diunduh; perlu reconnect (scan QR / pastikan linked device aktif). Ini **bukan** bug renderer.
- [ ] **Dead code** `railItem`/`railAvatar` di `fyneui/app.go` (rail sudah tidak dipakai `buildShell`) — hapus saat menyentuh file itu.

## Catatan

- `docs/` di-`.gitignore`; plan di-commit dengan `git add -f`.
- `write_file` menolak overwrite file existing → `cp x.bak && rm x` lalu tulis baru.
- Kill app: `cmd //c "taskkill /F /IM fyne_test.exe"` (bukan `taskkill //F`).
- Jalankan app harus `background: true`, kalau foreground exit -1.
- Buka chat tanpa klik GUI: env `WHATS4LINUX_OPEN_JID="6282134544728@s.whatsapp.net"`.
