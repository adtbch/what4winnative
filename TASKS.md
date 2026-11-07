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

**4 / 15 task selesai** — 0 / 26 komponen terverifikasi.

| Fase | Selesai | Total |
|---|---|---|
| 0 — Fondasi | 4 | 4 |
| 1 — Sidebar | 0 | 6 |
| 2 — Chat | 0 | 4 |
| 3 — Panel & layar | 0 | 2 |
| **Total** | **3** | **15** |

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

- [ ] **T4 · Sidebar header**
  - [ ] `newSidebarHeader(onNewChat, onMenu)` + `SetSelfAvatar`
  - [ ] `h-16` (64px), judul `text-xl font-bold`, 2 tombol ikon `p-2 rounded-full`, avatar 32px, border bawah `white/5`
  - [ ] `TestSidebarHeaderGolden` PASS · PNG `sidebar_header.png` diperiksa
  - [ ] `topBar()` lama di `sidebar.go` dihapus, `loadSelfAvatar` arahkan ke header baru
  - [ ] commit `feat(fyne): sidebar header component with golden test`
  - Spec: `ChatScreen.tsx:50-70` · File: `fyneui/sidebar_header.go`

- [x] **T5 · View tabs (chats/communities/channels/status)** — commit `04cfcd1`
  - [x] `newViewTabs(current, onChange)` + `SetCurrent` (tidak memicu `onChange`); tap pill aktif = no-op
  - [x] pill `rounded-full px-3 py-1 text-sm capitalize`; aktif `bg-[#e9edef] text-[#0b141a]`, idle `border-white/10 text-[#8696a0]`
  - [x] geometri **terukur** dari `getBoundingClientRect()`: pill h **30**, padding **4px 12px**, gap **8**, bar `pt-1 px-3 pb-2` = h **42**, font **14**
  - [x] `TestViewTabsGeometryMatchesReact` (h 30, lebar ±6, bar 42), `TestViewTabsCapitalize`, `TestViewTabsSetCurrentDoesNotFireOnChange`, `TestViewTabsTapActiveIsNoop`, `TestViewTabsGolden` PASS · PNG `view_tabs.png` diperiksa
  - [x] `tabPill`/`rebuildTabs` lama dihapus; search bar tetap disembunyikan saat `view == "status"`
  - Catatan: `rect.height` React terbaca **57** karena scrollbar horizontal 15px (`overflow-x-auto`, 4 pill butuh 338px di ruang 295px); `clientHeight`/`scrollHeight` = **42** — pakai 42. Border Fyne butuh stroke **2px** karena stroke 1px jatuh di setengah pixel dan jadi 2 baris alpha 50%.
  - Spec: `ChatScreen.tsx:828-849` · File: `fyneui/view_tabs.go`

- [ ] **T6 · Search bar**
  - [ ] `newSearchBar(placeholder, onChange)` + `SetPlaceholder`
  - [ ] luar `px-3 py-2`, dalam `rounded-full bg-[#242626] px-4 py-2`, ikon `mr-4`, input `text-sm`
  - [ ] `TestSearchBarGolden` PASS · PNG `search_bar.png` diperiksa (pill penuh bulat, tanpa kotak Entry)
  - [ ] `searchBar` lama dihapus, placeholder dinamis per view
  - [ ] commit `feat(fyne): search bar component with golden test`
  - Spec: `ChatScreen.tsx:84-95` · File: `fyneui/search_bar.go`

- [ ] **T7 · Chat row** — 🟡 sebagian (geometri + teks + avatar terverifikasi)
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
  - [ ] pin + badge unread belum dibandingkan dengan PNG React (fixture belum punya unread/pin terlihat)
  - [ ] `chatRowRefs`/`rowRefs` belum dihapus — refactor konstruktor murni menyusul
  - Spec: `ChatScreen.tsx:247-320` · File: `fyneui/sidebar.go`, `fyneui/layout.go`, `fyneui/icons.go`

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
  - [ ] **fix akar masalah:** tinggi bubble diambil dari return `htmlRichText`, **bukan** `box.MinSize()` (yang = 0)
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
