import { useEffect, useMemo, useRef, useState } from 'react'

const Y0 = 0
const Y1 = 1
const MIN_X_SPAN_RATIO = 0.02 // мин. окно — 2% от всего диапазона

function clampView(x0, x1, g0, g1) {
  let span = x1 - x0
  const gSpan = g1 - g0 || 1
  const minSpan = gSpan * MIN_X_SPAN_RATIO
  if (span < minSpan) {
    const c = (x0 + x1) / 2
    span = minSpan
    x0 = c - span / 2
    x1 = c + span / 2
  }
  if (x0 < g0) {
    x1 += g0 - x0
    x0 = g0
  }
  if (x1 > g1) {
    x0 -= x1 - g1
    x1 = g1
  }
  if (x0 < g0) x0 = g0
  if (x1 > g1) x1 = g1
  return { x0, x1 }
}

function formatTick(x) {
  if (Math.abs(x) >= 100) return x.toFixed(0)
  if (Math.abs(x) >= 10) return x.toFixed(1)
  return x.toFixed(2)
}

export default function CurvesExplorer({ bank, alpha }) {
  const items = bank?.items || []
  const ref = useRef(null)

  const global = useMemo(() => {
    let minX = Infinity
    let maxX = -Infinity
    for (const it of items) {
      for (const p of it.preview || []) {
        if (p.x < minX) minX = p.x
        if (p.x > maxX) maxX = p.x
      }
    }
    if (!Number.isFinite(minX)) return { g0: 0, g1: 1 }
    const m = (maxX - minX) * 0.02 || 0.5
    return { g0: minX - m, g1: maxX + m }
  }, [items])

  const [view, setView] = useState(null)

  useEffect(() => {
    setView({ x0: global.g0, x1: global.g1 })
  }, [global.g0, global.g1])

  const W = 800
  const H = 340
  const ml = 48
  const mr = 16
  const mt = 20
  const mb = 36
  const pw = W - ml - mr
  const ph = H - mt - mb

  const toScreen = (x, y) => {
    if (!view) return { sx: ml, sy: mt + ph }
    const sx = ml + ((x - view.x0) / (view.x1 - view.x0 || 1)) * pw
    const sy = mt + ph - ((y - Y0) / (Y1 - Y0)) * ph
    return { sx, sy }
  }

  const zoomAt = (clientX, factor) => {
    const el = ref.current
    if (!el || !view) return
    const rect = el.getBoundingClientRect()
    const t = (clientX - rect.left - (ml / W) * rect.width) / ((pw / W) * rect.width)
    const tClamped = Math.min(1, Math.max(0, t))
    const cx = view.x0 + tClamped * (view.x1 - view.x0)
    const span = (view.x1 - view.x0) * factor
    const x0 = cx - tClamped * span
    const x1 = cx + (1 - tClamped) * span
    setView(clampView(x0, x1, global.g0, global.g1))
  }

  const onWheel = (e) => {
    e.preventDefault()
    const factor = e.deltaY > 0 ? 1.14 : 1 / 1.14
    zoomAt(e.clientX, factor)
  }

  const drag = useRef(null)

  const onPointerDown = (e) => {
    drag.current = { x: e.clientX, view0: view.x0, view1: view.x1 }
    e.currentTarget.setPointerCapture(e.pointerId)
  }

  const onPointerMove = (e) => {
    if (!drag.current || !view) return
    const el = ref.current
    const rect = el.getBoundingClientRect()
    const dxPx = e.clientX - drag.current.x
    const span = drag.current.view1 - drag.current.view0
    const dxData = (-dxPx / rect.width) * (W / pw) * span
    setView(
      clampView(
        drag.current.view0 + dxData,
        drag.current.view1 + dxData,
        global.g0,
        global.g1,
      ),
    )
  }

  const onPointerUp = () => {
    drag.current = null
  }

  const resetView = () => setView({ x0: global.g0, x1: global.g1 })

  const zoomPct = view
    ? Math.round(((global.g1 - global.g0) / (view.x1 - view.x0 || 1)) * 100)
    : 100

  const xTicks = useMemo(() => {
    if (!view) return []
    const span = view.x1 - view.x0
    const raw = span / 6
    const mag = Math.pow(10, Math.floor(Math.log10(raw)))
    const step = Math.ceil(raw / mag) * mag
    const start = Math.ceil(view.x0 / step) * step
    const ticks = []
    for (let x = start; x <= view.x1; x += step) ticks.push(x)
    return ticks.slice(0, 8)
  }, [view])

  if (!items.length || !view) return null

  const clipId = 'plot-clip'

  return (
    <div className="explorer">
      <div className="explorer-toolbar">
        <span className="explorer-hint">Колёсико — приближение к курсору · перетаскивание — вдоль оси x</span>
        <div className="explorer-tools">
          <button type="button" className="btn btn-sm btn-ghost" onClick={() => zoomAt(ref.current?.getBoundingClientRect().left + pw / 2 + ml, 1 / 1.35)}>
            +
          </button>
          <button type="button" className="btn btn-sm btn-ghost" onClick={() => zoomAt(ref.current?.getBoundingClientRect().left + pw / 2 + ml, 1.35)}>
            −
          </button>
          <button type="button" className="btn btn-sm btn-ghost" onClick={resetView}>
            Вся область
          </button>
          <span className="explorer-zoom">{zoomPct}%</span>
        </div>
      </div>
      <div
        className="explorer-viewport"
        ref={ref}
        onWheel={onWheel}
        onPointerDown={onPointerDown}
        onPointerMove={onPointerMove}
        onPointerUp={onPointerUp}
        onPointerLeave={onPointerUp}
      >
        <svg viewBox={`0 0 ${W} ${H}`} width="100%" height={H} className="explorer-svg">
          <defs>
            <clipPath id={clipId}>
              <rect x={ml} y={mt} width={pw} height={ph} />
            </clipPath>
            <linearGradient id="plotBg" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor="#faf8ff" />
              <stop offset="100%" stopColor="#f3f0fa" />
            </linearGradient>
          </defs>
          <rect width={W} height={H} fill="url(#plotBg)" rx="10" />
          <rect x={ml} y={mt} width={pw} height={ph} fill="#fff" stroke="#ddd6e8" strokeWidth="1" rx="4" />

          {[0, 0.2, 0.4, 0.6, 0.8, 1].map((gy) => {
            const y = mt + ph - gy * ph
            return (
              <g key={gy}>
                <line x1={ml} y1={y} x2={ml + pw} y2={y} stroke="#eeeaf5" strokeWidth="1" />
                <text x={ml - 8} y={y + 4} textAnchor="end" className="axis-tick">
                  {gy.toFixed(1)}
                </text>
              </g>
            )
          })}

          {xTicks.map((x) => {
            const { sx } = toScreen(x, 0)
            return (
              <g key={x}>
                <line x1={sx} y1={mt + ph} x2={sx} y2={mt + ph + 5} stroke="#aaa" />
                <text x={sx} y={mt + ph + 18} textAnchor="middle" className="axis-tick">
                  {formatTick(x)}
                </text>
              </g>
            )
          })}

          <text x={ml + pw / 2} y={H - 4} textAnchor="middle" className="axis-label">
            x
          </text>
          <text
            x={14}
            y={mt + ph / 2}
            textAnchor="middle"
            transform={`rotate(-90, 14, ${mt + ph / 2})`}
            className="axis-label"
          >
            μ̂
          </text>

          <g clipPath={`url(#${clipId})`}>
            {alpha > 0 && alpha <= 1 && (
              <>
                <rect
                  x={ml}
                  y={toScreen(0, alpha).sy}
                  width={pw}
                  height={mt + ph - toScreen(0, alpha).sy}
                  fill="rgba(234, 88, 12, 0.06)"
                />
                <line
                  x1={ml}
                  y1={toScreen(0, alpha).sy}
                  x2={ml + pw}
                  y2={toScreen(0, alpha).sy}
                  stroke="#ea580c"
                  strokeWidth="1.5"
                  strokeDasharray="6 4"
                />
              </>
            )}
            {items.map((it) => {
              const pts = (it.preview || []).filter((p) => p.x >= view.x0 && p.x <= view.x1)
              if (pts.length < 2) return null
              const d = pts
                .map((p, i) => {
                  const { sx, sy } = toScreen(p.x, p.muNorm)
                  return `${i === 0 ? 'M' : 'L'} ${sx.toFixed(2)} ${sy.toFixed(2)}`
                })
                .join(' ')
              return (
                <path
                  key={it.index}
                  d={d}
                  fill="none"
                  stroke="#6366f1"
                  strokeWidth="1"
                  opacity="0.28"
                  strokeLinejoin="round"
                  strokeLinecap="round"
                />
              )
            })}
          </g>

          <text x={ml + 8} y={mt + 14} className="svg-label">
            {items.length} кривых · μ̂(x)=μ/max μ
          </text>
          {alpha > 0 && (
            <text x={ml + pw - 8} y={toScreen(0, alpha).sy - 6} textAnchor="end" className="svg-label alpha">
              α = {alpha.toFixed(2)}
            </text>
          )}
        </svg>
      </div>
    </div>
  )
}
