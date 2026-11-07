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
import type { ChatItem } from "./store/types"

const params = new URLSearchParams(location.search)
const scene = params.get("scene") || "chatrow"
const theme = (params.get("theme") || "dark") as "light" | "dark"

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
useChatStore.getState().setChats(FIXTURE)

function Scenes() {
  if (scene === "chatrow") return <ChatListScreen onOpenSettings={() => {}} />
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

  const pre = document.getElementById("metrics")
  if (pre) pre.textContent = JSON.stringify(out, null, 2)
  ;(window as unknown as { __PARITY__: unknown }).__PARITY__ = out
}

setTimeout(collect, 400)
