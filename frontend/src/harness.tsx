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

  const pre = document.getElementById("metrics")
  if (pre) pre.textContent = JSON.stringify(out, null, 2)
  ;(window as unknown as { __PARITY__: unknown }).__PARITY__ = out
}

setTimeout(collect, 400)
