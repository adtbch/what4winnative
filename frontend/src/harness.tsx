// Parity harness: renders the REAL React components with fixture data and no
// backend, then exposes geometry metrics so the Go golden tests can assert
// against measured truth instead of hand-guessed numbers.
//
// Run:  npx vite --config vite.harness.config.ts
// Open: http://localhost:5199/harness.html?scene=chatrow&theme=dark
import { createRoot } from "react-dom/client"
import "./style.css"
import { applyThemeClass } from "./lib/theme"
import { useChatStore } from "./store/useChatStore"
import { useAppSettingsStore } from "./store/useAppSettingsStore"
import { ChatListScreen } from "./screens/ChatScreen"
import { ChatHeader } from "./components/chat/ChatHeader"
import type { ChatItem } from "./store/types"

const params = new URLSearchParams(location.search)
const scene = params.get("scene") || "chatrow"
const theme = (params.get("theme") || "dark") as "light" | "dark"

// ARCHIVED adds one archived chat so the "Archived" entry row renders; the
// empty scene drops every chat so EmptyState renders.
const FIXTURE_ARCHIVED: ChatItem[] = [
  {
    id: "628111222333@s.whatsapp.net",
    name: "Grup Lama",
    subtitle: "udah selesai",
    type: "group",
    timestamp: 1762000000,
    archived: true,
  },
]

const FIXTURE: ChatItem[] = [
  {
    id: "6282134544728@s.whatsapp.net",
    name: "cintakuuu",
    subtitle: "iyaa aku juga sayang kamu",
    type: "contact",
    timestamp: 1762500000,
    unreadCount: 2,
  },
  {
    id: "120363999999999999@g.us",
    name: "Vapers Semarangan 18+",
    subtitle: "besok kumpul jam 8",
    type: "group",
    timestamp: 1762490000,
    sender: "Budi",
    pinned: true,
    unreadCount: 128,
  },
  {
    id: "628999888777@s.whatsapp.net",
    name: "+62 899-9888-777",
    subtitle: "📷 Photo",
    type: "contact",
    timestamp: 1762400000,
  },
]

applyThemeClass(theme)
useAppSettingsStore.setState({ theme, loaded: true })
useChatStore.getState().setChats(
  scene === "empty" ? [] : scene === "archived" ? [...FIXTURE, ...FIXTURE_ARCHIVED] : FIXTURE,
)

const HEADER_SCENES = new Set(["chatheader", "chatheader-group"])
const BUBBLE_SCENES = new Set(["bubble-sent","bubble-recv","bubble-group","bubble-emoji","bubble-edited"])

function Scenes() {
  if (scene === "chatheader" || scene === "chatheader-group") {
    const group = scene === "chatheader-group"
    return (
      <div className="w-full bg-light-bg dark:bg-[#0b141a]">
        <ChatHeader
          chatName={group ? "Vapers Semarangan 18+" : "cintakuuu"}
          chatSubtitle={group ? "You, Budi, ~ Sari, +62 812-3456-7890" : "online"}
          onBack={() => {}}
          onInfoClick={() => {}}
          onCallClick={group ? undefined : () => {}}
        />
      </div>
    )
  }
  if (BUBBLE_SCENES.has(scene)) {
    const fromMe = scene === "bubble-sent" || scene === "bubble-emoji" || scene === "bubble-edited"
    const group = scene === "bubble-group"
    const isSticker = false, firstInGroup = true, isFromMe = fromMe, isGroup = group
    const emojiOnly = scene === "bubble-emoji", edited = scene === "bubble-edited"
    const content = emojiOnly ? "😂😂😂" : "halo bro apa kabar hari ini"
    const clsx = (...a: any[]) => a.flatMap((x: any) => typeof x === 'string' ? x : x ? Object.keys(x).filter(k => x[k]) : []).join(' ')
    return (
      <div className="w-[520px] bg-[#0b141a] p-4" data-bubble-scene={scene}>
        <div className={clsx("flex group", isFromMe ? "justify-end" : "justify-start")}>
          {!isFromMe && isGroup && <div className="w-7 ml-3 shrink-0" />}
          <div className={clsx("max-w-[85%] rounded-xl px-2 pt-1 pb-1.5 relative min-w-0",
            !isFromMe && isGroup ? "ml-2 mr-5" : "mx-5",
            { "rounded-tl-sm": firstInGroup && !isFromMe && !isSticker, "rounded-tr-sm": firstInGroup && isFromMe && !isSticker,
              "bg-[#144d37]": isFromMe && !isSticker, "bg-[#242626]": !isFromMe && !isSticker })} data-bubble-box>
            {!isFromMe && isGroup && firstInGroup && <span className="text-xs font-bold mb-1 block" style={{color:'#53bdeb'}} data-bubble-sender>Budi</span>}
            {emojiOnly ? <span className="text-3xl leading-10" data-bubble-body>{content}</span> : <span className="text-sm" data-bubble-body>{content}</span>}
            <div className="flex justify-end items-center gap-1 mt-0.5" data-bubble-meta>
              {edited && <span className="text-[11px] opacity-60 mr-1">Edited</span>}
              <span className="text-[11px] opacity-60">10:05</span>
              {isFromMe && <span className="text-[11px]" style={{color:'#53bdeb'}} data-bubble-tick>✓✓</span>}
            </div>
          </div>
        </div>
      </div>
    )
  }
  return <ChatListScreen onOpenSettings={() => {}} />
}

createRoot(document.getElementById("root")!).render(<Scenes />)

// ---- metrics ---------------------------------------------------------------
// Emit measured geometry for every chat row so the Go side asserts real numbers.
function collect() {
  const out: Record<string, unknown> = {
    scene,
    theme,
    viewport: { w: innerWidth, h: innerHeight },
    rows: [] as unknown[],
  }

  const root = document.querySelector("#root > div")
  if (root) {
    const cs = getComputedStyle(root)
    out.root = {
      bg: cs.backgroundColor,
      color: cs.color,
      font: cs.fontFamily,
      fontSize: cs.fontSize,
    }
  }

  document.querySelectorAll("#root [data-parity-row]").forEach(el => {
    const r = el.getBoundingClientRect()
    const cs = getComputedStyle(el)
    out.rows.push({
      h: Math.round(r.height),
      w: Math.round(r.width),
      padding: cs.padding,
      bg: cs.backgroundColor,
    })
  })

  // Fallback: measure every direct row-like child of the scroll container.
  if ((out.rows as unknown[]).length === 0) {
    const list = document.querySelector("#root .overflow-y-auto")
    if (list) {
      Array.from(list.children).forEach(el => {
        const r = el.getBoundingClientRect()
        if (r.height > 20) {
          const cs = getComputedStyle(el)
          ;(out.rows as unknown[]).push({
            h: Math.round(r.height),
            w: Math.round(r.width),
            padding: cs.padding,
            bg: cs.backgroundColor,
            text: (el.textContent || "").slice(0, 40),
          })
        }
      })
    }
  }

  // Typography of the first row, measured from the live DOM. The name is the
  // <h3>, the preview is the sibling div that carries the subtitle HTML, and the
  // timestamp is the span inside the baseline row.
  const firstRow = document.querySelector("#root .overflow-y-auto > div")
  if (firstRow) {
    const pick = (sel: string) => {
      const el = firstRow.querySelector(sel)
      if (!el) return null
      const cs = getComputedStyle(el)
      return {
        fontSize: cs.fontSize,
        lineHeight: cs.lineHeight,
        fontWeight: cs.fontWeight,
        color: cs.color,
      }
    }
    out.typography = {
      name: pick("h3"),
      preview: pick(".text-sm"),
      time: pick(".text-xs"),
    }
  }

  // View tabs (chats/communities/channels/status): measured, not guessed.
  const tabButtons = Array.from(
    document.querySelectorAll("#root .overflow-x-auto button"),
  ).filter(b => /^(chats|communities|channels|status)$/i.test((b.textContent || "").trim()))
  if (tabButtons.length) {
    const box = tabButtons[0].parentElement!
    const br = box.getBoundingClientRect()
    const bs = getComputedStyle(box)
    out.tabs = {
      container: {
        h: Math.round(br.height),
        padding: bs.padding,
        gap: bs.gap,
        boxSizing: bs.boxSizing,
        display: bs.display,
        flexWrap: bs.flexWrap,
        scrollH: box.scrollHeight,
        clientH: box.clientHeight,
        offsetH: (box as HTMLElement).offsetHeight,
        childCount: box.children.length,
      },
      pills: tabButtons.map(b => {
        const r = b.getBoundingClientRect()
        const cs = getComputedStyle(b)
        return {
          label: (b.textContent || "").trim(),
          h: Math.round(r.height),
          w: Math.round(r.width),
          padding: cs.padding,
          radius: cs.borderRadius,
          fontSize: cs.fontSize,
          fontWeight: cs.fontWeight,
          borderWidth: cs.borderTopWidth,
          borderColor: cs.borderTopColor,
          bg: cs.backgroundColor,
          color: cs.color,
        }
      }),
    }
  }

  // Sidebar header: h-16 row with "WhatsApp", two icon buttons, self avatar.
  const h1 = document.querySelector("#root h1")
  if (h1) {
    const head = h1.parentElement as HTMLElement
    const hr = head.getBoundingClientRect()
    const hs = getComputedStyle(head)
    const cs = getComputedStyle(h1)
    const btns = Array.from(head.querySelectorAll("button"))
    const av = head.querySelector("button ~ div") as HTMLElement | null
    out.header = {
      box: {
        h: Math.round(hr.height),
        w: Math.round(hr.width),
        padding: hs.padding,
        borderBottom: hs.borderBottomWidth + " " + hs.borderBottomColor,
        justify: hs.justifyContent,
        align: hs.alignItems,
      },
      title: {
        text: (h1.textContent || "").trim(),
        fontSize: cs.fontSize,
        fontWeight: cs.fontWeight,
        lineHeight: cs.lineHeight,
        color: cs.color,
      },
      buttons: btns.map(b => {
        const r = b.getBoundingClientRect()
        const bs = getComputedStyle(b)
        const svg = b.querySelector("svg")
        const sr = svg?.getBoundingClientRect()
        return {
          title: b.getAttribute("title"),
          x: Math.round(r.x),
          y: Math.round(r.y),
          w: Math.round(r.width),
          h: Math.round(r.height),
          padding: bs.padding,
          radius: bs.borderRadius,
          icon: sr ? Math.round(sr.width) : null,
        }
      }),
      avatar: av
        ? (() => {
            const r = av.getBoundingClientRect()
            const as = getComputedStyle(av)
            return {
              x: Math.round(r.x),
              w: Math.round(r.width),
              h: Math.round(r.height),
              radius: as.borderRadius,
              bg: as.backgroundColor,
              marginLeft: as.marginLeft,
            }
          })()
        : null,
      gap: hs.gap,
    }
  }

  // Search bar: outer px-3 py-2, inner rounded-full pill with icon + input.
  const searchInput = document.querySelector(
    '#root input[placeholder="Search or start new chat"]',
  ) as HTMLInputElement | null
  if (searchInput) {
    const pill = searchInput.parentElement as HTMLElement
    const outer = pill.parentElement as HTMLElement
    const icon = pill.firstElementChild as HTMLElement
    const or = outer.getBoundingClientRect()
    const pr = pill.getBoundingClientRect()
    const ir = icon.getBoundingClientRect()
    const os = getComputedStyle(outer)
    const ps = getComputedStyle(pill)
    const is_ = getComputedStyle(icon)
    const ins = getComputedStyle(searchInput)
    const svg = icon.querySelector("svg")
    const sr = svg?.getBoundingClientRect()
    out.search = {
      outer: {
        h: Math.round(or.height),
        w: Math.round(or.width),
        padding: os.padding,
        bg: os.backgroundColor,
      },
      pill: {
        x: Math.round(pr.x),
        y: Math.round(pr.y),
        w: Math.round(pr.width),
        h: Math.round(pr.height),
        padding: ps.padding,
        radius: ps.borderRadius,
        bg: ps.backgroundColor,
        gap: ps.gap,
      },
      icon: {
        x: Math.round(ir.x),
        w: Math.round(ir.width),
        h: Math.round(ir.height),
        marginRight: is_.marginRight,
        color: is_.color,
        svg: sr ? Math.round(sr.width) : null,
      },
      input: {
        fontSize: ins.fontSize,
        lineHeight: ins.lineHeight,
        color: ins.color,
        placeholderColor: ins.getPropertyValue("--tw-placeholder-color"),
      },
    }
  }

  // Pin icon + unread badge inside the preview row: the last two children of
  // `flex items-center gap-2`. Measured so the Fyne row can place them exactly.
  // The pin is the only 14x14 svg in the row; the badge is the only span with a
  // #21c063 background.
  out.badges = []
  document.querySelectorAll("#root .overflow-y-auto > div").forEach(rowEl => {
    const rowRect = rowEl.getBoundingClientRect()
    const p = rowEl.querySelector('svg[width="14"][height="14"]') as SVGElement | null
    const b = Array.from(rowEl.querySelectorAll("span")).find(el => {
      const cs = getComputedStyle(el)
      return cs.backgroundColor === "rgb(33, 192, 99)"
    }) as HTMLElement | undefined
    if (!p && !b) return
    const info = (el: Element) => {
      const r = el.getBoundingClientRect()
      const cs = getComputedStyle(el)
      return {
        x: Math.round(r.x),
        y: Math.round(r.y - rowRect.y),
        w: Math.round(r.width),
        h: Math.round(r.height),
        padding: cs.padding,
        radius: cs.borderRadius,
        bg: cs.backgroundColor,
        color: cs.color,
        fontSize: cs.fontSize,
        fontWeight: cs.fontWeight,
        lineHeight: cs.lineHeight,
        minWidth: cs.minWidth,
      }
    }
    ;(out.badges as unknown[]).push({
      row: (rowEl.textContent || "").slice(0, 20),
      pin: p ? info(p) : null,
      badge: b ? info(b) : null,
    })
  })

  // Archived entry: the only button in the list area whose label is "Archived".
  const archBtn = Array.from(
    document.querySelectorAll("#root button"),
  ).find(b => (b.textContent || "").trim().startsWith("Archived")) as HTMLElement | undefined
  if (archBtn) {
    const r = archBtn.getBoundingClientRect()
    const cs = getComputedStyle(archBtn)
    const icon = archBtn.querySelector("svg")
    const ir = icon?.getBoundingClientRect()
    const label = Array.from(archBtn.querySelectorAll("span")).find(
      el => (el.textContent || "").trim() === "Archived",
    ) as HTMLElement | undefined
    const count = archBtn.lastElementChild as HTMLElement | null
    const ls = label ? getComputedStyle(label) : null
    const ccs = count ? getComputedStyle(count) : null
    out.archived = {
      box: {
        x: Math.round(r.x),
        y: Math.round(r.y),
        w: Math.round(r.width),
        h: Math.round(r.height),
        padding: cs.padding,
        gap: cs.gap,
      },
      icon: icon
        ? {
            x: Math.round(ir!.x - r.x),
            y: Math.round(ir!.y - r.y),
            w: Math.round(ir!.width),
            h: Math.round(ir!.height),
            color: getComputedStyle(icon!).color,
          }
        : null,
      label: ls
        ? { fontSize: ls.fontSize, fontWeight: ls.fontWeight, color: ls.color }
        : null,
      count: ccs
        ? {
            text: (count!.textContent || "").trim(),
            x: Math.round(count!.getBoundingClientRect().x - r.x),
            fontSize: ccs.fontSize,
            color: ccs.color,
          }
        : null,
    }
  }

  // Empty state: centered column with a 120px chat glyph, the message, and the
  // refresh button.
  const emptyP = Array.from(document.querySelectorAll("#root p")).find(p =>
    /No chats/i.test(p.textContent || ""),
  ) as HTMLElement | undefined
  if (emptyP) {
    const wrap = emptyP.parentElement as HTMLElement
    const wr = wrap.getBoundingClientRect()
    const ws = getComputedStyle(wrap)
    const pr = emptyP.getBoundingClientRect()
    const ps = getComputedStyle(emptyP)
    const btn = wrap.querySelector("button") as HTMLElement | null
    const br = btn?.getBoundingClientRect()
    const bs = btn ? getComputedStyle(btn) : null
    const svg = wrap.querySelector("svg")
    out.empty = {
      box: {
        x: Math.round(wr.x),
        y: Math.round(wr.y),
        w: Math.round(wr.width),
        h: Math.round(wr.height),
        padding: ws.padding,
        align: ws.alignItems,
        justify: ws.justifyContent,
      },
      text: {
        text: (emptyP.textContent || "").trim(),
        fontSize: ps.fontSize,
        lineHeight: ps.lineHeight,
        color: ps.color,
        x: Math.round(pr.x),
        y: Math.round(pr.y),
        w: Math.round(pr.width),
        h: Math.round(pr.height),
      },
      button: bs
        ? {
            text: (btn!.textContent || "").trim(),
            x: Math.round(br!.x),
            y: Math.round(br!.y),
            w: Math.round(br!.width),
            h: Math.round(br!.height),
            padding: bs.padding,
            radius: bs.borderRadius,
            bg: bs.backgroundColor,
            color: bs.color,
            fontSize: bs.fontSize,
            fontWeight: bs.fontWeight,
            marginTop: bs.marginTop,
          }
        : null,
      icon: svg
        ? {
            w: Math.round(svg.getBoundingClientRect().width),
            color: getComputedStyle(svg).color,
          }
        : null,
    }
  }

  // Chat header (ChatHeader.tsx): 12px padded row, 40px avatar, 16px name,
  // 12px subtitle, and the call/info buttons.
  if (HEADER_SCENES.has(scene)) {
    const h2 = document.querySelector("#root h2") as HTMLElement | null
    if (h2) {
      const head = h2.closest("div.border-b") as HTMLElement
      const hr = head.getBoundingClientRect()
      const hs = getComputedStyle(head)
      const av = head.querySelector("div.w-10") as HTMLElement | null
      const avWrap = av?.parentElement as HTMLElement | null
      const ar = av?.getBoundingClientRect()
      const as_ = av ? getComputedStyle(av) : null
      const nr = h2.getBoundingClientRect()
      const ns = getComputedStyle(h2)
      const subEl = h2.nextElementSibling as HTMLElement | null
      const sr = subEl?.getBoundingClientRect()
      const ss = subEl ? getComputedStyle(subEl) : null
      const btns = Array.from(head.querySelectorAll("button")).map(b => {
        const r = b.getBoundingClientRect()
        const bs = getComputedStyle(b)
        const svg = b.querySelector("svg")
        const gr = svg?.getBoundingClientRect()
        return {
          label: b.getAttribute("aria-label"),
          x: Math.round(r.x),
          y: Math.round(r.y - hr.y),
          w: Math.round(r.width),
          h: Math.round(r.height),
          padding: bs.padding,
          radius: bs.borderRadius,
          color: bs.color,
          icon: gr ? Math.round(gr.width) : null,
          iconX: gr ? Math.round(gr.x - r.x) : null,
        }
      })
      out.chatHeader = {
        box: {
          x: Math.round(hr.x),
          y: Math.round(hr.y),
          w: Math.round(hr.width),
          h: Math.round(hr.height),
          padding: hs.padding,
          bg: hs.backgroundColor,
          borderBottom: hs.borderBottomWidth + " " + hs.borderBottomColor,
          gap: hs.gap,
          justify: hs.justifyContent,
        },
        avatar: ar
          ? {
              x: Math.round(ar.x - hr.x),
              y: Math.round(ar.y - hr.y),
              w: Math.round(ar.width),
              h: Math.round(ar.height),
              radius: as_!.borderRadius,
              bg: as_.backgroundColor,
              gap: getComputedStyle(avWrap!).gap,
            }
          : null,
        name: {
          text: (h2.textContent || "").trim(),
          x: Math.round(nr.x - hr.x),
          y: Math.round(nr.y - hr.y),
          w: Math.round(nr.width),
          h: Math.round(nr.height),
          fontSize: ns.fontSize,
          fontWeight: ns.fontWeight,
          lineHeight: ns.lineHeight,
          color: ns.color,
        },
        subtitle: subEl
          ? {
              text: (subEl.textContent || "").trim(),
              x: Math.round(sr!.x - hr.x),
              y: Math.round(sr!.y - hr.y),
              w: Math.round(sr!.width),
              h: Math.round(sr!.height),
              fontSize: ss!.fontSize,
              lineHeight: ss!.lineHeight,
              color: ss!.color,
            }
          : null,
        buttons: btns,
      }
    }
  }

  if (BUBBLE_SCENES.has(scene)) {
    const root = document.querySelector('#root [data-bubble-scene]') as HTMLElement | null
    const box = root?.querySelector('[data-bubble-box]') as HTMLElement | null
    if (root && box) {
      const rr = root.getBoundingClientRect(); const br = box.getBoundingClientRect(); const cs = getComputedStyle(box)
      const meta = box.querySelector('[data-bubble-meta]') as HTMLElement | null
      const sender = box.querySelector('[data-bubble-sender]') as HTMLElement | null
      const body = box.querySelector('[data-bubble-body]') as HTMLElement | null
      const tick = box.querySelector('[data-bubble-tick]') as HTMLElement | null
      out.bubble = {
        scene,
        box: { x: Math.round(br.x-rr.x), y: Math.round(br.y-rr.y), w: Math.round(br.width), h: Math.round(br.height), bg: cs.backgroundColor, borderRadius: cs.borderRadius, borderTopLeftRadius: cs.borderTopLeftRadius, borderTopRightRadius: cs.borderTopRightRadius, padding: cs.padding },
        body: body ? { fontSize: getComputedStyle(body).fontSize, lineHeight: getComputedStyle(body).lineHeight } : null,
        meta: meta ? { fontSize: getComputedStyle(meta.querySelector('span')||meta).fontSize, text: (meta.textContent||'').trim() } : null,
        sender: sender ? { text: (sender.textContent||'').trim(), color: getComputedStyle(sender).color, fontSize: getComputedStyle(sender).fontSize } : null,
        tick: tick ? { color: getComputedStyle(tick).color } : null,
        fromMe: /bubble-sent|bubble-emoji|bubble-edited/.test(scene),
      }
    }
  }

  const pre = document.getElementById("metrics")
  if (pre) pre.textContent = JSON.stringify(out, null, 2)
  ;(window as unknown as { __PARITY__: unknown }).__PARITY__ = out
}

setTimeout(collect, 400)
