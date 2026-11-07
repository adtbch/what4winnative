// H.264 media plumbing for WhatsApp video calls.
//
// meowcaller carries encoded H.264 only — it never touches pixels. WebView2 (Edge
// 94+) ships WebCodecs, so the browser side does both directions:
//   camera -> VideoEncoder (annexb) -> base64 -> Go -> WhatsApp relay
//   WhatsApp relay -> Go -> base64 -> VideoDecoder (annexb) -> <canvas>
//
// Access units arrive as Annex-B: start-code prefixed NAL units, one frame per
// access unit. That is exactly what VideoEncoder/VideoDecoder expect with
// avc.format = "annexb", so no bitstream rewriting is needed.

const ENCODER_CODEC = "avc1.42E01E" // H.264 baseline, level 3.0 — WhatsApp's floor
const ENCODER_BITRATE = 600_000
const ENCODER_FRAMERATE = 24

export function videoCodecsSupported(): boolean {
  return (
    typeof window !== "undefined" &&
    typeof (window as any).VideoEncoder === "function" &&
    typeof (window as any).VideoDecoder === "function"
  )
}

function bytesToBase64(bytes: Uint8Array): string {
  let binary = ""
  const chunk = 0x8000
  for (let i = 0; i < bytes.length; i += chunk) {
    binary += String.fromCharCode.apply(
      null,
      bytes.subarray(i, i + chunk) as unknown as number[],
    )
  }
  return btoa(binary)
}

function base64ToBytes(b64: string): Uint8Array {
  const binary = atob(b64)
  const out = new Uint8Array(binary.length)
  for (let i = 0; i < binary.length; i++) out[i] = binary.charCodeAt(i)
  return out
}

// Annex-B NAL type 5 is an IDR slice: a self-contained keyframe. The decoder
// rejects inter frames that arrive before one, so callers gate on this.
function isKeyframe(accessUnit: Uint8Array): boolean {
  for (let i = 0; i + 3 < accessUnit.length; i++) {
    if (accessUnit[i] !== 0 || accessUnit[i + 1] !== 0) continue
    // 3-byte start code 00 00 01, or 4-byte 00 00 00 01
    let nal = -1
    if (accessUnit[i + 2] === 1) nal = i + 3
    else if (accessUnit[i + 2] === 0 && accessUnit[i + 3] === 1) nal = i + 4
    if (nal < 0 || nal >= accessUnit.length) continue
    if ((accessUnit[nal] & 0x1f) === 5) return true
  }
  return false
}

type SenderOptions = {
  send: (base64AccessUnit: string) => Promise<void>
}

export class CallVideoSender {
  private stream: MediaStream | null = null
  private video: HTMLVideoElement | null = null
  private encoder: VideoEncoder | null = null
  private frameHandle = 0
  private running = false
  private forceKeyframe = false
  private sendQueue: string[] = []
  private flushing = false

  constructor(private opts: SenderOptions) {}

  async start(preview?: HTMLVideoElement | null): Promise<void> {
    if (this.running) return
    if (!videoCodecsSupported()) {
      throw new Error("WebCodecs unavailable in this WebView2 build")
    }

    this.stream = await navigator.mediaDevices.getUserMedia({
      video: {
        width: { ideal: 640 },
        height: { ideal: 480 },
        frameRate: { ideal: ENCODER_FRAMERATE },
      },
      audio: false,
    })

    const video = preview ?? document.createElement("video")
    video.srcObject = this.stream
    video.muted = true
    video.playsInline = true
    await video.play()
    this.video = video

    this.encoder = new VideoEncoder({
      output: (chunk) => this.onChunk(chunk),
      error: (err) => console.error("video encoder error:", err),
    })
    this.encoder.configure({
      codec: ENCODER_CODEC,
      width: this.video.videoWidth || 640,
      height: this.video.videoHeight || 480,
      bitrate: ENCODER_BITRATE,
      framerate: ENCODER_FRAMERATE,
      latencyMode: "realtime",
      avc: { format: "annexb" },
    })

    this.running = true
    this.forceKeyframe = true
    this.pump()
  }

  // requestKeyframe makes the next encoded access unit an IDR. meowcaller fires
  // this when the peer reports packet loss (WhatsApp PLI/FIR).
  requestKeyframe() {
    this.forceKeyframe = true
  }

  private pump = () => {
    if (!this.running || !this.video || !this.encoder) return
    const video = this.video
    const tick = () => {
      if (!this.running || !this.video || !this.encoder) return
      if (this.video.readyState >= 2 && this.encoder.state === "configured") {
        const keyFrame = this.forceKeyframe
        this.forceKeyframe = false
        try {
          const frame = new VideoFrame(this.video, {
            timestamp: performance.now() * 1000,
          })
          this.encoder.encode(frame, { keyFrame })
          frame.close()
        } catch (err) {
          console.error("encode frame failed:", err)
        }
      }
      this.frameHandle = this.video.requestVideoFrameCallback(tick)
    }
    if ("requestVideoFrameCallback" in video) {
      this.frameHandle = video.requestVideoFrameCallback(tick)
    } else {
      const interval = window.setInterval(tick, 1000 / ENCODER_FRAMERATE)
      this.frameHandle = -1
      ;(this as any).fallbackInterval = interval
    }
  }

  private onChunk(chunk: EncodedVideoChunk) {
    const bytes = new Uint8Array(chunk.byteLength)
    chunk.copyTo(bytes)
    this.sendQueue.push(bytesToBase64(bytes))
    if (!this.flushing) void this.flush()
  }

  // Frames are pushed serially so the Go side never interleaves access units
  // from two goroutines, and a slow IPC call cannot reorder RTP timestamps.
  private async flush() {
    this.flushing = true
    try {
      while (this.sendQueue.length > 0) {
        const next = this.sendQueue.shift()
        if (!next) continue
        try {
          await this.opts.send(next)
        } catch (err) {
          console.error("send video frame failed:", err)
        }
      }
    } finally {
      this.flushing = false
    }
  }

  async stop(): Promise<void> {
    this.running = false
    const fallback = (this as any).fallbackInterval
    if (fallback) window.clearInterval(fallback)
    if (this.video && this.frameHandle > 0) {
      this.video.cancelVideoFrameCallback(this.frameHandle)
    }
    this.frameHandle = 0
    if (this.encoder) {
      try {
        const enc = this.encoder
        this.encoder = null
        if (enc.state !== "closed") {
          enc.close()
        }
      } catch {
        // already torn down
      }
    }
    if (this.video) {
      this.video.srcObject = null
      this.video = null
    }
    if (this.stream) {
      this.stream.getTracks().forEach(t => t.stop())
      this.stream = null
    }
    this.sendQueue = []
  }
}

type ReceiverOptions = {
  canvas: HTMLCanvasElement
  onOrientation?: (quarterTurns: number) => void
}

export class CallVideoReceiver {
  private decoder: VideoDecoder | null = null
  private waitingForKeyframe = true
  private orientation = 0

  constructor(private opts: ReceiverOptions) {}

  start() {
    if (!videoCodecsSupported()) {
      throw new Error("WebCodecs unavailable in this WebView2 build")
    }
    if (this.decoder) return
    this.decoder = new VideoDecoder({
      output: (frame) => this.draw(frame),
      error: (err) => {
        console.error("video decoder error:", err)
        this.waitingForKeyframe = true
      },
    })
    this.decoder.configure({
      codec: ENCODER_CODEC,
      optimizeForLatency: true,
      avc: { format: "annexb" },
    })
  }

  setOrientation(quarterTurns: number) {
    this.orientation = ((quarterTurns % 4) + 4) % 4
  }

  push(base64AccessUnit: string) {
    if (!this.decoder || this.decoder.state !== "configured") return
    const bytes = base64ToBytes(base64AccessUnit)
    if (bytes.length === 0) return
    const key = isKeyframe(bytes)
    if (this.waitingForKeyframe) {
      if (!key) return
      this.waitingForKeyframe = false
    }
    try {
      this.decoder.decode(
        new EncodedVideoChunk({
          type: key ? "key" : "delta",
          timestamp: performance.now() * 1000,
          data: bytes,
        }),
      )
    } catch (err) {
      console.error("decode access unit failed:", err)
      this.waitingForKeyframe = true
    }
  }

  private draw(frame: VideoFrame) {
    const canvas = this.opts.canvas
    const rotate = this.orientation !== 0
    const width = rotate ? frame.displayHeight : frame.displayWidth
    const height = rotate ? frame.displayWidth : frame.displayHeight
    if (canvas.width !== width || canvas.height !== height) {
      canvas.width = width
      canvas.height = height
    }
    const ctx = canvas.getContext("2d")
    if (!ctx) {
      frame.close()
      return
    }
    ctx.save()
    if (rotate) {
      ctx.translate(width / 2, height / 2)
      ctx.rotate((this.orientation * Math.PI) / 2)
      ctx.drawImage(
        frame,
        -frame.displayWidth / 2,
        -frame.displayHeight / 2,
        frame.displayWidth,
        frame.displayHeight,
      )
    } else {
      ctx.drawImage(frame, 0, 0, width, height)
    }
    ctx.restore()
    frame.close()
  }

  stop() {
    if (this.decoder) {
      try {
        if (this.decoder.state !== "closed") this.decoder.close()
      } catch {
        // already torn down
      }
      this.decoder = null
    }
    this.waitingForKeyframe = true
  }
}

export { base64ToBytes, bytesToBase64, isKeyframe }
