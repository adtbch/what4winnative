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

**0 / 14 task selesai** — 0 / 26 komponen terverifikasi.

---

## Fase 0 — Fondasi

- [ ] **T1 · Baseline build + token warna**
  - [ ] `go build ./...` bersih (fix import mati `cmd/run.go:15` `options/windows`)
  - [ ] `.gitignore`: `/fyne_test.exe`, `/fyne.exe`, `/nul`, `fyneui/testdata/out/*.png`
  - [ ] hapus `fyneui/*.bak`, `fyne_test.exe`, `fyne.exe`, `nul`
  - [ ] `go test ./fyneui/` PASS
  - [ ] commit `chore(fyne): baseline before component-by-component parity pass`
  - Spec: `frontend/src/style.css` `@theme` · File: `fyneui/theme_wa.go`

- [ ] **T2 · Harness render PNG + helper bersama**
  - [ ] `fyneui/golden_test.go`: `renderPNG`, `blankImage`, `diffRatio`, `assertNotBlank`, `assertPixel`, `crop`
  - [ ] `fyneui/layout.go`: `padLayout`, `truncateToWidth`, `avatarColor`, `initials` (pindah dari `sidebar.go`)
  - [ ] `TestHarnessRendersColor` PASS + `testdata/out/harness.png` diperiksa
  - [ ] commit `test(fyne): headless golden-render harness for component parity`
  - File: `fyneui/golden_test.go`, `fyneui/layout.go`

- [ ] **T3 · Test paritas token vs style.css**
  - [ ] `fyneui/tokens_test.go`: `TestTokensMatchStyleCSS` (dark-text, dark-muted, dark-bg, dark-tertiary, green, sent-bubble-dark-bg, received-bubble-dark-bg)
  - [ ] token melenceng diperbaiki di `theme_wa.go`
  - [ ] tambah `waDayChip` (putih 10%) + `waTime`
  - [ ] PASS + commit `test(fyne): assert palette tokens match style.css @theme`
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

- [ ] **T5 · View tabs (chats/communities/channels/status)**
  - [ ] `newViewTabs(current, onChange)` + `SetCurrent` (tidak memicu `onChange`)
  - [ ] pill `rounded-full px-3 py-1 text-sm capitalize`; aktif `bg-[#e9edef] text-[#0b141a]`, idle `border-white/10 text-[#8696a0]`
  - [ ] `TestViewTabsGolden` PASS · PNG `view_tabs.png` diperiksa
  - [ ] `tabPill`/`rebuildTabs` lama dihapus; search bar tetap disembunyikan saat `view == "status"`
  - [ ] commit `feat(fyne): view tabs component with golden test`
  - Spec: `ChatScreen.tsx:828-849` · File: `fyneui/view_tabs.go`

- [ ] **T6 · Search bar**
  - [ ] `newSearchBar(placeholder, onChange)` + `SetPlaceholder`
  - [ ] luar `px-3 py-2`, dalam `rounded-full bg-[#242626] px-4 py-2`, ikon `mr-4`, input `text-sm`
  - [ ] `TestSearchBarGolden` PASS · PNG `search_bar.png` diperiksa (pill penuh bulat, tanpa kotak Entry)
  - [ ] `searchBar` lama dihapus, placeholder dinamis per view
  - [ ] commit `feat(fyne): search bar component with golden test`
  - Spec: `ChatScreen.tsx:84-95` · File: `fyneui/search_bar.go`

- [ ] **T7 · Chat row**
  - [ ] `chatRowData{JID,Name,Preview,Time,Pinned,Unread,Avatar,AvatarBg}` + `newChatRow()`/`Set()`
  - [ ] avatar `w-12 h-12` + `mr-4`, nama `font-medium` truncate, preview `text-sm` muted, waktu `text-xs` (hijau `#21c063` kalau unread), pin, badge `min-w-5 h-5 rounded-full bg-[#21c063]`
  - [ ] `truncateToWidth` untuk nama & preview (canvas.Text tidak ellipsize sendiri)
  - [ ] `TestChatRowGolden` + `TestChatRowTruncatesLongName` PASS · PNG `chat_row.png`, `chat_row_long.png` diperiksa
  - [ ] `chatRowRefs`/`rowRefs` dihapus; app render 300+ baris tanpa panic
  - [ ] commit `feat(fyne): chat row component with truncation and golden tests`
  - Spec: `ChatScreen.tsx:247-320` · File: `fyneui/chat_row.go`

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
