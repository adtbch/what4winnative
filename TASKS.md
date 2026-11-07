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

**7 / 15 task selesai** — 8 / 26 komponen terverifikasi.

| Fase | Selesai | Total |
|---|---|---|
| 0 — Fondasi | 3 | 3 |
| 1 — Sidebar | 4 | 5 |
| 2 — Chat | 0 | 5 |
| 3 — Panel & layar | 0 | 2 |
| **Total** | **7** | **15** |

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

- [ ] **T8 · Archived entry + empty state**
  - [ ] `newArchivedEntry(count, onOpen)`: ikon arsip hijau `#21c063`, label "Archived", jumlah `text-xs` muted
  - [ ] `newSidebarEmptyState(hasChats, onRetry)`: "No chats match your search." / "No chats available. Start a conversation!" + tombol "Refresh Chats"
  - [ ] `TestArchivedEntryGolden` + `TestEmptyStateGolden` PASS · PNG diperiksa
  - [ ] `archivedRow`/placeholder lama diganti
  - [ ] commit `feat(fyne): archived entry and empty state components`
  - Spec: `ChatScreen.tsx:362-375, 861-874` · File: `fyneui/sidebar_states.go`

- [ ] **T9 · Chat header**
  - [ ] `newChatHeader(onBack, onInfo, onCall)` + `Set(chatHeaderData)`
  - [ ] avatar 40px, nama `text-[16px] font-medium truncate`, subtitle `text-xs` muted, tombol call (hanya non-group) + info, border bawah
  - [ ] `TestChatHeaderGolden` + `TestChatHeaderGroupHidesCall` PASS · PNG diperiksa
  - [ ] `headerName`/`headerSub`/`avatarTxt`/`avatarBg`/`avatarImg` di `ChatView` dihapus
  - [ ] commit `feat(fyne): chat header component with golden tests`
  - Spec: `ChatHeader.tsx` · File: `fyneui/chat_header.go`

---

## Fase 2 — Chat

- [ ] **T10 · Bubble teks — FIX BUG bubble kosong** ⚠️ prioritas
  - [ ] `newBubble(bubbleData{HTML,FromMe,Edited,Time,Sender,FirstInGroup,EmojiOnly,MaxWidth})` + `bubbleCorner()`
  - [x] **fix akar masalah (commit `0e11859`):** bukan `MinSize()==0`, tapi `widget.TextSegment.Style.Inline=false` — Fyne memecah paragraf di tiap segmen non-inline, jadi 1 kalimat jadi 1 kata per baris (tinggi 127px untuk 1 baris 19px). Plus `wrapRuns` mengukur kata tanpa spasi ekor (baris 532px terukur 559px), plus `textBlock` tidak menghitung `innerPadding` 8px di sel GridWrap
  - [ ] body kosong → italic "Empty Message"; emoji-only → 32px; sender label grup hanya saat `FirstInGroup`
  - [ ] meta `text-[11px] opacity-60`, tick `✓✓` biru `#53bdeb` kalau `FromMe`
  - [ ] background `waSent`/`waRecv`, corner 2px saat `FirstInGroup`
  - [ ] `TestBubbleTextVisible`, `TestBubbleEmptyContent`, `TestBubbleLongURLWraps`, `TestBubbleEmojiOnlyIsLarge`, `TestBubbleCorner` PASS
  - [ ] PNG `bubble_text.png`, `bubble_empty.png`, `bubble_long_url.png` diperiksa
  - [ ] app live: bubble yang tadinya kosong sekarang berisi teks
  - [ ] commit `fix(fyne): bubble text rendered blank; add bubble component + golden tests`
  - Spec: `MessageItem.tsx:406-560` · File: `fyneui/bubble.go`

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
